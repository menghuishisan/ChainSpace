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
// 预言机桥演示器
// 演示使用预言机网络进行跨链数据和资产传输
//
// 预言机桥特点:
// 1. 依赖去中心化预言机网络验证跨链消息
// 2. 预言机节点监听源链事件并达成共识
// 3. 使用加密经济激励确保诚实行为
//
// 安全模型:
// - 依赖预言机网络的诚实多数假设
// - 质押和惩罚机制
// - 多源验证
//
// 参考: Chainlink CCIP, Band Protocol, API3
// =============================================================================

// OracleNode 预言机节点
type OracleNode struct {
	NodeID          string    `json:"node_id"`
	Address         string    `json:"address"`
	Stake           *big.Int  `json:"stake"`
	Reputation      float64   `json:"reputation"`
	ReportsCount    int       `json:"reports_count"`
	AccurateCount   int       `json:"accurate_count"`
	SlashedCount    int       `json:"slashed_count"`
	IsActive        bool      `json:"is_active"`
	SupportedChains []string  `json:"supported_chains"`
	LastActiveAt    time.Time `json:"last_active_at"`
}

// OracleReport 预言机报告
type OracleReport struct {
	ReportID    string    `json:"report_id"`
	NodeID      string    `json:"node_id"`
	MessageID   string    `json:"message_id"`
	SourceChain string    `json:"source_chain"`
	DestChain   string    `json:"dest_chain"`
	Data        []byte    `json:"data"`
	DataHash    string    `json:"data_hash"`
	Signature   string    `json:"signature"`
	Timestamp   time.Time `json:"timestamp"`
	IsValid     bool      `json:"is_valid"`
}

// OracleConsensus 预言机共识
type OracleConsensus struct {
	MessageID     string          `json:"message_id"`
	Reports       []*OracleReport `json:"reports"`
	AgreedData    []byte          `json:"agreed_data"`
	AgreementRate float64         `json:"agreement_rate"`
	IsFinalized   bool            `json:"is_finalized"`
	FinalizedAt   time.Time       `json:"finalized_at"`
}

