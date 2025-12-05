package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeller_JSONSerialization(t *testing.T) {
	now := time.Now()
	seller := Seller{
		ID:          1,
		UserID:      10,
		ShopName:    "Test Shop",
		Description: "Best shop ever",
		Rating:      4.5,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(seller)
	require.NoError(t, err)

	var decoded Seller
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, seller.ID, decoded.ID)
	assert.Equal(t, seller.UserID, decoded.UserID)
	assert.Equal(t, seller.ShopName, decoded.ShopName)
	assert.Equal(t, seller.Rating, decoded.Rating)
	assert.Equal(t, seller.IsActive, decoded.IsActive)
}

func TestCreateSellerRequest_Fields(t *testing.T) {
	req := CreateSellerRequest{
		ShopName:    "My Store",
		Description: "Selling great stuff",
	}

	assert.Equal(t, "My Store", req.ShopName)
	assert.Equal(t, "Selling great stuff", req.Description)
}

func TestUpdateSellerRequest_Fields(t *testing.T) {
	req := UpdateSellerRequest{
		ShopName:    "Updated Store Name",
		Description: "New description",
	}

	assert.Equal(t, "Updated Store Name", req.ShopName)
	assert.Equal(t, "New description", req.Description)
}

func TestCreateSellerRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{"shop_name":"Tech Store","description":"Electronics and more"}`

	var req CreateSellerRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "Tech Store", req.ShopName)
	assert.Equal(t, "Electronics and more", req.Description)
}

func TestSeller_DefaultRating(t *testing.T) {
	seller := Seller{
		ID:       1,
		ShopName: "New Shop",
	}

	assert.Equal(t, 0.0, seller.Rating)
}

func TestSeller_ActiveStatus(t *testing.T) {
	activeSeller := Seller{IsActive: true}
	inactiveSeller := Seller{IsActive: false}

	assert.True(t, activeSeller.IsActive)
	assert.False(t, inactiveSeller.IsActive)
}

func TestUpdateSellerRequest_EmptyUpdate(t *testing.T) {
	req := UpdateSellerRequest{}

	assert.Empty(t, req.ShopName)
	assert.Empty(t, req.Description)
}

func TestSeller_RatingRange(t *testing.T) {
	tests := []struct {
		name   string
		rating float64
		valid  bool
	}{
		{"minimum rating", 0.0, true},
		{"mid rating", 2.5, true},
		{"maximum rating", 5.0, true},
		{"negative rating", -1.0, false},
		{"over maximum", 5.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seller := Seller{Rating: tt.rating}
			isValid := seller.Rating >= 0 && seller.Rating <= 5.0
			assert.Equal(t, tt.valid, isValid)
		})
	}
}
