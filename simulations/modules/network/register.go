package network

import (
	"github.com/chainspace/simulations/pkg/engine"
)

// RegisterAll 注册所有网络领域模块。
func RegisterAll(registry *engine.Registry) {
	registry.Register("topology", NewTopologyFactory())
	registry.Register("kademlia", NewKademliaFactory())
	registry.Register("discovery", NewDiscoveryFactory())
	registry.Register("gossip", NewGossipFactory())
	registry.Register("block_propagation", NewBlockPropagationFactory())
	registry.Register("partition", NewPartitionFactory())
	registry.Register("nat_traversal", NewNATTraversalFactory())
	registry.Register("eclipse_attack", NewEclipseAttackFactory())
	registry.Register("sybil_attack", NewSybilAttackFactory())
	registry.Register("bgp_hijack", NewBGPHijackFactory())
}

// GetModuleList 返回网络领域模块目录。
func GetModuleList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "topology",
			"name":        "网络拓扑演示器",
			"description": "演示 P2P 网络中的拓扑结构、连通性与节点故障影响。",
			"type":        "demo",
			"category":    "network",
		},
		{
			"id":          "kademlia",
			"name":        "Kademlia DHT 演示器",
			"description": "演示分布式哈希表中的 XOR 距离、K 桶与节点查找过程。",
			"type":        "process",
			"category":    "network",
		},
		{
			"id":          "discovery",
			"name":        "节点发现演示器",
			"description": "演示新节点如何通过 bootstrap、邻居交换与持续发现加入网络。",
			"type":        "process",
			"category":    "network",
		},
		{
			"id":          "gossip",
			"name":        "Gossip 协议演示器",
			"description": "演示节点如何通过 gossip 机制扩散消息和状态。",
			"type":        "process",
			"category":    "network",
		},
		{
			"id":          "block_propagation",
			"name":        "区块传播演示器",
			"description": "演示区块在 P2P 网络中的传播路径、延迟与优化策略。",
			"type":        "process",
			"category":    "network",
		},
		{
			"id":          "partition",
			"name":        "网络分区演示器",
			"description": "演示网络分区对共识结果、一致性与传播路径的影响。",
			"type":        "process",
			"category":    "network",
		},
		{
			"id":          "nat_traversal",
			"name":        "NAT 穿透演示器",
			"description": "演示 NAT 类型识别与打洞、回落中继等连接建立过程。",
			"type":        "demo",
			"category":    "network",
		},
		{
			"id":          "eclipse_attack",
			"name":        "Eclipse 攻击演示器",
			"description": "演示目标节点如何被恶意邻居包围并逐步失去真实网络视图。",
			"type":        "attack",
			"category":    "network",
		},
		{
			"id":          "sybil_attack",
			"name":        "Sybil 攻击演示器",
			"description": "演示攻击者如何通过大量伪造身份扰乱网络连接与传播路径。",
			"type":        "attack",
			"category":    "network",
		},
		{
			"id":          "bgp_hijack",
			"name":        "BGP 劫持演示器",
			"description": "演示路由劫持如何改变跨网络连接、传播时延和可达性。",
			"type":        "attack",
			"category":    "network",
		},
	}
}
