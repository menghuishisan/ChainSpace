package attacks

import (
	"fmt"
	"math/big"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// OverflowExample 描述一次溢出或下溢计算示例。
type OverflowExample struct {
	Type         string `json:"type"`
	BitSize      int    `json:"bit_size"`
	Value1       string `json:"value1"`
	Value2       string `json:"value2"`
	Operation    string `json:"operation"`
	Expected     string `json:"expected"`
	Actual       string `json:"actual"`
	IsVulnerable bool   `json:"is_vulnerable"`
}

// IntegerOverflowSimulator 演示整数溢出与下溢。
type IntegerOverflowSimulator struct {
	*base.BaseSimulator
	examples []*OverflowExample
}

// NewIntegerOverflowSimulator 创建整数溢出模拟器。
func NewIntegerOverflowSimulator() *IntegerOverflowSimulator {
	return &IntegerOverflowSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"integer_overflow",
			"整数溢出攻击演示器",
			"演示 Solidity 早期版本中常见的加法溢出、减法下溢和乘法回绕问题。",
			"attacks",
			types.ComponentAttack,
		),
		examples: make([]*OverflowExample, 0),
	}
}

// Init 初始化模拟器。
func (s *IntegerOverflowSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.examples = make([]*OverflowExample, 0)
	s.updateState()
	return nil
}

// SimulateOverflow 演示加法溢出。
func (s *IntegerOverflowSimulator) SimulateOverflow(bitSize int) *OverflowExample {
	maxVal := new(big.Int).Lsh(big.NewInt(1), uint(bitSize))
	maxVal.Sub(maxVal, big.NewInt(1))
	expected := new(big.Int).Add(maxVal, big.NewInt(1))
	modulus := new(big.Int).Lsh(big.NewInt(1), uint(bitSize))
	actual := new(big.Int).Mod(expected, modulus)

	example := &OverflowExample{
		Type:         "overflow",
		BitSize:      bitSize,
		Value1:       maxVal.String(),
		Value2:       "1",
		Operation:    "+",
		Expected:     expected.String(),
		Actual:       actual.String(),
		IsVulnerable: true,
	}
	s.examples = append(s.examples, example)
	s.EmitEvent("overflow_simulated", "", "", map[string]interface{}{
		"bit_size": bitSize,
		"result":   actual.String(),
	})
	s.updateState()
	return example
}

// SimulateUnderflow 演示减法下溢。
func (s *IntegerOverflowSimulator) SimulateUnderflow(bitSize int) *OverflowExample {
	maxVal := new(big.Int).Lsh(big.NewInt(1), uint(bitSize))
	maxVal.Sub(maxVal, big.NewInt(1))

	example := &OverflowExample{
		Type:         "underflow",
		BitSize:      bitSize,
		Value1:       "0",
		Value2:       "1",
		Operation:    "-",
		Expected:     "-1",
		Actual:       maxVal.String(),
		IsVulnerable: true,
	}
	s.examples = append(s.examples, example)
	s.EmitEvent("underflow_simulated", "", "", map[string]interface{}{
		"bit_size": bitSize,
		"result":   maxVal.String(),
	})
	s.updateState()
	return example
}

// SimulateMulOverflow 演示乘法溢出。
func (s *IntegerOverflowSimulator) SimulateMulOverflow(bitSize int) *OverflowExample {
	maxVal := new(big.Int).Lsh(big.NewInt(1), uint(bitSize))
	maxVal.Sub(maxVal, big.NewInt(1))
	val1 := new(big.Int).Rsh(maxVal, 1)
	val2 := big.NewInt(3)
	expected := new(big.Int).Mul(val1, val2)
	modulus := new(big.Int).Lsh(big.NewInt(1), uint(bitSize))
	actual := new(big.Int).Mod(expected, modulus)

	example := &OverflowExample{
		Type:         "mul_overflow",
		BitSize:      bitSize,
		Value1:       val1.String(),
		Value2:       val2.String(),
		Operation:    "*",
		Expected:     expected.String(),
		Actual:       actual.String(),
		IsVulnerable: expected.Cmp(actual) != 0,
	}
	s.examples = append(s.examples, example)
	s.EmitEvent("mul_overflow_simulated", "", "", map[string]interface{}{
		"bit_size":      bitSize,
		"is_vulnerable": example.IsVulnerable,
	})
	s.updateState()
	return example
}

// ShowDefenses 返回防御建议。
func (s *IntegerOverflowSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "使用 Solidity 0.8.0+", "description": "编译器默认开启溢出检查。"},
		{"name": "保留显式边界检查", "description": "对关键资金逻辑增加 require 校验。"},
		{"name": "谨慎使用 unchecked", "description": "只有在明确知道不会出错且为性能考虑时才使用。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *IntegerOverflowSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "BEC Token", "date": "2018", "issue": "批量转账中的整数溢出导致可无限增发。"},
		{"name": "PoWHC", "date": "2018", "issue": "早期代币合约中的下溢问题被用于提走资金。"},
	}
}

