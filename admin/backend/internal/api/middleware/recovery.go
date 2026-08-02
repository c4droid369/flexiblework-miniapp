package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/pkg/response"
)

// Recovery turns panics into 500 responses with a stable envelope shape and
// logs the stack trace at error level. Place it after RequestID so the panic
// log carries the request id.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorContext(
					c.Request.Context(), "panic recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				if !c.Writer.Written() {
					response.Fail(c, http.StatusInternalServerError, 500, "internal_error")
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
