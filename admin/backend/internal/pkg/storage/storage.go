// Package storage defines the Storage interface and a local-filesystem
// implementation. Adding OSS / MinIO / S3 means writing another file in this
// package — handlers and services don't change.
package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrNotFound is returned when the requested key does not exist.
var ErrNotFound = errors.New("storage: not found")

// Storage is the contract every backend implementation must satisfy. Keys
// are adapter-side identifiers (e.g., "2026/01/22/abc123.jpg"), NOT public
// URLs. Public URLs are produced by GetURL().
type Storage interface {
	// Put writes the contents of r under key. Returns the storage-side key.
	// Implementations should generate content-addressable keys internally
	// when key is empty.
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error)

	// Get opens a reader for the given key. Returns ErrNotFound if missing.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the key. Missing keys MUST NOT error.
	Delete(ctx context.Context, key string) error

	// GetURL returns a publicly fetchable URL for the key. LocalStorage
	// returns a relative path served by the /files static route.
	GetURL(key string) string
}

// ObjectMeta describes an uploaded blob. Returned by Put when the caller
// wants more than the key.
type ObjectMeta struct {
	Key         string
	Size        int64
	ContentType string
	UploadedAt  time.Time
}
