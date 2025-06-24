package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"github.com/golang-jwt/jwt"
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
	AddSuperpower(ctx context.Context, vkid int, task string) error
	GetCurrentGiftaway(ctx context.Context) (domain.Giftaway, error)
	AddGift(ctx context.Context, newGift domain.Gift) error
	UpdateGift(ctx context.Context, newGift domain.Gift) error
	DeleteGift(ctx context.Context, id int) error
}

type AdminShopService struct {
	shopStorage AdminShopStorage
	logger      *zap.SugaredLogger
	jwtKey      []byte
}

func NewAdminShopService(shopStorage AdminShopStorage, logger *zap.SugaredLogger) (*AdminShopService, error) {
	jwtKey := make([]byte, 16)
	_, err := rand.Read(jwtKey)
	if err != nil {
		logger.Errorf("(adminShopService.NewAdminShopService): %w", err)
		return nil, fmt.Errorf("(adminShopService.NewAdminShopService): %w", err)
	}

	return &AdminShopService{
		shopStorage: shopStorage,
		logger:      logger,
		jwtKey:      jwtKey,
	}, nil
}

func (s *AdminShopService) GetPromocodes(ctx context.Context) ([]domain.PromocodeAdminData, error) {
	promocodes, err := s.shopStorage.GetPromocodes(ctx)
	if err != nil {
		s.logger.Errorf("(adminShopService.GetPromocodes): %w", err)
		return nil, fmt.Errorf("(adminShopService.GetPromocodes): %w", err)
	}

	return promocodes, nil
}

func (s *AdminShopService) AddPromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error {
	err := s.shopStorage.AddPromocode(ctx, newPromocode)
	if err != nil {
		s.logger.Errorf("(adminShopService.AddPromocode): %w", err)
		return fmt.Errorf("(adminShopService.AddPromocode): %w", err)
	}

	return nil
}

func (s *AdminShopService) UpdatePromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error {
	err := s.shopStorage.UpdatePromocode(ctx, newPromocode)
	if err != nil {
		s.logger.Errorf("(adminShopService.UpdatePromocode): %w", err)
		return fmt.Errorf("(adminShopService.UpdatePromocode): %w", err)
	}

	return nil
}

func (s *AdminShopService) DeletePromocode(ctx context.Context, id int) error {
	err := s.shopStorage.DeletePromocode(ctx, id)
	if err != nil {
		s.logger.Errorf("(adminShopService.DeletePromocode): %w", err)
		return fmt.Errorf("(adminShopService.DeletePromocode): %w", err)
	}

	return nil
}

func (s *AdminShopService) GetProducts(ctx context.Context) ([]domain.ProductAdminData, error) {
	products, err := s.shopStorage.GetProducts(ctx)
	if err != nil {
		s.logger.Errorf("(adminShopService.GetProducts): %w", err)
		return nil, fmt.Errorf("(adminShopService.GetProducts): %w", err)
	}

	return products, nil
}

func (s *AdminShopService) AddProduct(ctx context.Context, newProduct domain.ProductAdminData) error {
	err := s.shopStorage.AddProduct(ctx, newProduct)
	if err != nil {
		s.logger.Errorf("(adminShopService.AddProduct): %w", err)
		return fmt.Errorf("(adminShopService.AddProduct): %w", err)
	}

	return nil
}

func (s *AdminShopService) UpdateProduct(ctx context.Context, newProduct domain.ProductAdminData) error {
	err := s.shopStorage.UpdateProduct(ctx, newProduct)
	if err != nil {
		s.logger.Errorf("(adminShopService.UpdateProduct): %w", err)
		return fmt.Errorf("(adminShopService.UpdateProduct): %w", err)
	}

	return nil
}

func (s *AdminShopService) DeleteProduct(ctx context.Context, id int) error {
	err := s.shopStorage.DeleteProduct(ctx, id)
	if err != nil {
		s.logger.Errorf("(adminShopService.DeleteProduct): %w", err)
		return fmt.Errorf("(adminShopService.DeleteProduct): %w", err)
	}

	return nil
}

