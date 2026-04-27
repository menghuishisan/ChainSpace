package evm

// =============================================================================
// EVM操作码定义
// =============================================================================

// OpCode 操作码类型
type OpCode byte

// 操作码常量
const (
	// 0x00 - 停止和算术运算
	STOP       OpCode = 0x00
	ADD        OpCode = 0x01
	MUL        OpCode = 0x02
	SUB        OpCode = 0x03
	DIV        OpCode = 0x04
	SDIV       OpCode = 0x05
	MOD        OpCode = 0x06
	SMOD       OpCode = 0x07
	ADDMOD     OpCode = 0x08
	MULMOD     OpCode = 0x09
	EXP        OpCode = 0x0A
	SIGNEXTEND OpCode = 0x0B

	// 0x10 - 比较和位运算
	LT     OpCode = 0x10
	GT     OpCode = 0x11
	SLT    OpCode = 0x12
	SGT    OpCode = 0x13
	EQ     OpCode = 0x14
	ISZERO OpCode = 0x15
	AND    OpCode = 0x16
	OR     OpCode = 0x17
	XOR    OpCode = 0x18
	NOT    OpCode = 0x19
	BYTE   OpCode = 0x1A
	SHL    OpCode = 0x1B
	SHR    OpCode = 0x1C
	SAR    OpCode = 0x1D

	// 0x20 - 哈希
	KECCAK256 OpCode = 0x20

	// 0x30 - 环境信息
	ADDRESS        OpCode = 0x30
	BALANCE        OpCode = 0x31
	ORIGIN         OpCode = 0x32
	CALLER         OpCode = 0x33
	CALLVALUE      OpCode = 0x34
	CALLDATALOAD   OpCode = 0x35
	CALLDATASIZE   OpCode = 0x36
	CALLDATACOPY   OpCode = 0x37
	CODESIZE       OpCode = 0x38
	CODECOPY       OpCode = 0x39
	GASPRICE       OpCode = 0x3A
	EXTCODESIZE    OpCode = 0x3B
	EXTCODECOPY    OpCode = 0x3C
	RETURNDATASIZE OpCode = 0x3D
	RETURNDATACOPY OpCode = 0x3E
	EXTCODEHASH    OpCode = 0x3F

	// 0x40 - 区块信息
	BLOCKHASH   OpCode = 0x40
	COINBASE    OpCode = 0x41
	TIMESTAMP   OpCode = 0x42
	NUMBER      OpCode = 0x43
	DIFFICULTY  OpCode = 0x44
	GASLIMIT    OpCode = 0x45
	CHAINID     OpCode = 0x46
	SELFBALANCE OpCode = 0x47
	BASEFEE     OpCode = 0x48

	// 0x50 - 栈、内存、存储、流程控制
	POP      OpCode = 0x50
	MLOAD    OpCode = 0x51
	MSTORE   OpCode = 0x52
	MSTORE8  OpCode = 0x53
	SLOAD    OpCode = 0x54
	SSTORE   OpCode = 0x55
	JUMP     OpCode = 0x56
	JUMPI    OpCode = 0x57
	PC       OpCode = 0x58
	MSIZE    OpCode = 0x59
	GAS      OpCode = 0x5A
	JUMPDEST OpCode = 0x5B

	// 0x5F - PUSH0 (EIP-3855)
	PUSH0 OpCode = 0x5F

	// 0x60-0x7F - PUSH1-PUSH32
	PUSH1  OpCode = 0x60
	PUSH2  OpCode = 0x61
	PUSH3  OpCode = 0x62
	PUSH4  OpCode = 0x63
	PUSH5  OpCode = 0x64
	PUSH6  OpCode = 0x65
	PUSH7  OpCode = 0x66
	PUSH8  OpCode = 0x67
	PUSH9  OpCode = 0x68
	PUSH10 OpCode = 0x69
	PUSH11 OpCode = 0x6A
	PUSH12 OpCode = 0x6B
	PUSH13 OpCode = 0x6C
	PUSH14 OpCode = 0x6D
	PUSH15 OpCode = 0x6E
	PUSH16 OpCode = 0x6F
	PUSH17 OpCode = 0x70
	PUSH18 OpCode = 0x71
	PUSH19 OpCode = 0x72
	PUSH20 OpCode = 0x73
	PUSH21 OpCode = 0x74
	PUSH22 OpCode = 0x75
	PUSH23 OpCode = 0x76
	PUSH24 OpCode = 0x77
	PUSH25 OpCode = 0x78
	PUSH26 OpCode = 0x79
	PUSH27 OpCode = 0x7A
	PUSH28 OpCode = 0x7B
	PUSH29 OpCode = 0x7C
	PUSH30 OpCode = 0x7D
	PUSH31 OpCode = 0x7E
	PUSH32 OpCode = 0x7F

	// 0x80-0x8F - DUP1-DUP16
	DUP1  OpCode = 0x80
	DUP2  OpCode = 0x81
	DUP3  OpCode = 0x82
	DUP4  OpCode = 0x83
	DUP5  OpCode = 0x84
	DUP6  OpCode = 0x85
	DUP7  OpCode = 0x86
	DUP8  OpCode = 0x87
	DUP9  OpCode = 0x88
	DUP10 OpCode = 0x89
	DUP11 OpCode = 0x8A
	DUP12 OpCode = 0x8B
	DUP13 OpCode = 0x8C
	DUP14 OpCode = 0x8D
	DUP15 OpCode = 0x8E
	DUP16 OpCode = 0x8F

	// 0x90-0x9F - SWAP1-SWAP16
	SWAP1  OpCode = 0x90
	SWAP2  OpCode = 0x91
	SWAP3  OpCode = 0x92
	SWAP4  OpCode = 0x93
	SWAP5  OpCode = 0x94
	SWAP6  OpCode = 0x95
	SWAP7  OpCode = 0x96
	SWAP8  OpCode = 0x97
	SWAP9  OpCode = 0x98
	SWAP10 OpCode = 0x99
	SWAP11 OpCode = 0x9A
	SWAP12 OpCode = 0x9B
	SWAP13 OpCode = 0x9C
	SWAP14 OpCode = 0x9D
	SWAP15 OpCode = 0x9E
	SWAP16 OpCode = 0x9F

	// 0xA0-0xA4 - LOG0-LOG4
	LOG0 OpCode = 0xA0
	LOG1 OpCode = 0xA1
	LOG2 OpCode = 0xA2
	LOG3 OpCode = 0xA3
	LOG4 OpCode = 0xA4

	// 0xF0-0xFF - 系统操作
	CREATE       OpCode = 0xF0
	CALL         OpCode = 0xF1
	CALLCODE     OpCode = 0xF2
	RETURN       OpCode = 0xF3
	DELEGATECALL OpCode = 0xF4
	CREATE2      OpCode = 0xF5
	STATICCALL   OpCode = 0xFA
	REVERT       OpCode = 0xFD
	INVALID      OpCode = 0xFE
	SELFDESTRUCT OpCode = 0xFF
)

