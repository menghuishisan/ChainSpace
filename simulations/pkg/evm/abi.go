package evm

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
)

// =============================================================================
// ABI编解码
// =============================================================================

// ABIType ABI类型
type ABIType string

const (
	ABITypeUint256    ABIType = "uint256"
	ABITypeUint128    ABIType = "uint128"
	ABITypeUint64     ABIType = "uint64"
	ABITypeUint32     ABIType = "uint32"
	ABITypeUint8      ABIType = "uint8"
	ABITypeInt256     ABIType = "int256"
	ABITypeAddress    ABIType = "address"
	ABITypeBool       ABIType = "bool"
	ABITypeBytes32    ABIType = "bytes32"
	ABITypeBytes      ABIType = "bytes"
	ABITypeString     ABIType = "string"
	ABITypeUint256Arr ABIType = "uint256[]"
	ABITypeAddressArr ABIType = "address[]"
)

// ABIParam ABI参数
type ABIParam struct {
	Name string  `json:"name"`
	Type ABIType `json:"type"`
}

// ABIFunction ABI函数
type ABIFunction struct {
	Name    string     `json:"name"`
	Inputs  []ABIParam `json:"inputs"`
	Outputs []ABIParam `json:"outputs"`
}

// ABIEvent ABI事件
type ABIEvent struct {
	Name      string     `json:"name"`
	Inputs    []ABIParam `json:"inputs"`
	Anonymous bool       `json:"anonymous"`
}

// ABIEncoder ABI编码器
type ABIEncoder struct{}

// NewABIEncoder 创建ABI编码器
func NewABIEncoder() *ABIEncoder {
	return &ABIEncoder{}
}

// FunctionSelector 计算函数选择器 (前4字节)
func (e *ABIEncoder) FunctionSelector(signature string) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(signature))
	return h.Sum(nil)[:4]
}

// FunctionSelectorHex 计算函数选择器(十六进制)
func (e *ABIEncoder) FunctionSelectorHex(signature string) string {
	return "0x" + hex.EncodeToString(e.FunctionSelector(signature))
}

// EventTopic 计算事件Topic
func (e *ABIEncoder) EventTopic(signature string) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(signature))
	return h.Sum(nil)
}

// EventTopicHex 计算事件Topic(十六进制)
func (e *ABIEncoder) EventTopicHex(signature string) string {
	return "0x" + hex.EncodeToString(e.EventTopic(signature))
}

// EncodeUint256 编码uint256
func (e *ABIEncoder) EncodeUint256(val *big.Int) []byte {
	result := make([]byte, 32)
	if val != nil {
		bytes := val.Bytes()
		copy(result[32-len(bytes):], bytes)
	}
	return result
}

// EncodeInt256 编码int256
func (e *ABIEncoder) EncodeInt256(val *big.Int) []byte {
	result := make([]byte, 32)
	if val == nil {
		return result
	}
	if val.Sign() >= 0 {
		bytes := val.Bytes()
		copy(result[32-len(bytes):], bytes)
	} else {
		// 二进制补码
		absBytes := new(big.Int).Abs(val).Bytes()
		for i := range result {
			result[i] = 0xFF
		}
		for i, b := range absBytes {
			result[32-len(absBytes)+i] = ^b
		}
		// 加1
		carry := byte(1)
		for i := 31; i >= 0 && carry > 0; i-- {
			sum := result[i] + carry
			result[i] = sum
			if sum >= carry {
				carry = 0
			}
		}
	}
	return result
}

// EncodeAddress 编码address
func (e *ABIEncoder) EncodeAddress(addr Address) []byte {
	result := make([]byte, 32)
	copy(result[12:], addr[:])
	return result
}

// EncodeBool 编码bool
func (e *ABIEncoder) EncodeBool(val bool) []byte {
	result := make([]byte, 32)
	if val {
		result[31] = 1
	}
	return result
}

// EncodeBytes32 编码bytes32
func (e *ABIEncoder) EncodeBytes32(data []byte) []byte {
	result := make([]byte, 32)
	if len(data) > 32 {
		data = data[:32]
	}
	copy(result, data)
	return result
}

