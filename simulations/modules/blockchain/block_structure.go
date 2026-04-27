package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// BlockField 表示当前区块可观察的字段。
type BlockField struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
	Editable    bool        `json:"editable"`
}

// DemoBlock 表示教学演示用区块。
type DemoBlock struct {
	Version       uint32    `json:"version"`
	PrevBlockHash string    `json:"prev_block_hash"`
	MerkleRoot    string    `json:"merkle_root"`
	Timestamp     time.Time `json:"timestamp"`
	Difficulty    uint32    `json:"difficulty"`
	Nonce         uint64    `json:"nonce"`
	Hash          string    `json:"hash"`
	Transactions  []string  `json:"transactions"`
}

// BlockStructureSimulator 演示区块字段变化、哈希变化和挖矿过程。
type BlockStructureSimulator struct {
	*base.BaseSimulator
	currentBlock *DemoBlock
	chain        []*DemoBlock
}

// NewBlockStructureSimulator 创建演示器。
func NewBlockStructureSimulator() *BlockStructureSimulator {
	sim := &BlockStructureSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"block_structure",
			"区块结构演示器",
			"展示区块头字段、交易列表、Merkle 根和区块哈希之间的关系。",
			"blockchain",
			types.ComponentDemo,
		),
		chain: make([]*DemoBlock, 0),
	}

	sim.AddParam(types.Param{
		Key:         "version",
		Name:        "区块版本",
		Description: "区块头中的版本字段。",
		Type:        types.ParamTypeInt,
		Default:     1,
	})
	sim.AddParam(types.Param{
		Key:         "difficulty",
		Name:        "难度目标",
		Description: "区块挖矿时需要满足的前导零数量。",
		Type:        types.ParamTypeInt,
		Default:     4,
		Min:         1,
		Max:         8,
	})

	return sim
}

// Init 初始化演示器。
func (s *BlockStructureSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	genesis := s.createGenesisBlock()
	s.chain = []*DemoBlock{genesis}
	s.currentBlock = s.createNewBlock(genesis)
	s.updateState()
	return nil
}

func (s *BlockStructureSimulator) createGenesisBlock() *DemoBlock {
	block := &DemoBlock{
		Version:       1,
		PrevBlockHash: "0000000000000000000000000000000000000000000000000000000000000000",
		Timestamp:     time.Unix(0, 0),
		Difficulty:    1,
		Nonce:         0,
		Transactions:  []string{},
	}
	block.MerkleRoot = s.computeMerkleRoot(block.Transactions)
	block.Hash = s.computeBlockHash(block)
	return block
}

func (s *BlockStructureSimulator) createNewBlock(prevBlock *DemoBlock) *DemoBlock {
	block := &DemoBlock{
		Version:       1,
		PrevBlockHash: prevBlock.Hash,
		Timestamp:     time.Now(),
		Difficulty:    4,
		Nonce:         0,
		Transactions:  []string{},
	}
	block.MerkleRoot = s.computeMerkleRoot(block.Transactions)
	block.Hash = s.computeBlockHash(block)
	return block
}

func (s *BlockStructureSimulator) computeBlockHash(block *DemoBlock) string {
	header := fmt.Sprintf(
		"%d%s%s%d%d%d",
		block.Version,
		block.PrevBlockHash,
		block.MerkleRoot,
		block.Timestamp.Unix(),
		block.Difficulty,
		block.Nonce,
	)
	hash := sha256.Sum256([]byte(header))
	return hex.EncodeToString(hash[:])
}

func (s *BlockStructureSimulator) computeMerkleRoot(txs []string) string {
	if len(txs) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}

	hashes := make([]string, 0, len(txs))
	for _, tx := range txs {
		hash := sha256.Sum256([]byte(tx))
		hashes = append(hashes, hex.EncodeToString(hash[:]))
	}

	for len(hashes) > 1 {
		newLevel := make([]string, 0)
		for i := 0; i < len(hashes); i += 2 {
			combined := hashes[i] + hashes[i]
			if i+1 < len(hashes) {
				combined = hashes[i] + hashes[i+1]
			}
			hash := sha256.Sum256([]byte(combined))
			newLevel = append(newLevel, hex.EncodeToString(hash[:]))
		}
		hashes = newLevel
	}

	return hashes[0]
}

