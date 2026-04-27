package model

import (
	"time"
)

// Contest 表示比赛主体。
type Contest struct {
	BaseModel
	SchoolID            *uint               `gorm:"index" json:"school_id"`
	CreatorID           uint                `gorm:"index;not null" json:"creator_id"`
	Title               string              `gorm:"size:200;not null" json:"title"`
	Description         string              `gorm:"type:text" json:"description"`
	Type                string              `gorm:"size:20;not null" json:"type"`
	Level               string              `gorm:"size:20;default:practice" json:"level"`
	Cover               string              `gorm:"size:500" json:"cover"`
	Rules               string              `gorm:"type:text" json:"rules"`
	BattleOrchestration BattleOrchestration `gorm:"type:jsonb" json:"battle_orchestration"`
	StartTime           time.Time           `gorm:"not null" json:"start_time"`
	EndTime             time.Time           `gorm:"not null" json:"end_time"`
	RegistrationStart   *time.Time          `json:"registration_start"`
	RegistrationEnd     *time.Time          `json:"registration_end"`
	Status              string              `gorm:"size:20;default:draft" json:"status"`
	IsPublic            bool                `gorm:"default:false" json:"is_public"`
	MaxParticipants     int                 `gorm:"default:0" json:"max_participants"`
	TeamMinSize         int                 `gorm:"default:1" json:"team_min_size"`
	TeamMaxSize         int                 `gorm:"default:1" json:"team_max_size"`
	DynamicScore        bool                `gorm:"default:true" json:"dynamic_score"`
	FirstBloodBonus     int                 `gorm:"default:0" json:"first_blood_bonus"`

	// 关联对象。
	School     *School            `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Creator    *User              `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Challenges []ContestChallenge `gorm:"foreignKey:ContestID" json:"challenges,omitempty"`
	Teams      []Team             `gorm:"foreignKey:ContestID" json:"teams,omitempty"`
}

// TableName 返回比赛表名。
func (Contest) TableName() string {
	return "contests"
}

// CurrentStatusAt 根据统一规则计算指定时间点的比赛状态。
// 这里把数据库原始状态与时间窗口合并，避免各处再次自行推导。
func (c *Contest) CurrentStatusAt(now time.Time) string {
	if c.Status == ContestStatusDraft {
		return ContestStatusDraft
	}
	if now.Before(c.StartTime) {
		return ContestStatusPublished
	}
	if !now.Before(c.EndTime) {
		return ContestStatusEnded
	}
	return ContestStatusOngoing
}

// CurrentStatus 返回当前时间下的统一比赛状态。
func (c *Contest) CurrentStatus() string {
	return c.CurrentStatusAt(time.Now())
}

// MatchesStatusAt 判断比赛在指定时间点是否匹配给定的对外状态语义。
func (c *Contest) MatchesStatusAt(status string, now time.Time) bool {
	if status == "" {
		return true
	}
	return c.CurrentStatusAt(now) == status
}

// IsOngoing 判断比赛当前是否处于进行中。
func (c *Contest) IsOngoing() bool {
	return c.CurrentStatus() == ContestStatusOngoing
}

// IsRegistrationOpen 判断当前是否仍处于允许报名的时间窗口。
func (c *Contest) IsRegistrationOpen() bool {
	now := time.Now()
	if c.RegistrationStart != nil && now.Before(*c.RegistrationStart) {
		return false
	}
	if c.RegistrationEnd != nil && now.After(*c.RegistrationEnd) {
		return false
	}

	return c.CurrentStatusAt(now) == ContestStatusPublished
}

