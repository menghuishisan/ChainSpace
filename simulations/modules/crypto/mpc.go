package crypto

import (
	"crypto/rand"
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
// 安全多方计算数据结构
// =============================================================================

// MPCShare 秘密份额
// Shamir秘密共享中每个参与方持有的份额
type MPCShare struct {
	PartyID string   `json:"party_id"` // 参与方ID
	X       *big.Int `json:"-"`        // 份额的x坐标
	Y       *big.Int `json:"-"`        // 份额的y坐标 (实际份额值)
	XHex    string   `json:"x"`        // x坐标十六进制
	YHex    string   `json:"y"`        // y坐标十六进制
}

// MPCParty MPC参与方
type MPCParty struct {
	ID        string      `json:"id"`         // 参与方ID
	Index     int         `json:"index"`      // 参与方索引 (1-based)
	Shares    []*MPCShare `json:"shares"`     // 持有的份额
	CreatedAt time.Time   `json:"created_at"` // 创建时间
}

// SecretSharingSession 秘密共享会话
type SecretSharingSession struct {
	ID         string               `json:"id"`          // 会话ID
	Threshold  int                  `json:"threshold"`   // 门限值t
	TotalParts int                  `json:"total_parts"` // 总份额数n
	Shares     map[string]*MPCShare `json:"shares"`      // 分发的份额
	Prime      *big.Int             `json:"-"`           // 有限域素数
	PrimeHex   string               `json:"prime"`       // 素数十六进制
	Polynomial []*big.Int           `json:"-"`           // 多项式系数 (秘密是常数项)
	CreatedAt  time.Time            `json:"created_at"`  // 创建时间
}

// MPCRecord 操作记录
type MPCRecord struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // share/reconstruct/add/multiply
	SessionID string    `json:"session_id"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================
// MPCSimulator 安全多方计算演示器
// =============================================================================

// MPCSimulator 安全多方计算演示器
// 实现Shamir秘密共享方案，支持:
// - (t,n)门限秘密共享: 需要至少t个份额才能恢复秘密
// - 秘密加法: 份额可以直接相加
// - 拉格朗日插值恢复: 使用多项式插值恢复秘密
//
// 安全性保证:
// - 少于t个份额无法获得秘密的任何信息 (信息论安全)
// - 任意t个份额可以完全恢复秘密
type MPCSimulator struct {
	*base.BaseSimulator
	mu         sync.RWMutex
	parties    map[string]*MPCParty             // 参与方
	sessions   map[string]*SecretSharingSession // 秘密共享会话
	history    []*MPCRecord                     // 操作记录
	prime      *big.Int                         // 默认素数域
	threshold  int                              // 默认门限
	totalParts int                              // 默认参与方数
}

// NewMPCSimulator 创建安全多方计算演示器
func NewMPCSimulator() *MPCSimulator {
	sim := &MPCSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"mpc",
			"安全多方计算演示器",
			"实现Shamir秘密共享，支持(t,n)门限方案和秘密恢复",
			"crypto",
			types.ComponentTool,
		),
		parties:  make(map[string]*MPCParty),
		sessions: make(map[string]*SecretSharingSession),
		history:  make([]*MPCRecord, 0),
	}

	sim.AddParam(types.Param{
		Key:         "threshold",
		Name:        "门限值",
		Description: "恢复秘密所需的最少份额数",
		Type:        types.ParamTypeInt,
		Default:     3,
		Min:         2,
		Max:         10,
	})
	sim.AddParam(types.Param{
		Key:         "total_parts",
		Name:        "总份额数",
		Description: "秘密分割的总份额数",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         3,
		Max:         20,
	})

	return sim
}

// Init 初始化演示器
func (s *MPCSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 解析参数
	s.threshold = 3
	s.totalParts = 5

	if v, ok := config.Params["threshold"]; ok {
		if n, ok := v.(float64); ok {
			s.threshold = int(n)
		}
	}
	if v, ok := config.Params["total_parts"]; ok {
		if n, ok := v.(float64); ok {
			s.totalParts = int(n)
		}
	}

	// 确保threshold <= totalParts
	if s.threshold > s.totalParts {
		s.threshold = s.totalParts
	}

	// 生成大素数 (256位)
	s.prime, _ = rand.Prime(rand.Reader, 256)

	// 创建参与方
	for i := 1; i <= s.totalParts; i++ {
		id := fmt.Sprintf("party-%d", i)
		s.parties[id] = &MPCParty{
			ID:        id,
			Index:     i,
			Shares:    make([]*MPCShare, 0),
			CreatedAt: time.Now(),
		}
	}

	s.updateState()
	return nil
}

// =============================================================================
// Shamir秘密共享核心实现
// =============================================================================

// ShareSecret 分割秘密
//
// Shamir秘密共享原理:
//  1. 选择一个随机多项式 f(x) = a_0 + a_1*x + a_2*x^2 + ... + a_{t-1}*x^{t-1}
//     其中 a_0 = secret (秘密是常数项)
//  2. 为每个参与方i计算份额 share_i = f(i)
//  3. 任意t个份额可以通过拉格朗日插值恢复f(0) = secret
//
// 安全性: 少于t个份额无法确定多项式，因此无法获得秘密的任何信息
func (s *MPCSimulator) ShareSecret(secretHex string) (*SecretSharingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 解析秘密
	secretBytes, err := hex.DecodeString(secretHex)
	if err != nil {
		// 如果不是十六进制，当作字符串处理
		secretBytes = []byte(secretHex)
	}
	secret := new(big.Int).SetBytes(secretBytes)

	// 确保秘密在有限域内
	secret.Mod(secret, s.prime)

	// 步骤1: 生成随机多项式
	// f(x) = secret + a_1*x + a_2*x^2 + ... + a_{t-1}*x^{t-1}
	polynomial := make([]*big.Int, s.threshold)
	polynomial[0] = secret // 常数项是秘密

	for i := 1; i < s.threshold; i++ {
		// 生成随机系数
		coeff, _ := rand.Int(rand.Reader, s.prime)
		polynomial[i] = coeff
	}

	// 步骤2: 为每个参与方计算份额
	sessionID := fmt.Sprintf("session-%d", len(s.sessions)+1)
	session := &SecretSharingSession{
		ID:         sessionID,
		Threshold:  s.threshold,
		TotalParts: s.totalParts,
		Shares:     make(map[string]*MPCShare),
		Prime:      s.prime,
		PrimeHex:   hex.EncodeToString(s.prime.Bytes()),
		Polynomial: polynomial,
		CreatedAt:  time.Now(),
	}

	for id, party := range s.parties {
		x := big.NewInt(int64(party.Index))
		y := s.evaluatePolynomial(polynomial, x)

		share := &MPCShare{
			PartyID: id,
			X:       x,
			Y:       y,
			XHex:    hex.EncodeToString(x.Bytes()),
			YHex:    hex.EncodeToString(y.Bytes()),
		}

		session.Shares[id] = share
		party.Shares = append(party.Shares, share)
	}

	s.sessions[sessionID] = session

	// 记录
	s.history = append(s.history, &MPCRecord{
		ID:        fmt.Sprintf("share-%d", len(s.history)+1),
		Type:      "share",
		SessionID: sessionID,
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("secret_shared", "", "", map[string]interface{}{
		"session_id":  sessionID,
		"threshold":   s.threshold,
		"total_parts": s.totalParts,
	})

	s.updateState()
	return session, nil
}

// evaluatePolynomial 计算多项式在x点的值
// f(x) = a_0 + a_1*x + a_2*x^2 + ... + a_{t-1}*x^{t-1} mod p
func (s *MPCSimulator) evaluatePolynomial(coeffs []*big.Int, x *big.Int) *big.Int {
	result := new(big.Int).Set(coeffs[0])
	xPower := new(big.Int).Set(x)

	for i := 1; i < len(coeffs); i++ {
		term := new(big.Int).Mul(coeffs[i], xPower)
		result.Add(result, term)
		result.Mod(result, s.prime)
		xPower.Mul(xPower, x)
		xPower.Mod(xPower, s.prime)
	}

	return result
}

// ReconstructSecret 恢复秘密
//
// 使用拉格朗日插值公式:
// f(0) = Σ y_i * L_i(0)
// 其中 L_i(0) = Π_{j≠i} (0 - x_j) / (x_i - x_j)
//
// 只需t个份额即可恢复秘密
func (s *MPCSimulator) ReconstructSecret(sessionID string, partyIDs []string) (*big.Int, error) {
	s.mu.RLock()
	session := s.sessions[sessionID]
	s.mu.RUnlock()

	if session == nil {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}

	if len(partyIDs) < session.Threshold {
		return nil, fmt.Errorf("份额不足: 需要%d个，只有%d个", session.Threshold, len(partyIDs))
	}

	// 收集份额
	xs := make([]*big.Int, 0)
	ys := make([]*big.Int, 0)

	for _, partyID := range partyIDs {
		share := session.Shares[partyID]
		if share == nil {
			return nil, fmt.Errorf("参与方没有份额: %s", partyID)
		}
		xs = append(xs, share.X)
		ys = append(ys, share.Y)

		if len(xs) >= session.Threshold {
			break
		}
	}

	// 拉格朗日插值
	secret := s.lagrangeInterpolate(xs, ys, big.NewInt(0))

	// 记录
	s.mu.Lock()
	s.history = append(s.history, &MPCRecord{
		ID:        fmt.Sprintf("reconstruct-%d", len(s.history)+1),
		Type:      "reconstruct",
		SessionID: sessionID,
		Success:   true,
		Timestamp: time.Now(),
	})
	s.mu.Unlock()

	s.EmitEvent("secret_reconstructed", "", "", map[string]interface{}{
		"session_id":  sessionID,
		"shares_used": len(xs),
		"secret":      hex.EncodeToString(secret.Bytes())[:16] + "...",
	})

	return secret, nil
}

// lagrangeInterpolate 拉格朗日插值
// 计算多项式在目标点的值
func (s *MPCSimulator) lagrangeInterpolate(xs, ys []*big.Int, target *big.Int) *big.Int {
	result := big.NewInt(0)

	for i := 0; i < len(xs); i++ {
		// 计算拉格朗日基多项式 L_i(target)
		numerator := big.NewInt(1)
		denominator := big.NewInt(1)

		for j := 0; j < len(xs); j++ {
			if i != j {
				// numerator *= (target - x_j)
				diff := new(big.Int).Sub(target, xs[j])
				numerator.Mul(numerator, diff)
				numerator.Mod(numerator, s.prime)

				// denominator *= (x_i - x_j)
				diff = new(big.Int).Sub(xs[i], xs[j])
				denominator.Mul(denominator, diff)
				denominator.Mod(denominator, s.prime)
			}
		}

		// 处理负数
		if numerator.Sign() < 0 {
			numerator.Add(numerator, s.prime)
		}
		if denominator.Sign() < 0 {
			denominator.Add(denominator, s.prime)
		}

		// L_i(target) = numerator / denominator mod p
		// = numerator * denominator^(-1) mod p
		denominatorInv := new(big.Int).ModInverse(denominator, s.prime)
		if denominatorInv == nil {
			continue
		}

		li := new(big.Int).Mul(numerator, denominatorInv)
		li.Mod(li, s.prime)

		// result += y_i * L_i(target)
		term := new(big.Int).Mul(ys[i], li)
		result.Add(result, term)
		result.Mod(result, s.prime)
	}

	// 处理负数结果
	if result.Sign() < 0 {
		result.Add(result, s.prime)
	}

	return result
}

// AddShares 份额加法
//
// Shamir秘密共享的加法同态性:
// 如果 share_a = f(i) 对应秘密a，share_b = g(i) 对应秘密b
// 则 share_a + share_b = f(i) + g(i) 对应秘密 a + b
//
// 这允许在不恢复秘密的情况下对秘密进行加法运算
func (s *MPCSimulator) AddShares(session1ID, session2ID string) (*SecretSharingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session1 := s.sessions[session1ID]
	session2 := s.sessions[session2ID]

	if session1 == nil || session2 == nil {
		return nil, fmt.Errorf("会话不存在")
	}

	// 创建新会话存储加法结果
	sessionID := fmt.Sprintf("session-%d", len(s.sessions)+1)
	newSession := &SecretSharingSession{
		ID:         sessionID,
		Threshold:  session1.Threshold,
		TotalParts: session1.TotalParts,
		Shares:     make(map[string]*MPCShare),
		Prime:      s.prime,
		PrimeHex:   hex.EncodeToString(s.prime.Bytes()),
		CreatedAt:  time.Now(),
	}

	// 对应份额相加
	for partyID, share1 := range session1.Shares {
		share2 := session2.Shares[partyID]
		if share2 == nil {
			continue
		}

		// 新份额 = share1 + share2 mod p
		newY := new(big.Int).Add(share1.Y, share2.Y)
		newY.Mod(newY, s.prime)

		newShare := &MPCShare{
			PartyID: partyID,
			X:       share1.X,
			Y:       newY,
			XHex:    share1.XHex,
			YHex:    hex.EncodeToString(newY.Bytes()),
		}
		newSession.Shares[partyID] = newShare
	}

	s.sessions[sessionID] = newSession

	// 记录
	s.history = append(s.history, &MPCRecord{
		ID:        fmt.Sprintf("add-%d", len(s.history)+1),
		Type:      "add",
		SessionID: sessionID,
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("shares_added", "", "", map[string]interface{}{
		"session1":    session1ID,
		"session2":    session2ID,
		"new_session": sessionID,
	})

	s.updateState()
	return newSession, nil
}

// VerifyShare 验证份额
// 验证一个份额是否属于某个会话
func (s *MPCSimulator) VerifyShare(sessionID, partyID string, shareYHex string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[sessionID]
	if session == nil {
		return false
	}

	share := session.Shares[partyID]
	if share == nil {
		return false
	}

	return share.YHex == shareYHex
}

// GetSession 获取会话信息
func (s *MPCSimulator) GetSession(sessionID string) *SecretSharingSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionID]
}

// updateState 更新状态
func (s *MPCSimulator) updateState() {
	partyList := make([]map[string]interface{}, 0)
	for _, p := range s.parties {
		partyList = append(partyList, map[string]interface{}{
			"id":          p.ID,
			"index":       p.Index,
			"share_count": len(p.Shares),
		})
	}

	s.SetGlobalData("threshold", s.threshold)
	s.SetGlobalData("total_parts", s.totalParts)
	s.SetGlobalData("prime", s.prime.Text(16)[:32]+"...")
	s.SetGlobalData("parties", partyList)
	s.SetGlobalData("session_count", len(s.sessions))
	s.SetGlobalData("history_count", len(s.history))

	summary := fmt.Sprintf("当前共有 %d 个参与方，已创建 %d 个秘密共享会话。", len(s.parties), len(s.sessions))
	nextHint := "可以继续分发秘密、重构秘密或比较不同会话的份额运算。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备多方计算",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"party_count": len(s.parties), "session_count": len(s.sessions)},
	)
}

func (s *MPCSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "share_secret":
		secretHex := "2a"
		if raw, ok := params["secret"].(string); ok && raw != "" {
			secretHex = raw
		}
		session, err := s.ShareSecret(secretHex)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已创建一个秘密共享会话。", map[string]interface{}{"session_id": session.ID}, &types.ActionFeedback{
			Summary:     "秘密已经按照阈值策略分发给多个参与方。",
			NextHint:    "继续使用多个参与方重构秘密，观察阈值机制如何保证正确恢复。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"session_id": session.ID, "threshold": session.Threshold},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported mpc action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// MPCFactory 安全多方计算演示器工厂
type MPCFactory struct{}

// Create 创建演示器实例
func (f *MPCFactory) Create() engine.Simulator {
	return NewMPCSimulator()
}

// GetDescription 获取描述
func (f *MPCFactory) GetDescription() types.Description {
	return NewMPCSimulator().GetDescription()
}

// NewMPCFactory 创建工厂实例
func NewMPCFactory() *MPCFactory {
	return &MPCFactory{}
}

var _ engine.SimulatorFactory = (*MPCFactory)(nil)
