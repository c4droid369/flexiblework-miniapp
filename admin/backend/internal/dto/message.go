package dto

import "time"

// MessageResp is the wire shape for a single in-app message.
type MessageResp struct {
	ID        uint64     `json:"id"`
	Type      int8       `json:"type"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Link      string     `json:"link"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// BroadcastMessageReq is the body of POST /admin/messages/broadcast. The
// service fans this out to one row per user. Use sparingly — a single fan-out
// for a few hundred students is fine; a campaign to thousands should chunk.
type BroadcastMessageReq struct {
	Title   string `json:"title"   binding:"required,min=1,max=128"`
	Content string `json:"content" binding:"required,min=1"`
	Link    string `json:"link"    binding:"max=255"`
	Type    int8   `json:"type"    binding:"omitempty,oneof=1 2 3 4"`
	// Audience filters — all optional. Empty audience = send to all users.
	UserType  string `json:"user_type"  binding:"omitempty,oneof=admin student employer all"`
}
