package transport

import (
	"context"
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

// SimulatedTransport 模拟传输层（内存队列）
type SimulatedTransport struct {
	mu         sync.RWMutex
	nodes      map[types.NodeID]*nodeChannel
	latencies  map[string]time.Duration // "from->to" -> latency
	packetLoss map[string]float64       // "from->to" -> loss rate
	partitions [][]types.NodeID
	config     TransportConfig
	stats      TransportStats
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
}

// nodeChannel 节点通道
type nodeChannel struct {
	inbox  chan *types.Message
	online bool
}

// NewSimulatedTransport 创建模拟传输层
func NewSimulatedTransport(config TransportConfig) *SimulatedTransport {
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	return &SimulatedTransport{
		nodes:      make(map[types.NodeID]*nodeChannel),
		latencies:  make(map[string]time.Duration),
		packetLoss: make(map[string]float64),
		config:     config,
		stats: TransportStats{
			NodeStats: make(map[types.NodeID]*NodeTransportStats),
		},
	}
}

// Start 启动传输层
func (t *SimulatedTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ctx, t.cancel = context.WithCancel(ctx)
	t.running = true
	return nil
}

// Stop 停止传输层
func (t *SimulatedTransport) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}
	t.running = false

	// 关闭所有通道
	for _, nc := range t.nodes {
		close(nc.inbox)
	}
	return nil
}

// AddNode 添加节点
func (t *SimulatedTransport) AddNode(nodeID types.NodeID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.nodes[nodeID]; exists {
		return nil
	}

	t.nodes[nodeID] = &nodeChannel{
		inbox:  make(chan *types.Message, t.config.BufferSize),
		online: true,
	}
	t.stats.NodeStats[nodeID] = &NodeTransportStats{}
	return nil
}

// RemoveNode 移除节点
func (t *SimulatedTransport) RemoveNode(nodeID types.NodeID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if nc, exists := t.nodes[nodeID]; exists {
		close(nc.inbox)
		delete(t.nodes, nodeID)
		delete(t.stats.NodeStats, nodeID)
	}
	return nil
}

// GetNodes 获取所有节点
func (t *SimulatedTransport) GetNodes() []types.NodeID {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var nodes []types.NodeID
	for id := range t.nodes {
		nodes = append(nodes, id)
	}
	return nodes
}

// Send 发送消息
func (t *SimulatedTransport) Send(msg *types.Message) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.running {
		return nil
	}

	// 检查节点是否存在
	targetNode, exists := t.nodes[msg.To]
	if !exists || !targetNode.online {
		return nil
	}

	// 检查网络分区
	if t.isPartitioned(msg.From, msg.To) {
		t.stats.MessagesDropped++
		return nil
	}

	// 检查丢包
	lossKey := string(msg.From) + "->" + string(msg.To)
	if loss, ok := t.packetLoss[lossKey]; ok && rand.Float64() < loss {
		t.stats.MessagesDropped++
		if stats, ok := t.stats.NodeStats[msg.From]; ok {
			stats.Dropped++
		}
		return nil
	}

	// 设置消息ID
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// 计算延迟
	latency := t.config.DefaultLatency
	latencyKey := string(msg.From) + "->" + string(msg.To)
	if l, ok := t.latencies[latencyKey]; ok {
		latency = l
	}

	// 异步发送（模拟延迟）
	go func() {
		if latency > 0 {
			time.Sleep(latency)
		}

		t.mu.RLock()
		nc, exists := t.nodes[msg.To]
		t.mu.RUnlock()

		if exists && nc.online {
			select {
			case nc.inbox <- msg:
				t.mu.Lock()
				t.stats.MessagesSent++
				if stats, ok := t.stats.NodeStats[msg.From]; ok {
					stats.Sent++
				}
				if stats, ok := t.stats.NodeStats[msg.To]; ok {
					stats.Received++
				}
				msgData, _ := json.Marshal(msg)
				t.stats.BytesSent += uint64(len(msgData))
				t.mu.Unlock()
			default:
				// 通道满，丢弃消息
				t.mu.Lock()
				t.stats.MessagesDropped++
				t.mu.Unlock()
			}
		}
	}()

	return nil
}

// Broadcast 广播消息
func (t *SimulatedTransport) Broadcast(msg *types.Broadcast) error {
	t.mu.RLock()
	nodes := make([]types.NodeID, 0, len(t.nodes))
	for id := range t.nodes {
		if id != msg.From {
			nodes = append(nodes, id)
		}
	}
	t.mu.RUnlock()

	// 设置消息ID
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// 向所有节点发送
	for _, nodeID := range nodes {
		unicast := &types.Message{
			ID:        msg.ID + "-" + string(nodeID),
			Type:      msg.Type,
			From:      msg.From,
			To:        nodeID,
			Payload:   msg.Payload,
			Timestamp: msg.Timestamp,
		}
		t.Send(unicast)
	}

	return nil
}

// Receive 接收消息
func (t *SimulatedTransport) Receive(nodeID types.NodeID) <-chan *types.Message {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if nc, exists := t.nodes[nodeID]; exists {
		return nc.inbox
	}
	return nil
}

// SetLatency 设置延迟
func (t *SimulatedTransport) SetLatency(from, to types.NodeID, latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := string(from) + "->" + string(to)
	t.latencies[key] = latency
}

// SetPacketLoss 设置丢包率
func (t *SimulatedTransport) SetPacketLoss(from, to types.NodeID, rate float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := string(from) + "->" + string(to)
	t.packetLoss[key] = rate
}

// CreatePartition 创建网络分区
func (t *SimulatedTransport) CreatePartition(groups [][]types.NodeID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.partitions = groups
}

// HealPartition 恢复网络分区
func (t *SimulatedTransport) HealPartition() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.partitions = nil
}

// isPartitioned 检查是否被分区隔离
func (t *SimulatedTransport) isPartitioned(from, to types.NodeID) bool {
	if len(t.partitions) == 0 {
		return false
	}

	fromGroup := -1
	toGroup := -1

	for i, group := range t.partitions {
		for _, nodeID := range group {
			if nodeID == from {
				fromGroup = i
			}
			if nodeID == to {
				toGroup = i
			}
		}
	}

	// 如果在不同分区，则被隔离
	return fromGroup != -1 && toGroup != -1 && fromGroup != toGroup
}

// GetStats 获取统计信息
func (t *SimulatedTransport) GetStats() TransportStats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats
}

// SetNodeOnline 设置节点在线状态
func (t *SimulatedTransport) SetNodeOnline(nodeID types.NodeID, online bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if nc, exists := t.nodes[nodeID]; exists {
		nc.online = online
	}
}

// IsNodeOnline 检查节点是否在线
func (t *SimulatedTransport) IsNodeOnline(nodeID types.NodeID) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if nc, exists := t.nodes[nodeID]; exists {
		return nc.online
	}
	return false
}

// ResetStats 重置统计
func (t *SimulatedTransport) ResetStats() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stats = TransportStats{
		NodeStats: make(map[types.NodeID]*NodeTransportStats),
	}
	for nodeID := range t.nodes {
		t.stats.NodeStats[nodeID] = &NodeTransportStats{}
	}
}
