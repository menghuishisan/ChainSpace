package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// BridgeAttackRecord 描述一次跨链桥攻击演示结果。
type BridgeAttackRecord struct {
	ID           string    `json:"id"`
	AttackType   string    `json:"attack_type"`
	AttackVector string    `json:"attack_vector"`
	Method       string    `json:"method"`
	Impact       string    `json:"impact"`
	Timestamp    time.Time `json:"timestamp"`
}

// BridgeAttackSimulator 演示跨链桥中的多签委托、签名绕过与初始化漏洞。
type BridgeAttackSimulator struct {
	*base.BaseSimulator
	attacks []*BridgeAttackRecord
}

// NewBridgeAttackSimulator 创建跨链桥攻击模拟器。
func NewBridgeAttackSimulator() *BridgeAttackSimulator {
	return &BridgeAttackSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"bridge_attack",
			"跨链桥攻击演示器",
			"演示跨链桥中的验证者委托、签名校验绕过和初始化配置错误等典型攻击面。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*BridgeAttackRecord, 0),
	}
}

// Init 初始化模拟器。
func (s *BridgeAttackSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.attacks = make([]*BridgeAttackRecord, 0)
	s.updateState()
	return nil
}

// SimulateMultisigCompromise 演示多签验证者被接管。
func (s *BridgeAttackSimulator) SimulateMultisigCompromise() *BridgeAttackRecord {
	record := &BridgeAttackRecord{
		ID:           fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:   "multisig_compromise",
		AttackVector: "验证者私钥失陷",
		Method:       "攻击者控制了足够多的验证者签名，可以伪造跨链消息并直接提走桥内资产。",
		Impact:       "源链锁定与目标链铸造失去对应关系，桥内资金被一次性抽空。",
		Timestamp:    time.Now(),
	}
	s.attacks = append(s.attacks, record)
	s.SetGlobalData("latest_bridge_attack", record)
	s.EmitEvent("bridge_multisig_compromise", "", "", map[string]interface{}{
		"attack": record,
	})
	s.updateState()
	return record
}

// SimulateSignatureBypass 演示签名或消息验证逻辑被绕过。
func (s *BridgeAttackSimulator) SimulateSignatureBypass() *BridgeAttackRecord {
	record := &BridgeAttackRecord{
		ID:           fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:   "signature_bypass",
		AttackVector: "验证逻辑缺陷",
		Method:       "攻击者构造出看似合法但未被正确验证的证明或签名集合。",
		Impact:       "桥合约错误接受伪造消息，在目标链铸造或释放本不该存在的资产。",
		Timestamp:    time.Now(),
	}
	s.attacks = append(s.attacks, record)
	s.SetGlobalData("latest_bridge_attack", record)
	s.EmitEvent("bridge_signature_bypass", "", "", map[string]interface{}{
		"attack": record,
	})
	s.updateState()
	return record
}

// SimulateInitializationAttack 演示初始化配置错误。
func (s *BridgeAttackSimulator) SimulateInitializationAttack() *BridgeAttackRecord {
	record := &BridgeAttackRecord{
		ID:           fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:   "initialization_attack",
		AttackVector: "初始化状态错误",
		Method:       "攻击者在桥合约未正确初始化时，抢先设置关键验证参数或管理员地址。",
		Impact:       "后续所有跨链消息校验都建立在错误信任根上，桥安全边界整体失效。",
		Timestamp:    time.Now(),
	}
	s.attacks = append(s.attacks, record)
	s.SetGlobalData("latest_bridge_attack", record)
	s.EmitEvent("bridge_initialization_attack", "", "", map[string]interface{}{
		"attack": record,
	})
	s.updateState()
	return record
}

// ShowDefenses 返回跨链桥防御策略。
func (s *BridgeAttackSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "分散验证者与硬件签名", "description": "降低多签密钥一次性泄露的概率。"},
		{"name": "严格验证消息结构", "description": "对签名集合、消息根和链 ID 做完整校验。"},
		{"name": "初始化锁定", "description": "部署后立即完成初始化并禁止重复初始化。"},
		{"name": "限额与暂停开关", "description": "即使发生攻击，也能缩小单次损失规模。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *BridgeAttackSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Ronin Bridge", "impact": "验证者密钥妥协导致桥资产被大规模转走。"},
		{"name": "Wormhole", "impact": "消息验证缺陷导致目标链被错误铸造大量资产。"},
	}
}

