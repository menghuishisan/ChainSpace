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

type ForkBlock struct {
	Hash        string       `json:"hash"`
	ParentHash  string       `json:"parent_hash"`
	Height      uint64       `json:"height"`
	Weight      uint64       `json:"weight"`
	Votes       int          `json:"votes"`
	Creator     types.NodeID `json:"creator"`
	Timestamp   time.Time    `json:"timestamp"`
	IsCanonical bool         `json:"is_canonical"`
	Children    []string     `json:"children"`
}

type ForkNode struct {
	ID             types.NodeID `json:"id"`
	HeadBlock      string       `json:"head_block"`
	JustifiedBlock string       `json:"justified_block"`
	FinalizedBlock string       `json:"finalized_block"`
	VotedFor       string       `json:"voted_for"`
	Stake          uint64       `json:"stake"`
}

type ForkChoiceRule string

const (
	ForkChoiceLongestChain ForkChoiceRule = "longest_chain"
	ForkChoiceGHOST        ForkChoiceRule = "ghost"
	ForkChoiceLMDGHOST     ForkChoiceRule = "lmd_ghost"
	ForkChoiceHeaviest     ForkChoiceRule = "heaviest"
)

type ForkChoiceSimulator struct {
	*base.BaseSimulator
	mu            sync.RWMutex
	nodes         map[types.NodeID]*ForkNode
	nodeList      []types.NodeID
	blocks        map[string]*ForkBlock
	genesis       *ForkBlock
	tips          []string
	rule          ForkChoiceRule
	canonicalHead string
	forkCount     int
	reorgCount    int
	totalStake    uint64
}

func NewForkChoiceSimulator() *ForkChoiceSimulator {
	sim := &ForkChoiceSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"fork_choice",
			"分叉选择演示器",
			"展示 GHOST、LMD-GHOST 等分叉选择规则，支持主动构造分叉并观察规范链切换。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:    make(map[types.NodeID]*ForkNode),
		nodeList: make([]types.NodeID, 0),
		blocks:   make(map[string]*ForkBlock),
		tips:     make([]string, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "参与分叉选择过程的节点数量。",
		Type:        types.ParamTypeInt,
		Default:     6,
		Min:         3,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "fork_choice_rule",
		Name:        "分叉选择规则",
		Description: "用于决定规范链头的规则。",
		Type:        types.ParamTypeSelect,
		Default:     "lmd_ghost",
		Options: []types.Option{
			{Label: "最长链", Value: "longest_chain"},
			{Label: "GHOST", Value: "ghost"},
			{Label: "LMD-GHOST", Value: "lmd_ghost"},
			{Label: "最重链", Value: "heaviest"},
		},
	})
	sim.AddParam(types.Param{
		Key:         "fork_probability",
		Name:        "分叉概率",
		Description: "每个 tick 主动在侧枝上继续出块的概率。",
		Type:        types.ParamTypeFloat,
		Default:     0.1,
		Min:         0.0,
		Max:         0.5,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *ForkChoiceSimulator) Init(config types.Config) error {
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

	s.rule = ForkChoiceLMDGHOST
	if v, ok := config.Params["fork_choice_rule"]; ok {
		if str, ok := v.(string); ok {
			s.rule = ForkChoiceRule(str)
		}
	}

	s.genesis = &ForkBlock{
		Hash:        "genesis",
		ParentHash:  "",
		Height:      0,
		Weight:      0,
		IsCanonical: true,
		Children:    make([]string, 0),
		Timestamp:   time.Now(),
	}
	s.blocks = map[string]*ForkBlock{s.genesis.Hash: s.genesis}
	s.tips = []string{s.genesis.Hash}
	s.canonicalHead = s.genesis.Hash
	s.forkCount = 0
	s.reorgCount = 0

	s.nodes = make(map[types.NodeID]*ForkNode)
	s.nodeList = make([]types.NodeID, 0, nodeCount)
	s.totalStake = 0

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("node-%d", i))
		node := &ForkNode{
			ID:             nodeID,
			HeadBlock:      s.genesis.Hash,
			JustifiedBlock: s.genesis.Hash,
			FinalizedBlock: s.genesis.Hash,
			VotedFor:       "",
			Stake:          100,
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
		s.totalStake += 100
	}

	s.updateAllStates()
	return nil
}

func (s *ForkChoiceSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tick%3 == 0 {
		creatorID := s.nodeList[rand.Intn(len(s.nodeList))]
		forkProb := 0.1
		if p := s.BaseSimulator.GetParams()["fork_probability"]; p.Value != nil {
			if v, ok := p.Value.(float64); ok {
				forkProb = v
			}
		}

		if rand.Float64() < forkProb && len(s.tips) > 0 {
			parentHash := s.tips[rand.Intn(len(s.tips))]
			s.createBlock(creatorID, parentHash)
		} else {
			s.createBlock(creatorID, s.canonicalHead)
		}
	}

	if tick%5 == 0 {
		for _, nodeID := range s.nodeList {
			s.castVote(nodeID)
		}
		s.updateCanonicalChain()
	}

	s.updateAllStates()
	return nil
}

