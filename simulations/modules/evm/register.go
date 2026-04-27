package evm

import "github.com/chainspace/simulations/pkg/engine"

// RegisterAll 注册所有EVM模块 (11个)
func RegisterAll(registry *engine.Registry) {
	// 核心执行 (4个)
	registry.Register("evm_executor", NewEVMExecutorFactory())  // EVM执行器
	registry.Register("disassembler", NewDisassemblerFactory()) // 反汇编器
	registry.Register("abi_codec", NewABICodecFactory())        // ABI编解码
	registry.Register("gas_profiler", NewGasProfilerFactory())  // Gas分析器

	// 调用与存储 (4个)
	registry.Register("call_trace", NewCallTraceFactory())         // 调用跟踪
	registry.Register("delegatecall", NewDelegatecallFactory())    // 委托调用
	registry.Register("storage_layout", NewStorageLayoutFactory()) // 存储布局
	registry.Register("state_diff", NewStateDiffFactory())         // 状态差异

	// 合约模式 (3个)
	registry.Register("create_create2", NewCreateCreate2Factory()) // 合约创建
	registry.Register("proxy_pattern", NewProxyPatternFactory())   // 代理模式
	registry.Register("event_log", NewEventLogFactory())           // 事件日志
}

// RegisterToEngine 注册到引擎 (11个)
func RegisterToEngine(eng *engine.Engine) {
	// 核心执行 (4个)
	eng.Register("evm_executor", NewEVMExecutorFactory())
	eng.Register("disassembler", NewDisassemblerFactory())
	eng.Register("abi_codec", NewABICodecFactory())
	eng.Register("gas_profiler", NewGasProfilerFactory())

	// 调用与存储 (4个)
	eng.Register("call_trace", NewCallTraceFactory())
	eng.Register("delegatecall", NewDelegatecallFactory())
	eng.Register("storage_layout", NewStorageLayoutFactory())
	eng.Register("state_diff", NewStateDiffFactory())

	// 合约模式 (3个)
	eng.Register("create_create2", NewCreateCreate2Factory())
	eng.Register("proxy_pattern", NewProxyPatternFactory())
	eng.Register("event_log", NewEventLogFactory())
}
