package service

import (
	"context"
	"errors"
	"time"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/repository"
)

// UserService is the user-management façade. Handlers depend on this, not on
// the repository directly.
type UserService struct {
	repo     repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewUserService(repo repository.UserRepository, roleRepo repository.RoleRepository) *UserService {
	return &UserService{repo: repo, roleRepo: roleRepo}
}

// Create validates uniqueness and inserts a new user with the supplied roles.
func (s *UserService) Create(ctx context.Context, req dto.CreateUserReq) (*dto.UserResp, error) {
	existing, err := s.repo.GetByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return nil, httperr.Conflict("username already exists")
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.Internal(err)
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, httperr.BadRequest("weak password")
	}
	u := &model.User{
		Username: req.Username, PasswordHash: hash, Nickname: req.Nickname,
		Email: req.Email, Phone: req.Phone, Avatar: req.Avatar,
		Remark: req.Remark, Status: model.UserStatusActive,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, httperr.Internal(err)
	}
	if len(req.RoleIDs) > 0 {
		if err := s.repo.AssignRoles(ctx, u.ID, req.RoleIDs); err != nil {
			return nil, httperr.Internal(err)
		}
	}
	return s.Get(ctx, u.ID)
}

// Update mutates the supplied fields and (optionally) reassigns roles.
func (s *UserService) Update(ctx context.Context, id uint64, req dto.UpdateUserReq) (*dto.UserResp, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.NotFound("user not found")
		}
		return nil, httperr.Internal(err)
	}
	if req.Nickname != nil {
		u.Nickname = *req.Nickname
	}
	if req.Email != nil {
		u.Email = *req.Email
	}
	if req.Phone != nil {
		u.Phone = *req.Phone
	}
	if req.Avatar != nil {
		u.Avatar = *req.Avatar
	}
	if req.Remark != nil {
		u.Remark = *req.Remark
	}
	if req.Status != nil {
		u.Status = model.UserStatus(*req.Status)
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, httperr.Internal(err)
	}
	if req.RoleIDs != nil {
		if err := s.repo.AssignRoles(ctx, id, req.RoleIDs); err != nil {
			return nil, httperr.Internal(err)
		}
	}
	return s.Get(ctx, id)
}

// Delete removes a user by id. Caller must be sure — this is irreversible.
func (s *UserService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("user not found")
		}
		return httperr.Internal(err)
	}
	return httperr.WriteRaw(s.repo.Delete(ctx, id))
}

// BatchDelete removes many users in one transaction.
func (s *UserService) BatchDelete(ctx context.Context, ids []uint64) error {
	return httperr.WriteRaw(s.repo.BatchDelete(ctx, ids))
}

// ResetPassword sets a new password (bcrypt-hashed) for a user.
func (s *UserService) ResetPassword(ctx context.Context, id uint64, newPwd string) error {
	hash, err := auth.HashPassword(newPwd)
	if err != nil {
		return httperr.BadRequest("weak password")
	}
	if err := s.repo.ResetPassword(ctx, id, hash); err != nil {
		return httperr.Internal(err)
	}
	return nil
}

// ChangeStatus toggles a user between active and disabled.
func (s *UserService) ChangeStatus(ctx context.Context, id uint64, status int8) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("user not found")
		}
		return httperr.Internal(err)
	}
	u.Status = model.UserStatus(status)
	return httperr.WriteRaw(s.repo.Update(ctx, u))
}

// Get returns one user by id with roles populated.
func (s *UserService) Get(ctx context.Context, id uint64) (*dto.UserResp, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.NotFound("user not found")
		}
		return nil, httperr.Internal(err)
	}
	return toUserResp(u), nil
}

// List returns a paginated, fuzzy-searched list of users.
func (s *UserService) List(ctx context.Context, page pagination.Page, search pagination.Search) ([]dto.UserResp, int64, error) {
	users, total, err := s.repo.List(ctx, page.Page, page.Size, search.Keyword)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.UserResp, 0, len(users))
	for i := range users {
		out = append(out, *toUserResp(&users[i]))
	}
	return out, total, nil
}

// AssignRoles is the dedicated "assign roles" endpoint. Empty list clears.
func (s *UserService) AssignRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	if _, err := s.repo.GetByID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("user not found")
		}
		return httperr.Internal(err)
	}
	return httperr.WriteRaw(s.repo.AssignRoles(ctx, userID, roleIDs))
}

func toUserResp(u *model.User) *dto.UserResp {
	r := &dto.UserResp{
		ID: u.ID, Username: u.Username, Nickname: u.Nickname,
		Email: u.Email, Phone: u.Phone, Avatar: u.Avatar,
		Status: int8(u.Status), LastLoginIP: u.LastLoginIP, Remark: u.Remark,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		Roles:     make([]dto.RoleBrief, 0, len(u.Roles)),
	}
	if u.LastLoginAt != nil {
		r.LastLoginAt = u.LastLoginAt.Format(time.RFC3339)
	}
	for _, role := range u.Roles {
		r.Roles = append(r.Roles, dto.RoleBrief{ID: role.ID, Code: role.Code, Name: role.Name})
	}
	return r
}
