package model

import (
	"time"

	"gorm.io/gorm"
)

type Experiment struct {
	BaseModel
	SchoolID      uint       `gorm:"index;not null" json:"school_id"`
	ChapterID     uint       `gorm:"index;not null" json:"chapter_id"`
	CreatorID     uint       `gorm:"index;not null" json:"creator_id"`
	Title         string     `gorm:"size:200;not null" json:"title"`
	Description   string     `gorm:"type:text" json:"description"`
	Type          string     `gorm:"size:30;not null" json:"type"`
	Mode          string     `gorm:"size:20;not null;default:single" json:"mode"`
	Difficulty    int        `gorm:"default:1" json:"difficulty"`
	MaxScore      int        `gorm:"default:100" json:"max_score"`
	PassScore     int        `gorm:"default:60" json:"pass_score"`
	AutoGrade     bool       `gorm:"default:false" json:"auto_grade"`
	GradingStrategy string   `gorm:"size:30;default:checkpoint" json:"grading_strategy"`
	EstimatedTime int        `gorm:"default:60" json:"estimated_time"`
	SortOrder     int        `gorm:"default:0" json:"sort_order"`
	Status        string     `gorm:"size:20;default:draft" json:"status"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	AllowLate     bool       `gorm:"default:false" json:"allow_late"`
	LateDeduction int        `gorm:"default:0" json:"late_deduction"`

	School      *School                `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Chapter     *Chapter               `gorm:"foreignKey:ChapterID" json:"chapter,omitempty"`
	Creator     *User                  `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Workspace   *ExperimentWorkspace   `gorm:"foreignKey:ExperimentID" json:"workspace,omitempty"`
	Topology    *ExperimentTopology    `gorm:"foreignKey:ExperimentID" json:"topology,omitempty"`
	Tools       []ExperimentTool       `gorm:"foreignKey:ExperimentID" json:"tools,omitempty"`
	InitScripts []ExperimentInitScript `gorm:"foreignKey:ExperimentID" json:"init_scripts,omitempty"`
	Collaboration *ExperimentCollaboration `gorm:"foreignKey:ExperimentID" json:"collaboration,omitempty"`
	Nodes       []ExperimentNode       `gorm:"foreignKey:ExperimentID" json:"nodes,omitempty"`
	Services    []ExperimentService    `gorm:"foreignKey:ExperimentID" json:"services,omitempty"`
	Assets      []ExperimentAsset      `gorm:"foreignKey:ExperimentID" json:"assets,omitempty"`
	Checkpoints []ExperimentCheckpoint `gorm:"foreignKey:ExperimentID" json:"checkpoints,omitempty"`
}

func (Experiment) TableName() string {
	return "experiments"
}

type ExperimentEnv struct {
	BaseModel
	EnvID              string                    `gorm:"size:50;uniqueIndex;not null" json:"env_id"`
	ExperimentID       uint                      `gorm:"index;not null" json:"experiment_id"`
	SessionID          *uint                     `gorm:"index" json:"session_id"`
	UserID             uint                      `gorm:"index;not null" json:"user_id"`
	SchoolID           uint                      `gorm:"index;not null" json:"school_id"`
	Status             string                    `gorm:"size:20;default:pending" json:"status"`
	SessionMode        string                    `gorm:"size:20;not null;default:single" json:"session_mode"`
	PrimaryInstanceKey string                    `gorm:"size:50" json:"primary_instance_key"`
	StartedAt          *time.Time                `json:"started_at"`
	ExpiresAt          *time.Time                `json:"expires_at"`
	ExtendCount        int                       `gorm:"default:0" json:"extend_count"`
	SnapshotAt         *time.Time                `json:"snapshot_at"`
	SnapshotURL        string                    `gorm:"size:500" json:"snapshot_url"`
	ErrorMessage       string                    `gorm:"type:text" json:"error_message"`

	Experiment       *Experiment                 `gorm:"foreignKey:ExperimentID" json:"experiment,omitempty"`
	Session          *ExperimentSession          `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	User             *User                       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RuntimeInstances []ExperimentRuntimeInstance `gorm:"foreignKey:ExperimentEnvID" json:"runtime_instances,omitempty"`
}

func (ExperimentEnv) TableName() string {
	return "experiment_envs"
}

func ActiveExperimentEnvStatuses() []string {
	return []string{
		EnvStatusPending,
		EnvStatusCreating,
		EnvStatusRunning,
		EnvStatusPaused,
	}
}

func (e *ExperimentEnv) IsRunning() bool {
	return e.Status == EnvStatusRunning
}

func (e *ExperimentEnv) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}
	return e.ExpiresAt.Before(time.Now())
}