func (s *AdminShopService) GetTasks(ctx context.Context) ([]domain.TaskAdminData, error) {
	tasks, err := s.shopStorage.GetTasks(ctx)
	if err != nil {
		s.logger.Errorf("(adminShopService.GetTasks): %w", err)
		return nil, fmt.Errorf("(adminShopService.GetTasks): %w", err)
	}

	return tasks, nil
}

func (s *AdminShopService) AddTask(ctx context.Context, newTask domain.TaskAdminData) error {
	token, err := s.createToken()
	if err != nil {
		s.logger.Errorf("(adminShopService.AddTask): %w", err)
		return fmt.Errorf("(adminShopService.AddTask): %w", err)
	}

	newTask.Token = token

	err = s.shopStorage.AddTask(ctx, newTask)
	if err != nil {
		s.logger.Errorf("(adminShopService.AddTask): %w", err)
		return fmt.Errorf("(adminShopService.AddTask): %w", err)
	}

	return nil
}

func (s *AdminShopService) UpdateTask(ctx context.Context, newTask domain.TaskAdminData) error {
	err := s.shopStorage.UpdateTask(ctx, newTask)
	if err != nil {
		s.logger.Errorf("(adminShopService.UpdateTask): %w", err)
		return fmt.Errorf("(adminShopService.UpdateTask): %w", err)
	}

	return nil
}

func (s *AdminShopService) DeleteTask(ctx context.Context, id int) error {
	err := s.shopStorage.DeleteTask(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to delete task: %w", err)
		return fmt.Errorf("(services.DeleteTask): %w", err)
	}

	return nil
}

func (s *AdminShopService) AddSuperpower(ctx context.Context, vkid int, task string) error {
	err := s.validateToken(task)
	if err != nil {
		s.logger.Errorf("(adminShopService.AddSuperpower): %w", err)
		return fmt.Errorf("(adminShopService.AddSuperpower): %w", err)
	}

	err = s.shopStorage.AddSuperpower(ctx, vkid, task)
	if err != nil {
		s.logger.Errorf("(adminShopService.AddSuperpower): %w", err)
		return fmt.Errorf("(adminShopService.AddSuperpower): %w", err)
	}

	return nil
}

func (s *AdminShopService) GetCurrentGiftaway(ctx context.Context) (domain.Giftaway, error) {
	giftaway, err := s.shopStorage.GetCurrentGiftaway(ctx)
	if err != nil {
		s.logger.Errorf("failed to get giftaway: %w", err)
		return domain.Giftaway{}, fmt.Errorf("(services.GetCurrentGiftaway): %w", err)
	}

	return giftaway, nil
}

func (s *AdminShopService) AddGift(ctx context.Context, newGift domain.Gift) error {
	err := s.shopStorage.AddGift(ctx, newGift)
	if err != nil {
		s.logger.Errorf("failed to add gift: %w", err)
		return fmt.Errorf("(services.AddGift): %w", err)
	}

	return nil
}

func (s *AdminShopService) UpdateGift(ctx context.Context, newGift domain.Gift) error {
	err := s.shopStorage.UpdateGift(ctx, newGift)
	if err != nil {
		s.logger.Errorf("failed to update gift: %w", err)
		return fmt.Errorf("(services.UpdateGift): %w", err)
	}

	return nil
}

func (s *AdminShopService) DeleteGift(ctx context.Context, id int) error {
	err := s.shopStorage.DeleteGift(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to delete gift: %w", err)
		return fmt.Errorf("(services.DeleteGift): %w", err)
	}

	return nil
}

func (s *AdminShopService) createToken() (string, error) {
	claims := jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		Issuer:   "mail-jumper",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", fmt.Errorf("(adminShopService.createToken): %w", err)
	}

	return signedToken, nil
}

func (s *AdminShopService) validateToken(token string) error {
	_, err := jwt.Parse(token,
		func(token *jwt.Token) (interface{}, error) {
			return s.jwtKey, nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}
