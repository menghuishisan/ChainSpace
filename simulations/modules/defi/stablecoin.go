package defi

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 稳定币演示器
// =============================================================================

// StablecoinType 稳定币类型
type StablecoinType string

const (
	StablecoinFiat        StablecoinType = "fiat_backed"   // 法币抵押 (USDC, USDT)
	StablecoinCrypto      StablecoinType = "crypto_backed" // 加密货币抵押 (DAI)
	StablecoinAlgorithmic StablecoinType = "algorithmic"   // 算法稳定币 (UST)
	StablecoinHybrid      StablecoinType = "hybrid"        // 混合型 (FRAX)
)

// CDP 抵押债仓 (Collateralized Debt Position)
type CDP struct {
	ID               string    `json:"id"`
	Owner            string    `json:"owner"`
	CollateralType   string    `json:"collateral_type"`
	CollateralAmount *big.Int  `json:"collateral_amount"`
	DebtAmount       *big.Int  `json:"debt_amount"`      // 铸造的稳定币数量
	CollateralRatio  float64   `json:"collateral_ratio"` // 抵押率
	LiquidationPrice float64   `json:"liquidation_price"`
	StabilityFee     float64   `json:"stability_fee"` // 稳定费率
	CreatedAt        time.Time `json:"created_at"`
}

// AlgorithmicState 算法稳定币状态
type AlgorithmicState struct {
	TotalSupply      *big.Int `json:"total_supply"`
	TargetPrice      float64  `json:"target_price"`
	CurrentPrice     float64  `json:"current_price"`
	RebaseMultiplier float64  `json:"rebase_multiplier"`
	SeignioragePool  *big.Int `json:"seigniorage_pool"`
	BondSupply       *big.Int `json:"bond_supply"`
	ShareSupply      *big.Int `json:"share_supply"`
}

// StablecoinSimulator 稳定币演示器
// 演示不同类型稳定币的机制:
//
// 1. 法币抵押型 (USDC, USDT)
//   - 1:1法币储备
//   - 中心化发行
//
// 2. 加密货币超额抵押型 (DAI/MakerDAO)
//   - 超额抵押ETH等资产
//   - CDP机制
//   - 清算机制保证偿付能力
//
// 3. 算法稳定币 (UST/Luna, Basis Cash)
//   - 无/部分抵押
//   - 通过算法调节供需
//   - 套利机制维持锚定
//
// 4. 混合型 (FRAX)
//   - 部分抵押 + 部分算法
type StablecoinSimulator struct {
	*base.BaseSimulator
	stablecoinType     StablecoinType
	symbol             string
	totalSupply        *big.Int
	targetPrice        float64
	currentPrice       float64
	cdps               map[string]*CDP
	minCollateralRatio float64
	liquidationRatio   float64
	stabilityFee       float64
	algorithmicState   *AlgorithmicState
}

// NewStablecoinSimulator 创建稳定币演示器
func NewStablecoinSimulator() *StablecoinSimulator {
	sim := &StablecoinSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"stablecoin",
			"稳定币演示器",
			"演示法币抵押、超额抵押、算法稳定币等不同类型稳定币的机制",
			"defi",
			types.ComponentDeFi,
		),
		cdps: make(map[string]*CDP),
	}

	sim.AddParam(types.Param{
		Key:         "stablecoin_type",
		Name:        "稳定币类型",
		Description: "稳定币的锚定机制",
		Type:        types.ParamTypeSelect,
		Default:     "crypto_backed",
		Options: []types.Option{
			{Label: "法币抵押型", Value: "fiat_backed"},
			{Label: "加密货币超额抵押型", Value: "crypto_backed"},
			{Label: "算法稳定币", Value: "algorithmic"},
			{Label: "混合型", Value: "hybrid"},
		},
	})

	sim.AddParam(types.Param{
		Key:         "min_collateral_ratio",
		Name:        "最小抵押率",
		Description: "CDP最小抵押率(%)",
		Type:        types.ParamTypeFloat,
		Default:     150.0,
		Min:         100,
		Max:         300,
	})

	return sim
}

