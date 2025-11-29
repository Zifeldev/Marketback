package repository

import (
	"context"
	"fmt"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, sellerID int, req *models.CreateProductRequest) (*models.Product, error) {
	query := `
		INSERT INTO products (seller_id, category_id, title, description, price, stock, sizes, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, seller_id, category_id, title, description, price, stock, sizes, image_url, status, created_at, updated_at
	`

	var product models.Product
	err := r.db.QueryRow(ctx, query,
		sellerID,
		req.CategoryID,
		req.Title,
		req.Description,
		req.Price,
		req.Stock,
		req.Sizes,
		req.ImageURL,
	).Scan(
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
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (*models.ProductWithDetails, error) {
	query := `
		SELECT 
			p.id, p.seller_id, p.category_id, p.title, p.description, 
			p.price, p.stock, p.sizes, p.image_url, p.status, 
			p.created_at, p.updated_at,
			COALESCE(s.shop_name, '') as seller_name,
			COALESCE(c.name, '') as category_name
		FROM products p
		LEFT JOIN sellers s ON p.seller_id = s.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`

	var product models.ProductWithDetails
	err := r.db.QueryRow(ctx, query, id).Scan(
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
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetAll(ctx context.Context, categoryID, sellerID *int, status string, pagination *models.PaginationParams) ([]*models.ProductWithDetails, int64, error) {
	countQuery := `
		SELECT COUNT(*) 
		FROM products p
		WHERE 1=1 AND p.category_id IS NOT NULL
	`

	countArgs := []interface{}{}
	argNum := 1

	if categoryID != nil {
		countQuery += fmt.Sprintf(" AND p.category_id = $%d", argNum)
		countArgs = append(countArgs, *categoryID)
		argNum++
	}

	if sellerID != nil {
		countQuery += fmt.Sprintf(" AND p.seller_id = $%d", argNum)
		countArgs = append(countArgs, *sellerID)
		argNum++
	}

	if status != "" {
		countQuery += fmt.Sprintf(" AND p.status = $%d", argNum)
		countArgs = append(countArgs, status)
		argNum++
	}

	var totalItems int64
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}


	query := `
		SELECT 
			p.id, p.seller_id, p.category_id, p.title, p.description, 
			p.price, p.stock, p.sizes, p.image_url, p.status, 
			p.created_at, p.updated_at,
			COALESCE(s.shop_name, '') as seller_name,
			COALESCE(c.name, '') as category_name
		FROM products p
		LEFT JOIN sellers s ON p.seller_id = s.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE 1=1 AND p.category_id IS NOT NULL
	`

	args := []interface{}{}
	argNum = 1

	if categoryID != nil {
		query += fmt.Sprintf(" AND p.category_id = $%d", argNum)
		args = append(args, *categoryID)
		argNum++
	}

	if sellerID != nil {
		query += fmt.Sprintf(" AND p.seller_id = $%d", argNum)
		args = append(args, *sellerID)
		argNum++
	}

	if status != "" {
		query += fmt.Sprintf(" AND p.status = $%d", argNum)
		args = append(args, status)
		argNum++
	}

	query += " ORDER BY p.created_at DESC"

	// Add pagination
	if pagination != nil {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, pagination.GetLimit(), pagination.GetOffset())
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
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
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, totalItems, nil
}

func (r *ProductRepository) Update(ctx context.Context, id int, req *models.UpdateProductRequest) (*models.Product, error) {
	query := `
		UPDATE products
		SET category_id = COALESCE($1, category_id),
		    title = COALESCE($2, title),
		    description = COALESCE($3, description),
		    price = COALESCE($4, price),
		    stock = COALESCE($5, stock),
		    sizes = COALESCE($6, sizes),
		    image_url = COALESCE($7, image_url),
		    status = COALESCE($8, status),
		    updated_at = NOW()
		WHERE id = $9
		RETURNING id, seller_id, category_id, title, description, price, stock, sizes, image_url, status, created_at, updated_at
	`

	var product models.Product
	err := r.db.QueryRow(ctx, query,
		req.CategoryID,
		req.Title,
		req.Description,
		req.Price,
		req.Stock,
		req.Sizes,
		req.ImageURL,
		req.Status,
		id,
	).Scan(
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
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *ProductRepository) GetBySellerID(ctx context.Context, sellerID int) ([]*models.Product, error) {
	query := `
		SELECT id, seller_id, category_id, title, description, price, stock, sizes, image_url, status, created_at, updated_at
		FROM products
		WHERE seller_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, sellerID)
	if err != nil {
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
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}
