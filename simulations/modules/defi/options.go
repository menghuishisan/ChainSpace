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
// 期权演示器
// =============================================================================

// OptionType 期权类型
type OptionType string

const (
	OptionCall OptionType = "call" // 看涨期权
	OptionPut  OptionType = "put"  // 看跌期权
)

// OptionStyle 期权风格
type OptionStyle string

const (
	OptionEuropean OptionStyle = "european" // 欧式期权
	OptionAmerican OptionStyle = "american" // 美式期权
)

// Option 期权合约
type Option struct {
	ID          string      `json:"id"`
	Underlying  string      `json:"underlying"`   // 标的资产
	OptionType  OptionType  `json:"option_type"`  // call/put
	Style       OptionStyle `json:"style"`        // 欧式/美式
	StrikePrice float64     `json:"strike_price"` // 行权价
	Premium     float64     `json:"premium"`      // 期权费
	Expiry      time.Time   `json:"expiry"`       // 到期时间
	Size        *big.Int    `json:"size"`         // 合约规模
	Writer      string      `json:"writer"`       // 卖方
	Holder      string      `json:"holder"`       // 买方
	Collateral  *big.Int    `json:"collateral"`   // 抵押品
	IsExercised bool        `json:"is_exercised"`
	IsExpired   bool        `json:"is_expired"`
}

// GreeksResult Greeks计算结果
type GreeksResult struct {
	Delta float64 `json:"delta"` // 标的价格敏感度
	Gamma float64 `json:"gamma"` // Delta变化率
	Theta float64 `json:"theta"` // 时间价值衰减
	Vega  float64 `json:"vega"`  // 波动率敏感度
	Rho   float64 `json:"rho"`   // 利率敏感度
}

// OptionsSimulator 期权演示器
// 演示DeFi期权的核心机制:
//
// 1. 期权基础
//   - 看涨期权(Call): 权利以行权价买入
//   - 看跌期权(Put): 权利以行权价卖出
//
// 2. 期权定价
//   - Black-Scholes模型
//   - 希腊字母(Greeks)
//
// 3. DeFi期权协议
//   - 抵押品机制
//   - 自动结算
//
// 参考: Opyn, Dopex, Lyra
type OptionsSimulator struct {
	*base.BaseSimulator
	underlying   string
	spotPrice    float64
	volatility   float64 // 隐含波动率
	riskFreeRate float64 // 无风险利率
	options      map[string]*Option
}

// NewOptionsSimulator 创建期权演示器
func NewOptionsSimulator() *OptionsSimulator {
	sim := &OptionsSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"options",
			"期权演示器",
			"演示期权定价(Black-Scholes)、Greeks、行权等DeFi期权机制",
			"defi",
			types.ComponentDeFi,
		),
		options: make(map[string]*Option),
	}

	sim.AddParam(types.Param{
		Key:         "spot_price",
		Name:        "现货价格",
		Description: "标的资产当前价格",
		Type:        types.ParamTypeFloat,
		Default:     2000,
		Min:         100,
		Max:         100000,
	})

	sim.AddParam(types.Param{
		Key:         "volatility",
		Name:        "波动率",
		Description: "年化隐含波动率(%)",
		Type:        types.ParamTypeFloat,
		Default:     80,
		Min:         10,
		Max:         200,
	})

	return sim
}

// Init 初始化
func (s *OptionsSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.underlying = "ETH"
	s.spotPrice = 2000
	s.volatility = 0.80
	s.riskFreeRate = 0.05

	if v, ok := config.Params["spot_price"]; ok {
		if f, ok := v.(float64); ok {
			s.spotPrice = f
		}
	}
	if v, ok := config.Params["volatility"]; ok {
		if f, ok := v.(float64); ok {
			s.volatility = f / 100
		}
	}

	s.options = make(map[string]*Option)

	s.updateState()
	return nil
}

// =============================================================================
// 期权基础解释
// =============================================================================

