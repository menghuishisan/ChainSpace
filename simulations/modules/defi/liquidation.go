package defi

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 清算机制演示器
// =============================================================================

// LiquidatablePosition 可清算持仓
type LiquidatablePosition struct {
	ID               string   `json:"id"`
	Owner            string   `json:"owner"`
	Protocol         string   `json:"protocol"`
	CollateralAsset  string   `json:"collateral_asset"`
	CollateralAmount *big.Int `json:"collateral_amount"`
	CollateralValue  *big.Int `json:"collateral_value"`
	DebtAsset        string   `json:"debt_asset"`
	DebtAmount       *big.Int `json:"debt_amount"`
	DebtValue        *big.Int `json:"debt_value"`
	HealthFactor     float64  `json:"health_factor"`
	LiquidationPrice float64  `json:"liquidation_price"`
}

// LiquidationRecord 清算记录
type LiquidationRecord struct {
	ID               string    `json:"id"`
	PositionID       string    `json:"position_id"`
	Liquidator       string    `json:"liquidator"`
	Borrower         string    `json:"borrower"`
	CollateralSeized *big.Int  `json:"collateral_seized"`
	DebtRepaid       *big.Int  `json:"debt_repaid"`
	Bonus            *big.Int  `json:"bonus"`
	GasUsed          uint64    `json:"gas_used"`
	Profit           *big.Int  `json:"profit"`
	Timestamp        time.Time `json:"timestamp"`
}

// LiquidationBot 清算机器人
type LiquidationBot struct {
	ID                string   `json:"id"`
	Address           string   `json:"address"`
	Balance           *big.Int `json:"balance"`
	TotalLiquidations int      `json:"total_liquidations"`
	TotalProfit       *big.Int `json:"total_profit"`
	IsActive          bool     `json:"is_active"`
}

// LiquidationSimulator 清算机制演示器
// 演示DeFi借贷协议的清算机制:
//
// 1. 清算触发条件
//   - 健康因子 < 1
//   - 抵押率低于清算线
//
// 2. 清算流程
//   - 清算人还债，获得抵押品+奖励
//   - 部分清算 vs 全额清算
//
// 3. 清算策略
//   - MEV机器人
//   - 闪电贷清算
//
// 参考: Aave, Compound, MakerDAO清算机制
type LiquidationSimulator struct {
	*base.BaseSimulator
	positions          map[string]*LiquidatablePosition
	liquidationRecords []*LiquidationRecord
	bots               map[string]*LiquidationBot
	liquidationBonus   float64 // 清算奖励
	closeFactor        float64 // 单次最大清算比例
	gasPrice           uint64
	ethPrice           float64
}

// NewLiquidationSimulator 创建清算演示器
func NewLiquidationSimulator() *LiquidationSimulator {
	sim := &LiquidationSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"liquidation",
			"清算机制演示器",
			"演示DeFi借贷协议的清算流程、清算奖励、MEV策略等",
			"defi",
			types.ComponentDeFi,
		),
		positions:          make(map[string]*LiquidatablePosition),
		liquidationRecords: make([]*LiquidationRecord, 0),
		bots:               make(map[string]*LiquidationBot),
	}

	sim.AddParam(types.Param{
		Key:         "liquidation_bonus",
		Name:        "清算奖励",
		Description: "清算人获得的额外抵押品比例(%)",
		Type:        types.ParamTypeFloat,
		Default:     5.0,
		Min:         1.0,
		Max:         15.0,
	})

	sim.AddParam(types.Param{
		Key:         "close_factor",
		Name:        "关闭因子",
		Description: "单次清算最大债务比例(%)",
		Type:        types.ParamTypeFloat,
		Default:     50.0,
		Min:         25.0,
		Max:         100.0,
	})

	return sim
}

// Init 初始化
func (s *LiquidationSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.liquidationBonus = 0.05
	s.closeFactor = 0.50
	s.gasPrice = 50 // gwei
	s.ethPrice = 2000

	if v, ok := config.Params["liquidation_bonus"]; ok {
		if f, ok := v.(float64); ok {
			s.liquidationBonus = f / 100
		}
	}
	if v, ok := config.Params["close_factor"]; ok {
		if f, ok := v.(float64); ok {
			s.closeFactor = f / 100
		}
	}

	s.positions = make(map[string]*LiquidatablePosition)
	s.liquidationRecords = make([]*LiquidationRecord, 0)
	s.bots = make(map[string]*LiquidationBot)

	// 创建示例持仓
	s.createSamplePositions()

	s.updateState()
	return nil
}

