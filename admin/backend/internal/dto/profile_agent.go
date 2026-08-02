package dto

import "time"

// AgentProfileResp is the wire shape for an agent profile. ReferralCode and
// rating are surfaced for the agent's own dashboard; bank fields are masked
// in non-admin responses (out of scope here — they're already optional and
// the agent's own dashboard is the only consumer in v1).
type AgentProfileResp struct {
	ID             uint64     `json:"id"`
	UserID         uint64     `json:"user_id"`
	RealName       string     `json:"real_name"`
	Phone          string     `json:"phone"`
	Wechat         string     `json:"wechat"`
	IDCardNoMask   string     `json:"id_card_no_mask"`
	IDCardFront    string     `json:"id_card_front"`
	IDCardBack     string     `json:"id_card_back"`
	CampusCard     string     `json:"campus_card"`
	CertStatus     int8       `json:"cert_status"`
	CertRemark     string     `json:"cert_remark"`
	CertifiedAt    *time.Time `json:"certified_at"`
	Bio            string     `json:"bio"`
	ReferralCode   string     `json:"referral_code"`
	BankAccount    string     `json:"bank_account"`
	BankName       string     `json:"bank_name"`
	TotalReferrals int        `json:"total_referrals"`
	TotalEarnings  float64    `json:"total_earnings"`
	Rating         float64    `json:"rating"`
	TotalJobs      int        `json:"total_jobs"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// UpsertAgentProfileReq is the body for POST /agent/profile. Pointer
// fields let the client send a partial update.
type UpsertAgentProfileReq struct {
	RealName    *string `json:"real_name"    binding:"omitempty,max=64"`
	Phone       *string `json:"phone"       binding:"omitempty,max=32"`
	Wechat      *string `json:"wechat"      binding:"omitempty,max=64"`
	IDCardNo    *string `json:"id_card_no"   binding:"omitempty,len=18"`
	IDCardFront *string `json:"id_card_front" binding:"omitempty,max=255"`
	IDCardBack  *string `json:"id_card_back"  binding:"omitempty,max=255"`
	CampusCard  *string `json:"campus_card"   binding:"omitempty,max=255"`
	Bio         *string `json:"bio"           binding:"omitempty,max=500"`
}

// SubmitAgentCertificationReq is the body for POST /agent/certification.
// Re-submits the identity + campus card and flips cert_status to 1.
type SubmitAgentCertificationReq struct {
	IDCardFront string `json:"id_card_front" binding:"required,max=255"`
	IDCardBack  string `json:"id_card_back"  binding:"required,max=255"`
	CampusCard  string `json:"campus_card"   binding:"required,max=255"`
	RealName    string `json:"real_name"    binding:"required,max=64"`
	Phone       string `json:"phone"         binding:"required,max=32"`
}