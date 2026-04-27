package evm

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	evmpkg "github.com/chainspace/simulations/pkg/evm"
	"github.com/chainspace/simulations/pkg/types"
)

type CallType string

const (
	CallTypeCall         CallType = "CALL"
	CallTypeDelegateCall CallType = "DELEGATECALL"
	CallTypeStaticCall   CallType = "STATICCALL"
	CallTypeCallCode     CallType = "CALLCODE"
	CallTypeCreate       CallType = "CREATE"
	CallTypeCreate2      CallType = "CREATE2"
)

type CallFrame struct {
	Type           CallType                 `json:"type"`
	From           string                   `json:"from"`
	To             string                   `json:"to"`
	Value          string                   `json:"value"`
	Gas            uint64                   `json:"gas"`
	GasUsed        uint64                   `json:"gas_used"`
	Input          string                   `json:"input"`
	Output         string                   `json:"output"`
	Error          string                   `json:"error,omitempty"`
	Depth          int                      `json:"depth"`
	Calls          []*CallFrame             `json:"calls,omitempty"`
	Logs           []map[string]interface{} `json:"logs,omitempty"`
	StorageChanges map[string]string        `json:"storage_changes,omitempty"`
}

type CallTraceSimulator struct {
	*base.BaseSimulator
	rootFrame    *CallFrame
	currentDepth int
	maxDepth     int
}

func NewCallTraceSimulator() *CallTraceSimulator {
	return &CallTraceSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"call_trace",
			"调用跟踪演示器",
			"可视化合约调用的完整调用栈、返回值、日志和状态变化。",
			"evm",
			types.ComponentDemo,
		),
	}
}

