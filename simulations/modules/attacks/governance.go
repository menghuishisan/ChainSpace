package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type Proposal struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Proposer     string    `json:"proposer"`
	ForVotes     *big.Int  `json:"for_votes"`
	AgainstVotes *big.Int  `json:"against_votes"`
	QuorumVotes  *big.Int  `json:"quorum_votes"`
	StartBlock   uint64    `json:"start_block"`
	EndBlock     uint64    `json:"end_block"`
	Executed     bool      `json:"executed"`
	Cancelled    bool      `json:"cancelled"`
	IsMalicious  bool      `json:"is_malicious"`
	Timestamp    time.Time `json:"timestamp"`
}

type GovernanceAttack struct {
	ID                string    `json:"id"`
	AttackType        string    `json:"attack_type"`
	Attacker          string    `json:"attacker"`
	VotingPower       *big.Int  `json:"voting_power"`
	VotingPowerSource string    `json:"voting_power_source"`
	TargetProposal    *Proposal `json:"target_proposal"`
	FlashloanAmount   *big.Int  `json:"flashloan_amount,omitempty"`
	Profit            *big.Int  `json:"profit"`
	Success           bool      `json:"success"`
	Timestamp         time.Time `json:"timestamp"`
}

type DAOState struct {
	TotalSupply     *big.Int            `json:"total_supply"`
	TreasuryBalance *big.Int            `json:"treasury_balance"`
	QuorumPercent   float64             `json:"quorum_percent"`
	VotingDelay     uint64              `json:"voting_delay"`
	VotingPeriod    uint64              `json:"voting_period"`
	TimelockDelay   uint64              `json:"timelock_delay"`
	Proposals       []*Proposal         `json:"proposals"`
	VotingPower     map[string]*big.Int `json:"voting_power"`
	HasSnapshot     bool                `json:"has_snapshot"`
}

type GovernanceSimulator struct {
	*base.BaseSimulator
	daoState *DAOState
	attacks  []*GovernanceAttack
}

func NewGovernanceSimulator() *GovernanceSimulator {
	sim := &GovernanceSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"governance",
			"治理攻击演示器",
			"演示闪电贷治理劫持、低法定人数攻击、治理贿赂和恶意提案执行。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*GovernanceAttack, 0),
	}

	sim.AddParam(types.Param{
		Key:         "treasury_balance",
		Name:        "国库余额",
		Description: "DAO 国库中可被治理提案影响的资产规模，单位为 ETH。",
		Type:        types.ParamTypeInt,
		Default:     1000000,
		Min:         1000,
		Max:         100000000,
	})
	sim.AddParam(types.Param{
		Key:         "has_snapshot",
		Name:        "启用快照",
		Description: "控制投票是否使用快照机制，以阻断闪电贷借票。",
		Type:        types.ParamTypeBool,
		Default:     false,
	})

	return sim
}

func (s *GovernanceSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	treasuryBalance := int64(1000000)
	if value, ok := config.Params["treasury_balance"]; ok {
		if typed, ok := value.(float64); ok {
			treasuryBalance = int64(typed)
		}
	}

	hasSnapshot := false
	if value, ok := config.Params["has_snapshot"]; ok {
		if typed, ok := value.(bool); ok {
			hasSnapshot = typed
		}
	}

	s.daoState = &DAOState{
		TotalSupply:     new(big.Int).Mul(big.NewInt(10000000), big.NewInt(1e18)),
		TreasuryBalance: new(big.Int).Mul(big.NewInt(treasuryBalance), big.NewInt(1e18)),
		QuorumPercent:   4.0,
		VotingDelay:     1,
		VotingPeriod:    17280,
		TimelockDelay:   172800,
		Proposals:       make([]*Proposal, 0),
		VotingPower:     make(map[string]*big.Int),
		HasSnapshot:     hasSnapshot,
	}
	s.daoState.VotingPower["whale1"] = new(big.Int).Mul(big.NewInt(500000), big.NewInt(1e18))
	s.daoState.VotingPower["whale2"] = new(big.Int).Mul(big.NewInt(300000), big.NewInt(1e18))
	s.daoState.VotingPower["attacker"] = new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))
	s.attacks = make([]*GovernanceAttack, 0)
	s.updateState()
	return nil
}

