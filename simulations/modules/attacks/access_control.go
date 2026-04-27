package attacks

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// AccessControlVuln 表示一类访问控制漏洞。
type AccessControlVuln struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Exploitable bool   `json:"exploitable"`
}

// ContractRole 表示合约中的角色与权限。
type ContractRole struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"is_active"`
}

// AccessControlAttack 记录一次访问控制攻击过程。
type AccessControlAttack struct {
	ID          string    `json:"id"`
	AttackType  string    `json:"attack_type"`
	Attacker    string    `json:"attacker"`
	Target      string    `json:"target"`
	Function    string    `json:"function"`
	Success     bool      `json:"success"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// AccessControlSimulator 演示常见的访问控制漏洞。
type AccessControlSimulator struct {
	*base.BaseSimulator
	owner       string
	roles       map[string]*ContractRole
	attacks     []*AccessControlAttack
	initialized bool
}

// NewAccessControlSimulator 创建访问控制攻击演示器。
func NewAccessControlSimulator() *AccessControlSimulator {
	sim := &AccessControlSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"access_control",
			"访问控制攻击演示器",
			"演示缺失权限校验、初始化劫持、tx.origin 误用和角色提升等访问控制漏洞。",
			"attacks",
			types.ComponentAttack,
		),
		roles:   make(map[string]*ContractRole),
		attacks: make([]*AccessControlAttack, 0),
	}

	sim.AddParam(types.Param{
		Key:         "enable_vulnerabilities",
		Name:        "启用漏洞模式",
		Description: "控制是否启用存在漏洞的演示逻辑，便于观察攻击过程。",
		Type:        types.ParamTypeBool,
		Default:     true,
	})

	return sim
}

// Init 初始化模拟器状态。
func (s *AccessControlSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.owner = "0xOwner"
	s.initialized = true
	s.roles = map[string]*ContractRole{
		"owner": {
			Name:        "owner",
			Address:     "0xOwner",
			Permissions: []string{"upgrade", "withdraw", "configure"},
			IsActive:    true,
		},
		"operator": {
			Name:        "operator",
			Address:     "0xOperator",
			Permissions: []string{"pause", "resume"},
			IsActive:    true,
		},
		"user": {
			Name:        "user",
			Address:     "0xUser",
			Permissions: []string{"deposit"},
			IsActive:    true,
		},
	}
	s.attacks = make([]*AccessControlAttack, 0)

	s.updateState()
	return nil
}

// SimulateMissingAccessControl 演示敏感函数缺失权限校验。
func (s *AccessControlSimulator) SimulateMissingAccessControl(attacker string) *AccessControlAttack {
	attack := &AccessControlAttack{
		ID:          fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:  "missing_access_control",
		Attacker:    attacker,
		Target:      "0xVault",
		Function:    "withdrawAll()",
		Success:     true,
		Description: "敏感函数没有权限校验，攻击者可以直接执行资金转移。",
		Timestamp:   time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("missing_access_control", "", "", map[string]interface{}{
		"attacker": attacker,
		"target":   attack.Target,
		"function": attack.Function,
		"summary":  attack.Description,
	})
	s.updateState()
	return attack
}

// SimulateInitializeAttack 演示初始化函数被抢先调用。
func (s *AccessControlSimulator) SimulateInitializeAttack(attacker string) *AccessControlAttack {
	attack := &AccessControlAttack{
		ID:          fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:  "initialize_hijack",
		Attacker:    attacker,
		Target:      "0xProxy",
		Function:    "initialize()",
		Success:     true,
		Description: "未受保护的初始化函数被抢先调用，攻击者获得管理员权限。",
		Timestamp:   time.Now(),
	}

	s.owner = attacker
	s.attacks = append(s.attacks, attack)
	s.EmitEvent("initialize_attack", "", "", map[string]interface{}{
		"attacker": attacker,
		"target":   attack.Target,
		"function": attack.Function,
		"summary":  attack.Description,
	})
	s.updateState()
	return attack
}

// SimulateTxOriginPhishing 演示 tx.origin 钓鱼场景。
func (s *AccessControlSimulator) SimulateTxOriginPhishing(victim, attacker string) *AccessControlAttack {
	attack := &AccessControlAttack{
		ID:          fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:  "tx_origin_phishing",
		Attacker:    attacker,
		Target:      victim,
		Function:    "executeByOrigin()",
		Success:     true,
		Description: "攻击者借助恶意中间合约诱导受害者发起交易，绕过 tx.origin 权限判断。",
		Timestamp:   time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("tx_origin_phishing", "", "", map[string]interface{}{
		"victim":   victim,
		"attacker": attacker,
		"summary":  attack.Description,
	})
	s.updateState()
	return attack
}

// SimulateRoleEscalation 演示角色提升漏洞。
func (s *AccessControlSimulator) SimulateRoleEscalation(attacker string) *AccessControlAttack {
	attack := &AccessControlAttack{
		ID:          fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:  "role_escalation",
		Attacker:    attacker,
		Target:      "0xRoleManager",
		Function:    "grantRole()",
		Success:     true,
		Description: "角色边界检查不足时，攻击者可以把自己提升为高权限角色。",
		Timestamp:   time.Now(),
	}

	s.roles["attacker"] = &ContractRole{
		Name:        "admin",
		Address:     attacker,
		Permissions: []string{"upgrade", "withdraw", "pause", "resume", "grant_role"},
		IsActive:    true,
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("role_escalation", "", "", map[string]interface{}{
		"attacker": attacker,
		"summary":  attack.Description,
	})
	s.updateState()
	return attack
}

// ShowVulnerablePatterns 返回典型脆弱模式。
func (s *AccessControlSimulator) ShowVulnerablePatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "缺失 onlyOwner / onlyRole",
			"description": "敏感函数直接暴露给任意外部地址调用。",
		},
		{
			"name":        "初始化函数无保护",
			"description": "部署后任何人都可以调用 initialize 接管合约。",
		},
		{
			"name":        "误用 tx.origin",
			"description": "把 tx.origin 当作权限判断依据，容易被中间合约钓鱼利用。",
		},
		{
			"name":        "角色边界过宽",
			"description": "授予角色时未校验目标地址是否在白名单和边界约束之内。",
		},
	}
}

