package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// StorageSlot 描述代理合约与实现合约之间的重要存储槽位。
type StorageSlot struct {
	Slot        string `json:"slot"`
	ProxyValue  string `json:"proxy_value"`
	LogicValue  string `json:"logic_value"`
	Description string `json:"description"`
}

// DelegatecallAttack 描述一次 delegatecall 相关攻击记录。
type DelegatecallAttack struct {
	ID            string        `json:"id"`
	AttackType    string        `json:"attack_type"`
	Result        string        `json:"result"`
	StorageSlots  []StorageSlot `json:"storage_slots"`
	ExecutionFlow []string      `json:"execution_flow"`
	Timestamp     time.Time     `json:"timestamp"`
}

// DelegatecallAttackSimulator 演示存储碰撞与 Parity 类 delegatecall 风险。
type DelegatecallAttackSimulator struct {
	*base.BaseSimulator
	attacks []*DelegatecallAttack
}

// NewDelegatecallAttackSimulator 创建 delegatecall 攻击模拟器。
func NewDelegatecallAttackSimulator() *DelegatecallAttackSimulator {
	return &DelegatecallAttackSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"delegatecall_attack",
			"delegatecall 攻击演示器",
			"演示代理合约中的存储碰撞、未初始化实现合约与危险委托调用带来的安全风险。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*DelegatecallAttack, 0),
	}
}

// Init 初始化模拟器。
func (s *DelegatecallAttackSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.attacks = make([]*DelegatecallAttack, 0)
	s.updateState()
	return nil
}

// ShowVulnerablePatterns 返回典型 delegatecall 风险模式。
func (s *DelegatecallAttackSimulator) ShowVulnerablePatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"pattern":     "代理与实现存储布局不一致",
			"description": "实现合约写入 slot_0 时，可能覆盖代理合约自己的 implementation 或 admin 字段。",
		},
		{
			"pattern":     "实现合约未初始化",
			"description": "攻击者可先初始化实现合约，再利用其权限执行危险逻辑。",
		},
		{
			"pattern":     "委托调用外部不可信逻辑",
			"description": "delegatecall 会在当前合约上下文中执行外部代码，任何写操作都会落在代理存储中。",
		},
	}
}

// SimulateStorageCollision 演示存储碰撞。
func (s *DelegatecallAttackSimulator) SimulateStorageCollision() *DelegatecallAttack {
	attack := &DelegatecallAttack{
		ID:         fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType: "storage_collision",
		Result:     "实现合约写入 owner 时，意外覆盖了代理合约的重要槽位。",
		StorageSlots: []StorageSlot{
			{Slot: "slot_0", ProxyValue: "implementation", LogicValue: "owner", Description: "最危险的冲突槽位"},
			{Slot: "slot_1", ProxyValue: "admin", LogicValue: "value", Description: "管理权限也可能被误覆盖"},
		},
		ExecutionFlow: []string{
			"攻击者通过代理调用实现合约的 setOwner。",
			"delegatecall 让实现合约逻辑在代理上下文中执行。",
			"实现合约写入 slot_0，结果覆盖了代理的 implementation。",
			"代理后续调用被劫持到攻击者指定的逻辑地址。",
		},
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("delegatecall_storage_collision", "", "", map[string]interface{}{
		"result": attack.Result,
		"slots":  attack.StorageSlots,
	})
	s.updateState()
	return attack
}

// SimulateParityHack 演示 Parity 类初始化与自毁问题。
func (s *DelegatecallAttackSimulator) SimulateParityHack() *DelegatecallAttack {
	attack := &DelegatecallAttack{
		ID:         fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType: "parity_hack",
		Result:     "攻击者先接管未初始化的库合约，再执行自毁，导致依赖它的钱包彻底失效。",
		ExecutionFlow: []string{
			"钱包逻辑被拆到共享库合约。",
			"库合约没有完成初始化保护。",
			"攻击者先调用初始化函数成为 owner。",
			"随后攻击者在库合约上执行 selfdestruct。",
			"所有依赖该库的代理钱包失去可用实现逻辑。",
		},
		Timestamp: time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("delegatecall_parity_hack", "", "", map[string]interface{}{
		"result": attack.Result,
		"flow":   attack.ExecutionFlow,
	})
	s.updateState()
	return attack
}

// ShowDefenses 返回防御建议。
func (s *DelegatecallAttackSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "固定代理存储槽", "description": "使用 EIP-1967 等标准槽位，避免与实现合约业务状态冲突。"},
		{"name": "严格初始化保护", "description": "实现合约和代理都要防止重复初始化或被未授权初始化。"},
		{"name": "限制 delegatecall 目标", "description": "不要向可变且不可信的地址执行 delegatecall。"},
		{"name": "升级前审计存储布局", "description": "每次升级都应核对 slot 顺序与类型兼容性。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *DelegatecallAttackSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "Parity Multisig Wallet",
			"date":   "2017",
			"impact": "库合约被错误初始化并自毁，造成大量资金永久冻结。",
		},
		{
			"name":   "多次代理升级事故",
			"impact": "多个升级代理项目因存储布局不兼容，导致 owner、implementation 或业务状态被污染。",
		},
	}
}

func (s *DelegatecallAttackSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))

	if len(s.attacks) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发 delegatecall 攻击，请选择具体场景观察代理状态如何被污染。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待 delegatecall 攻击场景。",
			"可以先触发存储碰撞或 Parity 类场景，观察代理上下文中的危险写操作如何发生。",
			0,
			map[string]interface{}{
				"attack_count": 0,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := make([]map[string]interface{}, 0, len(latest.ExecutionFlow))
	for index, item := range latest.ExecutionFlow {
		steps = append(steps, map[string]interface{}{
			"index":       index + 1,
			"title":       fmt.Sprintf("步骤 %d", index+1),
			"description": item,
		})
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest.Result)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"delegated",
		latest.Result,
		"重点观察代理上下文为何会被外部逻辑污染，以及关键槽位在哪一步被覆盖。",
		1.0,
		map[string]interface{}{
			"attack_type":   latest.AttackType,
			"slot_count":    len(latest.StorageSlots),
			"step_count":    len(steps),
			"storage_slots": latest.StorageSlots,
		},
	)
}

// ExecuteAction 执行 delegatecall 相关动作。
func (s *DelegatecallAttackSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_storage_collision":
		attack := s.SimulateStorageCollision()
		return actionResultWithFeedback(
			"已执行存储碰撞攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入实现合约写入污染代理存储槽的攻击流程。",
				NextHint:    "重点观察 slot_0 和管理员槽位如何被错误覆盖。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"slot_count":  len(attack.StorageSlots),
				},
			},
		), nil
	case "simulate_parity_hack":
		attack := s.SimulateParityHack()
		return actionResultWithFeedback(
			"已执行 Parity 类 delegatecall 攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入未初始化库合约被接管并被销毁的攻击流程。",
				NextHint:    "重点观察初始化保护缺失如何一步步演变成整个代理体系失效。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"step_count":  len(attack.ExecutionFlow),
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// DelegatecallAttackFactory 创建 delegatecall 模拟器。
type DelegatecallAttackFactory struct{}

func (f *DelegatecallAttackFactory) Create() engine.Simulator { return NewDelegatecallAttackSimulator() }
func (f *DelegatecallAttackFactory) GetDescription() types.Description {
	return NewDelegatecallAttackSimulator().GetDescription()
}
func NewDelegatecallAttackFactory() *DelegatecallAttackFactory { return &DelegatecallAttackFactory{} }

var _ engine.SimulatorFactory = (*DelegatecallAttackFactory)(nil)
