package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// FakeDepositType 表示假充值攻击类型。
type FakeDepositType string

const (
	FakeDepositFakeToken  FakeDepositType = "fake_token"
	FakeDepositZeroConf   FakeDepositType = "zero_conf"
	FakeDepositRBF        FakeDepositType = "rbf"
	FakeDepositFailedTx   FakeDepositType = "failed_tx"
	FakeDepositInternalTx FakeDepositType = "internal_tx"
)

// FakeDepositAttack 记录一次假充值攻击。
type FakeDepositAttack struct {
	ID            string          `json:"id"`
	Type          FakeDepositType `json:"type"`
	TokenName     string          `json:"token_name"`
	FakeContract  string          `json:"fake_contract,omitempty"`
	RealContract  string          `json:"real_contract,omitempty"`
	Amount        *big.Int        `json:"amount"`
	Confirmations int             `json:"confirmations"`
	TxStatus      int             `json:"tx_status"`
	Success       bool            `json:"success"`
	Timestamp     time.Time       `json:"timestamp"`
}

// FakeDepositSimulator 演示交易所和跨链桥场景中的假充值问题。
type FakeDepositSimulator struct {
	*base.BaseSimulator
	attacks        []*FakeDepositAttack
	officialTokens map[string]string
	depositChecks  map[string]bool
}

// NewFakeDepositSimulator 创建假充值攻击演示器。
func NewFakeDepositSimulator() *FakeDepositSimulator {
	sim := &FakeDepositSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"fake_deposit",
			"假充值攻击演示器",
			"演示假代币充值、零确认双花和失败交易入账等错误记账问题。",
			"attacks",
			types.ComponentAttack,
		),
		attacks:        make([]*FakeDepositAttack, 0),
		officialTokens: make(map[string]string),
		depositChecks:  make(map[string]bool),
	}

	sim.AddParam(types.Param{
		Key:         "check_contract_address",
		Name:        "校验合约地址",
		Description: "控制入账前是否校验代币或资产对应的官方合约地址。",
		Type:        types.ParamTypeBool,
		Default:     true,
	})

	sim.AddParam(types.Param{
		Key:         "required_confirmations",
		Name:        "确认数要求",
		Description: "控制系统在入账前需要等待的确认数量。",
		Type:        types.ParamTypeInt,
		Default:     12,
		Min:         0,
		Max:         100,
	})

	return sim
}

// Init 初始化模拟器。
func (s *FakeDepositSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.officialTokens = map[string]string{
		"USDT": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		"USDC": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		"DAI":  "0x6B175474E89094C44Da98b954EedeAC495271d0F",
	}
	s.depositChecks = map[string]bool{
		"contract_address": true,
		"tx_status":        true,
		"confirmations":    true,
	}

	if value, ok := config.Params["check_contract_address"]; ok {
		if typed, ok := value.(bool); ok {
			s.depositChecks["contract_address"] = typed
		}
	}

	s.attacks = make([]*FakeDepositAttack, 0)
	s.updateState()
	return nil
}

// ShowAttackTypes 展示不同攻击类型。
func (s *FakeDepositSimulator) ShowAttackTypes() []map[string]interface{} {
	return []map[string]interface{}{
		{"type": "假代币攻击", "severity": "Critical", "prevention": "验证官方合约地址白名单"},
		{"type": "零确认双花", "severity": "High", "prevention": "等待足够确认数后再入账"},
		{"type": "失败交易入账", "severity": "High", "prevention": "检查 receipt.status == 1"},
	}
}

// SimulateFakeTokenAttack 模拟假代币充值。
func (s *FakeDepositSimulator) SimulateFakeTokenAttack() *FakeDepositAttack {
	attack := &FakeDepositAttack{
		ID:           fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:         FakeDepositFakeToken,
		TokenName:    "USDT",
		FakeContract: "0xFakeUSDT123456789",
		RealContract: s.officialTokens["USDT"],
		Amount:       new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e6)),
		TxStatus:     1,
		Success:      !s.depositChecks["contract_address"],
		Timestamp:    time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("fake_token_attack", "", "", map[string]interface{}{
		"token":         attack.TokenName,
		"fake_contract": attack.FakeContract,
		"real_contract": attack.RealContract,
		"success":       attack.Success,
	})
	s.updateState()
	return attack
}

