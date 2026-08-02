package dto

// BatchDeleteReq is the body of POST /system/<resource>/batch-delete.
type BatchDeleteReq struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// IDResp is the minimal response after creating a single resource.
type IDResp struct {
	ID uint64 `json:"id"`
}
