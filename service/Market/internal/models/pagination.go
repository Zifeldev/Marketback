package models


type PaginationParams struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

func (p *PaginationParams) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetLimit()
}

func (p *PaginationParams) GetLimit() int {
	if p.PageSize < 1 {
		return 20 // default page size
	}
	if p.PageSize > 100 {
		return 100 // max page size
	}
	return p.PageSize
}

func NewPaginationMeta(page, pageSize int, totalItems int64) PaginationMeta {
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize > 0 {
		totalPages++
	}

	return PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
