package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"go.uber.org/zap"
)

type ShopService interface {
	GetTasks(ctx context.Context, vkid int) ([]domain.TaskData, error)
	GetCurrentGiftaway(ctx context.Context) (domain.Giftaway, error)
}

type ShopHandler struct {
	shopService ShopService
	logger      *zap.SugaredLogger
}

func NewShopHandler(shopService ShopService, logger *zap.SugaredLogger) (*ShopHandler, error) {
	return &ShopHandler{
		shopService: shopService,
		logger:      logger,
	}, nil
}

type TasksResponse struct {
	Tasks []domain.TaskData `json:"tasks"`
}

func (h *ShopHandler) GetTasks(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	vkIdStr, err := req.Cookie("vkid")
	if err != nil {
		vkIdStr = &http.Cookie{}
	}
	vkId, err := strconv.Atoi(vkIdStr.Value)
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

	tasks, err := h.shopService.GetTasks(ctx, vkId)
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

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data: TasksResponse{
			Tasks: tasks,
		},
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

func (h *ShopHandler) GetCurrentGiftaway(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	giftaway, err := h.shopService.GetCurrentGiftaway(ctx)
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

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   giftaway,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}
