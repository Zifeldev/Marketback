package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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

	query, args, err := psql.Insert("cart_items").
		Columns("cart_id", "product_id", "quantity", "size", "color").
		Values(cartID, req.ProductID, req.Quantity, req.Size, nil).
		Suffix("ON CONFLICT (cart_id, product_id, size, color) DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = NOW()").
		Suffix("RETURNING id, cart_id, product_id, quantity, COALESCE(size, '') as size, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build add item query: %w", err)
	}

	var item models.CartItem
	var returnedCartID int
	err = r.db.QueryRow(ctx, query, args...).Scan(
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
	selectQuery, selectArgs, err := psql.Select("id").
		From("carts").
		Where(sq.Eq{"user_id": userID}).
		Limit(1).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build select cart query: %w", err)
	}

	var id int
	err = r.db.QueryRow(ctx, selectQuery, selectArgs...).Scan(&id)
	if err == nil {
		return id, nil
	}

	insertQuery, insertArgs, err := psql.Insert("carts").
		Columns("user_id", "created_at", "updated_at").
		Values(userID, sq.Expr("NOW()"), sq.Expr("NOW()")).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build insert cart query: %w", err)
	}

	err = r.db.QueryRow(ctx, insertQuery, insertArgs...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *CartRepository) GetUserCart(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
	query, args, err := psql.Select(
		"ci.id", "c.user_id", "ci.product_id", "ci.quantity", "COALESCE(ci.size, '') as size", "ci.created_at", "ci.updated_at",
		"p.title as product_title",
		"p.price::float8 as product_price",
		"COALESCE(p.image_url, '') as product_image",
	).From("cart_items ci").
		Join("carts c ON ci.cart_id = c.id").
		Join("products p ON ci.product_id = p.id").
		Where(sq.Eq{"c.user_id": userID}).
		OrderBy("ci.created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build get cart query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get user cart")
		return nil, fmt.Errorf("failed to get user cart: %w", err)
	}
	defer rows.Close()

	var items []*models.CartItemWithDetails
	for rows.Next() {
		var item models.CartItemWithDetails
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
			&item.ProductImage,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}

func (r *CartRepository) UpdateItem(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	updateBuilder := psql.Update("cart_items").
		Set("quantity", req.Quantity).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.And{
			sq.Eq{"id": itemID},
			sq.Expr("cart_id = (SELECT id FROM carts WHERE user_id = ?)", userID),
		}).
		Suffix("RETURNING id, cart_id, product_id, quantity, COALESCE(size, '') as size, created_at, updated_at")

	if req.Size != "" {
		updateBuilder = updateBuilder.Set("size", req.Size)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build update cart item query: %w", err)
	}

	var item models.CartItem
	var returnedCartID int
	err = r.db.QueryRow(ctx, query, args...).Scan(
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
	query, args, err := psql.Delete("cart_items").
		Where(sq.And{
			sq.Eq{"id": itemID},
			sq.Expr("cart_id = (SELECT id FROM carts WHERE user_id = ?)", userID),
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete cart item query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
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
	query, args, err := psql.Delete("cart_items").
		Where(sq.Expr("cart_id = (SELECT id FROM carts WHERE user_id = ?)", userID)).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build clear cart query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to clear cart")
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}
