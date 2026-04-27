package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type DiscoveryNode struct {
	ID            string    `json:"id"`
	Address       string    `json:"address"`
	BootstrapNode bool      `json:"bootstrap_node"`
	KnownPeers    []string  `json:"known_peers"`
	JoinTime      time.Time `json:"join_time"`
	Status        string    `json:"status"`
}

type DiscoveryStep struct {
	Step        int      `json:"step"`
	Action      string   `json:"action"`
	FromNode    string   `json:"from_node"`
	ToNode      string   `json:"to_node"`
	PeersFound  []string `json:"peers_found"`
	Description string   `json:"description"`
}

type DiscoverySession struct {
	NewNode       string           `json:"new_node"`
	BootstrapUsed string           `json:"bootstrap_used"`
	Steps         []*DiscoveryStep `json:"steps"`
	FinalPeers    []string         `json:"final_peers"`
	Duration      time.Duration    `json:"duration"`
}

type DiscoverySimulator struct {
	*base.BaseSimulator
	nodes          map[string]*DiscoveryNode
	bootstrapNodes []string
	sessions       []*DiscoverySession
}

func NewDiscoverySimulator() *DiscoverySimulator {
	sim := &DiscoverySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"discovery",
			"节点发现演示器",
			"演示新节点如何通过 bootstrap、邻居交换与持续发现加入 P2P 网络。",
			"network",
			types.ComponentProcess,
		),
		nodes:          make(map[string]*DiscoveryNode),
		bootstrapNodes: make([]string, 0),
		sessions:       make([]*DiscoverySession, 0),
	}

	sim.AddParam(types.Param{
		Key:         "network_size",
		Name:        "网络规模",
		Description: "当前网络中的节点数量",
		Type:        types.ParamTypeInt,
		Default:     50,
		Min:         10,
		Max:         500,
	})

	sim.AddParam(types.Param{
		Key:         "bootstrap_count",
		Name:        "Bootstrap 节点数",
		Description: "负责提供入口连接的 bootstrap 节点数量",
		Type:        types.ParamTypeInt,
		Default:     3,
		Min:         1,
		Max:         10,
	})

	return sim
}

func (s *DiscoverySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	networkSize := 50
	bootstrapCount := 3
	if v, ok := config.Params["network_size"]; ok {
		if n, ok := v.(float64); ok {
			networkSize = int(n)
		}
	}
	if v, ok := config.Params["bootstrap_count"]; ok {
		if n, ok := v.(float64); ok {
			bootstrapCount = int(n)
		}
	}

	s.initializeNetwork(networkSize, bootstrapCount)
	s.updateState()
	return nil
}

func (s *DiscoverySimulator) initializeNetwork(size, bootstrapCount int) {
	s.nodes = make(map[string]*DiscoveryNode)
	s.bootstrapNodes = make([]string, 0)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < size; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		isBootstrap := i < bootstrapCount

		node := &DiscoveryNode{
			ID:            nodeID,
			Address:       fmt.Sprintf("192.168.1.%d:30303", i+1),
			BootstrapNode: isBootstrap,
			KnownPeers:    make([]string, 0),
			JoinTime:      time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second),
			Status:        "synced",
		}
		s.nodes[nodeID] = node
		if isBootstrap {
			s.bootstrapNodes = append(s.bootstrapNodes, nodeID)
		}
	}

	nodeIDs := make([]string, 0, len(s.nodes))
	for id := range s.nodes {
		nodeIDs = append(nodeIDs, id)
	}

	for _, node := range s.nodes {
		peerCount := 8 + rand.Intn(8)
		maxPeers := len(nodeIDs) - 1
		if peerCount > maxPeers {
			peerCount = maxPeers
		}
		for len(node.KnownPeers) < peerCount {
			peer := nodeIDs[rand.Intn(len(nodeIDs))]
			if peer != node.ID && !contains(node.KnownPeers, peer) {
				node.KnownPeers = append(node.KnownPeers, peer)
			}
		}
	}
}

func (s *DiscoverySimulator) ExplainDiscovery() map[string]interface{} {
	return map[string]interface{}{
		"overview": "新节点加入 P2P 网络之前，必须先找到入口节点，再逐步扩展自己的邻居集合。",
		"methods": []map[string]interface{}{
			{
				"method":      "Bootstrap 节点",
				"description": "客户端内置的一组入口节点，负责提供第一跳邻居信息。",
				"pros":        []string{"简单可靠", "接入路径清晰"},
				"cons":        []string{"可能带来中心化入口", "入口节点可能被封锁"},
				"examples":    []string{"以太坊 bootstrap 节点", "比特币 DNS 种子"},
			},
			{
				"method":      "DNS 种子",
				"description": "通过 DNS 查询获得一组可连接节点。",
				"pros":        []string{"维护成本低", "更新方便"},
				"cons":        []string{"依赖 DNS 基础设施", "可能遭受污染"},
				"examples":    []string{"seed.bitcoin.sipa.be", "比特币 DNS 种子"},
			},
			{
				"method":      "本地发现（mDNS）",
				"description": "通过局域网广播和多播 DNS 发现附近节点。",
				"pros":        []string{"无需外部服务", "局域网内延迟低"},
				"cons":        []string{"只适用于局域网环境"},
				"examples":    []string{"IPFS 本地发现", "libp2p mDNS"},
			},
			{
				"method":      "邻居交换",
				"description": "从已连接节点处继续请求更多节点信息。",
				"pros":        []string{"去中心化", "可持续扩展"},
				"cons":        []string{"需要先有初始连接"},
				"examples":    []string{"Kademlia FIND_NODE", "以太坊 RLPx 邻居交换"},
			},
		},
		"typical_flow": []string{
			"1. 新节点启动并读取预设的 bootstrap 节点列表",
			"2. 与一个 bootstrap 节点建立连接",
			"3. 请求该节点提供已知邻居列表",
			"4. 逐步连接更多邻居并继续扩展",
			"5. 达到目标连接规模后进入稳定同步状态",
		},
	}
}

