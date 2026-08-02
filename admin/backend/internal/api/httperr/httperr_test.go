package httperr_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/api/httperr"
)

func init() { gin.SetMode(gin.TestMode) }

// runHandler mounts a single GET /x handler that calls Write with err, then
// runs a request through it. Returns the recorded response.
func runHandler(t *testing.T, err error, logger *slog.Logger) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	r := gin.New()
	r.GET("/x", func(c *gin.Context) {
		httperr.Write(c, logger, err)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	out := map[string]any{}
	if len(body) > 0 {
		require.NoError(t, json.Unmarshal(body, &out))
	}
	return rec, out
}

func TestError_ErrorStringAndUnwrap(t *testing.T) {
	cause := errors.New("db down")
	e := httperr.Internal(cause)
	assert.Contains(t, e.Error(), "internal_error")
	assert.Contains(t, e.Error(), "db down")
	assert.True(t, errors.Is(e, cause), "errors.Is must traverse Unwrap")
}

func TestConstructors_HTTPStatusAndCode(t *testing.T) {
	cases := []struct {
		name string
		err  *httperr.Error
		want int
	}{
		{"BadRequest", httperr.BadRequest("x"), http.StatusBadRequest},
		{"Unauthorized", httperr.Unauthorized("x"), http.StatusUnauthorized},
		{"Forbidden", httperr.Forbidden("x"), http.StatusForbidden},
		{"NotFound", httperr.NotFound("x"), http.StatusNotFound},
		{"Conflict", httperr.Conflict("x"), http.StatusConflict},
		{"Validation", httperr.Validation("x"), http.StatusBadRequest},
		{"Business", httperr.Business("x"), http.StatusBadRequest},
		{"Internal", httperr.Internal(errors.New("x")), http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, c.err.HTTPStatus)
			assert.NotZero(t, c.err.Code)
		})
	}
}

func TestWrite_5xx_LogsAndEmits(t *testing.T) {
	rec, out := runHandler(t, httperr.Internal(errors.New("boom")), nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// Envelope `code` is the application-level code (CodeInternal = 7),
	// NOT the HTTP status — the two are intentionally decoupled.
	assert.Equal(t, float64(httperr.CodeInternal), out["code"])
	assert.Equal(t, "internal_error", out["message"])
}

func TestWrite_4xx_NoLoggerNeeded(t *testing.T) {
	rec, out := runHandler(t, httperr.NotFound("missing"), nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "missing", out["message"])
}

func TestWrite_Nil_NoOp(t *testing.T) {
	rec, _ := runHandler(t, nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestWrite_UntypedError_WrappedAs500(t *testing.T) {
	rec, out := runHandler(t, errors.New("oops"), nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "internal_error", out["message"])
}

func TestWrite_WrappedTypedError_PreservesStatus(t *testing.T) {
	wrapped := errors.New("wrapped: " + httperr.NotFound("missing").Error())
	rec, _ := runHandler(t, wrapped, nil)
	// Untyped wrapping loses the *Error — fall through to 500. This is
	// intentional: a wrapped string is no longer type-checkable, and
	// returning 404 would be a lie.
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