// ShowDefenses 返回建议的防御方式。
func (s *AccessControlSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "基于角色的权限模型",
			"description": "按职责拆分 owner、admin、operator 等角色，减少单点超权风险。",
			"code": `bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");
function pause() external onlyRole(OPERATOR_ROLE) {}`,
		},
		{
			"name":        "初始化保护",
			"description": "使用 initializer 或部署后立即锁定初始化函数。",
			"code": `function initialize(address admin) external initializer {
    owner = admin;
}`,
		},
		{
			"name":        "多签控制",
			"description": "关键操作通过多签审批，降低单个私钥泄露的影响。",
			"code": `require(approvals[txHash] >= requiredApprovals, "Not enough approvals");`,
		},
		{
			"name":        "时间锁保护",
			"description": "为高风险操作加入等待期，给监控与治理留出干预窗口。",
			"code": `function queueTransaction(...) public onlyOwner {
    uint256 eta = block.timestamp + delay;
    queuedTransactions[txHash] = eta;
}`,
		},
	}
}

// GetRealWorldCases 返回相关真实案例。
func (s *AccessControlSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "Parity Wallet Hack #1",
			"date":   "2017-07-19",
			"loss":   "约 3000 万美元",
			"issue":  "初始化函数缺少保护",
			"detail": "攻击者调用初始化函数成为钱包管理员，随后转移资金。",
		},
		{
			"name":   "Parity Wallet Hack #2",
			"date":   "2017-11-06",
			"loss":   "约 1.5 亿美元被冻结",
			"issue":  "共享库合约被错误初始化后又被自毁",
			"detail": "攻击者获得库合约权限后调用自毁，导致依赖钱包全部失效。",
		},
		{
			"name":   "Poly Network",
			"date":   "2021-08-10",
			"loss":   "约 6.1 亿美元",
			"issue":  "跨链消息校验逻辑存在访问控制缺陷",
			"detail": "攻击者伪造高权限执行消息并转走桥内资产。",
		},
		{
			"name":   "Ronin Bridge",
			"date":   "2022-03-23",
			"loss":   "约 6.25 亿美元",
			"issue":  "验证者密钥管理薄弱且权限过于集中",
			"detail": "攻击者控制多数签名权后伪造提款消息，转移桥内资金。",
		},
	}
}

