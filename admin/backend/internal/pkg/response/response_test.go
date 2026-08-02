package response_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/pkg/response"
)

func init() { gin.SetMode(gin.TestMode) }

func decode(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	body, _ := io.ReadAll(rec.Body)
	out := response.Envelope{}
	require.NoError(t, json.Unmarshal(body, &out))
	return out
}

func TestOK(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/", nil)

	response.OK(c, map[string]string{"hello": "world"})

	assert.Equal(t, http.StatusOK, rec.Code)
	env := decode(t, rec)
	assert.Equal(t, 0, env.Code)
	assert.Equal(t, "ok", env.Message)
	m, _ := env.Data.(map[string]any)
	assert.Equal(t, "world", m["hello"])
}

func TestOKCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/", nil)

	response.OKCreated(c, "id-1")

	assert.Equal(t, http.StatusCreated, rec.Code)
	env := decode(t, rec)
	assert.Equal(t, "created", env.Message)
}

func TestOKPage(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/", nil)

	list := []int{1, 2, 3}
	response.OKPage(c, list, 42, 2, 10)

	env := decode(t, rec)
	data, _ := env.Data.(map[string]any)
	assert.Equal(t, "ok", env.Message)
	assert.Equal(t, float64(42), data["total"])
	assert.Equal(t, float64(2), data["page"])
	assert.Equal(t, float64(10), data["size"])
}

func TestFail(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/", nil)

	response.Fail(c, http.StatusBadRequest, 1001, "bad")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	env := decode(t, rec)
	assert.Equal(t, 1001, env.Code)
	assert.Equal(t, "bad", env.Message)
}