type ExperimentSession struct {
	BaseModel
	SessionKey          string                    `gorm:"size:64;uniqueIndex;not null" json:"session_key"`
	ExperimentID        uint                      `gorm:"index;not null" json:"experiment_id"`
	SchoolID            uint                      `gorm:"index;not null" json:"school_id"`
	Mode                string                    `gorm:"size:20;not null" json:"mode"`
	Status              string                    `gorm:"size:20;default:pending" json:"status"`
	PrimaryEnvID        string                    `gorm:"size:50" json:"primary_env_id"`
	MaxMembers          int                       `gorm:"default:1" json:"max_members"`
	CurrentMemberCount  int                       `gorm:"default:0" json:"current_member_count"`
	StartedAt           *time.Time                `json:"started_at"`
	ExpiresAt           *time.Time                `json:"expires_at"`

	Experiment *Experiment                `gorm:"foreignKey:ExperimentID" json:"experiment,omitempty"`
	Members    []ExperimentSessionMember  `gorm:"foreignKey:SessionID" json:"members,omitempty"`
	Messages   []ExperimentSessionMessage `gorm:"foreignKey:SessionID" json:"messages,omitempty"`
}

func (ExperimentSession) TableName() string {
	return "experiment_sessions"
}

type ExperimentSessionMember struct {
	BaseModel
	SessionID          uint       `gorm:"index;not null" json:"session_id"`
	UserID             uint       `gorm:"index;not null" json:"user_id"`
	RoleKey            string     `gorm:"size:50" json:"role_key"`
	AssignedNodeKey    string     `gorm:"size:50" json:"assigned_node_key"`
	JoinStatus         string     `gorm:"size:20;default:joined" json:"join_status"`
	JoinedAt           time.Time  `gorm:"autoCreateTime" json:"joined_at"`

	Session *ExperimentSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	User    *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ExperimentSessionMember) TableName() string {
	return "experiment_session_members"
}

type ExperimentSessionMessage struct {
	BaseModel
	SessionID  uint   `gorm:"index;not null" json:"session_id"`
	UserID     uint   `gorm:"index;not null" json:"user_id"`
	Message    string `gorm:"type:text;not null" json:"message"`
	MessageType string `gorm:"size:20;default:text" json:"message_type"`

	Session *ExperimentSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	User    *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ExperimentSessionMessage) TableName() string {
	return "experiment_session_messages"
}

type Submission struct {
	BaseModel
	ExperimentID  uint       `gorm:"index;not null" json:"experiment_id"`
	StudentID     uint       `gorm:"index;not null" json:"student_id"`
	SchoolID      uint       `gorm:"index;not null" json:"school_id"`
	EnvID         string     `gorm:"size:50" json:"env_id"`
	Content       string     `gorm:"type:text" json:"content"`
	FileURL       string     `gorm:"size:500" json:"file_url"`
	SnapshotURL   string     `gorm:"size:500" json:"snapshot_url"`
	Score         *int       `json:"score"`
	AutoScore     *int       `json:"auto_score"`
	ManualScore   *int       `json:"manual_score"`
	Feedback      string     `gorm:"type:text" json:"feedback"`
	Status        string     `gorm:"size:20;default:pending" json:"status"`
	SubmittedAt   time.Time  `gorm:"autoCreateTime" json:"submitted_at"`
	GradedAt      *time.Time `json:"graded_at"`
	GraderID      *uint      `json:"grader_id"`
	IsLate        bool       `gorm:"default:false" json:"is_late"`
	AttemptNumber int        `gorm:"default:1" json:"attempt_number"`

	Experiment   *Experiment             `gorm:"foreignKey:ExperimentID" json:"experiment,omitempty"`
	Student      *User                   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Grader       *User                   `gorm:"foreignKey:GraderID" json:"grader,omitempty"`
	CheckResults []SubmissionCheckResult `gorm:"foreignKey:SubmissionID" json:"check_results,omitempty"`
}

func (Submission) TableName() string {
	return "submissions"
}

type DockerImage struct {
	BaseModel
	Name             string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Tag              string    `gorm:"size:50;not null" json:"tag"`
	Registry         string    `gorm:"size:200" json:"registry"`
	Description      string    `gorm:"type:text" json:"description"`
	Category         string    `gorm:"size:50" json:"category"`
	DefaultResources JSONMap   `gorm:"type:jsonb" json:"default_resources"`
	Features         JSONArray `gorm:"type:jsonb" json:"features"`
	EnvVars          JSONMap   `gorm:"type:jsonb" json:"env_vars"`
	Ports            JSONArray `gorm:"type:jsonb" json:"ports"`
	BaseImage        string    `gorm:"size:100" json:"base_image"`
	Size             int64     `gorm:"default:0" json:"size"`
	Status           string    `gorm:"size:20;default:active" json:"-"`
	IsActive         bool      `gorm:"-" json:"is_active"`
	IsBuiltIn        bool      `gorm:"default:false" json:"is_built_in"`
}

func (DockerImage) TableName() string {
	return "docker_images"
}

func (d *DockerImage) AfterFind(tx *gorm.DB) error {
	d.IsActive = d.Status == StatusActive
	return nil
}

func (d *DockerImage) FullName() string {
	if d.Registry != "" {
		return d.Registry + "/" + d.Name + ":" + d.Tag
	}
	return d.Name + ":" + d.Tag
}
