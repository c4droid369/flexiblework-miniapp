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

// MessageHandler covers the bell-icon stream (user-side) and the admin
// broadcast endpoint.
type MessageHandler struct{ svc *service.MessageService }

func NewMessageHandler(svc *service.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

// ListMine godoc
// @Summary      我的消息列表
// @Tags         Messages
// @Produce      json
// @Security     BearerAuth
// @Param        page        query     int     false  "页码"
// @Param        size        query     int     false  "每页大小"
// @Param        only_unread query     bool    false  "true=只返回未读"
// @Success      200         {object}  response.Envelope{data=response.PageData{list=[]dto.MessageResp}}
// @Router       /messages [get]
func (h *MessageHandler) ListMine(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	page := pagination.FromGin(c)
	onlyUnread := false
	if s := c.Query("only_unread"); s == "true" || s == "1" {
		onlyUnread = true
	}
	rows, total, err := h.svc.ListByUser(c.Request.Context(), uid, page.Page, page.Size, onlyUnread)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, rows, total, page.Page, page.Size)
}

// MarkRead godoc
// @Summary      标记消息为已读
// @Tags         Messages
// @Security     BearerAuth
// @Param        id   path      int  true  "消息 ID"
// @Success      200  {object}  response.Envelope
// @Router       /messages/{id}/read [post]
func (h *MessageHandler) MarkRead(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), uid, id); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// MarkAllRead godoc
// @Summary      全部消息标为已读
// @Tags         Messages
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope
// @Router       /messages/read-all [post]
func (h *MessageHandler) MarkAllRead(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	if err := h.svc.MarkAllRead(c.Request.Context(), uid); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// Broadcast godoc
// @Summary      广播系统消息(管理端)
// @Description  user_type: all|admin|student|employer。返回推送条数。
// @Tags         Admin/Messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.BroadcastMessageReq  true  "广播内容"
// @Success      200   {object}  response.Envelope{data=int}
// @Router       /admin/messages/broadcast [post]
func (h *MessageHandler) Broadcast(c *gin.Context) {
	var req dto.BroadcastMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	n, err := h.svc.Broadcast(c.Request.Context(), req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, n)
}

// silence unused-import warning when strconv is the only consumer.
var _ = strconv.Atoi
