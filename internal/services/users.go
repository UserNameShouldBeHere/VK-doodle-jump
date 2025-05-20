package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
)

type UsersStorage interface {
	UpdateUserRating(ctx context.Context, vkid int, newScore int) error
	GetTopUsers(ctx context.Context, count int) ([]domain.UserRating, error)
	GetNearbyUsers(ctx context.Context, vkid, count int) ([]domain.UserRating, error)
	UserScore(ctx context.Context, vkid int) (int, error)
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

func (s *UsersService) UpdateUserRating(ctx context.Context, vkid int, newScore int) error {
	err := s.storage.UpdateUserRating(ctx, vkid, newScore)
	if err != nil {
		s.logger.Errorf("(usersService.UpdateUserRating) %w", err)
		return fmt.Errorf("(usersService.UpdateUserRating) %w", err)
	}

	return nil
}

func (s *UsersService) GetTopUsers(ctx context.Context, count int) ([]domain.UserRating, error) {
	usersTop, err := s.storage.GetTopUsers(ctx, count)
	if err != nil {
		s.logger.Errorf("(usersService.GetTopUsers): %w", err)
		return nil, fmt.Errorf("(usersService.GetTopUsers): %w", err)
	}

	return usersTop, nil
}

func (s *UsersService) GetNearbyUsers(ctx context.Context, vkid, count int) ([]domain.UserRating, error) {
	usersTop, err := s.storage.GetNearbyUsers(ctx, vkid, count)
	if err != nil {
		s.logger.Errorf("(usersService.GetNearbyUsers): %w", err)
		return nil, fmt.Errorf("(usersService.GetNearbyUsers): %w", err)
	}

	return usersTop, nil
}

func (s *UsersService) UserScore(ctx context.Context, vkid int) (int, error) {
	score, err := s.storage.UserScore(ctx, vkid)
	if err != nil {
		s.logger.Errorf("(usersService.UserScore): %w", err)
		return 0, fmt.Errorf("(usersService.UserScore): %w", err)
	}

	return score, nil
}
