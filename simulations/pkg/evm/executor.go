package evm

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// =============================================================================
// EVM执行器
// =============================================================================

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success    bool   `json:"success"`
	ReturnData []byte `json:"return_data"`
	GasUsed    uint64 `json:"gas_used"`
	GasLeft    uint64 `json:"gas_left"`
	Logs       []*Log `json:"logs"`
	Error      string `json:"error,omitempty"`
}

// ExecutionTrace 执行跟踪
type ExecutionTrace struct {
	PC      uint64   `json:"pc"`
	OpCode  string   `json:"opcode"`
	Gas     uint64   `json:"gas"`
	GasCost uint64   `json:"gas_cost"`
	Depth   int      `json:"depth"`
	Stack   []string `json:"stack"`
	Memory  string   `json:"memory,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Executor EVM执行器
type Executor struct {
	state   *StateDB
	code    []byte
	ctx     *ExecutionContext
	traces  []*ExecutionTrace
	tracing bool
}

// NewExecutor 创建执行器
func NewExecutor(state *StateDB) *Executor {
	return &Executor{
		state:   state,
		traces:  make([]*ExecutionTrace, 0),
		tracing: false,
	}
}

// EnableTracing 启用跟踪
func (e *Executor) EnableTracing() {
	e.tracing = true
}

// GetTraces 获取执行跟踪
func (e *Executor) GetTraces() []*ExecutionTrace {
	return e.traces
}

// Execute 执行字节码
func (e *Executor) Execute(code []byte, ctx *ExecutionContext) *ExecutionResult {
	e.code = code
	e.ctx = ctx
	e.traces = make([]*ExecutionTrace, 0)

	result := &ExecutionResult{
		Success: true,
		GasUsed: 0,
	}

	// 执行循环
	for {
		// 检查PC是否越界
		if ctx.PC >= uint64(len(code)) {
			break
		}

		// 读取操作码
		op := OpCode(code[ctx.PC])
		info, ok := GetOpCodeInfo(op)
		if !ok {
			result.Success = false
			result.Error = fmt.Sprintf("invalid opcode: 0x%02x at PC=%d", op, ctx.PC)
			break
		}

		// 检查Gas
		if ctx.Gas < info.Gas {
			result.Success = false
			result.Error = "out of gas"
			break
		}

		// 记录跟踪
		if e.tracing {
			trace := &ExecutionTrace{
				PC:      ctx.PC,
				OpCode:  info.Name,
				Gas:     ctx.Gas,
				GasCost: info.Gas,
				Depth:   ctx.Depth,
				Stack:   e.stackToStrings(),
			}
			e.traces = append(e.traces, trace)
		}

		// 消耗Gas
		ctx.Gas -= info.Gas
		result.GasUsed += info.Gas

		// 执行操作码
		err := e.executeOpCode(op)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			break
		}

		// 检查是否终止
		if info.IsTerminal {
			if op == RETURN || op == REVERT {
				result.ReturnData = ctx.ReturnData
				if op == REVERT {
					result.Success = false
				}
			}
			break
		}

		// 移动PC (除了跳转指令)
		if !info.IsJump {
			ctx.PC += 1 + uint64(info.ImmSize)
		}
	}

	result.GasLeft = ctx.Gas
	result.Logs = e.state.GetLogs()
	return result
}

// executeOpCode 执行单个操作码
func (e *Executor) executeOpCode(op OpCode) error {
	ctx := e.ctx
	stack := ctx.Stack

	switch op {
	// 停止
	case STOP:
		return nil

	// 算术运算
	case ADD:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		result := new(big.Int).Add(a, b)
		result.And(result, maxUint256())
		return stack.Push(result)

	case MUL:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		result := new(big.Int).Mul(a, b)
		result.And(result, maxUint256())
		return stack.Push(result)

	case SUB:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		result := new(big.Int).Sub(a, b)
		if result.Sign() < 0 {
			result.Add(result, new(big.Int).Add(maxUint256(), big.NewInt(1)))
		}
		return stack.Push(result)

	case DIV:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		if b.Sign() == 0 {
			return stack.Push(big.NewInt(0))
		}
		return stack.Push(new(big.Int).Div(a, b))

	case MOD:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		if b.Sign() == 0 {
			return stack.Push(big.NewInt(0))
		}
		return stack.Push(new(big.Int).Mod(a, b))

	case EXP:
		base, _ := stack.Pop()
		exp, _ := stack.Pop()
		result := new(big.Int).Exp(base, exp, new(big.Int).Add(maxUint256(), big.NewInt(1)))
		return stack.Push(result)

	// 比较运算
	case LT:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		if a.Cmp(b) < 0 {
			return stack.Push(big.NewInt(1))
		}
		return stack.Push(big.NewInt(0))

	case GT:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		if a.Cmp(b) > 0 {
			return stack.Push(big.NewInt(1))
		}
		return stack.Push(big.NewInt(0))

	case EQ:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		if a.Cmp(b) == 0 {
			return stack.Push(big.NewInt(1))
		}
		return stack.Push(big.NewInt(0))

	case ISZERO:
		a, _ := stack.Pop()
		if a.Sign() == 0 {
			return stack.Push(big.NewInt(1))
		}
		return stack.Push(big.NewInt(0))

	// 位运算
	case AND:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		return stack.Push(new(big.Int).And(a, b))

	case OR:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		return stack.Push(new(big.Int).Or(a, b))

	case XOR:
		a, _ := stack.Pop()
		b, _ := stack.Pop()
		return stack.Push(new(big.Int).Xor(a, b))

	case NOT:
		a, _ := stack.Pop()
		return stack.Push(new(big.Int).Xor(a, maxUint256()))

	case SHL:
		shift, _ := stack.Pop()
		value, _ := stack.Pop()
		if shift.Cmp(big.NewInt(256)) >= 0 {
			return stack.Push(big.NewInt(0))
		}
		result := new(big.Int).Lsh(value, uint(shift.Uint64()))
		result.And(result, maxUint256())
		return stack.Push(result)

	case SHR:
		shift, _ := stack.Pop()
		value, _ := stack.Pop()
		if shift.Cmp(big.NewInt(256)) >= 0 {
			return stack.Push(big.NewInt(0))
		}
		return stack.Push(new(big.Int).Rsh(value, uint(shift.Uint64())))

	// 哈希
	case KECCAK256:
		offset, _ := stack.Pop()
		size, _ := stack.Pop()
		data := ctx.Memory.Get(offset.Uint64(), size.Uint64())
		hash := keccak256(data)
		return stack.Push(new(big.Int).SetBytes(hash))

	// 环境信息
	case ADDRESS:
		return stack.Push(new(big.Int).SetBytes(ctx.Address[:]))

	case BALANCE:
		addr, _ := stack.Pop()
		address := bigIntToAddress(addr)
		balance := e.state.GetBalance(address)
		return stack.Push(balance)

	case ORIGIN:
		return stack.Push(new(big.Int).SetBytes(ctx.Origin[:]))

	case CALLER:
		return stack.Push(new(big.Int).SetBytes(ctx.Caller[:]))

	case CALLVALUE:
		return stack.Push(new(big.Int).Set(ctx.Value))

	case CALLDATALOAD:
		offset, _ := stack.Pop()
		off := offset.Uint64()
		data := make([]byte, 32)
		if off < uint64(len(ctx.Input)) {
			copy(data, ctx.Input[off:])
		}
		return stack.Push(new(big.Int).SetBytes(data))

	case CALLDATASIZE:
		return stack.Push(big.NewInt(int64(len(ctx.Input))))

	case CALLDATACOPY:
		destOffset, _ := stack.Pop()
		offset, _ := stack.Pop()
		size, _ := stack.Pop()
		data := make([]byte, size.Uint64())
		off := offset.Uint64()
		if off < uint64(len(ctx.Input)) {
			copy(data, ctx.Input[off:])
		}
		ctx.Memory.Set(destOffset.Uint64(), size.Uint64(), data)
		return nil

	case CODESIZE:
		return stack.Push(big.NewInt(int64(len(e.code))))

	case CODECOPY:
		destOffset, _ := stack.Pop()
		offset, _ := stack.Pop()
		size, _ := stack.Pop()
		data := make([]byte, size.Uint64())
		off := offset.Uint64()
		if off < uint64(len(e.code)) {
			copy(data, e.code[off:])
		}
		ctx.Memory.Set(destOffset.Uint64(), size.Uint64(), data)
		return nil

	case GASPRICE:
		return stack.Push(new(big.Int).Set(ctx.GasPrice))

	case RETURNDATASIZE:
		return stack.Push(big.NewInt(int64(len(ctx.ReturnData))))

	case RETURNDATACOPY:
		destOffset, _ := stack.Pop()
		offset, _ := stack.Pop()
		size, _ := stack.Pop()
		data := make([]byte, size.Uint64())
		off := offset.Uint64()
		if off < uint64(len(ctx.ReturnData)) {
			copy(data, ctx.ReturnData[off:])
		}
		ctx.Memory.Set(destOffset.Uint64(), size.Uint64(), data)
		return nil

	// 区块信息
	case BLOCKHASH:
		num, _ := stack.Pop()
		// 简化实现
		_ = num
		return stack.Push(big.NewInt(0))

	case COINBASE:
		return stack.Push(new(big.Int).SetBytes(ctx.Coinbase[:]))

	case TIMESTAMP:
		return stack.Push(big.NewInt(int64(ctx.BlockTimestamp)))

	case NUMBER:
		return stack.Push(new(big.Int).Set(ctx.BlockNumber))

	case DIFFICULTY:
		return stack.Push(new(big.Int).Set(ctx.Difficulty))

	case GASLIMIT:
		return stack.Push(big.NewInt(int64(ctx.BlockGasLimit)))

	case CHAINID:
		return stack.Push(big.NewInt(1)) // 主网

	case SELFBALANCE:
		balance := e.state.GetBalance(ctx.Address)
		return stack.Push(balance)

	case BASEFEE:
		return stack.Push(new(big.Int).Set(ctx.BaseFee))

	// 栈操作
	case POP:
		_, _ = stack.Pop()
		return nil

	// 内存操作
	case MLOAD:
		offset, _ := stack.Pop()
		data := ctx.Memory.Get(offset.Uint64(), 32)
		return stack.Push(new(big.Int).SetBytes(data))

	case MSTORE:
		offset, _ := stack.Pop()
		value, _ := stack.Pop()
		ctx.Memory.Set32(offset.Uint64(), value)
		return nil

	case MSTORE8:
		offset, _ := stack.Pop()
		value, _ := stack.Pop()
		ctx.Memory.Set(offset.Uint64(), 1, []byte{byte(value.Uint64() & 0xFF)})
		return nil

	case MSIZE:
		return stack.Push(big.NewInt(int64(ctx.Memory.Len())))

	// 存储操作
	case SLOAD:
		key, _ := stack.Pop()
		keyHash := BigIntToHash(key)
		value := e.state.GetStorage(ctx.Address, keyHash)
		return stack.Push(HashToBigInt(value))

	case SSTORE:
		key, _ := stack.Pop()
		value, _ := stack.Pop()
		keyHash := BigIntToHash(key)
		valueHash := BigIntToHash(value)
		e.state.SetStorage(ctx.Address, keyHash, valueHash)
		return nil

	// 流程控制
	case JUMP:
		dest, _ := stack.Pop()
		ctx.PC = dest.Uint64()
		// 验证跳转目标
		if ctx.PC >= uint64(len(e.code)) || OpCode(e.code[ctx.PC]) != JUMPDEST {
			return fmt.Errorf("invalid jump destination")
		}
		return nil

	case JUMPI:
		dest, _ := stack.Pop()
		cond, _ := stack.Pop()
		if cond.Sign() != 0 {
			ctx.PC = dest.Uint64()
			if ctx.PC >= uint64(len(e.code)) || OpCode(e.code[ctx.PC]) != JUMPDEST {
				return fmt.Errorf("invalid jump destination")
			}
		} else {
			ctx.PC += 1
		}
		return nil

	case PC:
		return stack.Push(big.NewInt(int64(ctx.PC)))

	case GAS:
		return stack.Push(big.NewInt(int64(ctx.Gas)))

	case JUMPDEST:
		return nil

	// PUSH0
	case PUSH0:
		return stack.Push(big.NewInt(0))

	// PUSH1-PUSH32
	case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8,
		PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16,
		PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24,
		PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
		size := GetPushSize(op)
		start := ctx.PC + 1
		end := start + uint64(size)
		if end > uint64(len(e.code)) {
			end = uint64(len(e.code))
		}
		data := make([]byte, size)
		copy(data[size-int(end-start):], e.code[start:end])
		return stack.Push(new(big.Int).SetBytes(data))

	// DUP1-DUP16
	case DUP1, DUP2, DUP3, DUP4, DUP5, DUP6, DUP7, DUP8,
		DUP9, DUP10, DUP11, DUP12, DUP13, DUP14, DUP15, DUP16:
		n := GetDupN(op)
		return stack.Dup(n)

	// SWAP1-SWAP16
	case SWAP1, SWAP2, SWAP3, SWAP4, SWAP5, SWAP6, SWAP7, SWAP8,
		SWAP9, SWAP10, SWAP11, SWAP12, SWAP13, SWAP14, SWAP15, SWAP16:
		n := GetSwapN(op)
		return stack.Swap(n)

	// LOG0-LOG4
	case LOG0, LOG1, LOG2, LOG3, LOG4:
		offset, _ := stack.Pop()
		size, _ := stack.Pop()
		data := ctx.Memory.Get(offset.Uint64(), size.Uint64())

		topicCount := GetLogTopics(op)
		topics := make([]Hash, topicCount)
		for i := 0; i < topicCount; i++ {
			topic, _ := stack.Pop()
			topics[i] = BigIntToHash(topic)
		}

		log := &Log{
			Address: ctx.Address,
			Topics:  topics,
			Data:    data,
		}
		e.state.AddLog(log)
		return nil

	// RETURN
	case RETURN:
		offset, _ := stack.Pop()
		size, _ := stack.Pop()
		ctx.ReturnData = ctx.Memory.Get(offset.Uint64(), size.Uint64())
		return nil

	// REVERT
	case REVERT:
		offset, _ := stack.Pop()
		size, _ := stack.Pop()
		ctx.ReturnData = ctx.Memory.Get(offset.Uint64(), size.Uint64())
		return fmt.Errorf("execution reverted")

	// INVALID
	case INVALID:
		return fmt.Errorf("invalid opcode")

	default:
		return fmt.Errorf("unimplemented opcode: %s", GetOpCodeName(op))
	}
}

// stackToStrings 栈转字符串数组
func (e *Executor) stackToStrings() []string {
	data := e.ctx.Stack.Data()
	result := make([]string, len(data))
	for i, v := range data {
		result[i] = fmt.Sprintf("0x%s", v.Text(16))
	}
	return result
}

// =============================================================================
// 辅助函数
// =============================================================================

// maxUint256 返回uint256最大值
func maxUint256() *big.Int {
	max := new(big.Int)
	max.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	return max
}

// keccak256 计算Keccak256哈希
func keccak256(data []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	return h.Sum(nil)
}

// bigIntToAddress 大整数转地址
func bigIntToAddress(val *big.Int) Address {
	var addr Address
	bytes := val.Bytes()
	if len(bytes) > 20 {
		bytes = bytes[len(bytes)-20:]
	}
	copy(addr[20-len(bytes):], bytes)
	return addr
}
