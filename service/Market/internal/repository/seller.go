package repository

import (
	"context"
	"fmt"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SellerRepository struct {
	db *pgxpool.Pool
}

func NewSellerRepository(db *pgxpool.Pool) *SellerRepository {
	return &SellerRepository{db: db}
}

func (r *SellerRepository) Create(ctx context.Context, userID int, req *models.CreateSellerRequest) (*models.Seller, error) {
	query := `
		INSERT INTO sellers (user_id, shop_name, description)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, shop_name, description, rating, is_active, created_at, updated_at
	`

	var seller models.Seller
	err := r.db.QueryRow(ctx, query, userID, req.ShopName, req.Description).Scan(
		&seller.ID,
		&seller.UserID,
		&seller.ShopName,
		&seller.Description,
		&seller.Rating,
		&seller.IsActive,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create seller: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) GetByID(ctx context.Context, id int) (*models.Seller, error) {
	query := `
		SELECT id, user_id, shop_name, description, rating, is_active, created_at, updated_at
		FROM sellers
		WHERE id = $1
	`

	var seller models.Seller
	err := r.db.QueryRow(ctx, query, id).Scan(
		&seller.ID,
		&seller.UserID,
		&seller.ShopName,
		&seller.Description,
		&seller.Rating,
		&seller.IsActive,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get seller: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) GetByUserID(ctx context.Context, userID int) (*models.Seller, error) {
	query := `
		SELECT id, user_id, shop_name, description, rating, is_active, created_at, updated_at
		FROM sellers
		WHERE user_id = $1
	`

	var seller models.Seller
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&seller.ID,
		&seller.UserID,
		&seller.ShopName,
		&seller.Description,
		&seller.Rating,
		&seller.IsActive,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get seller by user ID: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) Update(ctx context.Context, id int, req *models.UpdateSellerRequest) (*models.Seller, error) {
	query := `
		UPDATE sellers
		SET shop_name = COALESCE(NULLIF($1, ''), shop_name),
		    description = COALESCE(NULLIF($2, ''), description),
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id, user_id, shop_name, description, rating, is_active, created_at, updated_at
	`

	var seller models.Seller
	err := r.db.QueryRow(ctx, query, req.ShopName, req.Description, id).Scan(
		&seller.ID,
		&seller.UserID,
		&seller.ShopName,
		&seller.Description,
		&seller.Rating,
		&seller.IsActive,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update seller: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) UpdateStatus(ctx context.Context, id int, isActive bool) error {
	query := `
		UPDATE sellers
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, isActive, id)
	if err != nil {
		return fmt.Errorf("failed to update seller status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("seller not found")
	}

	return nil
}

func (r *SellerRepository) GetAll(ctx context.Context) ([]*models.Seller, error) {
	query := `
		SELECT id, user_id, shop_name, description, rating, is_active, created_at, updated_at
		FROM sellers
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sellers: %w", err)
	}
	defer rows.Close()

	var sellers []*models.Seller
	for rows.Next() {
		var seller models.Seller
		if err := rows.Scan(
			&seller.ID,
			&seller.UserID,
			&seller.ShopName,
			&seller.Description,
			&seller.Rating,
			&seller.IsActive,
			&seller.CreatedAt,
			&seller.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan seller: %w", err)
		}
		sellers = append(sellers, &seller)
	}

	return sellers, nil
}
