// Package service holds the business logic. Services depend on repository
// interfaces and pkg/* primitives; they never import gin or sqlx.
package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/repository"
)

// AuthService owns login, refresh, logout, register, and the /auth/me
// aggregator. It pulls together user lookup, JWT issuance, refresh-token
// persistence, and menu/permission projection in one place so handlers stay
// thin.
type AuthService struct {
	users        repository.UserRepository
	refresh      repository.RefreshTokenRepository
	roleRepo     repository.RoleRepository
	menuRepo     repository.MenuRepository
	studentRepo  repository.StudentProfileRepository
	employerRepo repository.EmployerProfileRepository
	agentRepo    repository.AgentProfileRepository
	issuer       *auth.Issuer
	accessTTL    time.Duration
	logger       *slog.Logger
}

func NewAuthService(
	users repository.UserRepository,
	refresh repository.RefreshTokenRepository,
	roleRepo repository.RoleRepository,
	menuRepo repository.MenuRepository,
	studentRepo repository.StudentProfileRepository,
	employerRepo repository.EmployerProfileRepository,
	agentRepo repository.AgentProfileRepository,
	issuer *auth.Issuer,
	accessTTL time.Duration,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		users: users, refresh: refresh, roleRepo: roleRepo, menuRepo: menuRepo,
		studentRepo: studentRepo, employerRepo: employerRepo, agentRepo: agentRepo,
		issuer: issuer, accessTTL: accessTTL, logger: logger,
	}
}

// Login verifies credentials, updates last-login metadata, and issues an
// access + refresh token pair.
func (s *AuthService) Login(ctx context.Context, req dto.LoginReq, ip string) (*dto.LoginResp, error) {
	u, err := s.users.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.Unauthorized("invalid username or password")
		}
		return nil, httperr.Internal(err)
	}
	if !u.IsActive() {
		return nil, httperr.Forbidden("account disabled")
	}
	if err := auth.VerifyPassword(u.PasswordHash, req.Password); err != nil {
		return nil, httperr.Unauthorized("invalid username or password")
	}
	if err := s.users.UpdateLastLogin(ctx, u.ID, ip); err != nil {
		s.logger.WarnContext(ctx, "update last login", slog.Any("err", err))
	}

	perms, err := s.collectPermissions(ctx, u)
	if err != nil {
		return nil, httperr.Internal(err)
	}

	access, _, err := s.issuer.IssueAccess(u.ID, perms)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	refreshRaw, refreshExp, err := s.issuer.IssueRefresh(u.ID)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if err := s.refresh.Create(ctx, u.ID, refreshRaw, refreshExp); err != nil {
		return nil, httperr.Internal(err)
	}

	return &dto.LoginResp{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		TokenType:    "Bearer",
		Permissions:  perms,
	}, nil
}

