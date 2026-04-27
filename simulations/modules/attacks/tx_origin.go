package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// PhishingAttack 表示一次基于 tx.origin 的钓鱼攻击。
type PhishingAttack struct {
	ID                string    `json:"id"`
	AttackType        string    `json:"attack_type"`
	Victim            string    `json:"victim"`
	VictimBalance     *big.Int  `json:"victim_balance"`
	Attacker          string    `json:"attacker"`
	MaliciousContract string    `json:"malicious_contract"`
	TargetContract    string    `json:"target_contract"`
	StolenAmount      *big.Int  `json:"stolen_amount"`
	PhishingMethod    string    `json:"phishing_method"`
	Success           bool      `json:"success"`
	Timestamp         time.Time `json:"timestamp"`
}

// CallContext 记录一次调用链上下文。
type CallContext struct {
	Depth     int    `json:"depth"`
	Contract  string `json:"contract"`
	Function  string `json:"function"`
	TxOrigin  string `json:"tx_origin"`
	MsgSender string `json:"msg_sender"`
	MsgValue  string `json:"msg_value"`
}

// WalletState 表示目标钱包状态。
type WalletState struct {
	Address string   `json:"address"`
	Owner   string   `json:"owner"`
	Balance *big.Int `json:"balance"`
}

// TxOriginSimulator 演示 tx.origin 与 msg.sender 混淆带来的风险。
type TxOriginSimulator struct {
	*base.BaseSimulator
	wallets   map[string]*WalletState
	attacks   []*PhishingAttack
	callStack []*CallContext
}

// NewTxOriginSimulator 创建 tx.origin 模拟器。
func NewTxOriginSimulator() *TxOriginSimulator {
	sim := &TxOriginSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"tx_origin",
			"tx.origin 攻击演示器",
			"演示开发者错误依赖 tx.origin 进行权限校验时，如何被中间合约、NFT 回调或授权钓鱼利用。",
			"attacks",
			types.ComponentAttack,
		),
		wallets:   make(map[string]*WalletState),
		attacks:   make([]*PhishingAttack, 0),
		callStack: make([]*CallContext, 0),
	}

	sim.AddParam(types.Param{
		Key:         "victim_balance",
		Name:        "受害钱包余额",
		Description: "受害者钱包在攻击开始前持有的 ETH 数量。",
		Type:        types.ParamTypeInt,
		Default:     10,
		Min:         1,
		Max:         1000,
	})

	return sim
}

// Init 初始化钱包状态。
func (s *TxOriginSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	victimBalance := int64(10)
	if v, ok := config.Params["victim_balance"]; ok {
		if n, ok := v.(float64); ok {
			victimBalance = int64(n)
		}
	}

	s.wallets = make(map[string]*WalletState)
	s.wallets["victim_wallet"] = &WalletState{
		Address: "0xVictimWallet",
		Owner:   "0xVictim",
		Balance: new(big.Int).Mul(big.NewInt(victimBalance), big.NewInt(1e18)),
	}
	s.attacks = make([]*PhishingAttack, 0)
	s.callStack = make([]*CallContext, 0)

	s.updateState()
	return nil
}

// SimulateDirectPhishing 演示最典型的中间合约钓鱼。
func (s *TxOriginSimulator) SimulateDirectPhishing(victim, attacker string) *PhishingAttack {
	wallet := s.wallets["victim_wallet"]
	stolenAmount := new(big.Int).Set(wallet.Balance)

	s.callStack = []*CallContext{
		{Depth: 0, Contract: "EOA", Function: "点击钓鱼页面", TxOrigin: victim, MsgSender: victim, MsgValue: "0"},
		{Depth: 1, Contract: "MaliciousContract", Function: "claimReward()", TxOrigin: victim, MsgSender: victim, MsgValue: "0"},
		{Depth: 2, Contract: "VulnerableWallet", Function: "withdraw()", TxOrigin: victim, MsgSender: "MaliciousContract", MsgValue: "0"},
	}

	attack := &PhishingAttack{
		ID:                fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:        "direct_phishing",
		Victim:            victim,
		VictimBalance:     new(big.Int).Set(wallet.Balance),
		Attacker:          attacker,
		MaliciousContract: "0xMaliciousContract",
		TargetContract:    wallet.Address,
		StolenAmount:      stolenAmount,
		PhishingMethod:    "伪装奖励领取页面诱导受害者点击",
		Success:           true,
		Timestamp:         time.Now(),
	}

	wallet.Balance = big.NewInt(0)
	s.attacks = append(s.attacks, attack)

	s.EmitEvent("direct_phishing_attack", "", "", map[string]interface{}{
		"attack_flow": []string{
			"攻击者部署恶意合约并伪装成正常奖励页面。",
			fmt.Sprintf("受害者 %s 在前端点击“领取奖励”。", victim),
			"恶意合约继续调用目标钱包的 withdraw 函数。",
			fmt.Sprintf("目标钱包用 tx.origin(%s) 进行鉴权。", victim),
			"虽然 msg.sender 已变成恶意合约，但错误校验仍会通过。",
			fmt.Sprintf("钱包中的 %s ETH 被转走。", formatTxOriginEther(stolenAmount)),
		},
		"call_stack": s.callStack,
		"stolen":     formatTxOriginEther(stolenAmount) + " ETH",
	})

	s.updateState()
	return attack
}

