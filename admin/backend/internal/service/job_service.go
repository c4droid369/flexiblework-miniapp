package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// JobService is the public + employer/agent + admin view of the gig catalogue.
// Cross-entity lookups (employer, category) are denormalized into the DTO
// here so the mini-program gets a single round-trip per page.
//
// employer_id on a job is the *user_id* of whoever posted it — either an
// employer OR a campus agent. Both share this service; the cert gate at
// Create time checks whichever profile is present.
type JobService struct {
	jobRepo      repository.JobRepository
	categoryRepo repository.CategoryRepository
	userRepo     repository.UserRepository
	employerRepo repository.EmployerProfileRepository
	agentRepo    repository.AgentProfileRepository
}

func NewJobService(
	jobRepo repository.JobRepository,
	categoryRepo repository.CategoryRepository,
	userRepo repository.UserRepository,
	employerRepo repository.EmployerProfileRepository,
	agentRepo repository.AgentProfileRepository,
) *JobService {
	return &JobService{
		jobRepo: jobRepo, categoryRepo: categoryRepo, userRepo: userRepo,
		employerRepo: employerRepo, agentRepo: agentRepo,
	}
}

// Create persists a new job in status 1 (待审核) and bumps the poster's
// total_jobs counter (employer OR agent). It refuses the request if the
// caller has no certified business profile — admins decide which role to
// approve via the cert audit flow.
func (s *JobService) Create(ctx context.Context, employerUserID uint64, req dto.CreateJobReq) (*dto.JobResp, error) {
	posterKind, err := s.resolvePoster(ctx, employerUserID)
	if err != nil {
		return nil, err
	}
	if posterKind == "" {
		return nil, httperr.Forbidden("请先完成资质认证(雇主或校园代理)")
	}
	j := &model.Job{
		EmployerID:        employerUserID,
		CategoryID:        req.CategoryID,
		Title:             req.Title,
		Cover:             req.Cover,
		Description:       req.Description,
		Requirements:      req.Requirements,
		SalaryType:        req.SalaryType,
		SalaryMin:         req.SalaryMin,
		SalaryMax:         req.SalaryMax,
		SalaryUnit:        req.SalaryUnit,
		Location:          req.Location,
		WorkDateType:      req.WorkDateType,
		WorkDateStart:     req.WorkDateStart,
		WorkDateEnd:       req.WorkDateEnd,
		WorkTimeStart:     req.WorkTimeStart,
		WorkTimeEnd:       req.WorkTimeEnd,
		RecruitCount:      req.RecruitCount,
		GenderRequirement: req.GenderRequirement,
		SettlementType:    req.SettlementType,
		Status:            1,
	}
	if j.WorkDateType == 0 {
		j.WorkDateType = 1
	}
	if j.SettlementType == 0 {
		j.SettlementType = 1
	}
	if j.GenderRequirement == 0 {
		j.GenderRequirement = 0
	}
	if err := s.jobRepo.Create(ctx, j); err != nil {
		return nil, httperr.Internal(err)
	}
	if posterKind == "agent" {
		_ = s.agentRepo.IncTotalJobs(ctx, employerUserID, 1)
	} else {
		_ = s.employerRepo.IncTotalJobs(ctx, employerUserID, 1)
	}
	return s.Get(ctx, j.ID, false)
}

// resolvePoster returns "employer" / "agent" / "" (no business profile)
// and enforces the cert gate. Caller-side check is centralized here so
// every job-mutating verb (Create, Offline, …) can share the same logic.
func (s *JobService) resolvePoster(ctx context.Context, userID uint64) (string, error) {
	if ep, err := s.employerRepo.GetByUserID(ctx, userID); err == nil {
		if ep.CertStatus != 2 {
			return "", httperr.Forbidden("资质未通过认证,无法发布岗位")
		}
		return "employer", nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return "", httperr.Internal(err)
	}
	ap, err := s.agentRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil // no business profile — caller will return Forbidden
		}
		return "", httperr.Internal(err)
	}
	if ap.CertStatus != 2 {
		return "", httperr.Forbidden("资质未通过认证,无法发布岗位")
	}
	return "agent", nil
}

