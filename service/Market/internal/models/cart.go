package models

import "time"

type CartItem struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Size      string    `json:"size" db:"size"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CartItemWithDetails struct {
	CartItem
	ProductTitle string  `json:"product_title" db:"product_title"`
	ProductPrice float64 `json:"product_price" db:"product_price"`
	ProductImage string  `json:"product_image" db:"product_image"`
}

type AddToCartRequest struct {
	ProductID int    `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
	Size      string `json:"size"`
}

type UpdateCartItemRequest struct {
	Quantity int    `json:"quantity" binding:"required,gt=0"`
	Size     string `json:"size"`
}
