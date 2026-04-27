package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type MintAttackType string

const (
	MintAttackNoAccessControl MintAttackType = "no_access_control"
	MintAttackReentrancy      MintAttackType = "reentrancy"
	MintAttackOverflow        MintAttackType = "overflow"
	MintAttackLogicFlaw       MintAttackType = "logic_flaw"
	MintAttackBridgeExploit   MintAttackType = "bridge_exploit"
)

type TokenState struct {
	Name           string   `json:"name"`
	Symbol         string   `json:"symbol"`
	TotalSupply    *big.Int `json:"total_supply"`
	MaxSupply      *big.Int `json:"max_supply"`
	Price          *big.Int `json:"price"`
	MarketCap      *big.Int `json:"market_cap"`
	HasMintControl bool     `json:"has_mint_control"`
	MintableBy     []string `json:"mintable_by"`
}

type MintAttack struct {
	ID           string         `json:"id"`
	AttackType   MintAttackType `json:"attack_type"`
	TokenName    string         `json:"token_name"`
	MintedAmount *big.Int       `json:"minted_amount"`
	BeforeSupply *big.Int       `json:"before_supply"`
	AfterSupply  *big.Int       `json:"after_supply"`
	PriceBefore  *big.Int       `json:"price_before"`
	PriceAfter   *big.Int       `json:"price_after"`
	Profit       *big.Int       `json:"profit"`
	Success      bool           `json:"success"`
	Timestamp    time.Time      `json:"timestamp"`
}

type InfiniteMintSimulator struct {
	*base.BaseSimulator
	token   *TokenState
	attacks []*MintAttack
}

func NewInfiniteMintSimulator() *InfiniteMintSimulator {
	sim := &InfiniteMintSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"infinite_mint",
			"无限增发攻击演示器",
			"演示无访问控制铸币、重入铸币和桥接伪造导致的供应量失控。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*MintAttack, 0),
	}

	sim.AddParam(types.Param{
		Key:         "initial_supply",
		Name:        "初始供应量",
		Description: "代币初始供应量，单位为百万枚。",
		Type:        types.ParamTypeInt,
		Default:     100,
		Min:         1,
		Max:         1000,
	})
	sim.AddParam(types.Param{
		Key:         "has_mint_control",
		Name:        "启用铸币权限控制",
		Description: "控制是否启用安全的铸币权限边界。",
		Type:        types.ParamTypeBool,
		Default:     false,
	})

	return sim
}

func (s *InfiniteMintSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	initialSupply := int64(100)
	if value, ok := config.Params["initial_supply"]; ok {
		if typed, ok := value.(float64); ok {
			initialSupply = int64(typed)
		}
	}

	hasMintControl := false
	if value, ok := config.Params["has_mint_control"]; ok {
		if typed, ok := value.(bool); ok {
			hasMintControl = typed
		}
	}

	s.token = &TokenState{
		Name:           "VulnerableToken",
		Symbol:         "VULN",
		TotalSupply:    new(big.Int).Mul(big.NewInt(initialSupply*1000000), big.NewInt(1e18)),
		MaxSupply:      new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(1e18)),
		Price:          big.NewInt(100000000),
		HasMintControl: hasMintControl,
	}
	if hasMintControl {
		s.token.MintableBy = []string{"owner", "minter_role"}
	} else {
		s.token.MintableBy = []string{"anyone"}
	}
	s.token.MarketCap = new(big.Int).Mul(s.token.TotalSupply, s.token.Price)
	s.token.MarketCap.Div(s.token.MarketCap, big.NewInt(1e8))
	s.attacks = make([]*MintAttack, 0)
	s.updateState()
	return nil
}

func (s *InfiniteMintSimulator) ShowVulnerablePatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{"pattern": "无访问控制 mint", "severity": "Critical"},
		{"pattern": "重复领取奖励", "severity": "High"},
		{"pattern": "重入铸币", "severity": "Critical"},
		{"pattern": "桥接证明伪造", "severity": "Critical"},
	}
}

