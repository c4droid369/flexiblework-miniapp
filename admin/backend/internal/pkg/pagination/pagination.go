// Package pagination parses list query params and computes offset/limit for
// the repository layer. Centralized so every resource applies the same rules.
package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Defaults bound the page size to prevent accidental SELECT * of millions of
// rows. Tune in one place rather than per-handler.
const (
	DefaultPage = 1
	DefaultSize = 10
	MaxSize     = 200
)

type Page struct {
	Page  int
	Size  int
	Off   int
	Limit int
}

// FromGin reads `page` and `size` query params, applies defaults, and returns
// a Page ready to feed into GORM's Offset/Limit.
func FromGin(c *gin.Context) Page {
	p := Page{Page: DefaultPage, Size: DefaultSize}
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		p.Page = v
	}
	if v, err := strconv.Atoi(c.Query("size")); err == nil && v > 0 {
		p.Size = v
	}
	if p.Size > MaxSize {
		p.Size = MaxSize
	}
	p.Off = (p.Page - 1) * p.Size
	p.Limit = p.Size
	return p
}

// Search represents an optional fuzzy-match search term against a single
// column. Frontend sends `keyword=foo`; repository decides which columns to
// LIKE against.
type Search struct {
	Keyword string
}

// SearchFromGin reads the optional `keyword` query param.
func SearchFromGin(c *gin.Context) Search {
	return Search{Keyword: c.Query("keyword")}
}
