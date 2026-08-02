package model

import "time"

// StorageType identifies the backing storage backend. Currently only "local";
// future "oss", "s3", "minio" plug in via pkg/storage.
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
)

// File is the metadata for an uploaded blob. The actual bytes live in the
// storage backend; Path is the adapter-side key, NOT a public URL.
//
// Do NOT redeclare CreatedAt / UpdatedAt / DeletedAt here — Base already
// provides them with their `gorm:"index"` tags. Redefining creates a column
// conflict in MySQL ("Duplicate column name 'created_at'").
type File struct {
	Base
	Name         string      `gorm:"size:255;not null" json:"name"`
	OriginalName string      `gorm:"size:255;not null" json:"original_name"`
	Path         string      `gorm:"size:512;not null" json:"path"`
	Size         int64       `                        json:"size"`
	MimeType     string      `gorm:"size:128"          json:"mime_type"`
	Storage      StorageType `gorm:"size:32;index"     json:"storage"`
	UploaderID   uint64      `gorm:"index"             json:"uploader_id"`
}

func (File) TableName() string { return "files" }

// RefreshToken persists active refresh tokens so we can revoke them on
// rotation or logout. Hash is stored, not the raw token.
type RefreshToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"index"                    json:"user_id"`
	Hash      string    `gorm:"size:128;uniqueIndex"     json:"-"`
	ExpiresAt time.Time `gorm:"index"                    json:"expires_at"`
	Revoked   bool      `gorm:"default:false"            json:"revoked"`
	CreatedAt time.Time `                                json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
