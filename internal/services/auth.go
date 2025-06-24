package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/VK-doodle-jump/internal/errors"
	"github.com/golang-jwt/jwt"
)

type AuthStorage interface {
	SignIn(ctx context.Context, signInData domain.SignInData) (bool, error)
	Check(ctx context.Context, vkid int, accessToken string) (bool, error)
	Logout(ctx context.Context, vkid int) error
	GetUserData(ctx context.Context, vkid int) (domain.UserHeader, error)
	IsAdmin(ctx context.Context, vkid int) (bool, error)
}

type AuthService struct {
	authStorage AuthStorage
	logger      *zap.SugaredLogger

	oauthUrl string
	clientId string
	csrfKey  []byte
}

func NewAuthService(authStorage AuthStorage, logger *zap.SugaredLogger) (*AuthService, error) {
	csrfKey := make([]byte, 16)
	_, err := rand.Read(csrfKey)
	if err != nil {
		logger.Errorf("(adminShopService.NewAdminShopService): %w", err)
		return nil, fmt.Errorf("(adminShopService.NewAdminShopService): %w", err)
	}

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

func (s *AuthService) SignIn(ctx context.Context, req domain.SignInRequest) (domain.SignInData, bool, error) {
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
		s.logger.Errorf("(authService.SignIn) %v: %v", customErrors.ErrFailedToCallApi, err)
		return domain.SignInData{}, false, fmt.Errorf("(authService.SignIn) %w: %w", customErrors.ErrFailedToCallApi, err)
	}

	if errResp.Err != "" {
		s.logger.Errorf("(authService.SignIn) %v: %v", customErrors.ErrRecievedFromApi, errResp.Desc)
		return domain.SignInData{}, false, fmt.Errorf("(authService.SignIn) %w: %v", customErrors.ErrRecievedFromApi, errResp.Err)
	}

	if req.State != authResp.State {
		s.logger.Errorf("(authService.SignIn) %v", customErrors.ErrStateMismatch)
		return domain.SignInData{}, false, fmt.Errorf("(authService.SignIn) %w", customErrors.ErrStateMismatch)
	}

	infoUrl := s.oauthUrl + "/user_info"
	data = url.Values{}
	data.Set("client_id", s.clientId)
	data.Set("access_token", authResp.AccessToken)

	infoResp := &infoResponse{}
	errResp, err = s.apiCall(infoUrl, data, infoResp)
	if err != nil {
		s.logger.Errorf("(authService.SignIn) %v: %v", customErrors.ErrFailedToCallApi, err)
		return domain.SignInData{}, false, fmt.Errorf("(authService.SignIn) %w: %w", customErrors.ErrFailedToCallApi, err)
	}

	if errResp.Err != "" {
		s.logger.Errorf("(authService.SignIn) %v: %v", customErrors.ErrRecievedFromApi, errResp.Desc)
		return domain.SignInData{}, false, fmt.Errorf("(authService.SignIn) %w: %v", customErrors.ErrRecievedFromApi, errResp.Err)
	}

	signInData := domain.SignInData{
		User: domain.UserHeader{
			VkId:   authResp.UserId,
			Name:   fmt.Sprintf("%s %c.", infoResp.User.FirstName, []rune(infoResp.User.LastName)[0]),
			Avatar: infoResp.User.Avatar,
		},
		AccessToken: authResp.AccessToken,
	}

	isFirstTime, err := s.authStorage.SignIn(ctx, signInData)
	if err != nil {
		s.logger.Errorf("(authService.SignIn) %v", err)
		return domain.SignInData{}, false, fmt.Errorf("(authService.SignIn) %w", err)
	}

	signInData.RefreshToken = authResp.RefreshToken

	return signInData, isFirstTime, nil
}

