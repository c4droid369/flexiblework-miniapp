package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// OperationLogRepository persists handler invocations captured by middleware.
type OperationLogRepository interface {
	Create(ctx context.Context, l *model.OperationLog) error
	List(ctx context.Context, page, size int, keyword, action string) ([]model.OperationLog, int64, error)
	BatchDelete(ctx context.Context, ids []uint64) error
}

type opLogRepo struct{ db *gorm.DB }

func NewOperationLogRepository(db *gorm.DB) OperationLogRepository {
	return &opLogRepo{db: db}
}

func (r *opLogRepo) Create(ctx context.Context, l *model.OperationLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *opLogRepo) List(ctx context.Context, page, size int, keyword, action string) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64
	q := r.db.WithContext(ctx).Model(&model.OperationLog{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username LIKE ? OR path LIKE ? OR action LIKE ?", like, like, like)
	}
	if action != "" {
		q = q.Where("action LIKE ?", "%"+action+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("id DESC").Offset(off).Limit(size).Find(&logs).Error
	return logs, total, err
}

func (r *opLogRepo) BatchDelete(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&model.OperationLog{}, ids).Error
}
