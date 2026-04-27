package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	evmpkg "github.com/chainspace/simulations/pkg/evm"
	"github.com/chainspace/simulations/pkg/types"
	"golang.org/x/crypto/sha3"
)

// =============================================================================
// 事件日志演示器
// =============================================================================

// EventDefinition 事件定义
type EventDefinition struct {
	Name      string   `json:"name"`
	Signature string   `json:"signature"`
	Topic0    string   `json:"topic0"`
	Indexed   []string `json:"indexed"`
	Data      []string `json:"data"`
}

// DecodedLog 解码后的日志
type DecodedLog struct {
	Address   string                 `json:"address"`
	EventName string                 `json:"event_name"`
	Signature string                 `json:"signature"`
	Topics    []string               `json:"topics"`
	Data      string                 `json:"data"`
	Decoded   map[string]interface{} `json:"decoded"`
}

// EventLogSimulator 事件日志演示器
// 演示EVM事件日志的工作原理:
// - Topic结构
// - indexed vs non-indexed参数
// - 事件签名计算
// - 日志解码
type EventLogSimulator struct {
	*base.BaseSimulator
	events      map[string]*EventDefinition
	decodedLogs []*DecodedLog
}

// NewEventLogSimulator 创建事件日志演示器
func NewEventLogSimulator() *EventLogSimulator {
	sim := &EventLogSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"event_log",
			"事件日志演示器",
			"演示EVM事件日志的结构、编码和解码",
			"evm",
			types.ComponentDemo,
		),
		events:      make(map[string]*EventDefinition),
		decodedLogs: make([]*DecodedLog, 0),
	}

	return sim
}

// Init 初始化
func (s *EventLogSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	// 注册常见事件
	s.registerCommonEvents()

	s.updateState()
	return nil
}

// registerCommonEvents 注册常见事件
func (s *EventLogSimulator) registerCommonEvents() {
	// ERC20事件
	s.RegisterEvent("Transfer", "Transfer(address,address,uint256)",
		[]string{"from", "to"}, []string{"value"})
	s.RegisterEvent("Approval", "Approval(address,address,uint256)",
		[]string{"owner", "spender"}, []string{"value"})

	// ERC721事件
	s.RegisterEvent("Transfer721", "Transfer(address,address,uint256)",
		[]string{"from", "to", "tokenId"}, []string{})
	s.RegisterEvent("ApprovalForAll", "ApprovalForAll(address,address,bool)",
		[]string{"owner", "operator"}, []string{"approved"})

	// Uniswap事件
	s.RegisterEvent("Swap", "Swap(address,uint256,uint256,uint256,uint256,address)",
		[]string{"sender", "to"}, []string{"amount0In", "amount1In", "amount0Out", "amount1Out"})
	s.RegisterEvent("Sync", "Sync(uint112,uint112)",
		[]string{}, []string{"reserve0", "reserve1"})
	s.RegisterEvent("Mint", "Mint(address,uint256,uint256)",
		[]string{"sender"}, []string{"amount0", "amount1"})
	s.RegisterEvent("Burn", "Burn(address,uint256,uint256,address)",
		[]string{"sender", "to"}, []string{"amount0", "amount1"})

	// 治理事件
	s.RegisterEvent("ProposalCreated", "ProposalCreated(uint256,address,address[],uint256[],string[],bytes[],uint256,uint256,string)",
		[]string{}, []string{"proposalId", "proposer", "targets", "values", "signatures", "calldatas", "startBlock", "endBlock", "description"})
	s.RegisterEvent("VoteCast", "VoteCast(address,uint256,uint8,uint256,string)",
		[]string{"voter"}, []string{"proposalId", "support", "weight", "reason"})
}

// RegisterEvent 注册事件定义
func (s *EventLogSimulator) RegisterEvent(name, signature string, indexed, data []string) *EventDefinition {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(signature))
	topic0 := "0x" + hex.EncodeToString(h.Sum(nil))

	event := &EventDefinition{
		Name:      name,
		Signature: signature,
		Topic0:    topic0,
		Indexed:   indexed,
		Data:      data,
	}

	s.events[name] = event
	s.events[topic0] = event

	return event
}

// CalculateEventTopic 计算事件Topic
func (s *EventLogSimulator) CalculateEventTopic(signature string) map[string]interface{} {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(signature))
	topic := h.Sum(nil)

	result := map[string]interface{}{
		"signature":   signature,
		"topic0":      "0x" + hex.EncodeToString(topic),
		"explanation": "topic0 = keccak256(event_signature)",
	}

	s.EmitEvent("topic_calculated", "", "", result)

	return result
}

