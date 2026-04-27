package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// DOSAttackType 表示 DoS 攻击类型。
type DOSAttackType string

const (
	DOSTypeGasLimit      DOSAttackType = "gas_limit"
	DOSTypeUnboundedLoop DOSAttackType = "unbounded_loop"
	DOSTypeExternalCall  DOSAttackType = "external_call"
	DOSTypeBlockStuffing DOSAttackType = "block_stuffing"
	DOSTypeRevert        DOSAttackType = "revert"
)

// DOSAttack 描述一次 DoS 攻击。
type DOSAttack struct {
	ID        string        `json:"id"`
	Type      DOSAttackType `json:"type"`
	Target    string        `json:"target"`
	Method    string        `json:"method"`
	GasUsed   uint64        `json:"gas_used"`
	Success   bool          `json:"success"`
	Impact    string        `json:"impact"`
	Timestamp time.Time     `json:"timestamp"`
}

// DOSSimulator 演示合约级拒绝服务攻击。
type DOSSimulator struct {
	*base.BaseSimulator
	attacks []*DOSAttack
}

// NewDOSSimulator 创建 DoS 攻击演示器。
func NewDOSSimulator() *DOSSimulator {
	return &DOSSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"dos",
			"拒绝服务攻击演示器",
			"演示无界循环、外部调用阻塞和区块填充等常见 DoS 攻击。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*DOSAttack, 0),
	}
}

// Init 初始化模拟器。
func (s *DOSSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.attacks = make([]*DOSAttack, 0)
	s.updateState()
	return nil
}

// SimulateUnboundedLoop 模拟无界循环攻击。
func (s *DOSSimulator) SimulateUnboundedLoop(arraySize int) *DOSAttack {
	gasPerIteration := uint64(5000)
	totalGas := uint64(arraySize) * gasPerIteration
	blockGasLimit := uint64(30000000)

	attack := &DOSAttack{
		ID:        fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:      DOSTypeUnboundedLoop,
		Target:    "0xVulnerableContract",
		Method:    fmt.Sprintf("遍历 %d 个元素的动态数组", arraySize),
		GasUsed:   totalGas,
		Success:   totalGas > blockGasLimit,
		Impact:    "随着列表持续增长，函数最终无法在单个区块内完成。",
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("unbounded_loop_attack", "", "", map[string]interface{}{
		"array_size":  arraySize,
		"gas_needed":  totalGas,
		"block_limit": blockGasLimit,
		"summary":     attack.Impact,
	})
	s.updateState()
	return attack
}

// SimulateExternalCallDOS 模拟外部调用阻塞。
func (s *DOSSimulator) SimulateExternalCallDOS() *DOSAttack {
	attack := &DOSAttack{
		ID:        fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:      DOSTypeExternalCall,
		Target:    "0xAuctionContract",
		Method:    "恶意合约在 receive 中主动 revert",
		GasUsed:   0,
		Success:   true,
		Impact:    "一旦攻击者成为关键流程中的付款接收方，系统可能永久卡死。",
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("external_call_dos", "", "", map[string]interface{}{
		"target":  attack.Target,
		"method":  attack.Method,
		"summary": attack.Impact,
	})
	s.updateState()
	return attack
}

// SimulateBlockStuffing 模拟区块填充攻击。
func (s *DOSSimulator) SimulateBlockStuffing(targetBlocks int) *DOSAttack {
	attack := &DOSAttack{
		ID:        fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:      DOSTypeBlockStuffing,
		Target:    "network",
		Method:    fmt.Sprintf("用高费用交易填满 %d 个区块", targetBlocks),
		GasUsed:   uint64(targetBlocks) * 30000000,
		Success:   true,
		Impact:    "目标交易在高费用竞争期间难以进入区块，关键操作被延迟或错过窗口。",
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("block_stuffing", "", "", map[string]interface{}{
		"target_blocks": targetBlocks,
		"summary":       attack.Impact,
	})
	s.updateState()
	return attack
}

// ShowVulnerablePatterns 展示易受攻击模式。
func (s *DOSSimulator) ShowVulnerablePatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"pattern": "无界循环",
			"vulnerable": `for (uint i = 0; i < array.length; i++) {
    process(array[i]);
}`,
			"fixed": `function processRange(uint start, uint count) external {
    uint end = min(start + count, array.length);
    for (uint i = start; i < end; i++) {
        process(array[i]);
    }
}`,
		},
		{
			"pattern":    "依赖外部调用成功",
			"vulnerable": `payable(user).transfer(amount);`,
			"fixed":      `(bool success, ) = user.call{value: amount}("");`,
		},
	}
}

