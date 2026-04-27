package model

import (
	"time"
)

// User 用户
type User struct {
	BaseModel
	SchoolID      *uint      `gorm:"index" json:"school_id"`
	Password      string     `gorm:"size:100;not null" json:"-"`
	Email         string     `gorm:"size:100;index" json:"email"`
	Phone         string     `gorm:"size:20" json:"phone"`
	RealName      string     `gorm:"size:50" json:"real_name"`
	Avatar        string     `gorm:"size:500" json:"avatar"`
	Role          string     `gorm:"size:20;index;not null" json:"role"`
	StudentNo     string     `gorm:"size:50;index" json:"student_no"`
	ClassID       *uint      `gorm:"index" json:"class_id"`
	Status        string     `gorm:"size:20;default:active" json:"status"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	LastLoginIP   string     `gorm:"size:50" json:"last_login_ip"`
	MustChangePwd bool       `gorm:"default:false" json:"must_change_pwd"`

	// 关联
	School *School `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Class  *Class  `gorm:"foreignKey:ClassID" json:"class,omitempty"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// IsActive 是否激活
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsPlatformAdmin 是否为平台管理员
func (u *User) IsPlatformAdmin() bool {
	return u.Role == RolePlatformAdmin
}

// IsSchoolAdmin 是否为学校管理员
func (u *User) IsSchoolAdmin() bool {
	return u.Role == RoleSchoolAdmin
}

// IsTeacher 是否为教师
func (u *User) IsTeacher() bool {
	return u.Role == RoleTeacher
}

// IsStudent 是否为学生
func (u *User) IsStudent() bool {
	return u.Role == RoleStudent
}

// DisplayName 返回适合界面展示的人类可读名称。
func (u *User) DisplayName() string {
	if u == nil {
		return ""
	}
	if u.RealName != "" {
		return u.RealName
	}
	if u.StudentNo != "" {
		return u.StudentNo
	}
	return u.Phone
}

// Class 班级
type Class struct {
	BaseModel
	SchoolID    uint   `gorm:"index;not null" json:"school_id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Grade       string `gorm:"size:20" json:"grade"`
	Major       string `gorm:"size:100" json:"major"`
	Description string `gorm:"type:text" json:"description"`
	Status      string `gorm:"size:20;default:active" json:"status"`

	// 关联
	School   *School `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Students []User  `gorm:"foreignKey:ClassID" json:"students,omitempty"`
}

// TableName 表名
func (Class) TableName() string {
	return "classes"
}
