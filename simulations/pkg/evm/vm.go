package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
)

// =============================================================================
// EVM核心类型定义
// =============================================================================

// Word 256位字 (32字节)
type Word [32]byte

// Address 以太坊地址 (20字节)
type Address [20]byte

// Hash 哈希值 (32字节)
type Hash [32]byte

// Stack EVM栈
// 最大深度1024，每个元素256位
type Stack struct {
	data []*big.Int
}

// NewStack 创建新栈
func NewStack() *Stack {
	return &Stack{
		data: make([]*big.Int, 0, 1024),
	}
}

// Push 压栈
func (s *Stack) Push(val *big.Int) error {
	if len(s.data) >= 1024 {
		return fmt.Errorf("stack overflow")
	}
	s.data = append(s.data, new(big.Int).Set(val))
	return nil
}

// Pop 出栈
func (s *Stack) Pop() (*big.Int, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("stack underflow")
	}
	val := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return val, nil
}

// Peek 查看栈顶
func (s *Stack) Peek() (*big.Int, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("stack empty")
	}
	return s.data[len(s.data)-1], nil
}

// PeekN 查看第n个元素(从栈顶开始)
func (s *Stack) PeekN(n int) (*big.Int, error) {
	if n >= len(s.data) {
		return nil, fmt.Errorf("stack underflow")
	}
	return s.data[len(s.data)-1-n], nil
}

// Swap 交换栈顶和第n个元素
func (s *Stack) Swap(n int) error {
	if n >= len(s.data) {
		return fmt.Errorf("stack underflow")
	}
	top := len(s.data) - 1
	s.data[top], s.data[top-n] = s.data[top-n], s.data[top]
	return nil
}

// Dup 复制第n个元素到栈顶
func (s *Stack) Dup(n int) error {
	if n > len(s.data) {
		return fmt.Errorf("stack underflow")
	}
	val := s.data[len(s.data)-n]
	return s.Push(new(big.Int).Set(val))
}

// Len 栈深度
func (s *Stack) Len() int {
	return len(s.data)
}

// Data 获取栈数据(用于调试)
func (s *Stack) Data() []*big.Int {
	result := make([]*big.Int, len(s.data))
	copy(result, s.data)
	return result
}

// =============================================================================
// Memory EVM内存
// =============================================================================

// Memory EVM内存
// 按字节寻址，按32字节扩展
type Memory struct {
	data []byte
}

// NewMemory 创建新内存
func NewMemory() *Memory {
	return &Memory{
		data: make([]byte, 0),
	}
}

// Resize 调整内存大小
func (m *Memory) Resize(size uint64) {
	if uint64(len(m.data)) < size {
		// 按32字节对齐扩展
		newSize := ((size + 31) / 32) * 32
		newData := make([]byte, newSize)
		copy(newData, m.data)
		m.data = newData
	}
}

// Set 写入内存
func (m *Memory) Set(offset, size uint64, value []byte) {
	if size == 0 {
		return
	}
	m.Resize(offset + size)
	copy(m.data[offset:offset+size], value)
}

// Set32 写入32字节
func (m *Memory) Set32(offset uint64, val *big.Int) {
	m.Resize(offset + 32)
	data := val.Bytes()
	// 左对齐填充
	start := offset + 32 - uint64(len(data))
	for i := offset; i < start; i++ {
		m.data[i] = 0
	}
	copy(m.data[start:offset+32], data)
}

// Get 读取内存
func (m *Memory) Get(offset, size uint64) []byte {
	if size == 0 {
		return nil
	}
	m.Resize(offset + size)
	result := make([]byte, size)
	copy(result, m.data[offset:offset+size])
	return result
}

// GetPtr 获取内存指针(用于优化)
func (m *Memory) GetPtr(offset, size uint64) []byte {
	m.Resize(offset + size)
	return m.data[offset : offset+size]
}

// Len 内存大小
func (m *Memory) Len() int {
	return len(m.data)
}

// Data 获取内存数据(用于调试)
func (m *Memory) Data() []byte {
	result := make([]byte, len(m.data))
	copy(result, m.data)
	return result
}

// =============================================================================
// Storage EVM存储
// =============================================================================

// Storage 合约存储
// 键值对存储，每个槽32字节
type Storage struct {
	mu   sync.RWMutex
	data map[Hash]Hash
}

// NewStorage 创建新存储
func NewStorage() *Storage {
	return &Storage{
		data: make(map[Hash]Hash),
	}
}

// Get 读取存储槽
func (s *Storage) Get(key Hash) Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// Set 写入存储槽
func (s *Storage) Set(key, value Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value == (Hash{}) {
		delete(s.data, key)
	} else {
		s.data[key] = value
	}
}

// GetAll 获取所有存储(用于调试)
func (s *Storage) GetAll() map[Hash]Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[Hash]Hash)
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// =============================================================================
// Account 账户状态
// =============================================================================

// Account 账户
type Account struct {
	Address  Address  `json:"address"`
	Balance  *big.Int `json:"balance"`
	Nonce    uint64   `json:"nonce"`
	Code     []byte   `json:"code"`
	CodeHash Hash     `json:"code_hash"`
	Storage  *Storage `json:"-"`
}

// NewAccount 创建新账户
func NewAccount(addr Address) *Account {
	return &Account{
		Address: addr,
		Balance: big.NewInt(0),
		Nonce:   0,
		Code:    nil,
		Storage: NewStorage(),
	}
}

// IsContract 是否是合约账户
func (a *Account) IsContract() bool {
	return len(a.Code) > 0
}

// =============================================================================
// ExecutionContext 执行上下文
// =============================================================================