// Update mutates an existing job. Only the owning employer may call this,
// and only while the job is still 待审核 or 招聘中.
func (s *JobService) Update(ctx context.Context, employerUserID, jobID uint64, req dto.UpdateJobReq) (*dto.JobResp, error) {
	j, err := s.jobRepo.GetByID(ctx, jobID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("job not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if j.EmployerID != employerUserID {
		return nil, httperr.Forbidden("无权修改该岗位")
	}
	if j.Status != 1 && j.Status != 2 {
		return nil, httperr.Conflict("当前状态不可修改")
	}
	if req.CategoryID != nil {
		j.CategoryID = *req.CategoryID
	}
	if req.Title != nil {
		j.Title = *req.Title
	}
	if req.Cover != nil {
		j.Cover = *req.Cover
	}
	if req.Description != nil {
		j.Description = *req.Description
	}
	if req.Requirements != nil {
		j.Requirements = *req.Requirements
	}
	if req.SalaryType != nil {
		j.SalaryType = *req.SalaryType
	}
	if req.SalaryMin != nil {
		j.SalaryMin = *req.SalaryMin
	}
	if req.SalaryMax != nil {
		j.SalaryMax = *req.SalaryMax
	}
	if req.SalaryUnit != nil {
		j.SalaryUnit = *req.SalaryUnit
	}
	if req.Location != nil {
		j.Location = *req.Location
	}
	if req.WorkDateType != nil {
		j.WorkDateType = *req.WorkDateType
	}
	if req.WorkDateStart != nil {
		j.WorkDateStart = req.WorkDateStart
	}
	if req.WorkDateEnd != nil {
		j.WorkDateEnd = req.WorkDateEnd
	}
	if req.WorkTimeStart != nil {
		j.WorkTimeStart = *req.WorkTimeStart
	}
	if req.WorkTimeEnd != nil {
		j.WorkTimeEnd = *req.WorkTimeEnd
	}
	if req.RecruitCount != nil {
		j.RecruitCount = *req.RecruitCount
	}
	if req.GenderRequirement != nil {
		j.GenderRequirement = *req.GenderRequirement
	}
	if req.SettlementType != nil {
		j.SettlementType = *req.SettlementType
	}
	if err := s.jobRepo.Update(ctx, j); err != nil {
		return nil, httperr.Internal(err)
	}
	return s.Get(ctx, jobID, false)
}

func (s *JobService) Delete(ctx context.Context, employerUserID, jobID uint64) error {
	j, err := s.jobRepo.GetByID(ctx, jobID)
	if errors.Is(err, repository.ErrNotFound) {
		return httperr.NotFound("job not found")
	}
	if err != nil {
		return httperr.Internal(err)
	}
	if j.EmployerID != employerUserID {
		return httperr.Forbidden("无权删除该岗位")
	}
	if j.Status == 2 {
		return httperr.Conflict("招聘中的岗位请先下架")
	}
	return httperr.WriteRaw(s.jobRepo.Delete(ctx, jobID))
}

// Offline transitions 招聘中 → 已下架. Only the owning employer may trigger.
func (s *JobService) Offline(ctx context.Context, employerUserID, jobID uint64) error {
	j, err := s.jobRepo.GetByID(ctx, jobID)
	if errors.Is(err, repository.ErrNotFound) {
		return httperr.NotFound("job not found")
	}
	if err != nil {
		return httperr.Internal(err)
	}
	if j.EmployerID != employerUserID {
		return httperr.Forbidden("无权操作该岗位")
	}
	if j.Status != 2 {
		return httperr.Conflict("仅招聘中岗位可下架")
	}
	return httperr.WriteRaw(s.jobRepo.UpdateStatus(ctx, jobID, 3, ""))
}

// Get returns one job, optionally bumping its view count. The view bump is
// off for the employer's own view to keep their dashboard counts honest.
func (s *JobService) Get(ctx context.Context, jobID uint64, bumpView bool) (*dto.JobResp, error) {
	j, err := s.jobRepo.GetByID(ctx, jobID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("job not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if bumpView && j.Status == 2 {
		_ = s.jobRepo.IncViewCount(ctx, jobID)
		j.ViewCount++
	}
	return s.toResp(ctx, j)
}

// List is the public-facing job listing (defaults to status=2 招聘中).
func (s *JobService) List(ctx context.Context, page, size int, f repository.JobListFilter) ([]dto.JobResp, int64, error) {
	f.OnlyActive = true // public listing is always "招聘中"
	rows, total, err := s.jobRepo.List(ctx, page, size, f)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.JobResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// ListByEmployer is the employer's "我发布的" listing — no status filter
// because the dashboard surfaces every state.
func (s *JobService) ListByEmployer(ctx context.Context, employerUserID uint64, page, size int) ([]dto.JobResp, int64, error) {
	f := repository.JobListFilter{EmployerID: employerUserID}
	rows, total, err := s.jobRepo.List(ctx, page, size, f)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.JobResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// Audit is the admin's "通过/拒绝" action.
func (s *JobService) Audit(ctx context.Context, jobID uint64, action int8, remark string) (*dto.JobResp, error) {
	if action != 2 && action != 4 {
		return nil, httperr.BadRequest("action must be 2 (pass) or 4 (reject)")
	}
	j, err := s.jobRepo.GetByID(ctx, jobID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("job not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if j.Status != 1 {
		return nil, httperr.Conflict("仅待审核岗位可审核")
	}
	if err := s.jobRepo.UpdateStatus(ctx, jobID, action, remark); err != nil {
		return nil, httperr.Internal(err)
	}
	return s.Get(ctx, jobID, false)
}

// ListPending is the admin queue. Only 待审核 jobs are surfaced.
func (s *JobService) ListPending(ctx context.Context, page, size int) ([]dto.JobResp, int64, error) {
	f := repository.JobListFilter{Status: 1}
	rows, total, err := s.jobRepo.List(ctx, page, size, f)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.JobResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// MarkFull flips a job to status 5 (已招满). Called by the order service when
// the number of hired applications reaches job.recruit_count.
func (s *JobService) MarkFull(ctx context.Context, jobID uint64) error {
	return s.jobRepo.UpdateStatus(ctx, jobID, 5, "")
}

// toResp denormalizes employer + category into the wire DTO. Failed lookups
// fall back to empty strings — never block the list response on missing
// joined rows.
func (s *JobService) toResp(ctx context.Context, j *model.Job) (*dto.JobResp, error) {
	r := &dto.JobResp{
		ID: j.ID, EmployerID: j.EmployerID, CategoryID: j.CategoryID,
		Title: j.Title, Cover: j.Cover, Description: j.Description, Requirements: j.Requirements,
		SalaryType: j.SalaryType, SalaryMin: j.SalaryMin, SalaryMax: j.SalaryMax,
		SalaryUnit: j.SalaryUnit, SalaryText: formatSalary(j.SalaryMin, j.SalaryMax, j.SalaryUnit),
		Location: j.Location, WorkDateType: j.WorkDateType,
		WorkDateStart: j.WorkDateStart, WorkDateEnd: j.WorkDateEnd,
		WorkTimeStart: j.WorkTimeStart, WorkTimeEnd: j.WorkTimeEnd,
		RecruitCount: j.RecruitCount, GenderRequirement: j.GenderRequirement,
		SettlementType: j.SettlementType, Status: j.Status,
		AuditRemark: j.AuditRemark, AuditedAt: j.AuditedAt,
		ViewCount: j.ViewCount, ApplyCount: j.ApplyCount,
		CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt,
	}
	if c, err := s.categoryRepo.GetByID(ctx, j.CategoryID); err == nil {
		r.CategoryName = c.Name
	}
	if u, err := s.userRepo.GetByID(ctx, j.EmployerID); err == nil {
		r.EmployerName = nicknameOrUsername(u)
	}
	return r, nil
}

func formatSalary(min, max float64, unit string) string {
	if unit == "" {
		unit = "元"
	}
	if max > 0 && min != max {
		return strconv.FormatFloat(min, 'f', 0, 64) + "-" + strconv.FormatFloat(max, 'f', 0, 64) + unit
	}
	return strconv.FormatFloat(min, 'f', 0, 64) + unit
}

func nicknameOrUsername(u *model.User) string {
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Username
}

// guard against fmt unused when this file is the only consumer.
var _ = fmt.Sprintf
