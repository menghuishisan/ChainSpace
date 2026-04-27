package attacks

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// SelfdestructSimulator 演示 selfdestruct 相关风险。
type SelfdestructSimulator struct {
	*base.BaseSimulator
	history []map[string]interface{}
}

// NewSelfdestructSimulator 创建 selfdestruct 模拟器。
func NewSelfdestructSimulator() *SelfdestructSimulator {
	return &SelfdestructSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"selfdestruct",
			"selfdestruct 攻击演示器",
			"演示强制转账、游戏逻辑破坏、实现合约销毁与 CREATE2 重新部署等 selfdestruct 风险。",
			"attacks",
			types.ComponentAttack,
		),
		history: make([]map[string]interface{}, 0),
	}
}

// Init 初始化模拟器。
func (s *SelfdestructSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.history = make([]map[string]interface{}, 0)
	s.updateState()
	return nil
}

// SimulateForceSendETH 演示强制转账。
func (s *SelfdestructSimulator) SimulateForceSendETH(amountETH int64) map[string]interface{} {
	data := map[string]interface{}{
		"attack_type": "force_send_eth",
		"amount_eth":  amountETH,
		"summary":     "攻击合约通过 selfdestruct 强制向目标地址注入 ETH，即使目标合约没有 receive 或 fallback 也无法阻止。",
		"flow": []string{
			"攻击者部署临时合约并充入 ETH。",
			"临时合约在目标地址不可拒绝的情况下执行 selfdestruct。",
			"目标合约余额被动增加，原有业务不变量被打破。",
		},
	}
	s.history = append(s.history, data)
	s.EmitEvent("selfdestruct_force_send", "", "", data)
	s.updateState()
	return data
}

// SimulateGameBreaking 演示基于余额判断的游戏被破坏。
func (s *SelfdestructSimulator) SimulateGameBreaking() map[string]interface{} {
	data := map[string]interface{}{
		"attack_type": "game_breaking",
		"summary":     "如果合约逻辑依赖 address(this).balance，自毁强制转账会让业务状态被意外破坏。",
		"flow": []string{
			"游戏逻辑使用合约余额判断输赢或阶段。",
			"攻击者构造自毁合约并把 ETH 强制打入目标。",
			"目标余额与业务记录脱节，导致规则判断失真。",
		},
	}
	s.history = append(s.history, data)
	s.EmitEvent("selfdestruct_game_breaking", "", "", data)
	s.updateState()
	return data
}

// SimulateProxyAttack 演示实现合约被销毁后代理失效。
func (s *SelfdestructSimulator) SimulateProxyAttack() map[string]interface{} {
	data := map[string]interface{}{
		"attack_type": "proxy_attack",
		"summary":     "攻击者若能控制实现合约并触发自毁，代理后续 delegatecall 将失去正常逻辑目标。",
		"flow": []string{
			"代理合约将调用转发到实现合约。",
			"攻击者接管实现合约后执行 selfdestruct。",
			"代理继续转发调用，但实现逻辑已不存在或被替换。",
		},
	}
	s.history = append(s.history, data)
	s.EmitEvent("selfdestruct_proxy_attack", "", "", data)
	s.updateState()
	return data
}

// SimulateCreate2Redeploy 演示销毁后在同地址重新部署。
func (s *SelfdestructSimulator) SimulateCreate2Redeploy() map[string]interface{} {
	data := map[string]interface{}{
		"attack_type": "create2_redeploy",
		"summary":     "攻击者可在销毁旧合约后，借助 CREATE2 在相同地址重新部署不同逻辑。",
		"flow": []string{
			"旧合约通过 selfdestruct 退出。",
			"攻击者利用相同 salt 和部署器重新计算地址。",
			"新逻辑在同地址落地，外部依赖方可能误以为合约仍然可信。",
		},
	}
	s.history = append(s.history, data)
	s.EmitEvent("selfdestruct_create2_redeploy", "", "", data)
	s.updateState()
	return data
}

