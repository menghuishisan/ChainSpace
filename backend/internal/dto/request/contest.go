package request

import "github.com/chainspace/backend/internal/model"

type ChallengeHintPayload struct {
	Content string `json:"content" binding:"required"`
	Cost    int    `json:"cost" binding:"min=0"`
}

// CreateContestRequest 创建竞赛请求
type CreateContestRequest struct {
	Level               string                    `json:"level" binding:"omitempty,oneof=practice school cross_school platform"`
	Title               string                    `json:"title" binding:"required,max=200"`
	Description         string                    `json:"description" binding:"omitempty"`
	Type                string                    `json:"type" binding:"required,oneof=jeopardy agent_battle"`
	Cover               string                    `json:"cover" binding:"omitempty,max=500"`
	Rules               string                    `json:"rules" binding:"omitempty"`
	StartTime           string                    `json:"start_time" binding:"required"`
	EndTime             string                    `json:"end_time" binding:"required"`
	RegistrationStart   string                    `json:"registration_start" binding:"omitempty"`
	RegistrationEnd     string                    `json:"registration_end" binding:"omitempty"`
	IsPublic            bool                      `json:"is_public"`
	MaxParticipants     int                       `json:"max_participants" binding:"omitempty,min=0"`
	TeamMinSize         int                       `json:"team_min_size" binding:"omitempty,min=1"`
	TeamMaxSize         int                       `json:"team_max_size" binding:"omitempty,min=1"`
	DynamicScore        bool                      `json:"dynamic_score"`
	FirstBloodBonus     int                       `json:"first_blood_bonus" binding:"omitempty,min=0"`
	BattleOrchestration model.BattleOrchestration `json:"battle_orchestration"`
}

// UpdateContestRequest 更新竞赛请求
type UpdateContestRequest struct {
	Level               string                     `json:"level" binding:"omitempty,oneof=practice school cross_school platform"`
	Title               string                     `json:"title" binding:"omitempty,max=200"`
	Description         string                     `json:"description" binding:"omitempty"`
	Cover               string                     `json:"cover" binding:"omitempty,max=500"`
	Rules               string                     `json:"rules" binding:"omitempty"`
	StartTime           string                     `json:"start_time" binding:"omitempty"`
	EndTime             string                     `json:"end_time" binding:"omitempty"`
	RegistrationStart   string                     `json:"registration_start" binding:"omitempty"`
	RegistrationEnd     string                     `json:"registration_end" binding:"omitempty"`
	IsPublic            *bool                      `json:"is_public" binding:"omitempty"`
	MaxParticipants     *int                       `json:"max_participants" binding:"omitempty,min=0"`
	TeamMinSize         *int                       `json:"team_min_size" binding:"omitempty,min=1"`
	TeamMaxSize         *int                       `json:"team_max_size" binding:"omitempty,min=1"`
	DynamicScore        *bool                      `json:"dynamic_score" binding:"omitempty"`
	FirstBloodBonus     *int                       `json:"first_blood_bonus" binding:"omitempty,min=0"`
	Status              string                     `json:"status" binding:"omitempty,oneof=draft published ongoing ended"`
	BattleOrchestration *model.BattleOrchestration `json:"battle_orchestration" binding:"omitempty"`
}

// CreateChallengeRequest 创建题目请求
type CreateChallengeRequest struct {
	Title                  string                       `json:"title" binding:"required,max=200"`
	Description            string                       `json:"description" binding:"omitempty"`
	Category               string                       `json:"category" binding:"required,oneof=contract_vuln defi consensus crypto cross_chain nft reverse key_management misc"`
	RuntimeProfile         string                       `json:"runtime_profile" binding:"omitempty,oneof=static single_chain_instance fork_replay multi_service_lab"`
	Difficulty             int                          `json:"difficulty" binding:"omitempty,min=1,max=5"`
	BasePoints             int                          `json:"base_points" binding:"omitempty,min=1"`
	MinPoints              int                          `json:"min_points" binding:"omitempty,min=1"`
	DecayFactor            float64                      `json:"decay_factor" binding:"omitempty,min=0,max=1"`
	ContractCode           string                       `json:"contract_code" binding:"omitempty"`
	SetupCode              string                       `json:"setup_code" binding:"omitempty"`
	DeployScript           string                       `json:"deploy_script" binding:"omitempty"`
	CheckScript            string                       `json:"check_script" binding:"omitempty"`
	FlagType               string                       `json:"flag_type" binding:"omitempty,oneof=static dynamic"`
	FlagTemplate           string                       `json:"flag_template" binding:"omitempty,max=200"`
	Hints                  []ChallengeHintPayload       `json:"hints" binding:"omitempty,dive"`
	Attachments            []string                     `json:"attachments" binding:"omitempty,dive,max=500"`
	ChallengeOrchestration model.ChallengeOrchestration `json:"challenge_orchestration"`
	Tags                   []string                     `json:"tags" binding:"omitempty"`
	IsPublic               bool                         `json:"is_public"`
}

