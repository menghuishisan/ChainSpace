package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// TimestampManipulationSimulator 演示区块时间戳被轻微操纵后带来的业务风险。
type TimestampManipulationSimulator struct {
	*base.BaseSimulator
	records []map[string]interface{}
}

// NewTimestampManipulationSimulator 创建时间戳操纵模拟器。
func NewTimestampManipulationSimulator() *TimestampManipulationSimulator {
	return &TimestampManipulationSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"timestamp_manipulation",
			"时间戳操纵攻击演示器",
			"演示矿工或出块者利用 block.timestamp 的可调节空间，绕过时间锁、开奖窗口或结算条件。",
			"attacks",
			types.ComponentAttack,
		),
		records: make([]map[string]interface{}, 0),
	}
}

// Init 初始化模拟器。
func (s *TimestampManipulationSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.records = make([]map[string]interface{}, 0)
	s.updateState()
	return nil
}

// ShowVulnerablePatterns 返回典型不安全写法。
func (s *TimestampManipulationSimulator) ShowVulnerablePatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"pattern": "直接依赖 block.timestamp 解锁",
			"risk":    "出块者可以在允许范围内微调时间，从而提前或延后触发逻辑。",
			"code":    `require(block.timestamp >= unlockTime, "locked");`,
		},
		{
			"pattern": "用时间戳控制开奖",
			"risk":    "开奖结果可能被出块者在打包时人为偏向。",
			"code":    `uint lucky = uint(block.timestamp) % 100;`,
		},
		{
			"pattern": "结算窗口过窄",
			"risk":    "边界条件下几秒钟的偏移就可能让交易从失败变成成功。",
			"code":    `require(block.timestamp <= deadline, "expired");`,
		},
	}
}

// ShowDefenses 返回推荐防御策略。
func (s *TimestampManipulationSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":           "使用更宽的时间缓冲",
			"recommendation": "不要让协议对几秒钟的偏移过于敏感。",
		},
		{
			"name":        "改用区块高度或多条件判断",
			"description": "对严格顺序要求的逻辑，可结合 block.number 与额外状态做联合判断。",
		},
		{
			"name":        "高价值场景引入预言机或外部确认",
			"description": "开奖、结算等关键流程不要完全依赖单一区块时间戳。",
		},
	}
}

// SimulateTimelockBypass 演示通过轻微调整时间戳绕过时间锁。
func (s *TimestampManipulationSimulator) SimulateTimelockBypass(unlockTime int64) map[string]interface{} {
	manipulatedTime := unlockTime + 10
	data := map[string]interface{}{
		"scenario":         "时间锁绕过",
		"unlock_time":      unlockTime,
		"manipulated_time": manipulatedTime,
		"attack":           fmt.Sprintf("出块者将区块时间调高到 %d，使本应稍后才能执行的操作提前生效。", manipulatedTime),
		"summary":          "当业务逻辑只依赖 block.timestamp 且窗口过窄时，轻微操纵就足以改变结果。",
		"flow": []string{
			"协议检查当前区块时间是否达到 unlockTime。",
			"出块者在允许范围内抬高时间戳。",
			"时间锁判断被提前满足，本应等待的操作提前执行。",
		},
	}
	s.records = append(s.records, data)
	s.EmitEvent("timelock_bypass", "", "", data)
	s.updateState()
	return data
}

// GetRealWorldCases 返回真实案例。
func (s *TimestampManipulationSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "早期链上彩票与小游戏",
			"impact": "很多合约直接使用时间戳生成随机数，导致开奖结果容易被偏置。",
		},
		{
			"name":   "时间锁执行边界问题",
			"impact": "对执行时机要求过于严格的合约，容易因时间戳微调而提前或延后触发。",
		},
	}
}

// updateState 同步状态。
func (s *TimestampManipulationSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.records))

	if len(s.records) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发时间戳操纵场景，请观察出块者微调时间后如何改变时间锁结果。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待时间戳操纵场景。",
			"可以先触发一次时间锁绕过，观察微调时间戳是怎样改变结果的。",
			0,
			map[string]interface{}{
				"attack_count": len(s.records),
			},
		)
		return
	}

	latest := s.records[len(s.records)-1]
	flow, _ := latest["flow"].([]string)
	steps := make([]map[string]interface{}, 0, len(flow))
	for index, item := range flow {
		steps = append(steps, map[string]interface{}{
			"index":       index + 1,
			"title":       fmt.Sprintf("阶段 %d", index+1),
			"description": item,
		})
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest["summary"])
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"shifted",
		fmt.Sprintf("%v", latest["summary"]),
		"重点观察时间锁阈值与被操纵时间戳的差距有多小，以及它为什么足以改变执行结果。",
		1.0,
		map[string]interface{}{
			"scenario":         latest["scenario"],
			"unlock_time":      latest["unlock_time"],
			"manipulated_time": latest["manipulated_time"],
			"step_count":       len(steps),
		},
	)
}

// ExecuteAction 执行时间戳相关演示动作。
func (s *TimestampManipulationSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_timelock_bypass":
		unlockTime := actionInt64(params, "unlock_time", time.Now().Unix()+3600)
		result := s.SimulateTimelockBypass(unlockTime)
		return actionResultWithFeedback(
			"已执行时间锁绕过演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入通过微调时间戳来提前满足时间锁条件的攻击流程。",
				NextHint:    "重点观察时间锁窗口是否过窄，以及时间戳与业务条件绑定得是否过紧。",
				EffectScope: "execution",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// TimestampManipulationFactory 创建时间戳操纵模拟器。
type TimestampManipulationFactory struct{}

func (f *TimestampManipulationFactory) Create() engine.Simulator { return NewTimestampManipulationSimulator() }
func (f *TimestampManipulationFactory) GetDescription() types.Description {
	return NewTimestampManipulationSimulator().GetDescription()
}
func NewTimestampManipulationFactory() *TimestampManipulationFactory {
	return &TimestampManipulationFactory{}
}

var _ engine.SimulatorFactory = (*TimestampManipulationFactory)(nil)
