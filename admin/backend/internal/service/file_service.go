package service

import (
	"context"
	"errors"
	"mime"
	"path/filepath"
	"strings"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/model"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/pkg/storage"
	"github.com/admin-template/backend/internal/repository"
)

// FileService is the façade for upload, listing, and deletion of files. The
// actual bytes go through the Storage adapter; the row goes to GORM.
type FileService struct {
	repo  repository.FileRepository
	store storage.Storage
	maxSz int64
}

func NewFileService(repo repository.FileRepository, store storage.Storage, maxBytes int64) *FileService {
	return &FileService{repo: repo, store: store, maxSz: maxBytes}
}

// UploadInput is the contract between the handler and the service. The
// handler is responsible for parsing the multipart form; the service is
// responsible for the storage round-trip and the DB row.
type UploadInput struct {
	OriginalName string
	Size         int64
	ContentType  string
	Reader       interface{ Read([]byte) (int, error) }
	UploaderID   uint64
}

// UploadResult is what the handler returns to the frontend.
type UploadResult struct {
	ID   uint64 `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Upload stores the file via the adapter and records its metadata.
func (s *FileService) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	if in.Size > s.maxSz {
		return nil, httperr.BadRequest("file too large")
	}
	if strings.TrimSpace(in.OriginalName) == "" {
		return nil, httperr.BadRequest("missing filename")
	}
	ct := in.ContentType
	if ct == "" {
		ct = mime.TypeByExtension(filepath.Ext(in.OriginalName))
		if ct == "" {
			ct = "application/octet-stream"
		}
	}
	key, err := s.store.Put(ctx, "", &readerAdapter{r: in.Reader}, in.Size, ct)
	if err != nil {
		return nil, httperr.Internal(err)
	}
	row := &model.File{
		Name:         key,
		OriginalName: in.OriginalName,
		Path:         key,
		Size:         in.Size,
		MimeType:     ct,
		Storage:      model.StorageTypeLocal,
		UploaderID:   in.UploaderID,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		// Roll back the storage write to avoid orphaned blobs.
		_ = s.store.Delete(ctx, key)
		return nil, httperr.Internal(err)
	}
	return &UploadResult{ID: row.ID, URL: s.store.GetURL(key), Name: in.OriginalName, Size: in.Size}, nil
}

// List returns a paginated list of uploaded files.
func (s *FileService) List(ctx context.Context, page pagination.Page, search pagination.Search) ([]model.File, int64, error) {
	files, total, err := s.repo.List(ctx, page.Page, page.Size, search.Keyword)
	if err != nil {
		return nil, 0, httperr.Internal(err)
	}
	return files, total, nil
}

// Delete removes the file from storage and the DB row.
func (s *FileService) Delete(ctx context.Context, id uint64) error {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httperr.NotFound("file not found")
		}
		return httperr.Internal(err)
	}
	if err := s.store.Delete(ctx, f.Path); err != nil {
		return httperr.Internal(err)
	}
	return s.repo.Delete(ctx, id)
}

// readerAdapter wraps an io.Reader-shaped value without requiring us to
// import io in handler code. Returns io.EOF when the underlying read returns 0.
type readerAdapter struct {
	r interface{ Read([]byte) (int, error) }
}

func (a *readerAdapter) Read(p []byte) (int, error) { return a.r.Read(p) }
