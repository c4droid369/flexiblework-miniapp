package dto

import "time"

// CreateJobReq is the body of POST /employer/jobs. New jobs start at status 1
// (待审核); the service transitions to 2 (招聘中) once admin approves.
type CreateJobReq struct {
	CategoryID        uint64     `json:"category_id"        binding:"required"`
	Title             string     `json:"title"              binding:"required,min=2,max=128"`
	Cover             string     `json:"cover"              binding:"max=255"`
	Description       string     `json:"description"        binding:"required,min=10"`
	Requirements      string     `json:"requirements"       binding:"max=5000"`
	SalaryType        int8       `json:"salary_type"        binding:"required,oneof=1 2 3 4 5"`
	SalaryMin         float64    `json:"salary_min"         binding:"gte=0"`
	SalaryMax         float64    `json:"salary_max"         binding:"gte=0"`
	SalaryUnit        string     `json:"salary_unit"        binding:"required,max=16"`
	Location          string     `json:"location"           binding:"required,max=255"`
	WorkDateType      int8       `json:"work_date_type"     binding:"omitempty,oneof=1 2 3"`
	WorkDateStart     *time.Time `json:"work_date_start"`
	WorkDateEnd       *time.Time `json:"work_date_end"`
	WorkTimeStart     string     `json:"work_time_start"    binding:"omitempty,max=16"`
	WorkTimeEnd       string     `json:"work_time_end"      binding:"omitempty,max=16"`
	RecruitCount      int        `json:"recruit_count"      binding:"required,min=1,max=999"`
	GenderRequirement int8       `json:"gender_requirement" binding:"omitempty,oneof=0 1 2"`
	SettlementType    int8       `json:"settlement_type"    binding:"omitempty,oneof=1 2 3"`
}

// UpdateJobReq is the body of PUT /employer/jobs/:id. Only the employer who
// owns the job and only when status=1 (待审核) or 2 (招聘中) may update.
type UpdateJobReq struct {
	CategoryID        *uint64    `json:"category_id"`
	Title             *string    `json:"title"              binding:"omitempty,min=2,max=128"`
	Cover             *string    `json:"cover"              binding:"omitempty,max=255"`
	Description       *string    `json:"description"        binding:"omitempty,min=10"`
	Requirements      *string    `json:"requirements"       binding:"omitempty,max=5000"`
	SalaryType        *int8      `json:"salary_type"        binding:"omitempty,oneof=1 2 3 4 5"`
	SalaryMin         *float64   `json:"salary_min"`
	SalaryMax         *float64   `json:"salary_max"`
	SalaryUnit        *string    `json:"salary_unit"        binding:"omitempty,max=16"`
	Location          *string    `json:"location"           binding:"omitempty,max=255"`
	WorkDateType      *int8      `json:"work_date_type"     binding:"omitempty,oneof=1 2 3"`
	WorkDateStart     *time.Time `json:"work_date_start"`
	WorkDateEnd       *time.Time `json:"work_date_end"`
	WorkTimeStart     *string    `json:"work_time_start"    binding:"omitempty,max=16"`
	WorkTimeEnd       *string    `json:"work_time_end"      binding:"omitempty,max=16"`
	RecruitCount      *int       `json:"recruit_count"      binding:"omitempty,min=1,max=999"`
	GenderRequirement *int8      `json:"gender_requirement" binding:"omitempty,oneof=0 1 2"`
	SettlementType    *int8      `json:"settlement_type"    binding:"omitempty,oneof=1 2 3"`
}

// JobResp is the wire shape for a job. It denormalizes employer_name and
// category_name for the listing pages so the frontend avoids a join API.
type JobResp struct {
	ID               uint64     `json:"id"`
	EmployerID       uint64     `json:"employer_id"`
	EmployerName     string     `json:"employer_name"`
	CategoryID       uint64     `json:"category_id"`
	CategoryName     string     `json:"category_name"`
	Title            string     `json:"title"`
	Cover            string     `json:"cover"`
	Description      string     `json:"description"`
	Requirements     string     `json:"requirements"`
	SalaryType       int8       `json:"salary_type"`
	SalaryMin        float64    `json:"salary_min"`
	SalaryMax        float64    `json:"salary_max"`
	SalaryUnit       string     `json:"salary_unit"`
	SalaryText       string     `json:"salary_text"` // 拼好的可读文本 e.g. "25-30元/时"
	Location         string     `json:"location"`
	WorkDateType     int8       `json:"work_date_type"`
	WorkDateStart    *time.Time `json:"work_date_start"`
	WorkDateEnd      *time.Time `json:"work_date_end"`
	WorkTimeStart    string     `json:"work_time_start"`
	WorkTimeEnd      string     `json:"work_time_end"`
	RecruitCount     int        `json:"recruit_count"`
	GenderRequirement int8      `json:"gender_requirement"`
	SettlementType   int8       `json:"settlement_type"`
	Status           int8       `json:"status"`
	AuditRemark      string     `json:"audit_remark"`
	AuditedAt        *time.Time `json:"audited_at"`
	ViewCount        int        `json:"view_count"`
	ApplyCount       int        `json:"apply_count"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// AuditJobReq is the body of POST /admin/jobs/:id/audit. Pass=2, Reject=4.
type AuditJobReq struct {
	Action  int8   `json:"action"  binding:"required,oneof=2 4"` // 2=通过 4=拒绝
	Remark  string `json:"remark"  binding:"max=255"`
}
