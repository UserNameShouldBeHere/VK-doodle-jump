package handlers

import (
	"net/http"
	"strconv"

	"go.uber.org/zap"
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

func (h *GameHandler) GetTopUsers(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	vkIdStr, err := req.Cookie("vkid")
	if err != nil {
		vkIdStr = &http.Cookie{}
	}
	vkId, err := strconv.Atoi(vkIdStr.Value)
	if err != nil {
		h.logger.Errorf("failed to convert vkid to int: %v", err)

		vkId = 0
	}

	count, err := strconv.Atoi(req.URL.Query().Get("count"))
	if err != nil || count <= 0 {
		count = 10
	}

	usersTop, err := h.usersService.GetTopUsers(ctx, vkId, count)
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
		Data:   usersTop,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}
