package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/api/middleware"
	"github.com/admin-template/backend/internal/pkg/auth"
)

func init() { gin.SetMode(gin.TestMode) }

const testSecret = "this-is-a-32-byte-test-secret-ok"

// runAuth spins up a minimal Gin engine behind middleware.Auth, sends a
// request with the supplied Authorization header, and returns the recorder.
// The handler echoes the user id from the request context.
func runAuth(t *testing.T, header string) *httptest.ResponseRecorder {
	t.Helper()
	iss, err := auth.NewIssuer(testSecret, time.Hour, time.Hour, "test-iss")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.Use(middleware.Auth(iss, slog.Default()))
	r.GET("/x", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"uid": auth.UserIDFrom(c.Request.Context())})
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	r.ServeHTTP(rec, req)
	return rec
}

func TestAuth_MissingHeader_401(t *testing.T) {
	rec := runAuth(t, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_NonBearerScheme_401(t *testing.T) {
	rec := runAuth(t, "Basic dXNlcjpwYXNz")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_InvalidToken_401(t *testing.T) {
	rec := runAuth(t, "Bearer not-a-jwt")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_ValidToken_PassesAndSetsUID(t *testing.T) {
	iss, _ := auth.NewIssuer(testSecret, time.Hour, time.Hour, "test-iss")
	tok, _, err := iss.IssueAccess(99, []string{"user:view"})
	require.NoError(t, err)

	rec := runAuth(t, "Bearer "+tok)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"uid":99`)
}

func TestAuth_TokenFromWrongIssuer_401(t *testing.T) {
	other, _ := auth.NewIssuer(testSecret, time.Hour, time.Hour, "other-iss")
	tok, _, err := other.IssueAccess(1, nil)
	require.NoError(t, err)

	rec := runAuth(t, "Bearer "+tok)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
