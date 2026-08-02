package httperr

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// WriteRaw wraps an error and calls Write with the supplied logger. Use this
// when the service wants to pass through a non-typed error from the
// repository layer without rewriting it.
func WriteRaw(err error) error {
	if err == nil {
		return nil
	}
	var he *Error
	if errors.As(err, &he) {
		return err
	}
	return Internal(err)
}

// WriteNoop is the gin-handler variant that uses a no-op logger. Prefer the
// (logger, err) signature in middleware so 5xx errors get logged.
func WriteNoop(c *gin.Context, err error) {
	Write(c, slog.Default(), err)
}
