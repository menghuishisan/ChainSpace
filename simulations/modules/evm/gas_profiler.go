package evm

import (
	"fmt"
	"sort"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	evmpkg "github.com/chainspace/simulations/pkg/evm"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// Gas分析器
// =============================================================================

// GasBreakdown Gas分解
type GasBreakdown struct {
	OpCode  string  `json:"opcode"`
	Count   int     `json:"count"`
	GasUsed uint64  `json:"gas_used"`
	Percent float64 `json:"percent"`
}

// GasProfilerSimulator Gas分析器
// 分析合约执行的Gas消耗:
// - 各操作码的Gas消耗
// - 热点操作识别
// - 优化建议
type GasProfilerSimulator struct {
	*base.BaseSimulator
	breakdown map[string]*GasBreakdown
	totalGas  uint64
}

// NewGasProfilerSimulator 创建Gas分析器
func NewGasProfilerSimulator() *GasProfilerSimulator {
	sim := &GasProfilerSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"gas_profiler",
			"Gas分析器",
			"分析EVM执行的Gas消耗分布，识别优化机会",
			"evm",
			types.ComponentTool,
		),
		breakdown: make(map[string]*GasBreakdown),
	}

	return sim
}

// Init 初始化
func (s *GasProfilerSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// AnalyzeTraces 分析执行跟踪
func (s *GasProfilerSimulator) AnalyzeTraces(traces []*evmpkg.ExecutionTrace) map[string]interface{} {
	s.breakdown = make(map[string]*GasBreakdown)
	s.totalGas = 0

	for _, trace := range traces {
		opcode := trace.OpCode
		gasCost := trace.GasCost

		if _, ok := s.breakdown[opcode]; !ok {
			s.breakdown[opcode] = &GasBreakdown{
				OpCode: opcode,
			}
		}
		s.breakdown[opcode].Count++
		s.breakdown[opcode].GasUsed += gasCost
		s.totalGas += gasCost
	}

	// 计算百分比
	for _, b := range s.breakdown {
		if s.totalGas > 0 {
			b.Percent = float64(b.GasUsed) / float64(s.totalGas) * 100
		}
	}

	result := map[string]interface{}{
		"total_gas":         s.totalGas,
		"opcode_count":      len(s.breakdown),
		"instruction_count": len(traces),
	}

	s.EmitEvent("traces_analyzed", "", "", result)
	s.updateState()
	return result
}

// GetTopGasConsumers 获取Gas消耗最高的操作码
func (s *GasProfilerSimulator) GetTopGasConsumers(n int) []*GasBreakdown {
	items := make([]*GasBreakdown, 0, len(s.breakdown))
	for _, b := range s.breakdown {
		items = append(items, b)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].GasUsed > items[j].GasUsed
	})

	if n > len(items) {
		n = len(items)
	}
	return items[:n]
}

// GetOptimizationSuggestions 获取优化建议
func (s *GasProfilerSimulator) GetOptimizationSuggestions() []map[string]interface{} {
	suggestions := make([]map[string]interface{}, 0)

	// 检查SSTORE使用
	if b, ok := s.breakdown["SSTORE"]; ok && b.Count > 5 {
		suggestions = append(suggestions, map[string]interface{}{
			"issue":            "频繁存储写入",
			"opcode":           "SSTORE",
			"count":            b.Count,
			"gas_used":         b.GasUsed,
			"suggestion":       "考虑使用内存变量累积结果后一次性写入存储",
			"potential_saving": "每次SSTORE 5000-20000 gas",
		})
	}

	// 检查SLOAD使用
	if b, ok := s.breakdown["SLOAD"]; ok && b.Count > 10 {
		suggestions = append(suggestions, map[string]interface{}{
			"issue":            "频繁存储读取",
			"opcode":           "SLOAD",
			"count":            b.Count,
			"gas_used":         b.GasUsed,
			"suggestion":       "考虑将常用存储值缓存到内存变量",
			"potential_saving": "每次SLOAD 100-2100 gas",
		})
	}

	// 检查CALL使用
	if b, ok := s.breakdown["CALL"]; ok && b.Count > 3 {
		suggestions = append(suggestions, map[string]interface{}{
			"issue":            "多次外部调用",
			"opcode":           "CALL",
			"count":            b.Count,
			"gas_used":         b.GasUsed,
			"suggestion":       "考虑批量处理减少调用次数",
			"potential_saving": "每次CALL 100+ gas (不含内部执行)",
		})
	}

	// 检查MSTORE使用
	if b, ok := s.breakdown["MSTORE"]; ok && b.Count > 50 {
		suggestions = append(suggestions, map[string]interface{}{
			"issue":            "大量内存操作",
			"opcode":           "MSTORE",
			"count":            b.Count,
			"gas_used":         b.GasUsed,
			"suggestion":       "检查是否有不必要的内存复制",
			"potential_saving": "视具体情况而定",
		})
	}

	return suggestions
}

