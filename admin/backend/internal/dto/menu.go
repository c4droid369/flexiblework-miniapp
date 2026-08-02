package dto

// CreateMenuReq is the body of POST /system/menus.
type CreateMenuReq struct {
	ParentID  uint64 `json:"parent_id"`
	Type      int8   `json:"type"      binding:"required,oneof=1 2 3"`
	Name      string `json:"name"      binding:"required,max=64"`
	Title     string `json:"title"     binding:"max=64"`
	Path      string `json:"path"      binding:"max=255"`
	Component string `json:"component" binding:"max=255"`
	PermCode  string `json:"perm_code" binding:"max=64"`
	Icon      string `json:"icon"      binding:"max=64"`
	Sort      int    `json:"sort"`
	Visible   bool   `json:"visible"`
}

// UpdateMenuReq is the body of PUT /system/menus/:id.
type UpdateMenuReq struct {
	ParentID  *uint64 `json:"parent_id"`
	Type      *int8   `json:"type"      binding:"omitempty,oneof=1 2 3"`
	Name      *string `json:"name"      binding:"omitempty,max=64"`
	Title     *string `json:"title"     binding:"max=64"`
	Path      *string `json:"path"      binding:"max=255"`
	Component *string `json:"component" binding:"max=255"`
	PermCode  *string `json:"perm_code" binding:"max=64"`
	Icon      *string `json:"icon"      binding:"max=64"`
	Sort      *int    `json:"sort"`
	Visible   *bool   `json:"visible"`
	Status    *int8   `json:"status"    binding:"omitempty,oneof=1 2"`
}

// MenuResp is the wire shape for a single menu.
type MenuResp struct {
	ID        uint64     `json:"id"`
	ParentID  uint64     `json:"parent_id"`
	Type      int8       `json:"type"`
	Name      string     `json:"name"`
	Title     string     `json:"title"`
	Path      string     `json:"path"`
	Component string     `json:"component"`
	PermCode  string     `json:"perm_code"`
	Icon      string     `json:"icon"`
	Sort      int        `json:"sort"`
	Visible   bool       `json:"visible"`
	Status    int8       `json:"status"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Children  []MenuTree `json:"children,omitempty"`
}
