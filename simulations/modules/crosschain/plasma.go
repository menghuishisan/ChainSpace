package crosschain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// Plasma演示器
// 演示Plasma子链的核心机制
//
// 核心概念:
// 1. 子链: 独立的区块链，定期提交状态到主链
// 2. 退出机制: 用户可随时将资产退出到主链
// 3. 挑战期: 防止无效退出的挑战机制
// 4. 数据可用性: 子链运营者需保证数据可用
//
// 参考: Plasma MVP, Plasma Cash, OMG Network
// =============================================================================

// PlasmaBlockStatus 区块状态
type PlasmaBlockStatus string

const (
	PlasmaBlockPending   PlasmaBlockStatus = "pending"
	PlasmaBlockSubmitted PlasmaBlockStatus = "submitted"
	PlasmaBlockFinalized PlasmaBlockStatus = "finalized"
)

// ExitStatus 退出状态
type ExitStatus string

const (
	ExitPending    ExitStatus = "pending"
	ExitChallenged ExitStatus = "challenged"
	ExitFinalized  ExitStatus = "finalized"
	ExitCanceled   ExitStatus = "canceled"
)

// PlasmaBlock 子链区块
type PlasmaBlock struct {
	BlockNumber uint64            `json:"block_number"`
	MerkleRoot  string            `json:"merkle_root"`
	TxCount     int               `json:"tx_count"`
	Operator    string            `json:"operator"`
	Status      PlasmaBlockStatus `json:"status"`
	SubmittedAt time.Time         `json:"submitted_at"`
	L1TxHash    string            `json:"l1_tx_hash"`
}

// PlasmaUTXO UTXO
type PlasmaUTXO struct {
	UTXOID      string   `json:"utxo_id"`
	Owner       string   `json:"owner"`
	Amount      *big.Int `json:"amount"`
	Token       string   `json:"token"`
	BlockNumber uint64   `json:"block_number"`
	TxIndex     int      `json:"tx_index"`
	OutputIndex int      `json:"output_index"`
	Spent       bool     `json:"spent"`
}

// ExitRequest 退出请求
type ExitRequest struct {
	ExitID       string     `json:"exit_id"`
	UTXOID       string     `json:"utxo_id"`
	Owner        string     `json:"owner"`
	Amount       *big.Int   `json:"amount"`
	Status       ExitStatus `json:"status"`
	ExitBond     *big.Int   `json:"exit_bond"`
	RequestedAt  time.Time  `json:"requested_at"`
	ChallengeEnd time.Time  `json:"challenge_end"`
	FinalizedAt  time.Time  `json:"finalized_at"`
}

// PlasmaSimulator Plasma演示器
type PlasmaSimulator struct {
	*base.BaseSimulator
	blocks          map[uint64]*PlasmaBlock
	utxos           map[string]*PlasmaUTXO
	exits           map[string]*ExitRequest
	currentBlock    uint64
	exitBond        *big.Int
	challengePeriod time.Duration
}

// NewPlasmaSimulator 创建演示器
func NewPlasmaSimulator() *PlasmaSimulator {
	sim := &PlasmaSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"plasma",
			"Plasma演示器",
			"演示Plasma子链的区块提交、UTXO模型、退出机制、挑战等核心功能",
			"crosschain",
			types.ComponentProcess,
		),
		blocks:   make(map[uint64]*PlasmaBlock),
		utxos:    make(map[string]*PlasmaUTXO),
		exits:    make(map[string]*ExitRequest),
		exitBond: new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17)),
	}

	sim.AddParam(types.Param{
		Key:         "challenge_period_days",
		Name:        "挑战期(天)",
		Description: "退出挑战期时长",
		Type:        types.ParamTypeInt,
		Default:     7,
		Min:         1,
		Max:         14,
	})

	return sim
}

// Init 初始化
func (s *PlasmaSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.challengePeriod = 7 * 24 * time.Hour
	if v, ok := config.Params["challenge_period_days"]; ok {
		if n, ok := v.(float64); ok {
			s.challengePeriod = time.Duration(n) * 24 * time.Hour
		}
	}

	s.blocks = make(map[uint64]*PlasmaBlock)
	s.utxos = make(map[string]*PlasmaUTXO)
	s.exits = make(map[string]*ExitRequest)
	s.currentBlock = 0

	s.updateState()
	return nil
}

