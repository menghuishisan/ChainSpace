package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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

// ==============================================================================
// 环签名数据结构
// ==============================================================================

// RingMember 环成员
// 每个成员持有一个公钥，只有真正的签名者知道对应私钥
type RingMember struct {
	ID         string            `json:"id"`         // 成员ID
	PublicKey  *ecdsa.PublicKey  `json:"-"`          // 公钥
	PrivateKey *ecdsa.PrivateKey `json:"-"`          // 私钥(仅签名者有)
	PubKeyHex  string            `json:"public_key"` // 公钥十六进制
	Index      int               `json:"index"`      // 环中索引
}

// RingSignatureData 环签名数据
// 基于Spontaneous Anonymous Group (SAG) 签名方案
type RingSignatureData struct {
	Message    string    `json:"message"`     // 原始消息
	Hash       string    `json:"hash"`        // 消息哈希
	RingSize   int       `json:"ring_size"`   // 环大小
	KeyImage   string    `json:"key_image"`   // 密钥镜像(链接标签)
	C0         string    `json:"c0"`          // 初始挑战值
	S          []string  `json:"s"`           // 响应值数组
	PublicKeys []string  `json:"public_keys"` // 环中的公钥
	Timestamp  time.Time `json:"timestamp"`   // 时间戳
}

// RingRecord 操作记录
type RingRecord struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // sign/verify
	RingSize  int       `json:"ring_size"`
	KeyImage  string    `json:"key_image"`
	Success   bool      `json:"success"`   // 验证结果
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// ==============================================================================
// RingSigSimulator 环签名演示器
// ==============================================================================

// RingSigSimulator 环签名演示器
// 实现基于椭圆曲线的环签名方案
//
// 环签名特性:
// 1. 匿名性: 验证者无法确定环中哪个成员是真正的签名者
// 2. 不可伪造性: 只有环中某个成员的私钥持有者才能生成有效签名
// 3. 链接性(可选): 通过密钥镜像可以检测同一私钥的多次签名
//
// 应用场景:
// - 隐私加密货币(如Monero)
// - 匿名投票
// - 举报人保护
type RingSigSimulator struct {
	*base.BaseSimulator
	mu         sync.RWMutex
	curve      elliptic.Curve       // 椭圆曲线
	members    []*RingMember        // 环成员
	signerIdx  int                  // 真正签名者的索引
	signatures []*RingSignatureData // 签名历史
	history    []*RingRecord        // 操作记录
	keyImages  map[string]int       // 密钥镜像使用计数
}

// NewRingSigSimulator 创建环签名演示器
func NewRingSigSimulator() *RingSigSimulator {
	sim := &RingSigSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"ring_sig",
			"环签名演示器",
			"实现椭圆曲线环签名，展示匿名性和链接性特性",
			"crypto",
			types.ComponentTool,
		),
		curve:      elliptic.P256(),
		members:    make([]*RingMember, 0),
		signatures: make([]*RingSignatureData, 0),
		history:    make([]*RingRecord, 0),
		keyImages:  make(map[string]int),
	}

	sim.AddParam(types.Param{
		Key:         "ring_size",
		Name:        "环大小",
		Description: "环中公钥的数量(匿名集大小)",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         3,
		Max:         20,
	})

	return sim
}

// Init 初始化演示器
func (s *RingSigSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	ringSize := 5
	if v, ok := config.Params["ring_size"]; ok {
		if n, ok := v.(float64); ok {
			ringSize = int(n)
		}
	}

	s.setupRing(ringSize)
	s.updateState()
	return nil
}

