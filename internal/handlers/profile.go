package handlers

import "net/http"

type ProfileHandler struct{}

func NewProfileHandler() (*ProfileHandler, error) {
	return &ProfileHandler{}, nil
}

func (h *ProfileHandler) GetRating(w http.ResponseWriter, req *http.Request) {}

func (h *ProfileHandler) UpdateRating(w http.ResponseWriter, req *http.Request) {}
