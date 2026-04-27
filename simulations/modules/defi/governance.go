package defi

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// DeFi治理演示器
// =============================================================================

// ProposalState 提案状态
type ProposalState string

const (
	ProposalPending   ProposalState = "pending"
	ProposalActive    ProposalState = "active"
	ProposalSucceeded ProposalState = "succeeded"
	ProposalDefeated  ProposalState = "defeated"
	ProposalQueued    ProposalState = "queued"
	ProposalExecuted  ProposalState = "executed"
	ProposalCanceled  ProposalState = "canceled"
	ProposalExpired   ProposalState = "expired"
)

// Proposal 提案
type Proposal struct {
	ID           uint64              `json:"id"`
	Proposer     string              `json:"proposer"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Actions      []*GovAction        `json:"actions"`
	ForVotes     *big.Int            `json:"for_votes"`
	AgainstVotes *big.Int            `json:"against_votes"`
	AbstainVotes *big.Int            `json:"abstain_votes"`
	Voters       map[string]VoteInfo `json:"voters"`
	State        ProposalState       `json:"state"`
	StartBlock   uint64              `json:"start_block"`
	EndBlock     uint64              `json:"end_block"`
	ETA          time.Time           `json:"eta"` // 执行时间
	CreatedAt    time.Time           `json:"created_at"`
}

// GovAction 治理动作
type GovAction struct {
	Target   string   `json:"target"`   // 目标合约
	Value    *big.Int `json:"value"`    // ETH值
	Function string   `json:"function"` // 函数签名
	Data     []byte   `json:"data"`     // 调用数据
}

// VoteInfo 投票信息
type VoteInfo struct {
	Voter     string    `json:"voter"`
	Support   uint8     `json:"support"` // 0=against, 1=for, 2=abstain
	Weight    *big.Int  `json:"weight"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// GovernanceConfig 治理配置
type GovernanceConfig struct {
	VotingDelay       uint64   `json:"voting_delay"`       // 投票开始前等待区块数
	VotingPeriod      uint64   `json:"voting_period"`      // 投票持续区块数
	ProposalThreshold *big.Int `json:"proposal_threshold"` // 提案所需最少代币
	QuorumVotes       *big.Int `json:"quorum_votes"`       // 法定人数票数
	TimelockDelay     uint64   `json:"timelock_delay"`     // 时间锁延迟(秒)
}

// GovernanceSimulator DeFi治理演示器
// 演示DAO治理的核心机制:
//
// 1. 提案创建
//   - 需持有一定数量治理代币
//   - 提案包含链上执行的动作
//
// 2. 投票
//   - 代币持有者按持仓权重投票
//   - 支持/反对/弃权
//
// 3. 时间锁
//   - 通过的提案需等待时间锁
//   - 给用户退出时间
//
// 4. 执行
//   - 时间锁后执行提案动作
//
// 参考: Compound Governor, OpenZeppelin Governor
type GovernanceSimulator struct {
	*base.BaseSimulator
	proposals      map[uint64]*Proposal
	nextProposalID uint64
	config         *GovernanceConfig
	tokenBalances  map[string]*big.Int // 治理代币余额
	delegations    map[string]string   // 委托关系
	currentBlock   uint64
}

// NewGovernanceSimulator 创建治理演示器
func NewGovernanceSimulator() *GovernanceSimulator {
	sim := &GovernanceSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"governance",
			"DeFi治理演示器",
			"演示DAO治理的提案、投票、时间锁、执行等完整流程",
			"defi",
			types.ComponentDeFi,
		),
		proposals:     make(map[uint64]*Proposal),
		tokenBalances: make(map[string]*big.Int),
		delegations:   make(map[string]string),
	}

	sim.AddParam(types.Param{
		Key:         "voting_period",
		Name:        "投票周期",
		Description: "投票持续的区块数",
		Type:        types.ParamTypeInt,
		Default:     17280, // 约3天 (假设12秒/块)
		Min:         100,
		Max:         100000,
	})

	sim.AddParam(types.Param{
		Key:         "timelock_delay",
		Name:        "时间锁延迟",
		Description: "提案通过后等待执行的秒数",
		Type:        types.ParamTypeInt,
		Default:     172800, // 2天
		Min:         3600,
		Max:         604800,
	})

	return sim
}

