package evm

import (
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	evmpkg "github.com/chainspace/simulations/pkg/evm"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 反汇编器
// =============================================================================

// Instruction 指令
type Instruction struct {
	PC        uint64 `json:"pc"`
	OpCode    byte   `json:"opcode"`
	OpName    string `json:"op_name"`
	Immediate string `json:"immediate,omitempty"`
	Gas       uint64 `json:"gas"`
	StackIn   int    `json:"stack_in"`
	StackOut  int    `json:"stack_out"`
	Raw       string `json:"raw"`
}

// DisassemblerSimulator 反汇编器
// 将EVM字节码反汇编为可读的操作码序列
type DisassemblerSimulator struct {
	*base.BaseSimulator
	instructions []*Instruction
	jumpDests    map[uint64]bool
}

// NewDisassemblerSimulator 创建反汇编器
func NewDisassemblerSimulator() *DisassemblerSimulator {
	sim := &DisassemblerSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"disassembler",
			"反汇编器",
			"将EVM字节码反汇编为可读的操作码序列",
			"evm",
			types.ComponentTool,
		),
		instructions: make([]*Instruction, 0),
		jumpDests:    make(map[uint64]bool),
	}

	return sim
}

// Init 初始化
func (s *DisassemblerSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// Disassemble 反汇编字节码
func (s *DisassemblerSimulator) Disassemble(bytecodeHex string) ([]*Instruction, error) {
	if len(bytecodeHex) > 2 && bytecodeHex[:2] == "0x" {
		bytecodeHex = bytecodeHex[2:]
	}

	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return nil, fmt.Errorf("无效的字节码: %v", err)
	}

	s.instructions = make([]*Instruction, 0)
	s.jumpDests = make(map[uint64]bool)

	// 第一遍: 找到所有JUMPDEST
	pc := uint64(0)
	for pc < uint64(len(bytecode)) {
		op := evmpkg.OpCode(bytecode[pc])
		if op == evmpkg.JUMPDEST {
			s.jumpDests[pc] = true
		}
		info, ok := evmpkg.GetOpCodeInfo(op)
		if ok && info.ImmSize > 0 {
			pc += uint64(info.ImmSize)
		}
		pc++
	}

	// 第二遍: 反汇编
	pc = 0
	for pc < uint64(len(bytecode)) {
		op := evmpkg.OpCode(bytecode[pc])
		info, ok := evmpkg.GetOpCodeInfo(op)

		inst := &Instruction{
			PC:     pc,
			OpCode: byte(op),
			OpName: evmpkg.GetOpCodeName(op),
			Raw:    fmt.Sprintf("%02x", op),
		}

		if ok {
			inst.Gas = info.Gas
			inst.StackIn = info.StackPop
			inst.StackOut = info.StackPush

			// 处理立即数
			if info.ImmSize > 0 {
				immStart := pc + 1
				immEnd := immStart + uint64(info.ImmSize)
				if immEnd > uint64(len(bytecode)) {
					immEnd = uint64(len(bytecode))
				}
				imm := bytecode[immStart:immEnd]
				inst.Immediate = "0x" + hex.EncodeToString(imm)
				inst.Raw = hex.EncodeToString(bytecode[pc:immEnd])
				pc = immEnd
			} else {
				pc++
			}
		} else {
			inst.OpName = fmt.Sprintf("INVALID(0x%02x)", op)
			pc++
		}

		s.instructions = append(s.instructions, inst)
	}

	s.EmitEvent("disassembled", "", "", map[string]interface{}{
		"instruction_count": len(s.instructions),
		"jumpdest_count":    len(s.jumpDests),
		"bytecode_size":     len(bytecode),
	})

	s.updateState()
	return s.instructions, nil
}

// DisassembleWithAnnotations 带注释的反汇编
func (s *DisassemblerSimulator) DisassembleWithAnnotations(bytecodeHex string) []map[string]interface{} {
	instructions, err := s.Disassemble(bytecodeHex)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, len(instructions))
	for i, inst := range instructions {
		annotation := s.getAnnotation(inst)
		result[i] = map[string]interface{}{
			"pc":          fmt.Sprintf("0x%04x", inst.PC),
			"opcode":      inst.OpName,
			"immediate":   inst.Immediate,
			"gas":         inst.Gas,
			"annotation":  annotation,
			"is_jumpdest": s.jumpDests[inst.PC],
		}
	}

	return result
}

// getAnnotation 获取指令注释
func (s *DisassemblerSimulator) getAnnotation(inst *Instruction) string {
	op := evmpkg.OpCode(inst.OpCode)

	switch {
	case op == evmpkg.STOP:
		return "停止执行"
	case op >= evmpkg.PUSH1 && op <= evmpkg.PUSH32:
		return fmt.Sprintf("压入%d字节常量到栈", op-evmpkg.PUSH1+1)
	case op >= evmpkg.DUP1 && op <= evmpkg.DUP16:
		return fmt.Sprintf("复制栈中第%d个元素", op-evmpkg.DUP1+1)
	case op >= evmpkg.SWAP1 && op <= evmpkg.SWAP16:
		return fmt.Sprintf("交换栈顶和第%d个元素", op-evmpkg.SWAP1+2)
	case op >= evmpkg.LOG0 && op <= evmpkg.LOG4:
		return fmt.Sprintf("发出带%d个topic的事件", op-evmpkg.LOG0)
	case op == evmpkg.JUMPDEST:
		return "跳转目标标记"
	case op == evmpkg.JUMP:
		return "无条件跳转"
	case op == evmpkg.JUMPI:
		return "条件跳转"
	case op == evmpkg.SLOAD:
		return "读取存储槽"
	case op == evmpkg.SSTORE:
		return "写入存储槽"
	case op == evmpkg.MLOAD:
		return "读取内存"
	case op == evmpkg.MSTORE:
		return "写入内存(32字节)"
	case op == evmpkg.CALL:
		return "调用外部合约"
	case op == evmpkg.DELEGATECALL:
		return "委托调用(保持msg.sender)"
	case op == evmpkg.STATICCALL:
		return "静态调用(只读)"
	case op == evmpkg.CREATE:
		return "创建合约"
	case op == evmpkg.CREATE2:
		return "使用salt创建合约(确定性地址)"
	case op == evmpkg.RETURN:
		return "返回数据并结束"
	case op == evmpkg.REVERT:
		return "回滚并返回错误"
	case op == evmpkg.SELFDESTRUCT:
		return "销毁合约"
	case op == evmpkg.KECCAK256:
		return "计算Keccak256哈希"
	case op == evmpkg.CALLER:
		return "获取msg.sender"
	case op == evmpkg.CALLVALUE:
		return "获取msg.value"
	case op == evmpkg.CALLDATALOAD:
		return "读取calldata"
	default:
		return ""
	}
}

