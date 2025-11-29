package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) AddItem(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
	cartID, err := r.getOrCreateCartID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create cart: %w", err)
	}

	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity, size, color)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cart_id, product_id, size, color) 
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = NOW()
		RETURNING id, cart_id, product_id, quantity, size, created_at, updated_at
	`

	var item models.CartItem
	var returnedCartID int
	err = r.db.QueryRow(ctx, query, cartID, req.ProductID, req.Quantity, req.Size, nil).Scan(
		&item.ID,
		&returnedCartID,
		&item.ProductID,
		&item.Quantity,
		&item.Size,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to add item to cart")
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}

	item.UserID = userID
	return &item, nil
}

func (r *CartRepository) getOrCreateCartID(ctx context.Context, userID int) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `SELECT id FROM carts WHERE user_id = $1 LIMIT 1`, userID).Scan(&id)
	if err == nil {
		return id, nil
	}

	err = r.db.QueryRow(ctx, `INSERT INTO carts (user_id, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id`, userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *CartRepository) GetUserCart(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
	query := `
		SELECT 
			ci.id, c.user_id, ci.product_id, ci.quantity, ci.size, ci.created_at, ci.updated_at,
			p.title as product_title,
			p.price as product_price,
			p.image_url as product_image
		FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		JOIN products p ON ci.product_id = p.id
		WHERE c.user_id = $1
		ORDER BY ci.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get user cart")
		return nil, fmt.Errorf("failed to get user cart: %w", err)
	}
	defer rows.Close()

	var items []*models.CartItemWithDetails
	for rows.Next() {
		var item models.CartItemWithDetails
		var productImage sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProductID,
			&item.Quantity,
			&item.Size,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ProductTitle,
			&item.ProductPrice,
			&productImage,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		if productImage.Valid {
			item.ProductImage = productImage.String
		}
		items = append(items, &item)
	}

	return items, nil
}

func (r *CartRepository) UpdateItem(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	query := `
		UPDATE cart_items
		SET quantity = $1, size = COALESCE(NULLIF($2, ''), size), updated_at = NOW()
		WHERE id = $3 AND cart_id = (SELECT id FROM carts WHERE user_id = $4)
		RETURNING id, cart_id, product_id, quantity, size, created_at, updated_at
	`

	var item models.CartItem
	var returnedCartID int
	err := r.db.QueryRow(ctx, query, req.Quantity, req.Size, itemID, userID).Scan(
		&item.ID,
		&returnedCartID,
		&item.ProductID,
		&item.Quantity,
		&item.Size,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to update cart item")
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}

	item.UserID = userID
	return &item, nil
}

func (r *CartRepository) DeleteItem(ctx context.Context, itemID, userID int) error {
	query := `DELETE FROM cart_items WHERE id = $1 AND cart_id = (SELECT id FROM carts WHERE user_id = $2)`

	result, err := r.db.Exec(ctx, query, itemID, userID)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to delete cart item")
		return fmt.Errorf("failed to delete cart item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cart item not found")
	}

	return nil
}

func (r *CartRepository) ClearCart(ctx context.Context, userID int) error {
	query := `DELETE FROM cart_items WHERE cart_id = (SELECT id FROM carts WHERE user_id = $1)`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to clear cart")
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}
