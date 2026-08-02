package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/admin-template/backend/internal/api/httperr"
	"github.com/admin-template/backend/internal/pkg/auth"
	"github.com/admin-template/backend/internal/pkg/pagination"
	"github.com/admin-template/backend/internal/pkg/response"
	"github.com/admin-template/backend/internal/service"
)

// UploadHandler exposes /upload and /files-list endpoints.
type UploadHandler struct{ svc *service.FileService }

func NewUploadHandler(svc *service.FileService) *UploadHandler { return &UploadHandler{svc: svc} }

// Upload godoc
// @Summary      上传文件
// @Tags         Files
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "文件"
// @Success      200   {object}  response.Envelope{data=service.UploadResult}
// @Failure      400   {object}  response.Envelope
// @Router       /upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		httperr.Write(c, nil, httperr.BadRequest("missing file field"))
		return
	}
	src, err := fh.Open()
	if err != nil {
		httperr.Write(c, nil, httperr.Internal(err))
		return
	}
	defer func() { _ = src.Close() }()

	uid := auth.UserIDFrom(c.Request.Context())
	res, err := h.svc.Upload(c.Request.Context(), service.UploadInput{
		OriginalName: fh.Filename,
		Size:         fh.Size,
		ContentType:  fh.Header.Get("Content-Type"),
		Reader:       src,
		UploaderID:   uid,
	})
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, res)
}

// List godoc
// @Summary      已上传文件列表
// @Tags         Files
// @Produce      json
// @Security     BearerAuth
// @Param        page     query     int  false  "页码"
// @Param        size     query     int  false  "每页大小"
// @Param        keyword  query     string  false  "模糊搜索原始名/存储名"
// @Success      200      {object}  response.Envelope{data=response.PageData}
// @Router       /files-list [get]
func (h *UploadHandler) List(c *gin.Context) {
	page := pagination.FromGin(c)
	search := pagination.SearchFromGin(c)
	files, total, err := h.svc.List(c.Request.Context(), page, search)
	if err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OKPage(c, files, total, page.Page, page.Size)
}

// Delete godoc
// @Summary      删除文件
// @Tags         Files
// @Security     BearerAuth
// @Param        id   path      int  true  "文件 ID"
// @Success      200  {object}  response.Envelope
// @Router       /files-list/{id} [delete]
func (h *UploadHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httperr.Write(c, nil, err)
		return
	}
	response.OK(c, nil)
}
