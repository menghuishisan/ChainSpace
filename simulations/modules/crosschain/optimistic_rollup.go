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
// Optimistic Rollup演示器
// 演示Optimistic Rollup的核心机制
//
// 核心概念:
// 1. 乐观执行: 假设所有交易有效，直接执行
// 2. 欺诈证明: 挑战期内可提交欺诈证明
// 3. 挑战期: 7天等待期，允许验证者挑战
// 4. 状态承诺: 定期提交状态根到L1
//
// 参考: Optimism, Arbitrum
// =============================================================================

// RollupBatchStatus 批次状态
type RollupBatchStatus string

const (
	BatchPending    RollupBatchStatus = "pending"
	BatchSubmitted  RollupBatchStatus = "submitted"
	BatchChallenged RollupBatchStatus = "challenged"
	BatchFinalized  RollupBatchStatus = "finalized"
	BatchReverted   RollupBatchStatus = "reverted"
)

// RollupBatch 批次
type RollupBatch struct {
	BatchID       string            `json:"batch_id"`
	BatchIndex    uint64            `json:"batch_index"`
	Transactions  []string          `json:"transactions"`
	TxCount       int               `json:"tx_count"`
	StateRoot     string            `json:"state_root"`
	PrevStateRoot string            `json:"prev_state_root"`
	Sequencer     string            `json:"sequencer"`
	Status        RollupBatchStatus `json:"status"`
	SubmittedAt   time.Time         `json:"submitted_at"`
	ChallengeEnd  time.Time         `json:"challenge_end"`
	FinalizedAt   time.Time         `json:"finalized_at"`
	L1TxHash      string            `json:"l1_tx_hash"`
}

// FraudProof 欺诈证明
type FraudProof struct {
	ProofID        string    `json:"proof_id"`
	BatchID        string    `json:"batch_id"`
	Challenger     string    `json:"challenger"`
	InvalidTxIndex int       `json:"invalid_tx_index"`
	PreStateRoot   string    `json:"pre_state_root"`
	PostStateRoot  string    `json:"post_state_root"`
	ProofData      string    `json:"proof_data"`
	Verified       bool      `json:"verified"`
	SubmittedAt    time.Time `json:"submitted_at"`
}

// OptimisticRollupSimulator Optimistic Rollup演示器
type OptimisticRollupSimulator struct {
	*base.BaseSimulator
	batches          map[string]*RollupBatch
	fraudProofs      map[string]*FraudProof
	currentBatchIdx  uint64
	challengePeriod  time.Duration
	sequencerBond    *big.Int
	currentStateRoot string
}

// NewOptimisticRollupSimulator 创建演示器
func NewOptimisticRollupSimulator() *OptimisticRollupSimulator {
	sim := &OptimisticRollupSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"optimistic_rollup",
			"Optimistic Rollup演示器",
			"演示Optimistic Rollup的乐观执行、欺诈证明、挑战期等核心机制",
			"crosschain",
			types.ComponentProcess,
		),
		batches:       make(map[string]*RollupBatch),
		fraudProofs:   make(map[string]*FraudProof),
		sequencerBond: new(big.Int).Mul(big.NewInt(32), big.NewInt(1e18)),
	}

	sim.AddParam(types.Param{
		Key:         "challenge_period_days",
		Name:        "挑战期(天)",
		Description: "欺诈证明挑战期时长",
		Type:        types.ParamTypeInt,
		Default:     7,
		Min:         1,
		Max:         14,
	})

	return sim
}

// Init 初始化
func (s *OptimisticRollupSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.challengePeriod = 7 * 24 * time.Hour
	if v, ok := config.Params["challenge_period_days"]; ok {
		if n, ok := v.(float64); ok {
			s.challengePeriod = time.Duration(n) * 24 * time.Hour
		}
	}

	s.batches = make(map[string]*RollupBatch)
	s.fraudProofs = make(map[string]*FraudProof)
	s.currentBatchIdx = 0
	s.currentStateRoot = "0x" + hex.EncodeToString(make([]byte, 32))

	s.updateState()
	return nil
}

// ExplainOptimisticRollup 解释机制
func (s *OptimisticRollupSimulator) ExplainOptimisticRollup() map[string]interface{} {
	return map[string]interface{}{
		"overview": "Optimistic Rollup假设所有交易有效，通过欺诈证明保证安全性",
		"workflow": []map[string]string{
			{"step": "1. 收集交易", "desc": "Sequencer收集L2交易"},
			{"step": "2. 执行交易", "desc": "在L2执行并生成新状态根"},
			{"step": "3. 提交批次", "desc": "将状态根和交易数据提交到L1"},
			{"step": "4. 挑战期", "desc": fmt.Sprintf("等待%s，允许验证者挑战", s.challengePeriod)},
			{"step": "5. 最终确认", "desc": "挑战期结束无异议则状态最终确认"},
		},
		"fraud_proof": map[string]interface{}{
			"purpose":    "证明某笔交易执行结果错误",
			"components": []string{"交易数据", "前状态证明", "后状态证明", "执行证明"},
			"result":     "成功挑战：批次回滚，Sequencer质押被罚没",
		},
		"pros": []string{
			"EVM完全兼容",
			"数据在链上，安全性高",
			"Gas成本相对低",
		},
		"cons": []string{
			"提款需等待挑战期(7天)",
			"依赖至少一个诚实验证者",
		},
		"implementations": []map[string]string{
			{"name": "Optimism", "feature": "OVM, EVM等效"},
			{"name": "Arbitrum", "feature": "AVM, 交互式证明"},
			{"name": "Base", "feature": "基于OP Stack"},
		},
	}
}

