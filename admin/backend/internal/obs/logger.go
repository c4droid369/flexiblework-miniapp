// Package obs owns observability concerns: structured logging and database
// connectivity orchestration. Nothing else in the codebase may import a
// database driver directly.
package obs

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// ctxKey is unexported so other packages cannot collide on context values.
type ctxKey struct{ name string }

var requestIDKey = ctxKey{name: "request_id"}

// NewLogger builds a slog.Logger with the requested level and format. The
// returned handler is wrapped in a ctxHandler that pulls request_id out of the
// context and attaches it to every log record.
func NewLogger(level, format, env string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(strings.ToLower(level))); err != nil {
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl, AddSource: env == "development"}

	var h slog.Handler
	switch strings.ToLower(format) {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, opts)
	default:
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(&ctxHandler{Handler: h})
}

type ctxHandler struct{ slog.Handler }

func (h *ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(requestIDKey).(string); ok && id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

// WithRequestID returns a child context carrying the given request id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFrom extracts the request id set by middleware; empty if absent.
func RequestIDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