// SimulateNFTPhishing 演示通过 NFT 回调触发的钓鱼流程。
func (s *TxOriginSimulator) SimulateNFTPhishing(victim, attacker string) *PhishingAttack {
	wallet := s.wallets["victim_wallet"]
	stolenAmount := new(big.Int).Set(wallet.Balance)

	attack := &PhishingAttack{
		ID:                fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:        "nft_phishing",
		Victim:            victim,
		VictimBalance:     new(big.Int).Set(wallet.Balance),
		Attacker:          attacker,
		MaliciousContract: "0xMaliciousNFT",
		TargetContract:    wallet.Address,
		StolenAmount:      stolenAmount,
		PhishingMethod:    "利用 NFT 接收回调触发危险逻辑",
		Success:           true,
		Timestamp:         time.Now(),
	}

	wallet.Balance = big.NewInt(0)
	s.attacks = append(s.attacks, attack)
	s.callStack = []*CallContext{
		{Depth: 0, Contract: "EOA", Function: "接收 NFT", TxOrigin: victim, MsgSender: victim, MsgValue: "0"},
		{Depth: 1, Contract: "MaliciousNFT", Function: "onERC721Received()", TxOrigin: victim, MsgSender: victim, MsgValue: "0"},
		{Depth: 2, Contract: "VulnerableWallet", Function: "withdraw()", TxOrigin: victim, MsgSender: "MaliciousNFT", MsgValue: "0"},
	}

	s.EmitEvent("nft_phishing_attack", "", "", map[string]interface{}{
		"attack_flow": []string{
			"攻击者向受害者发送带有恶意回调逻辑的 NFT。",
			"受害者接收 NFT 时触发 onERC721Received。",
			"恶意 NFT 在回调内部继续调用依赖 tx.origin 的钱包函数。",
			"由于 tx.origin 仍然是受害者，钱包会误以为操作来自本人。",
			fmt.Sprintf("钱包中的 %s ETH 被恶意提取。", formatTxOriginEther(stolenAmount)),
		},
		"call_stack": s.callStack,
	})

	s.updateState()
	return attack
}

