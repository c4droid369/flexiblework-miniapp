// Package oplog provides the annotation-driven operation log extractor. The
// middleware uses these helpers to capture (action, method, path, params)
// and persist them through the OperationLogService.
package oplog

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/model"
)

// MaxBodyBytes caps the request body we record. Larger uploads get truncated
// to keep the log table sane.
const MaxBodyBytes = 4 * 1024

// Build reads the gin context after the handler runs and assembles an
// OperationLog entry ready to persist. The middleware must call this from a
// deferred function so latency captures the full handler execution.
func Build(c *gin.Context, userID uint64, username string, start time.Time, body string) *model.OperationLog {
	status := c.Writer.Status()
	latency := time.Since(start)
	entry := &model.OperationLog{
		UserID:         userID,
		Username:       username,
		Action:         c.Request.Method + " " + c.FullPath(),
		Method:         c.Request.Method,
		Path:           c.Request.URL.Path,
		IP:             c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestBody:    sanitize(body),
		ResponseStatus: status,
		LatencyMS:      latency.Milliseconds(),
		CreatedAt:      time.Now(),
	}
	if len(c.Errors) > 0 {
		entry.ErrorMessage = truncate(c.Errors.String(), 1024)
	}
	return entry
}

func sanitize(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > MaxBodyBytes {
		body = body[:MaxBodyBytes] + "...[truncated]"
	}
	return body
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// ReadBody reads the full request body and resets it for downstream reads.
// Returns "" if the body is not text-shaped (multipart, binary).
func ReadBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	ct := c.ContentType()
	if strings.HasPrefix(ct, "multipart/") || strings.HasPrefix(ct, "application/octet-stream") {
		return ""
	}
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))
	return string(buf)
}
