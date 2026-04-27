package evm

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 委托调用演示器
// =============================================================================

// CallContext 调用上下文
type CallContext struct {
	MsgSender   string `json:"msg_sender"`
	MsgValue    string `json:"msg_value"`
	Address     string `json:"address"`      // 执行上下文地址
	CodeAddress string `json:"code_address"` // 代码来源地址
	Storage     string `json:"storage"`      // 存储归属
}

// DelegatecallSimulator 委托调用演示器
// 对比CALL、DELEGATECALL、CALLCODE、STATICCALL的区别:
//
// CALL:
//   - msg.sender = 调用者合约
//   - 执行目标合约代码
//   - 修改目标合约存储
//
// DELEGATECALL:
//   - msg.sender = 原始调用者(保持不变)
//   - 执行目标合约代码
//   - 修改调用者合约存储
//
// CALLCODE (已弃用):
//   - msg.sender = 调用者合约
//   - 执行目标合约代码
//   - 修改调用者合约存储
//
// STATICCALL:
//   - 只读调用，不能修改状态
type DelegatecallSimulator struct {
	*base.BaseSimulator
	scenarios []map[string]interface{}
}

// NewDelegatecallSimulator 创建委托调用演示器
func NewDelegatecallSimulator() *DelegatecallSimulator {
	sim := &DelegatecallSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"delegatecall",
			"委托调用演示器",
			"对比CALL、DELEGATECALL、STATICCALL的区别",
			"evm",
			types.ComponentDemo,
		),
		scenarios: make([]map[string]interface{}, 0),
	}

	return sim
}

// Init 初始化
func (s *DelegatecallSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// CompareCallTypes 对比不同调用类型
func (s *DelegatecallSimulator) CompareCallTypes() []map[string]interface{} {
	user := "0xUser"
	contractA := "0xContractA"
	contractB := "0xContractB"

	comparisons := []map[string]interface{}{
		{
			"type":        "CALL",
			"description": "普通调用",
			"scenario":    fmt.Sprintf("%s 调用 %s，%s 再CALL %s", user, contractA, contractA, contractB),
			"context_in_b": CallContext{
				MsgSender:   contractA,
				MsgValue:    "可传递",
				Address:     contractB,
				CodeAddress: contractB,
				Storage:     contractB,
			},
			"use_case": "调用其他合约的函数",
		},
		{
			"type":        "DELEGATECALL",
			"description": "委托调用",
			"scenario":    fmt.Sprintf("%s 调用 %s，%s 再DELEGATECALL %s", user, contractA, contractA, contractB),
			"context_in_b": CallContext{
				MsgSender:   user, // 保持原始调用者!
				MsgValue:    "保持原值",
				Address:     contractA, // 执行环境是A!
				CodeAddress: contractB,
				Storage:     contractA, // 修改A的存储!
			},
			"use_case": "代理模式、库函数调用",
		},
		{
			"type":        "STATICCALL",
			"description": "静态调用",
			"scenario":    fmt.Sprintf("%s 调用 %s，%s 再STATICCALL %s", user, contractA, contractA, contractB),
			"context_in_b": CallContext{
				MsgSender:   contractA,
				MsgValue:    "0 (不能传ETH)",
				Address:     contractB,
				CodeAddress: contractB,
				Storage:     "只读，不能修改",
			},
			"use_case": "安全地读取其他合约状态",
		},
		{
			"type":        "CALLCODE (已弃用)",
			"description": "代码调用",
			"scenario":    fmt.Sprintf("%s 调用 %s，%s 再CALLCODE %s", user, contractA, contractA, contractB),
			"context_in_b": CallContext{
				MsgSender:   contractA, // 与DELEGATECALL不同!
				MsgValue:    "可传递",
				Address:     contractA,
				CodeAddress: contractB,
				Storage:     contractA,
			},
			"use_case": "已被DELEGATECALL取代",
		},
	}

	s.scenarios = comparisons

	s.EmitEvent("call_types_compared", "", "", map[string]interface{}{
		"types": []string{"CALL", "DELEGATECALL", "STATICCALL", "CALLCODE"},
	})

	s.updateState()
	return comparisons
}

// SimulateProxyPattern 模拟代理模式
func (s *DelegatecallSimulator) SimulateProxyPattern() map[string]interface{} {
	result := map[string]interface{}{
		"pattern": "Transparent Proxy Pattern",
		"components": map[string]interface{}{
			"proxy": map[string]interface{}{
				"address":         "0xProxy",
				"role":            "存储状态，转发调用",
				"storage_slot_0":  "implementation address",
				"storage_slot_1+": "业务数据",
			},
			"implementation": map[string]interface{}{
				"address": "0xImplementation",
				"role":    "提供逻辑代码",
				"storage": "不使用自己的存储",
			},
			"admin": map[string]interface{}{
				"address": "0xAdmin",
				"role":    "可升级implementation",
			},
		},
		"flow": []string{
			"1. 用户调用Proxy合约",
			"2. Proxy的fallback函数被触发",
			"3. Proxy通过DELEGATECALL调用Implementation",
			"4. Implementation的代码在Proxy的上下文中执行",
			"5. 所有状态修改都发生在Proxy的存储中",
		},
		"code_example": `// Proxy合约
fallback() external payable {
    address impl = implementation;
    assembly {
        calldatacopy(0, 0, calldatasize())
        let result := delegatecall(gas(), impl, 0, calldatasize(), 0, 0)
        returndatacopy(0, 0, returndatasize())
        switch result
        case 0 { revert(0, returndatasize()) }
        default { return(0, returndatasize()) }
    }
}`,
	}

	s.EmitEvent("proxy_pattern_simulated", "", "", map[string]interface{}{
		"pattern": "Transparent Proxy",
	})

	return result
}

// SimulateStorageCollision 模拟存储冲突
func (s *DelegatecallSimulator) SimulateStorageCollision() map[string]interface{} {
	return map[string]interface{}{
		"issue":       "存储冲突 (Storage Collision)",
		"description": "当Proxy和Implementation使用相同的存储槽时发生冲突",
		"example": map[string]interface{}{
			"proxy_storage": map[string]string{
				"slot_0": "owner (地址)",
				"slot_1": "implementation (地址)",
			},
			"impl_storage": map[string]string{
				"slot_0": "value (uint256)", // 冲突!
				"slot_1": "counter (uint256)",
			},
		},
		"problem": "Implementation写入slot_0会覆盖Proxy的owner!",
		"solutions": []map[string]string{
			{
				"name":        "EIP-1967",
				"description": "使用随机存储槽",
				"slot":        "bytes32(uint256(keccak256('eip1967.proxy.implementation')) - 1)",
			},
			{
				"name":        "Unstructured Storage",
				"description": "在任意位置存储管理变量",
			},
			{
				"name":        "Diamond Storage",
				"description": "每个facet使用唯一的存储结构",
			},
		},
	}
}

// GetSecurityConsiderations 获取安全注意事项
func (s *DelegatecallSimulator) GetSecurityConsiderations() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"risk":        "存储冲突",
			"description": "Proxy和Implementation存储布局必须兼容",
			"mitigation":  "使用EIP-1967标准存储槽",
		},
		{
			"risk":        "函数选择器冲突",
			"description": "Proxy管理函数可能与业务函数冲突",
			"mitigation":  "使用Transparent Proxy或UUPS模式",
		},
		{
			"risk":        "未初始化的Implementation",
			"description": "Implementation可能被攻击者初始化",
			"mitigation":  "在Implementation构造函数中禁用初始化",
		},
		{
			"risk":        "selfdestruct",
			"description": "Implementation中的selfdestruct会销毁Proxy",
			"mitigation":  "永远不要在implementation中使用selfdestruct",
		},
		{
			"risk":        "不安全的delegatecall目标",
			"description": "delegatecall到用户控制的地址",
			"mitigation":  "严格验证delegatecall目标地址",
		},
	}
}

