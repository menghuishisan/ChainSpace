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
// 收益聚合器演示器
// =============================================================================

// Strategy 策略
type Strategy struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Protocol       string    `json:"protocol"`   // 目标协议
	Asset          string    `json:"asset"`      // 资产
	APY            float64   `json:"apy"`        // 年化收益率
	TVL            *big.Int  `json:"tvl"`        // 总锁仓
	RiskLevel      int       `json:"risk_level"` // 风险等级 1-5
	IsActive       bool      `json:"is_active"`
	LastHarvest    time.Time `json:"last_harvest"`
	TotalHarvested *big.Int  `json:"total_harvested"`
}

// VaultDeposit 金库存款
type VaultDeposit struct {
	ID          string    `json:"id"`
	Depositor   string    `json:"depositor"`
	Amount      *big.Int  `json:"amount"`
	Shares      *big.Int  `json:"shares"`
	DepositTime time.Time `json:"deposit_time"`
}

// HarvestEvent 收割事件
type HarvestEvent struct {
	StrategyID     string    `json:"strategy_id"`
	Profit         *big.Int  `json:"profit"`
	PerformanceFee *big.Int  `json:"performance_fee"`
	ManagementFee  *big.Int  `json:"management_fee"`
	Timestamp      time.Time `json:"timestamp"`
}

// YieldAggregatorSimulator 收益聚合器演示器
// 演示DeFi收益聚合器的核心机制:
//
// 1. 自动复利
//   - 自动收割奖励并再投资
//   - 节省gas，提高收益
//
// 2. 策略切换
//   - 在多个协议间切换
//   - 追求最高收益
//
// 3. 金库机制
//   - 用户存入资产获得份额
//   - 份额价值随收益增长
//
// 参考: Yearn Finance, Beefy Finance, Harvest
type YieldAggregatorSimulator struct {
	*base.BaseSimulator
	vaultAsset     string
	totalAssets    *big.Int
	totalShares    *big.Int
	strategies     map[string]*Strategy
	deposits       map[string]*VaultDeposit
	harvests       []*HarvestEvent
	performanceFee float64 // 20%
	managementFee  float64 // 2%
	autoCompound   bool
}

// NewYieldAggregatorSimulator 创建收益聚合器演示器
func NewYieldAggregatorSimulator() *YieldAggregatorSimulator {
	sim := &YieldAggregatorSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"yield_aggregator",
			"收益聚合器演示器",
			"演示自动复利、策略切换、金库机制等收益聚合策略",
			"defi",
			types.ComponentDeFi,
		),
		strategies: make(map[string]*Strategy),
		deposits:   make(map[string]*VaultDeposit),
		harvests:   make([]*HarvestEvent, 0),
	}

	sim.AddParam(types.Param{
		Key:         "performance_fee",
		Name:        "绩效费",
		Description: "从收益中收取的绩效费(%)",
		Type:        types.ParamTypeFloat,
		Default:     20.0,
		Min:         0,
		Max:         50,
	})

	sim.AddParam(types.Param{
		Key:         "auto_compound",
		Name:        "自动复利",
		Description: "是否开启自动复利",
		Type:        types.ParamTypeBool,
		Default:     true,
	})

	return sim
}

