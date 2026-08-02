package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/exporter"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// UserHandler exposes /system/users endpoints. Permission middleware is
// applied at the router level via RequirePerm.
type UserHandler struct{ svc *service.UserService }

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc: svc} }

// List godoc
// @Summary      用户列表
// @Tags         System/Users
// @Produce      json
// @Security     BearerAuth
// @Param        page     query     int  false  "页码 (默认 1)"
// @Param        size     query     int  false  "每页大小 (默认 10, 最大 200)"
// @Param        keyword  query     string  false  "模糊搜索用户名/昵称/邮箱"
// @Success      200      {object}  response.Envelope{data=response.PageData}
// @Router       /system/users [get]
func (h *UserHandler) List(c *gin.Context) {
	page := pagination.FromGin(c)
	search := pagination.SearchFromGin(c)
	users, total, err := h.svc.List(c.Request.Context(), page, search)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, users, total, page.Page, page.Size)
}

// Get godoc
// @Summary      获取单个用户
// @Tags         System/Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "用户 ID"
// @Success      200  {object}  response.Envelope{data=dto.UserResp}
// @Failure      404  {object}  response.Envelope
// @Router       /system/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, u)
}

// Create godoc
// @Summary      创建用户
// @Tags         System/Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateUserReq  true  "新用户"
// @Success      201   {object}  response.Envelope{data=dto.UserResp}
// @Failure      400   {object}  response.Envelope
// @Failure      409   {object}  response.Envelope  "用户名已存在"
// @Router       /system/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	u, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKCreated(c, u)
}

// Update godoc
// @Summary      更新用户
// @Tags         System/Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "用户 ID"
// @Param        body  body      dto.UpdateUserReq  true  "要更新的字段（指针字段，未传则不更新）"
// @Success      200   {object}  response.Envelope{data=dto.UserResp}
// @Failure      400   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Router       /system/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	u, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, u)
}

// Delete godoc
// @Summary      删除单个用户
// @Tags         System/Users
// @Security     BearerAuth
// @Param        id   path      int  true  "用户 ID"
// @Success      200  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /system/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
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
// @Summary      批量删除用户
// @Tags         System/Users
// @Accept       json
// @Security     BearerAuth
// @Param        body  body      dto.BatchDeleteReq  true  "ID 列表"
// @Success      200   {object}  response.Envelope
// @Router       /system/users/batch-delete [post]
func (h *UserHandler) BatchDelete(c *gin.Context) {
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

// ResetPassword godoc
// @Summary      重置用户密码
// @Description  管理员无需知道旧密码即可设置新密码。
// @Tags         System/Users
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      int  true  "用户 ID"
// @Param        body  body      dto.ResetPasswordReq  true  "新密码"
// @Success      200   {object}  response.Envelope
// @Router       /system/users/{id}/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), id, req.NewPassword); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// ChangeStatus godoc
// @Summary      启用/禁用用户
// @Tags         System/Users
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      int  true  "用户 ID"
// @Param        body  body      dto.ChangeStatusReq  true  "目标状态：1=启用 2=禁用"
// @Success      200   {object}  response.Envelope
// @Router       /system/users/{id}/status [post]
func (h *UserHandler) ChangeStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ChangeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	if err := h.svc.ChangeStatus(c.Request.Context(), id, req.Status); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// AssignRoles godoc
// @Summary      给用户分配角色
// @Description  body {ids: [...]} 是角色 ID 列表，传空数组清空。
// @Tags         System/Users
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      int  true  "用户 ID"
// @Param        body  body      dto.BatchDeleteReq  true  "角色 ID 列表"
// @Success      200   {object}  response.Envelope
// @Router       /system/users/{id}/roles [post]
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	if err := h.svc.AssignRoles(c.Request.Context(), id, req.IDs); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// ExportExcel godoc
// @Summary      导出用户列表为 Excel
// @Tags         System/Users
// @Security     BearerAuth
// @Param        keyword  query     string  false  "模糊搜索（与 List 同语义）"
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success      200      {file}    file
// @Router       /system/users/export/excel [get]
func (h *UserHandler) ExportExcel(c *gin.Context) {
	page := pagination.Page{Page: 1, Size: pagination.MaxSize, Off: 0, Limit: pagination.MaxSize}
	search := pagination.SearchFromGin(c)
	users, _, err := h.svc.List(c.Request.Context(), page, search)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	sheet := exporter.Sheet{
		Name:    "Users",
		Headers: []string{"ID", "Username", "Nickname", "Email", "Phone", "Status", "Created"},
		Rows:    userRows(users),
	}
	if err := exporter.Excel(c, "users", sheet); err != nil {
		httperr.Write(c, nil, httperr.Internal(err))
	}
}

// ExportCSV godoc
// @Summary      导出用户列表为 CSV
// @Tags         System/Users
// @Security     BearerAuth
// @Param        keyword  query     string  false  "模糊搜索（与 List 同语义）"
// @Produce      text/csv
// @Success      200      {file}    file
// @Router       /system/users/export/csv [get]
func (h *UserHandler) ExportCSV(c *gin.Context) {
	page := pagination.Page{Page: 1, Size: pagination.MaxSize, Off: 0, Limit: pagination.MaxSize}
	search := pagination.SearchFromGin(c)
	users, _, err := h.svc.List(c.Request.Context(), page, search)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	sheet := exporter.Sheet{
		Name:    "Users",
		Headers: []string{"ID", "Username", "Nickname", "Email", "Phone", "Status", "Created"},
		Rows:    userRows(users),
	}
	if err := exporter.CSV(c, "users", sheet); err != nil {
		httperr.Write(c, nil, httperr.Internal(err))
	}
}

func userRows(users []dto.UserResp) [][]any {
	rows := make([][]any, 0, len(users))
	for _, u := range users {
		status := "active"
		if u.Status == 2 {
			status = "disabled"
		}
		rows = append(rows, []any{u.ID, u.Username, u.Nickname, u.Email, u.Phone, status, u.CreatedAt})
	}
	return rows
}

// parseID extracts the :id path param into a uint64. On failure it writes
// a 400 envelope and returns ok=false; handlers must return immediately.
func parseID(c *gin.Context) (uint64, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		httperr.Write(c, nil, httperr.BadRequest("invalid id"))
		return 0, false
	}
	return id, true
}
