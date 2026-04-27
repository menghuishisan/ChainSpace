package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// MiningSimulator 挖矿过程演示器
type MiningSimulator struct {
	*base.BaseSimulator
	mu          sync.Mutex
	blockData   string
	difficulty  uint64
	nonce       uint64
	targetHash  string
	currentHash string
	found       bool
	attempts    uint64
	startTime   time.Time
}

// NewMiningSimulator 创建挖矿演示器
func NewMiningSimulator() *MiningSimulator {
	sim := &MiningSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"mining",
			"挖矿过程演示器",
			"展示PoW挖矿的哈希计算和难度调整过程",
			"blockchain",
			types.ComponentProcess,
		),
		difficulty: 4,
	}

	sim.AddParam(types.Param{
		Key: "difficulty", Name: "难度", Type: types.ParamTypeInt,
		Default: 4, Min: 1, Max: 8,
	})
	sim.SetOnTick(sim.onTick)
	return sim
}

// Init 初始化
func (s *MiningSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	if v, ok := config.Params["difficulty"]; ok {
		if n, ok := v.(float64); ok {
			s.difficulty = uint64(n)
		}
	}
	s.blockData = "Block Data Example"
	s.computeTarget()
	s.updateState()
	return nil
}

func (s *MiningSimulator) computeTarget() {
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-s.difficulty*4))
	s.targetHash = fmt.Sprintf("%064x", target)
}

func (s *MiningSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.found {
		return nil
	}

	for i := 0; i < 100; i++ {
		s.nonce++
		s.attempts++
		hash := s.computeHash(s.blockData, s.nonce)
		s.currentHash = hash

		if s.meetsTarget(hash) {
			s.found = true
			s.EmitEvent("block_found", "", "", map[string]interface{}{
				"nonce": s.nonce, "hash": hash[:16], "attempts": s.attempts,
			})
			break
		}
	}
	s.updateState()
	return nil
}

func (s *MiningSimulator) computeHash(data string, nonce uint64) string {
	input := fmt.Sprintf("%s%d", data, nonce)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func (s *MiningSimulator) meetsTarget(hash string) bool {
	hashInt := new(big.Int)
	hashInt.SetString(hash, 16)
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-s.difficulty*4))
	return hashInt.Cmp(target) < 0
}

// StartMining 开始挖矿
func (s *MiningSimulator) StartMining(blockData string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blockData = blockData
	s.nonce = 0
	s.attempts = 0
	s.found = false
	s.startTime = time.Now()
	s.EmitEvent("mining_started", "", "", map[string]interface{}{
		"block_data": blockData, "difficulty": s.difficulty,
	})
}

// Reset 重置
func (s *MiningSimulator) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nonce = 0
	s.attempts = 0
	s.found = false
	s.updateState()
	return s.BaseSimulator.Reset()
}

func (s *MiningSimulator) updateState() {
	s.SetGlobalData("difficulty", s.difficulty)
	s.SetGlobalData("nonce", s.nonce)
	s.SetGlobalData("attempts", s.attempts)
	s.SetGlobalData("current_hash", s.currentHash)
	s.SetGlobalData("target_hash", s.targetHash[:16])
	s.SetGlobalData("found", s.found)

	summary := fmt.Sprintf("当前难度为 %d，已经尝试 %d 次 nonce。", s.difficulty, s.attempts)
	nextHint := "可以继续启动挖矿并推进 nonce，观察何时找到满足目标的哈希。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备挖矿",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"difficulty": s.difficulty, "attempts": s.attempts, "found": s.found},
	)
}

func (s *MiningSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "start_mining":
		blockData := "Block Data Example"
		if raw, ok := params["block_data"].(string); ok && raw != "" {
			blockData = raw
		}
		s.StartMining(blockData)
		return blockchainActionResult("已启动一次挖矿过程。", map[string]interface{}{"block_data": blockData}, &types.ActionFeedback{
			Summary:     "新的候选区块已进入挖矿阶段，系统会不断尝试 nonce。",
			NextHint:    "继续单步或自动推进，观察当前哈希何时低于目标。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"difficulty": s.difficulty, "found": s.found},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported mining action: %s", action)
	}
}

type MiningFactory struct{}

func (f *MiningFactory) Create() engine.Simulator { return NewMiningSimulator() }
func (f *MiningFactory) GetDescription() types.Description {
	return NewMiningSimulator().GetDescription()
}
func NewMiningFactory() *MiningFactory { return &MiningFactory{} }

var _ engine.SimulatorFactory = (*MiningFactory)(nil)
