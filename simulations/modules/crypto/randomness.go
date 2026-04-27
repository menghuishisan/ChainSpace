package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
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
// 随机数数据结构
// =============================================================================

// RandomnessSource 随机数来源
type RandomnessSource struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`   // blockhash/vrf/commit-reveal/beacon
	Input       string    `json:"input"`  // 输入数据
	Output      string    `json:"output"` // 随机数输出
	Proof       string    `json:"proof"`  // 可验证性证明(如有)
	BlockNumber uint64    `json:"block_number"`
	Timestamp   time.Time `json:"timestamp"`
}

// CommitRevealRound Commit-Reveal轮次
type CommitRevealRound struct {
	ID          string            `json:"id"`
	Phase       string            `json:"phase"`        // commit/reveal/finalize
	Commits     map[string]string `json:"commits"`      // 参与者 -> 承诺
	Reveals     map[string]string `json:"reveals"`      // 参与者 -> 揭示值
	FinalRandom string            `json:"final_random"` // 最终随机数
	Timestamp   time.Time         `json:"timestamp"`
}

// VRFOutput VRF输出
type VRFOutput struct {
	Alpha string `json:"alpha"` // 输入
	Beta  string `json:"beta"`  // 输出(随机数)
	Pi    string `json:"pi"`    // 证明
	Valid bool   `json:"valid"` // 验证结果
}

// =============================================================================
// RandomnessSimulator 链上随机数演示器
// =============================================================================

// RandomnessSimulator 链上随机数演示器
// 演示区块链中获取随机数的各种方案
//
// 1. 区块哈希 (Block Hash):
//   - 使用未来区块的哈希作为随机源
//   - 问题: 矿工可以选择不发布有利区块
//   - 适用: 低价值场景
//
// 2. Commit-Reveal:
//   - 参与者先提交随机数的哈希(Commit)
//   - 所有人提交后再揭示原值(Reveal)
//   - 最终随机数 = XOR(所有揭示值)
//   - 问题: 最后揭示者可以选择不揭示
//
// 3. VRF (可验证随机函数):
//   - 使用私钥生成确定性但不可预测的输出
//   - 可提供证明，任何人可验证
//   - 应用: Algorand、Chainlink VRF
//
// 4. 随机信标 (Random Beacon):
//   - 去中心化的随机数生成服务
//   - 如: drand、Chainlink VRF
type RandomnessSimulator struct {
	*base.BaseSimulator
	mu           sync.RWMutex
	blockNumber  uint64
	blockHashes  map[uint64]string
	commitRounds map[string]*CommitRevealRound
	vrfSecrets   map[string]*big.Int
	history      []*RandomnessSource
}

// NewRandomnessSimulator 创建随机数演示器
func NewRandomnessSimulator() *RandomnessSimulator {
	sim := &RandomnessSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"randomness",
			"链上随机数演示器",
			"演示Block Hash、Commit-Reveal、VRF等链上随机数方案",
			"crypto",
			types.ComponentTool,
		),
		blockHashes:  make(map[uint64]string),
		commitRounds: make(map[string]*CommitRevealRound),
		vrfSecrets:   make(map[string]*big.Int),
		history:      make([]*RandomnessSource, 0),
	}

	sim.AddParam(types.Param{
		Key:         "method",
		Name:        "随机数方案",
		Description: "选择随机数生成方案",
		Type:        types.ParamTypeSelect,
		Default:     "commit-reveal",
		Options: []types.Option{
			{Label: "区块哈希", Value: "blockhash"},
			{Label: "Commit-Reveal", Value: "commit-reveal"},
			{Label: "VRF", Value: "vrf"},
		},
	})

	return sim
}

// Init 初始化演示器
func (s *RandomnessSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 生成模拟区块哈希
	s.blockNumber = 100
	for i := uint64(0); i < 10; i++ {
		hash := make([]byte, 32)
		rand.Read(hash)
		s.blockHashes[s.blockNumber-i] = hex.EncodeToString(hash)
	}

	s.updateState()
	return nil
}

// =============================================================================
// 区块哈希方案
// =============================================================================

// GetBlockHash 获取区块哈希作为随机源
//
// 使用方式: randomness = hash(blockhash(blockNumber) + seed)
// 安全性: 矿工可以预知并可能影响结果
func (s *RandomnessSimulator) GetBlockHash(blockNum uint64, seed string) (*RandomnessSource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	blockHash := s.blockHashes[blockNum]
	if blockHash == "" {
		// 生成新的区块哈希
		hash := make([]byte, 32)
		rand.Read(hash)
		blockHash = hex.EncodeToString(hash)
		s.blockHashes[blockNum] = blockHash
	}

	// 组合区块哈希和种子
	combined := blockHash + seed
	finalHash := sha256.Sum256([]byte(combined))
	randomness := hex.EncodeToString(finalHash[:])

	result := &RandomnessSource{
		ID:          fmt.Sprintf("rnd-%d", len(s.history)+1),
		Type:        "blockhash",
		Input:       fmt.Sprintf("block:%d,seed:%s", blockNum, seed),
		Output:      randomness,
		BlockNumber: blockNum,
		Timestamp:   time.Now(),
	}

	s.history = append(s.history, result)

	s.EmitEvent("randomness_generated", "", "", map[string]interface{}{
		"method":       "blockhash",
		"block_number": blockNum,
		"randomness":   randomness[:16] + "...",
		"warning":      "矿工可能影响结果",
	})

	s.updateState()
	return result, nil
}

// =============================================================================
// Commit-Reveal方案
// =============================================================================

// StartCommitReveal 开始新的Commit-Reveal轮次
func (s *RandomnessSimulator) StartCommitReveal(roundID string) *CommitRevealRound {
	s.mu.Lock()
	defer s.mu.Unlock()

	round := &CommitRevealRound{
		ID:        roundID,
		Phase:     "commit",
		Commits:   make(map[string]string),
		Reveals:   make(map[string]string),
		Timestamp: time.Now(),
	}

	s.commitRounds[roundID] = round

	s.EmitEvent("commit_reveal_started", "", "", map[string]interface{}{
		"round_id": roundID,
		"phase":    "commit",
	})

	s.updateState()
	return round
}

// Commit 提交承诺 (哈希值)
// commitment = hash(secret || salt)
func (s *RandomnessSimulator) Commit(roundID, participantID, secret, salt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	round := s.commitRounds[roundID]
	if round == nil {
		return fmt.Errorf("轮次不存在: %s", roundID)
	}
	if round.Phase != "commit" {
		return fmt.Errorf("当前不是提交阶段")
	}

	// 计算承诺 = hash(secret || salt)
	data := secret + salt
	hash := sha256.Sum256([]byte(data))
	commitment := hex.EncodeToString(hash[:])

	round.Commits[participantID] = commitment

	s.EmitEvent("committed", types.NodeID(participantID), "", map[string]interface{}{
		"round_id":   roundID,
		"commitment": commitment[:16] + "...",
	})

	s.updateState()
	return nil
}

// StartRevealPhase 开始揭示阶段
func (s *RandomnessSimulator) StartRevealPhase(roundID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	round := s.commitRounds[roundID]
	if round == nil {
		return fmt.Errorf("轮次不存在: %s", roundID)
	}

	if len(round.Commits) < 2 {
		return fmt.Errorf("需要至少2个参与者")
	}

	round.Phase = "reveal"

	s.EmitEvent("reveal_phase_started", "", "", map[string]interface{}{
		"round_id":     roundID,
		"commit_count": len(round.Commits),
	})

	s.updateState()
	return nil
}

// Reveal 揭示原值
func (s *RandomnessSimulator) Reveal(roundID, participantID, secret, salt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	round := s.commitRounds[roundID]
	if round == nil {
		return fmt.Errorf("轮次不存在: %s", roundID)
	}
	if round.Phase != "reveal" {
		return fmt.Errorf("当前不是揭示阶段")
	}

	// 验证揭示值与承诺匹配
	data := secret + salt
	hash := sha256.Sum256([]byte(data))
	expectedCommitment := hex.EncodeToString(hash[:])

	if round.Commits[participantID] != expectedCommitment {
		return fmt.Errorf("揭示值与承诺不匹配")
	}

	round.Reveals[participantID] = secret

	s.EmitEvent("revealed", types.NodeID(participantID), "", map[string]interface{}{
		"round_id": roundID,
		"verified": true,
	})

	// 检查是否所有人都已揭示
	if len(round.Reveals) == len(round.Commits) {
		s.finalizeRound(round)
	}

	s.updateState()
	return nil
}

// finalizeRound 完成轮次，计算最终随机数
func (s *RandomnessSimulator) finalizeRound(round *CommitRevealRound) {
	// 最终随机数 = XOR(所有揭示值的哈希)
	result := make([]byte, 32)
	for _, secret := range round.Reveals {
		hash := sha256.Sum256([]byte(secret))
		for i := 0; i < 32; i++ {
			result[i] ^= hash[i]
		}
	}

	round.FinalRandom = hex.EncodeToString(result)
	round.Phase = "finalized"

	// 记录历史
	s.history = append(s.history, &RandomnessSource{
		ID:        fmt.Sprintf("rnd-%d", len(s.history)+1),
		Type:      "commit-reveal",
		Input:     fmt.Sprintf("round:%s,participants:%d", round.ID, len(round.Reveals)),
		Output:    round.FinalRandom,
		Timestamp: time.Now(),
	})

	s.EmitEvent("round_finalized", "", "", map[string]interface{}{
		"round_id":     round.ID,
		"participants": len(round.Reveals),
		"randomness":   round.FinalRandom[:16] + "...",
	})
}

// =============================================================================
// VRF方案
// =============================================================================

// GenerateVRFKey 生成VRF密钥
func (s *RandomnessSimulator) GenerateVRFKey(keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成256位随机私钥
	sk, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	if err != nil {
		return fmt.Errorf("生成密钥失败: %v", err)
	}

	s.vrfSecrets[keyID] = sk

	s.EmitEvent("vrf_key_generated", "", "", map[string]interface{}{
		"key_id": keyID,
	})

	s.updateState()
	return nil
}

// VRFProve 生成VRF输出和证明
//
// VRF特性:
// - 确定性: 相同输入产生相同输出
// - 不可预测: 不知私钥无法预测输出
// - 可验证: 任何人可用公钥验证
func (s *RandomnessSimulator) VRFProve(keyID, alpha string) (*VRFOutput, error) {
	s.mu.RLock()
	sk := s.vrfSecrets[keyID]
	s.mu.RUnlock()

	if sk == nil {
		return nil, fmt.Errorf("密钥不存在: %s", keyID)
	}

	// 简化的VRF: beta = HMAC(sk, alpha)
	// 真实VRF使用椭圆曲线
	h := hmac.New(sha256.New, sk.Bytes())
	h.Write([]byte(alpha))
	beta := h.Sum(nil)

	// 简化的证明: pi = HMAC(sk, alpha || beta)
	h.Reset()
	h.Write([]byte(alpha))
	h.Write(beta)
	pi := h.Sum(nil)

	output := &VRFOutput{
		Alpha: alpha,
		Beta:  hex.EncodeToString(beta),
		Pi:    hex.EncodeToString(pi),
		Valid: true,
	}

	// 记录历史
	s.mu.Lock()
	s.history = append(s.history, &RandomnessSource{
		ID:        fmt.Sprintf("rnd-%d", len(s.history)+1),
		Type:      "vrf",
		Input:     alpha,
		Output:    output.Beta,
		Proof:     output.Pi,
		Timestamp: time.Now(),
	})
	s.mu.Unlock()

	s.EmitEvent("vrf_proved", "", "", map[string]interface{}{
		"key_id":     keyID,
		"alpha":      alpha,
		"beta":       output.Beta[:16] + "...",
		"verifiable": true,
	})

	s.updateState()
	return output, nil
}

// RandomToNumber 将随机数转换为范围内的数字
// 用于抽奖、选择等场景
func (s *RandomnessSimulator) RandomToNumber(randomHex string, maxValue uint64) (uint64, error) {
	randomBytes, err := hex.DecodeString(randomHex)
	if err != nil {
		return 0, fmt.Errorf("无效的随机数: %v", err)
	}

	// 使用前8字节转换为uint64
	if len(randomBytes) < 8 {
		return 0, fmt.Errorf("随机数太短")
	}

	num := binary.BigEndian.Uint64(randomBytes[:8])
	result := num % maxValue

	s.EmitEvent("random_to_number", "", "", map[string]interface{}{
		"max_value": maxValue,
		"result":    result,
	})

	return result, nil
}

// updateState 更新状态
func (s *RandomnessSimulator) updateState() {
	s.SetGlobalData("block_number", s.blockNumber)
	s.SetGlobalData("history_count", len(s.history))
	s.SetGlobalData("active_rounds", len(s.commitRounds))
	s.SetGlobalData("vrf_keys", len(s.vrfSecrets))

	// 最近的随机数
	recent := s.history
	if len(recent) > 5 {
		recent = recent[len(recent)-5:]
	}
	recentList := make([]map[string]interface{}, 0)
	for _, r := range recent {
		recentList = append(recentList, map[string]interface{}{
			"id":     r.ID,
			"type":   r.Type,
			"output": r.Output[:min(16, len(r.Output))] + "...",
		})
	}
	s.SetGlobalData("recent_randomness", recentList)

	summary := fmt.Sprintf("当前区块高度 %d，已记录 %d 次随机数输出。", s.blockNumber, len(s.history))
	nextHint := "可以继续生成区块哈希随机数，或开启提交-揭示流程。"
	if len(s.commitRounds) > 0 {
		nextHint = "当前存在提交-揭示轮次，可以继续提交承诺或揭示随机值。"
	}

	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		"准备随机源",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{
			"history_count": len(s.history),
			"active_rounds": len(s.commitRounds),
			"vrf_keys":      len(s.vrfSecrets),
		},
	)
}

// ExecuteAction 执行随机数演示器的教学动作。
func (s *RandomnessSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "generate_blockhash":
		blockNumber := uint64(getIntParam(params, "block_number", int(s.blockNumber+1)))
		result, err := s.GetBlockHash(blockNumber, "ChainSpace")
		if err != nil {
			return nil, err
		}
		return cryptoActionResult(
			"已生成区块哈希随机数",
			map[string]interface{}{
				"block_number": blockNumber,
				"output":       result.Output,
			},
			&types.ActionFeedback{
				Summary:     "系统已根据指定区块高度生成新的随机输出。",
				NextHint:    "继续比较不同区块高度的输出，理解区块哈希随机源的局限。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"block_number": blockNumber},
			},
		), nil
	case "start_commit_reveal":
		roundID := getStringParam(params, "round_id", fmt.Sprintf("round-%d", len(s.commitRounds)+1))
		round := s.StartCommitReveal(roundID)
		return cryptoActionResult(
			"已启动提交-揭示轮次",
			map[string]interface{}{
				"round_id": roundID,
				"phase":    round.Phase,
			},
			&types.ActionFeedback{
				Summary:     "新的提交-揭示随机数轮次已创建，可继续提交承诺并在后续揭示随机值。",
				NextHint:    "继续增加参与方或揭示随机值，观察最终随机数如何合成。",
				EffectScope: "crypto",
				ResultState: map[string]interface{}{"round_id": roundID, "phase": round.Phase},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported randomness action: %s", action)
	}
}

func getIntParam(params map[string]interface{}, key string, fallback int) int {
	if params == nil {
		return fallback
	}
	value, ok := params[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return fallback
	}
}

// =============================================================================
// 工厂
// =============================================================================

// RandomnessFactory 随机数演示器工厂
type RandomnessFactory struct{}

// Create 创建演示器实例
func (f *RandomnessFactory) Create() engine.Simulator {
	return NewRandomnessSimulator()
}

// GetDescription 获取描述
func (f *RandomnessFactory) GetDescription() types.Description {
	return NewRandomnessSimulator().GetDescription()
}

// NewRandomnessFactory 创建工厂实例
func NewRandomnessFactory() *RandomnessFactory {
	return &RandomnessFactory{}
}

var _ engine.SimulatorFactory = (*RandomnessFactory)(nil)
