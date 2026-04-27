package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type GossipMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Origin    string    `json:"origin"`
	TTL       int       `json:"ttl"`
	Timestamp time.Time `json:"timestamp"`
}

type GossipNode struct {
	ID           string          `json:"id"`
	Peers        []string        `json:"peers"`
	ReceivedMsgs map[string]bool `json:"received_msgs"`
	ForwardCount int             `json:"forward_count"`
}

type PropagationStep struct {
	Round        int      `json:"round"`
	NodesReached int      `json:"nodes_reached"`
	NewNodes     []string `json:"new_nodes"`
	TotalNodes   int      `json:"total_nodes"`
	Coverage     float64  `json:"coverage"`
}

type GossipSimulation struct {
	Message       *GossipMessage     `json:"message"`
	Steps         []*PropagationStep `json:"steps"`
	TotalRounds   int                `json:"total_rounds"`
	FinalCoverage float64            `json:"final_coverage"`
	AvgLatency    float64            `json:"avg_latency_ms"`
}

type GossipSimulator struct {
	*base.BaseSimulator
	nodes       map[string]*GossipNode
	fanout      int
	ttl         int
	simulations []*GossipSimulation
}

func NewGossipSimulator() *GossipSimulator {
	sim := &GossipSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"gossip",
			"Gossip 协议演示器",
			"演示消息如何通过 gossip 机制在 P2P 网络中扩散。",
			"network",
			types.ComponentProcess,
		),
		nodes:       make(map[string]*GossipNode),
		simulations: make([]*GossipSimulation, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "网络中的节点数量",
		Type:        types.ParamTypeInt,
		Default:     100,
		Min:         10,
		Max:         1000,
	})
	sim.AddParam(types.Param{
		Key:         "fanout",
		Name:        "扇出",
		Description: "每一轮向多少个邻居转发消息",
		Type:        types.ParamTypeInt,
		Default:     6,
		Min:         1,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "ttl",
		Name:        "TTL",
		Description: "消息在网络中的最大传播轮数",
		Type:        types.ParamTypeInt,
		Default:     10,
		Min:         1,
		Max:         50,
	})

	return sim
}

func (s *GossipSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	nodeCount := 100
	s.fanout = 6
	s.ttl = 10

	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}
	if v, ok := config.Params["fanout"]; ok {
		if n, ok := v.(float64); ok {
			s.fanout = int(n)
		}
	}
	if v, ok := config.Params["ttl"]; ok {
		if n, ok := v.(float64); ok {
			s.ttl = int(n)
		}
	}

	s.initializeNetwork(nodeCount)
	s.updateState()
	return nil
}

