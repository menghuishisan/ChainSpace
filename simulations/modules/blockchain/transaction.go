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

// DemoTransaction 演示交易
type DemoTransaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     uint64    `json:"value"`
	Nonce     uint64    `json:"nonce"`
	GasPrice  uint64    `json:"gas_price"`
	GasLimit  uint64    `json:"gas_limit"`
	Data      string    `json:"data"`
	Signature string    `json:"signature"`
	Timestamp time.Time `json:"timestamp"`
}

// TransactionSimulator 交易演示器
type TransactionSimulator struct {
	*base.BaseSimulator
	transactions []*DemoTransaction
	accounts     map[string]uint64
	nonces       map[string]uint64
}

// NewTransactionSimulator 创建交易演示器
func NewTransactionSimulator() *TransactionSimulator {
	sim := &TransactionSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"transaction",
			"交易结构演示器",
			"展示交易的构造、签名和验证过程",
			"blockchain",
			types.ComponentTool,
		),
		transactions: make([]*DemoTransaction, 0),
		accounts:     make(map[string]uint64),
		nonces:       make(map[string]uint64),
	}
	return sim
}

// Init 初始化
func (s *TransactionSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.accounts = map[string]uint64{
		"alice": 1000, "bob": 500, "charlie": 300,
	}
	s.nonces = map[string]uint64{"alice": 0, "bob": 0, "charlie": 0}
	s.updateState()
	return nil
}

// CreateTransaction 创建交易
func (s *TransactionSimulator) CreateTransaction(from, to string, value, gasPrice, gasLimit uint64, data string) (*DemoTransaction, error) {
	balance := s.accounts[from]
	if balance < value+gasPrice*gasLimit {
		return nil, fmt.Errorf("insufficient balance")
	}

	nonce := s.nonces[from]
	tx := &DemoTransaction{
		From:      from,
		To:        to,
		Value:     value,
		Nonce:     nonce,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      data,
		Timestamp: time.Now(),
	}
	tx.Hash = s.computeTxHash(tx)
	s.EmitEvent("tx_created", "", "", map[string]interface{}{
		"hash": tx.Hash[:16], "from": from, "to": to, "value": value,
	})
	return tx, nil
}

// SignTransaction 签名交易
func (s *TransactionSimulator) SignTransaction(tx *DemoTransaction, privateKey string) {
	sigData := fmt.Sprintf("%s-%s-%s", tx.Hash, privateKey, tx.From)
	hash := sha256.Sum256([]byte(sigData))
	tx.Signature = hex.EncodeToString(hash[:])
	s.EmitEvent("tx_signed", "", "", map[string]interface{}{
		"hash": tx.Hash[:16], "signature": tx.Signature[:16],
	})
}

// ExecuteTransaction 执行交易
func (s *TransactionSimulator) ExecuteTransaction(tx *DemoTransaction) error {
	if tx.Signature == "" {
		return fmt.Errorf("transaction not signed")
	}
	if s.nonces[tx.From] != tx.Nonce {
		return fmt.Errorf("invalid nonce")
	}
	totalCost := tx.Value + tx.GasPrice*tx.GasLimit
	if s.accounts[tx.From] < totalCost {
		return fmt.Errorf("insufficient balance")
	}

	s.accounts[tx.From] -= totalCost
	s.accounts[tx.To] += tx.Value
	s.nonces[tx.From]++
	s.transactions = append(s.transactions, tx)

	s.EmitEvent("tx_executed", "", "", map[string]interface{}{
		"hash": tx.Hash[:16], "from": tx.From, "to": tx.To, "value": tx.Value,
	})
	s.updateState()
	return nil
}

func (s *TransactionSimulator) computeTxHash(tx *DemoTransaction) string {
	data := fmt.Sprintf("%s%s%d%d%d%d%s", tx.From, tx.To, tx.Value, tx.Nonce, tx.GasPrice, tx.GasLimit, tx.Data)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *TransactionSimulator) updateState() {
	s.SetGlobalData("accounts", s.accounts)
	s.SetGlobalData("nonces", s.nonces)
	s.SetGlobalData("tx_count", len(s.transactions))

	summary := fmt.Sprintf("当前账户数为 %d，已执行 %d 笔交易。", len(s.accounts), len(s.transactions))
	nextHint := "可以创建交易、签名并执行，观察余额、Nonce 与 Gas 成本如何变化。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备交易",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"tx_count": len(s.transactions), "account_count": len(s.accounts)},
	)
}

func (s *TransactionSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_transaction":
		from, to := "alice", "bob"
		value := uint64(10)
		if raw, ok := params["from"].(string); ok && raw != "" {
			from = raw
		}
		if raw, ok := params["to"].(string); ok && raw != "" {
			to = raw
		}
		if raw, ok := params["value"].(float64); ok && raw > 0 {
			value = uint64(raw)
		}
		tx, err := s.CreateTransaction(from, to, value, 1, 21000, "")
		if err != nil {
			return nil, err
		}
		s.SignTransaction(tx, from+"-private-key")
		if err := s.ExecuteTransaction(tx); err != nil {
			return nil, err
		}
		return blockchainActionResult("已完成一笔交易的创建、签名与执行。", map[string]interface{}{"tx": tx}, &types.ActionFeedback{
			Summary:     "交易已经从构造走到执行完成，余额和 Nonce 已同步更新。",
			NextHint:    "继续观察多笔交易执行后的余额变化和 Nonce 递增。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"tx_hash": tx.Hash, "from": from, "to": to},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported transaction action: %s", action)
	}
}

type TransactionFactory struct{}

func (f *TransactionFactory) Create() engine.Simulator { return NewTransactionSimulator() }
func (f *TransactionFactory) GetDescription() types.Description {
	return NewTransactionSimulator().GetDescription()
}
func NewTransactionFactory() *TransactionFactory { return &TransactionFactory{} }

var _ engine.SimulatorFactory = (*TransactionFactory)(nil)
