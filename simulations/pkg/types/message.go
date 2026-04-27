package types

import (
	"encoding/json"
	"time"
)

// ConsensusMessage 共识消息基础结构
type ConsensusMessage struct {
	Type      string          `json:"type"`
	From      NodeID          `json:"from"`
	To        NodeID          `json:"to,omitempty"`
	View      uint64          `json:"view"`
	Sequence  uint64          `json:"sequence"`
	Digest    Hash            `json:"digest"`
	Signature []byte          `json:"signature"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// PBFT消息类型
const (
	MsgTypePBFTRequest    = "pbft_request"
	MsgTypePBFTPrePrepare = "pbft_pre_prepare"
	MsgTypePBFTPrepare    = "pbft_prepare"
	MsgTypePBFTCommit     = "pbft_commit"
	MsgTypePBFTReply      = "pbft_reply"
	MsgTypePBFTViewChange = "pbft_view_change"
	MsgTypePBFTNewView    = "pbft_new_view"
	MsgTypePBFTCheckpoint = "pbft_checkpoint"
)

// PBFTPrePrepare PBFT PrePrepare消息
type PBFTPrePrepare struct {
	View     uint64 `json:"view"`
	Sequence uint64 `json:"sequence"`
	Digest   Hash   `json:"digest"`
	Request  []byte `json:"request"`
}

// PBFTPrepare PBFT Prepare消息
type PBFTPrepare struct {
	View     uint64 `json:"view"`
	Sequence uint64 `json:"sequence"`
	Digest   Hash   `json:"digest"`
	NodeID   NodeID `json:"node_id"`
}

// PBFTCommit PBFT Commit消息
type PBFTCommit struct {
	View     uint64 `json:"view"`
	Sequence uint64 `json:"sequence"`
	Digest   Hash   `json:"digest"`
	NodeID   NodeID `json:"node_id"`
}

// Raft消息类型
const (
	MsgTypeRaftRequestVote   = "raft_request_vote"
	MsgTypeRaftVoteReply     = "raft_vote_reply"
	MsgTypeRaftAppendEntries = "raft_append_entries"
	MsgTypeRaftAppendReply   = "raft_append_reply"
	MsgTypeRaftHeartbeat     = "raft_heartbeat"
)

// RaftRequestVote Raft投票请求
type RaftRequestVote struct {
	Term         uint64 `json:"term"`
	CandidateID  NodeID `json:"candidate_id"`
	LastLogIndex uint64 `json:"last_log_index"`
	LastLogTerm  uint64 `json:"last_log_term"`
}

// RaftVoteReply Raft投票响应
type RaftVoteReply struct {
	Term        uint64 `json:"term"`
	VoteGranted bool   `json:"vote_granted"`
}

// RaftAppendEntries Raft日志追加
type RaftAppendEntries struct {
	Term         uint64      `json:"term"`
	LeaderID     NodeID      `json:"leader_id"`
	PrevLogIndex uint64      `json:"prev_log_index"`
	PrevLogTerm  uint64      `json:"prev_log_term"`
	Entries      []RaftEntry `json:"entries"`
	LeaderCommit uint64      `json:"leader_commit"`
}

// RaftEntry Raft日志条目
type RaftEntry struct {
	Term    uint64          `json:"term"`
	Index   uint64          `json:"index"`
	Command json.RawMessage `json:"command"`
}

// RaftAppendReply Raft日志追加响应
type RaftAppendReply struct {
	Term         uint64 `json:"term"`
	Success      bool   `json:"success"`
	MatchIndex   uint64 `json:"match_index"`
	ConflictTerm uint64 `json:"conflict_term"`
}

// PoW消息类型
const (
	MsgTypePoWNewBlock  = "pow_new_block"
	MsgTypePoWGetBlocks = "pow_get_blocks"
	MsgTypePoWBlocks    = "pow_blocks"
)

// PoWNewBlock PoW新区块广播
type PoWNewBlock struct {
	Block *Block `json:"block"`
}

// PoS消息类型
const (
	MsgTypePoSProposal = "pos_proposal"
	MsgTypePoSVote     = "pos_vote"
	MsgTypePoSAttest   = "pos_attest"
)

// PoSProposal PoS提议
type PoSProposal struct {
	Slot      uint64 `json:"slot"`
	Proposer  NodeID `json:"proposer"`
	Block     *Block `json:"block"`
	Signature []byte `json:"signature"`
}

// PoSVote PoS投票
type PoSVote struct {
	Slot      uint64 `json:"slot"`
	BlockHash Hash   `json:"block_hash"`
	Validator NodeID `json:"validator"`
	Signature []byte `json:"signature"`
}

// NetworkMessage 网络层消息
type NetworkMessage struct {
	Type    string          `json:"type"`
	From    NodeID          `json:"from"`
	To      NodeID          `json:"to"`
	TTL     int             `json:"ttl"`
	Payload json.RawMessage `json:"payload"`
}

// 网络消息类型
const (
	MsgTypeNetPing      = "net_ping"
	MsgTypeNetPong      = "net_pong"
	MsgTypeNetFindNode  = "net_find_node"
	MsgTypeNetNeighbors = "net_neighbors"
	MsgTypeNetGossip    = "net_gossip"
)

// GossipMessage Gossip消息
type GossipMessage struct {
	ID        string          `json:"id"`
	Origin    NodeID          `json:"origin"`
	TTL       int             `json:"ttl"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Seen      map[NodeID]bool `json:"-"`
}

// MessageQueue 消息队列
type MessageQueue struct {
	messages []*Message
	size     int
}

// NewMessageQueue 创建消息队列
func NewMessageQueue(size int) *MessageQueue {
	return &MessageQueue{
		messages: make([]*Message, 0, size),
		size:     size,
	}
}

// Push 入队
func (q *MessageQueue) Push(msg *Message) bool {
	if len(q.messages) >= q.size {
		return false
	}
	q.messages = append(q.messages, msg)
	return true
}

// Pop 出队
func (q *MessageQueue) Pop() *Message {
	if len(q.messages) == 0 {
		return nil
	}
	msg := q.messages[0]
	q.messages = q.messages[1:]
	return msg
}

// Len 队列长度
func (q *MessageQueue) Len() int {
	return len(q.messages)
}

// IsEmpty 是否为空
func (q *MessageQueue) IsEmpty() bool {
	return len(q.messages) == 0
}