// ExplainLogStructure 解释日志结构
func (s *EventLogSimulator) ExplainLogStructure() map[string]interface{} {
	return map[string]interface{}{
		"structure": map[string]interface{}{
			"address": "发出事件的合约地址",
			"topics": map[string]string{
				"topic0":   "事件签名的keccak256哈希 (匿名事件除外)",
				"topic1-3": "indexed参数 (最多3个，每个32字节)",
			},
			"data": "非indexed参数，ABI编码",
		},
		"opcodes": map[string]interface{}{
			"LOG0": "无topic，只有data",
			"LOG1": "1个topic (通常是event signature)",
			"LOG2": "2个topics (signature + 1 indexed)",
			"LOG3": "3个topics (signature + 2 indexed)",
			"LOG4": "4个topics (signature + 3 indexed)",
		},
		"gas_cost": map[string]interface{}{
			"base":          375,
			"per_topic":     375,
			"per_byte_data": 8,
			"example_log2":  "375 + 2*375 + 32*8 = 1381 gas",
		},
		"indexed_vs_data": map[string]interface{}{
			"indexed": []string{
				"存储在topics中",
				"可以被过滤查询",
				"限制: 最多3个, 每个32字节",
				"动态类型(string/bytes)只存储哈希",
			},
			"data": []string{
				"存储在data字段",
				"不能被过滤",
				"无数量限制",
				"完整存储原始值",
			},
		},
	}
}

// SimulateTransferEvent 模拟Transfer事件
func (s *EventLogSimulator) SimulateTransferEvent(from, to string, amount string) *DecodedLog {
	encoder := evmpkg.NewABIEncoder()

	// topic0: 事件签名
	topic0 := encoder.EventTopicHex("Transfer(address,address,uint256)")

	// topic1: indexed from
	fromAddr := evmpkg.HexToAddress(from)
	topic1 := "0x" + hex.EncodeToString(encoder.EncodeAddress(fromAddr))

	// topic2: indexed to
	toAddr := evmpkg.HexToAddress(to)
	topic2 := "0x" + hex.EncodeToString(encoder.EncodeAddress(toAddr))

	// data: non-indexed amount
	amountBig, _ := new(big.Int).SetString(amount, 10)
	data := "0x" + hex.EncodeToString(encoder.EncodeUint256(amountBig))

	log := &DecodedLog{
		Address:   "0xTokenContract",
		EventName: "Transfer",
		Signature: "Transfer(address,address,uint256)",
		Topics:    []string{topic0, topic1, topic2},
		Data:      data,
		Decoded: map[string]interface{}{
			"from":  from,
			"to":    to,
			"value": amount,
		},
	}

	s.decodedLogs = append(s.decodedLogs, log)

	s.EmitEvent("transfer_event_simulated", "", "", map[string]interface{}{
		"from":   from,
		"to":     to,
		"amount": amount,
	})

	s.updateState()
	return log
}

// DecodeLog 解码日志
func (s *EventLogSimulator) DecodeLog(topics []string, data string) *DecodedLog {
	if len(topics) == 0 {
		return nil
	}

	decoder := evmpkg.NewABIDecoder()

	// 查找事件定义
	event := s.events[topics[0]]
	if event == nil {
		return &DecodedLog{
			Topics: topics,
			Data:   data,
			Decoded: map[string]interface{}{
				"error": "未知事件签名",
			},
		}
	}

	decoded := make(map[string]interface{})

	// 解码indexed参数 (从topics)
	for i, param := range event.Indexed {
		if i+1 < len(topics) {
			topicBytes, _ := hex.DecodeString(strings.TrimPrefix(topics[i+1], "0x"))
			// 根据参数类型解码
			decoded[param] = "0x" + hex.EncodeToString(topicBytes[12:]) // 假设是地址
		}
	}

	// 解码data参数
	if data != "" && data != "0x" {
		dataBytes, _ := hex.DecodeString(strings.TrimPrefix(data, "0x"))
		offset := 0
		for _, param := range event.Data {
			if offset+32 <= len(dataBytes) {
				decoded[param] = decoder.DecodeUint256(dataBytes[offset : offset+32]).String()
				offset += 32
			}
		}
	}

	log := &DecodedLog{
		EventName: event.Name,
		Signature: event.Signature,
		Topics:    topics,
		Data:      data,
		Decoded:   decoded,
	}

	s.decodedLogs = append(s.decodedLogs, log)
	s.updateState()

	return log
}

