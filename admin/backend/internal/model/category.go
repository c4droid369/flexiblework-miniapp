package model

// Category is the gig taxonomy (e.g. 促销派单 / 家教 / 餐饮服务). Rendered on
// the student's home page as a horizontal scroll and on the employer publish
// form as a picker. Status is 1=启用 2=禁用.
type Category struct {
	Base
	Name        string `gorm:"size:64;not null;index" json:"name"`
	Icon        string `gorm:"size:255"               json:"icon"`
	Sort        int    `gorm:"default:0"              json:"sort"`
	Status      int8   `gorm:"default:1;index"        json:"status"`
	Description string `gorm:"size:255"               json:"description"`
}

func (Category) TableName() string { return "categories" }
