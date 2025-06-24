package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"go.uber.org/zap"
)

type AuthService interface {
	SignIn(ctx context.Context, req domain.SignInRequest) (domain.SignInData, bool, error)
	Check(ctx context.Context, session domain.SignInData, state, deviceId string) (domain.SignInData, error)
	Logout(ctx context.Context, session domain.SignInData, state, deviceId string) error
	IsAdmin(ctx context.Context, vkid int) (bool, error)
	CreateCsrfToken() (string, error)
	ValidateCsfrToken(token string) error
}

type AuthHandler struct {
	authService AuthService
	logger      *zap.SugaredLogger
}

func NewAuthHandler(authService AuthService, logger *zap.SugaredLogger) (*AuthHandler, error) {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}, nil
}

type signInResponse struct {
	VkId        int    `json:"vkid"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	IsFirstTime bool   `json:"is_first_time"`
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		h.logger.Errorf("unable to read request body: %v", err)
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	var reqData domain.SignInRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		h.logger.Errorf("unable to unmarshall request data: %v", err)
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	signInData, isFirstTime, err := h.authService.SignIn(ctx, reqData)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	accessCookie := &http.Cookie{
		Name:     "access",
		Value:    signInData.AccessToken,
		Path:     "/api/v1",
		MaxAge:   3600,
		HttpOnly: true,
		// Secure:   true,
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    signInData.RefreshToken,
		Path:     "/api/v1",
		MaxAge:   3600 * 24 * 180,
		HttpOnly: true,
		// Secure:   true,
	}
	vkIdCookie := &http.Cookie{
		Name:     "vkid",
		Value:    fmt.Sprintf("%d", signInData.User.VkId),
		Path:     "/api/v1",
		MaxAge:   3600,
		HttpOnly: true,
		// Secure:   true,
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, vkIdCookie)

	token, err := h.authService.CreateCsrfToken()
	if err != nil {
		h.logger.Errorf("failed to generate csrf token: %v", err)

		err = WriteResponse(w, ResponseData{
			Status: http.StatusInternalServerError,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	w.Header().Set("X-CSRF-TOKEN", token)

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data: signInResponse{
			VkId:        signInData.User.VkId,
			Name:        signInData.User.Name,
			Avatar:      signInData.User.Avatar,
			IsFirstTime: isFirstTime,
		},
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

func (h *AuthHandler) Check(w http.ResponseWriter, req *http.Request) {
	accessToken, err := req.Cookie("access")
	if err != nil {
		accessToken = &http.Cookie{}
	}
	refreshToken, err := req.Cookie("refresh")
	if err != nil {
		refreshToken = &http.Cookie{}
	}
	vkIdStr, err := req.Cookie("vkid")
	if err != nil {
		vkIdStr = &http.Cookie{}
	}
	deviceId, err := req.Cookie("device_id")
	if err != nil {
		deviceId = &http.Cookie{}
	}

	vkId, err := strconv.Atoi(vkIdStr.Value)
	if err != nil {
		h.logger.Errorf("failed to convert vkid to int: %v", err)
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	ctx := req.Context()

	checkData := domain.SignInData{
		User: domain.UserHeader{
			VkId: vkId,
		},
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	}

	signInData, err := h.authService.Check(ctx, checkData, genState(50), deviceId.Value)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusUnauthorized,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	accessCookie := &http.Cookie{
		Name:     "access",
		Value:    signInData.AccessToken,
		Path:     "/api/v1",
		MaxAge:   3600,
		HttpOnly: true,
		// Secure:   true,
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    signInData.RefreshToken,
		Path:     "/api/v1",
		MaxAge:   3600 * 24 * 180,
		HttpOnly: true,
		// Secure:   true,

	}
	vkIdCookie := &http.Cookie{
		Name:     "vkid",
		Value:    fmt.Sprintf("%d", signInData.User.VkId),
		Path:     "/api/v1",
		MaxAge:   3600,
		HttpOnly: true,
		// Secure:   true,
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, vkIdCookie)

	token, err := h.authService.CreateCsrfToken()
	if err != nil {
		h.logger.Errorf("failed to generate csrf token: %v", err)

		err = WriteResponse(w, ResponseData{
			Status: http.StatusInternalServerError,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	w.Header().Set("X-CSRF-TOKEN", token)

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data: signInResponse{
			VkId:   signInData.User.VkId,
			Name:   signInData.User.Name,
			Avatar: signInData.User.Avatar,
		},
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, req *http.Request) {
	accessToken, err := req.Cookie("access")
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusUnauthorized,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to write response: %v", err)
		}
		return
	}
	refreshToken, err := req.Cookie("refresh")
	if err != nil {
		refreshToken = &http.Cookie{}
	}
	vkIdStr, err := req.Cookie("vkid")
	if err != nil {
		vkIdStr = &http.Cookie{}
	}
	deviceId, err := req.Cookie("device_id")
	if err != nil {
		deviceId = &http.Cookie{}
	}

	vkId, err := strconv.Atoi(vkIdStr.Value)
	if err != nil {
		h.logger.Errorf("failed to convert vkid to int: %v", err)
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	ctx := req.Context()

	checkData := domain.SignInData{
		User: domain.UserHeader{
			VkId: vkId,
		},
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	}

	err = h.authService.Logout(ctx, checkData, genState(50), deviceId.Value)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusInternalServerError,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	accessCookie := &http.Cookie{
		Name:     "access",
		Value:    "",
		Path:     "/api/v1",
		MaxAge:   0,
		HttpOnly: true,
		// Secure:   true,
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/api/v1",
		MaxAge:   0,
		HttpOnly: true,
		// Secure:   true,
	}
	vkIdCookie := &http.Cookie{
		Name:     "vkid",
		Value:    "",
		Path:     "/api/v1",
		MaxAge:   0,
		HttpOnly: true,
		// Secure:   true,
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, vkIdCookie)

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   nil,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

func genState(length int) string {
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"

	str := make([]rune, length)
	for i := range length {
		str[i] = rune(charSet[rand.Intn(len(charSet))])
	}

	return string(str)
}
