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
// 日蚀攻击演示器
// =============================================================================

// EclipseNode 日蚀攻击节点
type EclipseNode struct {
	ID            string    `json:"id"`
	IsMalicious   bool      `json:"is_malicious"`
	Connections   []string  `json:"connections"`
	IsEclipsed    bool      `json:"is_eclipsed"`
	BlockHeight   uint64    `json:"block_height"`
	LastBlockTime time.Time `json:"last_block_time"`
}

// EclipseAttackState 日蚀攻击状态
type EclipseAttackState struct {
	TargetNode          string   `json:"target_node"`
	MaliciousNodes      []string `json:"malicious_nodes"`
	OriginalPeers       []string `json:"original_peers"`
	Phase               string   `json:"phase"` // preparation, execution, exploitation
	ConnectionsHijacked int      `json:"connections_hijacked"`
	IsSuccessful        bool     `json:"is_successful"`
}

// EclipseAttackSimulator 日蚀攻击演示器
// 演示日蚀攻击的原理和过程:
//
// 攻击原理:
// 攻击者控制目标节点的所有网络连接，
// 使目标节点与真实网络隔离，只能看到攻击者提供的信息
//
// 攻击步骤:
// 1. 准备阶段 - 生成大量节点地址
// 2. 执行阶段 - 占据目标节点的所有连接槽
// 3. 利用阶段 - 向目标提供虚假信息
//
// 影响:
// 1. 双花攻击 - 对商家确认虚假交易
// 2. 0确认攻击 - 商家看不到真实交易
// 3. 挖矿算力浪费 - 矿工在无效链上挖矿
type EclipseAttackSimulator struct {
	*base.BaseSimulator
	nodes       map[string]*EclipseNode
	attackState *EclipseAttackState
}

// NewEclipseAttackSimulator 创建日蚀攻击演示器
func NewEclipseAttackSimulator() *EclipseAttackSimulator {
	sim := &EclipseAttackSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"eclipse_attack",
			"日蚀攻击演示器",
			"演示日蚀攻击如何隔离目标节点并进行攻击",
			"network",
			types.ComponentAttack,
		),
		nodes: make(map[string]*EclipseNode),
	}

	sim.AddParam(types.Param{
		Key:         "honest_nodes",
		Name:        "诚实节点数",
		Description: "网络中的诚实节点数量",
		Type:        types.ParamTypeInt,
		Default:     20,
		Min:         5,
		Max:         100,
	})

	sim.AddParam(types.Param{
		Key:         "malicious_nodes",
		Name:        "恶意节点数",
		Description: "攻击者控制的节点数量",
		Type:        types.ParamTypeInt,
		Default:     50,
		Min:         10,
		Max:         200,
	})

	return sim
}

// Init 初始化
func (s *EclipseAttackSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	honestNodes := 20
	maliciousNodes := 50

	if v, ok := config.Params["honest_nodes"]; ok {
		if n, ok := v.(float64); ok {
			honestNodes = int(n)
		}
	}
	if v, ok := config.Params["malicious_nodes"]; ok {
		if n, ok := v.(float64); ok {
			maliciousNodes = int(n)
		}
	}

	s.initializeNetwork(honestNodes, maliciousNodes)
	s.updateState()
	return nil
}

// initializeNetwork 初始化网络
func (s *EclipseAttackSimulator) initializeNetwork(honestCount, maliciousCount int) {
	s.nodes = make(map[string]*EclipseNode)
	rand.Seed(time.Now().UnixNano())

	// 创建诚实节点
	for i := 0; i < honestCount; i++ {
		nodeID := fmt.Sprintf("honest-%d", i)
		s.nodes[nodeID] = &EclipseNode{
			ID:            nodeID,
			IsMalicious:   false,
			Connections:   make([]string, 0),
			IsEclipsed:    false,
			BlockHeight:   100000,
			LastBlockTime: time.Now(),
		}
	}

	// 创建恶意节点
	for i := 0; i < maliciousCount; i++ {
		nodeID := fmt.Sprintf("malicious-%d", i)
		s.nodes[nodeID] = &EclipseNode{
			ID:            nodeID,
			IsMalicious:   true,
			Connections:   make([]string, 0),
			IsEclipsed:    false,
			BlockHeight:   100000,
			LastBlockTime: time.Now(),
		}
	}

	// 为诚实节点建立连接
	honestIDs := s.getHonestNodeIDs()
	for _, nodeID := range honestIDs {
		node := s.nodes[nodeID]
		for len(node.Connections) < 8 {
			peer := honestIDs[rand.Intn(len(honestIDs))]
			if peer != nodeID && !contains(node.Connections, peer) {
				node.Connections = append(node.Connections, peer)
			}
		}
	}
}

// getHonestNodeIDs 获取诚实节点ID
func (s *EclipseAttackSimulator) getHonestNodeIDs() []string {
	ids := make([]string, 0)
	for id, node := range s.nodes {
		if !node.IsMalicious {
			ids = append(ids, id)
		}
	}
	return ids
}

