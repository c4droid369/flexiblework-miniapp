package model

// RoleStatus represents the lifecycle state of a role.
type RoleStatus int8

const (
	RoleStatusActive   RoleStatus = 1
	RoleStatusDisabled RoleStatus = 2
)

// Role is a named permission bundle. Codes are referenced by application
// logic (e.g., "super_admin" bypasses permission checks). Menu associations
// are stored in role_menus.
type Role struct {
	Base
	Code        string     `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name        string     `gorm:"size:64;not null"            json:"name"`
	Description string     `gorm:"size:255"                    json:"description"`
	Sort        int        `gorm:"default:0"                   json:"sort"`
	Status      RoleStatus `gorm:"default:1;index"        json:"status"`
	Menus       []Menu     `gorm:"many2many:role_menus;"       json:"menus,omitempty"`
	Users       []User     `gorm:"many2many:user_roles;"       json:"users,omitempty"`
}

func (Role) TableName() string { return "roles" }

// IsActive mirrors User.IsActive for symmetry.
func (r *Role) IsActive() bool { return r.Status == RoleStatusActive }