// Init 初始化
func (s *StablecoinSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.stablecoinType = StablecoinCrypto
	s.symbol = "DAI"
	s.totalSupply = big.NewInt(0)
	s.targetPrice = 1.0
	s.currentPrice = 1.0
	s.minCollateralRatio = 150.0
	s.liquidationRatio = 130.0
	s.stabilityFee = 2.0 // 2% 年化

	if v, ok := config.Params["stablecoin_type"]; ok {
		if t, ok := v.(string); ok {
			s.stablecoinType = StablecoinType(t)
		}
	}
	if v, ok := config.Params["min_collateral_ratio"]; ok {
		if f, ok := v.(float64); ok {
			s.minCollateralRatio = f
		}
	}

	s.cdps = make(map[string]*CDP)

	// 初始化算法稳定币状态
	if s.stablecoinType == StablecoinAlgorithmic {
		totalSupply := new(big.Int)
		totalSupply.SetString("1000000000000000000000000", 10) // 1M * 1e18
		shareSupply := new(big.Int)
		shareSupply.SetString("100000000000000000000000", 10) // 100K * 1e18
		s.algorithmicState = &AlgorithmicState{
			TotalSupply:      totalSupply,
			TargetPrice:      1.0,
			CurrentPrice:     1.0,
			RebaseMultiplier: 1.0,
			SeignioragePool:  big.NewInt(0),
			BondSupply:       big.NewInt(0),
			ShareSupply:      shareSupply,
		}
	}

	s.updateState()
	return nil
}

// =============================================================================
// 稳定币类型解释
// =============================================================================

// ExplainStablecoinTypes 解释稳定币类型
func (s *StablecoinSimulator) ExplainStablecoinTypes() map[string]interface{} {
	return map[string]interface{}{
		"fiat_backed": map[string]interface{}{
			"name":     "法币抵押型稳定币",
			"examples": []string{"USDC", "USDT", "BUSD"},
			"mechanism": []string{
				"每发行1个稳定币，银行账户中存入$1",
				"用户可随时1:1赎回法币",
				"定期审计验证储备",
			},
			"pros":  []string{"简单易懂", "稳定性高"},
			"cons":  []string{"中心化", "监管风险", "银行风险"},
			"trust": "需要信任发行方",
		},
		"crypto_backed": map[string]interface{}{
			"name":     "加密货币超额抵押型",
			"examples": []string{"DAI", "LUSD", "sUSD"},
			"mechanism": []string{
				"用户存入ETH等抵押品创建CDP",
				"按抵押率铸造稳定币",
				"抵押率低于清算线时被清算",
				"清算机制保证系统偿付能力",
			},
			"pros":  []string{"去中心化", "透明", "抗审查"},
			"cons":  []string{"资本效率低", "价格波动风险", "复杂性"},
			"trust": "信任智能合约代码",
		},
		"algorithmic": map[string]interface{}{
			"name":     "算法稳定币",
			"examples": []string{"UST(已崩盘)", "AMPL", "Basis Cash"},
			"mechanism": []string{
				"无抵押或部分抵押",
				"价格>$1时增发，价格<$1时收缩",
				"依靠套利者维持锚定",
			},
			"pros":    []string{"资本效率高", "完全去中心化"},
			"cons":    []string{"死亡螺旋风险", "历史上多次失败"},
			"trust":   "信任算法和市场参与者",
			"warning": "UST/Luna崩盘损失超过400亿美元",
		},
		"hybrid": map[string]interface{}{
			"name":     "混合型稳定币",
			"examples": []string{"FRAX", "FEI"},
			"mechanism": []string{
				"部分法币/加密货币抵押",
				"部分算法调节",
				"抵押率随市场条件调整",
			},
			"pros": []string{"平衡效率和稳定性"},
			"cons": []string{"复杂性"},
		},
	}
}

// =============================================================================
// CDP操作 (超额抵押型)
// =============================================================================

// OpenCDP 开设抵押债仓
func (s *StablecoinSimulator) OpenCDP(owner, collateralType string, collateralAmount *big.Int, debtAmount *big.Int, collateralPrice float64) (*CDP, error) {
	// 计算抵押率
	collateralValue := float64(collateralAmount.Int64()) * collateralPrice
	debtValue := float64(debtAmount.Int64())
	collateralRatio := (collateralValue / debtValue) * 100

	if collateralRatio < s.minCollateralRatio {
		return nil, fmt.Errorf("抵押率%.1f%%低于最低要求%.1f%%", collateralRatio, s.minCollateralRatio)
	}

	cdpID := fmt.Sprintf("cdp-%s-%d", owner, time.Now().UnixNano())
	cdp := &CDP{
		ID:               cdpID,
		Owner:            owner,
		CollateralType:   collateralType,
		CollateralAmount: collateralAmount,
		DebtAmount:       debtAmount,
		CollateralRatio:  collateralRatio,
		LiquidationPrice: debtValue * s.liquidationRatio / 100 / float64(collateralAmount.Int64()),
		StabilityFee:     s.stabilityFee,
		CreatedAt:        time.Now(),
	}

	s.cdps[cdpID] = cdp
	s.totalSupply.Add(s.totalSupply, debtAmount)

	s.EmitEvent("cdp_opened", "", "", map[string]interface{}{
		"cdp_id":            cdpID,
		"owner":             owner,
		"collateral":        collateralAmount.String(),
		"debt":              debtAmount.String(),
		"collateral_ratio":  fmt.Sprintf("%.1f%%", collateralRatio),
		"liquidation_price": cdp.LiquidationPrice,
	})

	s.updateState()
	return cdp, nil
}

