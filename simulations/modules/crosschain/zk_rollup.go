package crosschain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// ZK Rollup演示器
// 演示ZK Rollup的核心机制
//
// 核心概念:
// 1. 有效性证明: 使用零知识证明验证状态转换正确性
// 2. 即时确认: 证明验证通过即可确认
// 3. 数据压缩: 只需存储状态差异
// 4. 证明生成: 电路编译、证明计算
//
// 参考: zkSync, StarkNet, Polygon zkEVM
// =============================================================================

// ZKProofType ZK证明类型
type ZKProofType string

const (
	ProofTypeSNARK ZKProofType = "snark"
	ProofTypeSTARK ZKProofType = "stark"
	ProofTypePLONK ZKProofType = "plonk"
)

// ZKBatch ZK批次
type ZKBatch struct {
	BatchID       string      `json:"batch_id"`
	BatchIndex    uint64      `json:"batch_index"`
	TxCount       int         `json:"tx_count"`
	StateRoot     string      `json:"state_root"`
	PrevStateRoot string      `json:"prev_state_root"`
	ProofType     ZKProofType `json:"proof_type"`
	Proof         *ZKProof    `json:"proof"`
	Verified      bool        `json:"verified"`
	Prover        string      `json:"prover"`
	SubmittedAt   time.Time   `json:"submitted_at"`
	VerifiedAt    time.Time   `json:"verified_at"`
	L1TxHash      string      `json:"l1_tx_hash"`
}

// ZKProof ZK证明
type ZKProof struct {
	ProofID          string        `json:"proof_id"`
	ProofType        ZKProofType   `json:"proof_type"`
	ProofData        string        `json:"proof_data"`
	PublicInputs     []string      `json:"public_inputs"`
	ProofSize        int           `json:"proof_size_bytes"`
	GenerationTime   time.Duration `json:"generation_time"`
	VerificationTime time.Duration `json:"verification_time"`
}

// ZKCircuit ZK电路
type ZKCircuit struct {
	Name          string `json:"name"`
	Constraints   int    `json:"constraints"`
	PublicInputs  int    `json:"public_inputs"`
	PrivateInputs int    `json:"private_inputs"`
	ProofSize     int    `json:"proof_size_bytes"`
	ProverTime    string `json:"prover_time"`
	VerifierTime  string `json:"verifier_time"`
}

// ZKRollupSimulator ZK Rollup演示器
type ZKRollupSimulator struct {
	*base.BaseSimulator
	batches          map[string]*ZKBatch
	circuits         map[string]*ZKCircuit
	currentBatchIdx  uint64
	proofType        ZKProofType
	currentStateRoot string
}

// NewZKRollupSimulator 创建演示器
func NewZKRollupSimulator() *ZKRollupSimulator {
	sim := &ZKRollupSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"zk_rollup",
			"ZK Rollup演示器",
			"演示ZK Rollup的有效性证明、即时确认、证明生成验证等核心机制",
			"crosschain",
			types.ComponentProcess,
		),
		batches:  make(map[string]*ZKBatch),
		circuits: make(map[string]*ZKCircuit),
	}

	sim.AddParam(types.Param{
		Key:         "proof_type",
		Name:        "证明类型",
		Description: "ZK证明系统类型",
		Type:        types.ParamTypeSelect,
		Default:     "plonk",
		Options: []types.Option{
			{Label: "SNARK (Groth16)", Value: "snark"},
			{Label: "STARK", Value: "stark"},
			{Label: "PLONK", Value: "plonk"},
		},
	})

	return sim
}

// Init 初始化
func (s *ZKRollupSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.proofType = ProofTypePLONK
	if v, ok := config.Params["proof_type"]; ok {
		if pt, ok := v.(string); ok {
			s.proofType = ZKProofType(pt)
		}
	}

	s.batches = make(map[string]*ZKBatch)
	s.circuits = make(map[string]*ZKCircuit)
	s.currentBatchIdx = 0
	s.currentStateRoot = "0x" + hex.EncodeToString(make([]byte, 32))

	s.initializeCircuits()
	s.updateState()
	return nil
}

func (s *ZKRollupSimulator) initializeCircuits() {
	s.circuits["transfer"] = &ZKCircuit{
		Name: "Transfer", Constraints: 5000, PublicInputs: 4, PrivateInputs: 2,
		ProofSize: 256, ProverTime: "2s", VerifierTime: "10ms",
	}
	s.circuits["swap"] = &ZKCircuit{
		Name: "Swap", Constraints: 15000, PublicInputs: 6, PrivateInputs: 4,
		ProofSize: 256, ProverTime: "5s", VerifierTime: "10ms",
	}
	s.circuits["batch"] = &ZKCircuit{
		Name: "BatchVerifier", Constraints: 1000000, PublicInputs: 3, PrivateInputs: 1000,
		ProofSize: 256, ProverTime: "30min", VerifierTime: "10ms",
	}
}