// Init 初始化
func (s *YieldAggregatorSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.vaultAsset = "USDC"
	s.totalAssets = big.NewInt(0)
	s.totalShares = big.NewInt(0)
	s.performanceFee = 0.20
	s.managementFee = 0.02
	s.autoCompound = true

	if v, ok := config.Params["performance_fee"]; ok {
		if f, ok := v.(float64); ok {
			s.performanceFee = f / 100
		}
	}
	if v, ok := config.Params["auto_compound"]; ok {
		if b, ok := v.(bool); ok {
			s.autoCompound = b
		}
	}

	// 初始化策略
	s.strategies = map[string]*Strategy{
		"aave-usdc": {
			ID:             "aave-usdc",
			Name:           "Aave USDC Lending",
			Protocol:       "Aave",
			Asset:          "USDC",
			APY:            3.5,
			TVL:            big.NewInt(0),
			RiskLevel:      1,
			IsActive:       true,
			TotalHarvested: big.NewInt(0),
		},
		"compound-usdc": {
			ID:             "compound-usdc",
			Name:           "Compound USDC",
			Protocol:       "Compound",
			Asset:          "USDC",
			APY:            2.8,
			TVL:            big.NewInt(0),
			RiskLevel:      1,
			IsActive:       false,
			TotalHarvested: big.NewInt(0),
		},
		"curve-3pool": {
			ID:             "curve-3pool",
			Name:           "Curve 3pool LP",
			Protocol:       "Curve",
			Asset:          "USDC",
			APY:            5.2,
			TVL:            big.NewInt(0),
			RiskLevel:      2,
			IsActive:       false,
			TotalHarvested: big.NewInt(0),
		},
		"convex-3pool": {
			ID:             "convex-3pool",
			Name:           "Convex 3pool",
			Protocol:       "Convex",
			Asset:          "USDC",
			APY:            8.5,
			TVL:            big.NewInt(0),
			RiskLevel:      3,
			IsActive:       false,
			TotalHarvested: big.NewInt(0),
		},
	}

	s.deposits = make(map[string]*VaultDeposit)
	s.harvests = make([]*HarvestEvent, 0)

	s.updateState()
	return nil
}

// =============================================================================
// 金库操作
// =============================================================================

// Deposit 存款
func (s *YieldAggregatorSimulator) Deposit(depositor string, amount *big.Int) (*VaultDeposit, error) {
	// 计算份额
	var shares *big.Int
	if s.totalShares.Cmp(big.NewInt(0)) == 0 {
		shares = new(big.Int).Set(amount)
	} else {
		// shares = amount * totalShares / totalAssets
		shares = new(big.Int).Mul(amount, s.totalShares)
		shares.Div(shares, s.totalAssets)
	}

	depositID := fmt.Sprintf("deposit-%s-%d", depositor, time.Now().UnixNano())
	deposit := &VaultDeposit{
		ID:          depositID,
		Depositor:   depositor,
		Amount:      amount,
		Shares:      shares,
		DepositTime: time.Now(),
	}

	s.deposits[depositID] = deposit
	s.totalAssets.Add(s.totalAssets, amount)
	s.totalShares.Add(s.totalShares, shares)

	// 分配到最优策略
	s.allocateToStrategies()

	s.EmitEvent("deposit", "", "", map[string]interface{}{
		"deposit_id":  depositID,
		"depositor":   depositor,
		"amount":      amount.String(),
		"shares":      shares.String(),
		"share_price": s.getSharePrice(),
	})

	s.updateState()
	return deposit, nil
}

// Withdraw 取款
func (s *YieldAggregatorSimulator) Withdraw(depositID string) (map[string]interface{}, error) {
	deposit, ok := s.deposits[depositID]
	if !ok {
		return nil, fmt.Errorf("存款不存在")
	}

	// 计算当前价值
	// value = shares * totalAssets / totalShares
	value := new(big.Int).Mul(deposit.Shares, s.totalAssets)
	value.Div(value, s.totalShares)

	profit := new(big.Int).Sub(value, deposit.Amount)

	s.totalAssets.Sub(s.totalAssets, value)
	s.totalShares.Sub(s.totalShares, deposit.Shares)
	delete(s.deposits, depositID)

	// 重新分配策略
	s.allocateToStrategies()

	result := map[string]interface{}{
		"deposit_id":     depositID,
		"original":       deposit.Amount.String(),
		"value":          value.String(),
		"profit":         profit.String(),
		"profit_percent": float64(profit.Int64()) / float64(deposit.Amount.Int64()) * 100,
	}

	s.EmitEvent("withdraw", "", "", result)

	s.updateState()
	return result, nil
}

// getSharePrice 获取份额价格
func (s *YieldAggregatorSimulator) getSharePrice() float64 {
	if s.totalShares.Cmp(big.NewInt(0)) == 0 {
		return 1.0
	}
	return float64(s.totalAssets.Int64()) / float64(s.totalShares.Int64())
}

