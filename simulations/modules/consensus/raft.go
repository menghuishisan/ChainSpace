package consensus

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

type RaftRole string

const (
	RaftRoleFollower  RaftRole = "follower"
	RaftRoleCandidate RaftRole = "candidate"
	RaftRoleLeader    RaftRole = "leader"
)

type RaftNode struct {
	ID               types.NodeID            `json:"id"`
	Role             RaftRole                `json:"role"`
	Term             uint64                  `json:"term"`
	VotedFor         types.NodeID            `json:"voted_for"`
	Log              []*RaftLogEntry         `json:"log"`
	CommitIndex      uint64                  `json:"commit_index"`
	LastApplied      uint64                  `json:"last_applied"`
	NextIndex        map[types.NodeID]uint64 `json:"next_index"`
	MatchIndex       map[types.NodeID]uint64 `json:"match_index"`
	VotesGranted     int                     `json:"votes_granted"`
	ElectionTimeout  int                     `json:"election_timeout"`
	HeartbeatTimeout int                     `json:"heartbeat_timeout"`
	TimeoutCounter   int                     `json:"timeout_counter"`
	IsOnline         bool                    `json:"is_online"`
}

type RaftLogEntry struct {
	Term    uint64      `json:"term"`
	Index   uint64      `json:"index"`
	Command interface{} `json:"command"`
}

