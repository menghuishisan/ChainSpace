package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

type Position struct {
	ID               string    `json:"id"`
	Owner            string    `json:"owner"`
	CollateralToken  string    `json:"collateral_token"`
	CollateralAmount *big.Int  `json:"collateral_amount"`
	CollateralPrice  *big.Int  `json:"collateral_price"`
	DebtToken        string    `json:"debt_token"`
	DebtAmount       *big.Int  `json:"debt_amount"`
	CollateralValue  *big.Int  `json:"collateral_value"`
	DebtValue        *big.Int  `json:"debt_value"`
	CollateralRatio  float64   `json:"collateral_ratio"`
	LiquidationRatio float64   `json:"liquidation_ratio"`
	Liquidatable     bool      `json:"liquidatable"`
	HealthFactor     float64   `json:"health_factor"`
	CreatedAt        time.Time `json:"created_at"`
}

type LiquidationAttack struct {
	ID               string    `json:"id"`
	AttackType       string    `json:"attack_type"`
	TargetPosition   *Position `json:"target_position"`
	OriginalPrice    *big.Int  `json:"original_price"`
	ManipulatedPrice *big.Int  `json:"manipulated_price"`
	FlashloanAmount  *big.Int  `json:"flashloan_amount"`
	LiquidationBonus *big.Int  `json:"liquidation_bonus"`
	Profit           *big.Int  `json:"profit"`
	GasCost          *big.Int  `json:"gas_cost"`
	Success          bool      `json:"success"`
	Timestamp        time.Time `json:"timestamp"`
}

type LendingProtocol struct {
	Name             string               `json:"name"`
	TotalDeposits    *big.Int             `json:"total_deposits"`
	TotalBorrows     *big.Int             `json:"total_borrows"`
	LiquidationBonus float64              `json:"liquidation_bonus"`
	LiquidationRatio float64              `json:"liquidation_ratio"`
	CloseFactor      float64              `json:"close_factor"`
	Positions        map[string]*Position `json:"positions"`
	OracleType       string               `json:"oracle_type"`
}

type LiquidationSimulator struct {
	*base.BaseSimulator
	protocol *LendingProtocol
	attacks  []*LiquidationAttack
}

func NewLiquidationSimulator() *LiquidationSimulator {
	sim := &LiquidationSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"liquidation",
			"清算攻击演示器",
			"演示价格操纵清算、抢跑清算、自清算和级联清算等场景。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*LiquidationAttack, 0),
	}

	sim.AddParam(types.Param{
		Key:         "liquidation_bonus",
		Name:        "清算奖励",
		Description: "清算人获得的额外奖励比例，单位为百分比。",
		Type:        types.ParamTypeFloat,
		Default:     5.0,
		Min:         1.0,
		Max:         15.0,
	})
	sim.AddParam(types.Param{
		Key:         "oracle_type",
		Name:        "预言机类型",
		Description: "控制价格来源，用于区分现货、TWAP 等机制。",
		Type:        types.ParamTypeString,
		Default:     "spot",
	})

	return sim
}

func (s *LiquidationSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	liquidationBonus := 5.0
	if value, ok := config.Params["liquidation_bonus"]; ok {
		if typed, ok := value.(float64); ok {
			liquidationBonus = typed
		}
	}

	oracleType := "spot"
	if value, ok := config.Params["oracle_type"]; ok {
		if typed, ok := value.(string); ok {
			oracleType = typed
		}
	}

	s.protocol = &LendingProtocol{
		Name:             "VulnerableLending",
		TotalDeposits:    new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18)),
		TotalBorrows:     new(big.Int).Mul(big.NewInt(500000), big.NewInt(1e18)),
		LiquidationBonus: liquidationBonus / 100,
		LiquidationRatio: 1.5,
		CloseFactor:      0.5,
		Positions:        make(map[string]*Position),
		OracleType:       oracleType,
	}

	s.protocol.Positions["victim1"] = &Position{
		ID:               "pos-1",
		Owner:            "0xVictim1",
		CollateralToken:  "ETH",
		CollateralAmount: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		CollateralPrice:  big.NewInt(200000000000),
		DebtToken:        "USDC",
		DebtAmount:       new(big.Int).Mul(big.NewInt(130000), big.NewInt(1e6)),
		LiquidationRatio: 1.5,
		CreatedAt:        time.Now(),
	}

	s.attacks = make([]*LiquidationAttack, 0)
	s.updatePositionValues()
	s.updateState()
	return nil
}