// SubmitBatch 提交批次
func (s *OptimisticRollupSimulator) SubmitBatch(sequencer string, transactions []string) (*RollupBatch, error) {
	s.currentBatchIdx++

	batchData := fmt.Sprintf("%s-%d-%d", sequencer, s.currentBatchIdx, time.Now().UnixNano())
	batchHash := sha256.Sum256([]byte(batchData))
	batchID := fmt.Sprintf("batch-%s", hex.EncodeToString(batchHash[:8]))

	stateData := fmt.Sprintf("%s-%s", s.currentStateRoot, batchID)
	stateHash := sha256.Sum256([]byte(stateData))
	newStateRoot := "0x" + hex.EncodeToString(stateHash[:])

	batch := &RollupBatch{
		BatchID:       batchID,
		BatchIndex:    s.currentBatchIdx,
		Transactions:  transactions,
		TxCount:       len(transactions),
		StateRoot:     newStateRoot,
		PrevStateRoot: s.currentStateRoot,
		Sequencer:     sequencer,
		Status:        BatchSubmitted,
		SubmittedAt:   time.Now(),
		ChallengeEnd:  time.Now().Add(s.challengePeriod),
		L1TxHash:      "0x" + hex.EncodeToString(batchHash[:]),
	}

	s.batches[batchID] = batch
	s.currentStateRoot = newStateRoot

	s.EmitEvent("batch_submitted", "", "", map[string]interface{}{
		"batch_id":      batchID,
		"batch_index":   s.currentBatchIdx,
		"tx_count":      len(transactions),
		"state_root":    newStateRoot,
		"challenge_end": batch.ChallengeEnd,
	})

	s.updateState()
	return batch, nil
}

// SubmitFraudProof 提交欺诈证明
func (s *OptimisticRollupSimulator) SubmitFraudProof(batchID, challenger string, invalidTxIdx int) (*FraudProof, error) {
	batch, ok := s.batches[batchID]
	if !ok {
		return nil, fmt.Errorf("批次不存在: %s", batchID)
	}

	if batch.Status != BatchSubmitted {
		return nil, fmt.Errorf("批次状态不允许挑战: %s", batch.Status)
	}

	if time.Now().After(batch.ChallengeEnd) {
		return nil, fmt.Errorf("挑战期已结束")
	}

	proofData := fmt.Sprintf("%s-%s-%d-%d", batchID, challenger, invalidTxIdx, time.Now().UnixNano())
	proofHash := sha256.Sum256([]byte(proofData))
	proofID := fmt.Sprintf("proof-%s", hex.EncodeToString(proofHash[:8]))

	proof := &FraudProof{
		ProofID:        proofID,
		BatchID:        batchID,
		Challenger:     challenger,
		InvalidTxIndex: invalidTxIdx,
		PreStateRoot:   batch.PrevStateRoot,
		PostStateRoot:  batch.StateRoot,
		ProofData:      hex.EncodeToString(proofHash[:]),
		Verified:       false,
		SubmittedAt:    time.Now(),
	}

	batch.Status = BatchChallenged
	s.fraudProofs[proofID] = proof

	s.EmitEvent("fraud_proof_submitted", "", "", map[string]interface{}{
		"proof_id":   proofID,
		"batch_id":   batchID,
		"challenger": challenger,
		"invalid_tx": invalidTxIdx,
	})

	s.updateState()
	return proof, nil
}

// VerifyFraudProof 验证欺诈证明
func (s *OptimisticRollupSimulator) VerifyFraudProof(proofID string, isValid bool) error {
	proof, ok := s.fraudProofs[proofID]
	if !ok {
		return fmt.Errorf("欺诈证明不存在: %s", proofID)
	}

	batch := s.batches[proof.BatchID]
	proof.Verified = true

	if isValid {
		batch.Status = BatchReverted
		s.currentStateRoot = batch.PrevStateRoot

		s.EmitEvent("batch_reverted", "", "", map[string]interface{}{
			"batch_id":       proof.BatchID,
			"challenger":     proof.Challenger,
			"reward":         s.sequencerBond.String(),
			"state_reverted": batch.PrevStateRoot,
		})
	} else {
		batch.Status = BatchSubmitted
		batch.ChallengeEnd = time.Now().Add(s.challengePeriod)

		s.EmitEvent("fraud_proof_rejected", "", "", map[string]interface{}{
			"proof_id": proofID,
			"batch_id": proof.BatchID,
		})
	}

	s.updateState()
	return nil
}