// setupRing 设置环
// 生成环成员的密钥对，随机选择一个作为真正的签名者
func (s *RingSigSimulator) setupRing(size int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.members = make([]*RingMember, 0, size)

	// 随机选择真正的签名者索引
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	s.signerIdx = int(new(big.Int).SetBytes(randBytes).Int64()) % size
	if s.signerIdx < 0 {
		s.signerIdx = -s.signerIdx
	}

	for i := 0; i < size; i++ {
		key, _ := ecdsa.GenerateKey(s.curve, rand.Reader)
		pubBytes := elliptic.Marshal(s.curve, key.PublicKey.X, key.PublicKey.Y)

		member := &RingMember{
			ID:        fmt.Sprintf("member-%d", i+1),
			PublicKey: &key.PublicKey,
			PubKeyHex: hex.EncodeToString(pubBytes),
			Index:     i,
		}

		// 只有真正的签名者保存私钥
		if i == s.signerIdx {
			member.PrivateKey = key
		}

		s.members = append(s.members, member)
	}

	s.EmitEvent("ring_setup", "", "", map[string]interface{}{
		"ring_size":     size,
		"anonymity_set": size,
		"note":          "真正的签名者身份已隐藏",
	})
}

// ==============================================================================
// 环签名核心实现 (SAG方案)
// ==============================================================================

