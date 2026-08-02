package service

import (
	"context"
	"errors"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// EmployerProfileService owns the employer-side profile and the
// business-license certification flow.
type EmployerProfileService struct {
	repo     repository.EmployerProfileRepository
	userRepo repository.UserRepository
}

func NewEmployerProfileService(repo repository.EmployerProfileRepository, userRepo repository.UserRepository) *EmployerProfileService {
	return &EmployerProfileService{repo: repo, userRepo: userRepo}
}

func (s *EmployerProfileService) GetMy(ctx context.Context, userID uint64) (*dto.EmployerProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return &dto.EmployerProfileResp{UserID: userID}, nil
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	return toEmployerProfileResp(p), nil
}

func (s *EmployerProfileService) UpsertMy(ctx context.Context, userID uint64, req dto.UpsertEmployerProfileReq) (*dto.EmployerProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		p = &model.EmployerProfile{UserID: userID}
	} else if err != nil {
		return nil, httperr.Internal(err)
	}
	applyEmployerUpdates(p, req)
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, httperr.Internal(err)
	}
	return toEmployerProfileResp(p), nil
}

func (s *EmployerProfileService) SubmitCertification(ctx context.Context, userID uint64, req dto.SubmitEmployerCertificationReq) (*dto.EmployerProfileResp, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		p = &model.EmployerProfile{UserID: userID}
	} else if err != nil {
		return nil, httperr.Internal(err)
	}
	if p.CertStatus == 1 {
		return nil, httperr.Conflict("已在审核中,请耐心等待")
	}
	if p.CertStatus == 2 {
		return nil, httperr.Conflict("已通过认证,无需重复提交")
	}
	p.CompanyName = req.CompanyName
	p.BusinessLicenseNo = req.BusinessLicenseNo
	p.BusinessLicenseImg = req.BusinessLicenseImg
	p.ContactName = req.ContactName
	p.ContactPhone = req.ContactPhone
	p.CertStatus = 1
	p.CertRemark = ""
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, httperr.Internal(err)
	}
	return toEmployerProfileResp(p), nil
}

func applyEmployerUpdates(p *model.EmployerProfile, req dto.UpsertEmployerProfileReq) {
	if req.CompanyName != nil {
		p.CompanyName = *req.CompanyName
	}
	if req.ContactName != nil {
		p.ContactName = *req.ContactName
	}
	if req.ContactPhone != nil {
		p.ContactPhone = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		p.ContactEmail = *req.ContactEmail
	}
	if req.BusinessLicenseNo != nil {
		p.BusinessLicenseNo = *req.BusinessLicenseNo
	}
	if req.BusinessLicenseImg != nil {
		p.BusinessLicenseImg = *req.BusinessLicenseImg
	}
	if req.Industry != nil {
		p.Industry = *req.Industry
	}
	if req.CompanySize != nil {
		p.CompanySize = *req.CompanySize
	}
	if req.CompanyAddress != nil {
		p.CompanyAddress = *req.CompanyAddress
	}
	if req.Intro != nil {
		p.Intro = *req.Intro
	}
}

func toEmployerProfileResp(p *model.EmployerProfile) *dto.EmployerProfileResp {
	return &dto.EmployerProfileResp{
		ID:                p.ID,
		UserID:            p.UserID,
		CompanyName:       p.CompanyName,
		ContactName:       p.ContactName,
		ContactPhone:      p.ContactPhone,
		ContactEmail:      p.ContactEmail,
		BusinessLicenseNo: p.BusinessLicenseNo,
		BusinessLicenseImg: p.BusinessLicenseImg,
		Industry:          p.Industry,
		CompanySize:       p.CompanySize,
		CompanyAddress:    p.CompanyAddress,
		Intro:             p.Intro,
		CertStatus:        p.CertStatus,
		CertRemark:        p.CertRemark,
		CertifiedAt:       p.CertifiedAt,
		Rating:            p.Rating,
		TotalJobs:         p.TotalJobs,
		CompletedOrders:   p.CompletedOrders,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// ListPendingCerts is the admin queue for employer 资质审核.
func (s *EmployerProfileService) ListPendingCerts(ctx context.Context) ([]dto.EmployerCertListItem, error) {
	rows, err := s.repo.ListByCertStatus(ctx, 1)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	out := make([]dto.EmployerCertListItem, 0, len(rows))
	for i := range rows {
		r := rows[i]
		item := dto.EmployerCertListItem{
			UserID: r.UserID, CompanyName: r.CompanyName,
			ContactName: r.ContactName, ContactPhone: r.ContactPhone,
			BusinessLicenseNo: r.BusinessLicenseNo, BusinessLicenseImg: r.BusinessLicenseImg,
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

// AuditCert is the admin's "通过/拒绝" on an employer 资质 submission.
func (s *EmployerProfileService) AuditCert(ctx context.Context, userID uint64, action int8, remark string) error {
	if action != 2 && action != 3 {
		return httperr.BadRequest("action must be 2 (pass) or 3 (reject)")
	}
	if _, err := s.repo.GetByUserID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("profile not found")
		}
		return httperr.Internal(err)
	}
	if err := s.repo.UpdateCertStatusAndStamp(ctx, userID, action, remark); err != nil {
		return httperr.Internal(err)
	}
	return nil
}
