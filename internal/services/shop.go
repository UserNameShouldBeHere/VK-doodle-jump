package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
)

type ShopStorage interface {
	GetTasks(ctx context.Context, vkid int) ([]domain.TaskData, error)
	GetCurrentGiftaway(ctx context.Context) (domain.Giftaway, error)
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

func (s *ShopService) GetCurrentGiftaway(ctx context.Context) (domain.Giftaway, error) {
	giftaway, err := s.shopStorage.GetCurrentGiftaway(ctx)
	if err != nil {
		s.logger.Errorf("(shopService.GetCurrentGiftaway): %w", err)
		return domain.Giftaway{}, fmt.Errorf("(shopService.GetCurrentGiftaway): %w", err)
	}

	return giftaway, nil
}
