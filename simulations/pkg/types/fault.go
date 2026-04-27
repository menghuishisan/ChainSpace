package types

import "time"

// Fault 故障定义
type Fault struct {
	ID        string                 `json:"id"`
	Type      FaultType              `json:"type"`
	Target    NodeID                 `json:"target"`
	Params    map[string]interface{} `json:"params"`
	StartTick uint64                 `json:"start_tick"`
	Duration  uint64                 `json:"duration"` // 0 = 永久
	Active    bool                   `json:"active"`
}

// FaultType 故障类型
type FaultType string

const (
	FaultTypeCrash        FaultType = "crash"         // 节点崩溃
	FaultTypeByzantine    FaultType = "byzantine"     // 拜占庭行为
	FaultTypeNetDelay     FaultType = "net_delay"     // 网络延迟
	FaultTypeNetLoss      FaultType = "net_loss"      // 网络丢包
	FaultTypeNetPartition FaultType = "net_partition" // 网络分区
	FaultTypeSlowdown     FaultType = "slowdown"      // 性能降级
)

// ByzantineBehavior 拜占庭行为类型
type ByzantineBehavior string

const (
	// 消息相关
	ByzantineIgnore     ByzantineBehavior = "ignore"      // 不响应
	ByzantineDelay      ByzantineBehavior = "delay"       // 延迟响应
	ByzantineDropRandom ByzantineBehavior = "drop_random" // 随机丢弃

	// 内容篡改
	ByzantineEquivocate  ByzantineBehavior = "equivocate"   // 对不同节点发送不同消息
	ByzantineInvalidSign ByzantineBehavior = "invalid_sign" // 无效签名
	ByzantineWrongDigest ByzantineBehavior = "wrong_digest" // 错误摘要
	ByzantineReplay      ByzantineBehavior = "replay"       // 重放旧消息

	// 协议违反
	ByzantineEarlyPropose ByzantineBehavior = "early_propose" // 提前提议
	ByzantineDoubleVote   ByzantineBehavior = "double_vote"   // 双重投票
	ByzantineSkipPhase    ByzantineBehavior = "skip_phase"    // 跳过阶段

	// 合谋攻击
	ByzantineCollusion ByzantineBehavior = "collusion" // 与其他拜占庭节点合谋
)

// ByzantineConfig 拜占庭配置
type ByzantineConfig struct {
	Behavior    ByzantineBehavior      `json:"behavior"`
	Probability float64                `json:"probability"` // 触发概率 0-1
	Targets     []NodeID               `json:"targets"`     // 针对特定节点
	Delay       time.Duration          `json:"delay"`       // 延迟时间
	Colluders   []NodeID               `json:"colluders"`   // 合谋者列表
	CustomData  map[string]interface{} `json:"custom_data"`
}

// Attack 攻击定义
type Attack struct {
	ID        string                 `json:"id"`
	Type      AttackType             `json:"type"`
	Target    string                 `json:"target"` // 目标（节点ID/合约地址/交易哈希）
	Params    map[string]interface{} `json:"params"`
	StartTick uint64                 `json:"start_tick"`
	Duration  uint64                 `json:"duration"` // 0 = 永久
	Active    bool                   `json:"active"`
}

// AttackType 攻击类型
type AttackType string

const (
	// 共识层攻击
	Attack51Percent      AttackType = "51_percent"
	AttackSelfishMining  AttackType = "selfish_mining"
	AttackLongRange      AttackType = "long_range"
	AttackNothingAtStake AttackType = "nothing_at_stake"
	AttackBribery        AttackType = "bribery"

	// 网络层攻击
	AttackEclipse          AttackType = "eclipse"
	AttackSybil            AttackType = "sybil"
	AttackBGPHijack        AttackType = "bgp_hijack"
	AttackNetworkPartition AttackType = "network_partition"

	// 合约层攻击
	AttackReentrancy      AttackType = "reentrancy"
	AttackOverflow        AttackType = "overflow"
	AttackAccessControl   AttackType = "access_control"
	AttackDelegatecall    AttackType = "delegatecall"
	AttackSelfDestruct    AttackType = "selfdestruct"
	AttackDOS             AttackType = "dos"
	AttackSignatureReplay AttackType = "signature_replay"
	AttackWeakRandomness  AttackType = "weak_randomness"
	AttackTimestamp       AttackType = "timestamp"

	// DeFi攻击
	AttackFlashloan    AttackType = "flashloan"
	AttackOracle       AttackType = "oracle"
	AttackSandwich     AttackType = "sandwich"
	AttackFrontrun     AttackType = "frontrun"
	AttackGovernance   AttackType = "governance"
	AttackLiquidation  AttackType = "liquidation"
	AttackInfiniteMint AttackType = "infinite_mint"

	// 跨链攻击
	AttackBridge      AttackType = "bridge"
	AttackFakeDeposit AttackType = "fake_deposit"
)

// AttackResult 攻击结果
type AttackResult struct {
	AttackID  string                 `json:"attack_id"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Impact    map[string]interface{} `json:"impact"`
	Tick      uint64                 `json:"tick"`
	Timestamp time.Time              `json:"timestamp"`
}

// FaultInjector 故障注入器接口
type FaultInjector interface {
	InjectFault(fault *Fault) error
	RemoveFault(faultID string) error
	GetActiveFaults() []*Fault
	ClearAllFaults() error
}

// AttackInjector 攻击注入器接口
type AttackInjector interface {
	InjectAttack(attack *Attack) error
	RemoveAttack(attackID string) error
	GetActiveAttacks() []*Attack
	ClearAllAttacks() error
	GetAttackResult(attackID string) *AttackResult
}
