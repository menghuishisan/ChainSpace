package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"golang.org/x/crypto/sha3"
)

// =============================================================================
// 数据结构定义
// =============================================================================

// SignatureAccount 签名账户
// 包含完整的密钥对信息和派生地址
type SignatureAccount struct {
	Name         string            `json:"name"`       // 账户名称
	Address      string            `json:"address"`    // 派生地址 (公钥哈希的前20字节)
	PublicKey    *ecdsa.PublicKey  `json:"-"`          // ECDSA公钥
	PrivateKey   *ecdsa.PrivateKey `json:"-"`          // ECDSA私钥
	PublicKeyHex string            `json:"public_key"` // 公钥十六进制
	CreatedAt    time.Time         `json:"created_at"` // 创建时间
}

// SignatureData 签名数据结构
// 符合以太坊签名格式 (r, s, v)
type SignatureData struct {
	R         *big.Int  `json:"-"`         // 签名R值
	S         *big.Int  `json:"-"`         // 签名S值
	V         uint8     `json:"v"`         // 恢复标识符
	RHex      string    `json:"r"`         // R值十六进制
	SHex      string    `json:"s"`         // S值十六进制
	Message   string    `json:"message"`   // 原始消息
	Hash      string    `json:"hash"`      // 消息哈希
	Signer    string    `json:"signer"`    // 签名者
	Timestamp time.Time `json:"timestamp"` // 签名时间
}

// SignatureRecord 签名记录
type SignatureRecord struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // sign/verify/recover
	Account   string    `json:"account"`
	Message   string    `json:"message"`
	Valid     bool      `json:"valid"`
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================
// SignatureSimulator 数字签名演示器
// =============================================================================

// SignatureSimulator 数字签名演示器
// 完整实现ECDSA签名算法，包括:
// - 密钥生成: 使用secp256k1/P256曲线
// - 签名生成: 标准ECDSA签名流程
// - 签名验证: 验证签名有效性
// - 地址恢复: 从签名恢复签名者公钥/地址
type SignatureSimulator struct {
	*base.BaseSimulator
	mu        sync.RWMutex
	accounts  map[string]*SignatureAccount // 账户映射
	curve     elliptic.Curve               // 椭圆曲线
	curveName string                       // 曲线名称
	history   []*SignatureRecord           // 操作历史
}

