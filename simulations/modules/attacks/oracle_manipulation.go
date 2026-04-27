package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// OracleAttack 表示一次预言机操纵攻击。
type OracleAttack struct {
	ID               string    `json:"id"`
	AttackType       string    `json:"attack_type"`
	OriginalPrice    *big.Int  `json:"original_price"`
	ManipulatedPrice *big.Int  `json:"manipulated_price"`
	FlashloanAmount  *big.Int  `json:"flashloan_amount"`
	Profit           *big.Int  `json:"profit"`
	Success          bool      `json:"success"`
	Timestamp        time.Time `json:"timestamp"`
}

// OracleManipulationSimulator 演示现货、TWAP 和多跳价格操纵。
type OracleManipulationSimulator struct {
	*base.BaseSimulator
	attacks      []*OracleAttack
	currentPrice *big.Int
}

func NewOracleManipulationSimulator() *OracleManipulationSimulator {
	return &OracleManipulationSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"oracle_manipulation",
			"预言机操纵攻击演示器",
			"演示现货价格操纵、TWAP 污染和多跳价格偏移导致的错误结算。",
			"attacks",
			types.ComponentAttack,
		),
		attacks:      make([]*OracleAttack, 0),
		currentPrice: big.NewInt(200000000000),
	}
}

func (s *OracleManipulationSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.attacks = make([]*OracleAttack, 0)
	s.currentPrice = big.NewInt(200000000000)
	s.updateState()
	return nil
}

func (s *OracleManipulationSimulator) SimulateSpotPriceManipulation(flashloanAmount int64) *OracleAttack {
	original := new(big.Int).Set(s.currentPrice)
	manipulated := new(big.Int).Mul(original, big.NewInt(65))
	manipulated.Div(manipulated, big.NewInt(100))

	attack := &OracleAttack{
		ID:               fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:       "spot_price_manipulation",
		OriginalPrice:    original,
		ManipulatedPrice: manipulated,
		FlashloanAmount:  new(big.Int).Mul(big.NewInt(flashloanAmount), big.NewInt(1e18)),
		Profit:           big.NewInt(8000),
		Success:          true,
		Timestamp:        time.Now(),
	}
	s.currentPrice = manipulated
	s.attacks = append(s.attacks, attack)
	s.EmitEvent("spot_price_manipulation", "", "", map[string]interface{}{
		"original_price":    formatBigInt(original),
		"manipulated_price": formatBigInt(manipulated),
		"summary":           "攻击者用闪电贷操纵现货池，导致预言机读取到错误价格。",
	})
	s.updateState()
	return attack
}

func (s *OracleManipulationSimulator) SimulateTWAPManipulation(manipulationBlocks int) *OracleAttack {
	original := new(big.Int).Set(s.currentPrice)
	manipulated := new(big.Int).Mul(original, big.NewInt(80))
	manipulated.Div(manipulated, big.NewInt(100))

	attack := &OracleAttack{
		ID:               fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:       "twap_manipulation",
		OriginalPrice:    original,
		ManipulatedPrice: manipulated,
		FlashloanAmount:  big.NewInt(int64(manipulationBlocks)),
		Profit:           big.NewInt(5000),
		Success:          true,
		Timestamp:        time.Now(),
	}
	s.currentPrice = manipulated
	s.attacks = append(s.attacks, attack)
	s.EmitEvent("twap_manipulation", "", "", map[string]interface{}{
		"blocks":  manipulationBlocks,
		"summary": "攻击者在多个区块内维持偏离价格，污染时间加权均价。",
	})
	s.updateState()
	return attack
}

func (s *OracleManipulationSimulator) SimulateMultiHopManipulation() *OracleAttack {
	original := new(big.Int).Set(s.currentPrice)
	manipulated := new(big.Int).Mul(original, big.NewInt(70))
	manipulated.Div(manipulated, big.NewInt(100))

	attack := &OracleAttack{
		ID:               fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:       "multi_hop_manipulation",
		OriginalPrice:    original,
		ManipulatedPrice: manipulated,
		FlashloanAmount:  big.NewInt(0),
		Profit:           big.NewInt(6000),
		Success:          true,
		Timestamp:        time.Now(),
	}
	s.currentPrice = manipulated
	s.attacks = append(s.attacks, attack)
	s.EmitEvent("multi_hop_manipulation", "", "", map[string]interface{}{
		"summary": "攻击者在多跳流动性路径中逐层传递价格偏移，最终影响目标协议。",
	})
	s.updateState()
	return attack
}

func (s *OracleManipulationSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "多源价格聚合", "description": "避免单一交易对价格被瞬时操纵后直接进入核心结算逻辑。"},
		{"name": "TWAP 与偏差保护", "description": "同时比较现价和时间加权均价，发现异常偏差时暂停关键操作。"},
		{"name": "流动性阈值检查", "description": "对低流动性池子的价格读数降权或直接禁用。"},
		{"name": "预言机冗余和回退", "description": "主预言机异常时切换到备用来源或保护模式。"},
	}
}

