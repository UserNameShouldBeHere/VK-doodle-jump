package handlers

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type MiddlewareHandler struct {
	serverAddr string
	logger     *zap.SugaredLogger
}

func NewMiddlewareHandler(serverAddr string, logger *zap.SugaredLogger) (*MiddlewareHandler, error) {
	return &MiddlewareHandler{
		serverAddr: serverAddr,
		logger:     logger,
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
