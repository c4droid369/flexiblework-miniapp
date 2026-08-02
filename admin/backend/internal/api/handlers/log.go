package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// LogHandler exposes /system/logs endpoints.
type LogHandler struct{ svc *service.OperationLogService }

func NewLogHandler(svc *service.OperationLogService) *LogHandler { return &LogHandler{svc: svc} }

// List godoc
// @Summary      操作日志列表
// @Tags         System/Logs
// @Produce      json
// @Security     BearerAuth
// @Param        page     query     int  false  "页码 (默认 1)"
// @Param        size     query     int  false  "每页大小 (默认 20)"
// @Param        keyword  query     string  false  "模糊搜索用户名/路径/操作"
// @Param        action   query     string  false  "按 action 模糊过滤"
// @Success      200      {object}  response.Envelope{data=response.PageData}
// @Router       /system/logs [get]
func (h *LogHandler) List(c *gin.Context) {
	page := pagination.FromGin(c)
	search := pagination.SearchFromGin(c)
	action := c.Query("action")
	logs, total, err := h.svc.List(c.Request.Context(), page, search, action)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, logs, total, page.Page, page.Size)
}

// BatchDelete godoc
// @Summary      批量删除日志
// @Tags         System/Logs
// @Accept       json
// @Security     BearerAuth
// @Param        body  body      dto.BatchDeleteReq  true  "ID 列表"
// @Success      200   {object}  response.Envelope
// @Router       /system/logs/batch-delete [post]
func (h *LogHandler) BatchDelete(c *gin.Context) {
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
