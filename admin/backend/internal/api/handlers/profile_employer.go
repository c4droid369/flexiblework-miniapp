package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// EmployerProfileHandler exposes /employer/profile endpoints.
type EmployerProfileHandler struct{ svc *service.EmployerProfileService }

func NewEmployerProfileHandler(svc *service.EmployerProfileService) *EmployerProfileHandler {
	return &EmployerProfileHandler{svc: svc}
}

// GetMy godoc
// @Summary      获取我的雇主资料
// @Tags         Employer/Profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=dto.EmployerProfileResp}
// @Router       /employer/profile [get]
func (h *EmployerProfileHandler) GetMy(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	p, err := h.svc.GetMy(c.Request.Context(), uid)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, p)
}

// UpsertMy godoc
// @Summary      创建/更新雇主资料
// @Tags         Employer/Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UpsertEmployerProfileReq  true  "部分更新"
// @Success      200   {object}  response.Envelope{data=dto.EmployerProfileResp}
// @Router       /employer/profile [post]
func (h *EmployerProfileHandler) UpsertMy(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	var req dto.UpsertEmployerProfileReq
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
// @Summary      提交雇主资质(营业执照)
// @Tags         Employer/Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.SubmitEmployerCertificationReq  true  "资质信息"
// @Success      200   {object}  response.Envelope{data=dto.EmployerProfileResp}
// @Router       /employer/certification [post]
func (h *EmployerProfileHandler) SubmitCertification(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	var req dto.SubmitEmployerCertificationReq
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
