package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
)

// mockCartRepoFull extends CartRepo for cart controller tests
type mockCartRepoFull struct {
	addFn    func(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error)
	getFn    func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error)
	updateFn func(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error)
	deleteFn func(ctx context.Context, itemID, userID int) error
	clearFn  func(ctx context.Context, userID int) error
}

func (m *mockCartRepoFull) AddItem(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
	return m.addFn(ctx, userID, req)
}
func (m *mockCartRepoFull) GetUserCart(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
	return m.getFn(ctx, userID)
}
func (m *mockCartRepoFull) UpdateItem(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	return m.updateFn(ctx, itemID, userID, req)
}
func (m *mockCartRepoFull) DeleteItem(ctx context.Context, itemID, userID int) error {
	return m.deleteFn(ctx, itemID, userID)
}
func (m *mockCartRepoFull) ClearCart(ctx context.Context, userID int) error {
	return m.clearFn(ctx, userID)
}

var _ repository.CartRepo = (*mockCartRepoFull)(nil)

func TestMarketController_GetCart_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	c.Request = httptest.NewRequest("GET", "/api/cart", nil)
	c.Set("user_id", 42)

	now := time.Now()
	items := []*models.CartItemWithDetails{{
		CartItem:     models.CartItem{ID: 1, UserID: 42, ProductID: 7, Quantity: 3, Size: "L", CreatedAt: now, UpdatedAt: now},
		ProductTitle: "Sneakers",
		ProductPrice: 99.90,
		ProductImage: "http://img/1.jpg",
	}}

	mrepo := &mockCartRepoFull{getFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
		require.Equal(t, 42, userID)
		return items, nil
	}, addFn: noopAdd, updateFn: noopUpdate, deleteFn: noopDelete, clearFn: noopClear}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.GetCart(c)

	require.Equal(t, 200, r.Code)
	require.Contains(t, r.Body.String(), "Sneakers")
}

func TestMarketController_UpdateCartItem_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"quantity":5,"size":"XL"}`
	c.Request = httptest.NewRequest("PUT", "/api/cart/items/15", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)
	c.Params = gin.Params{{Key: "id", Value: "15"}}

	updated := &models.CartItem{ID: 15, UserID: 42, ProductID: 7, Quantity: 5, Size: "XL", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mrepo := &mockCartRepoFull{updateFn: func(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
		require.Equal(t, 15, itemID)
		require.Equal(t, 42, userID)
		require.Equal(t, 5, req.Quantity)
		return updated, nil
	}, addFn: noopAdd, getFn: noopGet, deleteFn: noopDelete, clearFn: noopClear}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.UpdateCartItem(c)

	require.Equal(t, 200, r.Code)
	require.Contains(t, r.Body.String(), "\"id\":15")
}

func TestMarketController_UpdateCartItem_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest("PUT", "/api/cart/items/notnum", strings.NewReader(`{"quantity":1}`))
	c.Set("user_id", 1)
	mrepo := &mockCartRepoFull{updateFn: noopUpdate, addFn: noopAdd, getFn: noopGet, deleteFn: noopDelete, clearFn: noopClear}
	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.UpdateCartItem(c)
	require.Equal(t, 400, r.Code)
}

func TestMarketController_DeleteCartItem_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest("DELETE", "/api/cart/items/9", nil)
	c.Set("user_id", 99)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	called := false
	mrepo := &mockCartRepoFull{deleteFn: func(ctx context.Context, itemID, userID int) error {
		called = true
		require.Equal(t, 9, itemID)
		require.Equal(t, 99, userID)
		return nil
	}, addFn: noopAdd, getFn: noopGet, updateFn: noopUpdate, clearFn: noopClear}
	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.DeleteCartItem(c)
	require.True(t, called)
	require.Equal(t, 200, r.Code)
	require.Contains(t, r.Body.String(), "removed")
}

func TestMarketController_GetCart_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest("GET", "/api/cart", nil)
	c.Set("user_id", 7)
	mrepo := &mockCartRepoFull{getFn: func(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) {
		return []*models.CartItemWithDetails{}, nil
	}, addFn: noopAdd, updateFn: noopUpdate, deleteFn: noopDelete, clearFn: noopClear}
	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.GetCart(c)
	require.Equal(t, 200, r.Code)
	// Should be an empty JSON array
	var arr []interface{}
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &arr))
	require.Len(t, arr, 0)
}

// --- helpers ---
func noopAdd(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
	return nil, nil
}
func noopGet(ctx context.Context, userID int) ([]*models.CartItemWithDetails, error) { return nil, nil }
func noopUpdate(ctx context.Context, itemID, userID int, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	return nil, nil
}
func noopDelete(ctx context.Context, itemID, userID int) error { return nil }
func noopClear(ctx context.Context, userID int) error          { return nil }

// Prevent unused warnings when converting string id to int for negative test paths
func _atoi(s string) int { i, _ := strconv.Atoi(s); return i }

// --- AddToCart Unit Tests ---

func TestMarketController_AddToCart_Unit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"product_id":5,"quantity":2,"size":"M"}`
	c.Request = httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)

	now := time.Now()
	mrepo := &mockCartRepoFull{
		addFn: func(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
			require.Equal(t, 42, userID)
			require.Equal(t, 5, req.ProductID)
			require.Equal(t, 2, req.Quantity)
			require.Equal(t, "M", req.Size)
			return &models.CartItem{
				ID:        1,
				UserID:    userID,
				ProductID: req.ProductID,
				Quantity:  req.Quantity,
				Size:      req.Size,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		getFn:    noopGet,
		updateFn: noopUpdate,
		deleteFn: noopDelete,
		clearFn:  noopClear,
	}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 201, r.Code)
	require.Contains(t, r.Body.String(), `"product_id":5`)
	require.Contains(t, r.Body.String(), `"quantity":2`)
}

func TestMarketController_AddToCart_Unit_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"product_id":"not_a_number"}` // invalid
	c.Request = httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)

	mrepo := &mockCartRepoFull{
		addFn:    noopAdd,
		getFn:    noopGet,
		updateFn: noopUpdate,
		deleteFn: noopDelete,
		clearFn:  noopClear,
	}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 400, r.Code)
}

func TestMarketController_AddToCart_Unit_ZeroQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"product_id":5,"quantity":0}` // zero quantity
	c.Request = httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)

	mrepo := &mockCartRepoFull{
		addFn:    noopAdd,
		getFn:    noopGet,
		updateFn: noopUpdate,
		deleteFn: noopDelete,
		clearFn:  noopClear,
	}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.AddToCart(c)

	// Should fail validation for zero quantity
	require.True(t, r.Code >= 400)
}

func TestMarketController_AddToCart_Unit_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	body := `{"product_id":5,"quantity":1}`
	c.Request = httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 42)

	mrepo := &mockCartRepoFull{
		addFn: func(ctx context.Context, userID int, req *models.AddToCartRequest) (*models.CartItem, error) {
			return nil, errors.New("database error")
		},
		getFn:    noopGet,
		updateFn: noopUpdate,
		deleteFn: noopDelete,
		clearFn:  noopClear,
	}

	mc := NewMarketController(nil, nil, mrepo, nil, nil)
	mc.AddToCart(c)

	require.Equal(t, 500, r.Code)
}