// OpCodeInfo 操作码信息
type OpCodeInfo struct {
	Name       string // 操作码名称
	StackPop   int    // 弹出栈元素数
	StackPush  int    // 压入栈元素数
	Gas        uint64 // 基础Gas消耗
	ImmSize    int    // 立即数大小(字节)
	IsJump     bool   // 是否是跳转指令
	IsTerminal bool   // 是否终止执行
}

// OpCodeInfoMap 操作码信息表
var OpCodeInfoMap = map[OpCode]OpCodeInfo{
	// 停止和算术运算
	STOP:       {"STOP", 0, 0, 0, 0, false, true},
	ADD:        {"ADD", 2, 1, 3, 0, false, false},
	MUL:        {"MUL", 2, 1, 5, 0, false, false},
	SUB:        {"SUB", 2, 1, 3, 0, false, false},
	DIV:        {"DIV", 2, 1, 5, 0, false, false},
	SDIV:       {"SDIV", 2, 1, 5, 0, false, false},
	MOD:        {"MOD", 2, 1, 5, 0, false, false},
	SMOD:       {"SMOD", 2, 1, 5, 0, false, false},
	ADDMOD:     {"ADDMOD", 3, 1, 8, 0, false, false},
	MULMOD:     {"MULMOD", 3, 1, 8, 0, false, false},
	EXP:        {"EXP", 2, 1, 10, 0, false, false},
	SIGNEXTEND: {"SIGNEXTEND", 2, 1, 5, 0, false, false},

	// 比较和位运算
	LT:     {"LT", 2, 1, 3, 0, false, false},
	GT:     {"GT", 2, 1, 3, 0, false, false},
	SLT:    {"SLT", 2, 1, 3, 0, false, false},
	SGT:    {"SGT", 2, 1, 3, 0, false, false},
	EQ:     {"EQ", 2, 1, 3, 0, false, false},
	ISZERO: {"ISZERO", 1, 1, 3, 0, false, false},
	AND:    {"AND", 2, 1, 3, 0, false, false},
	OR:     {"OR", 2, 1, 3, 0, false, false},
	XOR:    {"XOR", 2, 1, 3, 0, false, false},
	NOT:    {"NOT", 1, 1, 3, 0, false, false},
	BYTE:   {"BYTE", 2, 1, 3, 0, false, false},
	SHL:    {"SHL", 2, 1, 3, 0, false, false},
	SHR:    {"SHR", 2, 1, 3, 0, false, false},
	SAR:    {"SAR", 2, 1, 3, 0, false, false},

	// 哈希
	KECCAK256: {"KECCAK256", 2, 1, 30, 0, false, false},

	// 环境信息
	ADDRESS:        {"ADDRESS", 0, 1, 2, 0, false, false},
	BALANCE:        {"BALANCE", 1, 1, 100, 0, false, false},
	ORIGIN:         {"ORIGIN", 0, 1, 2, 0, false, false},
	CALLER:         {"CALLER", 0, 1, 2, 0, false, false},
	CALLVALUE:      {"CALLVALUE", 0, 1, 2, 0, false, false},
	CALLDATALOAD:   {"CALLDATALOAD", 1, 1, 3, 0, false, false},
	CALLDATASIZE:   {"CALLDATASIZE", 0, 1, 2, 0, false, false},
	CALLDATACOPY:   {"CALLDATACOPY", 3, 0, 3, 0, false, false},
	CODESIZE:       {"CODESIZE", 0, 1, 2, 0, false, false},
	CODECOPY:       {"CODECOPY", 3, 0, 3, 0, false, false},
	GASPRICE:       {"GASPRICE", 0, 1, 2, 0, false, false},
	EXTCODESIZE:    {"EXTCODESIZE", 1, 1, 100, 0, false, false},
	EXTCODECOPY:    {"EXTCODECOPY", 4, 0, 100, 0, false, false},
	RETURNDATASIZE: {"RETURNDATASIZE", 0, 1, 2, 0, false, false},
	RETURNDATACOPY: {"RETURNDATACOPY", 3, 0, 3, 0, false, false},
	EXTCODEHASH:    {"EXTCODEHASH", 1, 1, 100, 0, false, false},

	// 区块信息
	BLOCKHASH:   {"BLOCKHASH", 1, 1, 20, 0, false, false},
	COINBASE:    {"COINBASE", 0, 1, 2, 0, false, false},
	TIMESTAMP:   {"TIMESTAMP", 0, 1, 2, 0, false, false},
	NUMBER:      {"NUMBER", 0, 1, 2, 0, false, false},
	DIFFICULTY:  {"DIFFICULTY", 0, 1, 2, 0, false, false},
	GASLIMIT:    {"GASLIMIT", 0, 1, 2, 0, false, false},
	CHAINID:     {"CHAINID", 0, 1, 2, 0, false, false},
	SELFBALANCE: {"SELFBALANCE", 0, 1, 5, 0, false, false},
	BASEFEE:     {"BASEFEE", 0, 1, 2, 0, false, false},

	// 栈、内存、存储、流程控制
	POP:      {"POP", 1, 0, 2, 0, false, false},
	MLOAD:    {"MLOAD", 1, 1, 3, 0, false, false},
	MSTORE:   {"MSTORE", 2, 0, 3, 0, false, false},
	MSTORE8:  {"MSTORE8", 2, 0, 3, 0, false, false},
	SLOAD:    {"SLOAD", 1, 1, 100, 0, false, false},
	SSTORE:   {"SSTORE", 2, 0, 100, 0, false, false},
	JUMP:     {"JUMP", 1, 0, 8, 0, true, false},
	JUMPI:    {"JUMPI", 2, 0, 10, 0, true, false},
	PC:       {"PC", 0, 1, 2, 0, false, false},
	MSIZE:    {"MSIZE", 0, 1, 2, 0, false, false},
	GAS:      {"GAS", 0, 1, 2, 0, false, false},
	JUMPDEST: {"JUMPDEST", 0, 0, 1, 0, false, false},

	// PUSH0
	PUSH0: {"PUSH0", 0, 1, 2, 0, false, false},

	// PUSH1-PUSH32
	PUSH1:  {"PUSH1", 0, 1, 3, 1, false, false},
	PUSH2:  {"PUSH2", 0, 1, 3, 2, false, false},
	PUSH3:  {"PUSH3", 0, 1, 3, 3, false, false},
	PUSH4:  {"PUSH4", 0, 1, 3, 4, false, false},
	PUSH5:  {"PUSH5", 0, 1, 3, 5, false, false},
	PUSH6:  {"PUSH6", 0, 1, 3, 6, false, false},
	PUSH7:  {"PUSH7", 0, 1, 3, 7, false, false},
	PUSH8:  {"PUSH8", 0, 1, 3, 8, false, false},
	PUSH9:  {"PUSH9", 0, 1, 3, 9, false, false},
	PUSH10: {"PUSH10", 0, 1, 3, 10, false, false},
	PUSH11: {"PUSH11", 0, 1, 3, 11, false, false},
	PUSH12: {"PUSH12", 0, 1, 3, 12, false, false},
	PUSH13: {"PUSH13", 0, 1, 3, 13, false, false},
	PUSH14: {"PUSH14", 0, 1, 3, 14, false, false},
	PUSH15: {"PUSH15", 0, 1, 3, 15, false, false},
	PUSH16: {"PUSH16", 0, 1, 3, 16, false, false},
	PUSH17: {"PUSH17", 0, 1, 3, 17, false, false},
	PUSH18: {"PUSH18", 0, 1, 3, 18, false, false},
	PUSH19: {"PUSH19", 0, 1, 3, 19, false, false},
	PUSH20: {"PUSH20", 0, 1, 3, 20, false, false},
	PUSH21: {"PUSH21", 0, 1, 3, 21, false, false},
	PUSH22: {"PUSH22", 0, 1, 3, 22, false, false},
	PUSH23: {"PUSH23", 0, 1, 3, 23, false, false},
	PUSH24: {"PUSH24", 0, 1, 3, 24, false, false},
	PUSH25: {"PUSH25", 0, 1, 3, 25, false, false},
	PUSH26: {"PUSH26", 0, 1, 3, 26, false, false},
	PUSH27: {"PUSH27", 0, 1, 3, 27, false, false},
	PUSH28: {"PUSH28", 0, 1, 3, 28, false, false},
	PUSH29: {"PUSH29", 0, 1, 3, 29, false, false},
	PUSH30: {"PUSH30", 0, 1, 3, 30, false, false},
	PUSH31: {"PUSH31", 0, 1, 3, 31, false, false},
	PUSH32: {"PUSH32", 0, 1, 3, 32, false, false},

	// DUP1-DUP16
	DUP1:  {"DUP1", 1, 2, 3, 0, false, false},
	DUP2:  {"DUP2", 2, 3, 3, 0, false, false},
	DUP3:  {"DUP3", 3, 4, 3, 0, false, false},
	DUP4:  {"DUP4", 4, 5, 3, 0, false, false},
	DUP5:  {"DUP5", 5, 6, 3, 0, false, false},
	DUP6:  {"DUP6", 6, 7, 3, 0, false, false},
	DUP7:  {"DUP7", 7, 8, 3, 0, false, false},
	DUP8:  {"DUP8", 8, 9, 3, 0, false, false},
	DUP9:  {"DUP9", 9, 10, 3, 0, false, false},
	DUP10: {"DUP10", 10, 11, 3, 0, false, false},
	DUP11: {"DUP11", 11, 12, 3, 0, false, false},
	DUP12: {"DUP12", 12, 13, 3, 0, false, false},
	DUP13: {"DUP13", 13, 14, 3, 0, false, false},
	DUP14: {"DUP14", 14, 15, 3, 0, false, false},
	DUP15: {"DUP15", 15, 16, 3, 0, false, false},
	DUP16: {"DUP16", 16, 17, 3, 0, false, false},

	// SWAP1-SWAP16
	SWAP1:  {"SWAP1", 2, 2, 3, 0, false, false},
	SWAP2:  {"SWAP2", 3, 3, 3, 0, false, false},
	SWAP3:  {"SWAP3", 4, 4, 3, 0, false, false},
	SWAP4:  {"SWAP4", 5, 5, 3, 0, false, false},
	SWAP5:  {"SWAP5", 6, 6, 3, 0, false, false},
	SWAP6:  {"SWAP6", 7, 7, 3, 0, false, false},
	SWAP7:  {"SWAP7", 8, 8, 3, 0, false, false},
	SWAP8:  {"SWAP8", 9, 9, 3, 0, false, false},
	SWAP9:  {"SWAP9", 10, 10, 3, 0, false, false},
	SWAP10: {"SWAP10", 11, 11, 3, 0, false, false},
	SWAP11: {"SWAP11", 12, 12, 3, 0, false, false},
	SWAP12: {"SWAP12", 13, 13, 3, 0, false, false},
	SWAP13: {"SWAP13", 14, 14, 3, 0, false, false},
	SWAP14: {"SWAP14", 15, 15, 3, 0, false, false},
	SWAP15: {"SWAP15", 16, 16, 3, 0, false, false},
	SWAP16: {"SWAP16", 17, 17, 3, 0, false, false},

	// LOG0-LOG4
	LOG0: {"LOG0", 2, 0, 375, 0, false, false},
	LOG1: {"LOG1", 3, 0, 750, 0, false, false},
	LOG2: {"LOG2", 4, 0, 1125, 0, false, false},
	LOG3: {"LOG3", 5, 0, 1500, 0, false, false},
	LOG4: {"LOG4", 6, 0, 1875, 0, false, false},

	// 系统操作
	CREATE:       {"CREATE", 3, 1, 32000, 0, false, false},
	CALL:         {"CALL", 7, 1, 100, 0, false, false},
	CALLCODE:     {"CALLCODE", 7, 1, 100, 0, false, false},
	RETURN:       {"RETURN", 2, 0, 0, 0, false, true},
	DELEGATECALL: {"DELEGATECALL", 6, 1, 100, 0, false, false},
	CREATE2:      {"CREATE2", 4, 1, 32000, 0, false, false},
	STATICCALL:   {"STATICCALL", 6, 1, 100, 0, false, false},
	REVERT:       {"REVERT", 2, 0, 0, 0, false, true},
	INVALID:      {"INVALID", 0, 0, 0, 0, false, true},
	SELFDESTRUCT: {"SELFDESTRUCT", 1, 0, 5000, 0, false, true},
}

