package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

// RealTransport 真实传输层（TCP）
type RealTransport struct {
	mu        sync.RWMutex
	localID   types.NodeID
	localAddr string
	peers     map[types.NodeID]*peerConn
	listener  net.Listener
	inbox     chan *types.Message
	config    TransportConfig
	stats     TransportStats
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// peerConn 对等连接
type peerConn struct {
	nodeID types.NodeID
	addr   string
	conn   net.Conn
	outbox chan *types.Message
	online bool
}

// NewRealTransport 创建真实传输层
func NewRealTransport(config TransportConfig) *RealTransport {
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	return &RealTransport{
		peers:  make(map[types.NodeID]*peerConn),
		inbox:  make(chan *types.Message, config.BufferSize),
		config: config,
		stats: TransportStats{
			NodeStats: make(map[types.NodeID]*NodeTransportStats),
		},
	}
}

// Start 启动传输层
func (t *RealTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ctx, t.cancel = context.WithCancel(ctx)

	// 启动监听
	listener, err := net.Listen("tcp", t.localAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	t.listener = listener
	t.running = true

	// 接受连接
	go t.acceptLoop()

	// 连接已知节点
	for nodeID, addr := range t.config.NodeAddresses {
		if nodeID != t.localID {
			go t.connectToPeer(nodeID, addr)
		}
	}

	return nil
}

// acceptLoop 接受连接循环
func (t *RealTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.ctx.Done():
				return
			default:
				continue
			}
		}
		go t.handleConnection(conn)
	}
}

// handleConnection 处理连接
func (t *RealTransport) handleConnection(conn net.Conn) {
	decoder := json.NewDecoder(conn)
	for {
		var msg types.Message
		if err := decoder.Decode(&msg); err != nil {
			conn.Close()
			return
		}

		t.mu.Lock()
		t.stats.MessagesReceived++
		if stats, ok := t.stats.NodeStats[msg.From]; ok {
			stats.Received++
		}
		t.mu.Unlock()

		select {
		case t.inbox <- &msg:
		default:
			// 通道满，丢弃
		}
	}
}

// connectToPeer 连接到对等节点
func (t *RealTransport) connectToPeer(nodeID types.NodeID, addr string) {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		peer := &peerConn{
			nodeID: nodeID,
			addr:   addr,
			conn:   conn,
			outbox: make(chan *types.Message, t.config.BufferSize),
			online: true,
		}

		t.mu.Lock()
		t.peers[nodeID] = peer
		t.mu.Unlock()

		// 启动发送循环
		go t.sendLoop(peer)
		// 启动接收循环
		go t.recvLoop(peer)
		return
	}
}

// sendLoop 发送循环
func (t *RealTransport) sendLoop(peer *peerConn) {
	encoder := json.NewEncoder(peer.conn)
	for msg := range peer.outbox {
		if err := encoder.Encode(msg); err != nil {
			peer.online = false
			peer.conn.Close()
			return
		}
		t.mu.Lock()
		t.stats.MessagesSent++
		if stats, ok := t.stats.NodeStats[peer.nodeID]; ok {
			stats.Sent++
		}
		t.mu.Unlock()
	}
}

// recvLoop 接收循环
func (t *RealTransport) recvLoop(peer *peerConn) {
	decoder := json.NewDecoder(peer.conn)
	for {
		var msg types.Message
		if err := decoder.Decode(&msg); err != nil {
			peer.online = false
			peer.conn.Close()
			return
		}

		t.mu.Lock()
		t.stats.MessagesReceived++
		t.mu.Unlock()

		select {
		case t.inbox <- &msg:
		default:
		}
	}
}

// Stop 停止传输层
func (t *RealTransport) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}
	t.running = false

	if t.listener != nil {
		t.listener.Close()
	}

	for _, peer := range t.peers {
		close(peer.outbox)
		if peer.conn != nil {
			peer.conn.Close()
		}
	}

	return nil
}

// AddNode 添加节点
func (t *RealTransport) AddNode(nodeID types.NodeID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.peers[nodeID]; exists {
		return nil
	}

	addr, ok := t.config.NodeAddresses[nodeID]
	if !ok {
		return fmt.Errorf("no address for node: %s", nodeID)
	}

	go t.connectToPeer(nodeID, addr)
	return nil
}

// RemoveNode 移除节点
func (t *RealTransport) RemoveNode(nodeID types.NodeID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if peer, exists := t.peers[nodeID]; exists {
		close(peer.outbox)
		if peer.conn != nil {
			peer.conn.Close()
		}
		delete(t.peers, nodeID)
	}
	return nil
}

// GetNodes 获取所有节点
func (t *RealTransport) GetNodes() []types.NodeID {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var nodes []types.NodeID
	for id := range t.peers {
		nodes = append(nodes, id)
	}
	return nodes
}

// Send 发送消息
func (t *RealTransport) Send(msg *types.Message) error {
	t.mu.RLock()
	peer, exists := t.peers[msg.To]
	t.mu.RUnlock()

	if !exists || !peer.online {
		return fmt.Errorf("peer not connected: %s", msg.To)
	}

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	select {
	case peer.outbox <- msg:
		return nil
	default:
		return fmt.Errorf("peer outbox full: %s", msg.To)
	}
}

// Broadcast 广播消息
func (t *RealTransport) Broadcast(msg *types.Broadcast) error {
	t.mu.RLock()
	peers := make([]*peerConn, 0, len(t.peers))
	for _, peer := range t.peers {
		if peer.online {
			peers = append(peers, peer)
		}
	}
	t.mu.RUnlock()

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	for _, peer := range peers {
		unicast := &types.Message{
			ID:        msg.ID + "-" + string(peer.nodeID),
			Type:      msg.Type,
			From:      msg.From,
			To:        peer.nodeID,
			Payload:   msg.Payload,
			Timestamp: msg.Timestamp,
		}
		select {
		case peer.outbox <- unicast:
		default:
		}
	}

	return nil
}

// Receive 接收消息
func (t *RealTransport) Receive(nodeID types.NodeID) <-chan *types.Message {
	return t.inbox
}

// SetLatency 设置延迟（真实模式不支持）
func (t *RealTransport) SetLatency(from, to types.NodeID, latency time.Duration) {
	// 真实模式下不支持人工设置延迟
}

// SetPacketLoss 设置丢包率（真实模式不支持）
func (t *RealTransport) SetPacketLoss(from, to types.NodeID, rate float64) {
	// 真实模式下不支持人工设置丢包
}

// CreatePartition 创建网络分区（真实模式不支持）
func (t *RealTransport) CreatePartition(groups [][]types.NodeID) {
	// 真实模式下不支持人工创建分区
}

// HealPartition 恢复网络分区
func (t *RealTransport) HealPartition() {
	// 真实模式下不支持
}

// GetStats 获取统计信息
func (t *RealTransport) GetStats() TransportStats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats
}

// SetLocalID 设置本地节点ID
func (t *RealTransport) SetLocalID(nodeID types.NodeID) {
	t.localID = nodeID
}

// SetLocalAddr 设置本地地址
func (t *RealTransport) SetLocalAddr(addr string) {
	t.localAddr = addr
}
