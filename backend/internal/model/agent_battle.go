package model

import "time"

// AgentBattleRound 描述对抗赛的单个轮次。
type AgentBattleRound struct {
	BaseModel
	ContestID    uint      `gorm:"index;not null" json:"contest_id"`
	RoundNumber  int       `gorm:"not null" json:"round_number"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `gorm:"size:20;default:pending" json:"status"`
	ChainRPCURL  string    `gorm:"size:200" json:"chain_rpc_url"`
	BlockHeight  int64     `gorm:"default:0" json:"block_height"`
	TxCount      int       `gorm:"default:0" json:"tx_count"`
	GasUsed      int64     `gorm:"default:0" json:"gas_used"`
	WinnerTeamID *uint     `json:"winner_team_id"`
	Summary      string    `gorm:"type:text" json:"summary"`

	Contest    *Contest `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	WinnerTeam *Team    `gorm:"foreignKey:WinnerTeamID" json:"winner_team,omitempty"`
}

// TableName 表名。
func (AgentBattleRound) TableName() string {
	return "agent_battle_rounds"
}

// 轮次状态常量。
const (
	RoundStatusPending  = "pending"
	RoundStatusRunning  = "running"
	RoundStatusFinished = "finished"
	RoundStatusFailed   = "failed"
)

// 对抗赛轮次阶段常量。
const (
	RoundPhasePending       = "pending"
	RoundPhaseUpgradeWindow = "upgrade_window"
	RoundPhaseLocked        = "locked"
	RoundPhaseExecuting     = "executing"
	RoundPhaseSettling      = "settling"
	RoundPhaseFinished      = "finished"
)

// AgentContract 描述队伍在某轮提交的策略智能体版本。
type AgentContract struct {
	BaseModel
	ContestID       uint       `gorm:"index;not null" json:"contest_id"`
	RoundID         uint       `gorm:"index;not null" json:"round_id"`
	TeamID          uint       `gorm:"index;not null" json:"team_id"`
	SubmitterID     uint       `gorm:"index;not null" json:"submitter_id"`
	ContractAddress string     `gorm:"size:50" json:"contract_address"`
	SourceCode      string     `gorm:"type:text" json:"source_code"`
	ByteCode        string     `gorm:"type:text" json:"byte_code"`
	ABI             string     `gorm:"type:text" json:"abi"`
	Version         int        `gorm:"default:1" json:"version"`
	Status          string     `gorm:"size:20;default:pending" json:"status"`
	IsActive        bool       `gorm:"default:false" json:"is_active"`
	DeployedAt      *time.Time `json:"deployed_at"`
	DeployTxHash    string     `gorm:"size:100" json:"deploy_tx_hash"`
	GasUsed         int64      `gorm:"default:0" json:"gas_used"`
	ErrorMessage    string     `gorm:"type:text" json:"error_message"`
	CompileLog      string     `gorm:"type:text" json:"compile_log"`

	Contest   *Contest          `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Round     *AgentBattleRound `gorm:"foreignKey:RoundID" json:"round,omitempty"`
	Team      *Team             `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Submitter *User             `gorm:"foreignKey:SubmitterID" json:"submitter,omitempty"`
}

// TableName 表名。
func (AgentContract) TableName() string {
	return "agent_contracts"
}

// 合约状态常量。
const (
	ContractStatusPending   = "pending"
	ContractStatusCompiling = "compiling"
	ContractStatusDeploying = "deploying"
	ContractStatusDeployed  = "deployed"
	ContractStatusFailed    = "failed"
)

// AgentBattleEvent 描述对抗赛中的结构化事件。
type AgentBattleEvent struct {
	BaseModel
	ContestID   uint      `gorm:"index;not null" json:"contest_id"`
	RoundID     uint      `gorm:"index;not null" json:"round_id"`
	TeamID      *uint     `gorm:"index" json:"team_id"`
	EventType   string    `gorm:"size:50;not null" json:"event_type"`
	TxHash      string    `gorm:"size:100" json:"tx_hash"`
	BlockNumber int64     `gorm:"default:0" json:"block_number"`
	FromAddress string    `gorm:"size:50" json:"from_address"`
	ToAddress   string    `gorm:"size:50" json:"to_address"`
	Value       string    `gorm:"size:100" json:"value"`
	GasUsed     int64     `gorm:"default:0" json:"gas_used"`
	EventData   JSONMap   `gorm:"type:jsonb" json:"event_data"`
	Timestamp   time.Time `gorm:"not null" json:"timestamp"`

	Contest *Contest          `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Round   *AgentBattleRound `gorm:"foreignKey:RoundID" json:"round,omitempty"`
	Team    *Team             `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName 表名。
func (AgentBattleEvent) TableName() string {
	return "agent_battle_events"
}

// 事件类型常量。
const (
	EventTypeDeployment    = "deployment"
	EventTypeTransaction   = "transaction"
	EventTypeContractCall  = "contract_call"
	EventTypeTokenTransfer = "token_transfer"
	EventTypeStateChange   = "state_change"
	EventTypeError         = "error"
	EventTypeReward        = "reward"
	EventTypePenalty       = "penalty"
)

// AgentBattleScore 描述某支队伍在某轮的得分结果。
type AgentBattleScore struct {
	BaseModel
	ContestID    uint    `gorm:"uniqueIndex:idx_contest_team_round;not null" json:"contest_id"`
	TeamID       uint    `gorm:"uniqueIndex:idx_contest_team_round;not null" json:"team_id"`
	RoundID      uint    `gorm:"uniqueIndex:idx_contest_team_round;not null" json:"round_id"`
	Score        int     `gorm:"default:0" json:"score"`
	Rank         int     `gorm:"default:0" json:"rank"`
	TokenBalance string  `gorm:"size:100;default:0" json:"token_balance"`
	TxCount      int     `gorm:"default:0" json:"tx_count"`
	SuccessCount int     `gorm:"default:0" json:"success_count"`
	FailCount    int     `gorm:"default:0" json:"fail_count"`
	GasSpent     int64   `gorm:"default:0" json:"gas_spent"`
	Profit       string  `gorm:"size:100;default:0" json:"profit"`
	Details      JSONMap `gorm:"type:jsonb" json:"details"`

	Contest *Contest          `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Team    *Team             `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Round   *AgentBattleRound `gorm:"foreignKey:RoundID" json:"round,omitempty"`
}

// TableName 表名。
func (AgentBattleScore) TableName() string {
	return "agent_battle_scores"
}
