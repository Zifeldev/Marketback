package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartItem_JSONSerialization(t *testing.T) {
	now := time.Now()
	item := CartItem{
		ID:        1,
		UserID:    10,
		ProductID: 5,
		Quantity:  3,
		Size:      "L",
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)

	var decoded CartItem
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, item.ID, decoded.ID)
	assert.Equal(t, item.UserID, decoded.UserID)
	assert.Equal(t, item.ProductID, decoded.ProductID)
	assert.Equal(t, item.Quantity, decoded.Quantity)
	assert.Equal(t, item.Size, decoded.Size)
}

func TestCartItemWithDetails_CalculateTotal(t *testing.T) {
	item := CartItemWithDetails{
		CartItem: CartItem{
			ID:        1,
			Quantity:  3,
			ProductID: 10,
		},
		ProductTitle: "Test Product",
		ProductPrice: 25.50,
		ProductImage: "http://example.com/image.jpg",
	}

	total := float64(item.Quantity) * item.ProductPrice
	assert.Equal(t, 76.50, total)
}

func TestAddToCartRequest_Fields(t *testing.T) {
	req := AddToCartRequest{
		ProductID: 15,
		Quantity:  2,
		Size:      "M",
	}

	assert.Equal(t, 15, req.ProductID)
	assert.Equal(t, 2, req.Quantity)
	assert.Equal(t, "M", req.Size)
}

func TestUpdateCartItemRequest_Fields(t *testing.T) {
	req := UpdateCartItemRequest{
		Quantity: 5,
		Size:     "XL",
	}

	assert.Equal(t, 5, req.Quantity)
	assert.Equal(t, "XL", req.Size)
}

func TestCartItemWithDetails_Embedding(t *testing.T) {
	now := time.Now()
	item := CartItemWithDetails{
		CartItem: CartItem{
			ID:        1,
			UserID:    10,
			ProductID: 5,
			Quantity:  2,
			Size:      "S",
			CreatedAt: now,
			UpdatedAt: now,
		},
		ProductTitle: "Shirt",
		ProductPrice: 29.99,
		ProductImage: "http://example.com/shirt.jpg",
	}

	assert.Equal(t, 1, item.ID)
	assert.Equal(t, 10, item.UserID)
	assert.Equal(t, "Shirt", item.ProductTitle)
	assert.Equal(t, 29.99, item.ProductPrice)
}

func TestAddToCartRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{"product_id":10,"quantity":3,"size":"L"}`

	var req AddToCartRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, 10, req.ProductID)
	assert.Equal(t, 3, req.Quantity)
	assert.Equal(t, "L", req.Size)
}

func TestAddToCartRequest_JSONUnmarshal_NoSize(t *testing.T) {
	jsonData := `{"product_id":5,"quantity":1}`

	var req AddToCartRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, 5, req.ProductID)
	assert.Equal(t, 1, req.Quantity)
	assert.Equal(t, "", req.Size)
}