func (s *DiscoverySimulator) SimulateNewNodeJoin() *DiscoverySession {
	newNodeID := fmt.Sprintf("new-node-%d", len(s.sessions))
	newNode := &DiscoveryNode{
		ID:            newNodeID,
		Address:       fmt.Sprintf("10.0.0.%d:30303", len(s.sessions)+1),
		BootstrapNode: false,
		KnownPeers:    make([]string, 0),
		JoinTime:      time.Now(),
		Status:        "discovering",
	}

	session := &DiscoverySession{
		NewNode: newNodeID,
		Steps:   make([]*DiscoveryStep, 0),
	}
	startTime := time.Now()

	bootstrapNode := s.bootstrapNodes[rand.Intn(len(s.bootstrapNodes))]
	session.BootstrapUsed = bootstrapNode

	session.Steps = append(session.Steps, &DiscoveryStep{
		Step:        1,
		Action:      "connect_bootstrap",
		FromNode:    newNodeID,
		ToNode:      bootstrapNode,
		Description: fmt.Sprintf("连接到 bootstrap 节点 %s", bootstrapNode),
	})
	s.EmitEvent("discovery_step", "", "", map[string]interface{}{
		"step":   1,
		"action": "connect_bootstrap",
		"target": bootstrapNode,
	})

	bootstrapPeers := s.nodes[bootstrapNode].KnownPeers
	initialPeers := bootstrapPeers[:min(5, len(bootstrapPeers))]
	session.Steps = append(session.Steps, &DiscoveryStep{
		Step:        2,
		Action:      "get_peers",
		FromNode:    newNodeID,
		ToNode:      bootstrapNode,
		PeersFound:  initialPeers,
		Description: fmt.Sprintf("从 %s 获取到 %d 个候选邻居", bootstrapNode, len(initialPeers)),
	})

	discovered := map[string]bool{bootstrapNode: true}
	toQuery := append([]string(nil), initialPeers...)
	stepNum := 3

	for len(toQuery) > 0 && stepNum < 10 {
		currentPeer := toQuery[0]
		toQuery = toQuery[1:]
		if discovered[currentPeer] {
			continue
		}
		discovered[currentPeer] = true

		peer, ok := s.nodes[currentPeer]
		if !ok {
			continue
		}

		newPeers := make([]string, 0)
		for _, candidate := range peer.KnownPeers {
			if !discovered[candidate] {
				newPeers = append(newPeers, candidate)
			}
		}
		if len(newPeers) > 3 {
			newPeers = newPeers[:3]
		}

		session.Steps = append(session.Steps, &DiscoveryStep{
			Step:        stepNum,
			Action:      "query_peer",
			FromNode:    newNodeID,
			ToNode:      currentPeer,
			PeersFound:  newPeers,
			Description: fmt.Sprintf("查询节点 %s，继续发现 %d 个新邻居", currentPeer, len(newPeers)),
		})

		toQuery = append(toQuery, newPeers...)
		newNode.KnownPeers = append(newNode.KnownPeers, currentPeer)
		stepNum++
	}

	newNode.Status = "connected"
	session.FinalPeers = append([]string(nil), newNode.KnownPeers...)
	session.Duration = time.Since(startTime)

	s.nodes[newNodeID] = newNode
	s.sessions = append(s.sessions, session)

	s.EmitEvent("node_joined", "", "", map[string]interface{}{
		"node_id":     newNodeID,
		"peers_found": len(session.FinalPeers),
		"steps":       len(session.Steps),
	})

	s.updateState()
	return session
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *DiscoverySimulator) GetBootstrapNodes() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(s.bootstrapNodes))
	for _, nodeID := range s.bootstrapNodes {
		node := s.nodes[nodeID]
		result = append(result, map[string]interface{}{
			"id":          node.ID,
			"address":     node.Address,
			"known_peers": len(node.KnownPeers),
		})
	}
	return result
}

