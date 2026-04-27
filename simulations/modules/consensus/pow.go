package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type PoWNode struct {
	ID           types.NodeID `json:"id"`
	HashPower    float64      `json:"hash_power"`
	BlocksMined  int          `json:"blocks_mined"`
	CurrentNonce uint64       `json:"current_nonce"`
	IsMining     bool         `json:"is_mining"`
	IsSelfish    bool         `json:"is_selfish"`
	HiddenBlocks []*PoWBlock  `json:"hidden_blocks,omitempty"`
	TotalReward  uint64       `json:"total_reward"`
}

type PoWBlock struct {
	Hash       string       `json:"hash"`
	PrevHash   string       `json:"prev_hash"`
	Height     uint64       `json:"height"`
	Nonce      uint64       `json:"nonce"`
	Difficulty uint64       `json:"difficulty"`
	Miner      types.NodeID `json:"miner"`
	Timestamp  time.Time    `json:"timestamp"`
	Txs        []string     `json:"txs"`
}

type PoWSimulator struct {
	*base.BaseSimulator
	mu              sync.RWMutex
	nodes           map[types.NodeID]*PoWNode
	nodeList        []types.NodeID
	chain           []*PoWBlock
	orphans         []*PoWBlock
	difficulty      uint64
	targetBlockTime int
	blockReward     uint64
	totalHashPower  float64
	pendingTxs      []string
	isUnder51Attack bool
	attackerPower   float64
}

func NewPoWSimulator() *PoWSimulator {
	sim := &PoWSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"pow",
			"PoW 工作量证明",
			"模拟 PoW 的挖矿竞争、难度调整、分叉处理以及 51% 攻击和自私挖矿场景。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:    make(map[types.NodeID]*PoWNode),
		nodeList: make([]types.NodeID, 0),
		chain:    make([]*PoWBlock, 0),
		orphans:  make([]*PoWBlock, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "矿工数量",
		Description: "参与挖矿的节点数量。",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         2,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "difficulty",
		Name:        "挖矿难度",
		Description: "用于控制出块难度的目标值。",
		Type:        types.ParamTypeInt,
		Default:     4,
		Min:         1,
		Max:         8,
	})
	sim.AddParam(types.Param{
		Key:         "target_block_time",
		Name:        "目标出块时间(tick)",
		Description: "期望的平均出块间隔。",
		Type:        types.ParamTypeInt,
		Default:     10,
		Min:         5,
		Max:         100,
	})
	sim.AddParam(types.Param{
		Key:         "block_reward",
		Name:        "区块奖励",
		Description: "成功出块后矿工获得的奖励。",
		Type:        types.ParamTypeInt,
		Default:     50,
		Min:         1,
		Max:         1000,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *PoWSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	nodeCount := 5
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	s.difficulty = 4
	if v, ok := config.Params["difficulty"]; ok {
		if n, ok := v.(float64); ok {
			s.difficulty = uint64(n)
		}
	}

	s.targetBlockTime = 10
	if v, ok := config.Params["target_block_time"]; ok {
		if n, ok := v.(float64); ok {
			s.targetBlockTime = int(n)
		}
	}

	s.blockReward = 50
	if v, ok := config.Params["block_reward"]; ok {
		if n, ok := v.(float64); ok {
			s.blockReward = uint64(n)
		}
	}

	s.nodes = make(map[types.NodeID]*PoWNode)
	s.nodeList = make([]types.NodeID, 0, nodeCount)
	s.chain = make([]*PoWBlock, 0)
	s.orphans = make([]*PoWBlock, 0)
	s.totalHashPower = 0
	s.isUnder51Attack = false
	s.attackerPower = 0

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("miner-%d", i))
		hashPower := 1.0 + rand.Float64()*2.0
		node := &PoWNode{
			ID:           nodeID,
			HashPower:    hashPower,
			IsMining:     true,
			HiddenBlocks: make([]*PoWBlock, 0),
		}
		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
		s.totalHashPower += hashPower
	}

	genesis := &PoWBlock{
		Hash:       s.computeHash("genesis", 0),
		PrevHash:   "",
		Height:     0,
		Difficulty: s.difficulty,
		Timestamp:  time.Now(),
	}
	s.chain = []*PoWBlock{genesis}

	s.updateAllNodeStates()
	s.updateGlobalState()
	return nil
}

