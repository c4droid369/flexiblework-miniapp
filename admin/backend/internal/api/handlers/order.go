package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// OrderHandler spans the full commercial transaction. The employer side has
// an extra "hire" verb to create an order from an approved application. The
// review-svc dependency is injected so /orders/:id/review can stay on this
// handler (it keeps the URL surface compact).
type OrderHandler struct {
	svc       *service.OrderService
	reviewSvc *service.ReviewService
}

func NewOrderHandler(svc *service.OrderService, reviewSvc *service.ReviewService) *OrderHandler {
	return &OrderHandler{svc: svc, reviewSvc: reviewSvc}
}

// Hire godoc
// @Summary      雇主录用报名 → 创建订单
// @Tags         Employer/Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "报名 ID"
// @Param        body  body      dto.CreateOrderReq  true  "金额(元)"
// @Success      201   {object}  response.Envelope{data=dto.OrderResp}
// @Router       /employer/applications/{id}/hire [post]
func (h *OrderHandler) Hire(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	appID, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	o, err := h.svc.Hire(c.Request.Context(), uid, appID, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, o)
}

// Pay godoc
// @Summary      Mock 支付订单(学生端)
// @Tags         Student/Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "订单 ID"
// @Param        body  body      dto.PayOrderReq  false  "支付方式,默认 mock_wechat"
// @Success      200   {object}  response.Envelope{data=dto.OrderResp}
// @Router       /orders/{id}/pay [post]
func (h *OrderHandler) Pay(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.PayOrderReq
	_ = c.ShouldBindJSON(&req)
	o, err := h.svc.Pay(c.Request.Context(), uid, id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// Checkin godoc
// @Summary      学生上岗打卡(上传凭证)
// @Tags         Student/Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "订单 ID"
// @Param        body  body      dto.CheckinOrderReq  true  "凭证图片 URL 数组(1-9 张)"
// @Success      200   {object}  response.Envelope{data=dto.OrderResp}
// @Router       /orders/{id}/checkin [post]
func (h *OrderHandler) Checkin(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CheckinOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	o, err := h.svc.Checkin(c.Request.Context(), uid, id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// Complete godoc
// @Summary      学生提交完成
// @Tags         Student/Orders
// @Security     BearerAuth
// @Param        id   path      int  true  "订单 ID"
// @Success      200  {object}  response.Envelope{data=dto.OrderResp}
// @Router       /orders/{id}/complete [post]
func (h *OrderHandler) Complete(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	o, err := h.svc.Complete(c.Request.Context(), uid, id)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// ListMine godoc
// @Summary      我的订单列表(学生端)
// @Tags         Student/Orders
// @Produce      json
// @Security     BearerAuth
// @Param        page   query     int  false  "页码"
// @Param        size   query     int  false  "每页大小"
// @Param        status query     int  false  "可选状态筛选"
// @Success      200    {object}  response.Envelope{data=response.PageData{list=[]dto.OrderResp}}
// @Router       /orders [get]
func (h *OrderHandler) ListMine(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	page := pagination.FromGin(c)
	status := int8(0)
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = int8(v)
		}
	}
	rows, total, err := h.svc.ListByStudent(c.Request.Context(), uid, page.Page, page.Size, status)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}

// GetMine godoc
// @Summary      获取我的订单详情
// @Tags         Student/Orders
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "订单 ID"
// @Success      200  {object}  response.Envelope{data=dto.OrderResp}
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetMine(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	o, err := h.svc.Get(c.Request.Context(), id, uid, false)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// CancelMine godoc
// @Summary      学生取消订单
// @Tags         Student/Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "订单 ID"
// @Param        body  body      dto.CancelOrderReq  true  "取消原因"
// @Success      200   {object}  response.Envelope{data=dto.OrderResp}
// @Router       /orders/{id}/cancel [post]
func (h *OrderHandler) CancelMine(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CancelOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	o, err := h.svc.Cancel(c.Request.Context(), uid, id, req, false)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// Review godoc
// @Summary      学生评价雇主
// @Tags         Student/Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "订单 ID"
// @Param        body  body      dto.CreateReviewReq  true  "评价内容(1-5 星 + 标签)"
// @Success      201   {object}  response.Envelope{data=dto.ReviewResp}
// @Router       /orders/{id}/review [post]
func (h *OrderHandler) Review(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CreateReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	rv, err := h.reviewSvc.CreateFromStudent(c.Request.Context(), uid, id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, rv)
}

// ListEmployer godoc
// @Summary      雇主端订单列表
// @Tags         Employer/Orders
// @Produce      json
// @Security     BearerAuth
// @Param        page   query     int  false  "页码"
// @Param        size   query     int  false  "每页大小"
// @Param        status query     int  false  "可选状态筛选"
// @Success      200    {object}  response.Envelope{data=response.PageData{list=[]dto.OrderResp}}
// @Router       /employer/orders [get]
func (h *OrderHandler) ListEmployer(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	page := pagination.FromGin(c)
	status := int8(0)
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = int8(v)
		}
	}
	rows, total, err := h.svc.ListByEmployer(c.Request.Context(), uid, page.Page, page.Size, status)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}

// GetEmployer godoc
// @Summary      雇主获取订单详情
// @Tags         Employer/Orders
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "订单 ID"
// @Success      200  {object}  response.Envelope{data=dto.OrderResp}
// @Router       /employer/orders/{id} [get]
func (h *OrderHandler) GetEmployer(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	o, err := h.svc.Get(c.Request.Context(), id, uid, false)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// Confirm godoc
// @Summary      雇主确认完成 → 结算
// @Tags         Employer/Orders
// @Security     BearerAuth
// @Param        id   path      int  true  "订单 ID"
// @Success      200  {object}  response.Envelope{data=dto.OrderResp}
// @Router       /employer/orders/{id}/confirm [post]
func (h *OrderHandler) Confirm(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	o, err := h.svc.Confirm(c.Request.Context(), uid, id)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// CancelEmployer godoc
// @Summary      雇主取消订单
// @Tags         Employer/Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "订单 ID"
// @Param        body  body      dto.CancelOrderReq  true  "取消原因"
// @Success      200   {object}  response.Envelope{data=dto.OrderResp}
// @Router       /employer/orders/{id}/cancel [post]
func (h *OrderHandler) CancelEmployer(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CancelOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	o, err := h.svc.Cancel(c.Request.Context(), uid, id, req, true)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, o)
}

// ReviewEmployer godoc
// @Summary      雇主评价学生
// @Tags         Employer/Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "订单 ID"
// @Param        body  body      dto.CreateReviewReq  true  "评价内容"
// @Success      201   {object}  response.Envelope{data=dto.ReviewResp}
// @Router       /employer/orders/{id}/review [post]
func (h *OrderHandler) ReviewEmployer(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CreateReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	rv, err := h.reviewSvc.CreateFromEmployer(c.Request.Context(), uid, id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, rv)
}

// ListAllAdmin godoc
// @Summary      全部订单(管理端监控)
// @Tags         Admin/Orders
// @Produce      json
// @Security     BearerAuth
// @Param        page   query     int  false  "页码"
// @Param        size   query     int  false  "每页大小"
// @Param        status query     int  false  "可选状态筛选"
// @Success      200    {object}  response.Envelope{data=response.PageData{list=[]dto.OrderResp}}
// @Router       /admin/orders [get]
func (h *OrderHandler) ListAllAdmin(c *gin.Context) {
	page := pagination.FromGin(c)
	status := int8(0)
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = int8(v)
		}
	}
	rows, total, err := h.svc.ListAll(c.Request.Context(), page.Page, page.Size, status)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}
