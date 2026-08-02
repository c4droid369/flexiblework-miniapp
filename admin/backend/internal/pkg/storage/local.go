package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage writes files under baseDir in a YYYY/MM/DD/<random>.<ext>
// layout. baseDir is the absolute filesystem path; URLPrefix is what callers
// see (e.g., "/files").
type LocalStorage struct {
	baseDir   string
	urlPrefix string
}

// NewLocalStorage creates the base directory if it doesn't exist.
func NewLocalStorage(baseDir, urlPrefix string) (*LocalStorage, error) {
	if baseDir == "" {
		return nil, errors.New("storage: baseDir required")
	}
	if urlPrefix == "" {
		urlPrefix = "/files"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: mkdir base: %w", err)
	}
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}
	return &LocalStorage{baseDir: abs, urlPrefix: strings.TrimRight(urlPrefix, "/")}, nil
}

// Put stores the file and returns the storage-side key.
func (s *LocalStorage) Put(_ context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	if key == "" {
		key = generateKey(contentType)
	}
	full := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	f, err := os.OpenFile(full, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return key, nil
}

// Get opens the file for reading. Returns ErrNotFound if missing.
func (s *LocalStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	full := filepath.Join(s.baseDir, filepath.FromSlash(key))
	f, err := os.Open(full)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	return f, err
}

// Delete removes the file. Missing files are not an error.
func (s *LocalStorage) Delete(_ context.Context, key string) error {
	full := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// GetURL returns the public URL the file is served from.
func (s *LocalStorage) GetURL(key string) string {
	if key == "" {
		return ""
	}
	return s.urlPrefix + "/" + strings.TrimLeft(key, "/")
}

func generateKey(contentType string) string {
	now := time.Now().UTC()
	ext := extFromContentType(contentType)
	rnd := make([]byte, 16)
	_, _ = rand.Read(rnd)
	return fmt.Sprintf("%04d/%02d/%02d/%s%s",
		now.Year(), now.Month(), now.Day(),
		hex.EncodeToString(rnd), ext)
}

func extFromContentType(ct string) string {
	switch strings.ToLower(ct) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "text/plain":
		return ".txt"
	case "application/zip":
		return ".zip"
	default:
		return ""
	}
}
