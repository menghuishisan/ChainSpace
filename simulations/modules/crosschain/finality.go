package crosschain

import (
	"fmt"
	"math"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 最终性演示器
// 演示区块链最终性的核心概念，这对跨链安全至关重要
//
// 最终性类型:
// 1. 概率性最终性: 重组概率随确认数指数下降(PoW)
// 2. 绝对最终性: 一旦确认就不可逆(BFT)
// 3. 经济最终性: 逆转需要巨大经济损失(PoS with Slashing)
//
// 跨链影响:
// - 跨链桥必须等待源链最终性
// - 不同链有不同的等待时间
// - 需要权衡安全性和用户体验
//
// 参考: Bitcoin 6确认, Ethereum 64 slots, Cosmos 即时最终性
// =============================================================================

// FinalityType 最终性类型
type FinalityType string

const (
	FinalityProbabilistic FinalityType = "probabilistic"
	FinalityAbsolute      FinalityType = "absolute"
	FinalityEconomic      FinalityType = "economic"
)

// ChainFinalityConfig 链最终性配置
type ChainFinalityConfig struct {
	ChainID          string        `json:"chain_id"`
	ChainName        string        `json:"chain_name"`
	FinalityType     FinalityType  `json:"finality_type"`
	ConsensusType    string        `json:"consensus_type"`
	BlockTime        time.Duration `json:"block_time"`
	ConfirmBlocks    int           `json:"confirm_blocks"`
	FinalityTime     time.Duration `json:"finality_time"`
	ReorgProbability float64       `json:"reorg_probability"`
	SlashingEnabled  bool          `json:"slashing_enabled"`
	ValidatorCount   int           `json:"validator_count"`
	SecurityBudget   float64       `json:"security_budget_usd"`
}

// FinalityProgress 最终性进度
type FinalityProgress struct {
	BlockHeight   uint64        `json:"block_height"`
	BlockHash     string        `json:"block_hash"`
	Confirmations int           `json:"confirmations"`
	ReorgRisk     float64       `json:"reorg_risk"`
	IsFinal       bool          `json:"is_final"`
	TimeToFinal   time.Duration `json:"time_to_final"`
	Timestamp     time.Time     `json:"timestamp"`
}

// ReorgEvent 重组事件
type ReorgEvent struct {
	ChainID       string    `json:"chain_id"`
	OldHeight     uint64    `json:"old_height"`
	NewHeight     uint64    `json:"new_height"`
	Depth         int       `json:"depth"`
	BlocksRemoved []string  `json:"blocks_removed"`
	BlocksAdded   []string  `json:"blocks_added"`
	Timestamp     time.Time `json:"timestamp"`
}

// FinalitySimulator 最终性演示器
type FinalitySimulator struct {
	*base.BaseSimulator
	chains       map[string]*ChainFinalityConfig
	progress     map[string]*FinalityProgress
	reorgHistory []*ReorgEvent
}

// NewFinalitySimulator 创建最终性演示器
func NewFinalitySimulator() *FinalitySimulator {
	sim := &FinalitySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"finality",
			"最终性演示器",
			"演示概率性最终性、绝对最终性、经济最终性等区块链最终性机制及其对跨链的影响",
			"crosschain",
			types.ComponentProcess,
		),
		chains:       make(map[string]*ChainFinalityConfig),
		progress:     make(map[string]*FinalityProgress),
		reorgHistory: make([]*ReorgEvent, 0),
	}

	return sim
}

// Init 初始化
func (s *FinalitySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.chains = make(map[string]*ChainFinalityConfig)
	s.progress = make(map[string]*FinalityProgress)
	s.reorgHistory = make([]*ReorgEvent, 0)

	s.initializeChains()
	s.updateState()
	return nil
}