// CrossChainRequest 跨链请求
type CrossChainRequest struct {
	RequestID   string           `json:"request_id"`
	Requester   string           `json:"requester"`
	SourceChain string           `json:"source_chain"`
	DestChain   string           `json:"dest_chain"`
	RequestType string           `json:"request_type"`
	Payload     []byte           `json:"payload"`
	Fee         *big.Int         `json:"fee"`
	Status      string           `json:"status"`
	Consensus   *OracleConsensus `json:"consensus"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt time.Time        `json:"completed_at"`
}

// OracleBridgeSimulator 预言机桥演示器
type OracleBridgeSimulator struct {
	*base.BaseSimulator
	nodes              map[string]*OracleNode
	reports            map[string]*OracleReport
	requests           map[string]*CrossChainRequest
	consensusThreshold float64
	minReports         int
	totalRequests      int
	totalFees          *big.Int
}

// NewOracleBridgeSimulator 创建预言机桥演示器
func NewOracleBridgeSimulator() *OracleBridgeSimulator {
	sim := &OracleBridgeSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"oracle_bridge",
			"预言机桥演示器",
			"演示使用预言机网络进行跨链数据传输和验证的机制",
			"crosschain",
			types.ComponentProcess,
		),
		nodes:     make(map[string]*OracleNode),
		reports:   make(map[string]*OracleReport),
		requests:  make(map[string]*CrossChainRequest),
		totalFees: big.NewInt(0),
	}

	sim.AddParam(types.Param{
		Key:         "consensus_threshold",
		Name:        "共识阈值",
		Description: "达成共识所需的节点同意比例",
		Type:        types.ParamTypeFloat,
		Default:     0.67,
		Min:         0.5,
		Max:         1.0,
	})

	sim.AddParam(types.Param{
		Key:         "min_reports",
		Name:        "最小报告数",
		Description: "达成共识所需的最小报告数量",
		Type:        types.ParamTypeInt,
		Default:     3,
		Min:         1,
		Max:         21,
	})

	return sim
}

// Init 初始化
func (s *OracleBridgeSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.consensusThreshold = 0.67
	s.minReports = 3

	if v, ok := config.Params["consensus_threshold"]; ok {
		if f, ok := v.(float64); ok {
			s.consensusThreshold = f
		}
	}
	if v, ok := config.Params["min_reports"]; ok {
		if n, ok := v.(float64); ok {
			s.minReports = int(n)
		}
	}

	s.nodes = make(map[string]*OracleNode)
	s.reports = make(map[string]*OracleReport)
	s.requests = make(map[string]*CrossChainRequest)
	s.totalFees = big.NewInt(0)
	s.totalRequests = 0

	s.initializeNodes()
	s.updateState()
	return nil
}

// initializeNodes 初始化预言机节点
func (s *OracleBridgeSimulator) initializeNodes() {
	nodes := []struct {
		name   string
		stake  int64
		chains []string
	}{
		{"Oracle-Alpha", 100000, []string{"ethereum", "polygon", "arbitrum", "optimism"}},
		{"Oracle-Beta", 80000, []string{"ethereum", "polygon", "bsc"}},
		{"Oracle-Gamma", 75000, []string{"ethereum", "arbitrum", "avalanche"}},
		{"Oracle-Delta", 70000, []string{"ethereum", "polygon"}},
		{"Oracle-Epsilon", 65000, []string{"ethereum", "bsc", "avalanche"}},
		{"Oracle-Zeta", 60000, []string{"ethereum", "polygon", "arbitrum"}},
		{"Oracle-Eta", 55000, []string{"ethereum", "optimism"}},
	}

	for _, n := range nodes {
		hash := sha256.Sum256([]byte(n.name))
		addr := fmt.Sprintf("0x%s", hex.EncodeToString(hash[:20]))
		s.nodes[n.name] = &OracleNode{
			NodeID:          n.name,
			Address:         addr,
			Stake:           new(big.Int).Mul(big.NewInt(n.stake), big.NewInt(1e18)),
			Reputation:      0.95,
			ReportsCount:    0,
			AccurateCount:   0,
			SlashedCount:    0,
			IsActive:        true,
			SupportedChains: n.chains,
			LastActiveAt:    time.Now(),
		}
	}
}

// =============================================================================
// 预言机桥机制解释
// =============================================================================

// ExplainOracleBridge 解释预言机桥
func (s *OracleBridgeSimulator) ExplainOracleBridge() map[string]interface{} {
	return map[string]interface{}{
		"overview": "预言机桥使用去中心化预言机网络验证和传递跨链消息",
		"how_it_works": []map[string]string{
			{"step": "1", "action": "用户在源链发起跨链请求"},
			{"step": "2", "action": "预言机节点监听源链事件"},
			{"step": "3", "action": "每个节点独立验证事件并生成报告"},
			{"step": "4", "action": "节点对报告进行签名"},
			{"step": "5", "action": "收集足够多的签名报告"},
			{"step": "6", "action": "达成共识后执行目标链操作"},
		},
		"security_model": map[string]interface{}{
			"trust_assumption":    "多数预言机节点诚实",
			"consensus_threshold": fmt.Sprintf("%.0f%%", s.consensusThreshold*100),
			"min_reports":         s.minReports,
			"incentives": []string{
				"质押机制: 节点需质押代币",
				"报告奖励: 正确报告获得奖励",
				"惩罚机制: 错误报告被罚没质押",
				"声誉系统: 历史表现影响权重",
			},
		},
		"vs_other_bridges": map[string]interface{}{
			"vs_multisig": map[string]string{
				"oracle":   "动态节点集，可扩展",
				"multisig": "固定验证者集合",
			},
			"vs_light_client": map[string]string{
				"oracle":       "更灵活，支持任意链",
				"light_client": "更安全，继承源链安全性",
			},
		},
		"real_implementations": []map[string]string{
			{"name": "Chainlink CCIP", "feature": "去中心化预言机网络+风险管理"},
			{"name": "Band Protocol", "feature": "跨链数据预言机"},
			{"name": "API3", "feature": "第一方预言机"},
		},
	}
}

// =============================================================================
// 预言机操作
// =============================================================================

// CreateRequest 创建跨链请求
func (s *OracleBridgeSimulator) CreateRequest(requester, sourceChain, destChain, requestType string, payload []byte, fee *big.Int) (*CrossChainRequest, error) {
	reqData := fmt.Sprintf("%s-%s-%s-%d", requester, sourceChain, destChain, time.Now().UnixNano())
	reqHash := sha256.Sum256([]byte(reqData))
	reqID := fmt.Sprintf("req-%s", hex.EncodeToString(reqHash[:8]))

	request := &CrossChainRequest{
		RequestID:   reqID,
		Requester:   requester,
		SourceChain: sourceChain,
		DestChain:   destChain,
		RequestType: requestType,
		Payload:     payload,
		Fee:         fee,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	s.requests[reqID] = request
	s.totalRequests++
	s.totalFees.Add(s.totalFees, fee)

	s.EmitEvent("request_created", "", "", map[string]interface{}{
		"request_id":   reqID,
		"requester":    requester,
		"source_chain": sourceChain,
		"dest_chain":   destChain,
		"request_type": requestType,
	})

	s.updateState()
	return request, nil
}

// SubmitReport 提交报告
func (s *OracleBridgeSimulator) SubmitReport(nodeID, requestID string, data []byte) (*OracleReport, error) {
	node, ok := s.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("节点不存在: %s", nodeID)
	}

	if !node.IsActive {
		return nil, fmt.Errorf("节点不活跃: %s", nodeID)
	}

	request, ok := s.requests[requestID]
	if !ok {
		return nil, fmt.Errorf("请求不存在: %s", requestID)
	}

	dataHash := sha256.Sum256(data)
	sigData := fmt.Sprintf("%s-%s-%s", nodeID, requestID, hex.EncodeToString(dataHash[:]))
	sigHash := sha256.Sum256([]byte(sigData))

	report := &OracleReport{
		ReportID:    fmt.Sprintf("report-%s-%s", nodeID, requestID),
		NodeID:      nodeID,
		MessageID:   requestID,
		SourceChain: request.SourceChain,
		DestChain:   request.DestChain,
		Data:        data,
		DataHash:    hex.EncodeToString(dataHash[:]),
		Signature:   hex.EncodeToString(sigHash[:]),
		Timestamp:   time.Now(),
		IsValid:     true,
	}

	s.reports[report.ReportID] = report
	node.ReportsCount++
	node.LastActiveAt = time.Now()

	s.EmitEvent("report_submitted", "", "", map[string]interface{}{
		"report_id":  report.ReportID,
		"node_id":    nodeID,
		"request_id": requestID,
		"data_hash":  report.DataHash[:16] + "...",
	})

	s.updateState()
	return report, nil
}

// CheckConsensus 检查共识
func (s *OracleBridgeSimulator) CheckConsensus(requestID string) (*OracleConsensus, error) {
	request, ok := s.requests[requestID]
	if !ok {
		return nil, fmt.Errorf("请求不存在: %s", requestID)
	}

	reports := make([]*OracleReport, 0)
	for _, report := range s.reports {
		if report.MessageID == requestID && report.IsValid {
			reports = append(reports, report)
		}
	}

	if len(reports) < s.minReports {
		return nil, fmt.Errorf("报告数不足: %d/%d", len(reports), s.minReports)
	}

	dataVotes := make(map[string]int)
	for _, report := range reports {
		dataVotes[report.DataHash]++
	}

	var maxVotes int
	var agreedHash string
	for hash, votes := range dataVotes {
		if votes > maxVotes {
			maxVotes = votes
			agreedHash = hash
		}
	}

	agreementRate := float64(maxVotes) / float64(len(reports))

	var agreedData []byte
	for _, report := range reports {
		if report.DataHash == agreedHash {
			agreedData = report.Data
			break
		}
	}

	consensus := &OracleConsensus{
		MessageID:     requestID,
		Reports:       reports,
		AgreedData:    agreedData,
		AgreementRate: agreementRate,
		IsFinalized:   agreementRate >= s.consensusThreshold,
	}

	if consensus.IsFinalized {
		consensus.FinalizedAt = time.Now()
		request.Consensus = consensus
		request.Status = "consensus_reached"

		for _, report := range reports {
			if node, ok := s.nodes[report.NodeID]; ok {
				if report.DataHash == agreedHash {
					node.AccurateCount++
				}
			}
		}

		s.EmitEvent("consensus_reached", "", "", map[string]interface{}{
			"request_id":     requestID,
			"agreement_rate": fmt.Sprintf("%.2f%%", agreementRate*100),
			"reports_count":  len(reports),
		})
	}

	s.updateState()
	return consensus, nil
}

// ExecuteRequest 执行请求
func (s *OracleBridgeSimulator) ExecuteRequest(requestID string) error {
	request, ok := s.requests[requestID]
	if !ok {
		return fmt.Errorf("请求不存在: %s", requestID)
	}

	if request.Status != "consensus_reached" {
		return fmt.Errorf("未达成共识: %s", request.Status)
	}

	request.Status = "executed"
	request.CompletedAt = time.Now()

	s.EmitEvent("request_executed", "", "", map[string]interface{}{
		"request_id": requestID,
		"duration":   request.CompletedAt.Sub(request.CreatedAt).String(),
	})

	s.updateState()
	return nil
}

// SimulateOracleBridgeFlow 模拟完整流程
func (s *OracleBridgeSimulator) SimulateOracleBridgeFlow(requester string, payload []byte) map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	fee := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e16))
	request, _ := s.CreateRequest(requester, "ethereum", "polygon", "transfer", payload, fee)
	steps = append(steps, map[string]interface{}{
		"step": 1, "action": "用户发起跨链请求",
		"request_id": request.RequestID,
	})

	activeNodes := make([]string, 0)
	for name, node := range s.nodes {
		if node.IsActive {
			activeNodes = append(activeNodes, name)
		}
	}

	for i := 0; i < s.minReports+2 && i < len(activeNodes); i++ {
		s.SubmitReport(activeNodes[i], request.RequestID, payload)
	}
	steps = append(steps, map[string]interface{}{
		"step": 2, "action": "预言机节点提交报告",
		"reports": s.minReports + 2,
	})

	consensus, _ := s.CheckConsensus(request.RequestID)
	steps = append(steps, map[string]interface{}{
		"step": 3, "action": "检查共识",
		"agreement_rate": fmt.Sprintf("%.2f%%", consensus.AgreementRate*100),
		"finalized":      consensus.IsFinalized,
	})

	if consensus.IsFinalized {
		s.ExecuteRequest(request.RequestID)
		steps = append(steps, map[string]interface{}{
			"step": 4, "action": "执行跨链操作",
			"status": request.Status,
		})
	}

	return map[string]interface{}{
		"request_id": request.RequestID,
		"steps":      steps,
		"duration":   request.CompletedAt.Sub(request.CreatedAt).String(),
	}
}

// GetStatistics 获取统计
func (s *OracleBridgeSimulator) GetStatistics() map[string]interface{} {
	activeNodes := 0
	for _, node := range s.nodes {
		if node.IsActive {
			activeNodes++
		}
	}

	return map[string]interface{}{
		"total_nodes":         len(s.nodes),
		"active_nodes":        activeNodes,
		"total_reports":       len(s.reports),
		"total_requests":      s.totalRequests,
		"total_fees":          s.totalFees.String(),
		"consensus_threshold": fmt.Sprintf("%.0f%%", s.consensusThreshold*100),
		"min_reports":         s.minReports,
	}
}

// updateState 更新状态
func (s *OracleBridgeSimulator) updateState() {
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("request_count", s.totalRequests)

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"oracle_bridge",
		"当前可以发起跨链请求，并观察预言机节点如何提交报告并形成共识。",
		"先创建一个请求，再提交报告，观察报告汇聚与共识形成的过程。",
		0,
		map[string]interface{}{
			"node_count":    len(s.nodes),
			"request_count": s.totalRequests,
		},
	)
}

func (s *OracleBridgeSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_request":
		request, err := s.CreateRequest("student", "ethereum", "polygon", "price_feed", []byte("eth/usdc"), big.NewInt(1))
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已创建预言机跨链请求",
			map[string]interface{}{"request_id": request.RequestID},
			&types.ActionFeedback{
				Summary:     "跨链请求已经进入等待报告阶段，可继续让预言机节点提交报告。",
				NextHint:    "执行 submit_report，观察多个节点报告如何逐步形成共识。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported oracle bridge action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// OracleBridgeFactory 预言机桥工厂
type OracleBridgeFactory struct{}

func (f *OracleBridgeFactory) Create() engine.Simulator { return NewOracleBridgeSimulator() }
func (f *OracleBridgeFactory) GetDescription() types.Description {
	return NewOracleBridgeSimulator().GetDescription()
}
func NewOracleBridgeFactory() *OracleBridgeFactory { return &OracleBridgeFactory{} }

var _ engine.SimulatorFactory = (*OracleBridgeFactory)(nil)
