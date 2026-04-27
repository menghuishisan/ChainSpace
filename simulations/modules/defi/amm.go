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

type AMMType string

const (
	AMMConstantProduct AMMType = "constant_product"
	AMMConstantSum     AMMType = "constant_sum"
	AMMStableSwap      AMMType = "stable_swap"
	AMMConcentrated    AMMType = "concentrated"
)

type PoolState struct {
	TokenA        string   `json:"token_a"`
	TokenB        string   `json:"token_b"`
	ReserveA      *big.Int `json:"reserve_a"`
	ReserveB      *big.Int `json:"reserve_b"`
	TotalLPShares *big.Int `json:"total_lp_shares"`
	Fee           float64  `json:"fee_percent"`
	K             *big.Int `json:"k"`
	Price         float64  `json:"price"`
	VolumeUSD     *big.Int `json:"volume_usd"`
	FeesCollected *big.Int `json:"fees_collected"`
}

type SwapResult struct {
	TokenIn         string   `json:"token_in"`
	TokenOut        string   `json:"token_out"`
	AmountIn        *big.Int `json:"amount_in"`
	AmountOut       *big.Int `json:"amount_out"`
	Fee             *big.Int `json:"fee"`
	PriceImpact     float64  `json:"price_impact_percent"`
	EffectivePrice  float64  `json:"effective_price"`
	SpotPriceBefore float64  `json:"spot_price_before"`
	SpotPriceAfter  float64  `json:"spot_price_after"`
	Slippage        float64  `json:"slippage_percent"`
}

type LiquidityAction struct {
	Type       string    `json:"type"`
	AmountA    *big.Int  `json:"amount_a"`
	AmountB    *big.Int  `json:"amount_b"`
	LPShares   *big.Int  `json:"lp_shares"`
	SharePrice *big.Int  `json:"share_price"`
	Timestamp  time.Time `json:"timestamp"`
}

type AMMSimulator struct {
	*base.BaseSimulator
	pool        *PoolState
	ammType     AMMType
	swapHistory []*SwapResult
	lpActions   []*LiquidityAction
	amplifier   float64
}

func NewAMMSimulator() *AMMSimulator {
	sim := &AMMSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"amm",
			"AMM 自动做市商演示器",
			"演示恒定乘积、恒定和、StableSwap 等自动做市曲线的工作机制。",
			"defi",
			types.ComponentDeFi,
		),
		swapHistory: make([]*SwapResult, 0),
		lpActions:   make([]*LiquidityAction, 0),
		amplifier:   100,
	}

	sim.AddParam(types.Param{
		Key:         "amm_type",
		Name:        "AMM 类型",
		Description: "自动做市商曲线类型。",
		Type:        types.ParamTypeSelect,
		Default:     "constant_product",
		Options: []types.Option{
			{Label: "恒定乘积 (Uniswap)", Value: "constant_product"},
			{Label: "恒定和", Value: "constant_sum"},
			{Label: "StableSwap (Curve)", Value: "stable_swap"},
		},
	})
	sim.AddParam(types.Param{
		Key:         "initial_liquidity",
		Name:        "初始流动性",
		Description: "每种代币的初始数量。",
		Type:        types.ParamTypeInt,
		Default:     1000000,
		Min:         1000,
		Max:         100000000,
	})
	sim.AddParam(types.Param{
		Key:         "fee_percent",
		Name:        "手续费率",
		Description: "交易手续费百分比。",
		Type:        types.ParamTypeFloat,
		Default:     0.3,
		Min:         0,
		Max:         1,
	})

	return sim
}

func (s *AMMSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	initialLiquidity := int64(1000000)
	feePercent := 0.3
	s.ammType = AMMConstantProduct
	s.amplifier = 100

	if v, ok := config.Params["initial_liquidity"]; ok {
		if n, ok := v.(float64); ok {
			initialLiquidity = int64(n)
		}
	}
	if v, ok := config.Params["fee_percent"]; ok {
		if n, ok := v.(float64); ok {
			feePercent = n
		}
	}
	if v, ok := config.Params["amm_type"]; ok {
		if t, ok := v.(string); ok {
			s.ammType = AMMType(t)
		}
	}

	reserveA := big.NewInt(initialLiquidity)
	reserveB := big.NewInt(initialLiquidity)
	k := new(big.Int).Mul(reserveA, reserveB)

	s.pool = &PoolState{
		TokenA:        "ETH",
		TokenB:        "USDC",
		ReserveA:      reserveA,
		ReserveB:      reserveB,
		TotalLPShares: big.NewInt(initialLiquidity),
		Fee:           feePercent,
		K:             k,
		Price:         float64(reserveB.Int64()) / float64(reserveA.Int64()),
		VolumeUSD:     big.NewInt(0),
		FeesCollected: big.NewInt(0),
	}

	s.swapHistory = make([]*SwapResult, 0)
	s.lpActions = make([]*LiquidityAction, 0)
	s.updateState()
	s.syncTeachingState()
	return nil
}

