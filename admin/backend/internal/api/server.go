// Package api owns the HTTP transport layer: gin engine, middleware chain,
// and handler registration. Handlers live in ./handlers; business logic in
// internal/service. Handlers MUST NOT call the database directly.
package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/admin-template/backend/internal/api/handlers"
	"github.com/admin-template/backend/internal/api/middleware"
)

// Server is the HTTP server handle returned to cmd/run.go for graceful
// shutdown. Fields are unexported — use New + Run.
type Server struct {
	cfg    Deps
	logger *slog.Logger
	srv    *http.Server
}

// New constructs the gin engine, wires every middleware in canonical order,
// and registers the route table. Returns an unstarted Server.
func New(deps Deps) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.MaxMultipartMemory = deps.Cfg.UploadMaxSize

	r.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.Logger(deps.Logger),
		middleware.CORS(deps.Cfg.CORSOrigins),
	)

	// Public, unauthenticated.
	r.GET("/healthz", handlers.Health(deps.DB))

	// Static file serving for local storage (Phase 6).
	r.Static("/files", deps.Cfg.StorageLocalDir)

	api := r.Group("/api/v1")
	RegisterAPIRoutes(api, deps)

	if deps.Cfg.SwaggerEnabled {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return &Server{
		cfg:    deps,
		logger: deps.Logger,
		srv: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", deps.Cfg.ServerHost, deps.Cfg.ServerPort),
			Handler:      r,
			ReadTimeout:  deps.Cfg.ReadTimeout,
			WriteTimeout: deps.Cfg.WriteTimeout,
		},
	}
}

// Run blocks until ctx is canceled, then performs graceful shutdown bounded
// by cfg.ShutdownTimeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.InfoContext(ctx, "http listening", slog.String("addr", s.srv.Addr))
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		s.logger.InfoContext(ctx, "shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Cfg.ShutdownTimeout)
		defer cancel()
		return s.srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
