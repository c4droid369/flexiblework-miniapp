package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/repository"
)

// ReviewService. The role (student→employer or employer→student) is derived
// from the order, not the request body, so the wire surface is just rating
// + content + tags.
type ReviewService struct {
	reviewRepo repository.ReviewRepository
	orderRepo  repository.OrderRepository
	userRepo   repository.UserRepository
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	orderRepo repository.OrderRepository,
	userRepo repository.UserRepository,
) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, orderRepo: orderRepo, userRepo: userRepo}
}

// CreateFromStudent is the student's "评价雇主" call.
func (s *ReviewService) CreateFromStudent(ctx context.Context, studentUserID, orderID uint64, req dto.CreateReviewReq) (*dto.ReviewResp, error) {
	return s.create(ctx, orderID, studentUserID, true, req)
}

// CreateFromEmployer is the employer's "评价学生" call.
func (s *ReviewService) CreateFromEmployer(ctx context.Context, employerUserID, orderID uint64, req dto.CreateReviewReq) (*dto.ReviewResp, error) {
	return s.create(ctx, orderID, employerUserID, false, req)
}

func (s *ReviewService) create(ctx context.Context, orderID, callerID uint64, byStudent bool, req dto.CreateReviewReq) (*dto.ReviewResp, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.NotFound("order not found")
	}
	if err != nil {
		return nil, httperr.Internal(err)
	}
	if o.Status != 5 {
		return nil, httperr.Conflict("仅已结算订单可评价")
	}
	var fromUserID, toUserID uint64
	var role int8
	if byStudent {
		if o.StudentID != callerID {
			return nil, httperr.Forbidden("无权评价")
		}
		fromUserID = o.StudentID
		toUserID = o.EmployerID
		role = 1
	} else {
		if o.EmployerID != callerID {
			return nil, httperr.Forbidden("无权评价")
		}
		fromUserID = o.EmployerID
		toUserID = o.StudentID
		role = 2
	}
	if existing, err := s.reviewRepo.GetByOrderAndFrom(ctx, orderID, fromUserID); err == nil && existing != nil {
		return nil, httperr.Conflict("你已评价过该订单")
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, httperr.Internal(err)
	}
	tagsJSON, _ := json.Marshal(req.Tags)
	rv := &model.Review{
		OrderID: orderID, FromUserID: fromUserID, ToUserID: toUserID,
		Role: role, Rating: req.Rating, Content: req.Content, Tags: string(tagsJSON),
	}
	if err := s.reviewRepo.Create(ctx, rv); err != nil {
		return nil, httperr.Internal(err)
	}
	return s.toResp(ctx, rv)
}

// ListByToUser is the public "TA 的评价" list. The owner of a profile
// (student or employer) views their own ratings; the mini-program detail
// pages call this with toUserID set to the profile owner.
func (s *ReviewService) ListByToUser(ctx context.Context, toUserID uint64, page, size int) ([]dto.ReviewResp, int64, error) {
	rows, total, err := s.reviewRepo.ListByToUser(ctx, page, size, toUserID)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	out := make([]dto.ReviewResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *r)
	}
	return out, total, nil
}

// ListByOrder is the order detail page's "双方评价" section.
func (s *ReviewService) ListByOrder(ctx context.Context, orderID uint64) ([]dto.ReviewResp, error) {
	rows, err := s.reviewRepo.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	out := make([]dto.ReviewResp, 0, len(rows))
	for i := range rows {
		r, err := s.toResp(ctx, &rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, nil
}

// Delete is admin-only moderation.
func (s *ReviewService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.reviewRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("review not found")
		}
		return httperr.Internal(err)
	}
	return httperr.WriteRaw(s.reviewRepo.Delete(ctx, id))
}

func (s *ReviewService) toResp(ctx context.Context, rv *model.Review) (*dto.ReviewResp, error) {
	r := &dto.ReviewResp{
		ID: rv.ID, OrderID: rv.OrderID,
		FromUserID: rv.FromUserID, ToUserID: rv.ToUserID,
		Role: rv.Role, Rating: rv.Rating, Content: rv.Content,
		CreatedAt: rv.CreatedAt,
	}
	if rv.Tags != "" {
		var arr []string
		_ = json.Unmarshal([]byte(rv.Tags), &arr)
		r.Tags = arr
	}
	if u, err := s.userRepo.GetByID(ctx, rv.FromUserID); err == nil {
		r.FromName = nicknameOrUsername(u)
		r.FromAvatar = u.Avatar
	}
	return r, nil
}
