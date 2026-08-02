package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// MenuHandler exposes /system/menus endpoints.
type MenuHandler struct{ svc *service.MenuService }

func NewMenuHandler(svc *service.MenuService) *MenuHandler { return &MenuHandler{svc: svc} }

// Tree godoc
// @Summary      菜单树（完整）
// @Description  返回全部菜单的层级结构（不过滤权限），仅用于管理后台配置。
// @Tags         System/Menus
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=[]dto.MenuTree}
// @Router       /system/menus [get]
func (h *MenuHandler) Tree(c *gin.Context) {
	tree, err := h.svc.Tree(c.Request.Context())
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, tree)
}

// Get godoc
// @Summary      获取单个菜单
// @Tags         System/Menus
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "菜单 ID"
// @Success      200  {object}  response.Envelope{data=dto.MenuResp}
// @Failure      404  {object}  response.Envelope
// @Router       /system/menus/{id} [get]
func (h *MenuHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	m, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, m)
}

// Create godoc
// @Summary      创建菜单
// @Tags         System/Menus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateMenuReq  true  "新菜单"
// @Success      201   {object}  response.Envelope{data=dto.MenuResp}
// @Router       /system/menus [post]
func (h *MenuHandler) Create(c *gin.Context) {
	var req dto.CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	m, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, m)
}

// Update godoc
// @Summary      更新菜单
// @Tags         System/Menus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "菜单 ID"
// @Param        body  body      dto.UpdateMenuReq  true  "要更新的字段"
// @Success      200   {object}  response.Envelope{data=dto.MenuResp}
// @Router       /system/menus/{id} [put]
func (h *MenuHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	m, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, m)
}

// Delete godoc
// @Summary      删除菜单（递归删除子项）
// @Tags         System/Menus
// @Security     BearerAuth
// @Param        id   path      int  true  "菜单 ID"
// @Success      200  {object}  response.Envelope
// @Router       /system/menus/{id} [delete]
func (h *MenuHandler) Delete(c *gin.Context) {
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
