package service

import (
	"context"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/repository"
)

// OperationLogService is the read-side façade for the operation log. The
// middleware writes directly to the repository to keep the hot path off the
// service; this service exists for the query endpoints.
type OperationLogService struct {
	repo repository.OperationLogRepository
}

func NewOperationLogService(repo repository.OperationLogRepository) *OperationLogService {
	return &OperationLogService{repo: repo}
}

// Record is the write-side entrypoint used by the middleware. Errors are
// logged by the caller; we never fail a user request because logging failed.
func (s *OperationLogService) Record(ctx context.Context, l *model.OperationLog) error {
	return s.repo.Create(ctx, l)
}

// List returns a paginated, keyword-filtered slice of operation logs.
func (s *OperationLogService) List(ctx context.Context, page pagination.Page, search pagination.Search, action string) ([]model.OperationLog, int64, error) {
	logs, total, err := s.repo.List(ctx, page.Page, page.Size, search.Keyword, action)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	return logs, total, nil
}

func (s *OperationLogService) BatchDelete(ctx context.Context, ids []uint64) error {
	return s.repo.BatchDelete(ctx, ids)
}
