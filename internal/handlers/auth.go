package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"go.uber.org/zap"
)

type AuthService interface {
	SignIn(ctx context.Context, req domain.SignInRequest) (domain.SignInData, error)
	Check(ctx context.Context, session domain.SignInData, state, deviceId string) (domain.SignInData, error)
	Logout(ctx context.Context, session domain.SignInData, state, deviceId string) error
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
	VkId   int    `json:"vkid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var reqData domain.SignInRequest
	err = json.Unmarshal(body, &reqData)
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

	signInData, err := h.authService.SignIn(ctx, reqData)
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

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

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

type checkRequest struct {
	VkId     int    `json:"vkid"`
	DeviceId string `json:"device_id"`
	State    string `json:"state"`
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
	fmt.Println(refreshToken)

	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var reqData checkRequest
	err = json.Unmarshal(body, &reqData)
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

	checkData := domain.SignInData{
		User: domain.UserHeader{
			VkId: reqData.VkId,
		},
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	}

	signInData, err := h.authService.Check(ctx, checkData, reqData.State, reqData.DeviceId)
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

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

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

	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var reqData checkRequest
	err = json.Unmarshal(body, &reqData)
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

	checkData := domain.SignInData{
		User: domain.UserHeader{
			VkId: reqData.VkId,
		},
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	}

	err = h.authService.Logout(ctx, checkData, reqData.State, reqData.DeviceId)
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

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   nil,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}