// SetField 修改当前候选区块的字段。
func (s *BlockStructureSimulator) SetField(field string, value interface{}) error {
	switch field {
	case "version":
		s.currentBlock.Version = uint32(parseNumericField(value))
	case "prev_block_hash":
		s.currentBlock.PrevBlockHash = parseStringField(value)
	case "timestamp":
		s.currentBlock.Timestamp = time.Unix(int64(parseNumericField(value)), 0)
	case "difficulty":
		s.currentBlock.Difficulty = uint32(parseNumericField(value))
	case "nonce":
		s.currentBlock.Nonce = uint64(parseNumericField(value))
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	oldHash := s.currentBlock.Hash
	s.currentBlock.Hash = s.computeBlockHash(s.currentBlock)

	s.EmitEvent("field_changed", "", "", map[string]interface{}{
		"field":    field,
		"value":    value,
		"old_hash": oldHash[:16],
		"new_hash": s.currentBlock.Hash[:16],
	})

	s.updateState()
	return nil
}

// AddTransaction 向当前候选区块添加交易。
func (s *BlockStructureSimulator) AddTransaction(tx string) {
	s.currentBlock.Transactions = append(s.currentBlock.Transactions, tx)
	oldMerkle := s.currentBlock.MerkleRoot
	s.currentBlock.MerkleRoot = s.computeMerkleRoot(s.currentBlock.Transactions)
	oldHash := s.currentBlock.Hash
	s.currentBlock.Hash = s.computeBlockHash(s.currentBlock)

	s.EmitEvent("transaction_added", "", "", map[string]interface{}{
		"tx":         tx,
		"old_merkle": oldMerkle[:16],
		"new_merkle": s.currentBlock.MerkleRoot[:16],
		"old_hash":   oldHash[:16],
		"new_hash":   s.currentBlock.Hash[:16],
	})

	s.updateState()
}

// MineBlock 模拟一次区块挖矿。
func (s *BlockStructureSimulator) MineBlock() bool {
	targetPrefix := ""
	for i := 0; i < int(s.currentBlock.Difficulty); i++ {
		targetPrefix += "0"
	}

	for i := uint64(0); i < 1000000; i++ {
		s.currentBlock.Nonce = i
		s.currentBlock.Hash = s.computeBlockHash(s.currentBlock)

		if len(s.currentBlock.Hash) >= int(s.currentBlock.Difficulty) &&
			s.currentBlock.Hash[:s.currentBlock.Difficulty] == targetPrefix {
			minedBlock := *s.currentBlock
			s.chain = append(s.chain, &minedBlock)
			s.currentBlock = s.createNewBlock(&minedBlock)

			s.EmitEvent("block_mined", "", "", map[string]interface{}{
				"nonce": i,
				"hash":  minedBlock.Hash,
			})
			s.updateState()
			return true
		}
	}
	return false
}

func parseStringField(value interface{}) string {
	if raw, ok := value.(string); ok {
		return raw
	}
	return fmt.Sprint(value)
}

func parseNumericField(value interface{}) float64 {
	switch raw := value.(type) {
	case float64:
		return raw
	case int:
		return float64(raw)
	case int64:
		return float64(raw)
	case string:
		var parsed float64
		fmt.Sscanf(raw, "%f", &parsed)
		return parsed
	default:
		return 0
	}
}

func (s *BlockStructureSimulator) updateState() {
	s.SetGlobalData("current_block", s.currentBlock)
	s.SetGlobalData("chain_length", len(s.chain))

	fields := []BlockField{
		{Name: "version", Value: s.currentBlock.Version, Description: "区块版本号", Editable: true},
		{Name: "prev_block_hash", Value: s.currentBlock.PrevBlockHash, Description: "前一区块哈希", Editable: true},
		{Name: "merkle_root", Value: s.currentBlock.MerkleRoot, Description: "交易 Merkle 根", Editable: false},
		{Name: "timestamp", Value: s.currentBlock.Timestamp.Unix(), Description: "时间戳", Editable: true},
		{Name: "difficulty", Value: s.currentBlock.Difficulty, Description: "难度目标", Editable: true},
		{Name: "nonce", Value: s.currentBlock.Nonce, Description: "随机数", Editable: true},
		{Name: "hash", Value: s.currentBlock.Hash, Description: "区块哈希", Editable: false},
	}
	s.SetGlobalData("fields", fields)

	summary := fmt.Sprintf("当前链长为 %d，候选区块包含 %d 笔交易。", len(s.chain), len(s.currentBlock.Transactions))
	nextHint := "可以修改区块字段、添加交易或尝试挖矿，观察区块哈希如何变化。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备区块",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{
			"chain_length": len(s.chain),
			"tx_count":     len(s.currentBlock.Transactions),
			"hash":         s.currentBlock.Hash,
		},
	)
}

// ExecuteAction 执行区块结构演示器动作。
func (s *BlockStructureSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "add_transaction":
		tx := fmt.Sprintf("tx-%d", time.Now().UnixNano())
		if raw, ok := params["tx"].(string); ok && raw != "" {
			tx = raw
		}
		s.AddTransaction(tx)
		return blockchainActionResult("已向当前区块添加一笔交易。", map[string]interface{}{"tx": tx}, &types.ActionFeedback{
			Summary:     "交易已写入候选区块，Merkle 根和区块哈希已随之更新。",
			NextHint:    "继续修改字段或执行挖矿，观察区块哈希如何随内容变化。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"tx_count": len(s.currentBlock.Transactions), "hash": s.currentBlock.Hash},
		}), nil
	case "mine_block":
		if !s.MineBlock() {
			return &types.ActionResult{
				Success: false,
				Message: "本轮未找到满足难度要求的区块哈希。",
			}, nil
		}
		return blockchainActionResult("已完成一次区块挖矿演示。", map[string]interface{}{"chain_length": len(s.chain)}, &types.ActionFeedback{
			Summary:     "候选区块已经满足难度要求并加入链中。",
			NextHint:    "继续观察新区块生成后，前一区块哈希如何衔接到下一轮候选区块。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"chain_length": len(s.chain)},
		}), nil
	case "set_field":
		field := "nonce"
		if raw, ok := params["field"].(string); ok && raw != "" {
			field = raw
		}
		if err := s.SetField(field, params["value"]); err != nil {
			return nil, err
		}
		return blockchainActionResult("已更新当前区块字段。", map[string]interface{}{"field": field, "hash": s.currentBlock.Hash}, &types.ActionFeedback{
			Summary:     "字段变化已经立即反映到区块哈希上。",
			NextHint:    "继续比较不同字段变化对区块哈希和链链接关系的影响。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"field": field, "hash": s.currentBlock.Hash},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported block action: %s", action)
	}
}

// BlockStructureFactory 提供工厂实现。
type BlockStructureFactory struct{}

func (f *BlockStructureFactory) Create() engine.Simulator {
	return NewBlockStructureSimulator()
}

func (f *BlockStructureFactory) GetDescription() types.Description {
	return NewBlockStructureSimulator().GetDescription()
}

func NewBlockStructureFactory() *BlockStructureFactory {
	return &BlockStructureFactory{}
}

var _ engine.SimulatorFactory = (*BlockStructureFactory)(nil)