// =============================================================================
// 策略管理
// =============================================================================

// allocateToStrategies 分配资产到策略
func (s *YieldAggregatorSimulator) allocateToStrategies() {
	// 清空所有策略TVL
	for _, strategy := range s.strategies {
		strategy.TVL = big.NewInt(0)
		strategy.IsActive = false
	}

	if s.totalAssets.Cmp(big.NewInt(0)) == 0 {
		return
	}

	// 选择最优策略
	bestStrategy := s.selectBestStrategy()
	if bestStrategy != nil {
		bestStrategy.TVL = new(big.Int).Set(s.totalAssets)
		bestStrategy.IsActive = true
	}
}

// selectBestStrategy 选择最优策略
func (s *YieldAggregatorSimulator) selectBestStrategy() *Strategy {
	var best *Strategy
	bestAPY := 0.0

	for _, strategy := range s.strategies {
		if strategy.APY > bestAPY {
			bestAPY = strategy.APY
			best = strategy
		}
	}

	return best
}

// SwitchStrategy 切换策略
func (s *YieldAggregatorSimulator) SwitchStrategy(newStrategyID string) error {
	newStrategy, ok := s.strategies[newStrategyID]
	if !ok {
		return fmt.Errorf("策略不存在")
	}

	// 找到当前活跃策略
	var currentStrategy *Strategy
	for _, strategy := range s.strategies {
		if strategy.IsActive {
			currentStrategy = strategy
			break
		}
	}

	if currentStrategy != nil {
		currentStrategy.IsActive = false
		currentStrategy.TVL = big.NewInt(0)
	}

	newStrategy.IsActive = true
	newStrategy.TVL = new(big.Int).Set(s.totalAssets)

	s.EmitEvent("strategy_switched", "", "", map[string]interface{}{
		"from_strategy": func() string {
			if currentStrategy != nil {
				return currentStrategy.ID
			}
			return "none"
		}(),
		"to_strategy": newStrategyID,
		"tvl":         s.totalAssets.String(),
	})

	s.updateState()
	return nil
}

// =============================================================================
// 收割和复利
// =============================================================================

// Harvest 收割收益
func (s *YieldAggregatorSimulator) Harvest() *HarvestEvent {
	// 找到活跃策略
	var activeStrategy *Strategy
	for _, strategy := range s.strategies {
		if strategy.IsActive && strategy.TVL.Cmp(big.NewInt(0)) > 0 {
			activeStrategy = strategy
			break
		}
	}

	if activeStrategy == nil {
		return nil
	}

	// 计算收益 (模拟一天的收益)
	dailyRate := activeStrategy.APY / 365 / 100
	profit := new(big.Int).Mul(activeStrategy.TVL, big.NewInt(int64(dailyRate*10000)))
	profit.Div(profit, big.NewInt(10000))

	// 计算费用
	performanceFee := new(big.Int).Mul(profit, big.NewInt(int64(s.performanceFee*10000)))
	performanceFee.Div(performanceFee, big.NewInt(10000))

	netProfit := new(big.Int).Sub(profit, performanceFee)

	// 自动复利
	if s.autoCompound {
		s.totalAssets.Add(s.totalAssets, netProfit)
		activeStrategy.TVL.Add(activeStrategy.TVL, netProfit)
	}

	activeStrategy.TotalHarvested.Add(activeStrategy.TotalHarvested, profit)
	activeStrategy.LastHarvest = time.Now()

	event := &HarvestEvent{
		StrategyID:     activeStrategy.ID,
		Profit:         profit,
		PerformanceFee: performanceFee,
		ManagementFee:  big.NewInt(0),
		Timestamp:      time.Now(),
	}

	s.harvests = append(s.harvests, event)

	s.EmitEvent("harvest", "", "", map[string]interface{}{
		"strategy":        activeStrategy.ID,
		"profit":          profit.String(),
		"performance_fee": performanceFee.String(),
		"net_profit":      netProfit.String(),
		"auto_compound":   s.autoCompound,
		"new_share_price": s.getSharePrice(),
	})

	s.updateState()
	return event
}