// updateState 更新状态
func (s *DelegatecallSimulator) updateState() {
	s.SetGlobalData("scenario_count", len(s.scenarios))

	stage := "pattern_ready"
	summary := "当前还没有加载调用类型对比，可以先生成 CALL、DELEGATECALL 和 STATICCALL 的上下文差异。"
	nextHint := "重点观察 msg.sender、执行上下文地址和存储归属在不同调用方式下如何变化。"
	progress := 0.2
	result := map[string]interface{}{"scenario_count": len(s.scenarios)}
	if len(s.scenarios) > 0 {
		stage = "comparison_completed"
		summary = fmt.Sprintf("最近一次已整理出 %d 种调用方式的上下文差异。", len(s.scenarios))
		nextHint = "继续查看代理模式和存储冲突案例，理解为什么 DELEGATECALL 容易引入高风险。"
		progress = 0.85
	}
	setEVMTeachingState(s.BaseSimulator, "evm", stage, summary, nextHint, progress, result)
}

// ExecuteAction 为委托调用实验提供交互动作。
func (s *DelegatecallSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "compare_call_types":
		result := s.CompareCallTypes()
		return evmActionResult(
			"已生成不同调用类型的对比结果。",
			map[string]interface{}{"scenario_count": len(result), "comparisons": result},
			&types.ActionFeedback{
				Summary:     "不同调用方式的执行上下文差异已经整理完成。",
				NextHint:    "重点观察 msg.sender 和 storage 归属如何变化，这正是代理和攻击风险的来源。",
				EffectScope: "evm",
				ResultState: map[string]interface{}{"scenario_count": len(result)},
			},
		), nil
	case "simulate_proxy_pattern":
		result := s.SimulateProxyPattern()
		return evmActionResult(
			"已模拟代理模式下的委托调用链路。",
			result,
			&types.ActionFeedback{
				Summary:     "代理模式中的调用转发和存储归属关系已经生成。",
				NextHint:    "继续对比实现合约代码与代理合约存储之间的关系，观察升级逻辑如何生效。",
				EffectScope: "evm",
				ResultState: result,
			},
		), nil
	case "simulate_storage_collision":
		result := s.SimulateStorageCollision()
		return evmActionResult(
			"已生成存储冲突示例。",
			result,
			&types.ActionFeedback{
				Summary:     "当前示例已经展示了代理与实现合约之间的存储槽冲突风险。",
				NextHint:    "重点观察 slot 对齐关系，以及 EIP-1967 如何避免关键管理槽被覆盖。",
				EffectScope: "evm",
				ResultState: result,
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported delegatecall action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// DelegatecallFactory 委托调用工厂
type DelegatecallFactory struct{}

func (f *DelegatecallFactory) Create() engine.Simulator {
	return NewDelegatecallSimulator()
}

func (f *DelegatecallFactory) GetDescription() types.Description {
	return NewDelegatecallSimulator().GetDescription()
}

func NewDelegatecallFactory() *DelegatecallFactory {
	return &DelegatecallFactory{}
}

var _ engine.SimulatorFactory = (*DelegatecallFactory)(nil)