func (s *BridgeAttackSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发跨链桥攻击，可以从验证者妥协、签名绕过或初始化漏洞开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"bridge",
			"idle",
			"等待桥攻击场景。",
			"可以先触发一种桥攻击，观察源链、桥验证层和目标链执行如何被串联破坏。",
			0,
			map[string]interface{}{
				"attack_count": 0,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "source_chain_request",
			"caller":      "source_chain",
			"function":    "lock_or_message",
			"target":      "bridge",
			"amount":      "1",
			"call_depth":  1,
			"description": "源链先发起锁定或跨链消息，正常情况下应由桥验证层严格校验。",
		},
		{
			"step":        2,
			"action":      latest.AttackType,
			"caller":      "attacker",
			"function":    latest.AttackVector,
			"target":      "bridge_verifier",
			"amount":      "1",
			"call_depth":  2,
			"description": latest.Method,
		},
		{
			"step":        3,
			"action":      "target_chain_effect",
			"caller":      "bridge",
			"function":    "mint_or_release",
			"target":      "target_chain",
			"amount":      "1",
			"call_depth":  3,
			"description": latest.Impact,
		},
	}

	s.SetGlobalData("latest_bridge_attack", latest)
	s.SetGlobalData("attack_summary", latest.Impact)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"bridge",
		"breached",
		latest.Impact,
		"重点观察攻击发生在桥验证生命周期的哪一步，以及目标链错误执行是如何产生的。",
		1.0,
		map[string]interface{}{
			"attack_type":   latest.AttackType,
			"attack_vector": latest.AttackVector,
			"step_count":    len(steps),
		},
	)
}

// ExecuteAction 执行跨链桥攻击动作。
func (s *BridgeAttackSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_multisig_compromise":
		record := s.SimulateMultisigCompromise()
		return actionResultWithFeedback(
			"已执行跨链桥多签妥协演示。",
			map[string]interface{}{"attack": record},
			&types.ActionFeedback{
				Summary:     "已进入验证者密钥失陷导致的桥验证失效流程。",
				NextHint:    "重点观察桥验证层何时失去信任，以及目标链何时出现错误资产释放。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"attack_type":   record.AttackType,
					"attack_vector": record.AttackVector,
				},
			},
		), nil
	case "simulate_signature_bypass":
		record := s.SimulateSignatureBypass()
		return actionResultWithFeedback(
			"已执行跨链桥签名绕过演示。",
			map[string]interface{}{"attack": record},
			&types.ActionFeedback{
				Summary:     "已进入伪造消息通过桥验证的攻击流程。",
				NextHint:    "重点观察错误消息如何穿过桥验证层，并在目标链触发错误铸造或释放。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"attack_type":   record.AttackType,
					"attack_vector": record.AttackVector,
				},
			},
		), nil
	case "simulate_initialization_attack":
		record := s.SimulateInitializationAttack()
		return actionResultWithFeedback(
			"已执行跨链桥初始化漏洞演示。",
			map[string]interface{}{"attack": record},
			&types.ActionFeedback{
				Summary:     "已进入错误初始化导致桥信任根失效的攻击流程。",
				NextHint:    "重点观察攻击者如何通过抢占初始化权限改变后续所有跨链校验结果。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"attack_type":   record.AttackType,
					"attack_vector": record.AttackVector,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// BridgeAttackFactory 创建跨链桥攻击模拟器。
type BridgeAttackFactory struct{}

func (f *BridgeAttackFactory) Create() engine.Simulator { return NewBridgeAttackSimulator() }
func (f *BridgeAttackFactory) GetDescription() types.Description { return NewBridgeAttackSimulator().GetDescription() }
func NewBridgeAttackFactory() *BridgeAttackFactory { return &BridgeAttackFactory{} }

var _ engine.SimulatorFactory = (*BridgeAttackFactory)(nil)
