package service

import (
	"context"
	"errors"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// MessageService owns the per-user notification stream. Admin broadcast is
// the only "write" path; everything else is just reads.
type MessageService struct {
	repo    repository.MessageRepository
	userRepo repository.UserRepository
}

func NewMessageService(repo repository.MessageRepository, userRepo repository.UserRepository) *MessageService {
	return &MessageService{repo: repo, userRepo: userRepo}
}

func (s *MessageService) ListByUser(ctx context.Context, userID uint64, page, size int, onlyUnread bool) ([]dto.MessageResp, int64, error) {
	rows, total, err := s.repo.ListByUser(ctx, page, size, userID, onlyUnread)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.MessageResp, 0, len(rows))
	for i := range rows {
		out = append(out, *toMessageResp(&rows[i]))
	}
	return out, total, nil
}

func (s *MessageService) MarkRead(ctx context.Context, userID, msgID uint64) error {
	m, err := s.repo.GetByID(ctx, msgID)
	if errors.Is(err, repository.ErrNotFound) {
		return httperr.NotFound("message not found")
	}
	if err != nil {
		return httperr.Internal(err)
	}
	if m.UserID != userID {
		return httperr.Forbidden("无权操作")
	}
	return httperr.WriteRaw(s.repo.MarkRead(ctx, msgID, userID))
}

func (s *MessageService) MarkAllRead(ctx context.Context, userID uint64) error {
	return httperr.WriteRaw(s.repo.MarkAllRead(ctx, userID))
}

// Broadcast fans a single message out to many users. Audience filter is
// "all" or by user_type (admin/student/employer). Implementation: query
// the user_ids from the user table filtered by role, then bulk-insert.
func (s *MessageService) Broadcast(ctx context.Context, req dto.BroadcastMessageReq) (int, error) {
	if req.Type == 0 {
		req.Type = 1
	}
	audience := req.UserType
	if audience == "" {
		audience = "all"
	}
	userIDs, err := s.resolveAudience(ctx, audience)
	if err != nil {
		return 0, httperr.Internal(err)
	}
	if len(userIDs) == 0 {
		return 0, nil
	}
	rows := make([]model.Message, 0, len(userIDs))
	for _, uid := range userIDs {
		rows = append(rows, model.Message{
			UserID: uid, Type: req.Type,
			Title: req.Title, Content: req.Content, Link: req.Link,
		})
	}
	if err := s.repo.CreateBatch(ctx, rows); err != nil {
		return 0, httperr.Internal(err)
	}
	return len(rows), nil
}

func (s *MessageService) resolveAudience(ctx context.Context, audience string) ([]uint64, error) {
	if audience == "all" {
		rows, _, err := s.userRepo.List(ctx, 1, 100000, "")
		if err != nil {
			return nil, err
		}
		out := make([]uint64, 0, len(rows))
		for _, u := range rows {
			out = append(out, u.ID)
		}
		return out, nil
	}
	rows, err := s.userRepo.ListByRole(ctx, audience)
	if err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(rows))
	for _, u := range rows {
		out = append(out, u.ID)
	}
	return out, nil
}

func toMessageResp(m *model.Message) *dto.MessageResp {
	return &dto.MessageResp{
		ID: m.ID, Type: m.Type, Title: m.Title, Content: m.Content,
		Link: m.Link, IsRead: m.IsRead, ReadAt: m.ReadAt, CreatedAt: m.CreatedAt,
	}
}
