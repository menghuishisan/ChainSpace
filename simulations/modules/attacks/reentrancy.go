package attacks

import (
	"fmt"
	"math/big"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// ContractState 表示受害合约的账户状态。
type ContractState struct {
	Balance  *big.Int            `json:"balance"`
	Balances map[string]*big.Int `json:"balances"`
}

// AttackStep 描述一次重入攻击过程中的单个步骤。
type AttackStep struct {
	Step        int    `json:"step"`
	Action      string `json:"action"`
	Caller      string `json:"caller"`
	Function    string `json:"function"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	CallDepth   int    `json:"call_depth"`
}

// ReentrancySimulator 演示重入攻击和安全提取流程。
type ReentrancySimulator struct {
	*base.BaseSimulator
	victimContract  *ContractState
	attackerBalance *big.Int
	attackSteps     []*AttackStep
	callDepth       int
	maxReentrancy   int
}

// NewReentrancySimulator 创建重入攻击模拟器。
func NewReentrancySimulator() *ReentrancySimulator {
	sim := &ReentrancySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"reentrancy",
			"重入攻击演示器",
			"演示以太坊合约中的重入攻击路径，并对比漏洞流程与 Checks-Effects-Interactions 安全流程。",
			"attacks",
			types.ComponentAttack,
		),
		attackSteps: make([]*AttackStep, 0),
	}

	sim.AddParam(types.Param{
		Key:         "victim_balance",
		Name:        "初始合约余额",
		Description: "受害合约在攻击前持有的 ETH 数量。",
		Type:        types.ParamTypeInt,
		Default:     10,
		Min:         1,
		Max:         100,
	})
	sim.AddParam(types.Param{
		Key:         "attacker_deposit",
		Name:        "攻击者初始存款",
		Description: "攻击者预先存入受害合约的 ETH 数量。",
		Type:        types.ParamTypeInt,
		Default:     1,
		Min:         1,
		Max:         10,
	})
	sim.AddParam(types.Param{
		Key:         "max_reentrancy",
		Name:        "最大重入层数",
		Description: "限制演示中最多发生多少层重入调用。",
		Type:        types.ParamTypeInt,
		Default:     5,
		Min:         1,
		Max:         20,
	})

	return sim
}

// Init 初始化合约余额和攻击参数。
func (s *ReentrancySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	victimBalance := int64(10)
	attackerDeposit := int64(1)
	s.maxReentrancy = 5

	if v, ok := config.Params["victim_balance"]; ok {
		if n, ok := v.(float64); ok {
			victimBalance = int64(n)
		}
	}
	if v, ok := config.Params["attacker_deposit"]; ok {
		if n, ok := v.(float64); ok {
			attackerDeposit = int64(n)
		}
	}
	if v, ok := config.Params["max_reentrancy"]; ok {
		if n, ok := v.(float64); ok {
			s.maxReentrancy = int(n)
		}
	}

	s.victimContract = &ContractState{
		Balance:  big.NewInt(victimBalance),
		Balances: make(map[string]*big.Int),
	}
	s.victimContract.Balances["attacker"] = big.NewInt(attackerDeposit)
	s.victimContract.Balance.Add(s.victimContract.Balance, big.NewInt(attackerDeposit))
	s.attackerBalance = big.NewInt(0)
	s.attackSteps = make([]*AttackStep, 0)
	s.callDepth = 0

	s.updateState()
	return nil
}

// SimulateVulnerableWithdraw 执行带漏洞的提取流程。
func (s *ReentrancySimulator) SimulateVulnerableWithdraw() {
	s.attackSteps = make([]*AttackStep, 0)
	s.callDepth = 0

	s.EmitEvent("attack_started", "", "", map[string]interface{}{
		"victim_balance":   s.victimContract.Balance.String(),
		"attacker_deposit": s.victimContract.Balances["attacker"].String(),
	})

	s.vulnerableWithdraw("attacker", new(big.Int).Set(s.victimContract.Balances["attacker"]))

	s.EmitEvent("attack_completed", "", "", map[string]interface{}{
		"victim_balance":   s.victimContract.Balance.String(),
		"attacker_gained":  s.attackerBalance.String(),
		"reentrancy_count": s.callDepth,
	})

	s.updateState()
}

// vulnerableWithdraw 模拟“先转账、后更新余额”的错误提取逻辑。
func (s *ReentrancySimulator) vulnerableWithdraw(caller string, amount *big.Int) {
	s.callDepth++

	s.attackSteps = append(s.attackSteps, &AttackStep{
		Step:        len(s.attackSteps) + 1,
		Action:      "call",
		Caller:      caller,
		Function:    "withdraw",
		Amount:      amount.String(),
		Description: fmt.Sprintf("第 %d 层调用进入 withdraw。", s.callDepth),
		CallDepth:   s.callDepth,
	})

	userBalance := s.victimContract.Balances[caller]
	if userBalance == nil || userBalance.Cmp(amount) < 0 {
		s.attackSteps = append(s.attackSteps, &AttackStep{
			Step:        len(s.attackSteps) + 1,
			Action:      "revert",
			Description: "用户余额不足，调用回退。",
			CallDepth:   s.callDepth,
		})
		return
	}

	if s.victimContract.Balance.Cmp(amount) < 0 {
		s.attackSteps = append(s.attackSteps, &AttackStep{
			Step:        len(s.attackSteps) + 1,
			Action:      "revert",
			Description: "受害合约余额不足，无法继续转账。",
			CallDepth:   s.callDepth,
		})
		return
	}

	s.attackSteps = append(s.attackSteps, &AttackStep{
		Step:        len(s.attackSteps) + 1,
		Action:      "transfer",
		Caller:      "victim_contract",
		Function:    "call.value",
		Amount:      amount.String(),
		Description: "受害合约先向攻击者转账，此时内部状态尚未更新。",
		CallDepth:   s.callDepth,
	})

	s.victimContract.Balance.Sub(s.victimContract.Balance, amount)
	s.attackerBalance.Add(s.attackerBalance, amount)

	if s.callDepth < s.maxReentrancy && s.victimContract.Balance.Cmp(amount) >= 0 {
		s.attackSteps = append(s.attackSteps, &AttackStep{
			Step:        len(s.attackSteps) + 1,
			Action:      "reentrancy",
			Caller:      caller,
			Function:    "receive -> withdraw",
			Amount:      amount.String(),
			Description: "攻击者在 receive 回调中再次调用 withdraw，形成重入。",
			CallDepth:   s.callDepth,
		})
		s.vulnerableWithdraw(caller, amount)
	}

	s.attackSteps = append(s.attackSteps, &AttackStep{
		Step:        len(s.attackSteps) + 1,
		Action:      "update_state",
		Caller:      "victim_contract",
		Function:    "balances[msg.sender] -= amount",
		Amount:      amount.String(),
		Description: "最后才更新攻击者余额，导致前面的重入已经多次提取资金。",
		CallDepth:   s.callDepth,
	})
	s.victimContract.Balances[caller].Sub(s.victimContract.Balances[caller], amount)
}

// SimulateSecureWithdraw 执行 Checks-Effects-Interactions 安全流程。
func (s *ReentrancySimulator) SimulateSecureWithdraw() {
	_ = s.Init(types.Config{Params: map[string]interface{}{}})
	s.attackSteps = make([]*AttackStep, 0)

	amount := new(big.Int).Set(s.victimContract.Balances["attacker"])

	s.attackSteps = append(s.attackSteps, &AttackStep{
		Step:        1,
		Action:      "check",
		Caller:      "attacker",
		Function:    "withdraw",
		Amount:      amount.String(),
		Description: "先检查攻击者是否有足够余额。",
		CallDepth:   1,
	})
	s.attackSteps = append(s.attackSteps, &AttackStep{
		Step:        2,
		Action:      "update_state",
		Caller:      "victim_contract",
		Function:    "balances[msg.sender] -= amount",
		Amount:      amount.String(),
		Description: "先更新内部余额，再执行外部交互。",
		CallDepth:   1,
	})
	s.victimContract.Balances["attacker"].Sub(s.victimContract.Balances["attacker"], amount)

	s.attackSteps = append(s.attackSteps, &AttackStep{
		Step:        3,
		Action:      "transfer",
		Caller:      "victim_contract",
		Function:    "transfer",
		Amount:      amount.String(),
		Description: "最后向攻击者转账，避免在回调里重复提取。",
		CallDepth:   1,
	})
	s.victimContract.Balance.Sub(s.victimContract.Balance, amount)
	s.attackerBalance.Add(s.attackerBalance, amount)

	s.attackSteps = append(s.attackSteps, &AttackStep{
		Step:        4,
		Action:      "complete",
		Caller:      "victim_contract",
		Function:    "finish",
		Amount:      amount.String(),
		Description: "安全提取完成，即使攻击者回调也无法再次提取资金。",
		CallDepth:   1,
	})

	s.EmitEvent("secure_withdraw_completed", "", "", map[string]interface{}{
		"pattern": "Checks-Effects-Interactions",
	})

	s.updateState()
}

// GetDefenses 返回典型防御措施。
func (s *ReentrancySimulator) GetDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "Checks-Effects-Interactions 模式",
			"description": "先检查条件，再更新状态，最后进行外部交互。",
			"code": `function withdraw(uint amount) {
    require(balances[msg.sender] >= amount);
    balances[msg.sender] -= amount;
    msg.sender.call{value: amount}("");
}`,
		},
		{
			"name":        "ReentrancyGuard",
			"description": "通过互斥锁阻止同一函数在一次执行流程中被重复进入。",
			"code": `bool private locked;
modifier nonReentrant() {
    require(!locked, "Reentrant call");
    locked = true;
    _;
    locked = false;
}`,
		},
		{
			"name":        "Pull Payment 模式",
			"description": "把提取行为交给用户主动领取，避免在业务流程中直接向外部地址转账。",
			"code": `mapping(address => uint) public pendingWithdrawals;

function withdraw() {
    uint amount = pendingWithdrawals[msg.sender];
    pendingWithdrawals[msg.sender] = 0;
    msg.sender.transfer(amount);
}`,
		},
	}
}

// GetRealWorldCases 返回真实世界案例。
func (s *ReentrancySimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "The DAO Hack",
			"date":   "2016-06-17",
			"loss":   "约 360 万 ETH",
			"impact": "导致以太坊社区发生硬分叉，并分化出 ETH 与 ETC。",
		},
		{
			"name":   "Cream Finance",
			"date":   "2021-08-30",
			"loss":   "约 1880 万美元",
			"impact": "攻击者利用回调和价格操纵路径组合获利，暴露了复杂协议中的重入风险。",
		},
		{
			"name":   "Siren Protocol",
			"date":   "2021-09-03",
			"loss":   "约 350 万美元",
			"impact": "说明即使是较新的 DeFi 协议，若状态更新顺序错误，仍然会被重入利用。",
		},
	}
}

// updateState 同步前端主舞台所需状态。
func (s *ReentrancySimulator) updateState() {
	s.SetGlobalData("victim_balance", s.victimContract.Balance.String())
	s.SetGlobalData("attacker_balance", s.attackerBalance.String())
	s.SetGlobalData("step_count", len(s.attackSteps))
	s.SetGlobalData("max_depth", s.callDepth)
	s.SetGlobalData("steps", s.attackSteps)

	summary := "当前尚未触发重入攻击，可以先观察漏洞流程与安全流程的差异。"
	progress := 0.0
	stage := "idle"

	if len(s.attackSteps) > 0 {
		stage = "executing"
		progress = 1.0
		summary = fmt.Sprintf("共记录 %d 个攻击步骤，最大重入深度为 %d。", len(s.attackSteps), s.callDepth)
	}

	s.SetGlobalData("attack_summary", summary)
	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		stage,
		summary,
		"重点观察外部转账发生在状态更新之前，重入因此得以持续发生。",
		progress,
		map[string]interface{}{
			"victim_balance":   s.victimContract.Balance.String(),
			"attacker_balance": s.attackerBalance.String(),
			"step_count":       len(s.attackSteps),
			"max_depth":        s.callDepth,
		},
	)
}

// ReentrancyFactory 创建重入攻击模拟器。
type ReentrancyFactory struct{}

func (f *ReentrancyFactory) Create() engine.Simulator {
	return NewReentrancySimulator()
}

func (f *ReentrancyFactory) GetDescription() types.Description {
	return NewReentrancySimulator().GetDescription()
}

func NewReentrancyFactory() *ReentrancyFactory {
	return &ReentrancyFactory{}
}

var _ engine.SimulatorFactory = (*ReentrancyFactory)(nil)

// ExecuteAction 执行前端触发的攻击动作。
func (s *ReentrancySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "start_attack":
		s.SimulateVulnerableWithdraw()
		return actionResultWithFeedback(
			"已执行重入攻击演示。",
			map[string]interface{}{
				"step_count": len(s.attackSteps),
				"max_depth":  s.callDepth,
			},
			&types.ActionFeedback{
				Summary:     "漏洞流程已启动，受害合约在更新余额前先向攻击者转账。",
				NextHint:    "观察回调如何再次进入 withdraw，以及受害合约余额如何被连续提取。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"victim_balance":   s.victimContract.Balance.String(),
					"attacker_balance": s.attackerBalance.String(),
				},
			},
		), nil
	case "show_secure_flow":
		s.SimulateSecureWithdraw()
		return actionResultWithFeedback(
			"已切换到安全提取流程。",
			map[string]interface{}{
				"step_count": len(s.attackSteps),
			},
			&types.ActionFeedback{
				Summary:     "已切换为 Checks-Effects-Interactions 安全流程。",
				NextHint:    "重点观察状态更新先于转账发生，因此攻击者无法再次提取资金。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"victim_balance":   s.victimContract.Balance.String(),
					"attacker_balance": s.attackerBalance.String(),
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported reentrancy action: %s", action)
	}
}
