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
// 轻客户端演示器
// 演示跨链轻客户端验证机制，是IBC等协议的核心组件
//
// 核心概念:
// 1. 区块头验证: 验证区块头的签名和连续性
// 2. 状态证明: 使用Merkle证明验证特定状态
// 3. 信任期: 验证者集合可信的时间窗口
// 4. 验证者集合更新: 处理验证者变更
//
// 安全假设:
// - 至少2/3的验证者是诚实的
// - 在信任期内至少同步一次
// - Merkle证明的密码学安全性
//
// 参考: IBC Light Client, ETH2 Light Client, Tendermint Light Client
// =============================================================================

// LCBlockHeader 轻客户端区块头
type LCBlockHeader struct {
	Height        uint64    `json:"height"`
	Hash          string    `json:"hash"`
	ParentHash    string    `json:"parent_hash"`
	StateRoot     string    `json:"state_root"`
	TxRoot        string    `json:"tx_root"`
	ReceiptsRoot  string    `json:"receipts_root"`
	ValidatorHash string    `json:"validator_hash"`
	NextValHash   string    `json:"next_validator_hash"`
	Timestamp     time.Time `json:"timestamp"`
	ProposerAddr  string    `json:"proposer_address"`
	Signatures    []string  `json:"signatures"`
	SignedPower   float64   `json:"signed_voting_power"`
}