// Register creates a new student/employer account, attaches the right role,
// and returns a fresh login response. Atomic-ish: the user row is created
// first, then the role assignment and profile row. If the role or profile
// insert fails, the user row is left in place — a partial state is
// preferable to a duplicate-name collision on retry.
func (s *AuthService) Register(ctx context.Context, req dto.RegisterReq, ip string) (*dto.LoginResp, error) {
	if existing, err := s.users.GetByUsername(ctx, req.Username); err == nil && existing != nil {
		return nil, httperr.Conflict("username already exists")
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.Internal(err)
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, httperr.BadRequest("weak password")
	}
	roleCode := req.UserType
	role, err := s.roleRepo.GetByCode(ctx, roleCode)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	nick := req.Nickname
	if nick == "" {
		nick = req.Username
	}
	u := &model.User{
		Username: req.Username, PasswordHash: hash,
		Nickname: nick, Phone: req.Phone, Email: req.Email,
		Status: model.UserStatusActive,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, httperr.Internal(err)
	}
	if err := s.users.AssignRoles(ctx, u.ID, []uint64{role.ID}); err != nil {
		return nil, httperr.Internal(err)
	}
	// Create the empty profile row so the frontend's first GET returns an
	// object instead of 404. Saves a round trip on the first profile load.
	// Agent referral codes are minted on first GetMy (in profile_agent_service)
	// so we don't need to duplicate the algorithm here.
	switch req.UserType {
	case "student":
		_ = s.studentRepo.Upsert(ctx, &model.StudentProfile{UserID: u.ID})
	case "employer":
		_ = s.employerRepo.Upsert(ctx, &model.EmployerProfile{UserID: u.ID})
	case "agent":
		_ = s.agentRepo.Upsert(ctx, &model.AgentProfile{UserID: u.ID})
	}
	// Reload to pick up the freshly assigned role.
	u, err = s.users.GetByID(ctx, u.ID)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	perms, err := s.collectPermissions(ctx, u)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	access, _, err := s.issuer.IssueAccess(u.ID, perms)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	refreshRaw, refreshExp, err := s.issuer.IssueRefresh(u.ID)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if err := s.refresh.Create(ctx, u.ID, refreshRaw, refreshExp); err != nil {
		return nil, httperr.Internal(err)
	}
	_ = s.users.UpdateLastLogin(ctx, u.ID, ip)
	return &dto.LoginResp{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		TokenType:    "Bearer",
		Permissions:  perms,
	}, nil
}

// Refresh exchanges a valid refresh token for a fresh pair. The old refresh
// is revoked (rotation). Reuse of a revoked token revokes ALL the user's
// tokens — defends against token theft.
func (s *AuthService) Refresh(ctx context.Context, raw string) (*dto.LoginResp, error) {
	claims, err := s.issuer.Parse(raw, auth.KindRefresh)
	if err != nil {
		return nil, httperr.Unauthorized("invalid refresh token")
	}
	rt, err := s.refresh.FindValid(ctx, raw)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			_ = s.refresh.RevokeAllForUser(ctx, claims.UserID)
			return nil, httperr.Unauthorized("refresh token revoked")
		}
		return nil, httperr.Internal(err)
	}
	if err := s.refresh.Revoke(ctx, rt.ID); err != nil {
		return nil, httperr.Internal(err)
	}
	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	perms, err := s.collectPermissions(ctx, u)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	access, _, err := s.issuer.IssueAccess(claims.UserID, perms)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	newRefresh, newExp, err := s.issuer.IssueRefresh(claims.UserID)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if err := s.refresh.Create(ctx, claims.UserID, newRefresh, newExp); err != nil {
		return nil, httperr.Internal(err)
	}
	return &dto.LoginResp{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		TokenType:    "Bearer",
		Permissions:  perms,
	}, nil
}

// Logout revokes the supplied refresh token (if any) and the user's whole
// chain. Stateless access tokens stay valid until expiry; the frontend
// discards them.
func (s *AuthService) Logout(ctx context.Context, userID uint64, rawRefresh string) error {
	if rawRefresh != "" {
		if rt, err := s.refresh.FindValid(ctx, rawRefresh); err == nil {
			_ = s.refresh.Revoke(ctx, rt.ID)
		}
	}
	return s.refresh.RevokeAllForUser(ctx, userID)
}

// Me projects the authenticated user for the frontend: profile, roles,
// permission codes, the menu tree (for dynamic route registration), and the
// derived user_type flags (admin / student / employer) the mini-program
// uses to decide which tabbar to render.
func (s *AuthService) Me(ctx context.Context, userID uint64) (*dto.MeResp, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httperr.Unauthorized("user no longer exists")
		}
		return nil, httperr.Internal(err)
	}
	roles := make([]dto.RoleBrief, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, dto.RoleBrief{ID: r.ID, Code: r.Code, Name: r.Name})
	}
	perms, err := s.collectPermissions(ctx, u)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	menus, err := s.menuRepo.TreeForRoles(ctx, roleIDs(u.Roles))
	if err != nil {
		return nil, httperr.Internal(err)
	}
	userTypes := deriveUserTypes(u.Roles)
	resp := &dto.MeResp{
		ID: u.ID, Username: u.Username, Nickname: u.Nickname,
		Email: u.Email, Phone: u.Phone, Avatar: u.Avatar,
		Status: int8(u.Status), Roles: roles, Permissions: perms, Menus: menus,
		UserTypes: userTypes,
	}
	if len(userTypes) > 0 {
		resp.UserType = userTypes[0]
	}
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format(time.RFC3339)
		resp.LastLoginAt = &s
	}
	return resp, nil
}

