package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategory_JSONSerialization(t *testing.T) {
	now := time.Now()
	category := Category{
		ID:          1,
		Name:        "Electronics",
		Description: "Electronic devices",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(category)
	require.NoError(t, err)

	var decoded Category
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, category.ID, decoded.ID)
	assert.Equal(t, category.Name, decoded.Name)
	assert.Equal(t, category.Description, decoded.Description)
}

func TestCreateCategoryRequest_Fields(t *testing.T) {
	req := CreateCategoryRequest{
		Name:        "Clothing",
		Description: "All types of clothing",
	}

	assert.Equal(t, "Clothing", req.Name)
	assert.Equal(t, "All types of clothing", req.Description)
}

func TestUpdateCategoryRequest_Fields(t *testing.T) {
	req := UpdateCategoryRequest{
		Name:        "Updated Category",
		Description: "Updated description",
	}

	assert.Equal(t, "Updated Category", req.Name)
	assert.Equal(t, "Updated description", req.Description)
}

func TestCreateCategoryRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{"name":"Sports","description":"Sports equipment"}`

	var req CreateCategoryRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "Sports", req.Name)
	assert.Equal(t, "Sports equipment", req.Description)
}

func TestUpdateCategoryRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{"name":"Updated Name"}`

	var req UpdateCategoryRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", req.Name)
	assert.Empty(t, req.Description)
}

func TestCategory_EmptyDescription(t *testing.T) {
	category := Category{
		ID:   1,
		Name: "No Description Category",
	}

	assert.Equal(t, 1, category.ID)
	assert.Equal(t, "No Description Category", category.Name)
	assert.Empty(t, category.Description)
}

func TestUpdateCategoryRequest_PartialUpdate(t *testing.T) {
	req := UpdateCategoryRequest{
		Name: "Only Name Updated",
	}

	assert.NotEmpty(t, req.Name)
	assert.Empty(t, req.Description)
}

func TestCategory_Timestamps(t *testing.T) {
	now := time.Now()
	category := Category{
		ID:        1,
		Name:      "Test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.False(t, category.CreatedAt.IsZero())
	assert.False(t, category.UpdatedAt.IsZero())
	assert.Equal(t, category.CreatedAt, category.UpdatedAt)
}
