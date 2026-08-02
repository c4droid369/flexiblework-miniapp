package service

import (
	"context"
	"errors"
	"time"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/repository"
)

// RoleService owns role CRUD + menu/user assignment.
type RoleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) Create(ctx context.Context, req dto.CreateRoleReq) (*dto.RoleResp, error) {
	if existing, err := s.repo.GetByCode(ctx, req.Code); err == nil && existing != nil {
		return nil, httperr.Conflict("role code already exists")
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.Internal(err)
	}
	r := &model.Role{
		Code: req.Code, Name: req.Name, Description: req.Description,
		Sort: req.Sort, Status: model.RoleStatusActive,
	}
	if req.Status != 0 {
		r.Status = model.RoleStatus(req.Status)
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return nil, httperr.Internal(err)
	}
	if len(req.MenuIDs) > 0 {
		if err := s.repo.AssignMenus(ctx, r.ID, req.MenuIDs); err != nil {
			return nil, httperr.Internal(err)
		}
	}
	return s.Get(ctx, r.ID)
}

func (s *RoleService) Update(ctx context.Context, id uint64, req dto.UpdateRoleReq) (*dto.RoleResp, error) {
	r, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.NotFound("role not found")
		}
		return nil, httperr.Internal(err)
	}
	if req.Name != nil {
		r.Name = *req.Name
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if req.Sort != nil {
		r.Sort = *req.Sort
	}
	if req.Status != nil {
		r.Status = model.RoleStatus(*req.Status)
	}
	if err := s.repo.Update(ctx, r); err != nil {
		return nil, httperr.Internal(err)
	}
	if req.MenuIDs != nil {
		if err := s.repo.AssignMenus(ctx, id, req.MenuIDs); err != nil {
			return nil, httperr.Internal(err)
		}
	}
	return s.Get(ctx, id)
}

func (s *RoleService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("role not found")
		}
		return httperr.Internal(err)
	}
	return s.repo.Delete(ctx, id)
}

func (s *RoleService) BatchDelete(ctx context.Context, ids []uint64) error {
	return s.repo.BatchDelete(ctx, ids)
}

func (s *RoleService) AssignMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	if _, err := s.repo.GetByID(ctx, roleID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("role not found")
		}
		return httperr.Internal(err)
	}
	return s.repo.AssignMenus(ctx, roleID, menuIDs)
}

func (s *RoleService) Get(ctx context.Context, id uint64) (*dto.RoleResp, error) {
	r, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.NotFound("role not found")
		}
		return nil, httperr.Internal(err)
	}
	return toRoleResp(r), nil
}

func (s *RoleService) List(ctx context.Context, page pagination.Page, search pagination.Search) ([]dto.RoleResp, int64, error) {
	roles, total, err := s.repo.List(ctx, page.Page, page.Size, search.Keyword)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.RoleResp, 0, len(roles))
	for i := range roles {
		out = append(out, *toRoleResp(&roles[i]))
	}
	return out, total, nil
}

func toRoleResp(r *model.Role) *dto.RoleResp {
	menuIDs := make([]uint64, 0, len(r.Menus))
	for _, m := range r.Menus {
		menuIDs = append(menuIDs, m.ID)
	}
	return &dto.RoleResp{
		ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description,
		Sort: r.Sort, Status: int8(r.Status),
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
		MenuIDs:   menuIDs,
	}
}
