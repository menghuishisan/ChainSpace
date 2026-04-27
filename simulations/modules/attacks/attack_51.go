package attacks

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// DoubleSpendTx 描述一次双花交易对。
type DoubleSpendTx struct {
	ID            string    `json:"id"`
	OriginalTx    string    `json:"original_tx"`
	OriginalTo    string    `json:"original_to"`
	ReplacementTx string    `json:"replacement_tx"`
	ReplacementTo string    `json:"replacement_to"`
	Amount        *big.Int  `json:"amount"`
	Confirmations int       `json:"confirmations"`
	DoubleSpent   bool      `json:"double_spent"`
	Timestamp     time.Time `json:"timestamp"`
}

// Attack51Record 描述一次 51% 攻击模拟结果。
type Attack51Record struct {
	ID               string         `json:"id"`
	AttackerHashrate float64        `json:"attacker_hashrate"`
	TargetConfirms   int            `json:"target_confirms"`
	PrivateChainLen  int            `json:"private_chain_len"`
	PublicChainLen   int            `json:"public_chain_len"`
	Success          bool           `json:"success"`
	DoubleSpendTx    *DoubleSpendTx `json:"double_spend_tx"`
	CostEstimate     string         `json:"cost_estimate"`
	Timestamp        time.Time      `json:"timestamp"`
}

// Attack51Simulator 演示多数算力攻击与双花流程。
type Attack51Simulator struct {
	*base.BaseSimulator
	attackerHashrate float64
	honestHashrate   float64
	attacks          []*Attack51Record
}

// NewAttack51Simulator 创建 51% 攻击模拟器。
func NewAttack51Simulator() *Attack51Simulator {
	sim := &Attack51Simulator{
		BaseSimulator: base.NewBaseSimulator(
			"attack_51",
			"51% 攻击演示器",
			"演示攻击者掌握多数算力后如何隐藏私有分支、等待确认并最终完成双花。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*Attack51Record, 0),
	}

	sim.AddParam(types.Param{
		Key:         "attacker_hashrate",
		Name:        "攻击者算力占比",
		Description: "攻击者控制的算力百分比，用于估算分叉赶超成功率。",
		Type:        types.ParamTypeFloat,
		Default:     51.0,
		Min:         10.0,
		Max:         90.0,
	})

	return sim
}

// Init 初始化算力分布。
func (s *Attack51Simulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.attackerHashrate = 51.0
	if v, ok := config.Params["attacker_hashrate"]; ok {
		if n, ok := v.(float64); ok {
			s.attackerHashrate = n
		}
	}
	s.honestHashrate = 100 - s.attackerHashrate
	s.attacks = make([]*Attack51Record, 0)
	s.updateState()
	return nil
}

// SimulateDoubleSpendAttack 演示双花攻击。
func (s *Attack51Simulator) SimulateDoubleSpendAttack(amount int64, targetConfirmations int) *Attack51Record {
	success := s.attackerHashrate >= 50
	if !success {
		ratio := s.attackerHashrate / 100
		success = math.Pow(ratio, float64(targetConfirmations)) > 0.1
	}

	privateLen := targetConfirmations
	if success {
		privateLen = targetConfirmations + 1
	}

	doubleSpend := &DoubleSpendTx{
		ID:            fmt.Sprintf("ds-%d", len(s.attacks)+1),
		OriginalTx:    "0xOriginalDeposit",
		OriginalTo:    "0xExchange",
		ReplacementTx: "0xReplacementToSelf",
		ReplacementTo: "0xAttacker",
		Amount:        new(big.Int).Mul(big.NewInt(amount), big.NewInt(1e18)),
		Confirmations: targetConfirmations,
		DoubleSpent:   success,
		Timestamp:     time.Now(),
	}

	record := &Attack51Record{
		ID:               fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackerHashrate: s.attackerHashrate,
		TargetConfirms:   targetConfirmations,
		PrivateChainLen:  privateLen,
		PublicChainLen:   targetConfirmations,
		Success:          success,
		DoubleSpendTx:    doubleSpend,
		CostEstimate:     s.estimateCost(privateLen),
		Timestamp:        time.Now(),
	}
	s.attacks = append(s.attacks, record)

	s.EmitEvent("double_spend_attack", "", "", map[string]interface{}{
		"success":           success,
		"attacker_hashrate": s.attackerHashrate,
		"target_confirms":   targetConfirmations,
		"private_chain_len": privateLen,
		"public_chain_len":  targetConfirmations,
		"amount":            amount,
		"attack_flow": []string{
			"1. 攻击者先向受害方广播一笔公开支付。",
			"2. 攻击者同时在私有分支中构造替代交易，把同一笔资金转回自己。",
			"3. 受害方等待目标确认数后交付商品或放行提款。",
			"4. 如果攻击者私链最终更长，就公开私链并替换原链历史。",
			map[bool]string{true: "5. 双花成功，原支付交易被回滚。", false: "5. 攻击失败，公开链保持优势。"}[success],
		},
	})
	s.updateState()
	return record
}

// ShowDefenses 返回常见防御措施。
func (s *Attack51Simulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "提高确认数", "description": "对高价值转账等待更多确认，增加私链赶超难度。"},
		{"name": "监测深度重组", "description": "交易所和商家应及时识别异常链重组并暂停高风险入账。"},
		{"name": "提升网络安全预算", "description": "低算力 PoW 网络更容易被租赁算力攻击。"},
		{"name": "检查点或快速终局机制", "description": "通过额外确认层降低历史被大幅改写的概率。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *Attack51Simulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Ethereum Classic", "period": "2019-2020", "impact": "多次遭受深度重组与双花攻击。"},
		{"name": "Bitcoin Gold", "date": "2018", "impact": "低算力网络在算力租赁市场下更容易成为目标。"},
	}
}

