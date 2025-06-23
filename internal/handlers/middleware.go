package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"go.uber.org/zap"
)

type MiddlewareHandler struct {
	serverAddr  string
	authService AuthService
	logger      *zap.SugaredLogger
}

func NewMiddlewareHandler(
	serverAddr string,
	authService AuthService,
	logger *zap.SugaredLogger) (*MiddlewareHandler, error) {

	return &MiddlewareHandler{
		serverAddr:  serverAddr,
		authService: authService,
		logger:      logger,
	}, nil
}

func (h *MiddlewareHandler) Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("http://%s", h.serverAddr))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, "+
			"Accept-Encoding, X-CSRF-Token, Authorization")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (h *MiddlewareHandler) Panic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				h.logger.Errorf("panic: %v", err)

				err = WriteResponse(w, ResponseData{
					Status: http.StatusInternalServerError,
					Data:   nil,
				})
				if err != nil {
					h.logger.Errorf("unable to decode http request: %v", err)
				}
			}
		}()

		next.ServeHTTP(w, req)
	})
}

func (h *MiddlewareHandler) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
			err = WriteResponse(w, ResponseData{
				Status: http.StatusBadRequest,
				Data:   nil,
			})
			if err != nil {
				h.logger.Errorf("unable to decode http request: %v", err)
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

		next.ServeHTTP(w, req)
	})
}

func (h *MiddlewareHandler) Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vkIdStr, err := req.Cookie("vkid")
		if err != nil {
			vkIdStr = &http.Cookie{}
		}

		vkId, err := strconv.Atoi(vkIdStr.Value)
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

		ctx := req.Context()

		ok, err := h.authService.IsAdmin(ctx, vkId)
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

		if !ok {
			err = WriteResponse(w, ResponseData{
				Status: http.StatusForbidden,
				Data:   nil,
			})
			if err != nil {
				h.logger.Errorf("unable to decode http request: %v", err)
			}
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (h *MiddlewareHandler) Csrf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			token := req.Header.Get("X-CSRF-TOKEN")

			err := h.authService.ValidateCsfrToken(token)
			if err != nil {
				err = WriteResponse(w, ResponseData{
					Status: http.StatusUnauthorized,
					Data:   nil,
				})
				if err != nil {
					h.logger.Errorf("unable to decode http request: %v", err)
				}
				return
			}
		}

		next.ServeHTTP(w, req)
	})
}
