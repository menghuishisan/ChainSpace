package consensus

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

type DAGVertex struct {
	Hash       string       `json:"hash"`
	Parents    []string     `json:"parents"`
	Creator    types.NodeID `json:"creator"`
	Round      uint64       `json:"round"`
	Timestamp  time.Time    `json:"timestamp"`
	Weight     uint64       `json:"weight"`
	Confirmed  bool         `json:"confirmed"`
	References []string     `json:"references"`
}

type DAGNode struct {
	ID             types.NodeID          `json:"id"`
	LocalDAG       map[string]*DAGVertex `json:"local_dag"`
	Tips           []string              `json:"tips"`
	Round          uint64                `json:"round"`
	VertexCount    int                   `json:"vertex_count"`
	ConfirmedCount int                   `json:"confirmed_count"`
}

type DAGSimulator struct {
	*base.BaseSimulator
	mu             sync.RWMutex
	nodes          map[types.NodeID]*DAGNode
	nodeList       []types.NodeID
	globalDAG      map[string]*DAGVertex
	tips           []string
	currentRound   uint64
	confirmedCount int
	tipSelection   string
	minParents     int
	maxParents     int
}

func NewDAGSimulator() *DAGSimulator {
	sim := &DAGSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"dag",
			"DAG 共识演示器",
			"模拟 DAG 并行出块、父顶点选择和确认累积过程。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:     make(map[types.NodeID]*DAGNode),
		nodeList:  make([]types.NodeID, 0),
		globalDAG: make(map[string]*DAGVertex),
		tips:      make([]string, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "参与 DAG 共识的节点数量。",
		Type:        types.ParamTypeInt,
		Default:     6,
		Min:         3,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "tip_selection",
		Name:        "Tip 选择算法",
		Description: "用于挑选父顶点的算法。",
		Type:        types.ParamTypeSelect,
		Default:     "weighted",
		Options: []types.Option{
			{Label: "随机选择", Value: "random"},
			{Label: "加权随机", Value: "weighted"},
			{Label: "MCMC", Value: "mcmc"},
		},
	})
	sim.AddParam(types.Param{
		Key:         "min_parents",
		Name:        "最少父顶点数",
		Description: "每个新顶点至少引用多少个父顶点。",
		Type:        types.ParamTypeInt,
		Default:     2,
		Min:         1,
		Max:         5,
	})
	sim.AddParam(types.Param{
		Key:         "max_parents",
		Name:        "最多父顶点数",
		Description: "每个新顶点最多引用多少个父顶点。",
		Type:        types.ParamTypeInt,
		Default:     4,
		Min:         2,
		Max:         8,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *DAGSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	nodeCount := 6
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}
	s.tipSelection = "weighted"
	if v, ok := config.Params["tip_selection"]; ok {
		if str, ok := v.(string); ok {
			s.tipSelection = str
		}
	}
	s.minParents = 2
	if v, ok := config.Params["min_parents"]; ok {
		if n, ok := v.(float64); ok {
			s.minParents = int(n)
		}
	}
	s.maxParents = 4
	if v, ok := config.Params["max_parents"]; ok {
		if n, ok := v.(float64); ok {
			s.maxParents = int(n)
		}
	}

	s.nodes = make(map[types.NodeID]*DAGNode)
	s.nodeList = make([]types.NodeID, 0, nodeCount)
	s.globalDAG = make(map[string]*DAGVertex)
	s.currentRound = 0
	s.confirmedCount = 0

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("node-%d", i))
		node := &DAGNode{
			ID:       nodeID,
			LocalDAG: make(map[string]*DAGVertex),
			Tips:     make([]string, 0),
			Round:    0,
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
	}

	genesis := &DAGVertex{
		Hash:      "genesis",
		Parents:   []string{},
		Round:     0,
		Timestamp: time.Now(),
		Weight:    1,
		Confirmed: true,
	}
	s.globalDAG[genesis.Hash] = genesis
	s.tips = []string{genesis.Hash}
	s.confirmedCount = 1

	for _, node := range s.nodes {
		node.LocalDAG[genesis.Hash] = genesis
		node.Tips = []string{genesis.Hash}
	}

	s.updateAllStates()
	return nil
}

func (s *DAGSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentRound = tick
	for _, nodeID := range s.nodeList {
		if rand.Float64() < 0.3 {
			s.createVertex(nodeID)
		}
	}
	if tick%5 == 0 {
		s.updateConfirmations()
	}

	s.updateAllStates()
	return nil
}

