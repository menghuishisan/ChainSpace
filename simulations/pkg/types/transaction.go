package types

import (
	"crypto/sha256"
	"encoding/json"
	"math/big"
	"time"
)

// Transaction 交易
type Transaction struct {
	Hash      Hash      `json:"hash"`
	From      Address   `json:"from"`
	To        Address   `json:"to"`
	Value     *big.Int  `json:"value"`
	Nonce     uint64    `json:"nonce"`
	GasPrice  *big.Int  `json:"gas_price"`
	GasLimit  uint64    `json:"gas_limit"`
	Data      []byte    `json:"data"`
	Signature Signature `json:"signature"`
	Timestamp time.Time `json:"timestamp"`
}

// TxType 交易类型
type TxType uint8

const (
	TxTypeTransfer TxType = iota // 转账
	TxTypeContract               // 合约调用
	TxTypeDeploy                 // 合约部署
)

// NewTransaction 创建新交易
func NewTransaction(from, to Address, value *big.Int, nonce uint64, data []byte) *Transaction {
	tx := &Transaction{
		From:      from,
		To:        to,
		Value:     value,
		Nonce:     nonce,
		GasPrice:  big.NewInt(1000000000), // 1 Gwei
		GasLimit:  21000,
		Data:      data,
		Timestamp: time.Now(),
	}
	tx.Hash = tx.CalculateHash()
	return tx
}

// CalculateHash 计算交易哈希
func (tx *Transaction) CalculateHash() Hash {
	data := struct {
		From     Address  `json:"from"`
		To       Address  `json:"to"`
		Value    *big.Int `json:"value"`
		Nonce    uint64   `json:"nonce"`
		GasPrice *big.Int `json:"gas_price"`
		GasLimit uint64   `json:"gas_limit"`
		Data     []byte   `json:"data"`
	}{
		From:     tx.From,
		To:       tx.To,
		Value:    tx.Value,
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		Data:     tx.Data,
	}

	encoded, _ := json.Marshal(data)
	return sha256.Sum256(encoded)
}

// Type 获取交易类型
func (tx *Transaction) Type() TxType {
	if tx.To.IsEmpty() {
		return TxTypeDeploy
	}
	if len(tx.Data) > 0 {
		return TxTypeContract
	}
	return TxTypeTransfer
}

// GasCost 计算Gas费用
func (tx *Transaction) GasCost() *big.Int {
	return new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(tx.GasLimit)))
}

// TxReceipt 交易收据
type TxReceipt struct {
	TxHash           Hash    `json:"tx_hash"`
	BlockHash        Hash    `json:"block_hash"`
	BlockNumber      uint64  `json:"block_number"`
	TransactionIndex uint64  `json:"transaction_index"`
	From             Address `json:"from"`
	To               Address `json:"to"`
	GasUsed          uint64  `json:"gas_used"`
	Status           uint64  `json:"status"` // 0 = failed, 1 = success
	Logs             []Log   `json:"logs"`
	ContractAddress  Address `json:"contract_address,omitempty"`
}

// Log 事件日志
type Log struct {
	Address     Address `json:"address"`
	Topics      []Hash  `json:"topics"`
	Data        []byte  `json:"data"`
	BlockNumber uint64  `json:"block_number"`
	TxHash      Hash    `json:"tx_hash"`
	TxIndex     uint64  `json:"tx_index"`
	LogIndex    uint64  `json:"log_index"`
}

// TxPool 交易池
type TxPool struct {
	Pending map[Address][]*Transaction `json:"pending"`
	Queued  map[Address][]*Transaction `json:"queued"`
}

// NewTxPool 创建交易池
func NewTxPool() *TxPool {
	return &TxPool{
		Pending: make(map[Address][]*Transaction),
		Queued:  make(map[Address][]*Transaction),
	}
}

// Add 添加交易到池
func (p *TxPool) Add(tx *Transaction) {
	p.Pending[tx.From] = append(p.Pending[tx.From], tx)
}

// GetPending 获取待处理交易
func (p *TxPool) GetPending(limit int) []*Transaction {
	var txs []*Transaction
	for _, pending := range p.Pending {
		for _, tx := range pending {
			txs = append(txs, tx)
			if len(txs) >= limit {
				return txs
			}
		}
	}
	return txs
}

// Remove 从池中移除交易
func (p *TxPool) Remove(hash Hash) {
	for addr, txs := range p.Pending {
		for i, tx := range txs {
			if tx.Hash == hash {
				p.Pending[addr] = append(txs[:i], txs[i+1:]...)
				return
			}
		}
	}
}

// UTXO 未花费交易输出（比特币模型）
type UTXO struct {
	TxHash Hash   `json:"tx_hash"`
	Index  uint32 `json:"index"`
	Value  uint64 `json:"value"`
	Script []byte `json:"script"`
}

// TxInput 交易输入（比特币）
type TxInput struct {
	PrevTxHash Hash   `json:"prev_tx_hash"`
	PrevIndex  uint32 `json:"prev_index"`
	ScriptSig  []byte `json:"script_sig"`
	Sequence   uint32 `json:"sequence"`
}

// TxOutput 交易输出（比特币）
type TxOutput struct {
	Value        uint64 `json:"value"`
	ScriptPubKey []byte `json:"script_pub_key"`
}

// BitcoinTx 比特币交易
type BitcoinTx struct {
	Hash     Hash       `json:"hash"`
	Version  uint32     `json:"version"`
	Inputs   []TxInput  `json:"inputs"`
	Outputs  []TxOutput `json:"outputs"`
	LockTime uint32     `json:"lock_time"`
}