func (s *OracleManipulationSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Mango Markets", "date": "2022", "issue": "预言机与现货价格被联动操纵"},
		{"name": "bZx", "date": "2020", "issue": "闪电贷操纵链上价格触发错误结算"},
	}
}

func (s *OracleManipulationSimulator) updateState() {
	s.SetGlobalData("current_price", formatBigInt(s.currentPrice))
	s.SetGlobalData("attack_count", len(s.attacks))
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("victim_balance", 0)
		s.SetGlobalData("attacker_balance", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发预言机操纵，可以从现货操纵、TWAP 污染或多跳偏移开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"economic",
			"idle",
			"等待预言机操纵场景。",
			"可以先触发一种价格操纵，观察错误价格如何逐步传导到依赖协议。",
			0,
			map[string]interface{}{
				"current_price": formatBigInt(s.currentPrice),
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "shift_price_source",
			"caller":      "attacker",
			"function":    latest.AttackType,
			"target":      "oracle_source",
			"amount":      formatBigInt(latest.FlashloanAmount),
			"call_depth":  1,
			"description": "攻击者先控制价格源或成交路径，让预言机开始读取偏离真实市场的价格。",
		},
		{
			"step":        2,
			"action":      "oracle_reads_manipulated_price",
			"caller":      "oracle",
			"function":    "update_price",
			"target":      "dependent_protocol",
			"amount":      formatBigInt(latest.ManipulatedPrice),
			"call_depth":  2,
			"description": "协议在错误时间窗口内读取到被扭曲的价格，产生错误结算依据。",
		},
		{
			"step":        3,
			"action":      "extract_profit",
			"caller":      "attacker",
			"function":    "settle_trade",
			"target":      "dependent_protocol",
			"amount":      formatBigInt(latest.Profit),
			"call_depth":  3,
			"description": "攻击者利用错误价格完成套利、借贷或清算获利。",
		},
	}

	summary := fmt.Sprintf("原始价格 %s，被操纵后变为 %s。", formatBigInt(latest.OriginalPrice), formatBigInt(latest.ManipulatedPrice))
	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", summary)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))
	s.SetGlobalData("victim_balance", parseAmountText(formatBigInt(latest.OriginalPrice)))
	s.SetGlobalData("attacker_balance", parseAmountText(formatBigInt(latest.Profit)))

	setAttackTeachingState(
		s.BaseSimulator,
		"economic",
		"distorted",
		summary,
		"重点观察错误价格在哪一步进入依赖协议，以及利润是如何实现的。",
		1.0,
		map[string]interface{}{
			"attack_type":       latest.AttackType,
			"original_price":    formatBigInt(latest.OriginalPrice),
			"manipulated_price": formatBigInt(latest.ManipulatedPrice),
			"profit":            formatBigInt(latest.Profit),
		},
	)
}

func (s *OracleManipulationSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_spot_price_manipulation":
		flashloanAmount := actionInt64(params, "flashloan_amount", 5000)
		attack := s.SimulateSpotPriceManipulation(flashloanAmount)
		return actionResultWithFeedback(
			"已执行现货价格操纵演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入通过现货池直接扭曲价格的攻击流程。",
				NextHint:    "重点观察错误价格如何进入预言机，并导致协议做出错误结算。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"original_price":    formatBigInt(attack.OriginalPrice),
					"manipulated_price": formatBigInt(attack.ManipulatedPrice),
				},
			},
		), nil
	case "simulate_twap_manipulation":
		manipulationBlocks := actionInt(params, "manipulation_blocks", 6)
		attack := s.SimulateTWAPManipulation(manipulationBlocks)
		return actionResultWithFeedback(
			"已执行 TWAP 操纵演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入多区块维持价格偏移的 TWAP 污染流程。",
				NextHint:    "重点观察错误价格如何在时间窗口中累积并被依赖协议接受。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"original_price":    formatBigInt(attack.OriginalPrice),
					"manipulated_price": formatBigInt(attack.ManipulatedPrice),
				},
			},
		), nil
	case "simulate_multi_hop_manipulation":
		attack := s.SimulateMultiHopManipulation()
		return actionResultWithFeedback(
			"已执行多跳价格操纵演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入多跳流动性路径传递价格偏移的攻击流程。",
				NextHint:    "重点观察多跳路径如何放大价格偏移，并最终影响目标协议。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"original_price":    formatBigInt(attack.OriginalPrice),
					"manipulated_price": formatBigInt(attack.ManipulatedPrice),
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

type OracleManipulationFactory struct{}

func (f *OracleManipulationFactory) Create() engine.Simulator { return NewOracleManipulationSimulator() }

func (f *OracleManipulationFactory) GetDescription() types.Description {
	return NewOracleManipulationSimulator().GetDescription()
}

func NewOracleManipulationFactory() *OracleManipulationFactory { return &OracleManipulationFactory{} }

var _ engine.SimulatorFactory = (*OracleManipulationFactory)(nil)
