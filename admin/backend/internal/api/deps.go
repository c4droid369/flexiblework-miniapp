package api

import (
	"log/slog"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/config"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/service"
)

// Deps bundles every dependency the HTTP layer needs. cmd/run.go builds it
// once and passes it to api.New. Adding a new resource means appending one
// *service.X field here — the router can then reference it without changing
// the New signature on every phase.
type Deps struct {
	Cfg     config.Config
	Logger  *slog.Logger
	DB      *gorm.DB
	Issuer  *auth.Issuer
	AuthSvc *service.AuthService
	UserSvc *service.UserService
	RoleSvc *service.RoleService
	MenuSvc *service.MenuService
	OpSvc   *service.OperationLogService
	FileSvc *service.FileService

	// Business services — campus gig work.
	StudentProfileSvc  *service.StudentProfileService
	EmployerProfileSvc *service.EmployerProfileService
	CategorySvc        *service.CategoryService
	JobSvc             *service.JobService
	AppSvc             *service.ApplicationService
	OrderSvc           *service.OrderService
	ReviewSvc          *service.ReviewService
	MessageSvc         *service.MessageService
}
