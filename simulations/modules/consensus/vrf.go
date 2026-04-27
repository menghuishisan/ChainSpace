package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type VRFProof struct {
	PublicKey string `json:"public_key"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	Proof     string `json:"proof"`
}

type VRFNode struct {
	ID           types.NodeID `json:"id"`
	PublicKey    string       `json:"public_key"`
	PrivateKey   string       `json:"private_key"`
	Stake        uint64       `json:"stake"`
	IsSelected   bool         `json:"is_selected"`
	SelectCount  int          `json:"select_count"`
	LastVRFProof *VRFProof    `json:"last_vrf_proof,omitempty"`
	Threshold    *big.Int     `json:"-"`
}

type VRFRound struct {
	RoundNum    uint64         `json:"round_num"`
	Seed        string         `json:"seed"`
	Selected    []types.NodeID `json:"selected"`
	TotalStake  uint64         `json:"total_stake"`
	TargetCount int            `json:"target_count"`
	Timestamp   time.Time      `json:"timestamp"`
}

type VRFSimulator struct {
	*base.BaseSimulator
	mu                 sync.RWMutex
	nodes              map[types.NodeID]*VRFNode
	nodeList           []types.NodeID
	rounds             []*VRFRound
	currentRound       uint64
	currentSeed        string
	totalStake         uint64
	targetCount        int
	sortitionThreshold float64
}

func NewVRFSimulator() *VRFSimulator {
	sim := &VRFSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"vrf",
			"VRF 随机选举演示器",
			"可验证随机函数的选举机制，支持秘密选举、概率证明和公平验证。",
			"consensus",
			types.ComponentProcess,
		),
		nodes:    make(map[types.NodeID]*VRFNode),
		nodeList: make([]types.NodeID, 0),
		rounds:   make([]*VRFRound, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "参与选举的节点数量。",
		Type:        types.ParamTypeInt,
		Default:     20,
		Min:         5,
		Max:         100,
	})
	sim.AddParam(types.Param{
		Key:         "target_committee",
		Name:        "目标委员会大小",
		Description: "期望选出的委员会成员数。",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         1,
		Max:         20,
	})
	sim.AddParam(types.Param{
		Key:         "stake_weighted",
		Name:        "权益加权",
		Description: "是否按权益加权选举概率。",
		Type:        types.ParamTypeBool,
		Default:     true,
	})

	sim.SetOnTick(sim.onTick)
	return sim
}

func (s *VRFSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	nodeCount := 20
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	s.targetCount = 5
	if v, ok := config.Params["target_committee"]; ok {
		if n, ok := v.(float64); ok {
			s.targetCount = int(n)
		}
	}

	s.sortitionThreshold = float64(s.targetCount) / float64(nodeCount)
	s.nodes = make(map[types.NodeID]*VRFNode)
	s.nodeList = make([]types.NodeID, 0, nodeCount)
	s.totalStake = 0

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(fmt.Sprintf("node-%d", i))
		stake := uint64(100 + i*10)
		privateKey := s.generatePrivateKey(nodeID)
		publicKey := s.derivePublicKey(privateKey)
		node := &VRFNode{
			ID:         nodeID,
			PublicKey:  publicKey,
			PrivateKey: privateKey,
			Stake:      stake,
			IsSelected: false,
		}

		s.nodes[nodeID] = node
		s.nodeList = append(s.nodeList, nodeID)
		s.totalStake += stake
	}

	s.currentRound = 0
	s.currentSeed = s.generateSeed(0)
	s.updateAllStates()
	return nil
}

func (s *VRFSimulator) generatePrivateKey(nodeID types.NodeID) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("private-%s", nodeID)))
	return hex.EncodeToString(hash[:])
}

func (s *VRFSimulator) derivePublicKey(privateKey string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("public-%s", privateKey)))
	return hex.EncodeToString(hash[:16])
}

func (s *VRFSimulator) generateSeed(round uint64) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("seed-%d-%d", round, time.Now().UnixNano())))
	return hex.EncodeToString(hash[:])
}

func (s *VRFSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tick%20 == 0 {
		s.runElection()
	}

	s.updateAllStates()
	return nil
}

func (s *VRFSimulator) runElection() {
	s.currentRound++
	s.currentSeed = s.generateSeed(s.currentRound)

	for _, node := range s.nodes {
		node.IsSelected = false
	}

	round := &VRFRound{
		RoundNum:    s.currentRound,
		Seed:        s.currentSeed,
		Selected:    make([]types.NodeID, 0),
		TotalStake:  s.totalStake,
		TargetCount: s.targetCount,
		Timestamp:   time.Now(),
	}

	for _, nodeID := range s.nodeList {
		node := s.nodes[nodeID]
		if node == nil {
			continue
		}

		proof := s.computeVRF(node, s.currentSeed)
		node.LastVRFProof = proof
		if s.checkSelection(node, proof) {
			node.IsSelected = true
			node.SelectCount++
			round.Selected = append(round.Selected, nodeID)
			s.EmitEvent("node_selected", nodeID, "", map[string]interface{}{
				"round":        s.currentRound,
				"vrf_output":   proof.Output[:16],
				"stake":        node.Stake,
				"total_select": node.SelectCount,
			})
		}
	}

	s.rounds = append(s.rounds, round)
	s.EmitEvent("election_complete", "", "", map[string]interface{}{
		"round":          s.currentRound,
		"selected_count": len(round.Selected),
		"target_count":   s.targetCount,
		"seed":           s.currentSeed[:16],
	})
}

func (s *VRFSimulator) computeVRF(node *VRFNode, seed string) *VRFProof {
	input := fmt.Sprintf("%s-%s-%d", seed, node.PrivateKey, s.currentRound)
	outputHash := sha256.Sum256([]byte(input))
	proofHash := sha256.Sum256([]byte(fmt.Sprintf("proof-%s", input)))

	return &VRFProof{
		PublicKey: node.PublicKey,
		Input:     seed,
		Output:    hex.EncodeToString(outputHash[:]),
		Proof:     hex.EncodeToString(proofHash[:]),
	}
}

func (s *VRFSimulator) checkSelection(node *VRFNode, proof *VRFProof) bool {
	vrfOutput := new(big.Int)
	vrfOutput.SetString(proof.Output, 16)

	maxValue := new(big.Int)
	maxValue.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

	ratio := new(big.Float).Quo(
		new(big.Float).SetInt(vrfOutput),
		new(big.Float).SetInt(maxValue),
	)

	threshold := s.sortitionThreshold * float64(node.Stake) / float64(s.totalStake) * float64(len(s.nodeList))
	ratioFloat, _ := ratio.Float64()
	return ratioFloat < threshold
}

func (s *VRFSimulator) VerifyVRF(nodeID types.NodeID, proof *VRFProof) bool {
	node := s.nodes[nodeID]
	if node == nil {
		return false
	}
	if proof.PublicKey != node.PublicKey {
		return false
	}

	expectedProof := s.computeVRF(node, proof.Input)
	return expectedProof.Output == proof.Output && expectedProof.Proof == proof.Proof
}

func (s *VRFSimulator) GetSelectionProbability(nodeID types.NodeID) float64 {
	node := s.nodes[nodeID]
	if node == nil {
		return 0
	}
	return s.sortitionThreshold * float64(node.Stake) / float64(s.totalStake) * float64(len(s.nodeList))
}

func (s *VRFSimulator) updateAllStates() {
	selectedCount := 0
	selectedIDs := make([]types.NodeID, 0)

	for nodeID, node := range s.nodes {
		status := "waiting"
		if node.IsSelected {
			status = "selected"
			selectedCount++
			selectedIDs = append(selectedIDs, nodeID)
		}

		s.SetNodeState(nodeID, &types.NodeState{
			ID:     nodeID,
			Status: status,
			Data: map[string]interface{}{
				"stake":          node.Stake,
				"public_key":     node.PublicKey[:16],
				"select_count":   node.SelectCount,
				"selection_prob": s.GetSelectionProbability(nodeID),
			},
		})
	}

	s.SetGlobalData("current_round", s.currentRound)
	s.SetGlobalData("current_seed", s.currentSeed)
	s.SetGlobalData("total_stake", s.totalStake)
	s.SetGlobalData("selected_count", selectedCount)
	s.SetGlobalData("target_count", s.targetCount)
	s.SetGlobalData("total_rounds", len(s.rounds))
	s.SetGlobalData("selected_ids", selectedIDs)

	if len(selectedIDs) > 0 {
		s.SetGlobalData("current_actor", selectedIDs[0])
	} else {
		s.SetGlobalData("current_actor", "")
	}

	s.SetGlobalData("committee_size", selectedCount)
	s.SetGlobalData("result_height", selectedCount)
	setConsensusTeachingState(
		s.BaseSimulator,
		"vrf_selection",
		"当前 VRF 正在按随机种子和权重选择本轮节点集合。",
		"继续观察随机种子、选中节点和目标委员会规模的变化。",
		60,
		map[string]interface{}{
			"current_round":  s.currentRound,
			"selected_count": selectedCount,
			"target_count":   s.targetCount,
			"selected_ids":   selectedIDs,
		},
	)
}

type VRFFactory struct{}

func (f *VRFFactory) Create() engine.Simulator          { return NewVRFSimulator() }
func (f *VRFFactory) GetDescription() types.Description { return NewVRFSimulator().GetDescription() }
func NewVRFFactory() *VRFFactory                        { return &VRFFactory{} }

func (s *VRFSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "run_lottery":
		s.mu.Lock()
		s.runElection()
		round := s.currentRound
		seed := s.currentSeed
		s.mu.Unlock()

		return consensusActionResult(
			"已执行一轮 VRF 抽签",
			map[string]interface{}{
				"round": round,
				"seed":  seed,
			},
			&types.ActionFeedback{
				Summary:     "VRF 已完成当前轮次抽签并选出候选结果。",
				NextHint:    "观察随机种子如何影响选中节点，以及后续轮次是否出现不同结果。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "vrf_executed"},
			},
		), nil
	case "refresh_seed":
		s.mu.Lock()
		s.currentSeed = s.generateSeed(s.currentRound + 1)
		s.SetGlobalData("current_seed", s.currentSeed)
		seed := s.currentSeed
		s.mu.Unlock()

		return consensusActionResult(
			"已刷新随机种子",
			map[string]interface{}{
				"seed": seed,
			},
			&types.ActionFeedback{
				Summary:     "当前轮使用的随机种子已经刷新。",
				NextHint:    "观察新的种子会如何改变后续抽签结果和候选集合。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{"status": "seed_refreshed"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported vrf action: %s", action)
	}
}

var _ engine.SimulatorFactory = (*VRFFactory)(nil)
