package blockchain

import (
	"fmt"
	"sort"
	"sync"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

// MempoolTx 内存池交易
type MempoolTx struct {
	Hash     string `json:"hash"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    uint64 `json:"value"`
	GasPrice uint64 `json:"gas_price"`
	GasLimit uint64 `json:"gas_limit"`
	Nonce    uint64 `json:"nonce"`
	Pending  bool   `json:"pending"`
}

// MempoolSimulator 交易池演示器
type MempoolSimulator struct {
	*base.BaseSimulator
	mu          sync.RWMutex
	pendingTxs  map[string]*MempoolTx
	queuedTxs   map[string]*MempoolTx
	nonces      map[string]uint64
	maxPoolSize int
	minGasPrice uint64
}

// NewMempoolSimulator 创建交易池演示器
func NewMempoolSimulator() *MempoolSimulator {
	sim := &MempoolSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"mempool",
			"交易池演示器",
			"展示交易池的排序、替换和打包策略",
			"blockchain",
			types.ComponentProcess,
		),
		pendingTxs:  make(map[string]*MempoolTx),
		queuedTxs:   make(map[string]*MempoolTx),
		nonces:      make(map[string]uint64),
		maxPoolSize: 100,
		minGasPrice: 1,
	}

	sim.AddParam(types.Param{
		Key: "max_pool_size", Name: "池容量", Type: types.ParamTypeInt,
		Default: 100, Min: 10, Max: 10000,
	})
	sim.AddParam(types.Param{
		Key: "min_gas_price", Name: "最低Gas价格", Type: types.ParamTypeInt,
		Default: 1, Min: 0, Max: 100,
	})
	sim.SetOnTick(sim.onTick)
	return sim
}

// Init 初始化
func (s *MempoolSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	if v, ok := config.Params["max_pool_size"]; ok {
		if n, ok := v.(float64); ok {
			s.maxPoolSize = int(n)
		}
	}
	if v, ok := config.Params["min_gas_price"]; ok {
		if n, ok := v.(float64); ok {
			s.minGasPrice = uint64(n)
		}
	}

	s.nonces = map[string]uint64{"alice": 0, "bob": 0, "charlie": 0}
	s.updateState()
	return nil
}

func (s *MempoolSimulator) onTick(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tick%10 == 0 && len(s.pendingTxs) > 0 {
		s.packBlock()
	}

	s.promoteQueuedTxs()
	s.updateState()
	return nil
}

// AddTransaction 添加交易
func (s *MempoolSimulator) AddTransaction(from, to string, value, gasPrice, gasLimit uint64) (*MempoolTx, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if gasPrice < s.minGasPrice {
		return nil, fmt.Errorf("gas price too low")
	}

	if len(s.pendingTxs)+len(s.queuedTxs) >= s.maxPoolSize {
		s.evictLowPriceTx()
	}

	expectedNonce := s.nonces[from]
	tx := &MempoolTx{
		Hash:     uuid.New().String()[:16],
		From:     from,
		To:       to,
		Value:    value,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Nonce:    expectedNonce,
		Pending:  true,
	}

	s.pendingTxs[tx.Hash] = tx
	s.nonces[from]++

	s.EmitEvent("tx_added", "", "", map[string]interface{}{
		"hash": tx.Hash, "from": from, "gas_price": gasPrice,
	})
	s.updateState()
	return tx, nil
}

// ReplaceTx 替换交易(提高Gas价格)
func (s *MempoolSimulator) ReplaceTx(oldHash string, newGasPrice uint64) (*MempoolTx, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldTx := s.pendingTxs[oldHash]
	if oldTx == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	if newGasPrice <= oldTx.GasPrice*110/100 {
		return nil, fmt.Errorf("new gas price must be at least 10%% higher")
	}

	newTx := &MempoolTx{
		Hash:     uuid.New().String()[:16],
		From:     oldTx.From,
		To:       oldTx.To,
		Value:    oldTx.Value,
		GasPrice: newGasPrice,
		GasLimit: oldTx.GasLimit,
		Nonce:    oldTx.Nonce,
		Pending:  true,
	}

	delete(s.pendingTxs, oldHash)
	s.pendingTxs[newTx.Hash] = newTx

	s.EmitEvent("tx_replaced", "", "", map[string]interface{}{
		"old_hash": oldHash, "new_hash": newTx.Hash, "new_gas_price": newGasPrice,
	})
	s.updateState()
	return newTx, nil
}

func (s *MempoolSimulator) packBlock() {
	var txList []*MempoolTx
	for _, tx := range s.pendingTxs {
		txList = append(txList, tx)
	}

	sort.Slice(txList, func(i, j int) bool {
		return txList[i].GasPrice > txList[j].GasPrice
	})

	packCount := 5
	if len(txList) < packCount {
		packCount = len(txList)
	}

	for i := 0; i < packCount; i++ {
		delete(s.pendingTxs, txList[i].Hash)
	}

	s.EmitEvent("block_packed", "", "", map[string]interface{}{
		"tx_count": packCount,
	})
}

func (s *MempoolSimulator) promoteQueuedTxs() {
	for hash, tx := range s.queuedTxs {
		if tx.Nonce == s.nonces[tx.From] {
			delete(s.queuedTxs, hash)
			s.pendingTxs[hash] = tx
			tx.Pending = true
		}
	}
}

func (s *MempoolSimulator) evictLowPriceTx() {
	var lowestHash string
	var lowestPrice uint64 = ^uint64(0)

	for hash, tx := range s.pendingTxs {
		if tx.GasPrice < lowestPrice {
			lowestPrice = tx.GasPrice
			lowestHash = hash
		}
	}

	if lowestHash != "" {
		delete(s.pendingTxs, lowestHash)
		s.EmitEvent("tx_evicted", "", "", map[string]interface{}{
			"hash": lowestHash, "gas_price": lowestPrice,
		})
	}
}

func (s *MempoolSimulator) updateState() {
	s.SetGlobalData("pending_count", len(s.pendingTxs))
	s.SetGlobalData("queued_count", len(s.queuedTxs))
	s.SetGlobalData("max_pool_size", s.maxPoolSize)
	s.SetGlobalData("min_gas_price", s.minGasPrice)

	summary := fmt.Sprintf("当前内存池中有 %d 笔待打包交易，%d 笔排队交易。", len(s.pendingTxs), len(s.queuedTxs))
	nextHint := "可以继续添加交易、提高 Gas 价格替换交易，或观察打包后池状态如何变化。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备交易池",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"pending_count": len(s.pendingTxs), "queued_count": len(s.queuedTxs)},
	)
}

func (s *MempoolSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "add_transaction":
		tx, err := s.AddTransaction("alice", "bob", 10, maxUint64(s.minGasPrice, 1), 21000)
		if err != nil {
			return nil, err
		}
		return blockchainActionResult("已向交易池添加一笔交易。", map[string]interface{}{"tx": tx}, &types.ActionFeedback{
			Summary:     "交易已进入 pending 队列，后续将根据 Gas 价格参与打包排序。",
			NextHint:    "继续添加更高 Gas 的交易，观察交易替换和打包优先级变化。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"pending_count": len(s.pendingTxs)},
		}), nil
	case "pack_block":
		s.packBlock()
		s.promoteQueuedTxs()
		s.updateState()
		return blockchainActionResult("已模拟一次交易打包。", map[string]interface{}{"pending_count": len(s.pendingTxs), "queued_count": len(s.queuedTxs)}, &types.ActionFeedback{
			Summary:     "优先级较高的交易已被打包，交易池状态已更新。",
			NextHint:    "继续观察不同 Gas 价格对打包顺序和替换行为的影响。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"pending_count": len(s.pendingTxs), "queued_count": len(s.queuedTxs)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported mempool action: %s", action)
	}
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

type MempoolFactory struct{}

func (f *MempoolFactory) Create() engine.Simulator { return NewMempoolSimulator() }
func (f *MempoolFactory) GetDescription() types.Description {
	return NewMempoolSimulator().GetDescription()
}
func NewMempoolFactory() *MempoolFactory { return &MempoolFactory{} }

var _ engine.SimulatorFactory = (*MempoolFactory)(nil)