// UpdateChallengeRequest 更新题目请求
type UpdateChallengeRequest struct {
	Title                  *string                       `json:"title" binding:"omitempty,max=200"`
	Description            *string                       `json:"description" binding:"omitempty"`
	Category               *string                       `json:"category" binding:"omitempty,oneof=contract_vuln defi consensus crypto cross_chain nft reverse key_management misc"`
	RuntimeProfile         *string                       `json:"runtime_profile" binding:"omitempty,oneof=static single_chain_instance fork_replay multi_service_lab"`
	Difficulty             *int                          `json:"difficulty" binding:"omitempty,min=1,max=5"`
	BasePoints             *int                          `json:"base_points" binding:"omitempty,min=1"`
	MinPoints              *int                          `json:"min_points" binding:"omitempty,min=1"`
	DecayFactor            *float64                      `json:"decay_factor" binding:"omitempty,min=0,max=1"`
	ContractCode           *string                       `json:"contract_code" binding:"omitempty"`
	SetupCode              *string                       `json:"setup_code" binding:"omitempty"`
	DeployScript           *string                       `json:"deploy_script" binding:"omitempty"`
	CheckScript            *string                       `json:"check_script" binding:"omitempty"`
	FlagType               *string                       `json:"flag_type" binding:"omitempty,oneof=static dynamic"`
	FlagTemplate           *string                       `json:"flag_template" binding:"omitempty,max=200"`
	Hints                  *[]ChallengeHintPayload       `json:"hints" binding:"omitempty,dive"`
	Attachments            *[]string                     `json:"attachments" binding:"omitempty,dive,max=500"`
	ChallengeOrchestration *model.ChallengeOrchestration `json:"challenge_orchestration" binding:"omitempty"`
	Tags                   *[]string                     `json:"tags" binding:"omitempty"`
	IsPublic               *bool                         `json:"is_public" binding:"omitempty"`
	Status                 *string                       `json:"status" binding:"omitempty,oneof=draft active archived"`
}

// AddChallengeToContestRequest 添加题目到竞赛请求
type AddChallengeToContestRequest struct {
	ChallengeID uint `json:"challenge_id" binding:"required"`
	Points      int  `json:"points" binding:"omitempty,min=1"`
	SortOrder   int  `json:"sort_order" binding:"omitempty,min=0"`
	IsVisible   bool `json:"is_visible"`
}

// UpdateContestChallengeRequest 更新竞赛题目请求
type UpdateContestChallengeRequest struct {
	Points    *int  `json:"points" binding:"omitempty,min=1"`
	SortOrder *int  `json:"sort_order" binding:"omitempty,min=0"`
	IsVisible *bool `json:"is_visible" binding:"omitempty"`
}

// CreateTeamRequest 创建队伍请求
type CreateTeamRequest struct {
	ContestID   uint   `json:"contest_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=100"`
	Avatar      string `json:"avatar" binding:"omitempty,max=500"`
	Description string `json:"description" binding:"omitempty"`
}

// UpdateTeamRequest 更新队伍请求
type UpdateTeamRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Avatar      string `json:"avatar" binding:"omitempty,max=500"`
	Description string `json:"description" binding:"omitempty"`
}

// JoinTeamRequest 加入队伍请求
type JoinTeamRequest struct {
	InviteCode string `json:"invite_code" binding:"required,max=50"`
}

type RequestChallengePublishRequest struct {
	Reason string `json:"reason" binding:"omitempty"`
}

// TransferLeaderRequest 转让队长请求
type TransferLeaderRequest struct {
	NewLeaderID uint `json:"new_leader_id" binding:"required"`
}

// KickMemberRequest 踢出成员请求
type KickMemberRequest struct {
	MemberID uint `json:"member_id" binding:"required"`
}

// SubmitFlagRequest 提交Flag请求
type SubmitFlagRequest struct {
	ChallengeID uint   `json:"challenge_id" binding:"required"`
	Flag        string `json:"flag" binding:"required,max=200"`
}

// StartChallengeEnvRequest 启动题目环境请求
type StartChallengeEnvRequest struct {
	ChallengeID uint `json:"challenge_id" binding:"required"`
}

// ListContestsRequest 竞赛列表请求
type ListContestsRequest struct {
	PaginationRequest
	Type     string `form:"type"`
	Level    string `form:"level"`
	Status   string `form:"status"`
	IsPublic *bool  `form:"is_public"`
	Keyword  string `form:"keyword"`
}

// ListChallengesRequest 题目列表请求
type ListChallengesRequest struct {
	PaginationRequest
	Category   string `form:"category"`
	Difficulty *int   `form:"difficulty" binding:"omitempty,min=1,max=5"`
	SourceType string `form:"source_type" binding:"omitempty,oneof=preset auto_converted user_created"`
	Status     string `form:"status"`
	IsPublic   *bool  `form:"is_public"`
	Keyword    string `form:"keyword"`
}

// ListTeamsRequest 队伍列表请求
type ListTeamsRequest struct {
	PaginationRequest
	ContestID uint   `form:"contest_id" binding:"required"`
	Status    string `form:"status"`
	Keyword   string `form:"keyword"`
}

// ListScoreboardRequest 排行榜请求
type ListScoreboardRequest struct {
	PaginationRequest
}

// InviteTeamMemberRequest 邀请队伍成员请求
type InviteTeamMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}
