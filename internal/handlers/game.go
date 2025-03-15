package handlers

import "net/http"

type GameHandler struct{}

func NewGameHandler() (*GameHandler, error) {
	return &GameHandler{}, nil
}

func (h *GameHandler) GetTopRating(w http.ResponseWriter, req *http.Request) {}
