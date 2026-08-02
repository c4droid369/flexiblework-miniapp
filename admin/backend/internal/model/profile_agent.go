package model

import (
	"time"
)

// AgentProfile extends the user record for the "校园代理" role. Distinct
// from EmployerProfile in that it carries agent-specific concerns:
//   - referral_code:  unique per-agent code, surfaces in invite links
//   - bank_*:         reserved for commission payouts (out of scope for v1)
//   - total_referrals/total_earnings: denormalized counters
//
// Cert workflow mirrors EmployerProfile (submit → review → approved/rejected).
// Once approved, the agent can post jobs through the same /employer/jobs
// endpoints — JobService.Create checks both profile types when verifying the
// "has a certified business identity" precondition.
type AgentProfile struct {
	Base
	UserID         uint64     `gorm:"uniqueIndex;not null" json:"user_id"`
	RealName       string     `gorm:"size:64"             json:"real_name"`
	Phone          string     `gorm:"size:32"             json:"phone"`
	Wechat         string     `gorm:"size:64"             json:"wechat"`
	IDCardNo       string     `gorm:"size:32"             json:"-"`            // 敏感
	IDCardFront    string     `gorm:"size:255"            json:"id_card_front"`
	IDCardBack     string     `gorm:"size:255"            json:"id_card_back"`
	CampusCard     string     `gorm:"size:255"            json:"campus_card"`  // 校园卡 / 学生证
	CertStatus     int8       `gorm:"default:0;index"     json:"cert_status"` // 0=未认证 1=审核中 2=已通过 3=已拒绝
	CertRemark     string     `gorm:"size:255"            json:"cert_remark"`
	CertifiedAt    *time.Time `                           json:"certified_at"`
	Bio            string     `gorm:"size:500"            json:"bio"`
	ReferralCode   string     `gorm:"size:16;uniqueIndex" json:"referral_code"` // 推荐码
	BankAccount    string     `gorm:"size:64"             json:"bank_account"`  // 预留:佣金结算
	BankName       string     `gorm:"size:64"             json:"bank_name"`     // 预留:开户行
	TotalReferrals int        `gorm:"default:0"           json:"total_referrals"`
	TotalEarnings  float64    `gorm:"default:0"           json:"total_earnings"` // 预留:累计佣金
	Rating         float64    `gorm:"default:5.0"         json:"rating"`
	TotalJobs      int        `gorm:"default:0"           json:"total_jobs"`
}

// TableName overrides GORM's pluralization.
func (AgentProfile) TableName() string { return "agent_profiles" }