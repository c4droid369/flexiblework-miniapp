package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// AuthHandler exposes /auth/* endpoints. All methods are 1-2 lines: parse,
// call service, write envelope.
type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

// Login godoc
// @Summary      登录
// @Description  用用户名密码换取 access + refresh token。返回的 permissions 数组同时进 JWT claims，2h 内免查 DB。
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginReq  true  "credentials"
// @Success      200   {object}  response.Envelope{data=dto.LoginResp}
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, resp)
}

// Refresh godoc
// @Summary      刷新 token
// @Description  用有效 refresh token 换新的 access + refresh。旧 refresh 被吊销（rotation）；如发现已被吊销的 token 复用，会吊销该用户整个 token 家族。
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshReq  true  "refresh token"
// @Success      200   {object}  response.Envelope{data=dto.LoginResp}
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	resp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, resp)
}

// Logout godoc
// @Summary      登出
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.LogoutReq  false  "可选的 refresh_token，传入则一并吊销"
// @Success      200   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutReq
	_ = c.ShouldBindJSON(&req)
	uid := auth.UserIDFrom(c.Request.Context())
	if err := h.svc.Logout(c.Request.Context(), uid, req.RefreshToken); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}

// Register godoc
// @Summary      学生/雇主注册
// @Description  注册即登录,返回 access+refresh token,userType 决定后续入口(tabbar)。会自动创建一个空的 profile 行,前端 GET 即可拿到。
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterReq  true  "注册信息(user_type: student|employer)"
// @Success      200   {object}  response.Envelope{data=dto.LoginResp}
// @Failure      400   {object}  response.Envelope
// @Failure      409   {object}  response.Envelope  "用户名已存在"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Write(c, nil, httperr.BadRequest("invalid request body"))
		return
	}
	resp, err := h.svc.Register(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, resp)
}

// Me godoc
// @Summary      当前用户信息
// @Description  包含 profile、roles、permissions（按钮权限码数组）、menus（用于前端动态路由）。
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  response.Envelope{data=dto.MeResp}
// @Failure      401   {object}  response.Envelope
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	uid := auth.UserIDFrom(c.Request.Context())
	me, err := h.svc.Me(c.Request.Context(), uid)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, me)
}
