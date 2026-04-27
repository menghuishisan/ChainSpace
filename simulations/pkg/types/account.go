package types

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
)

// Address 地址
type Address [20]byte

// EmptyAddress 空地址
var EmptyAddress = Address{}

// AddressFromHex 从十六进制字符串创建地址
func AddressFromHex(s string) (Address, error) {
	var addr Address
	// 移除0x前缀
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:]
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return addr, err
	}
	copy(addr[:], b)
	return addr, nil
}

// String 返回十六进制字符串
func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

// MarshalJSON JSON序列化
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON JSON反序列化
func (a *Address) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	addr, err := AddressFromHex(s)
	if err != nil {
		return err
	}
	*a = addr
	return nil
}

// IsEmpty 是否为空地址
func (a Address) IsEmpty() bool {
	return a == EmptyAddress
}

// Account 账户（以太坊账户模型）
type Account struct {
	Address  Address  `json:"address"`
	Balance  *big.Int `json:"balance"`
	Nonce    uint64   `json:"nonce"`
	CodeHash Hash     `json:"code_hash"`
	Storage  Storage  `json:"storage"`
}

// Storage 账户存储
type Storage map[Hash]Hash

// NewAccount 创建新账户
func NewAccount(addr Address) *Account {
	return &Account{
		Address: addr,
		Balance: big.NewInt(0),
		Nonce:   0,
		Storage: make(Storage),
	}
}

// IsContract 是否为合约账户
func (a *Account) IsContract() bool {
	return !a.CodeHash.IsEmpty()
}

// AccountState 账户状态集合
type AccountState struct {
	Accounts map[Address]*Account `json:"accounts"`
}

// NewAccountState 创建账户状态
func NewAccountState() *AccountState {
	return &AccountState{
		Accounts: make(map[Address]*Account),
	}
}

// GetAccount 获取账户（不存在则创建）
func (s *AccountState) GetAccount(addr Address) *Account {
	if acc, ok := s.Accounts[addr]; ok {
		return acc
	}
	acc := NewAccount(addr)
	s.Accounts[addr] = acc
	return acc
}

// GetBalance 获取余额
func (s *AccountState) GetBalance(addr Address) *big.Int {
	return s.GetAccount(addr).Balance
}

// SetBalance 设置余额
func (s *AccountState) SetBalance(addr Address, balance *big.Int) {
	s.GetAccount(addr).Balance = balance
}

// AddBalance 增加余额
func (s *AccountState) AddBalance(addr Address, amount *big.Int) {
	acc := s.GetAccount(addr)
	acc.Balance = new(big.Int).Add(acc.Balance, amount)
}

// SubBalance 减少余额
func (s *AccountState) SubBalance(addr Address, amount *big.Int) bool {
	acc := s.GetAccount(addr)
	if acc.Balance.Cmp(amount) < 0 {
		return false
	}
	acc.Balance = new(big.Int).Sub(acc.Balance, amount)
	return true
}

// GetNonce 获取Nonce
func (s *AccountState) GetNonce(addr Address) uint64 {
	return s.GetAccount(addr).Nonce
}

// SetNonce 设置Nonce
func (s *AccountState) SetNonce(addr Address, nonce uint64) {
	s.GetAccount(addr).Nonce = nonce
}

// IncrementNonce 增加Nonce
func (s *AccountState) IncrementNonce(addr Address) {
	s.GetAccount(addr).Nonce++
}

// Transfer 转账
func (s *AccountState) Transfer(from, to Address, amount *big.Int) bool {
	if !s.SubBalance(from, amount) {
		return false
	}
	s.AddBalance(to, amount)
	return true
}

// Signature 签名
type Signature struct {
	R []byte `json:"r"`
	S []byte `json:"s"`
	V uint8  `json:"v"`
}

// IsEmpty 签名是否为空
func (s Signature) IsEmpty() bool {
	return len(s.R) == 0 && len(s.S) == 0
}

// KeyPair 密钥对
type KeyPair struct {
	PrivateKey []byte  `json:"private_key"`
	PublicKey  []byte  `json:"public_key"`
	Address    Address `json:"address"`
}

// Wallet 钱包
type Wallet struct {
	Keys    []KeyPair `json:"keys"`
	Current int       `json:"current"`
}

// NewWallet 创建钱包
func NewWallet() *Wallet {
	return &Wallet{
		Keys:    make([]KeyPair, 0),
		Current: 0,
	}
}

// CurrentKey 获取当前密钥对
func (w *Wallet) CurrentKey() *KeyPair {
	if len(w.Keys) == 0 {
		return nil
	}
	return &w.Keys[w.Current]
}
