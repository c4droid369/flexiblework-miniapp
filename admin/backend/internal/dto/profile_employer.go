package dto

import "time"

// EmployerProfileResp is the wire shape for an employer profile.
type EmployerProfileResp struct {
	ID                uint64     `json:"id"`
	UserID            uint64     `json:"user_id"`
	CompanyName       string     `json:"company_name"`
	ContactName       string     `json:"contact_name"`
	ContactPhone      string     `json:"contact_phone"`
	ContactEmail      string     `json:"contact_email"`
	BusinessLicenseNo string     `json:"business_license_no"`
	BusinessLicenseImg string    `json:"business_license_img"`
	Industry          string     `json:"industry"`
	CompanySize       string     `json:"company_size"`
	CompanyAddress    string     `json:"company_address"`
	Intro             string     `json:"intro"`
	CertStatus        int8       `json:"cert_status"`
	CertRemark        string     `json:"cert_remark"`
	CertifiedAt       *time.Time `json:"certified_at"`
	Rating            float64    `json:"rating"`
	TotalJobs         int        `json:"total_jobs"`
	CompletedOrders   int        `json:"completed_orders"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UpsertEmployerProfileReq is the body for POST /employer/profile.
type UpsertEmployerProfileReq struct {
	CompanyName        *string `json:"company_name"         binding:"omitempty,max=128"`
	ContactName        *string `json:"contact_name"         binding:"omitempty,max=64"`
	ContactPhone       *string `json:"contact_phone"        binding:"omitempty,max=32"`
	ContactEmail       *string `json:"contact_email"        binding:"omitempty,email,max=128"`
	BusinessLicenseNo  *string `json:"business_license_no"  binding:"omitempty,max=64"`
	BusinessLicenseImg *string `json:"business_license_img" binding:"omitempty,max=255"`
	Industry           *string `json:"industry"             binding:"omitempty,max=64"`
	CompanySize        *string `json:"company_size"         binding:"omitempty,max=32"`
	CompanyAddress     *string `json:"company_address"      binding:"omitempty,max=255"`
	Intro              *string `json:"intro"                binding:"omitempty,max=500"`
}

// SubmitEmployerCertificationReq is the body for POST /employer/certification.
type SubmitEmployerCertificationReq struct {
	CompanyName        string `json:"company_name"         binding:"required,max=128"`
	BusinessLicenseNo  string `json:"business_license_no"  binding:"required,max=64"`
	BusinessLicenseImg string `json:"business_license_img" binding:"required,max=255"`
	ContactName        string `json:"contact_name"         binding:"required,max=64"`
	ContactPhone       string `json:"contact_phone"        binding:"required,max=32"`
}
