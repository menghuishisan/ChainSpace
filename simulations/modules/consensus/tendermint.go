package consensus

import (
	"fmt"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

type TendermintStep string

const (
	TendermintPropose   TendermintStep = "propose"
	TendermintPrevote   TendermintStep = "prevote"
	TendermintPrecommit TendermintStep = "precommit"
	TendermintCommit    TendermintStep = "commit"
)

type TendermintNode struct {
	ID             types.NodeID   `json:"id"`
	VotingPower    uint64         `json:"voting_power"`
	Height         uint64         `json:"height"`
	Round          uint64         `json:"round"`
	Step           TendermintStep `json:"step"`
	LockedValue    string         `json:"locked_value"`
	LockedRound    int64          `json:"locked_round"`
	ValidValue     string         `json:"valid_value"`
	ValidRound     int64          `json:"valid_round"`
	PrevoteCount   map[string]int `json:"prevote_count"`
	PrecommitCount map[string]int `json:"precommit_count"`
	IsByzantine    bool           `json:"is_byzantine"`
}

type TendermintBlock struct {
	Hash       string       `json:"hash"`
	ParentHash string       `json:"parent_hash"`
	Height     uint64       `json:"height"`
	Round      uint64       `json:"round"`
	Proposer   types.NodeID `json:"proposer"`
	Timestamp  time.Time    `json:"timestamp"`
}

type TendermintMessage struct {
	ID        string         `json:"id"`
	Type      TendermintStep `json:"type"`
	Height    uint64         `json:"height"`
	Round     uint64         `json:"round"`
	BlockHash string         `json:"block_hash"`
	From      types.NodeID   `json:"from"`
	Timestamp time.Time      `json:"timestamp"`
}

type TendermintSimulator struct {
	*base.BaseSimulator
	mu               sync.RWMutex
	nodes            map[types.NodeID]*TendermintNode
	nodeList         []types.NodeID
	chain            []*TendermintBlock
	currentHeight    uint64
	currentRound     uint64
	currentStep      TendermintStep
	proposal         *TendermintBlock
	prevotes         map[string][]types.NodeID
	precommits       map[string][]types.NodeID
	msgQueue         []*TendermintMessage
	totalPower       uint64
	threshold        uint64
	timeoutPropose   int
	timeoutPrevote   int
	timeoutPrecommit int
	stepTicks        int
}

func NewTendermintSimulator() *TendermintSimulator {
	sim := &TendermintSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"tendermint",
			"Tendermint 共识算法",
			"模拟 Tendermint 的提案、Prevote、Precommit、提交和轮次切换过程。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:      make(map[types.NodeID]*TendermintNode),
		nodeList:   make([]types.NodeID, 0),
		chain:      make([]*TendermintBlock, 0),
		prevotes:   make(map[string][]types.NodeID),
		precommits: make(map[string][]types.NodeID),
		msgQueue:   make([]*TendermintMessage, 0),
	}

	sim.AddParam(types.Param{
		Key:         "validator_count",
		Name:        "验证者数量",
		Description: "参与 Tendermint 共识的验证者数量。",
		Type:        types.ParamTypeInt,
		Default:     4,
		Min:         4,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "timeout_propose",
		Name:        "提案超时(tick)",
		Description: "等待提案的最长时间。",
		Type:        types.ParamTypeInt,
		Default:     10,
		Min:         5,
		Max:         50,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *TendermintSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	validatorCount := 4
	if v, ok := config.Params["validator_count"]; ok {
		if n, ok := v.(float64); ok {
			validatorCount = int(n)
		}
	}

	s.timeoutPropose = 10
	if v, ok := config.Params["timeout_propose"]; ok {
		if n, ok := v.(float64); ok {
			s.timeoutPropose = int(n)
		}
	}
	s.timeoutPrevote = 10
	s.timeoutPrecommit = 10

	s.nodes = make(map[types.NodeID]*TendermintNode)
	s.nodeList = make([]types.NodeID, 0, validatorCount)
	s.totalPower = 0

	for i := 0; i < validatorCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("validator-%d", i))
		node := &TendermintNode{
			ID:             nodeID,
			VotingPower:    100,
			Height:         1,
			Round:          0,
			Step:           TendermintPropose,
			LockedRound:    -1,
			ValidRound:     -1,
			PrevoteCount:   make(map[string]int),
			PrecommitCount: make(map[string]int),
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
		s.totalPower += node.VotingPower
	}

	s.threshold = s.totalPower * 2 / 3
	s.currentHeight = 1
	s.currentRound = 0
	s.currentStep = TendermintPropose
	s.stepTicks = 0
	s.proposal = nil
	s.prevotes = make(map[string][]types.NodeID)
	s.precommits = make(map[string][]types.NodeID)
	s.msgQueue = make([]*TendermintMessage, 0)

	genesis := &TendermintBlock{
		Hash:      "genesis",
		Height:    0,
		Timestamp: time.Now(),
	}
	s.chain = []*TendermintBlock{genesis}

	s.updateAllStates()
	return nil
}

func (s *TendermintSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stepTicks++
	s.processMessages()

	switch s.currentStep {
	case TendermintPropose:
		if s.stepTicks == 1 {
			s.doPropose()
		} else if s.stepTicks >= s.timeoutPropose {
			s.enterPrevote("")
		}
	case TendermintPrevote:
		if s.checkPrevoteQuorum() || s.stepTicks >= s.timeoutPrevote {
			s.enterPrecommit()
		}
	case TendermintPrecommit:
		if s.checkPrecommitQuorum() {
			s.commitBlock()
		} else if s.stepTicks >= s.timeoutPrecommit {
			s.nextRound()
		}
	}

	s.updateAllStates()
	return nil
}

func (s *TendermintSimulator) getProposer() types.NodeID {
	return s.nodeList[int(s.currentHeight+s.currentRound)%len(s.nodeList)]
}

func (s *TendermintSimulator) doPropose() {
	proposerID := s.getProposer()
	proposer := s.nodes[proposerID]
	if proposer == nil || proposer.IsByzantine {
		return
	}

	prevBlock := s.chain[len(s.chain)-1]
	block := &TendermintBlock{
		Hash:       fmt.Sprintf("block-%d-%d-%s", s.currentHeight, s.currentRound, uuid.New().String()[:8]),
		ParentHash: prevBlock.Hash,
		Height:     s.currentHeight,
		Round:      s.currentRound,
		Proposer:   proposerID,
		Timestamp:  time.Now(),
	}
	if proposer.ValidValue != "" && proposer.ValidRound >= 0 {
		block.Hash = proposer.ValidValue
	}

	s.proposal = block
	for range s.nodeList {
		s.msgQueue = append(s.msgQueue, &TendermintMessage{
			ID:        uuid.New().String(),
			Type:      TendermintPropose,
			Height:    s.currentHeight,
			Round:     s.currentRound,
			BlockHash: block.Hash,
			From:      proposerID,
			Timestamp: time.Now(),
		})
	}

	s.EmitEvent("propose", proposerID, "", map[string]interface{}{
		"height":     s.currentHeight,
		"round":      s.currentRound,
		"block_hash": block.Hash,
	})
}

func (s *TendermintSimulator) enterPrevote(blockHash string) {
	s.currentStep = TendermintPrevote
	s.stepTicks = 0

	for _, nodeID := range s.nodeList {
		node := s.nodes[nodeID]
		if node == nil || node.IsByzantine {
			continue
		}

		voteHash := blockHash
		if s.proposal != nil {
			if node.LockedRound == -1 || node.LockedValue == s.proposal.Hash {
				voteHash = s.proposal.Hash
			} else {
				voteHash = "nil"
			}
		} else {
			voteHash = "nil"
		}

		node.Step = TendermintPrevote
		s.prevotes[voteHash] = append(s.prevotes[voteHash], nodeID)
		s.msgQueue = append(s.msgQueue, &TendermintMessage{
			ID:        uuid.New().String(),
			Type:      TendermintPrevote,
			Height:    s.currentHeight,
			Round:     s.currentRound,
			BlockHash: voteHash,
			From:      nodeID,
			Timestamp: time.Now(),
		})

		s.EmitEvent("prevote", nodeID, "", map[string]interface{}{
			"height":     s.currentHeight,
			"round":      s.currentRound,
			"block_hash": voteHash,
		})
	}
}

func (s *TendermintSimulator) checkPrevoteQuorum() bool {
	for hash, voters := range s.prevotes {
		if hash == "nil" {
			continue
		}

		power := uint64(0)
		for _, voterID := range voters {
			if node := s.nodes[voterID]; node != nil {
				power += node.VotingPower
			}
		}
		if power > s.threshold {
			return true
		}
	}
	return false
}

func (s *TendermintSimulator) enterPrecommit() {
	s.currentStep = TendermintPrecommit
	s.stepTicks = 0

	quorumHash := ""
	for hash, voters := range s.prevotes {
		power := uint64(0)
		for _, voterID := range voters {
			if node := s.nodes[voterID]; node != nil {
				power += node.VotingPower
			}
		}
		if power > s.threshold {
			quorumHash = hash
			break
		}
	}

	for _, nodeID := range s.nodeList {
		node := s.nodes[nodeID]
		if node == nil || node.IsByzantine {
			continue
		}

		voteHash := "nil"
		if quorumHash != "" && quorumHash != "nil" {
			voteHash = quorumHash
			node.LockedValue = quorumHash
			node.LockedRound = int64(s.currentRound)
			node.ValidValue = quorumHash
			node.ValidRound = int64(s.currentRound)
		}

		node.Step = TendermintPrecommit
		s.precommits[voteHash] = append(s.precommits[voteHash], nodeID)
		s.EmitEvent("precommit", nodeID, "", map[string]interface{}{
			"height":     s.currentHeight,
			"round":      s.currentRound,
			"block_hash": voteHash,
		})
	}
}

func (s *TendermintSimulator) checkPrecommitQuorum() bool {
	for hash, voters := range s.precommits {
		if hash == "nil" {
			continue
		}

		power := uint64(0)
		for _, voterID := range voters {
			if node := s.nodes[voterID]; node != nil {
				power += node.VotingPower
			}
		}
		if power > s.threshold {
			return true
		}
	}
	return false
}

func (s *TendermintSimulator) commitBlock() {
	if s.proposal == nil {
		s.nextRound()
		return
	}

	s.chain = append(s.chain, s.proposal)
	s.EmitEvent("commit", "", "", map[string]interface{}{
		"height":     s.currentHeight,
		"round":      s.currentRound,
		"block_hash": s.proposal.Hash,
	})

	s.currentHeight++
	s.currentRound = 0
	s.currentStep = TendermintPropose
	s.stepTicks = 0
	s.proposal = nil
	s.prevotes = make(map[string][]types.NodeID)
	s.precommits = make(map[string][]types.NodeID)

	for _, node := range s.nodes {
		node.Height = s.currentHeight
		node.Round = 0
		node.Step = TendermintPropose
		node.LockedRound = -1
		node.LockedValue = ""
		node.PrevoteCount = make(map[string]int)
		node.PrecommitCount = make(map[string]int)
	}
}

func (s *TendermintSimulator) nextRound() {
	s.currentRound++
	s.currentStep = TendermintPropose
	s.stepTicks = 0
	s.proposal = nil
	s.prevotes = make(map[string][]types.NodeID)
	s.precommits = make(map[string][]types.NodeID)

	for _, node := range s.nodes {
		node.Round = s.currentRound
		node.Step = TendermintPropose
	}

	s.EmitEvent("new_round", "", "", map[string]interface{}{
		"height": s.currentHeight,
		"round":  s.currentRound,
	})
}

func (s *TendermintSimulator) processMessages() {
	processed := 0
	for processed < 10 && len(s.msgQueue) > 0 {
		s.msgQueue = s.msgQueue[1:]
		processed++
	}
}

func (s *TendermintSimulator) updateAllStates() {
	for nodeID, node := range s.nodes {
		s.SetNodeState(nodeID, &types.NodeState{
			ID:          nodeID,
			Status:      string(node.Step),
			IsByzantine: node.IsByzantine,
			Data: map[string]interface{}{
				"voting_power": node.VotingPower,
				"height":       node.Height,
				"round":        node.Round,
				"locked_round": node.LockedRound,
				"locked_value": node.LockedValue,
			},
		})
	}

	s.SetGlobalData("height", s.currentHeight)
	s.SetGlobalData("round", s.currentRound)
	s.SetGlobalData("step", s.currentStep)
	s.SetGlobalData("chain_length", len(s.chain))
	s.SetGlobalData("proposer", s.getProposer())
	s.SetGlobalData("threshold", s.threshold)
	s.SetGlobalData("current_actor", s.getProposer())
	s.SetGlobalData("committee_size", len(s.nodeList))
	s.SetGlobalData("result_height", s.currentHeight)
	setConsensusTeachingState(
		s.BaseSimulator,
		"tendermint_round",
		"当前 Tendermint 正在按 propose、prevote、precommit、commit 的顺序推进本轮共识。",
		"继续观察当前轮次和步骤变化，确认阈值是否足够让区块进入 commit。",
		70,
		map[string]interface{}{
			"height":        s.currentHeight,
			"round":         s.currentRound,
			"step":          s.currentStep,
			"proposer":      s.getProposer(),
			"result_height": s.currentHeight,
		},
	)
}

func (s *TendermintSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "submit_proposal":
		s.mu.Lock()
		s.doPropose()
		s.updateAllStates()
		proposer := s.getProposer()
		height := s.currentHeight
		round := s.currentRound
		s.mu.Unlock()

		return consensusActionResult(
			"已发起 Tendermint 提案",
			map[string]interface{}{
				"proposer": proposer,
				"height":   height,
				"round":    round,
			},
			&types.ActionFeedback{
				Summary:     "当前提议者已发起新一轮提案。",
				NextHint:    "观察 Prevote 与 Precommit 是否达到阈值，并进入提交阶段。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "proposal_submitted"},
			},
		), nil
	case "force_next_round":
		s.mu.Lock()
		s.nextRound()
		s.updateAllStates()
		height := s.currentHeight
		round := s.currentRound
		s.mu.Unlock()

		return consensusActionResult(
			"已进入下一轮",
			map[string]interface{}{
				"height": height,
				"round":  round,
			},
			&types.ActionFeedback{
				Summary:     "当前高度已进入新的共识轮次。",
				NextHint:    "观察新的提议者与投票路径，以及上一轮未完成的原因。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "next_round_started"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported tendermint action: %s", action)
	}
}

type TendermintFactory struct{}

func (f *TendermintFactory) Create() engine.Simulator { return NewTendermintSimulator() }

func (f *TendermintFactory) GetDescription() types.Description {
	return NewTendermintSimulator().GetDescription()
}

func NewTendermintFactory() *TendermintFactory { return &TendermintFactory{} }

var _ engine.SimulatorFactory = (*TendermintFactory)(nil)
