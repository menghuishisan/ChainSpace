package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// TrieNode MPT节点
type TrieNode struct {
	Hash     string               `json:"hash"`
	Type     string               `json:"type"` // branch, extension, leaf
	Key      string               `json:"key,omitempty"`
	Value    interface{}          `json:"value,omitempty"`
	Children map[string]*TrieNode `json:"children,omitempty"`
}

// StateTrieSimulator 状态树演示器
type StateTrieSimulator struct {
	*base.BaseSimulator
	root     *TrieNode
	accounts map[string]map[string]interface{}
}

// NewStateTrieSimulator 创建状态树演示器
func NewStateTrieSimulator() *StateTrieSimulator {
	sim := &StateTrieSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"state_trie",
			"状态树演示器",
			"展示以太坊MPT状态树的存储结构",
			"blockchain",
			types.ComponentTool,
		),
		accounts: make(map[string]map[string]interface{}),
	}
	return sim
}

// Init 初始化
func (s *StateTrieSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.root = &TrieNode{Type: "branch", Children: make(map[string]*TrieNode)}
	s.SetAccount("0x1234", 1000, 0, "")
	s.SetAccount("0x5678", 500, 0, "")
	return nil
}

// SetAccount 设置账户
func (s *StateTrieSimulator) SetAccount(addr string, balance uint64, nonce uint64, codeHash string) {
	account := map[string]interface{}{
		"balance": balance, "nonce": nonce, "codeHash": codeHash, "storageRoot": "",
	}
	s.accounts[addr] = account
	s.insertToTrie(addr, account)
	s.EmitEvent("account_set", "", "", map[string]interface{}{
		"address": addr, "balance": balance,
	})
	s.updateState()
}

// GetAccount 获取账户
func (s *StateTrieSimulator) GetAccount(addr string) map[string]interface{} {
	return s.accounts[addr]
}

func (s *StateTrieSimulator) insertToTrie(key string, value interface{}) {
	node := &TrieNode{Type: "leaf", Key: key, Value: value}
	node.Hash = s.hashNode(node)
	if s.root.Children == nil {
		s.root.Children = make(map[string]*TrieNode)
	}
	s.root.Children[key[:4]] = node
	s.root.Hash = s.hashNode(s.root)
}

func (s *StateTrieSimulator) hashNode(node *TrieNode) string {
	data := fmt.Sprintf("%s-%s-%v", node.Type, node.Key, node.Value)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func (s *StateTrieSimulator) updateState() {
	rootHash := ""
	if s.root != nil {
		rootHash = s.root.Hash
	}
	s.SetGlobalData("root_hash", rootHash)
	s.SetGlobalData("account_count", len(s.accounts))
	s.SetGlobalData("accounts", s.accounts)

	summary := fmt.Sprintf("当前状态树中有 %d 个账户节点。", len(s.accounts))
	nextHint := "可以继续写入账户，观察根哈希如何随着状态更新而变化。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备状态树",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"account_count": len(s.accounts), "root_hash": rootHash},
	)
}

func (s *StateTrieSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "set_account":
		address := "0x9999"
		balance := uint64(100)
		if raw, ok := params["address"].(string); ok && raw != "" {
			address = raw
		}
		if raw, ok := params["balance"].(float64); ok && raw >= 0 {
			balance = uint64(raw)
		}
		s.SetAccount(address, balance, 0, "")
		return blockchainActionResult("已写入一个状态树账户。", map[string]interface{}{"address": address, "balance": balance}, &types.ActionFeedback{
			Summary:     "账户状态已写入树中，根哈希已随之更新。",
			NextHint:    "继续修改不同账户余额，观察状态树根如何反映全局状态变化。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"address": address, "root_hash": s.root.Hash},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported state trie action: %s", action)
	}
}

type StateTrieFactory struct{}

func (f *StateTrieFactory) Create() engine.Simulator { return NewStateTrieSimulator() }
func (f *StateTrieFactory) GetDescription() types.Description {
	return NewStateTrieSimulator().GetDescription()
}
func NewStateTrieFactory() *StateTrieFactory { return &StateTrieFactory{} }

var _ engine.SimulatorFactory = (*StateTrieFactory)(nil)
