package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	evmpkg "github.com/chainspace/simulations/pkg/evm"
	"github.com/chainspace/simulations/pkg/types"
)

type ExecutionStep struct {
	PC      uint64            `json:"pc"`
	OpCode  string            `json:"opcode"`
	Gas     uint64            `json:"gas"`
	GasCost uint64            `json:"gas_cost"`
	Stack   []string          `json:"stack"`
	Memory  string            `json:"memory"`
	Storage map[string]string `json:"storage"`
}

type EVMExecutorSimulator struct {
	*base.BaseSimulator
	state    *evmpkg.StateDB
	executor *evmpkg.Executor
	steps    []*ExecutionStep
}

func NewEVMExecutorSimulator() *EVMExecutorSimulator {
	sim := &EVMExecutorSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"evm_executor",
			"EVM 执行器演示器",
			"逐步执行 EVM 字节码，并可视化栈、内存、存储和 Gas 变化。",
			"evm",
			types.ComponentProcess,
		),
		steps: make([]*ExecutionStep, 0),
	}

	sim.AddParam(types.Param{
		Key:         "gas_limit",
		Name:        "Gas 限制",
		Description: "执行时允许消耗的 Gas 上限",
		Type:        types.ParamTypeInt,
		Default:     100000,
		Min:         1000,
		Max:         10000000,
	})

	return sim
}

func (s *EVMExecutorSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.state = evmpkg.NewStateDB()
	s.executor = evmpkg.NewExecutor(s.state)
	s.executor.EnableTracing()
	s.updateState()
	return nil
}

func (s *EVMExecutorSimulator) ExecuteBytecode(bytecodeHex string, calldataHex string) (*evmpkg.ExecutionResult, error) {
	if len(bytecodeHex) > 2 && bytecodeHex[:2] == "0x" {
		bytecodeHex = bytecodeHex[2:]
	}
	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return nil, fmt.Errorf("无效的字节码: %v", err)
	}

	var calldata []byte
	if calldataHex != "" {
		if len(calldataHex) > 2 && calldataHex[:2] == "0x" {
			calldataHex = calldataHex[2:]
		}
		calldata, _ = hex.DecodeString(calldataHex)
	}

	ctx := evmpkg.NewExecutionContext()
	ctx.Input = calldata
	ctx.Address = evmpkg.HexToAddress("0x1234567890123456789012345678901234567890")
	ctx.Caller = evmpkg.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	ctx.Origin = ctx.Caller
	ctx.Value = big.NewInt(0)
	ctx.BlockNumber = big.NewInt(12345678)
	ctx.BlockTimestamp = uint64(time.Now().Unix())

	result := s.executor.Execute(bytecode, ctx)

	s.steps = make([]*ExecutionStep, 0)
	for _, trace := range s.executor.GetTraces() {
		s.steps = append(s.steps, &ExecutionStep{
			PC:      trace.PC,
			OpCode:  trace.OpCode,
			Gas:     trace.Gas,
			GasCost: trace.GasCost,
			Stack:   trace.Stack,
			Memory:  trace.Memory,
		})
	}

	s.EmitEvent("execution_complete", "", "", map[string]interface{}{
		"success":     result.Success,
		"gas_used":    result.GasUsed,
		"step_count":  len(s.steps),
		"return_data": hex.EncodeToString(result.ReturnData),
	})

	s.updateState()
	return result, nil
}

func (s *EVMExecutorSimulator) ExecuteSimple(example string) (*evmpkg.ExecutionResult, error) {
	var bytecode string
	switch example {
	case "add":
		bytecode = "600560030160005260206000f3"
	case "storage":
		bytecode = "602a600055"
	case "loop":
		bytecode = "600a5b6001900380600357"
	case "call":
		bytecode = "33600052"
	default:
		bytecode = example
	}
	return s.ExecuteBytecode(bytecode, "")
}

func (s *EVMExecutorSimulator) GetSteps() []*ExecutionStep {
	return s.steps
}

func (s *EVMExecutorSimulator) DisassembleBytecode(bytecodeHex string) []map[string]interface{} {
	if len(bytecodeHex) > 2 && bytecodeHex[:2] == "0x" {
		bytecodeHex = bytecodeHex[2:]
	}
	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	pc := 0
	for pc < len(bytecode) {
		op := evmpkg.OpCode(bytecode[pc])
		info, ok := evmpkg.GetOpCodeInfo(op)
		instruction := map[string]interface{}{
			"pc":     pc,
			"opcode": fmt.Sprintf("0x%02x", op),
			"name":   evmpkg.GetOpCodeName(op),
		}
		if ok && info.ImmSize > 0 {
			immStart := pc + 1
			immEnd := immStart + info.ImmSize
			if immEnd > len(bytecode) {
				immEnd = len(bytecode)
			}
			imm := bytecode[immStart:immEnd]
			instruction["immediate"] = "0x" + hex.EncodeToString(imm)
			pc = immEnd
		} else {
			pc++
		}
		result = append(result, instruction)
	}
	return result
}

func (s *EVMExecutorSimulator) updateState() {
	s.SetGlobalData("step_count", len(s.steps))
	if len(s.steps) > 0 {
		lastStep := s.steps[len(s.steps)-1]
		s.SetGlobalData("last_step", map[string]interface{}{
			"pc":     lastStep.PC,
			"opcode": lastStep.OpCode,
			"gas":    lastStep.Gas,
		})
		setEVMTeachingState(
			s.BaseSimulator,
			"evm",
			"execution_completed",
			fmt.Sprintf("最近一次字节码执行共产生 %d 个步骤，最后停在 %s。", len(s.steps), lastStep.OpCode),
			"继续观察栈、内存和存储是如何随着每一条指令逐步变化的。",
			0.9,
			map[string]interface{}{"step_count": len(s.steps), "last_pc": lastStep.PC, "last_opcode": lastStep.OpCode, "last_gas": lastStep.Gas},
		)
		return
	}
	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		"execution_ready",
		"当前还没有执行字节码，可以选择一个示例开始逐步执行。",
		"优先比较 add、storage 和 call 三种示例，理解不同 opcode 对栈和状态的影响。",
		0.2,
		map[string]interface{}{"step_count": 0},
	)
}

func (s *EVMExecutorSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "execute_example":
		example := "add"
		if raw, ok := params["example"].(string); ok && raw != "" {
			example = raw
		}
		result, err := s.ExecuteSimple(example)
		if err != nil {
			return nil, err
		}
		return evmActionResult(
			"已执行一个 EVM 示例。",
			map[string]interface{}{"example": example, "success": result.Success, "gas_used": result.GasUsed, "step_count": len(s.steps)},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("%s 示例已经执行完成。", example),
				NextHint:    "重点查看 opcode 顺序、gas 消耗以及状态变化是否符合预期。",
				EffectScope: "evm",
				ResultState: map[string]interface{}{"example": example, "gas_used": result.GasUsed},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported evm executor action: %s", action)
	}
}

type EVMExecutorFactory struct{}

func (f *EVMExecutorFactory) Create() engine.Simulator {
	return NewEVMExecutorSimulator()
}

func (f *EVMExecutorFactory) GetDescription() types.Description {
	return NewEVMExecutorSimulator().GetDescription()
}

func NewEVMExecutorFactory() *EVMExecutorFactory {
	return &EVMExecutorFactory{}
}

var _ engine.SimulatorFactory = (*EVMExecutorFactory)(nil)
