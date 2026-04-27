package network

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type TopologyType string

const (
	TopologyFullMesh   TopologyType = "full_mesh"
	TopologyRing       TopologyType = "ring"
	TopologyStar       TopologyType = "star"
	TopologyTree       TopologyType = "tree"
	TopologyRandom     TopologyType = "random"
	TopologySmallWorld TopologyType = "small_world"
	TopologyScaleFree  TopologyType = "scale_free"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type NetworkNode struct {
	ID          string   `json:"id"`
	Address     string   `json:"address"`
	Degree      int      `json:"degree"`
	Connections []string `json:"connections"`
	Position    Position `json:"position"`
	IsOnline    bool     `json:"is_online"`
}

type NetworkEdge struct {
	From    string  `json:"from"`
	To      string  `json:"to"`
	Latency int     `json:"latency"`
	Weight  float64 `json:"weight"`
}

type TopologyStats struct {
	NodeCount       int     `json:"node_count"`
	EdgeCount       int     `json:"edge_count"`
	AvgDegree       float64 `json:"avg_degree"`
	MaxDegree       int     `json:"max_degree"`
	MinDegree       int     `json:"min_degree"`
	ClusteringCoeff float64 `json:"clustering_coefficient"`
	AvgPathLength   float64 `json:"avg_path_length"`
	Diameter        int     `json:"diameter"`
	IsConnected     bool    `json:"is_connected"`
}

type TopologySimulator struct {
	*base.BaseSimulator
	nodes        map[string]*NetworkNode
	edges        []*NetworkEdge
	topologyType TopologyType
	stats        *TopologyStats
}

func NewTopologySimulator() *TopologySimulator {
	sim := &TopologySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"topology",
			"网络拓扑演示器",
			"演示 P2P 网络中的典型拓扑结构、连通性与节点故障影响。",
			"network",
			types.ComponentDemo,
		),
		nodes: make(map[string]*NetworkNode),
		edges: make([]*NetworkEdge, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "网络中的节点数量",
		Type:        types.ParamTypeInt,
		Default:     20,
		Min:         3,
		Max:         100,
	})

	sim.AddParam(types.Param{
		Key:         "topology_type",
		Name:        "拓扑类型",
		Description: "选择当前网络的拓扑结构",
		Type:        types.ParamTypeSelect,
		Default:     string(TopologyRandom),
		Options: []types.Option{
			{Label: "全连接", Value: string(TopologyFullMesh)},
			{Label: "环形", Value: string(TopologyRing)},
			{Label: "星形", Value: string(TopologyStar)},
			{Label: "树形", Value: string(TopologyTree)},
			{Label: "随机图", Value: string(TopologyRandom)},
			{Label: "小世界", Value: string(TopologySmallWorld)},
			{Label: "无标度", Value: string(TopologyScaleFree)},
		},
	})

	return sim
}

func (s *TopologySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	nodeCount := 20
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	s.topologyType = TopologyRandom
	if v, ok := config.Params["topology_type"]; ok {
		if raw, ok := v.(string); ok && raw != "" {
			s.topologyType = TopologyType(raw)
		}
	}

	s.buildTopology(nodeCount)
	s.calculateStats()
	s.updateState()
	return nil
}

func (s *TopologySimulator) buildTopology(nodeCount int) {
	s.nodes = make(map[string]*NetworkNode)
	s.edges = make([]*NetworkEdge, 0)

	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		angle := 2 * math.Pi * float64(i) / float64(nodeCount)
		s.nodes[nodeID] = &NetworkNode{
			ID:          nodeID,
			Address:     fmt.Sprintf("192.168.1.%d:30303", i+1),
			Degree:      0,
			Connections: make([]string, 0),
			Position: Position{
				X: 200 + 150*math.Cos(angle),
				Y: 200 + 150*math.Sin(angle),
			},
			IsOnline: true,
		}
	}

	switch s.topologyType {
	case TopologyFullMesh:
		s.buildFullMesh()
	case TopologyRing:
		s.buildRing()
	case TopologyStar:
		s.buildStar()
	case TopologyTree:
		s.buildTree()
	case TopologySmallWorld:
		s.buildSmallWorld(4, 0.3)
	case TopologyScaleFree:
		s.buildScaleFree(2)
	default:
		s.buildRandom(0.3)
	}
}

func (s *TopologySimulator) buildFullMesh() {
	nodeIDs := s.getNodeIDs()
	for i := 0; i < len(nodeIDs); i++ {
		for j := i + 1; j < len(nodeIDs); j++ {
			s.addEdge(nodeIDs[i], nodeIDs[j])
		}
	}
}

