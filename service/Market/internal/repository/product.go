package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, sellerID int, req *models.CreateProductRequest) (*models.Product, error) {
	query, args, err := psql.Insert("products").
		Columns("seller_id", "category_id", "title", "description", "price", "stock", "sizes", "image_url").
		Values(sellerID, req.CategoryID, req.Title, req.Description, req.Price, req.Stock, req.Sizes, req.ImageURL).
		Suffix("RETURNING id, seller_id, category_id, title, COALESCE(description, '') as description, price::float8, stock, sizes, COALESCE(image_url, '') as image_url, COALESCE(status, 'pending') as status, created_at, updated_at").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build insert query")
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	var product models.Product
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&product.ID,
		&product.SellerID,
		&product.CategoryID,
		&product.Title,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.Sizes,
		&product.ImageURL,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to create product")
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (*models.ProductWithDetails, error) {
	query, args, err := psql.Select(
		"p.id", "p.seller_id", "p.category_id", "p.title", "COALESCE(p.description, '') as description",
		"p.price::float8", "p.stock", "p.sizes", "COALESCE(p.image_url, '') as image_url", "COALESCE(p.status, 'pending') as status",
		"p.created_at", "p.updated_at",
		"COALESCE(s.shop_name, '') as seller_name",
		"COALESCE(c.name, '') as category_name",
	).From("products p").
		LeftJoin("sellers s ON p.seller_id = s.id").
		LeftJoin("categories c ON p.category_id = c.id").
		Where(sq.Eq{"p.id": id}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build select query")
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var product models.ProductWithDetails
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&product.ID,
		&product.SellerID,
		&product.CategoryID,
		&product.Title,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.Sizes,
		&product.ImageURL,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.SellerName,
		&product.CategoryName,
	)

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get product")
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetAll(ctx context.Context, categoryID, sellerID *int, status string, pagination *models.PaginationParams) ([]*models.ProductWithDetails, int64, error) {
	countBuilder := psql.Select("COUNT(*)").
		From("products p").
		Where("p.category_id IS NOT NULL")

	if categoryID != nil {
		countBuilder = countBuilder.Where(sq.Eq{"p.category_id": *categoryID})
	}
	if sellerID != nil {
		countBuilder = countBuilder.Where(sq.Eq{"p.seller_id": *sellerID})
	}
	if status != "" {
		countBuilder = countBuilder.Where(sq.Eq{"p.status": status})
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build count query")
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var totalItems int64
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalItems)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to count products")
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	selectBuilder := psql.Select(
		"p.id", "p.seller_id", "p.category_id", "p.title", "COALESCE(p.description, '') as description",
		"p.price::float8", "p.stock", "p.sizes", "COALESCE(p.image_url, '') as image_url", "COALESCE(p.status, 'pending') as status",
		"p.created_at", "p.updated_at",
		"COALESCE(s.shop_name, '') as seller_name",
		"COALESCE(c.name, '') as category_name",
	).
		From("products p").
		LeftJoin("sellers s ON p.seller_id = s.id").
		LeftJoin("categories c ON p.category_id = c.id").
		Where("p.category_id IS NOT NULL").
		OrderBy("p.created_at DESC")

	if categoryID != nil {
		selectBuilder = selectBuilder.Where(sq.Eq{"p.category_id": *categoryID})
	}
	if sellerID != nil {
		selectBuilder = selectBuilder.Where(sq.Eq{"p.seller_id": *sellerID})
	}
	if status != "" {
		selectBuilder = selectBuilder.Where(sq.Eq{"p.status": status})
	}

	if pagination != nil {
		selectBuilder = selectBuilder.Limit(uint64(pagination.GetLimit())).Offset(uint64(pagination.GetOffset()))
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build select query")
		return nil, 0, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get products")
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []*models.ProductWithDetails
	for rows.Next() {
		var product models.ProductWithDetails
		if err := rows.Scan(
			&product.ID,
			&product.SellerID,
			&product.CategoryID,
			&product.Title,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Sizes,
			&product.ImageURL,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.SellerName,
			&product.CategoryName,
		); err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to scan product")
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, totalItems, nil
}

func (r *ProductRepository) Update(ctx context.Context, id int, req *models.UpdateProductRequest) (*models.Product, error) {
	updateBuilder := psql.Update("products").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, seller_id, category_id, title, COALESCE(description, '') as description, price::float8, stock, sizes, COALESCE(image_url, '') as image_url, COALESCE(status, 'pending') as status, created_at, updated_at")

	if req.CategoryID != nil {
		updateBuilder = updateBuilder.Set("category_id", *req.CategoryID)
	}
	if req.Title != nil {
		updateBuilder = updateBuilder.Set("title", *req.Title)
	}
	if req.Description != nil {
		updateBuilder = updateBuilder.Set("description", *req.Description)
	}
	if req.Price != nil {
		updateBuilder = updateBuilder.Set("price", *req.Price)
	}
	if req.Stock != nil {
		updateBuilder = updateBuilder.Set("stock", *req.Stock)
	}
	if req.Sizes != nil {
		updateBuilder = updateBuilder.Set("sizes", *req.Sizes)
	}
	if req.ImageURL != nil {
		updateBuilder = updateBuilder.Set("image_url", *req.ImageURL)
	}
	if req.Status != nil {
		updateBuilder = updateBuilder.Set("status", *req.Status)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build update query")
		return nil, fmt.Errorf("failed to build update query: %w", err)
	}

	var product models.Product
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&product.ID,
		&product.SellerID,
		&product.CategoryID,
		&product.Title,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.Sizes,
		&product.ImageURL,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to update product")
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	query, args, err := psql.Delete("products").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build delete query")
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to delete product")
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		logger.GetLogger().WithField("product_id", id).Error("product not found")
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *ProductRepository) GetBySellerID(ctx context.Context, sellerID int) ([]*models.Product, error) {
	query, args, err := psql.Select(
		"id", "seller_id", "category_id", "title", "COALESCE(description, '') as description",
		"price::float8", "stock", "sizes", "COALESCE(image_url, '') as image_url", "COALESCE(status, 'pending') as status", "created_at", "updated_at",
	).From("products").
		Where(sq.Eq{"seller_id": sellerID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to build select query")
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to get products by seller")
		return nil, fmt.Errorf("failed to get products by seller: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(
			&product.ID,
			&product.SellerID,
			&product.CategoryID,
			&product.Title,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Sizes,
			&product.ImageURL,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			logger.GetLogger().WithField("err", err).Error("failed to scan product")
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}
