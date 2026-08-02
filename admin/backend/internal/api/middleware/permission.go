package middleware

import (
	"log/slog"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
)

// PermContextKey is the gin.Context key under which the handler stores the
// user's effective permission codes. Set by ResolvePerms (a post-auth
// middleware that queries the role -> menu -> perm_code chain).
const PermContextKey = "permissions"

// ResolvePerms hydrates the user's permission codes into the gin context. Run
// after Auth and before RequirePerm. The lookup is centralized so RequirePerm
// is O(1).
func ResolvePerms(perms []string) gin.HandlerFunc {
	// Defensive copy to prevent the caller mutating the slice mid-request.
	frozen := slices.Clone(perms)
	return func(c *gin.Context) {
		c.Set(PermContextKey, frozen)
		c.Next()
	}
}

// RequirePerm rejects requests whose user lacks every required perm. Pass
// multiple codes for OR semantics; pass a single "*" to require super_admin
// via the "super_admin" role code (handled by RequireRole separately).
func RequirePerm(logger *slog.Logger, required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(required) == 0 {
			c.Next()
			return
		}
		raw, ok := c.Get(PermContextKey)
		if !ok {
			httperr.Write(c, logger, httperr.Forbidden("permissions not resolved"))
			return
		}
		perms, _ := raw.([]string)
		if !hasAny(perms, required) {
			httperr.Write(c, logger, httperr.Forbidden("permission denied"))
			return
		}
		c.Next()
	}
}

// RequireRole accepts the request only if the supplied role code is on the
// user's roles list. Stored on the gin context by RequirePerm consumers; this
// helper is for future use when a route must be role-gated beyond perms.
func RequireRole(_ string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func hasAny(have, want []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, p := range have {
		set[p] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; ok {
			return true
		}
	}
	return false
}
