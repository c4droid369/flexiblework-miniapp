package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/oplog"
	"github.com/admin-template/backend/internal/service"
)

// OperationLog records one row per request after the handler returns. Mount
// after Auth so the user id is on the context. Failures are logged and
// swallowed — a logging error must never fail the user request.
func OperationLog(svc *service.OperationLogService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		body := oplog.ReadBody(c)

		c.Next()

		uid := auth.UserIDFrom(c.Request.Context())
		username := lookupUsername(c, uid)
		entry := oplog.Build(c, uid, username, start, body)
		if err := svc.Record(c.Request.Context(), entry); err != nil {
			logger.WarnContext(
				c.Request.Context(), "operation log write failed",
				slog.Any("err", err),
				slog.String("path", c.Request.URL.Path),
			)
		}
	}
}

// lookupUsername prefers a username stashed on the gin context by the auth
// service in a future improvement; falls back to "" for now.
func lookupUsername(c *gin.Context, _ uint64) string {
	if v, ok := c.Get("username"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
