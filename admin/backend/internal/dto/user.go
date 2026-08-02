package dto

// CreateUserReq is the body of POST /system/users.
type CreateUserReq struct {
	Username string   `json:"username" binding:"required,min=3,max=64"`
	Password string   `json:"password" binding:"required,min=6,max=128"`
	Nickname string   `json:"nickname" binding:"max=64"`
	Email    string   `json:"email"    binding:"omitempty,email,max=128"`
	Phone    string   `json:"phone"    binding:"max=32"`
	Avatar   string   `json:"avatar"   binding:"max=255"`
	Remark   string   `json:"remark"   binding:"max=255"`
	RoleIDs  []uint64 `json:"role_ids"`
}

// UpdateUserReq is the body of PUT /system/users/:id. Password changes go via
// the dedicated reset-password endpoint so the audit log is unambiguous.
type UpdateUserReq struct {
	Nickname *string  `json:"nickname" binding:"omitempty,max=64"`
	Email    *string  `json:"email"    binding:"omitempty,email,max=128"`
	Phone    *string  `json:"phone"    binding:"max=32"`
	Avatar   *string  `json:"avatar"   binding:"max=255"`
	Remark   *string  `json:"remark"   binding:"max=255"`
	Status   *int8    `json:"status"   binding:"omitempty,oneof=1 2"`
	RoleIDs  []uint64 `json:"role_ids"`
}

// UserResp is the wire shape returned for a single user. Never carries the
// password hash — the JSON tag `json:"-"` on the model field enforces it.
type UserResp struct {
	ID          uint64      `json:"id"`
	Username    string      `json:"username"`
	Nickname    string      `json:"nickname"`
	Email       string      `json:"email"`
	Phone       string      `json:"phone"`
	Avatar      string      `json:"avatar"`
	Status      int8        `json:"status"`
	LastLoginAt string      `json:"last_login_at,omitempty"`
	LastLoginIP string      `json:"last_login_ip"`
	Remark      string      `json:"remark"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	Roles       []RoleBrief `json:"roles"`
}

// ResetPasswordReq is the body of POST /system/users/:id/reset-password.
type ResetPasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
}

// ChangeStatusReq is the body of POST /system/users/:id/status.
type ChangeStatusReq struct {
	Status int8 `json:"status" binding:"required,oneof=1 2"`
}