// ExplainOptions 解释期权
func (s *OptionsSimulator) ExplainOptions() map[string]interface{} {
	return map[string]interface{}{
		"definition": "期权是一种衍生品合约，给予持有者在特定时间以特定价格买入/卖出标的资产的权利(而非义务)",
		"types": []map[string]interface{}{
			{
				"type":        "Call (看涨期权)",
				"right":       "以行权价买入标的资产",
				"when_profit": "标的价格 > 行权价 + 期权费",
				"max_loss":    "期权费",
				"max_profit":  "无限",
			},
			{
				"type":        "Put (看跌期权)",
				"right":       "以行权价卖出标的资产",
				"when_profit": "标的价格 < 行权价 - 期权费",
				"max_loss":    "期权费",
				"max_profit":  "行权价 - 期权费",
			},
		},
		"styles": []map[string]string{
			{"style": "欧式", "exercise": "只能在到期日行权"},
			{"style": "美式", "exercise": "到期前任何时间可行权"},
		},
		"terminology": map[string]string{
			"premium":      "期权费 - 购买期权的成本",
			"strike_price": "行权价 - 约定的买入/卖出价格",
			"expiry":       "到期时间",
			"in_the_money": "价内 - 立即行权有利润",
			"at_the_money": "平价 - 现货价=行权价",
			"out_of_money": "价外 - 立即行权无利润",
		},
	}
}

// =============================================================================
// Black-Scholes定价
// =============================================================================

