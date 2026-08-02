package model

// Review is a post-completion rating between a student and an employer for
// a settled order. Each order can produce up to two reviews (one per side);
// Role disambiguates the direction. Tags is a JSON array of short labels
// (e.g. ["准时","态度好"]) chosen from a fixed vocabulary.
type Review struct {
	Base
	OrderID    uint64 `gorm:"index;not null" json:"order_id"`
	FromUserID uint64 `gorm:"index;not null" json:"from_user_id"`
	ToUserID   uint64 `gorm:"index;not null" json:"to_user_id"`
	Role       int8   `gorm:"not null"       json:"role"`   // 1=学生对雇主 2=雇主对学生
	Rating     int    `gorm:"not null"       json:"rating"` // 1-5
	Content    string `gorm:"type:text"      json:"content"`
	Tags       string `gorm:"type:json"      json:"tags"`
}

func (Review) TableName() string { return "reviews" }
