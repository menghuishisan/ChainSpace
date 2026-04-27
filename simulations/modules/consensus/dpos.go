package consensus

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type DPoSDelegate struct {
	ID             types.NodeID `json:"id"`
	Votes          uint64       `json:"votes"`
	IsActive       bool         `json:"is_active"`
	BlocksProduced int          `json:"blocks_produced"`
	MissedBlocks   int          `json:"missed_blocks"`
	Rewards        uint64       `json:"rewards"`
	Reliability    float64      `json:"reliability"`
	RegisteredAt   uint64       `json:"registered_at"`
}

type DPoSVoter struct {
	ID       types.NodeID   `json:"id"`
	Balance  uint64         `json:"balance"`
	VotedFor []types.NodeID `json:"voted_for"`
}

type DPoSBlock struct {
	Hash      string       `json:"hash"`
	PrevHash  string       `json:"prev_hash"`
	Height    uint64       `json:"height"`
	Producer  types.NodeID `json:"producer"`
	Timestamp time.Time    `json:"timestamp"`
	Round     uint64       `json:"round"`
}

type DPoSSimulator struct {
	*base.BaseSimulator
	mu              sync.RWMutex
	delegates       map[types.NodeID]*DPoSDelegate
	voters          map[types.NodeID]*DPoSVoter
	delegateList    []types.NodeID
	activeDelegates []types.NodeID
	chain           []*DPoSBlock
	currentRound    uint64
	roundIndex      int
	maxDelegates    int
	blockInterval   int
	roundLength     int
}

func NewDPoSSimulator() *DPoSSimulator {
	sim := &DPoSSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"dpos",
			"DPoS 委托权益证明",
			"模拟委托投票、活跃委托者选举、轮值出块和失误漏块过程。",
			"consensus",
			types.ComponentProcess,
		),
		delegates:       make(map[types.NodeID]*DPoSDelegate),
		voters:          make(map[types.NodeID]*DPoSVoter),
		delegateList:    make([]types.NodeID, 0),
		activeDelegates: make([]types.NodeID, 0),
		chain:           make([]*DPoSBlock, 0),
	}

	sim.AddParam(types.Param{
		Key:         "delegate_count",
		Name:        "候选委托数",
		Description: "参与竞选的委托节点总数。",
		Type:        types.ParamTypeInt,
		Default:     30,
		Min:         10,
		Max:         100,
	})
	sim.AddParam(types.Param{
		Key:         "active_delegates",
		Name:        "活跃委托数",
		Description: "每轮负责出块的活跃委托者数量。",
		Type:        types.ParamTypeInt,
		Default:     21,
		Min:         3,
		Max:         51,
	})
	sim.AddParam(types.Param{
		Key:         "voter_count",
		Name:        "投票者数量",
		Description: "参与委托投票的用户数量。",
		Type:        types.ParamTypeInt,
		Default:     100,
		Min:         10,
		Max:         1000,
	})
	sim.AddParam(types.Param{
		Key:         "block_interval",
		Name:        "出块间隔(tick)",
		Description: "活跃委托者轮值出块的间隔。",
		Type:        types.ParamTypeInt,
		Default:     3,
		Min:         1,
		Max:         10,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *DPoSSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delegateCount := 30
	if v, ok := config.Params["delegate_count"]; ok {
		if n, ok := v.(float64); ok {
			delegateCount = int(n)
		}
	}

	s.maxDelegates = 21
	if v, ok := config.Params["active_delegates"]; ok {
		if n, ok := v.(float64); ok {
			s.maxDelegates = int(n)
		}
	}

	voterCount := 100
	if v, ok := config.Params["voter_count"]; ok {
		if n, ok := v.(float64); ok {
			voterCount = int(n)
		}
	}

	s.blockInterval = 3
	if v, ok := config.Params["block_interval"]; ok {
		if n, ok := v.(float64); ok {
			s.blockInterval = int(n)
		}
	}

	s.roundLength = s.maxDelegates * s.blockInterval
	s.delegates = make(map[types.NodeID]*DPoSDelegate)
	s.voters = make(map[types.NodeID]*DPoSVoter)
	s.delegateList = make([]types.NodeID, 0, delegateCount)
	s.activeDelegates = make([]types.NodeID, 0, s.maxDelegates)
	s.currentRound = 0
	s.roundIndex = 0

	for i := 0; i < delegateCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("delegate-%d", i))
		s.delegates[nodeID] = &DPoSDelegate{
			ID:          nodeID,
			IsActive:    false,
			Reliability: 0.95 + rand.Float64()*0.05,
		}
		s.delegateList = append(s.delegateList, nodeID)
	}

	for i := 0; i < voterCount; i++ {
		voterID := types.NodeID(fmt.Sprintf("voter-%d", i))
		voter := &DPoSVoter{
			ID:       voterID,
			Balance:  uint64(1000 + rand.Intn(9000)),
			VotedFor: make([]types.NodeID, 0),
		}

		numVotes := 1 + rand.Intn(5)
		for j := 0; j < numVotes && j < delegateCount; j++ {
			delegateID := s.delegateList[rand.Intn(delegateCount)]
			voter.VotedFor = append(voter.VotedFor, delegateID)
			s.delegates[delegateID].Votes += voter.Balance / uint64(numVotes)
		}

		s.voters[voterID] = voter
	}

	s.electDelegates()

	genesis := &DPoSBlock{
		Hash:      "genesis",
		Height:    0,
		Timestamp: time.Now(),
	}
	s.chain = []*DPoSBlock{genesis}
	s.updateAllStates()
	return nil
}

