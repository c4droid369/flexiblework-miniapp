package model

// MenuType discriminates a row's purpose. The frontend renders different
// widgets per type; the backend enforces permission checks only on Button.
type MenuType int8

const (
	MenuTypeDirectory MenuType = 1 // container; no page of its own
	MenuTypeMenu      MenuType = 2 // routable page
	MenuTypeButton    MenuType = 3 // API-level permission; not rendered as route
)

// Menu is a node in the permission tree. PermCode is populated only for
// Button rows (e.g., "user:create") and matched against handler annotations.
type Menu struct {
	Base
	ParentID  uint64     `gorm:"default:0;index"   json:"parent_id"`
	Type      MenuType   `gorm:"default:2;index"   json:"type"`
	Name      string     `gorm:"size:64;not null"  json:"name"`
	Title     string     `gorm:"size:64"           json:"title"` // display name; may differ from Name
	Path      string     `gorm:"size:255"          json:"path"`  // route path for Menu, ignored for Button
	Component string     `gorm:"size:255"          json:"component"`
	PermCode  string     `gorm:"size:64;index"     json:"perm_code"`
	Icon      string     `gorm:"size:64"           json:"icon"`
	Sort      int        `gorm:"default:0"         json:"sort"`
	Visible   bool       `gorm:"default:true"      json:"visible"`
	Status    RoleStatus `gorm:"default:1;index" json:"status"`
	Children  []Menu     `gorm:"-"                 json:"children,omitempty"`
}

func (Menu) TableName() string { return "menus" }