func (s *GovernanceSimulator) SimulateFlashloanGovernance(flashloanAmount int64) *GovernanceAttack {
	if s.daoState.HasSnapshot {
		s.EmitEvent("flashloan_blocked", "", "", map[string]interface{}{
			"reason": "投票使用快照，临时借来的治理代币不会立即生效。",
		})
		return nil
	}

	flashloan := new(big.Int).Mul(big.NewInt(flashloanAmount), big.NewInt(1e18))
	quorumNeeded := new(big.Int).Div(new(big.Int).Mul(s.daoState.TotalSupply, big.NewInt(int64(s.daoState.QuorumPercent))), big.NewInt(100))
	success := flashloan.Cmp(quorumNeeded) >= 0

	proposal := &Proposal{
		ID:           fmt.Sprintf("prop-%d", len(s.daoState.Proposals)+1),
		Title:        "Drain treasury",
		Description:  "恶意提案试图把国库资产转移给攻击者。",
		Proposer:     "0xAttacker",
		ForVotes:     flashloan,
		AgainstVotes: big.NewInt(0),
		QuorumVotes:  quorumNeeded,
		StartBlock:   12345678,
		EndBlock:     12345678 + s.daoState.VotingPeriod,
		Executed:     success,
		IsMalicious:  true,
		Timestamp:    time.Now(),
	}
	s.daoState.Proposals = append(s.daoState.Proposals, proposal)

	profit := big.NewInt(0)
	if success {
		profit = new(big.Int).Set(s.daoState.TreasuryBalance)
		s.daoState.TreasuryBalance = big.NewInt(0)
	}

	attack := &GovernanceAttack{
		ID:                fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:        "flashloan_governance",
		Attacker:          "0xAttacker",
		VotingPower:       flashloan,
		VotingPowerSource: "flashloan",
		TargetProposal:    proposal,
		FlashloanAmount:   flashloan,
		Profit:            profit,
		Success:           success,
		Timestamp:         time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("flashloan_governance_attack", "", "", map[string]interface{}{
		"success":       success,
		"flashloan":     formatBigInt(flashloan),
		"quorum_needed": formatBigInt(quorumNeeded),
		"profit":        formatBigInt(profit),
	})
	s.updateState()
	return attack
}

func (s *GovernanceSimulator) SimulateLowQuorumAttack() *GovernanceAttack {
	quorumNeeded := new(big.Int).Div(new(big.Int).Mul(s.daoState.TotalSupply, big.NewInt(1)), big.NewInt(100))
	proposal := &Proposal{
		ID:           fmt.Sprintf("prop-%d", len(s.daoState.Proposals)+1),
		Title:        "Pass proposal with weak quorum",
		Description:  "利用过低法定人数要求通过原本难以通过的提案。",
		Proposer:     "0xAttacker",
		ForVotes:     quorumNeeded,
		AgainstVotes: big.NewInt(0),
		QuorumVotes:  quorumNeeded,
		StartBlock:   12345678,
		EndBlock:     12345678 + s.daoState.VotingPeriod,
		Executed:     true,
		IsMalicious:  true,
		Timestamp:    time.Now(),
	}
	s.daoState.Proposals = append(s.daoState.Proposals, proposal)

	attack := &GovernanceAttack{
		ID:                fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:        "low_quorum_attack",
		Attacker:          "0xAttacker",
		VotingPower:       quorumNeeded,
		VotingPowerSource: "concentrated holdings",
		TargetProposal:    proposal,
		Profit:            big.NewInt(0),
		Success:           true,
		Timestamp:         time.Now(),
	}
	s.attacks = append(s.attacks, attack)
	s.EmitEvent("low_quorum_attack", "", "", map[string]interface{}{
		"quorum_used": formatBigInt(quorumNeeded),
		"summary":     "攻击者利用过低的法定人数要求通过恶意提案。",
	})
	s.updateState()
	return attack
}

func (s *GovernanceSimulator) SimulateBriberyAttack(bribeAmount int64) map[string]interface{} {
	data := map[string]interface{}{
		"attack_type":      "bribery_attack",
		"bribe_amount":     bribeAmount,
		"bribe_unit":       "USD per voting block",
		"summary":          "攻击者通过补贴投票者来影响提案结果。",
		"recommended_fix":  "使用投票锁定、延迟公开和信誉惩罚。",
		"target_proposals": []string{"treasury spend", "parameter change", "emergency upgrade"},
	}
	s.SetGlobalData("latest_bribery_attack", data)
	s.EmitEvent("bribery_attack", "", "", data)
	s.updateState()
	return data
}

func (s *GovernanceSimulator) SimulateMaliciousProposal() map[string]interface{} {
	data := map[string]interface{}{
		"attack_type":  "malicious_proposal",
		"proposal":     "upgrade implementation to attacker-controlled contract",
		"summary":      "攻击者在复杂提案中隐藏恶意执行逻辑。",
		"review_focus": []string{"upgrade target", "timelock payload", "delegatecall usage"},
	}
	s.SetGlobalData("latest_malicious_proposal", data)
	s.EmitEvent("malicious_proposal", "", "", data)
	s.updateState()
	return data
}

func (s *GovernanceSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "投票快照", "description": "在提案开始前固定投票权，防止闪电贷借票。"},
		{"name": "提案时间锁", "description": "提案通过后增加等待期，给社区留出响应窗口。"},
		{"name": "更高法定人数", "description": "避免少量票数就能通过关键治理决策。"},
		{"name": "提案审查与分级权限", "description": "对升级、转移国库等高风险操作进行额外审查。"},
	}
}

