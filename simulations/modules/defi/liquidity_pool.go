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
// 流动性池演示器
// =============================================================================

// LPPosition LP持仓
type LPPosition struct {
	ID            string    `json:"id"`
	Owner         string    `json:"owner"`
	Shares        *big.Int  `json:"shares"`
	InitialTokenA *big.Int  `json:"initial_token_a"`
	InitialTokenB *big.Int  `json:"initial_token_b"`
	InitialPrice  float64   `json:"initial_price"`
	EntryTime     time.Time `json:"entry_time"`
}

// ImpermanentLossResult 无常损失计算结果
type ImpermanentLossResult struct {
	PriceChange     float64  `json:"price_change_percent"`
	ImpermanentLoss float64  `json:"impermanent_loss_percent"`
	HoldValue       *big.Int `json:"hold_value"`
	LPValue         *big.Int `json:"lp_value"`
	ValueDifference *big.Int `json:"value_difference"`
	FeesEarned      *big.Int `json:"fees_earned"`
	NetPnL          *big.Int `json:"net_pnl"`
}

// PoolMetrics 池指标
type PoolMetrics struct {
	TVL         *big.Int `json:"tvl"`
	Volume24h   *big.Int `json:"volume_24h"`
	Fees24h     *big.Int `json:"fees_24h"`
	APR         float64  `json:"apr_percent"`
	Utilization float64  `json:"utilization_percent"`
	LPCount     int      `json:"lp_count"`
}

// LiquidityPoolSimulator 流动性池演示器
// 演示流动性提供的完整流程:
//
// 1. 添加流动性
//   - 按比例存入两种代币
//   - 获得LP代币代表份额
//
// 2. 移除流动性
//   - 销毁LP代币
//   - 按比例取回代币
//
// 3. 无常损失 (Impermanent Loss)
//   - 价格变化导致LP持仓价值低于单纯持有
//   - IL = 2*sqrt(price_ratio)/(1+price_ratio) - 1
//
// 4. 收益来源
//   - 交易手续费
//   - 流动性挖矿奖励
type LiquidityPoolSimulator struct {
	*base.BaseSimulator
	tokenA      string
	tokenB      string
	reserveA    *big.Int
	reserveB    *big.Int
	totalShares *big.Int
	positions   map[string]*LPPosition
	feePercent  float64
	feesA       *big.Int
	feesB       *big.Int
	metrics     *PoolMetrics
}

// NewLiquidityPoolSimulator 创建流动性池演示器
func NewLiquidityPoolSimulator() *LiquidityPoolSimulator {
	sim := &LiquidityPoolSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"liquidity_pool",
			"流动性池演示器",
			"演示流动性提供、LP代币、无常损失等核心概念",
			"defi",
			types.ComponentDeFi,
		),
		positions: make(map[string]*LPPosition),
	}

	sim.AddParam(types.Param{
		Key:         "token_a",
		Name:        "代币A",
		Description: "交易对的第一个代币",
		Type:        types.ParamTypeString,
		Default:     "ETH",
	})

	sim.AddParam(types.Param{
		Key:         "token_b",
		Name:        "代币B",
		Description: "交易对的第二个代币",
		Type:        types.ParamTypeString,
		Default:     "USDC",
	})

	sim.AddParam(types.Param{
		Key:         "fee_percent",
		Name:        "手续费率",
		Description: "交易手续费百分比",
		Type:        types.ParamTypeFloat,
		Default:     0.3,
		Min:         0.01,
		Max:         1.0,
	})

	return sim
}

// Init 初始化
func (s *LiquidityPoolSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.tokenA = "ETH"
	s.tokenB = "USDC"
	s.feePercent = 0.3

	if v, ok := config.Params["token_a"]; ok {
		if t, ok := v.(string); ok {
			s.tokenA = t
		}
	}
	if v, ok := config.Params["token_b"]; ok {
		if t, ok := v.(string); ok {
			s.tokenB = t
		}
	}
	if v, ok := config.Params["fee_percent"]; ok {
		if f, ok := v.(float64); ok {
			s.feePercent = f
		}
	}

	s.reserveA = big.NewInt(0)
	s.reserveB = big.NewInt(0)
	s.totalShares = big.NewInt(0)
	s.feesA = big.NewInt(0)
	s.feesB = big.NewInt(0)
	s.positions = make(map[string]*LPPosition)
	s.metrics = &PoolMetrics{
		TVL:       big.NewInt(0),
		Volume24h: big.NewInt(0),
		Fees24h:   big.NewInt(0),
	}

	s.updateState()
	return nil
}