// initializeChains 初始化链配置
func (s *FinalitySimulator) initializeChains() {
	s.chains = map[string]*ChainFinalityConfig{
		"bitcoin": {
			ChainID: "bitcoin", ChainName: "Bitcoin",
			FinalityType: FinalityProbabilistic, ConsensusType: "PoW",
			BlockTime: 10 * time.Minute, ConfirmBlocks: 6,
			FinalityTime: 60 * time.Minute, ReorgProbability: 0.001,
			SecurityBudget: 30000000000,
		},
		"ethereum": {
			ChainID: "ethereum", ChainName: "Ethereum",
			FinalityType: FinalityEconomic, ConsensusType: "PoS",
			BlockTime: 12 * time.Second, ConfirmBlocks: 64,
			FinalityTime: 13 * time.Minute, ReorgProbability: 0.0001,
			SlashingEnabled: true, ValidatorCount: 500000,
			SecurityBudget: 50000000000,
		},
		"cosmos": {
			ChainID: "cosmos", ChainName: "Cosmos Hub",
			FinalityType: FinalityAbsolute, ConsensusType: "Tendermint BFT",
			BlockTime: 6 * time.Second, ConfirmBlocks: 1,
			FinalityTime: 6 * time.Second, ReorgProbability: 0,
			SlashingEnabled: true, ValidatorCount: 180,
		},
		"polygon": {
			ChainID: "polygon", ChainName: "Polygon PoS",
			FinalityType: FinalityProbabilistic, ConsensusType: "PoS",
			BlockTime: 2 * time.Second, ConfirmBlocks: 256,
			FinalityTime: 8 * time.Minute, ReorgProbability: 0.01,
			SlashingEnabled: true, ValidatorCount: 100,
		},
		"solana": {
			ChainID: "solana", ChainName: "Solana",
			FinalityType: FinalityAbsolute, ConsensusType: "Tower BFT",
			BlockTime: 400 * time.Millisecond, ConfirmBlocks: 32,
			FinalityTime: 13 * time.Second, ReorgProbability: 0,
			ValidatorCount: 1900,
		},
		"avalanche": {
			ChainID: "avalanche", ChainName: "Avalanche C-Chain",
			FinalityType: FinalityAbsolute, ConsensusType: "Snowman",
			BlockTime: 2 * time.Second, ConfirmBlocks: 1,
			FinalityTime: 2 * time.Second, ReorgProbability: 0,
			ValidatorCount: 1200,
		},
		"arbitrum": {
			ChainID: "arbitrum", ChainName: "Arbitrum One",
			FinalityType: FinalityEconomic, ConsensusType: "Optimistic Rollup",
			BlockTime: 250 * time.Millisecond, ConfirmBlocks: 1,
			FinalityTime:     7 * 24 * time.Hour,
			ReorgProbability: 0.0001,
		},
	}
}

// =============================================================================
// 最终性机制解释
// =============================================================================

// ExplainFinality 解释最终性
func (s *FinalitySimulator) ExplainFinality() map[string]interface{} {
	return map[string]interface{}{
		"definition": "最终性是指交易/区块不可被逆转的保证程度",
		"importance_for_crosschain": []string{
			"跨链桥必须等待源链最终性才能安全执行",
			"不同链有不同的最终性时间",
			"过早执行可能导致双花攻击",
		},
		"types": []map[string]interface{}{
			{
				"type":        "概率性最终性 (Probabilistic)",
				"description": "重组概率随确认数指数下降，但永不为零",
				"formula":     "P(reorg) ≈ (attacker_power / total_power)^confirmations",
				"examples":    []string{"Bitcoin", "Ethereum PoW", "Litecoin"},
				"pros":        []string{"简单", "高度去中心化"},
				"cons":        []string{"需要等待", "不是100%确定", "易受51%攻击"},
			},
			{
				"type":        "绝对最终性 (Absolute/Instant)",
				"description": "一旦达成共识就不可逆转",
				"mechanism":   "2/3+验证者签名确认",
				"examples":    []string{"Cosmos/Tendermint", "Avalanche", "Algorand"},
				"pros":        []string{"即时确定", "无需等待"},
				"cons":        []string{"需要验证者协调", "可能有活性问题"},
			},
			{
				"type":        "经济最终性 (Economic)",
				"description": "逆转需要承担巨大经济损失",
				"mechanism":   "Slashing惩罚机制",
				"examples":    []string{"Ethereum PoS", "Polkadot"},
				"pros":        []string{"强经济保证", "可量化安全性"},
				"cons":        []string{"需要等待finalized", "依赖经济假设"},
			},
		},
		"attack_scenarios": []map[string]string{
			{"attack": "51%攻击", "target": "PoW链", "defense": "增加确认数"},
			{"attack": "长程攻击", "target": "PoS链", "defense": "弱主观性检查点"},
			{"attack": "重组攻击", "target": "跨链桥", "defense": "等待足够最终性"},
		},
	}
}

// =============================================================================
// 最终性操作
// =============================================================================