// createSamplePositions 创建示例持仓
func (s *LiquidationSimulator) createSamplePositions() {
	positions := []*LiquidatablePosition{
		{
			ID:               "pos-1",
			Owner:            "user-1",
			Protocol:         "Aave",
			CollateralAsset:  "ETH",
			CollateralAmount: big.NewInt(10000000000), // 10 ETH (simplified)
			CollateralValue:  big.NewInt(20000000000), // $20,000
			DebtAsset:        "USDC",
			DebtAmount:       big.NewInt(18000000000), // $18,000
			DebtValue:        big.NewInt(18000000000),
			HealthFactor:     0.92,
			LiquidationPrice: 1800,
		},
		{
			ID:               "pos-2",
			Owner:            "user-2",
			Protocol:         "Compound",
			CollateralAsset:  "WBTC",
			CollateralAmount: big.NewInt(100000000),   // 1 WBTC
			CollateralValue:  big.NewInt(40000000000), // $40,000
			DebtAsset:        "DAI",
			DebtAmount:       big.NewInt(35000000000), // 35000 DAI
			DebtValue:        big.NewInt(35000000000),
			HealthFactor:     0.95,
			LiquidationPrice: 35000,
		},
		{
			ID:               "pos-3",
			Owner:            "user-3",
			Protocol:         "MakerDAO",
			CollateralAsset:  "ETH",
			CollateralAmount: big.NewInt(50000000000), // 50 ETH
			CollateralValue:  big.NewInt(100000000000),
			DebtAsset:        "DAI",
			DebtAmount:       big.NewInt(70000000000), // 70000 DAI
			DebtValue:        big.NewInt(70000000000),
			HealthFactor:     1.15,
			LiquidationPrice: 1400,
		},
	}

	for _, pos := range positions {
		s.positions[pos.ID] = pos
	}
}

// =============================================================================
// 清算机制解释
// =============================================================================

// ExplainLiquidation 解释清算机制
func (s *LiquidationSimulator) ExplainLiquidation() map[string]interface{} {
	return map[string]interface{}{
		"purpose": "保护协议偿付能力，确保所有债务都有足够抵押",
		"trigger": map[string]interface{}{
			"health_factor": "健康因子 < 1 时触发",
			"formula":       "HealthFactor = (CollateralValue × LiquidationThreshold) / DebtValue",
			"example":       "存入$20,000 ETH (LT=80%)，借$16,000 → HF = 20000×0.8/16000 = 1.0",
		},
		"process": []string{
			"1. 监控链上持仓健康因子",
			"2. 发现HF < 1的持仓",
			"3. 调用liquidate函数还债",
			"4. 获得等值抵押品 + 清算奖励",
		},
		"parameters": map[string]interface{}{
			"liquidation_bonus": fmt.Sprintf("%.0f%% - 清算人获得的额外抵押品", s.liquidationBonus*100),
			"close_factor":      fmt.Sprintf("%.0f%% - 单次最多清算债务比例", s.closeFactor*100),
		},
		"participants": []map[string]string{
			{"role": "借款人", "risk": "可能损失部分抵押品"},
			{"role": "清算人", "benefit": "赚取清算奖励"},
			{"role": "协议", "benefit": "维持偿付能力"},
		},
	}
}

// =============================================================================
// 清算操作
// =============================================================================

// GetLiquidatablePositions 获取可清算持仓
func (s *LiquidationSimulator) GetLiquidatablePositions() []*LiquidatablePosition {
	liquidatable := make([]*LiquidatablePosition, 0)
	for _, pos := range s.positions {
		if pos.HealthFactor < 1 {
			liquidatable = append(liquidatable, pos)
		}
	}

	// 按健康因子排序(越低越优先)
	sort.Slice(liquidatable, func(i, j int) bool {
		return liquidatable[i].HealthFactor < liquidatable[j].HealthFactor
	})

	return liquidatable
}