// SimulateApprovalPhishing 演示授权钓鱼。
func (s *TxOriginSimulator) SimulateApprovalPhishing(victim, attacker string) *PhishingAttack {
	attack := &PhishingAttack{
		ID:                fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:        "approval_phishing",
		Victim:            victim,
		VictimBalance:     big.NewInt(0),
		Attacker:          attacker,
		MaliciousContract: "0xMaliciousDApp",
		TargetContract:    "ERC20 Tokens",
		StolenAmount:      big.NewInt(0),
		PhishingMethod:    "诱导无限授权后转走资产",
		Success:           true,
		Timestamp:         time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.callStack = []*CallContext{
		{Depth: 0, Contract: "EOA", Function: "approve()", TxOrigin: victim, MsgSender: victim, MsgValue: "0"},
		{Depth: 1, Contract: "ERC20", Function: "approve()", TxOrigin: victim, MsgSender: victim, MsgValue: "0"},
		{Depth: 2, Contract: "MaliciousDApp", Function: "transferFrom()", TxOrigin: attacker, MsgSender: attacker, MsgValue: "0"},
	}

	s.EmitEvent("approval_phishing", "", "", map[string]interface{}{
		"attack_flow": []string{
			"攻击者伪装成正常 DApp，引导用户连接钱包。",
			"前端提示用户执行 approve 授权。",
			"受害者误将无限额度授权给恶意合约。",
			"攻击者随后调用 transferFrom 转走代币。",
		},
		"dangerous_approval": `// 受害者误将无限额度授权给恶意合约
token.approve(maliciousContract, type(uint256).max);

// 攻击者随后转走资产
token.transferFrom(victim, attacker, victimBalance);`,
		"note":       "这类问题本质上不一定依赖 tx.origin，但经常和前端钓鱼、错误授权链路一起出现。",
		"call_stack": s.callStack,
	})

	s.updateState()
	return attack
}

// CompareTxOriginMsgSender 对比 tx.origin 与 msg.sender 的差异。
func (s *TxOriginSimulator) CompareTxOriginMsgSender() map[string]interface{} {
	return map[string]interface{}{
		"definitions": map[string]string{
			"tx.origin":  "本次交易最初的外部账户，贯穿整个调用链。",
			"msg.sender": "当前这一层调用的直接调用者，每进入一层合约都会变化。",
		},
		"call_chain_example": map[string]interface{}{
			"scenario": "User(EOA) -> ContractA -> ContractB -> ContractC",
			"values": []map[string]interface{}{
				{"location": "ContractA", "tx_origin": "User", "msg_sender": "User"},
				{"location": "ContractB", "tx_origin": "User", "msg_sender": "ContractA"},
				{"location": "ContractC", "tx_origin": "User", "msg_sender": "ContractB"},
			},
		},
		"key_insight":   "权限校验应该依赖当前调用者 msg.sender，而不是沿调用链一直不变的 tx.origin。",
		"security_rule": "任何授权、提款、转账等敏感逻辑都不应使用 tx.origin 做鉴权。",
		"solidity_docs": "Solidity 官方文档长期建议不要使用 tx.origin 进行权限判断。",
	}
}

// ShowVulnerablePatterns 返回典型易受攻击模式。
func (s *TxOriginSimulator) ShowVulnerablePatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":     "直接用 tx.origin 鉴权",
			"severity": "Critical",
			"vulnerable_code": `function withdraw() public {
    require(tx.origin == owner, "Not owner");
    payable(msg.sender).transfer(address(this).balance);
}`,
			"attack_vector": "只要受害者先调用恶意合约，恶意合约再转调此函数即可绕过鉴权。",
		},
		{
			"name":     "用 tx.origin 保护代币转账",
			"severity": "Critical",
			"vulnerable_code": `function transferTo(address to, uint256 amount) public {
    require(tx.origin == owner);
    token.transfer(to, amount);
}`,
			"attack_vector": "恶意合约可以借受害者的点击行为完成受保护的资产转移。",
		},
		{
			"name":     "用 tx.origin 修改管理员",
			"severity": "Critical",
			"vulnerable_code": `function setNewOwner(address newOwner) public {
    require(tx.origin == owner);
    owner = newOwner;
}`,
			"attack_vector": "一旦受害者发起交易，恶意中间合约就可以夺取合约控制权。",
		},
	}
}

// ShowDefenses 返回防御建议。
func (s *TxOriginSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":          "使用 msg.sender 替代 tx.origin",
			"effectiveness": "最高",
			"code": `modifier onlyOwner() {
    require(msg.sender == owner, "Not owner");
    _;
}

function withdraw() public onlyOwner {
    payable(msg.sender).transfer(address(this).balance);
}`,
		},
		{
			"name":          "采用 OpenZeppelin Ownable",
			"effectiveness": "高",
			"code": `import "@openzeppelin/contracts/access/Ownable.sol";

contract SecureWallet is Ownable {
    function withdraw() public onlyOwner {
        payable(owner()).transfer(address(this).balance);
    }
}`,
		},
		{
			"name":          "敏感操作改用签名授权",
			"effectiveness": "中高",
			"code": `function withdrawWithSignature(
    uint256 amount,
    uint256 nonce,
    bytes calldata signature
) public {
    bytes32 hash = keccak256(abi.encode(amount, nonce, address(this)));
    address signer = ECDSA.recover(hash, signature);
    require(signer == owner, "Invalid signature");
    require(nonces[owner]++ == nonce, "Invalid nonce");
    payable(msg.sender).transfer(amount);
}`,
		},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *TxOriginSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "早期钱包合约中的 tx.origin 鉴权问题",
			"period": "2016-2017",
			"impact": "多个教学案例与审计报告都指出这种写法会让中间合约轻易绕过权限校验。",
			"lesson": "不要把 tx.origin 当成“用户本人”的可靠标识。",
		},
		{
			"name":   "NFT / 授权钓鱼攻击潮",
			"period": "2021-2023",
			"impact": "大量攻击通过恶意前端或伪装资产转移入口诱导用户执行危险授权。",
			"methods": []string{
				"奖励领取钓鱼页面",
				"恶意 NFT 回调",
				"伪装授权 DApp",
				"Discord / Telegram 社工链接",
			},
		},
	}
}

