// Package model holds GORM struct definitions for every persistent entity.
// They are the DB shape, NOT the wire shape — handlers marshal them into DTOs
// at the API boundary.
//
// Adding a new model: define the struct here, then append it to All() so
// AutoMigrate picks it up.
package model

import (
	"time"

	"gorm.io/gorm"
)

// All returns every model registered for auto-migration. The order is
// significant for foreign-key relations: referenced tables come first.
func All() []any {
	return []any{
		new(User),
		new(Role),
		new(Menu),
		new(UserRole),
		new(RoleMenu),
		new(OperationLog),
		new(File),
		new(RefreshToken),
		// Business modules (campus gig work) — added 2026-08.
		// Order matches the entity graph: profile tables first (depend only on
		// users), then taxonomy, then transactional tables.
		new(StudentProfile),
		new(EmployerProfile),
		new(Category),
		new(Job),
		new(Application),
		new(Order),
		new(Review),
		new(Message),
	}
}

// Base is embedded by every model to provide consistent audit columns. The
// JSON tags keep the raw struct wire-friendly when debugging, but production
// responses must use DTOs to avoid leaking timestamps.
type Base struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"index"                    json:"created_at"`
	UpdatedAt time.Time      `                                json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                    json:"-"`
}
