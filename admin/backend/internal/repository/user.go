// Package repository defines persistence interfaces and their GORM
// implementations. Services depend on the interface, not the concrete type,
// so handlers can be tested against an in-memory fake.
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// ErrNotFound is returned when a unique lookup yields no row. Handlers
// translate this into a 404 via httperr.NotFound.
var ErrNotFound = errors.New("not found")

// UserRepository is the persistence contract for users.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, page, size int, keyword string) ([]model.User, int64, error)
	ListByRole(ctx context.Context, roleCode string) ([]model.User, error)
	AssignRoles(ctx context.Context, userID uint64, roleIDs []uint64) error
	UpdateLastLogin(ctx context.Context, id uint64, ip string) error
	ResetPassword(ctx context.Context, id uint64, newHash string) error
	BatchDelete(ctx context.Context, ids []uint64) error
}

type userRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *userRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, id).Error
	})
}

func (r *userRepo) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Preload("Roles").First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) List(ctx context.Context, page, size int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	q := r.db.WithContext(ctx).Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Preload("Roles").Order("id DESC").Offset(off).Limit(size).Find(&users).Error
	return users, total, err
}

// ListByRole returns every user that holds the supplied role code. Used by
// the broadcast service to compute audience — much cheaper than loading
// every user and filtering in Go.
func (r *userRepo) ListByRole(ctx context.Context, roleCode string) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ur ON ur.user_id = users.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("r.code = ?", roleCode).
		Find(&users).Error
	return users, err
}

func (r *userRepo) AssignRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]model.UserRole, 0, len(roleIDs))
		for _, rid := range roleIDs {
			rows = append(rows, model.UserRole{UserID: userID, RoleID: rid})
		}
		return tx.Create(&rows).Error
	})
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, id uint64, ip string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]any{"last_login_ip": ip, "last_login_at": gorm.Expr("NOW()")}).Error
}

func (r *userRepo) ResetPassword(ctx context.Context, id uint64, newHash string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Update("password_hash", newHash).Error
}

func (r *userRepo) BatchDelete(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id IN ?", ids).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, ids).Error
	})
}