// ExplainPlasma 解释机制
func (s *PlasmaSimulator) ExplainPlasma() map[string]interface{} {
	return map[string]interface{}{
		"overview": "Plasma是子链扩容方案，通过定期向主链提交状态根保证安全",
		"variants": []map[string]string{
			{"name": "Plasma MVP", "model": "UTXO", "feature": "最简实现"},
			{"name": "Plasma Cash", "model": "NFT", "feature": "每个代币唯一"},
			{"name": "Plasma Debit", "model": "账户", "feature": "支持账户模型"},
		},
		"workflow": []map[string]string{
			{"step": "1. 存款", "desc": "用户在主链存款，子链创建UTXO"},
			{"step": "2. 交易", "desc": "子链内UTXO转移"},
			{"step": "3. 提交", "desc": "Operator定期提交Merkle根到主链"},
			{"step": "4. 退出", "desc": "用户提交UTXO证明发起退出"},
			{"step": "5. 挑战", "desc": "挑战期内可挑战无效退出"},
		},
		"exit_game": map[string]interface{}{
			"purpose":  "确保用户可安全退出",
			"priority": "按UTXO年龄排序(越老越优先)",
			"challenge": []string{
				"UTXO已花费",
				"UTXO所有权不正确",
				"更新的UTXO存在",
			},
		},
		"limitations": []string{
			"数据可用性问题",
			"批量退出拥堵",
			"不支持复杂智能合约",
		},
	}
}

// Deposit 存款
func (s *PlasmaSimulator) Deposit(user string, amount *big.Int, token string) (*PlasmaUTXO, error) {
	utxoData := fmt.Sprintf("%s-%s-%d", user, amount.String(), time.Now().UnixNano())
	utxoHash := sha256.Sum256([]byte(utxoData))
	utxoID := fmt.Sprintf("utxo-%s", hex.EncodeToString(utxoHash[:8]))

	utxo := &PlasmaUTXO{
		UTXOID:      utxoID,
		Owner:       user,
		Amount:      new(big.Int).Set(amount),
		Token:       token,
		BlockNumber: 0,
		TxIndex:     0,
		OutputIndex: 0,
		Spent:       false,
	}

	s.utxos[utxoID] = utxo

	s.EmitEvent("plasma_deposit", "", "", map[string]interface{}{
		"utxo_id": utxoID,
		"user":    user,
		"amount":  amount.String(),
		"token":   token,
	})

	s.updateState()
	return utxo, nil
}

// SubmitBlock 提交区块
func (s *PlasmaSimulator) SubmitBlock(operator string, txCount int) (*PlasmaBlock, error) {
	s.currentBlock++

	blockData := fmt.Sprintf("%s-%d-%d", operator, s.currentBlock, time.Now().UnixNano())
	blockHash := sha256.Sum256([]byte(blockData))

	block := &PlasmaBlock{
		BlockNumber: s.currentBlock,
		MerkleRoot:  "0x" + hex.EncodeToString(blockHash[:]),
		TxCount:     txCount,
		Operator:    operator,
		Status:      PlasmaBlockSubmitted,
		SubmittedAt: time.Now(),
		L1TxHash:    "0x" + hex.EncodeToString(blockHash[16:]),
	}

	s.blocks[s.currentBlock] = block

	s.EmitEvent("plasma_block_submitted", "", "", map[string]interface{}{
		"block_number": s.currentBlock,
		"merkle_root":  block.MerkleRoot,
		"tx_count":     txCount,
	})

	s.updateState()
	return block, nil
}

// StartExit 发起退出
func (s *PlasmaSimulator) StartExit(utxoID string) (*ExitRequest, error) {
	utxo, ok := s.utxos[utxoID]
	if !ok {
		return nil, fmt.Errorf("UTXO不存在: %s", utxoID)
	}

	if utxo.Spent {
		return nil, fmt.Errorf("UTXO已花费")
	}

	exitData := fmt.Sprintf("%s-%d", utxoID, time.Now().UnixNano())
	exitHash := sha256.Sum256([]byte(exitData))
	exitID := fmt.Sprintf("exit-%s", hex.EncodeToString(exitHash[:8]))

	exit := &ExitRequest{
		ExitID:       exitID,
		UTXOID:       utxoID,
		Owner:        utxo.Owner,
		Amount:       new(big.Int).Set(utxo.Amount),
		Status:       ExitPending,
		ExitBond:     new(big.Int).Set(s.exitBond),
		RequestedAt:  time.Now(),
		ChallengeEnd: time.Now().Add(s.challengePeriod),
	}

	s.exits[exitID] = exit

	s.EmitEvent("plasma_exit_started", "", "", map[string]interface{}{
		"exit_id":       exitID,
		"utxo_id":       utxoID,
		"amount":        utxo.Amount.String(),
		"challenge_end": exit.ChallengeEnd,
	})

	s.updateState()
	return exit, nil
}

