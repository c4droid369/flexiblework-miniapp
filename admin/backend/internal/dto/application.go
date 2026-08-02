package dto

import "time"

// CreateApplicationReq is the body of POST /jobs/:id/apply.
type CreateApplicationReq struct {
	Message      string `json:"message"       binding:"max=500"`
	ContactPhone string `json:"contact_phone" binding:"required,max=32"`
}

// AuditApplicationReq is the body of POST /employer/applications/:id/audit.
// Action: 2=通过 3=拒绝.
type AuditApplicationReq struct {
	Action int8   `json:"action" binding:"required,oneof=2 3"`
	Remark string `json:"remark" binding:"max=255"`
}

// ApplicationResp is the wire shape for an application. JobTitle and
// EmployerName are denormalized for list rendering.
type ApplicationResp struct {
	ID           uint64     `json:"id"`
	JobID        uint64     `json:"job_id"`
	JobTitle     string     `json:"job_title"`
	StudentID    uint64     `json:"student_id"`
	StudentName  string     `json:"student_name"`
	StudentPhone string     `json:"student_phone"` // 仅雇主视角可见
	StudentSchool string    `json:"student_school"`
	Message      string     `json:"message"`
	ContactPhone string     `json:"contact_phone"`
	Status       int8       `json:"status"`
	AuditRemark  string     `json:"audit_remark"`
	AuditedAt    *time.Time `json:"audited_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
