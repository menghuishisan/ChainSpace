package attacks

import "github.com/chainspace/simulations/pkg/engine"

// RegisterAll 注册所有攻击模块。
func RegisterAll(registry *engine.Registry) {
	// 合约执行与权限利用类攻击。
	registry.Register("reentrancy", NewReentrancyFactory())                        // 重入攻击
	registry.Register("integer_overflow", NewIntegerOverflowFactory())             // 整数溢出
	registry.Register("access_control", NewAccessControlFactory())                 // 访问控制缺陷
	registry.Register("tx_origin", NewTxOriginFactory())                           // tx.origin 钓鱼
	registry.Register("delegatecall_attack", NewDelegatecallAttackFactory())       // 委托调用攻击
	registry.Register("selfdestruct", NewSelfdestructFactory())                    // selfdestruct 攻击
	registry.Register("dos", NewDOSFactory())                                      // 拒绝服务
	registry.Register("signature_replay", NewSignatureReplayFactory())             // 签名重放
	registry.Register("weak_randomness", NewWeakRandomnessFactory())               // 弱随机数
	registry.Register("timestamp_manipulation", NewTimestampManipulationFactory()) // 时间戳操纵

	// DeFi 与经济操纵类攻击。
	registry.Register("flashloan", NewFlashloanFactory())                    // 闪电贷攻击
	registry.Register("oracle_manipulation", NewOracleManipulationFactory()) // 预言机操纵
	registry.Register("sandwich", NewSandwichFactory())                      // 三明治攻击
	registry.Register("frontrun", NewFrontrunFactory())                      // 抢跑攻击
	registry.Register("governance", NewGovernanceFactory())                  // 治理攻击
	registry.Register("liquidation", NewLiquidationFactory())                // 清算攻击
	registry.Register("infinite_mint", NewInfiniteMintFactory())             // 无限增发

	// 共识与链安全类攻击。
	registry.Register("attack_51", NewAttack51Factory())              // 51% 攻击
	registry.Register("selfish_mining", NewSelfishMiningFactory())    // 自私挖矿
	registry.Register("long_range", NewLongRangeFactory())            // 长程攻击
	registry.Register("nothing_at_stake", NewNothingAtStakeFactory()) // 无利害关系
	registry.Register("bribery", NewBriberyFactory())                 // 贿赂攻击

	// 跨链与交易完整性类攻击。
	registry.Register("bridge_attack", NewBridgeAttackFactory()) // 跨链桥攻击
	registry.Register("fake_deposit", NewFakeDepositFactory())   // 假充值攻击
}

// RegisterToEngine 直接向引擎注册所有攻击模块。
func RegisterToEngine(eng *engine.Engine) {
	// 合约执行与权限利用类攻击。
	eng.Register("reentrancy", NewReentrancyFactory())
	eng.Register("integer_overflow", NewIntegerOverflowFactory())
	eng.Register("access_control", NewAccessControlFactory())
	eng.Register("tx_origin", NewTxOriginFactory())
	eng.Register("delegatecall_attack", NewDelegatecallAttackFactory())
	eng.Register("selfdestruct", NewSelfdestructFactory())
	eng.Register("dos", NewDOSFactory())
	eng.Register("signature_replay", NewSignatureReplayFactory())
	eng.Register("weak_randomness", NewWeakRandomnessFactory())
	eng.Register("timestamp_manipulation", NewTimestampManipulationFactory())

	// DeFi 与经济操纵类攻击。
	eng.Register("flashloan", NewFlashloanFactory())
	eng.Register("oracle_manipulation", NewOracleManipulationFactory())
	eng.Register("sandwich", NewSandwichFactory())
	eng.Register("frontrun", NewFrontrunFactory())
	eng.Register("governance", NewGovernanceFactory())
	eng.Register("liquidation", NewLiquidationFactory())
	eng.Register("infinite_mint", NewInfiniteMintFactory())

	// 共识与链安全类攻击。
	eng.Register("attack_51", NewAttack51Factory())
	eng.Register("selfish_mining", NewSelfishMiningFactory())
	eng.Register("long_range", NewLongRangeFactory())
	eng.Register("nothing_at_stake", NewNothingAtStakeFactory())
	eng.Register("bribery", NewBriberyFactory())

	// 跨链与交易完整性类攻击。
	eng.Register("bridge_attack", NewBridgeAttackFactory())
	eng.Register("fake_deposit", NewFakeDepositFactory())
}
