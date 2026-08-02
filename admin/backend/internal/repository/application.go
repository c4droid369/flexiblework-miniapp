package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// ApplicationRepository. The composite UNIQUE (job_id, student_id) index is
// created by EnsureSchema (below) — GORM struct tags cannot express it.
type ApplicationRepository interface {
	Create(ctx context.Context, a *model.Application) error
	Update(ctx context.Context, a *model.Application) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.Application, error)
	GetByJobAndStudent(ctx context.Context, jobID, studentID uint64) (*model.Application, error)
	ListByStudent(ctx context.Context, page, size int, studentID uint64, status int8) ([]model.Application, int64, error)
	ListByJob(ctx context.Context, page, size int, jobID uint64, status int8) ([]model.Application, int64, error)
	UpdateStatus(ctx context.Context, id uint64, status int8, auditRemark string) error
	CountByStatuses(ctx context.Context, jobID uint64, statuses []int8) (int64, error)
	EnsureSchema(ctx context.Context) error
}

type applicationRepo struct{ db *gorm.DB }

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepo{db: db}
}

// EnsureSchema creates the (job_id, student_id) composite unique index if it
// does not exist. GORM v2 has no struct-tag syntax for composite unique
// indexes, so we issue raw SQL guarded by an INFORMATION_SCHEMA check.
func (r *applicationRepo) EnsureSchema(ctx context.Context) error {
	const idx = "uniq_applications_job_student"
	const ddl = "CREATE UNIQUE INDEX " + idx + " ON applications (job_id, student_id)"
	var count int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		"applications", idx,
	).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return r.db.WithContext(ctx).Exec(ddl).Error
}

func (r *applicationRepo) Create(ctx context.Context, a *model.Application) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *applicationRepo) Update(ctx context.Context, a *model.Application) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *applicationRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Application{}, id).Error
}

func (r *applicationRepo) GetByID(ctx context.Context, id uint64) (*model.Application, error) {
	var a model.Application
	err := r.db.WithContext(ctx).First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *applicationRepo) GetByJobAndStudent(ctx context.Context, jobID, studentID uint64) (*model.Application, error) {
	var a model.Application
	err := r.db.WithContext(ctx).
		Where("job_id = ? AND student_id = ?", jobID, studentID).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *applicationRepo) ListByStudent(ctx context.Context, page, size int, studentID uint64, status int8) ([]model.Application, int64, error) {
	var out []model.Application
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Application{}).Where("student_id = ?", studentID)
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("id DESC").Offset(off).Limit(size).Find(&out).Error
	return out, total, err
}

func (r *applicationRepo) ListByJob(ctx context.Context, page, size int, jobID uint64, status int8) ([]model.Application, int64, error) {
	var out []model.Application
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Application{}).Where("job_id = ?", jobID)
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("id DESC").Offset(off).Limit(size).Find(&out).Error
	return out, total, err
}

func (r *applicationRepo) UpdateStatus(ctx context.Context, id uint64, status int8, auditRemark string) error {
	updates := map[string]any{"status": status, "audit_remark": auditRemark}
	if status == 2 || status == 3 {
		updates["audited_at"] = gorm.Expr("NOW()")
	}
	return r.db.WithContext(ctx).Model(&model.Application{}).
		Where("id = ?", id).Updates(updates).Error
}

func (r *applicationRepo) CountByStatuses(ctx context.Context, jobID uint64, statuses []int8) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Application{}).
		Where("job_id = ? AND status IN ?", jobID, statuses).
		Count(&n).Error
	return n, err
}