// Liquidate 执行清算
func (s *LiquidationSimulator) Liquidate(liquidator, positionID string, debtToCover *big.Int) (*LiquidationRecord, error) {
	pos, ok := s.positions[positionID]
	if !ok {
		return nil, fmt.Errorf("持仓不存在")
	}

	if pos.HealthFactor >= 1 {
		return nil, fmt.Errorf("持仓健康因子 %.2f >= 1，不可清算", pos.HealthFactor)
	}

	// 计算最大可清算债务
	maxDebt := new(big.Int).Mul(pos.DebtAmount, big.NewInt(int64(s.closeFactor*10000)))
	maxDebt.Div(maxDebt, big.NewInt(10000))

	if debtToCover.Cmp(maxDebt) > 0 {
		debtToCover = maxDebt
	}

	// 计算获得的抵押品 (包含奖励)
	debtRatio := float64(debtToCover.Int64()) / float64(pos.DebtAmount.Int64())
	collateralBase := new(big.Int).Mul(pos.CollateralAmount, big.NewInt(int64(debtRatio*10000)))
	collateralBase.Div(collateralBase, big.NewInt(10000))

	bonus := new(big.Int).Mul(collateralBase, big.NewInt(int64(s.liquidationBonus*10000)))
	bonus.Div(bonus, big.NewInt(10000))

	collateralSeized := new(big.Int).Add(collateralBase, bonus)

	// 确保不超过全部抵押品
	if collateralSeized.Cmp(pos.CollateralAmount) > 0 {
		collateralSeized = new(big.Int).Set(pos.CollateralAmount)
	}

	// 计算利润
	gasUsed := uint64(300000)
	gasCost := gasUsed * s.gasPrice * 1e9 / 1e18 * uint64(s.ethPrice)
	collateralValue := float64(collateralSeized.Int64()) * s.ethPrice / 1e18
	debtCost := float64(debtToCover.Int64()) / 1e6
	profit := big.NewInt(int64((collateralValue - debtCost - float64(gasCost)) * 1e6))

	// 更新持仓
	pos.CollateralAmount.Sub(pos.CollateralAmount, collateralSeized)
	pos.DebtAmount.Sub(pos.DebtAmount, debtToCover)

	// 重新计算健康因子
	if pos.DebtAmount.Cmp(big.NewInt(0)) == 0 {
		pos.HealthFactor = 999
	} else {
		pos.CollateralValue = big.NewInt(int64(float64(pos.CollateralAmount.Int64()) * s.ethPrice / 1e18 * 1e6))
		pos.DebtValue = pos.DebtAmount
		pos.HealthFactor = float64(pos.CollateralValue.Int64()) * 0.8 / float64(pos.DebtValue.Int64())
	}

	// 记录清算
	record := &LiquidationRecord{
		ID:               fmt.Sprintf("liq-%d", time.Now().UnixNano()),
		PositionID:       positionID,
		Liquidator:       liquidator,
		Borrower:         pos.Owner,
		CollateralSeized: collateralSeized,
		DebtRepaid:       debtToCover,
		Bonus:            bonus,
		GasUsed:          gasUsed,
		Profit:           profit,
		Timestamp:        time.Now(),
	}

	s.liquidationRecords = append(s.liquidationRecords, record)

	s.EmitEvent("liquidation_executed", "", "", map[string]interface{}{
		"position_id":       positionID,
		"liquidator":        liquidator,
		"debt_repaid":       debtToCover.String(),
		"collateral_seized": collateralSeized.String(),
		"bonus":             bonus.String(),
		"profit":            profit.String(),
		"new_health_factor": pos.HealthFactor,
	})

	s.updateState()
	return record, nil
}