// ExecuteAction 执行前端触发的动作。
func (s *TxOriginSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_direct_phishing":
		victim := actionString(params, "victim", "0xVictim")
		attacker := actionString(params, "attacker", "0xAttacker")
		attack := s.SimulateDirectPhishing(victim, attacker)
		return actionResultWithFeedback(
			"已执行直接钓鱼演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
				"victim":    victim,
				"attacker":  attacker,
			},
			&types.ActionFeedback{
				Summary:     "已进入受害者点击恶意页面后，被中间合约借用身份完成提款的攻击流程。",
				NextHint:    "重点观察 tx.origin 和 msg.sender 在不同调用深度中的变化。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"stolen":      formatTxOriginEther(attack.StolenAmount),
				},
			},
		), nil
	case "simulate_nft_phishing":
		victim := actionString(params, "victim", "0xVictim")
		attacker := actionString(params, "attacker", "0xAttacker")
		attack := s.SimulateNFTPhishing(victim, attacker)
		return actionResultWithFeedback(
			"已执行 NFT 钓鱼演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
				"victim":    victim,
				"attacker":  attacker,
			},
			&types.ActionFeedback{
				Summary:     "已进入通过 NFT 回调转调钱包函数的攻击流程。",
				NextHint:    "重点观察 NFT 回调如何把恶意逻辑带入依赖 tx.origin 的钱包。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"stolen":      formatTxOriginEther(attack.StolenAmount),
				},
			},
		), nil
	case "simulate_approval_phishing":
		victim := actionString(params, "victim", "0xVictim")
		attacker := actionString(params, "attacker", "0xAttacker")
		attack := s.SimulateApprovalPhishing(victim, attacker)
		return actionResultWithFeedback(
			"已执行授权钓鱼演示。",
			map[string]interface{}{
				"attack_id": attack.ID,
				"victim":    victim,
				"attacker":  attacker,
			},
			&types.ActionFeedback{
				Summary:     "已进入伪装 DApp 诱导无限授权，再通过 transferFrom 转走资产的攻击流程。",
				NextHint:    "重点观察错误授权与 tx.origin 类身份混淆在真实钓鱼链路中如何组合出现。",
				EffectScope: "execution",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"target":      attack.TargetContract,
				},
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// formatTxOriginEther 将 wei 格式化成 ETH 文本。
func formatTxOriginEther(wei *big.Int) string {
	ether := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		new(big.Float).SetInt64(1e18),
	)
	return fmt.Sprintf("%.4f", ether)
}

// updateState 同步前端需要的可视化状态。
func (s *TxOriginSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))
	s.SetGlobalData("wallet_count", len(s.wallets))
	s.SetGlobalData("attacks", s.attacks)

	if len(s.callStack) > 0 {
		s.SetGlobalData("call_stack", s.callStack)
	}

	totalLoss := big.NewInt(0)
	for _, attack := range s.attacks {
		if attack.StolenAmount != nil {
			totalLoss.Add(totalLoss, attack.StolenAmount)
		}
	}
	s.SetGlobalData("total_loss", formatTxOriginEther(totalLoss)+" ETH")

	if len(s.attacks) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发 tx.origin 钓鱼攻击，请观察调用链中 tx.origin 与 msg.sender 的区别。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待 tx.origin 钓鱼场景。",
			"可以先触发一次直接钓鱼或 NFT 钓鱼，观察调用链身份是如何被误判的。",
			0,
			map[string]interface{}{
				"wallet_count": len(s.wallets),
				"total_loss":   formatTxOriginEther(totalLoss),
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := make([]map[string]interface{}, 0, len(s.callStack))
	for _, item := range s.callStack {
		steps = append(steps, map[string]interface{}{
			"index":       item.Depth + 1,
			"title":       fmt.Sprintf("调用深度 %d", item.Depth),
			"description": fmt.Sprintf("%s 调用 %s，tx.origin=%s，msg.sender=%s。", item.Contract, item.Function, item.TxOrigin, item.MsgSender),
		})
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", fmt.Sprintf("攻击方式：%s，受害者因错误依赖 tx.origin 而被绕过鉴权。", latest.PhishingMethod))
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"phished",
		fmt.Sprintf("tx.origin 身份误判导致 %s 场景成立。", latest.AttackType),
		"重点观察调用链上 tx.origin 始终不变，而 msg.sender 在每一层都会变化。",
		1.0,
		map[string]interface{}{
			"attack_type": latest.AttackType,
			"victim":      latest.Victim,
			"attacker":    latest.Attacker,
			"stolen":      formatTxOriginEther(latest.StolenAmount),
			"call_depth":  len(s.callStack),
		},
	)
}

// TxOriginFactory 创建 tx.origin 模拟器。
type TxOriginFactory struct{}

func (f *TxOriginFactory) Create() engine.Simulator { return NewTxOriginSimulator() }
func (f *TxOriginFactory) GetDescription() types.Description {
	return NewTxOriginSimulator().GetDescription()
}
func NewTxOriginFactory() *TxOriginFactory { return &TxOriginFactory{} }

var _ engine.SimulatorFactory = (*TxOriginFactory)(nil)