// ExecutionContext EVM执行上下文
type ExecutionContext struct {
	// 区块上下文
	BlockNumber    *big.Int
	BlockTimestamp uint64
	BlockGasLimit  uint64
	Coinbase       Address
	Difficulty     *big.Int
	BaseFee        *big.Int

	// 交易上下文
	Origin   Address // tx.origin
	GasPrice *big.Int

	// 调用上下文
	Caller  Address  // msg.sender
	Value   *big.Int // msg.value
	Address Address  // 当前合约地址
	Input   []byte   // calldata

	// 执行状态
	Gas        uint64
	Stack      *Stack
	Memory     *Memory
	ReturnData []byte
	PC         uint64 // 程序计数器
	Depth      int    // 调用深度
}

// NewExecutionContext 创建执行上下文
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		BlockNumber:    big.NewInt(1),
		BlockTimestamp: 0,
		BlockGasLimit:  30000000,
		Coinbase:       Address{},
		Difficulty:     big.NewInt(1),
		BaseFee:        big.NewInt(1000000000), // 1 Gwei
		Origin:         Address{},
		GasPrice:       big.NewInt(1000000000),
		Caller:         Address{},
		Value:          big.NewInt(0),
		Gas:            1000000,
		Stack:          NewStack(),
		Memory:         NewMemory(),
		ReturnData:     nil,
		PC:             0,
		Depth:          0,
	}
}

// =============================================================================
// StateDB 状态数据库
// =============================================================================

// StateDB 状态数据库接口
type StateDB struct {
	mu       sync.RWMutex
	accounts map[Address]*Account
	logs     []*Log
}

// Log 事件日志
type Log struct {
	Address Address `json:"address"`
	Topics  []Hash  `json:"topics"`
	Data    []byte  `json:"data"`
}

// NewStateDB 创建状态数据库
func NewStateDB() *StateDB {
	return &StateDB{
		accounts: make(map[Address]*Account),
		logs:     make([]*Log, 0),
	}
}

// GetAccount 获取账户
func (s *StateDB) GetAccount(addr Address) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if acc, ok := s.accounts[addr]; ok {
		return acc
	}
	return nil
}

// GetOrCreateAccount 获取或创建账户
func (s *StateDB) GetOrCreateAccount(addr Address) *Account {
	s.mu.Lock()
	defer s.mu.Unlock()
	if acc, ok := s.accounts[addr]; ok {
		return acc
	}
	acc := NewAccount(addr)
	s.accounts[addr] = acc
	return acc
}

// SetBalance 设置余额
func (s *StateDB) SetBalance(addr Address, balance *big.Int) {
	acc := s.GetOrCreateAccount(addr)
	acc.Balance = new(big.Int).Set(balance)
}

// GetBalance 获取余额
func (s *StateDB) GetBalance(addr Address) *big.Int {
	acc := s.GetAccount(addr)
	if acc == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(acc.Balance)
}

// SetCode 设置合约代码
func (s *StateDB) SetCode(addr Address, code []byte) {
	acc := s.GetOrCreateAccount(addr)
	acc.Code = make([]byte, len(code))
	copy(acc.Code, code)
}

// GetCode 获取合约代码
func (s *StateDB) GetCode(addr Address) []byte {
	acc := s.GetAccount(addr)
	if acc == nil {
		return nil
	}
	return acc.Code
}

// GetStorage 获取存储
func (s *StateDB) GetStorage(addr Address, key Hash) Hash {
	acc := s.GetAccount(addr)
	if acc == nil || acc.Storage == nil {
		return Hash{}
	}
	return acc.Storage.Get(key)
}

// SetStorage 设置存储
func (s *StateDB) SetStorage(addr Address, key, value Hash) {
	acc := s.GetOrCreateAccount(addr)
	if acc.Storage == nil {
		acc.Storage = NewStorage()
	}
	acc.Storage.Set(key, value)
}

// AddLog 添加日志
func (s *StateDB) AddLog(log *Log) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, log)
}

// GetLogs 获取日志
func (s *StateDB) GetLogs() []*Log {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Log, len(s.logs))
	copy(result, s.logs)
	return result
}

// =============================================================================
// 辅助函数
// =============================================================================

// HexToAddress 十六进制转地址
func HexToAddress(s string) Address {
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}
	var addr Address
	bytes, _ := hex.DecodeString(s)
	if len(bytes) > 20 {
		bytes = bytes[len(bytes)-20:]
	}
	copy(addr[20-len(bytes):], bytes)
	return addr
}

// HexToHash 十六进制转哈希
func HexToHash(s string) Hash {
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}
	var hash Hash
	bytes, _ := hex.DecodeString(s)
	if len(bytes) > 32 {
		bytes = bytes[len(bytes)-32:]
	}
	copy(hash[32-len(bytes):], bytes)
	return hash
}

// AddressToHex 地址转十六进制
func AddressToHex(addr Address) string {
	return "0x" + hex.EncodeToString(addr[:])
}

// HashToHex 哈希转十六进制
func HashToHex(hash Hash) string {
	return "0x" + hex.EncodeToString(hash[:])
}

// BigIntToHash 大整数转哈希
func BigIntToHash(val *big.Int) Hash {
	var hash Hash
	bytes := val.Bytes()
	if len(bytes) > 32 {
		bytes = bytes[len(bytes)-32:]
	}
	copy(hash[32-len(bytes):], bytes)
	return hash
}

// HashToBigInt 哈希转大整数
func HashToBigInt(hash Hash) *big.Int {
	return new(big.Int).SetBytes(hash[:])
}
