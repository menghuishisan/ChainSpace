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
// 数据可用性演示器
// 演示区块链数据可用性层(DA Layer)的核心机制
//
// 核心概念:
// 1. 数据可用性: 确保区块数据可被任何人获取和验证
// 2. 纠删码: 使用Reed-Solomon编码实现数据冗余
// 3. 数据可用性采样(DAS): 轻节点验证数据可用性
// 4. 数据承诺: 使用KZG承诺证明数据完整性
//
// 参考: Celestia, EigenDA, Avail, EIP-4844
// =============================================================================

// DataBlobStatus 数据块状态
type DataBlobStatus string

const (
	BlobPending   DataBlobStatus = "pending"
	BlobPublished DataBlobStatus = "published"
	BlobVerified  DataBlobStatus = "verified"
	BlobExpired   DataBlobStatus = "expired"
)

// DataBlob 数据块
type DataBlob struct {
	BlobID        string         `json:"blob_id"`
	Namespace     string         `json:"namespace"`
	Data          []byte         `json:"-"`
	DataSize      int            `json:"data_size"`
	Commitment    string         `json:"commitment"`
	EncodedShares int            `json:"encoded_shares"`
	Status        DataBlobStatus `json:"status"`
	Publisher     string         `json:"publisher"`
	PublishedAt   time.Time      `json:"published_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
}

// ErasureCodedShare 纠删码分片
type ErasureCodedShare struct {
	ShareID       string `json:"share_id"`
	BlobID        string `json:"blob_id"`
	Index         int    `json:"index"`
	Data          []byte `json:"-"`
	DataSize      int    `json:"data_size"`
	IsParityShare bool   `json:"is_parity"`
}

// DASample 数据可用性采样
type DASample struct {
	SampleID     string    `json:"sample_id"`
	BlobID       string    `json:"blob_id"`
	Sampler      string    `json:"sampler"`
	ShareIndices []int     `json:"share_indices"`
	AllAvailable bool      `json:"all_available"`
	SampledAt    time.Time `json:"sampled_at"`
}

// DataAvailabilitySimulator 数据可用性演示器
type DataAvailabilitySimulator struct {
	*base.BaseSimulator
	blobs           map[string]*DataBlob
	shares          map[string][]*ErasureCodedShare
	samples         map[string]*DASample
	namespaces      map[string][]string
	encodingRate    float64
	retentionPeriod time.Duration
	sampleSize      int
}

// NewDataAvailabilitySimulator 创建演示器
func NewDataAvailabilitySimulator() *DataAvailabilitySimulator {
	sim := &DataAvailabilitySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"data_availability",
			"数据可用性演示器",
			"演示DA层的纠删码、数据可用性采样、KZG承诺等核心机制",
			"crosschain",
			types.ComponentDemo,
		),
		blobs:      make(map[string]*DataBlob),
		shares:     make(map[string][]*ErasureCodedShare),
		samples:    make(map[string]*DASample),
		namespaces: make(map[string][]string),
	}

	sim.AddParam(types.Param{
		Key:         "encoding_rate",
		Name:        "编码冗余率",
		Description: "纠删码冗余倍数(2表示2倍冗余)",
		Type:        types.ParamTypeFloat,
		Default:     2.0,
		Min:         1.5,
		Max:         4.0,
	})

	sim.AddParam(types.Param{
		Key:         "sample_size",
		Name:        "采样数量",
		Description: "DAS每次采样的分片数",
		Type:        types.ParamTypeInt,
		Default:     16,
		Min:         4,
		Max:         64,
	})

	return sim
}

// Init 初始化
func (s *DataAvailabilitySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.encodingRate = 2.0
	s.sampleSize = 16
	s.retentionPeriod = 30 * 24 * time.Hour

	if v, ok := config.Params["encoding_rate"]; ok {
		if n, ok := v.(float64); ok {
			s.encodingRate = n
		}
	}
	if v, ok := config.Params["sample_size"]; ok {
		if n, ok := v.(float64); ok {
			s.sampleSize = int(n)
		}
	}

	s.blobs = make(map[string]*DataBlob)
	s.shares = make(map[string][]*ErasureCodedShare)
	s.samples = make(map[string]*DASample)
	s.namespaces = make(map[string][]string)

	s.updateState()
	return nil
}

// ExplainDataAvailability 解释机制
func (s *DataAvailabilitySimulator) ExplainDataAvailability() map[string]interface{} {
	return map[string]interface{}{
		"overview": "数据可用性层确保Rollup交易数据可被任何人获取，是L2安全的关键",
		"problem": map[string]string{
			"issue":  "L2 Sequencer可能隐藏数据，阻止用户验证或退出",
			"impact": "用户资产被锁定，无法证明自己的余额",
		},
		"solutions": []map[string]interface{}{
			{
				"name":    "纠删码(Erasure Coding)",
				"desc":    "将数据编码为多个分片，只需部分分片即可恢复",
				"example": fmt.Sprintf("%.0f倍冗余：任意50%%分片可恢复全部数据", s.encodingRate),
			},
			{
				"name":    "数据可用性采样(DAS)",
				"desc":    "轻节点随机采样少量分片，概率验证数据可用",
				"example": fmt.Sprintf("采样%d个分片，如全部可用则高概率数据完整", s.sampleSize),
			},
			{
				"name":     "KZG承诺",
				"desc":     "多项式承诺方案，可高效验证数据片段属于完整数据",
				"property": "绑定性、隐藏性、高效验证",
			},
		},
		"implementations": []map[string]string{
			{"name": "Celestia", "type": "专用DA链", "feature": "模块化区块链"},
			{"name": "EigenDA", "type": "再质押DA", "feature": "利用ETH质押安全性"},
			{"name": "Avail", "type": "专用DA链", "feature": "Polygon生态"},
			{"name": "EIP-4844", "type": "以太坊内置", "feature": "Proto-Danksharding"},
		},
		"eip4844": map[string]interface{}{
			"name":            "Proto-Danksharding",
			"blob_size":       "128KB",
			"blobs_per_block": 6,
			"retention":       "~18天",
			"benefit":         "L2 Gas费用降低10-100倍",
		},
	}
}

// PublishBlob 发布数据块
func (s *DataAvailabilitySimulator) PublishBlob(publisher, namespace string, data []byte) (*DataBlob, error) {
	blobData := fmt.Sprintf("%s-%s-%d", publisher, namespace, time.Now().UnixNano())
	blobHash := sha256.Sum256([]byte(blobData))
	blobID := fmt.Sprintf("blob-%s", hex.EncodeToString(blobHash[:8]))

	commitmentData := append(data, blobHash[:]...)
	commitment := sha256.Sum256(commitmentData)

	totalShares := int(float64(len(data)/32+1) * s.encodingRate)

	blob := &DataBlob{
		BlobID:        blobID,
		Namespace:     namespace,
		Data:          data,
		DataSize:      len(data),
		Commitment:    "0x" + hex.EncodeToString(commitment[:]),
		EncodedShares: totalShares,
		Status:        BlobPublished,
		Publisher:     publisher,
		PublishedAt:   time.Now(),
		ExpiresAt:     time.Now().Add(s.retentionPeriod),
	}

	s.blobs[blobID] = blob
	s.namespaces[namespace] = append(s.namespaces[namespace], blobID)

	shares := s.encodeToShares(blobID, data, totalShares)
	s.shares[blobID] = shares

	s.EmitEvent("blob_published", "", "", map[string]interface{}{
		"blob_id":    blobID,
		"namespace":  namespace,
		"data_size":  len(data),
		"shares":     totalShares,
		"commitment": blob.Commitment[:20] + "...",
	})

	s.updateState()
	return blob, nil
}

// encodeToShares 编码为分片
func (s *DataAvailabilitySimulator) encodeToShares(blobID string, data []byte, totalShares int) []*ErasureCodedShare {
	shares := make([]*ErasureCodedShare, totalShares)
	dataShares := totalShares / int(s.encodingRate)

	for i := 0; i < totalShares; i++ {
		shareData := fmt.Sprintf("%s-share-%d", blobID, i)
		shareHash := sha256.Sum256([]byte(shareData))

		shares[i] = &ErasureCodedShare{
			ShareID:       fmt.Sprintf("share-%s-%d", blobID[:12], i),
			BlobID:        blobID,
			Index:         i,
			Data:          shareHash[:],
			DataSize:      32,
			IsParityShare: i >= dataShares,
		}
	}

	return shares
}

// SampleBlob 数据可用性采样
func (s *DataAvailabilitySimulator) SampleBlob(blobID, sampler string) (*DASample, error) {
	blob, ok := s.blobs[blobID]
	if !ok {
		return nil, fmt.Errorf("数据块不存在: %s", blobID)
	}

	shares := s.shares[blobID]
	if len(shares) == 0 {
		return nil, fmt.Errorf("分片不存在")
	}

	sampleIndices := make([]int, 0, s.sampleSize)
	for i := 0; i < s.sampleSize && i < len(shares); i++ {
		idx := (i * 7) % len(shares)
		sampleIndices = append(sampleIndices, idx)
	}

	allAvailable := true

	sampleData := fmt.Sprintf("%s-%s-%d", blobID, sampler, time.Now().UnixNano())
	sampleHash := sha256.Sum256([]byte(sampleData))

	sample := &DASample{
		SampleID:     fmt.Sprintf("sample-%s", hex.EncodeToString(sampleHash[:8])),
		BlobID:       blobID,
		Sampler:      sampler,
		ShareIndices: sampleIndices,
		AllAvailable: allAvailable,
		SampledAt:    time.Now(),
	}

	s.samples[sample.SampleID] = sample

	if allAvailable && blob.Status == BlobPublished {
		blob.Status = BlobVerified
	}

	s.EmitEvent("blob_sampled", "", "", map[string]interface{}{
		"sample_id":     sample.SampleID,
		"blob_id":       blobID,
		"sampler":       sampler,
		"samples":       len(sampleIndices),
		"all_available": allAvailable,
	})

	s.updateState()
	return sample, nil
}

// VerifyCommitment 验证承诺
func (s *DataAvailabilitySimulator) VerifyCommitment(blobID string, shareIndex int) map[string]interface{} {
	blob, ok := s.blobs[blobID]
	if !ok {
		return map[string]interface{}{"error": "数据块不存在"}
	}

	shares := s.shares[blobID]
	if shareIndex >= len(shares) {
		return map[string]interface{}{"error": "分片索引越界"}
	}

	share := shares[shareIndex]

	return map[string]interface{}{
		"blob_id":      blobID,
		"share_index":  shareIndex,
		"commitment":   blob.Commitment,
		"share_data":   hex.EncodeToString(share.Data[:8]) + "...",
		"verified":     true,
		"proof_type":   "KZG",
		"verification": "分片数据与承诺匹配",
	}
}

// GetStatistics 获取统计
func (s *DataAvailabilitySimulator) GetStatistics() map[string]interface{} {
	totalSize := 0
	verified := 0
	for _, b := range s.blobs {
		totalSize += b.DataSize
		if b.Status == BlobVerified {
			verified++
		}
	}

	return map[string]interface{}{
		"total_blobs":     len(s.blobs),
		"verified_blobs":  verified,
		"total_samples":   len(s.samples),
		"total_data_size": totalSize,
		"namespaces":      len(s.namespaces),
		"encoding_rate":   s.encodingRate,
		"sample_size":     s.sampleSize,
	}
}

func (s *DataAvailabilitySimulator) updateState() {
	s.SetGlobalData("blob_count", len(s.blobs))
	s.SetGlobalData("sample_count", len(s.samples))

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"data_availability",
		"当前可以发布 blob 并执行数据可用性采样。",
		"先发布一份数据，再观察采样和承诺验证如何共同证明数据可用。",
		0,
		map[string]interface{}{
			"blob_count":   len(s.blobs),
			"sample_count": len(s.samples),
		},
	)
}

func (s *DataAvailabilitySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "publish_blob":
		publisher := "publisher-1"
		namespace := "rollup"
		payload := []byte("sample-da-payload")
		if raw, ok := params["publisher"].(string); ok && raw != "" {
			publisher = raw
		}
		if raw, ok := params["namespace"].(string); ok && raw != "" {
			namespace = raw
		}
		blob, err := s.PublishBlob(publisher, namespace, payload)
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已发布数据 blob",
			map[string]interface{}{"blob_id": blob.BlobID, "commitment": blob.Commitment},
			&types.ActionFeedback{
				Summary:     "数据已经编码并发布，可继续进行抽样或承诺验证。",
				NextHint:    "执行一次 sample_blob，观察随机采样是否能覆盖足够的数据分片。",
				EffectScope: "crosschain",
			},
		), nil
	case "sample_blob":
		blobID := ""
		if raw, ok := params["blob_id"].(string); ok && raw != "" {
			blobID = raw
		}
		if blobID == "" {
			for id := range s.blobs {
				blobID = id
				break
			}
		}
		if blobID == "" {
			return nil, fmt.Errorf("没有可采样的 blob")
		}
		sample, err := s.SampleBlob(blobID, "sampler-1")
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已完成数据可用性采样",
			map[string]interface{}{"sample_id": sample.SampleID, "blob_id": sample.BlobID},
			&types.ActionFeedback{
				Summary:     "采样已经完成，可以继续验证承诺或比较不同 blob 的可用性。",
				NextHint:    "执行一次 verify_commitment，观察采样与承诺验证如何配合。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported data availability action: %s", action)
	}
}

// Factory
type DataAvailabilityFactory struct{}

func (f *DataAvailabilityFactory) Create() engine.Simulator { return NewDataAvailabilitySimulator() }
func (f *DataAvailabilityFactory) GetDescription() types.Description {
	return NewDataAvailabilitySimulator().GetDescription()
}
func NewDataAvailabilityFactory() *DataAvailabilityFactory { return &DataAvailabilityFactory{} }

var _ engine.SimulatorFactory = (*DataAvailabilityFactory)(nil)
