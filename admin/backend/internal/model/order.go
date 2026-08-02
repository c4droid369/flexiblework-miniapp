package model

import (
	"time"
)

// Order is the commercial record for a hired application. It is created when
// the employer approves an Application (Application.Status → 5). One row per
// hire; the OrderNo is the public-facing identifier (e.g. CG202608020001).
//
// Status lifecycle:
//   1 待支付 → 2 已支付 (mock pay) → 3 进行中 (student checkin) → 4 待确认完成 (student complete) → 5 已结算 (employer confirm)
//   Any of 1|2 → 6 已取消 → 7 已退款 (only after paid)
//
// WorkProof is a JSON array of image URLs the student uploads at checkin.
type Order struct {
	Base
	OrderNo       string     `gorm:"size:32;uniqueIndex;not null" json:"order_no"`
	JobID         uint64     `gorm:"index;not null"               json:"job_id"`
	ApplicationID uint64     `gorm:"index;not null"               json:"application_id"`
	EmployerID    uint64     `gorm:"index;not null"               json:"employer_id"`
	StudentID     uint64     `gorm:"index;not null"               json:"student_id"`
	Amount        float64    `gorm:"not null"                     json:"amount"`
	Status        int8       `gorm:"default:1;index"              json:"status"`
	PayMethod     string     `gorm:"size:32"                      json:"pay_method"`
	PaidAt        *time.Time `                                    json:"paid_at"`
	StartedAt     *time.Time `                                    json:"started_at"`
	CompletedAt   *time.Time `                                    json:"completed_at"`
	ConfirmedAt   *time.Time `                                    json:"confirmed_at"`
	SettledAt     *time.Time `                                    json:"settled_at"`
	WorkProof     string     `gorm:"type:json"                    json:"work_proof"` // JSON 字符串数组
	CancelReason  string     `gorm:"size:255"                     json:"cancel_reason"`
}

func (Order) TableName() string { return "orders" }