func (s *LiquidationSimulator) updatePositionValues() {
	for _, pos := range s.protocol.Positions {
		collateralValue := new(big.Int).Mul(pos.CollateralAmount, pos.CollateralPrice)
		collateralValue.Div(collateralValue, new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)))
		pos.CollateralValue = collateralValue
		pos.DebtValue = new(big.Int).Set(pos.DebtAmount)
		if pos.DebtValue.Sign() > 0 {
			ratio := new(big.Float).Quo(new(big.Float).SetInt(pos.CollateralValue), new(big.Float).SetInt(pos.DebtValue))
			pos.CollateralRatio, _ = ratio.Float64()
		}
		pos.HealthFactor = pos.CollateralRatio / pos.LiquidationRatio
		pos.Liquidatable = pos.HealthFactor < 1.0
	}
}

func (s *LiquidationSimulator) SimulatePriceManipulationLiquidation(flashloanETH int64) *LiquidationAttack {
	position := s.protocol.Positions["victim1"]
	originalPrice := new(big.Int).Set(position.CollateralPrice)
	flashloan := new(big.Int).Mul(big.NewInt(flashloanETH), big.NewInt(1e18))
	manipulatedPrice := new(big.Int).Mul(originalPrice, big.NewInt(75))
	manipulatedPrice.Div(manipulatedPrice, big.NewInt(100))

	position.CollateralPrice = manipulatedPrice
	s.updatePositionValues()

	attack := &LiquidationAttack{
		ID:               fmt.Sprintf("attack-%d", len(s.attacks)+1),
		AttackType:       "price_manipulation_liquidation",
		TargetPosition:   position,
		OriginalPrice:    originalPrice,
		ManipulatedPrice: manipulatedPrice,
		FlashloanAmount:  flashloan,
		LiquidationBonus: big.NewInt(5000),
		Profit:           big.NewInt(12000),
		GasCost:          big.NewInt(500),
		Success:          true,
		Timestamp:        time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("price_manipulation_liquidation", "", "", map[string]interface{}{
		"original_price":    formatBigInt(originalPrice),
		"manipulated_price": formatBigInt(manipulatedPrice),
		"summary":           "攻击者通过操纵价格让健康仓位进入可清算区间。",
	})
	s.updateState()
	return attack
}

func (s *LiquidationSimulator) SimulateFrontrunLiquidation() map[string]interface{} {
	data := map[string]interface{}{
		"attack_type": "frontrun_liquidation",
		"summary":     "攻击者在 mempool 中发现清算交易后，提高 gas 抢先执行。",
	}
	s.SetGlobalData("latest_frontrun_liquidation", data)
	s.EmitEvent("frontrun_liquidation", "", "", data)
	s.updateState()
	return data
}

func (s *LiquidationSimulator) SimulateSelfLiquidation() map[string]interface{} {
	data := map[string]interface{}{
		"attack_type": "self_liquidation",
		"summary":     "用户或关联地址主动触发自己的仓位清算，以获得奖励或迁移损失。",
	}
	s.SetGlobalData("latest_self_liquidation", data)
	s.EmitEvent("self_liquidation", "", "", data)
	s.updateState()
	return data
}

func (s *LiquidationSimulator) SimulateCascadeLiquidation() map[string]interface{} {
	data := map[string]interface{}{
		"attack_type": "cascade_liquidation",
		"summary":     "价格继续下跌导致多个仓位连续爆仓，形成连锁清算。",
	}
	s.SetGlobalData("latest_cascade_liquidation", data)
	s.EmitEvent("cascade_liquidation", "", "", data)
	s.updateState()
	return data
}

func (s *LiquidationSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "稳健预言机", "description": "采用 TWAP、多源价格和链下喂价降低瞬时操纵风险。"},
		{"name": "清算冷却与限速", "description": "限制单区块内清算规模，降低连锁冲击。"},
		{"name": "更高健康因子缓冲", "description": "提高最低抵押率，给用户预留补仓空间。"},
		{"name": "私有清算通道", "description": "减少清算交易在公开 mempool 中被抢跑的概率。"},
	}
}

func (s *LiquidationSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Maker Black Thursday", "date": "2020-03", "issue": "价格剧烈波动引发大规模清算"},
		{"name": "DeFi liquidation wars", "date": "2020-2023", "issue": "清算抢跑中 gas 竞价导致收益高度集中"},
	}
}