// AddCollateral 增加抵押品
func (s *StablecoinSimulator) AddCollateral(cdpID string, amount *big.Int, collateralPrice float64) error {
	cdp, ok := s.cdps[cdpID]
	if !ok {
		return fmt.Errorf("CDP不存在")
	}

	cdp.CollateralAmount.Add(cdp.CollateralAmount, amount)
	s.updateCDPRatio(cdp, collateralPrice)

	s.EmitEvent("collateral_added", "", "", map[string]interface{}{
		"cdp_id":    cdpID,
		"amount":    amount.String(),
		"new_ratio": fmt.Sprintf("%.1f%%", cdp.CollateralRatio),
	})

	s.updateState()
	return nil
}

// RepayDebt 偿还债务
func (s *StablecoinSimulator) RepayDebt(cdpID string, amount *big.Int, collateralPrice float64) error {
	cdp, ok := s.cdps[cdpID]
	if !ok {
		return fmt.Errorf("CDP不存在")
	}

	if amount.Cmp(cdp.DebtAmount) > 0 {
		amount = new(big.Int).Set(cdp.DebtAmount)
	}

	cdp.DebtAmount.Sub(cdp.DebtAmount, amount)
	s.totalSupply.Sub(s.totalSupply, amount)
	s.updateCDPRatio(cdp, collateralPrice)

	s.EmitEvent("debt_repaid", "", "", map[string]interface{}{
		"cdp_id":    cdpID,
		"amount":    amount.String(),
		"remaining": cdp.DebtAmount.String(),
	})

	s.updateState()
	return nil
}

// LiquidateCDP 清算CDP
func (s *StablecoinSimulator) LiquidateCDP(cdpID string, liquidator string, collateralPrice float64) (map[string]interface{}, error) {
	cdp, ok := s.cdps[cdpID]
	if !ok {
		return nil, fmt.Errorf("CDP不存在")
	}

	s.updateCDPRatio(cdp, collateralPrice)

	if cdp.CollateralRatio >= s.liquidationRatio {
		return nil, fmt.Errorf("CDP抵押率%.1f%% >= 清算线%.1f%%，不可清算", cdp.CollateralRatio, s.liquidationRatio)
	}

	// 清算: 清算人还债，获得抵押品(含奖励)
	liquidationBonus := 0.13 // 13%清算奖励
	collateralSeized := new(big.Int).Set(cdp.CollateralAmount)
	debtRepaid := new(big.Int).Set(cdp.DebtAmount)

	// 删除CDP
	delete(s.cdps, cdpID)
	s.totalSupply.Sub(s.totalSupply, debtRepaid)

	result := map[string]interface{}{
		"cdp_id":            cdpID,
		"liquidator":        liquidator,
		"debt_repaid":       debtRepaid.String(),
		"collateral_seized": collateralSeized.String(),
		"liquidation_bonus": fmt.Sprintf("%.0f%%", liquidationBonus*100),
		"final_ratio":       fmt.Sprintf("%.1f%%", cdp.CollateralRatio),
	}

	s.EmitEvent("cdp_liquidated", "", "", result)

	s.updateState()
	return result, nil
}

// updateCDPRatio 更新CDP抵押率
func (s *StablecoinSimulator) updateCDPRatio(cdp *CDP, collateralPrice float64) {
	if cdp.DebtAmount.Cmp(big.NewInt(0)) == 0 {
		cdp.CollateralRatio = 999999
		return
	}

	collateralValue := float64(cdp.CollateralAmount.Int64()) * collateralPrice
	debtValue := float64(cdp.DebtAmount.Int64())
	cdp.CollateralRatio = (collateralValue / debtValue) * 100
	cdp.LiquidationPrice = debtValue * s.liquidationRatio / 100 / float64(cdp.CollateralAmount.Int64())
}

// =============================================================================
// 算法稳定币操作
// =============================================================================

