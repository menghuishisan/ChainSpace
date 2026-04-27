package defi

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 借贷协议演示器
// =============================================================================

// CollateralAsset 抵押资产
type CollateralAsset struct {
	Symbol           string   `json:"symbol"`
	PriceUSD         *big.Int `json:"price_usd"`
	CollateralFactor float64  `json:"collateral_factor"` // 0-1, 例如0.75表示75%
	LiquidationBonus float64  `json:"liquidation_bonus"` // 清算奖励, 例如0.05表示5%
	TotalSupplied    *big.Int `json:"total_supplied"`
	TotalBorrowed    *big.Int `json:"total_borrowed"`
	SupplyAPY        float64  `json:"supply_apy"`
	BorrowAPY        float64  `json:"borrow_apy"`
	ReserveFactor    float64  `json:"reserve_factor"` // 协议抽成
}

// UserPosition 用户持仓
type UserPosition struct {
	User            string              `json:"user"`
	Supplies        map[string]*big.Int `json:"supplies"`      // 存款
	Borrows         map[string]*big.Int `json:"borrows"`       // 借款
	HealthFactor    float64             `json:"health_factor"` // 健康因子
	CollateralValue *big.Int            `json:"collateral_value"`
	BorrowValue     *big.Int            `json:"borrow_value"`
	AvailableBorrow *big.Int            `json:"available_borrow"`
}