func (s *DiscoverySimulator) GetRealWorldExamples() []map[string]interface{} {
	return []map[string]interface{}{
		{"network": "以太坊", "protocol": "discv4 / discv5", "bootstrap": "预置 enode URL", "discovery": "Kademlia 变体"},
		{"network": "比特币", "protocol": "addr 消息", "bootstrap": "DNS 种子", "discovery": "通过邻居交换已知节点"},
		{"network": "IPFS / libp2p", "protocol": "Kademlia DHT", "bootstrap": "IPFS Bootstrap 节点", "discovery": "mDNS + DHT + 中继"},
	}
}

func (s *DiscoverySimulator) updateState() {
	s.ClearNodeStates()
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("bootstrap_count", len(s.bootstrapNodes))
	s.SetGlobalData("session_count", len(s.sessions))
	s.SetGlobalData("bootstrap_nodes", s.GetBootstrapNodes())

	var latestSession *DiscoverySession
	if len(s.sessions) > 0 {
		latestSession = s.sessions[len(s.sessions)-1]
		s.SetGlobalData("latest_session", latestSession)
	}

	connectedCount := 0
	discoveringCount := 0
	for _, node := range s.nodes {
		if node.Status == "connected" || node.Status == "synced" {
			connectedCount++
		}
		if node.Status == "discovering" {
			discoveringCount++
		}

		nodeType := "full"
		if node.BootstrapNode {
			nodeType = "bootstrap"
		}

		s.SetNodeState(types.NodeID(node.ID), &types.NodeState{
			ID:     types.NodeID(node.ID),
			Status: node.Status,
			Data: map[string]interface{}{
				"name":           node.ID,
				"type":           nodeType,
				"peers":          node.KnownPeers,
				"latency":        35,
				"address":        node.Address,
				"bootstrap_node": node.BootstrapNode,
				"peer_count":     len(node.KnownPeers),
			},
		})
	}

	stage := "network_ready"
	summary := fmt.Sprintf("当前网络共有 %d 个节点，其中 %d 个为 bootstrap 节点。", len(s.nodes), len(s.bootstrapNodes))
	nextHint := "可以模拟一个新节点加入，观察它如何借助 bootstrap 和邻居交换逐步完成接入。"
	progress := 0.25
	result := map[string]interface{}{
		"node_count":        len(s.nodes),
		"bootstrap_count":   len(s.bootstrapNodes),
		"session_count":     len(s.sessions),
		"connected_nodes":   connectedCount,
		"discovering_nodes": discoveringCount,
	}

	if latestSession != nil {
		stage = "join_completed"
		summary = fmt.Sprintf("最近一次加入流程中，新节点 %s 通过 %s 完成接入，共发现 %d 个邻居。", latestSession.NewNode, latestSession.BootstrapUsed, len(latestSession.FinalPeers))
		nextHint = "继续观察发现步骤如何扩展邻居集合，以及 bootstrap 对初始接入速度的影响。"
		progress = 0.85
		result["latest_node"] = latestSession.NewNode
		result["latest_peers"] = len(latestSession.FinalPeers)
		result["latest_steps"] = len(latestSession.Steps)
		result["latest_duration_ms"] = latestSession.Duration.Milliseconds()
	}

	setNetworkTeachingState(s.BaseSimulator, "network", stage, summary, nextHint, progress, result)
}

func (s *DiscoverySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_join":
		session := s.SimulateNewNodeJoin()
		return networkActionResult(
			"已开始一轮新的节点发现流程。",
			map[string]interface{}{"new_node": session.NewNode, "steps": len(session.Steps)},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("新节点 %s 已通过 bootstrap 入口发起接入。", session.NewNode),
				NextHint:    "观察它如何从第一跳邻居继续扩展，直到形成稳定连接。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"new_node": session.NewNode, "steps": len(session.Steps)},
			},
		), nil
	case "reset_network":
		networkSize := len(s.nodes)
		bootstrapCount := len(s.bootstrapNodes)
		if raw, ok := params["network_size"].(float64); ok && int(raw) >= 1 {
			networkSize = int(raw)
		}
		if raw, ok := params["bootstrap_count"].(float64); ok && int(raw) >= 1 {
			bootstrapCount = int(raw)
		}

		s.sessions = make([]*DiscoverySession, 0)
		s.initializeNetwork(networkSize, bootstrapCount)
		s.updateState()

		return networkActionResult(
			"已重置当前节点发现网络。",
			map[string]interface{}{"network_size": networkSize, "bootstrap_count": bootstrapCount, "latest_session_n": 0},
			&types.ActionFeedback{
				Summary:     "节点发现网络已恢复到初始状态，之前的加入会话和过程数据已清空。",
				NextHint:    "可以重新模拟新节点加入，对比不同 bootstrap 数量下的接入路径差异。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"network_size": networkSize, "bootstrap_count": bootstrapCount},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported discovery action: %s", action)
	}
}

type DiscoveryFactory struct{}

func (f *DiscoveryFactory) Create() engine.Simulator {
	return NewDiscoverySimulator()
}

func (f *DiscoveryFactory) GetDescription() types.Description {
	return NewDiscoverySimulator().GetDescription()
}

func NewDiscoveryFactory() *DiscoveryFactory {
	return &DiscoveryFactory{}
}

var _ engine.SimulatorFactory = (*DiscoveryFactory)(nil)
