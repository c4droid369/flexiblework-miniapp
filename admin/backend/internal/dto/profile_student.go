package dto

import "time"

// StudentProfileResp is the wire shape for a student profile. IDCardNo is
// masked — only the last 4 chars are returned, the rest become asterisks —
// because the field is sensitive and the frontend only needs it for display.
type StudentProfileResp struct {
	ID            uint64     `json:"id"`
	UserID        uint64     `json:"user_id"`
	RealName      string     `json:"real_name"`
	Gender        int8       `json:"gender"`
	School        string     `json:"school"`
	College       string     `json:"college"`
	Major         string     `json:"major"`
	StudentNo     string     `json:"student_no"`
	IDCardNoMask  string     `json:"id_card_no_mask"`
	IDCardFront   string     `json:"id_card_front"`
	IDCardBack    string     `json:"id_card_back"`
	StudentCard   string     `json:"student_card"`
	CertStatus    int8       `json:"cert_status"`
	CertRemark    string     `json:"cert_remark"`
	CertifiedAt   *time.Time `json:"certified_at"`
	Bio           string     `json:"bio"`
	Skills        string     `json:"skills"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// UpsertStudentProfileReq is the body for POST /student/profile. Pointer
// fields let the client send a partial update without zeroing existing data.
type UpsertStudentProfileReq struct {
	RealName     *string `json:"real_name"  binding:"omitempty,max=64"`
	Gender       *int8   `json:"gender"     binding:"omitempty,oneof=0 1 2"`
	School       *string `json:"school"     binding:"omitempty,max=128"`
	College      *string `json:"college"    binding:"omitempty,max=128"`
	Major        *string `json:"major"      binding:"omitempty,max=128"`
	StudentNo    *string `json:"student_no" binding:"omitempty,max=64"`
	IDCardNo     *string `json:"id_card_no" binding:"omitempty,len=18"`
	IDCardFront  *string `json:"id_card_front"  binding:"omitempty,max=255"`
	IDCardBack   *string `json:"id_card_back"   binding:"omitempty,max=255"`
	StudentCard  *string `json:"student_card"   binding:"omitempty,max=255"`
	Bio          *string `json:"bio"            binding:"omitempty,max=500"`
	Skills       *string `json:"skills"`
}

// SubmitStudentCertificationReq is the body for POST /student/certification.
// It re-submits the supporting images; cert_status becomes 审核中.
type SubmitStudentCertificationReq struct {
	IDCardFront string `json:"id_card_front" binding:"required,max=255"`
	IDCardBack  string `json:"id_card_back"  binding:"required,max=255"`
	StudentCard string `json:"student_card"  binding:"required,max=255"`
}