func (s *AMMSimulator) ExplainAMMCurves() map[string]interface{} {
	return map[string]interface{}{
		"constant_product": map[string]interface{}{
			"formula": "x * y = k",
			"name":    "恒定乘积做市商",
			"used_by": []string{"Uniswap V1/V2", "SushiSwap", "PancakeSwap"},
			"price":   "dy/dx = y/x",
			"properties": []string{
				"永远不会耗尽流动性",
				"交易规模越大，滑点越明显",
				"适合波动资产",
			},
		},
		"constant_sum": map[string]interface{}{
			"formula": "x + y = k",
			"name":    "恒定和做市商",
			"price":   "始终接近 1:1",
			"properties": []string{
				"低滑点",
				"可能被单边耗尽",
				"更适合教学说明，不适合真实大规模流动市场",
			},
		},
		"stable_swap": map[string]interface{}{
			"formula": "A*n^n*sum(x_i) + D = A*D*n^n + D^(n+1)/(n^n*prod(x_i))",
			"name":    "StableSwap",
			"used_by": []string{"Curve", "Saddle"},
			"properties": []string{
				"锚定附近滑点更低",
				"偏离锚定后逐渐接近恒定乘积",
				"适合稳定币兑换",
			},
		},
	}
}

func (s *AMMSimulator) Swap(tokenIn string, amountIn *big.Int) *SwapResult {
	var reserveIn, reserveOut *big.Int
	var tokenOut string

	if tokenIn == s.pool.TokenA {
		reserveIn = new(big.Int).Set(s.pool.ReserveA)
		reserveOut = new(big.Int).Set(s.pool.ReserveB)
		tokenOut = s.pool.TokenB
	} else {
		reserveIn = new(big.Int).Set(s.pool.ReserveB)
		reserveOut = new(big.Int).Set(s.pool.ReserveA)
		tokenOut = s.pool.TokenA
	}

	spotPriceBefore := float64(reserveOut.Int64()) / float64(reserveIn.Int64())
	feeAmount := new(big.Int).Div(
		new(big.Int).Mul(amountIn, big.NewInt(int64(s.pool.Fee*1000))),
		big.NewInt(100000),
	)
	amountInAfterFee := new(big.Int).Sub(amountIn, feeAmount)

	var amountOut *big.Int
	switch s.ammType {
	case AMMConstantProduct, AMMConcentrated:
		amountOut = s.calculateConstantProduct(reserveIn, reserveOut, amountInAfterFee)
	case AMMConstantSum:
		amountOut = s.calculateConstantSum(amountInAfterFee)
	case AMMStableSwap:
		amountOut = s.calculateStableSwap(reserveIn, reserveOut, amountInAfterFee)
	default:
		amountOut = s.calculateConstantProduct(reserveIn, reserveOut, amountInAfterFee)
	}

	if tokenIn == s.pool.TokenA {
		s.pool.ReserveA.Add(s.pool.ReserveA, amountIn)
		s.pool.ReserveB.Sub(s.pool.ReserveB, amountOut)
	} else {
		s.pool.ReserveB.Add(s.pool.ReserveB, amountIn)
		s.pool.ReserveA.Sub(s.pool.ReserveA, amountOut)
	}

	s.pool.FeesCollected.Add(s.pool.FeesCollected, feeAmount)
	s.pool.K = new(big.Int).Mul(s.pool.ReserveA, s.pool.ReserveB)
	s.pool.Price = float64(s.pool.ReserveB.Int64()) / float64(s.pool.ReserveA.Int64())

	spotPriceAfter := float64(s.pool.ReserveB.Int64()) / float64(s.pool.ReserveA.Int64())
	if tokenIn == s.pool.TokenB {
		spotPriceBefore = 1 / spotPriceBefore
		spotPriceAfter = 1 / spotPriceAfter
	}

	effectivePrice := float64(amountIn.Int64()) / math.Max(float64(amountOut.Int64()), 1)
	priceImpact := math.Abs(spotPriceAfter-spotPriceBefore) / math.Max(spotPriceBefore, 0.000001) * 100
	slippage := math.Abs(effectivePrice-spotPriceBefore) / math.Max(spotPriceBefore, 0.000001) * 100

	result := &SwapResult{
		TokenIn:         tokenIn,
		TokenOut:        tokenOut,
		AmountIn:        amountIn,
		AmountOut:       amountOut,
		Fee:             feeAmount,
		PriceImpact:     priceImpact,
		EffectivePrice:  effectivePrice,
		SpotPriceBefore: spotPriceBefore,
		SpotPriceAfter:  spotPriceAfter,
		Slippage:        slippage,
	}

	s.swapHistory = append(s.swapHistory, result)
	s.EmitEvent("swap_executed", "", "", map[string]interface{}{
		"token_in":             tokenIn,
		"token_out":            tokenOut,
		"amount_in":            amountIn.String(),
		"amount_out":           amountOut.String(),
		"fee":                  feeAmount.String(),
		"price_impact_percent": priceImpact,
		"slippage_percent":     slippage,
	})

	s.updateState()
	s.syncTeachingState()
	return result
}

