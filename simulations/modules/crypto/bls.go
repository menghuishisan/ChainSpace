package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================

// BLS签名数据结构

// =============================================================================

// BLSKeyPair BLS密钥对
// BLS使用双线性配对的椭圆曲线，这里使用模拟实现来演示概念
type BLSKeyPair struct {
	ID         string    `json:"id"`          // 密钥ID
	PrivateKey *big.Int  `json:"-"`           // 私钥 sk ∈ Z_r
	PublicKey  *big.Int  `json:"-"`           // 公钥 pk = sk * G2
	PrivKeyHex string    `json:"private_key"` // 私钥十六进制
	PubKeyHex  string    `json:"public_key"`  // 公钥十六进制
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
}

// BLSSignature BLS签名
// σ = sk * H(m)，其中H是hash-to-curve函数
type BLSSignature struct {
	Message      string    `json:"message"`   // 原始消息
	Hash         string    `json:"hash"`      // 消息哈希
	Signature    *big.Int  `json:"-"`         // 签名值
	SignatureHex string    `json:"signature"` // 签名十六进制
	SignerID     string    `json:"signer_id"` // 签名者ID
	Timestamp    time.Time `json:"timestamp"` // 时间戳
}

// AggregatedSignature 聚合签名
// BLS的核心特性：多个签名可以聚合为一个固定大小的签名
type AggregatedSignature struct {
	Messages     []string  `json:"messages"`   // 消息列表
	SignerIDs    []string  `json:"signer_ids"` // 签名者列表
	AggSig       *big.Int  `json:"-"`          // 聚合签名 σ_agg = Σσ_i
	AggSigHex    string    `json:"agg_sig"`    // 聚合签名十六进制
	AggPubKey    *big.Int  `json:"-"`          // 聚合公钥 pk_agg = Σpk_i
	AggPubKeyHex string    `json:"agg_pubkey"` // 聚合公钥十六进制
	SignCount    int       `json:"sign_count"` // 签名数量
	Timestamp    time.Time `json:"timestamp"`  // 时间戳
}

