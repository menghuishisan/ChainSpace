package defi

import (
	"fmt"
	"math"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 利率模型演示器
// =============================================================================

// InterestModelType 利率模型类型
type InterestModelType string

const (
	ModelLinear   InterestModelType = "linear"    // 线性模型
	ModelJumpRate InterestModelType = "jump_rate" // 跳跃利率模型
	ModelKink     InterestModelType = "kink"      // 拐点模型
	ModelDynamic  InterestModelType = "dynamic"   // 动态模型
)

// InterestRatePoint 利率曲线点
type InterestRatePoint struct {
	Utilization float64 `json:"utilization_percent"`
	BorrowRate  float64 `json:"borrow_rate_percent"`
	SupplyRate  float64 `json:"supply_rate_percent"`
}

// ModelParameters 模型参数
type ModelParameters struct {
	BaseRate           float64 `json:"base_rate"`           // 基础利率
	Multiplier         float64 `json:"multiplier"`          // 斜率乘数
	JumpMultiplier     float64 `json:"jump_multiplier"`     // 跳跃乘数
	OptimalUtilization float64 `json:"optimal_utilization"` // 最优利用率
	ReserveFactor      float64 `json:"reserve_factor"`      // 储备金率
}

// InterestModelSimulator 利率模型演示器
// 演示DeFi借贷协议的利率模型:
//
// 1. 线性模型
//   - borrowRate = baseRate + utilization × multiplier
//   - 简单但不够灵活
//
// 2. 跳跃利率模型 (Compound/Aave)
//   - 在最优利用率之下缓慢增长
//   - 超过最优利用率后急剧上升
//   - 激励存款，抑制过度借款
//
// 3. 动态利率模型
//   - 根据市场条件自动调整参数
//
// 核心公式:
// supplyRate = borrowRate × utilization × (1 - reserveFactor)
type InterestModelSimulator struct {
	*base.BaseSimulator
	modelType   InterestModelType
	params      *ModelParameters
	rateHistory []*InterestRatePoint
}

// NewInterestModelSimulator 创建利率模型演示器
func NewInterestModelSimulator() *InterestModelSimulator {
	sim := &InterestModelSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"interest_model",
			"利率模型演示器",
			"演示线性、跳跃利率、动态利率等DeFi借贷利率模型",
			"defi",
			types.ComponentDeFi,
		),
		rateHistory: make([]*InterestRatePoint, 0),
	}

	sim.AddParam(types.Param{
		Key:         "model_type",
		Name:        "模型类型",
		Description: "利率计算模型",
		Type:        types.ParamTypeSelect,
		Default:     "jump_rate",
		Options: []types.Option{
			{Label: "线性模型", Value: "linear"},
			{Label: "跳跃利率模型", Value: "jump_rate"},
			{Label: "动态模型", Value: "dynamic"},
		},
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

	sim.AddParam(types.Param{
		Key:         "base_rate",
		Name:        "基础利率",
		Description: "利用率为0时的年化利率(%)",
		Type:        types.ParamTypeFloat,
		Default:     2.0,
		Min:         0,
		Max:         10,
	})

	return sim
}

// Init 初始化
func (s *InterestModelSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.modelType = ModelJumpRate
	s.params = &ModelParameters{
		BaseRate:           2.0,
		Multiplier:         4.0,
		JumpMultiplier:     75.0,
		OptimalUtilization: 80.0,
		ReserveFactor:      0.10,
	}

	if v, ok := config.Params["model_type"]; ok {
		if t, ok := v.(string); ok {
			s.modelType = InterestModelType(t)
		}
	}
	if v, ok := config.Params["optimal_utilization"]; ok {
		if f, ok := v.(float64); ok {
			s.params.OptimalUtilization = f
		}
	}
	if v, ok := config.Params["base_rate"]; ok {
		if f, ok := v.(float64); ok {
			s.params.BaseRate = f
		}
	}

	s.rateHistory = make([]*InterestRatePoint, 0)
	s.updateState()
	return nil
}

// =============================================================================
// 利率模型解释
// =============================================================================

// ExplainInterestModels 解释利率模型
func (s *InterestModelSimulator) ExplainInterestModels() map[string]interface{} {
	return map[string]interface{}{
		"purpose": "动态调节借贷利率，平衡供需",
		"models": []map[string]interface{}{
			{
				"name":    "线性模型",
				"formula": "borrowRate = baseRate + utilization × multiplier",
				"pros":    []string{"简单易懂", "可预测"},
				"cons":    []string{"不能有效应对高利用率"},
				"used_by": "早期协议",
			},
			{
				"name": "跳跃利率模型",
				"formula": map[string]string{
					"below_kink": "borrowRate = baseRate + (U/U_opt) × multiplier",
					"above_kink": "borrowRate = normalRate + ((U-U_opt)/(1-U_opt)) × jumpMultiplier",
				},
				"pros":    []string{"激励最优利用率", "防止流动性枯竭"},
				"cons":    []string{"参数设置复杂"},
				"used_by": "Compound, Aave",
			},
			{
				"name":    "动态模型",
				"formula": "参数根据市场条件自动调整",
				"pros":    []string{"适应性强", "更高效"},
				"cons":    []string{"复杂度高", "可能不稳定"},
				"used_by": "Aave V3的可变利率模式",
			},
		},
		"key_insight": "利率模型的目标是保持利用率在最优区间(通常80%左右)",
	}
}

// =============================================================================
// 利率计算
// =============================================================================

// CalculateBorrowRate 计算借款利率
func (s *InterestModelSimulator) CalculateBorrowRate(utilization float64) float64 {
	switch s.modelType {
	case ModelLinear:
		return s.calculateLinearRate(utilization)
	case ModelJumpRate:
		return s.calculateJumpRate(utilization)
	case ModelDynamic:
		return s.calculateDynamicRate(utilization)
	default:
		return s.calculateJumpRate(utilization)
	}
}

// calculateLinearRate 线性利率
func (s *InterestModelSimulator) calculateLinearRate(utilization float64) float64 {
	// borrowRate = baseRate + utilization × multiplier
	return s.params.BaseRate + (utilization/100)*s.params.Multiplier
}

// calculateJumpRate 跳跃利率
func (s *InterestModelSimulator) calculateJumpRate(utilization float64) float64 {
	optimalU := s.params.OptimalUtilization

	if utilization <= optimalU {
		// 低于最优利用率: 线性增长
		return s.params.BaseRate + (utilization/optimalU)*s.params.Multiplier
	}

	// 高于最优利用率: 急剧增长
	normalRate := s.params.BaseRate + s.params.Multiplier
	excessUtilization := utilization - optimalU
	maxExcess := 100 - optimalU
	return normalRate + (excessUtilization/maxExcess)*s.params.JumpMultiplier
}

// calculateDynamicRate 动态利率
func (s *InterestModelSimulator) calculateDynamicRate(utilization float64) float64 {
	// 动态模型: 根据利用率偏离程度调整斜率
	optimalU := s.params.OptimalUtilization
	deviation := math.Abs(utilization - optimalU)

	// 偏离越大，利率调整越激进
	dynamicMultiplier := s.params.Multiplier * (1 + deviation/50)

	if utilization <= optimalU {
		return s.params.BaseRate + (utilization/optimalU)*dynamicMultiplier
	}

	normalRate := s.params.BaseRate + dynamicMultiplier
	excessUtilization := utilization - optimalU
	maxExcess := 100 - optimalU
	dynamicJump := s.params.JumpMultiplier * (1 + deviation/25)
	return normalRate + (excessUtilization/maxExcess)*dynamicJump
}

// CalculateSupplyRate 计算存款利率
func (s *InterestModelSimulator) CalculateSupplyRate(utilization float64) float64 {
	borrowRate := s.CalculateBorrowRate(utilization)
	// supplyRate = borrowRate × utilization × (1 - reserveFactor)
	return borrowRate * (utilization / 100) * (1 - s.params.ReserveFactor)
}

// =============================================================================
// 利率曲线
// =============================================================================

// GenerateRateCurve 生成利率曲线
func (s *InterestModelSimulator) GenerateRateCurve() []*InterestRatePoint {
	curve := make([]*InterestRatePoint, 0)

	for u := 0.0; u <= 100; u += 5 {
		point := &InterestRatePoint{
			Utilization: u,
			BorrowRate:  s.CalculateBorrowRate(u),
			SupplyRate:  s.CalculateSupplyRate(u),
		}
		curve = append(curve, point)
	}

	return curve
}

// CompareModels 比较不同模型
func (s *InterestModelSimulator) CompareModels() map[string]interface{} {
	originalType := s.modelType

	results := make(map[string]interface{})

	for _, modelType := range []InterestModelType{ModelLinear, ModelJumpRate, ModelDynamic} {
		s.modelType = modelType
		curve := s.GenerateRateCurve()

		rates := make([]map[string]interface{}, 0)
		for _, point := range curve {
			rates = append(rates, map[string]interface{}{
				"utilization": point.Utilization,
				"borrow_rate": fmt.Sprintf("%.2f%%", point.BorrowRate),
				"supply_rate": fmt.Sprintf("%.2f%%", point.SupplyRate),
			})
		}

		results[string(modelType)] = rates
	}

	s.modelType = originalType
	return results
}

// AnalyzeUtilizationImpact 分析利用率影响
func (s *InterestModelSimulator) AnalyzeUtilizationImpact() map[string]interface{} {
	keyPoints := []float64{0, 25, 50, 75, 80, 85, 90, 95, 100}
	analysis := make([]map[string]interface{}, 0)

	for _, u := range keyPoints {
		borrowRate := s.CalculateBorrowRate(u)
		supplyRate := s.CalculateSupplyRate(u)
		spread := borrowRate - supplyRate

		analysis = append(analysis, map[string]interface{}{
			"utilization": fmt.Sprintf("%.0f%%", u),
			"borrow_rate": fmt.Sprintf("%.2f%%", borrowRate),
			"supply_rate": fmt.Sprintf("%.2f%%", supplyRate),
			"spread":      fmt.Sprintf("%.2f%%", spread),
			"zone":        s.getUtilizationZone(u),
		})
	}

	return map[string]interface{}{
		"model":      s.modelType,
		"parameters": s.params,
		"analysis":   analysis,
		"insight": map[string]string{
			"low_utilization":  "低利率吸引借款人，但存款人收益低",
			"optimal_zone":     "供需平衡，协议效率最高",
			"high_utilization": "高利率激励存款，抑制借款，保护流动性",
		},
	}
}

// getUtilizationZone 获取利用率区间
func (s *InterestModelSimulator) getUtilizationZone(utilization float64) string {
	if utilization < 50 {
		return "低利用率区"
	} else if utilization <= s.params.OptimalUtilization {
		return "最优区间"
	} else if utilization <= 90 {
		return "警告区间"
	}
	return "危险区间"
}

// SimulateRateChange 模拟利率变化
func (s *InterestModelSimulator) SimulateRateChange(currentUtilization, newUtilization float64) map[string]interface{} {
	oldBorrowRate := s.CalculateBorrowRate(currentUtilization)
	oldSupplyRate := s.CalculateSupplyRate(currentUtilization)
	newBorrowRate := s.CalculateBorrowRate(newUtilization)
	newSupplyRate := s.CalculateSupplyRate(newUtilization)

	result := map[string]interface{}{
		"utilization_change": fmt.Sprintf("%.1f%% → %.1f%%", currentUtilization, newUtilization),
		"borrow_rate_change": fmt.Sprintf("%.2f%% → %.2f%%", oldBorrowRate, newBorrowRate),
		"supply_rate_change": fmt.Sprintf("%.2f%% → %.2f%%", oldSupplyRate, newSupplyRate),
		"borrow_rate_delta":  fmt.Sprintf("%+.2f%%", newBorrowRate-oldBorrowRate),
		"supply_rate_delta":  fmt.Sprintf("%+.2f%%", newSupplyRate-oldSupplyRate),
	}

	point := &InterestRatePoint{
		Utilization: newUtilization,
		BorrowRate:  newBorrowRate,
		SupplyRate:  newSupplyRate,
	}
	s.rateHistory = append(s.rateHistory, point)

	s.EmitEvent("rate_changed", "", "", result)
	s.updateState()

	return result
}

// GetModelParameters 获取模型参数
func (s *InterestModelSimulator) GetModelParameters() map[string]interface{} {
	return map[string]interface{}{
		"model_type":          s.modelType,
		"base_rate":           fmt.Sprintf("%.1f%%", s.params.BaseRate),
		"multiplier":          fmt.Sprintf("%.1f%%", s.params.Multiplier),
		"jump_multiplier":     fmt.Sprintf("%.1f%%", s.params.JumpMultiplier),
		"optimal_utilization": fmt.Sprintf("%.0f%%", s.params.OptimalUtilization),
		"reserve_factor":      fmt.Sprintf("%.0f%%", s.params.ReserveFactor*100),
	}
}

// UpdateParameters 更新参数
func (s *InterestModelSimulator) UpdateParameters(baseRate, multiplier, jumpMultiplier, optimalU float64) {
	s.params.BaseRate = baseRate
	s.params.Multiplier = multiplier
	s.params.JumpMultiplier = jumpMultiplier
	s.params.OptimalUtilization = optimalU

	s.EmitEvent("parameters_updated", "", "", map[string]interface{}{
		"base_rate":           baseRate,
		"multiplier":          multiplier,
		"jump_multiplier":     jumpMultiplier,
		"optimal_utilization": optimalU,
	})

	s.updateState()
}

// updateState 更新状态
func (s *InterestModelSimulator) updateState() {
	s.SetGlobalData("model_type", string(s.modelType))
	s.SetGlobalData("base_rate", s.params.BaseRate)
	s.SetGlobalData("optimal_utilization", s.params.OptimalUtilization)
	s.SetGlobalData("history_count", len(s.rateHistory))

	summary := fmt.Sprintf("当前利率模型为 %s，已记录 %d 次利率变化。", s.modelType, len(s.rateHistory))
	nextHint := "可以继续改变利用率或切换模型，观察借款利率和存款利率如何联动变化。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"interest_curve",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"model_type": s.modelType, "history_count": len(s.rateHistory)},
	)
}