func (s *AMMSimulator) calculateConstantProduct(reserveIn, reserveOut, amountIn *big.Int) *big.Int {
	numerator := new(big.Int).Mul(reserveOut, amountIn)
	denominator := new(big.Int).Add(reserveIn, amountIn)
	return new(big.Int).Div(numerator, denominator)
}

func (s *AMMSimulator) calculateConstantSum(amountIn *big.Int) *big.Int {
	return new(big.Int).Set(amountIn)
}

func (s *AMMSimulator) calculateStableSwap(reserveIn, reserveOut, amountIn *big.Int) *big.Int {
	xp := []float64{float64(reserveIn.Int64()), float64(reserveOut.Int64())}
	A := s.amplifier
	D := s.calculateD(xp, A)
	newXp0 := xp[0] + float64(amountIn.Int64())
	newY := s.calculateY(newXp0, D, A)
	amountOut := xp[1] - newY
	if amountOut < 0 {
		amountOut = 0
	}
	return big.NewInt(int64(amountOut))
}

func (s *AMMSimulator) calculateD(xp []float64, A float64) float64 {
	n := float64(len(xp))
	sum := 0.0
	for _, x := range xp {
		sum += x
	}
	if sum == 0 {
		return 0
	}

	D := sum
	Ann := A * math.Pow(n, n)
	for i := 0; i < 255; i++ {
		Dp := D
		for _, x := range xp {
			Dp = Dp * D / (x * n)
		}
		prev := D
		numerator := (Ann*sum/A + Dp*n) * D
		denominator := (Ann/A-1)*D + (n+1)*Dp
		D = numerator / denominator
		if math.Abs(D-prev) <= 1 {
			break
		}
	}
	return D
}

func (s *AMMSimulator) calculateY(x float64, D float64, A float64) float64 {
	n := 2.0
	Ann := A * math.Pow(n, n)
	c := D * D * D / (4 * x * Ann)
	b := x + D/Ann
	y := D
	for i := 0; i < 255; i++ {
		prev := y
		y = (y*y + c) / (2*y + b - D)
		if math.Abs(y-prev) <= 1 {
			break
		}
	}
	return y
}

func (s *AMMSimulator) ExplainStableSwap() map[string]interface{} {
	return map[string]interface{}{
		"invariant": "A*n^n*Σx_i + D = A*D*n^n + D^(n+1)/(n^n*Πx_i)",
		"parameters": map[string]interface{}{
			"A": "放大系数，控制曲线形状",
			"D": "当前池子的近似不变量",
		},
	}
}