// BlackScholesPrice 计算Black-Scholes期权价格
func (s *OptionsSimulator) BlackScholesPrice(optionType OptionType, strike float64, daysToExpiry int) map[string]interface{} {
	S := s.spotPrice
	K := strike
	T := float64(daysToExpiry) / 365
	r := s.riskFreeRate
	sigma := s.volatility

	if T <= 0 {
		T = 0.0001 // 防止除零
	}

	d1 := (math.Log(S/K) + (r+sigma*sigma/2)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	var price float64
	var intrinsicValue float64

	if optionType == OptionCall {
		price = S*normCDF(d1) - K*math.Exp(-r*T)*normCDF(d2)
		intrinsicValue = math.Max(S-K, 0)
	} else {
		price = K*math.Exp(-r*T)*normCDF(-d2) - S*normCDF(-d1)
		intrinsicValue = math.Max(K-S, 0)
	}

	timeValue := price - intrinsicValue

	return map[string]interface{}{
		"option_type":     optionType,
		"spot_price":      S,
		"strike_price":    K,
		"days_to_expiry":  daysToExpiry,
		"volatility":      fmt.Sprintf("%.0f%%", sigma*100),
		"risk_free_rate":  fmt.Sprintf("%.1f%%", r*100),
		"premium":         fmt.Sprintf("%.2f", price),
		"intrinsic_value": fmt.Sprintf("%.2f", intrinsicValue),
		"time_value":      fmt.Sprintf("%.2f", timeValue),
		"d1":              d1,
		"d2":              d2,
		"moneyness": func() string {
			if optionType == OptionCall {
				if S > K*1.05 {
					return "ITM (价内)"
				} else if S < K*0.95 {
					return "OTM (价外)"
				}
				return "ATM (平价)"
			}
			if S < K*0.95 {
				return "ITM (价内)"
			} else if S > K*1.05 {
				return "OTM (价外)"
			}
			return "ATM (平价)"
		}(),
	}
}

// normCDF 标准正态分布累积分布函数
func normCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// normPDF 标准正态分布概率密度函数
func normPDF(x float64) float64 {
	return math.Exp(-x*x/2) / math.Sqrt(2*math.Pi)
}

// CalculateGreeks 计算希腊字母
func (s *OptionsSimulator) CalculateGreeks(optionType OptionType, strike float64, daysToExpiry int) *GreeksResult {
	S := s.spotPrice
	K := strike
	T := float64(daysToExpiry) / 365
	r := s.riskFreeRate
	sigma := s.volatility

	if T <= 0 {
		T = 0.0001
	}

	sqrtT := math.Sqrt(T)
	d1 := (math.Log(S/K) + (r+sigma*sigma/2)*T) / (sigma * sqrtT)
	d2 := d1 - sigma*sqrtT

	var delta, theta float64

	if optionType == OptionCall {
		delta = normCDF(d1)
		theta = -S*normPDF(d1)*sigma/(2*sqrtT) - r*K*math.Exp(-r*T)*normCDF(d2)
	} else {
		delta = normCDF(d1) - 1
		theta = -S*normPDF(d1)*sigma/(2*sqrtT) + r*K*math.Exp(-r*T)*normCDF(-d2)
	}

	gamma := normPDF(d1) / (S * sigma * sqrtT)
	vega := S * normPDF(d1) * sqrtT / 100             // 每1%波动率变化
	rho := K * T * math.Exp(-r*T) * normCDF(d2) / 100 // 每1%利率变化

	if optionType == OptionPut {
		rho = -K * T * math.Exp(-r*T) * normCDF(-d2) / 100
	}

	return &GreeksResult{
		Delta: delta,
		Gamma: gamma,
		Theta: theta / 365, // 每日theta
		Vega:  vega,
		Rho:   rho,
	}
}

// ExplainGreeks 解释希腊字母
func (s *OptionsSimulator) ExplainGreeks() map[string]interface{} {
	return map[string]interface{}{
		"overview": "希腊字母衡量期权价格对不同因素的敏感度",
		"greeks": []map[string]interface{}{
			{
				"name":     "Delta (Δ)",
				"measures": "标的价格变化$1时，期权价格变化",
				"range":    "Call: 0到1, Put: -1到0",
				"usage":    "对冲比例、方向性风险",
				"example":  "Delta=0.5: 标的涨$1，期权涨$0.50",
			},
			{
				"name":       "Gamma (Γ)",
				"measures":   "标的价格变化$1时，Delta的变化",
				"highest_at": "ATM期权，临近到期",
				"usage":      "Delta对冲的稳定性",
			},
			{
				"name":        "Theta (Θ)",
				"measures":    "每天时间流逝导致的期权价值损失",
				"sign":        "买方为负(时间损耗)，卖方为正(收取时间价值)",
				"accelerates": "临近到期时加速衰减",
			},
			{
				"name":       "Vega (ν)",
				"measures":   "波动率变化1%时，期权价格变化",
				"highest_at": "ATM期权，长期期权",
				"usage":      "波动率交易",
			},
			{
				"name":       "Rho (ρ)",
				"measures":   "利率变化1%时，期权价格变化",
				"importance": "对DeFi期权影响较小",
			},
		},
	}
}

// =============================================================================
// 期权交易
// =============================================================================

// WriteOption 卖出(开立)期权
func (s *OptionsSimulator) WriteOption(writer string, optionType OptionType, strike float64, daysToExpiry int, size *big.Int) (*Option, error) {
	// 计算期权费
	pricing := s.BlackScholesPrice(optionType, strike, daysToExpiry)
	premium := 0.0
	fmt.Sscanf(pricing["premium"].(string), "%f", &premium)

	// 计算所需抵押品
	var collateral *big.Int
	if optionType == OptionCall {
		// Call期权需要抵押标的资产
		collateral = new(big.Int).Set(size)
	} else {
		// Put期权需要抵押行权价值的稳定币
		collateral = big.NewInt(int64(strike * float64(size.Int64())))
	}

	optionID := fmt.Sprintf("option-%s-%d", writer, time.Now().UnixNano())
	option := &Option{
		ID:          optionID,
		Underlying:  s.underlying,
		OptionType:  optionType,
		Style:       OptionEuropean,
		StrikePrice: strike,
		Premium:     premium,
		Expiry:      time.Now().Add(time.Duration(daysToExpiry) * 24 * time.Hour),
		Size:        size,
		Writer:      writer,
		Holder:      "",
		Collateral:  collateral,
		IsExercised: false,
		IsExpired:   false,
	}

	s.options[optionID] = option

	s.EmitEvent("option_written", "", "", map[string]interface{}{
		"option_id":   optionID,
		"writer":      writer,
		"type":        optionType,
		"strike":      strike,
		"expiry_days": daysToExpiry,
		"premium":     premium,
		"collateral":  collateral.String(),
	})

	s.updateState()
	return option, nil
}

// BuyOption 买入期权
func (s *OptionsSimulator) BuyOption(optionID, buyer string) error {
	option, ok := s.options[optionID]
	if !ok {
		return fmt.Errorf("期权不存在")
	}

	if option.Holder != "" {
		return fmt.Errorf("期权已被购买")
	}

	option.Holder = buyer

	s.EmitEvent("option_bought", "", "", map[string]interface{}{
		"option_id":    optionID,
		"buyer":        buyer,
		"premium_paid": option.Premium * float64(option.Size.Int64()),
	})

	s.updateState()
	return nil
}

// ExerciseOption 行权
func (s *OptionsSimulator) ExerciseOption(optionID string) (map[string]interface{}, error) {
	option, ok := s.options[optionID]
	if !ok {
		return nil, fmt.Errorf("期权不存在")
	}

	if option.IsExercised {
		return nil, fmt.Errorf("期权已行权")
	}

	if time.Now().After(option.Expiry) {
		option.IsExpired = true
		return nil, fmt.Errorf("期权已过期")
	}

	// 计算行权收益
	var profit float64
	if option.OptionType == OptionCall {
		profit = (s.spotPrice - option.StrikePrice) * float64(option.Size.Int64())
	} else {
		profit = (option.StrikePrice - s.spotPrice) * float64(option.Size.Int64())
	}

	if profit <= 0 {
		return nil, fmt.Errorf("期权为价外，行权无利润")
	}

	option.IsExercised = true

	result := map[string]interface{}{
		"option_id":    optionID,
		"option_type":  option.OptionType,
		"strike_price": option.StrikePrice,
		"spot_price":   s.spotPrice,
		"size":         option.Size.String(),
		"gross_profit": profit,
		"net_profit":   profit - option.Premium*float64(option.Size.Int64()),
	}

	s.EmitEvent("option_exercised", "", "", result)

	s.updateState()
	return result, nil
}

// UpdateSpotPrice 更新现货价格
func (s *OptionsSimulator) UpdateSpotPrice(newPrice float64) {
	oldPrice := s.spotPrice
	s.spotPrice = newPrice

	s.EmitEvent("price_updated", "", "", map[string]interface{}{
		"old_price": oldPrice,
		"new_price": newPrice,
		"change":    fmt.Sprintf("%.2f%%", (newPrice-oldPrice)/oldPrice*100),
	})

	s.updateState()
}

// GetOptionChain 获取期权链
func (s *OptionsSimulator) GetOptionChain(daysToExpiry int) []map[string]interface{} {
	strikes := []float64{
		s.spotPrice * 0.8,
		s.spotPrice * 0.9,
		s.spotPrice * 0.95,
		s.spotPrice,
		s.spotPrice * 1.05,
		s.spotPrice * 1.1,
		s.spotPrice * 1.2,
	}

	chain := make([]map[string]interface{}, 0)
	for _, strike := range strikes {
		callPricing := s.BlackScholesPrice(OptionCall, strike, daysToExpiry)
		putPricing := s.BlackScholesPrice(OptionPut, strike, daysToExpiry)
		callGreeks := s.CalculateGreeks(OptionCall, strike, daysToExpiry)
		putGreeks := s.CalculateGreeks(OptionPut, strike, daysToExpiry)

		chain = append(chain, map[string]interface{}{
			"strike":       strike,
			"call_premium": callPricing["premium"],
			"call_delta":   fmt.Sprintf("%.2f", callGreeks.Delta),
			"put_premium":  putPricing["premium"],
			"put_delta":    fmt.Sprintf("%.2f", putGreeks.Delta),
			"moneyness":    callPricing["moneyness"],
		})
	}

	return chain
}

// updateState 更新状态
func (s *OptionsSimulator) updateState() {
	s.SetGlobalData("underlying", s.underlying)
	s.SetGlobalData("spot_price", s.spotPrice)
	s.SetGlobalData("volatility", s.volatility)
	s.SetGlobalData("option_count", len(s.options))

	summary := fmt.Sprintf("当前标的现价为 %.2f，共有 %d 张期权合约。", s.spotPrice, len(s.options))
	nextHint := "可以继续写入期权、买入期权或更新现货价格，观察内在价值与盈亏变化。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"options_pricing",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"spot_price": s.spotPrice, "option_count": len(s.options)},
	)
}

func (s *OptionsSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "update_spot_price":
		newPrice := s.spotPrice * 1.05
		if raw, ok := params["new_price"].(float64); ok && raw > 0 {
			newPrice = raw
		}
		s.UpdateSpotPrice(newPrice)
		return defiActionResult("已更新期权标的现货价格。", map[string]interface{}{"spot_price": s.spotPrice}, &types.ActionFeedback{
			Summary:     "现货价格变化已同步影响期权内在价值和行权收益。",
			NextHint:    "继续比较不同执行价的期权链，观察虚值、平值、实值状态如何变化。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"spot_price": s.spotPrice},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported options action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// OptionsFactory 期权工厂
type OptionsFactory struct{}

// Create 创建演示器
func (f *OptionsFactory) Create() engine.Simulator {
	return NewOptionsSimulator()
}

// GetDescription 获取描述
func (f *OptionsFactory) GetDescription() types.Description {
	return NewOptionsSimulator().GetDescription()
}

// NewOptionsFactory 创建工厂
func NewOptionsFactory() *OptionsFactory {
	return &OptionsFactory{}
}

var _ engine.SimulatorFactory = (*OptionsFactory)(nil)
