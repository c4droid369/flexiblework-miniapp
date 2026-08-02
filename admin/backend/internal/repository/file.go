package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// FileRepository persists file metadata. The bytes live in the storage
// backend; the row only tracks key + name + size.
type FileRepository interface {
	Create(ctx context.Context, f *model.File) error
	GetByID(ctx context.Context, id uint64) (*model.File, error)
	List(ctx context.Context, page, size int, keyword string) ([]model.File, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type fileRepo struct{ db *gorm.DB }

func NewFileRepository(db *gorm.DB) FileRepository { return &fileRepo{db: db} }

func (r *fileRepo) Create(ctx context.Context, f *model.File) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *fileRepo) GetByID(ctx context.Context, id uint64) (*model.File, error) {
	var f model.File
	err := r.db.WithContext(ctx).First(&f, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *fileRepo) List(ctx context.Context, page, size int, keyword string) ([]model.File, int64, error) {
	var files []model.File
	var total int64
	q := r.db.WithContext(ctx).Model(&model.File{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("original_name LIKE ? OR name LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("id DESC").Offset(off).Limit(size).Find(&files).Error
	return files, total, err
}

func (r *fileRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.File{}, id).Error
}
