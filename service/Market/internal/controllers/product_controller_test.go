package controllers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// mockProductRepo implements ProductRepo for tests
type mockProductRepo struct {
	getAllFn  func(ctx context.Context, categoryID, sellerID *int, status string, p *models.PaginationParams) ([]*models.ProductWithDetails, int64, error)
	getByIDFn func(ctx context.Context, id int) (*models.ProductWithDetails, error)
}

func (m *mockProductRepo) GetAll(ctx context.Context, categoryID, sellerID *int, status string, p *models.PaginationParams) ([]*models.ProductWithDetails, int64, error) {
	return m.getAllFn(ctx, categoryID, sellerID, status, p)
}
func (m *mockProductRepo) GetByID(ctx context.Context, id int) (*models.ProductWithDetails, error) {
	return m.getByIDFn(ctx, id)
}

var _ repository.ProductRepo = (*mockProductRepo)(nil)

// mockCategoryRepo minimal for controller construction
type mockCategoryRepo struct {
	getAllFn  func(ctx context.Context) ([]*models.Category, error)
	getByIDFn func(ctx context.Context, id int) (*models.Category, error)
}

func (m *mockCategoryRepo) GetAll(ctx context.Context) ([]*models.Category, error) {
	return m.getAllFn(ctx)
}
func (m *mockCategoryRepo) GetByID(ctx context.Context, id int) (*models.Category, error) {
	return m.getByIDFn(ctx, id)
}

var _ repository.CategoryRepo = (*mockCategoryRepo)(nil)

// mockOrderRepo minimal for controller construction
type mockOrderRepo struct {
	getUserFn func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error)
	getByIDFn func(ctx context.Context, orderID int) (*models.OrderWithItems, error)
}

func (m *mockOrderRepo) GetUserOrders(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
	return m.getUserFn(ctx, userID, pagination)
}
func (m *mockOrderRepo) GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	return m.getByIDFn(ctx, orderID)
}

var _ repository.OrderRepo = (*mockOrderRepo)(nil)

func TestMarketController_GetProducts_PaginationAndFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	// Query params
	req := httptest.NewRequest("GET", "/api/products?category_id=5&seller_id=9&status=active&page=2&page_size=3", nil)
	c.Request = req

	prod := &models.ProductWithDetails{Product: models.Product{ID: 101, SellerID: 9, CategoryID: 5, Title: "Boots", Price: 77.7, CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	var capturedCat, capturedSeller *int
	var capturedStatus string
	var capturedPage, capturedLimit int

	mProd := &mockProductRepo{getAllFn: func(ctx context.Context, categoryID, sellerID *int, status string, p *models.PaginationParams) ([]*models.ProductWithDetails, int64, error) {
		capturedCat, capturedSeller = categoryID, sellerID
		capturedStatus = status
		capturedPage = p.Page
		capturedLimit = p.GetLimit()
		return []*models.ProductWithDetails{prod}, 11, nil // totalItems=11
	}}
	mCat := &mockCategoryRepo{getAllFn: func(ctx context.Context) ([]*models.Category, error) { return nil, nil }, getByIDFn: func(ctx context.Context, id int) (*models.Category, error) { return nil, nil }}
	mOrder := &mockOrderRepo{
		getUserFn: func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
			return nil, 0, nil
		},
		getByIDFn: func(ctx context.Context, orderID int) (*models.OrderWithItems, error) { return nil, nil },
	}

	mc := NewMarketController(mProd, mCat, nil, mOrder, nil)
	mc.GetProducts(c)

	require.Equal(t, 200, r.Code)
	require.NotNil(t, capturedCat)
	require.NotNil(t, capturedSeller)
	require.Equal(t, "active", capturedStatus)
	require.Equal(t, 2, capturedPage)
	require.Equal(t, 3, capturedLimit)

	var resp struct {
		Data       []models.ProductWithDetails `json:"data"`
		Pagination models.PaginationMeta       `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, 101, resp.Data[0].ID)
	// totalItems=11, pageSize=3 => totalPages = 4
	require.Equal(t, int64(11), resp.Pagination.TotalItems)
	require.Equal(t, 4, resp.Pagination.TotalPages)
}

func TestMarketController_GetProducts_DefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	// No page_size/page
	req := httptest.NewRequest("GET", "/api/products", nil)
	c.Request = req
	mProd := &mockProductRepo{getAllFn: func(ctx context.Context, categoryID, sellerID *int, status string, p *models.PaginationParams) ([]*models.ProductWithDetails, int64, error) {
		if p.Page == 0 {
			p.Page = 1
		} // mirror controller's implicit sanitation
		require.Equal(t, 1, p.Page) // default
		require.Equal(t, 20, p.GetLimit())
		return []*models.ProductWithDetails{}, 0, nil
	}}
	mCat := &mockCategoryRepo{getAllFn: func(ctx context.Context) ([]*models.Category, error) { return nil, nil }, getByIDFn: func(ctx context.Context, id int) (*models.Category, error) { return nil, nil }}
	mOrder := &mockOrderRepo{
		getUserFn: func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
			return nil, 0, nil
		},
		getByIDFn: func(ctx context.Context, orderID int) (*models.OrderWithItems, error) { return nil, nil },
	}
	mc := NewMarketController(mProd, mCat, nil, mOrder, nil)
	mc.GetProducts(c)
	require.Equal(t, 200, r.Code)
}

// helper to silence unused import of strconv in case future tests use conversions
var _ = strconv.Atoi