func (s *CallTraceSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

func (s *CallTraceSimulator) SimulateCallChain(scenario string) *CallFrame {
	s.currentDepth = 0
	s.maxDepth = 0
	s.SetGlobalData("scenario", scenario)

	switch scenario {
	case "simple_transfer":
		s.rootFrame = s.simulateSimpleTransfer()
	case "token_transfer":
		s.rootFrame = s.simulateTokenTransfer()
	case "nested_calls":
		s.rootFrame = s.simulateNestedCalls()
	default:
		s.rootFrame = s.simulateDeFiSwap()
	}

	s.EmitEvent("trace_complete", "", "", map[string]interface{}{
		"scenario":  scenario,
		"max_depth": s.maxDepth,
	})

	s.updateState()
	return s.rootFrame
}

func (s *CallTraceSimulator) simulateSimpleTransfer() *CallFrame {
	return &CallFrame{
		Type:    CallTypeCall,
		From:    "0xaaaa...aaaa",
		To:      "0xbbbb...bbbb",
		Value:   "1000000000000000000",
		Gas:     21000,
		GasUsed: 21000,
		Input:   "0x",
		Output:  "0x",
		Depth:   0,
	}
}

func (s *CallTraceSimulator) simulateTokenTransfer() *CallFrame {
	encoder := evmpkg.NewABIEncoder()
	selector := encoder.FunctionSelectorHex("transfer(address,uint256)")
	return &CallFrame{
		Type:    CallTypeCall,
		From:    "0xaaaa...aaaa",
		To:      "0xtoken...contract",
		Value:   "0",
		Gas:     65000,
		GasUsed: 52000,
		Input:   selector + "000000000000000000000000bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0000000000000000000000000000000000000000000000000de0b6b3a7640000",
		Output:  "0x0000000000000000000000000000000000000000000000000000000000000001",
		Depth:   0,
		Logs: []map[string]interface{}{
			{"event": "Transfer(address,address,uint256)", "from": "0xaaaa...aaaa", "to": "0xbbbb...bbbb", "amount": "1000000000000000000"},
		},
		StorageChanges: map[string]string{
			"balances[0xaaaa]": "1000 -> 0",
			"balances[0xbbbb]": "0 -> 1000",
		},
	}
}

func (s *CallTraceSimulator) simulateDeFiSwap() *CallFrame {
	s.maxDepth = 3
	return &CallFrame{
		Type:    CallTypeCall,
		From:    "0xuser...addr",
		To:      "0xrouter...contract",
		Value:   "1000000000000000000",
		Gas:     300000,
		GasUsed: 180000,
		Input:   "0x7ff36ab5...",
		Output:  "0x...",
		Depth:   0,
		Calls: []*CallFrame{
			{
				Type:    CallTypeCall,
				From:    "0xrouter...contract",
				To:      "0xweth...contract",
				Value:   "1000000000000000000",
				Gas:     50000,
				GasUsed: 25000,
				Input:   "0xd0e30db0",
				Output:  "0x",
				Depth:   1,
				Logs:    []map[string]interface{}{{"event": "Deposit(address,uint256)"}},
			},
			{
				Type:    CallTypeCall,
				From:    "0xrouter...contract",
				To:      "0xpair...contract",
				Value:   "0",
				Gas:     150000,
				GasUsed: 120000,
				Input:   "0x022c0d9f...",
				Output:  "0x",
				Depth:   1,
				Calls: []*CallFrame{
					{
						Type:    CallTypeCall,
						From:    "0xpair...contract",
						To:      "0xtoken...contract",
						Value:   "0",
						Gas:     50000,
						GasUsed: 30000,
						Input:   "0xa9059cbb...",
						Output:  "0x01",
						Depth:   2,
						Logs:    []map[string]interface{}{{"event": "Transfer(address,address,uint256)"}},
					},
					{
						Type:    CallTypeStaticCall,
						From:    "0xpair...contract",
						To:      "0xtoken...contract",
						Value:   "0",
						Gas:     10000,
						GasUsed: 3000,
						Input:   "0x70a08231...",
						Output:  "0x...",
						Depth:   2,
					},
				},
				Logs: []map[string]interface{}{{"event": "Swap(address,uint256,uint256,uint256,uint256,address)"}},
			},
		},
	}
}

func (s *CallTraceSimulator) simulateNestedCalls() *CallFrame {
	s.maxDepth = 4
	return &CallFrame{
		Type:    CallTypeCall,
		From:    "0xuser",
		To:      "0xcontract_a",
		Value:   "0",
		Gas:     500000,
		GasUsed: 250000,
		Input:   "0xabcd1234",
		Depth:   0,
		Calls: []*CallFrame{
			{
				Type:    CallTypeDelegateCall,
				From:    "0xcontract_a",
				To:      "0ximpl_a",
				Gas:     400000,
				GasUsed: 100000,
				Input:   "0xabcd1234",
				Depth:   1,
				Calls: []*CallFrame{
					{
						Type:    CallTypeCall,
						From:    "0xcontract_a",
						To:      "0xcontract_b",
						Gas:     300000,
						GasUsed: 80000,
						Input:   "0xef567890",
						Depth:   2,
						Calls: []*CallFrame{
							{
								Type:    CallTypeStaticCall,
								From:    "0xcontract_b",
								To:      "0xoracle",
								Gas:     50000,
								GasUsed: 5000,
								Input:   "0x...",
								Output:  "0x...",
								Depth:   3,
							},
						},
					},
				},
			},
		},
	}
}

func (s *CallTraceSimulator) TraceToTree(frame *CallFrame, indent int) []string {
	if frame == nil {
		return nil
	}
	lines := make([]string, 0)
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}
	callDesc := fmt.Sprintf("%s%s %s -> %s", prefix, frame.Type, frame.From, frame.To)
	if frame.Value != "0" && frame.Value != "" {
		callDesc += fmt.Sprintf(" [%s wei]", frame.Value)
	}
	callDesc += fmt.Sprintf(" (gas: %d/%d)", frame.GasUsed, frame.Gas)
	if frame.Error != "" {
		callDesc += fmt.Sprintf(" ERROR: %s", frame.Error)
	}
	lines = append(lines, callDesc)
	for _, child := range frame.Calls {
		lines = append(lines, s.TraceToTree(child, indent+1)...)
	}
	return lines
}

func (s *CallTraceSimulator) GetTraceStatistics() map[string]interface{} {
	if s.rootFrame == nil {
		return nil
	}
	stats := map[string]interface{}{
		"total_gas":    uint64(0),
		"call_count":   0,
		"max_depth":    0,
		"call_types":   make(map[string]int),
		"failed_calls": 0,
	}
	s.collectStats(s.rootFrame, stats, 0)
	return stats
}

func (s *CallTraceSimulator) collectStats(frame *CallFrame, stats map[string]interface{}, depth int) {
	if frame == nil {
		return
	}
	stats["total_gas"] = stats["total_gas"].(uint64) + frame.GasUsed
	stats["call_count"] = stats["call_count"].(int) + 1
	if depth > stats["max_depth"].(int) {
		stats["max_depth"] = depth
	}
	stats["call_types"].(map[string]int)[string(frame.Type)]++
	if frame.Error != "" {
		stats["failed_calls"] = stats["failed_calls"].(int) + 1
	}
	for _, child := range frame.Calls {
		s.collectStats(child, stats, depth+1)
	}
}

