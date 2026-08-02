package dto

import "time"

// CreateReviewReq is the body of POST /orders/:id/review (student→employer)
// or /employer/orders/:id/review (employer→student). The role and toUser are
// derived server-side from the authenticated user + the order, never the
// body, to prevent spoofed reviews.
type CreateReviewReq struct {
	Rating  int      `json:"rating"  binding:"required,min=1,max=5"`
	Content string   `json:"content" binding:"max=1000"`
	Tags    []string `json:"tags"    binding:"omitempty,max=10,dive,max=20"`
}

// ReviewResp is the wire shape for a review. FromName is denormalized.
type ReviewResp struct {
	ID         uint64    `json:"id"`
	OrderID    uint64    `json:"order_id"`
	FromUserID uint64    `json:"from_user_id"`
	FromName   string    `json:"from_name"`
	FromAvatar string    `json:"from_avatar"`
	ToUserID   uint64    `json:"to_user_id"`
	Role       int8      `json:"role"`
	Rating     int       `json:"rating"`
	Content    string    `json:"content"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
}
