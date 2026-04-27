package response

import "time"

// AgentBattleRoundResponse 对抗赛轮次响应。
type AgentBattleRoundResponse struct {
	ID          uint      `json:"id"`
	ContestID   uint      `json:"contest_id"`
	RoundNumber int       `json:"round_number"`
	Status      string    `json:"status"`
	Phase       string    `json:"phase"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	UpgradeWindowEnd *time.Time `json:"upgrade_window_end,omitempty"`
	ChainRPCURL string    `json:"chain_rpc_url"`
	BlockHeight int64     `json:"block_height"`
	TxCount     int       `json:"tx_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// AgentContractResponse 策略智能体合约响应。
type AgentContractResponse struct {
	ID              uint       `json:"id"`
	ContestID       uint       `json:"contest_id"`
	TeamID          uint       `json:"team_id"`
	TeamName        string     `json:"team_name"`
	ContractAddress string     `json:"contract_address"`
	Status          string     `json:"status"`
	Version         int        `json:"version"`
	DeployedAt      *time.Time `json:"deployed_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// AgentBattleScoreResponse 对抗赛单轮得分响应。
type AgentBattleScoreResponse struct {
	Rank         int    `json:"rank"`
	TeamID       uint   `json:"team_id"`
	TeamName     string `json:"team_name"`
	Score        int    `json:"score"`
	TokenBalance string `json:"token_balance"`
	SuccessCount int    `json:"success_count"`
	FailCount    int    `json:"fail_count"`
	ResourceHeld int    `json:"resource_held"`
}

// TeamWorkspaceResponse 队伍工作区响应。
type TeamWorkspaceResponse struct {
	TeamID      uint   `json:"team_id"`
	TeamName    string `json:"team_name"`
	EnvID       string `json:"env_id"`
	PodName     string `json:"pod_name"`
	Status      string `json:"status"`
	AccessURL   string `json:"access_url"`
	ChainRPCURL string `json:"chain_rpc_url"`
	Tools       []RuntimeToolResponse `json:"tools,omitempty"`
}
