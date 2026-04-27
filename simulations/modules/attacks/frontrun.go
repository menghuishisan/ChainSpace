package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// PendingTransaction 表示等待打包的交易。
type PendingTransaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     string    `json:"value"`
	GasPrice  string    `json:"gas_price"`
	GasLimit  uint64    `json:"gas_limit"`
	Data      string    `json:"data"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// FrontrunAttack 记录一次抢跑攻击过程。
type FrontrunAttack struct {
	ID            string              `json:"id"`
	Type          string              `json:"type"`
	VictimTx      *PendingTransaction `json:"victim_tx"`
	AttackerTx    *PendingTransaction `json:"attacker_tx"`
	GasPriceDelta string              `json:"gas_price_delta"`
	Profit        string              `json:"profit"`
	Success       bool                `json:"success"`
	Timestamp     time.Time           `json:"timestamp"`
}

// MempoolState 表示当前观察到的 mempool 状态。
type MempoolState struct {
	PendingTxs  []*PendingTransaction `json:"pending_txs"`
	BaseFee     string                `json:"base_fee"`
	BlockNumber uint64                `json:"block_number"`
}

// FrontrunSimulator 演示 displacement、insertion 和 suppression 三类典型 mempool 抢跑攻击。
type FrontrunSimulator struct {
	*base.BaseSimulator
	mempool *MempoolState
	attacks []*FrontrunAttack
}

// NewFrontrunSimulator 创建模拟器。
func NewFrontrunSimulator() *FrontrunSimulator {
	sim := &FrontrunSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"frontrun",
			"抢跑攻击演示器",
			"演示 displacement、insertion 和 suppression 三类典型 mempool 抢跑攻击。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*FrontrunAttack, 0),
	}

	sim.AddParam(types.Param{
		Key:         "base_fee",
		Name:        "基础费用",
		Description: "用于估算交易排序优先级的基础 gas 费用，单位为 Gwei。",
		Type:        types.ParamTypeInt,
		Default:     30,
		Min:         1,
		Max:         500,
	})

	return sim
}

// Init 初始化模拟器。
func (s *FrontrunSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	baseFee := 30
	if value, ok := config.Params["base_fee"]; ok {
		if typed, ok := value.(float64); ok {
			baseFee = int(typed)
		}
	}

	s.mempool = &MempoolState{
		PendingTxs:  make([]*PendingTransaction, 0),
		BaseFee:     fmt.Sprintf("%d Gwei", baseFee),
		BlockNumber: 12345678,
	}
	s.attacks = make([]*FrontrunAttack, 0)

	s.updateState()
	return nil
}

// AddPendingTx 向模拟 mempool 添加一笔待处理交易。
func (s *FrontrunSimulator) AddPendingTx(tx *PendingTransaction) {
	tx.Timestamp = time.Now()
	s.mempool.PendingTxs = append(s.mempool.PendingTxs, tx)

	s.EmitEvent("tx_detected", "", "", map[string]interface{}{
		"hash":      tx.Hash,
		"type":      tx.Type,
		"gas_price": tx.GasPrice,
	})
	s.updateState()
}

// ScanForOpportunities 返回当前可被利用的 mempool 机会。
func (s *FrontrunSimulator) ScanForOpportunities() []map[string]interface{} {
	opportunities := make([]map[string]interface{}, 0)

	for _, tx := range s.mempool.PendingTxs {
		switch tx.Type {
		case "swap":
			opportunities = append(opportunities, map[string]interface{}{
				"tx_hash":     tx.Hash,
				"type":        "DEX swap",
				"opportunity": "可通过先买后卖构造三明治或插队攻击。",
				"profit_est":  "0.1% - 2% of swap amount",
			})
		case "mint":
			opportunities = append(opportunities, map[string]interface{}{
				"tx_hash":     tx.Hash,
				"type":        "NFT mint",
				"opportunity": "可通过提高 gas 优先拿到热门铸造名额。",
				"profit_est":  "depends on resale premium",
			})
		case "liquidate":
			opportunities = append(opportunities, map[string]interface{}{
				"tx_hash":     tx.Hash,
				"type":        "Liquidation",
				"opportunity": "可通过抢先清算获得更高奖励。",
				"profit_est":  "5% - 10% liquidation bonus",
			})
		}
	}

	s.EmitEvent("opportunities_found", "", "", map[string]interface{}{
		"count": len(opportunities),
	})
	return opportunities
}

// SimulateDisplacementAttack 演示替换式抢跑。
func (s *FrontrunSimulator) SimulateDisplacementAttack(victimTxHash string) *FrontrunAttack {
	victimTx := &PendingTransaction{
		Hash:      victimTxHash,
		From:      "0xVictim",
		To:        "0xDEX",
		Value:     "0",
		GasPrice:  "50 Gwei",
		GasLimit:  200000,
		Data:      "0xarb(...)",
		Type:      "arbitrage",
		Timestamp: time.Now(),
	}

	attackerTx := &PendingTransaction{
		Hash:      "0xAttackerDisplacement",
		From:      "0xAttacker",
		To:        victimTx.To,
		Value:     "0",
		GasPrice:  "60 Gwei",
		GasLimit:  victimTx.GasLimit,
		Data:      victimTx.Data,
		Type:      victimTx.Type,
		Timestamp: time.Now(),
	}

	attack := &FrontrunAttack{
		ID:            fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:          "displacement",
		VictimTx:      victimTx,
		AttackerTx:    attackerTx,
		GasPriceDelta: "+10 Gwei",
		Profit:        "抢走原本属于受害者的套利机会",
		Success:       true,
		Timestamp:     time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("displacement_attack", "", "", map[string]interface{}{
		"victim_tx":    victimTx.Hash,
		"attacker_tx":  attackerTx.Hash,
		"victim_gas":   victimTx.GasPrice,
		"attacker_gas": attackerTx.GasPrice,
		"result":       "攻击者用更高 gas 费用优先执行同类交易。",
	})
	s.updateState()
	return attack
}

// SimulateInsertionAttack 演示插队式抢跑。
func (s *FrontrunSimulator) SimulateInsertionAttack(victimSwapAmount string) *FrontrunAttack {
	victimTx := &PendingTransaction{
		Hash:      "0xVictimSwap",
		From:      "0xVictim",
		To:        "0xUniswapRouter",
		Value:     victimSwapAmount,
		GasPrice:  "40 Gwei",
		GasLimit:  250000,
		Data:      "0xswapExactETHForTokens",
		Type:      "swap",
		Timestamp: time.Now(),
	}

	attackerTx := &PendingTransaction{
		Hash:      "0xAttackerFrontrun",
		From:      "0xAttacker",
		To:        "0xUniswapRouter",
		Value:     "5 ETH",
		GasPrice:  "50 Gwei",
		GasLimit:  250000,
		Data:      "0xswapExactETHForTokens",
		Type:      "swap",
		Timestamp: time.Now(),
	}

	attack := &FrontrunAttack{
		ID:            fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:          "insertion",
		VictimTx:      victimTx,
		AttackerTx:    attackerTx,
		GasPriceDelta: "+10 Gwei",
		Profit:        "通过抬高价格后再回撤卖出获利",
		Success:       true,
		Timestamp:     time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("insertion_attack", "", "", map[string]interface{}{
		"victim_tx": victimTx.Hash,
		"summary":   "攻击者在目标交易前插入买单，随后在目标成交后反向卖出。",
	})
	s.updateState()
	return attack
}

// SimulateSuppressionAttack 演示压制式抢跑。
func (s *FrontrunSimulator) SimulateSuppressionAttack(targetTxHash string, reason string) *FrontrunAttack {
	victimTx := &PendingTransaction{
		Hash:      targetTxHash,
		From:      "0xVictim",
		To:        "0xTarget",
		Value:     "0",
		GasPrice:  "35 Gwei",
		GasLimit:  300000,
		Data:      "0xexecute()",
		Type:      "targeted",
		Timestamp: time.Now(),
	}

	attackerTx := &PendingTransaction{
		Hash:      "0xAttackerSuppression",
		From:      "0xAttacker",
		To:        "0xSpam",
		Value:     "0",
		GasPrice:  "120 Gwei",
		GasLimit:  300000,
		Data:      "0xspam()",
		Type:      "spam",
		Timestamp: time.Now(),
	}

	attack := &FrontrunAttack{
		ID:            fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:          "suppression",
		VictimTx:      victimTx,
		AttackerTx:    attackerTx,
		GasPriceDelta: "+85 Gwei",
		Profit:        reason,
		Success:       true,
		Timestamp:     time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("suppression_attack", "", "", map[string]interface{}{
		"method": "攻击者发送高费用垃圾交易填满区块容量。",
		"effect": "目标交易在关键窗口内无法进入区块。",
		"reason": reason,
	})
	s.updateState()
	return attack
}

// ExplainMechanism 返回 mempool 排序与抢跑原理。
func (s *FrontrunSimulator) ExplainMechanism() map[string]interface{} {
	return map[string]interface{}{
		"definition": "抢跑利用公开 mempool 中可见的待确认交易，通过更高优先费或私有通道改变执行顺序。",
		"how_mempool_works": []string{
			"用户先把交易广播到节点网络。",
			"交易进入 mempool 等待打包。",
			"区块构建者通常优先选择更高收益的交易。",
			"攻击者据此插队、替换或阻塞目标交易。",
		},
	}
}

// ExplainMEV 返回 MEV 背景说明。
func (s *FrontrunSimulator) ExplainMEV() map[string]interface{} {
	return map[string]interface{}{
		"definition": "MEV 是区块构建者通过排序、插入或剔除交易获得的额外价值。",
		"sources": []string{
			"DEX 套利",
			"清算竞争",
			"三明治攻击",
			"热门铸造抢跑",
		},
	}
}

// ShowDefenses 返回防御方式。
func (s *FrontrunSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "Flashbots Protect",
			"description": "通过私有交易通道发送订单，减少公开 mempool 暴露。",
		},
		{
			"name":        "批量拍卖",
			"description": "将一批订单统一清算，削弱单笔交易先后顺序的价值。",
		},
		{
			"name":        "Commit-Reveal",
			"description": "先提交承诺哈希，再公开交易详情，降低攻击者提前复制交易的能力。",
		},
		{
			"name":        "滑点与价格保护",
			"description": "对单笔交易设置严格滑点上限与私有路由。",
		},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *FrontrunSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "DeFi sandwich MEV",
			"period": "2020-2024",
			"detail": "大量 DEX 交易在公开 mempool 中被插队，受害者承受额外滑点。",
		},
		{
			"name":   "NFT mint gas wars",
			"period": "2021-2022",
			"detail": "热门 NFT 铸造中，高 gas 交易挤占普通用户的上链机会。",
		},
	}
}

// updateState 更新状态。
func (s *FrontrunSimulator) updateState() {
	if s.mempool == nil {
		return
	}
	s.SetGlobalData("pending_tx_count", len(s.mempool.PendingTxs))
	s.SetGlobalData("attack_count", len(s.attacks))
	s.SetGlobalData("base_fee", s.mempool.BaseFee)
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("victim_balance", 0)
		s.SetGlobalData("attacker_balance", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发抢跑攻击，请观察目标交易如何被发现、插队或压制。")
		setAttackTeachingState(
			s.BaseSimulator,
			"economic",
			"idle",
			"等待抢跑攻击场景。",
			"可以先触发一种 displacement、insertion 或 suppression 攻击，观察目标交易排序如何被改变。",
			0,
			map[string]interface{}{
				"pending_tx_count": len(s.mempool.PendingTxs),
				"base_fee":         s.mempool.BaseFee,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "discover_victim_tx",
			"caller":      "attacker",
			"function":    "scan_mempool",
			"target":      latest.VictimTx.Hash,
			"amount":      latest.VictimTx.Value,
			"call_depth":  1,
			"description": "攻击者先在 mempool 中发现目标交易，并判断它是否值得抢跑。",
		},
		{
			"step":        2,
			"action":      latest.Type,
			"caller":      latest.AttackerTx.From,
			"function":    "submit_priority_tx",
			"target":      latest.AttackerTx.Hash,
			"amount":      latest.GasPriceDelta,
			"call_depth":  2,
			"description": "攻击者构造更高优先级的交易，尝试改变原始交易的执行顺序。",
		},
		{
			"step":        3,
			"action":      "capture_mev",
			"caller":      latest.AttackerTx.From,
			"function":    "finalize_profit",
			"target":      latest.VictimTx.To,
			"amount":      latest.Profit,
			"call_depth":  3,
			"description": "如果攻击成功，受害者会承受更差价格或失去机会，攻击者则拿走额外收益。",
		},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest.Profit)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))
	s.SetGlobalData("victim_balance", parseAmountText(latest.VictimTx.Value))
	s.SetGlobalData("attacker_balance", parseAmountText(latest.Profit))

	setAttackTeachingState(
		s.BaseSimulator,
		"economic",
		"reordered",
		latest.Profit,
		"重点观察目标交易在哪一步被插队，以及攻击者如何从排序优势中提取 MEV。",
		1.0,
		map[string]interface{}{
			"attack_type":      latest.Type,
			"victim_tx":        latest.VictimTx.Hash,
			"attacker_tx":      latest.AttackerTx.Hash,
			"gas_price_delta":  latest.GasPriceDelta,
			"attacker_profit":  latest.Profit,
		},
	)
}

// ExecuteAction 执行动作面板交互。
func (s *FrontrunSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_displacement_attack":
		victimTxHash := actionString(params, "victim_tx_hash", "0xVictimTx")
		attack := s.SimulateDisplacementAttack(victimTxHash)
		return actionResultWithFeedback(
			"已执行替换式抢跑攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入复制目标交易并提高 gas 排序优先级的抢跑流程。",
				NextHint:    "重点观察攻击者如何用更高 gas 取代受害者的套利机会。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"victim_tx":       attack.VictimTx.Hash,
					"attacker_tx":     attack.AttackerTx.Hash,
					"gas_price_delta": attack.GasPriceDelta,
				},
			},
		), nil
	case "simulate_insertion_attack":
		victimSwapAmount := actionString(params, "victim_swap_amount", "100 ETH")
		attack := s.SimulateInsertionAttack(victimSwapAmount)
		return actionResultWithFeedback(
			"已执行插队式抢跑攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入目标交易前插入买单、后续回撤卖出的抢跑流程。",
				NextHint:    "重点观察价格曲线如何被提前推高，以及受害者如何在更差价格成交。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"victim_tx":       attack.VictimTx.Hash,
					"attacker_tx":     attack.AttackerTx.Hash,
					"gas_price_delta": attack.GasPriceDelta,
				},
			},
		), nil
	case "simulate_suppression_attack":
		targetTxHash := actionString(params, "target_tx_hash", "0xTargetTx")
		reason := actionString(params, "reason", "阻塞目标交易进入区块")
		attack := s.SimulateSuppressionAttack(targetTxHash, reason)
		return actionResultWithFeedback(
			"已执行压制式抢跑攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入用高费用垃圾交易压制目标交易的攻击流程。",
				NextHint:    "重点观察目标交易如何被持续挤出区块窗口。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"victim_tx":       attack.VictimTx.Hash,
					"attacker_tx":     attack.AttackerTx.Hash,
					"gas_price_delta": attack.GasPriceDelta,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported frontrun action: %s", action)
	}
}

// FrontrunFactory 模拟器工厂。
type FrontrunFactory struct{}

// Create 创建模拟器。
func (f *FrontrunFactory) Create() engine.Simulator { return NewFrontrunSimulator() }

// GetDescription 返回描述。
func (f *FrontrunFactory) GetDescription() types.Description { return NewFrontrunSimulator().GetDescription() }

// NewFrontrunFactory 创建工厂。
func NewFrontrunFactory() *FrontrunFactory { return &FrontrunFactory{} }

var _ engine.SimulatorFactory = (*FrontrunFactory)(nil)
