package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/admin-template/backend/internal/api/middleware"
)

// mountProtected builds a 1-route engine whose handler runs after Auth (no
// JWT, so it 401s) + RequirePerm(...). Returns the recorder.
func mountProtected(t *testing.T, perms ...string) *httptest.ResponseRecorder {
	t.Helper()
	logger := slog.Default()
	r := gin.New()
	r.GET(
		"/x",
		middleware.RequirePerm(logger, perms...),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	return rec
}

// mountProtectedWithPerms is like mountProtected but pre-seeds the gin
// context with the user's permission set (RequirePerm reads from there).
func mountProtectedWithPerms(t *testing.T, have, want []string) *httptest.ResponseRecorder {
	t.Helper()
	logger := slog.Default()
	r := gin.New()
	r.GET(
		"/x",
		func(c *gin.Context) {
			c.Set(middleware.PermContextKey, have)
			c.Next()
		},
		middleware.RequirePerm(logger, want...),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	return rec
}

func TestRequirePerm_NoRequired_AllowsThrough(t *testing.T) {
	rec := mountProtected(t /* no perms */)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequirePerm_PermNotResolved_403(t *testing.T) {
	logger := slog.Default()
	r := gin.New()
	r.GET("/x", middleware.RequirePerm(logger, "user:view"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "permissions not resolved")
}

func TestRequirePerm_PermMissing_403(t *testing.T) {
	rec := mountProtectedWithPerms(t, []string{"user:view"}, []string{"user:delete"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequirePerm_PermPresent_200(t *testing.T) {
	rec := mountProtectedWithPerms(t, []string{"user:view", "user:create"}, []string{"user:create"})
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequirePerm_ORSemantics_200(t *testing.T) {
	// any-of: present "user:create" matches even though "user:delete" is missing.
	rec := mountProtectedWithPerms(t, []string{"user:create"}, []string{"user:delete", "user:create"})
	assert.Equal(t, http.StatusOK, rec.Code)
}
