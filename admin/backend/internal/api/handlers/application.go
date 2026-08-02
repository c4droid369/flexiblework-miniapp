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

// ApplicationHandler covers student-side apply/cancel/list, employer-side
// audit and per-job application queue. Apply is mounted under /jobs/:id
// rather than /applications so the URL is self-describing.
type ApplicationHandler struct{ svc *service.ApplicationService }

func NewApplicationHandler(svc *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

// Apply godoc
// @Summary      学生报名岗位
// @Tags         Student/Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "岗位 ID"
// @Param        body  body      dto.CreateApplicationReq  true  "留言/联系电话"
// @Success      201   {object}  response.Envelope{data=dto.ApplicationResp}
// @Router       /jobs/{id}/apply [post]
func (h *ApplicationHandler) Apply(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	jobID, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CreateApplicationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	a, err := h.svc.Apply(c.Request.Context(), uid, jobID, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, a)
}

// ListMine godoc
// @Summary      我的报名列表(学生端)
// @Tags         Student/Applications
// @Produce      json
// @Security     BearerAuth
// @Param        page   query     int  false  "页码"
// @Param        size   query     int  false  "每页大小"
// @Param        status query     int  false  "可选,1=待审核 2=已通过 3=已拒绝 4=已取消 5=已转订单"
// @Success      200    {object}  response.Envelope{data=response.PageData{list=[]dto.ApplicationResp}}
// @Router       /applications [get]
func (h *ApplicationHandler) ListMine(c *gin.Context) {
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

// Get godoc
// @Summary      获取我的报名详情
// @Tags         Student/Applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "报名 ID"
// @Success      200  {object}  response.Envelope{data=dto.ApplicationResp}
// @Router       /applications/{id} [get]
func (h *ApplicationHandler) Get(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	a, err := h.svc.Get(c.Request.Context(), id, uid)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, a)
}

// Cancel godoc
// @Summary      取消我的报名
// @Tags         Student/Applications
// @Security     BearerAuth
// @Param        id   path      int  true  "报名 ID"
// @Success      200  {object}  response.Envelope
// @Router       /applications/{id}/cancel [post]
func (h *ApplicationHandler) Cancel(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Cancel(c.Request.Context(), uid, id); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// ListByJob godoc
// @Summary      某岗位收到的报名列表(雇主端)
// @Tags         Employer/Applications
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      int  true  "岗位 ID"
// @Param        page    query     int  false  "页码"
// @Param        size    query     int  false  "每页大小"
// @Param        status  query     int  false  "可选状态筛选"
// @Success      200     {object}  response.Envelope{data=response.PageData{list=[]dto.ApplicationResp}}
// @Router       /employer/jobs/{id}/applications [get]
func (h *ApplicationHandler) ListByJob(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	jobID, ok := parseID(c)
	if !ok {
		return
	}
	page := pagination.FromGin(c)
	status := int8(0)
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = int8(v)
		}
	}
	rows, total, err := h.svc.ListByJob(c.Request.Context(), uid, jobID, page.Page, page.Size, status)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}

// Audit godoc
// @Summary      审核报名(雇主端) — 通过或拒绝
// @Description  action: 2=通过 3=拒绝. 通过后雇主可以调用 /employer/applications/{id}/hire 创建订单。
// @Tags         Employer/Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "报名 ID"
// @Param        body  body      dto.AuditApplicationReq  true  "审核结果"
// @Success      200   {object}  response.Envelope{data=dto.ApplicationResp}
// @Router       /employer/applications/{id}/audit [post]
func (h *ApplicationHandler) Audit(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.AuditApplicationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	a, err := h.svc.Audit(c.Request.Context(), uid, id, req.Action, req.Remark)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, a)
}