func (s *DPoSSimulator) electDelegates() {
	type delegateVote struct {
		id    types.NodeID
		votes uint64
	}

	sorted := make([]delegateVote, 0, len(s.delegates))
	for _, delegate := range s.delegates {
		sorted = append(sorted, delegateVote{id: delegate.ID, votes: delegate.Votes})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].votes > sorted[j].votes
	})

	for _, delegate := range s.delegates {
		delegate.IsActive = false
	}

	s.activeDelegates = make([]types.NodeID, 0, s.maxDelegates)
	for i := 0; i < s.maxDelegates && i < len(sorted); i++ {
		delegateID := sorted[i].id
		s.activeDelegates = append(s.activeDelegates, delegateID)
		s.delegates[delegateID].IsActive = true
	}

	rand.Shuffle(len(s.activeDelegates), func(i, j int) {
		s.activeDelegates[i], s.activeDelegates[j] = s.activeDelegates[j], s.activeDelegates[i]
	})

	topDelegate := ""
	if len(s.activeDelegates) > 0 {
		topDelegate = string(s.activeDelegates[0])
	}

	s.EmitEvent("delegates_elected", "", "", map[string]interface{}{
		"active_count": len(s.activeDelegates),
		"top_delegate": topDelegate,
	})
}

func (s *DPoSSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tick%uint64(s.blockInterval) == 0 {
		s.produceBlock(tick)
	}
	if tick > 0 && tick%uint64(s.roundLength) == 0 {
		s.currentRound++
		s.roundIndex = 0
		s.electDelegates()
	}

	s.updateAllStates()
	return nil
}

func (s *DPoSSimulator) produceBlock(tick uint64) {
	if len(s.activeDelegates) == 0 {
		return
	}

	producerID := s.activeDelegates[s.roundIndex%len(s.activeDelegates)]
	producer := s.delegates[producerID]
	if producer == nil {
		return
	}

	if rand.Float64() > producer.Reliability {
		producer.MissedBlocks++
		s.EmitEvent("block_missed", producerID, "", map[string]interface{}{
			"round":        s.currentRound,
			"missed_total": producer.MissedBlocks,
		})
		s.roundIndex++
		return
	}

	prevBlock := s.chain[len(s.chain)-1]
	block := &DPoSBlock{
		Hash:      fmt.Sprintf("block-%d", len(s.chain)),
		PrevHash:  prevBlock.Hash,
		Height:    uint64(len(s.chain)),
		Producer:  producerID,
		Round:     s.currentRound,
		Timestamp: time.Now(),
	}

	s.chain = append(s.chain, block)
	producer.BlocksProduced++
	producer.Rewards += 100

	s.EmitEvent("block_produced", producerID, "", map[string]interface{}{
		"height": block.Height,
		"round":  s.currentRound,
	})

	s.roundIndex++
}

func (s *DPoSSimulator) Vote(voterID types.NodeID, delegateIDs []types.NodeID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	voter := s.voters[voterID]
	if voter == nil {
		return fmt.Errorf("voter not found: %s", voterID)
	}
	if len(delegateIDs) == 0 {
		return fmt.Errorf("delegate list is empty")
	}

	if len(voter.VotedFor) > 0 {
		oldWeight := voter.Balance / uint64(len(voter.VotedFor))
		for _, oldDelegateID := range voter.VotedFor {
			if delegate := s.delegates[oldDelegateID]; delegate != nil {
				delegate.Votes -= oldWeight
			}
		}
	}

	voter.VotedFor = delegateIDs
	voteWeight := voter.Balance / uint64(len(delegateIDs))
	for _, delegateID := range delegateIDs {
		if delegate := s.delegates[delegateID]; delegate != nil {
			delegate.Votes += voteWeight
		}
	}

	s.EmitEvent("vote_cast", voterID, "", map[string]interface{}{
		"delegates": delegateIDs,
		"weight":    voteWeight,
	})
	return nil
}

