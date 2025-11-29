package models

import "time"

type Product struct {
	ID          int       `json:"id" db:"id"`
	SellerID    int       `json:"seller_id" db:"seller_id"`
	CategoryID  int       `json:"category_id" db:"category_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	Sizes       SizesJSON `json:"sizes" db:"sizes"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ProductWithDetails struct {
	Product
	SellerName   string `json:"seller_name" db:"seller_name"`
	CategoryName string `json:"category_name" db:"category_name"`
}

type CreateProductRequest struct {
	CategoryID  int       `json:"category_id" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	Stock       int       `json:"stock" binding:"required,gte=0"`
	Sizes       SizesJSON `json:"sizes"`
	ImageURL    string    `json:"image_url"`
}

type UpdateProductRequest struct {
	CategoryID  *int       `json:"category_id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Price       *float64   `json:"price"`
	Stock       *int       `json:"stock"`
	Sizes       *SizesJSON `json:"sizes"`
	ImageURL    *string    `json:"image_url"`
	Status      *string    `json:"status"`
}
