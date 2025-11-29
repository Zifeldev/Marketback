package models

import "time"

type Seller struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	ShopName    string    `json:"shop_name" db:"shop_name"`
	Description string    `json:"description" db:"description"`
	Rating      float64   `json:"rating" db:"rating"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSellerRequest struct {
	ShopName    string `json:"shop_name" binding:"required"`
	Description string `json:"description"`
}

type UpdateSellerRequest struct {
	ShopName    string `json:"shop_name"`
	Description string `json:"description"`
}
