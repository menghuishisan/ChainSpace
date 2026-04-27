package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"golang.org/x/crypto/pbkdf2"
)

// DerivedKey 派生密钥
type DerivedKey struct {
	ID         string    `json:"id"`          // 密钥ID
	Algorithm  string    `json:"algorithm"`   // 算法
	InputType  string    `json:"input_type"`  // 输入类型
	Salt       string    `json:"salt"`        // 盐值
	Iterations int       `json:"iterations"`  // 迭代次数
	KeyLength  int       `json:"key_length"`  // 密钥长度
	DerivedKey string    `json:"derived_key"` // 派生密钥
	Timestamp  time.Time `json:"timestamp"`   // 时间戳
}

// KDFSimulator 密钥派生函数演示器
// 展示PBKDF2、HKDF等密钥派生函数
type KDFSimulator struct {
	*base.BaseSimulator
	history []*DerivedKey // 派生历史
}

// NewKDFSimulator 创建KDF演示器
func NewKDFSimulator() *KDFSimulator {
	sim := &KDFSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"kdf",
			"密钥派生函数演示器",
			"展示PBKDF2、HKDF等密钥派生函数的工作原理",
			"crypto",
			types.ComponentTool,
		),
		history: make([]*DerivedKey, 0),
	}

	sim.AddParam(types.Param{
		Key:         "algorithm",
		Name:        "算法",
		Description: "密钥派生算法",
		Type:        types.ParamTypeSelect,
		Default:     "PBKDF2-SHA256",
		Options: []types.Option{
			{Label: "PBKDF2-SHA256", Value: "PBKDF2-SHA256"},
			{Label: "PBKDF2-SHA512", Value: "PBKDF2-SHA512"},
			{Label: "HKDF-SHA256", Value: "HKDF-SHA256"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *KDFSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// PBKDF2 使用PBKDF2派生密钥
// password: 密码
// salt: 盐值(可选，为空则自动生成)
// iterations: 迭代次数
// keyLen: 输出密钥长度(字节)
// hashFunc: 哈希函数 (sha256/sha512)
func (s *KDFSimulator) PBKDF2(password, salt string, iterations, keyLen int, hashFunc string) (*DerivedKey, error) {
	// 验证参数
	if iterations < 1000 {
		iterations = 10000 // 最小安全迭代次数
	}
	if keyLen < 16 {
		keyLen = 32
	}

	// 生成盐值
	var saltBytes []byte
	if salt == "" {
		saltBytes = make([]byte, 16)
		rand.Read(saltBytes)
		salt = hex.EncodeToString(saltBytes)
	} else {
		saltBytes, _ = hex.DecodeString(salt)
		if len(saltBytes) == 0 {
			saltBytes = []byte(salt)
		}
	}

	// 选择哈希函数
	var h func() hash.Hash
	algorithm := "PBKDF2-SHA256"
	switch hashFunc {
	case "sha512":
		h = sha512.New
		algorithm = "PBKDF2-SHA512"
	default:
		h = sha256.New
	}

	// 派生密钥
	derivedBytes := pbkdf2.Key([]byte(password), saltBytes, iterations, keyLen, h)

	result := &DerivedKey{
		ID:         fmt.Sprintf("key-%d", len(s.history)+1),
		Algorithm:  algorithm,
		InputType:  "password",
		Salt:       hex.EncodeToString(saltBytes),
		Iterations: iterations,
		KeyLength:  keyLen,
		DerivedKey: hex.EncodeToString(derivedBytes),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)

	s.EmitEvent("key_derived", "", "", map[string]interface{}{
		"algorithm":  algorithm,
		"iterations": iterations,
		"key_length": keyLen,
		"key":        result.DerivedKey[:16] + "...",
	})

	s.updateState()
	return result, nil
}

// HKDF 使用HKDF派生密钥
// ikm: 输入密钥材料
// salt: 盐值
// info: 上下文信息
// keyLen: 输出密钥长度
func (s *KDFSimulator) HKDF(ikm, salt, info string, keyLen int) (*DerivedKey, error) {
	if keyLen < 16 {
		keyLen = 32
	}

	// Extract阶段
	var saltBytes []byte
	if salt == "" {
		saltBytes = make([]byte, sha256.Size)
	} else {
		saltBytes = []byte(salt)
	}

	h := hmac.New(sha256.New, saltBytes)
	h.Write([]byte(ikm))
	prk := h.Sum(nil) // 伪随机密钥

	// Expand阶段
	derived := s.hkdfExpand(prk, []byte(info), keyLen)

	result := &DerivedKey{
		ID:         fmt.Sprintf("key-%d", len(s.history)+1),
		Algorithm:  "HKDF-SHA256",
		InputType:  "ikm",
		Salt:       hex.EncodeToString(saltBytes),
		Iterations: 1,
		KeyLength:  keyLen,
		DerivedKey: hex.EncodeToString(derived),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)

	s.EmitEvent("key_derived", "", "", map[string]interface{}{
		"algorithm":  "HKDF-SHA256",
		"key_length": keyLen,
		"key":        result.DerivedKey[:16] + "...",
	})

	s.updateState()
	return result, nil
}

// hkdfExpand HKDF Expand阶段
func (s *KDFSimulator) hkdfExpand(prk, info []byte, length int) []byte {
	hashLen := sha256.Size
	n := (length + hashLen - 1) / hashLen

	okm := make([]byte, 0, length)
	t := []byte{}

	for i := 1; i <= n; i++ {
		h := hmac.New(sha256.New, prk)
		h.Write(t)
		h.Write(info)
		h.Write([]byte{byte(i)})
		t = h.Sum(nil)
		okm = append(okm, t...)
	}

	return okm[:length]
}

// DeriveFromMnemonic 从助记词派生密钥
func (s *KDFSimulator) DeriveFromMnemonic(mnemonic, passphrase string) (*DerivedKey, error) {
	// BIP39标准: 使用PBKDF2-SHA512，2048次迭代
	salt := "mnemonic" + passphrase
	derivedBytes := pbkdf2.Key([]byte(mnemonic), []byte(salt), 2048, 64, sha512.New)

	result := &DerivedKey{
		ID:         fmt.Sprintf("seed-%d", len(s.history)+1),
		Algorithm:  "BIP39-PBKDF2",
		InputType:  "mnemonic",
		Salt:       "mnemonic" + passphrase[:min(8, len(passphrase))] + "...",
		Iterations: 2048,
		KeyLength:  64,
		DerivedKey: hex.EncodeToString(derivedBytes),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)

	s.EmitEvent("seed_derived", "", "", map[string]interface{}{
		"algorithm": "BIP39",
		"seed":      result.DerivedKey[:16] + "...",
	})

	s.updateState()
	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// updateState 更新状态
func (s *KDFSimulator) updateState() {
	s.SetGlobalData("history_count", len(s.history))

	recentHistory := s.history
	if len(recentHistory) > 10 {
		recentHistory = recentHistory[len(recentHistory)-10:]
	}
	s.SetGlobalData("recent_history", recentHistory)

	summary := fmt.Sprintf("当前已派生 %d 组密钥。", len(s.history))
	nextHint := "可以继续对比 PBKDF2、HKDF 和助记词派生的差异。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备派生",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"history_count": len(s.history)},
	)
}

func (s *KDFSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "derive_pbkdf2":
		password := "chainspace"
		salt := "salt"
		if raw, ok := params["password"].(string); ok && raw != "" {
			password = raw
		}
		if raw, ok := params["salt"].(string); ok && raw != "" {
			salt = raw
		}
		result, err := s.PBKDF2(password, salt, 4096, 32, "sha256")
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已完成一次 PBKDF2 派生。", map[string]interface{}{"derived_key": result.DerivedKey}, &types.ActionFeedback{
			Summary:     "口令已经通过 PBKDF2 派生出新的密钥。",
			NextHint:    "继续比较 HKDF 或助记词派生，观察不同输入和迭代参数的差异。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"algorithm": result.Algorithm, "iterations": result.Iterations},
		}), nil
	case "derive_hkdf":
		ikm := "chainspace"
		if raw, ok := params["ikm"].(string); ok && raw != "" {
			ikm = raw
		}
		result, err := s.HKDF(ikm, "", "demo", 32)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已完成一次 HKDF 派生。", map[string]interface{}{"derived_key": result.DerivedKey}, &types.ActionFeedback{
			Summary:     "输入密钥材料已经通过 HKDF 派生出新的密钥。",
			NextHint:    "观察盐值、信息字段和长度变化对结果的影响。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"algorithm": result.Algorithm, "key_length": result.KeyLength},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported kdf action: %s", action)
	}
}

// KDFFactory KDF演示器工厂
type KDFFactory struct{}

func (f *KDFFactory) Create() engine.Simulator {
	return NewKDFSimulator()
}

func (f *KDFFactory) GetDescription() types.Description {
	return NewKDFSimulator().GetDescription()
}

func NewKDFFactory() *KDFFactory {
	return &KDFFactory{}
}

var _ engine.SimulatorFactory = (*KDFFactory)(nil)