// FindFunctionSelectors 查找函数选择器
func (s *DisassemblerSimulator) FindFunctionSelectors() []map[string]interface{} {
	selectors := make([]map[string]interface{}, 0)

	for i := 0; i < len(s.instructions)-1; i++ {
		inst := s.instructions[i]
		// 查找PUSH4后面跟着EQ的模式
		if inst.OpName == "PUSH4" && inst.Immediate != "" {
			// 这可能是一个函数选择器
			selector := inst.Immediate
			selectors = append(selectors, map[string]interface{}{
				"pc":       fmt.Sprintf("0x%04x", inst.PC),
				"selector": selector,
			})
		}
	}

	s.EmitEvent("selectors_found", "", "", map[string]interface{}{
		"count": len(selectors),
	})

	return selectors
}

// GetControlFlowGraph 获取控制流图
func (s *DisassemblerSimulator) GetControlFlowGraph() map[string]interface{} {
	blocks := make([]map[string]interface{}, 0)
	edges := make([]map[string]interface{}, 0)

	// 找到基本块边界
	blockStarts := make(map[uint64]bool)
	blockStarts[0] = true

	for _, inst := range s.instructions {
		op := evmpkg.OpCode(inst.OpCode)
		if op == evmpkg.JUMPDEST {
			blockStarts[inst.PC] = true
		}
		if op == evmpkg.JUMP || op == evmpkg.JUMPI {
			// 下一条指令是新块的开始
			if inst.PC+1 < uint64(len(s.instructions)) {
				blockStarts[inst.PC+1] = true
			}
		}
	}

	// 构建基本块
	var currentBlock []uint64
	for _, inst := range s.instructions {
		if blockStarts[inst.PC] && len(currentBlock) > 0 {
			blocks = append(blocks, map[string]interface{}{
				"start": fmt.Sprintf("0x%04x", currentBlock[0]),
				"end":   fmt.Sprintf("0x%04x", currentBlock[len(currentBlock)-1]),
				"size":  len(currentBlock),
			})
			currentBlock = nil
		}
		currentBlock = append(currentBlock, inst.PC)
	}
	if len(currentBlock) > 0 {
		blocks = append(blocks, map[string]interface{}{
			"start": fmt.Sprintf("0x%04x", currentBlock[0]),
			"end":   fmt.Sprintf("0x%04x", currentBlock[len(currentBlock)-1]),
			"size":  len(currentBlock),
		})
	}

	return map[string]interface{}{
		"blocks":      blocks,
		"edges":       edges,
		"block_count": len(blocks),
	}
}

// updateState 更新状态
func (s *DisassemblerSimulator) updateState() {
	s.SetGlobalData("instruction_count", len(s.instructions))
	s.SetGlobalData("jumpdest_count", len(s.jumpDests))

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		func() string {
			if len(s.instructions) > 0 {
				return "disassembly_completed"
			}
			return "disassembly_ready"
		}(),
		func() string {
			if len(s.instructions) > 0 {
				return fmt.Sprintf("当前已经反汇编出 %d 条指令。", len(s.instructions))
			}
			return "当前还没有反汇编任何字节码，可以先加载一个简单示例。"
		}(),
		"重点观察操作码顺序、立即数和控制流跳转，理解字节码是如何映射到执行逻辑的。",
		func() float64 {
			if len(s.instructions) > 0 {
				return 0.85
			}
			return 0.2
		}(),
		map[string]interface{}{"instruction_count": len(s.instructions), "jumpdest_count": len(s.jumpDests)},
	)
}

// ExecuteAction 为反汇编实验提供交互动作。
func (s *DisassemblerSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "disassemble_sample":
		bytecode := "6001600201"
		if raw, ok := params["bytecode"].(string); ok && raw != "" {
			bytecode = raw
		}
		result, err := s.Disassemble(bytecode)
		if err != nil {
			return nil, err
		}
		return evmActionResult("已完成一次反汇编。", map[string]interface{}{"instructions": result}, &types.ActionFeedback{
			Summary:     "字节码已经被拆解成逐条 opcode。",
			NextHint:    "继续查看带注释的反汇编结果和控制流图，理解每条指令在做什么。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"instruction_count": len(result)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported disassembler action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// DisassemblerFactory 反汇编器工厂
type DisassemblerFactory struct{}

func (f *DisassemblerFactory) Create() engine.Simulator {
	return NewDisassemblerSimulator()
}

func (f *DisassemblerFactory) GetDescription() types.Description {
	return NewDisassemblerSimulator().GetDescription()
}

func NewDisassemblerFactory() *DisassemblerFactory {
	return &DisassemblerFactory{}
}

var _ engine.SimulatorFactory = (*DisassemblerFactory)(nil)
