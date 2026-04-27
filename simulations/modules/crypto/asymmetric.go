package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// KeyPairInfo 密钥对信息
type KeyPairInfo struct {
	Curve         string    `json:"curve"`       // 椭圆曲线名称
	PrivateKeyHex string    `json:"private_key"` // 私钥(十六进制)
	PublicKeyHex  string    `json:"public_key"`  // 公钥(十六进制)
	AddressHex    string    `json:"address"`     // 派生地址
	CreatedAt     time.Time `json:"created_at"`  // 创建时间
}

// ECDSASignature ECDSA签名结构
type ECDSASignature struct {
	R       string `json:"r"`       // 签名R值
	S       string `json:"s"`       // 签名S值
	V       int    `json:"v"`       // 恢复ID
	Message string `json:"message"` // 原始消息
	Hash    string `json:"hash"`    // 消息哈希
}

// AsymmetricSimulator 非对称加密演示器
// 展示椭圆曲线加密(ECC)的密钥生成、签名和验证过程
type AsymmetricSimulator struct {
	*base.BaseSimulator
	privateKey *ecdsa.PrivateKey // 当前私钥
	publicKey  *ecdsa.PublicKey  // 当前公钥
	curve      elliptic.Curve    // 使用的椭圆曲线
	curveName  string            // 曲线名称
	keyInfo    *KeyPairInfo      // 密钥对信息
	signatures []*ECDSASignature // 签名历史
}