// SimulateZeroConfAttack 模拟零确认双花。
func (s *FakeDepositSimulator) SimulateZeroConfAttack(confirmations int) *FakeDepositAttack {
	attack := &FakeDepositAttack{
		ID:            fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:          FakeDepositZeroConf,
		TokenName:     "BTC",
		Amount:        new(big.Int).Mul(big.NewInt(10), big.NewInt(1e8)),
		Confirmations: confirmations,
		TxStatus:      1,
		Success:       confirmations == 0,
		Timestamp:     time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("zero_conf_attack", "", "", map[string]interface{}{
		"confirmations": confirmations,
		"success":       attack.Success,
		"summary":       "如果系统在 0 确认状态下就入账，攻击者可以通过替换交易完成双花。",
	})
	s.updateState()
	return attack
}

// SimulateFailedTxAttack 模拟失败交易入账。
func (s *FakeDepositSimulator) SimulateFailedTxAttack() *FakeDepositAttack {
	attack := &FakeDepositAttack{
		ID:        fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:      FakeDepositFailedTx,
		TokenName: "USDT",
		Amount:    new(big.Int).Mul(big.NewInt(50000), big.NewInt(1e6)),
		TxStatus:  0,
		Success:   !s.depositChecks["tx_status"],
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("failed_tx_attack", "", "", map[string]interface{}{
		"tx_status": attack.TxStatus,
		"success":   attack.Success,
		"summary":   "如果系统只检查交易存在而不检查状态，失败交易也可能被错误入账。",
	})
	s.updateState()
	return attack
}

// ShowDefenses 展示防御方式。
func (s *FakeDepositSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "验证官方合约地址", "description": "所有充值资产都必须与白名单合约地址匹配。"},
		{"name": "等待足够确认数", "description": "高风险资产或大额充值需要更长确认窗口。"},
		{"name": "检查交易状态", "description": "必须校验 receipt.status == 1。"},
		{"name": "分级审计", "description": "大额充值在自动化校验外还应加入人工审核。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *FakeDepositSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "Fake EOS deposits",
			"date":   "2018-09",
			"method": "同名假代币充值",
			"detail": "部分平台只校验事件名称和 symbol，没有核对官方合约地址。",
		},
		{
			"name":   "Zero-conf BTC double spend",
			"date":   "multiple years",
			"method": "零确认入账 + RBF 替换",
			"detail": "攻击者在商家放行后，用更高费率的交易替换原始付款。",
		},
	}
}