func (s *ForkChoiceSimulator) createBlock(creatorID types.NodeID, parentHash string) {
	parent := s.blocks[parentHash]
	if parent == nil {
		return
	}

	block := &ForkBlock{
		Hash:        fmt.Sprintf("block-%d-%s", len(s.blocks), uuid.New().String()[:8]),
		ParentHash:  parentHash,
		Height:      parent.Height + 1,
		Weight:      parent.Weight + 1,
		Creator:     creatorID,
		Timestamp:   time.Now(),
		IsCanonical: false,
		Children:    make([]string, 0),
	}

	s.blocks[block.Hash] = block
	parent.Children = append(parent.Children, block.Hash)
	s.updateTips(block)

	if parentHash != s.canonicalHead {
		s.forkCount++
		s.EmitEvent("fork_created", creatorID, "", map[string]interface{}{
			"block_hash":  block.Hash,
			"parent_hash": parentHash,
			"height":      block.Height,
			"fork_count":  s.forkCount,
		})
		return
	}

	s.EmitEvent("block_created", creatorID, "", map[string]interface{}{
		"block_hash": block.Hash,
		"height":     block.Height,
	})
}

func (s *ForkChoiceSimulator) updateTips(newBlock *ForkBlock) {
	newTips := make([]string, 0)
	for _, tip := range s.tips {
		if tip != newBlock.ParentHash {
			newTips = append(newTips, tip)
		}
	}
	newTips = append(newTips, newBlock.Hash)
	s.tips = newTips
}

func (s *ForkChoiceSimulator) castVote(nodeID types.NodeID) {
	node := s.nodes[nodeID]
	if node == nil {
		return
	}

	var bestTip string
	switch s.rule {
	case ForkChoiceLongestChain:
		bestTip = s.findLongestChainTip()
	case ForkChoiceGHOST:
		bestTip = s.findGHOSTTip()
	case ForkChoiceLMDGHOST:
		bestTip = s.findLMDGHOSTTip()
	case ForkChoiceHeaviest:
		bestTip = s.findHeaviestTip()
	default:
		bestTip = s.findLongestChainTip()
	}

	if bestTip == "" || bestTip == node.VotedFor {
		return
	}

	if oldBlock := s.blocks[node.VotedFor]; oldBlock != nil {
		oldBlock.Votes--
	}
	node.VotedFor = bestTip
	node.HeadBlock = bestTip
	if newBlock := s.blocks[bestTip]; newBlock != nil {
		newBlock.Votes++
	}
}

func (s *ForkChoiceSimulator) findLongestChainTip() string {
	var best string
	var bestHeight uint64
	for _, tipHash := range s.tips {
		if block := s.blocks[tipHash]; block != nil && block.Height > bestHeight {
			bestHeight = block.Height
			best = tipHash
		}
	}
	return best
}

func (s *ForkChoiceSimulator) findGHOSTTip() string {
	current := s.genesis.Hash
	for {
		block := s.blocks[current]
		if block == nil || len(block.Children) == 0 {
			return current
		}

		var bestChild string
		var bestWeight uint64
		for _, childHash := range block.Children {
			weight := s.getSubtreeWeight(childHash)
			if weight > bestWeight {
				bestWeight = weight
				bestChild = childHash
			}
		}
		if bestChild == "" {
			return current
		}
		current = bestChild
	}
}

func (s *ForkChoiceSimulator) findLMDGHOSTTip() string {
	current := s.genesis.Hash
	for {
		block := s.blocks[current]
		if block == nil || len(block.Children) == 0 {
			return current
		}

		var bestChild string
		var bestVotes int
		for _, childHash := range block.Children {
			votes := s.getSubtreeVotes(childHash)
			if votes > bestVotes {
				bestVotes = votes
				bestChild = childHash
			}
		}
		if bestChild == "" {
			return current
		}
		current = bestChild
	}
}

func (s *ForkChoiceSimulator) findHeaviestTip() string {
	var best string
	var bestWeight uint64
	for _, tipHash := range s.tips {
		weight := s.getChainWeight(tipHash)
		if weight > bestWeight {
			bestWeight = weight
			best = tipHash
		}
	}
	return best
}

func (s *ForkChoiceSimulator) getSubtreeWeight(hash string) uint64 {
	block := s.blocks[hash]
	if block == nil {
		return 0
	}

	weight := block.Weight
	for _, childHash := range block.Children {
		weight += s.getSubtreeWeight(childHash)
	}
	return weight
}

func (s *ForkChoiceSimulator) getSubtreeVotes(hash string) int {
	block := s.blocks[hash]
	if block == nil {
		return 0
	}

	votes := block.Votes
	for _, childHash := range block.Children {
		votes += s.getSubtreeVotes(childHash)
	}
	return votes
}

func (s *ForkChoiceSimulator) getChainWeight(hash string) uint64 {
	weight := uint64(0)
	current := hash
	for current != "" {
		block := s.blocks[current]
		if block == nil {
			break
		}
		weight += block.Weight
		current = block.ParentHash
	}
	return weight
}

