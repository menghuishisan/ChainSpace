package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// UTXO 未花费交易输出
type UTXO struct {
	TxID   string `json:"tx_id"`
	Index  int    `json:"index"`
	Value  uint64 `json:"value"`
	Owner  string `json:"owner"`
	Script string `json:"script"`
	Spent  bool   `json:"spent"`
}

// UTXOTx UTXO交易
type UTXOTx struct {
	TxID    string      `json:"tx_id"`
	Inputs  []*TxInput  `json:"inputs"`
	Outputs []*TxOutput `json:"outputs"`
}

// TxInput 交易输入
type TxInput struct {
	PrevTxID  string `json:"prev_tx_id"`
	PrevIndex int    `json:"prev_index"`
	Signature string `json:"signature"`
}

// TxOutput 交易输出
type TxOutput struct {
	Value  uint64 `json:"value"`
	Owner  string `json:"owner"`
	Script string `json:"script"`
}

// UTXOSimulator UTXO模型演示器
type UTXOSimulator struct {
	*base.BaseSimulator
	utxoSet map[string]*UTXO
	txs     []*UTXOTx
}

// NewUTXOSimulator 创建UTXO演示器
func NewUTXOSimulator() *UTXOSimulator {
	sim := &UTXOSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"utxo",
			"UTXO模型演示器",
			"展示比特币UTXO交易模型的工作原理",
			"blockchain",
			types.ComponentTool,
		),
		utxoSet: make(map[string]*UTXO),
		txs:     make([]*UTXOTx, 0),
	}
	return sim
}

// Init 初始化
func (s *UTXOSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.createCoinbase("alice", 100)
	s.createCoinbase("bob", 50)
	s.updateState()
	return nil
}

// createCoinbase 创建Coinbase交易
func (s *UTXOSimulator) createCoinbase(owner string, value uint64) {
	txID := s.generateTxID(fmt.Sprintf("coinbase-%s-%d", owner, len(s.txs)))
	utxo := &UTXO{
		TxID:  txID,
		Index: 0,
		Value: value,
		Owner: owner,
	}
	s.utxoSet[s.utxoKey(txID, 0)] = utxo
	s.EmitEvent("coinbase_created", "", "", map[string]interface{}{
		"tx_id": txID[:16], "owner": owner, "value": value,
	})
}

// CreateTransaction 创建UTXO交易
func (s *UTXOSimulator) CreateTransaction(from string, to string, amount uint64) (*UTXOTx, error) {
	inputs, total := s.selectUTXOs(from, amount)
	if total < amount {
		return nil, fmt.Errorf("insufficient funds: have %d, need %d", total, amount)
	}

	tx := &UTXOTx{
		Inputs:  make([]*TxInput, len(inputs)),
		Outputs: make([]*TxOutput, 0),
	}

	for i, utxo := range inputs {
		tx.Inputs[i] = &TxInput{
			PrevTxID:  utxo.TxID,
			PrevIndex: utxo.Index,
		}
	}

	tx.Outputs = append(tx.Outputs, &TxOutput{Value: amount, Owner: to})
	if change := total - amount; change > 0 {
		tx.Outputs = append(tx.Outputs, &TxOutput{Value: change, Owner: from})
	}

	tx.TxID = s.computeTxHash(tx)
	s.EmitEvent("utxo_tx_created", "", "", map[string]interface{}{
		"tx_id": tx.TxID[:16], "inputs": len(tx.Inputs), "outputs": len(tx.Outputs),
	})
	return tx, nil
}