// updateState 更新状态。
func (s *FakeDepositSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))
	s.SetGlobalData("checks_enabled", s.depositChecks)
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("victim_balance", 0)
		s.SetGlobalData("attacker_balance", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发假充值攻击，可以从假代币、零确认双花或失败交易入账开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"bridge",
			"idle",
			"等待假充值场景。",
			"可以先触发一种伪充值路径，观察充值校验链路在哪一步被绕过。",
			0,
			map[string]interface{}{
				"attack_count":    len(s.attacks),
				"checks_enabled":  s.depositChecks,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "submit_deposit_proof",
			"caller":      "attacker",
			"function":    string(latest.Type),
			"target":      "deposit_system",
			"amount":      formatBigInt(latest.Amount),
			"call_depth":  1,
			"description": "攻击者先伪造、替换或重复利用一笔看似合法的充值凭证。",
		},
		{
			"step":        2,
			"action":      "bypass_verification",
			"caller":      "deposit_system",
			"function":    "verify_transaction",
			"target":      latest.TokenName,
			"amount":      fmt.Sprintf("%d", latest.Confirmations),
			"call_depth":  2,
			"description": "如果系统没有严格校验合约地址、状态或确认数，就会把错误凭证当成真实充值。",
		},
		{
			"step":        3,
			"action":      "credit_balance",
			"caller":      "deposit_system",
			"function":    "credit_account",
			"target":      "attacker_balance",
			"amount":      formatBigInt(latest.Amount),
			"call_depth":  3,
			"description": map[bool]string{
				true:  "系统错误入账，攻击者获得了本不存在的资产。",
				false: "校验逻辑生效，错误充值未能通过。",
			}[latest.Success],
		},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", fmt.Sprintf("攻击类型 %s，系统是否成功阻止取决于充值校验链路。", latest.Type))
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))
	s.SetGlobalData("victim_balance", 0)
	s.SetGlobalData("attacker_balance", parseAmountText(formatBigInt(latest.Amount)))

	setAttackTeachingState(
		s.BaseSimulator,
		"bridge",
		map[bool]string{true: "credited", false: "blocked"}[latest.Success],
		fmt.Sprintf("充值校验流程面对 %s 场景。", latest.Type),
		"重点观察错误凭证在哪一步混入入账流程，以及哪些校验项决定了结果。",
		1.0,
		map[string]interface{}{
			"attack_type":   latest.Type,
			"token":         latest.TokenName,
			"confirmations": latest.Confirmations,
			"tx_status":     latest.TxStatus,
			"success":       latest.Success,
		},
	)
}

// ExecuteAction 执行动作。
func (s *FakeDepositSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_fake_token_attack":
		attack := s.SimulateFakeTokenAttack()
		return actionResultWithFeedback(
			"已完成假代币充值攻击演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
				"success":   attack.Success,
			},
			&types.ActionFeedback{
				Summary:     "已进入伪造代币合约地址并尝试骗过入账系统的攻击流程。",
				NextHint:    "重点观察官方合约地址校验是否开启，以及错误资产如何被误记入账户。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"token":   attack.TokenName,
					"success": attack.Success,
				},
			},
		), nil
	case "simulate_zero_conf_attack":
		confirmations := actionInt(params, "confirmations", 0)
		attack := s.SimulateZeroConfAttack(confirmations)
		return actionResultWithFeedback(
			"已完成零确认双花攻击演示。",
			map[string]interface{}{
				"attack_id":     attack.ID,
				"confirmations": confirmations,
				"success":       attack.Success,
			},
			&types.ActionFeedback{
				Summary:     "已进入在确认数不足时先入账、后替换原交易的攻击流程。",
				NextHint:    "重点观察确认数要求是否足够，以及为什么 0 确认会给双花留下空间。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"confirmations": confirmations,
					"success":       attack.Success,
				},
			},
		), nil
	case "simulate_failed_tx_attack":
		attack := s.SimulateFailedTxAttack()
		return actionResultWithFeedback(
			"已完成失败交易入账攻击演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
				"success":   attack.Success,
			},
			&types.ActionFeedback{
				Summary:     "已进入利用失败交易或回滚交易骗过入账系统的攻击流程。",
				NextHint:    "重点观察交易状态校验是否被正确执行，以及失败交易为何可能被误记账。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"tx_status": attack.TxStatus,
					"success":   attack.Success,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// FakeDepositFactory 创建工厂。
type FakeDepositFactory struct{}

// Create 创建模拟器。
func (f *FakeDepositFactory) Create() engine.Simulator {
	return NewFakeDepositSimulator()
}

// GetDescription 获取描述。
func (f *FakeDepositFactory) GetDescription() types.Description {
	return NewFakeDepositSimulator().GetDescription()
}

// NewFakeDepositFactory 创建工厂。
func NewFakeDepositFactory() *FakeDepositFactory {
	return &FakeDepositFactory{}
}

var _ engine.SimulatorFactory = (*FakeDepositFactory)(nil)
