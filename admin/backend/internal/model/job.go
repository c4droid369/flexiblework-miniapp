package model

import (
	"time"
)

// Job is a single posted gig. Status lifecycle:
//   0 草稿 (reserved for future "save draft" flow)
//   1 待审核 → 2 招聘中 (admin approves) | 4 审核未通过 (admin rejects)
//   2 招聘中 → 3 已下架 (employer offline) | 5 已招满 (recruit_count reached)
//
// View/apply counters are denormalized for cheap list rendering. They are
// maintained by the application/order services, not by triggers.
type Job struct {
	Base
	EmployerID        uint64     `gorm:"index;not null"     json:"employer_id"`
	CategoryID        uint64     `gorm:"index;not null"     json:"category_id"`
	Title             string     `gorm:"size:128;not null"  json:"title"`
	Cover             string     `gorm:"size:255"           json:"cover"`
	Description       string     `gorm:"type:text"          json:"description"`
	Requirements      string     `gorm:"type:text"          json:"requirements"`
	SalaryType        int8       `gorm:"not null"           json:"salary_type"`    // 1=时薪 2=日薪 3=周薪 4=月薪 5=按件
	SalaryMin         float64    `gorm:"default:0"          json:"salary_min"`
	SalaryMax         float64    `gorm:"default:0"          json:"salary_max"`
	SalaryUnit        string     `gorm:"size:16"            json:"salary_unit"`
	Location          string     `gorm:"size:255;index"     json:"location"`
	WorkDateType      int8       `gorm:"default:1"          json:"work_date_type"` // 1=长期 2=短期 3=单次
	WorkDateStart     *time.Time `                          json:"work_date_start"`
	WorkDateEnd       *time.Time `                          json:"work_date_end"`
	WorkTimeStart     string     `gorm:"size:16"            json:"work_time_start"` // "HH:MM"
	WorkTimeEnd       string     `gorm:"size:16"            json:"work_time_end"`
	RecruitCount      int        `gorm:"default:1;not null" json:"recruit_count"`
	GenderRequirement int8       `gorm:"default:0"          json:"gender_requirement"` // 0=不限 1=男 2=女
	SettlementType    int8       `gorm:"default:1"          json:"settlement_type"`    // 1=日结 2=周结 3=完工结
	Status            int8       `gorm:"default:1;index"    json:"status"`
	AuditRemark       string     `gorm:"size:255"           json:"audit_remark"`
	AuditedAt         *time.Time `                          json:"audited_at"`
	ViewCount         int        `gorm:"default:0"          json:"view_count"`
	ApplyCount        int        `gorm:"default:0"          json:"apply_count"`
}

func (Job) TableName() string { return "jobs" }
