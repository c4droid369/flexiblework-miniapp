package exporter_test

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/pkg/exporter"
)

func init() { gin.SetMode(gin.TestMode) }

// newCtx returns a gin.Context whose Writer is the underlying
// httptest.ResponseRecorder, so tests can read the response body.
func newCtx() (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	return c, rec
}

func TestCSV_BasicRoundTrip(t *testing.T) {
	c, rec := newCtx()
	sheet := exporter.Sheet{
		Name:    "Users",
		Headers: []string{"ID", "Name", "Active"},
		Rows: [][]any{
			{1, "alice", true},
			{2, "bob", false},
		},
	}
	require.NoError(t, exporter.CSV(c, "users", sheet))

	assert.Equal(t, "text/csv; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Header().Get("Content-Disposition"), `filename="users_`)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), `.csv"`)

	r := csv.NewReader(bytes.NewReader(rec.Body.Bytes()))
	records, err := r.ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 3, "1 header + 2 data rows")

	assert.Equal(t, []string{"ID", "Name", "Active"}, records[0])
	assert.Equal(t, []string{"1", "alice", "true"}, records[1])
	assert.Equal(t, []string{"2", "bob", "false"}, records[2])
}

func TestCSV_StringifiesTimeAndNil(t *testing.T) {
	c, rec := newCtx()
	ts := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	sheet := exporter.Sheet{
		Name:    "T",
		Headers: []string{"x", "y"},
		Rows:    [][]any{{nil, ts}, {"plain", "x"}},
	}
	require.NoError(t, exporter.CSV(c, "t", sheet))

	r := csv.NewReader(bytes.NewReader(rec.Body.Bytes()))
	records, _ := r.ReadAll()
	require.Len(t, records, 3)
	assert.Equal(t, "", records[1][0], "nil cell becomes empty string")
	assert.Equal(t, "2026-07-22T12:00:00Z", records[1][1], "time formatted as RFC3339")
	assert.Equal(t, "plain", records[2][0])
}

func TestExcel_BasicRoundTrip(t *testing.T) {
	c, rec := newCtx()
	sheet := exporter.Sheet{
		Name:    "S",
		Headers: []string{"A", "B"},
		Rows:    [][]any{{1, "x"}, {2, "y"}},
	}
	require.NoError(t, exporter.Excel(c, "out", sheet))

	assert.Contains(t, rec.Header().Get("Content-Type"), "spreadsheetml.sheet")
	assert.Contains(t, rec.Header().Get("Content-Disposition"), `.xlsx`)

	bs := rec.Body.Bytes()
	assert.Greater(t, len(bs), 100, "xlsx should be non-trivial in size")
	// XLSX is a ZIP archive — magic bytes are 'PK'.
	assert.Equal(t, byte('P'), bs[0])
	assert.Equal(t, byte('K'), bs[1])
}

func TestStringify_VariousTypes(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sheet := exporter.Sheet{
		Name:    "V",
		Headers: []string{"v"},
		Rows: [][]any{
			{nil},
			{"plain"},
			{42},
			{3.14},
			{ts},
		},
	}
	c, rec := newCtx()
	require.NoError(t, exporter.CSV(c, "v", sheet))

	body := rec.Body.Bytes()
	assert.NotEmpty(t, body)
	assert.Contains(t, string(body), "plain")
	assert.Contains(t, string(body), "42")
	assert.Contains(t, string(body), "3.14")
}
