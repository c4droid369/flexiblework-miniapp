// Package dto defines wire-format request and response shapes. They are
// decoupled from model.* so internal columns (password hash, etc.) never
// leak to the client.
package dto

import "time"

// LoginReq is the body of POST /auth/login.
type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=64" example:"admin"`
	Password string `json:"password" binding:"required,min=6,max=128" example:"admin123"`
}

// LoginResp is what /auth/login and /auth/refresh return.
type LoginResp struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`  // seconds
	TokenType    string   `json:"token_type"`  // always "Bearer"
	Permissions  []string `json:"permissions"` // perm_codes the user has
}

// RefreshReq is the body of POST /auth/refresh.
type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutReq is the body of POST /auth/logout. Refresh token is optional but
// recommended so the refresh row is also revoked.
type LogoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

// MeResp is what /auth/me returns — the current authenticated user.
// UserType is derived from the user's role codes ("admin" | "student" |
// "employer"); UserTypes lists all of them so a user with both "student"
// and "employer" roles can switch contexts.
type MeResp struct {
	ID          uint64      `json:"id"`
	Username    string      `json:"username"`
	Nickname    string      `json:"nickname"`
	Email       string      `json:"email"`
	Phone       string      `json:"phone"`
	Avatar      string      `json:"avatar"`
	Status      int8        `json:"status"`
	LastLoginAt *string     `json:"last_login_at,omitempty"`
	UserType    string      `json:"user_type"`  // primary user type (for default UI)
	UserTypes   []string    `json:"user_types"` // all types the user qualifies for
	Roles       []RoleBrief `json:"roles"`
	Permissions []string    `json:"permissions"`
	Menus       []MenuTree  `json:"menus"`
}

// RoleBrief is the role block returned in /auth/me.
type RoleBrief struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// MenuTree is a entry in the menu tree returned to the frontend for dynamic
// route registration. Only menus with type Directory or Menu are included;
// Button rows are surfaced via the Permissions slice instead.
type MenuTree struct {
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
	Children  []MenuTree `json:"children,omitempty"`
}

// RegisterReq is the body of POST /auth/register. UserType decides which
// role is assigned and which profile row is created.
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"max=64"`
	Phone    string `json:"phone"    binding:"max=32"`
	Email    string `json:"email"    binding:"omitempty,email,max=128"`
	UserType string `json:"user_type" binding:"required,oneof=student employer"`
}

// CertAuditReq is shared by /admin/student-certifications/:id/audit and
// /admin/employer-certifications/:id/audit. Action: 2=通过 3=拒绝.
type CertAuditReq struct {
	Action int8   `json:"action" binding:"required,oneof=2 3"`
	Remark string `json:"remark" binding:"max=255"`
}

// EmployerCertListItem bundles the employer profile with the user row so the
// admin list page can render in a single request.
type EmployerCertListItem struct {
	UserID             uint64    `json:"user_id"`
	Username           string    `json:"username"`
	Nickname           string    `json:"nickname"`
	CompanyName        string    `json:"company_name"`
	ContactName        string    `json:"contact_name"`
	ContactPhone       string    `json:"contact_phone"`
	BusinessLicenseNo  string    `json:"business_license_no"`
	BusinessLicenseImg string    `json:"business_license_img"`
	CertStatus         int8      `json:"cert_status"`
	CertRemark         string    `json:"cert_remark"`
	CreatedAt          time.Time `json:"created_at"`
}

// StudentCertListItem bundles the student profile for the admin list.
type StudentCertListItem struct {
	UserID      uint64    `json:"user_id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	RealName    string    `json:"real_name"`
	School      string    `json:"school"`
	College     string    `json:"college"`
	Major       string    `json:"major"`
	StudentNo   string    `json:"student_no"`
	IDCardFront string    `json:"id_card_front"`
	IDCardBack  string    `json:"id_card_back"`
	StudentCard string    `json:"student_card"`
	CertStatus  int8      `json:"cert_status"`
	CertRemark  string    `json:"cert_remark"`
	CreatedAt   time.Time `json:"created_at"`
}
