package repository

import (
	"context"
	"fmt"
	"time"

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

	for _, item := range items {
		var currentStock int
		lockQuery := `SELECT stock FROM products WHERE id = $1 FOR UPDATE`
		err := tx.QueryRow(ctx, lockQuery, item.ProductID).Scan(&currentStock)
		if err != nil {
			if err == pgx.ErrNoRows {
				logger.GetLogger().WithField("product_id", item.ProductID).Error("product not found")
				return nil, fmt.Errorf("product %d not found", item.ProductID)
			}
			logger.GetLogger().WithField("err", err).Error("failed to lock product for stock check")
			return nil, fmt.Errorf("failed to lock product for stock check: %w", err)
		}

		if currentStock < item.Quantity {
			logger.GetLogger().WithFields(map[string]interface{}{
				"product_id": item.ProductID,
				"requested":  item.Quantity,
				"available":  currentStock,
			}).Error("insufficient stock for product")
			return nil, fmt.Errorf("insufficient stock for product %d: requested %d, available %d",
				item.ProductID, item.Quantity, currentStock)
		}
	}

	for _, item := range items {
		updateStockQuery := `UPDATE products SET stock = stock - $1, updated_at = NOW() 
			WHERE id = $2 AND stock >= $1`

		result, err := tx.Exec(ctx, updateStockQuery, item.Quantity, item.ProductID)
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to update product stock")
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}

		if result.RowsAffected() != 1 {
			logger.GetLogger().WithFields(map[string]interface{}{
				"product_id":    item.ProductID,
				"affected_rows": result.RowsAffected(),
			}).Error("stock update affected unexpected number of rows")
			return nil, fmt.Errorf("failed to deduct stock for product %d: concurrent modification detected", item.ProductID)
		}
	}

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

func (r *OrderRepository) GetUserOrders(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
	countQuery, countArgs, err := psql.Select("COUNT(*)").
		From("orders").
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build count query")
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var totalItems int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalItems); err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to count orders")
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	if totalItems == 0 {
		return []*models.OrderWithItems{}, 0, nil
	}

	query, args, err := psql.Select(
		"o.id", "o.user_id", "o.total_amount::float8",
		"COALESCE(o.status, 'pending') as status",
		"COALESCE(o.payment_method, '') as payment_method",
		"COALESCE(o.payment_status, 'pending') as payment_status",
		"o.delivery_address", "o.created_at", "o.updated_at",
		"oi.id as item_id", "oi.product_id", "oi.quantity",
		"COALESCE(oi.size, '') as size", "oi.price::float8", "oi.created_at as item_created_at",
		"COALESCE(p.title, '') as product_title",
	).From("orders o").
		LeftJoin("order_items oi ON o.id = oi.order_id").
		LeftJoin("products p ON oi.product_id = p.id").
		Where(sq.Eq{"o.user_id": userID}).
		OrderBy("o.created_at DESC", "oi.id").
		Limit(uint64(pagination.GetLimit())).
		Offset(uint64(pagination.GetOffset())).
		ToSql()

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build user orders query")
		return nil, 0, fmt.Errorf("failed to build user orders query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get user orders")
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	ordersMap := make(map[int]*models.OrderWithItems)
	var orderIDs []int

	for rows.Next() {
		var order models.Order
		var itemID, productID, quantity *int
		var size, productTitle *string
		var itemPrice *float64
		var itemCreatedAt *time.Time

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
			&itemID,
			&productID,
			&quantity,
			&size,
			&itemPrice,
			&itemCreatedAt,
			&productTitle,
		); err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to scan order row")
			return nil, 0, fmt.Errorf("failed to scan order row: %w", err)
		}

		if _, exists := ordersMap[order.ID]; !exists {
			ordersMap[order.ID] = &models.OrderWithItems{
				Order: order,
				Items: []models.OrderItem{},
			}
			orderIDs = append(orderIDs, order.ID)
		}

		if itemID != nil {
			item := models.OrderItem{
				ID:        *itemID,
				OrderID:   order.ID,
				ProductID: *productID,
				Quantity:  *quantity,
				Price:     *itemPrice,
				CreatedAt: *itemCreatedAt,
			}
			if size != nil {
				item.Size = *size
			}
			ordersMap[order.ID].Items = append(ordersMap[order.ID].Items, item)
		}
	}

	result := make([]*models.OrderWithItems, 0, len(orderIDs))
	for _, id := range orderIDs {
		result = append(result, ordersMap[id])
	}

	return result, totalItems, nil
}

func (r *OrderRepository) GetAll(ctx context.Context, pagination *models.PaginationParams, status string) ([]*models.OrderWithItems, int64, error) {
	countBuilder := psql.Select("COUNT(*)").From("orders")
	if status != "" {
		countBuilder = countBuilder.Where(sq.Eq{"status": status})
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build count query")
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var totalItems int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalItems); err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to count orders")
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	if totalItems == 0 {
		return []*models.OrderWithItems{}, 0, nil
	}

	queryBuilder := psql.Select(
		"o.id", "o.user_id", "o.total_amount::float8",
		"COALESCE(o.status, 'pending') as status",
		"COALESCE(o.payment_method, '') as payment_method",
		"COALESCE(o.payment_status, 'pending') as payment_status",
		"o.delivery_address", "o.created_at", "o.updated_at",
		"oi.id as item_id", "oi.product_id", "oi.quantity",
		"COALESCE(oi.size, '') as size", "oi.price::float8", "oi.created_at as item_created_at",
		"COALESCE(p.title, '') as product_title",
	).From("orders o").
		LeftJoin("order_items oi ON o.id = oi.order_id").
		LeftJoin("products p ON oi.product_id = p.id")

	if status != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"o.status": status})
	}

	query, args, err := queryBuilder.
		OrderBy("o.created_at DESC", "oi.id").
		Limit(uint64(pagination.GetLimit())).
		Offset(uint64(pagination.GetOffset())).
		ToSql()

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build all orders query")
		return nil, 0, fmt.Errorf("failed to build all orders query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get all orders")
		return nil, 0, fmt.Errorf("failed to get all orders: %w", err)
	}
	defer rows.Close()

	ordersMap := make(map[int]*models.OrderWithItems)
	var orderIDs []int

	for rows.Next() {
		var order models.Order
		var itemID, productID, quantity *int
		var size, productTitle *string
		var itemPrice *float64
		var itemCreatedAt *time.Time

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
			&itemID,
			&productID,
			&quantity,
			&size,
			&itemPrice,
			&itemCreatedAt,
			&productTitle,
		); err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to scan order row")
			return nil, 0, fmt.Errorf("failed to scan order row: %w", err)
		}

		if _, exists := ordersMap[order.ID]; !exists {
			ordersMap[order.ID] = &models.OrderWithItems{
				Order: order,
				Items: []models.OrderItem{},
			}
			orderIDs = append(orderIDs, order.ID)
		}

		if itemID != nil {
			item := models.OrderItem{
				ID:        *itemID,
				OrderID:   order.ID,
				ProductID: *productID,
				Quantity:  *quantity,
				Price:     *itemPrice,
				CreatedAt: *itemCreatedAt,
			}
			if size != nil {
				item.Size = *size
			}
			ordersMap[order.ID].Items = append(ordersMap[order.ID].Items, item)
		}
	}

	result := make([]*models.OrderWithItems, 0, len(orderIDs))
	for _, id := range orderIDs {
		result = append(result, ordersMap[id])
	}

	return result, totalItems, nil
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
