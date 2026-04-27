package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// HashRecord 表示一次哈希计算记录。
type HashRecord struct {
	ID        string    `json:"id"`
	Algorithm string    `json:"algorithm"`
	Input     string    `json:"input"`
	InputHex  string    `json:"input_hex"`
	Output    string    `json:"output"`
	BitLength int       `json:"bit_length"`
	Timestamp time.Time `json:"timestamp"`
}

// HashSimulator 用于演示多种哈希算法及其性质。
type HashSimulator struct {
	*base.BaseSimulator
	history    []*HashRecord
	algorithms map[string]int
	avalanche  map[string]interface{}
}

// NewHashSimulator 创建哈希演示器。
func NewHashSimulator() *HashSimulator {
	sim := &HashSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"hash",
			"哈希算法演示器",
			"展示 MD5、SHA-1、SHA-256、SHA-512 的摘要计算、雪崩效应和完整性校验。",
			"crypto",
			types.ComponentTool,
		),
		history: make([]*HashRecord, 0),
		algorithms: map[string]int{
			"md5":    128,
			"sha1":   160,
			"sha256": 256,
			"sha512": 512,
		},
	}

	sim.AddParam(types.Param{
		Key:         "default_algorithm",
		Name:        "默认算法",
		Description: "默认使用的哈希算法。",
		Type:        types.ParamTypeSelect,
		Default:     "sha256",
		Options: []types.Option{
			{Label: "MD5 (128位)", Value: "md5"},
			{Label: "SHA-1 (160位)", Value: "sha1"},
			{Label: "SHA-256 (256位)", Value: "sha256"},
			{Label: "SHA-512 (512位)", Value: "sha512"},
		},
	})

	return sim
}

// Init 初始化演示器。
func (s *HashSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// Hash 计算指定算法的哈希值。
func (s *HashSimulator) Hash(algorithm string, data string) string {
	var hashBytes []byte

	switch strings.ToLower(algorithm) {
	case "md5":
		h := md5.Sum([]byte(data))
		hashBytes = h[:]
	case "sha1":
		h := sha1.Sum([]byte(data))
		hashBytes = h[:]
	case "sha256":
		h := sha256.Sum256([]byte(data))
		hashBytes = h[:]
	case "sha512":
		h := sha512.Sum512([]byte(data))
		hashBytes = h[:]
	default:
		h := sha256.Sum256([]byte(data))
		hashBytes = h[:]
		algorithm = "sha256"
	}

	result := hex.EncodeToString(hashBytes)
	record := &HashRecord{
		ID:        fmt.Sprintf("hash-%d", len(s.history)+1),
		Algorithm: algorithm,
		Input:     data,
		InputHex:  hex.EncodeToString([]byte(data)),
		Output:    result,
		BitLength: len(hashBytes) * 8,
		Timestamp: time.Now(),
	}
	s.history = append(s.history, record)

	s.EmitEvent("hash_computed", "", "", map[string]interface{}{
		"algorithm":  algorithm,
		"input_len":  len(data),
		"output":     result[:16] + "...",
		"bit_length": record.BitLength,
	})

	s.updateState()
	return result
}

// HashMultiple 使用多种算法计算同一输入的摘要。
func (s *HashSimulator) HashMultiple(data string) map[string]string {
	results := make(map[string]string)
	for algorithm := range s.algorithms {
		results[algorithm] = s.Hash(algorithm, data)
	}

	s.EmitEvent("hash_multiple", "", "", map[string]interface{}{
		"algorithms": len(results),
		"input_len":  len(data),
	})

	return results
}

// DemonstrateAvalanche 演示雪崩效应。
func (s *HashSimulator) DemonstrateAvalanche(data string, algorithm string) map[string]interface{} {
	original := s.Hash(algorithm, data)

	modified := data
	if len(data) > 0 {
		firstByte := data[0] ^ 1
		modified = string(firstByte) + data[1:]
	} else {
		modified = string(byte(1))
	}

	changed := s.Hash(algorithm, modified)
	differentBits := s.countDifferentBits(original, changed)
	totalBits := len(original) * 4
	changePercentage := float64(differentBits) / float64(totalBits) * 100

	result := map[string]interface{}{
		"original_input":   data,
		"modified_input":   modified,
		"original_hash":    original,
		"modified_hash":    changed,
		"different_bits":   differentBits,
		"total_bits":       totalBits,
		"change_percentage": changePercentage,
		"change_percent":   changePercentage,
	}

	s.avalanche = result
	s.EmitEvent("avalanche_demo", "", "", map[string]interface{}{
		"different_bits":    differentBits,
		"change_percentage": changePercentage,
	})

	s.updateState()
	return result
}

func (s *HashSimulator) countDifferentBits(hex1, hex2 string) int {
	if len(hex1) != len(hex2) {
		return -1
	}

	count := 0
	for i := 0; i < len(hex1); i++ {
		b1 := hexCharToByte(hex1[i])
		b2 := hexCharToByte(hex2[i])
		xor := b1 ^ b2
		for xor > 0 {
			count += int(xor & 1)
			xor >>= 1
		}
	}
	return count
}

func hexCharToByte(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 10
	}
	if c >= 'A' && c <= 'F' {
		return c - 'A' + 10
	}
	return 0
}

// VerifyIntegrity 校验输入与期望摘要是否一致。
func (s *HashSimulator) VerifyIntegrity(data, expectedHash, algorithm string) bool {
	computedHash := s.Hash(algorithm, data)
	valid := computedHash == expectedHash

	s.EmitEvent("integrity_verified", "", "", map[string]interface{}{
		"algorithm": algorithm,
		"valid":     valid,
	})

	return valid
}