// FinalizeBatch 最终确认批次
func (s *OptimisticRollupSimulator) FinalizeBatch(batchID string) error {
	batch, ok := s.batches[batchID]
	if !ok {
		return fmt.Errorf("批次不存在: %s", batchID)
	}

	if batch.Status != BatchSubmitted {
		return fmt.Errorf("批次状态不允许最终确认: %s", batch.Status)
	}

	if time.Now().Before(batch.ChallengeEnd) {
		return fmt.Errorf("挑战期未结束: 剩余%s", time.Until(batch.ChallengeEnd))
	}

	batch.Status = BatchFinalized
	batch.FinalizedAt = time.Now()

	s.EmitEvent("batch_finalized", "", "", map[string]interface{}{
		"batch_id":     batchID,
		"state_root":   batch.StateRoot,
		"finalized_at": batch.FinalizedAt,
	})

	s.updateState()
	return nil
}

// SimulateRollupFlow 模拟完整流程
func (s *OptimisticRollupSimulator) SimulateRollupFlow() map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	txs := []string{"tx1: transfer 100 ETH", "tx2: swap tokens", "tx3: mint NFT"}
	batch, _ := s.SubmitBatch("sequencer-1", txs)
	steps = append(steps, map[string]interface{}{
		"step": 1, "action": "Sequencer提交批次",
		"batch_id": batch.BatchID, "tx_count": len(txs),
		"state_root": batch.StateRoot,
	})

	steps = append(steps, map[string]interface{}{
		"step": 2, "action": "进入挑战期",
		"duration": s.challengePeriod.String(),
		"deadline": batch.ChallengeEnd,
	})

	steps = append(steps, map[string]interface{}{
		"step": 3, "action": "验证者可提交欺诈证明",
		"scenario_a": "无挑战 -> 批次最终确认",
		"scenario_b": "有效挑战 -> 批次回滚，Sequencer被罚",
	})

	return map[string]interface{}{
		"batch":   batch,
		"steps":   steps,
		"summary": "Optimistic Rollup通过经济激励确保安全：诚实行为获利，欺诈行为被罚",
	}
}

// GetStatistics 获取统计
func (s *OptimisticRollupSimulator) GetStatistics() map[string]interface{} {
	finalized, challenged, reverted := 0, 0, 0
	for _, b := range s.batches {
		switch b.Status {
		case BatchFinalized:
			finalized++
		case BatchChallenged:
			challenged++
		case BatchReverted:
			reverted++
		}
	}

	return map[string]interface{}{
		"total_batches":      len(s.batches),
		"finalized_batches":  finalized,
		"challenged_batches": challenged,
		"reverted_batches":   reverted,
		"fraud_proofs":       len(s.fraudProofs),
		"challenge_period":   s.challengePeriod.String(),
		"current_state_root": s.currentStateRoot,
	}
}

func (s *OptimisticRollupSimulator) updateState() {
	s.SetGlobalData("batch_count", len(s.batches))
	s.SetGlobalData("current_state_root", s.currentStateRoot)

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"optimistic_rollup",
		"当前可以提交批次并观察挑战期如何影响最终确认。",
		"先提交一个批次，再决定是否发起欺诈证明或直接完成最终确认。",
		0,
		map[string]interface{}{
			"batch_count":         len(s.batches),
			"current_state_root": s.currentStateRoot,
		},
	)
}

func (s *OptimisticRollupSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "submit_batch":
		sequencer := "sequencer-1"
		txCount := 16
		if raw, ok := params["sequencer"].(string); ok && raw != "" {
			sequencer = raw
		}
		if raw, ok := params["tx_count"].(float64); ok && raw > 0 {
			txCount = int(raw)
		}
		txs := make([]string, txCount)
		for i := 0; i < txCount; i++ {
			txs[i] = fmt.Sprintf("tx-%d", i+1)
		}
		batch, err := s.SubmitBatch(sequencer, txs)
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已提交 Optimistic Rollup 批次",
			map[string]interface{}{"batch_id": batch.BatchID, "tx_count": batch.TxCount},
			&types.ActionFeedback{
				Summary:     "批次已经进入挑战期，可继续发起欺诈证明或等待最终确认。",
				NextHint:    "尝试 finalize_batch，或者先构造欺诈证明观察挑战流程。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported optimistic rollup action: %s", action)
	}
}

// Factory
type OptimisticRollupFactory struct{}

func (f *OptimisticRollupFactory) Create() engine.Simulator { return NewOptimisticRollupSimulator() }
func (f *OptimisticRollupFactory) GetDescription() types.Description {
	return NewOptimisticRollupSimulator().GetDescription()
}
func NewOptimisticRollupFactory() *OptimisticRollupFactory { return &OptimisticRollupFactory{} }

var _ engine.SimulatorFactory = (*OptimisticRollupFactory)(nil)