// Challenge 表示比赛题目。
type Challenge struct {
	BaseModel
	CreatorID              uint                   `gorm:"index;not null" json:"creator_id"`
	SchoolID               *uint                  `gorm:"index" json:"school_id"`
	Title                  string                 `gorm:"size:200;not null" json:"title"`
	Description            string                 `gorm:"type:text" json:"description"`
	Category               string                 `gorm:"size:50;not null" json:"category"`
	RuntimeProfile         string                 `gorm:"size:40;default:single_chain_instance" json:"runtime_profile"`
	Difficulty             int                    `gorm:"default:1" json:"difficulty"`
	BasePoints             int                    `gorm:"default:100" json:"base_points"`
	MinPoints              int                    `gorm:"default:50" json:"min_points"`
	DecayFactor            float64                `gorm:"default:0.1" json:"decay_factor"`
	ContractCode           string                 `gorm:"type:text" json:"contract_code"`
	SetupCode              string                 `gorm:"type:text" json:"setup_code"`
	DeployScript           string                 `gorm:"type:text" json:"deploy_script"`
	CheckScript            string                 `gorm:"type:text" json:"check_script"`
	FlagTemplate           string                 `gorm:"size:200" json:"flag_template"`
	FlagType               string                 `gorm:"size:20;default:static" json:"flag_type"`
	FlagRegex              string                 `gorm:"size:500" json:"flag_regex"`
	FlagSecret             string                 `gorm:"size:200" json:"flag_secret"`
	ValidationConfig       JSONMap                `gorm:"type:jsonb" json:"validation_config"`
	Hints                  JSONArray              `gorm:"type:jsonb" json:"hints"`
	Attachments            JSONArray              `gorm:"type:jsonb" json:"attachments"`
	ChallengeOrchestration ChallengeOrchestration `gorm:"type:jsonb" json:"challenge_orchestration"`
	Tags                   StringList             `gorm:"type:jsonb" json:"tags"`
	SourceType             string                 `gorm:"size:20;default:user_created" json:"source_type"`
	SourceRef              string                 `gorm:"size:255" json:"source_ref"`
	Status                 string                 `gorm:"size:20;default:draft" json:"status"`
	IsPublic               bool                   `gorm:"default:false" json:"is_public"`
	SolveCount             int                    `gorm:"default:0" json:"solve_count"`
	AttemptCount           int                    `gorm:"default:0" json:"attempt_count"`

	// 关联对象。
	Creator *User   `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	School  *School `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
}

// TableName 返回题目表名。
func (Challenge) TableName() string {
	return "challenges"
}