// ShowCommonEvents 显示常见事件
func (s *EventLogSimulator) ShowCommonEvents() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	commonEvents := []struct {
		name      string
		signature string
		category  string
	}{
		{"Transfer (ERC20)", "Transfer(address,address,uint256)", "ERC20"},
		{"Approval (ERC20)", "Approval(address,address,uint256)", "ERC20"},
		{"Transfer (ERC721)", "Transfer(address,address,uint256)", "ERC721"},
		{"Approval (ERC721)", "Approval(address,address,uint256)", "ERC721"},
		{"ApprovalForAll", "ApprovalForAll(address,address,bool)", "ERC721/1155"},
		{"TransferSingle", "TransferSingle(address,address,address,uint256,uint256)", "ERC1155"},
		{"TransferBatch", "TransferBatch(address,address,address,uint256[],uint256[])", "ERC1155"},
		{"Swap", "Swap(address,uint256,uint256,uint256,uint256,address)", "Uniswap V2"},
		{"Sync", "Sync(uint112,uint112)", "Uniswap V2"},
		{"Deposit", "Deposit(address,uint256)", "WETH"},
		{"Withdrawal", "Withdrawal(address,uint256)", "WETH"},
		{"OwnershipTransferred", "OwnershipTransferred(address,address)", "Ownable"},
		{"Upgraded", "Upgraded(address)", "Proxy"},
		{"AdminChanged", "AdminChanged(address,address)", "Proxy"},
	}

	for _, e := range commonEvents {
		h := sha3.NewLegacyKeccak256()
		h.Write([]byte(e.signature))
		topic0 := "0x" + hex.EncodeToString(h.Sum(nil))

		result = append(result, map[string]interface{}{
			"name":      e.name,
			"signature": e.signature,
			"topic0":    topic0,
			"category":  e.category,
		})
	}

	return result
}

// ExplainAnonymousEvents 解释匿名事件
func (s *EventLogSimulator) ExplainAnonymousEvents() map[string]interface{} {
	return map[string]interface{}{
		"description": "匿名事件不记录事件签名到topic0",
		"syntax":      "event AnonymousEvent(uint256 indexed value) anonymous;",
		"features": []string{
			"topic0不是事件签名",
			"可以有4个indexed参数(而非3个)",
			"无法通过签名过滤",
			"gas稍低(少一个topic)",
		},
		"use_cases": []string{
			"需要4个indexed参数",
			"gas优化",
			"事件识别通过其他方式",
		},
	}
}

// updateState 更新状态
func (s *EventLogSimulator) updateState() {
	s.SetGlobalData("event_count", len(s.events)/2) // 每个事件注册两次
	s.SetGlobalData("decoded_log_count", len(s.decodedLogs))

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		func() string {
			if len(s.decodedLogs) > 0 {
				return "event_completed"
			}
			return "event_ready"
		}(),
		func() string {
			if len(s.decodedLogs) > 0 {
				return fmt.Sprintf("当前已经解析出 %d 条事件日志。", len(s.decodedLogs))
			}
			return "当前还没有生成事件日志，可以先模拟一次 Transfer 事件。"
		}(),
		"重点观察 topics、data 区域和 indexed 参数如何共同构成日志结构。",
		func() float64 {
			if len(s.decodedLogs) > 0 {
				return 0.85
			}
			return 0.2
		}(),
		map[string]interface{}{
			"event_count":       len(s.events) / 2,
			"decoded_log_count": len(s.decodedLogs),
		},
	)
}

// ExecuteAction 为事件日志实验提供交互动作。
func (s *EventLogSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_transfer_event":
		result := s.SimulateTransferEvent("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "1000")
		return evmActionResult(
			"已模拟一次 Transfer 事件。",
			map[string]interface{}{"log": result},
			&types.ActionFeedback{
				Summary:     "Transfer 事件的 topics 和 data 编码已经生成。",
				NextHint:    "继续观察 indexed 参数为什么进入 topics，以及匿名事件少了哪一部分信息。",
				EffectScope: "evm",
				ResultState: map[string]interface{}{"decoded": result != nil},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported event log action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// EventLogFactory 事件日志工厂
type EventLogFactory struct{}

func (f *EventLogFactory) Create() engine.Simulator {
	return NewEventLogSimulator()
}

func (f *EventLogFactory) GetDescription() types.Description {
	return NewEventLogSimulator().GetDescription()
}

func NewEventLogFactory() *EventLogFactory {
	return &EventLogFactory{}
}

var _ engine.SimulatorFactory = (*EventLogFactory)(nil)