// getMaliciousNodeIDs 获取恶意节点ID
func (s *EclipseAttackSimulator) getMaliciousNodeIDs() []string {
	ids := make([]string, 0)
	for id, node := range s.nodes {
		if node.IsMalicious {
			ids = append(ids, id)
		}
	}
	return ids
}

// ExplainAttack 解释日蚀攻击
func (s *EclipseAttackSimulator) ExplainAttack() map[string]interface{} {
	return map[string]interface{}{
		"name":       "日蚀攻击 (Eclipse Attack)",
		"target":     "单个节点",
		"goal":       "隔离目标节点，使其只能看到攻击者控制的视图",
		"difficulty": "需要控制大量IP地址",
		"attack_vectors": []map[string]string{
			{"vector": "地址表污染", "description": "用恶意地址填满目标的地址表"},
			{"vector": "连接饥饿", "description": "占用所有入站连接槽"},
			{"vector": "节点重启利用", "description": "在目标重启时抢占连接"},
		},
		"exploitation": []map[string]string{
			{"attack": "双花", "description": "向被日蚀的商家确认虚假交易"},
			{"attack": "0确认攻击", "description": "商家看不到真实网络的交易"},
			{"attack": "算力浪费", "description": "被日蚀的矿工在无效链上挖矿"},
			{"attack": "自私挖矿", "description": "配合自私挖矿策略"},
		},
		"requirements": []string{
			"大量不同IP地址 (僵尸网络/云服务器)",
			"目标节点使用默认配置",
			"持续时间维持连接",
		},
	}
}

// SimulateAttack 模拟日蚀攻击
func (s *EclipseAttackSimulator) SimulateAttack(targetNode string) *EclipseAttackState {
	target := s.nodes[targetNode]
	if target == nil {
		// 选择第一个诚实节点作为目标
		for id, node := range s.nodes {
			if !node.IsMalicious {
				targetNode = id
				target = node
				break
			}
		}
	}

	s.attackState = &EclipseAttackState{
		TargetNode:     targetNode,
		MaliciousNodes: s.getMaliciousNodeIDs(),
		OriginalPeers:  make([]string, len(target.Connections)),
		Phase:          "preparation",
	}
	copy(s.attackState.OriginalPeers, target.Connections)

	// Phase 1: 准备
	s.EmitEvent("eclipse_phase", "", "", map[string]interface{}{
		"phase":   "preparation",
		"target":  targetNode,
		"details": "攻击者生成大量节点地址，准备填充目标的地址表",
	})

	// Phase 2: 执行 - 替换连接
	s.attackState.Phase = "execution"
	maliciousIDs := s.getMaliciousNodeIDs()

	for i := 0; i < len(target.Connections) && i < len(maliciousIDs); i++ {
		target.Connections[i] = maliciousIDs[i]
		s.attackState.ConnectionsHijacked++
	}

	s.EmitEvent("eclipse_phase", "", "", map[string]interface{}{
		"phase":                "execution",
		"connections_hijacked": s.attackState.ConnectionsHijacked,
		"details":              "逐步替换目标节点的连接为恶意节点",
	})

	// Phase 3: 利用
	s.attackState.Phase = "exploitation"
	s.attackState.IsSuccessful = s.attackState.ConnectionsHijacked == len(s.attackState.OriginalPeers)
	target.IsEclipsed = s.attackState.IsSuccessful

	s.EmitEvent("eclipse_phase", "", "", map[string]interface{}{
		"phase":      "exploitation",
		"successful": s.attackState.IsSuccessful,
		"details":    "目标节点已被隔离，攻击者可以提供虚假信息",
	})

	s.updateState()
	return s.attackState
}

// SimulateDoubleSpend 模拟双花攻击
func (s *EclipseAttackSimulator) SimulateDoubleSpend() map[string]interface{} {
	if s.attackState == nil || !s.attackState.IsSuccessful {
		return map[string]interface{}{"error": "需要先成功执行日蚀攻击"}
	}

	return map[string]interface{}{
		"attack": "双花攻击",
		"target": s.attackState.TargetNode,
		"attack_flow": []string{
			"1. 攻击者向被日蚀的商家发送交易TX1",
			"2. 同时向真实网络发送冲突交易TX2",
			"3. 被日蚀商家只能看到TX1",
			"4. 商家确认TX1并交付商品",
			"5. TX2在真实网络被确认",
			"6. TX1作废，攻击者获得商品但不损失资金",
		},
		"victim_view": map[string]interface{}{
			"sees_tx1":     true,
			"sees_tx2":     false,
			"block_height": "滞后于真实网络",
		},
		"prevention": []string{
			"增加连接多样性",
			"使用多个独立节点验证",
			"等待更多确认数",
		},
	}
}

