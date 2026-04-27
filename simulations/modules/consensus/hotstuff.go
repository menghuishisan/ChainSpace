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

type HotStuffPhase string

const (
	HotStuffPrepare   HotStuffPhase = "prepare"
	HotStuffPreCommit HotStuffPhase = "pre_commit"
	HotStuffCommit    HotStuffPhase = "commit"
	HotStuffDecide    HotStuffPhase = "decide"
)

type QC struct {
	Type      HotStuffPhase  `json:"type"`
	ViewNum   uint64         `json:"view_num"`
	BlockHash string         `json:"block_hash"`
	Votes     []types.NodeID `json:"votes"`
}

type HotStuffNode struct {
	ID          types.NodeID  `json:"id"`
	IsLeader    bool          `json:"is_leader"`
	ViewNum     uint64        `json:"view_num"`
	Phase       HotStuffPhase `json:"phase"`
	PrepareQC   *QC           `json:"prepare_qc"`
	LockedQC    *QC           `json:"locked_qc"`
	CommitQC    *QC           `json:"commit_qc"`
	VoteCount   int           `json:"vote_count"`
	IsByzantine bool          `json:"is_byzantine"`
}

type HotStuffBlock struct {
	Hash       string       `json:"hash"`
	ParentHash string       `json:"parent_hash"`
	Height     uint64       `json:"height"`
	ViewNum    uint64       `json:"view_num"`
	Proposer   types.NodeID `json:"proposer"`
	QC         *QC          `json:"qc"`
	Timestamp  time.Time    `json:"timestamp"`
	Committed  bool         `json:"committed"`
}

