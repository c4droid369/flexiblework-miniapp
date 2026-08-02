package model

import (
	"time"

	"gorm.io/gorm"
)

// UserStatus represents the lifecycle state of a user account.
type UserStatus int8

const (
	UserStatusActive   UserStatus = 1
	UserStatusDisabled UserStatus = 2
)

// User is the authentication principal. PasswordHash is never returned via API
// — the corresponding DTO strips it.
type User struct {
	Base
	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"size:128;not null"           json:"-"`
	Nickname     string     `gorm:"size:64"                    json:"nickname"`
	Email        string     `gorm:"size:128;index"             json:"email"`
	Phone        string     `gorm:"size:32"                    json:"phone"`
	Avatar       string     `gorm:"size:255"                   json:"avatar"`
	Status       UserStatus `gorm:"default:1;index"            json:"status"`
	LastLoginAt  *time.Time `                                  json:"last_login_at"`
	LastLoginIP  string     `gorm:"size:64"                    json:"last_login_ip"`
	Remark       string     `gorm:"size:255"                   json:"remark"`
	WxOpenid     string     `gorm:"size:64;index"              json:"wx_openid,omitempty"` // 预留,WeChat 登录
	Roles        []Role     `gorm:"many2many:user_roles;"      json:"roles,omitempty"`
}

// TableName overrides GORM's pluralization.
func (User) TableName() string { return "users" }

// IsActive is a positive-form predicate; UI code asks `if user.IsActive()`
// instead of `if user.Status != UserStatusDisabled`.
func (u *User) IsActive() bool { return u.Status == UserStatusActive }

// BeforeCreate ensures the Base.DeletedAt zero-value is preserved as NULL.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.Status == 0 {
		u.Status = UserStatusActive
	}
	return nil
}
