package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 状态差异演示器
// =============================================================================

// AccountDiff 账户差异
type AccountDiff struct {
	Address     string           `json:"address"`
	BalanceDiff *BalanceChange   `json:"balance_diff,omitempty"`
	NonceDiff   *NonceChange     `json:"nonce_diff,omitempty"`
	CodeDiff    *CodeChange      `json:"code_diff,omitempty"`
	StorageDiff []*StorageChange `json:"storage_diff,omitempty"`
}

// BalanceChange 余额变化
type BalanceChange struct {
	Before string `json:"before"`
	After  string `json:"after"`
	Delta  string `json:"delta"`
}

// NonceChange Nonce变化
type NonceChange struct {
	Before uint64 `json:"before"`
	After  uint64 `json:"after"`
}

// CodeChange 代码变化
type CodeChange struct {
	Action string `json:"action"` // deployed, destroyed
	Size   int    `json:"size"`
}

// StorageChange 存储变化
type StorageChange struct {
	Slot   string `json:"slot"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// TransactionStateDiff 交易状态差异
type TransactionStateDiff struct {
	TxHash       string         `json:"tx_hash"`
	BlockNumber  uint64         `json:"block_number"`
	From         string         `json:"from"`
	To           string         `json:"to"`
	Value        string         `json:"value"`
	GasUsed      uint64         `json:"gas_used"`
	AccountDiffs []*AccountDiff `json:"account_diffs"`
	LogCount     int            `json:"log_count"`
	Timestamp    time.Time      `json:"timestamp"`
}

// StateDiffSimulator 状态差异演示器
// 追踪交易执行前后的状态变化
type StateDiffSimulator struct {
	*base.BaseSimulator
	diffs []*TransactionStateDiff
}

// NewStateDiffSimulator 创建状态差异演示器
func NewStateDiffSimulator() *StateDiffSimulator {
	sim := &StateDiffSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"state_diff",
			"状态差异演示器",
			"追踪交易执行前后的状态变化",
			"evm",
			types.ComponentDemo,
		),
		diffs: make([]*TransactionStateDiff, 0),
	}

	return sim
}

// Init 初始化
func (s *StateDiffSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// SimulateETHTransfer 模拟ETH转账的状态变化
func (s *StateDiffSimulator) SimulateETHTransfer(from, to string, amountWei string) *TransactionStateDiff {
	amount, _ := new(big.Int).SetString(amountWei, 10)
	gasCost := big.NewInt(21000 * 20e9) // 21000 gas * 20 Gwei

	// 假设初始余额
	fromBalanceBefore := big.NewInt(1e18) // 1 ETH
	toBalanceBefore := big.NewInt(0)

	fromBalanceAfter := new(big.Int).Sub(fromBalanceBefore, amount)
	fromBalanceAfter.Sub(fromBalanceAfter, gasCost)
	toBalanceAfter := new(big.Int).Add(toBalanceBefore, amount)

	diff := &TransactionStateDiff{
		TxHash:      "0x" + hex.EncodeToString(make([]byte, 32)),
		BlockNumber: 12345678,
		From:        from,
		To:          to,
		Value:       amountWei,
		GasUsed:     21000,
		AccountDiffs: []*AccountDiff{
			{
				Address: from,
				BalanceDiff: &BalanceChange{
					Before: fromBalanceBefore.String(),
					After:  fromBalanceAfter.String(),
					Delta:  new(big.Int).Sub(fromBalanceAfter, fromBalanceBefore).String(),
				},
				NonceDiff: &NonceChange{
					Before: 0,
					After:  1,
				},
			},
			{
				Address: to,
				BalanceDiff: &BalanceChange{
					Before: toBalanceBefore.String(),
					After:  toBalanceAfter.String(),
					Delta:  amount.String(),
				},
			},
		},
		LogCount:  0,
		Timestamp: time.Now(),
	}

	s.diffs = append(s.diffs, diff)

	s.EmitEvent("state_diff_created", "", "", map[string]interface{}{
		"type":             "eth_transfer",
		"accounts_changed": 2,
	})

	s.updateState()
	return diff
}

// SimulateERC20Transfer 模拟ERC20转账的状态变化
func (s *StateDiffSimulator) SimulateERC20Transfer(tokenAddr, from, to string, amount string) *TransactionStateDiff {
	// 计算存储槽
	// balances[from] -> slot = keccak256(from . 0)
	// balances[to] -> slot = keccak256(to . 0)

	fromSlot := fmt.Sprintf("0x%064x", 1) // 简化
	toSlot := fmt.Sprintf("0x%064x", 2)

	diff := &TransactionStateDiff{
		TxHash:      "0x" + hex.EncodeToString(make([]byte, 32)),
		BlockNumber: 12345678,
		From:        from,
		To:          tokenAddr,
		Value:       "0",
		GasUsed:     65000,
		AccountDiffs: []*AccountDiff{
			{
				Address: from,
				NonceDiff: &NonceChange{
					Before: 5,
					After:  6,
				},
				BalanceDiff: &BalanceChange{
					Before: "1000000000000000000",
					After:  "999870000000000000", // 减去gas
					Delta:  "-130000000000000",
				},
			},
			{
				Address: tokenAddr,
				StorageDiff: []*StorageChange{
					{
						Slot:   fromSlot,
						Before: fmt.Sprintf("0x%064s", "1000"),
						After:  fmt.Sprintf("0x%064s", "0"),
					},
					{
						Slot:   toSlot,
						Before: fmt.Sprintf("0x%064s", "0"),
						After:  fmt.Sprintf("0x%064s", "1000"),
					},
				},
			},
		},
		LogCount:  1, // Transfer event
		Timestamp: time.Now(),
	}

	s.diffs = append(s.diffs, diff)

	s.EmitEvent("state_diff_created", "", "", map[string]interface{}{
		"type":            "erc20_transfer",
		"storage_changes": 2,
		"log_count":       1,
	})

	s.updateState()
	return diff
}

// SimulateContractDeployment 模拟合约部署的状态变化
func (s *StateDiffSimulator) SimulateContractDeployment(deployer string, codeSize int) *TransactionStateDiff {
	contractAddr := "0x" + hex.EncodeToString(make([]byte, 20))

	diff := &TransactionStateDiff{
		TxHash:      "0x" + hex.EncodeToString(make([]byte, 32)),
		BlockNumber: 12345678,
		From:        deployer,
		To:          "", // 合约创建
		Value:       "0",
		GasUsed:     uint64(32000 + codeSize*200),
		AccountDiffs: []*AccountDiff{
			{
				Address: deployer,
				NonceDiff: &NonceChange{
					Before: 10,
					After:  11,
				},
				BalanceDiff: &BalanceChange{
					Before: "10000000000000000000",
					After:  "9990000000000000000",
					Delta:  "-10000000000000000",
				},
			},
			{
				Address: contractAddr,
				CodeDiff: &CodeChange{
					Action: "deployed",
					Size:   codeSize,
				},
				StorageDiff: []*StorageChange{
					{
						Slot:   "0x0000000000000000000000000000000000000000000000000000000000000000",
						Before: "0x0000000000000000000000000000000000000000000000000000000000000000",
						After:  "0x000000000000000000000000" + deployer[2:], // owner
					},
				},
			},
		},
		LogCount:  0,
		Timestamp: time.Now(),
	}

	s.diffs = append(s.diffs, diff)

	s.EmitEvent("state_diff_created", "", "", map[string]interface{}{
		"type":             "contract_deployment",
		"contract_address": contractAddr,
		"code_size":        codeSize,
	})

	s.updateState()
	return diff
}

// SimulateDeFiSwap 模拟DeFi交换的状态变化
func (s *StateDiffSimulator) SimulateDeFiSwap(user, router, pair, token0, token1 string) *TransactionStateDiff {
	diff := &TransactionStateDiff{
		TxHash:      "0x" + hex.EncodeToString(make([]byte, 32)),
		BlockNumber: 12345678,
		From:        user,
		To:          router,
		Value:       "1000000000000000000", // 1 ETH
		GasUsed:     180000,
		AccountDiffs: []*AccountDiff{
			{
				Address:   user,
				NonceDiff: &NonceChange{Before: 100, After: 101},
				BalanceDiff: &BalanceChange{
					Before: "10000000000000000000",
					After:  "8996400000000000000",
					Delta:  "-1003600000000000000", // 1 ETH + gas
				},
			},
			{
				Address: pair,
				StorageDiff: []*StorageChange{
					{Slot: "reserve0", Before: "1000000", After: "1001000"},
					{Slot: "reserve1", Before: "500000000000000000000", After: "499500000000000000000"},
					{Slot: "blockTimestampLast", Before: "1699000000", After: "1699001000"},
				},
			},
			{
				Address: token0,
				StorageDiff: []*StorageChange{
					{Slot: "balances[pair]", Before: "1000000", After: "1001000"},
				},
			},
			{
				Address: token1,
				StorageDiff: []*StorageChange{
					{Slot: "balances[pair]", Before: "500000000000000000000", After: "499500000000000000000"},
					{Slot: "balances[user]", Before: "0", After: "500000000000000000"},
				},
			},
		},
		LogCount:  4, // Deposit, Transfer, Transfer, Swap
		Timestamp: time.Now(),
	}

	s.diffs = append(s.diffs, diff)

	s.EmitEvent("state_diff_created", "", "", map[string]interface{}{
		"type":             "defi_swap",
		"accounts_changed": 4,
		"log_count":        4,
	})

	s.updateState()
	return diff
}

// FormatDiffAsText 格式化差异为文本
func (s *StateDiffSimulator) FormatDiffAsText(diff *TransactionStateDiff) []string {
	lines := make([]string, 0)

	lines = append(lines, fmt.Sprintf("Transaction: %s", diff.TxHash[:18]+"..."))
	lines = append(lines, fmt.Sprintf("Block: %d, Gas Used: %d", diff.BlockNumber, diff.GasUsed))
	lines = append(lines, "")

	for _, acc := range diff.AccountDiffs {
		lines = append(lines, fmt.Sprintf("Account: %s", acc.Address))

		if acc.BalanceDiff != nil {
			lines = append(lines, fmt.Sprintf("  Balance: %s -> %s (Δ%s)",
				acc.BalanceDiff.Before, acc.BalanceDiff.After, acc.BalanceDiff.Delta))
		}

		if acc.NonceDiff != nil {
			lines = append(lines, fmt.Sprintf("  Nonce: %d -> %d",
				acc.NonceDiff.Before, acc.NonceDiff.After))
		}

		if acc.CodeDiff != nil {
			lines = append(lines, fmt.Sprintf("  Code: %s (%d bytes)",
				acc.CodeDiff.Action, acc.CodeDiff.Size))
		}

		for _, storage := range acc.StorageDiff {
			lines = append(lines, fmt.Sprintf("  Storage[%s]: %s -> %s",
				storage.Slot[:10]+"...", storage.Before[:10]+"...", storage.After[:10]+"..."))
		}

		lines = append(lines, "")
	}

	return lines
}

// updateState 更新状态
func (s *StateDiffSimulator) updateState() {
	s.SetGlobalData("diff_count", len(s.diffs))

	if len(s.diffs) > 0 {
		last := s.diffs[len(s.diffs)-1]
		s.SetGlobalData("last_diff", map[string]interface{}{
			"tx_hash":   last.TxHash[:18] + "...",
			"gas_used":  last.GasUsed,
			"accounts":  len(last.AccountDiffs),
			"log_count": last.LogCount,
		})
		setEVMTeachingState(
			s.BaseSimulator,
			"evm",
			"state_diff_completed",
			fmt.Sprintf("最近一次状态差异涉及 %d 个账户，Gas 消耗为 %d。", len(last.AccountDiffs), last.GasUsed),
			"继续查看余额、Nonce、代码和存储槽的变化，判断一次交易究竟改写了哪些状态。",
			0.85,
			map[string]interface{}{
				"tx_hash":   last.TxHash,
				"gas_used":  last.GasUsed,
				"accounts":  len(last.AccountDiffs),
				"log_count": last.LogCount,
			},
		)
		return
	}

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		"state_diff_ready",
		"当前还没有生成状态差异，可以先模拟一次 ETH 转账、ERC20 转账或合约部署。",
		"重点对比交易执行前后哪些账户被修改、哪些存储槽发生了变化。",
		0.2,
		map[string]interface{}{"diff_count": 0},
	)
}

// ExecuteAction 为状态差异实验提供交互动作。
func (s *StateDiffSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_eth_transfer":
		from, _ := params["from"].(string)
		to, _ := params["to"].(string)
		amount, _ := params["amount"].(string)
		if from == "" {
			from = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}
		if to == "" {
			to = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		}
		if amount == "" {
			amount = "100000000000000000"
		}
		diff := s.SimulateETHTransfer(from, to, amount)
		return evmActionResult("已模拟一笔 ETH 转账状态差异。", map[string]interface{}{"tx_hash": diff.TxHash, "accounts": len(diff.AccountDiffs)}, &types.ActionFeedback{
			Summary:     "一笔原生转账对余额和 nonce 的影响已经生成。",
			NextHint:    "继续观察发送方如何同时承担转账金额和 gas 成本。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"tx_hash": diff.TxHash},
		}), nil
	case "simulate_erc20_transfer":
		diff := s.SimulateERC20Transfer("0xtoken", "0xsender", "0xreceiver", "1000")
		return evmActionResult("已模拟一笔 ERC20 转账状态差异。", map[string]interface{}{"tx_hash": diff.TxHash, "accounts": len(diff.AccountDiffs), "logs": diff.LogCount}, &types.ActionFeedback{
			Summary:     "ERC20 转账带来的账户与存储槽变化已经生成。",
			NextHint:    "重点观察 token 合约内部的 balances 映射如何被改写，以及 Transfer 事件如何出现。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"tx_hash": diff.TxHash},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported state diff action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// StateDiffFactory 状态差异工厂
type StateDiffFactory struct{}

func (f *StateDiffFactory) Create() engine.Simulator {
	return NewStateDiffSimulator()
}

func (f *StateDiffFactory) GetDescription() types.Description {
	return NewStateDiffSimulator().GetDescription()
}

func NewStateDiffFactory() *StateDiffFactory {
	return &StateDiffFactory{}
}

var _ engine.SimulatorFactory = (*StateDiffFactory)(nil)
