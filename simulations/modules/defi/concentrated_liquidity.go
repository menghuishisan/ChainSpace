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
// 集中流动性演示器 (Uniswap V3)
// =============================================================================

// PriceRange 价格区间
type PriceRange struct {
	TickLower  int64   `json:"tick_lower"`
	TickUpper  int64   `json:"tick_upper"`
	PriceLower float64 `json:"price_lower"`
	PriceUpper float64 `json:"price_upper"`
}

// ConcentratedPosition 集中流动性持仓
type ConcentratedPosition struct {
	ID          string      `json:"id"`
	Owner       string      `json:"owner"`
	TokenA      string      `json:"token_a"`
	TokenB      string      `json:"token_b"`
	Liquidity   *big.Int    `json:"liquidity"`
	Range       *PriceRange `json:"range"`
	DepositedA  *big.Int    `json:"deposited_a"`
	DepositedB  *big.Int    `json:"deposited_b"`
	FeesEarnedA *big.Int    `json:"fees_earned_a"`
	FeesEarnedB *big.Int    `json:"fees_earned_b"`
	InRange     bool        `json:"in_range"`
	CreatedAt   time.Time   `json:"created_at"`
}

// Tick Tick数据
type Tick struct {
	Index          int64    `json:"index"`
	LiquidityGross *big.Int `json:"liquidity_gross"`
	LiquidityNet   *big.Int `json:"liquidity_net"`
	Initialized    bool     `json:"initialized"`
}

// ConcentratedLiquiditySimulator 集中流动性演示器
// 演示Uniswap V3的集中流动性机制:
//
// 1. 价格区间
//   - LP在指定价格区间提供流动性
//   - 资本效率大幅提升
//
// 2. Tick系统
//   - 价格空间离散化
//   - tick间距决定精度
//
// 3. 范围订单
//   - 超出区间的流动性不活跃
//   - 可实现限价单效果
//
// 公式:
// - L = sqrt(xy) (虚拟流动性)
// - x = L * (sqrt(pb) - sqrt(p)) / (sqrt(p) * sqrt(pb))
// - y = L * (sqrt(p) - sqrt(pa))
type ConcentratedLiquiditySimulator struct {
	*base.BaseSimulator
	tokenA       string
	tokenB       string
	currentPrice float64
	currentTick  int64
	sqrtPriceX96 *big.Int
	liquidity    *big.Int
	positions    map[string]*ConcentratedPosition
	ticks        map[int64]*Tick
	tickSpacing  int64
	feeRate      float64
}

// NewConcentratedLiquiditySimulator 创建集中流动性演示器
func NewConcentratedLiquiditySimulator() *ConcentratedLiquiditySimulator {
	sim := &ConcentratedLiquiditySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"concentrated_liquidity",
			"集中流动性演示器",
			"演示Uniswap V3的集中流动性、价格区间、Tick系统等机制",
			"defi",
			types.ComponentDeFi,
		),
		positions: make(map[string]*ConcentratedPosition),
		ticks:     make(map[int64]*Tick),
	}

	sim.AddParam(types.Param{
		Key:         "initial_price",
		Name:        "初始价格",
		Description: "TokenB/TokenA的初始价格",
		Type:        types.ParamTypeFloat,
		Default:     2000,
		Min:         1,
		Max:         100000,
	})

	sim.AddParam(types.Param{
		Key:         "fee_tier",
		Name:        "费率等级",
		Description: "交易手续费率(%)",
		Type:        types.ParamTypeSelect,
		Default:     "0.3",
		Options: []types.Option{
			{Label: "0.01% (稳定币)", Value: "0.01"},
			{Label: "0.05% (蓝筹)", Value: "0.05"},
			{Label: "0.30% (标准)", Value: "0.3"},
			{Label: "1.00% (高波动)", Value: "1.0"},
		},
	})

	return sim
}

// Init 初始化
func (s *ConcentratedLiquiditySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.tokenA = "ETH"
	s.tokenB = "USDC"
	s.currentPrice = 2000
	s.feeRate = 0.003
	s.tickSpacing = 60 // 0.3%费率对应60tick间距

	if v, ok := config.Params["initial_price"]; ok {
		if f, ok := v.(float64); ok {
			s.currentPrice = f
		}
	}
	if v, ok := config.Params["fee_tier"]; ok {
		if f, ok := v.(string); ok {
			switch f {
			case "0.01":
				s.feeRate = 0.0001
				s.tickSpacing = 1
			case "0.05":
				s.feeRate = 0.0005
				s.tickSpacing = 10
			case "0.3":
				s.feeRate = 0.003
				s.tickSpacing = 60
			case "1.0":
				s.feeRate = 0.01
				s.tickSpacing = 200
			}
		}
	}

	s.currentTick = s.priceToTick(s.currentPrice)
	s.sqrtPriceX96 = s.priceToSqrtPriceX96(s.currentPrice)
	s.liquidity = big.NewInt(0)
	s.positions = make(map[string]*ConcentratedPosition)
	s.ticks = make(map[int64]*Tick)

	s.updateState()
	return nil
}

