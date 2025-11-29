package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/cache"
	"github.com/Zifeldev/marketback/service/Market/internal/metrics"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	db    *pgxpool.Pool
	cache *cache.RedisCache
}

func NewCategoryRepository(db *pgxpool.Pool, cache *cache.RedisCache) *CategoryRepository {
	return &CategoryRepository{
		db:    db,
		cache: cache,
	}
}

func (r *CategoryRepository) Create(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	query := `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`

	var category models.Category
	err := r.db.QueryRow(ctx, query, req.Name, req.Description).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int) (*models.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var category models.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]*models.Category, error) {
	cacheKey := "categories:all"
	var categories []*models.Category

	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &categories); err == nil {
			metrics.RedisHitsTotal.Inc()
			return categories, nil
		}
		metrics.RedisMissesTotal.Inc()
	}

	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	categories = []*models.Category{}
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &category)
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, categories, 10*time.Minute)
	}

	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, id int, req *models.UpdateCategoryRequest) (*models.Category, error) {
	query := `
		UPDATE categories
		SET name = COALESCE(NULLIF($1, ''), name),
		    description = COALESCE(NULLIF($2, ''), description),
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description, created_at, updated_at
	`

	var category models.Category
	err := r.db.QueryRow(ctx, query, req.Name, req.Description, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
