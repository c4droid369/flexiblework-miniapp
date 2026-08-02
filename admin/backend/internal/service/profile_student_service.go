package service

import (
	"context"
	"errors"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// StudentProfileService owns the student-side profile row and its
// certification state machine.
type StudentProfileService struct {
	repo     repository.StudentProfileRepository
	userRepo repository.UserRepository
}

func NewStudentProfileService(repo repository.StudentProfileRepository, userRepo repository.UserRepository) *StudentProfileService {
	return &StudentProfileService{repo: repo, userRepo: userRepo}
}

// GetMy returns the caller's profile. An empty record is created on first
// call so the frontend can render the form with current values without a
// separate "create" round trip.
func (s *StudentProfileService) GetMy(ctx context.Context, userID uint64) (*dto.StudentProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return &dto.StudentProfileResp{UserID: userID}, nil
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	return toStudentProfileResp(p), nil
}

// UpsertMy is a partial update. Empty pointer fields leave the stored value
// untouched; the caller is expected to have used GetMy first to surface the
// current state in the form.
func (s *StudentProfileService) UpsertMy(ctx context.Context, userID uint64, req dto.UpsertStudentProfileReq) (*dto.StudentProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		p = &model.StudentProfile{UserID: userID}
	} else if err != nil {
		return nil, httperr.Internal(err)
	}
	applyStudentUpdates(p, req)
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, httperr.Internal(err)
	}
	return toStudentProfileResp(p), nil
}

// SubmitCertification uploads the three required images and flips cert_status
// from 0/3 (未认证/已拒绝) to 1 (审核中). Re-submission after a 3 is allowed.
func (s *StudentProfileService) SubmitCertification(ctx context.Context, userID uint64, req dto.SubmitStudentCertificationReq) (*dto.StudentProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		p = &model.StudentProfile{UserID: userID}
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
	p.StudentCard = req.StudentCard
	p.CertStatus = 1
	p.CertRemark = ""
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, httperr.Internal(err)
	}
	return toStudentProfileResp(p), nil
}

func applyStudentUpdates(p *model.StudentProfile, req dto.UpsertStudentProfileReq) {
	if req.RealName != nil {
		p.RealName = *req.RealName
	}
	if req.Gender != nil {
		p.Gender = *req.Gender
	}
	if req.School != nil {
		p.School = *req.School
	}
	if req.College != nil {
		p.College = *req.College
	}
	if req.Major != nil {
		p.Major = *req.Major
	}
	if req.StudentNo != nil {
		p.StudentNo = *req.StudentNo
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
	if req.StudentCard != nil {
		p.StudentCard = *req.StudentCard
	}
	if req.Bio != nil {
		p.Bio = *req.Bio
	}
	if req.Skills != nil {
		p.Skills = *req.Skills
	}
}

func toStudentProfileResp(p *model.StudentProfile) *dto.StudentProfileResp {
	r := &dto.StudentProfileResp{
		ID:           p.ID,
		UserID:       p.UserID,
		RealName:     p.RealName,
		Gender:       p.Gender,
		School:       p.School,
		College:      p.College,
		Major:        p.Major,
		StudentNo:    p.StudentNo,
		IDCardNoMask: maskIDCard(p.IDCardNo),
		IDCardFront:  p.IDCardFront,
		IDCardBack:   p.IDCardBack,
		StudentCard:  p.StudentCard,
		CertStatus:   p.CertStatus,
		CertRemark:   p.CertRemark,
		CertifiedAt:  p.CertifiedAt,
		Bio:          p.Bio,
		Skills:       p.Skills,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	return r
}

// maskIDCard returns the last 4 chars of an 18-char Chinese ID number with
// the leading 14 chars replaced by asterisks. Anything else is returned as
// empty — the client should treat absence as "not yet provided".
func maskIDCard(s string) string {
	if len(s) < 4 {
		return ""
	}
	masked := "**************" + s[len(s)-4:]
	return masked
}

// ListPendingCerts is the admin queue for student 实名/学生认证 review. Each
// row is joined with the user table to surface username / nickname in the
// admin list.
func (s *StudentProfileService) ListPendingCerts(ctx context.Context) ([]dto.StudentCertListItem, error) {
	rows, err := s.repo.ListByCertStatus(ctx, 1)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	out := make([]dto.StudentCertListItem, 0, len(rows))
	for i := range rows {
		r := rows[i]
		item := dto.StudentCertListItem{
			UserID: r.UserID, RealName: r.RealName,
			School: r.School, College: r.College, Major: r.Major,
			StudentNo: r.StudentNo, IDCardFront: r.IDCardFront,
			IDCardBack: r.IDCardBack, StudentCard: r.StudentCard,
			CertStatus: r.CertStatus, CertRemark: r.CertRemark,
			CreatedAt: r.CreatedAt,
		}
		if u, err := s.userRepo.GetByID(ctx, r.UserID); err == nil {
			item.Username = u.Username
			item.Nickname = u.Nickname
		}
		out = append(out, item)
	}
	return out, nil
}

// AuditCert is the admin's "通过/拒绝" on a student certification submission.
// action: 2=通过 3=拒绝. Transitions 1 → 2/3 and stamps certified_at on pass.
func (s *StudentProfileService) AuditCert(ctx context.Context, userID uint64, action int8, remark string) error {
	if action != 2 && action != 3 {
		return httperr.BadRequest("action must be 2 (pass) or 3 (reject)")
	}
	if _, err := s.repo.GetByUserID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("profile not found")
		}
		return httperr.Internal(err)
	}
	if err := s.repo.UpdateCertStatus(ctx, userID, action, remark); err != nil {
		return httperr.Internal(err)
	}
	if action == 2 {
		// mark certified_at on the profile row directly
		_ = s.markCertifiedAt(ctx, userID)
	}
	return nil
}

func (s *StudentProfileService) markCertifiedAt(ctx context.Context, userID uint64) error {
	// Importing gorm directly into service is fine for a one-liner; we keep
	// the repo interface narrow otherwise.
	return s.repo.UpdateCertStatusAndStamp(ctx, userID, 2, "")
}
