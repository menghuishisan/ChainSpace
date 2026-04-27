package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// MerkleNode 表示树中的一个节点。
type MerkleNode struct {
	Hash   string      `json:"hash"`
	Left   *MerkleNode `json:"left,omitempty"`
	Right  *MerkleNode `json:"right,omitempty"`
	Data   string      `json:"data,omitempty"`
	IsLeaf bool        `json:"is_leaf"`
	Level  int         `json:"level"`
	Index  int         `json:"index"`
}

// MerkleProof 表示一条从叶子到根的证明路径。
type MerkleProof struct {
	LeafHash   string   `json:"leaf_hash"`
	LeafIndex  int      `json:"leaf_index"`
	Siblings   []string `json:"siblings"`
	Directions []string `json:"directions"`
	Root       string   `json:"root"`
}

// MerkleTreeSimulator 演示 Merkle 树的构建、证明和验证。
type MerkleTreeSimulator struct {
	*base.BaseSimulator
	leaves            []*MerkleNode
	root              *MerkleNode
	levels            [][]*MerkleNode
	lastProof         *MerkleProof
	lastProofVerified bool
}

// NewMerkleTreeSimulator 创建演示器。
func NewMerkleTreeSimulator() *MerkleTreeSimulator {
	sim := &MerkleTreeSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"merkle_tree",
			"Merkle 树演示器",
			"展示 Merkle 树的构建、证明生成和路径验证过程。",
			"blockchain",
			types.ComponentTool,
		),
		leaves: make([]*MerkleNode, 0),
		levels: make([][]*MerkleNode, 0),
	}
	return sim
}

// Init 初始化演示器。
func (s *MerkleTreeSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// AddLeaf 添加叶子节点。
func (s *MerkleTreeSimulator) AddLeaf(data string) {
	hash := s.hashData(data)
	leaf := &MerkleNode{
		Hash:   hash,
		Data:   data,
		IsLeaf: true,
		Level:  0,
		Index:  len(s.leaves),
	}
	s.leaves = append(s.leaves, leaf)
	s.EmitEvent("leaf_added", "", "", map[string]interface{}{
		"data":  data,
		"hash":  hash[:16],
		"index": leaf.Index,
	})
	s.rebuildTree()
}

func (s *MerkleTreeSimulator) rebuildTree() {
	if len(s.leaves) == 0 {
		s.root = nil
		s.levels = nil
		s.updateState()
		return
	}

	s.levels = make([][]*MerkleNode, 0)
	currentLevel := make([]*MerkleNode, len(s.leaves))
	copy(currentLevel, s.leaves)
	s.levels = append(s.levels, currentLevel)

	level := 1
	for len(currentLevel) > 1 {
		var nextLevel []*MerkleNode
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := left
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			}
			parent := &MerkleNode{
				Hash:  s.hashPair(left.Hash, right.Hash),
				Left:  left,
				Right: right,
				Level: level,
				Index: len(nextLevel),
			}
			nextLevel = append(nextLevel, parent)
		}
		currentLevel = nextLevel
		s.levels = append(s.levels, currentLevel)
		level++
	}

	s.root = currentLevel[0]
	s.updateState()
}

// GetProof 生成某个叶子的证明路径。
func (s *MerkleTreeSimulator) GetProof(leafIndex int) *MerkleProof {
	if leafIndex < 0 || leafIndex >= len(s.leaves) || s.root == nil {
		return nil
	}

	proof := &MerkleProof{
		LeafHash:   s.leaves[leafIndex].Hash,
		LeafIndex:  leafIndex,
		Siblings:   make([]string, 0),
		Directions: make([]string, 0),
		Root:       s.root.Hash,
	}

	idx := leafIndex
	for level := 0; level < len(s.levels)-1; level++ {
		levelNodes := s.levels[level]
		if idx%2 == 0 {
			if idx+1 < len(levelNodes) {
				proof.Siblings = append(proof.Siblings, levelNodes[idx+1].Hash)
			} else {
				proof.Siblings = append(proof.Siblings, levelNodes[idx].Hash)
			}
			proof.Directions = append(proof.Directions, "right")
		} else {
			proof.Siblings = append(proof.Siblings, levelNodes[idx-1].Hash)
			proof.Directions = append(proof.Directions, "left")
		}
		idx /= 2
	}

	s.EmitEvent("proof_generated", "", "", map[string]interface{}{
		"leaf_index":   leafIndex,
		"proof_length": len(proof.Siblings),
	})
	s.lastProof = proof
	s.lastProofVerified = false
	s.updateState()
	return proof
}

// VerifyProof 验证证明路径。
func (s *MerkleTreeSimulator) VerifyProof(proof *MerkleProof) bool {
	if proof == nil || len(proof.Siblings) != len(proof.Directions) {
		return false
	}

	currentHash := proof.LeafHash
	for i, sibling := range proof.Siblings {
		if proof.Directions[i] == "left" {
			currentHash = s.hashPair(sibling, currentHash)
		} else {
			currentHash = s.hashPair(currentHash, sibling)
		}
	}

	valid := currentHash == proof.Root
	s.EmitEvent("proof_verified", "", "", map[string]interface{}{
		"valid":         valid,
		"computed_root": currentHash[:16],
	})
	s.lastProof = proof
	s.lastProofVerified = valid
	s.updateState()
	return valid
}

