package dto

import "time"

// CreateOrderReq is the body of POST /employer/applications/:id/hire — it
// creates an order from an approved application. The price is set here.
type CreateOrderReq struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// PayOrderReq is the body of POST /orders/:id/pay. Method defaults to
// "mock_wechat" in the service when empty.
type PayOrderReq struct {
	Method string `json:"method" binding:"omitempty,max=32"`
}

// CheckinOrderReq is the body of POST /orders/:id/checkin.
type CheckinOrderReq struct {
	WorkProof []string `json:"work_proof" binding:"required,min=1,max=9,dive,max=255"`
}

// OrderResp is the wire shape for an order. JobTitle / StudentName /
// EmployerName are denormalized.
type OrderResp struct {
	ID            uint64     `json:"id"`
	OrderNo       string     `json:"order_no"`
	JobID         uint64     `json:"job_id"`
	JobTitle      string     `json:"job_title"`
	ApplicationID uint64     `json:"application_id"`
	EmployerID    uint64     `json:"employer_id"`
	EmployerName  string     `json:"employer_name"`
	StudentID     uint64     `json:"student_id"`
	StudentName   string     `json:"student_name"`
	Amount        float64    `json:"amount"`
	Status        int8       `json:"status"`
	PayMethod     string     `json:"pay_method"`
	PaidAt        *time.Time `json:"paid_at"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	ConfirmedAt   *time.Time `json:"confirmed_at"`
	SettledAt     *time.Time `json:"settled_at"`
	WorkProof     []string   `json:"work_proof"`
	CancelReason  string     `json:"cancel_reason"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CancelOrderReq is the body of POST /orders/:id/cancel (student) or
// /employer/orders/:id/cancel (employer). Reason is required.
type CancelOrderReq struct {
	Reason string `json:"reason" binding:"required,min=2,max=255"`
}
