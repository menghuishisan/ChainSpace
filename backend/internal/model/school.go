package model

import (
	"time"
)

// School 学校
type School struct {
	BaseModel
	Name        string     `gorm:"size:100;not null" json:"name"`
	Code        string     `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Logo        string     `gorm:"size:500" json:"logo"`
	Address     string     `gorm:"size:500" json:"address"`
	Contact     string     `gorm:"size:100" json:"contact"`
	Phone       string     `gorm:"size:20" json:"phone"`
	Email       string     `gorm:"size:100" json:"email"`
	Website     string     `gorm:"size:200" json:"website"`
	Description string     `gorm:"type:text" json:"description"`
	Status      string     `gorm:"size:20;default:active" json:"status"`
	ExpireAt    *time.Time `json:"expire_at"`

	// 关联
	Users   []User   `gorm:"foreignKey:SchoolID" json:"-"`
	Courses []Course `gorm:"foreignKey:SchoolID" json:"-"`
}

// TableName 表名
func (School) TableName() string {
	return "schools"
}

// IsActive 是否激活
func (s *School) IsActive() bool {
	if s.Status != StatusActive {
		return false
	}
	if s.ExpireAt != nil && s.ExpireAt.Before(time.Now()) {
		return false
	}
	return true
}
