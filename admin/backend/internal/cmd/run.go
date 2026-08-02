// Package cmd wires configuration, observability, persistence, and the HTTP
// server together. Keep this thin: anything testable belongs in internal/.
package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/admin-template/backend/internal/api"
	"github.com/admin-template/backend/internal/config"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/obs"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/storage"
	"github.com/admin-template/backend/internal/repository"
	"github.com/admin-template/backend/internal/seed"
	"github.com/admin-template/backend/internal/service"
)

// Execute is the single entrypoint called by cmd/server/main.go. It loads
// config, builds the logger, opens the DB pool, runs migrations, seeds default
// data, wires services, and starts the HTTP server with graceful shutdown.
func Execute(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := obs.NewLogger(cfg.LogLevel, cfg.LogFormat, cfg.Env)
	slog.SetDefault(logger)

	db, err := obs.OpenMySQL(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer func() {
		if err := obs.CloseMySQL(db); err != nil {
			logger.WarnContext(ctx, "close db", slog.Any("err", err))
		}
	}()

	if err := obs.AutoMigrate(ctx, db, model.All()...); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	issuer, err := auth.NewIssuer(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL, cfg.JWTIssuer)
	if err != nil {
		return fmt.Errorf("jwt issuer: %w", err)
	}

	// Repositories — system (RBAC).
	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	opLogRepo := repository.NewOperationLogRepository(db)
	fileRepo := repository.NewFileRepository(db)

	// Repositories — business (campus gig work).
	studentProfileRepo := repository.NewStudentProfileRepository(db)
	employerProfileRepo := repository.NewEmployerProfileRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	jobRepo := repository.NewJobRepository(db)
	appRepo := repository.NewApplicationRepository(db)
	if err := appRepo.EnsureSchema(ctx); err != nil {
		return fmt.Errorf("ensure applications schema: %w", err)
	}
	orderRepo := repository.NewOrderRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	store, err := storage.NewLocalStorage(cfg.StorageLocalDir, cfg.StorageBaseURL)
	if err != nil {
		return fmt.Errorf("storage init: %w", err)
	}

	// Services — system.
	authSvc := service.NewAuthService(userRepo, refreshRepo, roleRepo, menuRepo,
		studentProfileRepo, employerProfileRepo, issuer, cfg.JWTAccessTTL, logger)
	userSvc := service.NewUserService(userRepo, roleRepo)
	roleSvc := service.NewRoleService(roleRepo)
	menuSvc := service.NewMenuService(menuRepo)
	opLogSvc := service.NewOperationLogService(opLogRepo)
	fileSvc := service.NewFileService(fileRepo, store, cfg.UploadMaxSize)

	// Services — business.
	studentProfileSvc := service.NewStudentProfileService(studentProfileRepo, userRepo)
	employerProfileSvc := service.NewEmployerProfileService(employerProfileRepo, userRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	jobSvc := service.NewJobService(jobRepo, categoryRepo, userRepo, employerProfileRepo)
	appSvc := service.NewApplicationService(appRepo, jobRepo, userRepo, studentProfileRepo, messageRepo)
	orderSvc := service.NewOrderService(db, orderRepo, appRepo, appSvc, jobRepo, jobSvc, userRepo, employerProfileRepo, messageRepo)
	reviewSvc := service.NewReviewService(reviewRepo, orderRepo, userRepo)
	messageSvc := service.NewMessageService(messageRepo, userRepo)

	// Seed — idempotent. Creates admin user, roles, menus, sample data.
	if err := seed.Run(ctx, db, logger, seed.Options{
		AdminUsername: cfg.SeedAdminUsername,
		AdminPassword: cfg.SeedAdminPassword,
	}); err != nil {
		logger.WarnContext(ctx, "seed skipped", slog.Any("err", err))
	}

	deps := api.Deps{
		Cfg: cfg, Logger: logger, DB: db, Issuer: issuer,
		AuthSvc: authSvc, UserSvc: userSvc, RoleSvc: roleSvc,
		MenuSvc: menuSvc, OpSvc: opLogSvc, FileSvc: fileSvc,
		StudentProfileSvc:  studentProfileSvc,
		EmployerProfileSvc: employerProfileSvc,
		CategorySvc:        categorySvc,
		JobSvc:             jobSvc,
		AppSvc:             appSvc,
		OrderSvc:           orderSvc,
		ReviewSvc:          reviewSvc,
		MessageSvc:         messageSvc,
	}

	srv := api.New(deps)
	logger.InfoContext(
		ctx, "server starting",
		slog.String("addr", fmt.Sprintf("%d", cfg.ServerPort)),
		slog.String("env", cfg.Env),
	)
	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("server run: %w", err)
	}
	logger.InfoContext(ctx, "server stopped cleanly")
	return nil
}