// ExplainZKRollup 解释机制
func (s *ZKRollupSimulator) ExplainZKRollup() map[string]interface{} {
	return map[string]interface{}{
		"overview": "ZK Rollup使用零知识证明验证状态转换，无需挑战期即可确认",
		"workflow": []map[string]string{
			{"step": "1. 收集交易", "desc": "Sequencer收集L2交易"},
			{"step": "2. 执行交易", "desc": "执行并生成状态转换"},
			{"step": "3. 生成证明", "desc": "Prover生成ZK证明(耗时)"},
			{"step": "4. 提交验证", "desc": "提交证明到L1并验证"},
			{"step": "5. 即时确认", "desc": "验证通过即最终确认"},
		},
		"proof_systems": map[string]interface{}{
			"snark": map[string]string{
				"name": "zk-SNARK (Groth16)",
				"pros": "证明小(~256B), 验证快(~10ms)",
				"cons": "需要可信设置, 证明生成慢",
			},
			"stark": map[string]string{
				"name": "zk-STARK",
				"pros": "无需可信设置, 量子安全",
				"cons": "证明较大(~100KB)",
			},
			"plonk": map[string]string{
				"name": "PLONK",
				"pros": "通用可信设置, 证明适中",
				"cons": "验证稍慢",
			},
		},
		"pros": []string{
			"即时最终确认",
			"安全性由数学保证",
			"数据压缩率高",
		},
		"cons": []string{
			"证明生成计算密集",
			"EVM兼容性挑战",
			"电路开发复杂",
		},
		"implementations": []map[string]string{
			{"name": "zkSync Era", "type": "SNARK", "feature": "zkEVM"},
			{"name": "StarkNet", "type": "STARK", "feature": "Cairo语言"},
			{"name": "Polygon zkEVM", "type": "SNARK", "feature": "EVM等效"},
			{"name": "Scroll", "type": "SNARK", "feature": "zkEVM"},
		},
	}
}

// GenerateProof 生成ZK证明
func (s *ZKRollupSimulator) GenerateProof(txCount int) *ZKProof {
	startTime := time.Now()

	proofData := fmt.Sprintf("%s-%d-%d", s.proofType, txCount, time.Now().UnixNano())
	proofHash := sha256.Sum256([]byte(proofData))

	genTime := time.Duration(txCount*100) * time.Millisecond

	proof := &ZKProof{
		ProofID:   fmt.Sprintf("proof-%s", hex.EncodeToString(proofHash[:8])),
		ProofType: s.proofType,
		ProofData: hex.EncodeToString(proofHash[:]),
		PublicInputs: []string{
			s.currentStateRoot,
			fmt.Sprintf("tx_count:%d", txCount),
		},
		ProofSize:        256,
		GenerationTime:   time.Since(startTime) + genTime,
		VerificationTime: 10 * time.Millisecond,
	}

	return proof
}

// SubmitBatch 提交批次
func (s *ZKRollupSimulator) SubmitBatch(prover string, txCount int) (*ZKBatch, error) {
	s.currentBatchIdx++

	proof := s.GenerateProof(txCount)

	batchData := fmt.Sprintf("%s-%d-%d", prover, s.currentBatchIdx, time.Now().UnixNano())
	batchHash := sha256.Sum256([]byte(batchData))
	batchID := fmt.Sprintf("zkbatch-%s", hex.EncodeToString(batchHash[:8]))

	stateData := fmt.Sprintf("%s-%s", s.currentStateRoot, batchID)
	stateHash := sha256.Sum256([]byte(stateData))
	newStateRoot := "0x" + hex.EncodeToString(stateHash[:])

	batch := &ZKBatch{
		BatchID:       batchID,
		BatchIndex:    s.currentBatchIdx,
		TxCount:       txCount,
		StateRoot:     newStateRoot,
		PrevStateRoot: s.currentStateRoot,
		ProofType:     s.proofType,
		Proof:         proof,
		Verified:      false,
		Prover:        prover,
		SubmittedAt:   time.Now(),
		L1TxHash:      "0x" + hex.EncodeToString(batchHash[:]),
	}

	s.batches[batchID] = batch

	s.EmitEvent("zk_batch_submitted", "", "", map[string]interface{}{
		"batch_id":   batchID,
		"tx_count":   txCount,
		"proof_type": string(s.proofType),
		"proof_size": proof.ProofSize,
	})

	s.updateState()
	return batch, nil
}

