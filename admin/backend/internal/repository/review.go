package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// ReviewRepository. Each (order_id, from_user_id) pair may have at most one
// review — enforced by the GetByOrderAndFrom check in the service layer.
type ReviewRepository interface {
	Create(ctx context.Context, r *model.Review) error
	GetByID(ctx context.Context, id uint64) (*model.Review, error)
	GetByOrderAndFrom(ctx context.Context, orderID, fromUserID uint64) (*model.Review, error)
	ListByToUser(ctx context.Context, page, size int, toUserID uint64) ([]model.Review, int64, error)
	ListByOrder(ctx context.Context, orderID uint64) ([]model.Review, error)
	Delete(ctx context.Context, id uint64) error
}

type reviewRepo struct{ db *gorm.DB }

func NewReviewRepository(db *gorm.DB) ReviewRepository { return &reviewRepo{db: db} }

func (r *reviewRepo) Create(ctx context.Context, rv *model.Review) error {
	return r.db.WithContext(ctx).Create(rv).Error
}

func (r *reviewRepo) GetByID(ctx context.Context, id uint64) (*model.Review, error) {
	var rv model.Review
	err := r.db.WithContext(ctx).First(&rv, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepo) GetByOrderAndFrom(ctx context.Context, orderID, fromUserID uint64) (*model.Review, error) {
	var rv model.Review
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND from_user_id = ?", orderID, fromUserID).
		First(&rv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepo) ListByToUser(ctx context.Context, page, size int, toUserID uint64) ([]model.Review, int64, error) {
	var out []model.Review
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Review{}).Where("to_user_id = ?", toUserID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("id DESC").Offset(off).Limit(size).Find(&out).Error
	return out, total, err
}

func (r *reviewRepo) ListByOrder(ctx context.Context, orderID uint64) ([]model.Review, error) {
	var out []model.Review
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("id ASC").Find(&out).Error
	return out, err
}

func (r *reviewRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Review{}, id).Error
}