// SimulateCompounding 模拟复利效果
func (s *YieldAggregatorSimulator) SimulateCompounding(principal int64, apy float64, days int, compoundFrequency int) map[string]interface{} {
	// 不复利
	simpleInterest := float64(principal) * (apy / 100) * float64(days) / 365
	simpleTotal := float64(principal) + simpleInterest

	// 复利
	rate := apy / 100
	n := float64(compoundFrequency)
	t := float64(days) / 365
	compoundTotal := float64(principal) * pow(1+rate/n, n*t)
	compoundInterest := compoundTotal - float64(principal)

	return map[string]interface{}{
		"principal":          principal,
		"apy":                fmt.Sprintf("%.1f%%", apy),
		"days":               days,
		"compound_frequency": compoundFrequency,
		"simple_interest":    fmt.Sprintf("%.2f", simpleInterest),
		"simple_total":       fmt.Sprintf("%.2f", simpleTotal),
		"compound_interest":  fmt.Sprintf("%.2f", compoundInterest),
		"compound_total":     fmt.Sprintf("%.2f", compoundTotal),
		"compound_advantage": fmt.Sprintf("%.2f%%", (compoundInterest-simpleInterest)/simpleInterest*100),
	}
}

// pow 幂函数
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	// 处理小数部分
	remainder := exp - float64(int(exp))
	if remainder > 0 {
		result *= (1 + remainder*(base-1))
	}
	return result
}

// ExplainYieldAggregator 解释收益聚合器
func (s *YieldAggregatorSimulator) ExplainYieldAggregator() map[string]interface{} {
	return map[string]interface{}{
		"purpose": "自动优化DeFi收益策略",
		"core_features": []map[string]string{
			{"feature": "自动复利", "description": "定期收割奖励并再投资"},
			{"feature": "策略切换", "description": "在多个协议间切换寻找最优收益"},
			{"feature": "Gas优化", "description": "批量操作降低用户成本"},
			{"feature": "风险分散", "description": "可分散投资多个策略"},
		},
		"vault_mechanism": map[string]string{
			"deposit":     "用户存入资产获得份额代币",
			"share_price": "份额价格 = 总资产 / 总份额",
			"growth":      "收益增加总资产，份额价格上涨",
			"withdraw":    "按份额比例取回资产",
		},
		"fee_structure": map[string]interface{}{
			"performance_fee": fmt.Sprintf("%.0f%% (从收益中收取)", s.performanceFee*100),
			"management_fee":  fmt.Sprintf("%.0f%% (年化，从TVL收取)", s.managementFee*100),
		},
		"risks": []string{
			"智能合约风险 (聚合器+底层协议)",
			"策略风险 (无常损失、清算等)",
			"Gas成本可能吃掉小额存款收益",
		},
	}
}

// GetStrategies 获取所有策略
func (s *YieldAggregatorSimulator) GetStrategies() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	// 按APY排序
	strategies := make([]*Strategy, 0)
	for _, strategy := range s.strategies {
		strategies = append(strategies, strategy)
	}
	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].APY > strategies[j].APY
	})

	for _, strategy := range strategies {
		result = append(result, map[string]interface{}{
			"id":         strategy.ID,
			"name":       strategy.Name,
			"protocol":   strategy.Protocol,
			"apy":        fmt.Sprintf("%.1f%%", strategy.APY),
			"tvl":        strategy.TVL.String(),
			"risk_level": strategy.RiskLevel,
			"is_active":  strategy.IsActive,
		})
	}

	return result
}