// =============================================================================
// 核心概念解释
// =============================================================================

// ExplainConcentratedLiquidity 解释集中流动性
func (s *ConcentratedLiquiditySimulator) ExplainConcentratedLiquidity() map[string]interface{} {
	return map[string]interface{}{
		"comparison": map[string]interface{}{
			"uniswap_v2": map[string]string{
				"liquidity":          "分布在0到∞的整个价格范围",
				"capital_efficiency": "低 - 大部分流动性未被使用",
				"position":           "统一的LP代币",
			},
			"uniswap_v3": map[string]string{
				"liquidity":          "集中在指定价格区间",
				"capital_efficiency": "高 - 可达4000x提升",
				"position":           "NFT表示的唯一头寸",
			},
		},
		"key_formulas": map[string]string{
			"liquidity":        "L = sqrt(x * y)",
			"real_reserves":    "x_real = L * (1/sqrt(p) - 1/sqrt(p_upper))",
			"virtual_reserves": "x_virtual = L / sqrt(p)",
		},
		"capital_efficiency": map[string]interface{}{
			"example":       "在2000-2200 USDC/ETH区间提供流动性",
			"v2_equivalent": "需要~100x更多资金达到相同深度",
			"formula":       "效率提升 = 1 / ((1 - pa/pb)^2)",
		},
		"trade_offs": []string{
			"更高的无常损失风险",
			"需要主动管理头寸",
			"价格超出区间后不赚取手续费",
		},
	}
}

// ExplainTickSystem 解释Tick系统
func (s *ConcentratedLiquiditySimulator) ExplainTickSystem() map[string]interface{} {
	return map[string]interface{}{
		"concept": "Tick是离散化的价格点",
		"formula": "price = 1.0001^tick",
		"examples": []map[string]interface{}{
			{"tick": 0, "price": 1.0},
			{"tick": 1000, "price": 1.1052},
			{"tick": 10000, "price": 2.7183},
			{"tick": 69082, "price": 1000.0},
			{"tick": 85176, "price": 5000.0},
		},
		"tick_spacing": map[string]interface{}{
			"purpose":   "限制可用tick，减少gas消耗",
			"0.01%_fee": 1,
			"0.05%_fee": 10,
			"0.30%_fee": 60,
			"1.00%_fee": 200,
		},
		"liquidity_tracking": "每个初始化的tick存储跨越该tick时的流动性变化",
	}
}

// =============================================================================
// 流动性操作
// =============================================================================

// AddLiquidity 添加集中流动性
func (s *ConcentratedLiquiditySimulator) AddLiquidity(owner string, priceLower, priceUpper float64, amountA, amountB *big.Int) (*ConcentratedPosition, error) {
	if priceLower >= priceUpper {
		return nil, fmt.Errorf("价格下限必须小于上限")
	}

	// 转换为tick
	tickLower := s.roundTickDown(s.priceToTick(priceLower))
	tickUpper := s.roundTickUp(s.priceToTick(priceUpper))

	// 计算流动性
	sqrtPriceLower := math.Sqrt(priceLower)
	sqrtPriceUpper := math.Sqrt(priceUpper)
	sqrtPriceCurrent := math.Sqrt(s.currentPrice)

	var liquidity float64
	if s.currentPrice <= priceLower {
		// 当前价格低于区间，只需要tokenA
		liquidity = float64(amountA.Int64()) * sqrtPriceLower * sqrtPriceUpper / (sqrtPriceUpper - sqrtPriceLower)
	} else if s.currentPrice >= priceUpper {
		// 当前价格高于区间，只需要tokenB
		liquidity = float64(amountB.Int64()) / (sqrtPriceUpper - sqrtPriceLower)
	} else {
		// 当前价格在区间内，需要两种代币
		liquidityA := float64(amountA.Int64()) * sqrtPriceCurrent * sqrtPriceUpper / (sqrtPriceUpper - sqrtPriceCurrent)
		liquidityB := float64(amountB.Int64()) / (sqrtPriceCurrent - sqrtPriceLower)
		liquidity = math.Min(liquidityA, liquidityB)
	}

	// 创建头寸
	positionID := fmt.Sprintf("pos-%s-%d", owner, time.Now().UnixNano())
	position := &ConcentratedPosition{
		ID:        positionID,
		Owner:     owner,
		TokenA:    s.tokenA,
		TokenB:    s.tokenB,
		Liquidity: big.NewInt(int64(liquidity)),
		Range: &PriceRange{
			TickLower:  tickLower,
			TickUpper:  tickUpper,
			PriceLower: s.tickToPrice(tickLower),
			PriceUpper: s.tickToPrice(tickUpper),
		},
		DepositedA:  amountA,
		DepositedB:  amountB,
		FeesEarnedA: big.NewInt(0),
		FeesEarnedB: big.NewInt(0),
		InRange:     s.currentTick >= tickLower && s.currentTick < tickUpper,
		CreatedAt:   time.Now(),
	}

	s.positions[positionID] = position

	// 更新tick
	s.updateTick(tickLower, big.NewInt(int64(liquidity)), true)
	s.updateTick(tickUpper, big.NewInt(int64(liquidity)), false)

	// 更新活跃流动性
	if position.InRange {
		s.liquidity.Add(s.liquidity, position.Liquidity)
	}

	s.EmitEvent("liquidity_added", "", "", map[string]interface{}{
		"position_id": positionID,
		"owner":       owner,
		"price_range": fmt.Sprintf("%.2f - %.2f", position.Range.PriceLower, position.Range.PriceUpper),
		"liquidity":   position.Liquidity.String(),
		"in_range":    position.InRange,
	})

	s.updateState()
	return position, nil
}

