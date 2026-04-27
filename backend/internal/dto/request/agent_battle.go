package request

// DeployAgentContractRequest 部署智能体合约请求
type DeployAgentContractRequest struct {
	ContestID  uint   `json:"contest_id" binding:"required"`
	SourceCode string `json:"source_code" binding:"required"`
}

// ListAgentContractsRequest 智能体合约列表请求
type ListAgentContractsRequest struct {
	PaginationRequest
	ContestID uint   `form:"contest_id" binding:"required"`
	TeamID    uint   `form:"team_id"`
	Status    string `form:"status"`
}

// ListBattleRoundsRequest 对抗赛轮次列表请求
type ListBattleRoundsRequest struct {
	PaginationRequest
	ContestID uint   `form:"contest_id" binding:"required"`
	Status    string `form:"status"`
}

// ListBattleEventsRequest 对抗赛事件列表请求
type ListBattleEventsRequest struct {
	PaginationRequest
	ContestID uint   `form:"contest_id" binding:"required"`
	RoundID   uint   `form:"round_id"`
	TeamID    uint   `form:"team_id"`
	EventType string `form:"event_type"`
}

// ListBattleScoresRequest 对抗赛分数列表请求
type ListBattleScoresRequest struct {
	PaginationRequest
	ContestID uint `form:"contest_id" binding:"required"`
	RoundID   uint `form:"round_id"`
}

// CreateAgentBattleRoundRequest 创建对抗赛轮次请求
type CreateAgentBattleRoundRequest struct {
	RoundNumber int    `json:"round_number" binding:"required,min=1"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`
	Description string `json:"description"`
}

// ListAgentBattleRoundsRequest 轮次列表请求
type ListAgentBattleRoundsRequest struct {
	PaginationRequest
	Status string `form:"status"`
}

// ListAgentBattleEventsRequest 事件列表请求
type ListAgentBattleEventsRequest struct {
	PaginationRequest
	EventType string `form:"event_type"`
	FromBlock uint64 `form:"from_block"`
}

// UpgradeAgentContractRequest 升级智能体合约请求
type UpgradeAgentContractRequest struct {
	ContestID         uint   `json:"contest_id" binding:"required"`
	NewImplementation string `json:"new_implementation" binding:"required"`
}