func (s *AMMSimulator) AddLiquidity(amountA, amountB *big.Int) *LiquidityAction {
	var lpShares *big.Int
	if s.pool.TotalLPShares.Sign() == 0 {
		product := new(big.Int).Mul(amountA, amountB)
		lpShares = new(big.Int).Sqrt(product)
	} else {
		shareA := new(big.Int).Div(new(big.Int).Mul(amountA, s.pool.TotalLPShares), s.pool.ReserveA)
		shareB := new(big.Int).Div(new(big.Int).Mul(amountB, s.pool.TotalLPShares), s.pool.ReserveB)
		if shareA.Cmp(shareB) < 0 {
			lpShares = shareA
		} else {
			lpShares = shareB
		}
	}

	s.pool.ReserveA.Add(s.pool.ReserveA, amountA)
	s.pool.ReserveB.Add(s.pool.ReserveB, amountB)
	s.pool.TotalLPShares.Add(s.pool.TotalLPShares, lpShares)
	s.pool.K = new(big.Int).Mul(s.pool.ReserveA, s.pool.ReserveB)

	action := &LiquidityAction{
		Type:      "add",
		AmountA:   amountA,
		AmountB:   amountB,
		LPShares:  lpShares,
		Timestamp: time.Now(),
	}
	s.lpActions = append(s.lpActions, action)

	s.EmitEvent("liquidity_added", "", "", map[string]interface{}{
		"amount_a":  amountA.String(),
		"amount_b":  amountB.String(),
		"lp_shares": lpShares.String(),
	})

	s.updateState()
	s.syncTeachingState()
	return action
}

func (s *AMMSimulator) RemoveLiquidity(lpShares *big.Int) *LiquidityAction {
	amountA := new(big.Int).Div(new(big.Int).Mul(lpShares, s.pool.ReserveA), s.pool.TotalLPShares)
	amountB := new(big.Int).Div(new(big.Int).Mul(lpShares, s.pool.ReserveB), s.pool.TotalLPShares)

	s.pool.ReserveA.Sub(s.pool.ReserveA, amountA)
	s.pool.ReserveB.Sub(s.pool.ReserveB, amountB)
	s.pool.TotalLPShares.Sub(s.pool.TotalLPShares, lpShares)
	s.pool.K = new(big.Int).Mul(s.pool.ReserveA, s.pool.ReserveB)

	action := &LiquidityAction{
		Type:      "remove",
		AmountA:   amountA,
		AmountB:   amountB,
		LPShares:  lpShares,
		Timestamp: time.Now(),
	}
	s.lpActions = append(s.lpActions, action)

	s.EmitEvent("liquidity_removed", "", "", map[string]interface{}{
		"amount_a":  amountA.String(),
		"amount_b":  amountB.String(),
		"lp_shares": lpShares.String(),
	})

	s.updateState()
	s.syncTeachingState()
	return action
}

func (s *AMMSimulator) CalculatePriceImpact(tokenIn string, amountIn *big.Int) map[string]interface{} {
	var reserveIn, reserveOut *big.Int
	if tokenIn == s.pool.TokenA {
		reserveIn = s.pool.ReserveA
		reserveOut = s.pool.ReserveB
	} else {
		reserveIn = s.pool.ReserveB
		reserveOut = s.pool.ReserveA
	}

	spotPrice := float64(reserveOut.Int64()) / math.Max(float64(reserveIn.Int64()), 1)
	impacts := make([]map[string]interface{}, 0)
	for _, percent := range []float64{0.1, 0.5, 1, 2, 5, 10} {
		tradeSize := new(big.Int).Div(
			new(big.Int).Mul(reserveIn, big.NewInt(int64(percent*100))),
			big.NewInt(10000),
		)
		amountOut := s.calculateConstantProduct(reserveIn, reserveOut, tradeSize)
		effectivePrice := float64(tradeSize.Int64()) / math.Max(float64(amountOut.Int64()), 1)
		impact := (effectivePrice - spotPrice) / math.Max(spotPrice, 0.000001) * 100
		impacts = append(impacts, map[string]interface{}{
			"trade_percent": fmt.Sprintf("%.1f%%", percent),
			"trade_size":    tradeSize.String(),
			"amount_out":    amountOut.String(),
			"price_impact":  fmt.Sprintf("%.2f%%", impact),
		})
	}
	return map[string]interface{}{
		"spot_price":  spotPrice,
		"reserve_in":  reserveIn.String(),
		"reserve_out": reserveOut.String(),
		"impacts":     impacts,
	}
}

func (s *AMMSimulator) GetPoolStats() map[string]interface{} {
	return map[string]interface{}{
		"amm_type":       s.ammType,
		"token_a":        s.pool.TokenA,
		"token_b":        s.pool.TokenB,
		"reserve_a":      s.pool.ReserveA.String(),
		"reserve_b":      s.pool.ReserveB.String(),
		"price":          s.pool.Price,
		"k":              s.pool.K.String(),
		"total_lp":       s.pool.TotalLPShares.String(),
		"fee_percent":    s.pool.Fee,
		"fees_collected": s.pool.FeesCollected.String(),
		"swap_count":     len(s.swapHistory),
	}
}

