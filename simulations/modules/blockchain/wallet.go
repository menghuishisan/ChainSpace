package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// WalletAccount 钱包账户
type WalletAccount struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Balance    uint64 `json:"balance"`
}

// WalletSimulator 钱包演示器
type WalletSimulator struct {
	*base.BaseSimulator
	accounts map[string]*WalletAccount
}

// NewWalletSimulator 创建钱包演示器
func NewWalletSimulator() *WalletSimulator {
	sim := &WalletSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"wallet",
			"钱包演示器",
			"展示密钥生成、地址派生和签名验证",
			"blockchain",
			types.ComponentTool,
		),
		accounts: make(map[string]*WalletAccount),
	}
	return sim
}

// Init 初始化
func (s *WalletSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.CreateAccount("alice")
	s.CreateAccount("bob")
	s.accounts["alice"].Balance = 1000
	s.accounts["bob"].Balance = 500
	s.updateState()
	return nil
}

// CreateAccount 创建账户
func (s *WalletSimulator) CreateAccount(name string) (*WalletAccount, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	privBytes := privateKey.D.Bytes()
	pubBytes := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	hash := sha256.Sum256(pubBytes)
	address := "0x" + hex.EncodeToString(hash[:20])

	account := &WalletAccount{
		Name:       name,
		Address:    address,
		PublicKey:  hex.EncodeToString(pubBytes),
		PrivateKey: hex.EncodeToString(privBytes),
		Balance:    0,
	}
	s.accounts[name] = account

	s.EmitEvent("account_created", "", "", map[string]interface{}{
		"name": name, "address": address,
	})
	s.updateState()
	return account, nil
}

// SignMessage 签名消息
func (s *WalletSimulator) SignMessage(accountName string, message string) (string, error) {
	account := s.accounts[accountName]
	if account == nil {
		return "", fmt.Errorf("account not found")
	}

	hash := sha256.Sum256([]byte(message))
	privBytes, _ := hex.DecodeString(account.PrivateKey)

	// 简化签名
	sigData := append(hash[:], privBytes...)
	sigHash := sha256.Sum256(sigData)
	signature := hex.EncodeToString(sigHash[:])

	s.EmitEvent("message_signed", "", "", map[string]interface{}{
		"account": accountName, "message_hash": hex.EncodeToString(hash[:16]),
	})
	return signature, nil
}

// VerifySignature 验证签名
func (s *WalletSimulator) VerifySignature(accountName string, message string, signature string) bool {
	account := s.accounts[accountName]
	if account == nil {
		return false
	}

	hash := sha256.Sum256([]byte(message))
	privBytes, _ := hex.DecodeString(account.PrivateKey)
	sigData := append(hash[:], privBytes...)
	sigHash := sha256.Sum256(sigData)
	expected := hex.EncodeToString(sigHash[:])

	valid := signature == expected
	s.EmitEvent("signature_verified", "", "", map[string]interface{}{
		"account": accountName, "valid": valid,
	})
	return valid
}

// Transfer 转账
func (s *WalletSimulator) Transfer(from, to string, amount uint64) error {
	fromAcc := s.accounts[from]
	toAcc := s.accounts[to]
	if fromAcc == nil || toAcc == nil {
		return fmt.Errorf("account not found")
	}
	if fromAcc.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	fromAcc.Balance -= amount
	toAcc.Balance += amount

	s.EmitEvent("transfer", "", "", map[string]interface{}{
		"from": from, "to": to, "amount": amount,
	})
	s.updateState()
	return nil
}

func (s *WalletSimulator) updateState() {
	accountList := make([]map[string]interface{}, 0)
	for _, acc := range s.accounts {
		accountList = append(accountList, map[string]interface{}{
			"name": acc.Name, "address": acc.Address, "balance": acc.Balance,
		})
	}
	s.SetGlobalData("accounts", accountList)
	s.SetGlobalData("account_count", len(s.accounts))

	summary := fmt.Sprintf("当前钱包中共有 %d 个账户。", len(s.accounts))
	nextHint := "可以继续创建账户、签名消息或发起转账，观察地址与余额如何变化。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备钱包",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"account_count": len(s.accounts)},
	)
}

func (s *WalletSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "create_account":
		name := fmt.Sprintf("account-%d", len(s.accounts)+1)
		if raw, ok := params["name"].(string); ok && raw != "" {
			name = raw
		}
		account, err := s.CreateAccount(name)
		if err != nil {
			return nil, err
		}
		return blockchainActionResult("已创建一个钱包账户。", map[string]interface{}{"account": account}, &types.ActionFeedback{
			Summary:     "新账户已经生成，包含地址、公钥和私钥信息。",
			NextHint:    "继续执行消息签名或转账，观察地址与余额如何参与交易。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"name": account.Name, "address": account.Address},
		}), nil
	case "transfer":
		from, to := "alice", "bob"
		amount := uint64(10)
		if raw, ok := params["from"].(string); ok && raw != "" {
			from = raw
		}
		if raw, ok := params["to"].(string); ok && raw != "" {
			to = raw
		}
		if raw, ok := params["amount"].(float64); ok && raw > 0 {
			amount = uint64(raw)
		}
		if err := s.Transfer(from, to, amount); err != nil {
			return nil, err
		}
		return blockchainActionResult("已完成一次钱包转账。", map[string]interface{}{"from": from, "to": to, "amount": amount}, &types.ActionFeedback{
			Summary:     "转账双方的余额已经更新。",
			NextHint:    "继续对比转账前后的账户余额变化，理解地址和签名在钱包中的作用。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"from": from, "to": to, "amount": amount},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported wallet action: %s", action)
	}
}

type WalletFactory struct{}

func (f *WalletFactory) Create() engine.Simulator { return NewWalletSimulator() }
func (f *WalletFactory) GetDescription() types.Description {
	return NewWalletSimulator().GetDescription()
}
func NewWalletFactory() *WalletFactory { return &WalletFactory{} }

var _ engine.SimulatorFactory = (*WalletFactory)(nil)
