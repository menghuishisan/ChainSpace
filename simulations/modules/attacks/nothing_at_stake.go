package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// ValidatorVote 记录一名验证者在某条分叉上的投票。
type ValidatorVote struct {
	ValidatorID string    `json:"validator_id"`
	BlockHash   string    `json:"block_hash"`
	ForkID      string    `json:"fork_id"`
	Signature   string    `json:"signature"`
	Timestamp   time.Time `json:"timestamp"`
}

// DoubleVoteEvidence 记录双签证据。
type DoubleVoteEvidence struct {
	ValidatorID string         `json:"validator_id"`
	Vote1       *ValidatorVote `json:"vote1"`
	Vote2       *ValidatorVote `json:"vote2"`
	SlashAmount *big.Int       `json:"slash_amount"`
	Timestamp   time.Time      `json:"timestamp"`
}

// NothingAtStakeSimulator 演示 PoS 中的无利害关系问题。
type NothingAtStakeSimulator struct {
	*base.BaseSimulator
	validators      []string
	doubleVotes     []*DoubleVoteEvidence
	slashingEnabled bool
	slashAmount     *big.Int
}

// NewNothingAtStakeSimulator 创建无利害关系模拟器。
func NewNothingAtStakeSimulator() *NothingAtStakeSimulator {
	sim := &NothingAtStakeSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"nothing_at_stake",
			"无利害关系攻击演示器",
			"演示 PoS 验证者在没有有效惩罚时为何倾向于同时为多条分叉投票，以及 Slash 如何改变激励。",
			"attacks",
			types.ComponentAttack,
		),
		validators:  make([]string, 0),
		doubleVotes: make([]*DoubleVoteEvidence, 0),
	}

	sim.AddParam(types.Param{
		Key:         "slashing_enabled",
		Name:        "启用 Slash",
		Description: "控制双签后是否触发罚没。",
		Type:        types.ParamTypeBool,
		Default:     true,
	})
	sim.AddParam(types.Param{
		Key:         "validator_count",
		Name:        "验证者数量",
		Description: "参与分叉投票的验证者数量。",
		Type:        types.ParamTypeInt,
		Default:     10,
		Min:         3,
		Max:         100,
	})

	return sim
}

// Init 初始化验证者集合。
func (s *NothingAtStakeSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.slashingEnabled = true
	if v, ok := config.Params["slashing_enabled"]; ok {
		if b, ok := v.(bool); ok {
			s.slashingEnabled = b
		}
	}

	validatorCount := 10
	if v, ok := config.Params["validator_count"]; ok {
		if n, ok := v.(float64); ok {
			validatorCount = int(n)
		}
	}

	s.validators = make([]string, validatorCount)
	for i := 0; i < validatorCount; i++ {
		s.validators[i] = fmt.Sprintf("V%d", i+1)
	}
	s.doubleVotes = make([]*DoubleVoteEvidence, 0)
	s.slashAmount = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18))
	s.updateState()
	return nil
}

// SimulateDoubleVoting 演示一名验证者同时为两条分叉投票。
func (s *NothingAtStakeSimulator) SimulateDoubleVoting(validatorID string) map[string]interface{} {
	vote1 := &ValidatorVote{
		ValidatorID: validatorID,
		BlockHash:   "fork-a-head",
		ForkID:      "Fork-A",
		Signature:   fmt.Sprintf("sig_%s_a", validatorID),
		Timestamp:   time.Now(),
	}
	vote2 := &ValidatorVote{
		ValidatorID: validatorID,
		BlockHash:   "fork-b-head",
		ForkID:      "Fork-B",
		Signature:   fmt.Sprintf("sig_%s_b", validatorID),
		Timestamp:   time.Now(),
	}

	result := map[string]interface{}{
		"validator":      validatorID,
		"vote_fork_a":    vote1.BlockHash,
		"vote_fork_b":    vote2.BlockHash,
		"is_double_vote": true,
	}

	if s.slashingEnabled {
		evidence := &DoubleVoteEvidence{
			ValidatorID: validatorID,
			Vote1:       vote1,
			Vote2:       vote2,
			SlashAmount: s.slashAmount,
			Timestamp:   time.Now(),
		}
		s.doubleVotes = append(s.doubleVotes, evidence)
		result["slashed"] = true
		result["slash_amount"] = formatWei(s.slashAmount) + " ETH"
		result["consequence"] = "双签被链上证据捕获，验证者受到罚没。"
		s.EmitEvent("double_vote_slashed", "", "", result)
	} else {
		result["slashed"] = false
		result["consequence"] = "没有 Slash 时，验证者几乎总会选择两边都投。"
		s.EmitEvent("double_vote_unpunished", "", "", result)
	}

	s.updateState()
	return result
}

// SimulateRationalBehavior 演示不同惩罚机制下的理性行为差异。
func (s *NothingAtStakeSimulator) SimulateRationalBehavior() map[string]interface{} {
	doubleVoters := 0
	honestVoters := 0

	for _, validator := range s.validators {
		if s.slashingEnabled {
			honestVoters++
		} else {
			doubleVoters++
			s.SimulateDoubleVoting(validator)
		}
	}

	return map[string]interface{}{
		"slashing_enabled": s.slashingEnabled,
		"total_validators": len(s.validators),
		"double_voters":    doubleVoters,
		"honest_voters":    honestVoters,
		"analysis": map[string]string{
			"with_slashing":    "存在罚没时，理性验证者更倾向只为一条规范链投票。",
			"without_slashing": "没有罚没时，为所有分叉都投票几乎是更优策略。",
		},
	}
}