// NewSignatureSimulator 创建数字签名演示器
func NewSignatureSimulator() *SignatureSimulator {
	sim := &SignatureSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"signature",
			"数字签名演示器",
			"完整实现ECDSA数字签名，支持签名生成、验证和地址恢复",
			"crypto",
			types.ComponentTool,
		),
		accounts:  make(map[string]*SignatureAccount),
		curve:     elliptic.P256(),
		curveName: "P-256",
		history:   make([]*SignatureRecord, 0),
	}

	// 参数定义
	sim.AddParam(types.Param{
		Key:         "curve",
		Name:        "椭圆曲线",
		Description: "ECDSA使用的椭圆曲线",
		Type:        types.ParamTypeSelect,
		Default:     "P-256",
		Options: []types.Option{
			{Label: "P-256 (secp256r1)", Value: "P-256"},
			{Label: "P-384", Value: "P-384"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *SignatureSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 解析曲线参数
	if v, ok := config.Params["curve"]; ok {
		if curveName, ok := v.(string); ok {
			switch curveName {
			case "P-384":
				s.curve = elliptic.P384()
				s.curveName = "P-384"
			default:
				s.curve = elliptic.P256()
				s.curveName = "P-256"
			}
		}
	}

	// 创建示例账户
	s.CreateAccount("alice")
	s.CreateAccount("bob")
	s.CreateAccount("charlie")

	s.updateState()
	return nil
}

// =============================================================================
// 核心功能实现
// =============================================================================

// CreateAccount 创建新账户
// 生成ECDSA密钥对并计算派生地址
func (s *SignatureSimulator) CreateAccount(name string) (*SignatureAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查账户是否已存在
	if _, exists := s.accounts[name]; exists {
		return nil, fmt.Errorf("账户已存在: %s", name)
	}

	// 生成ECDSA密钥对
	// 私钥是一个随机大整数 d，满足 1 < d < n (n是曲线阶)
	// 公钥是椭圆曲线上的点 Q = d * G (G是基点)
	privateKey, err := ecdsa.GenerateKey(s.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("密钥生成失败: %v", err)
	}

	// 编码公钥 (非压缩格式: 0x04 || X || Y)
	pubBytes := elliptic.Marshal(s.curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	// 计算地址: 以太坊标准方式
	// 1. 去掉0x04前缀，只保留X和Y坐标 (64字节)
	// 2. 对公钥进行Keccak256哈希
	// 3. 取哈希的后20字节作为地址
	keccak := sha3.NewLegacyKeccak256()
	keccak.Write(pubBytes[1:]) // 去掉0x04前缀
	addrHash := keccak.Sum(nil)
	address := "0x" + hex.EncodeToString(addrHash[12:]) // 取后20字节

	account := &SignatureAccount{
		Name:         name,
		Address:      address,
		PublicKey:    &privateKey.PublicKey,
		PrivateKey:   privateKey,
		PublicKeyHex: hex.EncodeToString(pubBytes),
		CreatedAt:    time.Now(),
	}

	s.accounts[name] = account

	// 记录事件
	s.EmitEvent("account_created", "", "", map[string]interface{}{
		"name":       name,
		"address":    address,
		"curve":      s.curveName,
		"public_key": account.PublicKeyHex[:32] + "...",
	})

	s.updateState()
	return account, nil
}

// Sign 签名消息
// ECDSA签名算法步骤:
// 1. 计算消息哈希 e = H(m)
// 2. 生成随机数 k，计算点 R = k * G
// 3. 计算 r = R.x mod n
// 4. 计算 s = k^(-1) * (e + r * d) mod n
// 5. 签名为 (r, s)
func (s *SignatureSimulator) Sign(accountName, message string) (*SignatureData, error) {
	s.mu.RLock()
	account := s.accounts[accountName]
	s.mu.RUnlock()

	if account == nil {
		return nil, fmt.Errorf("账户不存在: %s", accountName)
	}

	// 步骤1: 计算消息哈希
	hash := sha256.Sum256([]byte(message))

	// 步骤2-4: ECDSA签名 (Go标准库实现)
	r, sigS, err := ecdsa.Sign(rand.Reader, account.PrivateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("签名失败: %v", err)
	}

	// 计算恢复标识符 v
	// v 用于从签名恢复公钥，值为 27 或 28 (以太坊格式)
	v := s.calculateRecoveryID(account.PublicKey, hash[:], r, sigS)

	sig := &SignatureData{
		R:         r,
		S:         sigS,
		V:         v,
		RHex:      hex.EncodeToString(r.Bytes()),
		SHex:      hex.EncodeToString(sigS.Bytes()),
		Message:   message,
		Hash:      hex.EncodeToString(hash[:]),
		Signer:    accountName,
		Timestamp: time.Now(),
	}

	// 记录历史
	s.mu.Lock()
	s.history = append(s.history, &SignatureRecord{
		ID:        fmt.Sprintf("sign-%d", len(s.history)+1),
		Type:      "sign",
		Account:   accountName,
		Message:   message,
		Valid:     true,
		Timestamp: time.Now(),
	})
	s.mu.Unlock()

	s.EmitEvent("message_signed", "", "", map[string]interface{}{
		"account": accountName,
		"hash":    sig.Hash[:16] + "...",
		"r":       sig.RHex[:16] + "...",
		"s":       sig.SHex[:16] + "...",
		"v":       v,
	})

	s.updateState()
	return sig, nil
}

// Verify 验证签名
// ECDSA验证算法步骤:
// 1. 计算消息哈希 e = H(m)
// 2. 计算 w = s^(-1) mod n
// 3. 计算 u1 = e * w mod n, u2 = r * w mod n
// 4. 计算点 R' = u1 * G + u2 * Q
// 5. 验证 r == R'.x mod n
func (s *SignatureSimulator) Verify(accountName, message string, sig *SignatureData) bool {
	s.mu.RLock()
	account := s.accounts[accountName]
	s.mu.RUnlock()

	if account == nil || sig == nil {
		return false
	}

	// 计算消息哈希
	hash := sha256.Sum256([]byte(message))

	// 解析签名值
	r := sig.R
	sigS := sig.S
	if r == nil || sigS == nil {
		r = new(big.Int)
		sigS = new(big.Int)
		r.SetString(sig.RHex, 16)
		sigS.SetString(sig.SHex, 16)
	}

	// ECDSA验证
	valid := ecdsa.Verify(account.PublicKey, hash[:], r, sigS)

	// 记录历史
	s.mu.Lock()
	s.history = append(s.history, &SignatureRecord{
		ID:        fmt.Sprintf("verify-%d", len(s.history)+1),
		Type:      "verify",
		Account:   accountName,
		Message:   message,
		Valid:     valid,
		Timestamp: time.Now(),
	})
	s.mu.Unlock()

	s.EmitEvent("signature_verified", "", "", map[string]interface{}{
		"account": accountName,
		"valid":   valid,
		"hash":    hex.EncodeToString(hash[:16]) + "...",
	})

	s.updateState()
	return valid
}

// VerifyWithPublicKey 使用公钥验证签名
func (s *SignatureSimulator) VerifyWithPublicKey(publicKeyHex, message string, sig *SignatureData) bool {
	// 解析公钥
	pubBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false
	}

	x, y := elliptic.Unmarshal(s.curve, pubBytes)
	if x == nil {
		return false
	}

	publicKey := &ecdsa.PublicKey{
		Curve: s.curve,
		X:     x,
		Y:     y,
	}

	// 计算消息哈希
	hash := sha256.Sum256([]byte(message))

	// 解析签名值
	r := new(big.Int)
	sigS := new(big.Int)
	r.SetString(sig.RHex, 16)
	sigS.SetString(sig.SHex, 16)

	return ecdsa.Verify(publicKey, hash[:], r, sigS)
}

// RecoverPublicKey 从签名恢复公钥
// 实现椭圆曲线公钥恢复算法 (ecrecover)
//
// 数学原理:
// 给定签名(r, s)和消息哈希e，恢复公钥Q的步骤:
// 1. 从r计算椭圆曲线上的点R (有两个可能的点，由v决定)
// 2. 计算 Q = r^(-1) * (s*R - e*G)
//
// 其中G是基点，验证: 如果用Q验证签名成功，则Q是正确的公钥
func (s *SignatureSimulator) RecoverPublicKey(message string, sig *SignatureData) (string, error) {
	// 计算消息哈希
	hash := sha256.Sum256([]byte(message))

	// 解析签名值
	r := new(big.Int)
	sigS := new(big.Int)
	r.SetString(sig.RHex, 16)
	sigS.SetString(sig.SHex, 16)

	// 获取曲线参数
	curveParams := s.curve.Params()
	N := curveParams.N // 曲线阶

	// 根据v值确定R点的Y坐标奇偶性
	// v = 27 表示Y是偶数，v = 28 表示Y是奇数
	isOddY := sig.V == 28

	// 步骤1: 从r恢复点R的坐标
	// R.x = r (mod N)
	Rx := new(big.Int).Set(r)

	// 计算R.y: 解方程 y^2 = x^3 + ax + b (mod p)
	// 对于P-256曲线: a = -3, b = curve.Params().B
	Ry, err := s.recoverY(Rx, isOddY)
	if err != nil {
		return "", fmt.Errorf("无法恢复R点Y坐标: %v", err)
	}

	// 验证R点在曲线上
	if !s.curve.IsOnCurve(Rx, Ry) {
		return "", errors.New("恢复的R点不在曲线上")
	}

	// 步骤2: 计算公钥 Q = r^(-1) * (s*R - e*G)
	// 首先计算 r^(-1) mod N
	rInv := new(big.Int).ModInverse(r, N)
	if rInv == nil {
		return "", errors.New("r没有模逆元")
	}

	// 计算 e (消息哈希作为大整数)
	e := new(big.Int).SetBytes(hash[:])

	// 计算 s*R
	sRx, sRy := s.curve.ScalarMult(Rx, Ry, sigS.Bytes())

	// 计算 e*G
	eGx, eGy := s.curve.ScalarBaseMult(e.Bytes())

	// 计算 -e*G (取Y坐标的负值)
	negEGy := new(big.Int).Sub(curveParams.P, eGy)

	// 计算 s*R - e*G = s*R + (-e*G)
	diffX, diffY := s.curve.Add(sRx, sRy, eGx, negEGy)

	// 计算 Q = r^(-1) * (s*R - e*G)
	Qx, Qy := s.curve.ScalarMult(diffX, diffY, rInv.Bytes())

	// 验证恢复的公钥
	recoveredPubKey := &ecdsa.PublicKey{
		Curve: s.curve,
		X:     Qx,
		Y:     Qy,
	}

	// 验证签名
	if !ecdsa.Verify(recoveredPubKey, hash[:], r, sigS) {
		// 尝试另一个Y值
		Ry, _ = s.recoverY(Rx, !isOddY)
		if s.curve.IsOnCurve(Rx, Ry) {
			sRx, sRy = s.curve.ScalarMult(Rx, Ry, sigS.Bytes())
			diffX, diffY = s.curve.Add(sRx, sRy, eGx, negEGy)
			Qx, Qy = s.curve.ScalarMult(diffX, diffY, rInv.Bytes())
			recoveredPubKey = &ecdsa.PublicKey{Curve: s.curve, X: Qx, Y: Qy}
			if !ecdsa.Verify(recoveredPubKey, hash[:], r, sigS) {
				return "", errors.New("公钥恢复验证失败")
			}
		}
	}

	// 编码公钥
	pubBytes := elliptic.Marshal(s.curve, Qx, Qy)
	return hex.EncodeToString(pubBytes), nil
}

// recoverY 从X坐标恢复Y坐标
// 解方程: y^2 = x^3 + ax + b (mod p)
// 对于P-256: a = -3
func (s *SignatureSimulator) recoverY(x *big.Int, isOdd bool) (*big.Int, error) {
	curveParams := s.curve.Params()
	P := curveParams.P
	B := curveParams.B

	// 计算 x^3
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Mod(x3, P)

	// 计算 ax (对于P-256, a = -3)
	threeX := new(big.Int).Mul(x, big.NewInt(3))
	threeX.Mod(threeX, P)

	// 计算 x^3 - 3x + b = x^3 + ax + b
	y2 := new(big.Int).Sub(x3, threeX)
	y2.Add(y2, B)
	y2.Mod(y2, P)

	// 计算平方根: y = y2^((p+1)/4) mod p (仅当p ≡ 3 mod 4时有效)
	// P-256的p满足此条件
	exp := new(big.Int).Add(P, big.NewInt(1))
	exp.Div(exp, big.NewInt(4))
	y := new(big.Int).Exp(y2, exp, P)

	// 验证 y^2 == y2
	ySquared := new(big.Int).Mul(y, y)
	ySquared.Mod(ySquared, P)
	if ySquared.Cmp(y2) != 0 {
		return nil, errors.New("X坐标无效，不存在对应的Y坐标")
	}

	// 根据奇偶性选择正确的Y
	if isOdd != (y.Bit(0) == 1) {
		y.Sub(P, y)
	}

	return y, nil
}

// RecoverAddress 从签名恢复地址
// 使用Keccak256计算以太坊标准地址
func (s *SignatureSimulator) RecoverAddress(message string, sig *SignatureData) (string, error) {
	pubKeyHex, err := s.RecoverPublicKey(message, sig)
	if err != nil {
		return "", err
	}

	pubBytes, _ := hex.DecodeString(pubKeyHex)
	// 使用Keccak256计算地址
	keccak := sha3.NewLegacyKeccak256()
	keccak.Write(pubBytes[1:]) // 去掉0x04前缀
	addrHash := keccak.Sum(nil)
	address := "0x" + hex.EncodeToString(addrHash[12:]) // 取后20字节

	s.EmitEvent("address_recovered", "", "", map[string]interface{}{
		"address": address,
	})

	return address, nil
}

// calculateRecoveryID 计算恢复标识符
//
// 恢复标识符v用于从签名恢复公钥时确定正确的Y坐标
// 在ECDSA签名中，给定r值可能对应曲线上的两个点(Y和-Y)
// v的值指示应该使用哪一个
//
// 以太坊标准:
// - v = 27: R点的Y坐标是偶数
// - v = 28: R点的Y坐标是奇数
//
// 计算方法: 尝试两个可能的Y值，找到能恢复出正确公钥的那个
func (s *SignatureSimulator) calculateRecoveryID(pubKey *ecdsa.PublicKey, hash []byte, r, sigS *big.Int) uint8 {
	curveParams := s.curve.Params()
	N := curveParams.N

	// 尝试v=27 (Y偶数)
	for v := uint8(27); v <= 28; v++ {
		isOddY := v == 28

		// 从r恢复R点
		Rx := new(big.Int).Set(r)
		Ry, err := s.recoverY(Rx, isOddY)
		if err != nil {
			continue
		}

		if !s.curve.IsOnCurve(Rx, Ry) {
			continue
		}

		// 计算恢复的公钥
		rInv := new(big.Int).ModInverse(r, N)
		if rInv == nil {
			continue
		}

		e := new(big.Int).SetBytes(hash)
		sRx, sRy := s.curve.ScalarMult(Rx, Ry, sigS.Bytes())
		eGx, eGy := s.curve.ScalarBaseMult(e.Bytes())
		negEGy := new(big.Int).Sub(curveParams.P, eGy)
		diffX, diffY := s.curve.Add(sRx, sRy, eGx, negEGy)
		Qx, Qy := s.curve.ScalarMult(diffX, diffY, rInv.Bytes())

		// 检查恢复的公钥是否与原公钥匹配
		if Qx.Cmp(pubKey.X) == 0 && Qy.Cmp(pubKey.Y) == 0 {
			return v
		}
	}

	// 默认返回27
	return 27
}

// GetAccount 获取账户信息
func (s *SignatureSimulator) GetAccount(name string) *SignatureAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accounts[name]
}

// GetHistory 获取操作历史
func (s *SignatureSimulator) GetHistory() []*SignatureRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.history
}

// updateState 更新状态
func (s *SignatureSimulator) updateState() {
	accountList := make([]map[string]interface{}, 0)
	s.mu.RLock()
	for name, acc := range s.accounts {
		accountList = append(accountList, map[string]interface{}{
			"name":    name,
			"address": acc.Address,
		})
	}
	historyCount := len(s.history)
	s.mu.RUnlock()

	s.SetGlobalData("curve", s.curveName)
	s.SetGlobalData("accounts", accountList)
	s.SetGlobalData("account_count", len(accountList))
	s.SetGlobalData("history_count", historyCount)

	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		func() string {
			if historyCount > 0 {
				return "signature_completed"
			}
			return "signature_ready"
		}(),
		func() string {
			if historyCount > 0 {
				return fmt.Sprintf("当前已经记录了 %d 次签名相关操作。", historyCount)
			}
			return "当前还没有执行签名实验，可以先创建账户并签署一条消息。"
		}(),
		"重点观察消息哈希、签名三元组以及验证和地址恢复之间的关系。",
		func() float64 {
			if historyCount > 0 {
				return 0.85
			}
			return 0.2
		}(),
		map[string]interface{}{"account_count": len(accountList), "history_count": historyCount, "curve": s.curveName},
	)
}