func (s *PoWSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, nodeID := range s.nodeList {
		node := s.nodes[nodeID]
		if !node.IsMining {
			continue
		}
		if s.tryMine(nodeID) {
			break
		}
	}

	if tick%100 == 0 {
		s.adjustDifficulty()
	}

	s.updateAllNodeStates()
	s.updateGlobalState()
	return nil
}

func (s *PoWSimulator) tryMine(minerID types.NodeID) bool {
	node := s.nodes[minerID]
	if node == nil {
		return false
	}

	probability := node.HashPower / s.totalHashPower / float64(s.targetBlockTime)
	if rand.Float64() > probability {
		node.CurrentNonce++
		return false
	}

	prevBlock := s.chain[len(s.chain)-1]
	newBlock := &PoWBlock{
		PrevHash:   prevBlock.Hash,
		Height:     prevBlock.Height + 1,
		Nonce:      node.CurrentNonce,
		Difficulty: s.difficulty,
		Miner:      minerID,
		Timestamp:  time.Now(),
	}
	newBlock.Hash = s.computeBlockHash(newBlock)

	if node.IsSelfish && len(node.HiddenBlocks) > 0 {
		s.handleSelfishMining(minerID, newBlock)
		return true
	}

	s.broadcastBlock(minerID, newBlock)
	return true
}

func (s *PoWSimulator) broadcastBlock(minerID types.NodeID, block *PoWBlock) {
	node := s.nodes[minerID]
	if node == nil {
		return
	}

	s.chain = append(s.chain, block)
	node.BlocksMined++
	node.TotalReward += s.blockReward
	node.CurrentNonce = 0

	s.EmitEvent("block_mined", minerID, "", map[string]interface{}{
		"height":     block.Height,
		"hash":       block.Hash[:16],
		"difficulty": block.Difficulty,
	})
}

func (s *PoWSimulator) handleSelfishMining(minerID types.NodeID, block *PoWBlock) {
	node := s.nodes[minerID]
	node.HiddenBlocks = append(node.HiddenBlocks, block)
	s.EmitEvent("selfish_mining_hidden", minerID, "", map[string]interface{}{
		"hidden_count": len(node.HiddenBlocks),
		"height":       block.Height,
	})
}

func (s *PoWSimulator) adjustDifficulty() {
	if len(s.chain) < 10 {
		return
	}

	recent := s.chain[len(s.chain)-10:]
	avgTime := float64(recent[len(recent)-1].Timestamp.Sub(recent[0].Timestamp).Milliseconds()) / 9.0
	targetTime := float64(s.targetBlockTime * 100)

	if avgTime < targetTime*0.8 {
		s.difficulty++
		s.EmitEvent("difficulty_increased", "", "", map[string]interface{}{
			"new_difficulty": s.difficulty,
		})
		return
	}

	if avgTime > targetTime*1.2 && s.difficulty > 1 {
		s.difficulty--
		s.EmitEvent("difficulty_decreased", "", "", map[string]interface{}{
			"new_difficulty": s.difficulty,
		})
	}
}

func (s *PoWSimulator) computeHash(data string, nonce uint64) string {
	input := fmt.Sprintf("%s%d", data, nonce)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func (s *PoWSimulator) computeBlockHash(block *PoWBlock) string {
	data := fmt.Sprintf("%s%d%d%s", block.PrevHash, block.Height, block.Nonce, block.Miner)
	return s.computeHash(data, block.Nonce)
}

func (s *PoWSimulator) meetsTarget(hash string, difficulty uint64) bool {
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-difficulty*4))
	hashInt := new(big.Int)
	hashInt.SetString(hash, 16)
	return hashInt.Cmp(target) < 0
}

func (s *PoWSimulator) updateAllNodeStates() {
	for nodeID, node := range s.nodes {
		s.SetNodeState(nodeID, &types.NodeState{
			ID:     nodeID,
			Status: s.getNodeStatus(node),
			Data: map[string]interface{}{
				"hash_power":     node.HashPower,
				"hash_power_pct": node.HashPower / s.totalHashPower * 100,
				"blocks_mined":   node.BlocksMined,
				"total_reward":   node.TotalReward,
				"current_nonce":  node.CurrentNonce,
				"is_selfish":     node.IsSelfish,
			},
		})
	}
}

func (s *PoWSimulator) getNodeStatus(node *PoWNode) string {
	if !node.IsMining {
		return "idle"
	}
	if node.IsSelfish {
		return "selfish_mining"
	}
	return "mining"
}