// BLSRecord 操作记录
type BLSRecord struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // keygen/sign/aggregate/verify
	SignerID  string    `json:"signer_id"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================

// BLSSimulator BLS签名演示器

// =============================================================================

// BLSSimulator BLS签名演示器
//
// BLS签名特性：
// 1. 签名聚合: 多个签名可以压缩为一个签名 σ_agg = σ_1 + σ_2 + ... + σ_n
// 2. 公钥聚合: 多个公钥可以压缩为一个公钥 pk_agg = pk_1 + pk_2 + ... + pk_n
// 3. 批量验证: 可以一次验证多个签名
// 4. 签名大小恒定: 无论聚合多少签名，结果大小不变
//
// 验证原理 (使用双线性配对e)：
// 单签名验证: e(σ, G2) == e(H(m), pk)
// 聚合验证: e(σ_agg, G2) == e(H(m1), pk1) * e(H(m2), pk2) * ...
//
// 应用场景：
// - 区块链共识(以太坊2.0验证者签名)
// - 多重签名钱包
// - 证书聚合
type BLSSimulator struct {
	*base.BaseSimulator
	mu         sync.RWMutex
	keyPairs   map[string]*BLSKeyPair // 密钥对
	signatures []*BLSSignature        // 签名列表
	aggregated []*AggregatedSignature // 聚合签名历史
	history    []*BLSRecord           // 操作记录
	order      *big.Int               // 群阶 r
	generator  *big.Int               // 生成元 G
}

// NewBLSSimulator 创建BLS签名演示器
func NewBLSSimulator() *BLSSimulator {
	sim := &BLSSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"bls",
			"BLS签名演示器",
			"实现BLS签名的聚合特性，支持签名聚合和批量验证",
			"crypto",
			types.ComponentTool,
		),
		keyPairs:   make(map[string]*BLSKeyPair),
		signatures: make([]*BLSSignature, 0),
		aggregated: make([]*AggregatedSignature, 0),
		history:    make([]*BLSRecord, 0),
	}

	sim.AddParam(types.Param{
		Key:         "signer_count",
		Name:        "签名者数量",
		Description: "初始化的签名者数量",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         2,
		Max:         20,
	})

	return sim
}

// Init 初始化演示器
func (s *BLSSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 生成群参数
	// 使用256位素数作为群阶
	s.order, _ = rand.Prime(rand.Reader, 256)
	s.generator = big.NewInt(2)

	signerCount := 5
	if v, ok := config.Params["signer_count"]; ok {
		if n, ok := v.(float64); ok {
			signerCount = int(n)
		}
	}

	// 生成密钥对
	for i := 0; i < signerCount; i++ {
		s.GenerateKeyPair(fmt.Sprintf("signer-%d", i+1))
	}

	s.updateState()
	return nil
}

// =============================================================================

// BLS核心实现

// =============================================================================

// GenerateKeyPair 生成BLS密钥对
//
// 密钥生成：
// 1. 随机选择私钥 sk ∈ Z_r
// 2. 计算公钥 pk = sk * G2 (G2是G2群的生成元)
func (s *BLSSimulator) GenerateKeyPair(id string) (*BLSKeyPair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.keyPairs[id]; exists {
		return nil, fmt.Errorf("密钥已存在: %s", id)
	}

	// 生成随机私钥 sk ∈ [1, r-1]
	sk, err := rand.Int(rand.Reader, new(big.Int).Sub(s.order, big.NewInt(1)))
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}
	sk.Add(sk, big.NewInt(1))

	// 计算公钥 pk = g^sk mod order (模拟椭圆曲线标量乘法)
	pk := new(big.Int).Exp(s.generator, sk, s.order)

	keyPair := &BLSKeyPair{
		ID:         id,
		PrivateKey: sk,
		PublicKey:  pk,
		PrivKeyHex: hex.EncodeToString(sk.Bytes()),
		PubKeyHex:  hex.EncodeToString(pk.Bytes()),
		CreatedAt:  time.Now(),
	}

	s.keyPairs[id] = keyPair

	// 记录
	s.history = append(s.history, &BLSRecord{
		ID:        fmt.Sprintf("keygen-%d", len(s.history)+1),
		Type:      "keygen",
		SignerID:  id,
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("keypair_generated", "", "", map[string]interface{}{
		"id":         id,
		"public_key": keyPair.PubKeyHex[:16] + "...",
	})

	s.updateState()
	return keyPair, nil
}

// Sign 签名消息
//
// BLS签名：
// 1. 计算消息哈希点 H(m) (hash-to-curve)
// 2. 计算签名 σ = sk * H(m)
func (s *BLSSimulator) Sign(signerID, message string) (*BLSSignature, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keyPair := s.keyPairs[signerID]
	if keyPair == nil {
		return nil, fmt.Errorf("签名者不存在: %s", signerID)
	}

	// 步骤1: 计算消息哈希 (模拟hash-to-curve)
	// 真实BLS使用hash-to-curve将消息映射到G1群上的点
	msgHash := sha256.Sum256([]byte(message))
	hashPoint := new(big.Int).SetBytes(msgHash[:])
	hashPoint.Mod(hashPoint, s.order)

	// 步骤2: 计算签名 σ = sk * H(m)
	// 在真实BLS中，这是G1群上的标量乘法
	signature := new(big.Int).Mul(keyPair.PrivateKey, hashPoint)
	signature.Mod(signature, s.order)

	sig := &BLSSignature{
		Message:      message,
		Hash:         hex.EncodeToString(msgHash[:]),
		Signature:    signature,
		SignatureHex: hex.EncodeToString(signature.Bytes()),
		SignerID:     signerID,
		Timestamp:    time.Now(),
	}

	s.signatures = append(s.signatures, sig)

	// 记录
	s.history = append(s.history, &BLSRecord{
		ID:        fmt.Sprintf("sign-%d", len(s.history)+1),
		Type:      "sign",
		SignerID:  signerID,
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("signed", types.NodeID(signerID), "", map[string]interface{}{
		"message":   message,
		"signature": sig.SignatureHex[:16] + "...",
	})

	s.updateState()
	return sig, nil
}

// VerifySingle 验证单个签名
//
// 验证: e(σ, G2) == e(H(m), pk)
// 使用双线性配对性质: e(a*P, Q) = e(P, a*Q) = e(P, Q)^a
func (s *BLSSimulator) VerifySingle(sig *BLSSignature) bool {
	if sig == nil {
		return false
	}

	s.mu.RLock()
	keyPair := s.keyPairs[sig.SignerID]
	s.mu.RUnlock()

	if keyPair == nil {
		return false
	}

	// 重新计算消息哈希点
	msgHash := sha256.Sum256([]byte(sig.Message))
	hashPoint := new(big.Int).SetBytes(msgHash[:])
	hashPoint.Mod(hashPoint, s.order)

	// 验证: σ == sk * H(m)
	// 等价于验证: σ * g^(-sk) == H(m) (简化的验证)
	// 或者直接比较: σ / H(m) == sk (但这会泄露私钥)
	//
	// 真实BLS验证使用配对: e(σ, G2) == e(H(m), pk)
	// 这里我们使用模运算模拟
	expected := new(big.Int).Mul(keyPair.PrivateKey, hashPoint)
	expected.Mod(expected, s.order)

	valid := sig.Signature.Cmp(expected) == 0

	s.EmitEvent("single_verified", "", "", map[string]interface{}{
		"signer_id": sig.SignerID,
		"valid":     valid,
	})

	return valid
}

// Aggregate 聚合多个签名
//
// BLS签名聚合：
// σ_agg = σ_1 + σ_2 + ... + σ_n
//
// 这是BLS的核心优势: n个签名压缩为1个，大小不变
func (s *BLSSimulator) Aggregate(sigs []*BLSSignature) (*AggregatedSignature, error) {
	if len(sigs) < 2 {
		return nil, fmt.Errorf("至少需要2个签名进行聚合")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 聚合签名: σ_agg = Σσ_i
	aggSig := big.NewInt(0)
	messages := make([]string, 0, len(sigs))
	signerIDs := make([]string, 0, len(sigs))

	for _, sig := range sigs {
		if sig.Signature == nil {
			sigBytes, _ := hex.DecodeString(sig.SignatureHex)
			sig.Signature = new(big.Int).SetBytes(sigBytes)
		}
		aggSig.Add(aggSig, sig.Signature)
		messages = append(messages, sig.Message)
		signerIDs = append(signerIDs, sig.SignerID)
	}
	aggSig.Mod(aggSig, s.order)

	// 聚合公钥: pk_agg = Σpk_i
	aggPubKey := big.NewInt(0)
	for _, signerID := range signerIDs {
		if kp := s.keyPairs[signerID]; kp != nil {
			aggPubKey.Add(aggPubKey, kp.PublicKey)
		}
	}
	aggPubKey.Mod(aggPubKey, s.order)

	result := &AggregatedSignature{
		Messages:     messages,
		SignerIDs:    signerIDs,
		AggSig:       aggSig,
		AggSigHex:    hex.EncodeToString(aggSig.Bytes()),
		AggPubKey:    aggPubKey,
		AggPubKeyHex: hex.EncodeToString(aggPubKey.Bytes()),
		SignCount:    len(sigs),
		Timestamp:    time.Now(),
	}

	s.aggregated = append(s.aggregated, result)

	// 记录
	s.history = append(s.history, &BLSRecord{
		ID:        fmt.Sprintf("agg-%d", len(s.history)+1),
		Type:      "aggregate",
		SignerID:  fmt.Sprintf("%d signers", len(sigs)),
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("aggregated", "", "", map[string]interface{}{
		"sign_count":        len(sigs),
		"agg_sig":           result.AggSigHex[:16] + "...",
		"compression_ratio": fmt.Sprintf("%dx", len(sigs)),
	})

	s.updateState()
	return result, nil
}

// VerifyAggregated 验证聚合签名
//
// 聚合验证 (相同消息的情况)：
// e(σ_agg, G2) == e(H(m), pk_agg)
//
// 不同消息的情况：
// e(σ_agg, G2) == Π e(H(m_i), pk_i)
func (s *BLSSimulator) VerifyAggregated(aggSig *AggregatedSignature) bool {
	if aggSig == nil || len(aggSig.SignerIDs) == 0 {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 计算预期的聚合签名
	expectedAgg := big.NewInt(0)

	for i, signerID := range aggSig.SignerIDs {
		kp := s.keyPairs[signerID]
		if kp == nil {
			return false
		}

		// 计算每个消息的哈希点
		var msgHash [32]byte
		if i < len(aggSig.Messages) {
			msgHash = sha256.Sum256([]byte(aggSig.Messages[i]))
		} else {
			msgHash = sha256.Sum256([]byte(aggSig.Messages[0]))
		}
		hashPoint := new(big.Int).SetBytes(msgHash[:])
		hashPoint.Mod(hashPoint, s.order)

		// 计算预期的单个签名
		expected := new(big.Int).Mul(kp.PrivateKey, hashPoint)
		expected.Mod(expected, s.order)

		expectedAgg.Add(expectedAgg, expected)
	}
	expectedAgg.Mod(expectedAgg, s.order)

	// 比较聚合签名
	actualAgg := aggSig.AggSig
	if actualAgg == nil {
		sigBytes, _ := hex.DecodeString(aggSig.AggSigHex)
		actualAgg = new(big.Int).SetBytes(sigBytes)
	}

	valid := actualAgg.Cmp(expectedAgg) == 0

	// 记录
	s.history = append(s.history, &BLSRecord{
		ID:        fmt.Sprintf("verify-%d", len(s.history)+1),
		Type:      "verify",
		SignerID:  fmt.Sprintf("%d signers", len(aggSig.SignerIDs)),
		Success:   valid,
		Timestamp: time.Now(),
	})

	s.EmitEvent("agg_verified", "", "", map[string]interface{}{
		"valid":      valid,
		"sign_count": aggSig.SignCount,
	})

	return valid
}

// GetCompressionRatio 获取压缩比
// BLS聚合签名的优势: n个签名压缩为1个
func (s *BLSSimulator) GetCompressionRatio(sigCount int) float64 {
	// BLS签名约48字节(G1点), 聚合后仍是48字节
	singleSigSize := 48
	originalSize := sigCount * singleSigSize
	aggregatedSize := singleSigSize
	return float64(originalSize) / float64(aggregatedSize)
}

// GetBandwidthSaving 获取带宽节省
func (s *BLSSimulator) GetBandwidthSaving(sigCount int) string {
	ratio := s.GetCompressionRatio(sigCount)
	saving := (1 - 1/ratio) * 100
	return fmt.Sprintf("%.1f%%", saving)
}

// updateState 更新状态
func (s *BLSSimulator) updateState() {
	signerList := make([]map[string]interface{}, 0)
	for id, kp := range s.keyPairs {
		signerList = append(signerList, map[string]interface{}{
			"id":         id,
			"public_key": kp.PubKeyHex[:16] + "...",
		})
	}

	s.SetGlobalData("keypair_count", len(s.keyPairs))
	s.SetGlobalData("signature_count", len(s.signatures))
	s.SetGlobalData("aggregated_count", len(s.aggregated))
	s.SetGlobalData("signers", signerList)
	s.SetGlobalData("history_count", len(s.history))

	summary := fmt.Sprintf("当前有 %d 个 BLS 签名参与者，已生成 %d 条单签名和 %d 条聚合签名。", len(s.keyPairs), len(s.signatures), len(s.aggregated))
	nextHint := "先生成签名者，再创建单签名和聚合签名，观察带宽压缩效果。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备聚合签名",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"keypair_count": len(s.keyPairs), "signature_count": len(s.signatures), "aggregated_count": len(s.aggregated)},
	)
}

func (s *BLSSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_signer":
		signerID := fmt.Sprintf("signer-%d", len(s.keyPairs)+1)
		if raw, ok := params["signer_id"].(string); ok && raw != "" {
			signerID = raw
		}
		keyPair, err := s.GenerateKeyPair(signerID)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已创建一个 BLS 签名者。", map[string]interface{}{"signer_id": signerID, "public_key": keyPair.PubKeyHex}, &types.ActionFeedback{
			Summary:     "新的 BLS 签名者已经准备就绪，可以继续生成单签名或聚合签名。",
			NextHint:    "继续对多位签名者生成同一消息的签名，再观察聚合签名的带宽优势。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"signer_id": signerID, "keypair_count": len(s.keyPairs)},
		}), nil
	case "sign_message":
		signerID := ""
		for id := range s.keyPairs {
			signerID = id
			break
		}
		if raw, ok := params["signer_id"].(string); ok && raw != "" {
			signerID = raw
		}
		if signerID == "" {
			return nil, fmt.Errorf("no signer available")
		}
		message := "ChainSpace"
		if raw, ok := params["message"].(string); ok && raw != "" {
			message = raw
		}
		sig, err := s.Sign(signerID, message)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已完成一次 BLS 签名。", map[string]interface{}{"signature": sig}, &types.ActionFeedback{
			Summary:     "单个签名者已经生成 BLS 签名，可以继续聚合多位签名者的结果。",
			NextHint:    "继续对多个签名者生成签名，再观察聚合后的压缩效果。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"signer_id": signerID, "signature_count": len(s.signatures)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported bls action: %s", action)
	}
}

// =============================================================================

// 工厂

// =============================================================================

// BLSFactory BLS签名演示器工厂
type BLSFactory struct{}

// Create 创建演示器实例
func (f *BLSFactory) Create() engine.Simulator {
	return NewBLSSimulator()
}

// GetDescription 获取描述
func (f *BLSFactory) GetDescription() types.Description {
	return NewBLSSimulator().GetDescription()
}

// NewBLSFactory 创建工厂实例
func NewBLSFactory() *BLSFactory {
	return &BLSFactory{}
}

var _ engine.SimulatorFactory = (*BLSFactory)(nil)
