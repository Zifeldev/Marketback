package service

import (
	"context"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
)

type MarketService struct {
	orderRepo *repository.OrderRepository
	cartRepo  *repository.CartRepository
}

func NewMarketService(orderRepo *repository.OrderRepository, cartRepo *repository.CartRepository) *MarketService {
	return &MarketService{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
	}
}

func (s *MarketService) CreateOrder(ctx context.Context, userID int, req *models.CreateOrderRequest) (*models.OrderWithItems, error) {
	cartItems, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, ErrEmptyCart
	}

	return s.orderRepo.Create(ctx, userID, req, cartItems)
}

var ErrEmptyCart = &ServiceError{Message: "cart is empty"}

type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}
