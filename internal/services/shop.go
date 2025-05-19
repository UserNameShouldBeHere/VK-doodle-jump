package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
)

type ShopStorage interface {
	GetTasks(ctx context.Context, vkid int) ([]domain.TaskData, error)
	PassTask(ctx context.Context, vkid int, taskId int) error
}

type ShopService struct {
	shopStorage ShopStorage
	logger      *zap.SugaredLogger
}

func NewShopService(shopStorage ShopStorage, logger *zap.SugaredLogger) (*ShopService, error) {
	return &ShopService{
		shopStorage: shopStorage,
		logger:      logger,
	}, nil
}

func (s *ShopService) GetTasks(ctx context.Context, vkid int) ([]domain.TaskData, error) {
	tasks, err := s.shopStorage.GetTasks(ctx, vkid)
	if err != nil {
		s.logger.Errorf("(shopService.GetTasks): %w", err)
		return nil, fmt.Errorf("(shopService.GetTasks): %w", err)
	}

	return tasks, nil
}

func (s *ShopService) PassTask(ctx context.Context, vkid int, taskId int) error {
	err := s.shopStorage.PassTask(ctx, vkid, taskId)
	if err != nil {
		s.logger.Errorf("(shopService.PassTask): %w", err)
		return fmt.Errorf("(shopService.PassTask): %w", err)
	}

	return nil
}
