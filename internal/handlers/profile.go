package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"github.com/gorilla/mux"
)

type UsersService interface {
	UpdateUserRating(ctx context.Context, vkid int, newScore int) error
	GetTopUsers(ctx context.Context, count int) ([]domain.LeagueTopUsers, error)
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
	vkIdStr := mux.Vars(req)["vkid"]

	vkid, err := strconv.Atoi(vkIdStr)
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

	var reqData UpdateRatingRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		h.logger.Errorf("unable to unmarshall request body: %v", err)
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	err = h.usersService.UpdateUserRating(ctx, vkid, reqData.Score)
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
