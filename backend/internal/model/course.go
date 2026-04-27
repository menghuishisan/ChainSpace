package model

import (
	"time"
)

// Course 课程
type Course struct {
	BaseModel
	SchoolID    uint       `gorm:"index;not null" json:"school_id"`
	TeacherID   uint       `gorm:"index;not null" json:"teacher_id"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Cover       string     `gorm:"size:500" json:"cover"`
	Code        string     `gorm:"size:20;index" json:"code"`
	InviteCode  string     `gorm:"size:20;uniqueIndex" json:"invite_code"`
	Category    string     `gorm:"size:50" json:"category"`
	Tags        StringList `gorm:"type:jsonb" json:"tags"`
	Status      string     `gorm:"size:20;default:draft" json:"status"`
	IsPublic    bool       `gorm:"default:false" json:"is_public"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	MaxStudents int        `gorm:"default:0" json:"max_students"`

	// 关联
	School   *School         `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Teacher  *User           `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Chapters []Chapter       `gorm:"foreignKey:CourseID" json:"chapters,omitempty"`
	Students []CourseStudent `gorm:"foreignKey:CourseID" json:"students,omitempty"`
}

// TableName 表名
func (Course) TableName() string {
	return "courses"
}

// IsPublished 是否已发布
func (c *Course) IsPublished() bool {
	return c.Status == "published"
}

// CourseStudent 课程学生关联
type CourseStudent struct {
	BaseModel
	CourseID   uint       `gorm:"uniqueIndex:idx_course_student;not null" json:"course_id"`
	StudentID  uint       `gorm:"uniqueIndex:idx_course_student;not null" json:"student_id"`
	JoinedAt   time.Time  `gorm:"autoCreateTime" json:"joined_at"`
	Progress   float64    `gorm:"default:0" json:"progress"`
	LastAccess *time.Time `json:"last_access"`
	Status     string     `gorm:"size:20;default:active" json:"status"`

	// 关联
	Course  *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Student *User   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}

// TableName 表名
func (CourseStudent) TableName() string {
	return "course_students"
}

// Chapter 章节
type Chapter struct {
	BaseModel
	CourseID    uint   `gorm:"index;not null" json:"course_id"`
	Title       string `gorm:"size:200;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	Status      string `gorm:"size:20;default:draft" json:"status"`

	// 关联
	Course      *Course      `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Materials   []Material   `gorm:"foreignKey:ChapterID" json:"materials,omitempty"`
	Experiments []Experiment `gorm:"foreignKey:ChapterID" json:"experiments,omitempty"`
}

// TableName 表名
func (Chapter) TableName() string {
	return "chapters"
}

// Material 学习资料
type Material struct {
	BaseModel
	ChapterID uint   `gorm:"index;not null" json:"chapter_id"`
	Title     string `gorm:"size:200;not null" json:"title"`
	Type      string `gorm:"size:20;not null" json:"type"`
	Content   string `gorm:"type:text" json:"content"`
	URL       string `gorm:"size:500" json:"url"`
	FileSize  int64  `gorm:"default:0" json:"file_size"`
	Duration  int    `gorm:"default:0" json:"duration"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	Status    string `gorm:"size:20;default:active" json:"status"`

	// 关联
	Chapter *Chapter `gorm:"foreignKey:ChapterID" json:"chapter,omitempty"`
}

// TableName 表名
func (Material) TableName() string {
	return "materials"
}

// MaterialProgress 学习进度
type MaterialProgress struct {
	BaseModel
	MaterialID   uint       `gorm:"uniqueIndex:idx_material_student;not null" json:"material_id"`
	StudentID    uint       `gorm:"uniqueIndex:idx_material_student;not null" json:"student_id"`
	Progress     float64    `gorm:"default:0" json:"progress"`
	Duration     int        `gorm:"default:0" json:"duration"`
	LastPosition int        `gorm:"default:0" json:"last_position"`
	Completed    bool       `gorm:"default:false" json:"completed"`
	CompletedAt  *time.Time `json:"completed_at"`

	// 关联
	Material *Material `gorm:"foreignKey:MaterialID" json:"material,omitempty"`
	Student  *User     `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}

// TableName 表名
func (MaterialProgress) TableName() string {
	return "material_progress"
}
