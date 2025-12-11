package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// --- Order Controller Tests ---

// mockOrderRepoFull implements OrderRepo interface for order tests
type mockOrderRepoFull struct {
	getUserOrdersFn func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error)
	getByIDFn       func(ctx context.Context, orderID int) (*models.OrderWithItems, error)
}

func (m *mockOrderRepoFull) GetUserOrders(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
	return m.getUserOrdersFn(ctx, userID, pagination)
}

func (m *mockOrderRepoFull) GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	return m.getByIDFn(ctx, orderID)
}

var _ repository.OrderRepo = (*mockOrderRepoFull)(nil)

func TestMarketController_GetUserOrders_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/orders", nil)
	c.Set("user_id", 42)

	now := time.Now()
	orders := []*models.OrderWithItems{
		{Order: models.Order{ID: 1, UserID: 42, TotalAmount: 100, Status: "pending", PaymentMethod: "card", CreatedAt: now}},
		{Order: models.Order{ID: 2, UserID: 42, TotalAmount: 200, Status: "delivered", PaymentMethod: "cash", CreatedAt: now}},
	}

	mOrder := &mockOrderRepoFull{
		getUserOrdersFn: func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
			require.Equal(t, 42, userID)
			return orders, 2, nil
		},
		getByIDFn: func(ctx context.Context, orderID int) (*models.OrderWithItems, error) { return nil, nil },
	}

	mc := NewMarketController(nil, nil, nil, mOrder, nil)
	mc.GetUserOrders(c)

	require.Equal(t, 200, r.Code)
}

func TestMarketController_GetUserOrders_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/orders", nil)
	c.Set("user_id", 99)

	mOrder := &mockOrderRepoFull{
		getUserOrdersFn: func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
			return []*models.OrderWithItems{}, 0, nil
		},
		getByIDFn: func(ctx context.Context, orderID int) (*models.OrderWithItems, error) { return nil, nil },
	}

	mc := NewMarketController(nil, nil, nil, mOrder, nil)
	mc.GetUserOrders(c)

	require.Equal(t, 200, r.Code)
}

func TestMarketController_GetUserOrders_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/orders", nil)
	c.Set("user_id", 42)

	mOrder := &mockOrderRepoFull{
		getUserOrdersFn: func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
			return nil, 0, errors.New("database error")
		},
		getByIDFn: func(ctx context.Context, orderID int) (*models.OrderWithItems, error) { return nil, nil },
	}

	mc := NewMarketController(nil, nil, nil, mOrder, nil)
	mc.GetUserOrders(c)

	require.Equal(t, 500, r.Code)
}

func TestMarketController_GetOrder_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/orders/1", nil)
	c.Set("user_id", 42)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	now := time.Now()
	orderWithItems := &models.OrderWithItems{
		Order: models.Order{
			ID:            1,
			UserID:        42,
			TotalAmount:   150,
			Status:        "pending",
			PaymentMethod: "card",
			DeliveryAddr:  "Test Address",
			CreatedAt:     now,
		},
		Items: []models.OrderItem{
			{ID: 1, OrderID: 1, ProductID: 10, Quantity: 2, Price: 50},
			{ID: 2, OrderID: 1, ProductID: 11, Quantity: 1, Price: 50},
		},
	}

	mOrder := &mockOrderRepoFull{
		getUserOrdersFn: func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
			return nil, 0, nil
		},
		getByIDFn: func(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
			require.Equal(t, 1, orderID)
			return orderWithItems, nil
		},
	}

	mc := NewMarketController(nil, nil, nil, mOrder, nil)
	mc.GetOrder(c)

	require.Equal(t, 200, r.Code)
	var resp models.OrderWithItems
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.ID)
	require.Len(t, resp.Items, 2)
}

func TestMarketController_GetOrder_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/orders/abc", nil)
	c.Set("user_id", 42)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	mc := NewMarketController(nil, nil, nil, nil, nil)
	mc.GetOrder(c)

	require.Equal(t, 400, r.Code)
}

func TestMarketController_GetOrder_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/orders/999", nil)
	c.Set("user_id", 42)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	mOrder := &mockOrderRepoFull{
		getUserOrdersFn: func(ctx context.Context, userID int, pagination *models.PaginationParams) ([]*models.OrderWithItems, int64, error) {
			return nil, 0, nil
		},
		getByIDFn: func(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
			return nil, errors.New("order not found")
		},
	}

	mc := NewMarketController(nil, nil, nil, mOrder, nil)
	mc.GetOrder(c)

	require.Equal(t, 404, r.Code)
}

// --- Category Tests ---

type mockCategoryRepoFull struct {
	getAllFn  func(ctx context.Context) ([]*models.Category, error)
	getByIDFn func(ctx context.Context, id int) (*models.Category, error)
}

func (m *mockCategoryRepoFull) GetAll(ctx context.Context) ([]*models.Category, error) {
	return m.getAllFn(ctx)
}

func (m *mockCategoryRepoFull) GetByID(ctx context.Context, id int) (*models.Category, error) {
	return m.getByIDFn(ctx, id)
}

var _ repository.CategoryRepo = (*mockCategoryRepoFull)(nil)

