package repository

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Zifeldev/marketback/service/Market/internal/cache"
	"github.com/Zifeldev/marketback/service/Market/internal/logger"
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
	query, args, err := psql.Insert("categories").
		Columns("name", "description").
		Values(req.Name, req.Description).
		Suffix("RETURNING id, name, description, created_at, updated_at").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build insert category query")
		return nil, fmt.Errorf("failed to build insert category query: %w", err)
	}

	var category models.Category
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to create category")
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int) (*models.Category, error) {
	query, args, err := psql.Select("id", "name", "description", "created_at", "updated_at").
		From("categories").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build select category query")
		return nil, fmt.Errorf("failed to build select category query: %w", err)
	}

	var category models.Category
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get category")
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

	query, args, err := psql.Select("id", "name", "description", "created_at", "updated_at").
		From("categories").
		OrderBy("name").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build select all categories query")
		return nil, fmt.Errorf("failed to build select all categories query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get categories")
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
			logger.GetLogger().WithField("err", err).Error("failed to scan category")
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
	updateBuilder := psql.Update("categories").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, name, description, created_at, updated_at")

	if req.Name != "" {
		updateBuilder = updateBuilder.Set("name", req.Name)
	}
	if req.Description != "" {
		updateBuilder = updateBuilder.Set("description", req.Description)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build update category query")
		return nil, fmt.Errorf("failed to build update category query: %w", err)
	}

	var category models.Category
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to update category")
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
	query, args, err := psql.Delete("categories").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build delete category query")
		return fmt.Errorf("failed to build delete category query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to delete category")
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		logger.GetLogger().WithField("category_id", id).Error("category not found")
		return fmt.Errorf("category not found")
	}

	return nil
}
