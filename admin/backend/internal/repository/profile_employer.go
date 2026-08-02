package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// EmployerProfileRepository persists employer-side profile data keyed by user_id.
type EmployerProfileRepository interface {
	GetByUserID(ctx context.Context, userID uint64) (*model.EmployerProfile, error)
	ListByCertStatus(ctx context.Context, status int8) ([]model.EmployerProfile, error)
	Upsert(ctx context.Context, p *model.EmployerProfile) error
	UpdateCertStatus(ctx context.Context, userID uint64, status int8, remark string) error
	UpdateCertStatusAndStamp(ctx context.Context, userID uint64, status int8, remark string) error
	IncTotalJobs(ctx context.Context, userID uint64, delta int) error
	IncCompletedOrders(ctx context.Context, userID uint64, delta int) error
}

type employerProfileRepo struct{ db *gorm.DB }

func NewEmployerProfileRepository(db *gorm.DB) EmployerProfileRepository {
	return &employerProfileRepo{db: db}
}

func (r *employerProfileRepo) GetByUserID(ctx context.Context, userID uint64) (*model.EmployerProfile, error) {
	var p model.EmployerProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *employerProfileRepo) ListByCertStatus(ctx context.Context, status int8) ([]model.EmployerProfile, error) {
	var out []model.EmployerProfile
	err := r.db.WithContext(ctx).Where("cert_status = ?", status).
		Order("id DESC").Find(&out).Error
	return out, err
}

func (r *employerProfileRepo) Upsert(ctx context.Context, p *model.EmployerProfile) error {
	if p.ID == 0 {
		return r.db.WithContext(ctx).Create(p).Error
	}
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *employerProfileRepo) UpdateCertStatus(ctx context.Context, userID uint64, status int8, remark string) error {
	return r.db.WithContext(ctx).Model(&model.EmployerProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{"cert_status": status, "cert_remark": remark}).Error
}

func (r *employerProfileRepo) UpdateCertStatusAndStamp(ctx context.Context, userID uint64, status int8, remark string) error {
	return r.db.WithContext(ctx).Model(&model.EmployerProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"cert_status":  status,
			"cert_remark":  remark,
			"certified_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *employerProfileRepo) IncTotalJobs(ctx context.Context, userID uint64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.EmployerProfile{}).
		Where("user_id = ?", userID).
		Update("total_jobs", gorm.Expr("total_jobs + ?", delta)).Error
}

func (r *employerProfileRepo) IncCompletedOrders(ctx context.Context, userID uint64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.EmployerProfile{}).
		Where("user_id = ?", userID).
		Update("completed_orders", gorm.Expr("completed_orders + ?", delta)).Error
}
