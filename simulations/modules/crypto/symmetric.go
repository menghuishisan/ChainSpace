package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// EncryptionRecord 加密记录
type EncryptionRecord struct {
	ID        string    `json:"id"`         // 记录ID
	Operation string    `json:"operation"`  // 操作类型: encrypt/decrypt
	Algorithm string    `json:"algorithm"`  // 算法名称
	Mode      string    `json:"mode"`       // 加密模式: GCM/CBC/CTR
	InputLen  int       `json:"input_len"`  // 输入长度
	OutputLen int       `json:"output_len"` // 输出长度
	KeySize   int       `json:"key_size"`   // 密钥长度(位)
	Timestamp time.Time `json:"timestamp"`  // 操作时间
}

// SymmetricSimulator 对称加密演示器
// 支持AES算法的多种工作模式，展示对称加密的工作原理
type SymmetricSimulator struct {
	*base.BaseSimulator
	key     []byte              // 当前密钥
	mode    string              // 当前模式
	history []*EncryptionRecord // 操作历史
	keySize int                 // 密钥大小(字节)
}

// NewSymmetricSimulator 创建对称加密演示器
func NewSymmetricSimulator() *SymmetricSimulator {
	sim := &SymmetricSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"symmetric",
			"对称加密演示器",
			"展示AES对称加密算法，支持GCM/CBC/CTR等模式",
			"crypto",
			types.ComponentTool,
		),
		mode:    "GCM",
		history: make([]*EncryptionRecord, 0),
		keySize: 32, // AES-256
	}

	// 添加参数定义
	sim.AddParam(types.Param{
		Key:         "key_size",
		Name:        "密钥长度",
		Description: "AES密钥的位数",
		Type:        types.ParamTypeSelect,
		Default:     "256",
		Options: []types.Option{
			{Label: "AES-128 (128位)", Value: "128"},
			{Label: "AES-192 (192位)", Value: "192"},
			{Label: "AES-256 (256位)", Value: "256"},
		},
	})
	sim.AddParam(types.Param{
		Key:         "mode",
		Name:        "加密模式",
		Description: "分组密码工作模式",
		Type:        types.ParamTypeSelect,
		Default:     "GCM",
		Options: []types.Option{
			{Label: "GCM (认证加密)", Value: "GCM"},
			{Label: "CBC (密码块链接)", Value: "CBC"},
			{Label: "CTR (计数器模式)", Value: "CTR"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *SymmetricSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 解析密钥长度参数
	if v, ok := config.Params["key_size"]; ok {
		switch v {
		case "128":
			s.keySize = 16
		case "192":
			s.keySize = 24
		case "256":
			s.keySize = 32
		}
	}

	// 解析模式参数
	if v, ok := config.Params["mode"]; ok {
		if mode, ok := v.(string); ok {
			s.mode = mode
		}
	}

	// 生成初始密钥
	s.GenerateKey(s.keySize)
	s.updateState()
	return nil
}

// GenerateKey 生成新密钥
// size: 密钥字节数 (16=AES-128, 24=AES-192, 32=AES-256)
func (s *SymmetricSimulator) GenerateKey(size int) string {
	// 验证密钥长度
	if size != 16 && size != 24 && size != 32 {
		size = 32 // 默认AES-256
	}

	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return ""
	}

	s.key = key
	s.keySize = size
	keyHex := hex.EncodeToString(key)

	s.EmitEvent("key_generated", "", "", map[string]interface{}{
		"key_size_bits": size * 8,
		"key_preview":   keyHex[:16] + "...",
	})

	s.updateState()
	return keyHex
}

// SetKey 设置密钥
func (s *SymmetricSimulator) SetKey(keyHex string) error {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("无效的十六进制密钥: %v", err)
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return fmt.Errorf("无效的密钥长度: %d字节，必须为16/24/32字节", len(key))
	}

	s.key = key
	s.keySize = len(key)
	s.updateState()
	return nil
}

// Encrypt 加密数据
// plaintext: 明文数据
// 返回: 十六进制编码的密文
func (s *SymmetricSimulator) Encrypt(plaintext string) (string, error) {
	if len(s.key) == 0 {
		return "", fmt.Errorf("密钥未初始化")
	}

	// 创建AES密码块
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("创建AES密码块失败: %v", err)
	}

	var ciphertext []byte

	switch s.mode {
	case "GCM":
		// GCM模式 - 提供认证加密
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", fmt.Errorf("创建GCM失败: %v", err)
		}

		// 生成随机Nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return "", fmt.Errorf("生成Nonce失败: %v", err)
		}

		// 加密，Nonce前置
		ciphertext = gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	case "CTR":
		// CTR模式
		ciphertext = make([]byte, aes.BlockSize+len(plaintext))
		iv := ciphertext[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return "", fmt.Errorf("生成IV失败: %v", err)
		}

		stream := cipher.NewCTR(block, iv)
		stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	default:
		// 默认使用GCM
		gcm, _ := cipher.NewGCM(block)
		nonce := make([]byte, gcm.NonceSize())
		io.ReadFull(rand.Reader, nonce)
		ciphertext = gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	}

	result := hex.EncodeToString(ciphertext)

	// 记录操作
	s.history = append(s.history, &EncryptionRecord{
		ID:        fmt.Sprintf("enc-%d", len(s.history)+1),
		Operation: "encrypt",
		Algorithm: "AES",
		Mode:      s.mode,
		InputLen:  len(plaintext),
		OutputLen: len(result),
		KeySize:   s.keySize * 8,
		Timestamp: time.Now(),
	})

	s.EmitEvent("encrypted", "", "", map[string]interface{}{
		"mode":           s.mode,
		"plaintext_len":  len(plaintext),
		"ciphertext_len": len(result),
		"key_size_bits":  s.keySize * 8,
	})

	s.updateState()
	return result, nil
}

