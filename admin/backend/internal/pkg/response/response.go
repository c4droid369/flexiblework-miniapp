// Package response defines the unified envelope every handler writes. Keeping
// the shape consistent means the frontend can rely on a single response type.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the single response shape across the API. Successful calls
// leave Code == 0 and Error empty; failures set both.
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageData wraps a list response with pagination metadata.
type PageData struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// OK writes a 200 envelope. Use OKCreated (201) for new resources.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Message: "ok", Data: data})
}

// OKCreated writes a 201 envelope.
func OKCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Code: 0, Message: "created", Data: data})
}

// OKPage writes a 200 envelope with PageData. Pass total from the COUNT(*).
func OKPage(c *gin.Context, list any, total int64, page, size int) {
	c.JSON(http.StatusOK, Envelope{
		Code: 0, Message: "ok",
		Data: PageData{List: list, Total: total, Page: page, Size: size},
	})
}

// Fail writes a non-2xx envelope. The HTTP status and the envelope Code can
// differ — e.g. HTTP 400 with code 1001 for a validation failure.
func Fail(c *gin.Context, httpStatus, code int, msg string) {
	c.AbortWithStatusJSON(httpStatus, Envelope{Code: code, Message: msg})
}
