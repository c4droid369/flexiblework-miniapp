package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// StudentProfileRepository persists student-side profile data keyed by user_id.
type StudentProfileRepository interface {
	GetByUserID(ctx context.Context, userID uint64) (*model.StudentProfile, error)
	ListByCertStatus(ctx context.Context, status int8) ([]model.StudentProfile, error)
	Upsert(ctx context.Context, p *model.StudentProfile) error
	UpdateCertStatus(ctx context.Context, userID uint64, status int8, remark string) error
	UpdateCertStatusAndStamp(ctx context.Context, userID uint64, status int8, remark string) error
}

type studentProfileRepo struct{ db *gorm.DB }

func NewStudentProfileRepository(db *gorm.DB) StudentProfileRepository {
	return &studentProfileRepo{db: db}
}

func (r *studentProfileRepo) GetByUserID(ctx context.Context, userID uint64) (*model.StudentProfile, error) {
	var p model.StudentProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *studentProfileRepo) ListByCertStatus(ctx context.Context, status int8) ([]model.StudentProfile, error) {
	var out []model.StudentProfile
	err := r.db.WithContext(ctx).Where("cert_status = ?", status).
		Order("id DESC").Find(&out).Error
	return out, err
}

func (r *studentProfileRepo) Upsert(ctx context.Context, p *model.StudentProfile) error {
	if p.ID == 0 {
		return r.db.WithContext(ctx).Create(p).Error
	}
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *studentProfileRepo) UpdateCertStatus(ctx context.Context, userID uint64, status int8, remark string) error {
	return r.db.WithContext(ctx).Model(&model.StudentProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{"cert_status": status, "cert_remark": remark}).Error
}

func (r *studentProfileRepo) UpdateCertStatusAndStamp(ctx context.Context, userID uint64, status int8, remark string) error {
	return r.db.WithContext(ctx).Model(&model.StudentProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"cert_status":  status,
			"cert_remark":  remark,
			"certified_at": gorm.Expr("NOW()"),
		}).Error
}