func (s *PoWSimulator) updateGlobalState() {
	s.SetGlobalData("chain_height", len(s.chain)-1)
	s.SetGlobalData("difficulty", s.difficulty)
	s.SetGlobalData("total_hash_power", s.totalHashPower)
	s.SetGlobalData("orphan_count", len(s.orphans))
	s.SetGlobalData("committee_size", len(s.nodeList))
	s.SetGlobalData("result_height", len(s.chain)-1)

	latestMiner := ""
	if len(s.chain) > 0 {
		latestMiner = string(s.chain[len(s.chain)-1].Miner)
		s.SetGlobalData("latest_block", s.chain[len(s.chain)-1].Hash[:16])
		s.SetGlobalData("current_actor", latestMiner)
	}

	setConsensusTeachingState(
		s.BaseSimulator,
		"pow_mining",
		"当前 PoW 网络正在进行出块竞争与链头选择。",
		"继续观察最新矿工、孤块数量和难度变化，判断是否出现分叉或重组。",
		65,
		map[string]interface{}{
			"chain_height":  len(s.chain) - 1,
			"difficulty":    s.difficulty,
			"orphan_count":  len(s.orphans),
			"current_actor": latestMiner,
		},
	)
}

func (s *PoWSimulator) Enable51Attack(attackerID types.NodeID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enable51AttackLocked(attackerID)
}

func (s *PoWSimulator) enable51AttackLocked(attackerID types.NodeID) error {
	node := s.nodes[attackerID]
	if node == nil {
		return fmt.Errorf("node not found: %s", attackerID)
	}

	otherPower := s.totalHashPower - node.HashPower
	node.HashPower = otherPower * 1.1
	s.totalHashPower = node.HashPower + otherPower
	s.isUnder51Attack = true
	s.attackerPower = node.HashPower / s.totalHashPower

	s.EmitEvent("51_attack_started", attackerID, "", map[string]interface{}{
		"attacker_power": s.attackerPower * 100,
	})
	return nil
}

func (s *PoWSimulator) EnableSelfishMining(minerID types.NodeID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enableSelfishMiningLocked(minerID)
}

func (s *PoWSimulator) enableSelfishMiningLocked(minerID types.NodeID) error {
	node := s.nodes[minerID]
	if node == nil {
		return fmt.Errorf("node not found: %s", minerID)
	}
	node.IsSelfish = true
	node.HiddenBlocks = make([]*PoWBlock, 0)

	s.EmitEvent("selfish_mining_enabled", minerID, "", map[string]interface{}{
		"hash_power": node.HashPower,
	})
	return nil
}

func (s *PoWSimulator) InjectFault(fault *types.Fault) error {
	if err := s.BaseSimulator.InjectFault(fault); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch fault.Type {
	case "51_percent":
		return s.enable51AttackLocked(fault.Target)
	case "selfish_mining":
		return s.enableSelfishMiningLocked(fault.Target)
	default:
		return nil
	}
}

// ExecuteAction 执行 PoW 教学动作。
func (s *PoWSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "mine_block_now":
		minerID := types.NodeID("miner-0")
		if raw, ok := params["miner"].(string); ok && raw != "" {
			minerID = types.NodeID(raw)
		}

		node := s.nodes[minerID]
		if node == nil {
			return nil, fmt.Errorf("miner not found: %s", minerID)
		}

		prevBlock := s.chain[len(s.chain)-1]
		block := &PoWBlock{
			PrevHash:   prevBlock.Hash,
			Height:     prevBlock.Height + 1,
			Nonce:      node.CurrentNonce,
			Difficulty: s.difficulty,
			Miner:      minerID,
			Timestamp:  time.Now(),
		}
		block.Hash = s.computeBlockHash(block)
		s.broadcastBlock(minerID, block)
		s.updateAllNodeStates()
		s.updateGlobalState()

		return consensusActionResult(
			"已触发矿工立即出块",
			map[string]interface{}{
				"miner":  minerID,
				"height": block.Height,
				"hash":   block.Hash,
			},
			&types.ActionFeedback{
				Summary:     "候选区块已经广播到网络，链头竞争可能随之发生变化。",
				NextHint:    "观察规范链高度、分叉情况以及链头选择是否发生变化。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "block_mined"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported pow action: %s", action)
	}
}

type PoWFactory struct{}

func (f *PoWFactory) Create() engine.Simulator { return NewPoWSimulator() }

func (f *PoWFactory) GetDescription() types.Description {
	return NewPoWSimulator().GetDescription()
}

func NewPoWFactory() *PoWFactory { return &PoWFactory{} }

var _ engine.SimulatorFactory = (*PoWFactory)(nil)