// ExecuteTransaction 执行交易
func (s *UTXOSimulator) ExecuteTransaction(tx *UTXOTx) error {
	for _, input := range tx.Inputs {
		key := s.utxoKey(input.PrevTxID, input.PrevIndex)
		utxo := s.utxoSet[key]
		if utxo == nil || utxo.Spent {
			return fmt.Errorf("invalid input: %s", key)
		}
	}

	for _, input := range tx.Inputs {
		key := s.utxoKey(input.PrevTxID, input.PrevIndex)
		s.utxoSet[key].Spent = true
		delete(s.utxoSet, key)
	}

	for i, output := range tx.Outputs {
		utxo := &UTXO{
			TxID:  tx.TxID,
			Index: i,
			Value: output.Value,
			Owner: output.Owner,
		}
		s.utxoSet[s.utxoKey(tx.TxID, i)] = utxo
	}

	s.txs = append(s.txs, tx)
	s.EmitEvent("utxo_tx_executed", "", "", map[string]interface{}{"tx_id": tx.TxID[:16]})
	s.updateState()
	return nil
}

func (s *UTXOSimulator) selectUTXOs(owner string, amount uint64) ([]*UTXO, uint64) {
	var selected []*UTXO
	var total uint64
	for _, utxo := range s.utxoSet {
		if utxo.Owner == owner && !utxo.Spent {
			selected = append(selected, utxo)
			total += utxo.Value
			if total >= amount {
				break
			}
		}
	}
	return selected, total
}

func (s *UTXOSimulator) utxoKey(txID string, index int) string {
	return fmt.Sprintf("%s:%d", txID, index)
}

func (s *UTXOSimulator) generateTxID(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *UTXOSimulator) computeTxHash(tx *UTXOTx) string {
	data := ""
	for _, in := range tx.Inputs {
		data += fmt.Sprintf("%s:%d", in.PrevTxID, in.PrevIndex)
	}
	for _, out := range tx.Outputs {
		data += fmt.Sprintf("%s:%d", out.Owner, out.Value)
	}
	return s.generateTxID(data)
}

func (s *UTXOSimulator) updateState() {
	balances := make(map[string]uint64)
	utxoCount := 0
	for _, utxo := range s.utxoSet {
		if !utxo.Spent {
			balances[utxo.Owner] += utxo.Value
			utxoCount++
		}
	}
	s.SetGlobalData("balances", balances)
	s.SetGlobalData("utxo_count", utxoCount)
	s.SetGlobalData("tx_count", len(s.txs))

	summary := fmt.Sprintf("当前 UTXO 集中共有 %d 个未花费输出，已执行 %d 笔交易。", utxoCount, len(s.txs))
	nextHint := "可以继续创建 UTXO 交易，观察输入选择、找零输出和 UTXO 集变化。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备 UTXO",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"utxo_count": utxoCount, "tx_count": len(s.txs)},
	)
}

func (s *UTXOSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_transaction":
		from, to := "alice", "bob"
		amount := uint64(10)
		if raw, ok := params["from"].(string); ok && raw != "" {
			from = raw
		}
		if raw, ok := params["to"].(string); ok && raw != "" {
			to = raw
		}
		if raw, ok := params["amount"].(float64); ok && raw > 0 {
			amount = uint64(raw)
		}
		tx, err := s.CreateTransaction(from, to, amount)
		if err != nil {
			return nil, err
		}
		if err := s.ExecuteTransaction(tx); err != nil {
			return nil, err
		}
		return blockchainActionResult("已完成一笔 UTXO 交易。", map[string]interface{}{"tx": tx}, &types.ActionFeedback{
			Summary:     "交易输入已被花费，新输出和找零输出已经进入 UTXO 集。",
			NextHint:    "继续观察不同金额下输入选择和找零输出如何变化。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"tx_id": tx.TxID, "output_count": len(tx.Outputs)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported utxo action: %s", action)
	}
}

type UTXOFactory struct{}

func (f *UTXOFactory) Create() engine.Simulator          { return NewUTXOSimulator() }
func (f *UTXOFactory) GetDescription() types.Description { return NewUTXOSimulator().GetDescription() }
func NewUTXOFactory() *UTXOFactory                       { return &UTXOFactory{} }

var _ engine.SimulatorFactory = (*UTXOFactory)(nil)