// EncodeBytes 编码动态bytes
func (e *ABIEncoder) EncodeBytes(data []byte) []byte {
	// 长度
	length := e.EncodeUint256(big.NewInt(int64(len(data))))
	// 数据 (32字节对齐)
	paddedLen := ((len(data) + 31) / 32) * 32
	padded := make([]byte, paddedLen)
	copy(padded, data)
	return append(length, padded...)
}

// EncodeString 编码string
func (e *ABIEncoder) EncodeString(s string) []byte {
	return e.EncodeBytes([]byte(s))
}

// EncodeUint256Array 编码uint256[]
func (e *ABIEncoder) EncodeUint256Array(vals []*big.Int) []byte {
	// 长度
	result := e.EncodeUint256(big.NewInt(int64(len(vals))))
	// 元素
	for _, val := range vals {
		result = append(result, e.EncodeUint256(val)...)
	}
	return result
}

// EncodeCall 编码函数调用
func (e *ABIEncoder) EncodeCall(signature string, args ...interface{}) ([]byte, error) {
	// 函数选择器
	selector := e.FunctionSelector(signature)

	// 解析参数类型
	types := parseSignatureTypes(signature)
	if len(types) != len(args) {
		return nil, fmt.Errorf("参数数量不匹配: 期望 %d, 实际 %d", len(types), len(args))
	}

	// 编码参数
	headSize := len(types) * 32
	heads := make([]byte, 0, headSize)
	tails := make([]byte, 0)

	for i, arg := range args {
		abiType := ABIType(types[i])
		encoded, isDynamic, err := e.encodeArg(abiType, arg)
		if err != nil {
			return nil, fmt.Errorf("编码参数 %d 失败: %v", i, err)
		}

		if isDynamic {
			// 动态类型: 头部存偏移量
			offset := headSize + len(tails)
			heads = append(heads, e.EncodeUint256(big.NewInt(int64(offset)))...)
			tails = append(tails, encoded...)
		} else {
			// 静态类型: 直接存值
			heads = append(heads, encoded...)
		}
	}

	return append(selector, append(heads, tails...)...), nil
}

// encodeArg 编码单个参数
func (e *ABIEncoder) encodeArg(abiType ABIType, arg interface{}) ([]byte, bool, error) {
	switch abiType {
	case ABITypeUint256, ABITypeUint128, ABITypeUint64, ABITypeUint32, ABITypeUint8:
		val, err := toBigInt(arg)
		if err != nil {
			return nil, false, err
		}
		return e.EncodeUint256(val), false, nil

	case ABITypeInt256:
		val, err := toBigInt(arg)
		if err != nil {
			return nil, false, err
		}
		return e.EncodeInt256(val), false, nil

	case ABITypeAddress:
		addr, err := toAddress(arg)
		if err != nil {
			return nil, false, err
		}
		return e.EncodeAddress(addr), false, nil

	case ABITypeBool:
		val, ok := arg.(bool)
		if !ok {
			return nil, false, fmt.Errorf("期望bool类型")
		}
		return e.EncodeBool(val), false, nil

	case ABITypeBytes32:
		data, err := toBytes(arg)
		if err != nil {
			return nil, false, err
		}
		return e.EncodeBytes32(data), false, nil

	case ABITypeBytes:
		data, err := toBytes(arg)
		if err != nil {
			return nil, false, err
		}
		return e.EncodeBytes(data), true, nil

	case ABITypeString:
		s, ok := arg.(string)
		if !ok {
			return nil, false, fmt.Errorf("期望string类型")
		}
		return e.EncodeString(s), true, nil

	default:
		return nil, false, fmt.Errorf("不支持的类型: %s", abiType)
	}
}

// ABIDecoder ABI解码器
type ABIDecoder struct{}

// NewABIDecoder 创建ABI解码器
func NewABIDecoder() *ABIDecoder {
	return &ABIDecoder{}
}

