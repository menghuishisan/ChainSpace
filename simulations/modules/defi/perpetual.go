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
// 永续合约演示器
// =============================================================================

// PositionSide 持仓方向
type PositionSide string

const (
	PositionLong  PositionSide = "long"
	PositionShort PositionSide = "short"
)

// PerpPosition 永续合约持仓
type PerpPosition struct {
	ID               string       `json:"id"`
	Trader           string       `json:"trader"`
	Side             PositionSide `json:"side"`
	Size             *big.Int     `json:"size"`        // 合约数量
	EntryPrice       float64      `json:"entry_price"` // 开仓均价
	Margin           *big.Int     `json:"margin"`      // 保证金
	Leverage         float64      `json:"leverage"`    // 杠杆倍数
	LiquidationPrice float64      `json:"liquidation_price"`
	UnrealizedPnL    float64      `json:"unrealized_pnl"`
	RealizedPnL      float64      `json:"realized_pnl"`
	FundingPaid      float64      `json:"funding_paid"`
	OpenTime         time.Time    `json:"open_time"`
}

// FundingRate 资金费率
type FundingRate struct {
	Rate       float64   `json:"rate"`
	Timestamp  time.Time `json:"timestamp"`
	MarkPrice  float64   `json:"mark_price"`
	IndexPrice float64   `json:"index_price"`
}

// PerpMarket 永续合约市场
type PerpMarket struct {
	Symbol          string    `json:"symbol"`
	IndexPrice      float64   `json:"index_price"`  // 指数价格(现货)
	MarkPrice       float64   `json:"mark_price"`   // 标记价格
	LastPrice       float64   `json:"last_price"`   // 最新成交价
	FundingRate     float64   `json:"funding_rate"` // 当前资金费率
	NextFundingTime time.Time `json:"next_funding_time"`
	OpenInterest    *big.Int  `json:"open_interest"` // 未平仓合约
	Volume24h       *big.Int  `json:"volume_24h"`
	LongPositions   *big.Int  `json:"long_positions"`
	ShortPositions  *big.Int  `json:"short_positions"`
}

// PerpetualSimulator 永续合约演示器
// 演示永续合约的核心机制:
//
// 1. 永续合约 vs 期货
//   - 无到期日，通过资金费率锚定现货价格
//
// 2. 资金费率
//   - 多头溢价时，多头付给空头
//   - 空头溢价时，空头付给多头
//   - 每8小时结算一次
//
// 3. 杠杆与清算
//   - 支持高杠杆(1-100x)
//   - 保证金不足时强制平仓
//
// 4. 标记价格
//   - 用于计算盈亏和清算价格
//   - 防止市场操纵导致错误清算
type PerpetualSimulator struct {
	*base.BaseSimulator
	market            *PerpMarket
	positions         map[string]*PerpPosition
	fundingHistory    []*FundingRate
	maxLeverage       float64
	maintenanceMargin float64 // 维持保证金率
	takerFee          float64
	makerFee          float64
}

// NewPerpetualSimulator 创建永续合约演示器
func NewPerpetualSimulator() *PerpetualSimulator {
	sim := &PerpetualSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"perpetual",
			"永续合约演示器",
			"演示永续合约的资金费率、杠杆、清算等核心机制",
			"defi",
			types.ComponentDeFi,
		),
		positions:      make(map[string]*PerpPosition),
		fundingHistory: make([]*FundingRate, 0),
	}

	sim.AddParam(types.Param{
		Key:         "max_leverage",
		Name:        "最大杠杆",
		Description: "允许的最大杠杆倍数",
		Type:        types.ParamTypeInt,
		Default:     100,
		Min:         1,
		Max:         125,
	})

	sim.AddParam(types.Param{
		Key:         "initial_price",
		Name:        "初始价格",
		Description: "BTC初始价格(USD)",
		Type:        types.ParamTypeFloat,
		Default:     40000,
		Min:         1000,
		Max:         100000,
	})

	return sim
}

