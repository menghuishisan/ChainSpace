package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// Commitment 承诺结构
type Commitment struct {
	ID        string    `json:"id"`         // 承诺ID
	Hash      string    `json:"hash"`       // 承诺哈希
	Value     string    `json:"value"`      // 原始值(打开后可见)
	Nonce     string    `json:"nonce"`      // 随机数(打开后可见)
	Opened    bool      `json:"opened"`     // 是否已打开
	Verified  bool      `json:"verified"`   // 是否已验证
	CreatedAt time.Time `json:"created_at"` // 创建时间
	OpenedAt  time.Time `json:"opened_at"`  // 打开时间
}

// CommitmentSimulator 承诺方案演示器
// 展示隐藏-绑定承诺的工作原理，支持哈希承诺
type CommitmentSimulator struct {
	*base.BaseSimulator
	commitments map[string]*Commitment // 承诺映射
	history     []*Commitment          // 历史记录
}

// NewCommitmentSimulator 创建承诺方案演示器
func NewCommitmentSimulator() *CommitmentSimulator {
	sim := &CommitmentSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"commitment",
			"承诺方案演示器",
			"展示哈希承诺和Pedersen承诺的隐藏-绑定特性",
			"crypto",
			types.ComponentTool,
		),
		commitments: make(map[string]*Commitment),
		history:     make([]*Commitment, 0),
	}
	return sim
}

// Init 初始化演示器
func (s *CommitmentSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// Commit 创建承诺
// value: 要承诺的值
// 返回: 承诺哈希
func (s *CommitmentSimulator) Commit(value string) string {
	// 生成随机数(nonce)
	nonceBytes := make([]byte, 32)
	rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)

	// 计算承诺: H(value || nonce)
	data := value + nonce
	hash := sha256.Sum256([]byte(data))
	hashHex := hex.EncodeToString(hash[:])

	// 创建承诺记录
	commitment := &Commitment{
		ID:        fmt.Sprintf("commit-%d", len(s.history)+1),
		Hash:      hashHex,
		Value:     value,
		Nonce:     nonce,
		Opened:    false,
		CreatedAt: time.Now(),
	}

	s.commitments[hashHex] = commitment
	s.history = append(s.history, commitment)

	// 发送事件
	s.EmitEvent("committed", "", "", map[string]interface{}{
		"id":   commitment.ID,
		"hash": hashHex[:16] + "...",
	})

	s.updateState()
	return hashHex
}

// Open 打开承诺
// commitHash: 承诺哈希
// 返回: 原始值和随机数
func (s *CommitmentSimulator) Open(commitHash string) (string, string, error) {
	commitment := s.commitments[commitHash]
	if commitment == nil {
		return "", "", fmt.Errorf("承诺不存在: %s", commitHash[:16])
	}

	if commitment.Opened {
		return commitment.Value, commitment.Nonce, nil
	}

	commitment.Opened = true
	commitment.OpenedAt = time.Now()

	// 发送事件
	s.EmitEvent("opened", "", "", map[string]interface{}{
		"id":    commitment.ID,
		"value": commitment.Value,
	})

	s.updateState()
	return commitment.Value, commitment.Nonce, nil
}

// Verify 验证承诺
// commitHash: 承诺哈希
// value: 声称的原始值
// nonce: 声称的随机数
func (s *CommitmentSimulator) Verify(commitHash, value, nonce string) bool {
	// 重新计算哈希
	data := value + nonce
	hash := sha256.Sum256([]byte(data))
	computedHash := hex.EncodeToString(hash[:])

	valid := computedHash == commitHash

	// 更新承诺状态
	if commitment := s.commitments[commitHash]; commitment != nil {
		commitment.Verified = valid
	}

	// 发送事件
	s.EmitEvent("verified", "", "", map[string]interface{}{
		"hash":  commitHash[:16] + "...",
		"valid": valid,
	})

	s.updateState()
	return valid
}

// GetCommitment 获取承诺信息
func (s *CommitmentSimulator) GetCommitment(commitHash string) *Commitment {
	return s.commitments[commitHash]
}

