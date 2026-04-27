package attacks

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// SandwichAttack 描述一次三明治攻击的关键结果。
type SandwichAttack struct {
	VictimTx       string   `json:"victim_tx"`
	FrontrunTx     string   `json:"frontrun_tx"`
	BackrunTx      string   `json:"backrun_tx"`
	VictimLoss     string   `json:"victim_loss"`
	AttackerProfit string   `json:"attacker_profit"`
	Steps          []string `json:"steps"`
}

// SandwichSimulator 演示基于 mempool 观察的三明治攻击。
type SandwichSimulator struct {
	*base.BaseSimulator
	history []SandwichAttack
}

// NewSandwichSimulator 创建三明治攻击模拟器。
func NewSandwichSimulator() *SandwichSimulator {
	return &SandwichSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"sandwich",
			"三明治攻击演示器",
			"演示攻击者如何在目标交易前后插入交易，从滑点和价格变化中获利。",
			"attacks",
			types.ComponentAttack,
		),
		history: make([]SandwichAttack, 0),
	}
}

// Init 初始化模拟器。
func (s *SandwichSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.history = make([]SandwichAttack, 0)
	s.updateState()
	return nil
}

// SimulateSandwichAttack 演示三明治攻击流程。
func (s *SandwichSimulator) SimulateSandwichAttack(victimAmount string, slippage float64) map[string]interface{} {
	attack := SandwichAttack{
		VictimTx:       fmt.Sprintf("受害者兑换 %s", victimAmount),
		FrontrunTx:     "攻击者提前买入目标代币",
		BackrunTx:      "攻击者在受害者成交后反向卖出",
		VictimLoss:     fmt.Sprintf("%.2f%% 的额外滑点", slippage*100),
		AttackerProfit: "通过价格差与滑点差获利",
		Steps: []string{
			"攻击者在 mempool 中发现大额兑换交易。",
			"攻击者先发送高优先级买单，推动池子价格变化。",
			"受害者交易以更差价格成交。",
			"攻击者随后卖出此前买入的资产，锁定利润。",
		},
	}

	s.history = append(s.history, attack)

	data := map[string]interface{}{
		"attack": attack,
		"pool_observation": []string{
			"价格曲线先被前置交易推高。",
			"受害者在更差的成交价位执行。",
			"回撤交易吃掉受害者造成的价格位移。",
		},
	}
	s.EmitEvent("sandwich_attack", "", "", data)
	s.updateState()
	return data
}

// ShowDefenses 返回防御建议。
func (s *SandwichSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "私有交易通道", "description": "避免将高价值订单直接暴露在公开 mempool 中。"},
		{"name": "限制可接受滑点", "description": "降低攻击者可榨取的价格空间。"},
		{"name": "批量撮合或 RFQ", "description": "减少订单逐笔暴露带来的被抢跑风险。"},
		{"name": "Commit-Reveal 类方案", "description": "在合适场景下延迟暴露关键交易参数。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *SandwichSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "DeFi 三明治攻击常态化", "impact": "大额 DEX 交易在公开 mempool 中极易遭到前后夹击。"},
		{"name": "MEV 供应链成熟化", "impact": "搜索者、构建者与验证者之间形成了稳定的排序获利链条。"},
	}
}

func (s *SandwichSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.history))

	if len(s.history) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发三明治攻击，请观察前置买入、受害成交和回撤卖出的完整过程。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"economic",
			"idle",
			"等待三明治攻击场景。",
			"可以先触发一次三明治攻击，观察前置交易、受害成交和回撤卖出的时间关系。",
			0,
			map[string]interface{}{
				"attack_count": 0,
			},
		)
		return
	}

	latest := s.history[len(s.history)-1]
	steps := make([]map[string]interface{}, 0, len(latest.Steps))
	for index, item := range latest.Steps {
		steps = append(steps, map[string]interface{}{
			"index":       index + 1,
			"title":       fmt.Sprintf("步骤 %d", index+1),
			"description": item,
		})
	}

	summary := fmt.Sprintf("受害者承受 %s，攻击者通过前后夹击完成套利。", latest.VictimLoss)
	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", summary)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"economic",
		"sandwiched",
		summary,
		"重点观察前置交易如何改变价格曲线，以及回撤交易如何兑现利润。",
		1.0,
		map[string]interface{}{
			"victim_loss":     latest.VictimLoss,
			"attacker_profit": latest.AttackerProfit,
			"step_count":      len(steps),
		},
	)
}

// ExecuteAction 执行三明治攻击演示动作。
func (s *SandwichSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_sandwich_attack":
		victimAmount := actionString(params, "victim_amount", "10000000000000000000")
		slippage := actionFloat64(params, "slippage", 0.02)
		data := s.SimulateSandwichAttack(victimAmount, slippage)
		return actionResultWithFeedback(
			"已执行三明治攻击演示。",
			data,
			&types.ActionFeedback{
				Summary:     "已进入前置买入、受害成交和回撤卖出的完整三明治流程。",
				NextHint:    "重点观察价格曲线在受害交易前后如何被两次攻击交易包夹。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"victim_amount": victimAmount,
					"slippage":      slippage,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// SandwichFactory 创建三明治攻击模拟器。
type SandwichFactory struct{}

func (f *SandwichFactory) Create() engine.Simulator { return NewSandwichSimulator() }
func (f *SandwichFactory) GetDescription() types.Description { return NewSandwichSimulator().GetDescription() }
func NewSandwichFactory() *SandwichFactory { return &SandwichFactory{} }

var _ engine.SimulatorFactory = (*SandwichFactory)(nil)