// ChallengeExit 挑战退出
func (s *PlasmaSimulator) ChallengeExit(exitID, challenger string) error {
	exit, ok := s.exits[exitID]
	if !ok {
		return fmt.Errorf("退出请求不存在: %s", exitID)
	}

	if exit.Status != ExitPending {
		return fmt.Errorf("退出状态不允许挑战: %s", exit.Status)
	}

	if time.Now().After(exit.ChallengeEnd) {
		return fmt.Errorf("挑战期已结束")
	}

	exit.Status = ExitCanceled

	s.EmitEvent("plasma_exit_challenged", "", "", map[string]interface{}{
		"exit_id":    exitID,
		"challenger": challenger,
		"bond_to":    challenger,
	})

	s.updateState()
	return nil
}

// FinalizeExit 完成退出
func (s *PlasmaSimulator) FinalizeExit(exitID string) error {
	exit, ok := s.exits[exitID]
	if !ok {
		return fmt.Errorf("退出请求不存在: %s", exitID)
	}

	if exit.Status != ExitPending {
		return fmt.Errorf("退出状态不允许完成: %s", exit.Status)
	}

	if time.Now().Before(exit.ChallengeEnd) {
		return fmt.Errorf("挑战期未结束: 剩余%s", time.Until(exit.ChallengeEnd))
	}

	exit.Status = ExitFinalized
	exit.FinalizedAt = time.Now()

	if utxo, ok := s.utxos[exit.UTXOID]; ok {
		utxo.Spent = true
	}

	s.EmitEvent("plasma_exit_finalized", "", "", map[string]interface{}{
		"exit_id": exitID,
		"amount":  exit.Amount.String(),
		"owner":   exit.Owner,
	})

	s.updateState()
	return nil
}

// GetStatistics 获取统计
func (s *PlasmaSimulator) GetStatistics() map[string]interface{} {
	pendingExits := 0
	for _, e := range s.exits {
		if e.Status == ExitPending {
			pendingExits++
		}
	}

	totalValue := big.NewInt(0)
	for _, u := range s.utxos {
		if !u.Spent {
			totalValue.Add(totalValue, u.Amount)
		}
	}

	return map[string]interface{}{
		"total_blocks":     len(s.blocks),
		"current_block":    s.currentBlock,
		"total_utxos":      len(s.utxos),
		"total_exits":      len(s.exits),
		"pending_exits":    pendingExits,
		"total_value":      totalValue.String(),
		"challenge_period": s.challengePeriod.String(),
	}
}

func (s *PlasmaSimulator) updateState() {
	s.SetGlobalData("block_count", len(s.blocks))
	s.SetGlobalData("utxo_count", len(s.utxos))

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"plasma",
		"当前可以先向 Plasma 链存款，再提交块或发起退出流程。",
		"先执行 deposit，再观察 UTXO、Plasma 块和退出请求如何逐步形成。",
		0,
		map[string]interface{}{
			"block_count": len(s.blocks),
			"utxo_count":  len(s.utxos),
		},
	)
}

func (s *PlasmaSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "deposit":
		user := "alice"
		token := "ETH"
		amount := big.NewInt(10)
		if raw, ok := params["user"].(string); ok && raw != "" {
			user = raw
		}
		if raw, ok := params["token"].(string); ok && raw != "" {
			token = raw
		}
		if raw, ok := params["amount"].(float64); ok && raw > 0 {
			amount = big.NewInt(int64(raw))
		}
		utxo, err := s.Deposit(user, amount, token)
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已向 Plasma 链存款",
			map[string]interface{}{"utxo_id": utxo.UTXOID, "owner": utxo.Owner},
			&types.ActionFeedback{
				Summary:     "新的 UTXO 已经生成，可继续提交 Plasma 区块或发起退出。",
				NextHint:    "执行 submit_block，观察 UTXO 如何被纳入链上批次。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported plasma action: %s", action)
	}
}

// Factory
type PlasmaFactory struct{}

func (f *PlasmaFactory) Create() engine.Simulator { return NewPlasmaSimulator() }
func (f *PlasmaFactory) GetDescription() types.Description {
	return NewPlasmaSimulator().GetDescription()
}
func NewPlasmaFactory() *PlasmaFactory { return &PlasmaFactory{} }

var _ engine.SimulatorFactory = (*PlasmaFactory)(nil)