type HotStuffMessage struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Phase     HotStuffPhase `json:"phase"`
	ViewNum   uint64        `json:"view_num"`
	BlockHash string        `json:"block_hash"`
	From      types.NodeID  `json:"from"`
	QC        *QC           `json:"qc,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

type HotStuffSimulator struct {
	*base.BaseSimulator
	mu           sync.RWMutex
	nodes        map[types.NodeID]*HotStuffNode
	nodeList     []types.NodeID
	chain        []*HotStuffBlock
	pendingBlock *HotStuffBlock
	viewNum      uint64
	phase        HotStuffPhase
	votes        map[string][]types.NodeID
	msgQueue     []*HotStuffMessage
	threshold    int
}

func NewHotStuffSimulator() *HotStuffSimulator {
	sim := &HotStuffSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"hotstuff",
			"HotStuff 共识算法",
			"模拟 HotStuff 的提案、QC 形成、提交和领导者轮换过程。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:    make(map[types.NodeID]*HotStuffNode),
		nodeList: make([]types.NodeID, 0),
		chain:    make([]*HotStuffBlock, 0),
		votes:    make(map[string][]types.NodeID),
		msgQueue: make([]*HotStuffMessage, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "参与 HotStuff 共识的节点数量。",
		Type:        types.ParamTypeInt,
		Default:     4,
		Min:         4,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "pipelining",
		Name:        "启用流水线",
		Description: "是否启用 HotStuff 流水线处理。",
		Type:        types.ParamTypeBool,
		Default:     true,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *HotStuffSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	nodeCount := 4
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	s.threshold = (nodeCount*2)/3 + 1
	s.nodes = make(map[types.NodeID]*HotStuffNode)
	s.nodeList = make([]types.NodeID, 0, nodeCount)
	s.viewNum = 1
	s.phase = HotStuffPrepare
	s.pendingBlock = nil
	s.votes = make(map[string][]types.NodeID)
	s.msgQueue = make([]*HotStuffMessage, 0)

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("node-%d", i))
		node := &HotStuffNode{
			ID:       nodeID,
			IsLeader: i == 0,
			ViewNum:  1,
			Phase:    HotStuffPrepare,
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
	}

	genesis := &HotStuffBlock{
		Hash:      "genesis",
		Height:    0,
		ViewNum:   0,
		Committed: true,
		Timestamp: time.Now(),
	}
	s.chain = []*HotStuffBlock{genesis}

	s.updateAllNodeStates()
	s.updateGlobalState()
	return nil
}

func (s *HotStuffSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processMessages()
	if tick%5 == 0 {
		s.propose(s.getLeader())
	}

	s.updateAllNodeStates()
	s.updateGlobalState()
	return nil
}

func (s *HotStuffSimulator) getLeader() types.NodeID {
	return s.nodeList[int(s.viewNum-1)%len(s.nodeList)]
}

func (s *HotStuffSimulator) propose(leaderID types.NodeID) {
	leader := s.nodes[leaderID]
	if leader == nil || !leader.IsLeader || s.pendingBlock != nil {
		return
	}

	prevBlock := s.chain[len(s.chain)-1]
	block := &HotStuffBlock{
		Hash:       fmt.Sprintf("block-%d-%s", len(s.chain), uuid.New().String()[:8]),
		ParentHash: prevBlock.Hash,
		Height:     uint64(len(s.chain)),
		ViewNum:    s.viewNum,
		Proposer:   leaderID,
		QC:         leader.PrepareQC,
		Timestamp:  time.Now(),
	}

	s.pendingBlock = block
	s.phase = HotStuffPrepare
	s.votes[block.Hash] = []types.NodeID{leaderID}

	for _, nodeID := range s.nodeList {
		if nodeID == leaderID {
			continue
		}
		s.msgQueue = append(s.msgQueue, &HotStuffMessage{
			ID:        uuid.New().String(),
			Type:      "propose",
			Phase:     HotStuffPrepare,
			ViewNum:   s.viewNum,
			BlockHash: block.Hash,
			From:      leaderID,
			Timestamp: time.Now(),
		})
	}

	s.EmitEvent("propose", leaderID, "", map[string]interface{}{
		"view_num":   s.viewNum,
		"block_hash": block.Hash,
		"height":     block.Height,
	})
}

func (s *HotStuffSimulator) processMessages() {
	if len(s.msgQueue) == 0 {
		return
	}

	msg := s.msgQueue[0]
	s.msgQueue = s.msgQueue[1:]

	switch msg.Type {
	case "propose":
		s.handlePropose(msg)
	case "vote":
		s.handleVote(msg)
	}
}

func (s *HotStuffSimulator) handlePropose(msg *HotStuffMessage) {
	for _, nodeID := range s.nodeList {
		if nodeID == msg.From {
			continue
		}

		node := s.nodes[nodeID]
		if node == nil || node.IsByzantine {
			continue
		}

		s.msgQueue = append(s.msgQueue, &HotStuffMessage{
			ID:        uuid.New().String(),
			Type:      "vote",
			Phase:     s.phase,
			ViewNum:   msg.ViewNum,
			BlockHash: msg.BlockHash,
			From:      nodeID,
			Timestamp: time.Now(),
		})
	}
}

func (s *HotStuffSimulator) handleVote(msg *HotStuffMessage) {
	if s.pendingBlock == nil || msg.BlockHash != s.pendingBlock.Hash {
		return
	}

	s.votes[msg.BlockHash] = append(s.votes[msg.BlockHash], msg.From)
	voteCount := len(s.votes[msg.BlockHash])
	s.EmitEvent("vote_received", msg.From, "", map[string]interface{}{
		"phase":      s.phase,
		"vote_count": voteCount,
		"threshold":  s.threshold,
	})

	if voteCount >= s.threshold {
		s.advancePhase()
	}
}

func (s *HotStuffSimulator) advancePhase() {
	if s.pendingBlock == nil {
		return
	}

	qc := &QC{
		Type:      s.phase,
		ViewNum:   s.viewNum,
		BlockHash: s.pendingBlock.Hash,
		Votes:     s.votes[s.pendingBlock.Hash],
	}

	leaderID := s.getLeader()
	leader := s.nodes[leaderID]

	switch s.phase {
	case HotStuffPrepare:
		leader.PrepareQC = qc
		s.phase = HotStuffPreCommit
		s.EmitEvent("prepare_qc_formed", leaderID, "", map[string]interface{}{"view_num": s.viewNum})
	case HotStuffPreCommit:
		leader.LockedQC = qc
		s.phase = HotStuffCommit
		s.EmitEvent("precommit_qc_formed", leaderID, "", map[string]interface{}{"view_num": s.viewNum})
	case HotStuffCommit:
		leader.CommitQC = qc
		s.phase = HotStuffDecide
		s.EmitEvent("commit_qc_formed", leaderID, "", map[string]interface{}{"view_num": s.viewNum})
	case HotStuffDecide:
		s.pendingBlock.Committed = true
		s.chain = append(s.chain, s.pendingBlock)
		s.EmitEvent("block_committed", leaderID, "", map[string]interface{}{
			"height":   s.pendingBlock.Height,
			"view_num": s.viewNum,
		})
		s.pendingBlock = nil
		s.votes = make(map[string][]types.NodeID)
		s.viewNum++
		s.phase = HotStuffPrepare
		s.rotateLeader()
		return
	}

	s.votes[s.pendingBlock.Hash] = []types.NodeID{leaderID}
}

func (s *HotStuffSimulator) rotateLeader() {
	for _, node := range s.nodes {
		node.IsLeader = false
		node.ViewNum = s.viewNum
		node.Phase = HotStuffPrepare
	}

	newLeaderID := s.getLeader()
	if newLeader := s.nodes[newLeaderID]; newLeader != nil {
		newLeader.IsLeader = true
	}

	s.EmitEvent("leader_rotated", newLeaderID, "", map[string]interface{}{
		"view_num": s.viewNum,
	})
}

func (s *HotStuffSimulator) updateAllNodeStates() {
	for nodeID, node := range s.nodes {
		s.SetNodeState(nodeID, &types.NodeState{
			ID:          nodeID,
			Status:      string(node.Phase),
			IsByzantine: node.IsByzantine,
			Data: map[string]interface{}{
				"is_leader":  node.IsLeader,
				"view_num":   node.ViewNum,
				"vote_count": node.VoteCount,
			},
		})
	}
}

func (s *HotStuffSimulator) updateGlobalState() {
	s.SetGlobalData("view_num", s.viewNum)
	s.SetGlobalData("phase", s.phase)
	s.SetGlobalData("chain_height", len(s.chain)-1)
	s.SetGlobalData("leader", s.getLeader())
	s.SetGlobalData("threshold", s.threshold)
	s.SetGlobalData("current_actor", s.getLeader())
	s.SetGlobalData("committee_size", len(s.nodeList))
	s.SetGlobalData("result_height", len(s.chain)-1)
	setConsensusTeachingState(
		s.BaseSimulator,
		"hotstuff_qc",
		"当前 HotStuff 正在围绕 propose、prepare、precommit、commit QC 推进链上确认。",
		"继续观察领导者和 QC 形成情况，确认区块是否顺利进入最终提交。",
		70,
		map[string]interface{}{
			"view_num":      s.viewNum,
			"phase":         s.phase,
			"leader":        s.getLeader(),
			"result_height": len(s.chain) - 1,
		},
	)
}

func (s *HotStuffSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "submit_proposal":
		s.mu.Lock()
		leaderID := s.getLeader()
		s.propose(leaderID)
		s.updateAllNodeStates()
		s.updateGlobalState()
		view := s.viewNum
		s.mu.Unlock()

		return consensusActionResult(
			"已发起新的 HotStuff 提案",
			map[string]interface{}{
				"leader": leaderID,
				"view":   view,
			},
			&types.ActionFeedback{
				Summary:     "当前领导者已广播新提案，QC 链将继续推进。",
				NextHint:    "观察 Prepare、Precommit 与 Commit QC 是否依次形成。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "proposal_submitted"},
			},
		), nil
	case "rotate_leader":
		s.mu.Lock()
		s.viewNum++
		s.rotateLeader()
		s.updateAllNodeStates()
		s.updateGlobalState()
		leaderID := s.getLeader()
		view := s.viewNum
		s.mu.Unlock()

		return consensusActionResult(
			"已轮换到下一位领导者",
			map[string]interface{}{
				"leader": leaderID,
				"view":   view,
			},
			&types.ActionFeedback{
				Summary:     "领导者已经轮换，新的视图正在接管提案流程。",
				NextHint:    "观察新领导者是否继续推动 QC 链和最终提交。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "leader_rotated"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported hotstuff action: %s", action)
	}
}

type HotStuffFactory struct{}

func (f *HotStuffFactory) Create() engine.Simulator { return NewHotStuffSimulator() }

func (f *HotStuffFactory) GetDescription() types.Description {
	return NewHotStuffSimulator().GetDescription()
}

func NewHotStuffFactory() *HotStuffFactory { return &HotStuffFactory{} }

var _ engine.SimulatorFactory = (*HotStuffFactory)(nil)
