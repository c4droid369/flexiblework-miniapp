package dto

// CreateRoleReq is the body of POST /system/roles.
type CreateRoleReq struct {
	Code        string   `json:"code"        binding:"required,min=1,max=64"`
	Name        string   `json:"name"        binding:"required,min=1,max=64"`
	Description string   `json:"description" binding:"max=255"`
	Sort        int      `json:"sort"`
	Status      int8     `json:"status"      binding:"omitempty,oneof=1 2"`
	MenuIDs     []uint64 `json:"menu_ids"`
}

// UpdateRoleReq is the body of PUT /system/roles/:id.
type UpdateRoleReq struct {
	Name        *string  `json:"name"        binding:"omitempty,max=64"`
	Description *string  `json:"description" binding:"max=255"`
	Sort        *int     `json:"sort"`
	Status      *int8    `json:"status"      binding:"omitempty,oneof=1 2"`
	MenuIDs     []uint64 `json:"menu_ids"`
}

// RoleResp is the wire shape for a single role.
type RoleResp struct {
	ID          uint64   `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Sort        int      `json:"sort"`
	Status      int8     `json:"status"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	MenuIDs     []uint64 `json:"menu_ids"`
}

// AssignMenusReq is the body of POST /system/roles/:id/menus.
type AssignMenusReq struct {
	MenuIDs []uint64 `json:"menu_ids" binding:"required"`
}

// AssignUsersReq is the body of POST /system/roles/:id/users.
type AssignUsersReq struct {
	UserIDs []uint64 `json:"user_ids" binding:"required"`
}
