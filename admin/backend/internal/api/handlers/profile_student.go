package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// StudentProfileHandler exposes /student/profile endpoints. Caller-side only;
// the admin audit endpoints live in the admin handler.
type StudentProfileHandler struct{ svc *service.StudentProfileService }

func NewStudentProfileHandler(svc *service.StudentProfileService) *StudentProfileHandler {
	return &StudentProfileHandler{svc: svc}
}

// GetMy godoc
// @Summary      获取我的学生资料
// @Tags         Student/Profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=dto.StudentProfileResp}
// @Router       /student/profile [get]
func (h *StudentProfileHandler) GetMy(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	p, err := h.svc.GetMy(c.Request.Context(), uid)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, p)
}

// UpsertMy godoc
// @Summary      创建/更新学生资料
// @Tags         Student/Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UpsertStudentProfileReq  true  "部分更新"
// @Success      200   {object}  response.Envelope{data=dto.StudentProfileResp}
// @Router       /student/profile [post]
func (h *StudentProfileHandler) UpsertMy(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	var req dto.UpsertStudentProfileReq
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
// @Summary      提交学生认证(身份证+学生证)
// @Tags         Student/Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.SubmitStudentCertificationReq  true  "三张图片 URL"
// @Success      200   {object}  response.Envelope{data=dto.StudentProfileResp}
// @Router       /student/certification [post]
func (h *StudentProfileHandler) SubmitCertification(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	var req dto.SubmitStudentCertificationReq
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