// CalculateReorgRisk 计算重组风险
func (s *FinalitySimulator) CalculateReorgRisk(chainID string, confirmations int) map[string]interface{} {
	chain, ok := s.chains[chainID]
	if !ok {
		return map[string]interface{}{"error": "链不存在"}
	}

	var reorgRisk float64
	var isFinal bool
	var timeToFinal time.Duration

	switch chain.FinalityType {
	case FinalityProbabilistic:
		attackerPower := 0.3
		reorgRisk = math.Pow(attackerPower, float64(confirmations))
		isFinal = confirmations >= chain.ConfirmBlocks
		if !isFinal {
			remainingBlocks := chain.ConfirmBlocks - confirmations
			timeToFinal = time.Duration(remainingBlocks) * chain.BlockTime
		}

	case FinalityAbsolute:
		if confirmations >= chain.ConfirmBlocks {
			reorgRisk = 0
			isFinal = true
		} else {
			reorgRisk = 1
			isFinal = false
			remainingBlocks := chain.ConfirmBlocks - confirmations
			timeToFinal = time.Duration(remainingBlocks) * chain.BlockTime
		}

	case FinalityEconomic:
		if confirmations >= chain.ConfirmBlocks {
			reorgRisk = 0
			isFinal = true
		} else {
			progress := float64(confirmations) / float64(chain.ConfirmBlocks)
			reorgRisk = chain.ReorgProbability * (1 - progress)
			isFinal = false
			remainingBlocks := chain.ConfirmBlocks - confirmations
			timeToFinal = time.Duration(remainingBlocks) * chain.BlockTime
		}
	}

	return map[string]interface{}{
		"chain":          chain.ChainName,
		"finality_type":  string(chain.FinalityType),
		"confirmations":  confirmations,
		"required":       chain.ConfirmBlocks,
		"reorg_risk":     fmt.Sprintf("%.10f%%", reorgRisk*100),
		"is_final":       isFinal,
		"time_to_final":  timeToFinal.String(),
		"block_time":     chain.BlockTime.String(),
		"recommendation": s.getRecommendation(chain, confirmations),
	}
}

// getRecommendation 获取建议
func (s *FinalitySimulator) getRecommendation(chain *ChainFinalityConfig, confirmations int) string {
	if confirmations >= chain.ConfirmBlocks {
		return "✅ 已达到推荐最终性，可安全执行跨链操作"
	}

	progress := float64(confirmations) / float64(chain.ConfirmBlocks) * 100
	if progress >= 80 {
		return fmt.Sprintf("⏳ 接近最终性(%.0f%%)，建议再等待%d个区块",
			progress, chain.ConfirmBlocks-confirmations)
	}

	return fmt.Sprintf("⚠️ 最终性不足(%.0f%%)，强烈建议等待%d个区块",
		progress, chain.ConfirmBlocks-confirmations)
}

// CompareChainFinality 比较链最终性
func (s *FinalitySimulator) CompareChainFinality() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, chain := range s.chains {
		result = append(result, map[string]interface{}{
			"chain":          chain.ChainName,
			"type":           string(chain.FinalityType),
			"consensus":      chain.ConsensusType,
			"block_time":     chain.BlockTime.String(),
			"confirm_blocks": chain.ConfirmBlocks,
			"finality_time":  chain.FinalityTime.String(),
			"reorg_risk":     fmt.Sprintf("%.6f%%", chain.ReorgProbability*100),
			"validators":     chain.ValidatorCount,
		})
	}

	return result
}

// SimulateFinalityProgress 模拟最终性进度
func (s *FinalitySimulator) SimulateFinalityProgress(chainID string) map[string]interface{} {
	chain, ok := s.chains[chainID]
	if !ok {
		return map[string]interface{}{"error": "链不存在"}
	}

	progress := make([]map[string]interface{}, 0)
	blockHeight := uint64(1000000)

	for i := 0; i <= chain.ConfirmBlocks; i++ {
		risk := s.CalculateReorgRisk(chainID, i)

		progress = append(progress, map[string]interface{}{
			"confirmations": i,
			"time_elapsed":  (time.Duration(i) * chain.BlockTime).String(),
			"reorg_risk":    risk["reorg_risk"],
			"is_final":      risk["is_final"],
			"status": func() string {
				if risk["is_final"].(bool) {
					return "✅ FINALIZED"
				}
				return "⏳ PENDING"
			}(),
		})
	}

	return map[string]interface{}{
		"chain":         chain.ChainName,
		"block_height":  blockHeight,
		"finality_type": string(chain.FinalityType),
		"progress":      progress,
		"summary": map[string]interface{}{
			"required_confirmations": chain.ConfirmBlocks,
			"finality_time":          chain.FinalityTime.String(),
			"block_time":             chain.BlockTime.String(),
		},
	}
}

