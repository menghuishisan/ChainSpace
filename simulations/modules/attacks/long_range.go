package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// LongRangeAttackType 表示长程攻击类型。
type LongRangeAttackType string

const (
	LRAttackPosterior LongRangeAttackType = "posterior"
)

// LongRangeAttack 描述一次长程攻击演示。
type LongRangeAttack struct {
	ID             string              `json:"id"`
	Type           LongRangeAttackType `json:"type"`
	StartBlock     uint64              `json:"start_block"`
	CurrentBlock   uint64              `json:"current_block"`
	AltChainLength uint64              `json:"alt_chain_length"`
	KeysAcquired   int                 `json:"keys_acquired"`
	Success        bool                `json:"success"`
	Timestamp      time.Time           `json:"timestamp"`
}

// LongRangeSimulator 演示 PoS 中的长程攻击。
type LongRangeSimulator struct {
	*base.BaseSimulator
	attacks      []*LongRangeAttack
	currentEpoch uint64
}

// NewLongRangeSimulator 创建长程攻击模拟器。
func NewLongRangeSimulator() *LongRangeSimulator {
	sim := &LongRangeSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"long_range",
			"长程攻击演示器",
			"演示攻击者收集历史验证者密钥后，从较早高度重构一条看似合法的替代链历史。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*LongRangeAttack, 0),
	}

	sim.AddParam(types.Param{
		Key:         "chain_age_years",
		Name:        "链历史长度",
		Description: "用于估算可回滚跨度的链运行年限。",
		Type:        types.ParamTypeInt,
		Default:     3,
		Min:         1,
		Max:         10,
	})

	return sim
}

// Init 初始化链历史尺度。
func (s *LongRangeSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	chainAge := 3
	if v, ok := config.Params["chain_age_years"]; ok {
		if n, ok := v.(float64); ok {
			chainAge = int(n)
		}
	}

	s.currentEpoch = uint64(chainAge) * 365
	s.attacks = make([]*LongRangeAttack, 0)
	s.updateState()
	return nil
}

// SimulatePosteriorAttack 演示收集历史密钥后的长程攻击。
func (s *LongRangeSimulator) SimulatePosteriorAttack(startEpoch uint64, keysAcquired int) *LongRangeAttack {
	success := keysAcquired >= 3
	altChainLength := uint64(0)
	if success && s.currentEpoch > startEpoch {
		altChainLength = (s.currentEpoch - startEpoch) * 7200
	}

	attack := &LongRangeAttack{
		ID:             fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:           LRAttackPosterior,
		StartBlock:     startEpoch * 7200,
		CurrentBlock:   s.currentEpoch * 7200,
		AltChainLength: altChainLength,
		KeysAcquired:   keysAcquired,
		Success:        success,
		Timestamp:      time.Now(),
	}
	s.attacks = append(s.attacks, attack)

	s.EmitEvent("posterior_attack", "", "", map[string]interface{}{
		"start_epoch":      startEpoch,
		"current_epoch":    s.currentEpoch,
		"keys_acquired":    keysAcquired,
		"success":          success,
		"epochs_rewritten": s.currentEpoch - startEpoch,
		"attack_flow": []string{
			"1. 攻击者收集历史时期已经退出网络的验证者私钥。",
			"2. 攻击者从更早的 epoch 开始重写替代链历史。",
			"3. 如果新节点缺少可信检查点，就可能无法区分伪造历史。",
			map[bool]string{true: "4. 伪造链在缺少外部锚点时可能被误接受。", false: "4. 私钥控制不足，无法让替代链获得足够合法性。"}[success],
		},
	})

	s.updateState()
	return attack
}

// SimulateNewNodeSync 演示新节点同步时是否会被伪造历史误导。
func (s *LongRangeSimulator) SimulateNewNodeSync(hasCheckpoint bool) map[string]interface{} {
	if len(s.attacks) == 0 {
		return map[string]interface{}{"result": "当前没有长程攻击历史可供新节点比较。"}
	}

	lastAttack := s.attacks[len(s.attacks)-1]
	if hasCheckpoint {
		return map[string]interface{}{
			"result":                "新节点依靠可信检查点拒绝了伪造历史。",
			"attack_chain_rejected": true,
			"checkpoint_epoch":      s.currentEpoch - 100,
			"related_attack":        lastAttack.ID,
		}
	}

	return map[string]interface{}{
		"result":            "新节点缺少可信检查点，存在被替代历史误导的风险。",
		"related_attack":    lastAttack.ID,
		"weak_subjectivity": true,
	}
}