// =============================================================================
// 流动性操作
// =============================================================================

// AddLiquidity 添加流动性
func (s *LiquidityPoolSimulator) AddLiquidity(owner string, amountA, amountB *big.Int) *LPPosition {
	var shares *big.Int

	if s.totalShares.Cmp(big.NewInt(0)) == 0 {
		// 首次添加: shares = sqrt(amountA * amountB)
		product := new(big.Int).Mul(amountA, amountB)
		shares = new(big.Int).Sqrt(product)

		// 锁定最小流动性 (防止精度攻击)
		minLiquidity := big.NewInt(1000)
		if shares.Cmp(minLiquidity) > 0 {
			shares.Sub(shares, minLiquidity)
			s.totalShares.Add(s.totalShares, minLiquidity) // 永久锁定
		}
	} else {
		// 按比例计算: shares = min(amountA/reserveA, amountB/reserveB) * totalShares
		shareA := new(big.Int).Div(
			new(big.Int).Mul(amountA, s.totalShares),
			s.reserveA,
		)
		shareB := new(big.Int).Div(
			new(big.Int).Mul(amountB, s.totalShares),
			s.reserveB,
		)

		if shareA.Cmp(shareB) < 0 {
			shares = shareA
		} else {
			shares = shareB
		}
	}

	// 更新储备
	s.reserveA.Add(s.reserveA, amountA)
	s.reserveB.Add(s.reserveB, amountB)
	s.totalShares.Add(s.totalShares, shares)

	// 创建持仓记录
	positionID := fmt.Sprintf("lp-%s-%d", owner, time.Now().UnixNano())
	position := &LPPosition{
		ID:            positionID,
		Owner:         owner,
		Shares:        shares,
		InitialTokenA: new(big.Int).Set(amountA),
		InitialTokenB: new(big.Int).Set(amountB),
		InitialPrice:  float64(amountB.Int64()) / float64(amountA.Int64()),
		EntryTime:     time.Now(),
	}

	s.positions[positionID] = position

	s.EmitEvent("liquidity_added", "", "", map[string]interface{}{
		"position_id": positionID,
		"owner":       owner,
		"amount_a":    amountA.String(),
		"amount_b":    amountB.String(),
		"shares":      shares.String(),
		"share_of_pool": fmt.Sprintf("%.2f%%",
			float64(shares.Int64())/float64(s.totalShares.Int64())*100),
	})

	s.updateMetrics()
	s.updateState()
	return position
}

// RemoveLiquidity 移除流动性
func (s *LiquidityPoolSimulator) RemoveLiquidity(positionID string, sharePercent float64) map[string]interface{} {
	position, ok := s.positions[positionID]
	if !ok {
		return map[string]interface{}{"error": "持仓不存在"}
	}

	// 计算要移除的份额
	sharesToRemove := new(big.Int).Div(
		new(big.Int).Mul(position.Shares, big.NewInt(int64(sharePercent*100))),
		big.NewInt(10000),
	)

	// 计算返还的代币
	amountA := new(big.Int).Div(
		new(big.Int).Mul(sharesToRemove, s.reserveA),
		s.totalShares,
	)
	amountB := new(big.Int).Div(
		new(big.Int).Mul(sharesToRemove, s.reserveB),
		s.totalShares,
	)

	// 更新储备
	s.reserveA.Sub(s.reserveA, amountA)
	s.reserveB.Sub(s.reserveB, amountB)
	s.totalShares.Sub(s.totalShares, sharesToRemove)
	position.Shares.Sub(position.Shares, sharesToRemove)

	// 如果份额为0，删除持仓
	if position.Shares.Cmp(big.NewInt(0)) == 0 {
		delete(s.positions, positionID)
	}

	// 计算无常损失
	il := s.CalculateImpermanentLoss(position)

	result := map[string]interface{}{
		"position_id":      positionID,
		"shares_removed":   sharesToRemove.String(),
		"amount_a":         amountA.String(),
		"amount_b":         amountB.String(),
		"impermanent_loss": il,
	}

	s.EmitEvent("liquidity_removed", "", "", result)

	s.updateMetrics()
	s.updateState()
	return result
}

