package attacks

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// SelfishMiningSimulator 演示自私挖矿攻击。
type SelfishMiningSimulator struct {
	*base.BaseSimulator
}

// NewSelfishMiningSimulator 创建自私挖矿模拟器。
func NewSelfishMiningSimulator() *SelfishMiningSimulator {
	return &SelfishMiningSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"selfish_mining",
			"自私挖矿攻击演示器",
			"演示攻击者隐藏私有链、等待优势扩大后再公开，以提高相对收益的过程。",
			"attacks",
			types.ComponentAttack,
		),
	}
}

// Init 初始化模拟器。
func (s *SelfishMiningSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.SetGlobalData("steps", []interface{}{})
	s.SetGlobalData("max_depth", 0)
	s.SetGlobalData("attack_count", 0)
	s.SetGlobalData("attack_summary", "当前尚未触发自私挖矿攻击，可以观察隐匿区块、链竞争和公开私链的完整过程。")
	setAttackTeachingState(
		s.BaseSimulator,
		"consensus",
		"idle",
		"等待自私挖矿场景。",
		"可以先模拟数轮私链积累，观察攻击者何时选择公开私链。",
		0,
		map[string]interface{}{
			"attack_count": 0,
		},
	)
	return nil
}

// SimulateRounds 演示多轮自私挖矿。
func (s *SelfishMiningSimulator) SimulateRounds(rounds int) map[string]interface{} {
	data := map[string]interface{}{
		"rounds": rounds,
		"phases": []string{
			"攻击者私下保留新区块，不立即广播。",
			"诚实矿工继续在旧公开链上挖矿。",
			"当攻击者私链领先时，再择机公开形成竞争。",
			"部分诚实矿工会切换到更长或更新的链头。",
		},
		"summary": "攻击者通过策略性公开区块，使自己的相对收益高于其真实算力占比。",
	}
	s.SetGlobalData("latest_selfish_mining", data)
	s.SetGlobalData("attack_count", 1)
	s.SetGlobalData("steps", []map[string]interface{}{
		{
			"step":        1,
			"action":      "withhold_block",
			"caller":      "attacker",
			"function":    "private_mine",
			"target":      "private_chain",
			"amount":      fmt.Sprintf("%d", rounds),
			"call_depth":  1,
			"description": "攻击者先隐藏新区块，不立即广播，尝试建立私有领先优势。",
		},
		{
			"step":        2,
			"action":      "race_public_chain",
			"caller":      "honest_miners",
			"function":    "mine_public",
			"target":      "public_chain",
			"amount":      fmt.Sprintf("%d", rounds),
			"call_depth":  2,
			"description": "诚实矿工继续在公开链上挖矿，而攻击者观察是否值得继续隐藏私链。",
		},
		{
			"step":        3,
			"action":      "publish_private_chain",
			"caller":      "attacker",
			"function":    "override_choice",
			"target":      "public_chain",
			"amount":      "1",
			"call_depth":  3,
			"description": "一旦私链时机成熟，攻击者公开私链，诱导部分诚实矿工切换到攻击者路径。",
		},
	})
	s.SetGlobalData("max_depth", 3)
	s.SetGlobalData("attack_summary", "攻击者通过策略性公开区块，使相对收益高于真实算力占比。")
	setAttackTeachingState(
		s.BaseSimulator,
		"consensus",
		"contested",
		"攻击者正在通过隐藏区块和择机公开来争夺主导链。",
		"重点观察私链何时公开，以及诚实矿工何时开始切换链头。",
		1.0,
		map[string]interface{}{
			"rounds":     rounds,
			"step_count": 3,
		},
	)
	s.EmitEvent("selfish_mining_rounds", "", "", data)
	return data
}

// ShowDefenses 返回防御建议。
func (s *SelfishMiningSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "改进分叉选择规则", "description": "降低攻击者隐藏链被后续追上的概率优势。"},
		{"name": "提升区块传播效率", "description": "缩短诚实矿工获得新区块的时间窗口。"},
		{"name": "监测异常孤块率", "description": "及时识别持续性的私链发布行为。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *SelfishMiningSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Eyal & Sirer 论文", "impact": "指出算力低于 50% 的矿工也可能在特定条件下获得超额收益。"},
	}
}

// ExecuteAction 执行自私挖矿动作。
func (s *SelfishMiningSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_rounds":
		rounds := actionInt(params, "rounds", 12)
		data := s.SimulateRounds(rounds)
		return actionResultWithFeedback(
			"已执行自私挖矿多轮模拟。",
			data,
			&types.ActionFeedback{
				Summary:     "已进入隐藏新区块、竞争公开链和择机公开私链的完整过程。",
				NextHint:    "重点观察攻击者何时公开私链，以及诚实矿工是否切换到攻击者路径。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{
					"rounds": rounds,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// SelfishMiningFactory 创建自私挖矿模拟器。
type SelfishMiningFactory struct{}

func (f *SelfishMiningFactory) Create() engine.Simulator { return NewSelfishMiningSimulator() }
func (f *SelfishMiningFactory) GetDescription() types.Description { return NewSelfishMiningSimulator().GetDescription() }
func NewSelfishMiningFactory() *SelfishMiningFactory { return &SelfishMiningFactory{} }

var _ engine.SimulatorFactory = (*SelfishMiningFactory)(nil)
