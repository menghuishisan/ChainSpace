package crosschain

import (
	"github.com/chainspace/simulations/pkg/engine"
)

// =============================================================================
// 跨链与L2模块注册
// 本模块包含12个跨链/L2相关的演示器:
//
// 跨链核心:
// 1. bridge           - 跨链桥演示器: 锁定-铸造、流动性桥等
// 2. atomic_swap      - 原子交换演示器: HTLC无信任交换
// 3. relay            - 中继链演示器: 跨链消息传递
// 4. light_client     - 轻客户端演示器: 区块头验证
// 5. ibc              - IBC协议演示器: Cosmos跨链通信
// 6. oracle_bridge    - 预言机桥演示器: 预言机网络验证
//
// Layer2扩容:
// 7. state_channel    - 状态通道演示器: 链下交易、争议解决
// 8. optimistic_rollup- Optimistic Rollup演示器: 欺诈证明
// 9. zk_rollup        - ZK Rollup演示器: 有效性证明
// 10. plasma          - Plasma演示器: 子链退出机制
// 11. data_availability- 数据可用性演示器: DA层
// 12. finality        - 最终性演示器: 跨链最终性
// =============================================================================

// RegisterAll 注册所有跨链模块到引擎
func RegisterAll(registry *engine.Registry) {
	// 跨链核心
	registry.Register("bridge", NewBridgeFactory())
	registry.Register("atomic_swap", NewAtomicSwapFactory())
	registry.Register("relay", NewRelayFactory())
	registry.Register("light_client", NewLightClientFactory())
	registry.Register("ibc", NewIBCFactory())
	registry.Register("oracle_bridge", NewOracleBridgeFactory())

	// Layer2扩容
	registry.Register("state_channel", NewStateChannelFactory())
	registry.Register("optimistic_rollup", NewOptimisticRollupFactory())
	registry.Register("zk_rollup", NewZKRollupFactory())
	registry.Register("plasma", NewPlasmaFactory())
	registry.Register("data_availability", NewDataAvailabilityFactory())
	registry.Register("finality", NewFinalityFactory())
}

// GetModuleList 获取模块列表
func GetModuleList() []map[string]string {
	return []map[string]string{
		{"id": "bridge", "name": "跨链桥演示器", "description": "锁定-铸造、流动性桥、多签验证", "category": "crosschain"},
		{"id": "atomic_swap", "name": "原子交换演示器", "description": "HTLC无信任跨链原子交换", "category": "crosschain"},
		{"id": "relay", "name": "中继链演示器", "description": "跨链消息传递、中继器机制", "category": "crosschain"},
		{"id": "light_client", "name": "轻客户端演示器", "description": "区块头验证、状态证明", "category": "crosschain"},
		{"id": "ibc", "name": "IBC协议演示器", "description": "Cosmos IBC跨链通信协议", "category": "crosschain"},
		{"id": "oracle_bridge", "name": "预言机桥演示器", "description": "预言机网络跨链验证", "category": "crosschain"},
		{"id": "state_channel", "name": "状态通道演示器", "description": "链下交易、争议解决", "category": "crosschain"},
		{"id": "optimistic_rollup", "name": "Optimistic Rollup演示器", "description": "欺诈证明、挑战期", "category": "crosschain"},
		{"id": "zk_rollup", "name": "ZK Rollup演示器", "description": "有效性证明、即时确认", "category": "crosschain"},
		{"id": "plasma", "name": "Plasma演示器", "description": "子链、退出机制", "category": "crosschain"},
		{"id": "data_availability", "name": "数据可用性演示器", "description": "DA层、纠删码、DAS", "category": "crosschain"},
		{"id": "finality", "name": "最终性演示器", "description": "跨链最终性机制", "category": "crosschain"},
	}
}

// GetModuleCount 获取模块数量
func GetModuleCount() int {
	return 12
}