// updateState 同步前端主舞台所需状态。
func (s *Attack51Simulator) updateState() {
	s.SetGlobalData("attacker_hashrate", s.attackerHashrate)
	s.SetGlobalData("honest_hashrate", s.honestHashrate)
	s.SetGlobalData("attack_count", len(s.attacks))

	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发 51% 攻击，可以观察双花攻击如何依赖私链赶超。")
		setAttackTeachingState(
			s.BaseSimulator,
			"consensus",
			"idle",
			"等待多数算力攻击场景。",
			"可以先触发一次双花攻击，观察公开链与私链的竞争过程。",
			0,
			map[string]interface{}{
				"attacker_hashrate": s.attackerHashrate,
				"honest_hashrate":   s.honestHashrate,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "broadcast_payment",
			"caller":      "attacker",
			"function":    "public_transaction",
			"target":      latest.DoubleSpendTx.OriginalTo,
			"amount":      latest.DoubleSpendTx.Amount.String(),
			"call_depth":  1,
			"description": "攻击者先向受害方广播一笔公开支付，诱导对方等待确认。",
		},
		{
			"step":        2,
			"action":      "mine_private_chain",
			"caller":      "attacker",
			"function":    "private_fork",
			"target":      "private_chain",
			"amount":      fmt.Sprintf("%d", latest.PrivateChainLen),
			"call_depth":  2,
			"description": "攻击者在私有分支中隐藏替代交易，等待私链长度追平并反超公开链。",
		},
		{
			"step":        3,
			"action":      "replace_history",
			"caller":      "attacker",
			"function":    "publish_private_chain",
			"target":      "public_chain",
			"amount":      fmt.Sprintf("%d", latest.TargetConfirms),
			"call_depth":  3,
			"description": map[bool]string{
				true:  "私有链成功取代公开链，原始支付被回滚，双花攻击成立。",
				false: "私有链未能取得优势，公开链仍然保持主导，双花攻击失败。",
			}[latest.Success],
		},
	}

	summary := fmt.Sprintf("目标确认数 %d，攻击者算力占比 %.1f%%。", latest.TargetConfirms, latest.AttackerHashrate)
	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", summary)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"consensus",
		map[bool]string{true: "replaced", false: "contested"}[latest.Success],
		summary,
		"重点观察私链长度是否超过公开链，以及原交易是否最终被回滚。",
		1.0,
		map[string]interface{}{
			"attacker_hashrate": latest.AttackerHashrate,
			"private_chain_len": latest.PrivateChainLen,
			"public_chain_len":  latest.PublicChainLen,
			"success":           latest.Success,
			"cost_estimate":     latest.CostEstimate,
		},
	)
}

// ExecuteAction 执行 51% 攻击动作。
func (s *Attack51Simulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_double_spend_attack":
		amount := actionInt64(params, "amount", 100)
		targetConfirmations := actionInt(params, "target_confirmations", 6)
		record := s.SimulateDoubleSpendAttack(amount, targetConfirmations)
		return actionResultWithFeedback(
			"已执行 51% 双花攻击演示。",
			map[string]interface{}{"attack": record},
			&types.ActionFeedback{
				Summary:     "已进入公开支付、私链挖掘和历史替换的双花攻击流程。",
				NextHint:    "重点观察私链是否成功赶超，以及原支付交易是否被最终回滚。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{
					"attacker_hashrate": record.AttackerHashrate,
					"private_chain_len": record.PrivateChainLen,
					"public_chain_len":  record.PublicChainLen,
					"success":           record.Success,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported 51 attack action: %s", action)
	}
}

func (s *Attack51Simulator) estimateCost(blocksMined int) string {
	hoursNeeded := float64(blocksMined) * 10 / 60
	costPerHour := 204000000 * 0.05
	totalCost := hoursNeeded * costPerHour
	return fmt.Sprintf("$%.0f", totalCost)
}

// Attack51Factory 创建 51% 攻击模拟器。
type Attack51Factory struct{}

func (f *Attack51Factory) Create() engine.Simulator { return NewAttack51Simulator() }
func (f *Attack51Factory) GetDescription() types.Description { return NewAttack51Simulator().GetDescription() }
func NewAttack51Factory() *Attack51Factory { return &Attack51Factory{} }

var _ engine.SimulatorFactory = (*Attack51Factory)(nil)
