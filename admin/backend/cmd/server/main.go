// Package main is the binary entrypoint. Anything beyond flag parsing lives in
// internal/cmd. This file MUST stay ≤ 50 LOC.
//
// @title                       Admin Template API
// @version                     1.0
// @description                 RBAC admin backend — auth, users/roles/menus CRUD, operation log, file upload, Excel/CSV export.
// @BasePath                    /api/v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/admin-template/backend/internal/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		slog.Error("fatal", slog.Any("err", err))
		os.Exit(1)
	}
}