func (s *GovernanceSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Beanstalk", "date": "2022", "issue": "闪电贷治理劫持", "loss": "数亿美元"},
		{"name": "Curve wars", "date": "ongoing", "issue": "治理贿赂市场化", "loss": "governance capture pressure"},
	}
}

func (s *GovernanceSimulator) updateState() {
	s.SetGlobalData("treasury_balance", formatBigInt(s.daoState.TreasuryBalance))
	s.SetGlobalData("proposal_count", len(s.daoState.Proposals))
	s.SetGlobalData("attack_count", len(s.attacks))
	s.SetGlobalData("has_snapshot", s.daoState.HasSnapshot)
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("victim_balance", 0)
		s.SetGlobalData("attacker_balance", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发治理攻击，可以从借票、低法定人数、贿赂或恶意提案开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"economic",
			"idle",
			"等待治理攻击场景。",
			"可以先触发一次借票或恶意提案，观察投票权如何被集中并最终影响国库或升级结果。",
			0,
			map[string]interface{}{
				"treasury_balance": formatBigInt(s.daoState.TreasuryBalance),
				"proposal_count":   len(s.daoState.Proposals),
				"has_snapshot":     s.daoState.HasSnapshot,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "gain_voting_power",
			"caller":      latest.Attacker,
			"function":    latest.VotingPowerSource,
			"target":      "governance",
			"amount":      formatBigInt(latest.VotingPower),
			"call_depth":  1,
			"description": "攻击者先集中或临时借入投票权，为后续恶意提案建立多数优势。",
		},
		{
			"step":        2,
			"action":      latest.AttackType,
			"caller":      latest.Attacker,
			"function":    "submit_or_influence_proposal",
			"target":      latest.TargetProposal.Title,
			"amount":      formatBigInt(latest.TargetProposal.ForVotes),
			"call_depth":  2,
			"description": latest.TargetProposal.Description,
		},
		{
			"step":        3,
			"action":      "governance_effect",
			"caller":      "dao",
			"function":    "execute_proposal",
			"target":      "treasury",
			"amount":      formatBigInt(latest.Profit),
			"call_depth":  3,
			"description": map[bool]string{
				true:  "提案成功执行，协议治理结果被攻击者操纵。",
				false: "提案未成功执行，攻击没有真正控制治理结果。",
			}[latest.Success],
		},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest.TargetProposal.Description)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))
	s.SetGlobalData("victim_balance", parseAmountText(formatBigInt(s.daoState.TreasuryBalance)))
	s.SetGlobalData("attacker_balance", parseAmountText(formatBigInt(latest.Profit)))

	setAttackTeachingState(
		s.BaseSimulator,
		"economic",
		map[bool]string{true: "captured", false: "blocked"}[latest.Success],
		latest.TargetProposal.Description,
		"重点观察投票权是如何被集中起来的，以及恶意提案在哪一步影响了协议结果。",
		1.0,
		map[string]interface{}{
			"attack_type":      latest.AttackType,
			"voting_power":     formatBigInt(latest.VotingPower),
			"quorum_votes":     formatBigInt(latest.TargetProposal.QuorumVotes),
			"proposal_title":   latest.TargetProposal.Title,
			"treasury_balance": formatBigInt(s.daoState.TreasuryBalance),
			"profit":           formatBigInt(latest.Profit),
		},
	)
}

