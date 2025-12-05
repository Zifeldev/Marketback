package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Zifeldev/marketback/service/Market/internal/logger"
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
	query, args, err := psql.Insert("sellers").
		Columns("user_id", "shop_name", "description").
		Values(userID, req.ShopName, req.Description).
		Suffix("RETURNING id, user_id, shop_name, description, rating::float8, is_active, created_at, updated_at").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build insert seller query")
		return nil, fmt.Errorf("failed to build insert seller query: %w", err)
	}

	var seller models.Seller
	err = r.db.QueryRow(ctx, query, args...).Scan(
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
		logger.GetLogger().WithField("err", err).Error("failed to create seller")
		return nil, fmt.Errorf("failed to create seller: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) GetByID(ctx context.Context, id int) (*models.Seller, error) {
	query := `SELECT id, user_id, shop_name, COALESCE(description, '') as description, rating::float8 as rating, is_active, created_at, updated_at FROM sellers WHERE id = $1`

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
		logger.GetLogger().WithField("err", err).Error("failed to get seller")
		return nil, fmt.Errorf("failed to get seller: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) GetByUserID(ctx context.Context, userID int) (*models.Seller, error) {
	query := `SELECT id, user_id, shop_name, COALESCE(description, '') as description, rating::float8 as rating, is_active, created_at, updated_at FROM sellers WHERE user_id = $1`

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
		logger.GetLogger().WithField("err", err).Error("failed to get seller by user ID")
		return nil, fmt.Errorf("failed to get seller by user ID: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) Update(ctx context.Context, id int, req *models.UpdateSellerRequest) (*models.Seller, error) {
	updateBuilder := psql.Update("sellers").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, user_id, shop_name, description, rating::float8, is_active, created_at, updated_at")

	if req.ShopName != "" {
		updateBuilder = updateBuilder.Set("shop_name", req.ShopName)
	}
	if req.Description != "" {
		updateBuilder = updateBuilder.Set("description", req.Description)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build update seller query")
		return nil, fmt.Errorf("failed to build update seller query: %w", err)
	}

	var seller models.Seller
	err = r.db.QueryRow(ctx, query, args...).Scan(
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
		logger.GetLogger().WithField("err", err).Error("failed to update seller")
		return nil, fmt.Errorf("failed to update seller: %w", err)
	}

	return &seller, nil
}

func (r *SellerRepository) UpdateStatus(ctx context.Context, id int, isActive bool) error {
	query, args, err := psql.Update("sellers").
		Set("is_active", isActive).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build update seller status query")
		return fmt.Errorf("failed to build update seller status query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to update seller status")
		return fmt.Errorf("failed to update seller status: %w", err)
	}

	if result.RowsAffected() == 0 {
		logger.GetLogger().WithField("seller_id", id).Error("seller not found")
		return fmt.Errorf("seller not found")
	}

	return nil
}

func (r *SellerRepository) GetAll(ctx context.Context) ([]*models.Seller, error) {
	query := `SELECT id, user_id, shop_name, COALESCE(description, '') as description, rating::float8 as rating, is_active, created_at, updated_at FROM sellers ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get sellers")
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
			logger.GetLogger().WithField("err", err).Error("failed to scan seller")
			return nil, fmt.Errorf("failed to scan seller: %w", err)
		}
		sellers = append(sellers, &seller)
	}

	return sellers, nil
}