// ContestChallenge 表示比赛与题目之间的关联。
type ContestChallenge struct {
	BaseModel
	ContestID     uint `gorm:"uniqueIndex:idx_contest_challenge;not null" json:"contest_id"`
	ChallengeID   uint `gorm:"uniqueIndex:idx_contest_challenge;not null" json:"challenge_id"`
	Points        int  `gorm:"default:100" json:"points"`
	CurrentPoints int  `gorm:"default:100" json:"current_points"`
	SortOrder     int  `gorm:"default:0" json:"sort_order"`
	IsVisible     bool `gorm:"default:true" json:"is_visible"`

	// 关联对象。
	Contest   *Contest   `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
}

// TableName 返回比赛题目关联表名。
func (ContestChallenge) TableName() string {
	return "contest_challenges"
}

// Team 表示比赛队伍。
type Team struct {
	BaseModel
	ContestID        uint   `gorm:"index;not null" json:"contest_id"`
	Name             string `gorm:"size:100;not null" json:"name"`
	Token            string `gorm:"size:50;uniqueIndex" json:"token"`
	LeaderID         uint   `gorm:"index;not null" json:"leader_id"`
	SchoolID         *uint  `gorm:"index" json:"school_id"`
	Avatar           string `gorm:"size:500" json:"avatar"`
	Description      string `gorm:"type:text" json:"description"`
	Status           string `gorm:"size:20;default:active" json:"status"`
	WorkspacePodName string `gorm:"size:100" json:"workspace_pod_name"`
	WorkspaceStatus  string `gorm:"size:20;default:inactive" json:"workspace_status"`

	// 关联对象。
	Contest *Contest     `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Leader  *User        `gorm:"foreignKey:LeaderID" json:"leader,omitempty"`
	Members []TeamMember `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

// TableName 返回队伍表名。
func (Team) TableName() string {
	return "teams"
}

// TeamMember 表示队伍成员。
type TeamMember struct {
	BaseModel
	TeamID   uint      `gorm:"uniqueIndex:idx_team_user;not null" json:"team_id"`
	UserID   uint      `gorm:"uniqueIndex:idx_team_user;not null" json:"user_id"`
	Role     string    `gorm:"size:20;default:member" json:"role"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`

	// 关联对象。
	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 返回队伍成员表名。
func (TeamMember) TableName() string {
	return "team_members"
}

// ContestRegistration 表示比赛报名记录。
type ContestRegistration struct {
	BaseModel
	ContestID    uint      `gorm:"uniqueIndex:idx_contest_user;not null" json:"contest_id"`
	UserID       uint      `gorm:"uniqueIndex:idx_contest_user;not null" json:"user_id"`
	TeamID       *uint     `gorm:"index" json:"team_id"`
	Status       string    `gorm:"size:20;default:pending" json:"status"`
	RegisteredAt time.Time `gorm:"autoCreateTime" json:"registered_at"`

	// 关联对象。
	Contest *Contest `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Team    *Team    `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName 返回报名表名。
func (ContestRegistration) TableName() string {
	return "contest_registrations"
}

// ChallengeEnv 表示题目环境。
type ChallengeEnv struct {
	BaseModel
	EnvID        string     `gorm:"size:50;uniqueIndex;not null" json:"env_id"`
	ContestID    uint       `gorm:"index;not null" json:"contest_id"`
	ChallengeID  uint       `gorm:"index;not null" json:"challenge_id"`
	UserID       uint       `gorm:"index;not null" json:"user_id"`
	TeamID       *uint      `gorm:"index" json:"team_id"`
	Status       string     `gorm:"size:20;default:pending" json:"status"`
	PodName      string     `gorm:"size:100" json:"pod_name"`
	ForkPodName  string     `gorm:"size:100" json:"fork_pod_name"`
	AccessURL    string     `gorm:"size:500" json:"access_url"`
	Flag         string     `gorm:"size:200" json:"flag"`
	StartedAt    *time.Time `json:"started_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`

	// 关联对象。
	Contest   *Contest   `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Team      *Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName 返回题目环境表名。
func (ChallengeEnv) TableName() string {
	return "challenge_envs"
}

// 题目状态常量。
const (
	ChallengeStatusDraft  = "draft"
	ChallengeStatusActive = "active"
)

// 队伍成员角色常量。
const (
	TeamRoleLeader = "leader"
	TeamRoleMember = "member"
)

// 队伍工作区状态常量。
const (
	TeamWorkspaceStatusInactive = "inactive"
	TeamWorkspaceStatusRunning  = "running"
	TeamWorkspaceStatusStopped  = "stopped"
)

// 题目环境状态常量。
const (
	ChallengeEnvStatusPending    = "pending"
	ChallengeEnvStatusCreating   = "creating"
	ChallengeEnvStatusRunning    = "running"
	ChallengeEnvStatusFailed     = "failed"
	ChallengeEnvStatusExpired    = "expired"
	ChallengeEnvStatusTerminated = "terminated"
)

// ActiveChallengeEnvStatuses 返回仍可视为活跃中的题目环境状态集合。
func ActiveChallengeEnvStatuses() []string {
	return []string{
		ChallengeEnvStatusPending,
		ChallengeEnvStatusCreating,
		ChallengeEnvStatusRunning,
		ChallengeEnvStatusFailed,
	}
}

// ContestSubmission 表示比赛提交记录。
type ContestSubmission struct {
	BaseModel
	ContestID   uint      `gorm:"index;not null" json:"contest_id"`
	ChallengeID uint      `gorm:"index;not null" json:"challenge_id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	TeamID      *uint     `gorm:"index" json:"team_id"`
	Flag        string    `gorm:"size:200" json:"flag"`
	IsCorrect   bool      `gorm:"default:false" json:"is_correct"`
	Points      int       `gorm:"default:0" json:"points"`
	SubmittedAt time.Time `gorm:"autoCreateTime" json:"submitted_at"`
	IP          string    `gorm:"size:50" json:"ip"`

	// 关联对象。
	Contest   *Contest   `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Team      *Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName 返回比赛提交表名。
func (ContestSubmission) TableName() string {
	return "contest_submissions"
}

// ContestScore 表示比赛积分。
type ContestScore struct {
	BaseModel
	ContestID       uint       `gorm:"uniqueIndex:idx_contest_team_score;not null" json:"contest_id"`
	TeamID          *uint      `gorm:"uniqueIndex:idx_contest_team_score" json:"team_id"`
	UserID          uint       `gorm:"index;not null" json:"user_id"`
	TotalScore      int        `gorm:"default:0" json:"total_score"`
	SolveCount      int        `gorm:"default:0" json:"solve_count"`
	LastSolveAt     *time.Time `json:"last_solve_at"`
	Rank            int        `gorm:"default:0" json:"rank"`
	FirstBloodCount int        `gorm:"default:0" json:"first_blood_count"`

	// 关联对象。
	Contest *Contest `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Team    *Team    `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 返回比赛积分表名。
func (ContestScore) TableName() string {
	return "contest_scores"
}

// VulnerabilitySource 表示漏洞数据源配置。
type VulnerabilitySource struct {
	BaseModel
	Name          string     `gorm:"size:200;not null" json:"name"`
	Type          string     `gorm:"size:50;not null" json:"type"`
	Config        JSONMap    `gorm:"type:jsonb" json:"config"`
	Description   string     `gorm:"type:text" json:"description"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	LastSyncAt    *time.Time `json:"last_sync_at"`
	LastSyncError string     `gorm:"type:text" json:"last_sync_error"`
}