func (s *DAGSimulator) createVertex(creatorID types.NodeID) {
	node := s.nodes[creatorID]
	if node == nil {
		return
	}

	parents := s.selectParents()
	if len(parents) < s.minParents {
		return
	}

	vertex := &DAGVertex{
		Hash:      fmt.Sprintf("v-%d-%s", len(s.globalDAG), uuid.New().String()[:8]),
		Parents:   parents,
		Creator:   creatorID,
		Round:     s.currentRound,
		Timestamp: time.Now(),
		Weight:    1,
		Confirmed: false,
	}
	for _, parentHash := range parents {
		if parent := s.globalDAG[parentHash]; parent != nil {
			vertex.Weight += parent.Weight
		}
	}

	s.globalDAG[vertex.Hash] = vertex
	node.LocalDAG[vertex.Hash] = vertex
	node.VertexCount++
	node.Round = s.currentRound
	s.updateTips(vertex)

	s.EmitEvent("vertex_created", creatorID, "", map[string]interface{}{
		"hash":         vertex.Hash,
		"parents":      len(parents),
		"weight":       vertex.Weight,
		"total_vertex": len(s.globalDAG),
	})
}

func (s *DAGSimulator) selectParents() []string {
	if len(s.tips) == 0 {
		return nil
	}

	numParents := s.minParents
	if s.maxParents > s.minParents {
		numParents += rand.Intn(s.maxParents-s.minParents+1)
	}
	if numParents > len(s.tips) {
		numParents = len(s.tips)
	}

	switch s.tipSelection {
	case "random":
		return s.randomSelect(numParents)
	case "weighted":
		return s.weightedSelect(numParents)
	case "mcmc":
		return s.mcmcSelect(numParents)
	default:
		return s.randomSelect(numParents)
	}
}

func (s *DAGSimulator) randomSelect(n int) []string {
	if n >= len(s.tips) {
		return append([]string{}, s.tips...)
	}
	indices := rand.Perm(len(s.tips))[:n]
	selected := make([]string, n)
	for i, idx := range indices {
		selected[i] = s.tips[idx]
	}
	return selected
}

func (s *DAGSimulator) weightedSelect(n int) []string {
	if len(s.tips) == 0 {
		return nil
	}

	totalWeight := uint64(0)
	for _, tipHash := range s.tips {
		if vertex := s.globalDAG[tipHash]; vertex != nil {
			totalWeight += vertex.Weight
		}
	}

	selected := make([]string, 0, n)
	usedTips := make(map[string]bool)
	for len(selected) < n && len(selected) < len(s.tips) {
		target := rand.Uint64() % totalWeight
		cumulative := uint64(0)
		for _, tipHash := range s.tips {
			if usedTips[tipHash] {
				continue
			}
			if vertex := s.globalDAG[tipHash]; vertex != nil {
				cumulative += vertex.Weight
				if target < cumulative {
					selected = append(selected, tipHash)
					usedTips[tipHash] = true
					break
				}
			}
		}
	}
	return selected
}

func (s *DAGSimulator) mcmcSelect(n int) []string {
	if len(s.tips) == 0 {
		return nil
	}

	selected := make([]string, 0, n)
	current := s.tips[rand.Intn(len(s.tips))]
	for i := 0; i < n*10 && len(selected) < n; i++ {
		next := s.tips[rand.Intn(len(s.tips))]

		currentWeight := uint64(1)
		nextWeight := uint64(1)
		if vertex := s.globalDAG[current]; vertex != nil {
			currentWeight = vertex.Weight
		}
		if vertex := s.globalDAG[next]; vertex != nil {
			nextWeight = vertex.Weight
		}

		if nextWeight >= currentWeight || rand.Float64() < float64(nextWeight)/float64(currentWeight) {
			current = next
		}

		if i%10 == 0 {
			exists := false
			for _, item := range selected {
				if item == current {
					exists = true
					break
				}
			}
			if !exists {
				selected = append(selected, current)
			}
		}
	}
	return selected
}

func (s *DAGSimulator) updateTips(newVertex *DAGVertex) {
	parentSet := make(map[string]bool)
	for _, parent := range newVertex.Parents {
		parentSet[parent] = true
	}

	newTips := make([]string, 0)
	for _, tip := range s.tips {
		if !parentSet[tip] {
			newTips = append(newTips, tip)
		}
	}
	newTips = append(newTips, newVertex.Hash)
	s.tips = newTips
}

