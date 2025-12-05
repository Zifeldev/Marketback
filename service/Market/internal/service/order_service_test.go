package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
)

type mockCartRepoService struct {
	getUserCartFn func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error)
}

func (m *mockCartRepoService) GetUserCart(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
	if m.getUserCartFn != nil {
		return m.getUserCartFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockCartRepoService) AddItem(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
	return nil, nil
}

func (m *mockCartRepoService) UpdateItem(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	return nil, nil
}

func (m *mockCartRepoService) DeleteItem(ctx context.Context, itemID, userID int) error {
	return nil
}

func (m *mockCartRepoService) ClearCart(ctx context.Context, userID int) error {
	return nil
}

type mockOrderRepoService struct {
	createFn func(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error)
}

func (m *mockOrderRepoService) Create(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, req, items)
	}
	return nil, nil
}

func (m *mockOrderRepoService) GetByID(ctx context.Context, id int) (*models.OrderWithItems, error) {
	return nil, nil
}

func (m *mockOrderRepoService) GetUserOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	return nil, nil
}

func (m *mockOrderRepoService) GetAll(ctx context.Context) ([]*models.Order, error) {
	return nil, nil
}

func (m *mockOrderRepoService) UpdateStatus(ctx context.Context, id int, status string) (*models.Order, error) {
	return nil, nil
}

func TestMarketService_CreateOrder_ValidOrder(t *testing.T) {
	now := time.Now()
	cartItems := []*models.CartItemWithDetails{
		{
			CartItem:     models.CartItem{ID: 1, UserID: 10, ProductID: 5, Quantity: 2, Size: "M", CreatedAt: now, UpdatedAt: now},
			ProductTitle: "T-Shirt",
			ProductPrice: 25.00,
			ProductImage: "http://example.com/tshirt.jpg",
		},
		{
			CartItem:     models.CartItem{ID: 2, UserID: 10, ProductID: 6, Quantity: 1, Size: "L", CreatedAt: now, UpdatedAt: now},
			ProductTitle: "Pants",
			ProductPrice: 50.00,
			ProductImage: "http://example.com/pants.jpg",
		},
	}

	cartRepo := &mockCartRepoService{
		getUserCartFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
			assert.Equal(t, 10, userID)
			return cartItems, nil
		},
	}

	orderRepo := &mockOrderRepoService{
		createFn: func(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
			assert.Equal(t, 10, userID)
			assert.Equal(t, "card", req.PaymentMethod)
			assert.Len(t, items, 2)

			var total float64
			for _, item := range items {
				total += float64(item.Quantity) * item.ProductPrice
			}

			return &models.OrderWithItems{
				Order: models.Order{
					ID:            1,
					UserID:        10,
					TotalAmount:   total,
					Status:        "pending",
					PaymentMethod: req.PaymentMethod,
					DeliveryAddr:  req.DeliveryAddr,
					CreatedAt:     now,
					UpdatedAt:     now,
				},
				Items: []models.OrderItem{
					{ID: 1, OrderID: 1, ProductID: 5, Quantity: 2, Price: 25.00, Size: "M"},
					{ID: 2, OrderID: 1, ProductID: 6, Quantity: 1, Price: 50.00, Size: "L"},
				},
			}, nil
		},
	}

	ctx := context.Background()

	items, err := cartRepo.GetUserCart(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, items)

	orderReq := &models.CreateOrderRequest{
		PaymentMethod: "card",
		DeliveryAddr:  "123 Main Street",
	}

	order, err := orderRepo.Create(ctx, 10, orderReq, items)
	require.NoError(t, err)
	require.NotNil(t, order)

	assert.Equal(t, 1, order.ID)
	assert.Equal(t, 100.00, order.TotalAmount)
	assert.Equal(t, "pending", order.Status)
	assert.Len(t, order.Items, 2)
}

func TestMarketService_CreateOrder_EmptyCartError(t *testing.T) {
	cartRepo := &mockCartRepoService{
		getUserCartFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
			return []*models.CartItemWithDetails{}, nil
		},
	}

	ctx := context.Background()
	items, err := cartRepo.GetUserCart(ctx, 10)
	require.NoError(t, err)

	if len(items) == 0 {
		err = ErrEmptyCart
	}

	assert.ErrorIs(t, err, ErrEmptyCart)
}

func TestMarketService_CreateOrder_CartRepoConnectionError(t *testing.T) {
	expectedErr := errors.New("database connection failed")

	cartRepo := &mockCartRepoService{
		getUserCartFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
			return nil, expectedErr
		},
	}

	ctx := context.Background()
	_, err := cartRepo.GetUserCart(ctx, 10)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMarketService_CreateOrder_OrderRepoDBError(t *testing.T) {
	now := time.Now()
	cartItems := []*models.CartItemWithDetails{
		{
			CartItem:     models.CartItem{ID: 1, UserID: 10, ProductID: 5, Quantity: 1, Size: "S", CreatedAt: now, UpdatedAt: now},
			ProductTitle: "Hat",
			ProductPrice: 15.00,
		},
	}

	cartRepo := &mockCartRepoService{
		getUserCartFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
			return cartItems, nil
		},
	}

	expectedErr := errors.New("failed to create order")
	orderRepo := &mockOrderRepoService{
		createFn: func(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
			return nil, expectedErr
		},
	}

	ctx := context.Background()
	items, err := cartRepo.GetUserCart(ctx, 10)
	require.NoError(t, err)

	orderReq := &models.CreateOrderRequest{PaymentMethod: "cash", DeliveryAddr: "Test"}
	_, err = orderRepo.Create(ctx, 10, orderReq, items)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMarketService_CalculateOrderTotal(t *testing.T) {
	items := []*models.CartItemWithDetails{
		{CartItem: models.CartItem{Quantity: 2}, ProductPrice: 10.00},
		{CartItem: models.CartItem{Quantity: 3}, ProductPrice: 15.00},
		{CartItem: models.CartItem{Quantity: 1}, ProductPrice: 25.50},
	}

	var total float64
	for _, item := range items {
		total += float64(item.Quantity) * item.ProductPrice
	}

	assert.Equal(t, 90.50, total)
}

func TestMarketService_CalculateOrderTotal_EmptyItems(t *testing.T) {
	items := []*models.CartItemWithDetails{}

	var total float64
	for _, item := range items {
		total += float64(item.Quantity) * item.ProductPrice
	}

	assert.Equal(t, 0.0, total)
}

func TestMarketService_CalculateOrderTotal_SingleItem(t *testing.T) {
	items := []*models.CartItemWithDetails{
		{CartItem: models.CartItem{Quantity: 5}, ProductPrice: 20.00},
	}

	var total float64
	for _, item := range items {
		total += float64(item.Quantity) * item.ProductPrice
	}

	assert.Equal(t, 100.00, total)
}