func (s *CallTraceSimulator) updateState() {
	if s.rootFrame != nil {
		stats := s.GetTraceStatistics()
		s.SetGlobalData("max_depth", stats["max_depth"])
		s.SetGlobalData("call_count", stats["call_count"])
		s.SetGlobalData("total_gas", stats["total_gas"])
		s.SetGlobalData("frames", s.flattenTrace(s.rootFrame))
		s.SetGlobalData("trace_tree", s.TraceToTree(s.rootFrame, 0))
		setEVMTeachingState(
			s.BaseSimulator,
			"evm",
			"trace_completed",
			fmt.Sprintf("最近一次调用跟踪共包含 %d 个调用帧，最大深度为 %d。", stats["call_count"].(int), stats["max_depth"].(int)),
			"继续观察每一层调用的 gas、返回值和存储变化，判断执行路径是否符合预期。",
			0.85,
			map[string]interface{}{"call_count": stats["call_count"], "max_depth": stats["max_depth"], "total_gas": stats["total_gas"]},
		)
		return
	}
	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		"trace_ready",
		"当前还没有执行调用跟踪，可以选择一个场景生成完整调用链。",
		"推荐先从简单转账开始，再比较 DeFi 交换和嵌套调用的层级差异。",
		0.2,
		map[string]interface{}{"call_count": 0, "max_depth": 0, "total_gas": 0},
	)
}

func (s *CallTraceSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_trace":
		scenario := callTraceStringParam(params, "scenario", "defi_swap")
		trace := s.SimulateCallChain(scenario)
		return evmActionResult(
			"已生成一条完整调用跟踪。",
			map[string]interface{}{"scenario": scenario, "trace": trace},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("%s 场景的调用链已经生成。", scenario),
				NextHint:    "重点查看调用深度、子调用类型、gas 消耗和状态变化如何层层传递。",
				EffectScope: "evm",
				ResultState: map[string]interface{}{"scenario": scenario},
			},
		), nil
	case "reset_trace":
		s.rootFrame = nil
		s.currentDepth = 0
		s.maxDepth = 0
		s.SetGlobalData("scenario", "")
		s.SetGlobalData("frames", []map[string]interface{}{})
		s.SetGlobalData("trace_tree", []string{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("call_count", 0)
		s.SetGlobalData("total_gas", 0)
		s.updateState()
		return evmActionResult(
			"已重置调用跟踪场景。",
			nil,
			&types.ActionFeedback{
				Summary:     "调用跟踪数据已清空，执行上下文回到初始状态。",
				NextHint:    "可以重新选择一个场景，对比不同调用模式的执行差异。",
				EffectScope: "evm",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported call trace action: %s", action)
	}
}

func (s *CallTraceSimulator) flattenTrace(frame *CallFrame) []map[string]interface{} {
	if frame == nil {
		return []map[string]interface{}{}
	}
	frames := make([]map[string]interface{}, 0)
	var visit func(current *CallFrame)
	visit = func(current *CallFrame) {
		frames = append(frames, map[string]interface{}{
			"title":           string(current.Type),
			"description":     fmt.Sprintf("%s -> %s", current.From, current.To),
			"opcode":          string(current.Type),
			"depth":           current.Depth,
			"from":            current.From,
			"to":              current.To,
			"value":           current.Value,
			"gas":             fmt.Sprintf("%d / %d", current.GasUsed, current.Gas),
			"gas_used":        current.GasUsed,
			"gas_limit":       current.Gas,
			"input":           current.Input,
			"output":          current.Output,
			"error":           current.Error,
			"log_count":       len(current.Logs),
			"storage_changes": current.StorageChanges,
			"call_type":       string(current.Type),
		})
		for _, child := range current.Calls {
			visit(child)
		}
	}
	visit(frame)
	return frames
}

func callTraceStringParam(params map[string]interface{}, key, fallback string) string {
	if params == nil {
		return fallback
	}
	if value, ok := params[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

type CallTraceFactory struct{}

func (f *CallTraceFactory) Create() engine.Simulator {
	return NewCallTraceSimulator()
}

func (f *CallTraceFactory) GetDescription() types.Description {
	return NewCallTraceSimulator().GetDescription()
}

func NewCallTraceFactory() *CallTraceFactory {
	return &CallTraceFactory{}
}

var _ engine.SimulatorFactory = (*CallTraceFactory)(nil)