// RemoveLiquidity 移除流动性
func (s *ConcentratedLiquiditySimulator) RemoveLiquidity(positionID string) (map[string]interface{}, error) {
	position, ok := s.positions[positionID]
	if !ok {
		return nil, fmt.Errorf("头寸不存在")
	}

	// 计算返还的代币
	amountA, amountB := s.calculateTokenAmounts(position)

	// 更新tick
	s.updateTick(position.Range.TickLower, position.Liquidity, false)
	s.updateTick(position.Range.TickUpper, position.Liquidity, true)

	// 更新活跃流动性
	if position.InRange {
		s.liquidity.Sub(s.liquidity, position.Liquidity)
	}

	delete(s.positions, positionID)

	result := map[string]interface{}{
		"position_id":   positionID,
		"amount_a":      amountA.String(),
		"amount_b":      amountB.String(),
		"fees_earned_a": position.FeesEarnedA.String(),
		"fees_earned_b": position.FeesEarnedB.String(),
	}

	s.EmitEvent("liquidity_removed", "", "", result)

	s.updateState()
	return result, nil
}

// calculateTokenAmounts 计算当前头寸的代币数量
func (s *ConcentratedLiquiditySimulator) calculateTokenAmounts(pos *ConcentratedPosition) (*big.Int, *big.Int) {
	sqrtPriceLower := math.Sqrt(pos.Range.PriceLower)
	sqrtPriceUpper := math.Sqrt(pos.Range.PriceUpper)
	sqrtPriceCurrent := math.Sqrt(s.currentPrice)
	L := float64(pos.Liquidity.Int64())

	var amountA, amountB float64

	if s.currentPrice <= pos.Range.PriceLower {
		// 全部是tokenA
		amountA = L * (sqrtPriceUpper - sqrtPriceLower) / (sqrtPriceLower * sqrtPriceUpper)
		amountB = 0
	} else if s.currentPrice >= pos.Range.PriceUpper {
		// 全部是tokenB
		amountA = 0
		amountB = L * (sqrtPriceUpper - sqrtPriceLower)
	} else {
		// 混合
		amountA = L * (sqrtPriceUpper - sqrtPriceCurrent) / (sqrtPriceCurrent * sqrtPriceUpper)
		amountB = L * (sqrtPriceCurrent - sqrtPriceLower)
	}

	return big.NewInt(int64(amountA)), big.NewInt(int64(amountB))
}

// =============================================================================
// 价格和Tick转换
// =============================================================================

// priceToTick 价格转tick
func (s *ConcentratedLiquiditySimulator) priceToTick(price float64) int64 {
	// tick = log(price) / log(1.0001)
	return int64(math.Log(price) / math.Log(1.0001))
}

// tickToPrice tick转价格
func (s *ConcentratedLiquiditySimulator) tickToPrice(tick int64) float64 {
	return math.Pow(1.0001, float64(tick))
}