// Init 初始化
func (s *GovernanceSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	votingPeriod := uint64(17280)
	timelockDelay := uint64(172800)

	if v, ok := config.Params["voting_period"]; ok {
		if n, ok := v.(float64); ok {
			votingPeriod = uint64(n)
		}
	}
	if v, ok := config.Params["timelock_delay"]; ok {
		if n, ok := v.(float64); ok {
			timelockDelay = uint64(n)
		}
	}

	proposalThreshold := new(big.Int)
	proposalThreshold.SetString("100000000000000000000000", 10) // 100000 * 1e18
	quorumVotes := new(big.Int)
	quorumVotes.SetString("4000000000000000000000000", 10) // 4000000 * 1e18

	s.config = &GovernanceConfig{
		VotingDelay:       1,
		VotingPeriod:      votingPeriod,
		ProposalThreshold: proposalThreshold,
		QuorumVotes:       quorumVotes,
		TimelockDelay:     timelockDelay,
	}

	s.proposals = make(map[uint64]*Proposal)
	s.nextProposalID = 1
	s.currentBlock = 1000000

	// 初始化测试用户余额
	s.tokenBalances = make(map[string]*big.Int)
	whale1 := new(big.Int)
	whale1.SetString("5000000000000000000000000", 10) // 5000000 * 1e18
	whale2 := new(big.Int)
	whale2.SetString("3000000000000000000000000", 10)
	whale3 := new(big.Int)
	whale3.SetString("2000000000000000000000000", 10)
	user1 := new(big.Int)
	user1.SetString("500000000000000000000000", 10)
	user2 := new(big.Int)
	user2.SetString("200000000000000000000000", 10)
	s.tokenBalances["whale1"] = whale1
	s.tokenBalances["whale2"] = whale2
	s.tokenBalances["whale3"] = whale3
	s.tokenBalances["user1"] = user1
	s.tokenBalances["user2"] = user2

	s.updateState()
	return nil
}

// =============================================================================
// 治理操作
// =============================================================================

// CreateProposal 创建提案
func (s *GovernanceSimulator) CreateProposal(proposer, title, description string, actions []*GovAction) (*Proposal, error) {
	// 检查提案门槛
	balance := s.getVotingPower(proposer)
	if balance.Cmp(s.config.ProposalThreshold) < 0 {
		return nil, fmt.Errorf("余额%s不足提案门槛%s", balance.String(), s.config.ProposalThreshold.String())
	}

	proposal := &Proposal{
		ID:           s.nextProposalID,
		Proposer:     proposer,
		Title:        title,
		Description:  description,
		Actions:      actions,
		ForVotes:     big.NewInt(0),
		AgainstVotes: big.NewInt(0),
		AbstainVotes: big.NewInt(0),
		Voters:       make(map[string]VoteInfo),
		State:        ProposalPending,
		StartBlock:   s.currentBlock + s.config.VotingDelay,
		EndBlock:     s.currentBlock + s.config.VotingDelay + s.config.VotingPeriod,
		CreatedAt:    time.Now(),
	}

	s.proposals[proposal.ID] = proposal
	s.nextProposalID++

	s.EmitEvent("proposal_created", "", "", map[string]interface{}{
		"proposal_id": proposal.ID,
		"proposer":    proposer,
		"title":       title,
		"start_block": proposal.StartBlock,
		"end_block":   proposal.EndBlock,
	})

	s.updateState()
	return proposal, nil
}

// CastVote 投票
func (s *GovernanceSimulator) CastVote(proposalID uint64, voter string, support uint8, reason string) error {
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return fmt.Errorf("提案不存在")
	}

	// 检查投票期
	if s.currentBlock < proposal.StartBlock {
		return fmt.Errorf("投票尚未开始")
	}
	if s.currentBlock > proposal.EndBlock {
		return fmt.Errorf("投票已结束")
	}

	// 检查是否已投票
	if _, voted := proposal.Voters[voter]; voted {
		return fmt.Errorf("已经投过票")
	}

	// 获取投票权重
	weight := s.getVotingPower(voter)
	if weight.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("没有投票权")
	}

	// 记录投票
	proposal.Voters[voter] = VoteInfo{
		Voter:     voter,
		Support:   support,
		Weight:    weight,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	switch support {
	case 0:
		proposal.AgainstVotes.Add(proposal.AgainstVotes, weight)
	case 1:
		proposal.ForVotes.Add(proposal.ForVotes, weight)
	case 2:
		proposal.AbstainVotes.Add(proposal.AbstainVotes, weight)
	}

	s.EmitEvent("vote_cast", "", "", map[string]interface{}{
		"proposal_id": proposalID,
		"voter":       voter,
		"support":     support,
		"weight":      weight.String(),
		"reason":      reason,
	})

	s.updateState()
	return nil
}