// ExecuteAction 为数字签名实验提供交互动作。
func (s *SignatureSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_account":
		name := "demo"
		if raw, ok := params["name"].(string); ok && raw != "" {
			name = raw
		}
		account, err := s.CreateAccount(name)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已创建一个签名账户。", map[string]interface{}{"name": account.Name, "address": account.Address}, &types.ActionFeedback{
			Summary:     "新的签名账户和密钥对已经生成。",
			NextHint:    "继续对这个账户发起签名和验证，观察哈希、签名和地址恢复之间的关系。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"address": account.Address},
		}), nil
	case "sign_message":
		account := "alice"
		message := "ChainSpace"
		if raw, ok := params["account"].(string); ok && raw != "" {
			account = raw
		}
		if raw, ok := params["message"].(string); ok && raw != "" {
			message = raw
		}
		sig, err := s.Sign(account, message)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已完成一次消息签名。", map[string]interface{}{"signature": sig}, &types.ActionFeedback{
			Summary:     "消息哈希和签名结果已经生成。",
			NextHint:    "继续执行验证或地址恢复，确认签名确实来自对应账户。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"signer": account, "hash": sig.Hash},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported signature action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// SignatureFactory 数字签名演示器工厂
type SignatureFactory struct{}

// Create 创建演示器实例
func (f *SignatureFactory) Create() engine.Simulator {
	return NewSignatureSimulator()
}

// GetDescription 获取描述
func (f *SignatureFactory) GetDescription() types.Description {
	return NewSignatureSimulator().GetDescription()
}

// NewSignatureFactory 创建工厂实例
func NewSignatureFactory() *SignatureFactory {
	return &SignatureFactory{}
}

var _ engine.SimulatorFactory = (*SignatureFactory)(nil)