func (s *LiquidationSimulator) updateState() {
	s.SetGlobalData("position_count", len(s.protocol.Positions))
	s.SetGlobalData("attack_count", len(s.attacks))
	s.SetGlobalData("oracle_type", s.protocol.OracleType)
	if len(s.attacks) == 0 {
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		s.SetGlobalData("victim_balance", 0)
		s.SetGlobalData("attacker_balance", 0)
		s.SetGlobalData("attack_summary", "当前尚未触发清算攻击，可以从价格操纵、抢跑清算、自清算或级联清算开始观察。")
		setAttackTeachingState(
			s.BaseSimulator,
			"economic",
			"idle",
			"等待清算攻击场景。",
			"可以先触发一次价格操纵清算，观察健康因子何时跌破阈值并进入清算窗口。",
			0,
			map[string]interface{}{
				"position_count": len(s.protocol.Positions),
				"oracle_type":    s.protocol.OracleType,
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{
			"step":        1,
			"action":      "prepare_liquidation_window",
			"caller":      "attacker",
			"function":    latest.AttackType,
			"target":      latest.TargetPosition.Owner,
			"amount":      formatBigInt(latest.FlashloanAmount),
			"call_depth":  1,
			"description": "攻击者先制造或捕捉清算窗口，让原本边缘健康的仓位进入危险区间。",
		},
		{
			"step":        2,
			"action":      "trigger_liquidation",
			"caller":      "attacker",
			"function":    "liquidate_position",
			"target":      latest.TargetPosition.ID,
			"amount":      formatBigInt(latest.ManipulatedPrice),
			"call_depth":  2,
			"description": "在价格或顺序被扭曲后，攻击者抢先执行清算，获得额外折价或奖励。",
		},
		{
			"step":        3,
			"action":      "capture_bonus",
			"caller":      "attacker",
			"function":    "settle_profit",
			"target":      latest.TargetPosition.CollateralToken,
			"amount":      formatBigInt(latest.Profit),
			"call_depth":  3,
			"description": "清算完成后，攻击者回收借款并保留清算奖励或折价收益。",
		},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", fmt.Sprintf("目标仓位 %s 当前健康度 %.2f。", latest.TargetPosition.ID, latest.TargetPosition.HealthFactor))
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))
	s.SetGlobalData("victim_balance", parseAmountText(formatBigInt(latest.TargetPosition.CollateralValue)))
	s.SetGlobalData("attacker_balance", parseAmountText(formatBigInt(latest.Profit)))

	setAttackTeachingState(
		s.BaseSimulator,
		"economic",
		"liquidated",
		fmt.Sprintf("目标仓位 %s 已进入清算窗口。", latest.TargetPosition.ID),
		"重点观察价格扭曲、健康因子下探和清算奖励兑现之间的因果链路。",
		1.0,
		map[string]interface{}{
			"target_position":  latest.TargetPosition.ID,
			"health_factor":    latest.TargetPosition.HealthFactor,
			"original_price":   formatBigInt(latest.OriginalPrice),
			"manipulated_price": formatBigInt(latest.ManipulatedPrice),
			"profit":           formatBigInt(latest.Profit),
		},
	)
}

func (s *LiquidationSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_price_manipulation_liquidation":
		flashloanETH := actionInt64(params, "flashloan_eth", 5000)
		attack := s.SimulatePriceManipulationLiquidation(flashloanETH)
		return actionResultWithFeedback(
			"已执行价格操纵清算演示。",
			map[string]interface{}{"attack": attack},
			&types.ActionFeedback{
				Summary:     "已进入通过价格操纵制造清算窗口并收割奖励的攻击流程。",
				NextHint:    "重点观察健康因子在哪一步跌破阈值，以及清算奖励是如何兑现的。",
				EffectScope: "economic",
				ResultState: map[string]interface{}{
					"health_factor": attack.TargetPosition.HealthFactor,
					"profit":        formatBigInt(attack.Profit),
				},
			},
		), nil
	case "simulate_frontrun_liquidation":
		result := s.SimulateFrontrunLiquidation()
		return actionResultWithFeedback(
			"已执行抢跑清算演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入在公开 mempool 中抢先执行清算的攻击流程。",
				NextHint:    "重点观察谁先进入区块，以及清算奖励为何高度依赖排序优势。",
				EffectScope: "economic",
				ResultState: result,
			},
		), nil
	case "simulate_self_liquidation":
		result := s.SimulateSelfLiquidation()
		return actionResultWithFeedback(
			"已执行自清算演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入用户或关联地址主动触发自身清算的攻击流程。",
				NextHint:    "重点观察自清算如何把本应外流的奖励重新回收到自己手中。",
				EffectScope: "economic",
				ResultState: result,
			},
		), nil
	case "simulate_cascade_liquidation":
		result := s.SimulateCascadeLiquidation()
		return actionResultWithFeedback(
			"已执行级联清算演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入价格下跌引发连续清算的攻击流程。",
				NextHint:    "重点观察单个清算如何继续压低市场并诱发下一轮清算。",
				EffectScope: "economic",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

type LiquidationFactory struct{}

func (f *LiquidationFactory) Create() engine.Simulator { return NewLiquidationSimulator() }

func (f *LiquidationFactory) GetDescription() types.Description { return NewLiquidationSimulator().GetDescription() }

func NewLiquidationFactory() *LiquidationFactory { return &LiquidationFactory{} }

var _ engine.SimulatorFactory = (*LiquidationFactory)(nil)
