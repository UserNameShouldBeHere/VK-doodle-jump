package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
)

type UsersStorage interface {
	UpdateUserRating(ctx context.Context, uuid string, newScore int) error
	GetTopUsers(ctx context.Context, count int) ([]domain.UserRating, error)
}

type UsersService struct {
	storage UsersStorage
	logger  *zap.SugaredLogger
}

func NewUsersService(storage UsersStorage, logger *zap.SugaredLogger) (*UsersService, error) {
	return &UsersService{
		storage: storage,
		logger:  logger,
	}, nil
}

func (s *UsersService) UpdateUserRating(ctx context.Context, uuid string, newScore int) error {
	err := s.storage.UpdateUserRating(ctx, uuid, newScore)
	if err != nil {
		s.logger.Errorf("failed to update user's rating: %v", err)
		return fmt.Errorf("(services.UpdateUserRating): %w", err)
	}

	return nil
}

func (s *UsersService) GetTopUsers(ctx context.Context, count int) ([]domain.UserRating, error) {
	usersTop, err := s.storage.GetTopUsers(ctx, count)
	if err != nil {
		s.logger.Errorf("failed to get top users: %v", err)
		return nil, fmt.Errorf("(services.GetTopUsers): %w", err)
	}

	return usersTop, nil
}