func (s *MerkleTreeSimulator) hashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *MerkleTreeSimulator) hashPair(left, right string) string {
	hash := sha256.Sum256([]byte(left + right))
	return hex.EncodeToString(hash[:])
}

func (s *MerkleTreeSimulator) updateState() {
	rootHash := ""
	if s.root != nil {
		rootHash = s.root.Hash
	}

	s.SetGlobalData("root_hash", rootHash)
	s.SetGlobalData("leaf_count", len(s.leaves))
	s.SetGlobalData("tree_height", len(s.levels))

	treeData := make([]interface{}, 0)
	for _, level := range s.levels {
		levelData := make([]map[string]interface{}, len(level))
		for i, node := range level {
			levelData[i] = map[string]interface{}{
				"hash":    node.Hash[:16],
				"index":   node.Index,
				"is_leaf": node.IsLeaf,
			}
		}
		treeData = append(treeData, levelData)
	}
	s.SetGlobalData("tree_structure", treeData)

	if s.lastProof != nil {
		s.SetGlobalData("latest_proof", s.lastProof)
		s.SetGlobalData("last_verification", map[string]interface{}{
			"valid": s.lastProofVerified,
		})
	} else {
		s.SetGlobalData("latest_proof", nil)
		s.SetGlobalData("last_verification", nil)
	}

	summary := fmt.Sprintf("当前 Merkle 树共有 %d 个叶子，树高为 %d。", len(s.leaves), len(s.levels))
	nextHint := "可以继续添加叶子、生成证明并验证路径，观察哈希如何逐层汇聚到根。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备树结构",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{
			"leaf_count":  len(s.leaves),
			"tree_height": len(s.levels),
			"root_hash":   rootHash,
		},
	)
}

// ExecuteAction 提供给前端的教学动作入口。
func (s *MerkleTreeSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "add_leaf":
		data := "transaction:alice->bob:10"
		if raw, ok := params["data"].(string); ok && raw != "" {
			data = raw
		}
		s.AddLeaf(data)
		return blockchainActionResult("已向 Merkle 树添加一个新叶子。", map[string]interface{}{"data": data, "leaf_count": len(s.leaves)}, &types.ActionFeedback{
			Summary:     "新的叶子节点已加入 Merkle 树，根哈希已随之变化。",
			NextHint:    "继续生成一条证明路径，观察叶子如何通过兄弟节点一路汇聚到根。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"leaf_count": len(s.leaves)},
		}), nil
	case "generate_proof":
		leafIndex := 0
		if raw, ok := params["leaf_index"].(float64); ok {
			leafIndex = int(raw)
		}
		proof := s.GetProof(leafIndex)
		if proof == nil {
			return &types.ActionResult{
				Success: false,
				Message: "指定叶子索引不存在，无法生成证明。",
			}, nil
		}
		return blockchainActionResult("已生成指定叶子的证明路径。", map[string]interface{}{"proof": proof}, &types.ActionFeedback{
			Summary:     "叶子到根的证明路径已经生成，可继续验证该路径。",
			NextHint:    "继续执行验证，确认路径重算后能否回到当前根哈希。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"leaf_index": leafIndex, "path_length": len(proof.Siblings)},
		}), nil
	case "verify_proof":
		if s.lastProof == nil {
			return &types.ActionResult{
				Success: false,
				Message: "请先生成证明，再执行验证。",
			}, nil
		}
		valid := s.VerifyProof(s.lastProof)
		message := "证明验证失败。"
		if valid {
			message = "证明验证通过。"
		}
		return blockchainActionResult(message, map[string]interface{}{"valid": valid}, &types.ActionFeedback{
			Summary:     "系统已完成一次 Merkle 证明验证。",
			NextHint:    "继续修改叶子或重新生成证明，比较不同路径对验证结果的影响。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"valid": valid},
		}), nil
	case "reset_tree":
		s.leaves = make([]*MerkleNode, 0)
		s.root = nil
		s.levels = make([][]*MerkleNode, 0)
		s.lastProof = nil
		s.lastProofVerified = false
		s.updateState()
		return blockchainActionResult("已重置当前 Merkle 树场景。", nil, &types.ActionFeedback{
			Summary:     "Merkle 树已清空，可以重新开始叶子添加和证明生成。",
			NextHint:    "从头添加叶子，重新观察根哈希如何一步步形成。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"leaf_count": 0},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported merkle action: %s", action)
	}
}

// MerkleTreeFactory 为模拟器提供工厂。
type MerkleTreeFactory struct{}

func (f *MerkleTreeFactory) Create() engine.Simulator { return NewMerkleTreeSimulator() }
func (f *MerkleTreeFactory) GetDescription() types.Description {
	return NewMerkleTreeSimulator().GetDescription()
}
func NewMerkleTreeFactory() *MerkleTreeFactory { return &MerkleTreeFactory{} }

var _ engine.SimulatorFactory = (*MerkleTreeFactory)(nil)
