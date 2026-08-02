package model

import "time"

// OperationLog records one handler invocation. The middleware in
// internal/api/middleware writes a row after the handler returns, including
// status code and latency. Heavy fields (request_body) are truncated.
type OperationLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint64    `gorm:"index"                    json:"user_id"`
	Username       string    `gorm:"size:64;index"            json:"username"`
	Action         string    `gorm:"size:128;index"           json:"action"`
	Method         string    `gorm:"size:16"                  json:"method"`
	Path           string    `gorm:"size:255;index"           json:"path"`
	IP             string    `gorm:"size:64"                  json:"ip"`
	UserAgent      string    `gorm:"size:512"                 json:"user_agent"`
	RequestBody    string    `gorm:"type:text"                json:"request_body"`
	ResponseStatus int       `                               json:"response_status"`
	LatencyMS      int64     `                               json:"latency_ms"`
	ErrorMessage   string    `gorm:"size:1024"                json:"error_message"`
	CreatedAt      time.Time `gorm:"index"                    json:"created_at"`
}

func (OperationLog) TableName() string { return "operation_logs" }

// UserRole is the m2m join between User and Role. Composite primary key.
type UserRole struct {
	UserID uint64 `gorm:"primaryKey;autoIncrement:false"`
	RoleID uint64 `gorm:"primaryKey;autoIncrement:false;index"`
}

func (UserRole) TableName() string { return "user_roles" }

// RoleMenu is the m2m join between Role and Menu.
type RoleMenu struct {
	RoleID uint64 `gorm:"primaryKey;autoIncrement:false"`
	MenuID uint64 `gorm:"primaryKey;autoIncrement:false;index"`
}

func (RoleMenu) TableName() string { return "role_menus" }
