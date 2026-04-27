package crypto

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 编码数据结构
// =============================================================================

// EncodedData 编码数据
type EncodedData struct {
	ID         string    `json:"id"`          // 数据ID
	Original   string    `json:"original"`    // 原始数据
	Encoding   string    `json:"encoding"`    // 编码类型
	Encoded    string    `json:"encoded"`     // 编码结果
	ByteLength int       `json:"byte_length"` // 原始字节长度
	EncodedLen int       `json:"encoded_len"` // 编码后长度
	Expansion  float64   `json:"expansion"`   // 膨胀率
	Timestamp  time.Time `json:"timestamp"`   // 时间戳
}

// EncodingRecord 操作记录
type EncodingRecord struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // encode/decode
	Encoding  string    `json:"encoding"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================
// EncodingSimulator 编码演示器
// =============================================================================

// EncodingSimulator 编码演示器
// 展示区块链中常用的编码方式:
//
// 1. Hex (十六进制):
//   - 每字节用2个十六进制字符表示
//   - 膨胀率: 2x
//   - 应用: 交易哈希、地址、签名等
//
// 2. Base64:
//   - 每3字节用4个字符表示
//   - 膨胀率: ~1.33x
//   - 应用: API传输、JWT、证书等
//
// 3. Base58:
//   - 类似Base64，但去除易混淆字符(0,O,I,l)
//   - 应用: 比特币地址、IPFS CID
//
// 4. Base32:
//   - 每5字节用8个字符表示
//   - 膨胀率: ~1.6x
//   - 应用: TOTP密钥、某些地址格式
//
// 5. Bech32:
//   - 比特币SegWit地址编码
//   - 包含错误检测码
//   - 应用: bc1开头的比特币地址
type EncodingSimulator struct {
	*base.BaseSimulator
	mu      sync.RWMutex
	history []*EncodedData    // 编码历史
	records []*EncodingRecord // 操作记录
}

// Base58字母表 (比特币使用)
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// NewEncodingSimulator 创建编码演示器
func NewEncodingSimulator() *EncodingSimulator {
	sim := &EncodingSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"encoding",
			"编码演示器",
			"展示Hex、Base64、Base58、Base32、Bech32等区块链常用编码",
			"crypto",
			types.ComponentTool,
		),
		history: make([]*EncodedData, 0),
		records: make([]*EncodingRecord, 0),
	}

	sim.AddParam(types.Param{
		Key:         "default_encoding",
		Name:        "默认编码",
		Description: "默认使用的编码类型",
		Type:        types.ParamTypeSelect,
		Default:     "hex",
		Options: []types.Option{
			{Label: "Hex (十六进制)", Value: "hex"},
			{Label: "Base64", Value: "base64"},
			{Label: "Base64 URL Safe", Value: "base64url"},
			{Label: "Base58", Value: "base58"},
			{Label: "Base32", Value: "base32"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *EncodingSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// =============================================================================
// 编码实现
// =============================================================================

// EncodeHex 十六进制编码
// 每个字节转换为2个十六进制字符
func (s *EncodingSimulator) EncodeHex(data string) (*EncodedData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes := []byte(data)
	encoded := hex.EncodeToString(bytes)

	result := &EncodedData{
		ID:         fmt.Sprintf("enc-%d", len(s.history)+1),
		Original:   data,
		Encoding:   "hex",
		Encoded:    encoded,
		ByteLength: len(bytes),
		EncodedLen: len(encoded),
		Expansion:  float64(len(encoded)) / float64(len(bytes)),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)
	s.records = append(s.records, &EncodingRecord{
		ID:        fmt.Sprintf("rec-%d", len(s.records)+1),
		Type:      "encode",
		Encoding:  "hex",
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("encoded", "", "", map[string]interface{}{
		"encoding":  "hex",
		"expansion": fmt.Sprintf("%.2fx", result.Expansion),
		"result":    truncateString(encoded, 32),
	})

	s.updateState()
	return result, nil
}

// DecodeHex 十六进制解码
func (s *EncodingSimulator) DecodeHex(encoded string) (string, error) {
	bytes, err := hex.DecodeString(encoded)
	if err != nil {
		s.records = append(s.records, &EncodingRecord{
			ID:        fmt.Sprintf("rec-%d", len(s.records)+1),
			Type:      "decode",
			Encoding:  "hex",
			Success:   false,
			Timestamp: time.Now(),
		})
		return "", fmt.Errorf("无效的十六进制字符串: %v", err)
	}

	s.records = append(s.records, &EncodingRecord{
		ID:        fmt.Sprintf("rec-%d", len(s.records)+1),
		Type:      "decode",
		Encoding:  "hex",
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("decoded", "", "", map[string]interface{}{
		"encoding": "hex",
		"result":   truncateString(string(bytes), 32),
	})

	return string(bytes), nil
}

// EncodeBase64 Base64编码
// 标准Base64，使用+和/
func (s *EncodingSimulator) EncodeBase64(data string) (*EncodedData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes := []byte(data)
	encoded := base64.StdEncoding.EncodeToString(bytes)

	result := &EncodedData{
		ID:         fmt.Sprintf("enc-%d", len(s.history)+1),
		Original:   data,
		Encoding:   "base64",
		Encoded:    encoded,
		ByteLength: len(bytes),
		EncodedLen: len(encoded),
		Expansion:  float64(len(encoded)) / float64(len(bytes)),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)
	s.records = append(s.records, &EncodingRecord{
		ID:        fmt.Sprintf("rec-%d", len(s.records)+1),
		Type:      "encode",
		Encoding:  "base64",
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("encoded", "", "", map[string]interface{}{
		"encoding":  "base64",
		"expansion": fmt.Sprintf("%.2fx", result.Expansion),
		"result":    truncateString(encoded, 32),
	})

	s.updateState()
	return result, nil
}

// DecodeBase64 Base64解码
func (s *EncodingSimulator) DecodeBase64(encoded string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("无效的Base64字符串: %v", err)
	}

	s.records = append(s.records, &EncodingRecord{
		ID:        fmt.Sprintf("rec-%d", len(s.records)+1),
		Type:      "decode",
		Encoding:  "base64",
		Success:   true,
		Timestamp: time.Now(),
	})

	return string(bytes), nil
}

// EncodeBase64URL URL安全的Base64编码
// 使用-和_替代+和/，无填充
func (s *EncodingSimulator) EncodeBase64URL(data string) (*EncodedData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes := []byte(data)
	encoded := base64.RawURLEncoding.EncodeToString(bytes)

	result := &EncodedData{
		ID:         fmt.Sprintf("enc-%d", len(s.history)+1),
		Original:   data,
		Encoding:   "base64url",
		Encoded:    encoded,
		ByteLength: len(bytes),
		EncodedLen: len(encoded),
		Expansion:  float64(len(encoded)) / float64(len(bytes)),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)

	s.EmitEvent("encoded", "", "", map[string]interface{}{
		"encoding":  "base64url",
		"expansion": fmt.Sprintf("%.2fx", result.Expansion),
	})

	s.updateState()
	return result, nil
}

// EncodeBase58 Base58编码
// 比特币地址使用的编码，去除了0,O,I,l等易混淆字符
func (s *EncodingSimulator) EncodeBase58(data string) (*EncodedData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes := []byte(data)
	encoded := s.base58Encode(bytes)

	result := &EncodedData{
		ID:         fmt.Sprintf("enc-%d", len(s.history)+1),
		Original:   data,
		Encoding:   "base58",
		Encoded:    encoded,
		ByteLength: len(bytes),
		EncodedLen: len(encoded),
		Expansion:  float64(len(encoded)) / float64(len(bytes)),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)
	s.records = append(s.records, &EncodingRecord{
		ID:        fmt.Sprintf("rec-%d", len(s.records)+1),
		Type:      "encode",
		Encoding:  "base58",
		Success:   true,
		Timestamp: time.Now(),
	})

	s.EmitEvent("encoded", "", "", map[string]interface{}{
		"encoding":  "base58",
		"expansion": fmt.Sprintf("%.2fx", result.Expansion),
		"result":    truncateString(encoded, 32),
	})

	s.updateState()
	return result, nil
}

// base58Encode Base58编码实现
func (s *EncodingSimulator) base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	// 计算前导零的数量
	zeros := 0
	for _, b := range input {
		if b == 0 {
			zeros++
		} else {
			break
		}
	}

	// 分配足够的空间 (log(256)/log(58) ≈ 1.37)
	size := len(input)*138/100 + 1
	output := make([]byte, size)

	// 大数除法
	for _, b := range input {
		carry := int(b)
		for i := size - 1; i >= 0; i-- {
			carry += 256 * int(output[i])
			output[i] = byte(carry % 58)
			carry /= 58
		}
	}

	// 跳过前导零
	i := 0
	for i < size && output[i] == 0 {
		i++
	}

	// 构建结果
	result := make([]byte, zeros+size-i)
	for j := 0; j < zeros; j++ {
		result[j] = '1' // Base58中'1'表示0
	}
	for j := zeros; i < size; i, j = i+1, j+1 {
		result[j] = base58Alphabet[output[i]]
	}

	return string(result)
}

// DecodeBase58 Base58解码
func (s *EncodingSimulator) DecodeBase58(encoded string) ([]byte, error) {
	if len(encoded) == 0 {
		return nil, nil
	}

	// 计算前导'1'的数量
	zeros := 0
	for _, c := range encoded {
		if c == '1' {
			zeros++
		} else {
			break
		}
	}

	// 分配空间
	size := len(encoded)*733/1000 + 1
	output := make([]byte, size)

	for _, c := range encoded {
		// 查找字符在字母表中的位置
		carry := strings.IndexByte(base58Alphabet, byte(c))
		if carry < 0 {
			return nil, fmt.Errorf("无效的Base58字符: %c", c)
		}

		for i := size - 1; i >= 0; i-- {
			carry += 58 * int(output[i])
			output[i] = byte(carry % 256)
			carry /= 256
		}
	}

	// 跳过前导零
	i := 0
	for i < size && output[i] == 0 {
		i++
	}

	// 构建结果
	result := make([]byte, zeros+size-i)
	copy(result[zeros:], output[i:])

	s.records = append(s.records, &EncodingRecord{
		ID:        fmt.Sprintf("rec-%d", len(s.records)+1),
		Type:      "decode",
		Encoding:  "base58",
		Success:   true,
		Timestamp: time.Now(),
	})

	return result, nil
}

// EncodeBase32 Base32编码
func (s *EncodingSimulator) EncodeBase32(data string) (*EncodedData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes := []byte(data)
	encoded := base32.StdEncoding.EncodeToString(bytes)

	result := &EncodedData{
		ID:         fmt.Sprintf("enc-%d", len(s.history)+1),
		Original:   data,
		Encoding:   "base32",
		Encoded:    encoded,
		ByteLength: len(bytes),
		EncodedLen: len(encoded),
		Expansion:  float64(len(encoded)) / float64(len(bytes)),
		Timestamp:  time.Now(),
	}

	s.history = append(s.history, result)

	s.EmitEvent("encoded", "", "", map[string]interface{}{
		"encoding":  "base32",
		"expansion": fmt.Sprintf("%.2fx", result.Expansion),
	})

	s.updateState()
	return result, nil
}

// CompareEncodings 比较不同编码方式
func (s *EncodingSimulator) CompareEncodings(data string) map[string]*EncodedData {
	results := make(map[string]*EncodedData)

	hexResult, _ := s.EncodeHex(data)
	results["hex"] = hexResult

	base64Result, _ := s.EncodeBase64(data)
	results["base64"] = base64Result

	base58Result, _ := s.EncodeBase58(data)
	results["base58"] = base58Result

	base32Result, _ := s.EncodeBase32(data)
	results["base32"] = base32Result

	s.EmitEvent("encoding_compared", "", "", map[string]interface{}{
		"original_length":  len(data),
		"hex_expansion":    fmt.Sprintf("%.2fx", hexResult.Expansion),
		"base64_expansion": fmt.Sprintf("%.2fx", base64Result.Expansion),
		"base58_expansion": fmt.Sprintf("%.2fx", base58Result.Expansion),
		"base32_expansion": fmt.Sprintf("%.2fx", base32Result.Expansion),
	})

	return results
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// updateState 更新状态
func (s *EncodingSimulator) updateState() {
	s.SetGlobalData("history_count", len(s.history))
	s.SetGlobalData("record_count", len(s.records))

	// 最近的编码历史
	recentHistory := s.history
	if len(recentHistory) > 5 {
		recentHistory = recentHistory[len(recentHistory)-5:]
	}

	historyList := make([]map[string]interface{}, 0)
	for _, h := range recentHistory {
		historyList = append(historyList, map[string]interface{}{
			"id":        h.ID,
			"encoding":  h.Encoding,
			"expansion": fmt.Sprintf("%.2fx", h.Expansion),
		})
	}
	s.SetGlobalData("recent_history", historyList)

	summary := fmt.Sprintf("当前已记录 %d 次编码操作和 %d 次解码结果。", len(s.history), len(s.records))
	nextHint := "可以继续比较 Hex、Base64、Base58、Base32 的长度膨胀差异。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备编码",
		summary,
		nextHint,
		0.4,
		map[string]interface{}{"history_count": len(s.history), "record_count": len(s.records)},
	)
}

func (s *EncodingSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "encode_base64":
		data := "ChainSpace"
		if raw, ok := params["data"].(string); ok && raw != "" {
			data = raw
		}
		result, err := s.EncodeBase64(data)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已完成一次 Base64 编码。", map[string]interface{}{"encoded": result.Encoded}, &types.ActionFeedback{
			Summary:     "输入数据已经转换为 Base64 表示。",
			NextHint:    "继续比较 Hex、Base58、Base32 的膨胀比例差异。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"encoding": result.Encoding, "expansion": result.Expansion},
		}), nil
	case "compare_encodings":
		data := "ChainSpace"
		if raw, ok := params["data"].(string); ok && raw != "" {
			data = raw
		}
		results := s.CompareEncodings(data)
		return cryptoActionResult("已完成多种编码方式对比。", map[string]interface{}{"results": results}, &types.ActionFeedback{
			Summary:     "同一输入已经转换为多种编码形式，可直接比较长度膨胀差异。",
			NextHint:    "继续观察不同编码在可读性、长度和兼容性上的特点。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"encoding_count": len(results)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported encoding action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// EncodingFactory 编码演示器工厂
type EncodingFactory struct{}

// Create 创建演示器实例
func (f *EncodingFactory) Create() engine.Simulator {
	return NewEncodingSimulator()
}

// GetDescription 获取描述
func (f *EncodingFactory) GetDescription() types.Description {
	return NewEncodingSimulator().GetDescription()
}

// NewEncodingFactory 创建工厂实例
func NewEncodingFactory() *EncodingFactory {
	return &EncodingFactory{}
}

var _ engine.SimulatorFactory = (*EncodingFactory)(nil)
