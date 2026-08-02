package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// AgentProfileRepository persists agent-side profile data keyed by user_id.
type AgentProfileRepository interface {
	GetByUserID(ctx context.Context, userID uint64) (*model.AgentProfile, error)
	GetByReferralCode(ctx context.Context, code string) (*model.AgentProfile, error)
	ListByCertStatus(ctx context.Context, status int8) ([]model.AgentProfile, error)
	Upsert(ctx context.Context, p *model.AgentProfile) error
	UpdateCertStatusAndStamp(ctx context.Context, userID uint64, status int8, remark string) error
	IncTotalJobs(ctx context.Context, userID uint64, delta int) error
}

type agentProfileRepo struct{ db *gorm.DB }

func NewAgentProfileRepository(db *gorm.DB) AgentProfileRepository {
	return &agentProfileRepo{db: db}
}

func (r *agentProfileRepo) GetByUserID(ctx context.Context, userID uint64) (*model.AgentProfile, error) {
	var p model.AgentProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *agentProfileRepo) GetByReferralCode(ctx context.Context, code string) (*model.AgentProfile, error) {
	var p model.AgentProfile
	err := r.db.WithContext(ctx).Where("referral_code = ?", code).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *agentProfileRepo) ListByCertStatus(ctx context.Context, status int8) ([]model.AgentProfile, error) {
	var out []model.AgentProfile
	err := r.db.WithContext(ctx).Where("cert_status = ?", status).
		Order("id DESC").Find(&out).Error
	return out, err
}

func (r *agentProfileRepo) Upsert(ctx context.Context, p *model.AgentProfile) error {
	if p.ID == 0 {
		return r.db.WithContext(ctx).Create(p).Error
	}
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *agentProfileRepo) UpdateCertStatusAndStamp(ctx context.Context, userID uint64, status int8, remark string) error {
	return r.db.WithContext(ctx).Model(&model.AgentProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"cert_status":  status,
			"cert_remark":  remark,
			"certified_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *agentProfileRepo) IncTotalJobs(ctx context.Context, userID uint64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.AgentProfile{}).
		Where("user_id = ?", userID).
		Update("total_jobs", gorm.Expr("total_jobs + ?", delta)).Error
}