func (s *GovernanceSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_flashloan_governance":
		flashloanAmount := actionInt64(params, "flashloan_amount", 1000000)
		attack := s.SimulateFlashloanGovernance(flashloanAmount)
		return actionResultWithFeedback(
			"已执行闪电贷治理攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入借入治理权、推动恶意提案并试图执行的完整流程。",
				NextHint:    "重点观察投票权在哪一步被集中，以及快照机制是否能阻断攻击。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"has_snapshot": s.daoState.HasSnapshot,
					"proposal_id":  attack.TargetProposal.ID,
				},
			},
		), nil
	case "simulate_low_quorum_attack":
		attack := s.SimulateLowQuorumAttack()
		return actionResultWithFeedback(
			"已执行低法定人数攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入利用过低法定人数快速通过恶意提案的攻击流程。",
				NextHint:    "重点观察攻击者只需多小的投票权就能让提案通过。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"proposal_id": attack.TargetProposal.ID,
					"quorum":      formatBigInt(attack.TargetProposal.QuorumVotes),
				},
			},
		), nil
	case "simulate_bribery_attack":
		bribeAmount := actionInt64(params, "bribe_amount", 50000)
		result := s.SimulateBriberyAttack(bribeAmount)
		return actionResultWithFeedback(
			"已执行治理贿赂攻击演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入通过补贴投票者影响提案结果的攻击流程。",
				NextHint:    "重点观察贿赂成本与目标提案价值之间的关系。",
				EffectScope: "economic",
				ResultState: result,
			},
		), nil
	case "simulate_malicious_proposal":
		result := s.SimulateMaliciousProposal()
		return actionResultWithFeedback(
			"已执行恶意提案演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入在复杂提案中隐藏恶意执行逻辑的攻击流程。",
				NextHint:    "重点观察升级目标、时间锁载荷和 delegatecall 使用是否存在异常。",
				EffectScope: "economic",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

type GovernanceFactory struct{}

func (f *GovernanceFactory) Create() engine.Simulator { return NewGovernanceSimulator() }

func (f *GovernanceFactory) GetDescription() types.Description { return NewGovernanceSimulator().GetDescription() }

func NewGovernanceFactory() *GovernanceFactory { return &GovernanceFactory{} }

var _ engine.SimulatorFactory = (*GovernanceFactory)(nil)