// ShowGasCosts 显示常见操作Gas消耗
func (s *GasProfilerSimulator) ShowGasCosts() []map[string]interface{} {
	costs := []map[string]interface{}{
		{"operation": "ADD/SUB/AND/OR", "gas": 3, "category": "算术/逻辑"},
		{"operation": "MUL/DIV", "gas": 5, "category": "算术"},
		{"operation": "EXP", "gas": "10 + 50*bytes", "category": "算术"},
		{"operation": "KECCAK256", "gas": "30 + 6*words", "category": "哈希"},
		{"operation": "SLOAD (cold)", "gas": 2100, "category": "存储"},
		{"operation": "SLOAD (warm)", "gas": 100, "category": "存储"},
		{"operation": "SSTORE (zero->non)", "gas": 20000, "category": "存储"},
		{"operation": "SSTORE (non->non)", "gas": 5000, "category": "存储"},
		{"operation": "SSTORE (non->zero)", "gas": "5000 - 15000 refund", "category": "存储"},
		{"operation": "CALL", "gas": "100 + memory + value", "category": "调用"},
		{"operation": "CREATE", "gas": 32000, "category": "合约创建"},
		{"operation": "CREATE2", "gas": "32000 + 6*size", "category": "合约创建"},
		{"operation": "LOG0-LOG4", "gas": "375 + 375*topics + 8*bytes", "category": "日志"},
		{"operation": "BALANCE (cold)", "gas": 2600, "category": "账户"},
		{"operation": "BALANCE (warm)", "gas": 100, "category": "账户"},
		{"operation": "EXTCODESIZE (cold)", "gas": 2600, "category": "账户"},
	}

	s.EmitEvent("gas_costs_shown", "", "", map[string]interface{}{
		"count": len(costs),
	})

	return costs
}

// CalculateTransactionCost 计算交易成本
func (s *GasProfilerSimulator) CalculateTransactionCost(gasUsed uint64, gasPriceGwei float64) map[string]interface{} {
	gasPriceWei := gasPriceGwei * 1e9
	costWei := float64(gasUsed) * gasPriceWei
	costEth := costWei / 1e18

	// 假设ETH价格
	ethPrices := map[string]float64{
		"USD": 2000,
		"CNY": 14000,
	}

	result := map[string]interface{}{
		"gas_used":       gasUsed,
		"gas_price_gwei": gasPriceGwei,
		"cost_eth":       fmt.Sprintf("%.6f", costEth),
		"cost_usd":       fmt.Sprintf("%.2f", costEth*ethPrices["USD"]),
		"cost_cny":       fmt.Sprintf("%.2f", costEth*ethPrices["CNY"]),
	}

	s.EmitEvent("cost_calculated", "", "", result)
	return result
}

// updateState 更新状态
func (s *GasProfilerSimulator) updateState() {
	s.SetGlobalData("total_gas", s.totalGas)
	s.SetGlobalData("opcode_count", len(s.breakdown))

	top := s.GetTopGasConsumers(5)
	topList := make([]map[string]interface{}, len(top))
	for i, b := range top {
		topList[i] = map[string]interface{}{
			"opcode":  b.OpCode,
			"gas":     b.GasUsed,
			"percent": fmt.Sprintf("%.1f%%", b.Percent),
		}
	}
	s.SetGlobalData("top_consumers", topList)

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		func() string {
			if len(s.breakdown) > 0 {
				return "gas_profile_completed"
			}
			return "gas_profile_ready"
		}(),
		func() string {
			if len(s.breakdown) > 0 {
				return fmt.Sprintf("当前已经分析出 %d 类主要 Gas 消耗来源。", len(s.breakdown))
			}
			return "当前还没有生成 Gas 分析，可以先查看常见操作码成本或估算一笔交易费用。"
		}(),
		"重点观察最耗 Gas 的操作码和可优化点，理解一次交易为什么昂贵。",
		func() float64 {
			if len(s.breakdown) > 0 {
				return 0.85
			}
			return 0.2
		}(),
		map[string]interface{}{"total_gas": s.totalGas, "opcode_count": len(s.breakdown)},
	)
}

// ExecuteAction 为 Gas 分析实验提供交互动作。
func (s *GasProfilerSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "calculate_tx_cost":
		result := s.CalculateTransactionCost(21000, 20)
		return evmActionResult("已计算一笔交易的 Gas 成本。", result, &types.ActionFeedback{
			Summary:     "交易成本已经按 Gas 用量和 Gas Price 计算完成。",
			NextHint:    "继续对比不同 Gas Price 或不同 opcode 路径下的总成本差异。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	case "show_gas_costs":
		result := s.ShowGasCosts()
		return evmActionResult("已加载常见操作码的 Gas 成本表。", map[string]interface{}{"costs": result}, &types.ActionFeedback{
			Summary:     "常见 opcode 的 Gas 成本已经整理完成。",
			NextHint:    "继续观察哪些操作最昂贵，以及为什么存储写入会远高于内存操作。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"count": len(result)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported gas profiler action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// GasProfilerFactory Gas分析器工厂
type GasProfilerFactory struct{}

func (f *GasProfilerFactory) Create() engine.Simulator {
	return NewGasProfilerSimulator()
}

func (f *GasProfilerFactory) GetDescription() types.Description {
	return NewGasProfilerSimulator().GetDescription()
}

func NewGasProfilerFactory() *GasProfilerFactory {
	return &GasProfilerFactory{}
}

var _ engine.SimulatorFactory = (*GasProfilerFactory)(nil)
