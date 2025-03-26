package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"github.com/gorilla/mux"
)

type UsersService interface {
	UpdateUserRating(ctx context.Context, uuid string, newScore int) error
	GetTopUsers(ctx context.Context, count int) ([]domain.UserRating, error)
}

type ProfileHandler struct {
	usersService UsersService
	logger       *zap.SugaredLogger
}

func NewProfileHandler(usersService UsersService, logger *zap.SugaredLogger) (*ProfileHandler, error) {
	return &ProfileHandler{
		usersService: usersService,
		logger:       logger,
	}, nil
}

type UpdateRatingRequest struct {
	Score int `json:"score"`
}

func (h *ProfileHandler) UpdateRating(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	uuid := mux.Vars(req)["uuid"]

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

	var reqData UpdateRatingRequest
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

	err = h.usersService.UpdateUserRating(ctx, uuid, reqData.Score)
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

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   nil,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}