// ShowDefenses 返回防御建议。
func (s *SelfdestructSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "避免依赖合约余额做核心判断", "description": "强制转账会使 address(this).balance 失真。"},
		{"name": "锁定实现合约初始化", "description": "防止实现合约被攻击者接管后执行危险逻辑。"},
		{"name": "升级前校验代码哈希", "description": "防止 CREATE2 重新部署后逻辑被悄悄替换。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *SelfdestructSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Parity 钱包事件", "impact": "共享库被销毁后，大量代理钱包失去可执行逻辑。"},
		{"name": "余额依赖型小游戏", "impact": "强制转账会破坏原本依赖余额守恒的业务条件。"},
	}
}

// updateState 同步状态。
func (s *SelfdestructSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.history))

	if len(s.history) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发 selfdestruct 场景，请选择强制转账、代理失效或 CREATE2 重部署进行观察。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待 selfdestruct 攻击场景。",
			"可以先触发一次强制转账或代理销毁，观察业务状态与合约余额如何脱节。",
			0,
			map[string]interface{}{
				"attack_count": len(s.history),
			},
		)
		return
	}

	latest := s.history[len(s.history)-1]
	flow, _ := latest["flow"].([]string)
	steps := make([]map[string]interface{}, 0, len(flow))
	for index, item := range flow {
		steps = append(steps, map[string]interface{}{
			"index":       index + 1,
			"title":       fmt.Sprintf("阶段 %d", index+1),
			"description": item,
		})
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest["summary"])
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"destroyed",
		fmt.Sprintf("%v", latest["summary"]),
		"重点观察 selfdestruct 在哪一步改变了余额、代理逻辑或部署上下文。",
		1.0,
		map[string]interface{}{
			"attack_type": latest["attack_type"],
			"step_count":  len(steps),
		},
	)
}

// ExecuteAction 执行 selfdestruct 相关动作。
func (s *SelfdestructSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_force_send_eth":
		amountETH := actionInt64(params, "amount_eth", 1)
		result := s.SimulateForceSendETH(amountETH)
		return actionResultWithFeedback(
			"已执行强制转账演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入通过 selfdestruct 强制向目标地址注入 ETH 的攻击流程。",
				NextHint:    "重点观察目标合约为何无法拒绝这笔 ETH，以及业务不变量如何被打破。",
				EffectScope: "execution",
				ResultState: result,
			},
		), nil
	case "simulate_game_breaking":
		result := s.SimulateGameBreaking()
		return actionResultWithFeedback(
			"已执行游戏逻辑破坏演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入通过强制转账破坏余额驱动规则的攻击流程。",
				NextHint:    "重点观察 address(this).balance 与业务记录为何出现偏差。",
				EffectScope: "execution",
				ResultState: result,
			},
		), nil
	case "simulate_proxy_attack":
		result := s.SimulateProxyAttack()
		return actionResultWithFeedback(
			"已执行代理实现合约销毁演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入实现合约被销毁、代理失去逻辑目标的攻击流程。",
				NextHint:    "重点观察 delegatecall 在失去实现逻辑后为何会使代理整体失效。",
				EffectScope: "execution",
				ResultState: result,
			},
		), nil
	case "simulate_create2_redeploy":
		result := s.SimulateCreate2Redeploy()
		return actionResultWithFeedback(
			"已执行 CREATE2 重新部署演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入销毁旧合约并在同地址重新部署不同逻辑的攻击流程。",
				NextHint:    "重点观察外部依赖方为什么可能误把新逻辑当成原合约。",
				EffectScope: "execution",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// SelfdestructFactory 创建 selfdestruct 模拟器。
type SelfdestructFactory struct{}

func (f *SelfdestructFactory) Create() engine.Simulator { return NewSelfdestructSimulator() }
func (f *SelfdestructFactory) GetDescription() types.Description { return NewSelfdestructSimulator().GetDescription() }
func NewSelfdestructFactory() *SelfdestructFactory { return &SelfdestructFactory{} }

var _ engine.SimulatorFactory = (*SelfdestructFactory)(nil)
