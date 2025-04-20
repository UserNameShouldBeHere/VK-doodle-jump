package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
)

type AdminShopStorage interface {
	GetPromocodes(ctx context.Context) ([]domain.PromocodeAdminData, error)
	AddPromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error
	UpdatePromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error
	DeletePromocode(ctx context.Context, id int) error
	GetProducts(ctx context.Context) ([]domain.ProductAdminData, error)
	AddProduct(ctx context.Context, newProduct domain.ProductAdminData) error
	UpdateProduct(ctx context.Context, newProduct domain.ProductAdminData) error
	DeleteProduct(ctx context.Context, id int) error
	GetTasks(ctx context.Context) ([]domain.TaskAdminData, error)
	AddTask(ctx context.Context, newTask domain.TaskAdminData) error
	UpdateTask(ctx context.Context, newTask domain.TaskAdminData) error
	DeleteTask(ctx context.Context, id int) error
}

type AdminShopService struct {
	shopStorage AdminShopStorage
	logger      *zap.SugaredLogger
}

func NewAdminShopService(shopStorage AdminShopStorage, logger *zap.SugaredLogger) (*AdminShopService, error) {
	return &AdminShopService{
		shopStorage: shopStorage,
		logger:      logger,
	}, nil
}

func (s *AdminShopService) GetPromocodes(ctx context.Context) ([]domain.PromocodeAdminData, error) {
	promocodes, err := s.shopStorage.GetPromocodes(ctx)
	if err != nil {
		s.logger.Errorf("failed to get promocodes: %v", err)
		return nil, fmt.Errorf("(services.GetPromocodes): %w", err)
	}

	return promocodes, nil
}

func (s *AdminShopService) AddPromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error {
	err := s.shopStorage.AddPromocode(ctx, newPromocode)
	if err != nil {
		s.logger.Errorf("failed to add promocode: %v", err)
		return fmt.Errorf("(services.AddPromocode): %w", err)
	}

	return nil
}

func (s *AdminShopService) UpdatePromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error {
	err := s.shopStorage.UpdatePromocode(ctx, newPromocode)
	if err != nil {
		s.logger.Errorf("failed to update promocode: %v", err)
		return fmt.Errorf("(services.UpdatePromocode): %w", err)
	}

	return nil
}

func (s *AdminShopService) DeletePromocode(ctx context.Context, id int) error {
	err := s.shopStorage.DeletePromocode(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to delete promocode: %v", err)
		return fmt.Errorf("(services.DeletePromocode): %w", err)
	}

	return nil
}

func (s *AdminShopService) GetProducts(ctx context.Context) ([]domain.ProductAdminData, error) {
	products, err := s.shopStorage.GetProducts(ctx)
	if err != nil {
		s.logger.Errorf("failed to get products: %v", err)
		return nil, fmt.Errorf("(services.GetProducts): %w", err)
	}

	return products, nil
}

func (s *AdminShopService) AddProduct(ctx context.Context, newProduct domain.ProductAdminData) error {
	err := s.shopStorage.AddProduct(ctx, newProduct)
	if err != nil {
		s.logger.Errorf("failed to add product: %v", err)
		return fmt.Errorf("(services.AddProduct): %w", err)
	}

	return nil
}

func (s *AdminShopService) UpdateProduct(ctx context.Context, newProduct domain.ProductAdminData) error {
	err := s.shopStorage.UpdateProduct(ctx, newProduct)
	if err != nil {
		s.logger.Errorf("failed to update product: %v", err)
		return fmt.Errorf("(services.UpdateProduct): %w", err)
	}

	return nil
}

func (s *AdminShopService) DeleteProduct(ctx context.Context, id int) error {
	err := s.shopStorage.DeleteProduct(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to delete product: %v", err)
		return fmt.Errorf("(services.DeleteProduct): %w", err)
	}

	return nil
}

func (s *AdminShopService) GetTasks(ctx context.Context) ([]domain.TaskAdminData, error) {
	tasks, err := s.shopStorage.GetTasks(ctx)
	if err != nil {
		s.logger.Errorf("failed to get tasks: %v", err)
		return nil, fmt.Errorf("(services.GetTasks): %w", err)
	}

	return tasks, nil
}

func (s *AdminShopService) AddTask(ctx context.Context, newTask domain.TaskAdminData) error {
	err := s.shopStorage.AddTask(ctx, newTask)
	if err != nil {
		s.logger.Errorf("failed to add task: %v", err)
		return fmt.Errorf("(services.AddTask): %w", err)
	}

	return nil
}

func (s *AdminShopService) UpdateTask(ctx context.Context, newTask domain.TaskAdminData) error {
	err := s.shopStorage.UpdateTask(ctx, newTask)
	if err != nil {
		s.logger.Errorf("failed to update task: %v", err)
		return fmt.Errorf("(services.UpdateTask): %w", err)
	}

	return nil
}

func (s *AdminShopService) DeleteTask(ctx context.Context, id int) error {
	err := s.shopStorage.DeleteTask(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to delete task: %v", err)
		return fmt.Errorf("(services.DeleteTask): %w", err)
	}

	return nil
}