func (s *InfiniteMintSimulator) SimulateInfiniteMint(attackType MintAttackType, mintMultiplier int64) *MintAttack {
	if s.token.HasMintControl && attackType == MintAttackNoAccessControl {
		s.EmitEvent("attack_blocked", "", "", map[string]interface{}{
			"reason": "当前配置启用了铸币权限控制。",
		})
		return nil
	}

	beforeSupply := new(big.Int).Set(s.token.TotalSupply)
	beforePrice := new(big.Int).Set(s.token.Price)
	mintAmount := new(big.Int).Mul(s.token.TotalSupply, big.NewInt(mintMultiplier))
	s.token.TotalSupply.Add(s.token.TotalSupply, mintAmount)

	newPrice := new(big.Int).Mul(beforePrice, beforeSupply)
	newPrice.Div(newPrice, s.token.TotalSupply)
	s.token.Price = newPrice

	profit := new(big.Int).Mul(mintAmount, newPrice)
	profit.Div(profit, big.NewInt(1e18))

	attack := &MintAttack{
		ID:           fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:   attackType,
		TokenName:    s.token.Name,
		MintedAmount: mintAmount,
		BeforeSupply: beforeSupply,
		AfterSupply:  new(big.Int).Set(s.token.TotalSupply),
		PriceBefore:  beforePrice,
		PriceAfter:   newPrice,
		Profit:       profit,
		Success:      true,
		Timestamp:    time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("infinite_mint_attack", "", "", map[string]interface{}{
		"attack_type":     string(attackType),
		"minted":          formatTokenAmount(mintAmount),
		"before_supply":   formatTokenAmount(beforeSupply),
		"after_supply":    formatTokenAmount(s.token.TotalSupply),
		"price_before":    formatTokenAmount(beforePrice),
		"price_after":     formatTokenAmount(newPrice),
		"attacker_profit": formatTokenAmount(profit),
	})
	s.updateState()
	return attack
}

func (s *InfiniteMintSimulator) SimulateDumpAttack() map[string]interface{} {
	data := map[string]interface{}{
		"attack": "mint_then_dump",
		"flow": []string{
			"利用漏洞铸造大量代币。",
			"把新增供应抛向市场。",
			"价格迅速下跌。",
			"普通持有者资产被稀释。",
		},
		"summary": "攻击者先通过无限增发制造供给冲击，再通过抛售将账面收益兑现。",
	}
	s.SetGlobalData("latest_dump_attack", data)
	s.EmitEvent("dump_attack", "", "", data)
	s.updateState()
	return data
}

func (s *InfiniteMintSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "访问控制", "description": "限制 mint 入口，只允许 owner 或 MINTER_ROLE 调用。"},
		{"name": "供应量上限", "description": "加入最大供应量约束，防止无限增发。"},
		{"name": "时间锁与多签", "description": "高风险铸币操作经过延迟执行和多签审批。"},
		{"name": "领取去重", "description": "奖励、空投等领取逻辑必须记录已领取状态。"},
	}
}

func (s *InfiniteMintSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Cover Protocol", "date": "2020-12", "issue": "无限增发漏洞", "loss": "token price collapse"},
		{"name": "Wormhole", "date": "2022-02", "issue": "桥接铸造校验绕过", "loss": "hundreds of millions USD"},
	}
}