// Init 初始化
func (s *PerpetualSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.maxLeverage = 100
	s.maintenanceMargin = 0.005 // 0.5%
	s.takerFee = 0.0006         // 0.06%
	s.makerFee = 0.0002         // 0.02%

	initialPrice := 40000.0
	if v, ok := config.Params["max_leverage"]; ok {
		if f, ok := v.(float64); ok {
			s.maxLeverage = f
		}
	}
	if v, ok := config.Params["initial_price"]; ok {
		if f, ok := v.(float64); ok {
			initialPrice = f
		}
	}

	s.market = &PerpMarket{
		Symbol:          "BTC-PERP",
		IndexPrice:      initialPrice,
		MarkPrice:       initialPrice,
		LastPrice:       initialPrice,
		FundingRate:     0.0001, // 0.01%
		NextFundingTime: time.Now().Add(8 * time.Hour),
		OpenInterest:    big.NewInt(0),
		Volume24h:       big.NewInt(0),
		LongPositions:   big.NewInt(0),
		ShortPositions:  big.NewInt(0),
	}

	s.positions = make(map[string]*PerpPosition)
	s.fundingHistory = make([]*FundingRate, 0)

	s.updateState()
	return nil
}

// =============================================================================
// 资金费率
// =============================================================================

// ExplainFundingRate 解释资金费率
func (s *PerpetualSimulator) ExplainFundingRate() map[string]interface{} {
	return map[string]interface{}{
		"purpose": "使永续合约价格锚定现货指数价格",
		"mechanism": []string{
			"每8小时结算一次资金费",
			"当永续价格 > 指数价格: 正费率，多头付给空头",
			"当永续价格 < 指数价格: 负费率，空头付给多头",
			"费用 = 持仓价值 × 资金费率",
		},
		"formula": map[string]string{
			"funding_rate": "FundingRate = Clamp((MarkPrice - IndexPrice) / IndexPrice / 24, -0.75%, 0.75%)",
			"payment":      "FundingPayment = PositionSize × MarkPrice × FundingRate",
		},
		"example": map[string]interface{}{
			"position_size": "1 BTC",
			"mark_price":    40000,
			"funding_rate":  "0.01%",
			"payment":       "1 × 40000 × 0.0001 = $4 (多头支付)",
		},
		"impact": []string{
			"正费率激励做空，抑制做多",
			"负费率激励做多，抑制做空",
			"套利者会在现货和永续之间套利",
		},
	}
}

// CalculateFundingRate 计算资金费率
func (s *PerpetualSimulator) CalculateFundingRate() *FundingRate {
	// 资金费率 = (标记价格 - 指数价格) / 指数价格 / 24
	premium := (s.market.MarkPrice - s.market.IndexPrice) / s.market.IndexPrice
	rate := premium / 24

	// 限制在 ±0.75%
	if rate > 0.0075 {
		rate = 0.0075
	} else if rate < -0.0075 {
		rate = -0.0075
	}

	fundingRate := &FundingRate{
		Rate:       rate,
		Timestamp:  time.Now(),
		MarkPrice:  s.market.MarkPrice,
		IndexPrice: s.market.IndexPrice,
	}

	s.market.FundingRate = rate
	s.fundingHistory = append(s.fundingHistory, fundingRate)

	return fundingRate
}

// SettleFunding 结算资金费
func (s *PerpetualSimulator) SettleFunding() map[string]interface{} {
	totalLongPayment := 0.0
	totalShortPayment := 0.0

	for _, pos := range s.positions {
		positionValue := float64(pos.Size.Int64()) * s.market.MarkPrice
		payment := positionValue * s.market.FundingRate

		if pos.Side == PositionLong {
			pos.FundingPaid += payment
			totalLongPayment += payment
		} else {
			pos.FundingPaid -= payment // 空头收到
			totalShortPayment -= payment
		}
	}

	result := map[string]interface{}{
		"funding_rate":         fmt.Sprintf("%.4f%%", s.market.FundingRate*100),
		"long_total_paid":      totalLongPayment,
		"short_total_received": -totalShortPayment,
		"timestamp":            time.Now(),
	}

	s.EmitEvent("funding_settled", "", "", result)

	// 设置下次结算时间
	s.market.NextFundingTime = time.Now().Add(8 * time.Hour)

	s.updateState()
	return result
}

// =============================================================================
// 交易操作
// =============================================================================

