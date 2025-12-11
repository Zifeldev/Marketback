package controllers

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
)

// mockCartRepo implements repository.CartRepo for testing
type mockCartRepo struct {
	addFn func(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error)
}

func (m *mockCartRepo) AddItem(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
	return m.addFn(ctx, userID, req)
}
func (m *mockCartRepo) GetUserCart(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
	return nil, nil
}
func (m *mockCartRepo) UpdateItem(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	return nil, nil
}
func (m *mockCartRepo) DeleteItem(ctx context.Context, itemID, userID int) error { return nil }
func (m *mockCartRepo) ClearCart(ctx context.Context, userID int) error          { return nil }

func TestMarketController_AddToCart_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	// prepare request body
	body := `{"product_id":1,"quantity":2,"size":"M"}`
	req := httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	// set user_id in context
	c.Set("user_id", 7)

	expected := &models.CartItem{ID: 10, UserID: 7, ProductID: 1, Quantity: 2, Size: "M", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mrepo := &mockCartRepo{addFn: func(ctx context.Context, userID int, r *models.AddToCartRequest) (*models.CartItem, error) {
		require.Equal(t, 7, userID)
		require.Equal(t, 1, r.ProductID)
		return expected, nil
	}}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)

	mc.AddToCart(c)

	require.Equal(t, 201, recorder.Code)
	// body contains returned item id
	require.Contains(t, recorder.Body.String(), "\"id\":10")
}

func TestMarketController_AddToCart_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest("POST", "/api/cart/items", strings.NewReader("bad json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", 7)

	mrepo := &mockCartRepo{addFn: func(ctx context.Context, userID int, r *models.AddToCartRequest) (*models.CartItem, error) {
		return nil, nil
	}}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 400, recorder.Code)
}

func TestMarketController_AddToCart_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	body := `{"product_id":1,"quantity":2,"size":"M"}`
	req := httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", 7)

	mrepo := &mockCartRepo{addFn: func(ctx context.Context, userID int, r *models.AddToCartRequest) (*models.CartItem, error) {
		return nil, fmt.Errorf("db error")
	}}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 500, recorder.Code)
	// AppError возвращает унифицированное сообщение, не раскрывая внутренние ошибки клиенту
	require.Contains(t, recorder.Body.String(), "INTERNAL_ERROR")
	require.Contains(t, recorder.Body.String(), "failed to add item to cart")
}