// ExplainReorgRisk 解释重组风险
func (s *FinalitySimulator) ExplainReorgRisk() map[string]interface{} {
	return map[string]interface{}{
		"what_is_reorg": "区块链重组是指已确认的区块被更长/更重的链替代",
		"causes": []string{
			"网络分区后合并",
			"矿工/验证者自然分叉",
			"51%攻击或长程攻击",
		},
		"consequences": []string{
			"交易被逆转(回滚)",
			"双花攻击成功",
			"跨链桥资金损失",
		},
		"crosschain_attack_scenario": map[string]interface{}{
			"name": "源链重组攻击",
			"steps": []string{
				"1. 攻击者在源链发起跨链存款",
				"2. 跨链桥在目标链释放资金",
				"3. 攻击者重组源链，取消存款交易",
				"4. 攻击者同时拥有源链和目标链资金",
			},
			"prevention": "等待足够确认数达到最终性",
		},
		"real_incidents": []map[string]string{
			{"chain": "Ethereum Classic", "date": "2020-08", "depth": "7000+区块"},
			{"chain": "Bitcoin Gold", "date": "2018-05", "loss": "$18M双花"},
			{"chain": "Verge", "date": "2018-04", "depth": "多次深度重组"},
		},
	}
}

// GetChainInfo 获取链信息
func (s *FinalitySimulator) GetChainInfo(chainID string) map[string]interface{} {
	chain, ok := s.chains[chainID]
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"chain_id":          chain.ChainID,
		"chain_name":        chain.ChainName,
		"finality_type":     string(chain.FinalityType),
		"consensus":         chain.ConsensusType,
		"block_time":        chain.BlockTime.String(),
		"confirm_blocks":    chain.ConfirmBlocks,
		"finality_time":     chain.FinalityTime.String(),
		"reorg_probability": chain.ReorgProbability,
		"slashing_enabled":  chain.SlashingEnabled,
		"validator_count":   chain.ValidatorCount,
	}
}

// GetStatistics 获取统计
func (s *FinalitySimulator) GetStatistics() map[string]interface{} {
	probabilistic := 0
	absolute := 0
	economic := 0

	for _, chain := range s.chains {
		switch chain.FinalityType {
		case FinalityProbabilistic:
			probabilistic++
		case FinalityAbsolute:
			absolute++
		case FinalityEconomic:
			economic++
		}
	}

	return map[string]interface{}{
		"chain_count":          len(s.chains),
		"probabilistic_chains": probabilistic,
		"absolute_chains":      absolute,
		"economic_chains":      economic,
		"reorg_events":         len(s.reorgHistory),
	}
}

// updateState 更新状态
func (s *FinalitySimulator) updateState() {
	s.SetGlobalData("chain_count", len(s.chains))

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"finality",
		"当前正在比较不同链的最终性速度与重组风险。",
		"执行一次最终性模拟，观察确认数增加后重组风险如何下降。",
		0,
		map[string]interface{}{
			"chain_count": len(s.chains),
		},
	)
}

func (s *FinalitySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_finality":
		chainID := "ethereum"
		if raw, ok := params["chain_id"].(string); ok && raw != "" {
			chainID = raw
		}
		result := s.SimulateFinalityProgress(chainID)
		return crosschainActionResult(
			"已模拟链上最终性推进",
			result,
			&types.ActionFeedback{
				Summary:     "当前链的确认数与重组风险已经推演完成，可继续比较不同链的最终性差异。",
				NextHint:    "切换到另一条链再模拟一次，对比最终确认时间和安全边界。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported finality action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// FinalityFactory 最终性工厂
type FinalityFactory struct{}

func (f *FinalityFactory) Create() engine.Simulator { return NewFinalitySimulator() }
func (f *FinalityFactory) GetDescription() types.Description {
	return NewFinalitySimulator().GetDescription()
}
func NewFinalityFactory() *FinalityFactory { return &FinalityFactory{} }

var _ engine.SimulatorFactory = (*FinalityFactory)(nil)
