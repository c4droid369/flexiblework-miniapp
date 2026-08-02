package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// MessageRepository owns the per-user in-app message stream.
type MessageRepository interface {
	Create(ctx context.Context, m *model.Message) error
	CreateBatch(ctx context.Context, ms []model.Message) error
	GetByID(ctx context.Context, id uint64) (*model.Message, error)
	ListByUser(ctx context.Context, page, size int, userID uint64, onlyUnread bool) ([]model.Message, int64, error)
	MarkRead(ctx context.Context, id uint64, userID uint64) error
	MarkAllRead(ctx context.Context, userID uint64) error
}

type messageRepo struct{ db *gorm.DB }

func NewMessageRepository(db *gorm.DB) MessageRepository { return &messageRepo{db: db} }

func (r *messageRepo) Create(ctx context.Context, m *model.Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *messageRepo) CreateBatch(ctx context.Context, ms []model.Message) error {
	if len(ms) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&ms).Error
}

func (r *messageRepo) GetByID(ctx context.Context, id uint64) (*model.Message, error) {
	var m model.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepo) ListByUser(ctx context.Context, page, size int, userID uint64, onlyUnread bool) ([]model.Message, int64, error) {
	var out []model.Message
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Message{}).Where("user_id = ?", userID)
	if onlyUnread {
		q = q.Where("is_read = ?", false)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off := (page - 1) * size
	err := q.Order("id DESC").Offset(off).Limit(size).Find(&out).Error
	return out, total, err
}

func (r *messageRepo) MarkRead(ctx context.Context, id uint64, userID uint64) error {
	// Scoped by user_id so a user can only mark their own messages.
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{"is_read": true, "read_at": gorm.Expr("NOW()")}).Error
}

func (r *messageRepo) MarkAllRead(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{"is_read": true, "read_at": gorm.Expr("NOW()")}).Error
}