// DemonstrateBinding 演示绑定性
// 尝试用不同的值打开同一个承诺，证明绑定性
func (s *CommitmentSimulator) DemonstrateBinding(commitHash, fakeValue string) bool {
	commitment := s.commitments[commitHash]
	if commitment == nil {
		return false
	}

	// 尝试找到一个nonce使得H(fakeValue || nonce') = commitHash
	// 这在计算上是不可行的
	for i := 0; i < 1000; i++ {
		nonceBytes := make([]byte, 32)
		rand.Read(nonceBytes)
		testNonce := hex.EncodeToString(nonceBytes)

		data := fakeValue + testNonce
		hash := sha256.Sum256([]byte(data))
		if hex.EncodeToString(hash[:]) == commitHash {
			// 找到碰撞(实际上不可能)
			return true
		}
	}

	s.EmitEvent("binding_demo", "", "", map[string]interface{}{
		"hash":       commitHash[:16] + "...",
		"fake_value": fakeValue,
		"success":    false,
		"message":    "无法找到碰撞，承诺具有绑定性",
	})

	return false
}

// updateState 更新状态
func (s *CommitmentSimulator) updateState() {
	openedCount := 0
	verifiedCount := 0
	for _, c := range s.commitments {
		if c.Opened {
			openedCount++
		}
		if c.Verified {
			verifiedCount++
		}
	}

	s.SetGlobalData("total_commitments", len(s.commitments))
	s.SetGlobalData("opened_count", openedCount)
	s.SetGlobalData("verified_count", verifiedCount)

	// 最近10条历史
	recentHistory := s.history
	if len(recentHistory) > 10 {
		recentHistory = recentHistory[len(recentHistory)-10:]
	}
	s.SetGlobalData("recent_history", recentHistory)

	summary := fmt.Sprintf("当前共有 %d 份承诺，已打开 %d 份，已验证 %d 份。", len(s.commitments), openedCount, verifiedCount)
	nextHint := "先提交一个承诺，再打开或验证它，观察绑定性和隐藏性。"
	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备承诺",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"total_commitments": len(s.commitments), "opened_count": openedCount, "verified_count": verifiedCount},
	)
}

func (s *CommitmentSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_commitment":
		value := "secret"
		if raw, ok := params["value"].(string); ok && raw != "" {
			value = raw
		}
		hash := s.Commit(value)
		return cryptoActionResult("已创建一份承诺。", map[string]interface{}{"commitment": hash}, &types.ActionFeedback{
			Summary:     "承诺值已经生成，下一步可以打开或验证它。",
			NextHint:    "继续执行打开或验证，观察承诺如何在不暴露原文的情况下先固定结果。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"commitment": hash},
		}), nil
	case "open_commitment":
		commitHash := ""
		if raw, ok := params["commitment"].(string); ok && raw != "" {
			commitHash = raw
		}
		if commitHash == "" {
			for hash := range s.commitments {
				commitHash = hash
				break
			}
		}
		if commitHash == "" {
			return nil, fmt.Errorf("no commitment available")
		}
		value, nonce, err := s.Open(commitHash)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已打开一份承诺。", map[string]interface{}{"value": value, "nonce": nonce, "commitment": commitHash}, &types.ActionFeedback{
			Summary:     "承诺对应的原始值和随机数已经揭示。",
			NextHint:    "继续验证该承诺，确认哈希能否重建到相同的承诺值。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"commitment": commitHash},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported commitment action: %s", action)
	}
}

// CommitmentFactory 承诺方案演示器工厂
type CommitmentFactory struct{}

// Create 创建演示器实例
func (f *CommitmentFactory) Create() engine.Simulator {
	return NewCommitmentSimulator()
}

// GetDescription 获取描述
func (f *CommitmentFactory) GetDescription() types.Description {
	return NewCommitmentSimulator().GetDescription()
}

// NewCommitmentFactory 创建工厂实例
func NewCommitmentFactory() *CommitmentFactory {
	return &CommitmentFactory{}
}

var _ engine.SimulatorFactory = (*CommitmentFactory)(nil)
