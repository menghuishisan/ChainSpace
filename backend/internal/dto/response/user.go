package response

import (
	"time"

	"github.com/chainspace/backend/internal/model"
)

// UserResponse 用户响应
type UserResponse struct {
	ID            uint       `json:"id"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	RealName      string     `json:"real_name"`
	Avatar        string     `json:"avatar"`
	Role          string     `json:"role"`
	StudentNo     string     `json:"student_no,omitempty"`
	ClassID       *uint      `json:"class_id,omitempty"`
	ClassName     string     `json:"class_name,omitempty"`
	SchoolID      *uint      `json:"school_id,omitempty"`
	SchoolName    string     `json:"school_name,omitempty"`
	Status        string     `json:"status"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	MustChangePwd bool       `json:"must_change_pwd"`
	CreatedAt     time.Time  `json:"created_at"`
}

// FromUser 从User模型转换
func (r *UserResponse) FromUser(u *model.User) *UserResponse {
	r.ID = u.ID
	r.Email = u.Email
	r.Phone = u.Phone
	r.RealName = u.RealName
	r.Avatar = u.Avatar
	r.Role = u.Role
	r.StudentNo = u.StudentNo
	r.ClassID = u.ClassID
	r.SchoolID = u.SchoolID
	r.Status = u.Status
	r.LastLoginAt = u.LastLoginAt
	r.MustChangePwd = u.MustChangePwd
	r.CreatedAt = u.CreatedAt

	if u.School != nil {
		r.SchoolName = u.School.Name
	}
	if u.Class != nil {
		r.ClassName = u.Class.Name
	}

	return r
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	ID          uint       `json:"id"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	RealName    string     `json:"real_name"`
	Avatar      string     `json:"avatar"`
	Role        string     `json:"role"`
	StudentNo   string     `json:"student_no,omitempty"`
	ClassName   string     `json:"class_name,omitempty"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// SchoolResponse 学校响应
type SchoolResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	Logo         string     `json:"logo"`
	Address      string     `json:"address"`
	Contact      string     `json:"contact"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email"`
	Website      string     `json:"website"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	ExpireAt     *time.Time `json:"expire_at,omitempty"`
	TeacherCount int64      `json:"teacher_count"`
	StudentCount int64      `json:"student_count"`
	CreatedAt    time.Time  `json:"created_at"`
}

// FromSchool 从School模型转换
func (r *SchoolResponse) FromSchool(s *model.School) *SchoolResponse {
	r.ID = s.ID
	r.Name = s.Name
	r.Code = s.Code
	r.Logo = s.Logo
	r.Address = s.Address
	r.Contact = s.Contact
	r.Phone = s.Phone
	r.Email = s.Email
	r.Website = s.Website
	r.Description = s.Description
	r.Status = s.Status
	r.ExpireAt = s.ExpireAt
	r.CreatedAt = s.CreatedAt
	return r
}

// ClassResponse 班级响应
type ClassResponse struct {
	ID           uint      `json:"id"`
	SchoolID     uint      `json:"school_id"`
	Name         string    `json:"name"`
	Grade        string    `json:"grade"`
	Major        string    `json:"major"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	StudentCount int64     `json:"student_count,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// FromClass 从Class模型转换
func (r *ClassResponse) FromClass(c *model.Class) *ClassResponse {
	r.ID = c.ID
	r.SchoolID = c.SchoolID
	r.Name = c.Name
	r.Grade = c.Grade
	r.Major = c.Major
	r.Description = c.Description
	r.Status = c.Status
	r.CreatedAt = c.CreatedAt
	return r
}

// BatchImportResult 批量导入结果
type BatchImportResult struct {
	Total   int                `json:"total"`
	Success int                `json:"success"`
	Failed  int                `json:"failed"`
	Errors  []BatchImportError `json:"errors,omitempty"`
}

// BatchImportError 批量导入错误
type BatchImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}
