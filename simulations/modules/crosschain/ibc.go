package crosschain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// IBC协议演示器
// 演示Cosmos生态的跨链通信协议(Inter-Blockchain Communication)
//
// IBC核心组件:
// 1. 轻客户端: 验证对方链的区块头
// 2. 连接(Connection): 两条链之间的信任关系
// 3. 通道(Channel): 应用层的消息传递通道
// 4. 端口(Port): 应用模块的标识
//
// IBC优势:
// - 无需信任第三方
// - 由轻客户端保证安全
// - 支持任意消息传递
// - 标准化协议
//
// 参考: Cosmos SDK IBC模块, ICS标准
// =============================================================================

// IBCClientState IBC客户端状态
type IBCClientState struct {
	ClientID       string        `json:"client_id"`
	ChainID        string        `json:"chain_id"`
	LatestHeight   uint64        `json:"latest_height"`
	TrustingPeriod time.Duration `json:"trusting_period"`
	FrozenHeight   uint64        `json:"frozen_height"`
	Status         string        `json:"status"`
}

// IBCConnection IBC连接
type IBCConnection struct {
	ConnectionID   string   `json:"connection_id"`
	ClientID       string   `json:"client_id"`
	CounterpartyID string   `json:"counterparty_connection_id"`
	State          string   `json:"state"`
	Versions       []string `json:"versions"`
}

// IBCChannel IBC通道
type IBCChannel struct {
	ChannelID           string `json:"channel_id"`
	PortID              string `json:"port_id"`
	ConnectionID        string `json:"connection_id"`
	CounterpartyChannel string `json:"counterparty_channel"`
	CounterpartyPort    string `json:"counterparty_port"`
	State               string `json:"state"`
	Ordering            string `json:"ordering"`
	Version             string `json:"version"`
}

// IBCPacket IBC数据包
type IBCPacket struct {
	Sequence         uint64    `json:"sequence"`
	SourcePort       string    `json:"source_port"`
	SourceChannel    string    `json:"source_channel"`
	DestPort         string    `json:"dest_port"`
	DestChannel      string    `json:"dest_channel"`
	Data             []byte    `json:"data"`
	TimeoutHeight    uint64    `json:"timeout_height"`
	TimeoutTimestamp time.Time `json:"timeout_timestamp"`
	Status           string    `json:"status"`
}