// ShowDefenses 返回防御措施。
func (s *NothingAtStakeSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Slash 双签", "description": "通过罚没权益，让双签行为从“几乎零成本”变成高代价。"},
		{"name": "快速终局", "description": "让分叉更快被确认，从而减少同时为多条链投票的空间。"},
		{"name": "举报奖励", "description": "鼓励网络参与者上报双签证据。"},
	}
}

// GetRealWorldCases 返回相关背景案例。
func (s *NothingAtStakeSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Ethereum PoS Slash 设计", "impact": "以罚没和证据机制限制验证者双签。"},
		{"name": "早期 PoS 论文讨论", "impact": "无利害关系问题是 PoS 安全设计中的核心挑战之一。"},
	}
}

// updateState 同步前端状态。
func (s *NothingAtStakeSimulator) updateState() {
	s.SetGlobalData("validator_count", len(s.validators))
	s.SetGlobalData("slashing_enabled", s.slashingEnabled)
	s.SetGlobalData("double_vote_count", len(s.doubleVotes))
	if len(s.doubleVotes) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发双签行为，可以观察 Slash 如何改变验证者在分叉中的选择。")
		setAttackTeachingState(
			s.BaseSimulator,
			"consensus",
			"idle",
			"等待无利害关系场景。",
			"可以先触发一次双签，再比较启用 Slash 与关闭 Slash 时理性行为的差异。",
			0,
			map[string]interface{}{
				"validator_count":   len(s.validators),
				"slashing_enabled":  s.slashingEnabled,
				"double_vote_count": len(s.doubleVotes),
			},
		)
		return
	}

	latest := s.doubleVotes[len(s.doubleVotes)-1]
	steps := []map[string]interface{}{
		{"step": 1, "action": "vote_on_fork_a", "caller": latest.ValidatorID, "function": "sign_vote", "target": latest.Vote1.ForkID, "amount": latest.Vote1.BlockHash, "call_depth": 1, "description": "验证者先在第一条分叉上签名。"},
		{"step": 2, "action": "vote_on_fork_b", "caller": latest.ValidatorID, "function": "sign_vote", "target": latest.Vote2.ForkID, "amount": latest.Vote2.BlockHash, "call_depth": 2, "description": "如果没有足够惩罚，同一验证者也会倾向于为另一条分叉签名。"},
		{"step": 3, "action": "detect_or_ignore_double_vote", "caller": "network", "function": "slash_or_ignore", "target": latest.ValidatorID, "amount": formatWei(latest.SlashAmount), "call_depth": 3, "description": map[bool]string{true: "网络发现双签并执行罚没，攻击成本显著提高。", false: "如果没有 Slash 机制，双签几乎没有成本。"}[s.slashingEnabled]},
	}

	summary := fmt.Sprintf("验证者 %s 在两条分叉上都签名。", latest.ValidatorID)
	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", summary)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"consensus",
		map[bool]string{true: "slashed", false: "double_voting"}[s.slashingEnabled],
		summary,
		"重点观察同一个验证者为何愿意两边都投票，以及 Slash 怎样改变这条激励链路。",
		1.0,
		map[string]interface{}{
			"validator_id":      latest.ValidatorID,
			"slashing_enabled":  s.slashingEnabled,
			"double_vote_count": len(s.doubleVotes),
			"slash_amount":      formatWei(latest.SlashAmount),
		},
	)
}

// ExecuteAction 执行动作。
func (s *NothingAtStakeSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_double_voting":
		validatorID := actionString(params, "validator_id", "V1")
		result := s.SimulateDoubleVoting(validatorID)
		return actionResultWithFeedback(
			"已执行双签演示。",
			map[string]interface{}{"result": result},
			&types.ActionFeedback{
				Summary:     "已进入同一验证者同时为两条分叉签名的攻击流程。",
				NextHint:    "重点观察 Slash 是否启用，以及双签在有无惩罚时的行为差异。",
				EffectScope: "consensus",
				ResultState: result,
			},
		), nil
	case "simulate_rational_behavior":
		result := s.SimulateRationalBehavior()
		return actionResultWithFeedback(
			"已执行理性行为分析演示。",
			map[string]interface{}{"result": result},
			&types.ActionFeedback{
				Summary:     "已对比启用 Slash 与关闭 Slash 时验证者的理性选择差异。",
				NextHint:    "重点观察双签数量与诚实投票数量如何随惩罚机制变化。",
				EffectScope: "consensus",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported nothing-at-stake action: %s", action)
	}
}

func formatWei(wei *big.Int) string {
	if wei == nil {
		return "0"
	}
	f := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt64(1e18))
	result, _ := f.Float64()
	return fmt.Sprintf("%.2f", result)
}

// NothingAtStakeFactory 创建模拟器。
type NothingAtStakeFactory struct{}

func (f *NothingAtStakeFactory) Create() engine.Simulator { return NewNothingAtStakeSimulator() }
func (f *NothingAtStakeFactory) GetDescription() types.Description {
	return NewNothingAtStakeSimulator().GetDescription()
}
func NewNothingAtStakeFactory() *NothingAtStakeFactory { return &NothingAtStakeFactory{} }

var _ engine.SimulatorFactory = (*NothingAtStakeFactory)(nil)