// updateState 更新状态
func (s *YieldAggregatorSimulator) updateState() {
	s.SetGlobalData("total_assets", s.totalAssets.String())
	s.SetGlobalData("total_shares", s.totalShares.String())
	s.SetGlobalData("share_price", s.getSharePrice())
	s.SetGlobalData("deposit_count", len(s.deposits))
	s.SetGlobalData("harvest_count", len(s.harvests))

	activeStrategy := "none"
	for _, strategy := range s.strategies {
		if strategy.IsActive {
			activeStrategy = strategy.ID
			break
		}
	}

	summary := "当前收益聚合器处于待分配状态，可以先存入资产再观察策略配置与复利效果。"
	nextHint := "先执行一次存款，再观察份额价格、活跃策略和收割记录的变化。"
	progress := 0.0
	if len(s.deposits) > 0 {
		summary = fmt.Sprintf("当前金库存有 %d 笔存款，活跃策略为 %s。", len(s.deposits), activeStrategy)
		nextHint = "继续执行 harvest 或切换策略，观察总资产和份额价格如何变化。"
		progress = 65
	}
	if len(s.harvests) > 0 {
		progress = 100
	}

	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"yield_aggregation",
		summary,
		nextHint,
		progress,
		map[string]interface{}{
			"total_assets":    s.totalAssets.String(),
			"share_price":     s.getSharePrice(),
			"deposit_count":   len(s.deposits),
			"harvest_count":   len(s.harvests),
			"active_strategy": activeStrategy,
		},
	)
}

// ExecuteAction 执行收益聚合器教学动作
func (s *YieldAggregatorSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "deposit":
		depositor := "depositor-1"
		if raw, ok := params["depositor"].(string); ok && raw != "" {
			depositor = raw
		}
		amount := big.NewInt(100000)
		if raw, ok := params["amount"].(float64); ok {
			amount = big.NewInt(int64(raw))
		}
		deposit, err := s.Deposit(depositor, amount)
		if err != nil {
			return nil, err
		}
		return defiActionResult(
			"已向收益聚合器存入资产",
			map[string]interface{}{
				"deposit_id":  deposit.ID,
				"depositor":   deposit.Depositor,
				"shares":      deposit.Shares.String(),
				"share_price": s.getSharePrice(),
			},
			&types.ActionFeedback{
				Summary:     "资产已经进入金库并转换为份额，接下来可以观察策略分配和份额价格变化。",
				NextHint:    "执行一次 harvest 或切换策略，观察复利与策略切换的影响。",
				EffectScope: "defi",
			},
		), nil
	case "harvest":
		event := s.Harvest()
		if event == nil {
			return nil, fmt.Errorf("当前没有可收割收益的活跃策略")
		}
		return defiActionResult(
			"已执行收益收割",
			map[string]interface{}{
				"strategy_id": event.StrategyID,
				"profit":      event.Profit.String(),
			},
			&types.ActionFeedback{
				Summary:     "本轮收益已经收割并按规则计入金库，可继续观察份额价格和总资产变化。",
				NextHint:    "继续执行多轮收割，比较自动复利开启和关闭时的差异。",
				EffectScope: "defi",
			},
		), nil
	case "switch_strategy":
		strategyID := ""
		if raw, ok := params["strategy_id"].(string); ok && raw != "" {
			strategyID = raw
		}
		if strategyID == "" {
			for id := range s.strategies {
				strategyID = id
				break
			}
		}
		if err := s.SwitchStrategy(strategyID); err != nil {
			return nil, err
		}
		return defiActionResult(
			"已切换收益策略",
			map[string]interface{}{
				"strategy_id": strategyID,
			},
			&types.ActionFeedback{
				Summary:     "资金已经重新分配到新的活跃策略，可继续观察收益率和风险变化。",
				NextHint:    "执行一次 harvest，比较新旧策略的收益表现。",
				EffectScope: "defi",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported yield aggregator action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// YieldAggregatorFactory 收益聚合器工厂
type YieldAggregatorFactory struct{}

// Create 创建演示器
func (f *YieldAggregatorFactory) Create() engine.Simulator {
	return NewYieldAggregatorSimulator()
}

// GetDescription 获取描述
func (f *YieldAggregatorFactory) GetDescription() types.Description {
	return NewYieldAggregatorSimulator().GetDescription()
}

// NewYieldAggregatorFactory 创建工厂
func NewYieldAggregatorFactory() *YieldAggregatorFactory {
	return &YieldAggregatorFactory{}
}

var _ engine.SimulatorFactory = (*YieldAggregatorFactory)(nil)
