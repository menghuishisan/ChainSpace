package network

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type BlockData struct {
	Hash      string    `json:"hash"`
	Height    uint64    `json:"height"`
	Size      int       `json:"size_bytes"`
	TxCount   int       `json:"tx_count"`
	Miner     string    `json:"miner"`
	Timestamp time.Time `json:"timestamp"`
}

type PropagationNode struct {
	ID             string               `json:"id"`
	Region         string               `json:"region"`
	Latency        int                  `json:"latency_ms"`
	Bandwidth      int                  `json:"bandwidth_mbps"`
	ReceivedBlocks map[string]time.Time `json:"-"`
	Peers          []string             `json:"peers"`
}

type BlockPropagationStep struct {
	Time         time.Duration `json:"time_ms"`
	NodesReached int           `json:"nodes_reached"`
	Coverage     float64       `json:"coverage_percent"`
	NewNodes     []string      `json:"new_nodes"`
}

type BlockPropagationResult struct {
	Block            *BlockData               `json:"block"`
	StartTime        time.Time                `json:"start_time"`
	PropagationTimes map[string]time.Duration `json:"propagation_times"`
	Percentiles      map[string]time.Duration `json:"percentiles"`
	AvgTime          time.Duration            `json:"avg_time"`
	MaxTime          time.Duration            `json:"max_time"`
	Steps            []*BlockPropagationStep  `json:"steps"`
}

type BlockPropagationSimulator struct {
	*base.BaseSimulator
	nodes           map[string]*PropagationNode
	results         []*BlockPropagationResult
	propagationMode string
}

func NewBlockPropagationSimulator() *BlockPropagationSimulator {
	sim := &BlockPropagationSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"block_propagation",
			"区块传播演示器",
			"演示区块在 P2P 网络中的传播过程、延迟和传播优化策略。",
			"network",
			types.ComponentProcess,
		),
		nodes:   make(map[string]*PropagationNode),
		results: make([]*BlockPropagationResult, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "网络中的节点数量",
		Type:        types.ParamTypeInt,
		Default:     100,
		Min:         20,
		Max:         1000,
	})
	sim.AddParam(types.Param{
		Key:         "block_size",
		Name:        "区块大小",
		Description: "区块大小（KB）",
		Type:        types.ParamTypeInt,
		Default:     1000,
		Min:         100,
		Max:         10000,
	})
	sim.AddParam(types.Param{
		Key:         "propagation_mode",
		Name:        "传播模式",
		Description: "区块传播时采用的优化模式",
		Type:        types.ParamTypeSelect,
		Default:     "full_block",
		Options: []types.Option{
			{Label: "完整区块", Value: "full_block"},
			{Label: "区块头优先", Value: "headers_first"},
			{Label: "紧凑区块", Value: "compact_blocks"},
			{Label: "Graphene", Value: "graphene"},
		},
	})

	return sim
}

func (s *BlockPropagationSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	nodeCount := 100
	s.propagationMode = "full_block"
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}
	if v, ok := config.Params["propagation_mode"]; ok {
		if mode, ok := v.(string); ok && mode != "" {
			s.propagationMode = mode
		}
	}

	s.initializeNetwork(nodeCount)
	s.updateState()
	return nil
}

func (s *BlockPropagationSimulator) initializeNetwork(size int) {
	s.nodes = make(map[string]*PropagationNode)
	rand.Seed(time.Now().UnixNano())

	regions := []string{"US-East", "US-West", "EU-West", "EU-East", "Asia-Pacific", "South-America"}
	nodeIDs := make([]string, size)
	for i := 0; i < size; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		nodeIDs[i] = nodeID
		s.nodes[nodeID] = &PropagationNode{
			ID:             nodeID,
			Region:         regions[rand.Intn(len(regions))],
			Latency:        20 + rand.Intn(180),
			Bandwidth:      10 + rand.Intn(90),
			ReceivedBlocks: make(map[string]time.Time),
			Peers:          make([]string, 0),
		}
	}

	for _, node := range s.nodes {
		peerCount := 8 + rand.Intn(8)
		for len(node.Peers) < peerCount {
			peer := nodeIDs[rand.Intn(len(nodeIDs))]
			if peer != node.ID && !contains(node.Peers, peer) {
				node.Peers = append(node.Peers, peer)
			}
		}
	}
}

