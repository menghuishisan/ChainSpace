package response

import (
	"encoding/json"
	"time"

	"github.com/chainspace/backend/internal/model"
)

type ChallengeHintResponse struct {
	Content string `json:"content"`
	Cost    int    `json:"cost"`
}

// ContestResponse 比赛详情响应。
type ContestResponse struct {
	ID                  uint                      `json:"id"`
	SchoolID            *uint                     `json:"school_id,omitempty"`
	CreatorID           uint                      `json:"creator_id"`
	CreatorName         string                    `json:"creator_name,omitempty"`
	Title               string                    `json:"title"`
	Description         string                    `json:"description"`
	Type                string                    `json:"type"`
	Level               string                    `json:"level"`
	Cover               string                    `json:"cover"`
	Rules               string                    `json:"rules,omitempty"`
	BattleOrchestration model.BattleOrchestration `json:"battle_orchestration"`
	StartTime           time.Time                 `json:"start_time"`
	EndTime             time.Time                 `json:"end_time"`
	RegistrationStart   *time.Time                `json:"registration_start,omitempty"`
	RegistrationEnd     *time.Time                `json:"registration_end,omitempty"`
	DynamicScore        bool                      `json:"dynamic_score"`
	FirstBloodBonus     int                       `json:"first_blood_bonus"`
	Status              string                    `json:"status"`
	IsPublic            bool                      `json:"is_public"`
	MaxParticipants     int                       `json:"max_participants"`
	TeamMinSize         int                       `json:"team_min_size"`
	TeamMaxSize         int                       `json:"team_max_size"`
	ParticipantCount    int64                     `json:"participant_count,omitempty"`
	ChallengeCount      int64                     `json:"challenge_count,omitempty"`
	IsRegistered        bool                      `json:"is_registered"`
	CreatedAt           time.Time                 `json:"created_at"`
}

// ContestSummaryResponse 队伍视图中的比赛摘要。
type ContestSummaryResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ChallengeResponse 题目响应。
type ChallengeResponse struct {
	ID                     uint                         `json:"id"`
	CreatorID              uint                         `json:"creator_id"`
	SchoolID               *uint                        `json:"school_id,omitempty"`
	CreatorName            string                       `json:"creator_name,omitempty"`
	Title                  string                       `json:"title"`
	Description            string                       `json:"description"`
	Category               string                       `json:"category"`
	RuntimeProfile         string                       `json:"runtime_profile"`
	Difficulty             int                          `json:"difficulty"`
	BasePoints             int                          `json:"base_points"`
	Points                 int                          `json:"points"`
	MinPoints              int                          `json:"min_points"`
	DecayFactor            float64                      `json:"decay_factor"`
	FlagType               string                       `json:"flag_type"`
	FlagTemplate           string                       `json:"flag_template,omitempty"`
	SolveCount             int                          `json:"solve_count"`
	AttemptCount           int                          `json:"attempt_count"`
	FirstBlood             string                       `json:"first_blood,omitempty"`
	FirstBloodTime         *time.Time                   `json:"first_blood_time,omitempty"`
	AwardedPoints          *int                         `json:"awarded_points,omitempty"`
	ContractCode           string                       `json:"contract_code,omitempty"`
	SetupCode              string                       `json:"setup_code,omitempty"`
	DeployScript           string                       `json:"deploy_script,omitempty"`
	CheckScript            string                       `json:"check_script,omitempty"`
	ValidationConfig       model.JSONMap                `json:"validation_config,omitempty"`
	Hints                  []ChallengeHintResponse      `json:"hints,omitempty"`
	Attachments            []string                     `json:"attachments,omitempty"`
	ChallengeOrchestration model.ChallengeOrchestration `json:"challenge_orchestration"`
	Tags                   []string                     `json:"tags,omitempty"`
	SourceType             string                       `json:"source_type,omitempty"`
	SourceRef              string                       `json:"source_ref,omitempty"`
	Status                 string                       `json:"status"`
	IsPublic               bool                         `json:"is_public"`
	IsSolved               bool                         `json:"is_solved,omitempty"`
	CreatedAt              time.Time                    `json:"created_at"`
}

type BuildChallengeResponseOptions struct {
	Points              *int
	SolveCount          *int
	AttemptCount        *int
	FirstBlood          string
	FirstBloodTime      *time.Time
	AwardedPoints       *int
	IsSolved            *bool
	IncludeManageFields bool
}

