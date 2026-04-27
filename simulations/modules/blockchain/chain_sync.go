package blockchain

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// SyncNode 同步节点
type SyncNode struct {
	ID          types.NodeID   `json:"id"`
	ChainHeight uint64         `json:"chain_height"`
	SyncStatus  string         `json:"sync_status"`
	Peers       []types.NodeID `json:"peers"`
	Blocks      []string       `json:"blocks"`
}

// ChainSyncSimulator 链同步演示器
type ChainSyncSimulator struct {
	*base.BaseSimulator
	mu             sync.RWMutex
	nodes          map[types.NodeID]*SyncNode
	nodeList       []types.NodeID
	canonicalChain []string
}

// NewChainSyncSimulator 创建链同步演示器
func NewChainSyncSimulator() *ChainSyncSimulator {
	sim := &ChainSyncSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"chain_sync",
			"链同步演示器",
			"展示区块链节点间的同步过程",
			"blockchain",
			types.ComponentProcess,
		),
		nodes:    make(map[types.NodeID]*SyncNode),
		nodeList: make([]types.NodeID, 0),
	}

	sim.AddParam(types.Param{
		Key: "node_count", Name: "节点数量", Type: types.ParamTypeInt,
		Default: 5, Min: 3, Max: 20,
	})
	sim.SetOnTick(sim.onTick)
	return sim
}

// Init 初始化
func (s *ChainSyncSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	nodeCount := 5
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	s.canonicalChain = []string{"genesis"}
	for i := 1; i <= 10; i++ {
		s.canonicalChain = append(s.canonicalChain, fmt.Sprintf("block-%d", i))
	}

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("node-%d", i))
		syncHeight := rand.Intn(len(s.canonicalChain))
		node := &SyncNode{
			ID:          nodeID,
			ChainHeight: uint64(syncHeight),
			SyncStatus:  "syncing",
			Peers:       make([]types.NodeID, 0),
			Blocks:      s.canonicalChain[:syncHeight+1],
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
	}

	for _, node := range s.nodes {
		for _, peerID := range s.nodeList {
			if peerID != node.ID && rand.Float64() < 0.5 {
				node.Peers = append(node.Peers, peerID)
			}
		}
	}

	s.updateAllStates()
	return nil
}

func (s *ChainSyncSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, nodeID := range s.nodeList {
		node := s.nodes[nodeID]
		if node.ChainHeight < uint64(len(s.canonicalChain)-1) {
			s.syncNode(node)
		}
	}

	if tick%20 == 0 {
		newBlock := fmt.Sprintf("block-%d", len(s.canonicalChain))
		s.canonicalChain = append(s.canonicalChain, newBlock)
		s.EmitEvent("new_block", "", "", map[string]interface{}{
			"block": newBlock, "height": len(s.canonicalChain) - 1,
		})
	}

	s.updateAllStates()
	return nil
}

func (s *ChainSyncSimulator) syncNode(node *SyncNode) {
	if len(node.Peers) == 0 {
		return
	}

	peerID := node.Peers[rand.Intn(len(node.Peers))]
	peer := s.nodes[peerID]
	if peer == nil {
		return
	}

	if peer.ChainHeight > node.ChainHeight {
		nextHeight := node.ChainHeight + 1
		if int(nextHeight) < len(s.canonicalChain) {
			node.Blocks = append(node.Blocks, s.canonicalChain[nextHeight])
			node.ChainHeight = nextHeight
			node.SyncStatus = "syncing"

			s.EmitEvent("block_synced", node.ID, peerID, map[string]interface{}{
				"block": s.canonicalChain[nextHeight], "height": nextHeight,
			})
		}
	}

	if node.ChainHeight == uint64(len(s.canonicalChain)-1) {
		node.SyncStatus = "synced"
	}
}

func (s *ChainSyncSimulator) updateAllStates() {
	for nodeID, node := range s.nodes {
		s.SetNodeState(nodeID, &types.NodeState{
			ID:     nodeID,
			Status: node.SyncStatus,
			Data: map[string]interface{}{
				"chain_height": node.ChainHeight,
				"peer_count":   len(node.Peers),
			},
		})
	}
	s.SetGlobalData("canonical_height", len(s.canonicalChain)-1)
	s.SetGlobalData("node_count", len(s.nodes))

	summary := fmt.Sprintf("当前共有 %d 个节点，规范链高度为 %d。", len(s.nodes), len(s.canonicalChain)-1)
	nextHint := "可以继续单步推进同步过程，观察节点如何逐步追平规范链。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备同步",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"canonical_height": len(s.canonicalChain) - 1, "node_count": len(s.nodes)},
	)
}

func (s *ChainSyncSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "sync_round":
		s.mu.Lock()
		for _, nodeID := range s.nodeList {
			node := s.nodes[nodeID]
			if node.ChainHeight < uint64(len(s.canonicalChain)-1) {
				s.syncNode(node)
			}
		}
		s.updateAllStates()
		s.mu.Unlock()
		return blockchainActionResult("已执行一轮链同步。", map[string]interface{}{"canonical_height": len(s.canonicalChain) - 1}, &types.ActionFeedback{
			Summary:     "节点已经尝试从对等方同步新的区块。",
			NextHint:    "继续观察不同节点追平规范链的速度，以及何时全部转为 synced 状态。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"canonical_height": len(s.canonicalChain) - 1},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported chain sync action: %s", action)
	}
}

type ChainSyncFactory struct{}

func (f *ChainSyncFactory) Create() engine.Simulator { return NewChainSyncSimulator() }
func (f *ChainSyncFactory) GetDescription() types.Description {
	return NewChainSyncSimulator().GetDescription()
}
func NewChainSyncFactory() *ChainSyncFactory { return &ChainSyncFactory{} }

var _ engine.SimulatorFactory = (*ChainSyncFactory)(nil)
