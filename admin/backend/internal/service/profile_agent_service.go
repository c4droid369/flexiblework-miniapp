package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// AgentProfileService owns the agent-side profile and its cert workflow.
// Referral codes are generated on first profile creation so the agent has
// a stable share link from the start.
type AgentProfileService struct {
	repo     repository.AgentProfileRepository
	userRepo repository.UserRepository
}

func NewAgentProfileService(repo repository.AgentProfileRepository, userRepo repository.UserRepository) *AgentProfileService {
	return &AgentProfileService{repo: repo, userRepo: userRepo}
}

// GetMy returns the caller's profile. Empty record on first call so the
// frontend can render the form with no 404 round trip. Also lazily mints
// the referral code if it's missing — the code is stable (derived from
// userID) so subsequent calls return the same value.
func (s *AgentProfileService) GetMy(ctx context.Context, userID uint64) (*dto.AgentProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return &dto.AgentProfileResp{UserID: userID, ReferralCode: generateReferralCode(userID)}, nil
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if p.ReferralCode == "" {
		p.ReferralCode = generateReferralCode(userID)
		_ = s.repo.Upsert(ctx, p) // best-effort persist
	}
	return toAgentProfileResp(p), nil
}

// UpsertMy partial-updates the profile. On first creation a referral code is
// minted from the user ID so the agent can immediately share.
func (s *AgentProfileService) UpsertMy(ctx context.Context, userID uint64, req dto.UpsertAgentProfileReq) (*dto.AgentProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		p = &model.AgentProfile{UserID: userID}
		if p.ReferralCode == "" {
			p.ReferralCode = generateReferralCode(userID)
		}
	} else if err != nil {
		return nil, httperr.Internal(err)
	}
	applyAgentUpdates(p, req)
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, httperr.Internal(err)
	}
	return toAgentProfileResp(p), nil
}

// SubmitCertification uploads the identity + campus card and transitions
// cert_status from 0/3 → 1.
func (s *AgentProfileService) SubmitCertification(ctx context.Context, userID uint64, req dto.SubmitAgentCertificationReq) (*dto.AgentProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		p = &model.AgentProfile{UserID: userID, ReferralCode: generateReferralCode(userID)}
	} else if err != nil {
		return nil, httperr.Internal(err)
	}
	if p.CertStatus == 1 {
		return nil, httperr.Conflict("已在审核中,请耐心等待")
	}
	if p.CertStatus == 2 {
		return nil, httperr.Conflict("已通过认证,无需重复提交")
	}
	p.IDCardFront = req.IDCardFront
	p.IDCardBack = req.IDCardBack
	p.CampusCard = req.CampusCard
	p.RealName = req.RealName
	p.Phone = req.Phone
	p.CertStatus = 1
	p.CertRemark = ""
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, httperr.Internal(err)
	}
	return toAgentProfileResp(p), nil
}

// ListPendingCerts is the admin queue for agent certs.
func (s *AgentProfileService) ListPendingCerts(ctx context.Context) ([]dto.EmployerCertListItem, error) {
	// Reuse the EmployerCertListItem shape — admin cert UI doesn't need to
	// distinguish visually between employer / agent at a glance.
	rows, err := s.repo.ListByCertStatus(ctx, 1)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	out := make([]dto.EmployerCertListItem, 0, len(rows))
	for i := range rows {
		r := rows[i]
		item := dto.EmployerCertListItem{
			UserID:             r.UserID,
			CompanyName:        r.RealName + " (校园代理)",
			ContactName:        r.RealName,
			ContactPhone:       r.Phone,
			BusinessLicenseNo:  r.ReferralCode,
			BusinessLicenseImg: r.CampusCard,
			CertStatus:         r.CertStatus,
			CertRemark:         r.CertRemark,
			CreatedAt:          r.CreatedAt,
		}
		if u, err := s.userRepo.GetByID(ctx, r.UserID); err == nil {
			item.Username = u.Username
			item.Nickname = u.Nickname
		}
		out = append(out, item)
	}
	return out, nil
}

// AuditCert is the admin's "通过/拒绝" on an agent cert.
func (s *AgentProfileService) AuditCert(ctx context.Context, userID uint64, action int8, remark string) error {
	if action != 2 && action != 3 {
		return httperr.BadRequest("action must be 2 (pass) or 3 (reject)")
	}
	if _, err := s.repo.GetByUserID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("profile not found")
		}
		return httperr.Internal(err)
	}
	return httperr.WriteRaw(s.repo.UpdateCertStatusAndStamp(ctx, userID, action, remark))
}

func applyAgentUpdates(p *model.AgentProfile, req dto.UpsertAgentProfileReq) {
	if req.RealName != nil {
		p.RealName = *req.RealName
	}
	if req.Phone != nil {
		p.Phone = *req.Phone
	}
	if req.Wechat != nil {
		p.Wechat = *req.Wechat
	}
	if req.IDCardNo != nil {
		p.IDCardNo = *req.IDCardNo
	}
	if req.IDCardFront != nil {
		p.IDCardFront = *req.IDCardFront
	}
	if req.IDCardBack != nil {
		p.IDCardBack = *req.IDCardBack
	}
	if req.CampusCard != nil {
		p.CampusCard = *req.CampusCard
	}
	if req.Bio != nil {
		p.Bio = *req.Bio
	}
}

func toAgentProfileResp(p *model.AgentProfile) *dto.AgentProfileResp {
	return &dto.AgentProfileResp{
		ID:             p.ID,
		UserID:         p.UserID,
		RealName:       p.RealName,
		Phone:          p.Phone,
		Wechat:         p.Wechat,
		IDCardNoMask:   maskIDCard(p.IDCardNo),
		IDCardFront:    p.IDCardFront,
		IDCardBack:     p.IDCardBack,
		CampusCard:     p.CampusCard,
		CertStatus:     p.CertStatus,
		CertRemark:     p.CertRemark,
		CertifiedAt:    p.CertifiedAt,
		Bio:            p.Bio,
		ReferralCode:   p.ReferralCode,
		BankAccount:    p.BankAccount,
		BankName:       p.BankName,
		TotalReferrals: p.TotalReferrals,
		TotalEarnings:  p.TotalEarnings,
		Rating:         p.Rating,
		TotalJobs:      p.TotalJobs,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// generateReferralCode produces a deterministic 6-char code from the user ID.
// Not cryptographically secret — its purpose is collision-free display in
// invite links. Format: "A" + base36(userID) + 2-char checksum from a
// stable hash. The leading 'A' lets the agent recognize it visually.
func generateReferralCode(userID uint64) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := int64(userID)
	if id <= 0 {
		id = int64(time.Now().Unix() & 0xFFFFFF)
	}
	// base36 encode
	var sb strings.Builder
	for id > 0 {
		sb.WriteByte(charset[id%36])
		id /= 36
	}
	body := sb.String()
	// 2-char checksum
	cs := (userID * 2654435761) & 0xFFF
	ck := charset[cs%36]
	ck2 := charset[(cs>>6)%36]
	return fmt.Sprintf("A%s%c%c", body, ck, ck2)
}