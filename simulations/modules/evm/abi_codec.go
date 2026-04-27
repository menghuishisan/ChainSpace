package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	evmpkg "github.com/chainspace/simulations/pkg/evm"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// ABI编解码演示器
// =============================================================================

// ABICodecSimulator ABI编解码演示器
// 演示以太坊ABI编解码:
// - 函数选择器计算
// - 参数编码规则
// - 动态类型处理
// - 事件Topic计算
type ABICodecSimulator struct {
	*base.BaseSimulator
	encoder *evmpkg.ABIEncoder
	decoder *evmpkg.ABIDecoder
	history []map[string]interface{}
}

// NewABICodecSimulator 创建ABI编解码演示器
func NewABICodecSimulator() *ABICodecSimulator {
	sim := &ABICodecSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"abi_codec",
			"ABI编解码演示器",
			"演示以太坊ABI编解码规则，包括函数选择器、参数编码等",
			"evm",
			types.ComponentTool,
		),
		encoder: evmpkg.NewABIEncoder(),
		decoder: evmpkg.NewABIDecoder(),
		history: make([]map[string]interface{}, 0),
	}

	return sim
}

// Init 初始化
func (s *ABICodecSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// CalculateSelector 计算函数选择器
func (s *ABICodecSimulator) CalculateSelector(signature string) map[string]interface{} {
	selector := s.encoder.FunctionSelectorHex(signature)

	result := map[string]interface{}{
		"signature": signature,
		"selector":  selector,
		"bytes":     len(selector)/2 - 1, // 去掉0x
	}

	s.history = append(s.history, map[string]interface{}{
		"type":   "selector",
		"input":  signature,
		"output": selector,
	})

	s.EmitEvent("selector_calculated", "", "", result)
	s.updateState()
	return result
}

// CalculateEventTopic 计算事件Topic
func (s *ABICodecSimulator) CalculateEventTopic(signature string) map[string]interface{} {
	topic := s.encoder.EventTopicHex(signature)

	result := map[string]interface{}{
		"signature": signature,
		"topic":     topic,
	}

	s.history = append(s.history, map[string]interface{}{
		"type":   "event_topic",
		"input":  signature,
		"output": topic,
	})

	s.EmitEvent("topic_calculated", "", "", result)
	s.updateState()
	return result
}

// EncodeUint256 编码uint256
func (s *ABICodecSimulator) EncodeUint256(value string) map[string]interface{} {
	val := new(big.Int)
	val.SetString(value, 10)

	encoded := s.encoder.EncodeUint256(val)

	result := map[string]interface{}{
		"type":    "uint256",
		"value":   value,
		"encoded": "0x" + hex.EncodeToString(encoded),
		"length":  len(encoded),
	}

	s.EmitEvent("value_encoded", "", "", result)
	s.updateState()
	return result
}

// EncodeAddress 编码address
func (s *ABICodecSimulator) EncodeAddress(address string) map[string]interface{} {
	addr := evmpkg.HexToAddress(address)
	encoded := s.encoder.EncodeAddress(addr)

	result := map[string]interface{}{
		"type":    "address",
		"value":   address,
		"encoded": "0x" + hex.EncodeToString(encoded),
		"length":  len(encoded),
	}

	s.EmitEvent("value_encoded", "", "", result)
	s.updateState()
	return result
}

// EncodeString 编码string (动态类型)
func (s *ABICodecSimulator) EncodeString(value string) map[string]interface{} {
	encoded := s.encoder.EncodeString(value)

	result := map[string]interface{}{
		"type":        "string",
		"value":       value,
		"encoded":     "0x" + hex.EncodeToString(encoded),
		"length":      len(encoded),
		"is_dynamic":  true,
		"explanation": "动态类型: 前32字节是长度，后面是数据(32字节对齐)",
	}

	s.EmitEvent("value_encoded", "", "", result)
	s.updateState()
	return result
}

// EncodeCall 编码函数调用
func (s *ABICodecSimulator) EncodeCall(signature string, args []interface{}) (map[string]interface{}, error) {
	// 转换参数
	convertedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case float64:
			convertedArgs[i] = big.NewInt(int64(v))
		case string:
			if len(v) == 42 && v[:2] == "0x" {
				convertedArgs[i] = v // address
			} else {
				// 尝试解析为数字
				val := new(big.Int)
				if _, ok := val.SetString(v, 10); ok {
					convertedArgs[i] = val
				} else {
					convertedArgs[i] = v // string
				}
			}
		default:
			convertedArgs[i] = v
		}
	}

	encoded, err := s.encoder.EncodeCall(signature, convertedArgs...)
	if err != nil {
		return nil, err
	}

	selector := s.encoder.FunctionSelectorHex(signature)

	result := map[string]interface{}{
		"signature": signature,
		"selector":  selector,
		"args":      args,
		"calldata":  "0x" + hex.EncodeToString(encoded),
		"length":    len(encoded),
	}

	s.history = append(s.history, map[string]interface{}{
		"type":   "call_encode",
		"input":  signature,
		"output": result["calldata"],
	})

	s.EmitEvent("call_encoded", "", "", result)
	s.updateState()
	return result, nil
}