func BuildChallengeResponse(challenge *model.Challenge, options *BuildChallengeResponseOptions) ChallengeResponse {
	resp := ChallengeResponse{
		ID:                     challenge.ID,
		CreatorID:              challenge.CreatorID,
		SchoolID:               challenge.SchoolID,
		Title:                  challenge.Title,
		Description:            challenge.Description,
		Category:               challenge.Category,
		RuntimeProfile:         challenge.RuntimeProfile,
		Difficulty:             challenge.Difficulty,
		BasePoints:             challenge.BasePoints,
		Points:                 challenge.BasePoints,
		MinPoints:              challenge.MinPoints,
		DecayFactor:            challenge.DecayFactor,
		FlagType:               challenge.FlagType,
		SolveCount:             challenge.SolveCount,
		AttemptCount:           challenge.AttemptCount,
		ContractCode:           challenge.ContractCode,
		Hints:                  normalizeChallengeHints(challenge.Hints),
		Attachments:            normalizeChallengeAttachments(challenge.Attachments),
		ChallengeOrchestration: challenge.ChallengeOrchestration,
		Tags:                   []string(challenge.Tags),
		SourceType:             challenge.SourceType,
		SourceRef:              challenge.SourceRef,
		Status:                 challenge.Status,
		IsPublic:               challenge.IsPublic,
		CreatedAt:              challenge.CreatedAt,
	}

	if challenge.Creator != nil {
		resp.CreatorName = challenge.Creator.DisplayName()
	}

	if options != nil {
		if options.Points != nil {
			resp.Points = *options.Points
		}
		if options.SolveCount != nil {
			resp.SolveCount = *options.SolveCount
		}
		if options.AttemptCount != nil {
			resp.AttemptCount = *options.AttemptCount
		}
		resp.FirstBlood = options.FirstBlood
		resp.FirstBloodTime = options.FirstBloodTime
		resp.AwardedPoints = options.AwardedPoints
		if options.IsSolved != nil {
			resp.IsSolved = *options.IsSolved
		}
		if options.IncludeManageFields {
			resp.FlagTemplate = challenge.FlagTemplate
			resp.SetupCode = challenge.SetupCode
			resp.DeployScript = challenge.DeployScript
			resp.CheckScript = challenge.CheckScript
			resp.ValidationConfig = challenge.ValidationConfig
		}
	}

	return resp
}

func normalizeChallengeHints(items model.JSONArray) []ChallengeHintResponse {
	if len(items) == 0 {
		return nil
	}

	payload, err := json.Marshal(items)
	if err != nil {
		return nil
	}

	var hints []ChallengeHintResponse
	if err := json.Unmarshal(payload, &hints); err != nil {
		return nil
	}
	return hints
}

func normalizeChallengeAttachments(items model.JSONArray) []string {
	if len(items) == 0 {
		return nil
	}

	attachments := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok || value == "" {
			continue
		}
		attachments = append(attachments, value)
	}
	return attachments
}

// ContestChallengeResponse 比赛题目关联响应。
type ContestChallengeResponse struct {
	ID            uint              `json:"id"`
	ContestID     uint              `json:"contest_id"`
	ChallengeID   uint              `json:"challenge_id"`
	Points        int               `json:"points"`
	CurrentPoints int               `json:"current_points"`
	SortOrder     int               `json:"sort_order"`
	IsVisible     bool              `json:"is_visible"`
	Challenge     ChallengeResponse `json:"challenge"`
	IsSolved      bool              `json:"is_solved,omitempty"`
	SolveCount    int64             `json:"solve_count,omitempty"`
}