// updateState 同步状态。
func (s *IntegerOverflowSimulator) updateState() {
	s.SetGlobalData("example_count", len(s.examples))
	if len(s.examples) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发整数溢出场景，可以从上溢、下溢或乘法回绕开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待整数溢出场景。",
			"可以先选择一种运算，观察边界值如何让链上计算结果发生回绕。",
			0,
			map[string]interface{}{
				"example_count": len(s.examples),
			},
		)
		return
	}

	latest := s.examples[len(s.examples)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "prepare_boundary_value",
			"caller":      "attacker",
			"function":    latest.Operation,
			"target":      "integer_boundary",
			"amount":      latest.Value1,
			"call_depth":  1,
			"description": "先把整数推到边界值附近，构造最容易触发回绕的输入。",
		},
		{
			"step":        2,
			"action":      latest.Type,
			"caller":      "attacker",
			"function":    latest.Operation,
			"target":      "arithmetic_result",
			"amount":      latest.Value2,
			"call_depth":  2,
			"description": "执行危险运算，观察结果是正常增长还是发生回绕。",
		},
		{
			"step":        3,
			"action":      "compare_expected_actual",
			"caller":      "simulator",
			"function":    "validate_result",
			"target":      latest.Actual,
			"amount":      latest.Expected,
			"call_depth":  3,
			"description": "对比理论结果与链上实际结果，确认是否出现整数上溢、下溢或乘法溢出。",
		},
	}

	s.SetGlobalData("latest_example", latest)
	s.SetGlobalData("attack_summary", fmt.Sprintf("期望结果 %s，实际结果 %s。", latest.Expected, latest.Actual))
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"wrapped",
		fmt.Sprintf("%s 运算出现了边界回绕。", latest.Type),
		"重点观察理论结果和实际结果为何出现偏差，以及回绕是否会影响资金或权限逻辑。",
		1.0,
		map[string]interface{}{
			"type":          latest.Type,
			"bit_size":      latest.BitSize,
			"expected":      latest.Expected,
			"actual":        latest.Actual,
			"is_vulnerable": latest.IsVulnerable,
		},
	)
}

// ExecuteAction 执行前端动作。
func (s *IntegerOverflowSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_overflow":
		bitSize := actionInt(params, "bit_size", 256)
		example := s.SimulateOverflow(bitSize)
		return actionResultWithFeedback(
			"已执行整数上溢演示。",
			map[string]interface{}{"example": example},
			&types.ActionFeedback{
				Summary:     "已进入把数值推到上界并触发回绕的攻击流程。",
				NextHint:    "重点观察链上实际结果为何从最大值再回到 0。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"bit_size": bitSize,
					"actual":   example.Actual,
				},
			},
		), nil
	case "simulate_underflow":
		bitSize := actionInt(params, "bit_size", 256)
		example := s.SimulateUnderflow(bitSize)
		return actionResultWithFeedback(
			"已执行整数下溢演示。",
			map[string]interface{}{"example": example},
			&types.ActionFeedback{
				Summary:     "已进入从 0 继续减法导致结果跳到最大值的攻击流程。",
				NextHint:    "重点观察负值为何在无检查时表现为极大的正整数。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"bit_size": bitSize,
					"actual":   example.Actual,
				},
			},
		), nil
	case "simulate_mul_overflow":
		bitSize := actionInt(params, "bit_size", 256)
		example := s.SimulateMulOverflow(bitSize)
		return actionResultWithFeedback(
			"已执行乘法溢出演示。",
			map[string]interface{}{"example": example},
			&types.ActionFeedback{
				Summary:     "已进入乘法结果超过位宽后发生回绕的攻击流程。",
				NextHint:    "重点观察乘法结果和实际存储值之间的偏差。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"bit_size": bitSize,
					"actual":   example.Actual,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported integer overflow action: %s", action)
	}
}

// IntegerOverflowFactory 创建模拟器。
type IntegerOverflowFactory struct{}

func (f *IntegerOverflowFactory) Create() engine.Simulator { return NewIntegerOverflowSimulator() }
func (f *IntegerOverflowFactory) GetDescription() types.Description {
	return NewIntegerOverflowSimulator().GetDescription()
}
func NewIntegerOverflowFactory() *IntegerOverflowFactory { return &IntegerOverflowFactory{} }

var _ engine.SimulatorFactory = (*IntegerOverflowFactory)(nil)
