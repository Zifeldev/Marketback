package repository

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var totalAmount float64
	for _, item := range items {
		totalAmount += item.ProductPrice * float64(item.Quantity)
	}

	orderQuery := `
		INSERT INTO orders (user_id, total_amount, payment_method, delivery_address)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, total_amount, status, payment_method, payment_status, delivery_address, created_at, updated_at
	`

	var order models.Order
	err = tx.QueryRow(ctx, orderQuery, userID, totalAmount, req.PaymentMethod, req.DeliveryAddr).Scan(
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
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	orderItems := []models.OrderItem{}
	for _, cartItem := range items {
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, size, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, order_id, product_id, quantity, size, price, created_at
		`

		var orderItem models.OrderItem
		err = tx.QueryRow(ctx, itemQuery,
			order.ID,
			cartItem.ProductID,
			cartItem.Quantity,
			cartItem.Size,
			cartItem.ProductPrice,
		).Scan(
			&orderItem.ID,
			&orderItem.OrderID,
			&orderItem.ProductID,
			&orderItem.Quantity,
			&orderItem.Size,
			&orderItem.Price,
			&orderItem.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		orderItems = append(orderItems, orderItem)

		updateStockQuery := `
			UPDATE products
			SET stock = stock - $1
			WHERE id = $2 AND stock >= $1
		`
		result, err := tx.Exec(ctx, updateStockQuery, cartItem.Quantity, cartItem.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
		if result.RowsAffected() == 0 {
			return nil, fmt.Errorf("insufficient stock for product %d", cartItem.ProductID)
		}
	}

	clearCartQuery := `DELETE FROM carts WHERE user_id = $1`
	_, err = tx.Exec(ctx, clearCartQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear cart: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.OrderWithItems{
		Order: order,
		Items: orderItems,
	}, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	orderQuery := `
		SELECT id, user_id, total_amount, status, payment_method, payment_status, delivery_address, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order models.Order
	err := r.db.QueryRow(ctx, orderQuery, orderID).Scan(
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
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	itemsQuery := `
		SELECT id, order_id, product_id, quantity, size, price, created_at
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Query(ctx, itemsQuery, orderID)
	if err != nil {
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
	query := `
		SELECT id, user_id, total_amount, status, payment_method, payment_status, delivery_address, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
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
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *OrderRepository) GetAll(ctx context.Context) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, payment_method, payment_status, delivery_address, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
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
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int, status string) (*models.Order, error) {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, total_amount, status, payment_method, payment_status, delivery_address, created_at, updated_at
	`

	var order models.Order
	err := r.db.QueryRow(ctx, query, status, orderID).Scan(
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
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return &order, nil
}
