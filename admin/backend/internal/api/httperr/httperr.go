// Package httperr funnels every handler error through one writer so the wire
// shape stays consistent and the error code mapping is in one place.
package httperr

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/pkg/response"
)

// Code is the application-level error code returned to the client.
type Code int

const (
	CodeUnknown Code = iota
	CodeBadRequest
	CodeUnauthorized
	CodeForbidden
	CodeNotFound
	CodeConflict
	CodeValidation
	CodeInternal
	CodeBusiness // catch-all for service-layer domain errors
)

// Error is a typed error carrying an HTTP status, application code, and
// message. Anything not conforming to this interface is treated as internal.
type Error struct {
	HTTPStatus int
	Code       Code
	Message    string
	Err        error // wrapped cause for logs; never returned over the wire
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap exposes the cause for errors.Is/As chains.
func (e *Error) Unwrap() error { return e.Err }

// New constructs an Error with the given HTTP status and code.
func New(httpStatus int, code Code, msg string) *Error {
	return &Error{HTTPStatus: httpStatus, Code: code, Message: msg}
}

// Wrap attaches a cause to an Error.
func Wrap(httpStatus int, code Code, msg string, err error) *Error {
	return &Error{HTTPStatus: httpStatus, Code: code, Message: msg, Err: err}
}

// BadRequest returns a 400 Error with CodeBadRequest.
func BadRequest(msg string) *Error { return New(http.StatusBadRequest, CodeBadRequest, msg) }

// Unauthorized returns a 401 Error with CodeUnauthorized.
func Unauthorized(msg string) *Error { return New(http.StatusUnauthorized, CodeUnauthorized, msg) }

// Forbidden returns a 403 Error with CodeForbidden.
func Forbidden(msg string) *Error { return New(http.StatusForbidden, CodeForbidden, msg) }

// NotFound returns a 404 Error with CodeNotFound.
func NotFound(msg string) *Error { return New(http.StatusNotFound, CodeNotFound, msg) }

// Conflict returns a 409 Error with CodeConflict.
func Conflict(msg string) *Error { return New(http.StatusConflict, CodeConflict, msg) }

// Validation returns a 400 Error with CodeValidation.
func Validation(msg string) *Error { return New(http.StatusBadRequest, CodeValidation, msg) }

// Internal wraps an error in a 500 Error with CodeInternal. The original
// error is preserved for log inspection via errors.Unwrap.
func Internal(err error) *Error {
	return Wrap(http.StatusInternalServerError, CodeInternal, "internal_error", err)
}

// Business returns a 400 Error with CodeBusiness for domain-rule violations
// that aren't strictly validation problems.
func Business(msg string) *Error { return New(http.StatusBadRequest, CodeBusiness, msg) }

// Write inspects err, extracts *Error if present, and writes the appropriate
// envelope. If err is not *Error, it logs and writes a generic 500.
//
// logger may be nil — it falls back to slog.Default(). This keeps the
// handler call sites short (`httperr.Write(c, nil, err)`) when the handler
// has no logger of its own; the request still gets logged via the default
// logger set in cmd/run.go.
func Write(c *gin.Context, logger *slog.Logger, err error) {
	if err == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	var he *Error
	if errors.As(err, &he) {
		if he.HTTPStatus >= 500 {
			logger.ErrorContext(
				c.Request.Context(), "handler error",
				slog.Int("status", he.HTTPStatus),
				slog.Int("code", int(he.Code)),
				slog.String("msg", he.Message),
				slog.Any("cause", he.Err),
			)
		}
		response.Fail(c, he.HTTPStatus, int(he.Code), he.Message)
		return
	}
	logger.ErrorContext(c.Request.Context(), "untyped error", slog.Any("err", err))
	response.Fail(c, http.StatusInternalServerError, int(CodeInternal), "internal_error")
}