// NewAsymmetricSimulator 创建非对称加密演示器
func NewAsymmetricSimulator() *AsymmetricSimulator {
	sim := &AsymmetricSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"asymmetric",
			"非对称加密演示器",
			"展示ECDSA椭圆曲线加密算法，支持密钥生成、签名和验证",
			"crypto",
			types.ComponentTool,
		),
		curve:      elliptic.P256(),
		curveName:  "P-256",
		signatures: make([]*ECDSASignature, 0),
	}

	// 添加参数定义
	sim.AddParam(types.Param{
		Key:         "curve",
		Name:        "椭圆曲线",
		Description: "使用的椭圆曲线类型",
		Type:        types.ParamTypeSelect,
		Default:     "P-256",
		Options: []types.Option{
			{Label: "P-256 (secp256r1)", Value: "P-256"},
			{Label: "P-384 (secp384r1)", Value: "P-384"},
			{Label: "P-521 (secp521r1)", Value: "P-521"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *AsymmetricSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 解析曲线参数
	if v, ok := config.Params["curve"]; ok {
		if curveName, ok := v.(string); ok {
			s.setCurve(curveName)
		}
	}

	// 生成初始密钥对
	s.GenerateKeyPair()
	return nil
}

// setCurve 设置椭圆曲线
func (s *AsymmetricSimulator) setCurve(name string) {
	switch name {
	case "P-256":
		s.curve = elliptic.P256()
		s.curveName = "P-256"
	case "P-384":
		s.curve = elliptic.P384()
		s.curveName = "P-384"
	case "P-521":
		s.curve = elliptic.P521()
		s.curveName = "P-521"
	default:
		s.curve = elliptic.P256()
		s.curveName = "P-256"
	}
}

// GenerateKeyPair 生成新的密钥对
func (s *AsymmetricSimulator) GenerateKeyPair() (*KeyPairInfo, error) {
	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(s.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %v", err)
	}

	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey

	// 编码公钥和私钥
	privBytes := privateKey.D.Bytes()
	pubBytes := append(s.publicKey.X.Bytes(), s.publicKey.Y.Bytes()...)

	// 计算地址 (类似以太坊地址派生)
	addrHash := sha256.Sum256(pubBytes)
	address := hex.EncodeToString(addrHash[:20])

	// 创建密钥信息
	s.keyInfo = &KeyPairInfo{
		Curve:         s.curveName,
		PrivateKeyHex: hex.EncodeToString(privBytes),
		PublicKeyHex:  hex.EncodeToString(pubBytes),
		AddressHex:    "0x" + address,
		CreatedAt:     time.Now(),
	}

	// 发送事件
	s.EmitEvent("keypair_generated", "", "", map[string]interface{}{
		"curve":   s.curveName,
		"address": s.keyInfo.AddressHex,
		"pub_key": s.keyInfo.PublicKeyHex[:32] + "...",
	})

	s.updateState()
	return s.keyInfo, nil
}

// Sign 签名消息
// message: 要签名的消息
// 返回: 签名结构
func (s *AsymmetricSimulator) Sign(message string) (*ECDSASignature, error) {
	if s.privateKey == nil {
		return nil, fmt.Errorf("私钥未初始化")
	}

	// 计算消息哈希
	hash := sha256.Sum256([]byte(message))

	// ECDSA签名
	r, sigS, err := ecdsa.Sign(rand.Reader, s.privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("签名失败: %v", err)
	}

	// 创建签名结构
	sig := &ECDSASignature{
		R:       hex.EncodeToString(r.Bytes()),
		S:       hex.EncodeToString(sigS.Bytes()),
		V:       27, // 简化的恢复ID
		Message: message,
		Hash:    hex.EncodeToString(hash[:]),
	}

	s.signatures = append(s.signatures, sig)

	// 发送事件
	s.EmitEvent("message_signed", "", "", map[string]interface{}{
		"message_len": len(message),
		"hash":        sig.Hash[:16] + "...",
		"r":           sig.R[:16] + "...",
	})

	s.updateState()
	return sig, nil
}

// Verify 验证签名
// message: 原始消息
// rHex, sHex: 签名的R和S值(十六进制)
func (s *AsymmetricSimulator) Verify(message, rHex, sHex string) bool {
	if s.publicKey == nil {
		return false
	}

	// 计算消息哈希
	hash := sha256.Sum256([]byte(message))

	// 解析R和S
	r := new(big.Int)
	sigS := new(big.Int)
	r.SetString(rHex, 16)
	sigS.SetString(sHex, 16)

	// 验证签名
	valid := ecdsa.Verify(s.publicKey, hash[:], r, sigS)

	// 发送事件
	s.EmitEvent("signature_verified", "", "", map[string]interface{}{
		"valid":       valid,
		"message_len": len(message),
	})

	return valid
}

// VerifySignature 验证签名结构
func (s *AsymmetricSimulator) VerifySignature(sig *ECDSASignature) bool {
	return s.Verify(sig.Message, sig.R, sig.S)
}

// GetPublicKey 获取公钥
func (s *AsymmetricSimulator) GetPublicKey() string {
	if s.keyInfo != nil {
		return s.keyInfo.PublicKeyHex
	}
	return ""
}

// GetAddress 获取地址
func (s *AsymmetricSimulator) GetAddress() string {
	if s.keyInfo != nil {
		return s.keyInfo.AddressHex
	}
	return ""
}

// DeriveSharedSecret ECDH密钥协商
// otherPublicKeyHex: 对方公钥(十六进制)
func (s *AsymmetricSimulator) DeriveSharedSecret(otherPublicKeyHex string) (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("私钥未初始化")
	}

	// 解析对方公钥
	pubBytes, err := hex.DecodeString(otherPublicKeyHex)
	if err != nil {
		return "", fmt.Errorf("无效的公钥格式: %v", err)
	}

	// 分离X和Y坐标
	keyLen := len(pubBytes) / 2
	x := new(big.Int).SetBytes(pubBytes[:keyLen])
	y := new(big.Int).SetBytes(pubBytes[keyLen:])

	// ECDH计算共享密钥
	sharedX, _ := s.curve.ScalarMult(x, y, s.privateKey.D.Bytes())
	shared := sha256.Sum256(sharedX.Bytes())

	s.EmitEvent("shared_secret_derived", "", "", map[string]interface{}{
		"shared_preview": hex.EncodeToString(shared[:8]) + "...",
	})

	return hex.EncodeToString(shared[:]), nil
}

// updateState 更新状态
func (s *AsymmetricSimulator) updateState() {
	s.SetGlobalData("curve", s.curveName)
	s.SetGlobalData("signature_count", len(s.signatures))

	if s.keyInfo != nil {
		s.SetGlobalData("address", s.keyInfo.AddressHex)
		s.SetGlobalData("public_key_preview", s.keyInfo.PublicKeyHex[:32]+"...")
	}

	// 最近5条签名
	recentSigs := s.signatures
	if len(recentSigs) > 5 {
		recentSigs = recentSigs[len(recentSigs)-5:]
	}
	s.SetGlobalData("recent_signatures", recentSigs)

	summary := fmt.Sprintf("当前曲线为 %s，已记录 %d 次签名相关操作。", s.curveName, len(s.signatures))
	nextHint := "先生成密钥对，再进行签名和验签，观察公私钥与消息摘要的关系。"
	if s.keyInfo != nil {
		nextHint = "可以继续执行签名或共享密钥推导，观察公钥如何参与验证与协商。"
	}
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备密钥",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"curve": s.curveName, "signature_count": len(s.signatures)},
	)
}

func (s *AsymmetricSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "generate_keypair":
		keyInfo, err := s.GenerateKeyPair()
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已生成一组非对称密钥。", map[string]interface{}{"address": keyInfo.AddressHex, "curve": keyInfo.Curve}, &types.ActionFeedback{
			Summary:     "新的公私钥对已经生成，可以继续进行签名或密钥协商。",
			NextHint:    "尝试签名一条消息，观察消息哈希和签名结果如何对应。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"curve": keyInfo.Curve, "address": keyInfo.AddressHex},
		}), nil
	case "sign_message":
		message := "ChainSpace"
		if raw, ok := params["message"].(string); ok && raw != "" {
			message = raw
		}
		sig, err := s.Sign(message)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已完成一次非对称签名。", map[string]interface{}{"signature": sig}, &types.ActionFeedback{
			Summary:     "消息已被私钥签名，可以继续验证该签名是否与公钥匹配。",
			NextHint:    "继续执行验签或共享密钥推导，观察公钥在不同场景中的作用。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"hash": sig.Hash},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported asymmetric action: %s", action)
	}
}

// AsymmetricFactory 非对称加密演示器工厂
type AsymmetricFactory struct{}

// Create 创建演示器实例
func (f *AsymmetricFactory) Create() engine.Simulator {
	return NewAsymmetricSimulator()
}

// GetDescription 获取描述
func (f *AsymmetricFactory) GetDescription() types.Description {
	return NewAsymmetricSimulator().GetDescription()
}

// NewAsymmetricFactory 创建工厂实例
func NewAsymmetricFactory() *AsymmetricFactory {
	return &AsymmetricFactory{}
}

var _ engine.SimulatorFactory = (*AsymmetricFactory)(nil)
