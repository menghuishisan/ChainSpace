package model

import (
	"database/sql/driver"
	"encoding/json"
)

// ChallengeWorkspaceSpec 定义解题赛题目的主工作区。
type ChallengeWorkspaceSpec struct {
	Image            string            `json:"image"`
	DisplayName      string            `json:"display_name"`
	Template         string            `json:"template"`
	InteractionTools []string          `json:"interaction_tools"`
	Resources        map[string]string `json:"resources"`
	InitScripts      []string          `json:"init_scripts"`
}

// ChallengeServicePort 定义题目附加服务的端口暴露方式。
type ChallengeServicePort struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	ExposeAs string `json:"expose_as"`
}

// ChallengeServiceSpec 定义解题赛题目的附加服务组件。
type ChallengeServiceSpec struct {
	Key         string                 `json:"key"`
	Image       string                 `json:"image"`
	Purpose     string                 `json:"purpose"`
	Description string                 `json:"description"`
	Ports       []ChallengeServicePort `json:"ports"`
	Env         JSONMap                `json:"env"`
}

// ChallengeTopologySpec 描述题目环境的拓扑和暴露方式。
type ChallengeTopologySpec struct {
	Mode          string   `json:"mode"`
	ExposedEntrys []string `json:"exposed_entries"`
	SharedNetwork bool     `json:"shared_network"`
}

// ChallengeForkSpec 描述题目环境的 Fork 复现配置。
type ChallengeForkSpec struct {
	Enabled      bool   `json:"enabled"`
	Chain        string `json:"chain"`
	ChainID      int    `json:"chain_id"`
	Label        string `json:"label"`
	RPCURL       string `json:"rpc_url"`
	BlockNumber  uint64 `json:"block_number"`
	TargetTxHash string `json:"target_tx_hash"`
}

// ChallengeScenarioSpec 描述真实漏洞题的复现场景。
type ChallengeScenarioSpec struct {
	ContractAddress string   `json:"contract_address"`
	AttackGoal      string   `json:"attack_goal"`
	InitSteps       []string `json:"init_steps"`
	SolveSteps      []string `json:"solve_steps"`
	DefenseGoal     string   `json:"defense_goal"`
}

// ChallengeLifecycleSpec 描述题目环境的生命周期策略。
type ChallengeLifecycleSpec struct {
	TimeLimitMinutes int  `json:"time_limit_minutes"`
	AutoDestroy      bool `json:"auto_destroy"`
	ReuseRunningEnv  bool `json:"reuse_running_env"`
}

// ChallengeValidationSpec 描述环境题的验证方式。
type ChallengeValidationSpec struct {
	Mode        string `json:"mode"`
	Description string `json:"description"`
}

// ChallengeOrchestration 定义解题赛题目的完整环境编排。
type ChallengeOrchestration struct {
	Mode             string                  `json:"mode"`
	NeedsEnvironment bool                    `json:"needs_environment"`
	Workspace        ChallengeWorkspaceSpec  `json:"workspace"`
	Services         []ChallengeServiceSpec  `json:"services"`
	Topology         ChallengeTopologySpec   `json:"topology"`
	Fork             ChallengeForkSpec       `json:"fork"`
	Scenario         ChallengeScenarioSpec   `json:"scenario"`
	Lifecycle        ChallengeLifecycleSpec  `json:"lifecycle"`
	Validation       ChallengeValidationSpec `json:"validation"`
}

// Scan 实现 Scanner 接口。
func (o *ChallengeOrchestration) Scan(value interface{}) error {
	if value == nil {
		*o = ChallengeOrchestration{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, o)
}

// Value 实现 Valuer 接口。
func (o ChallengeOrchestration) Value() (driver.Value, error) {
	return json.Marshal(o)
}

// BattleSharedChainSpec 描述对抗赛的共享链环境。
type BattleSharedChainSpec struct {
	Image           string            `json:"image"`
	ChainType       string            `json:"chain_type"`
	NetworkID       int               `json:"network_id"`
	BlockTime       int               `json:"block_time"`
	InitialBalances map[string]string `json:"initial_balances"`
	RPCURL          string            `json:"rpc_url"`
}

// BattleJudgeSpec 描述对抗赛的裁判与评分组件。
type BattleJudgeSpec struct {
	Image             string         `json:"image"`
	JudgeContract     string         `json:"judge_contract"`
	TokenContract     string         `json:"token_contract"`
	StrategyInterface string         `json:"strategy_interface"`
	ResourceModel     string         `json:"resource_model"`
	ScoringModel      string         `json:"scoring_model"`
	ScoreWeights      map[string]int `json:"score_weights"`
	AllowedActions    []string       `json:"allowed_actions"`
	ForbiddenCalls    []string       `json:"forbidden_calls"`
}

// BattleTeamWorkspaceSpec 描述每支队伍的开发环境模板。
type BattleTeamWorkspaceSpec struct {
	Image            string            `json:"image"`
	DisplayName      string            `json:"display_name"`
	InteractionTools []string          `json:"interaction_tools"`
	Resources        map[string]string `json:"resources"`
}

// BattleSpectateSpec 描述观战和回放能力。
type BattleSpectateSpec struct {
	EnableMonitor bool `json:"enable_monitor"`
	EnableReplay  bool `json:"enable_replay"`
}

// BattleLifecycleSpec 描述对抗赛轮次与生命周期配置。
type BattleLifecycleSpec struct {
	RoundDurationSeconds int  `json:"round_duration_seconds"`
	UpgradeWindowSeconds int  `json:"upgrade_window_seconds"`
	TotalRounds          int  `json:"total_rounds"`
	AutoCleanup          bool `json:"auto_cleanup"`
}

// BattleOrchestration 定义对抗赛的完整环境与赛制编排。
type BattleOrchestration struct {
	SharedChain   BattleSharedChainSpec   `json:"shared_chain"`
	Judge         BattleJudgeSpec         `json:"judge"`
	TeamWorkspace BattleTeamWorkspaceSpec `json:"team_workspace"`
	Spectate      BattleSpectateSpec      `json:"spectate"`
	Lifecycle     BattleLifecycleSpec     `json:"lifecycle"`
}

// Scan 实现 Scanner 接口。
func (o *BattleOrchestration) Scan(value interface{}) error {
	if value == nil {
		*o = BattleOrchestration{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, o)
}

// Value 实现 Valuer 接口。
func (o BattleOrchestration) Value() (driver.Value, error) {
	return json.Marshal(o)
}
