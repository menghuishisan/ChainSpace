package attacks

import (
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// BriberyType 表示贿赂攻击类型。
type BriberyType string

const (
	BriberyTypePEpsilon   BriberyType = "p_epsilon"
	BriberyTypeDarkPool   BriberyType = "dark_pool"
	BriberyTypeGovernance BriberyType = "governance"
	BriberyTypeCrossChain BriberyType = "cross_chain"
	BriberyTypeMEV        BriberyType = "mev"
)

// BriberyAttack 记录一次贿赂攻击模拟结果。
type BriberyAttack struct {
	ID             string      `json:"id"`
	Type           BriberyType `json:"type"`
	BribeAmount    *big.Int    `json:"bribe_amount"`
	TargetCount    int         `json:"target_count"`
	AcceptedCount  int         `json:"accepted_count"`
	ExpectedProfit *big.Int    `json:"expected_profit"`
	ActualCost     *big.Int    `json:"actual_cost"`
	Success        bool        `json:"success"`
	Timestamp      time.Time   `json:"timestamp"`
}

// BriberySimulator 演示共识和治理场景中的贿赂攻击。
type BriberySimulator struct {
	*base.BaseSimulator
	attacks      []*BriberyAttack
	blockReward  *big.Int
	networkStake *big.Int
}

// NewBriberySimulator 创建贿赂攻击演示器。
func NewBriberySimulator() *BriberySimulator {
	sim := &BriberySimulator{
		BaseSimulator: base.NewBaseSimulator(
			"bribery",
			"贿赂攻击演示器",
			"演示条件性贿赂、治理贿赂和 MEV 贿赂如何改变系统参与者的理性选择。",
			"attacks",
			types.ComponentAttack,
		),
		attacks: make([]*BriberyAttack, 0),
	}

	sim.AddParam(types.Param{
		Key:         "block_reward",
		Name:        "区块奖励",
		Description: "用于估算贿赂成本的基础区块奖励，单位为 ETH。",
		Type:        types.ParamTypeFloat,
		Default:     2.0,
		Min:         0.1,
		Max:         10.0,
	})

	return sim
}

// Init 初始化模拟器。
func (s *BriberySimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	blockReward := 2.0
	if value, ok := config.Params["block_reward"]; ok {
		if typed, ok := value.(float64); ok {
			blockReward = typed
		}
	}

	s.blockReward = new(big.Int).Mul(big.NewInt(int64(blockReward*1000)), big.NewInt(1e15))
	s.networkStake = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18))
	s.attacks = make([]*BriberyAttack, 0)
	s.updateState()
	return nil
}

// ExplainBriberyTypes 返回不同贿赂攻击的定义。
func (s *BriberySimulator) ExplainBriberyTypes() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":        "P+epsilon 贿赂",
			"description": "攻击者只在分叉成功时支付额外奖励，失败时几乎不承担成本。",
			"mechanism": []string{
				"部署条件性支付合约",
				"承诺给支持攻击分叉的参与者额外收益",
				"足够多的节点接受后，攻击变成自我实现",
			},
		},
		{
			"type":        "治理贿赂",
			"description": "通过购买投票权或临时借入治理代币改变协议决策。",
			"mechanism": []string{
				"识别关键提案",
				"计算阈值票数",
				"按票报价以影响治理结果",
			},
		},
		{
			"type":        "MEV 贿赂",
			"description": "搜索者向构建者或验证者支付额外收益，换取交易排序优势。",
			"mechanism": []string{
				"提高 gas 价格",
				"通过私有通道提交 bundle",
				"直接给构建者支付小费",
			},
		},
	}
}