// TableName 返回漏洞数据源表名。
func (VulnerabilitySource) TableName() string {
	return "vulnerability_sources"
}

// Vulnerability 表示漏洞数据。
type Vulnerability struct {
	BaseModel
	SourceID                 uint       `gorm:"index;not null" json:"source_id"`
	ExternalID               string     `gorm:"size:100;index" json:"external_id"`
	Title                    string     `gorm:"size:500;not null" json:"title"`
	Description              string     `gorm:"type:text" json:"description"`
	Severity                 string     `gorm:"size:20" json:"severity"`
	Category                 string     `gorm:"size:50" json:"category"`
	Technique                string     `gorm:"size:200" json:"technique"`
	Chain                    string     `gorm:"size:200" json:"chain"`
	Amount                   float64    `json:"amount"`
	AttackDate               time.Time  `json:"attack_date"`
	BlockNumber              uint64     `gorm:"default:0" json:"block_number"`
	ContractAddress          string     `gorm:"size:100" json:"contract_address"`
	AttackTxHash             string     `gorm:"size:100" json:"attack_tx_hash"`
	ForkBlockNumber          uint64     `gorm:"default:0" json:"fork_block_number"`
	RelatedContracts         StringList `gorm:"type:jsonb" json:"related_contracts"`
	RelatedTokens            StringList `gorm:"type:jsonb" json:"related_tokens"`
	AttackerAddresses        StringList `gorm:"type:jsonb" json:"attacker_addresses"`
	VictimAddresses          StringList `gorm:"type:jsonb" json:"victim_addresses"`
	EvidenceLinks            StringList `gorm:"type:jsonb" json:"evidence_links"`
	Reference                string     `gorm:"size:500" json:"reference"`
	VulnCode                 string     `gorm:"type:text" json:"vuln_code"`
	AttackCode               string     `gorm:"type:text" json:"attack_code"`
	Analysis                 string     `gorm:"type:text" json:"analysis"`
	Tags                     StringList `gorm:"type:jsonb" json:"tags"`
	SourceSnapshot           JSONMap    `gorm:"type:jsonb" json:"source_snapshot"`
	Metadata                 JSONMap    `gorm:"type:jsonb" json:"metadata"`
	EnrichStatus             string     `gorm:"size:20;default:pending" json:"enrich_status"`
	ConversionScore          int        `gorm:"default:0" json:"conversion_score"`
	ConversionNotes          string     `gorm:"type:text" json:"conversion_notes"`
	RuntimeProfileSuggestion string     `gorm:"size:40;default:fork_replay" json:"runtime_profile_suggestion"`
	Status                   string     `gorm:"size:20;default:discovered" json:"status"`
	ConvertedID              *uint      `json:"converted_id"`

	// 关联对象。
	Source    *VulnerabilitySource `gorm:"foreignKey:SourceID" json:"source,omitempty"`
	Converted *Challenge           `gorm:"foreignKey:ConvertedID" json:"converted,omitempty"`
}