func (s *InfiniteMintSimulator) updateState() {
	s.SetGlobalData("token_name", s.token.Name)
	s.SetGlobalData("total_supply", formatTokenAmount(s.token.TotalSupply))
	s.SetGlobalData("price", formatTokenAmount(s.token.Price))
	s.SetGlobalData("has_mint_control", s.token.HasMintControl)
	s.SetGlobalData("attack_count", len(s.attacks))

	if len(s.attacks) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发无限增发攻击，请观察供应量、价格和收益在攻击前后的联动变化。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"economic",
			"idle",
			"等待无限增发场景。",
			"可以先触发一次增发攻击，观察供应膨胀、价格下跌和收益兑现之间的链路。",
			0,
			map[string]interface{}{
				"token_name":        s.token.Name,
				"has_mint_control":  s.token.HasMintControl,
				"total_supply":      formatTokenAmount(s.token.TotalSupply),
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{"index": 1, "title": "准备入口", "description": "攻击者识别可被利用的铸币入口或桥接验证漏洞。"},
		{"index": 2, "title": "执行增发", "description": fmt.Sprintf("额外铸造 %s 代币，供应量从 %s 增长到 %s。", formatTokenAmount(latest.MintedAmount), formatTokenAmount(latest.BeforeSupply), formatTokenAmount(latest.AfterSupply))},
		{"index": 3, "title": "冲击价格", "description": fmt.Sprintf("市场价格从 %s 下跌到 %s，持币者资产被稀释。", formatTokenAmount(latest.PriceBefore), formatTokenAmount(latest.PriceAfter))},
		{"index": 4, "title": "兑现收益", "description": fmt.Sprintf("攻击者理论可兑现收益约 %s。", formatTokenAmount(latest.Profit))},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", fmt.Sprintf("攻击类型 %s，重点观察供应量膨胀、价格下跌与收益兑现的完整链路。", latest.AttackType))
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"economic",
		"minted",
		fmt.Sprintf("代币在 %s 场景下发生了异常增发。", latest.AttackType),
		"重点观察供应量变化如何立刻传导到价格和攻击收益上。",
		1.0,
		map[string]interface{}{
			"attack_type":   latest.AttackType,
			"minted_amount": formatTokenAmount(latest.MintedAmount),
			"before_supply": formatTokenAmount(latest.BeforeSupply),
			"after_supply":  formatTokenAmount(latest.AfterSupply),
			"price_before":  formatTokenAmount(latest.PriceBefore),
			"price_after":   formatTokenAmount(latest.PriceAfter),
			"profit":        formatTokenAmount(latest.Profit),
		},
	)
}

func (s *InfiniteMintSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_infinite_mint":
		attackType := MintAttackType(actionString(params, "attack_type", string(MintAttackNoAccessControl)))
		mintMultiplier := actionInt64(params, "mint_multiplier", 10)
		attack := s.SimulateInfiniteMint(attackType, mintMultiplier)
		if attack == nil {
			return actionResultWithFeedback(
				"当前配置下无限增发已被权限控制阻断。",
				map[string]interface{}{
					"attack_type": attackType,
					"blocked":     true,
				},
				&types.ActionFeedback{
					Summary:     "铸币权限边界已经生效，攻击流程被提前阻断。",
					NextHint:    "可以比较开启和关闭铸币权限控制时，攻击结果会有什么不同。",
					EffectScope: "economic",
					ResultState: map[string]interface{}{
						"attack_type": attackType,
						"blocked":     true,
					},
				},
			), nil
		}
		return actionResultWithFeedback(
			"已执行无限增发攻击演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入供应量膨胀、价格回落和收益兑现的完整攻击流程。",
				NextHint:    "重点观察供应变化和价格变化在主舞台上的联动。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"attack_type": attack.AttackType,
					"profit":      formatTokenAmount(attack.Profit),
				},
			},
		), nil
	case "simulate_dump_attack":
		result := s.SimulateDumpAttack()
		return actionResultWithFeedback(
			"已执行增发后抛售演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入通过抛售新铸代币来兑现账面收益的攻击流程。",
				NextHint:    "重点观察价格冲击如何在抛售阶段进一步放大。",
				EffectScope: "economic",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

type InfiniteMintFactory struct{}

func (f *InfiniteMintFactory) Create() engine.Simulator { return NewInfiniteMintSimulator() }

func (f *InfiniteMintFactory) GetDescription() types.Description { return NewInfiniteMintSimulator().GetDescription() }

func NewInfiniteMintFactory() *InfiniteMintFactory { return &InfiniteMintFactory{} }

var _ engine.SimulatorFactory = (*InfiniteMintFactory)(nil)