// Decrypt 解密数据
// ciphertextHex: 十六进制编码的密文
// 返回: 明文数据
func (s *SymmetricSimulator) Decrypt(ciphertextHex string) (string, error) {
	if len(s.key) == 0 {
		return "", fmt.Errorf("密钥未初始化")
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("无效的十六进制密文: %v", err)
	}

	// 创建AES密码块
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("创建AES密码块失败: %v", err)
	}

	var plaintext []byte

	switch s.mode {
	case "GCM":
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", fmt.Errorf("创建GCM失败: %v", err)
		}

		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return "", fmt.Errorf("密文太短")
		}

		nonce, cipherData := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err = gcm.Open(nil, nonce, cipherData, nil)
		if err != nil {
			return "", fmt.Errorf("解密失败(可能密钥错误或数据被篡改): %v", err)
		}

	case "CTR":
		if len(ciphertext) < aes.BlockSize {
			return "", fmt.Errorf("密文太短")
		}

		iv := ciphertext[:aes.BlockSize]
		cipherData := ciphertext[aes.BlockSize:]

		plaintext = make([]byte, len(cipherData))
		stream := cipher.NewCTR(block, iv)
		stream.XORKeyStream(plaintext, cipherData)

	default:
		gcm, _ := cipher.NewGCM(block)
		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return "", fmt.Errorf("密文太短")
		}
		nonce, cipherData := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err = gcm.Open(nil, nonce, cipherData, nil)
		if err != nil {
			return "", err
		}
	}

	// 记录操作
	s.history = append(s.history, &EncryptionRecord{
		ID:        fmt.Sprintf("dec-%d", len(s.history)+1),
		Operation: "decrypt",
		Algorithm: "AES",
		Mode:      s.mode,
		InputLen:  len(ciphertextHex),
		OutputLen: len(plaintext),
		KeySize:   s.keySize * 8,
		Timestamp: time.Now(),
	})

	s.EmitEvent("decrypted", "", "", map[string]interface{}{
		"mode":          s.mode,
		"plaintext_len": len(plaintext),
	})

	s.updateState()
	return string(plaintext), nil
}

// SetMode 设置加密模式
func (s *SymmetricSimulator) SetMode(mode string) {
	if mode == "GCM" || mode == "CBC" || mode == "CTR" {
		s.mode = mode
		s.updateState()
	}
}

// GetHistory 获取操作历史
func (s *SymmetricSimulator) GetHistory() []*EncryptionRecord {
	return s.history
}

// updateState 更新状态
func (s *SymmetricSimulator) updateState() {
	keyPreview := ""
	if len(s.key) > 0 {
		keyPreview = hex.EncodeToString(s.key[:8]) + "..."
	}

	s.SetGlobalData("key_preview", keyPreview)
	s.SetGlobalData("key_size_bits", s.keySize*8)
	s.SetGlobalData("mode", s.mode)
	s.SetGlobalData("history_count", len(s.history))

	// 最近10条历史
	recentHistory := s.history
	if len(recentHistory) > 10 {
		recentHistory = recentHistory[len(recentHistory)-10:]
	}
	s.SetGlobalData("recent_history", recentHistory)

	summary := fmt.Sprintf("当前使用 AES-%d %s，已记录 %d 次加解密操作。", s.keySize*8, s.mode, len(s.history))
	nextHint := "先生成或设置密钥，再执行加密和解密，对比不同模式的行为差异。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备密钥",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"mode": s.mode, "history_count": len(s.history), "key_size": s.keySize * 8},
	)
}

func (s *SymmetricSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "generate_key":
		size := s.keySize
		if raw, ok := params["size"].(float64); ok && int(raw) > 0 {
			size = int(raw)
		}
		key := s.GenerateKey(size)
		return cryptoActionResult("已生成一把对称密钥。", map[string]interface{}{"key": key, "size": size * 8}, &types.ActionFeedback{
			Summary:     "新的 AES 密钥已经生成，可以继续执行加密和解密操作。",
			NextHint:    "输入一段明文并执行加密，观察不同模式下密文结构的差异。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"mode": s.mode, "key_size": size * 8},
		}), nil
	case "encrypt_text":
		plaintext := "ChainSpace"
		if raw, ok := params["plaintext"].(string); ok && raw != "" {
			plaintext = raw
		}
		ciphertext, err := s.Encrypt(plaintext)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已完成一次对称加密。", map[string]interface{}{"ciphertext": ciphertext}, &types.ActionFeedback{
			Summary:     "明文已经通过当前模式加密为密文。",
			NextHint:    "继续执行解密，确认相同密钥和模式下可以恢复原始数据。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"mode": s.mode, "plaintext_len": len(plaintext)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported symmetric action: %s", action)
	}
}

// SymmetricFactory 对称加密演示器工厂
type SymmetricFactory struct{}

// Create 创建演示器实例
func (f *SymmetricFactory) Create() engine.Simulator {
	return NewSymmetricSimulator()
}

// GetDescription 获取描述
func (f *SymmetricFactory) GetDescription() types.Description {
	return NewSymmetricSimulator().GetDescription()
}

// NewSymmetricFactory 创建工厂实例
func NewSymmetricFactory() *SymmetricFactory {
	return &SymmetricFactory{}
}

var _ engine.SimulatorFactory = (*SymmetricFactory)(nil)
