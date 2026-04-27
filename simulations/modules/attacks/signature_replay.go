package attacks

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// Signature 表示一份可被重放利用的签名数据。
type Signature struct {
	R       string `json:"r"`
	S       string `json:"s"`
	V       uint8  `json:"v"`
	Message string `json:"message"`
	Signer  string `json:"signer"`
	Used    bool   `json:"used"`
}

// ReplayAttack 描述一次签名重放攻击。
type ReplayAttack struct {
	ID          string     `json:"id"`
	AttackType  string     `json:"attack_type"`
	OriginalTx  string     `json:"original_tx"`
	ReplayedOn  string     `json:"replayed_on"`
	Signature   *Signature `json:"signature"`
	Success     bool       `json:"success"`
	PreventedBy string     `json:"prevented_by,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}

// SignatureReplaySimulator 演示同链与跨链签名重放风险。
type SignatureReplaySimulator struct {
	*base.BaseSimulator
	signatures map[string]*Signature
	nonces     map[string]uint64
	attacks    []*ReplayAttack
}

// NewSignatureReplaySimulator 创建签名重放模拟器。
func NewSignatureReplaySimulator() *SignatureReplaySimulator {
	sim := &SignatureReplaySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"signature_replay",
			"签名重放攻击演示器",
			"演示同链与跨链签名重放风险，帮助理解 nonce、domain separator 与链 ID 在防御中的作用。",
			"attacks",
			types.ComponentAttack,
		),
		signatures: make(map[string]*Signature),
		nonces:     make(map[string]uint64),
		attacks:    make([]*ReplayAttack, 0),
	}
	return sim
}

// Init 重置历史攻击记录。
func (s *SignatureReplaySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.signatures = make(map[string]*Signature)
	s.nonces = make(map[string]uint64)
	s.attacks = make([]*ReplayAttack, 0)
	s.updateState()
	return nil
}

// SimulateSameChainReplay 演示同链历史签名被重复利用。
func (s *SignatureReplaySimulator) SimulateSameChainReplay(signer string) *ReplayAttack {
	sig := &Signature{
		R:       "0x" + hex.EncodeToString(make([]byte, 32)),
		S:       "0x" + hex.EncodeToString(make([]byte, 32)),
		V:       27,
		Message: "transfer(0xRecipient, 100 ETH)",
		Signer:  signer,
		Used:    true,
	}
	s.signatures["original"] = sig

	attack := &ReplayAttack{
		ID:         fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType: "same_chain_replay",
		OriginalTx: "Transfer 100 ETH",
		ReplayedOn: "同一条链上的历史交易接口",
		Signature:  sig,
		Success:    true,
		Timestamp:  time.Now(),
	}
	s.attacks = append(s.attacks, attack)

	s.EmitEvent("same_chain_replay", "", "", map[string]interface{}{
		"vulnerable_code": `function executeWithSignature(
    address to,
    uint256 amount,
    bytes calldata signature
) external {
    bytes32 hash = keccak256(abi.encode(to, amount));
    address signer = ECDSA.recover(hash, signature);
    require(signer == owner, "Invalid signature");

    // 没有 nonce、防重放标记和过期时间
    token.transfer(to, amount);
}`,
		"attack": "攻击者重放同一份历史签名，再次触发转账。",
	})

	s.updateState()
	return attack
}

// SimulateCrossChainReplay 演示不同链之间重放签名。
func (s *SignatureReplaySimulator) SimulateCrossChainReplay() *ReplayAttack {
	attack := &ReplayAttack{
		ID:         fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType: "cross_chain_replay",
		OriginalTx: "Ethereum 主网上已执行交易",
		ReplayedOn: "Ethereum Classic / BSC 等兼容链",
		Success:    true,
		Timestamp:  time.Now(),
	}
	s.attacks = append(s.attacks, attack)

	s.EmitEvent("cross_chain_replay", "", "", map[string]interface{}{
		"scenario": "兼容链之间缺少链 ID 绑定的重放风险",
		"attack": []string{
			"1. 用户在源链签署一笔合法交易。",
			"2. 攻击者获取交易签名和参数。",
			"3. 目标链若未绑定 chainId，则会接受同样的签名。",
			"4. 同一份签名在另一条链上再次生效。",
		},
		"eip155_solution": "EIP-155 通过 chainId 将签名绑定到具体链环境。",
	})

	s.updateState()
	return attack
}

// ShowVulnerablePatterns 返回典型重放漏洞模式。
func (s *SignatureReplaySimulator) ShowVulnerablePatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"pattern":    "缺少 nonce",
			"issue":      "同一份签名可以被重复利用。",
			"vulnerable": `hash = keccak256(to, amount)`,
			"fixed":      `hash = keccak256(to, amount, nonce++)`,
		},
		{
			"pattern": "未绑定 chainId",
			"issue":   "签名可能在其他兼容链上被重放。",
			"fixed":   "使用 EIP-712 / EIP-155 将签名绑定到链环境。",
		},
		{
			"pattern": "未绑定合约地址",
			"issue":   "签名可在不同部署实例中被复用。",
			"fixed":   `hash = keccak256(to, amount, nonce, address(this))`,
		},
	}
}

// ShowDefenses 返回推荐防御方案。
func (s *SignatureReplaySimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name": "显式 nonce",
			"code": `mapping(address => uint256) public nonces;

function execute(..., bytes calldata sig) external {
    bytes32 hash = keccak256(abi.encode(
        to, amount, nonces[signer]++
    ));
    // verify signature
}`,
		},
		{
			"name":        "EIP-712 Domain Separator",
			"description": "将链 ID、合约地址与版本信息纳入签名域，避免跨链或跨合约重放。",
			"code": `bytes32 DOMAIN_SEPARATOR = keccak256(abi.encode(
    keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
    keccak256("MyContract"),
    keccak256("1"),
    block.chainid,
    address(this)
));`,
		},
		{
			"name": "签名使用标记",
			"code": `mapping(bytes32 => bool) public usedSignatures;

require(!usedSignatures[sigHash], "Signature used");
usedSignatures[sigHash] = true;`,
		},
		{
			"name": "设置过期时间",
			"code": `require(block.timestamp <= deadline, "Expired");`,
		},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *SignatureReplaySimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "DAO 分叉后的跨链重放问题",
			"date":   "2016-07",
			"impact": "ETH / ETC 分叉后，一段时间内相同签名可能在两条链上都有效。",
			"result": "社区最终引入 EIP-155 强化链级别的签名隔离。",
		},
		{
			"name":  "Wintermute 相关签名环境配置问题",
			"date":  "2022",
			"issue": "说明签名域、地址空间和部署环境绑定不严时，仍然可能被利用。",
		},
	}
}

// ExecuteAction 执行前端动作。
func (s *SignatureReplaySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_same_chain_replay":
		signer := actionString(params, "signer", "0xSigner")
		attack := s.SimulateSameChainReplay(signer)
		return actionResultWithFeedback(
			"已执行同链签名重放演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
				"signer":    signer,
			},
			&types.ActionFeedback{
				Summary:     "已进入复用历史签名再次触发敏感操作的攻击流程。",
				NextHint:    "重点观察签名是否绑定了 nonce、过期时间和使用标记。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"replayed_on": attack.ReplayedOn,
				},
			},
		), nil
	case "simulate_cross_chain_replay":
		attack := s.SimulateCrossChainReplay()
		return actionResultWithFeedback(
			"已执行跨链签名重放演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
			},
			&types.ActionFeedback{
				Summary:     "已进入在不同兼容链之间重复利用同一份签名的攻击流程。",
				NextHint:    "重点观察 chainId 和 domain separator 是否正确绑定到了签名上下文。",
				EffectScope: "bridge",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"replayed_on": attack.ReplayedOn,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// updateState 同步核心状态。
func (s *SignatureReplaySimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))
	s.SetGlobalData("signature_count", len(s.signatures))
	s.SetGlobalData("attacks", s.attacks)
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发签名重放攻击，可以观察历史签名如何在同链或跨链再次生效。")
		setAttackTeachingState(
			s.BaseSimulator,
			"bridge",
			"idle",
			"等待签名重放场景。",
			"可以先触发一次同链或跨链重放，观察签名上下文中缺失了哪些关键绑定信息。",
			0,
			map[string]interface{}{
				"signature_count": len(s.signatures),
				"attack_count":    len(s.attacks),
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "capture_signature",
			"caller":      "attacker",
			"function":    "read_signed_payload",
			"target":      latest.OriginalTx,
			"amount":      "1",
			"call_depth":  1,
			"description": "攻击者先拿到一份原本只应被使用一次的签名载荷。",
		},
		{
			"step":        2,
			"action":      latest.AttackType,
			"caller":      "attacker",
			"function":    "replay_signature",
			"target":      latest.ReplayedOn,
			"amount":      "1",
			"call_depth":  2,
			"description": "如果合约没有把 nonce、chainId 或 domain 正确绑定，历史签名就可能再次生效。",
		},
		{
			"step":        3,
			"action":      "observe_result",
			"caller":      "contract",
			"function":    "verify_and_execute",
			"target":      latest.ReplayedOn,
			"amount":      "1",
			"call_depth":  3,
			"description": map[bool]string{
				true:  "重放成功，旧签名再次驱动交易或权限操作。",
				false: "重放被阻止，说明签名已经和环境或使用次数正确绑定。",
			}[latest.Success],
		},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", fmt.Sprintf("重放场景：%s。", latest.ReplayedOn))
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"bridge",
		map[bool]string{true: "replayed", false: "rejected"}[latest.Success],
		fmt.Sprintf("签名在 %s 上被再次利用。", latest.ReplayedOn),
		"重点观察签名是否绑定了链环境、合约地址与 nonce。",
		1.0,
		map[string]interface{}{
			"attack_type":    latest.AttackType,
			"replayed_on":    latest.ReplayedOn,
			"signature_used": latest.Signature != nil,
			"success":        latest.Success,
		},
	)
}

// SignatureReplayFactory 创建签名重放模拟器。
type SignatureReplayFactory struct{}

func (f *SignatureReplayFactory) Create() engine.Simulator { return NewSignatureReplaySimulator() }
func (f *SignatureReplayFactory) GetDescription() types.Description {
	return NewSignatureReplaySimulator().GetDescription()
}
func NewSignatureReplayFactory() *SignatureReplayFactory { return &SignatureReplayFactory{} }

var _ engine.SimulatorFactory = (*SignatureReplayFactory)(nil)
