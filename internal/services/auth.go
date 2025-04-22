package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"go.uber.org/zap"
)

type AuthStorage interface {
	SignIn(ctx context.Context, signInData domain.SignInData) error
	Check(ctx context.Context, vkid int, accessToken string) (bool, error)
	Logout(ctx context.Context, vkid int) error
	GetUserData(ctx context.Context, vkid int) (domain.UserHeader, error)
}

type AuthService struct {
	authStorage AuthStorage
	logger      *zap.SugaredLogger

	oauthUrl string
	clientId string
}

func NewAuthService(authStorage AuthStorage, logger *zap.SugaredLogger) (*AuthService, error) {
	return &AuthService{
		authStorage: authStorage,
		logger:      logger,
		oauthUrl:    "https://id.vk.com/oauth2",
		clientId:    "53445034",
	}, nil
}

type errResponse struct {
	Err   string `json:"error"`
	Desc  string `json:"error_description"`
	State string `json:"state,omitempty"`
}

type authResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	UserId       int    `json:"user_id"`
	State        string `json:"state"`
	Scope        string `json:"scope"`
}

type userInfo struct {
	UserId    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
}

type infoResponse struct {
	User userInfo `json:"user"`
}

func (s *AuthService) SignIn(ctx context.Context, req domain.SignInRequest) (domain.SignInData, error) {
	authUrl := s.oauthUrl + "/auth"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code_verifier", req.CodeVerifier)
	data.Set("redirect_uri", req.RedirectUrl)
	data.Set("code", req.Code)
	data.Set("client_id", s.clientId)
	data.Set("device_id", req.DeviceId)
	data.Set("state", req.State)

	authResp := &authResponse{}
	errResp, err := s.apiCall(authUrl, data, authResp)
	if err != nil {
		s.logger.Errorf("failed to make api call: %v", err)
		return domain.SignInData{}, err
	}

	if errResp.Err != "" {
		s.logger.Errorf("error recieved from api: %v", err)
		return domain.SignInData{}, fmt.Errorf(errResp.Err)
	}

	if req.State != authResp.State {
		s.logger.Errorf("response state doesn't match: %v", err)
		return domain.SignInData{}, fmt.Errorf("response state doesn't match")
	}

	infoUrl := s.oauthUrl + "/user_info"
	data = url.Values{}
	data.Set("client_id", s.clientId)
	data.Set("access_token", authResp.AccessToken)

	infoResp := &infoResponse{}
	errResp, err = s.apiCall(infoUrl, data, infoResp)
	if err != nil {
		s.logger.Errorf("failed to make api call: %v", err)
		return domain.SignInData{}, err
	}

	if errResp.Err != "" {
		s.logger.Errorf("error recieved from api: %v", err)
		return domain.SignInData{}, fmt.Errorf(errResp.Err)
	}

	signInData := domain.SignInData{
		User: domain.UserHeader{
			VkId:   authResp.UserId,
			Name:   fmt.Sprintf("%s %c.", infoResp.User.FirstName, []rune(infoResp.User.LastName)[0]),
			Avatar: infoResp.User.Avatar,
		},
		AccessToken: authResp.AccessToken,
	}

	err = s.authStorage.SignIn(ctx, signInData)
	if err != nil {
		s.logger.Errorf("failed sign in user: %v", err)
		return domain.SignInData{}, err
	}

	signInData.RefreshToken = authResp.RefreshToken

	return signInData, nil
}

