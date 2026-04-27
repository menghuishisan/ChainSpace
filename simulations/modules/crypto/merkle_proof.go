package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// MerkleProofNode 证明节点
type MerkleProofNode struct {
	Hash      string `json:"hash"`      // 节点哈希
	Direction string `json:"direction"` // 方向: left/right
	Level     int    `json:"level"`     // 层级
}

// MerkleProofData Merkle证明数据
type MerkleProofData struct {
	LeafData  string             `json:"leaf_data"`  // 叶子数据
	LeafHash  string             `json:"leaf_hash"`  // 叶子哈希
	LeafIndex int                `json:"leaf_index"` // 叶子索引
	Path      []*MerkleProofNode `json:"path"`       // 证明路径
	RootHash  string             `json:"root_hash"`  // 根哈希
}

// MerkleProofSimulator Merkle证明演示器
// 独立于blockchain模块的密码学Merkle证明演示
type MerkleProofSimulator struct {
	*base.BaseSimulator
	leaves     []string   // 叶子数据列表
	leafHashes []string   // 叶子哈希列表
	tree       [][]string // 完整树结构
	rootHash   string     // 根哈希
}

// NewMerkleProofSimulator 创建Merkle证明演示器
func NewMerkleProofSimulator() *MerkleProofSimulator {
	sim := &MerkleProofSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"merkle_proof",
			"Merkle证明演示器",
			"展示Merkle树的构建和证明生成/验证过程",
			"crypto",
			types.ComponentTool,
		),
		leaves:     make([]string, 0),
		leafHashes: make([]string, 0),
		tree:       make([][]string, 0),
	}
	return sim
}

// Init 初始化演示器
func (s *MerkleProofSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	// 添加示例数据
	s.AddLeaf("Transaction 1")
	s.AddLeaf("Transaction 2")
	s.AddLeaf("Transaction 3")
	s.AddLeaf("Transaction 4")
	return nil
}

// AddLeaf 添加叶子节点
func (s *MerkleProofSimulator) AddLeaf(data string) {
	s.leaves = append(s.leaves, data)
	hash := s.hashData(data)
	s.leafHashes = append(s.leafHashes, hash)
	s.rebuildTree()

	s.EmitEvent("leaf_added", "", "", map[string]interface{}{
		"data":       data,
		"hash":       hash[:16] + "...",
		"leaf_count": len(s.leaves),
	})
}

// rebuildTree 重建Merkle树
func (s *MerkleProofSimulator) rebuildTree() {
	if len(s.leafHashes) == 0 {
		s.tree = nil
		s.rootHash = ""
		s.updateState()
		return
	}

	// 初始化树的第一层(叶子层)
	s.tree = make([][]string, 0)
	currentLevel := make([]string, len(s.leafHashes))
	copy(currentLevel, s.leafHashes)
	s.tree = append(s.tree, currentLevel)

	// 逐层构建
	for len(currentLevel) > 1 {
		var nextLevel []string
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := left
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			}
			parentHash := s.hashPair(left, right)
			nextLevel = append(nextLevel, parentHash)
		}
		s.tree = append(s.tree, nextLevel)
		currentLevel = nextLevel
	}

	s.rootHash = currentLevel[0]
	s.updateState()
}

// GenerateProof 生成Merkle证明
func (s *MerkleProofSimulator) GenerateProof(leafIndex int) (*MerkleProofData, error) {
	if leafIndex < 0 || leafIndex >= len(s.leaves) {
		return nil, fmt.Errorf("无效的叶子索引: %d", leafIndex)
	}

	proof := &MerkleProofData{
		LeafData:  s.leaves[leafIndex],
		LeafHash:  s.leafHashes[leafIndex],
		LeafIndex: leafIndex,
		Path:      make([]*MerkleProofNode, 0),
		RootHash:  s.rootHash,
	}

	idx := leafIndex
	for level := 0; level < len(s.tree)-1; level++ {
		levelNodes := s.tree[level]
		var siblingHash string
		var direction string

		if idx%2 == 0 {
			// 当前节点在左边，兄弟在右边
			if idx+1 < len(levelNodes) {
				siblingHash = levelNodes[idx+1]
			} else {
				siblingHash = levelNodes[idx]
			}
			direction = "right"
		} else {
			// 当前节点在右边，兄弟在左边
			siblingHash = levelNodes[idx-1]
			direction = "left"
		}

		proof.Path = append(proof.Path, &MerkleProofNode{
			Hash:      siblingHash,
			Direction: direction,
			Level:     level,
		})

		idx = idx / 2
	}

	s.EmitEvent("proof_generated", "", "", map[string]interface{}{
		"leaf_index":   leafIndex,
		"proof_length": len(proof.Path),
	})

	return proof, nil
}

