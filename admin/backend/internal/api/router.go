// Package api route registration. Public auth endpoints live under /auth;
// authenticated resources mount under their own sub-routers.
package api

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/handlers"
	"github.com/admin-template/backend/internal/api/middleware"
)

// RegisterAPIRoutes wires every public and authenticated endpoint. Adding a
// new resource is one block in this function — find the right group, mount
// the route, attach the permission middleware.
func RegisterAPIRoutes(r *gin.RouterGroup, d Deps) {
	// ----- Public auth -----
	authH := handlers.NewAuthHandler(d.AuthSvc)
	r.POST("/auth/login", authH.Login)
	r.POST("/auth/refresh", authH.Refresh)
	r.POST("/auth/register", authH.Register)

	// ----- Authenticated auth -----
	authed := r.Group("/auth", middleware.Auth(d.Issuer, d.Logger))
	authed.POST("/logout", authH.Logout)
	authed.GET("/me", authH.Me)

	// ----- System management (admin RBAC) -----
	sys := r.Group(
		"/system",
		middleware.Auth(d.Issuer, d.Logger),
		middleware.OperationLog(d.OpSvc, d.Logger),
	)
	users := handlers.NewUserHandler(d.UserSvc)
	sys.GET("/users", middleware.RequirePerm(d.Logger, "user:view"), users.List)
	sys.GET("/users/:id", middleware.RequirePerm(d.Logger, "user:view"), users.Get)
	sys.POST("/users", middleware.RequirePerm(d.Logger, "user:create"), users.Create)
	sys.PUT("/users/:id", middleware.RequirePerm(d.Logger, "user:update"), users.Update)
	sys.DELETE("/users/:id", middleware.RequirePerm(d.Logger, "user:delete"), users.Delete)
	sys.POST("/users/batch-delete", middleware.RequirePerm(d.Logger, "user:delete"), users.BatchDelete)
	sys.POST("/users/:id/reset-password", middleware.RequirePerm(d.Logger, "user:reset-password"), users.ResetPassword)
	sys.POST("/users/:id/status", middleware.RequirePerm(d.Logger, "user:update"), users.ChangeStatus)
	sys.POST("/users/:id/roles", middleware.RequirePerm(d.Logger, "user:update"), users.AssignRoles)
	sys.GET("/users/export/excel", middleware.RequirePerm(d.Logger, "user:view"), users.ExportExcel)
	sys.GET("/users/export/csv", middleware.RequirePerm(d.Logger, "user:view"), users.ExportCSV)

	roles := handlers.NewRoleHandler(d.RoleSvc)
	sys.GET("/roles", middleware.RequirePerm(d.Logger, "role:view"), roles.List)
	sys.GET("/roles/:id", middleware.RequirePerm(d.Logger, "role:view"), roles.Get)
	sys.POST("/roles", middleware.RequirePerm(d.Logger, "role:create"), roles.Create)
	sys.PUT("/roles/:id", middleware.RequirePerm(d.Logger, "role:update"), roles.Update)
	sys.DELETE("/roles/:id", middleware.RequirePerm(d.Logger, "role:delete"), roles.Delete)
	sys.POST("/roles/batch-delete", middleware.RequirePerm(d.Logger, "role:delete"), roles.BatchDelete)
	sys.POST("/roles/:id/menus", middleware.RequirePerm(d.Logger, "role:assign-menu"), roles.AssignMenus)
	sys.GET("/roles/export/excel", middleware.RequirePerm(d.Logger, "role:view"), roles.ExportExcel)
	sys.GET("/roles/export/csv", middleware.RequirePerm(d.Logger, "role:view"), roles.ExportCSV)

	menus := handlers.NewMenuHandler(d.MenuSvc)
	sys.GET("/menus", middleware.RequirePerm(d.Logger, "menu:view"), menus.Tree)
	sys.GET("/menus/:id", middleware.RequirePerm(d.Logger, "menu:view"), menus.Get)
	sys.POST("/menus", middleware.RequirePerm(d.Logger, "menu:create"), menus.Create)
	sys.PUT("/menus/:id", middleware.RequirePerm(d.Logger, "menu:update"), menus.Update)
	sys.DELETE("/menus/:id", middleware.RequirePerm(d.Logger, "menu:delete"), menus.Delete)

	logs := handlers.NewLogHandler(d.OpSvc)
	sys.GET("/logs", middleware.RequirePerm(d.Logger, "log:view"), logs.List)
	sys.POST("/logs/batch-delete", middleware.RequirePerm(d.Logger, "log:delete"), logs.BatchDelete)

	// ----- File management -----
	up := handlers.NewUploadHandler(d.FileSvc)
	upload := r.Group("", middleware.Auth(d.Issuer, d.Logger))
	upload.POST("/upload", middleware.RequirePerm(d.Logger, "file:upload"), up.Upload)
	upload.GET("/files-list", middleware.RequirePerm(d.Logger, "file:view"), up.List)
	upload.DELETE("/files-list/:id", middleware.RequirePerm(d.Logger, "file:delete"), up.Delete)

	// ====================================================================
	// Business routes — campus gig work mini-program
	// ====================================================================

	// ----- Public business reads (no auth) -----
	categoryH := handlers.NewCategoryHandler(d.CategorySvc)
	jobH := handlers.NewJobHandler(d.JobSvc)
	r.GET("/categories", categoryH.ListPublic)
	r.GET("/jobs", jobH.ListPublic)
	r.GET("/jobs/:id", jobH.GetPublic)

	// ----- Authenticated business group -----
	biz := r.Group("", middleware.Auth(d.Issuer, d.Logger))

	// Student side
	studentProfileH := handlers.NewStudentProfileHandler(d.StudentProfileSvc)
	biz.GET("/student/profile",
		middleware.RequirePerm(d.Logger, "profile:view"), studentProfileH.GetMy)
	biz.POST("/student/profile",
		middleware.RequirePerm(d.Logger, "profile:update"), studentProfileH.UpsertMy)
	biz.POST("/student/certification",
		middleware.RequirePerm(d.Logger, "cert:submit"), studentProfileH.SubmitCertification)

	// Agent side (campus agent — same /employer/* endpoints for jobs/applications/orders)
	agentProfileH := handlers.NewAgentProfileHandler(d.AgentProfileSvc)
	biz.GET("/agent/profile",
		middleware.RequirePerm(d.Logger, "profile:view"), agentProfileH.GetMy)
	biz.POST("/agent/profile",
		middleware.RequirePerm(d.Logger, "profile:update"), agentProfileH.UpsertMy)
	biz.POST("/agent/certification",
		middleware.RequirePerm(d.Logger, "cert:submit"), agentProfileH.SubmitCertification)

	// Apply (under /jobs/:id) + student applications
	appH := handlers.NewApplicationHandler(d.AppSvc)
	biz.POST("/jobs/:id/apply",
		middleware.RequirePerm(d.Logger, "job:apply"), appH.Apply)
	biz.GET("/applications",
		middleware.RequirePerm(d.Logger, "application:view"), appH.ListMine)
	biz.GET("/applications/:id",
		middleware.RequirePerm(d.Logger, "application:view"), appH.Get)
	biz.POST("/applications/:id/cancel",
		middleware.RequirePerm(d.Logger, "application:cancel"), appH.Cancel)

	// Student orders
	orderH := handlers.NewOrderHandler(d.OrderSvc, d.ReviewSvc)
	biz.GET("/orders",
		middleware.RequirePerm(d.Logger, "order:view"), orderH.ListMine)
	biz.GET("/orders/:id",
		middleware.RequirePerm(d.Logger, "order:view"), orderH.GetMine)
	biz.POST("/orders/:id/pay",
		middleware.RequirePerm(d.Logger, "order:pay"), orderH.Pay)
	biz.POST("/orders/:id/checkin",
		middleware.RequirePerm(d.Logger, "order:checkin"), orderH.Checkin)
	biz.POST("/orders/:id/complete",
		middleware.RequirePerm(d.Logger, "order:complete"), orderH.Complete)
	biz.POST("/orders/:id/cancel",
		middleware.RequirePerm(d.Logger, "order:cancel"), orderH.CancelMine)
	biz.POST("/orders/:id/review",
		middleware.RequirePerm(d.Logger, "review:create"), orderH.Review)

	// Reviews — list per order (shared between both sides)
	reviewH := handlers.NewReviewHandler(d.ReviewSvc)
	biz.GET("/orders/:id/reviews",
		middleware.RequirePerm(d.Logger, "review:view"), reviewH.ListByOrder)

	// Employer side
	employerProfileH := handlers.NewEmployerProfileHandler(d.EmployerProfileSvc)
	biz.GET("/employer/profile",
		middleware.RequirePerm(d.Logger, "profile:view"), employerProfileH.GetMy)
	biz.POST("/employer/profile",
		middleware.RequirePerm(d.Logger, "profile:update"), employerProfileH.UpsertMy)
	biz.POST("/employer/certification",
		middleware.RequirePerm(d.Logger, "cert:submit"), employerProfileH.SubmitCertification)

	biz.GET("/employer/jobs",
		middleware.RequirePerm(d.Logger, "job:create"), jobH.ListMine)
	biz.POST("/employer/jobs",
		middleware.RequirePerm(d.Logger, "job:create"), jobH.Create)
	biz.PUT("/employer/jobs/:id",
		middleware.RequirePerm(d.Logger, "job:update"), jobH.Update)
	biz.DELETE("/employer/jobs/:id",
		middleware.RequirePerm(d.Logger, "job:delete"), jobH.Delete)
	biz.POST("/employer/jobs/:id/offline",
		middleware.RequirePerm(d.Logger, "job:offline"), jobH.Offline)
	biz.GET("/employer/jobs/:id/applications",
		middleware.RequirePerm(d.Logger, "application:view"), appH.ListByJob)
	biz.POST("/employer/applications/:id/audit",
		middleware.RequirePerm(d.Logger, "application:audit"), appH.Audit)
	biz.POST("/employer/applications/:id/hire",
		middleware.RequirePerm(d.Logger, "application:audit"), orderH.Hire)

	biz.GET("/employer/orders",
		middleware.RequirePerm(d.Logger, "order:view"), orderH.ListEmployer)
	biz.GET("/employer/orders/:id",
		middleware.RequirePerm(d.Logger, "order:view"), orderH.GetEmployer)
	biz.POST("/employer/orders/:id/confirm",
		middleware.RequirePerm(d.Logger, "order:settle"), orderH.Confirm)
	biz.POST("/employer/orders/:id/cancel",
		middleware.RequirePerm(d.Logger, "order:cancel"), orderH.CancelEmployer)
	biz.POST("/employer/orders/:id/review",
		middleware.RequirePerm(d.Logger, "review:create"), orderH.ReviewEmployer)

	// Messages (any logged-in user)
	messageH := handlers.NewMessageHandler(d.MessageSvc)
	biz.GET("/messages",
		middleware.RequirePerm(d.Logger, "message:view"), messageH.ListMine)
	biz.POST("/messages/:id/read",
		middleware.RequirePerm(d.Logger, "message:view"), messageH.MarkRead)
	biz.POST("/messages/read-all",
		middleware.RequirePerm(d.Logger, "message:view"), messageH.MarkAllRead)

	// ----- Admin (operation-logged + per-perm) -----
	adminBiz := r.Group(
		"/",
		middleware.Auth(d.Issuer, d.Logger),
		middleware.OperationLog(d.OpSvc, d.Logger),
	)
	adminH := handlers.NewAdminHandler(d.StudentProfileSvc, d.EmployerProfileSvc, d.AgentProfileSvc)

	adminBiz.GET("/admin/categories",
		middleware.RequirePerm(d.Logger, "category:view"), categoryH.ListAdmin)
	adminBiz.POST("/admin/categories",
		middleware.RequirePerm(d.Logger, "category:create"), categoryH.Create)
	adminBiz.PUT("/admin/categories/:id",
		middleware.RequirePerm(d.Logger, "category:update"), categoryH.Update)
	adminBiz.DELETE("/admin/categories/:id",
		middleware.RequirePerm(d.Logger, "category:delete"), categoryH.Delete)

	adminBiz.GET("/admin/jobs",
		middleware.RequirePerm(d.Logger, "job:audit"), jobH.ListPendingAdmin)
	adminBiz.POST("/admin/jobs/:id/audit",
		middleware.RequirePerm(d.Logger, "job:audit"), jobH.Audit)

	adminBiz.GET("/admin/student-certifications",
		middleware.RequirePerm(d.Logger, "cert:audit"), adminH.ListPendingStudentCerts)
	adminBiz.POST("/admin/student-certifications/:id/audit",
		middleware.RequirePerm(d.Logger, "cert:audit"), adminH.AuditStudentCert)
	adminBiz.GET("/admin/employer-certifications",
		middleware.RequirePerm(d.Logger, "cert:audit"), adminH.ListPendingEmployerCerts)
	adminBiz.POST("/admin/employer-certifications/:id/audit",
		middleware.RequirePerm(d.Logger, "cert:audit"), adminH.AuditEmployerCert)
	adminBiz.GET("/admin/agent-certifications",
		middleware.RequirePerm(d.Logger, "cert:audit"), adminH.ListPendingAgentCerts)
	adminBiz.POST("/admin/agent-certifications/:id/audit",
		middleware.RequirePerm(d.Logger, "cert:audit"), adminH.AuditAgentCert)

	adminBiz.GET("/admin/orders",
		middleware.RequirePerm(d.Logger, "order:monitor"), orderH.ListAllAdmin)

	adminBiz.GET("/admin/reviews",
		middleware.RequirePerm(d.Logger, "review:view"), reviewH.ListAllAdmin)
	adminBiz.DELETE("/admin/reviews/:id",
		middleware.RequirePerm(d.Logger, "review:delete"), reviewH.Delete)

	adminBiz.POST("/admin/messages/broadcast",
		middleware.RequirePerm(d.Logger, "message:broadcast"), messageH.Broadcast)
}
