package model

import (
	"time"
)

// Application is a student's application to a job. One application per
// (job_id, student_id) is enforced by a composite unique index created via
// raw migration in the repository init (see migration.go).
//
// Status lifecycle:
//   1 待审核 → 2 已通过 (employer approves) | 3 已拒绝 (employer rejects) | 4 已取消 (student cancels)
//   2 已通过 → 5 已转订单 (employer hires → an Order row is created)
type Application struct {
	Base
	JobID        uint64     `gorm:"index;not null"  json:"job_id"`
	StudentID    uint64     `gorm:"index;not null"  json:"student_id"`
	Message      string     `gorm:"size:500"        json:"message"`
	ContactPhone string     `gorm:"size:32"         json:"contact_phone"`
	Status       int8       `gorm:"default:1;index" json:"status"`
	AuditRemark  string     `gorm:"size:255"        json:"audit_remark"`
	AuditedAt    *time.Time `                       json:"audited_at"`
}

func (Application) TableName() string { return "applications" }
