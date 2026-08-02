package service

import (
	"context"
	"errors"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// CategoryService is a thin CRUD over the gig taxonomy.
type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, req dto.CreateCategoryReq) (*dto.CategoryResp, error) {
	if req.Status == 0 {
		req.Status = 1
	}
	c := &model.Category{
		Name: req.Name, Icon: req.Icon, Sort: req.Sort,
		Status: req.Status, Description: req.Description,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, httperr.Internal(err)
	}
	return toCategoryResp(c), nil
}

func (s *CategoryService) Update(ctx context.Context, id uint64, req dto.UpdateCategoryReq) (*dto.CategoryResp, error) {
	c, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("category not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Icon != nil {
		c.Icon = *req.Icon
	}
	if req.Sort != nil {
		c.Sort = *req.Sort
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	if req.Description != nil {
		c.Description = *req.Description
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, httperr.Internal(err)
	}
	return toCategoryResp(c), nil
}

func (s *CategoryService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("category not found")
		}
		return httperr.Internal(err)
	}
	return httperr.WriteRaw(s.repo.Delete(ctx, id))
}

// List returns every category, optionally filtered by status. The dataset is
// small so no pagination; frontend renders them all.
func (s *CategoryService) List(ctx context.Context, status int8) ([]dto.CategoryResp, error) {
	rows, err := s.repo.List(ctx, status)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	out := make([]dto.CategoryResp, 0, len(rows))
	for i := range rows {
		out = append(out, *toCategoryResp(&rows[i]))
	}
	return out, nil
}

func (s *CategoryService) Get(ctx context.Context, id uint64) (*dto.CategoryResp, error) {
	c, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("category not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	return toCategoryResp(c), nil
}

func toCategoryResp(c *model.Category) *dto.CategoryResp {
	return &dto.CategoryResp{
		ID: c.ID, Name: c.Name, Icon: c.Icon, Sort: c.Sort,
		Status: c.Status, Description: c.Description,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}
