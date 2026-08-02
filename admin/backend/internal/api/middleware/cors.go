package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS sets Access-Control-* headers on every request and short-circuits
// OPTIONS preflight with 204. MUST be registered before Auth so preflight
// never reaches the token check.
func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := splitTrim(allowedOrigins, ",")
	allowAll := len(origins) == 1 && origins[0] == "*"

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		switch {
		case allowAll:
			c.Header("Access-Control-Allow-Origin", "*")
		case contains(origins, origin):
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func splitTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := parts[:0]
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func contains(xs []string, target string) bool {
	for _, x := range xs {
		if x == target {
			return true
		}
	}
	return false
}
