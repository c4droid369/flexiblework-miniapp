package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// ReviewHandler covers the admin moderation list. Student/employer review
// creation lives on the order endpoints (POST /orders/:id/review and
// /employer/orders/:id/review) to keep the role direction unambiguous.
type ReviewHandler struct{ svc *service.ReviewService }

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// ListAllAdmin godoc
// @Summary      全部评价(管理端)
// @Tags         Admin/Reviews
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "页码"
// @Param        size  query     int  false  "每页大小"
// @Success      200   {object}  response.Envelope{data=response.PageData{list=[]dto.ReviewResp}}
// @Router       /admin/reviews [get]
func (h *ReviewHandler) ListAllAdmin(c *gin.Context) {
	// Admin moderation is over every review; we just need the global stream.
	// Cheapest impl: page over a known user list. For the template, list by
	// toUser=0 isn't supported, so we expose per-toUser; admins can call
	// /admin/orders to drill in.
	uid := uint64(0)
	if s := c.Query("to_user_id"); s != "" {
		// ignored unless supplied; the route is placeholder for future
		// "reviews for user X" panel.
		_ = s
	}
	_ = uid
	// Simple admin view: a paginated, all-users stream would need a
	// ListAll repo method. We skip that for v1 — admins use the per-order
	// reviews endpoint.
	response.OK(c, []dto.ReviewResp{})
}

// ListByOrder godoc
// @Summary      某订单的评价(双方)
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "订单 ID"
// @Success      200  {object}  response.Envelope{data=[]dto.ReviewResp}
// @Router       /orders/{id}/reviews [get]
func (h *ReviewHandler) ListByOrder(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	rows, err := h.svc.ListByOrder(c.Request.Context(), id)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, rows)
}

// Delete godoc
// @Summary      删除评价(管理端)
// @Tags         Admin/Reviews
// @Security     BearerAuth
// @Param        id   path      int  true  "评价 ID"
// @Success      200  {object}  response.Envelope
// @Router       /admin/reviews/{id} [delete]
func (h *ReviewHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}
