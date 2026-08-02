package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/repository"
	"github.com/admin-template/backend/internal/service"
)

// JobHandler splits its endpoints across public, employer-side, and admin.
// Permission codes are applied in router.go.
type JobHandler struct{ svc *service.JobService }

func NewJobHandler(svc *service.JobService) *JobHandler {
	return &JobHandler{svc: svc}
}

// ListPublic godoc
// @Summary      公开岗位列表(仅招聘中)
// @Tags         Jobs
// @Produce      json
// @Param        page          query     int     false  "页码 (默认 1)"
// @Param        size          query     int     false  "每页大小 (默认 10)"
// @Param        category_id   query     int     false  "分类筛选"
// @Param        location      query     string  false  "工作地点(模糊)"
// @Param        salary_min    query     number  false  "最低薪资(>=)"
// @Param        salary_max    query     number  false  "最高薪资(<=)"
// @Param        keyword       query     string  false  "标题/描述模糊搜索"
// @Success      200           {object}  response.Envelope{data=response.PageData{list=[]dto.JobResp}}
// @Router       /jobs [get]
func (h *JobHandler) ListPublic(c *gin.Context) {
	page := pagination.FromGin(c)
	f := parseJobFilter(c)
	rows, total, err := h.svc.List(c.Request.Context(), page.Page, page.Size, f)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}

// GetPublic godoc
// @Summary      公开岗位详情(自动 +1 view_count)
// @Tags         Jobs
// @Produce      json
// @Param        id   path      int  true  "岗位 ID"
// @Success      200  {object}  response.Envelope{data=dto.JobResp}
// @Router       /jobs/{id} [get]
func (h *JobHandler) GetPublic(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	j, err := h.svc.Get(c.Request.Context(), id, true)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, j)
}

// Create godoc
// @Summary      雇主发布岗位
// @Description  新岗位状态为"待审核",管理端通过后变成"招聘中"。
// @Tags         Employer/Jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateJobReq  true  "新岗位"
// @Success      201   {object}  response.Envelope{data=dto.JobResp}
// @Router       /employer/jobs [post]
func (h *JobHandler) Create(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	var req dto.CreateJobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	j, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, j)
}

// Update godoc
// @Summary      雇主更新自己的岗位
// @Tags         Employer/Jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "岗位 ID"
// @Param        body  body      dto.UpdateJobReq  true  "更新字段"
// @Success      200   {object}  response.Envelope{data=dto.JobResp}
// @Router       /employer/jobs/{id} [put]
func (h *JobHandler) Update(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateJobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	j, err := h.svc.Update(c.Request.Context(), uid, id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, j)
}

// Delete godoc
// @Summary      雇主删除自己的岗位
// @Tags         Employer/Jobs
// @Security     BearerAuth
// @Param        id   path      int  true  "岗位 ID"
// @Success      200  {object}  response.Envelope
// @Router       /employer/jobs/{id} [delete]
func (h *JobHandler) Delete(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uid, id); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// Offline godoc
// @Summary      雇主下架自己的岗位
// @Tags         Employer/Jobs
// @Security     BearerAuth
// @Param        id   path      int  true  "岗位 ID"
// @Success      200  {object}  response.Envelope
// @Router       /employer/jobs/{id}/offline [post]
func (h *JobHandler) Offline(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Offline(c.Request.Context(), uid, id); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// ListMine godoc
// @Summary      我发布的岗位列表(雇主端)
// @Tags         Employer/Jobs
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "页码"
// @Param        size  query     int  false  "每页大小"
// @Success      200   {object}  response.Envelope{data=response.PageData{list=[]dto.JobResp}}
// @Router       /employer/jobs [get]
func (h *JobHandler) ListMine(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	page := pagination.FromGin(c)
	rows, total, err := h.svc.ListByEmployer(c.Request.Context(), uid, page.Page, page.Size)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}

// Audit godoc
// @Summary      审核岗位(管理端)
// @Tags         Admin/Jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "岗位 ID"
// @Param        body  body      dto.AuditJobReq  true  "action: 2=通过 4=拒绝"
// @Success      200   {object}  response.Envelope{data=dto.JobResp}
// @Router       /admin/jobs/{id}/audit [post]
func (h *JobHandler) Audit(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.AuditJobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	j, err := h.svc.Audit(c.Request.Context(), id, req.Action, req.Remark)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, j)
}

// ListPendingAdmin godoc
// @Summary      待审核岗位列表(管理端)
// @Tags         Admin/Jobs
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "页码"
// @Param        size  query     int  false  "每页大小"
// @Success      200   {object}  response.Envelope{data=response.PageData{list=[]dto.JobResp}}
// @Router       /admin/jobs [get]
func (h *JobHandler) ListPendingAdmin(c *gin.Context) {
	page := pagination.FromGin(c)
	rows, total, err := h.svc.ListPending(c.Request.Context(), page.Page, page.Size)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}

func parseJobFilter(c *gin.Context) repository.JobListFilter {
	f := repository.JobListFilter{}
	if s := c.Query("category_id"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			f.CategoryID = v
		}
	}
	f.Location = c.Query("location")
	if s := c.Query("salary_min"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			f.SalaryMin = v
		}
	}
	if s := c.Query("salary_max"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			f.SalaryMax = v
		}
	}
	f.Keyword = c.Query("keyword")
	return f
}