// AdvanceBlocks 推进区块
func (s *GovernanceSimulator) AdvanceBlocks(blocks uint64) {
	s.currentBlock += blocks

	// 更新提案状态
	for _, proposal := range s.proposals {
		s.updateProposalState(proposal)
	}

	s.updateState()
}

// updateProposalState 更新提案状态
func (s *GovernanceSimulator) updateProposalState(proposal *Proposal) {
	if proposal.State == ProposalExecuted || proposal.State == ProposalCanceled {
		return
	}

	if s.currentBlock < proposal.StartBlock {
		proposal.State = ProposalPending
	} else if s.currentBlock <= proposal.EndBlock {
		proposal.State = ProposalActive
	} else {
		// 投票结束，计算结果
		totalVotes := new(big.Int).Add(proposal.ForVotes, proposal.AgainstVotes)
		totalVotes.Add(totalVotes, proposal.AbstainVotes)

		if totalVotes.Cmp(s.config.QuorumVotes) < 0 {
			proposal.State = ProposalDefeated // 未达法定人数
		} else if proposal.ForVotes.Cmp(proposal.AgainstVotes) > 0 {
			if proposal.State != ProposalQueued && proposal.State != ProposalExecuted {
				proposal.State = ProposalSucceeded
			}
		} else {
			proposal.State = ProposalDefeated
		}
	}
}

// QueueProposal 将提案加入时间锁队列
func (s *GovernanceSimulator) QueueProposal(proposalID uint64) error {
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return fmt.Errorf("提案不存在")
	}

	if proposal.State != ProposalSucceeded {
		return fmt.Errorf("提案状态不是Succeeded")
	}

	proposal.State = ProposalQueued
	proposal.ETA = time.Now().Add(time.Duration(s.config.TimelockDelay) * time.Second)

	s.EmitEvent("proposal_queued", "", "", map[string]interface{}{
		"proposal_id": proposalID,
		"eta":         proposal.ETA,
	})

	s.updateState()
	return nil
}

// ExecuteProposal 执行提案
func (s *GovernanceSimulator) ExecuteProposal(proposalID uint64) error {
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return fmt.Errorf("提案不存在")
	}

	if proposal.State != ProposalQueued {
		return fmt.Errorf("提案未在队列中")
	}

	if time.Now().Before(proposal.ETA) {
		return fmt.Errorf("时间锁未到期")
	}

	// 模拟执行动作
	for i, action := range proposal.Actions {
		s.EmitEvent("action_executed", "", "", map[string]interface{}{
			"proposal_id":  proposalID,
			"action_index": i,
			"target":       action.Target,
			"function":     action.Function,
		})
	}

	proposal.State = ProposalExecuted

	s.EmitEvent("proposal_executed", "", "", map[string]interface{}{
		"proposal_id": proposalID,
	})

	s.updateState()
	return nil
}

// =============================================================================
// 委托
// =============================================================================

// Delegate 委托投票权
func (s *GovernanceSimulator) Delegate(delegator, delegatee string) error {
	s.delegations[delegator] = delegatee

	s.EmitEvent("delegation", "", "", map[string]interface{}{
		"delegator": delegator,
		"delegatee": delegatee,
	})

	s.updateState()
	return nil
}

// getVotingPower 获取投票权
func (s *GovernanceSimulator) getVotingPower(account string) *big.Int {
	power := big.NewInt(0)

	// 自己的余额
	if balance, ok := s.tokenBalances[account]; ok {
		// 如果没有委托给别人，使用自己的余额
		if s.delegations[account] == "" || s.delegations[account] == account {
			power.Add(power, balance)
		}
	}

	// 收到的委托
	for delegator, delegatee := range s.delegations {
		if delegatee == account && delegator != account {
			if balance, ok := s.tokenBalances[delegator]; ok {
				power.Add(power, balance)
			}
		}
	}

	return power
}

