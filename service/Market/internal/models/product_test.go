package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSizesJSON_Value(t *testing.T) {
	tests := []struct {
		name     string
		sizes    SizesJSON
		expected string
	}{
		{
			name:     "multiple sizes",
			sizes:    SizesJSON{"S", "M", "L", "XL"},
			expected: `["S","M","L","XL"]`,
		},
		{
			name:     "single size",
			sizes:    SizesJSON{"M"},
			expected: `["M"]`,
		},
		{
			name:     "empty sizes",
			sizes:    SizesJSON{},
			expected: `[]`,
		},
		{
			name:     "nil sizes",
			sizes:    nil,
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.sizes.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(value.([]byte)))
		})
	}
}

func TestSizesJSON_Scan_Bytes(t *testing.T) {
	var sizes SizesJSON
	err := sizes.Scan([]byte(`["S","M","L"]`))
	require.NoError(t, err)
	assert.Equal(t, SizesJSON{"S", "M", "L"}, sizes)
}

func TestSizesJSON_Scan_String(t *testing.T) {
	var sizes SizesJSON
	err := sizes.Scan(`["XS","S"]`)
	require.NoError(t, err)
	assert.Equal(t, SizesJSON{"XS", "S"}, sizes)
}

func TestSizesJSON_Scan_Nil(t *testing.T) {
	var sizes SizesJSON
	err := sizes.Scan(nil)
	require.NoError(t, err)
	assert.Equal(t, SizesJSON{}, sizes)
}

func TestSizesJSON_Scan_EmptyArray(t *testing.T) {
	var sizes SizesJSON
	err := sizes.Scan([]byte(`[]`))
	require.NoError(t, err)
	assert.Equal(t, SizesJSON{}, sizes)
}

func TestProduct_JSONSerialization(t *testing.T) {
	product := Product{
		ID:          1,
		SellerID:    10,
		CategoryID:  5,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       100,
		Sizes:       SizesJSON{"S", "M", "L"},
		ImageURL:    "http://example.com/image.jpg",
		Status:      "active",
	}

	data, err := json.Marshal(product)
	require.NoError(t, err)

	var decoded Product
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, product.ID, decoded.ID)
	assert.Equal(t, product.Title, decoded.Title)
	assert.Equal(t, product.Price, decoded.Price)
	assert.Equal(t, product.Sizes, decoded.Sizes)
}

func TestCreateProductRequest_Validation(t *testing.T) {
	req := CreateProductRequest{
		CategoryID:  1,
		Title:       "New Product",
		Description: "Product description",
		Price:       50.00,
		Stock:       10,
		Sizes:       SizesJSON{"M", "L"},
		ImageURL:    "http://example.com/img.png",
	}

	assert.Equal(t, 1, req.CategoryID)
	assert.Equal(t, "New Product", req.Title)
	assert.Equal(t, 50.00, req.Price)
	assert.Equal(t, 10, req.Stock)
	assert.Len(t, req.Sizes, 2)
}

func TestUpdateProductRequest_PartialUpdate(t *testing.T) {
	title := "Updated Title"
	price := 75.50
	stock := 200

	req := UpdateProductRequest{
		Title: &title,
		Price: &price,
		Stock: &stock,
	}

	assert.NotNil(t, req.Title)
	assert.Equal(t, "Updated Title", *req.Title)
	assert.NotNil(t, req.Price)
	assert.Equal(t, 75.50, *req.Price)
	assert.NotNil(t, req.Stock)
	assert.Equal(t, 200, *req.Stock)
	assert.Nil(t, req.CategoryID)
	assert.Nil(t, req.Description)
	assert.Nil(t, req.Status)
}

func TestProductWithDetails_Embedding(t *testing.T) {
	product := ProductWithDetails{
		Product: Product{
			ID:       1,
			SellerID: 10,
			Title:    "Test",
			Price:    100.00,
		},
		SellerName:   "Test Shop",
		CategoryName: "Electronics",
	}

	assert.Equal(t, 1, product.ID)
	assert.Equal(t, "Test", product.Title)
	assert.Equal(t, "Test Shop", product.SellerName)
	assert.Equal(t, "Electronics", product.CategoryName)
}