// LCValidatorSet 验证者集合
type LCValidatorSet struct {
	Hash       string         `json:"hash"`
	TotalPower int64          `json:"total_power"`
	Validators []*LCValidator `json:"validators"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// LCValidator 验证者
type LCValidator struct {
	Address     string `json:"address"`
	PubKey      string `json:"public_key"`
	VotingPower int64  `json:"voting_power"`
	IsActive    bool   `json:"is_active"`
}

// LightClientState 轻客户端状态
type LightClientState struct {
	ChainID          string           `json:"chain_id"`
	ChainName        string           `json:"chain_name"`
	LatestHeight     uint64           `json:"latest_height"`
	LatestHash       string           `json:"latest_hash"`
	TrustedHeight    uint64           `json:"trusted_height"`
	TrustedHash      string           `json:"trusted_hash"`
	TrustingPeriod   time.Duration    `json:"trusting_period"`
	UnbondingPeriod  time.Duration    `json:"unbonding_period"`
	MaxClockDrift    time.Duration    `json:"max_clock_drift"`
	ValidatorSet     *LCValidatorSet  `json:"validator_set"`
	NextValidatorSet *LCValidatorSet  `json:"next_validator_set"`
	Headers          []*LCBlockHeader `json:"headers"`
	LastUpdated      time.Time        `json:"last_updated"`
	IsFrozen         bool             `json:"is_frozen"`
	FreezeReason     string           `json:"freeze_reason"`
}

// StateProof 状态证明
type StateProof struct {
	Key         string   `json:"key"`
	Value       string   `json:"value"`
	Height      uint64   `json:"height"`
	MerkleProof []string `json:"merkle_proof"`
	StateRoot   string   `json:"state_root"`
}

// VerificationResult 验证结果
type VerificationResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Height    uint64 `json:"height"`
	StateRoot string `json:"state_root"`
	ErrorCode string `json:"error_code,omitempty"`
}

// LightClientSimulator 轻客户端演示器
type LightClientSimulator struct {
	*base.BaseSimulator
	clients         map[string]*LightClientState
	trustingPeriod  time.Duration
	unbondingPeriod time.Duration
	maxClockDrift   time.Duration
}

// NewLightClientSimulator 创建轻客户端演示器
func NewLightClientSimulator() *LightClientSimulator {
	sim := &LightClientSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"light_client",
			"轻客户端演示器",
			"演示跨链轻客户端的区块头验证、状态证明验证、信任期管理等机制",
			"crosschain",
			types.ComponentProcess,
		),
		clients: make(map[string]*LightClientState),
	}

	sim.AddParam(types.Param{
		Key:         "trusting_period_days",
		Name:        "信任期(天)",
		Description: "轻客户端的信任期时长",
		Type:        types.ParamTypeInt,
		Default:     14,
		Min:         1,
		Max:         30,
	})

	sim.AddParam(types.Param{
		Key:         "unbonding_period_days",
		Name:        "解绑期(天)",
		Description: "验证者解绑所需时间",
		Type:        types.ParamTypeInt,
		Default:     21,
		Min:         7,
		Max:         30,
	})

	return sim
}

// Init 初始化
func (s *LightClientSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.trustingPeriod = 14 * 24 * time.Hour
	s.unbondingPeriod = 21 * 24 * time.Hour
	s.maxClockDrift = 10 * time.Second

	if v, ok := config.Params["trusting_period_days"]; ok {
		if n, ok := v.(float64); ok {
			s.trustingPeriod = time.Duration(n) * 24 * time.Hour
		}
	}
	if v, ok := config.Params["unbonding_period_days"]; ok {
		if n, ok := v.(float64); ok {
			s.unbondingPeriod = time.Duration(n) * 24 * time.Hour
		}
	}

	s.clients = make(map[string]*LightClientState)

	s.initializeClient("cosmos", "Cosmos Hub", 100)
	s.initializeClient("ethereum", "Ethereum", 32)
	s.initializeClient("osmosis", "Osmosis", 150)

	s.updateState()
	return nil
}

// initializeClient 初始化轻客户端
func (s *LightClientSimulator) initializeClient(chainID, chainName string, validatorCount int) {
	validators := make([]*LCValidator, validatorCount)
	totalPower := int64(0)

	for i := 0; i < validatorCount; i++ {
		power := int64(100000 - i*500)
		if power < 10000 {
			power = 10000
		}
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s-val-%d", chainID, i)))
		validators[i] = &LCValidator{
			Address:     fmt.Sprintf("0x%s", hex.EncodeToString(hash[:20])),
			PubKey:      hex.EncodeToString(hash[:]),
			VotingPower: power,
			IsActive:    true,
		}
		totalPower += power
	}

	valSetHash := sha256.Sum256([]byte(fmt.Sprintf("%s-valset-%d", chainID, time.Now().UnixNano())))
	valSet := &LCValidatorSet{
		Hash:       hex.EncodeToString(valSetHash[:]),
		TotalPower: totalPower,
		Validators: validators,
		UpdatedAt:  time.Now(),
	}

	genesisHeader := s.createBlockHeader(chainID, 1, "0x0000", valSet.Hash)

	client := &LightClientState{
		ChainID:          chainID,
		ChainName:        chainName,
		LatestHeight:     1,
		LatestHash:       genesisHeader.Hash,
		TrustedHeight:    1,
		TrustedHash:      genesisHeader.Hash,
		TrustingPeriod:   s.trustingPeriod,
		UnbondingPeriod:  s.unbondingPeriod,
		MaxClockDrift:    s.maxClockDrift,
		ValidatorSet:     valSet,
		NextValidatorSet: valSet,
		Headers:          []*LCBlockHeader{genesisHeader},
		LastUpdated:      time.Now(),
		IsFrozen:         false,
	}

	s.clients[chainID] = client
}

// createBlockHeader 创建区块头
func (s *LightClientSimulator) createBlockHeader(chainID string, height uint64, parentHash, valHash string) *LCBlockHeader {
	data := fmt.Sprintf("%s-%d-%s-%d", chainID, height, parentHash, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	stateRoot := sha256.Sum256([]byte(data + "-state"))
	txRoot := sha256.Sum256([]byte(data + "-tx"))
	receiptsRoot := sha256.Sum256([]byte(data + "-receipts"))

	return &LCBlockHeader{
		Height:        height,
		Hash:          hex.EncodeToString(hash[:]),
		ParentHash:    parentHash,
		StateRoot:     hex.EncodeToString(stateRoot[:]),
		TxRoot:        hex.EncodeToString(txRoot[:]),
		ReceiptsRoot:  hex.EncodeToString(receiptsRoot[:]),
		ValidatorHash: valHash,
		NextValHash:   valHash,
		Timestamp:     time.Now(),
		ProposerAddr:  "0xproposer",
		Signatures:    []string{"sig1", "sig2", "sig3"},
		SignedPower:   0.75,
	}
}

// =============================================================================
// 轻客户端机制解释
// =============================================================================

// ExplainLightClient 解释轻客户端
func (s *LightClientSimulator) ExplainLightClient() map[string]interface{} {
	return map[string]interface{}{
		"purpose": "无需运行完整节点即可验证区块链状态的组件",
		"vs_full_node": map[string]interface{}{
			"full_node": map[string]string{
				"storage":   "完整区块链数据(数百GB)",
				"bandwidth": "持续同步所有区块和交易",
				"security":  "完全独立验证所有交易",
			},
			"light_client": map[string]string{
				"storage":   "仅区块头(<1GB)",
				"bandwidth": "仅同步区块头",
				"security":  "依赖验证者诚实假设(2/3+)",
			},
		},
		"core_components": []map[string]string{
			{"name": "区块头", "desc": "包含StateRoot、TxRoot等摘要信息"},
			{"name": "验证者集合", "desc": "当前负责出块的验证者列表及投票权"},
			{"name": "信任期", "desc": "验证者集合可信的时间窗口"},
			{"name": "Merkle证明", "desc": "证明特定数据存在于状态树中"},
		},
		"verification_process": []string{
			"1. 获取目标链的区块头",
			"2. 验证区块头签名(需要2/3+投票权)",
			"3. 验证区块头与父块的连续性",
			"4. 检查时间戳合理性(时钟漂移限制)",
			"5. 检查信任期是否过期",
			"6. 使用StateRoot验证Merkle证明",
		},
		"trust_model": map[string]interface{}{
			"assumption":       "至少2/3的验证者是诚实的",
			"trusting_period":  s.trustingPeriod.String(),
			"unbonding_period": s.unbondingPeriod.String(),
			"relationship":     "trusting_period < unbonding_period (确保作恶可被惩罚)",
		},
		"attack_vectors": []map[string]string{
			{"attack": "长程攻击", "desc": "获取旧私钥创建分叉", "defense": "信任期+弱主观性"},
			{"attack": "日蚀攻击", "desc": "隔离节点提供虚假数据", "defense": "多源验证"},
			{"attack": "信任期过期", "desc": "超过信任期不更新", "defense": "定期同步"},
		},
	}
}

// =============================================================================
// 轻客户端操作
// =============================================================================

// UpdateClient 更新轻客户端
func (s *LightClientSimulator) UpdateClient(chainID string, header *LCBlockHeader) (*VerificationResult, error) {
	client, ok := s.clients[chainID]
	if !ok {
		return nil, fmt.Errorf("轻客户端不存在: %s", chainID)
	}

	if client.IsFrozen {
		return &VerificationResult{
			Success:   false,
			Message:   "客户端已冻结: " + client.FreezeReason,
			ErrorCode: "CLIENT_FROZEN",
		}, nil
	}

	result := s.verifyHeader(client, header)
	if !result.Success {
		s.EmitEvent("header_rejected", "", "", map[string]interface{}{
			"chain_id": chainID,
			"height":   header.Height,
			"reason":   result.Message,
		})
		return result, nil
	}

	client.Headers = append(client.Headers, header)
	client.LatestHeight = header.Height
	client.LatestHash = header.Hash
	client.LastUpdated = time.Now()

	if header.Height > client.TrustedHeight+10 {
		client.TrustedHeight = header.Height
		client.TrustedHash = header.Hash
	}

	s.EmitEvent("client_updated", "", "", map[string]interface{}{
		"chain_id":       chainID,
		"new_height":     header.Height,
		"trusted_height": client.TrustedHeight,
		"signed_power":   header.SignedPower,
	})

	s.updateState()
	return result, nil
}

// verifyHeader 验证区块头
func (s *LightClientSimulator) verifyHeader(client *LightClientState, header *LCBlockHeader) *VerificationResult {
	if header.Height != client.LatestHeight+1 {
		return &VerificationResult{
			Success:   false,
			Message:   fmt.Sprintf("高度不连续: 期望%d, 收到%d", client.LatestHeight+1, header.Height),
			Height:    header.Height,
			ErrorCode: "HEIGHT_MISMATCH",
		}
	}

	if header.ParentHash != client.LatestHash {
		return &VerificationResult{
			Success:   false,
			Message:   "父哈希不匹配",
			Height:    header.Height,
			ErrorCode: "PARENT_HASH_MISMATCH",
		}
	}

	if header.SignedPower < 0.6667 {
		return &VerificationResult{
			Success:   false,
			Message:   fmt.Sprintf("签名投票权不足: %.2f%% < 66.67%%", header.SignedPower*100),
			Height:    header.Height,
			ErrorCode: "INSUFFICIENT_VOTING_POWER",
		}
	}

	lastHeader := client.Headers[len(client.Headers)-1]
	if header.Timestamp.Before(lastHeader.Timestamp) {
		return &VerificationResult{
			Success:   false,
			Message:   "时间戳早于上一区块",
			Height:    header.Height,
			ErrorCode: "TIMESTAMP_INVALID",
		}
	}

	if time.Since(client.LastUpdated) > client.TrustingPeriod {
		return &VerificationResult{
			Success:   false,
			Message:   "信任期已过期，需要从信任源重新初始化",
			Height:    header.Height,
			ErrorCode: "TRUSTING_PERIOD_EXPIRED",
		}
	}

	return &VerificationResult{
		Success:   true,
		Message:   "验证通过",
		Height:    header.Height,
		StateRoot: header.StateRoot,
	}
}

// VerifyStateProof 验证状态证明
func (s *LightClientSimulator) VerifyStateProof(chainID string, proof *StateProof) (*VerificationResult, error) {
	client, ok := s.clients[chainID]
	if !ok {
		return nil, fmt.Errorf("轻客户端不存在: %s", chainID)
	}

	var targetHeader *LCBlockHeader
	for _, h := range client.Headers {
		if h.Height == proof.Height {
			targetHeader = h
			break
		}
	}

	if targetHeader == nil {
		return &VerificationResult{
			Success:   false,
			Message:   fmt.Sprintf("区块头不存在: height=%d", proof.Height),
			ErrorCode: "HEADER_NOT_FOUND",
		}, nil
	}

	if proof.StateRoot != targetHeader.StateRoot {
		return &VerificationResult{
			Success:   false,
			Message:   "StateRoot不匹配",
			ErrorCode: "STATE_ROOT_MISMATCH",
		}, nil
	}

	proofValid := len(proof.MerkleProof) > 0

	s.EmitEvent("state_proof_verified", "", "", map[string]interface{}{
		"chain_id":   chainID,
		"height":     proof.Height,
		"key":        proof.Key,
		"valid":      proofValid,
		"state_root": proof.StateRoot[:16] + "...",
	})

	if proofValid {
		return &VerificationResult{
			Success:   true,
			Message:   "状态证明验证通过",
			Height:    proof.Height,
			StateRoot: proof.StateRoot,
		}, nil
	}

	return &VerificationResult{
		Success:   false,
		Message:   "Merkle证明验证失败",
		ErrorCode: "MERKLE_PROOF_INVALID",
	}, nil
}

// FreezeClient 冻结客户端(检测到作恶)
func (s *LightClientSimulator) FreezeClient(chainID, reason string) error {
	client, ok := s.clients[chainID]
	if !ok {
		return fmt.Errorf("轻客户端不存在: %s", chainID)
	}

	client.IsFrozen = true
	client.FreezeReason = reason

	s.EmitEvent("client_frozen", "", "", map[string]interface{}{
		"chain_id": chainID,
		"reason":   reason,
	})

	s.updateState()
	return nil
}

// SimulateBlockSync 模拟区块同步
func (s *LightClientSimulator) SimulateBlockSync(chainID string, numBlocks int) map[string]interface{} {
	client, ok := s.clients[chainID]
	if !ok {
		return map[string]interface{}{"error": "客户端不存在"}
	}

	results := make([]map[string]interface{}, 0)

	for i := 0; i < numBlocks; i++ {
		header := s.createBlockHeader(chainID, client.LatestHeight+1, client.LatestHash, client.ValidatorSet.Hash)
		result, _ := s.UpdateClient(chainID, header)

		results = append(results, map[string]interface{}{
			"height":  header.Height,
			"hash":    header.Hash[:16] + "...",
			"success": result.Success,
			"message": result.Message,
		})

		if !result.Success {
			break
		}
	}

	return map[string]interface{}{
		"chain_id":       chainID,
		"blocks_synced":  len(results),
		"latest_height":  client.LatestHeight,
		"trusted_height": client.TrustedHeight,
		"results":        results,
	}
}

// GetClientState 获取客户端状态
func (s *LightClientSimulator) GetClientState(chainID string) map[string]interface{} {
	client, ok := s.clients[chainID]
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"chain_id":         client.ChainID,
		"chain_name":       client.ChainName,
		"latest_height":    client.LatestHeight,
		"latest_hash":      client.LatestHash[:16] + "...",
		"trusted_height":   client.TrustedHeight,
		"trusting_period":  client.TrustingPeriod.String(),
		"unbonding_period": client.UnbondingPeriod.String(),
		"headers_stored":   len(client.Headers),
		"validator_count":  len(client.ValidatorSet.Validators),
		"total_power":      client.ValidatorSet.TotalPower,
		"is_frozen":        client.IsFrozen,
		"last_updated":     client.LastUpdated,
	}
}

// GetStatistics 获取统计
func (s *LightClientSimulator) GetStatistics() map[string]interface{} {
	totalHeaders := 0
	frozenClients := 0
	for _, client := range s.clients {
		totalHeaders += len(client.Headers)
		if client.IsFrozen {
			frozenClients++
		}
	}

	return map[string]interface{}{
		"client_count":    len(s.clients),
		"total_headers":   totalHeaders,
		"frozen_clients":  frozenClients,
		"trusting_period": s.trustingPeriod.String(),
	}
}

// updateState 更新状态
func (s *LightClientSimulator) updateState() {
	s.SetGlobalData("client_count", len(s.clients))

	totalHeaders := 0
	for _, client := range s.clients {
		totalHeaders += len(client.Headers)
	}
	s.SetGlobalData("total_headers", totalHeaders)

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"light_client",
		"当前可以更新轻客户端区块头，并验证状态证明。",
		"先推进一轮 header 更新，再验证状态证明，观察信任边界如何影响跨链验证。",
		0,
		map[string]interface{}{
			"client_count":  len(s.clients),
			"total_headers": totalHeaders,
		},
	)
}

func (s *LightClientSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "update_client":
		chainID := "ethereum"
		client, ok := s.clients[chainID]
		if !ok {
			for id := range s.clients {
				chainID = id
				client = s.clients[id]
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("没有可更新的轻客户端")
		}
		header := &LCBlockHeader{
			Height:        client.LatestHeight + 1,
			Hash:          client.LatestHash + "-next",
			ParentHash:    client.LatestHash,
			StateRoot:     client.TrustedHash,
			TxRoot:        "tx-root",
			ReceiptsRoot:  "receipts-root",
			ValidatorHash: client.ValidatorSet.Hash,
			NextValHash:   client.NextValidatorSet.Hash,
			Timestamp:     time.Now(),
			ProposerAddr:  "validator-1",
			Signatures:    []string{"sig-1", "sig-2", "sig-3"},
			SignedPower:   0.8,
		}
		result, err := s.UpdateClient(chainID, header)
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已更新轻客户端头部",
			map[string]interface{}{"chain_id": chainID, "height": header.Height, "success": result.Success},
			&types.ActionFeedback{
				Summary:     "新的区块头已经纳入轻客户端验证范围，可继续验证状态证明。",
				NextHint:    "执行 verify_state_proof，观察轻客户端如何用状态根完成验证。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported light client action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// LightClientFactory 轻客户端工厂
type LightClientFactory struct{}

func (f *LightClientFactory) Create() engine.Simulator { return NewLightClientSimulator() }
func (f *LightClientFactory) GetDescription() types.Description {
	return NewLightClientSimulator().GetDescription()
}
func NewLightClientFactory() *LightClientFactory { return &LightClientFactory{} }

var _ engine.SimulatorFactory = (*LightClientFactory)(nil)