func (s *TopologySimulator) buildRing() {
	nodeIDs := s.getNodeIDs()
	for i := 0; i < len(nodeIDs); i++ {
		next := (i + 1) % len(nodeIDs)
		s.addEdge(nodeIDs[i], nodeIDs[next])
	}
}

func (s *TopologySimulator) buildStar() {
	nodeIDs := s.getNodeIDs()
	if len(nodeIDs) == 0 {
		return
	}
	center := nodeIDs[0]
	for i := 1; i < len(nodeIDs); i++ {
		s.addEdge(center, nodeIDs[i])
	}
}

func (s *TopologySimulator) buildTree() {
	nodeIDs := s.getNodeIDs()
	for i := 1; i < len(nodeIDs); i++ {
		parent := (i - 1) / 2
		s.addEdge(nodeIDs[parent], nodeIDs[i])
	}
}

func (s *TopologySimulator) buildRandom(probability float64) {
	rand.Seed(time.Now().UnixNano())
	nodeIDs := s.getNodeIDs()
	for i := 0; i < len(nodeIDs); i++ {
		for j := i + 1; j < len(nodeIDs); j++ {
			if rand.Float64() < probability {
				s.addEdge(nodeIDs[i], nodeIDs[j])
			}
		}
	}
}

func (s *TopologySimulator) buildSmallWorld(k int, beta float64) {
	rand.Seed(time.Now().UnixNano())
	nodeIDs := s.getNodeIDs()
	n := len(nodeIDs)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		for j := 1; j <= k/2; j++ {
			neighbor := (i + j) % n
			s.addEdge(nodeIDs[i], nodeIDs[neighbor])
		}
	}

	for _, edge := range s.edges {
		if rand.Float64() < beta {
			newTarget := nodeIDs[rand.Intn(n)]
			if newTarget != edge.From && newTarget != edge.To {
				edge.To = newTarget
			}
		}
	}
}

func (s *TopologySimulator) buildScaleFree(m int) {
	rand.Seed(time.Now().UnixNano())
	nodeIDs := s.getNodeIDs()
	n := len(nodeIDs)
	if n <= m {
		s.buildFullMesh()
		return
	}

	for i := 0; i <= m; i++ {
		for j := i + 1; j <= m; j++ {
			s.addEdge(nodeIDs[i], nodeIDs[j])
		}
	}

	for i := m + 1; i < n; i++ {
		totalDegree := 0
		for _, node := range s.nodes {
			totalDegree += node.Degree
		}

		connected := 0
		for connected < m {
			for _, nodeID := range nodeIDs[:i] {
				if connected >= m {
					break
				}
				node := s.nodes[nodeID]
				if totalDegree == 0 {
					continue
				}
				probability := float64(node.Degree) / float64(totalDegree)
				if rand.Float64() < probability && !s.hasEdge(nodeIDs[i], nodeID) {
					s.addEdge(nodeIDs[i], nodeID)
					connected++
				}
			}
		}
	}
}

func (s *TopologySimulator) addEdge(from, to string) {
	edge := &NetworkEdge{
		From:    from,
		To:      to,
		Latency: 10 + rand.Intn(90),
		Weight:  1,
	}
	s.edges = append(s.edges, edge)

	s.nodes[from].Connections = append(s.nodes[from].Connections, to)
	s.nodes[from].Degree++
	s.nodes[to].Connections = append(s.nodes[to].Connections, from)
	s.nodes[to].Degree++
}

func (s *TopologySimulator) hasEdge(from, to string) bool {
	for _, edge := range s.edges {
		if (edge.From == from && edge.To == to) || (edge.From == to && edge.To == from) {
			return true
		}
	}
	return false
}

func (s *TopologySimulator) getNodeIDs() []string {
	ids := make([]string, 0, len(s.nodes))
	for id := range s.nodes {
		ids = append(ids, id)
	}
	return ids
}