// =============================================================================
// 信息查询
// =============================================================================

// GetProposal 获取提案信息
func (s *GovernanceSimulator) GetProposal(proposalID uint64) map[string]interface{} {
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return nil
	}

	totalVotes := new(big.Int).Add(proposal.ForVotes, proposal.AgainstVotes)
	totalVotes.Add(totalVotes, proposal.AbstainVotes)

	return map[string]interface{}{
		"id":            proposal.ID,
		"title":         proposal.Title,
		"proposer":      proposal.Proposer,
		"state":         proposal.State,
		"for_votes":     proposal.ForVotes.String(),
		"against_votes": proposal.AgainstVotes.String(),
		"abstain_votes": proposal.AbstainVotes.String(),
		"total_votes":   totalVotes.String(),
		"quorum":        s.config.QuorumVotes.String(),
		"voter_count":   len(proposal.Voters),
		"start_block":   proposal.StartBlock,
		"end_block":     proposal.EndBlock,
		"current_block": s.currentBlock,
	}
}

// ExplainGovernance 解释治理机制
func (s *GovernanceSimulator) ExplainGovernance() map[string]interface{} {
	return map[string]interface{}{
		"overview": "链上治理允许代币持有者对协议变更进行投票",
		"lifecycle": []map[string]string{
			{"stage": "1. 创建", "description": "持有足够代币的用户提交提案"},
			{"stage": "2. 延迟", "description": "等待用户准备投票"},
			{"stage": "3. 投票", "description": "代币持有者投票支持/反对"},
			{"stage": "4. 时间锁", "description": "通过的提案等待执行"},
			{"stage": "5. 执行", "description": "链上执行提案动作"},
		},
		"key_parameters": map[string]interface{}{
			"proposal_threshold": "提案所需最少代币",
			"voting_period":      "投票持续时间",
			"quorum":             "法定投票人数",
			"timelock":           "执行前等待时间",
		},
		"security_features": []string{
			"时间锁给用户退出时间",
			"法定人数防止少数人操纵",
			"提案门槛防止垃圾提案",
		},
	}
}

// updateState 更新状态
func (s *GovernanceSimulator) updateState() {
	s.SetGlobalData("proposal_count", len(s.proposals))
	s.SetGlobalData("current_block", s.currentBlock)

	activeCount := 0
	for _, p := range s.proposals {
		if p.State == ProposalActive {
			activeCount++
		}
	}
	s.SetGlobalData("active_proposals", activeCount)

	summary := fmt.Sprintf("当前区块高度为 %d，共有 %d 个提案，其中 %d 个处于活跃投票阶段。", s.currentBlock, len(s.proposals), activeCount)
	nextHint := "可以继续创建提案、投票或执行提案，观察治理状态如何推进。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"governance_lifecycle",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"proposal_count": len(s.proposals), "active_proposals": activeCount, "current_block": s.currentBlock},
	)
}

func (s *GovernanceSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "advance_block":
		s.currentBlock++
		s.updateState()
		return defiActionResult("已推进一个治理区块。", map[string]interface{}{"current_block": s.currentBlock}, &types.ActionFeedback{
			Summary:     "治理时间线已经向前推进，提案状态可能发生变化。",
			NextHint:    "继续观察提案是否进入投票、时间锁或执行阶段。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"current_block": s.currentBlock},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported governance action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// DeFiGovernanceFactory 治理工厂
type DeFiGovernanceFactory struct{}

// Create 创建演示器
func (f *DeFiGovernanceFactory) Create() engine.Simulator {
	return NewGovernanceSimulator()
}

// GetDescription 获取描述
func (f *DeFiGovernanceFactory) GetDescription() types.Description {
	return NewGovernanceSimulator().GetDescription()
}

// NewDeFiGovernanceFactory 创建工厂
func NewDeFiGovernanceFactory() *DeFiGovernanceFactory {
	return &DeFiGovernanceFactory{}
}

var _ engine.SimulatorFactory = (*DeFiGovernanceFactory)(nil)
