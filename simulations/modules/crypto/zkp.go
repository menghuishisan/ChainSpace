package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 零知识证明数据结构
// =============================================================================

// SchnorrProof Schnorr协议零知识证明
//
// Schnorr协议是一个经典的Σ-协议，用于证明"我知道离散对数x，使得y = g^x mod p"
// 协议满足三个性质:
// 1. 完备性(Completeness): 诚实的证明者总能说服验证者
// 2. 可靠性(Soundness): 欺骗者成功概率可忽略
// 3. 零知识性(Zero-Knowledge): 验证者学不到x的任何信息
type SchnorrProof struct {
	T         *big.Int  `json:"-"`         // 承诺 t = g^r mod p
	C         *big.Int  `json:"-"`         // 挑战 c = H(g, y, t)
	Z         *big.Int  `json:"-"`         // 响应 z = r + c*x mod q
	THex      string    `json:"t"`         // 承诺十六进制
	CHex      string    `json:"c"`         // 挑战十六进制
	ZHex      string    `json:"z"`         // 响应十六进制
	ProverID  string    `json:"prover_id"` // 证明者ID
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// ZKPParams Schnorr协议公共参数
// 使用RFC 3526定义的安全参数
type ZKPParams struct {
	P    *big.Int `json:"-"` // 大素数 p
	Q    *big.Int `json:"-"` // 素数阶 q，满足 q | (p-1)
	G    *big.Int `json:"-"` // 生成元 g，阶为q
	PHex string   `json:"p"` // p的十六进制
	QHex string   `json:"q"` // q的十六进制
	GHex string   `json:"g"` // g的十六进制
}

// ZKPProver 证明者
type ZKPProver struct {
	ID         string    `json:"id"`         // 证明者ID
	Secret     *big.Int  `json:"-"`          // 私密值 x
	PublicY    *big.Int  `json:"-"`          // 公开值 y = g^x mod p
	PublicYHex string    `json:"public_y"`   // 公开值十六进制
	CreatedAt  time.Time `json:"created_at"` // 创建时间
}

// ZKPRecord 操作记录
type ZKPRecord struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // generate/verify/simulate
	ProverID  string    `json:"prover_id"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================
// ZKPSimulator 零知识证明演示器
// =============================================================================

// ZKPSimulator 零知识证明演示器
// 实现完整的Schnorr协议，包括:
// - 参数生成: 生成安全的群参数
// - 证明生成: 使用Fiat-Shamir变换的非交互式证明
// - 证明验证: 验证证明的正确性
// - 模拟器: 演示零知识性
type ZKPSimulator struct {
	*base.BaseSimulator
	mu      sync.RWMutex
	params  *ZKPParams            // 公共参数
	provers map[string]*ZKPProver // 证明者
	proofs  []*SchnorrProof       // 证明历史
	history []*ZKPRecord          // 操作记录
}

// NewZKPSimulator 创建零知识证明演示器
func NewZKPSimulator() *ZKPSimulator {
	sim := &ZKPSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"zkp",
			"零知识证明演示器",
			"实现Schnorr协议，演示完备性、可靠性和零知识性三大性质",
			"crypto",
			types.ComponentTool,
		),
		provers: make(map[string]*ZKPProver),
		proofs:  make([]*SchnorrProof, 0),
		history: make([]*ZKPRecord, 0),
	}

	sim.AddParam(types.Param{
		Key:         "security_bits",
		Name:        "安全级别",
		Description: "群参数的位数",
		Type:        types.ParamTypeSelect,
		Default:     "256",
		Options: []types.Option{
			{Label: "128位 (演示)", Value: "128"},
			{Label: "256位 (标准)", Value: "256"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *ZKPSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 生成安全参数
	bits := 256
	if v, ok := config.Params["security_bits"]; ok {
		if str, ok := v.(string); ok && str == "128" {
			bits = 128
		}
	}
	s.params = s.generateSecureParams(bits)

	// 创建示例证明者
	s.CreateProver("alice")
	s.CreateProver("bob")

	s.updateState()
	return nil
}

// generateSecureParams 生成安全的群参数
// 生成安全素数 p = 2q + 1，其中q也是素数
// 这确保了Schnorr群的阶为素数q
func (s *ZKPSimulator) generateSecureParams(bits int) *ZKPParams {
	var p, q *big.Int
	var err error

	// 生成安全素数
	// 循环直到找到满足条件的素数对
	for {
		// 生成随机素数q
		q, err = rand.Prime(rand.Reader, bits-1)
		if err != nil {
			continue
		}

		// 计算 p = 2q + 1
		p = new(big.Int).Mul(q, big.NewInt(2))
		p.Add(p, big.NewInt(1))

		// 验证p是否为素数
		if p.ProbablyPrime(20) {
			break
		}
	}

	// 寻找生成元g
	// g必须满足: g^q = 1 mod p 且 g ≠ 1
	// 对于安全素数，g = h^2 mod p (h是任意非零元素) 是阶为q的生成元
	var g *big.Int
	for h := big.NewInt(2); ; h.Add(h, big.NewInt(1)) {
		// g = h^2 mod p
		g = new(big.Int).Exp(h, big.NewInt(2), p)
		if g.Cmp(big.NewInt(1)) != 0 {
			// 验证g的阶是q
			gq := new(big.Int).Exp(g, q, p)
			if gq.Cmp(big.NewInt(1)) == 0 {
				break
			}
		}
	}

	return &ZKPParams{
		P:    p,
		Q:    q,
		G:    g,
		PHex: hex.EncodeToString(p.Bytes()),
		QHex: hex.EncodeToString(q.Bytes()),
		GHex: hex.EncodeToString(g.Bytes()),
	}
}

// CreateProver 创建证明者
// 生成随机私密值x，计算公开值 y = g^x mod p
func (s *ZKPSimulator) CreateProver(id string) (*ZKPProver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.provers[id]; exists {
		return nil, fmt.Errorf("证明者已存在: %s", id)
	}

	// 生成随机私密值 x ∈ [1, q-1]
	secret, err := rand.Int(rand.Reader, new(big.Int).Sub(s.params.Q, big.NewInt(1)))
	if err != nil {
		return nil, fmt.Errorf("生成私密值失败: %v", err)
	}
	secret.Add(secret, big.NewInt(1)) // 确保 x >= 1

	// 计算公开值 y = g^x mod p
	publicY := new(big.Int).Exp(s.params.G, secret, s.params.P)

	prover := &ZKPProver{
		ID:         id,
		Secret:     secret,
		PublicY:    publicY,
		PublicYHex: hex.EncodeToString(publicY.Bytes()),
		CreatedAt:  time.Now(),
	}

	s.provers[id] = prover

	s.EmitEvent("prover_created", "", "", map[string]interface{}{
		"id":       id,
		"public_y": prover.PublicYHex[:32] + "...",
	})

	s.updateState()
	return prover, nil
}

// GenerateProof 生成Schnorr零知识证明
//
// 协议步骤 (使用Fiat-Shamir变换实现非交互式):
// 1. 承诺: 选择随机数 r ∈ [1, q-1]，计算 t = g^r mod p
// 2. 挑战: 计算 c = H(g || y || t) mod q (Fiat-Shamir变换)
// 3. 响应: 计算 z = r + c*x mod q
//
// 证明为 (t, c, z)
func (s *ZKPSimulator) GenerateProof(proverID string) (*SchnorrProof, error) {
	s.mu.RLock()
	prover := s.provers[proverID]
	s.mu.RUnlock()

	if prover == nil {
		return nil, fmt.Errorf("证明者不存在: %s", proverID)
	}

	// 步骤1: 承诺阶段
	// 选择随机数 r ∈ [1, q-1]
	r, err := rand.Int(rand.Reader, new(big.Int).Sub(s.params.Q, big.NewInt(1)))
	if err != nil {
		return nil, fmt.Errorf("生成随机数失败: %v", err)
	}
	r.Add(r, big.NewInt(1))

	// 计算承诺 t = g^r mod p
	t := new(big.Int).Exp(s.params.G, r, s.params.P)

	// 步骤2: 挑战阶段 (Fiat-Shamir变换)
	// c = H(g || y || t) mod q
	// 这将交互式协议转换为非交互式
	hashInput := make([]byte, 0)
	hashInput = append(hashInput, s.params.G.Bytes()...)
	hashInput = append(hashInput, prover.PublicY.Bytes()...)
	hashInput = append(hashInput, t.Bytes()...)
	hashOutput := sha256.Sum256(hashInput)
	c := new(big.Int).SetBytes(hashOutput[:])
	c.Mod(c, s.params.Q)

	// 步骤3: 响应阶段
	// z = r + c*x mod q
	z := new(big.Int).Mul(c, prover.Secret)
	z.Add(z, r)
	z.Mod(z, s.params.Q)

	proof := &SchnorrProof{
		T:         t,
		C:         c,
		Z:         z,
		THex:      hex.EncodeToString(t.Bytes()),
		CHex:      hex.EncodeToString(c.Bytes()),
		ZHex:      hex.EncodeToString(z.Bytes()),
		ProverID:  proverID,
		Timestamp: time.Now(),
	}

	// 记录
	s.mu.Lock()
	s.proofs = append(s.proofs, proof)
	s.history = append(s.history, &ZKPRecord{
		ID:        fmt.Sprintf("gen-%d", len(s.history)+1),
		Type:      "generate",
		ProverID:  proverID,
		Success:   true,
		Timestamp: time.Now(),
	})
	s.mu.Unlock()

	s.EmitEvent("proof_generated", types.NodeID(proverID), "", map[string]interface{}{
		"t": proof.THex[:16] + "...",
		"c": proof.CHex[:16] + "...",
		"z": proof.ZHex[:16] + "...",
	})

	s.updateState()
	return proof, nil
}

// VerifyProof 验证Schnorr零知识证明
//
// 验证步骤:
// 1. 重新计算挑战: c' = H(g || y || t) mod q
// 2. 验证挑战相等: c' == c
// 3. 验证等式: g^z == t * y^c mod p
//
// 数学推导 (为什么验证等式成立):
// g^z = g^(r + c*x) = g^r * g^(c*x) = g^r * (g^x)^c = t * y^c mod p
func (s *ZKPSimulator) VerifyProof(proof *SchnorrProof, publicYHex string) bool {
	if proof == nil || publicYHex == "" {
		return false
	}

	// 解析公开值y
	yBytes, err := hex.DecodeString(publicYHex)
	if err != nil {
		return false
	}
	y := new(big.Int).SetBytes(yBytes)

	// 解析证明值
	t := proof.T
	c := proof.C
	z := proof.Z
	if t == nil || c == nil || z == nil {
		tBytes, _ := hex.DecodeString(proof.THex)
		cBytes, _ := hex.DecodeString(proof.CHex)
		zBytes, _ := hex.DecodeString(proof.ZHex)
		t = new(big.Int).SetBytes(tBytes)
		c = new(big.Int).SetBytes(cBytes)
		z = new(big.Int).SetBytes(zBytes)
	}

	// 步骤1: 重新计算挑战
	hashInput := make([]byte, 0)
	hashInput = append(hashInput, s.params.G.Bytes()...)
	hashInput = append(hashInput, y.Bytes()...)
	hashInput = append(hashInput, t.Bytes()...)
	hashOutput := sha256.Sum256(hashInput)
	cPrime := new(big.Int).SetBytes(hashOutput[:])
	cPrime.Mod(cPrime, s.params.Q)

	// 步骤2: 验证挑战相等
	if c.Cmp(cPrime) != 0 {
		s.recordVerification(proof.ProverID, false)
		return false
	}

	// 步骤3: 验证等式 g^z == t * y^c mod p
	// 左边: g^z mod p
	left := new(big.Int).Exp(s.params.G, z, s.params.P)

	// 右边: t * y^c mod p
	yc := new(big.Int).Exp(y, c, s.params.P)
	right := new(big.Int).Mul(t, yc)
	right.Mod(right, s.params.P)

	valid := left.Cmp(right) == 0

	s.recordVerification(proof.ProverID, valid)

	s.EmitEvent("proof_verified", "", "", map[string]interface{}{
		"prover_id": proof.ProverID,
		"valid":     valid,
		"equation":  "g^z == t * y^c mod p",
	})

	return valid
}

// SimulateProof 模拟证明 (演示零知识性)
//
// 零知识性的关键: 验证者可以自己生成与真实证明不可区分的证明
// 这证明了协议不会泄露任何关于x的信息
//
// 模拟步骤 (注意顺序与真实证明相反):
// 1. 随机选择 z 和 c
// 2. 计算 t = g^z * y^(-c) mod p
// 3. 输出 (t, c, z)
//
// 这个模拟的证明可以通过验证，但模拟者并不知道x!
func (s *ZKPSimulator) SimulateProof(publicYHex string) (*SchnorrProof, error) {
	// 解析公开值y
	yBytes, err := hex.DecodeString(publicYHex)
	if err != nil {
		return nil, fmt.Errorf("无效的公开值: %v", err)
	}
	y := new(big.Int).SetBytes(yBytes)

	// 步骤1: 随机选择 z 和 c
	z, _ := rand.Int(rand.Reader, s.params.Q)
	c, _ := rand.Int(rand.Reader, s.params.Q)

	// 步骤2: 计算 t = g^z * y^(-c) mod p
	// y^(-c) = y^(q-c) mod p (因为y的阶是q)
	negC := new(big.Int).Sub(s.params.Q, c)
	yNegC := new(big.Int).Exp(y, negC, s.params.P)
	gz := new(big.Int).Exp(s.params.G, z, s.params.P)
	t := new(big.Int).Mul(gz, yNegC)
	t.Mod(t, s.params.P)

	proof := &SchnorrProof{
		T:         t,
		C:         c,
		Z:         z,
		THex:      hex.EncodeToString(t.Bytes()),
		CHex:      hex.EncodeToString(c.Bytes()),
		ZHex:      hex.EncodeToString(z.Bytes()),
		ProverID:  "simulator",
		Timestamp: time.Now(),
	}

	// 记录
	s.mu.Lock()
	s.history = append(s.history, &ZKPRecord{
		ID:        fmt.Sprintf("sim-%d", len(s.history)+1),
		Type:      "simulate",
		ProverID:  "simulator",
		Success:   true,
		Timestamp: time.Now(),
	})
	s.mu.Unlock()

	s.EmitEvent("proof_simulated", "", "", map[string]interface{}{
		"note": "模拟器不知道秘密x，但生成的证明可以通过验证",
	})

	s.updateState()
	return proof, nil
}

// recordVerification 记录验证结果
func (s *ZKPSimulator) recordVerification(proverID string, valid bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, &ZKPRecord{
		ID:        fmt.Sprintf("ver-%d", len(s.history)+1),
		Type:      "verify",
		ProverID:  proverID,
		Success:   valid,
		Timestamp: time.Now(),
	})
	s.updateState()
}

// GetProver 获取证明者信息
func (s *ZKPSimulator) GetProver(id string) *ZKPProver {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provers[id]
}

// updateState 更新状态
func (s *ZKPSimulator) updateState() {
	proverList := make([]map[string]interface{}, 0)
	for id, p := range s.provers {
		proverList = append(proverList, map[string]interface{}{
			"id":       id,
			"public_y": p.PublicYHex[:32] + "...",
		})
	}

	s.SetGlobalData("p", s.params.PHex[:32]+"...")
	s.SetGlobalData("q", s.params.QHex[:32]+"...")
	s.SetGlobalData("g", s.params.GHex)
	s.SetGlobalData("provers", proverList)
	s.SetGlobalData("proof_count", len(s.proofs))
	s.SetGlobalData("history_count", len(s.history))

	summary := fmt.Sprintf("当前有 %d 个证明者，已生成 %d 份证明。", len(s.provers), len(s.proofs))
	nextHint := "先创建证明者，再生成 Schnorr 证明。"
	if len(s.proofs) > 0 {
		nextHint = "可以继续验证最新证明，观察验证方如何确认其正确性。"
	}

	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备证明",
		summary,
		nextHint,
		0.4,
		map[string]interface{}{
			"prover_count": len(s.provers),
			"proof_count":  len(s.proofs),
		},
	)
}

// ExecuteAction 执行零知识证明演示器的教学动作。
func (s *ZKPSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_prover":
		proverID := getStringParam(params, "prover_id", fmt.Sprintf("prover-%d", len(s.provers)+1))
		prover, err := s.CreateProver(proverID)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult(
			"已创建证明者",
			map[string]interface{}{
				"prover_id": proverID,
				"public_y":  prover.PublicYHex,
			},
			&types.ActionFeedback{
				Summary:     "证明者已生成秘密值和对应公开值，能够继续发起 Schnorr 证明。",
				NextHint:    "下一步可以生成证明，观察承诺、挑战和响应如何协同工作。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"prover_id": proverID},
			},
		), nil
	case "generate_proof":
		proverID := getStringParam(params, "prover_id", "")
		if strings.TrimSpace(proverID) == "" {
			for id := range s.provers {
				proverID = id
				break
			}
		}
		if strings.TrimSpace(proverID) == "" {
			return nil, fmt.Errorf("no prover available")
		}
		proof, err := s.GenerateProof(proverID)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult(
			"已生成零知识证明",
			map[string]interface{}{
				"prover_id": proverID,
				"t":         proof.THex,
				"c":         proof.CHex,
				"z":         proof.ZHex,
			},
			&types.ActionFeedback{
				Summary:     "证明者已生成一份可验证但不会泄露秘密的 Schnorr 证明。",
				NextHint:    "接下来可以验证该证明，观察验证方如何在不知道秘密的情况下完成确认。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"prover_id": proverID, "proof_count": len(s.proofs)},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported zkp action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// ZKPFactory 零知识证明演示器工厂
type ZKPFactory struct{}

// Create 创建演示器实例
func (f *ZKPFactory) Create() engine.Simulator {
	return NewZKPSimulator()
}

// GetDescription 获取描述
func (f *ZKPFactory) GetDescription() types.Description {
	return NewZKPSimulator().GetDescription()
}

// NewZKPFactory 创建工厂实例
func NewZKPFactory() *ZKPFactory {
	return &ZKPFactory{}
}

var _ engine.SimulatorFactory = (*ZKPFactory)(nil)