// IBCTransfer IBC转账
type IBCTransfer struct {
	ID          string    `json:"id"`
	Sender      string    `json:"sender"`
	Receiver    string    `json:"receiver"`
	Amount      string    `json:"amount"`
	Denom       string    `json:"denom"`
	SourceChain string    `json:"source_chain"`
	DestChain   string    `json:"dest_chain"`
	PacketSeq   uint64    `json:"packet_sequence"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// IBCSimulator IBC协议演示器
type IBCSimulator struct {
	*base.BaseSimulator
	clients     map[string]*IBCClientState
	connections map[string]*IBCConnection
	channels    map[string]*IBCChannel
	packets     map[string]*IBCPacket
	transfers   map[string]*IBCTransfer
	packetSeq   uint64
}

// NewIBCSimulator 创建IBC协议演示器
func NewIBCSimulator() *IBCSimulator {
	sim := &IBCSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"ibc",
			"IBC协议演示器",
			"演示Cosmos IBC跨链通信协议的连接建立、通道创建、数据包传输等核心功能",
			"crosschain",
			types.ComponentProcess,
		),
		clients:     make(map[string]*IBCClientState),
		connections: make(map[string]*IBCConnection),
		channels:    make(map[string]*IBCChannel),
		packets:     make(map[string]*IBCPacket),
		transfers:   make(map[string]*IBCTransfer),
	}

	return sim
}

// Init 初始化
func (s *IBCSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.clients = make(map[string]*IBCClientState)
	s.connections = make(map[string]*IBCConnection)
	s.channels = make(map[string]*IBCChannel)
	s.packets = make(map[string]*IBCPacket)
	s.transfers = make(map[string]*IBCTransfer)
	s.packetSeq = 0

	s.initializeSampleIBC()
	s.updateState()
	return nil
}

// initializeSampleIBC 初始化示例IBC
func (s *IBCSimulator) initializeSampleIBC() {
	s.clients["07-tendermint-0"] = &IBCClientState{
		ClientID: "07-tendermint-0", ChainID: "osmosis-1",
		LatestHeight: 10000000, TrustingPeriod: 14 * 24 * time.Hour,
		Status: "Active",
	}
	s.clients["07-tendermint-1"] = &IBCClientState{
		ClientID: "07-tendermint-1", ChainID: "cosmoshub-4",
		LatestHeight: 15000000, TrustingPeriod: 14 * 24 * time.Hour,
		Status: "Active",
	}

	s.connections["connection-0"] = &IBCConnection{
		ConnectionID: "connection-0", ClientID: "07-tendermint-0",
		CounterpartyID: "connection-0", State: "OPEN",
		Versions: []string{"1"},
	}

	s.channels["channel-0"] = &IBCChannel{
		ChannelID: "channel-0", PortID: "transfer",
		ConnectionID:        "connection-0",
		CounterpartyChannel: "channel-0", CounterpartyPort: "transfer",
		State: "OPEN", Ordering: "UNORDERED", Version: "ics20-1",
	}
}

// =============================================================================
// IBC机制解释
// =============================================================================

// ExplainIBC 解释IBC协议
func (s *IBCSimulator) ExplainIBC() map[string]interface{} {
	return map[string]interface{}{
		"overview": "IBC是Cosmos生态的标准跨链通信协议，允许异构区块链之间安全传递任意消息",
		"architecture": map[string]interface{}{
			"layers": []map[string]string{
				{"layer": "应用层", "desc": "ICS-20(转账)、ICS-721(NFT)等应用"},
				{"layer": "传输层", "desc": "通道(Channel)和端口(Port)"},
				{"layer": "连接层", "desc": "连接(Connection)管理"},
				{"layer": "客户端层", "desc": "轻客户端验证"},
			},
		},
		"core_components": []map[string]interface{}{
			{
				"name":    "轻客户端 (Light Client)",
				"purpose": "验证对方链的区块头",
				"types":   []string{"Tendermint", "Solomachine", "Localhost"},
				"trust":   "信任对方链的共识(2/3+验证者)",
			},
			{
				"name":      "连接 (Connection)",
				"purpose":   "建立两条链之间的信任关系",
				"handshake": []string{"INIT", "TRY", "ACK", "CONFIRM"},
			},
			{
				"name":     "通道 (Channel)",
				"purpose":  "应用层的消息传递通道",
				"ordering": []string{"ORDERED(有序)", "UNORDERED(无序)"},
			},
			{
				"name":      "数据包 (Packet)",
				"purpose":   "跨链传输的消息单元",
				"lifecycle": []string{"发送", "接收", "确认/超时"},
			},
		},
		"security_model": map[string]interface{}{
			"trust_assumption": "信任两条链的共识机制(各2/3+验证者诚实)",
			"no_third_party":   "不依赖任何第三方验证者或中继者",
			"relayer_role":     "中继者只负责传递消息，不影响安全性",
			"verification":     "目标链通过轻客户端独立验证消息",
		},
		"advantages": []string{
			"标准化: ICS(Interchain Standards)规范",
			"安全: 由轻客户端保证，无需信任中继者",
			"通用: 支持任意消息传递",
			"可组合: 多种应用可共享连接和通道",
		},
	}
}

// =============================================================================
// IBC操作
// =============================================================================

// CreateClient 创建客户端
func (s *IBCSimulator) CreateClient(chainID string, height uint64) (*IBCClientState, error) {
	clientID := fmt.Sprintf("07-tendermint-%d", len(s.clients))

	client := &IBCClientState{
		ClientID:       clientID,
		ChainID:        chainID,
		LatestHeight:   height,
		TrustingPeriod: 14 * 24 * time.Hour,
		Status:         "Active",
	}

	s.clients[clientID] = client

	s.EmitEvent("client_created", "", "", map[string]interface{}{
		"client_id": clientID,
		"chain_id":  chainID,
		"height":    height,
	})

	s.updateState()
	return client, nil
}

// OpenConnection 打开连接
func (s *IBCSimulator) OpenConnection(clientID string) (*IBCConnection, error) {
	if _, ok := s.clients[clientID]; !ok {
		return nil, fmt.Errorf("客户端不存在: %s", clientID)
	}

	connID := fmt.Sprintf("connection-%d", len(s.connections))

	conn := &IBCConnection{
		ConnectionID:   connID,
		ClientID:       clientID,
		CounterpartyID: connID,
		State:          "OPEN",
		Versions:       []string{"1"},
	}

	s.connections[connID] = conn

	s.EmitEvent("connection_opened", "", "", map[string]interface{}{
		"connection_id": connID,
		"client_id":     clientID,
	})

	s.updateState()
	return conn, nil
}

// OpenChannel 打开通道
func (s *IBCSimulator) OpenChannel(connectionID, portID, version string) (*IBCChannel, error) {
	if _, ok := s.connections[connectionID]; !ok {
		return nil, fmt.Errorf("连接不存在: %s", connectionID)
	}

	channelID := fmt.Sprintf("channel-%d", len(s.channels))

	channel := &IBCChannel{
		ChannelID:           channelID,
		PortID:              portID,
		ConnectionID:        connectionID,
		CounterpartyChannel: channelID,
		CounterpartyPort:    portID,
		State:               "OPEN",
		Ordering:            "UNORDERED",
		Version:             version,
	}

	s.channels[channelID] = channel

	s.EmitEvent("channel_opened", "", "", map[string]interface{}{
		"channel_id":    channelID,
		"port_id":       portID,
		"connection_id": connectionID,
	})

	s.updateState()
	return channel, nil
}

// SendPacket 发送数据包
func (s *IBCSimulator) SendPacket(sourceChannel, sourcePort string, data []byte, timeoutHeight uint64) (*IBCPacket, error) {
	channel, ok := s.channels[sourceChannel]
	if !ok {
		return nil, fmt.Errorf("通道不存在: %s", sourceChannel)
	}

	s.packetSeq++

	packet := &IBCPacket{
		Sequence:         s.packetSeq,
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		DestPort:         channel.CounterpartyPort,
		DestChannel:      channel.CounterpartyChannel,
		Data:             data,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: time.Now().Add(1 * time.Hour),
		Status:           "sent",
	}

	packetKey := fmt.Sprintf("%s/%s/%d", sourcePort, sourceChannel, s.packetSeq)
	s.packets[packetKey] = packet

	s.EmitEvent("packet_sent", "", "", map[string]interface{}{
		"sequence":       s.packetSeq,
		"source_channel": sourceChannel,
		"source_port":    sourcePort,
	})

	s.updateState()
	return packet, nil
}

// SimulateIBCTransfer 模拟IBC转账
func (s *IBCSimulator) SimulateIBCTransfer(sender, receiver, amount, denom, sourceChain, destChain string) map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	steps = append(steps, map[string]interface{}{
		"step": 1, "action": "用户发起IBC转账",
		"sender": sender, "receiver": receiver, "amount": amount, "denom": denom,
	})

	transferData := fmt.Sprintf(`{"sender":"%s","receiver":"%s","amount":"%s","denom":"%s"}`,
		sender, receiver, amount, denom)
	packet, _ := s.SendPacket("channel-0", "transfer", []byte(transferData), 10000100)

	steps = append(steps, map[string]interface{}{
		"step": 2, "action": "源链transfer模块创建IBC数据包",
		"packet_sequence": packet.Sequence,
	})

	steps = append(steps, map[string]interface{}{
		"step": 3, "action": "中继者监听源链事件，获取数据包",
		"relayer": "relayer-1",
	})

	steps = append(steps, map[string]interface{}{
		"step": 4, "action": "中继者向目标链提交数据包和证明",
		"proof_type": "Merkle proof of packet commitment",
	})

	steps = append(steps, map[string]interface{}{
		"step": 5, "action": "目标链轻客户端验证证明",
		"verification": "验证源链区块头和Merkle证明",
	})

	steps = append(steps, map[string]interface{}{
		"step": 6, "action": "目标链transfer模块铸造IBC代币",
		"ibc_denom": fmt.Sprintf("ibc/%s", hex.EncodeToString(sha256.New().Sum([]byte(denom)))[:16]),
	})

	steps = append(steps, map[string]interface{}{
		"step": 7, "action": "中继者提交确认到源链",
		"ack": "success",
	})

	transferID := fmt.Sprintf("transfer-%d", time.Now().UnixNano())
	transfer := &IBCTransfer{
		ID: transferID, Sender: sender, Receiver: receiver,
		Amount: amount, Denom: denom,
		SourceChain: sourceChain, DestChain: destChain,
		PacketSeq: packet.Sequence, Status: "completed",
		CreatedAt: time.Now(),
	}
	s.transfers[transferID] = transfer

	return map[string]interface{}{
		"transfer_id":     transferID,
		"packet_sequence": packet.Sequence,
		"source_chain":    sourceChain,
		"dest_chain":      destChain,
		"amount":          amount,
		"denom":           denom,
		"steps":           steps,
		"status":          "completed",
	}
}

// GetStatistics 获取统计
func (s *IBCSimulator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"clients":     len(s.clients),
		"connections": len(s.connections),
		"channels":    len(s.channels),
		"packets":     len(s.packets),
		"transfers":   len(s.transfers),
	}
}

// updateState 更新状态
func (s *IBCSimulator) updateState() {
	s.SetGlobalData("clients", len(s.clients))
	s.SetGlobalData("connections", len(s.connections))
	s.SetGlobalData("channels", len(s.channels))

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"ibc",
		"当前可以依次建立 IBC client、connection 和 channel。",
		"先创建轻客户端，再打开 connection 与 channel，观察 IBC 握手链路。",
		0,
		map[string]interface{}{
			"clients":     len(s.clients),
			"connections": len(s.connections),
			"channels":    len(s.channels),
		},
	)
}

func (s *IBCSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "setup_channel":
		client, err := s.CreateClient("cosmoshub-4", 1000)
		if err != nil {
			return nil, err
		}
		conn, err := s.OpenConnection(client.ClientID)
		if err != nil {
			return nil, err
		}
		channel, err := s.OpenChannel(conn.ConnectionID, "transfer", "ics20-1")
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已完成 IBC 通道建立",
			map[string]interface{}{"client_id": client.ClientID, "connection_id": conn.ConnectionID, "channel_id": channel.ChannelID},
			&types.ActionFeedback{
				Summary:     "IBC 握手链路已经建立完成，可继续发送 packet 观察跨链数据传输。",
				NextHint:    "继续发送一条 packet，观察它在 channel 上的流转与确认。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported ibc action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// IBCFactory IBC工厂
type IBCFactory struct{}

func (f *IBCFactory) Create() engine.Simulator          { return NewIBCSimulator() }
func (f *IBCFactory) GetDescription() types.Description { return NewIBCSimulator().GetDescription() }
func NewIBCFactory() *IBCFactory                        { return &IBCFactory{} }

var _ engine.SimulatorFactory = (*IBCFactory)(nil)