// VerifyBatch 验证批次
func (s *ZKRollupSimulator) VerifyBatch(batchID string) error {
	batch, ok := s.batches[batchID]
	if !ok {
		return fmt.Errorf("批次不存在: %s", batchID)
	}

	if batch.Verified {
		return fmt.Errorf("批次已验证")
	}

	batch.Verified = true
	batch.VerifiedAt = time.Now()
	s.currentStateRoot = batch.StateRoot

	s.EmitEvent("zk_batch_verified", "", "", map[string]interface{}{
		"batch_id":    batchID,
		"state_root":  batch.StateRoot,
		"verify_time": batch.Proof.VerificationTime.String(),
		"finalized":   true,
	})

	s.updateState()
	return nil
}

// CompareWithOptimistic 对比Optimistic Rollup
func (s *ZKRollupSimulator) CompareWithOptimistic() map[string]interface{} {
	return map[string]interface{}{
		"comparison": []map[string]interface{}{
			{
				"aspect":     "最终确认时间",
				"zk_rollup":  "即时(证明验证后)",
				"optimistic": "7天(挑战期)",
				"winner":     "ZK Rollup",
			},
			{
				"aspect":     "计算成本",
				"zk_rollup":  "高(证明生成)",
				"optimistic": "低(仅执行)",
				"winner":     "Optimistic",
			},
			{
				"aspect":     "EVM兼容性",
				"zk_rollup":  "较难(需zkEVM)",
				"optimistic": "完全兼容",
				"winner":     "Optimistic",
			},
			{
				"aspect":     "安全假设",
				"zk_rollup":  "数学保证",
				"optimistic": "至少1个诚实验证者",
				"winner":     "ZK Rollup",
			},
		},
		"recommendation": "ZK Rollup适合高价值场景; Optimistic适合通用DApp",
	}
}

// GetStatistics 获取统计
func (s *ZKRollupSimulator) GetStatistics() map[string]interface{} {
	verified := 0
	totalTxs := 0
	for _, b := range s.batches {
		if b.Verified {
			verified++
		}
		totalTxs += b.TxCount
	}

	return map[string]interface{}{
		"total_batches":    len(s.batches),
		"verified_batches": verified,
		"total_txs":        totalTxs,
		"proof_type":       string(s.proofType),
		"state_root":       s.currentStateRoot,
	}
}

func (s *ZKRollupSimulator) updateState() {
	s.SetGlobalData("batch_count", len(s.batches))
	s.SetGlobalData("proof_type", string(s.proofType))

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"zk_rollup",
		"当前可以提交批次并观察证明生成与验证如何共同推进状态根更新。",
		"先提交一个批次，再验证它的证明，观察状态根如何完成更新。",
		0,
		map[string]interface{}{
			"batch_count": len(s.batches),
			"proof_type":  string(s.proofType),
		},
	)
}

func (s *ZKRollupSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "submit_batch":
		prover := "prover-1"
		txCount := 32
		if raw, ok := params["prover"].(string); ok && raw != "" {
			prover = raw
		}
		if raw, ok := params["tx_count"].(float64); ok && raw > 0 {
			txCount = int(raw)
		}
		batch, err := s.SubmitBatch(prover, txCount)
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已提交 ZK Rollup 批次",
			map[string]interface{}{"batch_id": batch.BatchID, "tx_count": batch.TxCount},
			&types.ActionFeedback{
				Summary:     "新的批次已经提交并生成证明，可继续验证该批次。",
				NextHint:    "执行 verify_batch，观察状态根何时真正写入最终结果。",
				EffectScope: "crosschain",
			},
		), nil
	case "verify_batch":
		batchID := ""
		if raw, ok := params["batch_id"].(string); ok && raw != "" {
			batchID = raw
		}
		if batchID == "" {
			for id := range s.batches {
				batchID = id
				break
			}
		}
		if batchID == "" {
			return nil, fmt.Errorf("没有可验证的 ZK 批次")
		}
		if err := s.VerifyBatch(batchID); err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已验证 ZK 批次",
			map[string]interface{}{"batch_id": batchID},
			&types.ActionFeedback{
				Summary:     "批次证明已经验证完成，新的状态根已进入最终状态。",
				NextHint:    "继续提交下一批次，对比不同证明类型的提交与验证节奏。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported zk rollup action: %s", action)
	}
}

// Factory
type ZKRollupFactory struct{}

func (f *ZKRollupFactory) Create() engine.Simulator { return NewZKRollupSimulator() }
func (f *ZKRollupFactory) GetDescription() types.Description {
	return NewZKRollupSimulator().GetDescription()
}
func NewZKRollupFactory() *ZKRollupFactory { return &ZKRollupFactory{} }

var _ engine.SimulatorFactory = (*ZKRollupFactory)(nil)
