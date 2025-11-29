package models

import "time"

type Order struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	TotalAmount   float64   `json:"total_amount" db:"total_amount"`
	Status        string    `json:"status" db:"status"`
	PaymentMethod string    `json:"payment_method" db:"payment_method"`
	PaymentStatus string    `json:"payment_status" db:"payment_status"`
	DeliveryAddr  string    `json:"delivery_address" db:"delivery_address"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID        int       `json:"id" db:"id"`
	OrderID   int       `json:"order_id" db:"order_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Size      string    `json:"size" db:"size"`
	Price     float64   `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

type CreateOrderRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required"`
	DeliveryAddr  string `json:"delivery_address" binding:"required"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
