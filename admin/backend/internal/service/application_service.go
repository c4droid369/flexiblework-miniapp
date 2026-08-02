package service

import (
	"context"
	"errors"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// ApplicationService handles the student→job application lifecycle.
// It needs both UserRepo (to denormalize names) and the message + job repos
// (to send notifications and bump the apply counter).
type ApplicationService struct {
	appRepo      repository.ApplicationRepository
	jobRepo      repository.JobRepository
	userRepo     repository.UserRepository
	studentRepo  repository.StudentProfileRepository
	messageRepo  repository.MessageRepository
}

func NewApplicationService(
	appRepo repository.ApplicationRepository,
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
	studentRepo repository.StudentProfileRepository,
	messageRepo repository.MessageRepository,
) *ApplicationService {
	return &ApplicationService{appRepo: appRepo, jobRepo: jobRepo, userRepo: userRepo, studentRepo: studentRepo, messageRepo: messageRepo}
}

// Apply creates a new application. Refuses if:
//   - the job is not 招聘中
//   - the student already applied to the same job (unique index is also a
//     belt-and-braces guarantee at the DB layer)
//   - the student is not 实名认证
func (s *ApplicationService) Apply(ctx context.Context, studentUserID, jobID uint64, req dto.CreateApplicationReq) (*dto.ApplicationResp, error) {
	sp, err := s.studentRepo.GetByUserID(ctx, studentUserID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.Forbidden("请先完善学生资料")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if sp.CertStatus != 2 {
		return nil, httperr.Forbidden("请先完成学生认证再报名")
	}
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("岗位不存在")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if job.Status != 2 {
		return nil, httperr.Conflict("该岗位暂不可报名")
	}
	if job.GenderRequirement != 0 && job.GenderRequirement != sp.Gender {
		return nil, httperr.Conflict("不符合岗位性别要求")
	}
	if _, err := s.appRepo.GetByJobAndStudent(ctx, jobID, studentUserID); err == nil {
		return nil, httperr.Conflict("你已报名过该岗位")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.Internal(err)
	}
	a := &model.Application{
		JobID: jobID, StudentID: studentUserID,
		Message: req.Message, ContactPhone: req.ContactPhone,
		Status: 1,
	}
	if err := s.appRepo.Create(ctx, a); err != nil {
		return nil, httperr.Internal(err)
	}
	_ = s.jobRepo.IncApplyCount(ctx, jobID, 1)
	_ = s.notifyEmployer(ctx, a, job)
	return s.Get(ctx, a.ID, studentUserID)
}

// Cancel transitions 待审核 → 已取消. Only the student who applied may call
// this; once the employer has approved, the student must request a release
// via the order (not implemented in v1).
func (s *ApplicationService) Cancel(ctx context.Context, studentUserID, appID uint64) error {
	a, err := s.appRepo.GetByID(ctx, appID)
	if errors.Is(err, repository.ErrNotFound) {
		return httperr.NotFound("application not found")
	}
	if err != nil {
		return httperr.Internal(err)
	}
	if a.StudentID != studentUserID {
		return httperr.Forbidden("无权操作")
	}
	if a.Status != 1 {
		return httperr.Conflict("仅待审核的报名可取消")
	}
	return httperr.WriteRaw(s.appRepo.UpdateStatus(ctx, appID, 4, "学生主动取消"))
}

// Audit is the employer's approve/reject. Approving flips application to
// status 2 (已通过) — order creation is a separate call (Hire) to give the
// employer a chance to set the price.
func (s *ApplicationService) Audit(ctx context.Context, employerUserID, appID uint64, action int8, remark string) (*dto.ApplicationResp, error) {
	if action != 2 && action != 3 {
		return nil, httperr.BadRequest("action must be 2 (pass) or 3 (reject)")
	}
	a, err := s.appRepo.GetByID(ctx, appID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("application not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	job, err := s.jobRepo.GetByID(ctx, a.JobID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("job not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if job.EmployerID != employerUserID {
		return nil, httperr.Forbidden("无权操作")
	}
	if a.Status != 1 {
		return nil, httperr.Conflict("仅待审核的报名可处理")
	}
	if err := s.appRepo.UpdateStatus(ctx, appID, action, remark); err != nil {
		return nil, httperr.Internal(err)
	}
	_ = s.notifyStudent(ctx, a, action, job)
	return s.Get(ctx, appID, a.StudentID)
}

// MarkHired flips a 2 (已通过) application to 5 (已转订单). Called from
// OrderService after a hire is committed.
func (s *ApplicationService) MarkHired(ctx context.Context, appID uint64) error {
	return s.appRepo.UpdateStatus(ctx, appID, 5, "已录用,生成订单")
}

// Get returns one application. CallerID is used to decide whether to expose
// the student's phone number (only to the job's employer).
func (s *ApplicationService) Get(ctx context.Context, appID, callerID uint64) (*dto.ApplicationResp, error) {
	a, err := s.appRepo.GetByID(ctx, appID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("application not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if a.StudentID != callerID {
		// Non-owners only see their own — the employer uses ListByJob instead.
		return nil, httperr.Forbidden("无权查看")
	}
	return s.toResp(ctx, a, false)
}

// ListByStudent is the student's "我报名的" list.
func (s *ApplicationService) ListByStudent(ctx context.Context, studentUserID uint64, page, size int, status int8) ([]dto.ApplicationResp, int64, error) {
	rows, total, err := s.appRepo.ListByStudent(ctx, page, size, studentUserID, status)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.ApplicationResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i], false)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// ListByJob is the employer's "收到的报名" list for one of their jobs.
func (s *ApplicationService) ListByJob(ctx context.Context, employerUserID, jobID uint64, page, size int, status int8) ([]dto.ApplicationResp, int64, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, 0, httperr.NotFound("job not found")
	}
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	if job.EmployerID != employerUserID {
		return nil, 0, httperr.Forbidden("无权查看")
	}
	rows, total, err := s.appRepo.ListByJob(ctx, page, size, jobID, status)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.ApplicationResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i], true)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// CountByJobAndStatuses exposes the underlying count so OrderService can
// decide whether the job is "已招满" after creating a new order.
func (s *ApplicationService) CountByJobAndStatuses(ctx context.Context, jobID uint64, statuses []int8) (int64, error) {
	return s.appRepo.CountByStatuses(ctx, jobID, statuses)
}

func (s *ApplicationService) toResp(ctx context.Context, a *model.Application, revealPhone bool) (*dto.ApplicationResp, error) {
	r := &dto.ApplicationResp{
		ID: a.ID, JobID: a.JobID, StudentID: a.StudentID,
		Message: a.Message, ContactPhone: a.ContactPhone,
		Status: a.Status, AuditRemark: a.AuditRemark, AuditedAt: a.AuditedAt,
		CreatedAt: a.CreatedAt,
	}
	if u, err := s.userRepo.GetByID(ctx, a.StudentID); err == nil {
		r.StudentName = nicknameOrUsername(u)
	}
	if sp, err := s.studentRepo.GetByUserID(ctx, a.StudentID); err == nil {
		r.StudentSchool = sp.School
	}
	if j, err := s.jobRepo.GetByID(ctx, a.JobID); err == nil {
		r.JobTitle = j.Title
	}
	return r, nil
}

func (s *ApplicationService) notifyEmployer(ctx context.Context, a *model.Application, job *model.Job) error {
	sp, err := s.studentRepo.GetByUserID(ctx, a.StudentID)
	if err != nil {
		return nil
	}
	name := ""
	if u, err := s.userRepo.GetByID(ctx, a.StudentID); err == nil {
		name = nicknameOrUsername(u)
	}
	msg := &model.Message{
		UserID: job.EmployerID, Type: 2,
		Title:   "新报名提醒",
		Content: name + " 报名了岗位《" + job.Title + "》" + sp.School,
		Link:    "/pages/employer/applications/index?job_id=" + itoa(job.ID),
	}
	return s.messageRepo.Create(ctx, msg)
}

func (s *ApplicationService) notifyStudent(ctx context.Context, a *model.Application, action int8, job *model.Job) error {
	var title, content string
	if action == 2 {
		title = "报名审核通过"
		content = "你报名的岗位《" + job.Title + "》已被雇主通过"
	} else {
		title = "报名未通过"
		content = "很遗憾,你报名的岗位《" + job.Title + "》未通过"
	}
	msg := &model.Message{
		UserID: a.StudentID, Type: 2,
		Title:   title,
		Content: content,
		Link:    "/pages/student/applications/index",
	}
	return s.messageRepo.Create(ctx, msg)
}

func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
