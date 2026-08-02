package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// OrderService owns the commercial transaction: hire → pay → checkin →
// complete → confirm. It stitches together Job, Application, Profile and
// Message services to keep denormalized counters and notifications in sync.
type OrderService struct {
	db              *gorm.DB
	orderRepo       repository.OrderRepository
	appRepo         repository.ApplicationRepository
	appSvc          *ApplicationService
	jobRepo         repository.JobRepository
	jobSvc          *JobService
	userRepo        repository.UserRepository
	employerRepo    repository.EmployerProfileRepository
	messageRepo     repository.MessageRepository
}

func NewOrderService(
	db *gorm.DB,
	orderRepo repository.OrderRepository,
	appRepo repository.ApplicationRepository,
	appSvc *ApplicationService,
	jobRepo repository.JobRepository,
	jobSvc *JobService,
	userRepo repository.UserRepository,
	employerRepo repository.EmployerProfileRepository,
	messageRepo repository.MessageRepository,
) *OrderService {
	return &OrderService{
		db: db, orderRepo: orderRepo, appRepo: appRepo, appSvc: appSvc,
		jobRepo: jobRepo, jobSvc: jobSvc, userRepo: userRepo,
		employerRepo: employerRepo, messageRepo: messageRepo,
	}
}

// Hire creates an order from an approved application. Transitions:
//   application 2 (已通过) → 5 (已转订单)
//   new order 1 (待支付)
//
// Side effects: if the number of hired applications reaches job.recruit_count,
// the job is flipped to status 5 (已招满).
func (s *OrderService) Hire(ctx context.Context, employerUserID, applicationID uint64, req dto.CreateOrderReq) (*dto.OrderResp, error) {
	app, err := s.appRepo.GetByID(ctx, applicationID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("application not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	job, err := s.jobRepo.GetByID(ctx, app.JobID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("job not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if job.EmployerID != employerUserID {
		return nil, httperr.Forbidden("无权操作")
	}
	if app.Status != 2 {
		return nil, httperr.Conflict("仅已通过的报名可录用")
	}
	if job.Status != 2 {
		return nil, httperr.Conflict("岗位不在招聘中")
	}
	// One order per application — refuse if an order already exists.
	if existing, err := s.orderRepo.GetByApplication(ctx, applicationID); err == nil && existing != nil {
		return nil, httperr.Conflict("该报名已生成订单")
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.Internal(err)
	}

	o := &model.Order{
		OrderNo:       generateOrderNo(),
		JobID:         job.ID,
		ApplicationID: app.ID,
		EmployerID:    employerUserID,
		StudentID:     app.StudentID,
		Amount:        req.Amount,
		Status:        1,
		WorkProof:     "[]",
	}
	if err := s.orderRepo.Create(ctx, o); err != nil {
		return nil, httperr.Internal(err)
	}
	if err := s.appRepo.UpdateStatus(ctx, applicationID, 5, "已录用,生成订单"); err != nil {
		// Order row is created; the partial state is logged but not rolled
		// back — a re-attempt by the employer will hit the duplicate guard.
		return nil, httperr.Internal(err)
	}

	// Recruit-count check → "已招满".
	hired, err := s.appRepo.CountByStatuses(ctx, job.ID, []int8{5})
	if err == nil && hired >= int64(job.RecruitCount) {
		_ = s.jobRepo.UpdateStatus(ctx, job.ID, 5, "招满自动结束")
	}
	_ = s.notifyStudent(ctx, o, "新订单待支付", "雇主已录用你,订单等待支付")
	return s.Get(ctx, o.ID, employerUserID, true)
}

// Pay is a mock payment. The state machine flips 1 → 2.
func (s *OrderService) Pay(ctx context.Context, studentUserID, orderID uint64, req dto.PayOrderReq) (*dto.OrderResp, error) {
	o, err := s.loadForStudent(ctx, orderID, studentUserID)
	if err != nil {
		return nil, err
	}
	if o.Status != 1 {
		return nil, httperr.Conflict("仅待支付订单可支付")
	}
	method := req.Method
	if method == "" {
		method = "mock_wechat"
	}
	if err := s.orderRepo.UpdateFields(ctx, orderID, map[string]any{
		"status":     2,
		"pay_method": method,
		"paid_at":    gorm.Expr("NOW()"),
	}); err != nil {
		return nil, httperr.Internal(err)
	}
	_ = s.notifyEmployer(ctx, o, "订单已支付", "学生已支付订单,可以开始工作")
	return s.Get(ctx, orderID, studentUserID, true)
}

// Checkin is the student's "我到岗了" action. Flips 2 → 3.
func (s *OrderService) Checkin(ctx context.Context, studentUserID, orderID uint64, req dto.CheckinOrderReq) (*dto.OrderResp, error) {
	o, err := s.loadForStudent(ctx, orderID, studentUserID)
	if err != nil {
		return nil, err
	}
	if o.Status != 2 {
		return nil, httperr.Conflict("仅已支付订单可打卡")
	}
	proof, err := json.Marshal(req.WorkProof)
	if err != nil {
		return nil, httperr.BadRequest("work_proof invalid")
	}
	if err := s.orderRepo.UpdateFields(ctx, orderID, map[string]any{
		"status":     3,
		"started_at": gorm.Expr("NOW()"),
		"work_proof": string(proof),
	}); err != nil {
		return nil, httperr.Internal(err)
	}
	_ = s.notifyEmployer(ctx, o, "学生已上岗", "学生已到达岗位并提交了凭证")
	return s.Get(ctx, orderID, studentUserID, true)
}

// Complete is the student's "我做完了" action. Flips 3 → 4 and waits for
// the employer to confirm (or dispute).
func (s *OrderService) Complete(ctx context.Context, studentUserID, orderID uint64) (*dto.OrderResp, error) {
	o, err := s.loadForStudent(ctx, orderID, studentUserID)
	if err != nil {
		return nil, err
	}
	if o.Status != 3 {
		return nil, httperr.Conflict("仅进行中的订单可提交完成")
	}
	if err := s.orderRepo.UpdateFields(ctx, orderID, map[string]any{
		"status":       4,
		"completed_at": gorm.Expr("NOW()"),
	}); err != nil {
		return nil, httperr.Internal(err)
	}
	_ = s.notifyEmployer(ctx, o, "学生已提交完成", "学生已完成工作,请确认结算")
	return s.Get(ctx, orderID, studentUserID, true)
}

// Confirm is the employer's "OK, 结算" action. Flips 4 → 5 and bumps the
// employer's completed_orders counter.
func (s *OrderService) Confirm(ctx context.Context, employerUserID, orderID uint64) (*dto.OrderResp, error) {
	o, err := s.loadForEmployer(ctx, orderID, employerUserID)
	if err != nil {
		return nil, err
	}
	if o.Status != 4 {
		return nil, httperr.Conflict("仅待确认完成的订单可结算")
	}
	if err := s.orderRepo.UpdateFields(ctx, orderID, map[string]any{
		"status":       5,
		"confirmed_at": gorm.Expr("NOW()"),
		"settled_at":   gorm.Expr("NOW()"),
	}); err != nil {
		return nil, httperr.Internal(err)
	}
	_ = s.employerRepo.IncCompletedOrders(ctx, employerUserID, 1)
	_ = s.notifyStudent(ctx, o, "订单已结算", "雇主已确认完成,订单已结算,快去评价吧")
	return s.Get(ctx, orderID, employerUserID, true)
}

// Cancel transitions 1|2 → 6. If the order was already paid, status moves
// to 7 (已退款) immediately. Either party can cancel, but only before
// checkin.
func (s *OrderService) Cancel(ctx context.Context, callerID, orderID uint64, req dto.CancelOrderReq, byEmployer bool) (*dto.OrderResp, error) {
	var o *model.Order
	var err error
	if byEmployer {
		o, err = s.loadForEmployer(ctx, orderID, callerID)
	} else {
		o, err = s.loadForStudent(ctx, orderID, callerID)
	}
	if err != nil {
		return nil, err
	}
	if o.Status != 1 && o.Status != 2 {
		return nil, httperr.Conflict("仅待支付/已支付订单可取消")
	}
	newStatus := 6
	if o.Status == 2 {
		newStatus = 7 // mock refund: paid order → 已退款
	}
	if err := s.orderRepo.UpdateFields(ctx, orderID, map[string]any{
		"status":        newStatus,
		"cancel_reason": req.Reason,
	}); err != nil {
		return nil, httperr.Internal(err)
	}
	// Notify the other side.
	if byEmployer {
		_ = s.notifyStudent(ctx, o, "订单已取消", "雇主取消了订单:"+req.Reason)
	} else {
		_ = s.notifyEmployer(ctx, o, "订单已取消", "学生取消了订单:"+req.Reason)
	}
	return s.Get(ctx, orderID, callerID, true)
}

// Get returns one order. revealToOtherSide controls whether the caller is
// the order's other party (and would not normally be able to read it). The
// default is false; admin overrides.
func (s *OrderService) Get(ctx context.Context, orderID, callerID uint64, isAdmin bool) (*dto.OrderResp, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("order not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if !isAdmin && o.StudentID != callerID && o.EmployerID != callerID {
		return nil, httperr.Forbidden("无权查看")
	}
	return s.toResp(ctx, o)
}

// ListByStudent surfaces the student's orders, optionally filtered by status.
func (s *OrderService) ListByStudent(ctx context.Context, studentUserID uint64, page, size int, status int8) ([]dto.OrderResp, int64, error) {
	rows, total, err := s.orderRepo.ListByStudent(ctx, page, size, studentUserID, status)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.OrderResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// ListByEmployer surfaces the employer's orders.
func (s *OrderService) ListByEmployer(ctx context.Context, employerUserID uint64, page, size int, status int8) ([]dto.OrderResp, int64, error) {
	rows, total, err := s.orderRepo.ListByEmployer(ctx, page, size, employerUserID, status)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.OrderResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// ListAll is the admin's view of all orders.
func (s *OrderService) ListAll(ctx context.Context, page, size int, status int8) ([]dto.OrderResp, int64, error) {
	rows, total, err := s.orderRepo.ListAll(ctx, page, size, status)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.OrderResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

func (s *OrderService) loadForStudent(ctx context.Context, orderID, studentID uint64) (*model.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("order not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if o.StudentID != studentID {
		return nil, httperr.Forbidden("无权操作")
	}
	return o, nil
}

func (s *OrderService) loadForEmployer(ctx context.Context, orderID, employerID uint64) (*model.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("order not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if o.EmployerID != employerID {
		return nil, httperr.Forbidden("无权操作")
	}
	return o, nil
}

func (s *OrderService) toResp(ctx context.Context, o *model.Order) (*dto.OrderResp, error) {
	r := &dto.OrderResp{
		ID: o.ID, OrderNo: o.OrderNo,
		JobID: o.JobID, ApplicationID: o.ApplicationID,
		EmployerID: o.EmployerID, StudentID: o.StudentID,
		Amount: o.Amount, Status: o.Status,
		PayMethod: o.PayMethod,
		PaidAt: o.PaidAt, StartedAt: o.StartedAt,
		CompletedAt: o.CompletedAt, ConfirmedAt: o.ConfirmedAt, SettledAt: o.SettledAt,
		CancelReason: o.CancelReason,
		CreatedAt:    o.CreatedAt,
	}
	if o.WorkProof != "" {
		var arr []string
		if err := json.Unmarshal([]byte(o.WorkProof), &arr); err == nil {
			r.WorkProof = arr
		}
	}
	if j, err := s.jobRepo.GetByID(ctx, o.JobID); err == nil {
		r.JobTitle = j.Title
	}
	if u, err := s.userRepo.GetByID(ctx, o.EmployerID); err == nil {
		r.EmployerName = nicknameOrUsername(u)
	}
	if u, err := s.userRepo.GetByID(ctx, o.StudentID); err == nil {
		r.StudentName = nicknameOrUsername(u)
	}
	return r, nil
}

func (s *OrderService) notifyStudent(ctx context.Context, o *model.Order, title, content string) error {
	return s.messageRepo.Create(ctx, &model.Message{
		UserID: o.StudentID, Type: 3,
		Title: title, Content: content,
		Link: "/pages/student/orders/detail?id=" + itoa(o.ID),
	})
}

func (s *OrderService) notifyEmployer(ctx context.Context, o *model.Order, title, content string) error {
	return s.messageRepo.Create(ctx, &model.Message{
		UserID: o.EmployerID, Type: 3,
		Title: title, Content: content,
		Link: "/pages/employer/orders/detail?id=" + itoa(o.ID),
	})
}

// generateOrderNo is "CG" + yyyyMMdd + 6-digit zero-padded daily counter.
// Counter is per-process (in-memory); under multi-instance deployment swap
// for a DB sequence. Adequate for the template's single-instance mode.
var orderSeq uint64

func generateOrderNo() string {
	now := time.Now()
	seq := orderSeq
	orderSeq++
	return fmt.Sprintf("CG%s%06d", now.Format("20060102"), seq%1000000)
}
