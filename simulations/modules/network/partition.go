package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 网络分区演示器
// =============================================================================

// PartitionType 分区类型
type PartitionType string

const (
	PartitionSymmetric  PartitionType = "symmetric"  // 对称分区
	PartitionAsymmetric PartitionType = "asymmetric" // 非对称分区
	PartitionPartial    PartitionType = "partial"    // 部分分区
)

// NetworkPartition 网络分区
type NetworkPartition struct {
	ID        string        `json:"id"`
	Type      PartitionType `json:"type"`
	GroupA    []string      `json:"group_a"`
	GroupB    []string      `json:"group_b"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	IsActive  bool          `json:"is_active"`
}

// PartitionNode 分区节点
type PartitionNode struct {
	ID          string   `json:"id"`
	Group       string   `json:"group"` // "A", "B", or "bridge"
	Peers       []string `json:"peers"`
	CanReach    []string `json:"can_reach"`
	BlockHeight uint64   `json:"block_height"`
}

// PartitionEffect 分区效果
type PartitionEffect struct {
	GroupABlocks      uint64 `json:"group_a_blocks"`
	GroupBBlocks      uint64 `json:"group_b_blocks"`
	ChainDivergence   uint64 `json:"chain_divergence"`
	TransactionsSplit int    `json:"transactions_split"`
	ReorgRequired     bool   `json:"reorg_required"`
}

// PartitionSimulator 网络分区演示器
// 演示网络分区对区块链的影响:
//
// 分区类型:
// 1. 对称分区 - 两个组完全隔离
// 2. 非对称分区 - A可达B但B不可达A
// 3. 部分分区 - 部分节点作为桥接
//
// 影响:
// 1. 链分叉 - 两个组各自出块
// 2. 双花风险 - 分区期间交易可能在两边确认
// 3. 重组 - 分区恢复后需要重组
type PartitionSimulator struct {
	*base.BaseSimulator
	nodes           map[string]*PartitionNode
	partitions      []*NetworkPartition
	activePartition *NetworkPartition
}

// NewPartitionSimulator 创建网络分区演示器
func NewPartitionSimulator() *PartitionSimulator {
	sim := &PartitionSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"partition",
			"网络分区演示器",
			"演示网络分区对区块链共识和一致性的影响",
			"network",
			types.ComponentProcess,
		),
		nodes:      make(map[string]*PartitionNode),
		partitions: make([]*NetworkPartition, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "网络中的节点数量",
		Type:        types.ParamTypeInt,
		Default:     20,
		Min:         6,
		Max:         100,
	})

	sim.AddParam(types.Param{
		Key:         "partition_type",
		Name:        "分区类型",
		Description: "网络分区的类型",
		Type:        types.ParamTypeSelect,
		Default:     "symmetric",
		Options: []types.Option{
			{Label: "对称分区", Value: "symmetric"},
			{Label: "非对称分区", Value: "asymmetric"},
			{Label: "部分分区", Value: "partial"},
		},
	})

	return sim
}

// Init 初始化
func (s *PartitionSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	nodeCount := 20
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	s.initializeNetwork(nodeCount)
	s.updateState()
	return nil
}

// initializeNetwork 初始化网络
func (s *PartitionSimulator) initializeNetwork(size int) {
	s.nodes = make(map[string]*PartitionNode)
	rand.Seed(time.Now().UnixNano())

	nodeIDs := make([]string, size)
	for i := 0; i < size; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		nodeIDs[i] = nodeID

		s.nodes[nodeID] = &PartitionNode{
			ID:          nodeID,
			Group:       "",
			Peers:       make([]string, 0),
			CanReach:    make([]string, 0),
			BlockHeight: 1000,
		}
	}

	// 建立连接
	for _, node := range s.nodes {
		peerCount := 4 + rand.Intn(4)
		for len(node.Peers) < peerCount {
			peer := nodeIDs[rand.Intn(len(nodeIDs))]
			if peer != node.ID && !contains(node.Peers, peer) {
				node.Peers = append(node.Peers, peer)
				node.CanReach = append(node.CanReach, peer)
			}
		}
	}
}

// ExplainPartition 解释网络分区
func (s *PartitionSimulator) ExplainPartition() map[string]interface{} {
	return map[string]interface{}{
		"definition": "网络分区是指网络中的节点被分成多个无法相互通信的组",
		"causes": []string{
			"网络设备故障(路由器、交换机)",
			"海底光缆断裂",
			"BGP配置错误",
			"DDoS攻击",
			"自然灾害",
			"政府审查/防火墙",
		},
		"types": []map[string]string{
			{"type": "symmetric", "description": "两组完全隔离，互相不可达"},
			{"type": "asymmetric", "description": "单向分区，A可达B但B不可达A"},
			{"type": "partial", "description": "部分节点作为桥接，可能延迟传播"},
		},
		"blockchain_impact": []string{
			"链分叉: 两个组各自产生区块",
			"双花风险: 同一交易可能在两边确认",
			"共识延迟: 无法达成全网共识",
			"重组: 分区恢复后较短链被废弃",
		},
		"cap_theorem": "根据CAP定理，分区时必须在一致性(C)和可用性(A)之间选择",
	}
}

// CreatePartition 创建网络分区
func (s *PartitionSimulator) CreatePartition(partitionType PartitionType, splitRatio float64) *NetworkPartition {
	nodeIDs := make([]string, 0, len(s.nodes))
	for id := range s.nodes {
		nodeIDs = append(nodeIDs, id)
	}

	splitPoint := int(float64(len(nodeIDs)) * splitRatio)
	groupA := nodeIDs[:splitPoint]
	groupB := nodeIDs[splitPoint:]

	partition := &NetworkPartition{
		ID:        fmt.Sprintf("partition-%d", len(s.partitions)+1),
		Type:      partitionType,
		GroupA:    groupA,
		GroupB:    groupB,
		StartTime: time.Now(),
		IsActive:  true,
	}

	// 更新节点分组
	for _, nodeID := range groupA {
		s.nodes[nodeID].Group = "A"
	}
	for _, nodeID := range groupB {
		s.nodes[nodeID].Group = "B"
	}

	// 根据分区类型更新可达性
	switch partitionType {
	case PartitionSymmetric:
		s.applySymmetricPartition(groupA, groupB)
	case PartitionAsymmetric:
		s.applyAsymmetricPartition(groupA, groupB)
	case PartitionPartial:
		s.applyPartialPartition(groupA, groupB)
	}

	s.activePartition = partition
	s.partitions = append(s.partitions, partition)

	s.EmitEvent("partition_created", "", "", map[string]interface{}{
		"partition_id": partition.ID,
		"type":         partitionType,
		"group_a_size": len(groupA),
		"group_b_size": len(groupB),
	})

	s.updateState()
	return partition
}

// applySymmetricPartition 应用对称分区
func (s *PartitionSimulator) applySymmetricPartition(groupA, groupB []string) {
	groupASet := make(map[string]bool)
	for _, id := range groupA {
		groupASet[id] = true
	}

	for _, nodeID := range groupA {
		node := s.nodes[nodeID]
		node.CanReach = filterPeers(node.Peers, groupASet)
	}

	groupBSet := make(map[string]bool)
	for _, id := range groupB {
		groupBSet[id] = true
	}

	for _, nodeID := range groupB {
		node := s.nodes[nodeID]
		node.CanReach = filterPeers(node.Peers, groupBSet)
	}
}

// applyAsymmetricPartition 应用非对称分区
func (s *PartitionSimulator) applyAsymmetricPartition(groupA, groupB []string) {
	// A可以到达所有节点
	for _, nodeID := range groupA {
		node := s.nodes[nodeID]
		node.CanReach = node.Peers
	}

	// B只能到达B组节点
	groupBSet := make(map[string]bool)
	for _, id := range groupB {
		groupBSet[id] = true
	}

	for _, nodeID := range groupB {
		node := s.nodes[nodeID]
		node.CanReach = filterPeers(node.Peers, groupBSet)
	}
}

// applyPartialPartition 应用部分分区
func (s *PartitionSimulator) applyPartialPartition(groupA, groupB []string) {
	// 选择一个桥接节点
	if len(groupA) > 0 {
		bridgeNode := groupA[len(groupA)-1]
		s.nodes[bridgeNode].Group = "bridge"
		s.nodes[bridgeNode].CanReach = s.nodes[bridgeNode].Peers // 桥接节点可达所有
	}

	// 其他节点应用对称分区
	s.applySymmetricPartition(groupA[:len(groupA)-1], groupB)
}

// filterPeers 过滤节点
func filterPeers(peers []string, allowedSet map[string]bool) []string {
	result := make([]string, 0)
	for _, peer := range peers {
		if allowedSet[peer] {
			result = append(result, peer)
		}
	}
	return result
}

// SimulatePartitionEffect 模拟分区效果
func (s *PartitionSimulator) SimulatePartitionEffect(blocksDuringPartition int) *PartitionEffect {
	if s.activePartition == nil {
		return nil
	}

	effect := &PartitionEffect{}

	// 模拟两组各自出块
	for i := 0; i < blocksDuringPartition; i++ {
		// A组出块
		for _, nodeID := range s.activePartition.GroupA {
			s.nodes[nodeID].BlockHeight++
		}
		effect.GroupABlocks++

		// B组出块
		for _, nodeID := range s.activePartition.GroupB {
			s.nodes[nodeID].BlockHeight++
		}
		effect.GroupBBlocks++
	}

	effect.ChainDivergence = effect.GroupABlocks + effect.GroupBBlocks
	effect.TransactionsSplit = int(effect.ChainDivergence) * 100 // 假设每块100交易
	effect.ReorgRequired = true

	s.EmitEvent("partition_effect", "", "", map[string]interface{}{
		"group_a_blocks":   effect.GroupABlocks,
		"group_b_blocks":   effect.GroupBBlocks,
		"chain_divergence": effect.ChainDivergence,
		"reorg_required":   effect.ReorgRequired,
	})

	return effect
}

// HealPartition 恢复分区
func (s *PartitionSimulator) HealPartition() map[string]interface{} {
	if s.activePartition == nil {
		return map[string]interface{}{"error": "没有活动的分区"}
	}

	s.activePartition.IsActive = false
	s.activePartition.Duration = time.Since(s.activePartition.StartTime)

	// 恢复所有节点的可达性
	for _, node := range s.nodes {
		node.CanReach = node.Peers
		node.Group = ""
	}

	// 确定获胜链 (假设A组更长)
	winningGroup := "A"
	reorgBlocks := s.activePartition.Duration.Seconds() / 12 // 假设12秒出块

	result := map[string]interface{}{
		"partition_id":  s.activePartition.ID,
		"duration":      s.activePartition.Duration.String(),
		"winning_group": winningGroup,
		"reorg_blocks":  int(reorgBlocks),
		"orphaned_txs":  int(reorgBlocks) * 100,
	}

	s.EmitEvent("partition_healed", "", "", result)

	s.activePartition = nil
	s.updateState()

	return result
}

// GetRealWorldCases 获取真实案例
func (s *PartitionSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"event":      "以太坊Ropsten测试网分区",
			"date":       "2020",
			"cause":      "客户端bug导致节点分裂",
			"impact":     "测试网暂时分叉",
			"resolution": "修复bug并重新同步",
		},
		{
			"event":      "比特币意外分叉",
			"date":       "2013-03-11",
			"cause":      "0.8版本数据库升级导致与旧版本不兼容",
			"impact":     "链分叉约6小时",
			"resolution": "矿工回滚到0.7版本",
		},
		{
			"event":      "Solana网络中断",
			"date":       "2021-09-14",
			"cause":      "验证者资源耗尽导致网络分区",
			"impact":     "网络停止17小时",
			"resolution": "协调验证者重启",
		},
	}
}

// updateState 更新状态
func (s *PartitionSimulator) updateState() {
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("partition_count", len(s.partitions))
	s.SetGlobalData("has_active_partition", s.activePartition != nil)

	stage := "network_ready"
	summary := "当前网络未发生分区，可以创建一轮新的网络分区并观察链状态如何分裂。"
	nextHint := "先创建分区，再观察组内可达性、链分叉和恢复后的重组结果。"
	progress := 0.2
	result := map[string]interface{}{
		"node_count":       len(s.nodes),
		"partition_count":  len(s.partitions),
		"partition_active": false,
	}

	if s.activePartition != nil {
		stage = "partition_active"
		summary = fmt.Sprintf(
			"当前存在一轮 %s 分区，A 组 %d 个节点，B 组 %d 个节点。",
			s.activePartition.Type,
			len(s.activePartition.GroupA),
			len(s.activePartition.GroupB),
		)
		nextHint = "继续模拟分区期间的出块和交易分裂，然后再恢复网络，观察是否需要重组。"
		progress = 0.7
		result["partition_active"] = true
		result["partition_type"] = s.activePartition.Type
		result["group_a_size"] = len(s.activePartition.GroupA)
		result["group_b_size"] = len(s.activePartition.GroupB)
	}

	setNetworkTeachingState(s.BaseSimulator, "network", stage, summary, nextHint, progress, result)
}

// ExecuteAction 为网络分区实验提供交互动作。
func (s *PartitionSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_partition":
		partitionType := PartitionSymmetric
		splitRatio := 0.5
		if raw, ok := params["partition_type"].(string); ok && raw != "" {
			partitionType = PartitionType(raw)
		}
		if raw, ok := params["split_ratio"].(float64); ok && raw > 0.1 && raw < 0.9 {
			splitRatio = raw
		}
		partition := s.CreatePartition(partitionType, splitRatio)
		return networkActionResult(
			"已创建一轮网络分区。",
			map[string]interface{}{
				"partition_id": partition.ID,
				"partition_type": partition.Type,
				"group_a_size": len(partition.GroupA),
				"group_b_size": len(partition.GroupB),
			},
			&types.ActionFeedback{
				Summary:     "网络已经被拆分成两个分区，后续传播和出块将不再保持全局一致。",
				NextHint:    "继续模拟分区期间的链分裂，再恢复网络观察最终重组结果。",
				EffectScope: "network",
				ResultState: map[string]interface{}{
					"partition_type": partition.Type,
					"group_a_size":   len(partition.GroupA),
					"group_b_size":   len(partition.GroupB),
				},
			},
		), nil
	case "simulate_partition_effect":
		blocks := 3
		if raw, ok := params["blocks"].(float64); ok && int(raw) >= 1 {
			blocks = int(raw)
		}
		effect := s.SimulatePartitionEffect(blocks)
		if effect == nil {
			return &types.ActionResult{Success: false, Message: "当前没有活动中的网络分区。"}, nil
		}
		return networkActionResult(
			"已模拟分区期间的链分裂影响。",
			map[string]interface{}{
				"group_a_blocks": effect.GroupABlocks,
				"group_b_blocks": effect.GroupBBlocks,
				"chain_divergence": effect.ChainDivergence,
			},
			&types.ActionFeedback{
				Summary:     "两个分区已经各自推进链状态，网络恢复后可能触发重组。",
				NextHint:    "恢复分区后重点观察哪一组获胜、多少区块被回滚以及多少交易被丢弃。",
				EffectScope: "network",
				ResultState: map[string]interface{}{
					"chain_divergence": effect.ChainDivergence,
				},
			},
		), nil
	case "heal_partition":
		result := s.HealPartition()
		if _, ok := result["error"]; ok {
			return &types.ActionResult{Success: false, Message: "当前没有活动中的网络分区。"}, nil
		}
		return networkActionResult(
			"已恢复网络分区。",
			result,
			&types.ActionFeedback{
				Summary:     "网络重新连通，系统开始比较两个分区的链状态并决定是否重组。",
				NextHint:    "重点观察获胜分区、回滚区块数以及孤儿交易数量。",
				EffectScope: "network",
				ResultState: result,
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported partition action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// PartitionFactory 网络分区工厂
type PartitionFactory struct{}

// Create 创建演示器
func (f *PartitionFactory) Create() engine.Simulator {
	return NewPartitionSimulator()
}

// GetDescription 获取描述
func (f *PartitionFactory) GetDescription() types.Description {
	return NewPartitionSimulator().GetDescription()
}

// NewPartitionFactory 创建工厂
func NewPartitionFactory() *PartitionFactory {
	return &PartitionFactory{}
}

var _ engine.SimulatorFactory = (*PartitionFactory)(nil)