// SimulatePriceChange 模拟价格变化
func (s *LiquidationSimulator) SimulatePriceChange(newEthPrice float64) map[string]interface{} {
	oldPrice := s.ethPrice
	s.ethPrice = newEthPrice

	newLiquidatable := make([]string, 0)

	for _, pos := range s.positions {
		if pos.CollateralAsset == "ETH" {
			// 更新抵押品价值
			pos.CollateralValue = big.NewInt(int64(float64(pos.CollateralAmount.Int64()) * newEthPrice / 1e18 * 1e6))

			// 重新计算健康因子
			if pos.DebtAmount.Cmp(big.NewInt(0)) > 0 {
				oldHF := pos.HealthFactor
				pos.HealthFactor = float64(pos.CollateralValue.Int64()) * 0.8 / float64(pos.DebtValue.Int64())

				if oldHF >= 1 && pos.HealthFactor < 1 {
					newLiquidatable = append(newLiquidatable, pos.ID)
				}
			}
		}
	}

	result := map[string]interface{}{
		"old_price":        oldPrice,
		"new_price":        newEthPrice,
		"price_change":     fmt.Sprintf("%.2f%%", (newEthPrice-oldPrice)/oldPrice*100),
		"new_liquidatable": newLiquidatable,
		"total_at_risk":    len(s.GetLiquidatablePositions()),
	}

	s.EmitEvent("price_changed", "", "", result)
	s.updateState()

	return result
}

// =============================================================================
// 清算策略
// =============================================================================

// ExplainLiquidationStrategies 解释清算策略
func (s *LiquidationSimulator) ExplainLiquidationStrategies() map[string]interface{} {
	return map[string]interface{}{
		"basic_liquidation": map[string]interface{}{
			"description": "使用自有资金还债获得抵押品",
			"steps": []string{
				"1. 监控可清算持仓",
				"2. 准备足够资金",
				"3. 调用liquidate",
				"4. 出售获得的抵押品",
			},
			"capital_required": "需要足够资金覆盖债务",
		},
		"flash_loan_liquidation": map[string]interface{}{
			"description": "使用闪电贷无需自有资金",
			"steps": []string{
				"1. 借入闪电贷",
				"2. 执行清算",
				"3. 出售抵押品",
				"4. 归还闪电贷",
				"5. 保留利润",
			},
			"capital_required": "仅需gas费",
			"advantage":        "资本效率极高",
		},
		"mev_strategies": map[string]interface{}{
			"description": "利用MEV提升清算成功率",
			"methods": []string{
				"Flashbots直接提交给矿工",
				"更高gas费抢先交易",
				"bundle多笔交易",
			},
			"risks": []string{
				"gas战争导致利润降低",
				"交易可能失败",
			},
		},
	}
}

// SimulateFlashLoanLiquidation 模拟闪电贷清算
func (s *LiquidationSimulator) SimulateFlashLoanLiquidation(positionID string) map[string]interface{} {
	pos, ok := s.positions[positionID]
	if !ok {
		return map[string]interface{}{"error": "持仓不存在"}
	}

	if pos.HealthFactor >= 1 {
		return map[string]interface{}{"error": "持仓不可清算"}
	}

	// 模拟闪电贷清算流程
	maxDebt := new(big.Int).Mul(pos.DebtAmount, big.NewInt(int64(s.closeFactor*10000)))
	maxDebt.Div(maxDebt, big.NewInt(10000))

	flashLoanFee := new(big.Int).Mul(maxDebt, big.NewInt(9))
	flashLoanFee.Div(flashLoanFee, big.NewInt(10000)) // 0.09%费用

	// 计算获得的抵押品
	debtRatio := s.closeFactor
	collateralBase := new(big.Int).Mul(pos.CollateralAmount, big.NewInt(int64(debtRatio*10000)))
	collateralBase.Div(collateralBase, big.NewInt(10000))

	bonus := new(big.Int).Mul(collateralBase, big.NewInt(int64(s.liquidationBonus*10000)))
	bonus.Div(bonus, big.NewInt(10000))

	collateralSeized := new(big.Int).Add(collateralBase, bonus)

	// 出售抵押品获得稳定币
	collateralValueUSD := float64(collateralSeized.Int64()) * s.ethPrice / 1e18

	// 计算利润
	gasUsed := uint64(500000) // 闪电贷需要更多gas
	gasCostUSD := float64(gasUsed) * float64(s.gasPrice) * 1e-9 * s.ethPrice

	profit := collateralValueUSD - float64(maxDebt.Int64())/1e6 - float64(flashLoanFee.Int64())/1e6 - gasCostUSD

	return map[string]interface{}{
		"position_id":       positionID,
		"flash_loan_amount": maxDebt.String(),
		"flash_loan_fee":    flashLoanFee.String(),
		"collateral_seized": collateralSeized.String(),
		"collateral_value":  fmt.Sprintf("$%.2f", collateralValueUSD),
		"gas_cost":          fmt.Sprintf("$%.2f", gasCostUSD),
		"net_profit":        fmt.Sprintf("$%.2f", profit),
		"profitable":        profit > 0,
		"flow": []string{
			fmt.Sprintf("1. 从Aave借入 %s USDC闪电贷", maxDebt.String()),
			fmt.Sprintf("2. 清算持仓，还债 %s USDC", maxDebt.String()),
			fmt.Sprintf("3. 获得 %s ETH (含%.0f%%奖励)", collateralSeized.String(), s.liquidationBonus*100),
			fmt.Sprintf("4. 在DEX出售ETH获得 $%.2f", collateralValueUSD),
			fmt.Sprintf("5. 归还闪电贷 + %.4f%%手续费", 0.09),
			fmt.Sprintf("6. 利润: $%.2f", profit),
		},
	}
}