// SimulatePEpsilonAttack 模拟 P+epsilon 贿赂攻击。
func (s *BriberySimulator) SimulatePEpsilonAttack(epsilon float64, minerCount int) *BriberyAttack {
	bribePerMiner := new(big.Int).Mul(s.blockReward, big.NewInt(int64(epsilon*100)))
	bribePerMiner.Div(bribePerMiner, big.NewInt(100))

	totalBribe := new(big.Int).Mul(bribePerMiner, big.NewInt(int64(minerCount)))
	acceptanceRate := 0.7 + epsilon/10
	if acceptanceRate > 0.95 {
		acceptanceRate = 0.95
	}

	acceptedCount := int(float64(minerCount) * acceptanceRate)
	success := acceptedCount > minerCount/2
	actualCost := big.NewInt(0)
	if success {
		actualCost = totalBribe
	}

	attack := &BriberyAttack{
		ID:             fmt.Sprintf("attack-%d", len(s.attacks)+1),
		Type:           BriberyTypePEpsilon,
		BribeAmount:    bribePerMiner,
		TargetCount:    minerCount,
		AcceptedCount:  acceptedCount,
		ExpectedProfit: new(big.Int).Mul(s.blockReward, big.NewInt(10)),
		ActualCost:     actualCost,
		Success:        success,
		Timestamp:      time.Now(),
	}

	s.attacks = append(s.attacks, attack)
	s.EmitEvent("p_epsilon_attack", "", "", map[string]interface{}{
		"success":         success,
		"epsilon":         epsilon,
		"target_count":    minerCount,
		"accepted_count":  acceptedCount,
		"bribe_per_miner": formatEther(bribePerMiner),
		"total_bribe":     formatEther(totalBribe),
	})
	s.updateState()
	return attack
}

// SimulateGovernanceBribery 模拟治理贿赂。
func (s *BriberySimulator) SimulateGovernanceBribery(bribePerVote float64, votesNeeded int64) map[string]interface{} {
	totalCost := float64(votesNeeded) * bribePerVote
	data := map[string]interface{}{
		"attack":         "governance_bribery",
		"bribe_per_vote": bribePerVote,
		"votes_needed":   votesNeeded,
		"total_cost":     totalCost,
		"summary":        "攻击者通过购买投票权改变治理结果。",
	}
	s.SetGlobalData("latest_governance_bribery", data)
	s.updateState()
	return data
}

// SimulateMEVBribery 模拟 MEV 贿赂。
func (s *BriberySimulator) SimulateMEVBribery() map[string]interface{} {
	data := map[string]interface{}{
		"attack":      "mev_bribery",
		"description": "搜索者向构建者支付额外收益，以获得交易排序优势。",
		"types": []map[string]interface{}{
			{"name": "Priority Gas Auction", "cost": "gas"},
			{"name": "Bundle Tips", "cost": "direct bribe"},
			{"name": "Private Relay", "cost": "inclusion payment"},
		},
	}
	s.SetGlobalData("latest_mev_bribery", data)
	s.updateState()
	return data
}

// ShowDefenses 展示防御方式。
func (s *BriberySimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "快速最终性", "description": "缩短重组窗口，降低贿赂攻击的可执行时间。"},
		{"name": "随机委员会", "description": "让攻击者难以提前锁定需要贿赂的对象。"},
		{"name": "惩罚机制", "description": "通过 slash 或信誉惩罚提高接受贿赂的成本。"},
		{"name": "秘密投票与延迟公开", "description": "降低验证者是否履约可被外部验证的概率。"},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *BriberySimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":   "Curve Wars",
			"type":   "治理贿赂",
			"scale":  "数亿美元级别投票激励",
			"detail": "协议通过激励 veToken 持有者引导流动性方向。",
		},
		{
			"name":   "MEV Supply Chain",
			"type":   "MEV 贿赂",
			"scale":  "每日数百万美元级排序收益",
			"detail": "搜索者、构建者和验证者之间形成稳定的贿赂链路。",
		},
	}
}

