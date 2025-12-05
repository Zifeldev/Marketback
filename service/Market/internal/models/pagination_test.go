package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationParams_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		params   PaginationParams
		expected int
	}{
		{
			name:     "first page",
			params:   PaginationParams{Page: 1, PageSize: 10},
			expected: 0,
		},
		{
			name:     "second page",
			params:   PaginationParams{Page: 2, PageSize: 10},
			expected: 10,
		},
		{
			name:     "third page with 20 items",
			params:   PaginationParams{Page: 3, PageSize: 20},
			expected: 40,
		},
		{
			name:     "zero page defaults to first",
			params:   PaginationParams{Page: 0, PageSize: 10},
			expected: 0,
		},
		{
			name:     "negative page defaults to first",
			params:   PaginationParams{Page: -1, PageSize: 10},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.GetOffset()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationParams_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		params   PaginationParams
		expected int
	}{
		{
			name:     "normal page size",
			params:   PaginationParams{Page: 1, PageSize: 25},
			expected: 25,
		},
		{
			name:     "zero page size defaults to 20",
			params:   PaginationParams{Page: 1, PageSize: 0},
			expected: 20,
		},
		{
			name:     "negative page size defaults to 20",
			params:   PaginationParams{Page: 1, PageSize: -5},
			expected: 20,
		},
		{
			name:     "exceeds max page size",
			params:   PaginationParams{Page: 1, PageSize: 150},
			expected: 100,
		},
		{
			name:     "exactly max page size",
			params:   PaginationParams{Page: 1, PageSize: 100},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.GetLimit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPaginationMeta(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		totalItems int64
		expected   PaginationMeta
	}{
		{
			name:       "exact division",
			page:       1,
			pageSize:   10,
			totalItems: 100,
			expected: PaginationMeta{
				Page:       1,
				PageSize:   10,
				TotalItems: 100,
				TotalPages: 10,
			},
		},
		{
			name:       "with remainder",
			page:       2,
			pageSize:   10,
			totalItems: 25,
			expected: PaginationMeta{
				Page:       2,
				PageSize:   10,
				TotalItems: 25,
				TotalPages: 3,
			},
		},
		{
			name:       "single page",
			page:       1,
			pageSize:   50,
			totalItems: 30,
			expected: PaginationMeta{
				Page:       1,
				PageSize:   50,
				TotalItems: 30,
				TotalPages: 1,
			},
		},
		{
			name:       "empty result",
			page:       1,
			pageSize:   10,
			totalItems: 0,
			expected: PaginationMeta{
				Page:       1,
				PageSize:   10,
				TotalItems: 0,
				TotalPages: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPaginationMeta(tt.page, tt.pageSize, tt.totalItems)
			assert.Equal(t, tt.expected, result)
		})
	}
}
