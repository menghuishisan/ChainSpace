package response

import (
	"time"

	"github.com/chainspace/backend/internal/model"
)

// CourseResponse 课程响应
type CourseResponse struct {
	ID           uint       `json:"id"`
	SchoolID     uint       `json:"school_id"`
	TeacherID    uint       `json:"teacher_id"`
	TeacherName  string     `json:"teacher_name,omitempty"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Cover        string     `json:"cover"`
	Code         string     `json:"code"`
	InviteCode   string     `json:"invite_code,omitempty"`
	Category     string     `json:"category"`
	Tags         []string   `json:"tags"`
	Status       string     `json:"status"`
	IsPublic     bool       `json:"is_public"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	MaxStudents  int        `json:"max_students"`
	StudentCount int64      `json:"student_count,omitempty"`
	ChapterCount int64      `json:"chapter_count,omitempty"`
	Progress     float64    `json:"progress,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// FromCourse 从Course模型转换
func (r *CourseResponse) FromCourse(c *model.Course) *CourseResponse {
	r.ID = c.ID
	r.SchoolID = c.SchoolID
	r.TeacherID = c.TeacherID
	r.Title = c.Title
	r.Description = c.Description
	r.Cover = c.Cover
	r.Code = c.Code
	r.InviteCode = c.InviteCode
	r.Category = c.Category
	r.Tags = c.Tags
	r.Status = c.Status
	r.IsPublic = c.IsPublic
	r.StartDate = c.StartDate
	r.EndDate = c.EndDate
	r.MaxStudents = c.MaxStudents
	r.CreatedAt = c.CreatedAt

	if c.Teacher != nil {
		r.TeacherName = c.Teacher.DisplayName()
	}

	return r
}

// CourseDetailResponse 课程详情响应
type CourseDetailResponse struct {
	CourseResponse
	Chapters []ChapterResponse `json:"chapters,omitempty"`
}

// ChapterResponse 章节响应
type ChapterResponse struct {
	ID              uint               `json:"id"`
	CourseID        uint               `json:"course_id"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	SortOrder       int                `json:"sort_order"`
	Status          string             `json:"status"`
	MaterialCount   int64              `json:"material_count,omitempty"`
	ExperimentCount int64              `json:"experiment_count,omitempty"`
	Progress        float64            `json:"progress,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	Materials       []MaterialResponse `json:"materials,omitempty"`
}

// FromChapter 从Chapter模型转换
func (r *ChapterResponse) FromChapter(c *model.Chapter) *ChapterResponse {
	r.ID = c.ID
	r.CourseID = c.CourseID
	r.Title = c.Title
	r.Description = c.Description
	r.SortOrder = c.SortOrder
	r.Status = c.Status
	r.CreatedAt = c.CreatedAt
	return r
}

// MaterialResponse 资料响应
type MaterialResponse struct {
	ID        uint      `json:"id"`
	ChapterID uint      `json:"chapter_id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Content   string    `json:"content,omitempty"`
	URL       string    `json:"url,omitempty"`
	FileSize  int64     `json:"file_size,omitempty"`
	Duration  int       `json:"duration,omitempty"`
	SortOrder int       `json:"sort_order"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress,omitempty"`
	Completed bool      `json:"completed,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FromMaterial 从Material模型转换
func (r *MaterialResponse) FromMaterial(m *model.Material) *MaterialResponse {
	r.ID = m.ID
	r.ChapterID = m.ChapterID
	r.Title = m.Title
	r.Type = m.Type
	r.Content = m.Content
	r.URL = m.URL
	r.FileSize = m.FileSize
	r.Duration = m.Duration
	r.SortOrder = m.SortOrder
	r.Status = m.Status
	r.CreatedAt = m.CreatedAt
	return r
}

// CourseStudentResponse 课程学生响应
type CourseStudentResponse struct {
	ID         uint       `json:"id"`
	StudentID  uint       `json:"student_id"`
	RealName   string     `json:"real_name"`
	StudentNo  string     `json:"student_no"`
	ClassName  string     `json:"class_name,omitempty"`
	Progress   float64    `json:"progress"`
	LastAccess *time.Time `json:"last_access,omitempty"`
	JoinedAt   time.Time  `json:"joined_at"`
	Status     string     `json:"status"`
}

// MaterialProgressResponse 学习进度响应
type MaterialProgressResponse struct {
	MaterialID   uint       `json:"material_id"`
	Progress     float64    `json:"progress"`
	Duration     int        `json:"duration"`
	LastPosition int        `json:"last_position"`
	Completed    bool       `json:"completed"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// CourseProgressResponse 课程学习进度响应
type CourseProgressResponse struct {
	CourseID             uint `json:"course_id"`
	TotalMaterials       int  `json:"total_materials"`
	CompletedMaterials   int  `json:"completed_materials"`
	TotalExperiments     int  `json:"total_experiments"`
	CompletedExperiments int  `json:"completed_experiments"`
	ProgressPercent      int  `json:"progress_percent"`
}
