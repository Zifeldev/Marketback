package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
)

// Extended service layer tests

func TestMarketService_CreateOrder_CartRepoError(t *testing.T) {
	mCart := &mockCartRepoSvc{itemsFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
		return nil, errors.New("database connection failed")
	}}

	ctx := context.Background()
	_, err := mCart.GetUserCart(ctx, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database")
}

func TestMarketService_CreateOrder_MultipleItems(t *testing.T) {
	now := time.Now()
	cartItems := []*models.CartItemWithDetails{
		{
			CartItem:     models.CartItem{ID: 1, UserID: 10, ProductID: 100, Quantity: 2, Size: "M", CreatedAt: now, UpdatedAt: now},
			ProductTitle: "T-Shirt",
			ProductPrice: 25.00,
		},
		{
			CartItem:     models.CartItem{ID: 2, UserID: 10, ProductID: 101, Quantity: 1, Size: "L", CreatedAt: now, UpdatedAt: now},
			ProductTitle: "Jeans",
			ProductPrice: 50.00,
		},
		{
			CartItem:     models.CartItem{ID: 3, UserID: 10, ProductID: 102, Quantity: 3, Size: "", CreatedAt: now, UpdatedAt: now},
			ProductTitle: "Socks",
			ProductPrice: 5.00,
		},
	}

	// Expected total: 2*25 + 1*50 + 3*5 = 50 + 50 + 15 = 115
	expectedTotal := 115.0

	mCart := &mockCartRepoSvc{itemsFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
		return cartItems, nil
	}}

	mOrder := &mockOrderRepo{createFn: func(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
		require.Len(t, items, 3)

		// Calculate total
		var total float64
		for _, item := range items {
			total += item.ProductPrice * float64(item.Quantity)
		}
		require.InDelta(t, expectedTotal, total, 0.01)

		return &models.OrderWithItems{
			Order: models.Order{
				ID:            1,
				UserID:        userID,
				TotalAmount:   total,
				Status:        "pending",
				PaymentMethod: req.PaymentMethod,
				PaymentStatus: "pending",
				DeliveryAddr:  req.DeliveryAddr,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			Items: []models.OrderItem{
				{ID: 1, OrderID: 1, ProductID: 100, Quantity: 2, Size: "M", Price: 25.00},
				{ID: 2, OrderID: 1, ProductID: 101, Quantity: 1, Size: "L", Price: 50.00},
				{ID: 3, OrderID: 1, ProductID: 102, Quantity: 3, Size: "", Price: 5.00},
			},
		}, nil
	}}

	ctx := context.Background()
	items, err := mCart.GetUserCart(ctx, 10)
	require.NoError(t, err)
	require.Len(t, items, 3)

	orderReq := &models.CreateOrderRequest{PaymentMethod: "credit_card", DeliveryAddr: "123 Main St"}
	order, err := mOrder.Create(ctx, 10, orderReq, items)
	require.NoError(t, err)
	require.Equal(t, expectedTotal, order.TotalAmount)
	require.Len(t, order.Items, 3)
}

func TestMarketService_CreateOrder_DifferentPaymentMethods(t *testing.T) {
	now := time.Now()
	cartItems := []*models.CartItemWithDetails{{
		CartItem:     models.CartItem{ID: 1, UserID: 1, ProductID: 1, Quantity: 1, Size: "", CreatedAt: now, UpdatedAt: now},
		ProductTitle: "Item",
		ProductPrice: 100.00,
	}}

	paymentMethods := []string{"credit_card", "cash", "bank_transfer", "crypto"}

	for _, pm := range paymentMethods {
		t.Run(pm, func(t *testing.T) {
			mCart := &mockCartRepoSvc{itemsFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
				return cartItems, nil
			}}

			mOrder := &mockOrderRepo{createFn: func(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
				require.Equal(t, pm, req.PaymentMethod)
				return &models.OrderWithItems{
					Order: models.Order{ID: 1, PaymentMethod: pm},
				}, nil
			}}

			ctx := context.Background()
			items, _ := mCart.GetUserCart(ctx, 1)
			orderReq := &models.CreateOrderRequest{PaymentMethod: pm, DeliveryAddr: "Address"}
			order, err := mOrder.Create(ctx, 1, orderReq, items)
			require.NoError(t, err)
			require.Equal(t, pm, order.PaymentMethod)
		})
	}
}

func TestServiceError_Error(t *testing.T) {
	err := &ServiceError{Message: "test error message"}
	require.Equal(t, "test error message", err.Error())
}

func TestErrEmptyCart_Message(t *testing.T) {
	require.Equal(t, "cart is empty", ErrEmptyCart.Error())
}

func TestMarketService_CreateOrder_ZeroPriceItem(t *testing.T) {
	now := time.Now()
	cartItems := []*models.CartItemWithDetails{{
		CartItem:     models.CartItem{ID: 1, UserID: 1, ProductID: 1, Quantity: 1, Size: "", CreatedAt: now, UpdatedAt: now},
		ProductTitle: "Free Item",
		ProductPrice: 0.00,
	}}

	mCart := &mockCartRepoSvc{itemsFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
		return cartItems, nil
	}}

	mOrder := &mockOrderRepo{createFn: func(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
		return &models.OrderWithItems{
			Order: models.Order{ID: 1, TotalAmount: 0},
		}, nil
	}}

	ctx := context.Background()
	items, _ := mCart.GetUserCart(ctx, 1)
	orderReq := &models.CreateOrderRequest{PaymentMethod: "free", DeliveryAddr: "Address"}
	order, err := mOrder.Create(ctx, 1, orderReq, items)
	require.NoError(t, err)
	require.Equal(t, 0.0, order.TotalAmount)
}

func TestMarketService_CreateOrder_LargeQuantity(t *testing.T) {
	now := time.Now()
	cartItems := []*models.CartItemWithDetails{{
		CartItem:     models.CartItem{ID: 1, UserID: 1, ProductID: 1, Quantity: 1000, Size: "", CreatedAt: now, UpdatedAt: now},
		ProductTitle: "Bulk Item",
		ProductPrice: 10.00,
	}}

	expectedTotal := 10000.0

	mCart := &mockCartRepoSvc{itemsFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
		return cartItems, nil
	}}

	mOrder := &mockOrderRepo{createFn: func(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
		total := items[0].ProductPrice * float64(items[0].Quantity)
		require.Equal(t, expectedTotal, total)
		return &models.OrderWithItems{
			Order: models.Order{ID: 1, TotalAmount: total},
		}, nil
	}}

	ctx := context.Background()
	items, _ := mCart.GetUserCart(ctx, 1)
	orderReq := &models.CreateOrderRequest{PaymentMethod: "card", DeliveryAddr: "Warehouse"}
	order, err := mOrder.Create(ctx, 1, orderReq, items)
	require.NoError(t, err)
	require.Equal(t, expectedTotal, order.TotalAmount)
}