func (s *AMMSimulator) updateState() {
	s.SetGlobalData("amm_type", string(s.ammType))
	s.SetGlobalData("reserve_a", s.pool.ReserveA.String())
	s.SetGlobalData("reserve_b", s.pool.ReserveB.String())
	s.SetGlobalData("price", s.pool.Price)
	s.SetGlobalData("fee_percent", s.pool.Fee)
	s.SetGlobalData("constant_product", s.pool.K.String())
	s.SetGlobalData("total_lp", s.pool.TotalLPShares.String())
	s.SetGlobalData("swap_count", len(s.swapHistory))
}

type AMMFactory struct{}

func (f *AMMFactory) Create() engine.Simulator {
	return NewAMMSimulator()
}

func (f *AMMFactory) GetDescription() types.Description {
	return NewAMMSimulator().GetDescription()
}

func NewAMMFactory() *AMMFactory {
	return &AMMFactory{}
}

func (s *AMMSimulator) syncTeachingState() {
	summary := "当前池子处于静态状态，可以发起兑换或调整流动性观察曲线变化。"
	nextHint := "先执行一次兑换，观察曲线位置、储备和滑点如何同时变化。"
	progress := 0.0
	if len(s.swapHistory) > 0 {
		latest := s.swapHistory[len(s.swapHistory)-1]
		summary = fmt.Sprintf(
			"最近一次兑换中，%s -> %s 让池子价格从 %.4f 变为 %.4f。",
			latest.TokenIn,
			latest.TokenOut,
			latest.SpotPriceBefore,
			latest.SpotPriceAfter,
		)
		nextHint = "继续观察下一次交易是否沿着同一曲线继续推动池子状态。"
		progress = 75
	}
	if len(s.lpActions) > 0 {
		progress = 100
	}

	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"amm_repricing",
		summary,
		nextHint,
		progress,
		map[string]interface{}{
			"price":      s.pool.Price,
			"reserve_a":  s.pool.ReserveA.String(),
			"reserve_b":  s.pool.ReserveB.String(),
			"swap_count": len(s.swapHistory),
			"lp_actions": len(s.lpActions),
		},
	)
}

var _ engine.SimulatorFactory = (*AMMFactory)(nil)

// ExecuteAction 执行 AMM 教学动作。
func (s *AMMSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "swap":
		tokenIn := "ETH"
		if raw, ok := params["token_in"].(string); ok && raw != "" {
			tokenIn = raw
		}
		amount := big.NewInt(10000)
		if raw, ok := params["amount_in"].(float64); ok {
			amount = big.NewInt(int64(raw))
		}
		result := s.Swap(tokenIn, amount)
		s.updateState()
		return defiActionResult(
			"已执行一次 AMM 兑换",
			map[string]interface{}{
				"amount_out": result.AmountOut.String(),
			},
			&types.ActionFeedback{
				Summary:     "输入资产已进入池子，储备和价格曲线已发生变化。",
				NextHint:    "观察输出数量、储备重配和价格偏移是否符合预期。",
				EffectScope: "defi",
				ResultState: map[string]interface{}{"status": "swap_completed"},
			},
		), nil
	case "add_liquidity":
		amountA := big.NewInt(20000)
		amountB := big.NewInt(20000)
		if raw, ok := params["amount_a"].(float64); ok {
			amountA = big.NewInt(int64(raw))
		}
		if raw, ok := params["amount_b"].(float64); ok {
			amountB = big.NewInt(int64(raw))
		}
		s.AddLiquidity(amountA, amountB)
		s.updateState()
		return defiActionResult(
			"已添加流动性",
			nil,
			&types.ActionFeedback{
				Summary:     "新的流动性已经注入池子。",
				NextHint:    "观察储备比例、LP 份额和价格稳定性如何变化。",
				EffectScope: "defi",
				ResultState: map[string]interface{}{"status": "liquidity_added"},
			},
		), nil
	case "remove_liquidity":
		lpShares := big.NewInt(10000)
		if raw, ok := params["lp_shares"].(float64); ok {
			lpShares = big.NewInt(int64(raw))
		}
		s.RemoveLiquidity(lpShares)
		s.updateState()
		return defiActionResult(
			"已移除流动性",
			nil,
			&types.ActionFeedback{
				Summary:     "部分流动性已经从池子中撤出。",
				NextHint:    "观察池子深度、价格敏感度和剩余 LP 份额的变化。",
				EffectScope: "defi",
				ResultState: map[string]interface{}{"status": "liquidity_removed"},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported amm action: %s", action)
	}
}