// ShowDefenses 返回防御方案。
func (s *LongRangeSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "可信检查点", "description": "让新节点同步时拥有可靠的历史锚点。"},
		{"name": "密钥轮换与失效机制", "description": "降低历史验证密钥长期泄露后的利用价值。"},
		{"name": "弱主观性同步", "description": "要求新节点从可信来源获取近期终局状态。"},
	}
}

// GetRealWorldCases 返回背景案例。
func (s *LongRangeSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "PoS 长程攻击理论讨论", "impact": "说明与 PoW 相比，PoS 更需要额外的历史锚点与弱主观性假设。"},
	}
}

// updateState 同步状态。
func (s *LongRangeSimulator) updateState() {
	s.SetGlobalData("current_epoch", s.currentEpoch)
	s.SetGlobalData("attack_count", len(s.attacks))
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发长程攻击，可以观察攻击者如何从更早历史重构替代链。")
		setAttackTeachingState(
			s.BaseSimulator,
			"consensus",
			"idle",
			"等待长程攻击场景。",
			"可以先触发一次历史重写攻击，再观察新节点同步时是否被伪造历史误导。",
			0,
			map[string]interface{}{
				"current_epoch": s.currentEpoch,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{"step": 1, "action": "collect_old_keys", "caller": "attacker", "function": "acquire_retired_keys", "target": "historical_validators", "amount": fmt.Sprintf("%d", latest.KeysAcquired), "call_depth": 1, "description": "攻击者先收集历史时期已经退出网络的验证者私钥。"},
		{"step": 2, "action": "rewrite_old_history", "caller": "attacker", "function": "forge_alt_chain", "target": "alt_chain", "amount": fmt.Sprintf("%d", latest.AltChainLength), "call_depth": 2, "description": "攻击者从更早高度开始重写一条替代历史。"},
		{"step": 3, "action": "mislead_new_node", "caller": "new_node", "function": "sync_history", "target": "alt_chain", "amount": fmt.Sprintf("%d", latest.CurrentBlock), "call_depth": 3, "description": map[bool]string{true: "如果新节点缺少可信检查点，就可能接受这条伪造历史。", false: "由于控制能力不足或存在可信锚点，替代历史未能被接受。"}[latest.Success]},
	}

	summary := fmt.Sprintf("攻击从高度 %d 开始重写，当前链高度 %d。", latest.StartBlock, latest.CurrentBlock)
	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", summary)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"consensus",
		map[bool]string{true: "rewriting", false: "rejected"}[latest.Success],
		summary,
		"重点观察可信检查点是否存在，以及新节点在同步时会不会被替代历史误导。",
		1.0,
		map[string]interface{}{
			"start_block":      latest.StartBlock,
			"current_block":    latest.CurrentBlock,
			"alt_chain_length": latest.AltChainLength,
			"keys_acquired":    latest.KeysAcquired,
			"success":          latest.Success,
		},
	)
}

// ExecuteAction 执行动作。
func (s *LongRangeSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_posterior_attack":
		startEpoch := uint64(actionInt(params, "start_epoch", 200))
		keysAcquired := actionInt(params, "keys_acquired", 3)
		attack := s.SimulatePosteriorAttack(startEpoch, keysAcquired)
		return actionResultWithFeedback(
			"已执行长程攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入收集历史密钥、重写旧历史和误导新节点的长程攻击流程。",
				NextHint:    "重点观察攻击者从哪个高度开始重写，以及新节点为何可能接受伪造历史。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{
					"start_block":      attack.StartBlock,
					"current_block":    attack.CurrentBlock,
					"alt_chain_length": attack.AltChainLength,
				},
			},
		), nil
	case "simulate_new_node_sync":
		hasCheckpoint := actionBool(params, "has_checkpoint", true)
		result := s.SimulateNewNodeSync(hasCheckpoint)
		return actionResultWithFeedback(
			"已执行新节点同步演示。",
			map[string]interface{}{"result": result},
			&types.ActionFeedback{
				Summary:     "已模拟新节点面对替代历史时的同步判断。",
				NextHint:    "重点观察可信检查点是否存在，以及弱主观性假设如何改变同步结果。",
				EffectScope: "consensus",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported long range action: %s", action)
	}
}

// LongRangeFactory 创建模拟器。
type LongRangeFactory struct{}

func (f *LongRangeFactory) Create() engine.Simulator { return NewLongRangeSimulator() }
func (f *LongRangeFactory) GetDescription() types.Description { return NewLongRangeSimulator().GetDescription() }
func NewLongRangeFactory() *LongRangeFactory { return &LongRangeFactory{} }

var _ engine.SimulatorFactory = (*LongRangeFactory)(nil)
