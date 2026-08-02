package model

import (
	"time"
)

// StudentProfile extends the user record with student-specific identity, school
// info, and certification state. Sensitive fields (id_card_no) are stripped from
// the JSON wire form — DTOs are responsible for surfacing the masked version.
type StudentProfile struct {
	Base
	UserID      uint64     `gorm:"uniqueIndex;not null" json:"user_id"`
	RealName    string     `gorm:"size:64"             json:"real_name"`
	Gender      int8       `gorm:"default:0"           json:"gender"`      // 0=未填 1=男 2=女
	School      string     `gorm:"size:128;index"      json:"school"`
	College     string     `gorm:"size:128"            json:"college"`
	Major       string     `gorm:"size:128"            json:"major"`
	StudentNo   string     `gorm:"size:64"             json:"student_no"`
	IDCardNo    string     `gorm:"size:32"             json:"-"` // 敏感字段,DTO 层脱敏后输出
	IDCardFront string     `gorm:"size:255"            json:"id_card_front"`
	IDCardBack  string     `gorm:"size:255"            json:"id_card_back"`
	StudentCard string     `gorm:"size:255"            json:"student_card"`
	CertStatus  int8       `gorm:"default:0;index"     json:"cert_status"` // 0=未认证 1=审核中 2=已通过 3=已拒绝
	CertRemark  string     `gorm:"size:255"            json:"cert_remark"`
	CertifiedAt *time.Time `                           json:"certified_at"`
	Bio         string     `gorm:"size:500"            json:"bio"`
	Skills      string     `gorm:"type:json"           json:"skills"` // JSON 数组,字符串标签
}

// TableName overrides GORM's pluralization.
func (StudentProfile) TableName() string { return "student_profiles" }
