package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"golang.org/x/crypto/sha3"
)

// =============================================================================
// 存储布局演示器
// =============================================================================

// StorageSlot 存储槽
type StorageSlot struct {
	Slot     string `json:"slot"`
	Variable string `json:"variable"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Offset   int    `json:"offset"` // 槽内偏移(字节)
	Size     int    `json:"size"`   // 大小(字节)
}

// StorageLayoutSimulator 存储布局演示器
// 演示Solidity合约的存储布局:
// - 基本类型的槽分配
// - 结构体和数组的存储
// - 映射的哈希计算
// - 紧凑打包规则
type StorageLayoutSimulator struct {
	*base.BaseSimulator
	slots    map[string]*StorageSlot
	nextSlot int
}

// NewStorageLayoutSimulator 创建存储布局演示器
func NewStorageLayoutSimulator() *StorageLayoutSimulator {
	sim := &StorageLayoutSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"storage_layout",
			"存储布局演示器",
			"演示Solidity合约的存储槽分配规则",
			"evm",
			types.ComponentDemo,
		),
		slots:    make(map[string]*StorageSlot),
		nextSlot: 0,
	}

	return sim
}

// Init 初始化
func (s *StorageLayoutSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// AddVariable 添加状态变量
func (s *StorageLayoutSimulator) AddVariable(name, varType string) *StorageSlot {
	size := s.getTypeSize(varType)

	slot := &StorageSlot{
		Slot:     fmt.Sprintf("%d", s.nextSlot),
		Variable: name,
		Type:     varType,
		Offset:   0,
		Size:     size,
	}

	s.slots[name] = slot
	s.nextSlot++

	s.EmitEvent("variable_added", "", "", map[string]interface{}{
		"name": name,
		"type": varType,
		"slot": slot.Slot,
		"size": size,
	})

	s.updateState()
	return slot
}

// CalculateMappingSlot 计算映射的存储槽
// slot = keccak256(key . mappingSlot)
func (s *StorageLayoutSimulator) CalculateMappingSlot(mappingSlot int, key string) map[string]interface{} {
	// 准备key (32字节对齐)
	keyBytes := make([]byte, 32)
	if len(key) > 2 && key[:2] == "0x" {
		decoded, _ := hex.DecodeString(key[2:])
		copy(keyBytes[32-len(decoded):], decoded)
	} else {
		// 假设是地址
		decoded, _ := hex.DecodeString(key)
		copy(keyBytes[32-len(decoded):], decoded)
	}

	// 准备槽号 (32字节)
	slotBytes := make([]byte, 32)
	slotBig := big.NewInt(int64(mappingSlot))
	slotData := slotBig.Bytes()
	copy(slotBytes[32-len(slotData):], slotData)

	// 拼接: key . slot
	data := append(keyBytes, slotBytes...)

	// 计算keccak256
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	hash := h.Sum(nil)

	result := map[string]interface{}{
		"mapping_slot": mappingSlot,
		"key":          key,
		"data":         "0x" + hex.EncodeToString(data),
		"result_slot":  "0x" + hex.EncodeToString(hash),
		"formula":      "keccak256(key . mappingSlot)",
	}

	s.EmitEvent("mapping_slot_calculated", "", "", result)
	return result
}

// CalculateNestedMappingSlot 计算嵌套映射的存储槽
// slot = keccak256(key2 . keccak256(key1 . mappingSlot))
func (s *StorageLayoutSimulator) CalculateNestedMappingSlot(mappingSlot int, key1, key2 string) map[string]interface{} {
	// 先计算第一层
	first := s.CalculateMappingSlot(mappingSlot, key1)
	firstSlot := first["result_slot"].(string)

	// 解析第一层结果
	firstSlotBytes, _ := hex.DecodeString(firstSlot[2:])

	// 准备key2
	key2Bytes := make([]byte, 32)
	if len(key2) > 2 && key2[:2] == "0x" {
		decoded, _ := hex.DecodeString(key2[2:])
		copy(key2Bytes[32-len(decoded):], decoded)
	}

	// 拼接: key2 . firstSlot
	data := append(key2Bytes, firstSlotBytes...)

	// 计算keccak256
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	hash := h.Sum(nil)

	result := map[string]interface{}{
		"mapping_slot":      mappingSlot,
		"key1":              key1,
		"key2":              key2,
		"intermediate_slot": firstSlot,
		"result_slot":       "0x" + hex.EncodeToString(hash),
		"formula":           "keccak256(key2 . keccak256(key1 . mappingSlot))",
	}

	s.EmitEvent("nested_mapping_slot_calculated", "", "", result)
	return result
}

// CalculateArraySlot 计算动态数组元素的存储槽
// slot = keccak256(arraySlot) + index
func (s *StorageLayoutSimulator) CalculateArraySlot(arraySlot int, index int) map[string]interface{} {
	// 准备槽号
	slotBytes := make([]byte, 32)
	slotBig := big.NewInt(int64(arraySlot))
	slotData := slotBig.Bytes()
	copy(slotBytes[32-len(slotData):], slotData)

	// 计算keccak256
	h := sha3.NewLegacyKeccak256()
	h.Write(slotBytes)
	baseSlot := new(big.Int).SetBytes(h.Sum(nil))

	// 加上索引
	elementSlot := new(big.Int).Add(baseSlot, big.NewInt(int64(index)))

	result := map[string]interface{}{
		"array_slot":   arraySlot,
		"index":        index,
		"base_slot":    "0x" + baseSlot.Text(16),
		"element_slot": "0x" + elementSlot.Text(16),
		"formula":      "keccak256(arraySlot) + index",
		"note":         "数组长度存储在arraySlot本身",
	}

	s.EmitEvent("array_slot_calculated", "", "", result)
	return result
}

// DemonstratePacking 演示紧凑打包
func (s *StorageLayoutSimulator) DemonstratePacking() []map[string]interface{} {
	examples := []map[string]interface{}{
		{
			"variables": []string{"uint256 a", "uint256 b"},
			"slots":     2,
			"reason":    "uint256占满整个槽(32字节)",
		},
		{
			"variables": []string{"uint128 a", "uint128 b"},
			"slots":     1,
			"reason":    "两个uint128可以打包到一个槽(各16字节)",
		},
		{
			"variables": []string{"uint8 a", "uint8 b", "uint8 c", "address d"},
			"slots":     1,
			"reason":    "3个uint8(3字节) + address(20字节) = 23字节 < 32字节",
		},
		{
			"variables": []string{"uint256 a", "uint8 b"},
			"slots":     2,
			"reason":    "uint256独占一个槽，uint8开启新槽",
		},
		{
			"variables": []string{"bool a", "bool b", "bool c", "uint240 d"},
			"slots":     1,
			"reason":    "3个bool(3字节) + uint240(30字节) = 33字节 > 32字节，所以需要2个槽",
		},
	}

	s.EmitEvent("packing_demonstrated", "", "", map[string]interface{}{
		"examples": len(examples),
	})

	return examples
}

// getTypeSize 获取类型大小(字节)
func (s *StorageLayoutSimulator) getTypeSize(varType string) int {
	switch varType {
	case "uint8", "int8", "bool":
		return 1
	case "uint16", "int16":
		return 2
	case "uint32", "int32":
		return 4
	case "uint64", "int64":
		return 8
	case "uint128", "int128":
		return 16
	case "address":
		return 20
	case "uint256", "int256", "bytes32":
		return 32
	default:
		return 32
	}
}

// updateState 更新状态
func (s *StorageLayoutSimulator) updateState() {
	slotList := make([]map[string]interface{}, 0)
	for _, slot := range s.slots {
		slotList = append(slotList, map[string]interface{}{
			"slot":     slot.Slot,
			"variable": slot.Variable,
			"type":     slot.Type,
		})
	}
	s.SetGlobalData("slots", slotList)
	s.SetGlobalData("next_slot", s.nextSlot)

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		func() string {
			if len(s.slots) > 0 {
				return "layout_completed"
			}
			return "layout_ready"
		}(),
		func() string {
			if len(s.slots) > 0 {
				return fmt.Sprintf("当前已经计算出 %d 个存储槽布局。", len(s.slots))
			}
			return "当前还没有计算存储布局，可以先添加变量并观察 slot 分配。"
		}(),
		"继续观察映射、数组和打包规则如何影响 slot 计算结果。",
		func() float64 {
			if len(s.slots) > 0 {
				return 0.85
			}
			return 0.2
		}(),
		map[string]interface{}{"slot_count": len(s.slots), "next_slot": s.nextSlot},
	)
}

// ExecuteAction 为存储布局实验提供交互动作。
func (s *StorageLayoutSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "add_variable":
		name := "value"
		varType := "uint256"
		if raw, ok := params["name"].(string); ok && raw != "" {
			name = raw
		}
		if raw, ok := params["type"].(string); ok && raw != "" {
			varType = raw
		}
		slot := s.AddVariable(name, varType)
		return evmActionResult("已添加一个存储变量。", map[string]interface{}{"slot": slot}, &types.ActionFeedback{
			Summary:     "变量已经加入当前布局，新的 slot 分配结果可直接用于前端可视化。",
			NextHint:    "继续添加不同尺寸的变量，观察打包规则是否让多个值共享同一个 slot。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"slot_index": slot.Slot},
		}), nil
	case "calculate_mapping_slot":
		result := s.CalculateMappingSlot(0, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		return evmActionResult("已计算映射槽位置。", result, &types.ActionFeedback{
			Summary:     "映射键对应的 keccak 槽位已经生成。",
			NextHint:    "继续比较嵌套映射和数组槽位，理解复杂结构的寻址方式。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported storage layout action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// StorageLayoutFactory 存储布局工厂
type StorageLayoutFactory struct{}

func (f *StorageLayoutFactory) Create() engine.Simulator {
	return NewStorageLayoutSimulator()
}

func (f *StorageLayoutFactory) GetDescription() types.Description {
	return NewStorageLayoutSimulator().GetDescription()
}

func NewStorageLayoutFactory() *StorageLayoutFactory {
	return &StorageLayoutFactory{}
}

var _ engine.SimulatorFactory = (*StorageLayoutFactory)(nil)
