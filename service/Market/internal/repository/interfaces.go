package repository

import (
	"context"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
)

type CartRepo interface {
	AddItem(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error)
	GetUserCart(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error)
	UpdateItem(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error)
	DeleteItem(ctx context.Context, itemID, userID int) error
	ClearCart(ctx context.Context, userID int) error
}

type ProductRepo interface {
	GetAll(ctx context.Context, categoryID, sellerID *int, status string, pagination *models.PaginationParams) ([]*models.ProductWithDetails, int64, error)
	GetByID(ctx context.Context, id int) (*models.ProductWithDetails, error)
}

type CategoryRepo interface {
	GetAll(ctx context.Context) ([]*models.Category, error)
	GetByID(ctx context.Context, id int) (*models.Category, error)
}

type OrderRepo interface {
	GetUserOrders(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error)
	GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error)
}