// DecodeReturnData 解码返回数据
func (s *ABICodecSimulator) DecodeReturnData(dataHex string, returnType string) map[string]interface{} {
	if len(dataHex) > 2 && dataHex[:2] == "0x" {
		dataHex = dataHex[2:]
	}
	data, _ := hex.DecodeString(dataHex)

	var decoded interface{}

	switch returnType {
	case "uint256":
		decoded = s.decoder.DecodeUint256(data).String()
	case "address":
		decoded = evmpkg.AddressToHex(s.decoder.DecodeAddress(data))
	case "bool":
		decoded = s.decoder.DecodeBool(data)
	case "bytes32":
		decoded = "0x" + hex.EncodeToString(s.decoder.DecodeBytes32(data))
	case "string":
		decoded = s.decoder.DecodeString(data, 0)
	default:
		decoded = "0x" + hex.EncodeToString(data)
	}

	result := map[string]interface{}{
		"raw_data":    "0x" + dataHex,
		"return_type": returnType,
		"decoded":     decoded,
	}

	s.EmitEvent("data_decoded", "", "", result)
	s.updateState()
	return result
}

// ShowCommonSelectors 显示常见函数选择器
func (s *ABICodecSimulator) ShowCommonSelectors() []map[string]string {
	signatures := []string{
		"transfer(address,uint256)",
		"approve(address,uint256)",
		"transferFrom(address,address,uint256)",
		"balanceOf(address)",
		"allowance(address,address)",
		"totalSupply()",
		"name()",
		"symbol()",
		"decimals()",
		"ownerOf(uint256)",
		"safeTransferFrom(address,address,uint256)",
		"setApprovalForAll(address,bool)",
		"isApprovedForAll(address,address)",
		"mint(address,uint256)",
		"burn(uint256)",
	}

	result := make([]map[string]string, len(signatures))
	for i, sig := range signatures {
		result[i] = map[string]string{
			"signature": sig,
			"selector":  s.encoder.FunctionSelectorHex(sig),
		}
	}

	return result
}

// updateState 更新状态
func (s *ABICodecSimulator) updateState() {
	s.SetGlobalData("history_count", len(s.history))

	if len(s.history) > 5 {
		s.SetGlobalData("recent_history", s.history[len(s.history)-5:])
	} else {
		s.SetGlobalData("recent_history", s.history)
	}

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		func() string {
			if len(s.history) > 0 {
				return "abi_completed"
			}
			return "abi_ready"
		}(),
		func() string {
			if len(s.history) > 0 {
				return fmt.Sprintf("最近一次 ABI 编解码实验已记录 %d 条结果。", len(s.history))
			}
			return "当前还没有进行 ABI 编解码，可以先计算函数选择器或编码一次调用参数。"
		}(),
		"重点观察函数签名如何变成 4 字节选择器，以及参数是如何按 32 字节槽编码的。",
		func() float64 {
			if len(s.history) > 0 {
				return 0.85
			}
			return 0.2
		}(),
		map[string]interface{}{"history_count": len(s.history)},
	)
}

// ExecuteAction 为 ABI 编解码实验提供交互动作。
func (s *ABICodecSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "calculate_selector":
		signature := "transfer(address,uint256)"
		if raw, ok := params["signature"].(string); ok && raw != "" {
			signature = raw
		}
		result := s.CalculateSelector(signature)
		return evmActionResult("已计算函数选择器。", result, &types.ActionFeedback{
			Summary:     "函数签名对应的 4 字节选择器已经生成。",
			NextHint:    "继续对比事件 Topic 和完整 calldata 编码，理解 ABI 编码规则的差异。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	case "encode_call":
		signature := "transfer(address,uint256)"
		result, err := s.EncodeCall(signature, []interface{}{"0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "1000"})
		if err != nil {
			return nil, err
		}
		return evmActionResult("已编码一次合约调用。", result, &types.ActionFeedback{
			Summary:     "一次完整的 ABI 调用编码已经生成。",
			NextHint:    "重点观察选择器、静态参数和动态参数在 calldata 中的位置布局。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported abi codec action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// ABICodecFactory ABI编解码工厂
type ABICodecFactory struct{}

func (f *ABICodecFactory) Create() engine.Simulator {
	return NewABICodecSimulator()
}

func (f *ABICodecFactory) GetDescription() types.Description {
	return NewABICodecSimulator().GetDescription()
}

func NewABICodecFactory() *ABICodecFactory {
	return &ABICodecFactory{}
}

var _ engine.SimulatorFactory = (*ABICodecFactory)(nil)