// =============================================================================
// 无常损失计算
// =============================================================================

// CalculateImpermanentLoss 计算无常损失
func (s *LiquidityPoolSimulator) CalculateImpermanentLoss(position *LPPosition) *ImpermanentLossResult {
	if s.reserveA.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	currentPrice := float64(s.reserveB.Int64()) / float64(s.reserveA.Int64())
	priceRatio := currentPrice / position.InitialPrice

	// 无常损失公式: IL = 2*sqrt(r)/(1+r) - 1
	// 其中 r = 当前价格/初始价格
	sqrtRatio := math.Sqrt(priceRatio)
	ilPercent := (2*sqrtRatio/(1+priceRatio) - 1) * 100

	// 计算持有价值 vs LP价值
	// 持有价值 = initialA * currentPriceA + initialB
	holdValueA := position.InitialTokenA
	holdValueB := position.InitialTokenB

	// LP当前价值
	lpAmountA := new(big.Int).Div(
		new(big.Int).Mul(position.Shares, s.reserveA),
		s.totalShares,
	)
	lpAmountB := new(big.Int).Div(
		new(big.Int).Mul(position.Shares, s.reserveB),
		s.totalShares,
	)

	// 转换为同一计价单位 (TokenB)
	holdValue := new(big.Int).Add(
		new(big.Int).Mul(holdValueA, big.NewInt(int64(currentPrice))),
		holdValueB,
	)
	lpValue := new(big.Int).Add(
		new(big.Int).Mul(lpAmountA, big.NewInt(int64(currentPrice))),
		lpAmountB,
	)

	valueDiff := new(big.Int).Sub(lpValue, holdValue)

	return &ImpermanentLossResult{
		PriceChange:     (priceRatio - 1) * 100,
		ImpermanentLoss: ilPercent,
		HoldValue:       holdValue,
		LPValue:         lpValue,
		ValueDifference: valueDiff,
		FeesEarned:      big.NewInt(0), // 简化
		NetPnL:          valueDiff,
	}
}

// ExplainImpermanentLoss 解释无常损失
func (s *LiquidityPoolSimulator) ExplainImpermanentLoss() map[string]interface{} {
	return map[string]interface{}{
		"definition": "无常损失是指LP持仓价值与单纯持有代币相比的差额",
		"formula":    "IL = 2*sqrt(price_ratio)/(1+price_ratio) - 1",
		"key_points": []string{
			"只有在价格变化时才会发生",
			"价格恢复到初始值时，无常损失消失",
			"价格变化越大，无常损失越大",
			"无常损失是对称的：涨100%和跌50%的IL相同",
		},
		"examples": []map[string]interface{}{
			{"price_change": "±25%", "impermanent_loss": "0.6%"},
			{"price_change": "±50%", "impermanent_loss": "2.0%"},
			{"price_change": "±75%", "impermanent_loss": "3.8%"},
			{"price_change": "±100%", "impermanent_loss": "5.7%"},
			{"price_change": "±200%", "impermanent_loss": "13.4%"},
			{"price_change": "±300%", "impermanent_loss": "20.0%"},
			{"price_change": "±400%", "impermanent_loss": "25.5%"},
		},
		"mitigation": []string{
			"选择相关性高的交易对 (如稳定币对)",
			"使用集中流动性 (Uniswap V3)",
			"依靠交易费用覆盖无常损失",
			"使用无常损失保险",
		},
		"why_impermanent": "如果价格恢复到初始水平，损失会消失，所以叫'无常'",
	}
}

// SimulateILScenarios 模拟不同价格变化下的无常损失
func (s *LiquidityPoolSimulator) SimulateILScenarios() []map[string]interface{} {
	scenarios := make([]map[string]interface{}, 0)

	priceChanges := []float64{-75, -50, -25, 0, 25, 50, 75, 100, 200, 300, 400}

	for _, change := range priceChanges {
		priceRatio := 1 + change/100
		if priceRatio <= 0 {
			continue
		}

		sqrtRatio := math.Sqrt(priceRatio)
		il := (2*sqrtRatio/(1+priceRatio) - 1) * 100

		scenarios = append(scenarios, map[string]interface{}{
			"price_change_percent": change,
			"price_ratio":          priceRatio,
			"impermanent_loss":     fmt.Sprintf("%.2f%%", math.Abs(il)),
			"lp_vs_hold":           fmt.Sprintf("%.2f%%", il),
		})
	}

	return scenarios
}