// GetOpCodeInfo 获取操作码信息
func GetOpCodeInfo(op OpCode) (OpCodeInfo, bool) {
	info, ok := OpCodeInfoMap[op]
	return info, ok
}

// GetOpCodeName 获取操作码名称
func GetOpCodeName(op OpCode) string {
	if info, ok := OpCodeInfoMap[op]; ok {
		return info.Name
	}
	return "UNKNOWN"
}

// IsPush 是否是PUSH指令
func IsPush(op OpCode) bool {
	return op >= PUSH1 && op <= PUSH32
}

// IsDup 是否是DUP指令
func IsDup(op OpCode) bool {
	return op >= DUP1 && op <= DUP16
}

// IsSwap 是否是SWAP指令
func IsSwap(op OpCode) bool {
	return op >= SWAP1 && op <= SWAP16
}

// IsLog 是否是LOG指令
func IsLog(op OpCode) bool {
	return op >= LOG0 && op <= LOG4
}

// GetPushSize 获取PUSH指令的数据大小
func GetPushSize(op OpCode) int {
	if op == PUSH0 {
		return 0
	}
	if IsPush(op) {
		return int(op - PUSH1 + 1)
	}
	return 0
}

// GetDupN 获取DUP指令的N值
func GetDupN(op OpCode) int {
	if IsDup(op) {
		return int(op - DUP1 + 1)
	}
	return 0
}

// GetSwapN 获取SWAP指令的N值
func GetSwapN(op OpCode) int {
	if IsSwap(op) {
		return int(op - SWAP1 + 1)
	}
	return 0
}

// GetLogTopics 获取LOG指令的Topics数
func GetLogTopics(op OpCode) int {
	if IsLog(op) {
		return int(op - LOG0)
	}
	return 0
}
