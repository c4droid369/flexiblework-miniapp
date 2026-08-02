package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/pkg/storage"
)

func newStore(t *testing.T) (*storage.LocalStorage, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := storage.NewLocalStorage(dir, "/files")
	require.NoError(t, err)
	return s, dir
}

func TestNewLocalStorage_CreatesBaseDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "uploads")
	_, err := storage.NewLocalStorage(dir, "/files")
	require.NoError(t, err)
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNewLocalStorage_EmptyBaseDirRejected(t *testing.T) {
	_, err := storage.NewLocalStorage("", "/files")
	require.Error(t, err)
}

func TestLocalStorage_PutGetRoundTrip(t *testing.T) {
	s, dir := newStore(t)
	ctx := context.Background()

	key, err := s.Put(ctx, "", strings.NewReader("hello world"), 11, "text/plain")
	require.NoError(t, err)
	assert.Contains(t, key, "/") // YYYY/MM/DD/<rand>

	// File actually exists on disk under <dir>/<key>.
	full := filepath.Join(dir, filepath.FromSlash(key))
	body, err := os.ReadFile(full)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(body))

	// Get returns the same bytes through the reader API.
	r, err := s.Get(ctx, key)
	require.NoError(t, err)
	defer func() { _ = r.Close() }()
	buf := make([]byte, 64)
	n, _ := r.Read(buf)
	assert.Equal(t, "hello world", string(buf[:n]))
}

func TestLocalStorage_GetMissingReturnsErrNotFound(t *testing.T) {
	s, _ := newStore(t)
	_, err := s.Get(context.Background(), "nope/missing.txt")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestLocalStorage_DeleteIsIdempotent(t *testing.T) {
	s, dir := newStore(t)
	ctx := context.Background()

	key, err := s.Put(ctx, "", strings.NewReader("bye"), 3, "text/plain")
	require.NoError(t, err)

	require.NoError(t, s.Delete(ctx, key))
	// Second delete must NOT error — missing keys are not an error.
	require.NoError(t, s.Delete(ctx, key))

	// File is gone from disk.
	_, err = os.Stat(filepath.Join(dir, filepath.FromSlash(key)))
	assert.True(t, os.IsNotExist(err))
}

func TestLocalStorage_GetURL(t *testing.T) {
	s, _ := newStore(t)
	assert.Equal(t, "/files/2026/07/x.png", s.GetURL("2026/07/x.png"))
	assert.Equal(t, "/files/x.png", s.GetURL("/x.png"))
	assert.Empty(t, s.GetURL(""))
}

func TestLocalStorage_DefaultURLPrefixWhenEmpty(t *testing.T) {
	s, err := storage.NewLocalStorage(t.TempDir(), "")
	require.NoError(t, err)
	assert.Equal(t, "/files/x.png", s.GetURL("x.png"))
}

func TestLocalStorage_ContentTypePicksExtension(t *testing.T) {
	s, _ := newStore(t)
	ctx := context.Background()

	for _, ct := range []string{"image/jpeg", "image/png", "application/pdf"} {
		key, err := s.Put(ctx, "", strings.NewReader("x"), 1, ct)
		require.NoError(t, err)
		assert.Contains(t, key, ".", "ct=%s should map to an extension", ct)
	}
}