// updateState 更新状态。
func (s *BriberySimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.attacks))
	s.SetGlobalData("block_reward", formatEther(s.blockReward))

	if len(s.attacks) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未触发贿赂攻击，请观察收益承诺如何改变矿工、验证者或投票人的理性选择。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"consensus",
			"idle",
			"等待贿赂攻击场景。",
			"可以先触发一次 P+epsilon、治理贿赂或 MEV 贿赂，观察收益承诺如何改变参与者决策。",
			0,
			map[string]interface{}{
				"block_reward": formatEther(s.blockReward),
				"attack_count": len(s.attacks),
			},
		)
		return
	}

	latest := s.attacks[len(s.attacks)-1]
	steps := []map[string]interface{}{
		{"index": 1, "title": "提出激励", "description": fmt.Sprintf("攻击者向 %d 个目标提出额外收益承诺。", latest.TargetCount)},
		{"index": 2, "title": "接受报价", "description": fmt.Sprintf("共有 %d 个目标接受了贿赂。", latest.AcceptedCount)},
		{"index": 3, "title": "形成结果", "description": fmt.Sprintf("预期收益 %s，真实成本 %s。", formatEther(latest.ExpectedProfit), formatEther(latest.ActualCost))},
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", fmt.Sprintf("本轮贿赂类型为 %s，重点观察收益承诺是否足以跨越参与者的理性门槛。", latest.Type))
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"consensus",
		map[bool]string{true: "accepted", false: "rejected"}[latest.Success],
		fmt.Sprintf("目标参与者对 %s 贿赂路径作出响应。", latest.Type),
		"重点观察接受贿赂的节点数量是否足够改变最终结果。",
		1.0,
		map[string]interface{}{
			"attack_type":     latest.Type,
			"target_count":    latest.TargetCount,
			"accepted_count":  latest.AcceptedCount,
			"expected_profit": formatEther(latest.ExpectedProfit),
			"actual_cost":     formatEther(latest.ActualCost),
			"success":         latest.Success,
		},
	)
}

// ExecuteAction 执行动作。
func (s *BriberySimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_p_epsilon_attack":
		epsilon := actionFloat64(params, "epsilon", 0.2)
		minerCount := actionInt(params, "miner_count", 10)
		attack := s.SimulatePEpsilonAttack(epsilon, minerCount)
		return actionResultWithFeedback(
			"已完成 P+epsilon 贿赂攻击演示。",
			map[string]interface{}{
				"attack_id":      attack.ID,
				"accepted_count": attack.AcceptedCount,
				"target_count":   attack.TargetCount,
				"success":        attack.Success,
			},
			&types.ActionFeedback{
				Summary:     "已进入条件性收益承诺驱动节点倒向攻击分叉的流程。",
				NextHint:    "重点观察接受贿赂的节点数量是否超过关键阈值。",
				EffectScope: "consensus",
				ResultState: map[string]interface{}{
					"accepted_count": attack.AcceptedCount,
					"target_count":   attack.TargetCount,
					"success":        attack.Success,
				},
			},
		), nil
	case "simulate_governance_bribery":
		bribePerVote := actionFloat64(params, "bribe_per_vote", 50)
		votesNeeded := actionInt64(params, "votes_needed", 100000)
		result := s.SimulateGovernanceBribery(bribePerVote, votesNeeded)
		return actionResultWithFeedback(
			"已完成治理贿赂演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入通过购买投票权来改变治理结果的攻击流程。",
				NextHint:    "重点观察单票价格、总票数需求和提案价值之间的关系。",
				EffectScope: "economic",
				ResultState: result,
			},
		), nil
	case "simulate_mev_bribery":
		result := s.SimulateMEVBribery()
		return actionResultWithFeedback(
			"已完成 MEV 贿赂演示。",
			result,
			&types.ActionFeedback{
				Summary:     "已进入搜索者向构建者或验证者购买排序优势的攻击流程。",
				NextHint:    "重点观察排序收益为何足以驱动额外支付。",
				EffectScope: "consensus",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// BriberyFactory 创建贿赂攻击模拟器。
type BriberyFactory struct{}

// Create 创建模拟器。
func (f *BriberyFactory) Create() engine.Simulator {
	return NewBriberySimulator()
}

// GetDescription 获取描述。
func (f *BriberyFactory) GetDescription() types.Description {
	return NewBriberySimulator().GetDescription()
}

// NewBriberyFactory 创建工厂。
func NewBriberyFactory() *BriberyFactory {
	return &BriberyFactory{}
}

var _ engine.SimulatorFactory = (*BriberyFactory)(nil)
