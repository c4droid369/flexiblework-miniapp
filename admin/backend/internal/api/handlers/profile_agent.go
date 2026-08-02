package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// AgentProfileHandler exposes /agent/profile endpoints. Agent-side job
// posting reuses /employer/jobs (the perm codes and cert gate are shared).
type AgentProfileHandler struct{ svc *service.AgentProfileService }

func NewAgentProfileHandler(svc *service.AgentProfileService) *AgentProfileHandler {
	return &AgentProfileHandler{svc: svc}
}

// GetMy godoc
// @Summary      获取我的代理资料
// @Tags         Agent/Profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=dto.AgentProfileResp}
// @Router       /agent/profile [get]
func (h *AgentProfileHandler) GetMy(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	p, err := h.svc.GetMy(c.Request.Context(), uid)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, p)
}

// UpsertMy godoc
// @Summary      创建/更新代理资料
// @Tags         Agent/Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UpsertAgentProfileReq  true  "部分更新"
// @Success      200   {object}  response.Envelope{data=dto.AgentProfileResp}
// @Router       /agent/profile [post]
func (h *AgentProfileHandler) UpsertMy(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	var req dto.UpsertAgentProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	p, err := h.svc.UpsertMy(c.Request.Context(), uid, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, p)
}

// SubmitCertification godoc
// @Summary      提交代理认证(身份证 + 校园卡)
// @Tags         Agent/Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.SubmitAgentCertificationReq  true  "证件图片 + 实名信息"
// @Success      200   {object}  response.Envelope{data=dto.AgentProfileResp}
// @Router       /agent/certification [post]
func (h *AgentProfileHandler) SubmitCertification(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	var req dto.SubmitAgentCertificationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	p, err := h.svc.SubmitCertification(c.Request.Context(), uid, req)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, p)
}