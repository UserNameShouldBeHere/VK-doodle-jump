package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
)

type GameHandler struct {
	usersService UsersService
	logger       *zap.SugaredLogger
}

func NewGameHandler(usersService UsersService, logger *zap.SugaredLogger) (*GameHandler, error) {
	return &GameHandler{
		usersService: usersService,
		logger:       logger,
	}, nil
}

type UsersTopResponse struct {
	Users []domain.UserRating `json:"users"`
}

func (h *GameHandler) GetTopUsers(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	offset, err := strconv.Atoi(mux.Vars(req)["offset"])
	if err != nil || offset <= 0 {
		offset = 10
	}

	usersTop, err := h.usersService.GetTopUsers(ctx, offset)
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
		Data: UsersTopResponse{
			Users: usersTop,
		},
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}