// GetLiquidationStats 获取清算统计
func (s *LiquidationSimulator) GetLiquidationStats() map[string]interface{} {
	totalDebtRepaid := big.NewInt(0)
	totalCollateralSeized := big.NewInt(0)
	totalProfit := big.NewInt(0)

	for _, record := range s.liquidationRecords {
		totalDebtRepaid.Add(totalDebtRepaid, record.DebtRepaid)
		totalCollateralSeized.Add(totalCollateralSeized, record.CollateralSeized)
		totalProfit.Add(totalProfit, record.Profit)
	}

	return map[string]interface{}{
		"total_liquidations":      len(s.liquidationRecords),
		"total_debt_repaid":       totalDebtRepaid.String(),
		"total_collateral_seized": totalCollateralSeized.String(),
		"total_liquidator_profit": totalProfit.String(),
		"positions_at_risk":       len(s.GetLiquidatablePositions()),
		"current_eth_price":       s.ethPrice,
	}
}

// updateState 更新状态
func (s *LiquidationSimulator) updateState() {
	s.SetGlobalData("position_count", len(s.positions))
	s.SetGlobalData("liquidation_count", len(s.liquidationRecords))
	s.SetGlobalData("at_risk_count", len(s.GetLiquidatablePositions()))
	s.SetGlobalData("eth_price", s.ethPrice)

	summary := fmt.Sprintf("当前有 %d 个仓位，其中 %d 个处于可清算风险区。", len(s.positions), len(s.GetLiquidatablePositions()))
	nextHint := "可以继续调整价格或模拟闪电贷清算，观察仓位何时进入清算区间。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"liquidation_watch",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"position_count": len(s.positions), "at_risk_count": len(s.GetLiquidatablePositions()), "eth_price": s.ethPrice},
	)
}

func (s *LiquidationSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "move_price":
		newPrice := s.ethPrice * 0.9
		if raw, ok := params["new_price"].(float64); ok && raw > 0 {
			newPrice = raw
		}
		result := s.SimulatePriceChange(newPrice)
		return defiActionResult("已更新清算场景价格。", result, &types.ActionFeedback{
			Summary:     "仓位健康因子已根据新价格重新计算。",
			NextHint:    "继续观察哪些仓位进入可清算区间，以及闪电贷清算是否仍然有利可图。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"eth_price": s.ethPrice, "at_risk_count": len(s.GetLiquidatablePositions())},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported liquidation action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// LiquidationFactory 清算工厂
type LiquidationFactory struct{}

// Create 创建演示器
func (f *LiquidationFactory) Create() engine.Simulator {
	return NewLiquidationSimulator()
}

// GetDescription 获取描述
func (f *LiquidationFactory) GetDescription() types.Description {
	return NewLiquidationSimulator().GetDescription()
}

// NewLiquidationFactory 创建工厂
func NewLiquidationFactory() *LiquidationFactory {
	return &LiquidationFactory{}
}

var _ engine.SimulatorFactory = (*LiquidationFactory)(nil)
