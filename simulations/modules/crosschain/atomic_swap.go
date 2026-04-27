package crosschain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 原子交换演示器
// 演示哈希时间锁合约(HTLC)实现的无信任跨链原子交换
//
// 核心概念:
// 1. 哈希锁 (HashLock): 需要知道原像才能解锁
// 2. 时间锁 (TimeLock): 超时后可以退款
// 3. 原子性: 要么双方都完成交换，要么都取回资金
//
// 交换流程:
// 1. Alice生成secret，计算hash=SHA256(secret)
// 2. Alice在链A创建HTLC，锁定资产，设置hash和timelock1
// 3. Bob验证Alice的HTLC，在链B创建HTLC，使用相同hash，更短的timelock2
// 4. Alice用secret从Bob的HTLC取出资产(secret公开)
// 5. Bob从链上获得secret，用它从Alice的HTLC取出资产
//
// 安全性分析:
// - 时间锁差异: timelock1 > timelock2，保证Bob有足够时间认领
// - 原子性保证: 要么Alice揭示secret双方完成，要么超时双方退款
// - 无需信任: 不依赖任何第三方
//
// 参考: Bitcoin HTLC, Ethereum HTLC, Lightning Network
// =============================================================================

// HTLCState HTLC状态
type HTLCState string

const (
	HTLCStatePending  HTLCState = "pending"
	HTLCStateLocked   HTLCState = "locked"
	HTLCStateClaimed  HTLCState = "claimed"
	HTLCStateRefunded HTLCState = "refunded"
	HTLCStateExpired  HTLCState = "expired"
)

// HTLCContract 哈希时间锁合约
type HTLCContract struct {
	ID           string    `json:"id"`
	ContractAddr string    `json:"contract_address"`
	Sender       string    `json:"sender"`
	Recipient    string    `json:"recipient"`
	Amount       *big.Int  `json:"amount"`
	Token        string    `json:"token"`
	HashLock     string    `json:"hash_lock"`
	TimeLock     time.Time `json:"time_lock"`
	Secret       string    `json:"secret"`
	State        HTLCState `json:"state"`
	Chain        string    `json:"chain"`
	TxHash       string    `json:"tx_hash"`
	ClaimTxHash  string    `json:"claim_tx_hash"`
	RefundTxHash string    `json:"refund_tx_hash"`
	CreatedAt    time.Time `json:"created_at"`
	ClaimedAt    time.Time `json:"claimed_at"`
	RefundedAt   time.Time `json:"refunded_at"`
}

// AtomicSwapRecord 原子交换记录
type AtomicSwapRecord struct {
	ID                string        `json:"id"`
	Initiator         string        `json:"initiator"`
	Participant       string        `json:"participant"`
	InitiatorHTLC     *HTLCContract `json:"initiator_htlc"`
	ParticipantHTLC   *HTLCContract `json:"participant_htlc"`
	InitiatorAsset    string        `json:"initiator_asset"`
	ParticipantAsset  string        `json:"participant_asset"`
	InitiatorAmount   *big.Int      `json:"initiator_amount"`
	ParticipantAmount *big.Int      `json:"participant_amount"`
	ExchangeRate      float64       `json:"exchange_rate"`
	State             HTLCState     `json:"state"`
	Secret            string        `json:"secret"`
	SecretHash        string        `json:"secret_hash"`
	CreatedAt         time.Time     `json:"created_at"`
	CompletedAt       time.Time     `json:"completed_at"`
}