// ShowDefenses 展示防御方式。
func (s *DOSSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Pull over Push", "description": "让用户主动提取资产，避免批量外部转账阻塞主流程。"},
		{"name": "分页处理", "description": "把可能无限增长的列表操作拆成多次执行。"},
		{"name": "失败隔离", "description": "单个外部调用失败不应导致全局状态回滚。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *DOSSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "GovernMental", "date": "2016", "issue": "无界循环导致提款逻辑无法完成"},
		{"name": "King of the Ether", "date": "2016", "issue": "外部退款失败导致竞价流程被锁死"},
	}
}

func (s *DOSSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))

	if len(s.attacks) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发 DoS 攻击，请观察 gas 膨胀、外部调用阻塞或区块填充如何中断主流程。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待 DoS 攻击场景。",
			"可以先触发一种阻塞路径，观察合约主流程在哪一步失去可用性。",
			0,
			map[string]interface{}{
				"attack_count": 0,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{"index": 1, "title": "识别瓶颈", "description": fmt.Sprintf("攻击者锁定目标 %s 的关键执行路径。", latest.Target)},
		{"index": 2, "title": "放大阻塞", "description": latest.Method},
		{"index": 3, "title": "主流程失效", "description": latest.Impact},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest.Impact)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"blocked",
		latest.Impact,
		"重点观察阻塞发生在哪条关键路径，以及系统为何无法继续向前推进。",
		1.0,
		map[string]interface{}{
			"attack_type": latest.Type,
			"target":      latest.Target,
			"gas_used":    latest.GasUsed,
			"success":     latest.Success,
		},
	)
}

// ExecuteAction 执行动作。
func (s *DOSSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_unbounded_loop":
		arraySize := actionInt(params, "array_size", 8000)
		attack := s.SimulateUnboundedLoop(arraySize)
		return actionResultWithFeedback(
			"已完成无界循环 DoS 演示。",
			map[string]interface{}{
				"attack_id":  attack.ID,
				"array_size": arraySize,
				"success":    attack.Success,
			},
			&types.ActionFeedback{
				Summary:     "已进入 gas 持续膨胀导致主函数无法完成的阻塞流程。",
				NextHint:    "重点观察消耗的 gas 是否超过区块上限，以及主流程在哪一步停止推进。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"gas_used": attack.GasUsed,
					"success":  attack.Success,
				},
			},
		), nil
	case "simulate_external_call_dos":
		attack := s.SimulateExternalCallDOS()
		return actionResultWithFeedback(
			"已完成外部调用阻塞 DoS 演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
				"success":   attack.Success,
			},
			&types.ActionFeedback{
				Summary:     "已进入恶意接收方通过 revert 卡死主流程的攻击路径。",
				NextHint:    "重点观察关键外部调用失败后，为什么主流程整体无法继续完成。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"target":  attack.Target,
					"success": attack.Success,
				},
			},
		), nil
	case "simulate_block_stuffing":
		targetBlocks := actionInt(params, "target_blocks", 3)
		attack := s.SimulateBlockStuffing(targetBlocks)
		return actionResultWithFeedback(
			"已完成区块填充 DoS 演示。",
			map[string]interface{}{
				"attack_id":     attack.ID,
				"target_blocks": targetBlocks,
				"success":       attack.Success,
			},
			&types.ActionFeedback{
				Summary:     "已进入高费用抢占区块空间的阻塞流程。",
				NextHint:    "重点观察目标交易如何在费用竞争中被持续挤出区块。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"target_blocks": targetBlocks,
					"gas_used":      attack.GasUsed,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// DOSFactory 创建模拟器工厂。
type DOSFactory struct{}

// Create 创建模拟器。
func (f *DOSFactory) Create() engine.Simulator { return NewDOSSimulator() }

// GetDescription 获取描述。
func (f *DOSFactory) GetDescription() types.Description { return NewDOSSimulator().GetDescription() }

// NewDOSFactory 创建工厂。
func NewDOSFactory() *DOSFactory { return &DOSFactory{} }

var _ engine.SimulatorFactory = (*DOSFactory)(nil)
