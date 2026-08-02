// Package middleware groups every Gin middleware used by the API. The order
// of registration in api/server.go is the source of truth for cross-cutting
// behavior.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/admin-template/backend/internal/obs"
)

// RequestID assigns a UUIDv7 to every request, surfaces it via the
// X-Request-ID response header, and stores it on the context for downstream
// loggers and handlers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.Must(uuid.NewV7()).String()
		}
		c.Request = c.Request.WithContext(obs.WithRequestID(c.Request.Context(), id))
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