// SwapOffer 交换订单
type SwapOffer struct {
	ID           string    `json:"id"`
	Maker        string    `json:"maker"`
	MakerAsset   string    `json:"maker_asset"`
	MakerAmount  *big.Int  `json:"maker_amount"`
	MakerChain   string    `json:"maker_chain"`
	TakerAsset   string    `json:"taker_asset"`
	TakerAmount  *big.Int  `json:"taker_amount"`
	TakerChain   string    `json:"taker_chain"`
	ExchangeRate float64   `json:"exchange_rate"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// AtomicSwapSimulator 原子交换演示器
type AtomicSwapSimulator struct {
	*base.BaseSimulator
	htlcs           map[string]*HTLCContract
	swaps           map[string]*AtomicSwapRecord
	offers          map[string]*SwapOffer
	initiatorLock   time.Duration
	participantLock time.Duration
	completedSwaps  int
	failedSwaps     int
	totalVolume     *big.Int
}

// NewAtomicSwapSimulator 创建原子交换演示器
func NewAtomicSwapSimulator() *AtomicSwapSimulator {
	sim := &AtomicSwapSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"atomic_swap",
			"原子交换演示器",
			"演示哈希时间锁合约(HTLC)实现的无信任跨链原子交换",
			"crosschain",
			types.ComponentProcess,
		),
		htlcs:       make(map[string]*HTLCContract),
		swaps:       make(map[string]*AtomicSwapRecord),
		offers:      make(map[string]*SwapOffer),
		totalVolume: big.NewInt(0),
	}

	sim.AddParam(types.Param{
		Key:         "initiator_lock_hours",
		Name:        "发起方时间锁(小时)",
		Description: "发起方HTLC的时间锁时长",
		Type:        types.ParamTypeInt,
		Default:     48,
		Min:         2,
		Max:         168,
	})

	sim.AddParam(types.Param{
		Key:         "participant_lock_hours",
		Name:        "参与方时间锁(小时)",
		Description: "参与方HTLC的时间锁时长(必须小于发起方)",
		Type:        types.ParamTypeInt,
		Default:     24,
		Min:         1,
		Max:         84,
	})

	return sim
}

// Init 初始化
func (s *AtomicSwapSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.initiatorLock = 48 * time.Hour
	s.participantLock = 24 * time.Hour

	if v, ok := config.Params["initiator_lock_hours"]; ok {
		if n, ok := v.(float64); ok {
			s.initiatorLock = time.Duration(n) * time.Hour
		}
	}
	if v, ok := config.Params["participant_lock_hours"]; ok {
		if n, ok := v.(float64); ok {
			s.participantLock = time.Duration(n) * time.Hour
		}
	}

	if s.participantLock >= s.initiatorLock {
		s.participantLock = s.initiatorLock / 2
	}

	s.htlcs = make(map[string]*HTLCContract)
	s.swaps = make(map[string]*AtomicSwapRecord)
	s.offers = make(map[string]*SwapOffer)
	s.completedSwaps = 0
	s.failedSwaps = 0
	s.totalVolume = big.NewInt(0)

	s.initializeSampleOffers()
	s.updateState()
	return nil
}

// initializeSampleOffers 初始化示例订单
func (s *AtomicSwapSimulator) initializeSampleOffers() {
	offers := []struct {
		maker       string
		makerAsset  string
		makerAmount int64
		makerChain  string
		takerAsset  string
		takerAmount int64
		takerChain  string
	}{
		{"Alice", "BTC", 1, "Bitcoin", "ETH", 15, "Ethereum"},
		{"Bob", "ETH", 10, "Ethereum", "BTC", 1, "Bitcoin"},
		{"Charlie", "LTC", 100, "Litecoin", "ETH", 5, "Ethereum"},
	}

	for i, o := range offers {
		offerID := fmt.Sprintf("offer-%d", i+1)
		makerAmt := new(big.Int).Mul(big.NewInt(o.makerAmount), big.NewInt(1e8))
		takerAmt := new(big.Int).Mul(big.NewInt(o.takerAmount), big.NewInt(1e8))
		s.offers[offerID] = &SwapOffer{
			ID:           offerID,
			Maker:        o.maker,
			MakerAsset:   o.makerAsset,
			MakerAmount:  makerAmt,
			MakerChain:   o.makerChain,
			TakerAsset:   o.takerAsset,
			TakerAmount:  takerAmt,
			TakerChain:   o.takerChain,
			ExchangeRate: float64(o.takerAmount) / float64(o.makerAmount),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
			CreatedAt:    time.Now(),
		}
	}
}

// =============================================================================
// HTLC机制解释
// =============================================================================

// ExplainHTLC 解释HTLC机制
func (s *AtomicSwapSimulator) ExplainHTLC() map[string]interface{} {
	return map[string]interface{}{
		"name": "哈希时间锁合约 (Hash Time-Locked Contract)",
		"components": []map[string]interface{}{
			{
				"name":        "HashLock (哈希锁)",
				"description": "资金被hash(secret)锁定，需要提供原像secret才能解锁",
				"formula":     "hashLock = SHA256(secret)",
				"security":    "只有知道secret的人才能取出资金",
			},
			{
				"name":        "TimeLock (时间锁)",
				"description": "在指定时间之前，只能通过hashLock解锁；超时后发送方可退款",
				"purpose":     "防止资金永久锁定",
				"typical":     "发起方48小时，参与方24小时",
			},
		},
		"atomic_swap_protocol": []map[string]string{
			{"step": "1", "actor": "Alice", "action": "生成32字节随机secret，计算hash=SHA256(secret)"},
			{"step": "2", "actor": "Alice", "action": "在Bitcoin创建HTLC，锁定1BTC，设置hash和48小时时间锁"},
			{"step": "3", "actor": "Bob", "action": "验证Alice的HTLC参数，确认金额和hash"},
			{"step": "4", "actor": "Bob", "action": "在Ethereum创建HTLC，锁定15ETH，使用相同hash，24小时时间锁"},
			{"step": "5", "actor": "Alice", "action": "使用secret从Bob的HTLC取出15ETH（secret在链上公开）"},
			{"step": "6", "actor": "Bob", "action": "从链上获取secret，用它从Alice的HTLC取出1BTC"},
			{"step": "7", "actor": "Both", "action": "交换完成，Alice有15ETH，Bob有1BTC"},
		},
		"security_analysis": map[string]interface{}{
			"atomicity": "要么双方都完成交换，要么都取回资金",
			"trustless": "不依赖任何第三方托管或仲裁",
			"timelock_rule": map[string]string{
				"rule":    "发起方时间锁 > 参与方时间锁",
				"reason":  "确保参与方有足够时间在看到secret后认领资金",
				"example": "Alice: 48h, Bob: 24h → Bob有24小时窗口",
			},
			"failure_modes": []map[string]string{
				{"scenario": "Alice不揭示secret", "result": "超时后双方都可退款"},
				{"scenario": "Bob不创建HTLC", "result": "Alice超时后取回资金"},
				{"scenario": "网络延迟", "result": "时间锁差异提供安全边际"},
			},
		},
		"limitations": []string{
			"需要双方在线协调",
			"时间锁期间资金被锁定",
			"不支持部分成交",
			"汇率在锁定期间可能变化",
		},
	}
}

// =============================================================================
// 密钥生成
// =============================================================================

// GenerateSecret 生成秘密和哈希锁
func (s *AtomicSwapSimulator) GenerateSecret() (secret, hashLock string) {
	secretBytes := make([]byte, 32)
	seed := time.Now().UnixNano()
	for i := range secretBytes {
		seed = seed*1103515245 + 12345
		secretBytes[i] = byte(seed >> 16)
	}
	secret = hex.EncodeToString(secretBytes)

	hash := sha256.Sum256(secretBytes)
	hashLock = hex.EncodeToString(hash[:])

	return secret, hashLock
}

// VerifySecret 验证秘密
func (s *AtomicSwapSimulator) VerifySecret(secret, hashLock string) bool {
	secretBytes, err := hex.DecodeString(secret)
	if err != nil {
		return false
	}

	hash := sha256.Sum256(secretBytes)
	computed := hex.EncodeToString(hash[:])

	return computed == hashLock
}

// =============================================================================
// HTLC操作
// =============================================================================

// CreateHTLC 创建HTLC
func (s *AtomicSwapSimulator) CreateHTLC(sender, recipient, chain, token string, amount *big.Int, hashLock string, duration time.Duration) (*HTLCContract, error) {
	if amount.Sign() <= 0 {
		return nil, fmt.Errorf("金额必须大于0")
	}

	if len(hashLock) != 64 {
		return nil, fmt.Errorf("无效的hashLock长度: %d", len(hashLock))
	}

	htlcData := fmt.Sprintf("%s-%s-%s-%d", sender, chain, hashLock[:16], time.Now().UnixNano())
	htlcHash := sha256.Sum256([]byte(htlcData))
	htlcID := fmt.Sprintf("htlc-%s", hex.EncodeToString(htlcHash[:8]))
	contractAddr := fmt.Sprintf("0x%s", hex.EncodeToString(htlcHash[8:28]))
	txHash := fmt.Sprintf("0x%s", hex.EncodeToString(htlcHash[:]))

	htlc := &HTLCContract{
		ID:           htlcID,
		ContractAddr: contractAddr,
		Sender:       sender,
		Recipient:    recipient,
		Amount:       amount,
		Token:        token,
		HashLock:     hashLock,
		TimeLock:     time.Now().Add(duration),
		State:        HTLCStateLocked,
		Chain:        chain,
		TxHash:       txHash,
		CreatedAt:    time.Now(),
	}

	s.htlcs[htlcID] = htlc

	s.EmitEvent("htlc_created", "", "", map[string]interface{}{
		"htlc_id":       htlcID,
		"sender":        sender,
		"recipient":     recipient,
		"chain":         chain,
		"token":         token,
		"amount":        amount.String(),
		"hash_lock":     hashLock[:16] + "...",
		"time_lock":     htlc.TimeLock,
		"contract_addr": contractAddr,
	})

	s.updateState()
	return htlc, nil
}

// ClaimHTLC 认领HTLC
func (s *AtomicSwapSimulator) ClaimHTLC(htlcID, secret string) error {
	htlc, ok := s.htlcs[htlcID]
	if !ok {
		return fmt.Errorf("HTLC不存在: %s", htlcID)
	}

	if htlc.State != HTLCStateLocked {
		return fmt.Errorf("HTLC状态不是locked: %s", htlc.State)
	}

	if time.Now().After(htlc.TimeLock) {
		htlc.State = HTLCStateExpired
		return fmt.Errorf("HTLC已过期")
	}

	if !s.VerifySecret(secret, htlc.HashLock) {
		return fmt.Errorf("secret不匹配hashLock")
	}

	htlc.Secret = secret
	htlc.State = HTLCStateClaimed
	htlc.ClaimedAt = time.Now()

	claimData := fmt.Sprintf("%s-claim-%d", htlcID, time.Now().UnixNano())
	claimHash := sha256.Sum256([]byte(claimData))
	htlc.ClaimTxHash = fmt.Sprintf("0x%s", hex.EncodeToString(claimHash[:]))

	s.EmitEvent("htlc_claimed", "", "", map[string]interface{}{
		"htlc_id":       htlcID,
		"recipient":     htlc.Recipient,
		"secret":        secret[:16] + "...",
		"amount":        htlc.Amount.String(),
		"claim_tx_hash": htlc.ClaimTxHash,
	})

	s.updateState()
	return nil
}

// RefundHTLC 退款HTLC
func (s *AtomicSwapSimulator) RefundHTLC(htlcID string) error {
	htlc, ok := s.htlcs[htlcID]
	if !ok {
		return fmt.Errorf("HTLC不存在: %s", htlcID)
	}

	if htlc.State != HTLCStateLocked && htlc.State != HTLCStateExpired {
		return fmt.Errorf("HTLC状态不允许退款: %s", htlc.State)
	}

	if time.Now().Before(htlc.TimeLock) {
		return fmt.Errorf("时间锁未到期，剩余: %s", time.Until(htlc.TimeLock))
	}

	htlc.State = HTLCStateRefunded
	htlc.RefundedAt = time.Now()

	refundData := fmt.Sprintf("%s-refund-%d", htlcID, time.Now().UnixNano())
	refundHash := sha256.Sum256([]byte(refundData))
	htlc.RefundTxHash = fmt.Sprintf("0x%s", hex.EncodeToString(refundHash[:]))

	s.EmitEvent("htlc_refunded", "", "", map[string]interface{}{
		"htlc_id":        htlcID,
		"sender":         htlc.Sender,
		"amount":         htlc.Amount.String(),
		"refund_tx_hash": htlc.RefundTxHash,
	})

	s.updateState()
	return nil
}

// =============================================================================
// 原子交换操作
// =============================================================================

// InitiateSwap 发起原子交换
func (s *AtomicSwapSimulator) InitiateSwap(initiator, participant string,
	initiatorChain, initiatorToken string, initiatorAmount *big.Int,
	participantChain, participantToken string, participantAmount *big.Int) (*AtomicSwapRecord, string, error) {

	secret, hashLock := s.GenerateSecret()

	initiatorHTLC, err := s.CreateHTLC(
		initiator, participant, initiatorChain, initiatorToken,
		initiatorAmount, hashLock, s.initiatorLock,
	)
	if err != nil {
		return nil, "", fmt.Errorf("创建发起方HTLC失败: %v", err)
	}

	swapData := fmt.Sprintf("%s-%s-%d", initiator, participant, time.Now().UnixNano())
	swapHash := sha256.Sum256([]byte(swapData))
	swapID := fmt.Sprintf("swap-%s", hex.EncodeToString(swapHash[:8]))

	swap := &AtomicSwapRecord{
		ID:                swapID,
		Initiator:         initiator,
		Participant:       participant,
		InitiatorHTLC:     initiatorHTLC,
		InitiatorAsset:    fmt.Sprintf("%s on %s", initiatorToken, initiatorChain),
		ParticipantAsset:  fmt.Sprintf("%s on %s", participantToken, participantChain),
		InitiatorAmount:   initiatorAmount,
		ParticipantAmount: participantAmount,
		ExchangeRate:      float64(participantAmount.Int64()) / float64(initiatorAmount.Int64()),
		State:             HTLCStatePending,
		SecretHash:        hashLock,
		CreatedAt:         time.Now(),
	}

	s.swaps[swapID] = swap

	s.EmitEvent("swap_initiated", "", "", map[string]interface{}{
		"swap_id":            swapID,
		"initiator":          initiator,
		"participant":        participant,
		"initiator_asset":    swap.InitiatorAsset,
		"participant_asset":  swap.ParticipantAsset,
		"initiator_amount":   initiatorAmount.String(),
		"participant_amount": participantAmount.String(),
		"hash_lock":          hashLock[:16] + "...",
	})

	s.updateState()
	return swap, secret, nil
}

// ParticipateSwap 参与原子交换
func (s *AtomicSwapSimulator) ParticipateSwap(swapID string, chain, token string, amount *big.Int) error {
	swap, ok := s.swaps[swapID]
	if !ok {
		return fmt.Errorf("交换不存在: %s", swapID)
	}

	if swap.ParticipantHTLC != nil {
		return fmt.Errorf("参与方HTLC已创建")
	}

	if swap.State != HTLCStatePending {
		return fmt.Errorf("交换状态不是pending: %s", swap.State)
	}

	participantHTLC, err := s.CreateHTLC(
		swap.Participant, swap.Initiator, chain, token,
		amount, swap.SecretHash, s.participantLock,
	)
	if err != nil {
		return fmt.Errorf("创建参与方HTLC失败: %v", err)
	}

	swap.ParticipantHTLC = participantHTLC
	swap.State = HTLCStateLocked

	s.EmitEvent("swap_participated", "", "", map[string]interface{}{
		"swap_id":     swapID,
		"participant": swap.Participant,
		"htlc_id":     participantHTLC.ID,
		"chain":       chain,
		"amount":      amount.String(),
	})

	s.updateState()
	return nil
}

// CompleteSwap 完成原子交换
func (s *AtomicSwapSimulator) CompleteSwap(swapID, secret string) error {
	swap, ok := s.swaps[swapID]
	if !ok {
		return fmt.Errorf("交换不存在: %s", swapID)
	}

	if swap.ParticipantHTLC == nil {
		return fmt.Errorf("参与方HTLC未创建")
	}

	if swap.State != HTLCStateLocked {
		return fmt.Errorf("交换状态不是locked: %s", swap.State)
	}

	if err := s.ClaimHTLC(swap.ParticipantHTLC.ID, secret); err != nil {
		return fmt.Errorf("发起方认领失败: %v", err)
	}

	if err := s.ClaimHTLC(swap.InitiatorHTLC.ID, secret); err != nil {
		return fmt.Errorf("参与方认领失败: %v", err)
	}

	swap.Secret = secret
	swap.State = HTLCStateClaimed
	swap.CompletedAt = time.Now()
	s.completedSwaps++
	s.totalVolume.Add(s.totalVolume, swap.InitiatorAmount)
	s.totalVolume.Add(s.totalVolume, swap.ParticipantAmount)

	s.EmitEvent("swap_completed", "", "", map[string]interface{}{
		"swap_id":     swapID,
		"initiator":   swap.Initiator,
		"participant": swap.Participant,
		"duration":    swap.CompletedAt.Sub(swap.CreatedAt).String(),
	})

	s.updateState()
	return nil
}

// SimulateFullSwap 模拟完整原子交换流程
func (s *AtomicSwapSimulator) SimulateFullSwap() map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	btcAmount := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e8))
	ethAmount := new(big.Int).Mul(big.NewInt(15), big.NewInt(1e18))

	swap, secret, _ := s.InitiateSwap(
		"Alice", "Bob",
		"Bitcoin", "BTC", btcAmount,
		"Ethereum", "ETH", ethAmount,
	)
	steps = append(steps, map[string]interface{}{
		"step":      1,
		"actor":     "Alice",
		"action":    "生成secret，在Bitcoin创建HTLC锁定1BTC",
		"htlc_id":   swap.InitiatorHTLC.ID,
		"timelock":  s.initiatorLock.String(),
		"hash_lock": swap.SecretHash[:16] + "...",
	})

	steps = append(steps, map[string]interface{}{
		"step":   2,
		"actor":  "Bob",
		"action": "验证Alice的HTLC: 金额、hash、时间锁",
		"verify": []string{"金额=1BTC", "hash正确", "时间锁=48h"},
	})

	s.ParticipateSwap(swap.ID, "Ethereum", "ETH", ethAmount)
	steps = append(steps, map[string]interface{}{
		"step":     3,
		"actor":    "Bob",
		"action":   "在Ethereum创建HTLC锁定15ETH",
		"htlc_id":  swap.ParticipantHTLC.ID,
		"timelock": s.participantLock.String(),
	})

	steps = append(steps, map[string]interface{}{
		"step":   4,
		"actor":  "Alice",
		"action": "用secret从Bob的HTLC取出15ETH",
		"secret": secret[:16] + "...",
		"result": "secret在链上公开",
	})

	steps = append(steps, map[string]interface{}{
		"step":   5,
		"actor":  "Bob",
		"action": "从链上获取secret，用它从Alice的HTLC取出1BTC",
		"result": "交换完成",
	})

	s.CompleteSwap(swap.ID, secret)

	return map[string]interface{}{
		"swap_id":     swap.ID,
		"initiator":   "Alice",
		"participant": "Bob",
		"exchange":    "1 BTC <-> 15 ETH",
		"duration":    swap.CompletedAt.Sub(swap.CreatedAt).String(),
		"steps":       steps,
		"result":      "原子交换成功完成",
		"final_state": map[string]string{
			"Alice": "原有1BTC，现有15ETH",
			"Bob":   "原有15ETH，现有1BTC",
		},
	}
}

// SimulateFailedSwap 模拟失败的原子交换
func (s *AtomicSwapSimulator) SimulateFailedSwap(scenario string) map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	btcAmount := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e8))
	ethAmount := new(big.Int).Mul(big.NewInt(15), big.NewInt(1e18))

	swap, _, _ := s.InitiateSwap(
		"Alice", "Bob",
		"Bitcoin", "BTC", btcAmount,
		"Ethereum", "ETH", ethAmount,
	)

	switch scenario {
	case "participant_no_response":
		steps = append(steps, map[string]interface{}{
			"step": 1, "actor": "Alice", "action": "在Bitcoin创建HTLC锁定1BTC",
		})
		steps = append(steps, map[string]interface{}{
			"step": 2, "actor": "Bob", "action": "不响应，不创建HTLC",
		})
		steps = append(steps, map[string]interface{}{
			"step": 3, "actor": "Alice", "action": "等待48小时时间锁过期",
		})
		steps = append(steps, map[string]interface{}{
			"step": 4, "actor": "Alice", "action": "调用refund()取回1BTC",
		})
		return map[string]interface{}{
			"scenario": "参与方不响应",
			"result":   "Alice取回资金，无损失",
			"steps":    steps,
		}

	case "initiator_no_claim":
		s.ParticipateSwap(swap.ID, "Ethereum", "ETH", ethAmount)
		steps = append(steps, map[string]interface{}{
			"step": 1, "actor": "Alice", "action": "在Bitcoin创建HTLC锁定1BTC",
		})
		steps = append(steps, map[string]interface{}{
			"step": 2, "actor": "Bob", "action": "在Ethereum创建HTLC锁定15ETH",
		})
		steps = append(steps, map[string]interface{}{
			"step": 3, "actor": "Alice", "action": "不揭示secret，不认领ETH",
		})
		steps = append(steps, map[string]interface{}{
			"step": 4, "actor": "Bob", "action": "等待24小时时间锁过期，取回15ETH",
		})
		steps = append(steps, map[string]interface{}{
			"step": 5, "actor": "Alice", "action": "等待48小时时间锁过期，取回1BTC",
		})
		return map[string]interface{}{
			"scenario": "发起方不认领",
			"result":   "双方都取回资金，无损失",
			"steps":    steps,
		}
	}

	return map[string]interface{}{"error": "未知场景"}
}

// GetSwapInfo 获取交换信息
func (s *AtomicSwapSimulator) GetSwapInfo(swapID string) map[string]interface{} {
	swap, ok := s.swaps[swapID]
	if !ok {
		return nil
	}

	result := map[string]interface{}{
		"id":                 swap.ID,
		"initiator":          swap.Initiator,
		"participant":        swap.Participant,
		"initiator_asset":    swap.InitiatorAsset,
		"participant_asset":  swap.ParticipantAsset,
		"initiator_amount":   swap.InitiatorAmount.String(),
		"participant_amount": swap.ParticipantAmount.String(),
		"exchange_rate":      swap.ExchangeRate,
		"state":              string(swap.State),
		"created_at":         swap.CreatedAt,
	}

	if swap.InitiatorHTLC != nil {
		result["initiator_htlc"] = map[string]interface{}{
			"id":        swap.InitiatorHTLC.ID,
			"state":     string(swap.InitiatorHTLC.State),
			"time_lock": swap.InitiatorHTLC.TimeLock,
		}
	}

	if swap.ParticipantHTLC != nil {
		result["participant_htlc"] = map[string]interface{}{
			"id":        swap.ParticipantHTLC.ID,
			"state":     string(swap.ParticipantHTLC.State),
			"time_lock": swap.ParticipantHTLC.TimeLock,
		}
	}

	return result
}

// GetStatistics 获取统计信息
func (s *AtomicSwapSimulator) GetStatistics() map[string]interface{} {
	pendingSwaps := 0
	activeHTLCs := 0

	for _, swap := range s.swaps {
		if swap.State == HTLCStatePending || swap.State == HTLCStateLocked {
			pendingSwaps++
		}
	}

	for _, htlc := range s.htlcs {
		if htlc.State == HTLCStateLocked {
			activeHTLCs++
		}
	}

	return map[string]interface{}{
		"total_swaps":      len(s.swaps),
		"completed_swaps":  s.completedSwaps,
		"failed_swaps":     s.failedSwaps,
		"pending_swaps":    pendingSwaps,
		"total_htlcs":      len(s.htlcs),
		"active_htlcs":     activeHTLCs,
		"total_volume":     s.totalVolume.String(),
		"active_offers":    len(s.offers),
		"initiator_lock":   s.initiatorLock.String(),
		"participant_lock": s.participantLock.String(),
	}
}

// updateState 更新状态
func (s *AtomicSwapSimulator) updateState() {
	s.SetGlobalData("htlc_count", len(s.htlcs))
	s.SetGlobalData("swap_count", len(s.swaps))
	s.SetGlobalData("completed_swaps", s.completedSwaps)
	s.SetGlobalData("total_volume", s.totalVolume.String())

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"atomic_swap",
		"当前可以发起原子交换并观察 HTLC 如何在两条链上形成闭环。",
		"先发起一次交换，再观察秘密、哈希锁和退款路径如何配合保证原子性。",
		0,
		map[string]interface{}{
			"htlc_count":       len(s.htlcs),
			"swap_count":       len(s.swaps),
			"completed_swaps":  s.completedSwaps,
			"total_volume":     s.totalVolume.String(),
		},
	)
}

func (s *AtomicSwapSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "initiate_swap":
		swap, secret, err := s.InitiateSwap(
			"alice", "bob",
			"bitcoin", "BTC", big.NewInt(1),
			"litecoin", "LTC", big.NewInt(10),
		)
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已发起原子交换",
			map[string]interface{}{"swap_id": swap.ID, "secret": secret},
			&types.ActionFeedback{
				Summary:     "新的 HTLC 交换已经建立，可继续观察参与方锁定与赎回流程。",
				NextHint:    "继续参与交换或模拟退款，比较成功兑换与超时回退两条路径。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported atomic swap action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// AtomicSwapFactory 原子交换工厂
type AtomicSwapFactory struct{}

// Create 创建演示器
func (f *AtomicSwapFactory) Create() engine.Simulator {
	return NewAtomicSwapSimulator()
}

// GetDescription 获取描述
func (f *AtomicSwapFactory) GetDescription() types.Description {
	return NewAtomicSwapSimulator().GetDescription()
}

// NewAtomicSwapFactory 创建工厂
func NewAtomicSwapFactory() *AtomicSwapFactory {
	return &AtomicSwapFactory{}
}

var _ engine.SimulatorFactory = (*AtomicSwapFactory)(nil)
