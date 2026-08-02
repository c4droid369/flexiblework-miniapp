package pagination_test

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/pkg/pagination"
)

func ctxWithQuery(qs string) *gin.Context {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/?"+qs, nil)
	return c
}

func TestFromGin_DefaultsWhenMissing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)

	p := pagination.FromGin(c)
	assert.Equal(t, pagination.DefaultPage, p.Page)
	assert.Equal(t, pagination.DefaultSize, p.Size)
	assert.Equal(t, 0, p.Off)
	assert.Equal(t, pagination.DefaultSize, p.Limit)
}

func TestFromGin_CustomValues(t *testing.T) {
	c := ctxWithQuery("page=3&size=25")
	p := pagination.FromGin(c)
	assert.Equal(t, 3, p.Page)
	assert.Equal(t, 25, p.Size)
	assert.Equal(t, 50, p.Off)
	assert.Equal(t, 25, p.Limit)
}

func TestFromGin_CapsAtMaxSize(t *testing.T) {
	c := ctxWithQuery("page=1&size=9999")
	p := pagination.FromGin(c)
	assert.Equal(t, pagination.MaxSize, p.Size, "size > MaxSize must be capped")
}

func TestFromGin_IgnoresGarbageValues(t *testing.T) {
	c := ctxWithQuery("page=abc&size=-3")
	p := pagination.FromGin(c)
	assert.Equal(t, pagination.DefaultPage, p.Page)
	assert.Equal(t, pagination.DefaultSize, p.Size)
}

func TestSearchFromGin(t *testing.T) {
	c := ctxWithQuery("keyword=alice")
	s := pagination.SearchFromGin(c)
	assert.Equal(t, "alice", s.Keyword)

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = httptest.NewRequest("GET", "/", nil)
	assert.Empty(t, pagination.SearchFromGin(c2).Keyword)
}

func TestFromGin_RoundTripOffCalculation(t *testing.T) {
	cases := []struct {
		page, size, wantOff int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 10, 20},
		{5, 25, 100},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			c := ctxWithQuery(url.Values{"page": {itoa(tc.page)}, "size": {itoa(tc.size)}}.Encode())
			p := pagination.FromGin(c)
			require.Equal(t, tc.wantOff, p.Off)
		})
	}
}

// itoa — avoid strconv import just for one helper.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
