package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TenantModel 多租户模型（带school_id）
type TenantModel struct {
	BaseModel
	SchoolID uint `gorm:"index;not null" json:"school_id"`
}

// 状态常量
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusPending  = "pending"
	StatusDisabled = "disabled"
)

// 角色常量
const (
	RolePlatformAdmin = "platform_admin"
	RoleSchoolAdmin   = "school_admin"
	RoleTeacher       = "teacher"
	RoleStudent       = "student"
)

// 实验类型常量
const (
	ExperimentTypeVisualization = "visualization"
	ExperimentTypeCodeDev       = "code_dev"
	ExperimentTypeCommandOp     = "command_op"
	ExperimentTypeDataAnalysis  = "data_analysis"
	ExperimentTypeToolUsage     = "tool_usage"
	ExperimentTypeConfigDebug   = "config_debug"
	ExperimentTypeReverse       = "reverse"
	ExperimentTypeTroubleshoot  = "troubleshoot"
	ExperimentTypeCollaboration = "collaboration"
)

// 实验模式常量
const (
	ExperimentModeSingle        = "single"
	ExperimentModeMultiNode     = "multi_node"
	ExperimentModeCollaboration = "collaboration"
)

// 实验环境状态常量
const (
	EnvStatusPending    = "pending"
	EnvStatusCreating   = "creating"
	EnvStatusRunning    = "running"
	EnvStatusPaused     = "paused"
	EnvStatusTerminated = "terminated"
	EnvStatusFailed     = "failed"
)

// 实验状态常量
const (
	ExperimentStatusDraft     = "draft"
	ExperimentStatusPublished = "published"
)

// 提交状态常量
const (
	SubmissionStatusPending = "pending"
	SubmissionStatusGrading = "grading"
	SubmissionStatusGraded  = "graded"
)

// 竞赛类型常量
const (
	ContestTypeJeopardy    = "jeopardy"
	ContestTypeAgentBattle = "agent_battle"
)

// 竞赛状态常量
const (
	ContestStatusDraft     = "draft"
	ContestStatusPublished = "published"
	ContestStatusOngoing   = "ongoing"
	ContestStatusEnded     = "ended"
)

// 挑战分类常量
const (
	ChallengeCategoryContract   = "contract_vuln"
	ChallengeCategoryDefi       = "defi"
	ChallengeCategoryConsensus  = "consensus"
	ChallengeCategoryCrypto     = "crypto"
	ChallengeCategoryCrossChain = "cross_chain"
	ChallengeCategoryNFT        = "nft"
	ChallengeCategoryReverse    = "reverse"
	ChallengeCategoryKeyMgmt    = "key_management"
	ChallengeCategoryMisc       = "misc"
)

// 解题赛运行形态常量
const (
	ChallengeRuntimeStatic              = "static"
	ChallengeRuntimeSingleChainInstance = "single_chain_instance"
	ChallengeRuntimeForkReplay          = "fork_replay"
	ChallengeRuntimeMultiServiceLab     = "multi_service_lab"
)

// 对抗赛模式常量
const (
	BattleModeStrategyBattle = "strategy_battle"
)

// 竞赛级别常量
const (
	ContestLevelPractice    = "practice"
	ContestLevelSchool      = "school"
	ContestLevelCrossSchool = "cross_school"
	ContestLevelPlatform    = "platform"
)

// 题目来源类型常量
const (
	ChallengeSourcePreset        = "preset"
	ChallengeSourceAutoConverted = "auto_converted"
	ChallengeSourceUserCreated   = "user_created"
)

// 通知类型常量
const (
	NotifyTypeSystem     = "system"
	NotifyTypeCourse     = "course"
	NotifyTypeExperiment = "experiment"
	NotifyTypeContest    = "contest"
	NotifyTypeDiscuss    = "discuss"
)

// 资料类型常量
const (
	MaterialTypeVideo    = "video"
	MaterialTypeDocument = "document"
	MaterialTypeLink     = "link"
)

// JSONMap JSON对象类型
type JSONMap map[string]interface{}

// Scan 实现Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现Valuer接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// JSONArray JSON数组类型
type JSONArray []interface{}

// Scan 实现Scanner接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = JSONArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现Valuer接口
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

// StringList 字符串列表（JSON存储）
type StringList []string

// Scan 实现Scanner接口
func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现Valuer接口
func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}