// ShowDefenses 显示防御方法
func (s *EclipseAttackSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"defense":     "增加出站连接",
			"description": "主动连接更多节点，减少被动依赖入站连接",
			"bitcoin":     "Bitcoin Core 增加了出站连接数",
		},
		{
			"defense":     "地址表分区",
			"description": "将地址按网络来源分区存储，防止单一来源填满",
			"bitcoin":     "使用tried和new两个表",
		},
		{
			"defense":     "固定种子节点",
			"description": "保持与可信种子节点的连接",
		},
		{
			"defense":     "Feeler连接",
			"description": "定期测试新地址的可达性",
		},
		{
			"defense":     "多节点验证",
			"description": "使用多个独立节点交叉验证信息",
		},
	}
}

// GetRealWorldCases 获取真实案例
func (s *EclipseAttackSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"paper":   "Eclipse Attacks on Bitcoin's Peer-to-Peer Network",
			"authors": "Heilman et al.",
			"year":    2015,
			"finding": "只需少量IP即可日蚀比特币节点",
			"impact":  "导致比特币Core多项安全改进",
		},
		{
			"paper":   "Low-Resource Eclipse Attacks on Ethereum's Peer-to-Peer Network",
			"authors": "Marcus et al.",
			"year":    2018,
			"finding": "以太坊的节点发现更容易被利用",
			"impact":  "Geth和Parity改进了节点发现机制",
		},
	}
}

// updateState 更新状态
func (s *EclipseAttackSimulator) updateState() {
	honestCount := 0
	maliciousCount := 0
	eclipsedCount := 0

	for _, node := range s.nodes {
		if node.IsMalicious {
			maliciousCount++
		} else {
			honestCount++
			if node.IsEclipsed {
				eclipsedCount++
			}
		}
	}

	s.SetGlobalData("honest_nodes", honestCount)
	s.SetGlobalData("malicious_nodes", maliciousCount)
	s.SetGlobalData("eclipsed_nodes", eclipsedCount)

	stage := "network_ready"
	summary := "当前还没有执行 Eclipse 攻击，目标节点仍然拥有正常的邻居视图。"
	nextHint := "发起一轮 Eclipse 攻击后，重点观察目标节点的连接如何被恶意节点逐步替换。"
	progress := 0.2
	result := map[string]interface{}{
		"honest_nodes":    honestCount,
		"malicious_nodes": maliciousCount,
		"eclipsed_nodes":  eclipsedCount,
	}

	if s.attackState != nil {
		stage = "attack_active"
		summary = fmt.Sprintf("当前正在对节点 %s 执行 Eclipse 攻击，已劫持 %d 条连接。", s.attackState.TargetNode, s.attackState.ConnectionsHijacked)
		nextHint = "继续观察目标节点是否完全失去真实邻居视图，以及攻击是否足以支撑双花。"
		progress = 0.85
		result["target_node"] = s.attackState.TargetNode
		result["connections_hijacked"] = s.attackState.ConnectionsHijacked
		result["successful"] = s.attackState.IsSuccessful
	}

	setNetworkTeachingState(s.BaseSimulator, "network", stage, summary, nextHint, progress, result)
}

// ExecuteAction 为 Eclipse 攻击实验提供交互动作。
func (s *EclipseAttackSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_attack":
		targetNode, _ := params["target_node"].(string)
		state := s.SimulateAttack(targetNode)
		return networkActionResult(
			"已执行一轮 Eclipse 攻击。",
			map[string]interface{}{
				"target_node":           state.TargetNode,
				"connections_hijacked":  state.ConnectionsHijacked,
				"successful":            state.IsSuccessful,
			},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("目标节点 %s 的邻居连接已经被恶意节点大量替换。", state.TargetNode),
				NextHint:    "继续观察目标节点所见网络视图与真实网络之间的偏差，以及双花攻击是否成立。",
				EffectScope: "network",
				ResultState: map[string]interface{}{
					"successful": state.IsSuccessful,
				},
			},
		), nil
	case "simulate_double_spend":
		result := s.SimulateDoubleSpend()
		if _, ok := result["error"]; ok {
			return &types.ActionResult{Success: false, Message: "需要先成功执行 Eclipse 攻击。"}, nil
		}
		return networkActionResult(
			"已模拟 Eclipse 攻击下的双花过程。",
			result,
			&types.ActionFeedback{
				Summary:     "双花路径已经建立，受害节点只看到了被操纵后的局部视图。",
				NextHint:    "重点比较受害者视图与真实链状态之间的差异，以及交易何时被回滚。",
				EffectScope: "network",
				ResultState: result,
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported eclipse attack action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// EclipseAttackFactory 日蚀攻击工厂
type EclipseAttackFactory struct{}

// Create 创建演示器
func (f *EclipseAttackFactory) Create() engine.Simulator {
	return NewEclipseAttackSimulator()
}

// GetDescription 获取描述
func (f *EclipseAttackFactory) GetDescription() types.Description {
	return NewEclipseAttackSimulator().GetDescription()
}

// NewEclipseAttackFactory 创建工厂
func NewEclipseAttackFactory() *EclipseAttackFactory {
	return &EclipseAttackFactory{}
}

var _ engine.SimulatorFactory = (*EclipseAttackFactory)(nil)