// updateState 更新前端展示状态。
func (s *AccessControlSimulator) updateState() {
	s.SetGlobalData("owner", s.owner)
	s.SetGlobalData("initialized", s.initialized)
	s.SetGlobalData("attack_count", len(s.attacks))

	roleList := make([]map[string]interface{}, 0, len(s.roles))
	for _, role := range s.roles {
		roleList = append(roleList, map[string]interface{}{
			"name":        role.Name,
			"address":     role.Address,
			"permissions": role.Permissions,
			"is_active":   role.IsActive,
		})
	}
	s.SetGlobalData("roles", roleList)

	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发访问控制攻击，可以从缺失权限校验、初始化劫持、tx.origin 钓鱼或角色提升开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待访问控制攻击场景。",
			"可以先触发一种越权路径，观察攻击者如何从入口识别逐步推进到敏感效果发生。",
			0,
			map[string]interface{}{
				"owner":       s.owner,
				"role_count":  len(roleList),
				"initialized": s.initialized,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "identify_entry",
			"caller":      latest.Attacker,
			"function":    latest.Function,
			"target":      latest.Target,
			"amount":      "0",
			"call_depth":  1,
			"description": "攻击者先锁定缺少权限边界保护的敏感入口。",
		},
		{
			"step":        2,
			"action":      latest.AttackType,
			"caller":      latest.Attacker,
			"function":    latest.Function,
			"target":      latest.Target,
			"amount":      "0",
			"call_depth":  2,
			"description": latest.Description,
		},
		{
			"step":        3,
			"action":      "privileged_effect",
			"caller":      latest.Attacker,
			"function":    latest.Function,
			"target":      latest.Target,
			"amount":      "0",
			"call_depth":  3,
			"description": "攻击者获得了本不应拥有的执行权限，敏感操作被错误放行。",
		},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest.Description)
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"privileged",
		latest.Description,
		"重点观察攻击者是如何越过权限边界，并最终触发敏感函数的。",
		1.0,
		map[string]interface{}{
			"attack_type":  latest.AttackType,
			"target":       latest.Target,
			"function":     latest.Function,
			"role_count":   len(roleList),
			"current_owner": s.owner,
		},
	)
}

// ExecuteAction 执行动作面板交互。
func (s *AccessControlSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_missing_access_control":
		attacker := actionString(params, "attacker", "0xAttacker")
		attack := s.SimulateMissingAccessControl(attacker)
		return actionResultWithFeedback(
			"已执行缺失访问控制攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "攻击者直接调用了缺少权限校验的敏感函数。",
				NextHint:    "重点观察敏感入口为何没有被角色边界拦住，以及后续越权效果如何发生。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"target":      attack.Target,
				},
			},
		), nil
	case "simulate_initialize_attack":
		attacker := actionString(params, "attacker", "0xAttacker")
		attack := s.SimulateInitializeAttack(attacker)
		return actionResultWithFeedback(
			"已执行初始化劫持攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "攻击者抢先调用初始化函数并接管管理员权限。",
				NextHint:    "重点观察初始化入口为何未被锁定，以及 owner 是如何被篡改的。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"owner":       s.owner,
				},
			},
		), nil
	case "simulate_tx_origin_phishing":
		victim := actionString(params, "victim", "0xVictim")
		attacker := actionString(params, "attacker", "0xAttacker")
		attack := s.SimulateTxOriginPhishing(victim, attacker)
		return actionResultWithFeedback(
			"已执行 tx.origin 钓鱼攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "攻击者通过恶意中间合约绕过了 tx.origin 权限判断。",
				NextHint:    "重点观察受害者地址和攻击者地址如何在错误的身份判断里被混淆。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"target":      attack.Target,
				},
			},
		), nil
	case "simulate_role_escalation":
		attacker := actionString(params, "attacker", "0xAttacker")
		attack := s.SimulateRoleEscalation(attacker)
		return actionResultWithFeedback(
			"已执行角色提升攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "攻击者已被提升为高权限角色。",
				NextHint:    "重点观察角色边界为何失效，以及新增角色权限会对哪些敏感操作产生影响。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"roles":       len(s.roles),
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported access control action: %s", action)
	}
}

// AccessControlFactory 访问控制模拟器工厂。
type AccessControlFactory struct{}

// Create 创建模拟器实例。
func (f *AccessControlFactory) Create() engine.Simulator {
	return NewAccessControlSimulator()
}

// GetDescription 返回模拟器描述。
func (f *AccessControlFactory) GetDescription() types.Description {
	return NewAccessControlSimulator().GetDescription()
}

// NewAccessControlFactory 创建工厂。
func NewAccessControlFactory() *AccessControlFactory {
	return &AccessControlFactory{}
}

var _ engine.SimulatorFactory = (*AccessControlFactory)(nil)