// SimulateAlgorithmicRebase 模拟算法稳定币Rebase
func (s *StablecoinSimulator) SimulateAlgorithmicRebase(currentPrice float64) map[string]interface{} {
	if s.algorithmicState == nil {
		return map[string]interface{}{"error": "非算法稳定币类型"}
	}

	s.algorithmicState.CurrentPrice = currentPrice
	deviation := currentPrice - s.algorithmicState.TargetPrice

	var action string
	var supplyChange float64

	if deviation > 0.05 { // 价格高于$1.05
		// 增发
		supplyChange = deviation * 0.1 // 每偏离1%增发0.1%
		action = "expansion"
		s.algorithmicState.RebaseMultiplier *= (1 + supplyChange)
	} else if deviation < -0.05 { // 价格低于$0.95
		// 收缩
		supplyChange = deviation * 0.1
		action = "contraction"
		s.algorithmicState.RebaseMultiplier *= (1 + supplyChange)
	} else {
		action = "neutral"
	}

	result := map[string]interface{}{
		"current_price":  currentPrice,
		"target_price":   s.algorithmicState.TargetPrice,
		"deviation":      fmt.Sprintf("%.2f%%", deviation*100),
		"action":         action,
		"supply_change":  fmt.Sprintf("%.2f%%", supplyChange*100),
		"new_multiplier": s.algorithmicState.RebaseMultiplier,
	}

	s.EmitEvent("rebase", "", "", result)
	return result
}

// ExplainDeathSpiral 解释死亡螺旋
func (s *StablecoinSimulator) ExplainDeathSpiral() map[string]interface{} {
	return map[string]interface{}{
		"name":        "死亡螺旋 (Death Spiral)",
		"description": "算法稳定币脱锚后可能进入的恶性循环",
		"mechanism": []string{
			"1. 稳定币价格开始下跌",
			"2. 持有者恐慌抛售",
			"3. 为维持锚定，算法铸造更多治理代币",
			"4. 治理代币价格暴跌",
			"5. 进一步降低对稳定币的信心",
			"6. 更多抛售，价格进一步下跌",
			"7. 循环加剧直到崩盘",
		},
		"case_study": map[string]interface{}{
			"name":    "UST/Luna崩盘",
			"date":    "2022年5月",
			"loss":    "超过400亿美元",
			"trigger": "大额UST抛售导致脱锚",
			"result":  "UST从$1跌至<$0.01，Luna归零",
		},
		"lessons": []string{
			"纯算法稳定币存在根本性风险",
			"缺乏真实抵押难以应对极端情况",
			"信心丧失会触发自我强化的崩盘",
		},
	}
}

// updateState 更新状态
func (s *StablecoinSimulator) updateState() {
	s.SetGlobalData("stablecoin_type", string(s.stablecoinType))
	s.SetGlobalData("total_supply", s.totalSupply.String())
	s.SetGlobalData("current_price", s.currentPrice)
	s.SetGlobalData("cdp_count", len(s.cdps))

	summary := fmt.Sprintf("当前稳定币总供应量为 %s，共有 %d 个 CDP。", s.totalSupply.String(), len(s.cdps))
	nextHint := "可以继续创建 CDP、补充抵押品或偿还债务，观察抵押率和清算阈值变化。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"stablecoin_management",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"total_supply": s.totalSupply.String(), "cdp_count": len(s.cdps), "current_price": s.currentPrice},
	)
}

func (s *StablecoinSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "open_cdp":
		owner := "alice"
		collateralType := "ETH"
		collateralAmount := big.NewInt(200)
		mintAmount := big.NewInt(100)
		price := s.currentPrice
		if raw, ok := params["owner"].(string); ok && raw != "" {
			owner = raw
		}
		if raw, ok := params["collateral_type"].(string); ok && raw != "" {
			collateralType = raw
		}
		if raw, ok := params["collateral_amount"].(float64); ok && raw > 0 {
			collateralAmount = big.NewInt(int64(raw))
		}
		if raw, ok := params["mint_amount"].(float64); ok && raw > 0 {
			mintAmount = big.NewInt(int64(raw))
		}
		if raw, ok := params["collateral_price"].(float64); ok && raw > 0 {
			price = raw
		}
		cdp, err := s.OpenCDP(owner, collateralType, collateralAmount, mintAmount, price)
		if err != nil {
			return nil, err
		}
		return defiActionResult("已创建一个 CDP。", map[string]interface{}{"cdp_id": cdp.ID}, &types.ActionFeedback{
			Summary:     "新的抵押债仓已经建立，稳定币已被铸造出来。",
			NextHint:    "继续调整抵押品价格或补充抵押，观察抵押率和清算价格如何变化。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"cdp_id": cdp.ID, "total_supply": s.totalSupply.String()},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported stablecoin action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// StablecoinFactory 稳定币工厂
type StablecoinFactory struct{}

// Create 创建演示器
func (f *StablecoinFactory) Create() engine.Simulator {
	return NewStablecoinSimulator()
}

// GetDescription 获取描述
func (f *StablecoinFactory) GetDescription() types.Description {
	return NewStablecoinSimulator().GetDescription()
}

// NewStablecoinFactory 创建工厂
func NewStablecoinFactory() *StablecoinFactory {
	return &StablecoinFactory{}
}

var _ engine.SimulatorFactory = (*StablecoinFactory)(nil)