func (s *AuthService) Check(
	ctx context.Context,
	session domain.SignInData,
	state string,
	deviceId string) (domain.SignInData, error) {

	ok, err := s.authStorage.Check(ctx, session.User.VkId, session.AccessToken)
	if err != nil {
		s.logger.Errorf("(authService.Check) %v", err)
		return domain.SignInData{}, fmt.Errorf("(authService.Check) %w", err)
	}
	if ok {
		userHeader, err := s.authStorage.GetUserData(ctx, session.User.VkId)
		if err != nil {
			s.logger.Errorf("(authService.Check) %v", err)
			return domain.SignInData{}, fmt.Errorf("(authService.Check) %w", err)
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
		s.logger.Errorf("(authService.Check) %v: %v", customErrors.ErrFailedToCallApi, err)
		return domain.SignInData{}, fmt.Errorf("(authService.Check) %w: %w", customErrors.ErrFailedToCallApi, err)
	}

	if errResp.Err != "" {
		s.logger.Errorf("(authService.Check) %v: %v", customErrors.ErrRecievedFromApi, errResp.Desc)
		return domain.SignInData{}, fmt.Errorf("(authService.Check) %w: %v", customErrors.ErrRecievedFromApi, errResp.Err)
	}

	if state != refreshResp.State {
		s.logger.Errorf("(authService.Check) %v", customErrors.ErrStateMismatch)
		return domain.SignInData{}, fmt.Errorf("(authService.Check) %w", customErrors.ErrStateMismatch)
	}

	infoUrl := s.oauthUrl + "/user_info"
	data = url.Values{}
	data.Set("client_id", s.clientId)
	data.Set("access_token", refreshResp.AccessToken)

	infoResp := &infoResponse{}
	errResp, err = s.apiCall(infoUrl, data, infoResp)
	if err != nil {
		s.logger.Errorf("(authService.Check) %v: %v", customErrors.ErrFailedToCallApi, err)
		return domain.SignInData{}, fmt.Errorf("(authService.Check) %w: %w", customErrors.ErrFailedToCallApi, err)
	}

	if errResp.Err != "" {
		s.logger.Errorf("(authService.Check) %v: %v", customErrors.ErrRecievedFromApi, errResp.Desc)
		return domain.SignInData{}, fmt.Errorf("(authService.Check) %w: %v", customErrors.ErrRecievedFromApi, errResp.Err)
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

	_, err = s.authStorage.SignIn(ctx, signInData)
	if err != nil {
		s.logger.Errorf("(authService.Check) %v", err)
		return domain.SignInData{}, fmt.Errorf("(authService.Check) %w", err)
	}

	return signInData, nil
}

func (s *AuthService) Logout(ctx context.Context, session domain.SignInData, state, deviceId string) error {
	newSession, err := s.Check(ctx, session, state, deviceId)
	if err != nil {
		s.logger.Errorf("(authService.Logout) %v", err)
		return fmt.Errorf("(authService.Logout) %w", err)
	}

	authUrl := s.oauthUrl + "/logout"
	data := url.Values{}
	data.Set("client_id", s.clientId)
	data.Set("access_token", newSession.AccessToken)

	authResp := &authResponse{}
	errResp, err := s.apiCall(authUrl, data, authResp)
	if err != nil {
		s.logger.Errorf("(authService.Logout) %v: %v", customErrors.ErrFailedToCallApi, err)
		return fmt.Errorf("(authService.Logout) %w: %w", customErrors.ErrFailedToCallApi, err)
	}

	if errResp.Err != "" {
		s.logger.Errorf("(authService.Logout) %v: %v", customErrors.ErrRecievedFromApi, errResp.Desc)
		return fmt.Errorf("(authService.Logout) %w: %v", customErrors.ErrRecievedFromApi, errResp.Err)
	}

	err = s.authStorage.Logout(ctx, newSession.User.VkId)
	if err != nil {
		s.logger.Errorf("(authService.Logout) %v", err)
		return fmt.Errorf("(authService.Logout) %w", err)
	}

	return nil
}

func (s *AuthService) IsAdmin(ctx context.Context, vkid int) (bool, error) {
	ok, err := s.authStorage.IsAdmin(ctx, vkid)
	if err != nil {
		s.logger.Errorf("(authService.IsAdmin) %v", err)
		return false, fmt.Errorf("(authService.IsAdmin) %w", err)
	}

	return ok, nil
}

func (s *AuthService) apiCall(apiUrl string, urlValues url.Values, target interface{}) (errResponse, error) {
	client := &http.Client{}
	r, err := http.NewRequest(http.MethodPost, apiUrl, strings.NewReader(urlValues.Encode()))
	if err != nil {
		return errResponse{}, fmt.Errorf("failed to create http request: %w", err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return errResponse{}, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errResponse{}, fmt.Errorf("non 200 response code")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errResponse{}, fmt.Errorf("failed read body: %w", err)
	}

	body := io.NopCloser(bytes.NewReader(data))
	err = json.NewDecoder(body).Decode(target)
	if err != nil {
		return errResponse{}, fmt.Errorf("failed to decode http response: %w", err)
	}

	body = io.NopCloser(bytes.NewReader(data))
	respErr := &errResponse{}
	err = json.NewDecoder(body).Decode(respErr)
	if err != nil {
		return errResponse{}, fmt.Errorf("failed to decode http response: %w", err)
	}

	return *respErr, nil
}

func (s *AuthService) CreateCsrfToken() (string, error) {
	claims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 6).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.csrfKey)
	if err != nil {
		return "", fmt.Errorf("(AuthHandler.createToken): %w", err)
	}

	return signedToken, nil
}

func (s *AuthService) ValidateCsfrToken(token string) error {
	parsedToken, err := jwt.Parse(token,
		func(token *jwt.Token) (interface{}, error) {
			return s.csrfKey, nil
		},
	)
	if err != nil || !parsedToken.Valid {
		return fmt.Errorf("(AuthHandler.ValidateCsfrToken) invalid csrf token: %w", err)
	}

	return nil
}