// roundTickDown 向下取整到tickSpacing
func (s *ConcentratedLiquiditySimulator) roundTickDown(tick int64) int64 {
	if tick < 0 {
		return tick - (tick % s.tickSpacing) - s.tickSpacing
	}
	return tick - (tick % s.tickSpacing)
}

// roundTickUp 向上取整到tickSpacing
func (s *ConcentratedLiquiditySimulator) roundTickUp(tick int64) int64 {
	remainder := tick % s.tickSpacing
	if remainder == 0 {
		return tick
	}
	if tick > 0 {
		return tick + s.tickSpacing - remainder
	}
	return tick - remainder
}

// priceToSqrtPriceX96 价格转sqrtPriceX96
func (s *ConcentratedLiquiditySimulator) priceToSqrtPriceX96(price float64) *big.Int {
	sqrtPrice := math.Sqrt(price)
	q96 := new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)
	sqrtPriceX96 := new(big.Float).Mul(big.NewFloat(sqrtPrice), new(big.Float).SetInt(q96))
	result := new(big.Int)
	sqrtPriceX96.Int(result)
	return result
}

// updateTick 更新tick
func (s *ConcentratedLiquiditySimulator) updateTick(tickIndex int64, liquidityDelta *big.Int, isLower bool) {
	tick, ok := s.ticks[tickIndex]
	if !ok {
		tick = &Tick{
			Index:          tickIndex,
			LiquidityGross: big.NewInt(0),
			LiquidityNet:   big.NewInt(0),
			Initialized:    false,
		}
		s.ticks[tickIndex] = tick
	}

	tick.LiquidityGross.Add(tick.LiquidityGross, liquidityDelta)
	if isLower {
		tick.LiquidityNet.Add(tick.LiquidityNet, liquidityDelta)
	} else {
		tick.LiquidityNet.Sub(tick.LiquidityNet, liquidityDelta)
	}
	tick.Initialized = tick.LiquidityGross.Cmp(big.NewInt(0)) > 0
}

// SimulatePriceMove 模拟价格变动
func (s *ConcentratedLiquiditySimulator) SimulatePriceMove(newPrice float64) map[string]interface{} {
	oldPrice := s.currentPrice
	oldTick := s.currentTick

	s.currentPrice = newPrice
	s.currentTick = s.priceToTick(newPrice)
	s.sqrtPriceX96 = s.priceToSqrtPriceX96(newPrice)

	// 更新头寸的in_range状态
	positionsAffected := 0
	for _, pos := range s.positions {
		wasInRange := pos.InRange
		pos.InRange = s.currentTick >= pos.Range.TickLower && s.currentTick < pos.Range.TickUpper

		if wasInRange != pos.InRange {
			positionsAffected++
			if pos.InRange {
				s.liquidity.Add(s.liquidity, pos.Liquidity)
			} else {
				s.liquidity.Sub(s.liquidity, pos.Liquidity)
			}
		}
	}

	result := map[string]interface{}{
		"old_price":          oldPrice,
		"new_price":          newPrice,
		"old_tick":           oldTick,
		"new_tick":           s.currentTick,
		"positions_affected": positionsAffected,
		"active_liquidity":   s.liquidity.String(),
	}

	s.EmitEvent("price_moved", "", "", result)
	s.updateState()
	return result
}

// GetPositionInfo 获取头寸信息
func (s *ConcentratedLiquiditySimulator) GetPositionInfo(positionID string) map[string]interface{} {
	pos, ok := s.positions[positionID]
	if !ok {
		return nil
	}

	amountA, amountB := s.calculateTokenAmounts(pos)

	return map[string]interface{}{
		"position_id":   pos.ID,
		"owner":         pos.Owner,
		"price_range":   fmt.Sprintf("%.2f - %.2f", pos.Range.PriceLower, pos.Range.PriceUpper),
		"liquidity":     pos.Liquidity.String(),
		"current_a":     amountA.String(),
		"current_b":     amountB.String(),
		"deposited_a":   pos.DepositedA.String(),
		"deposited_b":   pos.DepositedB.String(),
		"in_range":      pos.InRange,
		"fees_earned_a": pos.FeesEarnedA.String(),
		"fees_earned_b": pos.FeesEarnedB.String(),
	}
}