func TestMarketController_GetCategories_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/categories", nil)

	now := time.Now()
	categories := []*models.Category{
		{ID: 1, Name: "Electronics", Description: "Electronic devices", CreatedAt: now},
		{ID: 2, Name: "Clothing", Description: "Clothes", CreatedAt: now},
	}

	mCat := &mockCategoryRepoFull{
		getAllFn: func(ctx context.Context) ([]*models.Category, error) {
			return categories, nil
		},
		getByIDFn: func(ctx context.Context, id int) (*models.Category, error) { return nil, nil },
	}

	mc := NewMarketController(nil, mCat, nil, nil, nil)
	mc.GetCategories(c)

	require.Equal(t, 200, r.Code)
	var resp []models.Category
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
	require.Len(t, resp, 2)
}

func TestMarketController_GetCategories_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/categories", nil)

	mCat := &mockCategoryRepoFull{
		getAllFn: func(ctx context.Context) ([]*models.Category, error) {
			return nil, errors.New("database error")
		},
		getByIDFn: func(ctx context.Context, id int) (*models.Category, error) { return nil, nil },
	}

	mc := NewMarketController(nil, mCat, nil, nil, nil)
	mc.GetCategories(c)

	require.Equal(t, 500, r.Code)
}

// --- GetProduct Tests ---

func TestMarketController_GetProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/products/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	now := time.Now()
	product := &models.ProductWithDetails{
		Product: models.Product{
			ID:          1,
			SellerID:    1,
			CategoryID:  1,
			Title:       "Test Product",
			Description: "Description",
			Price:       99.99,
			Stock:       10,
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		SellerName:   "Test Seller",
		CategoryName: "Test Category",
	}

	mProd := &mockProductRepo{
		getByIDFn: func(ctx context.Context, id int) (*models.ProductWithDetails, error) {
			require.Equal(t, 1, id)
			return product, nil
		},
		getAllFn: func(ctx context.Context, categoryID, sellerID *int, status string, p *models.PaginationParams) ([]*models.ProductWithDetails, int64, error) {
			return nil, 0, nil
		},
	}

	mc := NewMarketController(mProd, nil, nil, nil, nil)
	mc.GetProduct(c)

	require.Equal(t, 200, r.Code)
	var resp models.ProductWithDetails
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
	require.Equal(t, "Test Product", resp.Title)
	require.Equal(t, "Test Seller", resp.SellerName)
}

func TestMarketController_GetProduct_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/products/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	mc := NewMarketController(nil, nil, nil, nil, nil)
	mc.GetProduct(c)

	require.Equal(t, 400, r.Code)
}

func TestMarketController_GetProduct_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/products/999", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	mProd := &mockProductRepo{
		getByIDFn: func(ctx context.Context, id int) (*models.ProductWithDetails, error) {
			return nil, errors.New("product not found")
		},
		getAllFn: func(ctx context.Context, categoryID, sellerID *int, status string, p *models.PaginationParams) ([]*models.ProductWithDetails, int64, error) {
			return nil, 0, nil
		},
	}

	mc := NewMarketController(mProd, nil, nil, nil, nil)
	mc.GetProduct(c)

	require.Equal(t, 404, r.Code)
}

// --- AddToCart Extended Tests ---

func TestMarketController_AddToCart_SuccessWithSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"product_id":1,"quantity":2,"size":"M"}`
	c.Request = httptest.NewRequest("POST", "/api/cart", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)

	now := time.Now()
	mCart := &mockCartRepoFull{
		addFn: func(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
			require.Equal(t, 42, userID)
			require.Equal(t, 1, req.ProductID)
			require.Equal(t, 2, req.Quantity)
			require.Equal(t, "M", req.Size)
			return &models.CartItem{
				ID:        1,
				UserID:    42,
				ProductID: 1,
				Quantity:  2,
				Size:      "M",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		getFn:    noopGet,
		updateFn: noopUpdate,
		deleteFn: noopDelete,
		clearFn:  noopClear,
	}

	mc := NewMarketController(nil, nil, mCart, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 201, r.Code)
	var resp models.CartItem
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.ProductID)
	require.Equal(t, 2, resp.Quantity)
}

func TestMarketController_AddToCart_InvalidQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"product_id":1,"quantity":0}` // quantity must be > 0
	c.Request = httptest.NewRequest("POST", "/api/cart", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)

	mCart := &mockCartRepoFull{
		addFn:    noopAdd,
		getFn:    noopGet,
		updateFn: noopUpdate,
		deleteFn: noopDelete,
		clearFn:  noopClear,
	}

	mc := NewMarketController(nil, nil, mCart, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 400, r.Code)
}

func TestMarketController_AddToCart_MissingProductID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"quantity":2}` // missing product_id
	c.Request = httptest.NewRequest("POST", "/api/cart", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)

	mCart := &mockCartRepoFull{
		addFn:    noopAdd,
		getFn:    noopGet,
		updateFn: noopUpdate,
		deleteFn: noopDelete,
		clearFn:  noopClear,
	}

	mc := NewMarketController(nil, nil, mCart, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 400, r.Code)
}