func (s *TopologySimulator) calculateStats() {
	s.stats = &TopologyStats{
		NodeCount: len(s.nodes),
		EdgeCount: len(s.edges),
	}

	if len(s.nodes) == 0 {
		return
	}

	totalDegree := 0
	s.stats.MinDegree = int(^uint(0) >> 1)
	for _, node := range s.nodes {
		totalDegree += node.Degree
		if node.Degree > s.stats.MaxDegree {
			s.stats.MaxDegree = node.Degree
		}
		if node.Degree < s.stats.MinDegree {
			s.stats.MinDegree = node.Degree
		}
	}

	s.stats.AvgDegree = float64(totalDegree) / float64(len(s.nodes))
	s.stats.IsConnected = s.stats.MinDegree > 0

	switch s.topologyType {
	case TopologyFullMesh:
		s.stats.ClusteringCoeff = 1
		s.stats.AvgPathLength = 1
		s.stats.Diameter = 1
	case TopologyRing:
		s.stats.ClusteringCoeff = 0
		s.stats.AvgPathLength = float64(len(s.nodes)) / 4
		s.stats.Diameter = len(s.nodes) / 2
	case TopologyStar:
		s.stats.ClusteringCoeff = 0
		s.stats.AvgPathLength = 2
		s.stats.Diameter = 2
	case TopologySmallWorld:
		s.stats.ClusteringCoeff = 0.42
		s.stats.AvgPathLength = math.Log(float64(len(s.nodes))) + 0.5
		s.stats.Diameter = max(2, int(math.Ceil(math.Log(float64(len(s.nodes))))))
	case TopologyScaleFree:
		s.stats.ClusteringCoeff = 0.18
		s.stats.AvgPathLength = math.Log(math.Log(float64(max(3, len(s.nodes)))))
		s.stats.Diameter = max(2, int(math.Ceil(math.Log(float64(len(s.nodes))))))
	default:
		s.stats.ClusteringCoeff = 0.12
		s.stats.AvgPathLength = math.Log(float64(max(3, len(s.nodes))))
		s.stats.Diameter = max(2, int(math.Ceil(math.Log(float64(len(s.nodes))))))
	}
}

func (s *TopologySimulator) GetTopologyInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":  s.topologyType,
		"stats": s.stats,
		"nodes": s.nodes,
		"edges": s.edges,
	}
}

func (s *TopologySimulator) CompareTopologies() []map[string]interface{} {
	return []map[string]interface{}{
		{"type": "full_mesh", "name": "全连接网络", "edges": "n(n-1)/2", "avg_degree": "n-1", "diameter": 1, "pros": []string{"延迟最低", "冗余度高"}, "cons": []string{"边数量爆炸", "扩展性差"}, "use_case": "小规模、强一致性或低延迟网络"},
		{"type": "ring", "name": "环形网络", "edges": "n", "avg_degree": 2, "diameter": "n/2", "pros": []string{"结构简单", "边开销低"}, "cons": []string{"路径较长", "抗故障能力弱"}, "use_case": "令牌环或教学演示基础结构"},
		{"type": "star", "name": "星形网络", "edges": "n-1", "avg_degree": "2(n-1)/n", "diameter": 2, "pros": []string{"管理简单", "中心节点协调成本低"}, "cons": []string{"中心节点单点故障"}, "use_case": "中心协调型网络"},
		{"type": "random", "name": "随机图", "edges": "p*n(n-1)/2", "diameter": "O(log n)", "pros": []string{"抗随机故障", "易于理论分析"}, "cons": []string{"不完全符合真实网络度分布"}, "use_case": "随机连接或对比实验"},
		{"type": "small_world", "name": "小世界网络", "properties": []string{"高聚类", "短平均路径"}, "diameter": "O(log n)", "pros": []string{"传播快", "局部连接稳定"}, "cons": []string{"构建和维护较复杂"}, "use_case": "社交网络、真实 P2P 邻接"},
		{"type": "scale_free", "name": "无标度网络", "properties": []string{"幂律度分布", "存在超级节点"}, "diameter": "O(log log n)", "pros": []string{"路径更短", "抗随机故障能力强"}, "cons": []string{"核心节点更易成为攻击目标"}, "use_case": "互联网骨干、现实世界复杂网络"},
	}
}

func (s *TopologySimulator) SimulateNodeFailure(nodeID string) map[string]interface{} {
	node, ok := s.nodes[nodeID]
	if !ok {
		return map[string]interface{}{"error": "节点不存在"}
	}

	node.IsOnline = false
	affectedNodes := append([]string(nil), node.Connections...)
	s.EmitEvent("node_failure", "", "", map[string]interface{}{
		"node_id":        nodeID,
		"degree":         node.Degree,
		"affected_nodes": affectedNodes,
	})
	s.calculateStats()

	return map[string]interface{}{
		"node_id":           nodeID,
		"degree":            node.Degree,
		"affected_nodes":    affectedNodes,
		"network_connected": s.stats.IsConnected,
	}
}