// updateState 更新状态
func (s *ConcentratedLiquiditySimulator) updateState() {
	s.SetGlobalData("current_price", s.currentPrice)
	s.SetGlobalData("current_tick", s.currentTick)
	s.SetGlobalData("active_liquidity", s.liquidity.String())
	s.SetGlobalData("position_count", len(s.positions))

	inRangeCount := 0
	for _, pos := range s.positions {
		if pos.InRange {
			inRangeCount++
		}
	}
	s.SetGlobalData("in_range_positions", inRangeCount)

	summary := "当前集中流动性池处于价格区间观察状态，可添加头寸或推动价格穿越不同区间。"
	nextHint := "先添加一笔价格区间头寸，再推动价格移动，观察头寸是否从活跃状态切换为区间外。"
	progress := 0.0
	if len(s.positions) > 0 {
		summary = fmt.Sprintf("当前共有 %d 个头寸，其中 %d 个头寸仍在有效价格区间内。", len(s.positions), inRangeCount)
		nextHint = "继续移动价格，观察 tick 穿越后活跃流动性和头寸状态的变化。"
		progress = 60
	}

	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"concentrated_liquidity",
		summary,
		nextHint,
		progress,
		map[string]interface{}{
			"current_price":      s.currentPrice,
			"current_tick":       s.currentTick,
			"position_count":     len(s.positions),
			"in_range_positions": inRangeCount,
			"active_liquidity":   s.liquidity.String(),
		},
	)
}

// ExecuteAction 执行集中流动性教学动作
func (s *ConcentratedLiquiditySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "add_liquidity":
		owner := "lp-1"
		if raw, ok := params["owner"].(string); ok && raw != "" {
			owner = raw
		}
		priceLower := s.currentPrice * 0.9
		priceUpper := s.currentPrice * 1.1
		if raw, ok := params["price_lower"].(float64); ok {
			priceLower = raw
		}
		if raw, ok := params["price_upper"].(float64); ok {
			priceUpper = raw
		}
		amountA := big.NewInt(10000)
		amountB := big.NewInt(10000)
		if raw, ok := params["amount_a"].(float64); ok {
			amountA = big.NewInt(int64(raw))
		}
		if raw, ok := params["amount_b"].(float64); ok {
			amountB = big.NewInt(int64(raw))
		}
		position, err := s.AddLiquidity(owner, priceLower, priceUpper, amountA, amountB)
		if err != nil {
			return nil, err
		}
		return defiActionResult(
			"已添加集中流动性头寸",
			map[string]interface{}{
				"position_id": position.ID,
				"tick_lower":  position.Range.TickLower,
				"tick_upper":  position.Range.TickUpper,
			},
			&types.ActionFeedback{
				Summary:     "新的流动性头寸已经进入价格区间，可继续推动价格观察头寸状态变化。",
				NextHint:    "执行一次价格移动，观察头寸是否仍处于有效区间内。",
				EffectScope: "defi",
			},
		), nil
	case "move_price":
		newPrice := s.currentPrice * 1.05
		if raw, ok := params["new_price"].(float64); ok {
			newPrice = raw
		}
		result := s.SimulatePriceMove(newPrice)
		return defiActionResult(
			"已模拟价格移动",
			result,
			&types.ActionFeedback{
				Summary:     "价格已经跨越新的 tick 区间，活跃流动性和头寸状态可能发生变化。",
				NextHint:    "查看哪些头寸仍在区间内，以及活跃流动性是否随之变化。",
				EffectScope: "defi",
			},
		), nil
	case "remove_liquidity":
		positionID := ""
		if raw, ok := params["position_id"].(string); ok {
			positionID = raw
		}
		if positionID == "" {
			for id := range s.positions {
				positionID = id
				break
			}
		}
		if positionID == "" {
			return nil, fmt.Errorf("没有可移除的集中流动性头寸")
		}
		result, err := s.RemoveLiquidity(positionID)
		if err != nil {
			return nil, err
		}
		return defiActionResult(
			"已移除集中流动性头寸",
			result,
			&types.ActionFeedback{
				Summary:     "头寸已经退出池子，相关资产和手续费收益已结算。",
				NextHint:    "继续添加新头寸，比较不同价格区间下的资本效率。",
				EffectScope: "defi",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported concentrated liquidity action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// ConcentratedLiquidityFactory 集中流动性工厂
type ConcentratedLiquidityFactory struct{}

// Create 创建演示器
func (f *ConcentratedLiquidityFactory) Create() engine.Simulator {
	return NewConcentratedLiquiditySimulator()
}

// GetDescription 获取描述
func (f *ConcentratedLiquidityFactory) GetDescription() types.Description {
	return NewConcentratedLiquiditySimulator().GetDescription()
}

// NewConcentratedLiquidityFactory 创建工厂
func NewConcentratedLiquidityFactory() *ConcentratedLiquidityFactory {
	return &ConcentratedLiquidityFactory{}
}

var _ engine.SimulatorFactory = (*ConcentratedLiquidityFactory)(nil)
