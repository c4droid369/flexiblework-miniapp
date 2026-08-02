package service

import (
	"context"
	"errors"
	"time"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// MenuService owns menu CRUD and exposes the menu tree to other services.
type MenuService struct {
	repo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

func (s *MenuService) Create(ctx context.Context, req dto.CreateMenuReq) (*dto.MenuResp, error) {
	m := &model.Menu{
		ParentID: req.ParentID, Type: model.MenuType(req.Type),
		Name: req.Name, Title: req.Title, Path: req.Path, Component: req.Component,
		PermCode: req.PermCode, Icon: req.Icon, Sort: req.Sort, Visible: req.Visible,
		Status: model.RoleStatusActive,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, httperr.Internal(err)
	}
	return s.Get(ctx, m.ID)
}

func (s *MenuService) Update(ctx context.Context, id uint64, req dto.UpdateMenuReq) (*dto.MenuResp, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.NotFound("menu not found")
		}
		return nil, httperr.Internal(err)
	}
	if req.ParentID != nil {
		m.ParentID = *req.ParentID
	}
	if req.Type != nil {
		m.Type = model.MenuType(*req.Type)
	}
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Title != nil {
		m.Title = *req.Title
	}
	if req.Path != nil {
		m.Path = *req.Path
	}
	if req.Component != nil {
		m.Component = *req.Component
	}
	if req.PermCode != nil {
		m.PermCode = *req.PermCode
	}
	if req.Icon != nil {
		m.Icon = *req.Icon
	}
	if req.Sort != nil {
		m.Sort = *req.Sort
	}
	if req.Visible != nil {
		m.Visible = *req.Visible
	}
	if req.Status != nil {
		m.Status = model.RoleStatus(*req.Status)
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, httperr.Internal(err)
	}
	return s.Get(ctx, id)
}

func (s *MenuService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("menu not found")
		}
		return httperr.Internal(err)
	}
	return s.repo.Delete(ctx, id)
}

// Tree returns the full menu tree (admin view, no role filter).
func (s *MenuService) Tree(ctx context.Context) ([]dto.MenuTree, error) {
	menus, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	idx := make(map[uint64]*dto.MenuTree, len(menus))
	roots := []*dto.MenuTree{}
	for i := range menus {
		m := menus[i]
		node := &dto.MenuTree{
			ID: m.ID, ParentID: m.ParentID, Type: int8(m.Type),
			Name: m.Name, Title: m.Title, Path: m.Path, Component: m.Component,
			PermCode: m.PermCode, Icon: m.Icon, Sort: m.Sort, Visible: m.Visible,
		}
		idx[m.ID] = node
		if m.ParentID == 0 {
			roots = append(roots, node)
		}
	}
	for i := range menus {
		m := menus[i]
		if m.ParentID == 0 {
			continue
		}
		if parent, ok := idx[m.ParentID]; ok {
			cp := *idx[m.ID]
			parent.Children = append(parent.Children, cp)
		}
	}
	out := make([]dto.MenuTree, 0, len(roots))
	for _, r := range roots {
		out = append(out, *r)
	}
	return out, nil
}

func (s *MenuService) Get(ctx context.Context, id uint64) (*dto.MenuResp, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.NotFound("menu not found")
		}
		return nil, httperr.Internal(err)
	}
	return toMenuResp(m), nil
}

func toMenuResp(m *model.Menu) *dto.MenuResp {
	return &dto.MenuResp{
		ID: m.ID, ParentID: m.ParentID, Type: int8(m.Type),
		Name: m.Name, Title: m.Title, Path: m.Path, Component: m.Component,
		PermCode: m.PermCode, Icon: m.Icon, Sort: m.Sort, Visible: m.Visible,
		Status:    int8(m.Status),
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}
