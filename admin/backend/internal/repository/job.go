package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// JobListFilter narrows a JobRepository.List query. Zero values mean "no
// filter". Caller passes values straight from the query string after parsing.
type JobListFilter struct {
	CategoryID uint64
	Location   string
	SalaryMin  float64
	SalaryMax  float64
	Status     int8    // 0 = any
	EmployerID uint64  // for /employer/jobs
	Keyword    string  // matches title/description
	OnlyActive bool    // true → status in (2) 招聘中 (public listing default)
}

// JobRepository owns the gigs CRUD plus the denormalized counter maintenance
// the platform relies on for cheap rendering.
type JobRepository interface {
	Create(ctx context.Context, j *model.Job) error
	Update(ctx context.Context, j *model.Job) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.Job, error)
	List(ctx context.Context, page, size int, f JobListFilter) ([]model.Job, int64, error)
	UpdateStatus(ctx context.Context, id uint64, status int8, auditRemark string) error
	IncViewCount(ctx context.Context, id uint64) error
	IncApplyCount(ctx context.Context, id uint64, delta int) error
}

type jobRepo struct{ db *gorm.DB }

func NewJobRepository(db *gorm.DB) JobRepository { return &jobRepo{db: db} }

func (r *jobRepo) Create(ctx context.Context, j *model.Job) error {
	return r.db.WithContext(ctx).Create(j).Error
}

func (r *jobRepo) Update(ctx context.Context, j *model.Job) error {
	return r.db.WithContext(ctx).Save(j).Error
}

func (r *jobRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Job{}, id).Error
}

func (r *jobRepo) GetByID(ctx context.Context, id uint64) (*model.Job, error) {
	var j model.Job
	err := r.db.WithContext(ctx).First(&j, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *jobRepo) List(ctx context.Context, page, size int, f JobListFilter) ([]model.Job, int64, error) {
	var jobs []model.Job
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Job{})

	if f.OnlyActive {
		q = q.Where("status = ?", 2)
	} else if f.Status > 0 {
		q = q.Where("status = ?", f.Status)
	}
	if f.CategoryID > 0 {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.EmployerID > 0 {
		q = q.Where("employer_id = ?", f.EmployerID)
	}
	if f.Location != "" {
		q = q.Where("location LIKE ?", "%"+f.Location+"%")
	}
	if f.SalaryMin > 0 {
		q = q.Where("salary_max >= ?", f.SalaryMin)
	}
	if f.SalaryMax > 0 {
		q = q.Where("salary_min <= ?", f.SalaryMax)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("title LIKE ? OR description LIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("id DESC").Offset(off).Limit(size).Find(&jobs).Error
	return jobs, total, err
}

func (r *jobRepo) UpdateStatus(ctx context.Context, id uint64, status int8, auditRemark string) error {
	updates := map[string]any{"status": status, "audit_remark": auditRemark}
	if status == 2 || status == 4 {
		updates["audited_at"] = gorm.Expr("NOW()")
	}
	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).Updates(updates).Error
}

func (r *jobRepo) IncViewCount(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *jobRepo) IncApplyCount(ctx context.Context, id uint64, delta int) error {
	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).
		Update("apply_count", gorm.Expr("apply_count + ?", delta)).Error
}
