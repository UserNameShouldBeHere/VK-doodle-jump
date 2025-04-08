package handlers

import (
	"fmt"
	"net/http"
)

type MiddlewareHandler struct {
	serverAddr string
}

func NewMiddlewareHandler(serverAddr string) (*MiddlewareHandler, error) {
	return &MiddlewareHandler{
		serverAddr: serverAddr,
	}, nil
}

func (h *MiddlewareHandler) Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("http://%s", h.serverAddr))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, "+
			"Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