// OpenPosition 开仓
func (s *PerpetualSimulator) OpenPosition(trader string, side PositionSide, size *big.Int, leverage float64) (*PerpPosition, error) {
	if leverage > s.maxLeverage {
		return nil, fmt.Errorf("杠杆倍数超过最大限制 %.0fx", s.maxLeverage)
	}
	if leverage < 1 {
		leverage = 1
	}

	// 计算所需保证金
	positionValue := float64(size.Int64()) * s.market.MarkPrice
	requiredMargin := positionValue / leverage

	// 计算清算价格
	var liquidationPrice float64
	if side == PositionLong {
		// 多头: 清算价 = 开仓价 × (1 - 1/杠杆 + 维持保证金率)
		liquidationPrice = s.market.MarkPrice * (1 - 1/leverage + s.maintenanceMargin)
	} else {
		// 空头: 清算价 = 开仓价 × (1 + 1/杠杆 - 维持保证金率)
		liquidationPrice = s.market.MarkPrice * (1 + 1/leverage - s.maintenanceMargin)
	}

	positionID := fmt.Sprintf("pos-%s-%d", trader, time.Now().UnixNano())
	position := &PerpPosition{
		ID:               positionID,
		Trader:           trader,
		Side:             side,
		Size:             size,
		EntryPrice:       s.market.MarkPrice,
		Margin:           big.NewInt(int64(requiredMargin)),
		Leverage:         leverage,
		LiquidationPrice: liquidationPrice,
		UnrealizedPnL:    0,
		RealizedPnL:      0,
		FundingPaid:      0,
		OpenTime:         time.Now(),
	}

	s.positions[positionID] = position

	// 更新市场数据
	s.market.OpenInterest.Add(s.market.OpenInterest, size)
	if side == PositionLong {
		s.market.LongPositions.Add(s.market.LongPositions, size)
	} else {
		s.market.ShortPositions.Add(s.market.ShortPositions, size)
	}

	s.EmitEvent("position_opened", "", "", map[string]interface{}{
		"position_id":       positionID,
		"trader":            trader,
		"side":              side,
		"size":              size.String(),
		"entry_price":       s.market.MarkPrice,
		"leverage":          leverage,
		"margin":            requiredMargin,
		"liquidation_price": liquidationPrice,
	})

	s.updateState()
	return position, nil
}

// ClosePosition 平仓
func (s *PerpetualSimulator) ClosePosition(positionID string) (map[string]interface{}, error) {
	position, ok := s.positions[positionID]
	if !ok {
		return nil, fmt.Errorf("持仓不存在")
	}

	// 计算已实现盈亏
	exitPrice := s.market.MarkPrice
	var pnl float64
	if position.Side == PositionLong {
		pnl = float64(position.Size.Int64()) * (exitPrice - position.EntryPrice)
	} else {
		pnl = float64(position.Size.Int64()) * (position.EntryPrice - exitPrice)
	}

	pnl -= position.FundingPaid // 扣除资金费

	// 更新市场数据
	s.market.OpenInterest.Sub(s.market.OpenInterest, position.Size)
	if position.Side == PositionLong {
		s.market.LongPositions.Sub(s.market.LongPositions, position.Size)
	} else {
		s.market.ShortPositions.Sub(s.market.ShortPositions, position.Size)
	}

	delete(s.positions, positionID)

	result := map[string]interface{}{
		"position_id":  positionID,
		"exit_price":   exitPrice,
		"entry_price":  position.EntryPrice,
		"pnl":          pnl,
		"funding_paid": position.FundingPaid,
		"roi":          pnl / float64(position.Margin.Int64()) * 100,
	}

	s.EmitEvent("position_closed", "", "", result)

	s.updateState()
	return result, nil
}

// UpdatePrice 更新价格
func (s *PerpetualSimulator) UpdatePrice(newMarkPrice, newIndexPrice float64) {
	s.market.MarkPrice = newMarkPrice
	s.market.IndexPrice = newIndexPrice
	s.market.LastPrice = newMarkPrice

	// 更新所有持仓的未实现盈亏
	for _, pos := range s.positions {
		s.updatePositionPnL(pos)
	}

	// 检查清算
	s.checkLiquidations()

	s.CalculateFundingRate()
	s.updateState()
}

// updatePositionPnL 更新持仓盈亏
func (s *PerpetualSimulator) updatePositionPnL(pos *PerpPosition) {
	if pos.Side == PositionLong {
		pos.UnrealizedPnL = float64(pos.Size.Int64()) * (s.market.MarkPrice - pos.EntryPrice)
	} else {
		pos.UnrealizedPnL = float64(pos.Size.Int64()) * (pos.EntryPrice - s.market.MarkPrice)
	}
}

// checkLiquidations 检查清算
func (s *PerpetualSimulator) checkLiquidations() {
	toLiquidate := make([]string, 0)

	for id, pos := range s.positions {
		if pos.Side == PositionLong && s.market.MarkPrice <= pos.LiquidationPrice {
			toLiquidate = append(toLiquidate, id)
		} else if pos.Side == PositionShort && s.market.MarkPrice >= pos.LiquidationPrice {
			toLiquidate = append(toLiquidate, id)
		}
	}

	for _, id := range toLiquidate {
		s.liquidatePosition(id)
	}
}