func (s *GossipSimulator) initializeNetwork(nodeCount int) {
	s.nodes = make(map[string]*GossipNode)
	rand.Seed(time.Now().UnixNano())

	nodeIDs := make([]string, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		nodeIDs[i] = nodeID
		s.nodes[nodeID] = &GossipNode{
			ID:           nodeID,
			Peers:        make([]string, 0),
			ReceivedMsgs: make(map[string]bool),
			ForwardCount: 0,
		}
	}

	for _, node := range s.nodes {
		peerCount := 10 + rand.Intn(11)
		for len(node.Peers) < peerCount {
			peer := nodeIDs[rand.Intn(len(nodeIDs))]
			if peer != node.ID && !contains(node.Peers, peer) {
				node.Peers = append(node.Peers, peer)
				if !contains(s.nodes[peer].Peers, node.ID) {
					s.nodes[peer].Peers = append(s.nodes[peer].Peers, node.ID)
				}
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *GossipSimulator) ExplainGossip() map[string]interface{} {
	return map[string]interface{}{
		"name":    "Gossip 协议",
		"analogy": "像流言或病毒一样逐轮扩散消息",
		"mechanism": []string{
			"1. 某个节点收到一条新消息",
			"2. 它随机选择若干邻居进行转发",
			"3. 新收到消息的节点继续重复该过程",
			"4. 直到消息 TTL 耗尽或覆盖大部分网络",
		},
		"variants": []map[string]string{
			{"name": "Push", "description": "主动推送给邻居"},
			{"name": "Pull", "description": "主动向邻居拉取最新消息"},
			{"name": "Push-Pull", "description": "结合推送和拉取，加快收敛"},
		},
		"applications": []string{"区块和交易传播", "成员状态同步", "故障探测", "分布式元数据同步"},
	}
}

func (s *GossipSimulator) SimulatePropagation(originNode string) *GossipSimulation {
	for _, node := range s.nodes {
		node.ReceivedMsgs = make(map[string]bool)
		node.ForwardCount = 0
	}

	if originNode == "" || s.nodes[originNode] == nil {
		for id := range s.nodes {
			originNode = id
			break
		}
	}

	msg := &GossipMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Content:   "Test gossip message",
		Origin:    originNode,
		TTL:       s.ttl,
		Timestamp: time.Now(),
	}

	simulation := &GossipSimulation{
		Message: msg,
		Steps:   make([]*PropagationStep, 0),
	}

	s.nodes[originNode].ReceivedMsgs[msg.ID] = true
	currentInfected := map[string]bool{originNode: true}
	totalInfected := 1

	for round := 0; round < s.ttl; round++ {
		newInfected := make(map[string]bool)

		for nodeID := range currentInfected {
			node := s.nodes[nodeID]
			peers := s.selectRandomPeers(node.Peers, s.fanout)
			for _, peerID := range peers {
				peer := s.nodes[peerID]
				if !peer.ReceivedMsgs[msg.ID] {
					peer.ReceivedMsgs[msg.ID] = true
					newInfected[peerID] = true
					totalInfected++
				}
			}
			node.ForwardCount++
		}

		newNodesList := make([]string, 0, len(newInfected))
		for id := range newInfected {
			newNodesList = append(newNodesList, id)
		}

		step := &PropagationStep{
			Round:        round + 1,
			NodesReached: totalInfected,
			NewNodes:     newNodesList,
			TotalNodes:   len(s.nodes),
			Coverage:     float64(totalInfected) / float64(len(s.nodes)) * 100,
		}
		simulation.Steps = append(simulation.Steps, step)

		s.EmitEvent("gossip_round", "", "", map[string]interface{}{
			"round":         round + 1,
			"new_nodes":     len(newInfected),
			"total_reached": totalInfected,
			"coverage":      fmt.Sprintf("%.1f%%", step.Coverage),
		})

		if len(newInfected) == 0 {
			break
		}
		currentInfected = newInfected
	}

	simulation.TotalRounds = len(simulation.Steps)
	simulation.FinalCoverage = float64(totalInfected) / float64(len(s.nodes)) * 100
	simulation.AvgLatency = float64(simulation.TotalRounds) * 100

	s.simulations = append(s.simulations, simulation)
	s.updateState()
	return simulation
}

func (s *GossipSimulator) selectRandomPeers(peers []string, count int) []string {
	if len(peers) <= count {
		return peers
	}
	selected := make([]string, 0, count)
	indices := rand.Perm(len(peers))
	for i := 0; i < count && i < len(indices); i++ {
		selected = append(selected, peers[indices[i]])
	}
	return selected
}

func (s *GossipSimulator) CompareFanout() []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	originalFanout := s.fanout

	for _, fanout := range []int{2, 4, 6, 8, 10} {
		s.fanout = fanout
		sim := s.SimulatePropagation("")
		results = append(results, map[string]interface{}{
			"fanout":            fanout,
			"rounds":            sim.TotalRounds,
			"final_coverage":    fmt.Sprintf("%.1f%%", sim.FinalCoverage),
			"messages_per_node": fanout * sim.TotalRounds,
		})
	}

	s.fanout = originalFanout
	return results
}

func (s *GossipSimulator) GetRealWorldExamples() []map[string]interface{} {
	return []map[string]interface{}{
		{"protocol": "Bitcoin", "usage": "交易和区块传播", "details": "通过 inv/getdata 避免重复发送完整数据"},
		{"protocol": "Ethereum", "usage": "交易和区块传播", "details": "devp2p 协议支持更高效的块头和交易扩散"},
		{"protocol": "Cassandra", "usage": "节点状态同步", "details": "使用 gossip 持续同步成员状态并检测故障"},
		{"protocol": "Hyperledger Fabric", "usage": "私有数据分发", "details": "利用 gossip 在组织内同步区块和状态"},
	}
}

func (s *GossipSimulator) updateState() {
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("fanout", s.fanout)
	s.SetGlobalData("ttl", s.ttl)
	s.SetGlobalData("simulation_count", len(s.simulations))

	latestCoverage := 0.0
	latestRounds := 0
	if len(s.simulations) > 0 {
		latest := s.simulations[len(s.simulations)-1]
		s.SetGlobalData("latest_simulation", latest)
		latestCoverage = latest.FinalCoverage
		latestRounds = latest.TotalRounds
	}

	setNetworkTeachingState(
		s.BaseSimulator,
		"network",
		func() string {
			if len(s.simulations) > 0 {
				return "propagation_completed"
			}
			return "network_ready"
		}(),
		func() string {
			if len(s.simulations) > 0 {
				return fmt.Sprintf("最近一次 gossip 扩散在 %d 轮内覆盖了 %.1f%% 的节点。", latestRounds, latestCoverage)
			}
			return "当前网络已经就绪，可以发起一次 gossip 扩散过程。"
		}(),
		func() string {
			if len(s.simulations) > 0 {
				return "继续调整 fanout 或 TTL，对比扩散轮数、覆盖率和消息开销的变化。"
			}
			return "发起一次传播，观察消息如何从源节点逐轮扩散到更多邻居。"
		}(),
		func() float64 {
			if len(s.simulations) > 0 {
				return latestCoverage / 100
			}
			return 0.2
		}(),
		map[string]interface{}{
			"node_count":       len(s.nodes),
			"fanout":           s.fanout,
			"ttl":              s.ttl,
			"simulation_count": len(s.simulations),
			"latest_coverage":  latestCoverage,
			"latest_rounds":    latestRounds,
		},
	)
}

func (s *GossipSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_propagation":
		origin := ""
		if raw, ok := params["origin"].(string); ok {
			origin = raw
		}
		result := s.SimulatePropagation(origin)
		return networkActionResult(
			"已启动一轮 gossip 消息传播。",
			map[string]interface{}{"origin": result.Message.Origin, "rounds": result.TotalRounds, "coverage": result.FinalCoverage},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("消息已从 %s 发起，并在 %d 轮内完成当前扩散。", result.Message.Origin, result.TotalRounds),
				NextHint:    "观察每一轮新增节点数如何变化，以及 fanout 对覆盖速度的影响。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"origin": result.Message.Origin, "coverage": result.FinalCoverage},
			},
		), nil
	case "reset_network":
		nodeCount := len(s.nodes)
		if raw, ok := params["node_count"].(float64); ok && int(raw) >= 10 {
			nodeCount = int(raw)
		}
		s.simulations = make([]*GossipSimulation, 0)
		s.initializeNetwork(nodeCount)
		s.updateState()
		return networkActionResult(
			"已重置 gossip 网络场景。",
			map[string]interface{}{"node_count": nodeCount},
			&types.ActionFeedback{
				Summary:     "网络已经恢复到初始状态，之前的传播记录已清空。",
				NextHint:    "可以重新发起扩散，对比不同 fanout 或 TTL 设置的影响。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"node_count": nodeCount},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported gossip action: %s", action)
	}
}

type GossipFactory struct{}

func (f *GossipFactory) Create() engine.Simulator {
	return NewGossipSimulator()
}

func (f *GossipFactory) GetDescription() types.Description {
	return NewGossipSimulator().GetDescription()
}

func NewGossipFactory() *GossipFactory {
	return &GossipFactory{}
}

var _ engine.SimulatorFactory = (*GossipFactory)(nil)