// Sign 生成环签名
//
// SAG签名算法步骤:
// 1. 计算消息哈希 m = H(message)
// 2. 计算密钥镜像 I = x * H_p(P) (x是私钥，P是公钥)
// 3. 选择随机数 α，计算 L_π = α*G
// 4. 计算初始挑战 c_{π+1} = H(m, L_π)
// 5. 对于 i = π+1 到 n-1, 0 到 π-1:
//   - 选择随机 s_i
//   - 计算 L_i = s_i*G + c_i*P_i
//   - 计算 c_{i+1} = H(m, L_i)
//
// 6. 计算 s_π = α - c_π * x mod n
// 7. 签名为 (I, c_0, s_0, s_1, ..., s_{n-1})
func (s *RingSigSimulator) Sign(message string) (*RingSignatureData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.members) == 0 {
		return nil, fmt.Errorf("环未初始化")
	}

	signer := s.members[s.signerIdx]
	if signer.PrivateKey == nil {
		return nil, fmt.Errorf("签名者私钥不可用")
	}

	n := len(s.members)
	curveN := s.curve.Params().N

	// 步骤1: 计算消息哈希
	msgHash := sha256.Sum256([]byte(message))

	// 步骤2: 计算密钥镜像 (简化版: I = H(x || P))
	// 真实实现需要 hash-to-curve
	keyImageInput := append(signer.PrivateKey.D.Bytes(), elliptic.Marshal(s.curve, signer.PublicKey.X, signer.PublicKey.Y)...)
	keyImageHash := sha256.Sum256(keyImageInput)
	keyImage := hex.EncodeToString(keyImageHash[:])

	// 步骤3: 选择随机数 α
	alpha, _ := rand.Int(rand.Reader, curveN)

	// 计算 L_π = α * G
	Lx, Ly := s.curve.ScalarBaseMult(alpha.Bytes())

	// 初始化响应数组和挑战数组
	sValues := make([]*big.Int, n)
	cValues := make([]*big.Int, n)

	// 步骤4: 计算 c_{π+1}
	cInput := append(msgHash[:], Lx.Bytes()...)
	cInput = append(cInput, Ly.Bytes()...)
	cHash := sha256.Sum256(cInput)
	cValues[(s.signerIdx+1)%n] = new(big.Int).SetBytes(cHash[:])
	cValues[(s.signerIdx+1)%n].Mod(cValues[(s.signerIdx+1)%n], curveN)

	// 步骤5: 环遍历，从 π+1 到 π-1
	for i := 1; i < n; i++ {
		idx := (s.signerIdx + i) % n
		nextIdx := (idx + 1) % n

		// 选择随机 s_i
		sValues[idx], _ = rand.Int(rand.Reader, curveN)

		// 计算 L_i = s_i * G + c_i * P_i
		// s_i * G
		sGx, sGy := s.curve.ScalarBaseMult(sValues[idx].Bytes())
		// c_i * P_i
		cPx, cPy := s.curve.ScalarMult(
			s.members[idx].PublicKey.X,
			s.members[idx].PublicKey.Y,
			cValues[idx].Bytes(),
		)
		// L_i = s_i * G + c_i * P_i
		Lx, Ly = s.curve.Add(sGx, sGy, cPx, cPy)

		// 计算 c_{i+1} = H(m, L_i)
		if nextIdx != s.signerIdx {
			cInput = append(msgHash[:], Lx.Bytes()...)
			cInput = append(cInput, Ly.Bytes()...)
			cHash = sha256.Sum256(cInput)
			cValues[nextIdx] = new(big.Int).SetBytes(cHash[:])
			cValues[nextIdx].Mod(cValues[nextIdx], curveN)
		}
	}

	// 步骤6: 计算 s_π = α - c_π * x mod n
	// 首先需要计算 c_π (环的最后一步)
	lastIdx := (s.signerIdx + n - 1) % n
	sGx, sGy := s.curve.ScalarBaseMult(sValues[lastIdx].Bytes())
	cPx, cPy := s.curve.ScalarMult(
		s.members[lastIdx].PublicKey.X,
		s.members[lastIdx].PublicKey.Y,
		cValues[lastIdx].Bytes(),
	)
	Lx, Ly = s.curve.Add(sGx, sGy, cPx, cPy)
	cInput = append(msgHash[:], Lx.Bytes()...)
	cInput = append(cInput, Ly.Bytes()...)
	cHash = sha256.Sum256(cInput)
	cValues[s.signerIdx] = new(big.Int).SetBytes(cHash[:])
	cValues[s.signerIdx].Mod(cValues[s.signerIdx], curveN)

	// s_π = α - c_π * x mod n
	cx := new(big.Int).Mul(cValues[s.signerIdx], signer.PrivateKey.D)
	sValues[s.signerIdx] = new(big.Int).Sub(alpha, cx)
	sValues[s.signerIdx].Mod(sValues[s.signerIdx], curveN)

	// 构建签名
	sHexValues := make([]string, n)
	pubKeyHexes := make([]string, n)
	for i := 0; i < n; i++ {
		sHexValues[i] = hex.EncodeToString(sValues[i].Bytes())
		pubKeyHexes[i] = s.members[i].PubKeyHex
	}

	sig := &RingSignatureData{
		Message:    message,
		Hash:       hex.EncodeToString(msgHash[:]),
		RingSize:   n,
		KeyImage:   keyImage,
		C0:         hex.EncodeToString(cValues[0].Bytes()),
		S:          sHexValues,
		PublicKeys: pubKeyHexes,
		Timestamp:  time.Now(),
	}

	s.signatures = append(s.signatures, sig)
	s.keyImages[keyImage]++

	// 记录
	s.history = append(s.history, &RingRecord{
		ID:        fmt.Sprintf("sign-%d", len(s.history)+1),
		Type:      "sign",
		RingSize:  n,
		KeyImage:  keyImage,
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("ring_signed", "", "", map[string]interface{}{
		"ring_size": n,
		"key_image": keyImage[:16] + "...",
		"message":   message,
	})

	s.updateState()
	return sig, nil
}

// Verify 验证环签名
//
// 验证步骤:
// 1. 解析签名中的公钥和响应值
// 2. 从 c_0 开始，计算每个 L_i 和 c_{i+1}
// 3. 检查最终计算的 c_n 是否等于 c_0
func (s *RingSigSimulator) Verify(sig *RingSignatureData) bool {
	if sig == nil || len(sig.S) == 0 || len(sig.PublicKeys) == 0 {
		return false
	}

	n := len(sig.S)
	if n != len(sig.PublicKeys) {
		return false
	}

	curveN := s.curve.Params().N

	// 解析消息哈希
	msgHash, _ := hex.DecodeString(sig.Hash)

	// 解析 c_0
	c0Bytes, _ := hex.DecodeString(sig.C0)
	c := new(big.Int).SetBytes(c0Bytes)

	// 解析公钥
	publicKeys := make([]*ecdsa.PublicKey, n)
	for i, pkHex := range sig.PublicKeys {
		pkBytes, _ := hex.DecodeString(pkHex)
		x, y := elliptic.Unmarshal(s.curve, pkBytes)
		if x == nil {
			return false
		}
		publicKeys[i] = &ecdsa.PublicKey{Curve: s.curve, X: x, Y: y}
	}

	// 验证环
	for i := 0; i < n; i++ {
		// 解析 s_i
		sBytes, _ := hex.DecodeString(sig.S[i])
		si := new(big.Int).SetBytes(sBytes)

		// 计算 L_i = s_i * G + c * P_i
		sGx, sGy := s.curve.ScalarBaseMult(si.Bytes())
		cPx, cPy := s.curve.ScalarMult(publicKeys[i].X, publicKeys[i].Y, c.Bytes())
		Lx, Ly := s.curve.Add(sGx, sGy, cPx, cPy)

		// 计算 c_{i+1} = H(m, L_i)
		cInput := append(msgHash, Lx.Bytes()...)
		cInput = append(cInput, Ly.Bytes()...)
		cHash := sha256.Sum256(cInput)
		c = new(big.Int).SetBytes(cHash[:])
		c.Mod(c, curveN)
	}

	// 验证 c_n == c_0
	c0 := new(big.Int).SetBytes(c0Bytes)
	valid := c.Cmp(c0) == 0

	// 记录
	s.mu.Lock()
	s.history = append(s.history, &RingRecord{
		ID:        fmt.Sprintf("verify-%d", len(s.history)+1),
		Type:      "verify",
		RingSize:  n,
		KeyImage:  sig.KeyImage,
		Success:   valid,
		Timestamp: time.Now(),
	})
	s.mu.Unlock()

	s.EmitEvent("ring_verified", "", "", map[string]interface{}{
		"valid":     valid,
		"ring_size": n,
	})

	return valid
}

// CheckLinkability 检查链接性(双花检测)
// 通过密钥镜像检测同一私钥的多次签名
func (s *RingSigSimulator) CheckLinkability(keyImage string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := s.keyImages[keyImage]
	isLinked := count > 1

	s.EmitEvent("linkability_check", "", "", map[string]interface{}{
		"key_image":   keyImage[:16] + "...",
		"usage_count": count,
		"is_linked":   isLinked,
	})

	return count, isLinked
}

// GetAnonymitySet 获取匿名集大小
func (s *RingSigSimulator) GetAnonymitySet() int {
	return len(s.members)
}

// updateState 更新状态
func (s *RingSigSimulator) updateState() {
	memberList := make([]map[string]interface{}, 0)
	for _, m := range s.members {
		memberList = append(memberList, map[string]interface{}{
			"id":         m.ID,
			"index":      m.Index,
			"public_key": m.PubKeyHex[:32] + "...",
		})
	}

	s.SetGlobalData("ring_size", len(s.members))
	s.SetGlobalData("anonymity_set", len(s.members))
	s.SetGlobalData("members", memberList)
	s.SetGlobalData("signature_count", len(s.signatures))
	s.SetGlobalData("history_count", len(s.history))

	summary := fmt.Sprintf("当前环大小为 %d，已生成 %d 条环签名。", len(s.members), len(s.signatures))
	nextHint := "可以继续生成环签名并检查密钥镜像，观察匿名性与可链接性的平衡。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备环签名",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"ring_size": len(s.members), "signature_count": len(s.signatures)},
	)
}

func (s *RingSigSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "generate_signature":
		message := "ChainSpace"
		if raw, ok := params["message"].(string); ok && raw != "" {
			message = raw
		}
		sig, err := s.Sign(message)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已生成一条环签名。", map[string]interface{}{"signature": sig}, &types.ActionFeedback{
			Summary:     "新的环签名已经生成，签名者身份被隐藏在整个环中。",
			NextHint:    "继续执行链接性检查，观察同一密钥镜像是否会暴露重复使用。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"ring_size": len(s.members), "signature_count": len(s.signatures)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported ring signature action: %s", action)
	}
}

// ==============================================================================
// 工厂
// ==============================================================================

// RingSigFactory 环签名演示器工厂
type RingSigFactory struct{}

// Create 创建演示器实例
func (f *RingSigFactory) Create() engine.Simulator {
	return NewRingSigSimulator()
}

// GetDescription 获取描述
func (f *RingSigFactory) GetDescription() types.Description {
	return NewRingSigSimulator().GetDescription()
}

// NewRingSigFactory 创建工厂实例
func NewRingSigFactory() *RingSigFactory {
	return &RingSigFactory{}
}

var _ engine.SimulatorFactory = (*RingSigFactory)(nil)