func (s *DPoSSimulator) updateAllStates() {
	for nodeID, delegate := range s.delegates {
		status := "candidate"
		if delegate.IsActive {
			status = "active"
		}
		s.SetNodeState(nodeID, &types.NodeState{
			ID:     nodeID,
			Status: status,
			Data: map[string]interface{}{
				"votes":           delegate.Votes,
				"blocks_produced": delegate.BlocksProduced,
				"missed_blocks":   delegate.MissedBlocks,
				"rewards":         delegate.Rewards,
				"reliability":     delegate.Reliability,
			},
		})
	}

	currentProducer := ""
	if len(s.activeDelegates) > 0 {
		currentProducer = string(s.activeDelegates[s.roundIndex%len(s.activeDelegates)])
	}

	s.SetGlobalData("chain_height", len(s.chain)-1)
	s.SetGlobalData("current_round", s.currentRound)
	s.SetGlobalData("active_delegates", len(s.activeDelegates))
	s.SetGlobalData("total_delegates", len(s.delegates))
	s.SetGlobalData("active_delegate_ids", s.activeDelegates)
	s.SetGlobalData("delegate_target", s.maxDelegates)
	s.SetGlobalData("voter_count", len(s.voters))
	s.SetGlobalData("current_producer", currentProducer)
	s.SetGlobalData("current_actor", currentProducer)
	s.SetGlobalData("committee_size", len(s.activeDelegates))
	s.SetGlobalData("result_height", len(s.chain)-1)
	setConsensusTeachingState(
		s.BaseSimulator,
		"dpos_rotation",
		"当前 DPoS 正在轮换活跃代表并观察出块、漏块和投票变化。",
		"继续观察活跃代表集合与当前出块者，判断选举结果如何影响链上高度推进。",
		65,
		map[string]interface{}{
			"chain_height":     len(s.chain) - 1,
			"current_round":    s.currentRound,
			"active_delegates": len(s.activeDelegates),
			"current_producer": currentProducer,
		},
	)
}

func (s *DPoSSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "re_elect_delegates":
		s.mu.Lock()
		s.electDelegates()
		s.updateAllStates()
		round := s.currentRound
		s.mu.Unlock()

		return consensusActionResult(
			"已重新选举活跃委托者",
			map[string]interface{}{
				"round": round,
			},
			&types.ActionFeedback{
				Summary:     "DPoS 已完成新一轮委托者选举。",
				NextHint:    "观察活跃委托者集合和轮值顺序是否发生变化。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "delegates_elected"},
			},
		), nil
	case "cast_vote":
		voterIndex := 0
		if raw, ok := params["voter"].(float64); ok {
			voterIndex = int(raw)
		}

		delegateID := types.NodeID("delegate-0")
		if raw, ok := params["delegate"].(string); ok && raw != "" {
			delegateID = types.NodeID(raw)
		}

		voterID := types.NodeID(fmt.Sprintf("voter-%d", voterIndex))
		if err := s.Vote(voterID, []types.NodeID{delegateID}); err != nil {
			return nil, err
		}

		s.mu.Lock()
		s.updateAllStates()
		s.mu.Unlock()

		return consensusActionResult(
			"已提交新的委托投票",
			map[string]interface{}{
				"voter":    voterID,
				"delegate": delegateID,
			},
			&types.ActionFeedback{
				Summary:     "新的投票已经计入委托者选举。",
				NextHint:    "观察投票权重如何改变活跃委托者的组成。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "vote_cast"},
			},
		), nil
	case "produce_block_now":
		s.mu.Lock()
		s.produceBlock(uint64(time.Now().UnixNano()))
		s.updateAllStates()
		s.mu.Unlock()

		return consensusActionResult(
			"已触发当前轮值委托者立即出块",
			nil,
			&types.ActionFeedback{
				Summary:     "当前轮值委托者已经开始出块。",
				NextHint:    "观察出块结果是否推进了链高，以及下一位轮值委托者是否切换。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "block_produced"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported dpos action: %s", action)
	}
}

type DPoSFactory struct{}

func (f *DPoSFactory) Create() engine.Simulator          { return NewDPoSSimulator() }
func (f *DPoSFactory) GetDescription() types.Description { return NewDPoSSimulator().GetDescription() }
func NewDPoSFactory() *DPoSFactory                       { return &DPoSFactory{} }

var _ engine.SimulatorFactory = (*DPoSFactory)(nil)