func (s *BlockPropagationSimulator) ExplainPropagationModes() []map[string]interface{} {
	return []map[string]interface{}{
		{"mode": "full_block", "name": "完整区块传播", "description": "直接广播整个区块数据", "bandwidth": "高", "latency": "高"},
		{"mode": "headers_first", "name": "区块头优先", "description": "先广播区块头，再按需拉取完整区块", "bandwidth": "中", "latency": "中"},
		{"mode": "compact_blocks", "name": "紧凑区块", "description": "仅发送区块头和短交易 ID，由接收方重建区块", "bandwidth": "低", "latency": "低"},
		{"mode": "graphene", "name": "Graphene", "description": "通过过滤器和集合编码进一步压缩传播数据", "bandwidth": "极低", "latency": "低"},
	}
}

func (s *BlockPropagationSimulator) SimulatePropagation(minerNode string, blockSizeKB int) *BlockPropagationResult {
	block := &BlockData{
		Hash:      fmt.Sprintf("0x%x", rand.Int63()),
		Height:    uint64(rand.Intn(1000000)),
		Size:      blockSizeKB * 1024,
		TxCount:   blockSizeKB / 250,
		Miner:     minerNode,
		Timestamp: time.Now(),
	}

	if minerNode == "" || s.nodes[minerNode] == nil {
		for id := range s.nodes {
			minerNode = id
			block.Miner = minerNode
			break
		}
	}

	result := &BlockPropagationResult{
		Block:            block,
		StartTime:        time.Now(),
		PropagationTimes: make(map[string]time.Duration),
		Percentiles:      make(map[string]time.Duration),
		Steps:            make([]*BlockPropagationStep, 0),
	}

	for _, node := range s.nodes {
		node.ReceivedBlocks = make(map[string]time.Time)
	}

	s.nodes[minerNode].ReceivedBlocks[block.Hash] = result.StartTime
	result.PropagationTimes[minerNode] = 0

	currentNodes := map[string]bool{minerNode: true}
	simulatedTime := time.Duration(0)

	for step := 0; step < 50 && len(result.PropagationTimes) < len(s.nodes); step++ {
		newNodes := make(map[string]bool)

		for nodeID := range currentNodes {
			node := s.nodes[nodeID]
			transmitTime := s.calculateTransmitTime(block.Size, s.propagationMode)
			for _, peerID := range node.Peers {
				peer := s.nodes[peerID]
				if _, received := peer.ReceivedBlocks[block.Hash]; !received {
					arrivalTime := simulatedTime + transmitTime + time.Duration(peer.Latency)*time.Millisecond
					peer.ReceivedBlocks[block.Hash] = result.StartTime.Add(arrivalTime)
					result.PropagationTimes[peerID] = arrivalTime
					newNodes[peerID] = true
				}
			}
		}

		if len(newNodes) > 0 {
			newNodesList := make([]string, 0, len(newNodes))
			for id := range newNodes {
				newNodesList = append(newNodesList, id)
			}
			result.Steps = append(result.Steps, &BlockPropagationStep{
				Time:         simulatedTime,
				NodesReached: len(result.PropagationTimes),
				Coverage:     float64(len(result.PropagationTimes)) / float64(len(s.nodes)) * 100,
				NewNodes:     newNodesList,
			})
		}

		currentNodes = newNodes
		simulatedTime += 100 * time.Millisecond
	}

	var values []time.Duration
	var totalTime time.Duration
	for _, t := range result.PropagationTimes {
		values = append(values, t)
		totalTime += t
		if t > result.MaxTime {
			result.MaxTime = t
		}
	}
	if len(values) > 0 {
		sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		result.AvgTime = totalTime / time.Duration(len(values))
		result.Percentiles["p50"] = values[len(values)/2]
		result.Percentiles["p90"] = values[min(len(values)-1, int(float64(len(values))*0.9))]
		result.Percentiles["p99"] = values[min(len(values)-1, int(float64(len(values))*0.99))]
	}

	s.EmitEvent("block_propagated", "", "", map[string]interface{}{
		"block_hash":    block.Hash[:16] + "...",
		"block_size":    fmt.Sprintf("%d KB", blockSizeKB),
		"mode":          s.propagationMode,
		"nodes_reached": len(result.PropagationTimes),
		"avg_time":      result.AvgTime.String(),
		"max_time":      result.MaxTime.String(),
	})

	s.results = append(s.results, result)
	s.updateState()
	return result
}

func (s *BlockPropagationSimulator) calculateTransmitTime(blockSize int, mode string) time.Duration {
	var dataSize int
	switch mode {
	case "full_block":
		dataSize = blockSize
	case "headers_first":
		dataSize = 80 + blockSize/10
	case "compact_blocks":
		dataSize = 80 + (blockSize/250)*6
	case "graphene":
		dataSize = 80 + blockSize/400
	default:
		dataSize = blockSize
	}
	transmitMs := float64(dataSize*8) / float64(10*1000*1000) * 1000
	return time.Duration(transmitMs) * time.Millisecond
}