func (s *AuthService) collectPermissions(ctx context.Context, u *model.User) ([]string, error) {
	if len(u.Roles) == 0 {
		return []string{}, nil
	}
	// Menu-based perms (admin side).
	menuPerms, err := s.menuRepo.PermCodesForRoles(ctx, roleIDs(u.Roles))
	if err != nil {
		return nil, err
	}
	// Role-code-based perms (student / employer). Hardcoded in code so the
	// template doesn't need a deep menu tree for every business action; the
	// admin menu tree still drives the admin UI navigation.
	rolePerms := roleCodePermissions(u.Roles)
	// Merge (dedup).
	seen := make(map[string]struct{}, len(menuPerms)+len(rolePerms))
	out := make([]string, 0, len(menuPerms)+len(rolePerms))
	for _, p := range menuPerms {
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	for _, p := range rolePerms {
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	return out, nil
}

// roleCodePermissions returns the business permission codes that map to a
// role code, independent of the menu table. Used so student/employer/agent
// roles don't have to thread their permissions through a hidden menu tree.
//
// agent is intentionally same-shape as employer — agents post jobs, audit
// applications, settle orders through the same /employer/jobs endpoints.
// JobService.Create checks both profile types when verifying the cert gate.
func roleCodePermissions(roles []model.Role) []string {
	student := []string{
		"profile:view", "profile:update", "cert:submit",
		"job:view", "category:view", "job:apply",
		"application:view", "application:cancel",
		"order:view", "order:pay", "order:checkin", "order:complete", "order:cancel",
		"review:create", "review:view",
		"message:view",
	}
	business := []string{
		"profile:view", "profile:update", "cert:submit",
		"job:view", "category:view",
		"job:create", "job:update", "job:delete", "job:offline",
		"application:view", "application:audit",
		"order:view", "order:settle", "order:cancel",
		"review:create", "review:view",
		"message:view",
	}
	var out []string
	for _, r := range roles {
		switch r.Code {
		case "student":
			out = append(out, student...)
		case "employer", "agent":
			// agent mirrors employer — they share the same business flow.
			out = append(out, business...)
		}
	}
	return out
}

func roleIDs(rs []model.Role) []uint64 {
	ids := make([]uint64, 0, len(rs))
	for _, r := range rs {
		ids = append(ids, r.ID)
	}
	return ids
}

// deriveUserTypes maps the user's role codes to the user_type vocabulary
// the mini-program consumes. The order is fixed: admin first, then
// business roles, then student — so a super-admin who is also a student gets
// UserType="admin" but UserTypes contains both.
func deriveUserTypes(roles []model.Role) []string {
	const admin, employer, agent, student = "admin", "employer", "agent", "student"
	var have [4]bool
	for _, r := range roles {
		switch r.Code {
		case "super_admin", "admin":
			have[0] = true
		case "employer":
			have[1] = true
		case "agent":
			have[2] = true
		case "student":
			have[3] = true
		}
	}
	out := make([]string, 0, 4)
	if have[0] {
		out = append(out, admin)
	}
	if have[1] {
		out = append(out, employer)
	}
	if have[2] {
		out = append(out, agent)
	}
	if have[3] {
		out = append(out, student)
	}
	return out
}