func (s *ForkChoiceSimulator) updateCanonicalChain() {
	var newHead string
	switch s.rule {
	case ForkChoiceLongestChain:
		newHead = s.findLongestChainTip()
	case ForkChoiceGHOST:
		newHead = s.findGHOSTTip()
	case ForkChoiceLMDGHOST:
		newHead = s.findLMDGHOSTTip()
	case ForkChoiceHeaviest:
		newHead = s.findHeaviestTip()
	}

	if newHead == "" || newHead == s.canonicalHead {
		return
	}

	for _, block := range s.blocks {
		block.IsCanonical = false
	}

	current := newHead
	for current != "" {
		block := s.blocks[current]
		if block == nil {
			break
		}
		block.IsCanonical = true
		current = block.ParentHash
	}

	if s.canonicalHead != "" && !s.isAncestor(s.canonicalHead, newHead) {
		s.reorgCount++
		s.EmitEvent("chain_reorg", "", "", map[string]interface{}{
			"old_head":    s.canonicalHead,
			"new_head":    newHead,
			"reorg_count": s.reorgCount,
		})
	}

	s.canonicalHead = newHead
}

func (s *ForkChoiceSimulator) isAncestor(ancestorHash, descendantHash string) bool {
	current := descendantHash
	for current != "" {
		if current == ancestorHash {
			return true
		}
		block := s.blocks[current]
		if block == nil {
			break
		}
		current = block.ParentHash
	}
	return false
}

func (s *ForkChoiceSimulator) TriggerFork(parentHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.blocks[parentHash]; !ok {
		return fmt.Errorf("parent block not found: %s", parentHash)
	}

	creatorID := s.nodeList[rand.Intn(len(s.nodeList))]
	s.createBlock(creatorID, parentHash)
	s.updateAllStates()
	return nil
}

func (s *ForkChoiceSimulator) updateAllStates() {
	for nodeID, node := range s.nodes {
		s.SetNodeState(nodeID, &types.NodeState{
			ID:     nodeID,
			Status: "active",
			Data: map[string]interface{}{
				"head_block": node.HeadBlock,
				"voted_for":  node.VotedFor,
				"stake":      node.Stake,
			},
		})
	}

	canonicalHeight := uint64(0)
	canonicalCreator := types.NodeID("")
	if block := s.blocks[s.canonicalHead]; block != nil {
		canonicalHeight = block.Height
		canonicalCreator = block.Creator
	}

	s.SetGlobalData("canonical_head", s.canonicalHead)
	s.SetGlobalData("canonical_height", canonicalHeight)
	s.SetGlobalData("total_blocks", len(s.blocks))
	s.SetGlobalData("tips_count", len(s.tips))
	s.SetGlobalData("fork_count", s.forkCount)
	s.SetGlobalData("reorg_count", s.reorgCount)
	s.SetGlobalData("fork_choice_rule", s.rule)
	s.SetGlobalData("active_heads", s.tips)
	s.SetGlobalData("current_actor", canonicalCreator)
	s.SetGlobalData("result_height", canonicalHeight)
	setConsensusTeachingState(
		s.BaseSimulator,
		"fork_choice",
		"当前网络正在比较多个链头，并按规则选择规范链。",
		"继续观察 tips 数量、重组次数和规范链头变化，理解分叉是如何被收敛的。",
		70,
		map[string]interface{}{
			"canonical_head":   s.canonicalHead,
			"canonical_height": canonicalHeight,
			"tips_count":       len(s.tips),
			"reorg_count":      s.reorgCount,
		},
	)
}

type ForkChoiceFactory struct{}

func (f *ForkChoiceFactory) Create() engine.Simulator { return NewForkChoiceSimulator() }

func (f *ForkChoiceFactory) GetDescription() types.Description {
	return NewForkChoiceSimulator().GetDescription()
}

func NewForkChoiceFactory() *ForkChoiceFactory { return &ForkChoiceFactory{} }

func (s *ForkChoiceSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_fork":
		parentHash := s.canonicalHead
		if raw, ok := params["parent"].(string); ok && raw != "" {
			parentHash = raw
		}
		if err := s.TriggerFork(parentHash); err != nil {
			return nil, err
		}

		return consensusActionResult(
			"已创建新的分叉分支",
			map[string]interface{}{
				"parent": parentHash,
			},
			&types.ActionFeedback{
				Summary:     "新的候选链分支已经生成。",
				NextHint:    "观察规范链选择规则是否会改变当前链头。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "fork_created"},
			},
		), nil
	case "recompute_canonical":
		s.mu.Lock()
		s.updateCanonicalChain()
		s.updateAllStates()
		head := s.canonicalHead
		s.mu.Unlock()

		return consensusActionResult(
			"已重新计算规范链头",
			map[string]interface{}{
				"canonical_head": head,
			},
			&types.ActionFeedback{
				Summary:     "系统已重新评估所有分支并更新规范链头。",
				NextHint:    "观察分叉竞争结果是否导致重组，以及最终链高如何变化。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "canonical_recomputed"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported fork_choice action: %s", action)
	}
}

var _ engine.SimulatorFactory = (*ForkChoiceFactory)(nil)