func (s *BlockPropagationSimulator) ComparePropagationModes(blockSizeKB int) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	originalMode := s.propagationMode
	for _, mode := range []string{"full_block", "headers_first", "compact_blocks", "graphene"} {
		s.propagationMode = mode
		result := s.SimulatePropagation("", blockSizeKB)
		results = append(results, map[string]interface{}{
			"mode":     mode,
			"avg_time": result.AvgTime.Milliseconds(),
			"max_time": result.MaxTime.Milliseconds(),
			"coverage": fmt.Sprintf("%.1f%%", float64(len(result.PropagationTimes))/float64(len(s.nodes))*100),
		})
	}
	s.propagationMode = originalMode
	return results
}

func (s *BlockPropagationSimulator) updateState() {
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("propagation_mode", s.propagationMode)
	s.SetGlobalData("simulation_count", len(s.results))

	latestCoverage := 0.0
	latestAvg := int64(0)
	latestMax := int64(0)
	if len(s.results) > 0 {
		latest := s.results[len(s.results)-1]
		s.SetGlobalData("latest_result", latest)
		latestCoverage = float64(len(latest.PropagationTimes)) / float64(max(1, len(s.nodes))) * 100
		latestAvg = latest.AvgTime.Milliseconds()
		latestMax = latest.MaxTime.Milliseconds()
	}

	setNetworkTeachingState(
		s.BaseSimulator,
		"network",
		func() string {
			if len(s.results) > 0 {
				return "propagation_completed"
			}
			return "network_ready"
		}(),
		func() string {
			if len(s.results) > 0 {
				return fmt.Sprintf("最近一次区块传播采用 %s 模式，覆盖 %.1f%% 的节点。", s.propagationMode, latestCoverage)
			}
			return "当前网络已就绪，可以发起一轮区块传播实验。"
		}(),
		func() string {
			if len(s.results) > 0 {
				return "继续切换传播模式，对比平均传播时间、最大延迟和覆盖率差异。"
			}
			return "发起传播后，重点观察不同模式下的覆盖速度和延迟分布。"
		}(),
		func() float64 {
			if len(s.results) > 0 {
				return latestCoverage / 100
			}
			return 0.2
		}(),
		map[string]interface{}{
			"node_count":        len(s.nodes),
			"propagation_mode":  s.propagationMode,
			"simulation_count":  len(s.results),
			"latest_coverage":   latestCoverage,
			"latest_avg_time":   latestAvg,
			"latest_max_time":   latestMax,
		},
	)
}

func (s *BlockPropagationSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_propagation":
		blockSizeKB := 1000
		if raw, ok := params["block_size"].(float64); ok && int(raw) >= 100 {
			blockSizeKB = int(raw)
		}
		miner := ""
		if raw, ok := params["miner"].(string); ok {
			miner = raw
		}
		result := s.SimulatePropagation(miner, blockSizeKB)
		return networkActionResult(
			"已开始一轮区块传播实验。",
			map[string]interface{}{"miner": result.Block.Miner, "coverage": float64(len(result.PropagationTimes)) / float64(len(s.nodes)) * 100, "avg_time_ms": result.AvgTime.Milliseconds()},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("区块已从 %s 发出，当前模式为 %s。", result.Block.Miner, s.propagationMode),
				NextHint:    "观察不同传播模式下的覆盖速度、平均延迟和最长传播时间。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"miner": result.Block.Miner, "mode": s.propagationMode},
			},
		), nil
	case "reset_network":
		nodeCount := len(s.nodes)
		if raw, ok := params["node_count"].(float64); ok && int(raw) >= 20 {
			nodeCount = int(raw)
		}
		s.results = make([]*BlockPropagationResult, 0)
		s.initializeNetwork(nodeCount)
		s.updateState()
		return networkActionResult(
			"已重置区块传播网络。",
			map[string]interface{}{"node_count": nodeCount},
			&types.ActionFeedback{
				Summary:     "区块传播网络已恢复初始状态，之前的传播结果已清空。",
				NextHint:    "可以重新模拟传播，对比不同区块大小和传播模式的影响。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"node_count": nodeCount},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported block propagation action: %s", action)
	}
}

type BlockPropagationFactory struct{}

func (f *BlockPropagationFactory) Create() engine.Simulator {
	return NewBlockPropagationSimulator()
}

func (f *BlockPropagationFactory) GetDescription() types.Description {
	return NewBlockPropagationSimulator().GetDescription()
}

func NewBlockPropagationFactory() *BlockPropagationFactory {
	return &BlockPropagationFactory{}
}

var _ engine.SimulatorFactory = (*BlockPropagationFactory)(nil)
