package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// BlockHeader 区块头
type BlockHeader struct {
	Hash       string `json:"hash"`
	ParentHash string `json:"parent_hash"`
	Height     uint64 `json:"height"`
	StateRoot  string `json:"state_root"`
	TxRoot     string `json:"tx_root"`
}

// SPVProof SPV证明
type SPVProof struct {
	TxHash      string   `json:"tx_hash"`
	BlockHash   string   `json:"block_hash"`
	BlockHeight uint64   `json:"block_height"`
	MerklePath  []string `json:"merkle_path"`
	Directions  []string `json:"directions"`
}

// LightClientSimulator 轻客户端演示器
type LightClientSimulator struct {
	*base.BaseSimulator
	headers     []*BlockHeader
	trustedRoot string
	verifiedTxs []string
}

// NewLightClientSimulator 创建轻客户端演示器
func NewLightClientSimulator() *LightClientSimulator {
	sim := &LightClientSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"light_client",
			"轻客户端/SPV演示器",
			"展示SPV轻客户端的区块头验证和交易证明",
			"blockchain",
			types.ComponentTool,
		),
		headers:     make([]*BlockHeader, 0),
		verifiedTxs: make([]string, 0),
	}
	return sim
}

// Init 初始化
func (s *LightClientSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	genesis := &BlockHeader{
		Hash: "genesis", ParentHash: "", Height: 0,
		StateRoot: s.hash("state-0"), TxRoot: s.hash("tx-0"),
	}
	s.headers = []*BlockHeader{genesis}
	s.trustedRoot = genesis.Hash

	for i := 1; i <= 10; i++ {
		parent := s.headers[i-1]
		header := &BlockHeader{
			ParentHash: parent.Hash,
			Height:     uint64(i),
			StateRoot:  s.hash(fmt.Sprintf("state-%d", i)),
			TxRoot:     s.hash(fmt.Sprintf("tx-%d", i)),
		}
		header.Hash = s.hashHeader(header)
		s.headers = append(s.headers, header)
	}

	s.updateState()
	return nil
}

// SyncHeader 同步区块头
func (s *LightClientSimulator) SyncHeader(header *BlockHeader) error {
	if len(s.headers) > 0 {
		lastHeader := s.headers[len(s.headers)-1]
		if header.ParentHash != lastHeader.Hash {
			return fmt.Errorf("invalid parent hash")
		}
		if header.Height != lastHeader.Height+1 {
			return fmt.Errorf("invalid height")
		}
	}

	computedHash := s.hashHeader(header)
	if header.Hash != computedHash {
		return fmt.Errorf("invalid header hash")
	}

	s.headers = append(s.headers, header)
	s.EmitEvent("header_synced", "", "", map[string]interface{}{
		"height": header.Height, "hash": header.Hash[:16],
	})
	s.updateState()
	return nil
}

// VerifySPVProof 验证SPV证明
func (s *LightClientSimulator) VerifySPVProof(proof *SPVProof) bool {
	var blockHeader *BlockHeader
	for _, h := range s.headers {
		if h.Hash == proof.BlockHash {
			blockHeader = h
			break
		}
	}
	if blockHeader == nil {
		s.EmitEvent("spv_verify_failed", "", "", map[string]interface{}{
			"reason": "block not found",
		})
		return false
	}

	currentHash := proof.TxHash
	for i, sibling := range proof.MerklePath {
		if proof.Directions[i] == "left" {
			currentHash = s.hash(sibling + currentHash)
		} else {
			currentHash = s.hash(currentHash + sibling)
		}
	}

	valid := currentHash == blockHeader.TxRoot
	if valid {
		s.verifiedTxs = append(s.verifiedTxs, proof.TxHash)
	}

	s.EmitEvent("spv_verified", "", "", map[string]interface{}{
		"tx_hash": proof.TxHash[:16], "valid": valid, "block_height": blockHeader.Height,
	})
	s.updateState()
	return valid
}

// GetHeaderByHeight 根据高度获取区块头
func (s *LightClientSimulator) GetHeaderByHeight(height uint64) *BlockHeader {
	if int(height) < len(s.headers) {
		return s.headers[height]
	}
	return nil
}

func (s *LightClientSimulator) hash(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func (s *LightClientSimulator) hashHeader(h *BlockHeader) string {
	data := fmt.Sprintf("%s%d%s%s", h.ParentHash, h.Height, h.StateRoot, h.TxRoot)
	return s.hash(data)
}

func (s *LightClientSimulator) updateState() {
	s.SetGlobalData("header_count", len(s.headers))
	s.SetGlobalData("latest_height", len(s.headers)-1)
	s.SetGlobalData("verified_tx_count", len(s.verifiedTxs))
	if len(s.headers) > 0 {
		s.SetGlobalData("latest_hash", s.headers[len(s.headers)-1].Hash)
	}

	summary := fmt.Sprintf("当前轻客户端已同步 %d 个区块头，已验证 %d 笔交易。", len(s.headers), len(s.verifiedTxs))
	nextHint := "可以继续同步新的区块头，或验证一条交易证明路径。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备轻客户端",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"header_count": len(s.headers), "verified_tx_count": len(s.verifiedTxs)},
	)
}

func (s *LightClientSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "sync_next_header":
		last := s.headers[len(s.headers)-1]
		header := &BlockHeader{
			ParentHash: last.Hash,
			Height:     last.Height + 1,
			StateRoot:  s.hash(fmt.Sprintf("state-%d", last.Height+1)),
			TxRoot:     s.hash(fmt.Sprintf("tx-%d", last.Height+1)),
		}
		header.Hash = s.hashHeader(header)
		if err := s.SyncHeader(header); err != nil {
			return nil, err
		}
		return blockchainActionResult("已同步一个新的区块头。", map[string]interface{}{"height": header.Height, "hash": header.Hash}, &types.ActionFeedback{
			Summary:     "轻客户端已经接收并验证了一个新的区块头。",
			NextHint:    "继续验证某笔交易的 SPV 证明，观察仅凭区块头如何确认交易存在。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"height": header.Height},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported light client action: %s", action)
	}
}

type LightClientFactory struct{}

func (f *LightClientFactory) Create() engine.Simulator { return NewLightClientSimulator() }
func (f *LightClientFactory) GetDescription() types.Description {
	return NewLightClientSimulator().GetDescription()
}
func NewLightClientFactory() *LightClientFactory { return &LightClientFactory{} }

var _ engine.SimulatorFactory = (*LightClientFactory)(nil)
