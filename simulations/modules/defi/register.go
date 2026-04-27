package defi

import (
	"github.com/chainspace/simulations/pkg/engine"
)

// RegisterAll 注册所有DeFi模块
func RegisterAll(registry *engine.Registry) {
	// 1. AMM自动做市商
	registry.Register("amm", NewAMMFactory())

	// 2. 集中流动性 (Uniswap V3)
	registry.Register("concentrated_liquidity", NewConcentratedLiquidityFactory())

	// 3. 流动性池
	registry.Register("liquidity_pool", NewLiquidityPoolFactory())

	// 4. 借贷协议
	registry.Register("lending", NewLendingFactory())

	// 5. 利率模型
	registry.Register("interest_model", NewInterestModelFactory())

	// 6. 稳定币
	registry.Register("stablecoin", NewStablecoinFactory())

	// 7. 清算机制
	registry.Register("liquidation", NewLiquidationFactory())

	// 8. 治理
	registry.Register("governance", NewDeFiGovernanceFactory())

	// 9. veToken
	registry.Register("ve_token", NewVeTokenFactory())

	// 10. 收益聚合器
	registry.Register("yield_aggregator", NewYieldAggregatorFactory())

	// 11. 保险
	registry.Register("insurance", NewInsuranceFactory())

	// 12. 永续合约
	registry.Register("perpetual", NewPerpetualFactory())

	// 13. 期权
	registry.Register("options", NewOptionsFactory())
}

// GetModuleList 获取模块列表
func GetModuleList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "amm",
			"name":        "AMM自动做市商演示器",
			"description": "演示恒定乘积、恒定和、Curve等AMM曲线的工作原理",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "concentrated_liquidity",
			"name":        "集中流动性演示器",
			"description": "演示Uniswap V3的集中流动性、价格区间、Tick系统等机制",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "liquidity_pool",
			"name":        "流动性池演示器",
			"description": "演示流动性提供、LP代币、无常损失等核心概念",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "lending",
			"name":        "借贷协议演示器",
			"description": "演示超额抵押借贷、利率模型、清算机制等DeFi借贷核心概念",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "interest_model",
			"name":        "利率模型演示器",
			"description": "演示线性、跳跃利率、动态利率等DeFi借贷利率模型",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "stablecoin",
			"name":        "稳定币演示器",
			"description": "演示法币抵押、超额抵押、算法稳定币等不同类型稳定币的机制",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "liquidation",
			"name":        "清算机制演示器",
			"description": "演示DeFi借贷协议的清算流程、清算奖励、MEV策略等",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "governance",
			"name":        "DeFi治理演示器",
			"description": "演示DAO治理的提案、投票、时间锁、执行等完整流程",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "ve_token",
			"name":        "veToken演示器",
			"description": "演示投票托管代币的锁定、投票权重、Gauge投票等机制",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "yield_aggregator",
			"name":        "收益聚合器演示器",
			"description": "演示自动复利、策略切换、金库机制等收益聚合策略",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "insurance",
			"name":        "DeFi保险演示器",
			"description": "演示保险池、保费定价、理赔投票等DeFi保险机制",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "perpetual",
			"name":        "永续合约演示器",
			"description": "演示永续合约的资金费率、杠杆、清算等核心机制",
			"type":        "defi",
			"category":    "defi",
		},
		{
			"id":          "options",
			"name":        "期权演示器",
			"description": "演示期权定价(Black-Scholes)、Greeks、行权等DeFi期权机制",
			"type":        "defi",
			"category":    "defi",
		},
	}
}
