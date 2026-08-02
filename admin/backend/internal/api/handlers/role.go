package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/exporter"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// RoleHandler exposes /system/roles endpoints.
type RoleHandler struct{ svc *service.RoleService }

func NewRoleHandler(svc *service.RoleService) *RoleHandler { return &RoleHandler{svc: svc} }

// List godoc
// @Summary      角色列表
// @Tags         System/Roles
// @Produce      json
// @Security     BearerAuth
// @Param        page     query     int  false  "页码 (默认 1)"
// @Param        size     query     int  false  "每页大小 (默认 10, 最大 200)"
// @Param        keyword  query     string  false  "模糊搜索名称/编码"
// @Success      200      {object}  response.Envelope{data=response.PageData}
// @Router       /system/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	page := pagination.FromGin(c)
	search := pagination.SearchFromGin(c)
	roles, total, err := h.svc.List(c.Request.Context(), page, search)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, roles, total, page.Page, page.Size)
}

// Get godoc
// @Summary      获取单个角色
// @Tags         System/Roles
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "角色 ID"
// @Success      200  {object}  response.Envelope{data=dto.RoleResp}
// @Failure      404  {object}  response.Envelope
// @Router       /system/roles/{id} [get]
func (h *RoleHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	r, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, r)
}

// Create godoc
// @Summary      创建角色
// @Tags         System/Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateRoleReq  true  "新角色"
// @Success      201   {object}  response.Envelope{data=dto.RoleResp}
// @Failure      400   {object}  response.Envelope
// @Failure      409   {object}  response.Envelope  "角色编码已存在"
// @Router       /system/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleReq
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
// @Summary      更新角色
// @Tags         System/Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "角色 ID"
// @Param        body  body      dto.UpdateRoleReq  true  "要更新的字段"
// @Success      200   {object}  response.Envelope{data=dto.RoleResp}
// @Failure      404   {object}  response.Envelope
// @Router       /system/roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateRoleReq
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
// @Summary      删除单个角色
// @Tags         System/Roles
// @Security     BearerAuth
// @Param        id   path      int  true  "角色 ID"
// @Success      200  {object}  response.Envelope
// @Router       /system/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
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

// BatchDelete godoc
// @Summary      批量删除角色
// @Tags         System/Roles
// @Accept       json
// @Security     BearerAuth
// @Param        body  body      dto.BatchDeleteReq  true  "ID 列表"
// @Success      200   {object}  response.Envelope
// @Router       /system/roles/batch-delete [post]
func (h *RoleHandler) BatchDelete(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	if err := h.svc.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// AssignMenus godoc
// @Summary      给角色分配菜单
// @Description  覆盖式写入；传入空数组可清空。
// @Tags         System/Roles
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      int  true  "角色 ID"
// @Param        body  body      dto.AssignMenusReq  true  "菜单 ID 列表"
// @Success      200   {object}  response.Envelope
// @Router       /system/roles/{id}/menus [post]
func (h *RoleHandler) AssignMenus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.AssignMenusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	if err := h.svc.AssignMenus(c.Request.Context(), id, req.MenuIDs); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// ExportExcel godoc
// @Summary      导出角色列表为 Excel
// @Tags         System/Roles
// @Security     BearerAuth
// @Param        keyword  query     string  false  "模糊搜索"
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success      200      {file}    file
// @Router       /system/roles/export/excel [get]
func (h *RoleHandler) ExportExcel(c *gin.Context) {
	page := pagination.Page{Page: 1, Size: pagination.MaxSize, Off: 0, Limit: pagination.MaxSize}
	search := pagination.SearchFromGin(c)
	roles, _, err := h.svc.List(c.Request.Context(), page, search)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	sheet := exporter.Sheet{
		Name:    "Roles",
		Headers: []string{"ID", "Code", "Name", "Description", "Sort", "Status", "Created"},
		Rows:    roleRows(roles),
	}
	if err := exporter.Excel(c, "roles", sheet); err != nil {
		httperr.Write(c, nil, httperr.Internal(err))
	}
}

// ExportCSV godoc
// @Summary      导出角色列表为 CSV
// @Tags         System/Roles
// @Security     BearerAuth
// @Param        keyword  query     string  false  "模糊搜索"
// @Produce      text/csv
// @Success      200      {file}    file
// @Router       /system/roles/export/csv [get]
func (h *RoleHandler) ExportCSV(c *gin.Context) {
	page := pagination.Page{Page: 1, Size: pagination.MaxSize, Off: 0, Limit: pagination.MaxSize}
	search := pagination.SearchFromGin(c)
	roles, _, err := h.svc.List(c.Request.Context(), page, search)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	sheet := exporter.Sheet{
		Name:    "Roles",
		Headers: []string{"ID", "Code", "Name", "Description", "Sort", "Status", "Created"},
		Rows:    roleRows(roles),
	}
	if err := exporter.CSV(c, "roles", sheet); err != nil {
		httperr.Write(c, nil, httperr.Internal(err))
	}
}

func roleRows(roles []dto.RoleResp) [][]any {
	rows := make([][]any, 0, len(roles))
	for _, r := range roles {
		status := "active"
		if r.Status == 2 {
			status = "disabled"
		}
		rows = append(rows, []any{r.ID, r.Code, r.Name, r.Description, r.Sort, status, r.CreatedAt})
	}
	return rows
}
