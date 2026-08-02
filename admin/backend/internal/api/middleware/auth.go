package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/pkg/auth"
)

// Auth validates the Authorization: Bearer <token> header and stores the
// user id and permission codes on the gin context. Per-route permission
// checks use RequirePerm.
func Auth(issuer *auth.Issuer, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearer(c.GetHeader("Authorization"))
		if raw == "" {
			httperr.Write(c, logger, httperr.Unauthorized("missing bearer token"))
			return
		}
		claims, err := issuer.Parse(raw, auth.KindAccess)
		if err != nil {
			httperr.Write(c, logger, httperr.Unauthorized("invalid or expired token"))
			return
		}
		c.Request = c.Request.WithContext(auth.WithUserID(c.Request.Context(), claims.UserID))
		c.Set("user_id", claims.UserID)
		// Permissions are baked into the access token at issue time so the
		// RequirePerm check stays O(1) and DB-free per request.
		c.Set(PermContextKey, claims.Permissions)
		c.Next()
	}
}

func extractBearer(h string) string {
	const prefix = "Bearer "
	if len(h) <= len(prefix) {
		return ""
	}
	if !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}