// TableName 返回漏洞表名。
func (Vulnerability) TableName() string {
	return "vulnerabilities"
}

// 漏洞状态常量。
const (
	VulnStatusDiscovered = "discovered"
	VulnStatusEnriched   = "enriched"
	VulnStatusConverted  = "converted"
	VulnStatusSkipped    = "skipped"
	VulnStatusFailed     = "failed"
)

// 漏洞增强状态常量。
const (
	VulnEnrichStatusPending = "pending"
	VulnEnrichStatusRunning = "running"
	VulnEnrichStatusDone    = "done"
	VulnEnrichStatusFailed  = "failed"
)

// CrossSchoolApplication 表示跨校申请。
type CrossSchoolApplication struct {
	BaseModel
	FromSchoolID uint       `gorm:"index;not null" json:"from_school_id"`
	ToSchoolID   uint       `gorm:"index;not null" json:"to_school_id"`
	ApplicantID  uint       `gorm:"index;not null" json:"applicant_id"`
	Type         string     `gorm:"size:30;not null" json:"type"`
	TargetID     uint       `gorm:"not null" json:"target_id"`
	TargetType   string     `gorm:"size:30;not null" json:"target_type"`
	Reason       string     `gorm:"type:text" json:"reason"`
	Status       string     `gorm:"size:20;default:pending" json:"status"`
	ReviewerID   *uint      `json:"reviewer_id"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	RejectReason string     `gorm:"type:text" json:"reject_reason"`

	// 关联对象。
	FromSchool *School `gorm:"foreignKey:FromSchoolID" json:"from_school,omitempty"`
	ToSchool   *School `gorm:"foreignKey:ToSchoolID" json:"to_school,omitempty"`
	Applicant  *User   `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
	Reviewer   *User   `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
}

// TableName 返回跨校申请表名。
func (CrossSchoolApplication) TableName() string {
	return "cross_school_applications"
}

// 跨校申请类型常量。
const (
	CrossSchoolTypeContest   = "contest"
	CrossSchoolTypeChallenge = "challenge"
)

// 跨校申请状态常量。
const (
	CrossSchoolStatusPending  = "pending"
	CrossSchoolStatusApproved = "approved"
	CrossSchoolStatusRejected = "rejected"
)

// 题目公开申请状态常量。
const (
	ChallengePublishStatusPending  = "pending"
	ChallengePublishStatusApproved = "approved"
	ChallengePublishStatusRejected = "rejected"
)

// ChallengePublishRequest 表示题目公开申请。
type ChallengePublishRequest struct {
	BaseModel
	ChallengeID   uint       `gorm:"index;not null" json:"challenge_id"`
	ApplicantID   uint       `gorm:"index;not null" json:"applicant_id"`
	Reason        string     `gorm:"type:text" json:"reason"`
	Status        string     `gorm:"size:20;default:pending" json:"status"`
	ReviewerID    *uint      `json:"reviewer_id"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	ReviewComment string     `gorm:"type:text" json:"review_comment"`

	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
	Applicant *User      `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
}

// TableName 返回题目公开申请表名。
func (ChallengePublishRequest) TableName() string {
	return "challenge_publish_requests"
}