func (s *TopologySimulator) updateState() {
	s.ClearNodeStates()
	s.SetGlobalData("topology_type", string(s.topologyType))
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("edge_count", len(s.edges))

	onlineCount := 0
	offlineCount := 0
	if s.stats != nil {
		s.SetGlobalData("avg_degree", s.stats.AvgDegree)
		s.SetGlobalData("is_connected", s.stats.IsConnected)
		s.SetGlobalData("topology_stats", s.stats)
	}

	for _, node := range s.nodes {
		if node.IsOnline {
			onlineCount++
		} else {
			offlineCount++
		}

		nodeType := "full"
		switch s.topologyType {
		case TopologyStar:
			if node.ID == "node-0" {
				nodeType = "hub"
			} else {
				nodeType = "leaf"
			}
		case TopologyTree:
			nodeType = "tree"
		}

		s.SetNodeState(types.NodeID(node.ID), &types.NodeState{
			ID:     types.NodeID(node.ID),
			Status: func() string { if node.IsOnline { return "online" }; return "offline" }(),
			Data: map[string]interface{}{
				"name":      node.ID,
				"type":      nodeType,
				"peers":     node.Connections,
				"latency":   40,
				"degree":    node.Degree,
				"address":   node.Address,
				"position":  node.Position,
				"is_online": node.IsOnline,
			},
		})
	}

	s.SetGlobalData("online_node_count", onlineCount)
	s.SetGlobalData("offline_node_count", offlineCount)

	stage := "topology_ready"
	summary := fmt.Sprintf("当前网络为 %s 拓扑，共有 %d 个节点、%d 条连边，在线节点 %d 个。", s.topologyType, len(s.nodes), len(s.edges), onlineCount)
	nextHint := "可以切换拓扑结构或让关键节点上下线，观察连通性与平均度的变化。"
	progress := 0.35
	if offlineCount > 0 {
		stage = "fault_impact"
		summary = fmt.Sprintf("当前有 %d 个节点离线，网络正在暴露拓扑脆弱性与传播路径变化。", offlineCount)
		nextHint = "重点观察节点是否失去邻居、网络是否仍然连通，以及哪些结构最容易受影响。"
		progress = 0.7
	}

	setNetworkTeachingState(
		s.BaseSimulator,
		"network",
		stage,
		summary,
		nextHint,
		progress,
		map[string]interface{}{
			"topology_type":    string(s.topologyType),
			"node_count":       len(s.nodes),
			"edge_count":       len(s.edges),
			"online_nodes":     onlineCount,
			"offline_nodes":    offlineCount,
			"is_connected":     s.stats != nil && s.stats.IsConnected,
			"average_degree":   func() float64 { if s.stats != nil { return s.stats.AvgDegree }; return 0 }(),
			"clustering_coeff": func() float64 { if s.stats != nil { return s.stats.ClusteringCoeff }; return 0 }(),
		},
	)
}

func (s *TopologySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "rebuild_topology":
		nodeCount := len(s.nodes)
		if raw, ok := params["node_count"].(float64); ok && int(raw) >= 3 {
			nodeCount = int(raw)
		}
		if raw, ok := params["topology_type"].(string); ok && raw != "" {
			s.topologyType = TopologyType(raw)
		}

		s.buildTopology(nodeCount)
		s.calculateStats()
		s.updateState()

		return networkActionResult(
			"已按当前参数重新生成网络拓扑。",
			map[string]interface{}{"topology_type": string(s.topologyType), "node_count": nodeCount},
			&types.ActionFeedback{
				Summary:     "拓扑结构已更新，节点连边和布局会随之变化。",
				NextHint:    "继续观察平均度、连通性以及不同拓扑下的传播差异。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"topology_type": string(s.topologyType), "node_count": nodeCount},
			},
		), nil
	case "toggle_node":
		target, _ := params["target"].(string)
		node, ok := s.nodes[target]
		if !ok {
			return &types.ActionResult{Success: false, Message: "目标节点不存在。"}, nil
		}

		node.IsOnline = !node.IsOnline
		s.EmitEvent("node_toggled", "", "", map[string]interface{}{"node_id": node.ID, "is_online": node.IsOnline})
		s.calculateStats()
		s.updateState()

		statusLabel := "离线"
		if node.IsOnline {
			statusLabel = "在线"
		}

		return networkActionResult(
			fmt.Sprintf("已将节点 %s 切换为%s状态。", node.ID, statusLabel),
			map[string]interface{}{"node_id": node.ID, "is_online": node.IsOnline, "degree": node.Degree},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("节点 %s 的在线状态已经变化，网络连通性和传播路径也会随之更新。", node.ID),
				NextHint:    "重点观察该节点邻居是否失去路径，以及整个网络是否仍然连通。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"node_id": node.ID, "is_online": node.IsOnline},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported topology action: %s", action)
	}
}

type TopologyFactory struct{}

func (f *TopologyFactory) Create() engine.Simulator {
	return NewTopologySimulator()
}

func (f *TopologyFactory) GetDescription() types.Description {
	return NewTopologySimulator().GetDescription()
}

func NewTopologyFactory() *TopologyFactory {
	return &TopologyFactory{}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ engine.SimulatorFactory = (*TopologyFactory)(nil)