// =============================================================================
// 池指标
// =============================================================================

// updateMetrics 更新池指标
func (s *LiquidityPoolSimulator) updateMetrics() {
	// 假设TokenB是稳定币，TVL = 2 * reserveB
	s.metrics.TVL = new(big.Int).Mul(s.reserveB, big.NewInt(2))
	s.metrics.LPCount = len(s.positions)

	// APR计算: (fees_24h / tvl) * 365 * 100
	if s.metrics.TVL.Cmp(big.NewInt(0)) > 0 {
		dailyFees := float64(s.metrics.Fees24h.Int64())
		tvl := float64(s.metrics.TVL.Int64())
		s.metrics.APR = (dailyFees / tvl) * 365 * 100
	}
}

// GetPoolInfo 获取池信息
func (s *LiquidityPoolSimulator) GetPoolInfo() map[string]interface{} {
	price := float64(0)
	if s.reserveA.Cmp(big.NewInt(0)) > 0 {
		price = float64(s.reserveB.Int64()) / float64(s.reserveA.Int64())
	}

	return map[string]interface{}{
		"token_a":      s.tokenA,
		"token_b":      s.tokenB,
		"reserve_a":    s.reserveA.String(),
		"reserve_b":    s.reserveB.String(),
		"price":        price,
		"total_shares": s.totalShares.String(),
		"fee_percent":  s.feePercent,
		"lp_count":     len(s.positions),
		"tvl":          s.metrics.TVL.String(),
		"apr":          fmt.Sprintf("%.2f%%", s.metrics.APR),
	}
}

// GetPositions 获取所有持仓
func (s *LiquidityPoolSimulator) GetPositions() []*LPPosition {
	positions := make([]*LPPosition, 0, len(s.positions))
	for _, p := range s.positions {
		positions = append(positions, p)
	}
	return positions
}

// updateState 更新状态
func (s *LiquidityPoolSimulator) updateState() {
	s.SetGlobalData("token_pair", fmt.Sprintf("%s/%s", s.tokenA, s.tokenB))
	s.SetGlobalData("reserve_a", s.reserveA.String())
	s.SetGlobalData("reserve_b", s.reserveB.String())
	s.SetGlobalData("total_shares", s.totalShares.String())
	s.SetGlobalData("lp_count", len(s.positions))

	summary := fmt.Sprintf("当前池子为 %s/%s，已有 %d 个 LP 持仓。", s.tokenA, s.tokenB, len(s.positions))
	nextHint := "可以继续添加或移除流动性，比较 LP 份额、TVL 和无常损失场景变化。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"lp_management",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"lp_count": len(s.positions), "total_shares": s.totalShares.String()},
	)
}

func (s *LiquidityPoolSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "add_liquidity":
		owner := "lp-1"
		if raw, ok := params["owner"].(string); ok && raw != "" {
			owner = raw
		}
		amountA := big.NewInt(1000)
		amountB := big.NewInt(1000)
		if raw, ok := params["amount_a"].(float64); ok && raw > 0 {
			amountA = big.NewInt(int64(raw))
		}
		if raw, ok := params["amount_b"].(float64); ok && raw > 0 {
			amountB = big.NewInt(int64(raw))
		}
		position := s.AddLiquidity(owner, amountA, amountB)
		return defiActionResult("已添加一笔流动性头寸。", map[string]interface{}{"position": position}, &types.ActionFeedback{
			Summary:     "LP 头寸已经建立，份额和储备均已更新。",
			NextHint:    "继续观察价格变化后该头寸的无常损失与持有收益差异。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"position_id": position.ID, "lp_count": len(s.positions)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported liquidity pool action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// LiquidityPoolFactory 流动性池工厂
type LiquidityPoolFactory struct{}

// Create 创建演示器
func (f *LiquidityPoolFactory) Create() engine.Simulator {
	return NewLiquidityPoolSimulator()
}

// GetDescription 获取描述
func (f *LiquidityPoolFactory) GetDescription() types.Description {
	return NewLiquidityPoolSimulator().GetDescription()
}

// NewLiquidityPoolFactory 创建工厂
func NewLiquidityPoolFactory() *LiquidityPoolFactory {
	return &LiquidityPoolFactory{}
}

var _ engine.SimulatorFactory = (*LiquidityPoolFactory)(nil)
