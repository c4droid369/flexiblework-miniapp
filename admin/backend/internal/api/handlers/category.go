package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// CategoryHandler is the public + admin CRUD for the gig taxonomy. The public
// list endpoint (no auth) is mounted in the unauthenticated router group; the
// admin CRUD endpoints sit behind the admin permission codes.
type CategoryHandler struct{ svc *service.CategoryService }

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// ListPublic godoc
// @Summary      公开的分类列表(无需登录)
// @Tags         Categories
// @Produce      json
// @Param        status  query     int  false  "可选,只返回启用(1)或禁用(2)"
// @Success      200     {object}  response.Envelope{data=[]dto.CategoryResp}
// @Router       /categories [get]
func (h *CategoryHandler) ListPublic(c *gin.Context) {
	status := int8(0)
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = int8(v)
		}
	}
	rows, err := h.svc.List(c.Request.Context(), status)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, rows)
}

// Create godoc
// @Summary      创建分类(管理端)
// @Tags         Admin/Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateCategoryReq  true  "新分类"
// @Success      201   {object}  response.Envelope{data=dto.CategoryResp}
// @Router       /admin/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	r, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, r)
}

// Update godoc
// @Summary      更新分类(管理端)
// @Tags         Admin/Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "分类 ID"
// @Param        body  body      dto.UpdateCategoryReq  true  "更新字段"
// @Success      200   {object}  response.Envelope{data=dto.CategoryResp}
// @Router       /admin/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	r, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, r)
}

// Delete godoc
// @Summary      删除分类(管理端)
// @Tags         Admin/Categories
// @Security     BearerAuth
// @Param        id   path      int  true  "分类 ID"
// @Success      200  {object}  response.Envelope
// @Router       /admin/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
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

// ListAdmin godoc
// @Summary      管理端分类列表
// @Tags         Admin/Categories
// @Produce      json
// @Security     BearerAuth
// @Param        status  query     int  false  "可选,1=启用 2=禁用"
// @Success      200     {object}  response.Envelope{data=[]dto.CategoryResp}
// @Router       /admin/categories [get]
func (h *CategoryHandler) ListAdmin(c *gin.Context) {
	h.ListPublic(c)
}
