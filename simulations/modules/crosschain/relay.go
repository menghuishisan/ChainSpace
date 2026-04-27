package crosschain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 中继链演示器
// 演示跨链消息传递机制，包括:
// 1. 消息格式: 标准化的跨链消息结构
// 2. 中继器: 监听源链事件，生成证明，提交到目标链
// 3. 验证: 验证消息来源、Merkle证明、防重放
//
// 安全模型:
// - 多签中继: 多个中继器签名确认
// - 轻客户端: 目标链验证源链区块头
// - 乐观中继: 挑战期机制
//
// 参考实现: Cosmos IBC, LayerZero, Axelar, Chainlink CCIP
// =============================================================================

// MessageStatus 消息状态
type MessageStatus string

const (
	MsgStatusPending  MessageStatus = "pending"
	MsgStatusRelayed  MessageStatus = "relayed"
	MsgStatusVerified MessageStatus = "verified"
	MsgStatusExecuted MessageStatus = "executed"
	MsgStatusFailed   MessageStatus = "failed"
	MsgStatusExpired  MessageStatus = "expired"
)

// CrossChainMessage 跨链消息
type CrossChainMessage struct {
	ID           string        `json:"id"`
	Nonce        uint64        `json:"nonce"`
	SourceChain  string        `json:"source_chain"`
	DestChain    string        `json:"dest_chain"`
	Sender       string        `json:"sender"`
	Receiver     string        `json:"receiver"`
	Payload      []byte        `json:"payload"`
	PayloadType  string        `json:"payload_type"`
	GasLimit     uint64        `json:"gas_limit"`
	Fee          *big.Int      `json:"fee"`
	Status       MessageStatus `json:"status"`
	SourceTxHash string        `json:"source_tx_hash"`
	DestTxHash   string        `json:"dest_tx_hash"`
	Proof        *MessageProof `json:"proof"`
	Signatures   []string      `json:"signatures"`
	CreatedAt    time.Time     `json:"created_at"`
	RelayedAt    time.Time     `json:"relayed_at"`
	ExecutedAt   time.Time     `json:"executed_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

// MessageProof 消息证明
type MessageProof struct {
	BlockHeight  uint64   `json:"block_height"`
	BlockHash    string   `json:"block_hash"`
	TxIndex      int      `json:"tx_index"`
	MerkleProof  []string `json:"merkle_proof"`
	ReceiptProof []string `json:"receipt_proof"`
	StateRoot    string   `json:"state_root"`
}

// RelayerNode 中继器节点
type RelayerNode struct {
	Address         string    `json:"address"`
	Name            string    `json:"name"`
	Stake           *big.Int  `json:"stake"`
	Commission      float64   `json:"commission"`
	RelayedCount    int       `json:"relayed_count"`
	SuccessCount    int       `json:"success_count"`
	FailedCount     int       `json:"failed_count"`
	Rewards         *big.Int  `json:"rewards"`
	Slashed         *big.Int  `json:"slashed"`
	IsActive        bool      `json:"is_active"`
	SupportedChains []string  `json:"supported_chains"`
	LastActiveAt    time.Time `json:"last_active_at"`
}

// RelayChannel 中继通道
type RelayChannel struct {
	ID            string    `json:"id"`
	SourceChain   string    `json:"source_chain"`
	DestChain     string    `json:"dest_chain"`
	SourceGateway string    `json:"source_gateway"`
	DestGateway   string    `json:"dest_gateway"`
	State         string    `json:"state"`
	Nonce         uint64    `json:"nonce"`
	MessageCount  int       `json:"message_count"`
	TotalVolume   *big.Int  `json:"total_volume"`
	CreatedAt     time.Time `json:"created_at"`
}

// ChainGateway 链网关
type ChainGateway struct {
	ChainID       string        `json:"chain_id"`
	ChainName     string        `json:"chain_name"`
	GatewayAddr   string        `json:"gateway_address"`
	Finality      int           `json:"finality_blocks"`
	BlockTime     time.Duration `json:"block_time"`
	SupportedMsgs []string      `json:"supported_messages"`
	LastBlockNum  uint64        `json:"last_block_num"`
	LastBlockHash string        `json:"last_block_hash"`
}

// RelaySimulator 中继链演示器
type RelaySimulator struct {
	*base.BaseSimulator
	messages      map[string]*CrossChainMessage
	relayers      map[string]*RelayerNode
	channels      map[string]*RelayChannel
	gateways      map[string]*ChainGateway
	nonces        map[string]uint64
	requiredSigs  int
	messageExpiry time.Duration
	rewardPool    *big.Int
	totalRelayed  int
	totalFees     *big.Int
}

// NewRelaySimulator 创建中继链演示器
func NewRelaySimulator() *RelaySimulator {
	sim := &RelaySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"relay",
			"中继链演示器",
			"演示跨链消息传递、中继器机制、消息验证等核心功能",
			"crosschain",
			types.ComponentProcess,
		),
		messages:   make(map[string]*CrossChainMessage),
		relayers:   make(map[string]*RelayerNode),
		channels:   make(map[string]*RelayChannel),
		gateways:   make(map[string]*ChainGateway),
		nonces:     make(map[string]uint64),
		rewardPool: big.NewInt(0),
		totalFees:  big.NewInt(0),
	}

	sim.AddParam(types.Param{
		Key:         "relayer_count",
		Name:        "中继器数量",
		Description: "活跃中继器的数量",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         1,
		Max:         20,
	})

	sim.AddParam(types.Param{
		Key:         "required_signatures",
		Name:        "所需签名数",
		Description: "消息验证所需的中继器签名数",
		Type:        types.ParamTypeInt,
		Default:     3,
		Min:         1,
		Max:         10,
	})

	sim.AddParam(types.Param{
		Key:         "message_expiry_hours",
		Name:        "消息过期时间(小时)",
		Description: "未执行消息的过期时间",
		Type:        types.ParamTypeInt,
		Default:     24,
		Min:         1,
		Max:         168,
	})

	return sim
}

// Init 初始化
func (s *RelaySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	relayerCount := 5
	s.requiredSigs = 3
	s.messageExpiry = 24 * time.Hour

	if v, ok := config.Params["relayer_count"]; ok {
		if n, ok := v.(float64); ok {
			relayerCount = int(n)
		}
	}
	if v, ok := config.Params["required_signatures"]; ok {
		if n, ok := v.(float64); ok {
			s.requiredSigs = int(n)
		}
	}
	if v, ok := config.Params["message_expiry_hours"]; ok {
		if n, ok := v.(float64); ok {
			s.messageExpiry = time.Duration(n) * time.Hour
		}
	}

	s.messages = make(map[string]*CrossChainMessage)
	s.relayers = make(map[string]*RelayerNode)
	s.channels = make(map[string]*RelayChannel)
	s.gateways = make(map[string]*ChainGateway)
	s.nonces = make(map[string]uint64)
	s.rewardPool = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18))
	s.totalFees = big.NewInt(0)
	s.totalRelayed = 0

	s.initializeGateways()
	s.initializeRelayers(relayerCount)
	s.initializeChannels()

	s.updateState()
	return nil
}

// initializeGateways 初始化网关
func (s *RelaySimulator) initializeGateways() {
	gateways := []struct {
		id, name, addr string
		finality       int
		blockTime      time.Duration
	}{
		{"ethereum", "Ethereum", "0x1111...gateway", 12, 12 * time.Second},
		{"polygon", "Polygon", "0x2222...gateway", 256, 2 * time.Second},
		{"arbitrum", "Arbitrum", "0x3333...gateway", 1, 250 * time.Millisecond},
		{"optimism", "Optimism", "0x4444...gateway", 1, 2 * time.Second},
		{"avalanche", "Avalanche", "0x5555...gateway", 1, 2 * time.Second},
		{"bsc", "BNB Chain", "0x6666...gateway", 15, 3 * time.Second},
	}

	for _, g := range gateways {
		s.gateways[g.id] = &ChainGateway{
			ChainID:       g.id,
			ChainName:     g.name,
			GatewayAddr:   g.addr,
			Finality:      g.finality,
			BlockTime:     g.blockTime,
			SupportedMsgs: []string{"transfer", "call", "deploy"},
			LastBlockNum:  1000000,
		}
	}
}

// initializeRelayers 初始化中继器
func (s *RelaySimulator) initializeRelayers(count int) {
	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("Relayer-%d", i)
		hash := sha256.Sum256([]byte(name))
		addr := fmt.Sprintf("0x%s", hex.EncodeToString(hash[:20]))

		chains := []string{"ethereum", "polygon", "arbitrum"}
		if i <= 3 {
			chains = append(chains, "optimism", "avalanche", "bsc")
		}

		s.relayers[name] = &RelayerNode{
			Address:         addr,
			Name:            name,
			Stake:           new(big.Int).Mul(big.NewInt(int64(100000-i*10000)), big.NewInt(1e18)),
			Commission:      0.05 + float64(i)*0.01,
			RelayedCount:    0,
			SuccessCount:    0,
			FailedCount:     0,
			Rewards:         big.NewInt(0),
			Slashed:         big.NewInt(0),
			IsActive:        true,
			SupportedChains: chains,
			LastActiveAt:    time.Now(),
		}
	}
}

// initializeChannels 初始化通道
func (s *RelaySimulator) initializeChannels() {
	pairs := []struct{ src, dst string }{
		{"ethereum", "polygon"},
		{"ethereum", "arbitrum"},
		{"ethereum", "optimism"},
		{"polygon", "arbitrum"},
		{"avalanche", "bsc"},
	}

	for _, p := range pairs {
		channelID := fmt.Sprintf("channel-%s-%s", p.src, p.dst)
		s.channels[channelID] = &RelayChannel{
			ID:            channelID,
			SourceChain:   p.src,
			DestChain:     p.dst,
			SourceGateway: s.gateways[p.src].GatewayAddr,
			DestGateway:   s.gateways[p.dst].GatewayAddr,
			State:         "open",
			Nonce:         0,
			MessageCount:  0,
			TotalVolume:   big.NewInt(0),
			CreatedAt:     time.Now().Add(-30 * 24 * time.Hour),
		}
	}
}

// =============================================================================
// 中继机制解释
// =============================================================================

// ExplainRelayMechanism 解释中继机制
func (s *RelaySimulator) ExplainRelayMechanism() map[string]interface{} {
	return map[string]interface{}{
		"overview": "中继器负责在不同区块链之间传递消息，是跨链通信的核心基础设施",
		"message_lifecycle": []map[string]string{
			{"stage": "1. 发送", "action": "用户在源链调用Gateway.sendMessage()"},
			{"stage": "2. 事件", "action": "Gateway发出MessageSent事件"},
			{"stage": "3. 索引", "action": "中继器监听并索引源链事件"},
			{"stage": "4. 等待", "action": "等待源链达到最终性"},
			{"stage": "5. 生成证明", "action": "生成Merkle证明或收集签名"},
			{"stage": "6. 提交", "action": "中继器将消息和证明提交到目标链"},
			{"stage": "7. 验证", "action": "目标链Gateway验证证明有效性"},
			{"stage": "8. 执行", "action": "调用目标合约处理消息"},
		},
		"security_models": []map[string]interface{}{
			{
				"model":       "多签验证 (Multisig)",
				"description": "多个中继器独立验证并签名",
				"threshold":   fmt.Sprintf("%d/%d", s.requiredSigs, len(s.relayers)),
				"trust":       "信任多数中继器诚实",
				"examples":    []string{"Axelar", "Wormhole"},
			},
			{
				"model":       "轻客户端 (Light Client)",
				"description": "目标链运行源链轻客户端，验证区块头",
				"trust":       "信任源链共识",
				"examples":    []string{"IBC", "Rainbow Bridge"},
			},
			{
				"model":       "乐观验证 (Optimistic)",
				"description": "假设消息有效，挑战期内可被质疑",
				"trust":       "至少有一个诚实的监视者",
				"examples":    []string{"Nomad"},
			},
			{
				"model":       "ZK证明 (Zero-Knowledge)",
				"description": "使用ZK-SNARK证明状态转换",
				"trust":       "密码学假设",
				"examples":    []string{"zkBridge", "Succinct"},
			},
		},
		"message_format": map[string]string{
			"nonce":        "防重放的序列号",
			"source_chain": "源链标识",
			"dest_chain":   "目标链标识",
			"sender":       "源链发送者地址",
			"receiver":     "目标链接收者地址",
			"payload":      "消息内容(ABI编码)",
			"gas_limit":    "目标链执行gas限制",
		},
	}
}

// =============================================================================
// 消息操作
// =============================================================================

// SendMessage 发送跨链消息
func (s *RelaySimulator) SendMessage(sourceChain, destChain, sender, receiver string, payload []byte, payloadType string, gasLimit uint64) (*CrossChainMessage, error) {
	srcGateway, ok := s.gateways[sourceChain]
	if !ok {
		return nil, fmt.Errorf("源链网关不存在: %s", sourceChain)
	}
	if _, ok := s.gateways[destChain]; !ok {
		return nil, fmt.Errorf("目标链网关不存在: %s", destChain)
	}

	channelKey := fmt.Sprintf("channel-%s-%s", sourceChain, destChain)
	channel := s.channels[channelKey]
	if channel == nil {
		return nil, fmt.Errorf("通道不存在: %s -> %s", sourceChain, destChain)
	}

	nonceKey := fmt.Sprintf("%s-%s", sourceChain, sender)
	nonce := s.nonces[nonceKey]
	s.nonces[nonceKey] = nonce + 1

	msgData := fmt.Sprintf("%s-%s-%s-%d-%d", sourceChain, destChain, sender, nonce, time.Now().UnixNano())
	msgHash := sha256.Sum256([]byte(msgData))
	msgID := fmt.Sprintf("msg-%s", hex.EncodeToString(msgHash[:8]))
	txHash := fmt.Sprintf("0x%s", hex.EncodeToString(msgHash[:]))

	fee := new(big.Int).Mul(big.NewInt(int64(gasLimit)), big.NewInt(1000000000))

	msg := &CrossChainMessage{
		ID:           msgID,
		Nonce:        nonce,
		SourceChain:  sourceChain,
		DestChain:    destChain,
		Sender:       sender,
		Receiver:     receiver,
		Payload:      payload,
		PayloadType:  payloadType,
		GasLimit:     gasLimit,
		Fee:          fee,
		Status:       MsgStatusPending,
		SourceTxHash: txHash,
		Signatures:   make([]string, 0),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.messageExpiry),
	}

	msg.Proof = &MessageProof{
		BlockHeight: srcGateway.LastBlockNum,
		BlockHash:   fmt.Sprintf("0x%x", srcGateway.LastBlockNum),
		TxIndex:     0,
	}

	s.messages[msgID] = msg
	channel.MessageCount++
	s.totalFees.Add(s.totalFees, fee)

	s.EmitEvent("message_sent", "", "", map[string]interface{}{
		"msg_id":       msgID,
		"source_chain": sourceChain,
		"dest_chain":   destChain,
		"sender":       sender,
		"receiver":     receiver,
		"payload_type": payloadType,
		"gas_limit":    gasLimit,
		"nonce":        nonce,
		"finality":     srcGateway.Finality,
	})

	s.updateState()
	return msg, nil
}

// RelayMessage 中继消息
func (s *RelaySimulator) RelayMessage(msgID, relayerName string) error {
	msg, ok := s.messages[msgID]
	if !ok {
		return fmt.Errorf("消息不存在: %s", msgID)
	}

	if msg.Status != MsgStatusPending {
		return fmt.Errorf("消息状态不是pending: %s", msg.Status)
	}

	if time.Now().After(msg.ExpiresAt) {
		msg.Status = MsgStatusExpired
		return fmt.Errorf("消息已过期")
	}

	relayer, ok := s.relayers[relayerName]
	if !ok {
		return fmt.Errorf("中继器不存在: %s", relayerName)
	}

	if !relayer.IsActive {
		return fmt.Errorf("中继器不活跃: %s", relayerName)
	}

	for _, sig := range msg.Signatures {
		if sig == relayerName {
			return fmt.Errorf("中继器已签名: %s", relayerName)
		}
	}

	sigData := fmt.Sprintf("%s-%s-%d", msgID, relayerName, time.Now().UnixNano())
	sigHash := sha256.Sum256([]byte(sigData))
	signature := hex.EncodeToString(sigHash[:])

	msg.Signatures = append(msg.Signatures, relayerName)
	relayer.RelayedCount++
	relayer.LastActiveAt = time.Now()

	if len(msg.Signatures) >= s.requiredSigs {
		msg.Status = MsgStatusRelayed
		msg.RelayedAt = time.Now()
	}

	s.EmitEvent("message_relayed", "", "", map[string]interface{}{
		"msg_id":     msgID,
		"relayer":    relayerName,
		"signature":  signature[:16] + "...",
		"signatures": len(msg.Signatures),
		"required":   s.requiredSigs,
	})

	s.updateState()
	return nil
}

// VerifyMessage 验证消息
func (s *RelaySimulator) VerifyMessage(msgID string) error {
	msg, ok := s.messages[msgID]
	if !ok {
		return fmt.Errorf("消息不存在: %s", msgID)
	}

	if msg.Status != MsgStatusRelayed {
		return fmt.Errorf("消息未被中继: %s", msg.Status)
	}

	if len(msg.Signatures) < s.requiredSigs {
		return fmt.Errorf("签名不足: %d/%d", len(msg.Signatures), s.requiredSigs)
	}

	if msg.Proof == nil {
		return fmt.Errorf("缺少证明")
	}

	msg.Status = MsgStatusVerified

	s.EmitEvent("message_verified", "", "", map[string]interface{}{
		"msg_id":     msgID,
		"signatures": len(msg.Signatures),
	})

	s.updateState()
	return nil
}

// ExecuteMessage 执行消息
func (s *RelaySimulator) ExecuteMessage(msgID string) error {
	msg, ok := s.messages[msgID]
	if !ok {
		return fmt.Errorf("消息不存在: %s", msgID)
	}

	if msg.Status != MsgStatusVerified && msg.Status != MsgStatusRelayed {
		return fmt.Errorf("消息未验证: %s", msg.Status)
	}

	execData := fmt.Sprintf("%s-exec-%d", msgID, time.Now().UnixNano())
	execHash := sha256.Sum256([]byte(execData))
	msg.DestTxHash = fmt.Sprintf("0x%s", hex.EncodeToString(execHash[:]))

	msg.Status = MsgStatusExecuted
	msg.ExecutedAt = time.Now()
	s.totalRelayed++

	reward := new(big.Int).Div(msg.Fee, big.NewInt(int64(len(msg.Signatures))))
	for _, signerName := range msg.Signatures {
		if relayer, ok := s.relayers[signerName]; ok {
			relayer.SuccessCount++
			relayer.Rewards.Add(relayer.Rewards, reward)
		}
	}

	s.EmitEvent("message_executed", "", "", map[string]interface{}{
		"msg_id":       msgID,
		"dest_tx_hash": msg.DestTxHash,
		"receiver":     msg.Receiver,
		"duration":     msg.ExecutedAt.Sub(msg.CreatedAt).String(),
	})

	s.updateState()
	return nil
}

// SimulateMessageFlow 模拟完整消息流程
func (s *RelaySimulator) SimulateMessageFlow() map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	msg, _ := s.SendMessage(
		"ethereum", "polygon",
		"0xAlice", "0xBobContract",
		[]byte("transfer(0xCharlie, 1000)"),
		"contract_call",
		200000,
	)
	steps = append(steps, map[string]interface{}{
		"step":   1,
		"action": "Alice在Ethereum发送跨链消息",
		"msg_id": msg.ID,
		"status": string(msg.Status),
	})

	srcGateway := s.gateways["ethereum"]
	steps = append(steps, map[string]interface{}{
		"step":     2,
		"action":   fmt.Sprintf("等待%d个区块确认", srcGateway.Finality),
		"finality": srcGateway.Finality,
		"time":     (time.Duration(srcGateway.Finality) * srcGateway.BlockTime).String(),
	})

	relayerNames := make([]string, 0)
	for name := range s.relayers {
		relayerNames = append(relayerNames, name)
	}
	sort.Strings(relayerNames)

	for i := 0; i < s.requiredSigs && i < len(relayerNames); i++ {
		s.RelayMessage(msg.ID, relayerNames[i])
	}
	steps = append(steps, map[string]interface{}{
		"step":       3,
		"action":     fmt.Sprintf("中继器收集签名(%d/%d)", s.requiredSigs, len(s.relayers)),
		"signatures": len(msg.Signatures),
		"status":     string(msg.Status),
	})

	s.VerifyMessage(msg.ID)
	steps = append(steps, map[string]interface{}{
		"step":   4,
		"action": "目标链网关验证签名和证明",
		"status": string(msg.Status),
	})

	s.ExecuteMessage(msg.ID)
	steps = append(steps, map[string]interface{}{
		"step":         5,
		"action":       "执行消息调用目标合约",
		"dest_tx_hash": msg.DestTxHash,
		"status":       string(msg.Status),
	})

	return map[string]interface{}{
		"msg_id":     msg.ID,
		"steps":      steps,
		"total_time": msg.ExecutedAt.Sub(msg.CreatedAt).String(),
	}
}

// GetMessageInfo 获取消息信息
func (s *RelaySimulator) GetMessageInfo(msgID string) map[string]interface{} {
	msg, ok := s.messages[msgID]
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"id":           msg.ID,
		"nonce":        msg.Nonce,
		"source_chain": msg.SourceChain,
		"dest_chain":   msg.DestChain,
		"sender":       msg.Sender,
		"receiver":     msg.Receiver,
		"payload_type": msg.PayloadType,
		"status":       string(msg.Status),
		"signatures":   len(msg.Signatures),
		"source_tx":    msg.SourceTxHash,
		"dest_tx":      msg.DestTxHash,
		"created_at":   msg.CreatedAt,
		"executed_at":  msg.ExecutedAt,
	}
}

// GetStatistics 获取统计信息
func (s *RelaySimulator) GetStatistics() map[string]interface{} {
	pendingMsgs := 0
	executedMsgs := 0
	for _, msg := range s.messages {
		if msg.Status == MsgStatusPending || msg.Status == MsgStatusRelayed {
			pendingMsgs++
		} else if msg.Status == MsgStatusExecuted {
			executedMsgs++
		}
	}

	activeRelayers := 0
	for _, r := range s.relayers {
		if r.IsActive {
			activeRelayers++
		}
	}

	return map[string]interface{}{
		"total_messages":    len(s.messages),
		"pending_messages":  pendingMsgs,
		"executed_messages": executedMsgs,
		"total_relayed":     s.totalRelayed,
		"total_fees":        s.totalFees.String(),
		"relayer_count":     len(s.relayers),
		"active_relayers":   activeRelayers,
		"channel_count":     len(s.channels),
		"gateway_count":     len(s.gateways),
		"required_sigs":     s.requiredSigs,
	}
}

// updateState 更新状态
func (s *RelaySimulator) updateState() {
	s.SetGlobalData("message_count", len(s.messages))
	s.SetGlobalData("relayer_count", len(s.relayers))
	s.SetGlobalData("channel_count", len(s.channels))
	s.SetGlobalData("total_relayed", s.totalRelayed)

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"relay",
		"当前可以发送跨链消息，并观察中继、验证和执行如何逐步推进。",
		"先发送一条消息，再观察中继签名、验证通过和目标链执行的完整过程。",
		0,
		map[string]interface{}{
			"message_count": len(s.messages),
			"relayer_count": len(s.relayers),
			"total_relayed": s.totalRelayed,
		},
	)
}

func (s *RelaySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_message_flow":
		result := s.SimulateMessageFlow()
		return crosschainActionResult(
			"已模拟一轮跨链消息中继流程",
			result,
			&types.ActionFeedback{
				Summary:     "消息已经完成从发送到中继验证的推演，可继续查看每一步的签名与执行结果。",
				NextHint:    "再次执行并比较不同链路下的消息推进节奏。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported relay action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// RelayFactory 中继工厂
type RelayFactory struct{}

func (f *RelayFactory) Create() engine.Simulator { return NewRelaySimulator() }
func (f *RelayFactory) GetDescription() types.Description {
	return NewRelaySimulator().GetDescription()
}
func NewRelayFactory() *RelayFactory { return &RelayFactory{} }

var _ engine.SimulatorFactory = (*RelayFactory)(nil)
