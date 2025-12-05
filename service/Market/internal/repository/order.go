package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, userID int, req *models.CreateOrderRequest, items []*models.CartItemWithDetails) (*models.OrderWithItems, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var totalAmount float64
	for _, item := range items {
		totalAmount += item.ProductPrice * float64(item.Quantity)
	}

	orderQuery, orderArgs, err := psql.Insert("orders").
		Columns("user_id", "total_amount", "payment_method", "delivery_address").
		Values(userID, totalAmount, req.PaymentMethod, req.DeliveryAddr).
		Suffix("RETURNING id, user_id, total_amount::float8, COALESCE(status, 'pending') as status, COALESCE(payment_method, '') as payment_method, COALESCE(payment_status, 'pending') as payment_status, delivery_address, created_at, updated_at").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build order insert query")
		return nil, fmt.Errorf("failed to build order insert query: %w", err)
	}

	var order models.Order
	err = tx.QueryRow(ctx, orderQuery, orderArgs...).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.DeliveryAddr,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to create order")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	orderItems := []models.OrderItem{}
	for _, cartItem := range items {
		itemQuery, itemArgs, err := psql.Insert("order_items").
			Columns("order_id", "product_id", "quantity", "size", "price").
			Values(order.ID, cartItem.ProductID, cartItem.Quantity, cartItem.Size, cartItem.ProductPrice).
			Suffix("RETURNING id, order_id, product_id, quantity, COALESCE(size, '') as size, price::float8, created_at").
			ToSql()
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to build order item insert query")
			return nil, fmt.Errorf("failed to build order item insert query: %w", err)
		}

		var orderItem models.OrderItem
		err = tx.QueryRow(ctx, itemQuery, itemArgs...).Scan(
			&orderItem.ID,
			&orderItem.OrderID,
			&orderItem.ProductID,
			&orderItem.Quantity,
			&orderItem.Size,
			&orderItem.Price,
			&orderItem.CreatedAt,
		)
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to create order item")
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		orderItems = append(orderItems, orderItem)

		updateStockQuery, updateStockArgs, err := psql.Update("products").
			Set("stock", sq.Expr("stock - ?", cartItem.Quantity)).
			Where(sq.And{
				sq.Eq{"id": cartItem.ProductID},
				sq.GtOrEq{"stock": cartItem.Quantity},
			}).
			ToSql()
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to build stock update query")
			return nil, fmt.Errorf("failed to build stock update query: %w", err)
		}

		result, err := tx.Exec(ctx, updateStockQuery, updateStockArgs...)
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to update product stock")
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
		if result.RowsAffected() == 0 {
			logger.GetLogger().WithField("product_id", cartItem.ProductID).Error("insufficient stock for product")
			return nil, fmt.Errorf("insufficient stock for product %d", cartItem.ProductID)
		}
	}

	clearCartQuery, clearCartArgs, err := psql.Delete("carts").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build clear cart query")
		return nil, fmt.Errorf("failed to build clear cart query: %w", err)
	}

	_, err = tx.Exec(ctx, clearCartQuery, clearCartArgs...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to clear cart")
		return nil, fmt.Errorf("failed to clear cart: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to commit transaction")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.OrderWithItems{
		Order: order,
		Items: orderItems,
	}, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	orderQuery, orderArgs, err := psql.Select(
		"id", "user_id", "total_amount::float8", "COALESCE(status, 'pending') as status", "COALESCE(payment_method, '') as payment_method",
		"COALESCE(payment_status, 'pending') as payment_status", "delivery_address", "created_at", "updated_at",
	).From("orders").
		Where(sq.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build order select query")
		return nil, fmt.Errorf("failed to build order select query: %w", err)
	}

	var order models.Order
	err = r.db.QueryRow(ctx, orderQuery, orderArgs...).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.DeliveryAddr,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get order")
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	itemsQuery, itemsArgs, err := psql.Select(
		"id", "order_id", "product_id", "quantity", "COALESCE(size, '') as size", "price::float8", "created_at",
	).From("order_items").
		Where(sq.Eq{"order_id": orderID}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build order items select query")
		return nil, fmt.Errorf("failed to build order items select query: %w", err)
	}

	rows, err := r.db.Query(ctx, itemsQuery, itemsArgs...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get order items")
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Size,
			&item.Price,
			&item.CreatedAt,
		); err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to scan order item")
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return &models.OrderWithItems{
		Order: order,
		Items: items,
	}, nil
}

func (r *OrderRepository) GetUserOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	query, args, err := psql.Select(
		"id", "user_id", "total_amount::float8", "COALESCE(status, 'pending') as status", "COALESCE(payment_method, '') as payment_method",
		"COALESCE(payment_status, 'pending') as payment_status", "delivery_address", "created_at", "updated_at",
	).From("orders").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build user orders select query")
		return nil, fmt.Errorf("failed to build user orders select query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get user orders")
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TotalAmount,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.DeliveryAddr,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to scan order")
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *OrderRepository) GetAll(ctx context.Context) ([]*models.Order, error) {
	query, args, err := psql.Select(
		"id", "user_id", "total_amount::float8", "COALESCE(status, 'pending') as status", "COALESCE(payment_method, '') as payment_method",
		"COALESCE(payment_status, 'pending') as payment_status", "delivery_address", "created_at", "updated_at",
	).From("orders").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build all orders select query")
		return nil, fmt.Errorf("failed to build all orders select query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get all orders")
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TotalAmount,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.DeliveryAddr,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to scan order")
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int, status string) (*models.Order, error) {
	query, args, err := psql.Update("orders").
		Set("status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": orderID}).
		Suffix("RETURNING id, user_id, total_amount::float8, COALESCE(status, 'pending') as status, COALESCE(payment_method, '') as payment_method, COALESCE(payment_status, 'pending') as payment_status, delivery_address, created_at, updated_at").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build update status query")
		return nil, fmt.Errorf("failed to build update status query: %w", err)
	}

	var order models.Order
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.DeliveryAddr,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			logger.GetLogger().WithField("order_id", orderID).Error("order not found")
			return nil, fmt.Errorf("order not found")
		}
		logger.GetLogger().WithField("err", err).Error("failed to update order status")
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return &order, nil
}