type RaftMessage struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Term         uint64          `json:"term"`
	From         types.NodeID    `json:"from"`
	To           types.NodeID    `json:"to"`
	LeaderID     types.NodeID    `json:"leader_id,omitempty"`
	PrevLogIndex uint64          `json:"prev_log_index,omitempty"`
	PrevLogTerm  uint64          `json:"prev_log_term,omitempty"`
	Entries      []*RaftLogEntry `json:"entries,omitempty"`
	LeaderCommit uint64          `json:"leader_commit,omitempty"`
	LastLogIndex uint64          `json:"last_log_index,omitempty"`
	LastLogTerm  uint64          `json:"last_log_term,omitempty"`
	VoteGranted  bool            `json:"vote_granted,omitempty"`
	Success      bool            `json:"success,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
}

type RaftSimulator struct {
	*base.BaseSimulator
	mu           sync.RWMutex
	nodes        map[types.NodeID]*RaftNode
	nodeList     []types.NodeID
	nodeCount    int
	leaderID     types.NodeID
	messageQueue []*RaftMessage
	committed    []interface{}
}

func NewRaftSimulator() *RaftSimulator {
	sim := &RaftSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"raft",
			"Raft 共识算法",
			"Raft 一致性算法的教学模拟器，支持领导者选举、日志复制和故障恢复。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:        make(map[types.NodeID]*RaftNode),
		nodeList:     make([]types.NodeID, 0),
		messageQueue: make([]*RaftMessage, 0),
		committed:    make([]interface{}, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "Raft 集群中的节点数量。",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         3,
		Max:         11,
	})
	sim.AddParam(types.Param{
		Key:         "election_timeout_min",
		Name:        "选举超时下限(tick)",
		Description: "选举超时的最小值。",
		Type:        types.ParamTypeInt,
		Default:     15,
		Min:         5,
		Max:         50,
	})
	sim.AddParam(types.Param{
		Key:         "election_timeout_max",
		Name:        "选举超时上限(tick)",
		Description: "选举超时的最大值。",
		Type:        types.ParamTypeInt,
		Default:     30,
		Min:         10,
		Max:         100,
	})
	sim.AddParam(types.Param{
		Key:         "heartbeat_interval",
		Name:        "心跳间隔(tick)",
		Description: "Leader 发送心跳的间隔。",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         1,
		Max:         20,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *RaftSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	nodeCount := 5
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	s.nodeCount = nodeCount
	s.nodes = make(map[types.NodeID]*RaftNode)
	s.nodeList = make([]types.NodeID, 0, nodeCount)
	s.leaderID = ""
	s.messageQueue = make([]*RaftMessage, 0)
	s.committed = make([]interface{}, 0)

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("node-%d", i))
		node := &RaftNode{
			ID:               nodeID,
			Role:             RaftRoleFollower,
			Term:             0,
			Log:              make([]*RaftLogEntry, 0),
			CommitIndex:      0,
			LastApplied:      0,
			NextIndex:        make(map[types.NodeID]uint64),
			MatchIndex:       make(map[types.NodeID]uint64),
			ElectionTimeout:  s.randomElectionTimeout(),
			HeartbeatTimeout: 5,
			IsOnline:         true,
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
		s.updateNodeState(nodeID)
	}

	s.SetGlobalData("term", uint64(0))
	s.SetGlobalData("leader", "")
	s.SetGlobalData("request_count", 0)
	s.SetGlobalData("committed_count", 0)
	s.SetGlobalData("committee_size", s.nodeCount)
	s.SetGlobalData("current_actor", "")
	s.SetGlobalData("result_height", 0)
	setConsensusTeachingState(
		s.BaseSimulator,
		"raft_idle",
		"当前 Raft 集群处于待命状态，可以触发选举或提交日志观察主从复制过程。",
		"先等待或手动触发一次选举，观察候选者如何获得多数并成为 Leader。",
		0,
		map[string]interface{}{"term": 0, "leader": "", "result_height": 0},
	)

	return nil
}

func (s *RaftSimulator) randomElectionTimeout() int {
	return 15 + rand.Intn(16)
}

func (s *RaftSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processMessages()

	for _, nodeID := range s.nodeList {
		node := s.nodes[nodeID]
		if !node.IsOnline {
			continue
		}

		node.TimeoutCounter++
		switch node.Role {
		case RaftRoleFollower, RaftRoleCandidate:
			if node.TimeoutCounter >= node.ElectionTimeout {
				s.startElection(nodeID)
			}
		case RaftRoleLeader:
			if node.TimeoutCounter >= node.HeartbeatTimeout {
				s.sendHeartbeats(nodeID)
				node.TimeoutCounter = 0
			}
		}
	}

	s.updateAllNodeStates()
	return nil
}

func (s *RaftSimulator) startElection(nodeID types.NodeID) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	node.Term++
	node.Role = RaftRoleCandidate
	node.VotedFor = nodeID
	node.VotesGranted = 1
	node.TimeoutCounter = 0
	node.ElectionTimeout = s.randomElectionTimeout()

	s.EmitEvent("election_started", nodeID, "", map[string]interface{}{
		"term": node.Term,
	})

	lastLogIndex := uint64(len(node.Log))
	lastLogTerm := uint64(0)
	if len(node.Log) > 0 {
		lastLogTerm = node.Log[len(node.Log)-1].Term
	}

	for _, targetID := range s.nodeList {
		if targetID == nodeID {
			continue
		}

		msg := &RaftMessage{
			ID:           uuid.New().String(),
			Type:         "RequestVote",
			Term:         node.Term,
			From:         nodeID,
			To:           targetID,
			LastLogIndex: lastLogIndex,
			LastLogTerm:  lastLogTerm,
			Timestamp:    time.Now(),
		}
		s.messageQueue = append(s.messageQueue, msg)
		s.EmitEvent("request_vote", nodeID, targetID, map[string]interface{}{
			"term": node.Term,
		})
	}
}

func (s *RaftSimulator) processMessages() {
	if len(s.messageQueue) == 0 {
		return
	}

	msg := s.messageQueue[0]
	s.messageQueue = s.messageQueue[1:]
	targetNode := s.nodes[msg.To]
	if targetNode == nil || !targetNode.IsOnline {
		return
	}

	switch msg.Type {
	case "RequestVote":
		s.handleRequestVote(msg)
	case "RequestVoteReply":
		s.handleRequestVoteReply(msg)
	case "AppendEntries":
		s.handleAppendEntries(msg)
	case "AppendEntriesReply":
		s.handleAppendEntriesReply(msg)
	}
}

func (s *RaftSimulator) handleRequestVote(msg *RaftMessage) {
	node := s.nodes[msg.To]
	if node == nil {
		return
	}

	voteGranted := false
	if msg.Term > node.Term {
		node.Term = msg.Term
		node.Role = RaftRoleFollower
		node.VotedFor = ""
	}

	if msg.Term >= node.Term && (node.VotedFor == "" || node.VotedFor == msg.From) {
		lastLogIndex := uint64(len(node.Log))
		lastLogTerm := uint64(0)
		if len(node.Log) > 0 {
			lastLogTerm = node.Log[len(node.Log)-1].Term
		}
		if msg.LastLogTerm > lastLogTerm || (msg.LastLogTerm == lastLogTerm && msg.LastLogIndex >= lastLogIndex) {
			voteGranted = true
			node.VotedFor = msg.From
			node.TimeoutCounter = 0
		}
	}

	reply := &RaftMessage{
		ID:          uuid.New().String(),
		Type:        "RequestVoteReply",
		Term:        node.Term,
		From:        msg.To,
		To:          msg.From,
		VoteGranted: voteGranted,
		Timestamp:   time.Now(),
	}
	s.messageQueue = append(s.messageQueue, reply)
	s.EmitEvent("vote_response", msg.To, msg.From, map[string]interface{}{
		"term":         node.Term,
		"vote_granted": voteGranted,
	})
}

func (s *RaftSimulator) handleRequestVoteReply(msg *RaftMessage) {
	node := s.nodes[msg.To]
	if node == nil || node.Role != RaftRoleCandidate {
		return
	}

	if msg.Term > node.Term {
		node.Term = msg.Term
		node.Role = RaftRoleFollower
		node.VotedFor = ""
		return
	}

	if msg.VoteGranted && msg.Term == node.Term {
		node.VotesGranted++
		if node.VotesGranted > s.nodeCount/2 {
			s.becomeLeader(msg.To)
		}
	}
}

func (s *RaftSimulator) becomeLeader(nodeID types.NodeID) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	node.Role = RaftRoleLeader
	node.TimeoutCounter = 0
	s.leaderID = nodeID

	lastLogIndex := uint64(len(node.Log))
	for _, id := range s.nodeList {
		node.NextIndex[id] = lastLogIndex + 1
		node.MatchIndex[id] = 0
	}

	s.SetGlobalData("leader", string(nodeID))
	s.SetGlobalData("term", node.Term)
	s.SetGlobalData("current_actor", string(nodeID))

	s.EmitEvent("leader_elected", nodeID, "", map[string]interface{}{
		"term": node.Term,
	})

	s.sendHeartbeats(nodeID)
}

func (s *RaftSimulator) dropPendingAppendEntries(leaderID types.NodeID) {
	filtered := make([]*RaftMessage, 0, len(s.messageQueue))
	for _, msg := range s.messageQueue {
		if msg.Type == "AppendEntries" && msg.From == leaderID {
			continue
		}
		filtered = append(filtered, msg)
	}
	s.messageQueue = filtered
}

func (s *RaftSimulator) sendHeartbeats(leaderID types.NodeID) {
	leader := s.nodes[leaderID]
	if leader == nil || leader.Role != RaftRoleLeader {
		return
	}

	for _, targetID := range s.nodeList {
		if targetID == leaderID {
			continue
		}

		prevLogIndex := leader.NextIndex[targetID] - 1
		prevLogTerm := uint64(0)
		if prevLogIndex > 0 && int(prevLogIndex) <= len(leader.Log) {
			prevLogTerm = leader.Log[prevLogIndex-1].Term
		}

		var entries []*RaftLogEntry
		if int(leader.NextIndex[targetID]) <= len(leader.Log) {
			entries = leader.Log[leader.NextIndex[targetID]-1:]
		}

		msg := &RaftMessage{
			ID:           uuid.New().String(),
			Type:         "AppendEntries",
			Term:         leader.Term,
			From:         leaderID,
			To:           targetID,
			LeaderID:     leaderID,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: leader.CommitIndex,
			Timestamp:    time.Now(),
		}
		s.messageQueue = append(s.messageQueue, msg)
		s.EmitEvent("append_entries", leaderID, targetID, map[string]interface{}{
			"term":        leader.Term,
			"entry_count": len(entries),
		})
	}
}

func (s *RaftSimulator) handleAppendEntries(msg *RaftMessage) {
	node := s.nodes[msg.To]
	if node == nil {
		return
	}

	success := false
	if msg.Term >= node.Term {
		node.Term = msg.Term
		node.Role = RaftRoleFollower
		node.VotedFor = ""
		node.TimeoutCounter = 0

		if msg.PrevLogIndex == 0 || (int(msg.PrevLogIndex) <= len(node.Log) && node.Log[msg.PrevLogIndex-1].Term == msg.PrevLogTerm) {
			success = true
			if len(msg.Entries) > 0 {
				node.Log = append(node.Log[:msg.PrevLogIndex], msg.Entries...)
			}
			if msg.LeaderCommit > node.CommitIndex {
				lastNewEntry := uint64(len(node.Log))
				if msg.LeaderCommit < lastNewEntry {
					node.CommitIndex = msg.LeaderCommit
				} else {
					node.CommitIndex = lastNewEntry
				}
			}
		}
	}

	reply := &RaftMessage{
		ID:        uuid.New().String(),
		Type:      "AppendEntriesReply",
		Term:      node.Term,
		From:      msg.To,
		To:        msg.From,
		Success:   success,
		Timestamp: time.Now(),
	}
	s.messageQueue = append(s.messageQueue, reply)
}

func (s *RaftSimulator) handleAppendEntriesReply(msg *RaftMessage) {
	node := s.nodes[msg.To]
	if node == nil || node.Role != RaftRoleLeader {
		return
	}

	if msg.Term > node.Term {
		node.Term = msg.Term
		node.Role = RaftRoleFollower
		node.VotedFor = ""
		s.leaderID = ""
		s.SetGlobalData("leader", "")
		s.SetGlobalData("current_actor", "")
		return
	}

	if msg.Success {
		node.NextIndex[msg.From] = uint64(len(node.Log)) + 1
		node.MatchIndex[msg.From] = uint64(len(node.Log))
		latestIndex := uint64(len(node.Log))
		replicated := 1
		for _, targetID := range s.nodeList {
			if targetID == msg.To {
				continue
			}
			if node.MatchIndex[targetID] >= latestIndex {
				replicated++
			}
		}
		if latestIndex > 0 && replicated > s.nodeCount/2 && node.CommitIndex < latestIndex {
			node.CommitIndex = latestIndex
			s.SetGlobalData("committed_count", int(node.CommitIndex))
			s.SetGlobalData("result_height", int(node.CommitIndex))
			s.dropPendingAppendEntries(msg.To)
			s.sendHeartbeats(msg.To)
		}
	} else {
		if node.NextIndex[msg.From] > 1 {
			node.NextIndex[msg.From]--
		}
	}
}

func (s *RaftSimulator) SubmitCommand(command interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.leaderID == "" {
		return fmt.Errorf("no leader")
	}
	leader := s.nodes[s.leaderID]
	if leader == nil {
		return fmt.Errorf("leader not found")
	}

	entry := &RaftLogEntry{
		Term:    leader.Term,
		Index:   uint64(len(leader.Log)) + 1,
		Command: command,
	}
	leader.Log = append(leader.Log, entry)

	s.EmitEvent("command_submitted", s.leaderID, "", map[string]interface{}{
		"index":   entry.Index,
		"command": command,
	})
	s.SetGlobalData("request_count", entry.Index)
	s.SetGlobalData("current_actor", string(s.leaderID))
	s.dropPendingAppendEntries(s.leaderID)
	s.sendHeartbeats(s.leaderID)
	return nil
}

func (s *RaftSimulator) updateNodeState(nodeID types.NodeID) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	s.SetNodeState(nodeID, &types.NodeState{
		ID:     nodeID,
		Status: string(node.Role),
		Data: map[string]interface{}{
			"term":           node.Term,
			"voted_for":      node.VotedFor,
			"last_log_index": len(node.Log),
			"log_length":     len(node.Log),
			"commit_index":   node.CommitIndex,
			"votes_granted":  node.VotesGranted,
			"next_index":     node.NextIndex,
			"match_index":    node.MatchIndex,
			"is_online":      node.IsOnline,
		},
	})
}

func (s *RaftSimulator) updateAllNodeStates() {
	for nodeID := range s.nodes {
		s.updateNodeState(nodeID)
	}
}

func (s *RaftSimulator) SetNodeOnline(nodeID types.NodeID, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if node := s.nodes[nodeID]; node != nil {
		node.IsOnline = online
		if !online && s.leaderID == nodeID {
			s.leaderID = ""
			s.SetGlobalData("leader", "")
			s.SetGlobalData("current_actor", "")
		}
		s.EmitEvent("node_status_changed", nodeID, "", map[string]interface{}{
			"online": online,
		})
	}
}

type RaftFactory struct{}

func (f *RaftFactory) Create() engine.Simulator {
	return NewRaftSimulator()
}

func (f *RaftFactory) GetDescription() types.Description {
	return NewRaftSimulator().GetDescription()
}

func NewRaftFactory() *RaftFactory {
	return &RaftFactory{}
}

var _ engine.SimulatorFactory = (*RaftFactory)(nil)

// ExecuteAction 执行 Raft 教学动作。
func (s *RaftSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "submit_command":
		command := "set x=1"
		if raw, ok := params["command"].(string); ok && raw != "" {
			command = raw
		}
		if err := s.SubmitCommand(command); err != nil {
			return nil, err
		}
		s.updateAllNodeStates()
		return consensusActionResult(
			"已提交新的 Raft 日志命令",
			nil,
			&types.ActionFeedback{
				Summary:     "新的命令已经提交给当前领导者。",
				NextHint:    "观察日志复制是否传播到跟随者，以及提交索引是否继续推进。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "command_submitted"},
			},
		), nil
	case "trigger_election":
		target := types.NodeID("node-0")
		if raw, ok := params["target"].(string); ok && raw != "" {
			target = types.NodeID(raw)
		}
		s.mu.Lock()
		s.startElection(target)
		s.updateAllNodeStates()
		s.mu.Unlock()
		return consensusActionResult(
			"已触发新的选主流程",
			nil,
			&types.ActionFeedback{
				Summary:     "目标节点已开始新的选举。",
				NextHint:    "观察候选者是否获得多数投票并成为新的领导者。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "election_triggered"},
			},
		), nil
	case "simulate_leader_failure":
		if s.leaderID == "" {
			return nil, fmt.Errorf("no leader to fail")
		}
		s.SetNodeOnline(s.leaderID, false)
		s.updateAllNodeStates()
		return consensusActionResult(
			"已让当前领导者离线",
			nil,
			&types.ActionFeedback{
				Summary:     "当前领导者已失效，复制流程会受到影响。",
				NextHint:    "观察是否会重新选主，以及跟随者日志复制是否停滞。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "leader_failed"},
			},
		), nil
	case "restore_node":
		target := types.NodeID("node-0")
		if raw, ok := params["target"].(string); ok && raw != "" {
			target = types.NodeID(raw)
		}
		s.SetNodeOnline(target, true)
		s.updateAllNodeStates()
		return consensusActionResult(
			"已恢复节点在线状态",
			nil,
			&types.ActionFeedback{
				Summary:     "目标节点重新加入集群。",
				NextHint:    "观察该节点是否追上最新日志，并重新参与复制或投票。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "node_restored"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported raft action: %s", action)
	}
}