// GetHistory 返回历史记录。
func (s *HashSimulator) GetHistory() []*HashRecord {
	return s.history
}

// ClearHistory 清空历史记录。
func (s *HashSimulator) ClearHistory() {
	s.history = make([]*HashRecord, 0)
	s.updateState()
}

// updateState 更新状态。
func (s *HashSimulator) updateState() {
	s.SetGlobalData("algorithms", s.algorithms)
	s.SetGlobalData("history_count", len(s.history))
	s.SetGlobalData("avalanche", s.avalanche)

	recentHistory := s.history
	if len(recentHistory) > 20 {
		recentHistory = recentHistory[len(recentHistory)-20:]
	}
	s.SetGlobalData("recent_history", recentHistory)

	summary := fmt.Sprintf("已记录 %d 次哈希计算", len(s.history))
	nextHint := "尝试切换算法或修改输入内容，观察摘要如何变化。"
	if s.avalanche != nil {
		if changePercent, ok := s.avalanche["change_percent"].(float64); ok {
			summary = fmt.Sprintf("最近一次雪崩实验的变化比例为 %.2f%%", changePercent)
		} else {
			summary = "最近一次雪崩实验已完成，可重点观察摘要变化。"
		}
		nextHint = "继续执行完整性校验，观察输入篡改后摘要是否失配。"
	}

	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备输入",
		summary,
		nextHint,
		0.35,
		map[string]interface{}{
			"history_count": len(s.history),
			"algorithms":    len(s.algorithms),
		},
	)
}

// ExecuteAction 执行哈希演示器的教学动作。
func (s *HashSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "compute_hash":
		algorithm := strings.ToLower(getStringParam(params, "algorithm", "sha256"))
		input := getStringParam(params, "input", "ChainSpace")
		output := s.Hash(algorithm, input)
		return cryptoActionResult(
			"已完成一次哈希计算",
			map[string]interface{}{
				"algorithm": algorithm,
				"input":     input,
				"output":    output,
			},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("使用 %s 对输入生成了新的摘要。", strings.ToUpper(algorithm)),
				NextHint:    "继续对比不同算法或调整输入内容，观察输出变化。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"algorithm": algorithm, "history_count": len(s.history)},
			},
		), nil
	case "compare_algorithms":
		input := getStringParam(params, "input", "ChainSpace")
		results := s.HashMultiple(input)
		return cryptoActionResult(
			"已完成多算法对比",
			map[string]interface{}{
				"input":   input,
				"results": results,
			},
			&types.ActionFeedback{
				Summary:     "相同输入已生成多种哈希结果，可直接比较摘要差异和输出长度。",
				NextHint:    "继续观察不同算法的位数差异，理解摘要长度与碰撞难度的关系。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"algorithm_count": len(results)},
			},
		), nil
	case "demo_avalanche":
		algorithm := strings.ToLower(getStringParam(params, "algorithm", "sha256"))
		input := getStringParam(params, "input", "ChainSpace")
		result := s.DemonstrateAvalanche(input, algorithm)
		return cryptoActionResult(
			"已演示雪崩效应",
			result,
			&types.ActionFeedback{
				Summary:     "输入的微小变化已导致摘要结果发生显著改变。",
				NextHint:    "重点观察变化位数和变化比例，理解哈希扩散特性。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"has_avalanche": true},
			},
		), nil
	case "verify_integrity":
		algorithm := strings.ToLower(getStringParam(params, "algorithm", "sha256"))
		input := getStringParam(params, "input", "ChainSpace")
		expectedHash := getStringParam(params, "expected_hash", s.Hash(algorithm, input))
		valid := s.VerifyIntegrity(input, expectedHash, algorithm)
		description := "输入内容与期望摘要不一致，完整性验证失败。"
		if valid {
			description = "输入内容与期望摘要一致，完整性验证通过。"
		}
		return cryptoActionResult(
			"已完成完整性校验",
			map[string]interface{}{
				"algorithm":     algorithm,
				"input":         input,
				"expected_hash": expectedHash,
				"valid":         valid,
			},
			&types.ActionFeedback{
				Summary:     description,
				NextHint:    "可以修改输入或期望摘要，再次验证校验结果。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"valid": valid},
			},
		), nil
	case "clear_history":
		s.ClearHistory()
		s.avalanche = nil
		s.updateState()
		return cryptoActionResult(
			"已清空哈希历史",
			nil,
			&types.ActionFeedback{
				Summary:     "哈希计算记录和雪崩实验结果均已重置。",
				NextHint:    "可以重新输入数据，开始新一轮哈希实验。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"history_count": 0},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported hash action: %s", action)
	}
}

func getStringParam(params map[string]interface{}, key, fallback string) string {
	if params == nil {
		return fallback
	}
	value, ok := params[key]
	if !ok {
		return fallback
	}
	if str, ok := value.(string); ok && strings.TrimSpace(str) != "" {
		return str
	}
	return fallback
}

// HashFactory 哈希演示器工厂。
type HashFactory struct{}

// Create 创建演示器实例。
func (f *HashFactory) Create() engine.Simulator {
	return NewHashSimulator()
}

// GetDescription 返回描述信息。
func (f *HashFactory) GetDescription() types.Description {
	return NewHashSimulator().GetDescription()
}

// NewHashFactory 创建工厂实例。
func NewHashFactory() *HashFactory {
	return &HashFactory{}
}

var _ engine.SimulatorFactory = (*HashFactory)(nil)