// VerifyProof 验证Merkle证明
func (s *MerkleProofSimulator) VerifyProof(proof *MerkleProofData) bool {
	currentHash := proof.LeafHash

	for _, node := range proof.Path {
		if node.Direction == "left" {
			currentHash = s.hashPair(node.Hash, currentHash)
		} else {
			currentHash = s.hashPair(currentHash, node.Hash)
		}
	}

	valid := currentHash == proof.RootHash

	s.EmitEvent("proof_verified", "", "", map[string]interface{}{
		"valid":         valid,
		"computed_root": currentHash[:16] + "...",
		"expected_root": proof.RootHash[:16] + "...",
	})

	return valid
}

// hashData 计算数据哈希
func (s *MerkleProofSimulator) hashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// hashPair 计算两个哈希的组合哈希
func (s *MerkleProofSimulator) hashPair(left, right string) string {
	combined := left + right
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// GetTreeStructure 获取树结构
func (s *MerkleProofSimulator) GetTreeStructure() [][]string {
	return s.tree
}

// updateState 更新状态
func (s *MerkleProofSimulator) updateState() {
	s.SetGlobalData("leaf_count", len(s.leaves))
	s.SetGlobalData("tree_height", len(s.tree))
	s.SetGlobalData("root_hash", s.rootHash)
	s.SetGlobalData("leaves", s.leaves)

	// 简化的树结构展示
	treeView := make([]map[string]interface{}, 0)
	for level, nodes := range s.tree {
		levelView := make([]string, len(nodes))
		for i, h := range nodes {
			if len(h) > 16 {
				levelView[i] = h[:16] + "..."
			} else {
				levelView[i] = h
			}
		}
		treeView = append(treeView, map[string]interface{}{
			"level": level,
			"nodes": levelView,
		})
	}
	s.SetGlobalData("tree_view", treeView)

	setCryptoTeachingState(
		s.BaseSimulator,
		"crypto",
		func() string {
			if len(s.leaves) > 0 {
				return "tree_ready"
			}
			return "tree_empty"
		}(),
		func() string {
			if len(s.leaves) > 0 {
				return fmt.Sprintf("当前 Merkle 树共有 %d 个叶子，树高为 %d。", len(s.leaves), len(s.tree))
			}
			return "当前 Merkle 树为空，可以先添加叶子再生成证明。"
		}(),
		"重点观察叶子哈希如何逐层向上汇聚成根哈希，以及证明路径如何被验证。",
		func() float64 {
			if len(s.leaves) > 0 {
				return 0.75
			}
			return 0.2
		}(),
		map[string]interface{}{"leaf_count": len(s.leaves), "tree_height": len(s.tree), "root_hash": s.rootHash},
	)
}

// ExecuteAction 为 Merkle 证明实验提供交互动作。
func (s *MerkleProofSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "add_leaf":
		data := fmt.Sprintf("Leaf %d", len(s.leaves)+1)
		if raw, ok := params["data"].(string); ok && raw != "" {
			data = raw
		}
		s.AddLeaf(data)
		return cryptoActionResult("已添加一个叶子节点。", map[string]interface{}{"data": data, "leaf_count": len(s.leaves), "root_hash": s.rootHash}, &types.ActionFeedback{
			Summary:     "叶子节点已经加入树中，根哈希会随之更新。",
			NextHint:    "继续生成一条证明路径，观察当前叶子如何通过兄弟节点一路汇聚到根。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"leaf_count": len(s.leaves), "root_hash": s.rootHash},
		}), nil
	case "generate_proof":
		index := 0
		if raw, ok := params["leaf_index"].(float64); ok && int(raw) >= 0 {
			index = int(raw)
		}
		proof, err := s.GenerateProof(index)
		if err != nil {
			return nil, err
		}
		return cryptoActionResult("已生成一条 Merkle 证明。", map[string]interface{}{"proof": proof}, &types.ActionFeedback{
			Summary:     "叶子到根的证明路径已经生成。",
			NextHint:    "继续执行验证，确认按照左右顺序重新哈希后能否回到当前根哈希。",
			EffectScope: "crypto",
			ResultState: map[string]interface{}{"path_length": len(proof.Path)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported merkle proof action: %s", action)
	}
}

// MerkleProofFactory Merkle证明演示器工厂
type MerkleProofFactory struct{}

// Create 创建演示器实例
func (f *MerkleProofFactory) Create() engine.Simulator {
	return NewMerkleProofSimulator()
}

// GetDescription 获取描述
func (f *MerkleProofFactory) GetDescription() types.Description {
	return NewMerkleProofSimulator().GetDescription()
}

// NewMerkleProofFactory 创建工厂实例
func NewMerkleProofFactory() *MerkleProofFactory {
	return &MerkleProofFactory{}
}

var _ engine.SimulatorFactory = (*MerkleProofFactory)(nil)
