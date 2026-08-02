package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// CategoryRepository manages the small gig taxonomy. The dataset is bounded
// (a few dozen rows in production), so List is un-paginated.
type CategoryRepository interface {
	Create(ctx context.Context, c *model.Category) error
	Update(ctx context.Context, c *model.Category) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.Category, error)
	List(ctx context.Context, status int8) ([]model.Category, error)
}

type categoryRepo struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, c *model.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *categoryRepo) Update(ctx context.Context, c *model.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *categoryRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

func (r *categoryRepo) GetByID(ctx context.Context, id uint64) (*model.Category, error) {
	var c model.Category
	err := r.db.WithContext(ctx).First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) List(ctx context.Context, status int8) ([]model.Category, error) {
	var out []model.Category
	q := r.db.WithContext(ctx).Model(&model.Category{}).Order("sort ASC, id ASC")
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