// LendingAction 借贷操作
type LendingAction struct {
	Type      string    `json:"type"` // supply, withdraw, borrow, repay, liquidate
	User      string    `json:"user"`
	Asset     string    `json:"asset"`
	Amount    *big.Int  `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

// LiquidationEvent 清算事件
type LiquidationEvent struct {
	Liquidator       string    `json:"liquidator"`
	Borrower         string    `json:"borrower"`
	DebtAsset        string    `json:"debt_asset"`
	CollateralAsset  string    `json:"collateral_asset"`
	DebtRepaid       *big.Int  `json:"debt_repaid"`
	CollateralSeized *big.Int  `json:"collateral_seized"`
	LiquidatorProfit *big.Int  `json:"liquidator_profit"`
	Timestamp        time.Time `json:"timestamp"`
}

// LendingSimulator 借贷协议演示器
// 演示DeFi借贷协议的核心机制:
//
// 1. 超额抵押借贷
//   - 存入抵押品
//   - 根据抵押因子计算可借额度
//   - 借出资产
//
// 2. 利率模型
//   - 利用率决定利率
//   - 存款利率 vs 借款利率
//
// 3. 清算机制
//   - 健康因子 < 1 时可被清算
//   - 清算人获得清算奖励
//
// 参考: Aave, Compound
type LendingSimulator struct {
	*base.BaseSimulator
	assets             map[string]*CollateralAsset
	positions          map[string]*UserPosition
	actions            []*LendingAction
	liquidations       []*LiquidationEvent
	baseRate           float64 // 基础利率
	multiplier         float64 // 利率乘数
	jumpMultiplier     float64 // 跳跃乘数
	optimalUtilization float64 // 最优利用率
}

// NewLendingSimulator 创建借贷协议演示器
func NewLendingSimulator() *LendingSimulator {
	sim := &LendingSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"lending",
			"借贷协议演示器",
			"演示超额抵押借贷、利率模型、清算机制等DeFi借贷核心概念",
			"defi",
			types.ComponentDeFi,
		),
		assets:       make(map[string]*CollateralAsset),
		positions:    make(map[string]*UserPosition),
		actions:      make([]*LendingAction, 0),
		liquidations: make([]*LiquidationEvent, 0),
	}

	sim.AddParam(types.Param{
		Key:         "base_rate",
		Name:        "基础利率",
		Description: "利用率为0时的基础利率(%)",
		Type:        types.ParamTypeFloat,
		Default:     2.0,
		Min:         0,
		Max:         10,
	})

	sim.AddParam(types.Param{
		Key:         "optimal_utilization",
		Name:        "最优利用率",
		Description: "利率曲线拐点(%)",
		Type:        types.ParamTypeFloat,
		Default:     80.0,
		Min:         50,
		Max:         95,
	})

	return sim
}

// Init 初始化
func (s *LendingSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.baseRate = 2.0
	s.multiplier = 4.0
	s.jumpMultiplier = 75.0
	s.optimalUtilization = 80.0

	if v, ok := config.Params["base_rate"]; ok {
		if f, ok := v.(float64); ok {
			s.baseRate = f
		}
	}
	if v, ok := config.Params["optimal_utilization"]; ok {
		if f, ok := v.(float64); ok {
			s.optimalUtilization = f
		}
	}

	// 初始化资产
	s.assets = map[string]*CollateralAsset{
		"ETH": {
			Symbol:           "ETH",
			PriceUSD:         big.NewInt(200000000000), // $2000 * 1e8
			CollateralFactor: 0.80,
			LiquidationBonus: 0.05,
			TotalSupplied:    big.NewInt(10000000000000), // 10000 tokens (simplified)
			TotalBorrowed:    big.NewInt(5000000000000),
			ReserveFactor:    0.10,
		},
		"USDC": {
			Symbol:           "USDC",
			PriceUSD:         big.NewInt(100000000), // $1 * 1e8
			CollateralFactor: 0.85,
			LiquidationBonus: 0.04,
			TotalSupplied:    big.NewInt(50000000000000), // 50M * 1e6
			TotalBorrowed:    big.NewInt(30000000000000),
			ReserveFactor:    0.10,
		},
		"WBTC": {
			Symbol:           "WBTC",
			PriceUSD:         big.NewInt(4000000000000), // $40000 * 1e8
			CollateralFactor: 0.75,
			LiquidationBonus: 0.06,
			TotalSupplied:    big.NewInt(50000000000), // 500 * 1e8
			TotalBorrowed:    big.NewInt(20000000000),
			ReserveFactor:    0.10,
		},
	}

	// 更新利率
	for symbol := range s.assets {
		s.updateAssetRates(symbol)
	}

	s.positions = make(map[string]*UserPosition)
	s.updateState()
	return nil
}

// =============================================================================
// 利率模型
// =============================================================================

// CalculateUtilizationRate 计算利用率
func (s *LendingSimulator) CalculateUtilizationRate(asset *CollateralAsset) float64 {
	if asset.TotalSupplied.Cmp(big.NewInt(0)) == 0 {
		return 0
	}
	return float64(asset.TotalBorrowed.Int64()) / float64(asset.TotalSupplied.Int64()) * 100
}

// CalculateBorrowRate 计算借款利率 (跳跃利率模型)
// Compound/Aave使用的分段线性利率模型
func (s *LendingSimulator) CalculateBorrowRate(utilization float64) float64 {
	if utilization <= s.optimalUtilization {
		// 低于最优利用率: 线性增长
		return s.baseRate + (utilization/s.optimalUtilization)*s.multiplier
	}
	// 高于最优利用率: 急剧增长
	normalRate := s.baseRate + s.multiplier
	excessUtilization := utilization - s.optimalUtilization
	maxExcess := 100 - s.optimalUtilization
	return normalRate + (excessUtilization/maxExcess)*s.jumpMultiplier
}

// CalculateSupplyRate 计算存款利率
func (s *LendingSimulator) CalculateSupplyRate(borrowRate, utilization float64, reserveFactor float64) float64 {
	// supplyRate = borrowRate * utilization * (1 - reserveFactor)
	return borrowRate * (utilization / 100) * (1 - reserveFactor)
}

// updateAssetRates 更新资产利率
func (s *LendingSimulator) updateAssetRates(symbol string) {
	asset := s.assets[symbol]
	utilization := s.CalculateUtilizationRate(asset)
	asset.BorrowAPY = s.CalculateBorrowRate(utilization)
	asset.SupplyAPY = s.CalculateSupplyRate(asset.BorrowAPY, utilization, asset.ReserveFactor)
}

// ExplainInterestRateModel 解释利率模型
func (s *LendingSimulator) ExplainInterestRateModel() map[string]interface{} {
	return map[string]interface{}{
		"model": "跳跃利率模型 (Jump Rate Model)",
		"formula": map[string]string{
			"below_optimal": "borrowRate = baseRate + (utilization/optimalUtilization) * multiplier",
			"above_optimal": "borrowRate = normalRate + ((utilization-optimal)/(100-optimal)) * jumpMultiplier",
			"supply_rate":   "supplyRate = borrowRate * utilization * (1 - reserveFactor)",
		},
		"parameters": map[string]interface{}{
			"base_rate":           fmt.Sprintf("%.1f%%", s.baseRate),
			"multiplier":          fmt.Sprintf("%.1f%%", s.multiplier),
			"jump_multiplier":     fmt.Sprintf("%.1f%%", s.jumpMultiplier),
			"optimal_utilization": fmt.Sprintf("%.0f%%", s.optimalUtilization),
		},
		"purpose": []string{
			"低利用率时保持低利率，鼓励借款",
			"接近100%利用率时急剧提高利率",
			"激励存款人存入，借款人还款",
			"防止流动性枯竭",
		},
	}
}

// =============================================================================
// 借贷操作
// =============================================================================

// Supply 存款
func (s *LendingSimulator) Supply(user, asset string, amount *big.Int) error {
	assetInfo, ok := s.assets[asset]
	if !ok {
		return fmt.Errorf("资产不存在: %s", asset)
	}

	// 获取或创建用户持仓
	position := s.getOrCreatePosition(user)

	// 更新存款
	if position.Supplies[asset] == nil {
		position.Supplies[asset] = big.NewInt(0)
	}
	position.Supplies[asset].Add(position.Supplies[asset], amount)
	assetInfo.TotalSupplied.Add(assetInfo.TotalSupplied, amount)

	// 记录操作
	s.actions = append(s.actions, &LendingAction{
		Type:      "supply",
		User:      user,
		Asset:     asset,
		Amount:    amount,
		Timestamp: time.Now(),
	})

	s.updateAssetRates(asset)
	s.updatePosition(user)

	s.EmitEvent("supply", "", "", map[string]interface{}{
		"user":   user,
		"asset":  asset,
		"amount": amount.String(),
	})

	s.updateState()
	return nil
}

// Withdraw 取款
func (s *LendingSimulator) Withdraw(user, asset string, amount *big.Int) error {
	position, ok := s.positions[user]
	if !ok {
		return fmt.Errorf("用户没有持仓")
	}

	supply := position.Supplies[asset]
	if supply == nil || supply.Cmp(amount) < 0 {
		return fmt.Errorf("存款不足")
	}

	// 检查取款后健康因子
	testPosition := s.clonePosition(position)
	testPosition.Supplies[asset].Sub(testPosition.Supplies[asset], amount)
	s.calculateHealthFactor(testPosition)
	if testPosition.HealthFactor < 1 && testPosition.BorrowValue.Cmp(big.NewInt(0)) > 0 {
		return fmt.Errorf("取款后健康因子低于1")
	}

	// 执行取款
	position.Supplies[asset].Sub(position.Supplies[asset], amount)
	s.assets[asset].TotalSupplied.Sub(s.assets[asset].TotalSupplied, amount)

	s.actions = append(s.actions, &LendingAction{
		Type:      "withdraw",
		User:      user,
		Asset:     asset,
		Amount:    amount,
		Timestamp: time.Now(),
	})

	s.updateAssetRates(asset)
	s.updatePosition(user)

	s.EmitEvent("withdraw", "", "", map[string]interface{}{
		"user":   user,
		"asset":  asset,
		"amount": amount.String(),
	})

	s.updateState()
	return nil
}

// Borrow 借款
func (s *LendingSimulator) Borrow(user, asset string, amount *big.Int) error {
	position := s.getOrCreatePosition(user)

	// 检查借款额度
	s.updatePosition(user)
	if position.AvailableBorrow.Cmp(amount) < 0 {
		return fmt.Errorf("超出借款额度")
	}

	assetInfo := s.assets[asset]
	if assetInfo.TotalSupplied.Cmp(new(big.Int).Add(assetInfo.TotalBorrowed, amount)) < 0 {
		return fmt.Errorf("流动性不足")
	}

	// 执行借款
	if position.Borrows[asset] == nil {
		position.Borrows[asset] = big.NewInt(0)
	}
	position.Borrows[asset].Add(position.Borrows[asset], amount)
	assetInfo.TotalBorrowed.Add(assetInfo.TotalBorrowed, amount)

	s.actions = append(s.actions, &LendingAction{
		Type:      "borrow",
		User:      user,
		Asset:     asset,
		Amount:    amount,
		Timestamp: time.Now(),
	})

	s.updateAssetRates(asset)
	s.updatePosition(user)

	s.EmitEvent("borrow", "", "", map[string]interface{}{
		"user":          user,
		"asset":         asset,
		"amount":        amount.String(),
		"health_factor": position.HealthFactor,
	})

	s.updateState()
	return nil
}

// Repay 还款
func (s *LendingSimulator) Repay(user, asset string, amount *big.Int) error {
	position, ok := s.positions[user]
	if !ok {
		return fmt.Errorf("用户没有持仓")
	}

	borrow := position.Borrows[asset]
	if borrow == nil || borrow.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("没有借款")
	}

	// 计算实际还款金额
	repayAmount := amount
	if repayAmount.Cmp(borrow) > 0 {
		repayAmount = new(big.Int).Set(borrow)
	}

	position.Borrows[asset].Sub(position.Borrows[asset], repayAmount)
	s.assets[asset].TotalBorrowed.Sub(s.assets[asset].TotalBorrowed, repayAmount)

	s.actions = append(s.actions, &LendingAction{
		Type:      "repay",
		User:      user,
		Asset:     asset,
		Amount:    repayAmount,
		Timestamp: time.Now(),
	})

	s.updateAssetRates(asset)
	s.updatePosition(user)

	s.EmitEvent("repay", "", "", map[string]interface{}{
		"user":          user,
		"asset":         asset,
		"amount":        repayAmount.String(),
		"health_factor": position.HealthFactor,
	})

	s.updateState()
	return nil
}

// =============================================================================
// 清算
// =============================================================================

// Liquidate 清算
func (s *LendingSimulator) Liquidate(liquidator, borrower, debtAsset, collateralAsset string, debtAmount *big.Int) (*LiquidationEvent, error) {
	position, ok := s.positions[borrower]
	if !ok {
		return nil, fmt.Errorf("借款人不存在")
	}

	s.updatePosition(borrower)
	if position.HealthFactor >= 1 {
		return nil, fmt.Errorf("健康因子>=1，不可清算")
	}

	debt := position.Borrows[debtAsset]
	if debt == nil || debt.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("借款人没有该债务")
	}

	collateral := position.Supplies[collateralAsset]
	if collateral == nil || collateral.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("借款人没有该抵押品")
	}

	// 最多清算50%的债务
	maxLiquidation := new(big.Int).Div(debt, big.NewInt(2))
	if debtAmount.Cmp(maxLiquidation) > 0 {
		debtAmount = maxLiquidation
	}

	// 计算获得的抵押品 (包含清算奖励)
	debtAssetInfo := s.assets[debtAsset]
	collateralAssetInfo := s.assets[collateralAsset]

	debtValueUSD := new(big.Int).Mul(debtAmount, debtAssetInfo.PriceUSD)
	collateralWithBonus := new(big.Int).Mul(
		debtValueUSD,
		big.NewInt(int64((1+collateralAssetInfo.LiquidationBonus)*10000)),
	)
	collateralWithBonus.Div(collateralWithBonus, big.NewInt(10000))
	collateralSeized := new(big.Int).Div(collateralWithBonus, collateralAssetInfo.PriceUSD)

	// 确保不超过抵押品总量
	if collateralSeized.Cmp(collateral) > 0 {
		collateralSeized = new(big.Int).Set(collateral)
	}

	// 执行清算
	position.Borrows[debtAsset].Sub(position.Borrows[debtAsset], debtAmount)
	position.Supplies[collateralAsset].Sub(position.Supplies[collateralAsset], collateralSeized)
	debtAssetInfo.TotalBorrowed.Sub(debtAssetInfo.TotalBorrowed, debtAmount)
	collateralAssetInfo.TotalSupplied.Sub(collateralAssetInfo.TotalSupplied, collateralSeized)

	// 计算清算人利润
	baseCollateral := new(big.Int).Div(debtValueUSD, collateralAssetInfo.PriceUSD)
	liquidatorProfit := new(big.Int).Sub(collateralSeized, baseCollateral)

	event := &LiquidationEvent{
		Liquidator:       liquidator,
		Borrower:         borrower,
		DebtAsset:        debtAsset,
		CollateralAsset:  collateralAsset,
		DebtRepaid:       debtAmount,
		CollateralSeized: collateralSeized,
		LiquidatorProfit: liquidatorProfit,
		Timestamp:        time.Now(),
	}

	s.liquidations = append(s.liquidations, event)
	s.updatePosition(borrower)

	s.EmitEvent("liquidation", "", "", map[string]interface{}{
		"liquidator":        liquidator,
		"borrower":          borrower,
		"debt_repaid":       debtAmount.String(),
		"collateral_seized": collateralSeized.String(),
		"profit":            liquidatorProfit.String(),
	})

	s.updateState()
	return event, nil
}

// =============================================================================
// 辅助函数
// =============================================================================

// getOrCreatePosition 获取或创建持仓
func (s *LendingSimulator) getOrCreatePosition(user string) *UserPosition {
	if position, ok := s.positions[user]; ok {
		return position
	}

	position := &UserPosition{
		User:     user,
		Supplies: make(map[string]*big.Int),
		Borrows:  make(map[string]*big.Int),
	}
	s.positions[user] = position
	return position
}

// updatePosition 更新持仓信息
func (s *LendingSimulator) updatePosition(user string) {
	position, ok := s.positions[user]
	if !ok {
		return
	}

	s.calculateHealthFactor(position)
}

// calculateHealthFactor 计算健康因子
func (s *LendingSimulator) calculateHealthFactor(position *UserPosition) {
	collateralValue := big.NewInt(0)
	borrowValue := big.NewInt(0)

	// 计算抵押品价值 (带抵押因子)
	for asset, amount := range position.Supplies {
		if amount.Cmp(big.NewInt(0)) > 0 {
			assetInfo := s.assets[asset]
			value := new(big.Int).Mul(amount, assetInfo.PriceUSD)
			adjustedValue := new(big.Int).Mul(value, big.NewInt(int64(assetInfo.CollateralFactor*10000)))
			adjustedValue.Div(adjustedValue, big.NewInt(10000))
			collateralValue.Add(collateralValue, adjustedValue)
		}
	}

	// 计算借款价值
	for asset, amount := range position.Borrows {
		if amount.Cmp(big.NewInt(0)) > 0 {
			assetInfo := s.assets[asset]
			value := new(big.Int).Mul(amount, assetInfo.PriceUSD)
			borrowValue.Add(borrowValue, value)
		}
	}

	position.CollateralValue = collateralValue
	position.BorrowValue = borrowValue

	// 健康因子 = 抵押品价值 / 借款价值
	if borrowValue.Cmp(big.NewInt(0)) == 0 {
		position.HealthFactor = math.MaxFloat64
		position.AvailableBorrow = new(big.Int).Set(collateralValue)
	} else {
		position.HealthFactor = float64(collateralValue.Int64()) / float64(borrowValue.Int64())
		available := new(big.Int).Sub(collateralValue, borrowValue)
		if available.Cmp(big.NewInt(0)) < 0 {
			available = big.NewInt(0)
		}
		position.AvailableBorrow = available
	}
}

// clonePosition 克隆持仓
func (s *LendingSimulator) clonePosition(p *UserPosition) *UserPosition {
	clone := &UserPosition{
		User:     p.User,
		Supplies: make(map[string]*big.Int),
		Borrows:  make(map[string]*big.Int),
	}
	for k, v := range p.Supplies {
		clone.Supplies[k] = new(big.Int).Set(v)
	}
	for k, v := range p.Borrows {
		clone.Borrows[k] = new(big.Int).Set(v)
	}
	return clone
}

// GetAssetInfo 获取资产信息
func (s *LendingSimulator) GetAssetInfo(symbol string) map[string]interface{} {
	asset := s.assets[symbol]
	if asset == nil {
		return nil
	}

	utilization := s.CalculateUtilizationRate(asset)

	return map[string]interface{}{
		"symbol":            asset.Symbol,
		"price_usd":         asset.PriceUSD.String(),
		"collateral_factor": fmt.Sprintf("%.0f%%", asset.CollateralFactor*100),
		"liquidation_bonus": fmt.Sprintf("%.0f%%", asset.LiquidationBonus*100),
		"total_supplied":    asset.TotalSupplied.String(),
		"total_borrowed":    asset.TotalBorrowed.String(),
		"utilization":       fmt.Sprintf("%.1f%%", utilization),
		"supply_apy":        fmt.Sprintf("%.2f%%", asset.SupplyAPY),
		"borrow_apy":        fmt.Sprintf("%.2f%%", asset.BorrowAPY),
	}
}

// updateState 更新状态
func (s *LendingSimulator) updateState() {
	s.SetGlobalData("asset_count", len(s.assets))
	s.SetGlobalData("position_count", len(s.positions))
	s.SetGlobalData("action_count", len(s.actions))
	s.SetGlobalData("liquidation_count", len(s.liquidations))

	summary := fmt.Sprintf("当前有 %d 个借贷市场，%d 个用户持仓。", len(s.assets), len(s.positions))
	nextHint := "可以继续存入抵押品、借出资产或触发清算，观察健康因子如何变化。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"lending_positioning",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"asset_count": len(s.assets), "position_count": len(s.positions), "liquidation_count": len(s.liquidations)},
	)
}

func (s *LendingSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "supply":
		user, asset := "alice", "ETH"
		amount := big.NewInt(100)
		if raw, ok := params["user"].(string); ok && raw != "" {
			user = raw
		}
		if raw, ok := params["asset"].(string); ok && raw != "" {
			asset = raw
		}
		if raw, ok := params["amount"].(float64); ok && raw > 0 {
			amount = big.NewInt(int64(raw))
		}
		if err := s.Supply(user, asset, amount); err != nil {
			return nil, err
		}
		return defiActionResult("已完成一次抵押存入。", map[string]interface{}{"user": user, "asset": asset, "amount": amount.String()}, &types.ActionFeedback{
			Summary:     "用户抵押品和可借额度已经更新。",
			NextHint:    "继续执行借款，观察健康因子如何随抵押和债务变化。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"user": user, "asset": asset},
		}), nil
	case "borrow":
		user, asset := "alice", "USDC"
		amount := big.NewInt(50)
		if raw, ok := params["user"].(string); ok && raw != "" {
			user = raw
		}
		if raw, ok := params["asset"].(string); ok && raw != "" {
			asset = raw
		}
		if raw, ok := params["amount"].(float64); ok && raw > 0 {
			amount = big.NewInt(int64(raw))
		}
		if err := s.Borrow(user, asset, amount); err != nil {
			return nil, err
		}
		return defiActionResult("已完成一次借款。", map[string]interface{}{"user": user, "asset": asset, "amount": amount.String()}, &types.ActionFeedback{
			Summary:     "借款仓位和健康因子已更新。",
			NextHint:    "继续观察还款或价格变化对健康因子和清算风险的影响。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"user": user, "asset": asset},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported lending action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// LendingFactory 借贷工厂
type LendingFactory struct{}

// Create 创建演示器
func (f *LendingFactory) Create() engine.Simulator {
	return NewLendingSimulator()
}

// GetDescription 获取描述
func (f *LendingFactory) GetDescription() types.Description {
	return NewLendingSimulator().GetDescription()
}

// NewLendingFactory 创建工厂
func NewLendingFactory() *LendingFactory {
	return &LendingFactory{}
}

var _ engine.SimulatorFactory = (*LendingFactory)(nil)
