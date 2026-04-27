package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

type PBFTState string

const (
	PBFTStateIdle       PBFTState = "idle"
	PBFTStatePrePrepare PBFTState = "pre_prepare"
	PBFTStatePrepare    PBFTState = "prepare"
	PBFTStateCommit     PBFTState = "commit"
	PBFTStateReply      PBFTState = "reply"
)

type PBFTNode struct {
	ID           types.NodeID            `json:"id"`
	IsPrimary    bool                    `json:"is_primary"`
	View         uint64                  `json:"view"`
	Sequence     uint64                  `json:"sequence"`
	State        PBFTState               `json:"state"`
	IsByzantine  bool                    `json:"is_byzantine"`
	ByzBehavior  types.ByzantineBehavior `json:"byz_behavior,omitempty"`
	PrepareCount map[string]int          `json:"prepare_count"`
	CommitCount  map[string]int          `json:"commit_count"`
	CommittedSeq uint64                  `json:"committed_seq"`
	MessageLog   []*PBFTMessage          `json:"message_log"`
}

type PBFTMessage struct {
	ID        string       `json:"id"`
	Type      string       `json:"type"`
	View      uint64       `json:"view"`
	Sequence  uint64       `json:"sequence"`
	Digest    string       `json:"digest"`
	From      types.NodeID `json:"from"`
	To        types.NodeID `json:"to,omitempty"`
	Request   []byte       `json:"request,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

type PBFTSimulator struct {
	*base.BaseSimulator
	mu           sync.RWMutex
	nodes        map[types.NodeID]*PBFTNode
	nodeList     []types.NodeID
	nodeCount    int
	faultCount   int
	currentView  uint64
	currentSeq   uint64
	pendingReqs  []*PBFTMessage
	messageQueue []*PBFTMessage
	committed    []string
}

func NewPBFTSimulator() *PBFTSimulator {
	sim := &PBFTSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"pbft",
			"PBFT共识算法",
			"实用拜占庭容错算法的教学模拟器，支持预准备、准备、提交、视图切换和拜占庭行为注入。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:        make(map[types.NodeID]*PBFTNode),
		nodeList:     make([]types.NodeID, 0),
		messageQueue: make([]*PBFTMessage, 0),
		pendingReqs:  make([]*PBFTMessage, 0),
		committed:    make([]string, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "PBFT 网络中的节点总数，应满足 n >= 3f + 1。",
		Type:        types.ParamTypeInt,
		Default:     4,
		Min:         4,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "byzantine_count",
		Name:        "拜占庭节点数量",
		Description: "初始化时注入的拜占庭节点数量。",
		Type:        types.ParamTypeInt,
		Default:     0,
		Min:         0,
		Max:         6,
	})
	sim.AddParam(types.Param{
		Key:         "auto_request",
		Name:        "自动发起请求",
		Description: "是否自动发起客户端请求。",
		Type:        types.ParamTypeBool,
		Default:     true,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *PBFTSimulator) Init(config types.Config) error {
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

	byzantineCount := 0
	if v, ok := config.Params["byzantine_count"]; ok {
		if n, ok := v.(float64); ok {
			byzantineCount = int(n)
		}
	}

	s.nodeCount = nodeCount
	s.faultCount = (nodeCount - 1) / 3
	s.currentView = 0
	s.currentSeq = 0
	s.nodes = make(map[types.NodeID]*PBFTNode)
	s.nodeList = make([]types.NodeID, 0, nodeCount)
	s.messageQueue = make([]*PBFTMessage, 0)
	s.pendingReqs = make([]*PBFTMessage, 0)
	s.committed = make([]string, 0)

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("node-%d", i))
		node := &PBFTNode{
			ID:           nodeID,
			IsPrimary:    i == 0,
			View:         0,
			Sequence:     0,
			State:        PBFTStateIdle,
			IsByzantine:  i < byzantineCount,
			PrepareCount: make(map[string]int),
			CommitCount:  make(map[string]int),
			MessageLog:   make([]*PBFTMessage, 0),
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)

		s.SetNodeState(nodeID, &types.NodeState{
			ID:          nodeID,
			Status:      string(PBFTStateIdle),
			IsByzantine: node.IsByzantine,
			Data: map[string]interface{}{
				"is_primary":    node.IsPrimary,
				"view":          node.View,
				"sequence":      node.Sequence,
				"committed_seq": node.CommittedSeq,
			},
		})
	}

	s.SetGlobalData("view", s.currentView)
	s.SetGlobalData("sequence", s.currentSeq)
	s.SetGlobalData("node_count", s.nodeCount)
	s.SetGlobalData("fault_tolerance", s.faultCount)
	s.SetGlobalData("request_count", 0)
	s.SetGlobalData("committed_count", 0)
	s.SetGlobalData("current_actor", s.getPrimaryID())
	s.SetGlobalData("committee_size", s.nodeCount)
	s.SetGlobalData("result_height", 0)

	s.updateNodeStates()
	return nil
}

func (s *PBFTSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processMessageQueue()

	autoReq := true
	if p := s.BaseSimulator.GetParams()["auto_request"]; p.Value != nil {
		if v, ok := p.Value.(bool); ok {
			autoReq = v
		}
	}

	if autoReq && (tick == 1 || tick%4 == 0) && len(s.pendingReqs) == 0 {
		s.submitRequest([]byte(fmt.Sprintf("request-%d", tick)))
	}

	s.updateNodeStates()
	return nil
}

func (s *PBFTSimulator) submitRequest(data []byte) {
	digest := s.computeDigest(data)
	s.currentSeq++

	req := &PBFTMessage{
		ID:        uuid.New().String(),
		Type:      "REQUEST",
		Sequence:  s.currentSeq,
		Digest:    digest,
		Request:   data,
		Timestamp: time.Now(),
	}
	s.pendingReqs = append(s.pendingReqs, req)
	s.SetGlobalData("request_count", s.currentSeq)

	primaryID := s.getPrimaryID()
	s.SetGlobalData("current_actor", primaryID)
	s.EmitEvent("client_request", "", primaryID, map[string]interface{}{
		"sequence": s.currentSeq,
		"digest":   digest,
	})

	s.handlePrePrepare(primaryID, req)
}

func (s *PBFTSimulator) handlePrePrepare(primaryID types.NodeID, req *PBFTMessage) {
	primary := s.nodes[primaryID]
	if primary == nil || !primary.IsPrimary {
		return
	}
	if primary.IsByzantine && primary.ByzBehavior == types.ByzantineIgnore {
		return
	}

	prePrepare := &PBFTMessage{
		ID:        uuid.New().String(),
		Type:      "PRE-PREPARE",
		View:      s.currentView,
		Sequence:  req.Sequence,
		Digest:    req.Digest,
		From:      primaryID,
		Request:   req.Request,
		Timestamp: time.Now(),
	}

	primary.State = PBFTStatePrePrepare
	primary.Sequence = req.Sequence
	primary.PrepareCount[req.Digest] = 1
	primary.CommitCount = make(map[string]int)
	primary.MessageLog = append(primary.MessageLog, prePrepare)

	for _, nodeID := range s.nodeList {
		if nodeID != primaryID {
			s.queueMessage(prePrepare, nodeID)
		}
	}

	s.EmitEvent("pre_prepare", primaryID, "", map[string]interface{}{
		"view":     s.currentView,
		"sequence": req.Sequence,
		"digest":   req.Digest,
	})
}

func (s *PBFTSimulator) processMessageQueue() {
	if len(s.messageQueue) == 0 {
		return
	}

	msg := s.messageQueue[0]
	s.messageQueue = s.messageQueue[1:]

	if msg.To != "" {
		s.deliverMessage(msg.To, msg)
		return
	}

	for _, nodeID := range s.nodeList {
		if nodeID == msg.From {
			continue
		}
		s.deliverMessage(nodeID, msg)
	}
}

func (s *PBFTSimulator) deliverMessage(nodeID types.NodeID, msg *PBFTMessage) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}
	if node.IsByzantine && node.ByzBehavior == types.ByzantineIgnore {
		return
	}

	node.MessageLog = append(node.MessageLog, msg)

	switch msg.Type {
	case "PRE-PREPARE":
		s.handlePrepare(nodeID, msg)
	case "PREPARE":
		s.handlePrepareReceive(nodeID, msg)
	case "COMMIT":
		s.handleCommitReceive(nodeID, msg)
	}
}

func (s *PBFTSimulator) handlePrepare(nodeID types.NodeID, prePrepare *PBFTMessage) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	node.State = PBFTStatePrepare
	node.Sequence = prePrepare.Sequence
	node.PrepareCount[prePrepare.Digest]++

	prepare := &PBFTMessage{
		ID:        uuid.New().String(),
		Type:      "PREPARE",
		View:      prePrepare.View,
		Sequence:  prePrepare.Sequence,
		Digest:    prePrepare.Digest,
		From:      nodeID,
		Timestamp: time.Now(),
	}

	if node.IsByzantine && node.ByzBehavior == types.ByzantineEquivocate {
		prepare.Digest = s.computeDigest([]byte("fake"))
	}

	node.MessageLog = append(node.MessageLog, prepare)
	s.queueMessage(prepare, "")

	s.EmitEvent("prepare", nodeID, "", map[string]interface{}{
		"view":     prepare.View,
		"sequence": prepare.Sequence,
		"digest":   prepare.Digest,
	})
}

func (s *PBFTSimulator) handlePrepareReceive(nodeID types.NodeID, prepare *PBFTMessage) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	if node.State == PBFTStatePrePrepare {
		node.State = PBFTStatePrepare
	}
	if prepare.Sequence > node.Sequence {
		node.Sequence = prepare.Sequence
	}
	node.PrepareCount[prepare.Digest]++

	if node.PrepareCount[prepare.Digest] >= 2*s.faultCount+1 && node.State == PBFTStatePrepare {
		s.handleCommit(nodeID, prepare)
	}
}

func (s *PBFTSimulator) handleCommit(nodeID types.NodeID, prepare *PBFTMessage) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	node.State = PBFTStateCommit
	node.Sequence = prepare.Sequence
	node.CommitCount[prepare.Digest]++

	commit := &PBFTMessage{
		ID:        uuid.New().String(),
		Type:      "COMMIT",
		View:      prepare.View,
		Sequence:  prepare.Sequence,
		Digest:    prepare.Digest,
		From:      nodeID,
		Timestamp: time.Now(),
	}

	node.MessageLog = append(node.MessageLog, commit)
	s.queueMessage(commit, "")

	s.EmitEvent("commit", nodeID, "", map[string]interface{}{
		"view":     commit.View,
		"sequence": commit.Sequence,
		"digest":   commit.Digest,
	})
}

func (s *PBFTSimulator) handleCommitReceive(nodeID types.NodeID, commit *PBFTMessage) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	if commit.Sequence > node.Sequence {
		node.Sequence = commit.Sequence
	}
	node.CommitCount[commit.Digest]++

	if node.CommitCount[commit.Digest] >= 2*s.faultCount+1 && node.State == PBFTStateCommit {
		s.handleReply(nodeID, commit)
	}
}

func (s *PBFTSimulator) handleReply(nodeID types.NodeID, commit *PBFTMessage) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	node.State = PBFTStateReply
	node.Sequence = commit.Sequence
	node.CommittedSeq = commit.Sequence

	alreadyCommitted := false
	for _, digest := range s.committed {
		if digest == commit.Digest {
			alreadyCommitted = true
			break
		}
	}

	if !alreadyCommitted {
		s.committed = append(s.committed, commit.Digest)
		s.SetGlobalData("committed_count", len(s.committed))
		s.SetGlobalData("result_height", len(s.committed))

		s.EmitEvent("committed", nodeID, "", map[string]interface{}{
			"sequence": commit.Sequence,
			"digest":   commit.Digest,
		})
	}

	node.State = PBFTStateIdle
	node.PrepareCount = make(map[string]int)
	node.CommitCount = make(map[string]int)
	if len(s.pendingReqs) > 0 {
		s.pendingReqs = s.pendingReqs[1:]
	}
}

func (s *PBFTSimulator) queueMessage(msg *PBFTMessage, target types.NodeID) {
	cloned := *msg
	cloned.To = target
	s.messageQueue = append(s.messageQueue, &cloned)
}

func (s *PBFTSimulator) updateNodeStates() {
	for nodeID, node := range s.nodes {
		currentPrepareVotes := 0
		for _, count := range node.PrepareCount {
			if count > currentPrepareVotes {
				currentPrepareVotes = count
			}
		}

		currentCommitVotes := 0
		for _, count := range node.CommitCount {
			if count > currentCommitVotes {
				currentCommitVotes = count
			}
		}

		s.SetNodeState(nodeID, &types.NodeState{
			ID:          nodeID,
			Status:      string(node.State),
			IsByzantine: node.IsByzantine,
			Data: map[string]interface{}{
				"is_primary":    node.IsPrimary,
				"view":          node.View,
				"sequence":      node.Sequence,
				"last_log_index": len(node.MessageLog),
				"log_length":     len(node.MessageLog),
				"committed_seq":  node.CommittedSeq,
				"prepare_count":  currentPrepareVotes,
				"commit_count":   currentCommitVotes,
			},
		})
	}

	s.syncTeachingState()
}

func (s *PBFTSimulator) syncTeachingState() {
	primaryID := s.getPrimaryID()
	primary := s.nodes[primaryID]
	stage := "idle"
	summary := "当前系统正在等待新的客户端请求。"
	nextHint := "可以发起新的客户端请求，观察提案广播如何开始。"
	progress := 0.0

	if primary != nil {
		stage = string(primary.State)
		switch primary.State {
		case PBFTStatePrePrepare:
			summary = "主节点已经广播预准备消息，副本节点将开始对同一请求达成准备共识。"
			nextHint = "继续观察 Prepare 票数是否逐步达到法定阈值。"
			progress = 25
		case PBFTStatePrepare:
			summary = "副本节点正在汇聚 Prepare 票，系统正在判断是否能够进入 Commit 阶段。"
			nextHint = "重点看是否存在故障节点导致票数不足或消息延迟。"
			progress = 50
		case PBFTStateCommit:
			summary = "系统已经进入 Commit 阶段，节点正在确认请求是否可以最终提交。"
			nextHint = "继续观察 Commit 票数是否达到阈值，以及是否形成链上结果。"
			progress = 75
		case PBFTStateReply:
			summary = "当前轮请求已经完成提交，系统正在向客户端返回结果。"
			nextHint = "观察本轮结果是否成功落地，以及是否自然进入下一轮请求。"
			progress = 100
		}
	}

	setConsensusTeachingState(
		s.BaseSimulator,
		stage,
		summary,
		nextHint,
		progress,
		map[string]interface{}{
			"current_actor":   string(primaryID),
			"result_height":   len(s.committed),
			"request_count":   s.currentSeq,
			"committed_count": len(s.committed),
		},
	)
}

func (s *PBFTSimulator) getPrimaryID() types.NodeID {
	if len(s.nodeList) == 0 {
		return ""
	}
	idx := int(s.currentView) % len(s.nodeList)
	return s.nodeList[idx]
}

func (s *PBFTSimulator) computeDigest(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *PBFTSimulator) InjectFault(fault *types.Fault) error {
	if err := s.BaseSimulator.InjectFault(fault); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if node, ok := s.nodes[fault.Target]; ok {
		node.IsByzantine = true
		if behavior, ok := fault.Params["behavior"].(string); ok {
			node.ByzBehavior = types.ByzantineBehavior(behavior)
		}
		s.updateNodeStates()
		s.EmitEvent("fault_injected", fault.Target, "", map[string]interface{}{
			"fault_type": fault.Type,
			"behavior":   node.ByzBehavior,
		})
	}

	return nil
}

func (s *PBFTSimulator) TriggerViewChange() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentView++
	s.SetGlobalData("view", s.currentView)

	for _, node := range s.nodes {
		node.View = s.currentView
		node.IsPrimary = false
	}

	newPrimaryID := s.getPrimaryID()
	s.SetGlobalData("current_actor", newPrimaryID)
	if newPrimary := s.nodes[newPrimaryID]; newPrimary != nil {
		newPrimary.IsPrimary = true
	}

	s.EmitEvent("view_change", "", "", map[string]interface{}{
		"new_view":    s.currentView,
		"new_primary": newPrimaryID,
	})
}

type PBFTFactory struct{}

func (f *PBFTFactory) Create() engine.Simulator {
	return NewPBFTSimulator()
}

func (f *PBFTFactory) GetDescription() types.Description {
	return NewPBFTSimulator().GetDescription()
}

func NewPBFTFactory() *PBFTFactory {
	return &PBFTFactory{}
}

var _ engine.SimulatorFactory = (*PBFTFactory)(nil)

// ExecuteAction 执行 PBFT 教学动作。
func (s *PBFTSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "submit_request":
		payload := []byte(fmt.Sprintf("manual-request-%d", time.Now().UnixNano()))
		if raw, ok := params["payload"].(string); ok && raw != "" {
			payload = []byte(raw)
		}
		s.mu.Lock()
		s.submitRequest(payload)
		s.updateNodeStates()
		s.mu.Unlock()
		return consensusActionResult(
			"已发起新的客户端请求",
			nil,
			&types.ActionFeedback{
				Summary:     "客户端请求已进入 PBFT 共识流程。",
				NextHint:    "观察主节点广播预准备消息，以及副本节点如何进入准备阶段。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "request_submitted"},
			},
		), nil
	case "trigger_view_change":
		s.TriggerViewChange()
		s.mu.Lock()
		s.updateNodeStates()
		s.mu.Unlock()
		return consensusActionResult(
			"已触发 PBFT 视图切换",
			nil,
			&types.ActionFeedback{
				Summary:     "系统已切换到新的视图编号。",
				NextHint:    "观察新的主节点是否接管提案，以及提交进度是否恢复推进。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "view_changed"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported pbft action: %s", action)
	}
}