func (s *AuthService) Check(ctx context.Context, session domain.SignInData, state, deviceId string) (domain.SignInData, error) {
	ok, err := s.authStorage.Check(ctx, session.User.VkId, session.AccessToken)
	if err != nil {
		s.logger.Errorf("failed check user: %v", err)
		return domain.SignInData{}, err
	}
	if ok {
		userHeader, err := s.authStorage.GetUserData(ctx, session.User.VkId)
		if err != nil {
			s.logger.Errorf("failed to get user: %v", err)
			return domain.SignInData{}, err
		}

		return domain.SignInData{
			User:         userHeader,
			AccessToken:  session.AccessToken,
			RefreshToken: session.RefreshToken,
		}, nil
	}

	authUrl := s.oauthUrl + "/auth"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", session.RefreshToken)
	data.Set("client_id", s.clientId)
	data.Set("device_id", deviceId)
	data.Set("state", state)
	data.Set("scope", "vkid.personal_info email")

	refreshResp := &authResponse{}
	errResp, err := s.apiCall(authUrl, data, refreshResp)
	if err != nil {
		s.logger.Errorf("failed to make api call: %v", err)
		return domain.SignInData{}, err
	}

	if errResp.Err != "" {
		s.logger.Errorf("error recieved from api: %v", err)
		return domain.SignInData{}, fmt.Errorf("unathorized")
	}

	if state != refreshResp.State {
		return domain.SignInData{}, fmt.Errorf("response state doesn't match")
	}

	infoUrl := s.oauthUrl + "/user_info"
	data = url.Values{}
	data.Set("client_id", s.clientId)
	data.Set("access_token", refreshResp.AccessToken)

	infoResp := &infoResponse{}
	errResp, err = s.apiCall(infoUrl, data, infoResp)
	if err != nil {
		s.logger.Errorf("failed to make api call: %v", err)
		return domain.SignInData{}, err
	}

	if errResp.Err != "" {
		s.logger.Errorf("error recieved from api: %v", err)
		return domain.SignInData{}, fmt.Errorf("unathorized")
	}

	signInData := domain.SignInData{
		User: domain.UserHeader{
			VkId:   refreshResp.UserId,
			Name:   fmt.Sprintf("%s %c.", infoResp.User.FirstName, []rune(infoResp.User.LastName)[0]),
			Avatar: infoResp.User.Avatar,
		},
		AccessToken:  refreshResp.AccessToken,
		RefreshToken: refreshResp.RefreshToken,
	}

	err = s.authStorage.SignIn(ctx, signInData)
	if err != nil {
		s.logger.Errorf("failed sign in user: %v", err)
		return domain.SignInData{}, err
	}

	return signInData, nil
}

func (s *AuthService) Logout(ctx context.Context, session domain.SignInData, state, deviceId string) error {
	newSession, err := s.Check(ctx, session, state, deviceId)
	if err != nil {
		return fmt.Errorf("unauthorized")
	}

	authUrl := s.oauthUrl + "/logout"
	data := url.Values{}
	data.Set("client_id", s.clientId)
	data.Set("access_token", newSession.AccessToken)

	authResp := &authResponse{}
	errResp, err := s.apiCall(authUrl, data, authResp)
	if err != nil {
		s.logger.Errorf("failed to make api call: %v", err)
		return err
	}

	if errResp.Err != "" {
		s.logger.Errorf("error recieved from api: %v", err)
		return fmt.Errorf(errResp.Err)
	}

	err = s.authStorage.Logout(ctx, newSession.User.VkId)
	if err != nil {
		s.logger.Errorf("failed to logout user: %v", err)
		return err
	}

	return nil
}

func (s *AuthService) apiCall(apiUrl string, urlValues url.Values, target interface{}) (errResponse, error) {
	client := &http.Client{}
	r, err := http.NewRequest(http.MethodPost, apiUrl, strings.NewReader(urlValues.Encode()))
	if err != nil {
		s.logger.Errorf("failed to create http request: %v", err)
		return errResponse{}, err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		s.logger.Errorf("failed to send http request: %v", err)
		return errResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errResponse{}, fmt.Errorf("unexpected wrong status")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("failed read body: %v", err)
		return errResponse{}, err
	}

	body := io.NopCloser(bytes.NewReader(data))
	err = json.NewDecoder(body).Decode(target)
	if err != nil {
		s.logger.Errorf("failed to decode http response: %v", err)
		return errResponse{}, err
	}

	body = io.NopCloser(bytes.NewReader(data))
	respErr := &errResponse{}
	err = json.NewDecoder(body).Decode(respErr)
	if err != nil {
		s.logger.Errorf("failed to decode http response: %v", err)
		return errResponse{}, err
	}

	return *respErr, nil
}
