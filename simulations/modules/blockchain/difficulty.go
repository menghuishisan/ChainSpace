package blockchain

import (
	"fmt"
	"math"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// DifficultyBlock 难度调整区块
type DifficultyBlock struct {
	Height     uint64    `json:"height"`
	Difficulty uint64    `json:"difficulty"`
	Timestamp  time.Time `json:"timestamp"`
	BlockTime  int64     `json:"block_time"`
}

// DifficultySimulator 难度调整演示器
type DifficultySimulator struct {
	*base.BaseSimulator
	blocks          []*DifficultyBlock
	targetBlockTime int64
	adjustInterval  int
	currentDiff     uint64
}

// NewDifficultySimulator 创建难度调整演示器
func NewDifficultySimulator() *DifficultySimulator {
	sim := &DifficultySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"difficulty",
			"难度调整演示器",
			"展示比特币和以太坊的难度调整算法",
			"blockchain",
			types.ComponentProcess,
		),
		blocks:          make([]*DifficultyBlock, 0),
		targetBlockTime: 10,
		adjustInterval:  10,
		currentDiff:     1000,
	}

	sim.AddParam(types.Param{
		Key: "target_block_time", Name: "目标出块时间(秒)", Type: types.ParamTypeInt,
		Default: 10, Min: 1, Max: 600,
	})
	sim.AddParam(types.Param{
		Key: "adjust_interval", Name: "调整间隔(区块数)", Type: types.ParamTypeInt,
		Default: 10, Min: 1, Max: 2016,
	})
	sim.SetOnTick(sim.onTick)
	return sim
}

// Init 初始化
func (s *DifficultySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	if v, ok := config.Params["target_block_time"]; ok {
		if n, ok := v.(float64); ok {
			s.targetBlockTime = int64(n)
		}
	}
	if v, ok := config.Params["adjust_interval"]; ok {
		if n, ok := v.(float64); ok {
			s.adjustInterval = int(n)
		}
	}

	genesis := &DifficultyBlock{
		Height: 0, Difficulty: s.currentDiff, Timestamp: time.Now(), BlockTime: 0,
	}
	s.blocks = []*DifficultyBlock{genesis}
	s.updateState()
	return nil
}

func (s *DifficultySimulator) onTick(tick uint64) error {
	if tick%5 == 0 {
		s.mineBlock()
	}
	return nil
}

// mineBlock 挖矿产生区块
func (s *DifficultySimulator) mineBlock() {
	lastBlock := s.blocks[len(s.blocks)-1]

	// 模拟出块时间波动
	variance := float64(s.targetBlockTime) * 0.5
	actualTime := s.targetBlockTime + int64((2*variance)*float64(time.Now().UnixNano()%1000)/1000-variance)
	if actualTime < 1 {
		actualTime = 1
	}

	newBlock := &DifficultyBlock{
		Height:     lastBlock.Height + 1,
		Difficulty: s.currentDiff,
		Timestamp:  lastBlock.Timestamp.Add(time.Duration(actualTime) * time.Second),
		BlockTime:  actualTime,
	}

	s.blocks = append(s.blocks, newBlock)

	if len(s.blocks)%s.adjustInterval == 0 {
		s.adjustDifficulty()
	}

	s.EmitEvent("block_mined", "", "", map[string]interface{}{
		"height": newBlock.Height, "difficulty": newBlock.Difficulty, "block_time": actualTime,
	})
	s.updateState()
}

// adjustDifficulty 调整难度
func (s *DifficultySimulator) adjustDifficulty() {
	if len(s.blocks) < s.adjustInterval {
		return
	}

	startIdx := len(s.blocks) - s.adjustInterval
	startBlock := s.blocks[startIdx]
	endBlock := s.blocks[len(s.blocks)-1]

	actualTime := endBlock.Timestamp.Sub(startBlock.Timestamp).Seconds()
	expectedTime := float64(s.adjustInterval) * float64(s.targetBlockTime)

	ratio := expectedTime / actualTime
	ratio = math.Max(0.25, math.Min(4.0, ratio))

	oldDiff := s.currentDiff
	s.currentDiff = uint64(float64(s.currentDiff) * ratio)
	if s.currentDiff < 1 {
		s.currentDiff = 1
	}

	s.EmitEvent("difficulty_adjusted", "", "", map[string]interface{}{
		"old_difficulty": oldDiff,
		"new_difficulty": s.currentDiff,
		"ratio":          ratio,
		"actual_time":    actualTime,
		"expected_time":  expectedTime,
	})
}

func (s *DifficultySimulator) updateState() {
	s.SetGlobalData("current_difficulty", s.currentDiff)
	s.SetGlobalData("block_count", len(s.blocks))
	s.SetGlobalData("target_block_time", s.targetBlockTime)
	s.SetGlobalData("adjust_interval", s.adjustInterval)

	if len(s.blocks) > 0 {
		last := s.blocks[len(s.blocks)-1]
		s.SetGlobalData("latest_height", last.Height)
		s.SetGlobalData("latest_block_time", last.BlockTime)
	}

	recentBlocks := s.blocks
	if len(recentBlocks) > 20 {
		recentBlocks = recentBlocks[len(recentBlocks)-20:]
	}
	s.SetGlobalData("recent_blocks", recentBlocks)

	summary := fmt.Sprintf("当前难度为 %d，已记录 %d 个区块。", s.currentDiff, len(s.blocks))
	nextHint := "可以继续推进出块，观察目标出块时间和实际出块时间如何影响难度调整。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备难度调整",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"current_difficulty": s.currentDiff, "block_count": len(s.blocks)},
	)
}

func (s *DifficultySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "mine_block":
		s.mineBlock()
		return blockchainActionResult("已模拟一次区块出块。", map[string]interface{}{"current_difficulty": s.currentDiff, "block_count": len(s.blocks)}, &types.ActionFeedback{
			Summary:     "新区块已加入统计序列，系统会按间隔评估是否需要调整难度。",
			NextHint:    "继续观察一段时间内的出块速度，再比较难度如何随之变化。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"current_difficulty": s.currentDiff, "block_count": len(s.blocks)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported difficulty action: %s", action)
	}
}

type DifficultyFactory struct{}

func (f *DifficultyFactory) Create() engine.Simulator { return NewDifficultySimulator() }
func (f *DifficultyFactory) GetDescription() types.Description {
	return NewDifficultySimulator().GetDescription()
}
func NewDifficultyFactory() *DifficultyFactory { return &DifficultyFactory{} }

var _ engine.SimulatorFactory = (*DifficultyFactory)(nil)