func (s *InterestModelSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_rate_change":
		currentUtilization := 50.0
		newUtilization := 80.0
		if raw, ok := params["current_utilization"].(float64); ok {
			currentUtilization = raw
		}
		if raw, ok := params["new_utilization"].(float64); ok {
			newUtilization = raw
		}
		result := s.SimulateRateChange(currentUtilization, newUtilization)
		return defiActionResult("已模拟一次利用率变化。", result, &types.ActionFeedback{
			Summary:     "借款利率和存款利率已随着利用率变化重新计算。",
			NextHint:    "继续切换不同模型，比较线性、跳跃和动态模型的曲线差异。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"model_type": s.modelType, "new_utilization": newUtilization},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported interest model action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// InterestModelFactory 利率模型工厂
type InterestModelFactory struct{}

// Create 创建演示器
func (f *InterestModelFactory) Create() engine.Simulator {
	return NewInterestModelSimulator()
}

// GetDescription 获取描述
func (f *InterestModelFactory) GetDescription() types.Description {
	return NewInterestModelSimulator().GetDescription()
}

// NewInterestModelFactory 创建工厂
func NewInterestModelFactory() *InterestModelFactory {
	return &InterestModelFactory{}
}

var _ engine.SimulatorFactory = (*InterestModelFactory)(nil)
