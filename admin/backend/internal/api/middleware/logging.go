package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger writes one structured line per request after c.Next() returns. The
// sloglint rule requires typed attrs — keep slog.String / slog.Int, not Any.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.InfoContext(
			c.Request.Context(), "http request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int("bytes", c.Writer.Size()),
			slog.Duration("elapsed", time.Since(start)),
			slog.String("ip", c.ClientIP()),
		)
	}
}
