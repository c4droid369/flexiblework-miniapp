package model

import (
	"time"
)

// EmployerProfile extends the user record with employer/merchant identity, biz
// license info, certification state, and denormalized reputation counters
// (rating, total_jobs, completed_orders) for cheap list rendering.
type EmployerProfile struct {
	Base
	UserID             uint64     `gorm:"uniqueIndex;not null" json:"user_id"`
	CompanyName        string     `gorm:"size:128"            json:"company_name"`
	ContactName        string     `gorm:"size:64"             json:"contact_name"`
	ContactPhone       string     `gorm:"size:32"             json:"contact_phone"`
	ContactEmail       string     `gorm:"size:128"            json:"contact_email"`
	BusinessLicenseNo  string     `gorm:"size:64"             json:"business_license_no"`
	BusinessLicenseImg string     `gorm:"size:255"            json:"business_license_img"`
	Industry           string     `gorm:"size:64"             json:"industry"`
	CompanySize        string     `gorm:"size:32"             json:"company_size"`
	CompanyAddress     string     `gorm:"size:255"            json:"company_address"`
	Intro              string     `gorm:"size:500"            json:"intro"`
	CertStatus         int8       `gorm:"default:0;index"     json:"cert_status"` // 0=未认证 1=审核中 2=已通过 3=已拒绝
	CertRemark         string     `gorm:"size:255"            json:"cert_remark"`
	CertifiedAt        *time.Time `                           json:"certified_at"`
	Rating             float64    `gorm:"default:5.0"         json:"rating"`
	TotalJobs          int        `gorm:"default:0"           json:"total_jobs"`
	CompletedOrders    int        `gorm:"default:0"           json:"completed_orders"`
}

// TableName overrides GORM's pluralization.
func (EmployerProfile) TableName() string { return "employer_profiles" }