// TeamResponse 队伍响应。
type TeamResponse struct {
	ID          uint                    `json:"id"`
	ContestID   uint                    `json:"contest_id"`
	Name        string                  `json:"name"`
	InviteCode  string                  `json:"invite_code,omitempty"`
	CaptainID   uint                    `json:"captain_id"`
	LeaderName  string                  `json:"leader_name,omitempty"`
	Avatar      string                  `json:"avatar"`
	Description string                  `json:"description"`
	Status      string                  `json:"status"`
	MemberCount int64                   `json:"member_count,omitempty"`
	Members     []TeamMemberResponse    `json:"members,omitempty"`
	Contest     *ContestSummaryResponse `json:"contest,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
}

// TeamMemberResponse 队伍成员响应。
type TeamMemberResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	DisplayName string    `json:"display_name"`
	RealName    string    `json:"real_name"`
	Avatar      string    `json:"avatar"`
	Role        string    `json:"role"`
	IsCaptain   bool      `json:"is_captain"`
	JoinedAt    time.Time `json:"joined_at"`
}

// ScoreboardResponse 排行榜完整响应。
type ScoreboardResponse struct {
	List    []ScoreboardEntry `json:"list"`
	MyRank  int               `json:"my_rank"`
	MyScore int               `json:"my_score"`
}

// ScoreboardEntry 排行榜条目。
type ScoreboardEntry struct {
	Rank            int        `json:"rank"`
	TeamID          *uint      `json:"team_id,omitempty"`
	TeamName        string     `json:"team_name,omitempty"`
	UserID          uint       `json:"user_id,omitempty"`
	DisplayName     string     `json:"display_name,omitempty"`
	TotalScore      int        `json:"total_score"`
	SolveCount      int        `json:"solve_count"`
	LastSolveAt     *time.Time `json:"last_solve_at,omitempty"`
	FirstBloodCount int        `json:"first_blood_count,omitempty"`
}

// FlagSubmitResponse Flag 提交响应。
type FlagSubmitResponse struct {
	Correct bool   `json:"correct"`
	Message string `json:"message"`
	Points  int    `json:"points,omitempty"`
}

// ContestSubmissionResponse 比赛提交响应。
type ContestSubmissionResponse struct {
	ID          uint      `json:"id"`
	ContestID   uint      `json:"contest_id"`
	ChallengeID uint      `json:"challenge_id"`
	UserID      uint      `json:"user_id"`
	TeamID      *uint     `json:"team_id,omitempty"`
	IsCorrect   bool      `json:"is_correct"`
	Points      int       `json:"points"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// SpectateDataResponse 观战数据响应。
type SpectateDataResponse struct {
	ContestName  string              `json:"contest_name"`
	CurrentRound int                 `json:"current_round"`
	RoundStatus  string              `json:"round_status"`
	RoundPhase   string              `json:"round_phase"`
	RoundEndTime *time.Time          `json:"round_end_time,omitempty"`
	Teams        []SpectateTeamInfo  `json:"teams"`
	RecentEvents []SpectateEventInfo `json:"recent_events"`
}

// SpectateTeamInfo 观战队伍信息。
type SpectateTeamInfo struct {
	TeamID       uint   `json:"team_id"`
	TeamName     string `json:"team_name"`
	Score        int    `json:"total_score"`
	ResourceHeld int    `json:"resource_held"`
	IsAlive      bool   `json:"is_alive"`
}

// SpectateEventInfo 观战事件信息。
type SpectateEventInfo struct {
	Block         uint64    `json:"block"`
	Time          time.Time `json:"time"`
	RoundNumber   int       `json:"round_number,omitempty"`
	EventType     string    `json:"event_type"`
	ActorTeam     string    `json:"actor_team,omitempty"`
	TargetTeam    string    `json:"target_team,omitempty"`
	ActionResult  string    `json:"action_result,omitempty"`
	ScoreDelta    int64     `json:"score_delta,omitempty"`
	ResourceDelta int64     `json:"resource_delta,omitempty"`
	Description   string    `json:"description"`
}

// ReplayDataResponse 回放数据响应。
type ReplayDataResponse struct {
	RoundID    uint             `json:"round_id"`
	StartBlock uint64           `json:"start_block"`
	EndBlock   uint64           `json:"end_block"`
	Snapshots  []ReplaySnapshot `json:"snapshots"`
}

// ReplaySnapshot 回放快照。
type ReplaySnapshot struct {
	Block  uint64            `json:"block"`
	Teams  []ReplayTeamState `json:"teams"`
	Events []ReplayEventInfo `json:"events"`
}

// ReplayTeamState 回放队伍状态。
type ReplayTeamState struct {
	TeamID   uint `json:"team_id"`
	Score    int  `json:"score"`
	Resource int  `json:"resource"`
}

// ReplayEventInfo 回放事件信息。
type ReplayEventInfo struct {
	EventType     string `json:"event_type"`
	ActorTeam     string `json:"actor_team,omitempty"`
	TargetTeam    string `json:"target_team,omitempty"`
	ActionResult  string `json:"action_result,omitempty"`
	ScoreDelta    int64  `json:"score_delta,omitempty"`
	ResourceDelta int64  `json:"resource_delta,omitempty"`
	Description   string `json:"description"`
}

// BattleStatusResponse 对抗赛状态响应。
type BattleStatusResponse struct {
	Status       string             `json:"status"`
	CurrentBlock uint64             `json:"current_block"`
	CurrentRound int                `json:"current_round"`
	TotalRounds  int                `json:"total_rounds"`
	RoundPhase   string             `json:"round_phase"`
	MyRank       int                `json:"my_rank"`
	MyScore      int64              `json:"my_score"`
	AgentStatus  *BattleAgentStatus `json:"agent_status,omitempty"`
	Teams        []BattleTeamStatus `json:"teams"`
	RecentEvents []BattleEventInfo  `json:"recent_events"`
}

// BattleAgentStatus 当前用户所属队伍的策略智能体状态。
type BattleAgentStatus struct {
	Version    string     `json:"version"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
	IsValid    bool       `json:"is_valid"`
}

// BattleTeamStatus 对抗赛队伍状态。
type BattleTeamStatus struct {
	TeamID          uint   `json:"team_id"`
	TeamName        string `json:"team_name"`
	ContractAddress string `json:"contract_address"`
	IsAlive         bool   `json:"is_alive"`
	ResourceHeld    int    `json:"resource_held,omitempty"`
	TotalScore      int64  `json:"total_score"`
}

// BattleEventInfo 对抗赛结构化事件信息。
type BattleEventInfo struct {
	Block         uint64 `json:"block"`
	RoundNumber   int    `json:"round_number,omitempty"`
	EventType     string `json:"event_type"`
	ActorTeam     string `json:"actor_team,omitempty"`
	TargetTeam    string `json:"target_team,omitempty"`
	ActionResult  string `json:"action_result,omitempty"`
	ScoreDelta    int64  `json:"score_delta,omitempty"`
	ResourceDelta int64  `json:"resource_delta,omitempty"`
	Description   string `json:"description"`
}

// BattleConfigResponse 对抗赛配置响应。
type BattleConfigResponse struct {
	ContestID        uint              `json:"contest_id"`
	ContestName      string            `json:"contest_name"`
	ChainRPC         string            `json:"chain_rpc"`
	Rounds           []BattleRoundInfo `json:"rounds"`
	CurrentRound     int               `json:"current_round"`
	TeamWorkspaceURL string            `json:"team_workspace_url,omitempty"`
}

// BattleRoundInfo 对抗赛轮次信息。
type BattleRoundInfo struct {
	RoundNumber int        `json:"round_number"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Status      string     `json:"status"`
}

// FinalRankResponse 最终排名响应。
type FinalRankResponse struct {
	Rank       int    `json:"rank"`
	TeamID     uint   `json:"team_id"`
	TeamName   string `json:"team_name"`
	TotalScore int64  `json:"total_score"`
}

// ChallengeEnvResponse 题目环境响应。
type ChallengeEnvResponse struct {
	ID             uint                     `json:"id"`
	EnvID          string                   `json:"env_id"`
	ContestID      uint                     `json:"contest_id"`
	ChallengeID    uint                     `json:"challenge_id"`
	Status         string                   `json:"status"`
	AccessURL      string                   `json:"access_url,omitempty"`
	Tools          []RuntimeToolResponse    `json:"tools,omitempty"`
	ServiceEntries []RuntimeServiceResponse `json:"service_entries,omitempty"`
	StartedAt      *time.Time               `json:"started_at,omitempty"`
	ExpiresAt      *time.Time               `json:"expires_at,omitempty"`
	ErrorMessage   string                   `json:"error_message,omitempty"`
	Remaining      int                      `json:"remaining"`
}

// ContestRecordResponse 比赛记录响应。
type ContestRecordResponse struct {
	ContestID   uint   `json:"contest_id"`
	ContestName string `json:"contest_name"`
	ContestType string `json:"contest_type"`
	TeamName    string `json:"team_name,omitempty"`
	Rank        int    `json:"rank,omitempty"`
	TotalScore  int    `json:"total_score"`
	Status      string `json:"status"`
}
