package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// AdminHandler aggregates the admin-only endpoints that don't fit elsewhere:
// the cert-review queues (student + employer) and the audit calls.
type AdminHandler struct {
	studentProfileSvc  *service.StudentProfileService
	employerProfileSvc *service.EmployerProfileService
}

func NewAdminHandler(s *service.StudentProfileService, e *service.EmployerProfileService) *AdminHandler {
	return &AdminHandler{studentProfileSvc: s, employerProfileSvc: e}
}

// ListPendingStudentCerts godoc
// @Summary      待审核学生认证
// @Tags         Admin/Cert
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=[]dto.StudentCertListItem}
// @Router       /admin/student-certifications [get]
func (h *AdminHandler) ListPendingStudentCerts(c *gin.Context) {
	rows, err := h.studentProfileSvc.ListPendingCerts(c.Request.Context())
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, rows)
}

// AuditStudentCert godoc
// @Summary      审核学生认证
// @Tags         Admin/Cert
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "学生 user_id"
// @Param        body  body      dto.CertAuditReq  true  "action: 2=通过 3=拒绝"
// @Success      200   {object}  response.Envelope
// @Router       /admin/student-certifications/{id}/audit [post]
func (h *AdminHandler) AuditStudentCert(c *gin.Context) {
	uid, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CertAuditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	if err := h.studentProfileSvc.AuditCert(c.Request.Context(), uid, req.Action, req.Remark); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// ListPendingEmployerCerts godoc
// @Summary      待审核雇主资质
// @Tags         Admin/Cert
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=[]dto.EmployerCertListItem}
// @Router       /admin/employer-certifications [get]
func (h *AdminHandler) ListPendingEmployerCerts(c *gin.Context) {
	rows, err := h.employerProfileSvc.ListPendingCerts(c.Request.Context())
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, rows)
}

// AuditEmployerCert godoc
// @Summary      审核雇主资质
// @Tags         Admin/Cert
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "雇主 user_id"
// @Param        body  body      dto.CertAuditReq  true  "action: 2=通过 3=拒绝"
// @Success      200   {object}  response.Envelope
// @Router       /admin/employer-certifications/{id}/audit [post]
func (h *AdminHandler) AuditEmployerCert(c *gin.Context) {
	uid, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CertAuditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	if err := h.employerProfileSvc.AuditCert(c.Request.Context(), uid, req.Action, req.Remark); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// silence unused-import warning when auth is only referenced in the
// underlying services.
var _ = auth.UserIDFrom
