package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// FlashloanStep 表示一次闪电贷攻击中的单个步骤。
type FlashloanStep struct {
	Step        int    `json:"step"`
	Action      string `json:"action"`
	Protocol    string `json:"protocol"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
}

// FlashloanAttack 表示一次闪电贷攻击场景。
type FlashloanAttack struct {
	Name         string           `json:"name"`
	BorrowAmount string           `json:"borrow_amount"`
	Profit       string           `json:"profit"`
	Steps        []*FlashloanStep `json:"steps"`
	Protocols    []string         `json:"protocols"`
	Timestamp    time.Time        `json:"timestamp"`
}

// FlashloanSimulator 演示价格操纵、治理劫持和清算窗口利用等闪电贷攻击。
type FlashloanSimulator struct {
	*base.BaseSimulator
	attacks []*FlashloanAttack
}

// NewFlashloanSimulator 创建模拟器。
func NewFlashloanSimulator() *FlashloanSimulator {
	return &FlashloanSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"flashloan",
			"闪电贷攻击演示器",
			"演示价格操纵、治理劫持和清算窗口利用等典型闪电贷攻击过程。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*FlashloanAttack, 0),
	}
}

// Init 初始化模拟器。
func (s *FlashloanSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.attacks = make([]*FlashloanAttack, 0)
	s.updateState()
	return nil
}

// SimulatePriceManipulation 演示价格操纵攻击。
func (s *FlashloanSimulator) SimulatePriceManipulation() *FlashloanAttack {
	attack := &FlashloanAttack{
		Name:         "价格操纵",
		BorrowAmount: "10000 ETH",
		Profit:       "500 ETH",
		Protocols:    []string{"Aave", "Uniswap", "Victim Protocol"},
		Steps: []*FlashloanStep{
			{Step: 1, Action: "borrow", Protocol: "Aave", Amount: "10000 ETH", Description: "借入大额流动性，获得短时操纵市场的资金。"},
			{Step: 2, Action: "swap", Protocol: "Uniswap", Amount: "10000 ETH", Description: "大额买卖触发短时价格偏移。"},
			{Step: 3, Action: "exploit", Protocol: "Victim Protocol", Amount: "", Description: "利用被污染的预言机价格触发错误清算或错误定价。"},
			{Step: 4, Action: "swap_back", Protocol: "Uniswap", Amount: "", Description: "平仓并尽量恢复市场价格。"},
			{Step: 5, Action: "repay", Protocol: "Aave", Amount: "10000 ETH + fee", Description: "归还闪电贷本息。"},
			{Step: 6, Action: "profit", Protocol: "", Amount: "500 ETH", Description: "保留操纵和结算错误带来的利润。"},
		},
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("price_manipulation_simulated", "", "", map[string]interface{}{
		"profit": attack.Profit,
		"steps":  len(attack.Steps),
	})
	s.updateState()
	return attack
}

// SimulateGovernanceAttack 演示治理劫持攻击。
func (s *FlashloanSimulator) SimulateGovernanceAttack() *FlashloanAttack {
	attack := &FlashloanAttack{
		Name:         "治理劫持",
		BorrowAmount: "1000000 GOV",
		Profit:       "获得协议控制权",
		Protocols:    []string{"Aave", "Victim DAO"},
		Steps: []*FlashloanStep{
			{Step: 1, Action: "borrow", Protocol: "Aave", Amount: "1000000 GOV", Description: "在单个区块内借入大量治理代币。"},
			{Step: 2, Action: "delegate", Protocol: "Victim DAO", Amount: "", Description: "把借入代币的投票权委托给攻击者。"},
			{Step: 3, Action: "propose", Protocol: "Victim DAO", Amount: "", Description: "提交恶意提案或修改提案参数。"},
			{Step: 4, Action: "vote", Protocol: "Victim DAO", Amount: "", Description: "利用临时获得的投票权快速通过提案。"},
			{Step: 5, Action: "execute", Protocol: "Victim DAO", Amount: "", Description: "执行恶意提案，转移控制权或协议资产。"},
			{Step: 6, Action: "repay", Protocol: "Aave", Amount: "1000000 GOV", Description: "归还闪电贷，攻击结果已经保留。"},
		},
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("governance_attack_simulated", "", "", map[string]interface{}{
		"borrow_amount": attack.BorrowAmount,
		"steps":         len(attack.Steps),
	})
	s.updateState()
	return attack
}

// SimulateLiquidationAttack 演示清算窗口利用。
func (s *FlashloanSimulator) SimulateLiquidationAttack() *FlashloanAttack {
	attack := &FlashloanAttack{
		Name:         "清算窗口利用",
		BorrowAmount: "5000 ETH",
		Profit:       "获得折价抵押品",
		Protocols:    []string{"Aave", "Compound", "Uniswap"},
		Steps: []*FlashloanStep{
			{Step: 1, Action: "borrow", Protocol: "Aave", Amount: "5000 ETH", Description: "借入大额资产，为操纵市场做准备。"},
			{Step: 2, Action: "manipulate", Protocol: "Uniswap", Amount: "", Description: "压低抵押品价格，使目标仓位进入可清算区间。"},
			{Step: 3, Action: "liquidate", Protocol: "Compound", Amount: "", Description: "抢先清算受害者仓位并获得奖励。"},
			{Step: 4, Action: "swap_back", Protocol: "Uniswap", Amount: "", Description: "卖出奖励资产并回收借款资产。"},
			{Step: 5, Action: "repay", Protocol: "Aave", Amount: "5000 ETH + fee", Description: "归还闪电贷并锁定利润。"},
		},
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("liquidation_attack_simulated", "", "", map[string]interface{}{
		"target": "Compound positions",
		"steps":  len(attack.Steps),
	})
	s.updateState()
	return attack
}

// ExplainFlashloanMechanism 返回闪电贷机制说明。
func (s *FlashloanSimulator) ExplainFlashloanMechanism() map[string]interface{} {
	return map[string]interface{}{
		"definition": "闪电贷允许在单笔交易内借入大额资金，并要求在同一笔交易结束前归还。",
		"how_it_works": []string{
			"借款人在交易开始时获得流动性。",
			"借款人在同一笔交易内执行套利、治理或清算逻辑。",
			"交易结束前必须归还本息。",
			"如果未能归还，整笔交易会回滚。",
		},
		"providers": map[string]string{
			"Aave":     "标准闪电贷接口",
			"Balancer": "基于池子流动性的闪电贷",
			"Uniswap":  "通过池子储备实现的 flash swap",
		},
	}
}

// ShowDefenses 返回防御方式。
func (s *FlashloanSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "TWAP 价格源", "description": "使用时间加权平均价，而不是单点现价。"},
		{"name": "外部预言机", "description": "通过链下预言机或多源价格做交叉校验。"},
		{"name": "治理延迟", "description": "关键提案通过后增加时间锁，避免瞬时借票。"},
		{"name": "操作限额", "description": "对单区块内的关键状态变更加阈值和速率限制。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *FlashloanSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "bZx", "date": "2020", "method": "价格操纵", "loss": "数百万美元"},
		{"name": "Beanstalk", "date": "2022", "method": "治理劫持", "loss": "数亿美元"},
	}
}

// updateState 更新前端主舞台所需状态。
func (s *FlashloanSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("victim_balance", 0)
		s.SetGlobalData("attacker_balance", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发闪电贷攻击，可以从价格操纵、治理劫持或清算窗口利用开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"economic",
			"idle",
			"等待执行闪电贷攻击场景。",
			"可以先选择一种攻击路径，观察借入资金、操纵过程和最终利润如何形成。",
			0,
			map[string]interface{}{
				"attack_count": 0,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	summary := fmt.Sprintf("%s 已完成，重点观察资金借入、市场扭曲和利润兑现的完整链路。", latest.Name)
	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", summary)
	s.SetGlobalData("steps", latest.Steps)
	s.SetGlobalData("max_depth", len(latest.Steps))
	s.SetGlobalData("victim_balance", parseAmountText(latest.BorrowAmount))
	s.SetGlobalData("attacker_balance", parseAmountText(latest.Profit))

	setAttackTeachingState(
		s.BaseSimulator,
		"economic",
		"settled",
		summary,
		"观察价格、仓位或治理状态在哪一步发生偏移，以及最终利润在何处兑现。",
		1.0,
		map[string]interface{}{
			"attack_name":      latest.Name,
			"borrow_amount":    latest.BorrowAmount,
			"profit":           latest.Profit,
			"protocol_count":   len(latest.Protocols),
			"step_count":       len(latest.Steps),
			"victim_balance":   parseAmountText(latest.BorrowAmount),
			"attacker_balance": parseAmountText(latest.Profit),
		},
	)
}

// ExecuteAction 执行动作。
func (s *FlashloanSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_price_manipulation":
		attack := s.SimulatePriceManipulation()
		return actionResultWithFeedback(
			"已执行闪电贷价格操纵演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已借入大额资金并触发价格操纵流程。",
				NextHint:    "重点观察受害协议在哪一步读取了被污染的价格，以及利润如何回收。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"borrow_amount": attack.BorrowAmount,
					"profit":        attack.Profit,
				},
			},
		), nil
	case "simulate_governance_attack":
		attack := s.SimulateGovernanceAttack()
		return actionResultWithFeedback(
			"已执行闪电贷治理攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入借票、提案、投票和执行的治理劫持流程。",
				NextHint:    "重点观察攻击者如何在一个短时间窗口内完成投票控制并保留结果。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"borrow_amount": attack.BorrowAmount,
					"profit":        attack.Profit,
				},
			},
		), nil
	case "simulate_liquidation_attack":
		attack := s.SimulateLiquidationAttack()
		return actionResultWithFeedback(
			"已执行闪电贷清算攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入价格压低、触发清算和回收奖励的完整流程。",
				NextHint:    "重点观察价格扭曲和目标仓位进入可清算区间的关键节点。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"borrow_amount": attack.BorrowAmount,
					"profit":        attack.Profit,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// FlashloanFactory 模拟器工厂。
type FlashloanFactory struct{}

// Create 创建模拟器。
func (f *FlashloanFactory) Create() engine.Simulator { return NewFlashloanSimulator() }

// GetDescription 返回描述。
func (f *FlashloanFactory) GetDescription() types.Description { return NewFlashloanSimulator().GetDescription() }

// NewFlashloanFactory 创建工厂。
func NewFlashloanFactory() *FlashloanFactory { return &FlashloanFactory{} }

var _ engine.SimulatorFactory = (*FlashloanFactory)(nil)