// liquidatePosition 强制平仓
func (s *PerpetualSimulator) liquidatePosition(positionID string) {
	pos := s.positions[positionID]
	if pos == nil {
		return
	}

	s.EmitEvent("position_liquidated", "", "", map[string]interface{}{
		"position_id":       positionID,
		"trader":            pos.Trader,
		"side":              pos.Side,
		"size":              pos.Size.String(),
		"entry_price":       pos.EntryPrice,
		"liquidation_price": pos.LiquidationPrice,
		"mark_price":        s.market.MarkPrice,
		"margin_lost":       pos.Margin.String(),
	})

	// 更新市场数据
	s.market.OpenInterest.Sub(s.market.OpenInterest, pos.Size)
	if pos.Side == PositionLong {
		s.market.LongPositions.Sub(s.market.LongPositions, pos.Size)
	} else {
		s.market.ShortPositions.Sub(s.market.ShortPositions, pos.Size)
	}

	delete(s.positions, positionID)
}

// GetMarketInfo 获取市场信息
func (s *PerpetualSimulator) GetMarketInfo() map[string]interface{} {
	longShortRatio := float64(0)
	if s.market.ShortPositions.Cmp(big.NewInt(0)) > 0 {
		longShortRatio = float64(s.market.LongPositions.Int64()) / float64(s.market.ShortPositions.Int64())
	}

	return map[string]interface{}{
		"symbol":           s.market.Symbol,
		"index_price":      s.market.IndexPrice,
		"mark_price":       s.market.MarkPrice,
		"last_price":       s.market.LastPrice,
		"funding_rate":     fmt.Sprintf("%.4f%%", s.market.FundingRate*100),
		"next_funding":     s.market.NextFundingTime,
		"open_interest":    s.market.OpenInterest.String(),
		"long_positions":   s.market.LongPositions.String(),
		"short_positions":  s.market.ShortPositions.String(),
		"long_short_ratio": longShortRatio,
		"position_count":   len(s.positions),
	}
}

// updateState 更新状态
func (s *PerpetualSimulator) updateState() {
	s.SetGlobalData("mark_price", s.market.MarkPrice)
	s.SetGlobalData("index_price", s.market.IndexPrice)
	s.SetGlobalData("funding_rate", s.market.FundingRate)
	s.SetGlobalData("open_interest", s.market.OpenInterest.String())
	s.SetGlobalData("position_count", len(s.positions))

	summary := fmt.Sprintf("当前标记价格为 %.2f，未平仓总量为 %s。", s.market.MarkPrice, s.market.OpenInterest.String())
	nextHint := "可以开仓、调价或观察资金费率变化，理解仓位盈亏和清算路径。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"perpetual_positioning",
		summary,
		nextHint,
		0.5,
		map[string]interface{}{"mark_price": s.market.MarkPrice, "open_interest": s.market.OpenInterest.String(), "position_count": len(s.positions)},
	)
}

func (s *PerpetualSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "open_position":
		trader := "alice"
		side := PositionLong
		size := big.NewInt(10)
		leverage := 5.0
		if raw, ok := params["trader"].(string); ok && raw != "" {
			trader = raw
		}
		if raw, ok := params["side"].(string); ok && raw == "short" {
			side = PositionShort
		}
		if raw, ok := params["size"].(float64); ok && raw > 0 {
			size = big.NewInt(int64(raw))
		}
		if raw, ok := params["leverage"].(float64); ok && raw > 0 {
			leverage = raw
		}
		position, err := s.OpenPosition(trader, side, size, leverage)
		if err != nil {
			return nil, err
		}
		return defiActionResult("已开立一笔永续合约仓位。", map[string]interface{}{"position_id": position.ID}, &types.ActionFeedback{
			Summary:     "新的多头或空头仓位已经建立，开仓价格和清算价格已确定。",
			NextHint:    "继续更新标记价格，观察未实现盈亏和清算风险如何变化。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"position_id": position.ID, "position_count": len(s.positions)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported perpetual action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// PerpetualFactory 永续合约工厂
type PerpetualFactory struct{}

// Create 创建演示器
func (f *PerpetualFactory) Create() engine.Simulator {
	return NewPerpetualSimulator()
}

// GetDescription 获取描述
func (f *PerpetualFactory) GetDescription() types.Description {
	return NewPerpetualSimulator().GetDescription()
}

// NewPerpetualFactory 创建工厂
func NewPerpetualFactory() *PerpetualFactory {
	return &PerpetualFactory{}
}

var _ engine.SimulatorFactory = (*PerpetualFactory)(nil)