// DecodeUint256 解码uint256
func (d *ABIDecoder) DecodeUint256(data []byte) *big.Int {
	if len(data) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(data):], data)
		data = padded
	}
	return new(big.Int).SetBytes(data[:32])
}

// DecodeInt256 解码int256
func (d *ABIDecoder) DecodeInt256(data []byte) *big.Int {
	if len(data) < 32 {
		return big.NewInt(0)
	}
	val := new(big.Int).SetBytes(data[:32])
	// 检查符号位
	if data[0]&0x80 != 0 {
		// 负数
		max := new(big.Int).Lsh(big.NewInt(1), 256)
		val.Sub(val, max)
	}
	return val
}

// DecodeAddress 解码address
func (d *ABIDecoder) DecodeAddress(data []byte) Address {
	var addr Address
	if len(data) >= 32 {
		copy(addr[:], data[12:32])
	}
	return addr
}

// DecodeBool 解码bool
func (d *ABIDecoder) DecodeBool(data []byte) bool {
	if len(data) < 32 {
		return false
	}
	return data[31] != 0
}

// DecodeBytes32 解码bytes32
func (d *ABIDecoder) DecodeBytes32(data []byte) []byte {
	result := make([]byte, 32)
	if len(data) >= 32 {
		copy(result, data[:32])
	}
	return result
}

// DecodeBytes 解码动态bytes
func (d *ABIDecoder) DecodeBytes(data []byte, offset int) []byte {
	if len(data) < offset+32 {
		return nil
	}
	length := d.DecodeUint256(data[offset : offset+32]).Uint64()
	start := offset + 32
	if uint64(len(data)) < uint64(start)+length {
		return nil
	}
	result := make([]byte, length)
	copy(result, data[start:uint64(start)+length])
	return result
}

// DecodeString 解码string
func (d *ABIDecoder) DecodeString(data []byte, offset int) string {
	return string(d.DecodeBytes(data, offset))
}

// =============================================================================
// 辅助函数
// =============================================================================

// parseSignatureTypes 解析函数签名中的类型
func parseSignatureTypes(signature string) []string {
	// 找到括号
	start := strings.Index(signature, "(")
	end := strings.LastIndex(signature, ")")
	if start < 0 || end < 0 || end <= start {
		return nil
	}

	typesStr := signature[start+1 : end]
	if typesStr == "" {
		return nil
	}

	return strings.Split(typesStr, ",")
}

// toBigInt 转换为big.Int
func toBigInt(arg interface{}) (*big.Int, error) {
	switch v := arg.(type) {
	case *big.Int:
		return v, nil
	case int:
		return big.NewInt(int64(v)), nil
	case int64:
		return big.NewInt(v), nil
	case uint64:
		return new(big.Int).SetUint64(v), nil
	case string:
		val := new(big.Int)
		if strings.HasPrefix(v, "0x") {
			val.SetString(v[2:], 16)
		} else {
			val.SetString(v, 10)
		}
		return val, nil
	default:
		return nil, fmt.Errorf("无法转换为big.Int: %T", arg)
	}
}

// toAddress 转换为Address
func toAddress(arg interface{}) (Address, error) {
	switch v := arg.(type) {
	case Address:
		return v, nil
	case string:
		return HexToAddress(v), nil
	case []byte:
		var addr Address
		if len(v) > 20 {
			v = v[len(v)-20:]
		}
		copy(addr[20-len(v):], v)
		return addr, nil
	default:
		return Address{}, fmt.Errorf("无法转换为Address: %T", arg)
	}
}

// toBytes 转换为[]byte
func toBytes(arg interface{}) ([]byte, error) {
	switch v := arg.(type) {
	case []byte:
		return v, nil
	case string:
		if strings.HasPrefix(v, "0x") {
			return hex.DecodeString(v[2:])
		}
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("无法转换为[]byte: %T", arg)
	}
}

// Uint64ToBytes 转换uint64为字节
func Uint64ToBytes(val uint64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, val)
	return result
}
