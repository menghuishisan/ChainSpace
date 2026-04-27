package transport

import (
	"context"
	"time"

	"github.com/chainspace/simulations/pkg/types"
)

// Transport 传输层接口
type Transport interface {
	// 生命周期
	Start(ctx context.Context) error
	Stop() error

	// 消息发送
	Send(msg *types.Message) error
	Broadcast(msg *types.Broadcast) error

	// 消息接收
	Receive(nodeID types.NodeID) <-chan *types.Message

	// 节点管理
	AddNode(nodeID types.NodeID) error
	RemoveNode(nodeID types.NodeID) error
	GetNodes() []types.NodeID

	// 网络模拟
	SetLatency(from, to types.NodeID, latency time.Duration)
	SetPacketLoss(from, to types.NodeID, rate float64)
	CreatePartition(groups [][]types.NodeID)
	HealPartition()

	// 状态
	GetStats() TransportStats
}

// TransportStats 传输统计
type TransportStats struct {
	MessagesSent     uint64                               `json:"messages_sent"`
	MessagesReceived uint64                               `json:"messages_received"`
	MessagesDropped  uint64                               `json:"messages_dropped"`
	BytesSent        uint64                               `json:"bytes_sent"`
	BytesReceived    uint64                               `json:"bytes_received"`
	AvgLatency       time.Duration                        `json:"avg_latency"`
	NodeStats        map[types.NodeID]*NodeTransportStats `json:"node_stats"`
}

// NodeTransportStats 节点传输统计
type NodeTransportStats struct {
	Sent       uint64        `json:"sent"`
	Received   uint64        `json:"received"`
	Dropped    uint64        `json:"dropped"`
	AvgLatency time.Duration `json:"avg_latency"`
}

// TransportConfig 传输配置
type TransportConfig struct {
	Mode           types.RunMode           `json:"mode"`
	DefaultLatency time.Duration           `json:"default_latency"`
	DefaultLoss    float64                 `json:"default_loss"`
	BufferSize     int                     `json:"buffer_size"`
	NodeAddresses  map[types.NodeID]string `json:"node_addresses"` // Real模式使用
}

// NewTransport 创建传输层
func NewTransport(config TransportConfig) Transport {
	switch config.Mode {
	case types.ModeReal:
		return NewRealTransport(config)
	default:
		return NewSimulatedTransport(config)
	}
}

// MessageHandler 消息处理器
type MessageHandler func(msg *types.Message) error

// BroadcastHandler 广播处理器
type BroadcastHandler func(msg *types.Broadcast) error

// Node 节点接口
type Node interface {
	ID() types.NodeID
	Send(to types.NodeID, msg *types.Message) error
	Broadcast(msg *types.Broadcast) error
	Receive() <-chan *types.Message
	Start(ctx context.Context) error
	Stop() error
}

// BaseNode 基础节点实现
type BaseNode struct {
	id        types.NodeID
	transport Transport
	inbox     chan *types.Message
	handlers  map[string]MessageHandler
}

// NewBaseNode 创建基础节点
func NewBaseNode(id types.NodeID, transport Transport) *BaseNode {
	return &BaseNode{
		id:        id,
		transport: transport,
		inbox:     make(chan *types.Message, 1000),
		handlers:  make(map[string]MessageHandler),
	}
}

// ID 获取节点ID
func (n *BaseNode) ID() types.NodeID {
	return n.id
}

// Send 发送消息
func (n *BaseNode) Send(to types.NodeID, msgType string, payload []byte) error {
	msg := &types.Message{
		Type:      msgType,
		From:      n.id,
		To:        to,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return n.transport.Send(msg)
}

// Broadcast 广播消息
func (n *BaseNode) Broadcast(msgType string, payload []byte) error {
	broadcast := &types.Broadcast{
		Type:      msgType,
		From:      n.id,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return n.transport.Broadcast(broadcast)
}

// Receive 接收消息通道
func (n *BaseNode) Receive() <-chan *types.Message {
	return n.transport.Receive(n.id)
}

// RegisterHandler 注册消息处理器
func (n *BaseNode) RegisterHandler(msgType string, handler MessageHandler) {
	n.handlers[msgType] = handler
}

// ProcessMessage 处理消息
func (n *BaseNode) ProcessMessage(msg *types.Message) error {
	handler, ok := n.handlers[msg.Type]
	if !ok {
		return nil // 没有处理器，忽略消息
	}
	return handler(msg)
}

// Start 启动节点
func (n *BaseNode) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-n.Receive():
				n.ProcessMessage(msg)
			}
		}
	}()
	return nil
}
