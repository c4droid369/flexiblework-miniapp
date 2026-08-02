package dto

import "time"

// CreateCategoryReq is the body of POST /admin/categories.
type CreateCategoryReq struct {
	Name        string `json:"name"        binding:"required,min=1,max=64"`
	Icon        string `json:"icon"        binding:"max=255"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"      binding:"omitempty,oneof=1 2"`
	Description string `json:"description" binding:"max=255"`
}

// UpdateCategoryReq is the body of PUT /admin/categories/:id.
type UpdateCategoryReq struct {
	Name        *string `json:"name"        binding:"omitempty,min=1,max=64"`
	Icon        *string `json:"icon"        binding:"omitempty,max=255"`
	Sort        *int    `json:"sort"`
	Status      *int8   `json:"status"      binding:"omitempty,oneof=1 2"`
	Description *string `json:"description" binding:"omitempty,max=255"`
}

// CategoryResp is the wire shape for a single category.
type CategoryResp struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
