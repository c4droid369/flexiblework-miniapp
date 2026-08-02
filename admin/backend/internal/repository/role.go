package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// RoleRepository persists roles and their menu assignments.
type RoleRepository interface {
	Create(ctx context.Context, r *model.Role) error
	Update(ctx context.Context, r *model.Role) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
	List(ctx context.Context, page, size int, keyword string) ([]model.Role, int64, error)
	AssignMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error
	BatchDelete(ctx context.Context, ids []uint64) error
}

type roleRepo struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) RoleRepository { return &roleRepo{db: db} }

func (r *roleRepo) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepo) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Role{}, id).Error
	})
}

func (r *roleRepo) GetByID(ctx context.Context, id uint64) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Menus").First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) List(ctx context.Context, page, size int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Role{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("sort ASC, id ASC").Offset(off).Limit(size).Find(&roles).Error
	return roles, total, err
}

func (r *roleRepo) AssignMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		rows := make([]model.RoleMenu, 0, len(menuIDs))
		for _, mid := range menuIDs {
			rows = append(rows, model.RoleMenu{RoleID: roleID, MenuID: mid})
		}
		return tx.Create(&rows).Error
	})
}

func (r *roleRepo) BatchDelete(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id IN ?", ids).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id IN ?", ids).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Role{}, ids).Error
	})
}