func (s *DAGSimulator) updateConfirmations() {
	for hash, vertex := range s.globalDAG {
		if vertex.Confirmed {
			continue
		}

		confirmerCount := 0
		for _, candidate := range s.globalDAG {
			if s.isAncestor(hash, candidate.Hash) {
				confirmerCount++
			}
		}

		if confirmerCount >= len(s.nodes)/2 {
			vertex.Confirmed = true
			s.confirmedCount++
			s.EmitEvent("vertex_confirmed", vertex.Creator, "", map[string]interface{}{
				"hash":            hash,
				"confirmer_count": confirmerCount,
			})
		}
	}
}

func (s *DAGSimulator) isAncestor(ancestorHash, descendantHash string) bool {
	if ancestorHash == descendantHash {
		return false
	}
	if s.globalDAG[descendantHash] == nil {
		return false
	}

	visited := make(map[string]bool)
	queue := []string{descendantHash}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		vertex := s.globalDAG[current]
		if vertex == nil {
			continue
		}

		for _, parent := range vertex.Parents {
			if parent == ancestorHash {
				return true
			}
			if !visited[parent] {
				queue = append(queue, parent)
			}
		}
	}

	return false
}

func (s *DAGSimulator) updateAllStates() {
	for nodeID, node := range s.nodes {
		s.SetNodeState(nodeID, &types.NodeState{
			ID:     nodeID,
			Status: "active",
			Data: map[string]interface{}{
				"vertex_count":    node.VertexCount,
				"confirmed_count": node.ConfirmedCount,
				"tips_count":      len(node.Tips),
				"round":           node.Round,
			},
		})
	}

	latestVertex := ""
	currentActor := types.NodeID("")
	latestRound := uint64(0)
	for _, vertex := range s.globalDAG {
		if vertex.Round >= latestRound {
			latestRound = vertex.Round
			latestVertex = vertex.Hash
			currentActor = vertex.Creator
		}
	}

	s.SetGlobalData("total_vertices", len(s.globalDAG))
	s.SetGlobalData("tips_count", len(s.tips))
	s.SetGlobalData("confirmed_count", s.confirmedCount)
	s.SetGlobalData("current_round", s.currentRound)
	s.SetGlobalData("tip_selection", s.tipSelection)
	s.SetGlobalData("latest_vertex", latestVertex)
	s.SetGlobalData("current_actor", currentActor)
	s.SetGlobalData("result_height", s.confirmedCount)
	setConsensusTeachingState(
		s.BaseSimulator,
		"dag_growth",
		"当前 DAG 正在扩展顶点并累积确认关系。",
		"继续观察 tips 数量、最新顶点和 confirmed 数量，判断图结构如何推进最终确认。",
		65,
		map[string]interface{}{
			"total_vertices": len(s.globalDAG),
			"tips_count":     len(s.tips),
			"confirmed":      s.confirmedCount,
			"latest_vertex":  latestVertex,
		},
	)
}

type DAGFactory struct{}

func (f *DAGFactory) Create() engine.Simulator          { return NewDAGSimulator() }
func (f *DAGFactory) GetDescription() types.Description { return NewDAGSimulator().GetDescription() }
func NewDAGFactory() *DAGFactory                        { return &DAGFactory{} }

func (s *DAGSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_vertex":
		creatorID := types.NodeID("node-0")
		if raw, ok := params["creator"].(string); ok && raw != "" {
			creatorID = types.NodeID(raw)
		}
		s.mu.Lock()
		s.createVertex(creatorID)
		s.updateAllStates()
		s.mu.Unlock()

		return consensusActionResult(
			"已创建新的 DAG 顶点",
			map[string]interface{}{
				"creator": creatorID,
			},
			&types.ActionFeedback{
				Summary:     "新的顶点已经加入 DAG 图结构。",
				NextHint:    "观察该顶点的引用关系，以及确认计数是否继续累积。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "vertex_created"},
			},
		), nil
	case "update_confirmations":
		s.mu.Lock()
		s.updateConfirmations()
		s.updateAllStates()
		confirmed := s.confirmedCount
		s.mu.Unlock()

		return consensusActionResult(
			"已刷新顶点确认状态",
			map[string]interface{}{
				"confirmed": confirmed,
			},
			&types.ActionFeedback{
				Summary:     "DAG 中已确认的顶点数量已经更新。",
				NextHint:    "观察 tips 集合和确认关系是否发生变化。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "confirmations_updated"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported dag action: %s", action)
	}
}

var _ engine.SimulatorFactory = (*DAGFactory)(nil)
