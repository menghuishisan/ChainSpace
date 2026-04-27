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
// DeFi保险演示器
// =============================================================================

// CoverType 保险类型
type CoverType string

const (
	CoverSmartContract CoverType = "smart_contract" // 智能合约风险
	CoverProtocol      CoverType = "protocol"       // 协议风险
	CoverCustody       CoverType = "custody"        // 托管风险
	CoverStablecoin    CoverType = "stablecoin"     // 稳定币脱锚
	CoverOracle        CoverType = "oracle"         // 预言机故障
)

// InsurancePolicy 保险单
type InsurancePolicy struct {
	ID             string    `json:"id"`
	Holder         string    `json:"holder"`
	CoverType      CoverType `json:"cover_type"`
	Protocol       string    `json:"protocol"`     // 被保协议
	CoverAmount    *big.Int  `json:"cover_amount"` // 保额
	Premium        *big.Int  `json:"premium"`      // 保费
	PremiumRate    float64   `json:"premium_rate"` // 年化费率
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	IsActive       bool      `json:"is_active"`
	ClaimSubmitted bool      `json:"claim_submitted"`
	ClaimApproved  bool      `json:"claim_approved"`
}

// InsurancePool 保险池
type InsurancePool struct {
	ID             string   `json:"id"`
	Protocol       string   `json:"protocol"`
	TotalStaked    *big.Int `json:"total_staked"`   // 质押总额
	ActiveCover    *big.Int `json:"active_cover"`   // 活跃保额
	CapacityUsed   float64  `json:"capacity_used"`  // 容量使用率
	PremiumEarned  *big.Int `json:"premium_earned"` // 收取保费
	ClaimsPaid     *big.Int `json:"claims_paid"`    // 已支付理赔
	StakerCount    int      `json:"staker_count"`
	RiskScore      int      `json:"risk_score"`       // 风险评分 1-100
	MinPremiumRate float64  `json:"min_premium_rate"` // 最低费率
}

// Claim 理赔申请
type Claim struct {
	ID           string    `json:"id"`
	PolicyID     string    `json:"policy_id"`
	Claimant     string    `json:"claimant"`
	Amount       *big.Int  `json:"amount"`
	Reason       string    `json:"reason"`
	Evidence     string    `json:"evidence"`
	Status       string    `json:"status"` // pending, voting, approved, rejected
	VotesFor     int       `json:"votes_for"`
	VotesAgainst int       `json:"votes_against"`
	SubmittedAt  time.Time `json:"submitted_at"`
	ResolvedAt   time.Time `json:"resolved_at"`
}

// InsuranceSimulator DeFi保险演示器
// 演示DeFi保险的核心机制:
//
// 1. 保险池
//   - 质押者提供资金支持保险
//   - 收取保费作为收益
//   - 承担理赔风险
//
// 2. 保险购买
//   - 选择协议和保额
//   - 支付保费获得保障
//
// 3. 理赔流程
//   - 提交理赔申请
//   - 社区投票评估
//   - 批准后支付赔偿
//
// 参考: Nexus Mutual, Cover Protocol, InsurAce
type InsuranceSimulator struct {
	*base.BaseSimulator
	pools           map[string]*InsurancePool
	policies        map[string]*InsurancePolicy
	claims          map[string]*Claim
	capacityFactor  float64 // 抵押品可承保倍数
	claimVotePeriod time.Duration
}

// NewInsuranceSimulator 创建保险演示器
func NewInsuranceSimulator() *InsuranceSimulator {
	sim := &InsuranceSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"insurance",
			"DeFi保险演示器",
			"演示保险池、保费定价、理赔投票等DeFi保险机制",
			"defi",
			types.ComponentDeFi,
		),
		pools:    make(map[string]*InsurancePool),
		policies: make(map[string]*InsurancePolicy),
		claims:   make(map[string]*Claim),
	}

	sim.AddParam(types.Param{
		Key:         "capacity_factor",
		Name:        "容量因子",
		Description: "质押资金可承保的倍数",
		Type:        types.ParamTypeFloat,
		Default:     4.0,
		Min:         1.0,
		Max:         10.0,
	})

	return sim
}

// Init 初始化
func (s *InsuranceSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.capacityFactor = 4.0
	s.claimVotePeriod = 72 * time.Hour

	if v, ok := config.Params["capacity_factor"]; ok {
		if f, ok := v.(float64); ok {
			s.capacityFactor = f
		}
	}

	// 初始化保险池 (使用较小的数值避免int64溢出)
	s.pools = make(map[string]*InsurancePool)

	aaveStaked := new(big.Int)
	aaveStaked.SetString("10000000000000000000000000", 10) // 10M * 1e18
	aaveCover := new(big.Int)
	aaveCover.SetString("20000000000000000000000000", 10)
	s.pools["aave"] = &InsurancePool{
		ID:             "aave",
		Protocol:       "Aave",
		TotalStaked:    aaveStaked,
		ActiveCover:    aaveCover,
		PremiumEarned:  big.NewInt(0),
		ClaimsPaid:     big.NewInt(0),
		StakerCount:    150,
		RiskScore:      15,
		MinPremiumRate: 0.026,
	}

	compoundStaked := new(big.Int)
	compoundStaked.SetString("8000000000000000000000000", 10)
	compoundCover := new(big.Int)
	compoundCover.SetString("15000000000000000000000000", 10)
	s.pools["compound"] = &InsurancePool{
		ID:             "compound",
		Protocol:       "Compound",
		TotalStaked:    compoundStaked,
		ActiveCover:    compoundCover,
		PremiumEarned:  big.NewInt(0),
		ClaimsPaid:     big.NewInt(0),
		StakerCount:    120,
		RiskScore:      18,
		MinPremiumRate: 0.028,
	}

	curveStaked := new(big.Int)
	curveStaked.SetString("5000000000000000000000000", 10)
	curveCover := new(big.Int)
	curveCover.SetString("12000000000000000000000000", 10)
	s.pools["curve"] = &InsurancePool{
		ID:             "curve",
		Protocol:       "Curve",
		TotalStaked:    curveStaked,
		ActiveCover:    curveCover,
		PremiumEarned:  big.NewInt(0),
		ClaimsPaid:     big.NewInt(0),
		StakerCount:    80,
		RiskScore:      22,
		MinPremiumRate: 0.032,
	}

	uniswapStaked := new(big.Int)
	uniswapStaked.SetString("6000000000000000000000000", 10)
	uniswapCover := new(big.Int)
	uniswapCover.SetString("18000000000000000000000000", 10)
	s.pools["uniswap"] = &InsurancePool{
		ID:             "uniswap",
		Protocol:       "Uniswap",
		TotalStaked:    uniswapStaked,
		ActiveCover:    uniswapCover,
		PremiumEarned:  big.NewInt(0),
		ClaimsPaid:     big.NewInt(0),
		StakerCount:    100,
		RiskScore:      12,
		MinPremiumRate: 0.022,
	}

	// 更新容量使用率
	for _, pool := range s.pools {
		capacity := new(big.Int).Mul(pool.TotalStaked, big.NewInt(int64(s.capacityFactor)))
		pool.CapacityUsed = float64(pool.ActiveCover.Int64()) / float64(capacity.Int64()) * 100
	}

	s.policies = make(map[string]*InsurancePolicy)
	s.claims = make(map[string]*Claim)

	s.updateState()
	return nil
}

// =============================================================================
// 保险机制解释
// =============================================================================

// ExplainDeFiInsurance 解释DeFi保险
func (s *InsuranceSimulator) ExplainDeFiInsurance() map[string]interface{} {
	return map[string]interface{}{
		"purpose": "为DeFi用户提供智能合约和协议风险保障",
		"covered_risks": []map[string]string{
			{"risk": "智能合约漏洞", "example": "重入攻击、逻辑错误"},
			{"risk": "协议攻击", "example": "闪电贷攻击、价格操纵"},
			{"risk": "预言机故障", "example": "价格偏差导致错误清算"},
			{"risk": "稳定币脱锚", "example": "UST崩盘"},
			{"risk": "托管风险", "example": "中心化托管方跑路"},
		},
		"participants": []map[string]string{
			{"role": "质押者", "action": "质押资金到保险池，承担风险，获得保费收益"},
			{"role": "保险购买者", "action": "支付保费获得保障"},
			{"role": "评估师", "action": "评估协议风险，决定保费费率"},
			{"role": "理赔评审员", "action": "投票决定理赔是否有效"},
		},
		"not_covered": []string{
			"市场价格波动损失",
			"无常损失",
			"用户自身操作失误",
			"网络钓鱼/私钥泄露",
		},
	}
}

// =============================================================================
// 保险购买
// =============================================================================

// BuyCover 购买保险
func (s *InsuranceSimulator) BuyCover(buyer, protocol string, coverType CoverType, amount *big.Int, durationDays int) (*InsurancePolicy, error) {
	pool, ok := s.pools[protocol]
	if !ok {
		return nil, fmt.Errorf("协议保险池不存在")
	}

	// 检查容量
	capacity := new(big.Int).Mul(pool.TotalStaked, big.NewInt(int64(s.capacityFactor)))
	newActiveCover := new(big.Int).Add(pool.ActiveCover, amount)
	if newActiveCover.Cmp(capacity) > 0 {
		return nil, fmt.Errorf("保险容量不足")
	}

	// 计算保费
	annualPremiumRate := pool.MinPremiumRate * (1 + pool.CapacityUsed/100)
	premium := new(big.Int).Mul(amount, big.NewInt(int64(annualPremiumRate*float64(durationDays)/365*10000)))
	premium.Div(premium, big.NewInt(10000))

	policyID := fmt.Sprintf("policy-%s-%d", buyer, time.Now().UnixNano())
	policy := &InsurancePolicy{
		ID:          policyID,
		Holder:      buyer,
		CoverType:   coverType,
		Protocol:    protocol,
		CoverAmount: amount,
		Premium:     premium,
		PremiumRate: annualPremiumRate,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Duration(durationDays) * 24 * time.Hour),
		IsActive:    true,
	}

	s.policies[policyID] = policy
	pool.ActiveCover.Add(pool.ActiveCover, amount)
	pool.PremiumEarned.Add(pool.PremiumEarned, premium)
	pool.CapacityUsed = float64(pool.ActiveCover.Int64()) / float64(capacity.Int64()) * 100

	s.EmitEvent("cover_bought", "", "", map[string]interface{}{
		"policy_id":    policyID,
		"buyer":        buyer,
		"protocol":     protocol,
		"cover_type":   coverType,
		"cover_amount": amount.String(),
		"premium":      premium.String(),
		"duration":     durationDays,
		"premium_rate": fmt.Sprintf("%.2f%%", annualPremiumRate*100),
	})

	s.updateState()
	return policy, nil
}

// =============================================================================
// 质押
// =============================================================================

// StakeToPool 质押到保险池
func (s *InsuranceSimulator) StakeToPool(staker, protocol string, amount *big.Int) error {
	pool, ok := s.pools[protocol]
	if !ok {
		return fmt.Errorf("保险池不存在")
	}

	pool.TotalStaked.Add(pool.TotalStaked, amount)
	pool.StakerCount++

	// 更新容量使用率
	capacity := new(big.Int).Mul(pool.TotalStaked, big.NewInt(int64(s.capacityFactor)))
	pool.CapacityUsed = float64(pool.ActiveCover.Int64()) / float64(capacity.Int64()) * 100

	s.EmitEvent("staked", "", "", map[string]interface{}{
		"staker":       staker,
		"protocol":     protocol,
		"amount":       amount.String(),
		"new_capacity": capacity.String(),
	})

	s.updateState()
	return nil
}

// =============================================================================
// 理赔
// =============================================================================

// SubmitClaim 提交理赔
func (s *InsuranceSimulator) SubmitClaim(policyID, reason, evidence string) (*Claim, error) {
	policy, ok := s.policies[policyID]
	if !ok {
		return nil, fmt.Errorf("保单不存在")
	}

	if !policy.IsActive {
		return nil, fmt.Errorf("保单已过期")
	}

	if policy.ClaimSubmitted {
		return nil, fmt.Errorf("已提交过理赔")
	}

	claimID := fmt.Sprintf("claim-%s-%d", policyID, time.Now().UnixNano())
	claim := &Claim{
		ID:          claimID,
		PolicyID:    policyID,
		Claimant:    policy.Holder,
		Amount:      policy.CoverAmount,
		Reason:      reason,
		Evidence:    evidence,
		Status:      "pending",
		SubmittedAt: time.Now(),
	}

	s.claims[claimID] = claim
	policy.ClaimSubmitted = true

	s.EmitEvent("claim_submitted", "", "", map[string]interface{}{
		"claim_id":  claimID,
		"policy_id": policyID,
		"amount":    policy.CoverAmount.String(),
		"reason":    reason,
	})

	s.updateState()
	return claim, nil
}

// VoteOnClaim 对理赔投票
func (s *InsuranceSimulator) VoteOnClaim(claimID string, voter string, approve bool) error {
	claim, ok := s.claims[claimID]
	if !ok {
		return fmt.Errorf("理赔不存在")
	}

	if claim.Status != "pending" && claim.Status != "voting" {
		return fmt.Errorf("理赔不在投票阶段")
	}

	claim.Status = "voting"

	if approve {
		claim.VotesFor++
	} else {
		claim.VotesAgainst++
	}

	s.EmitEvent("claim_voted", "", "", map[string]interface{}{
		"claim_id":      claimID,
		"voter":         voter,
		"approve":       approve,
		"votes_for":     claim.VotesFor,
		"votes_against": claim.VotesAgainst,
	})

	s.updateState()
	return nil
}

// ResolveClaim 解决理赔
func (s *InsuranceSimulator) ResolveClaim(claimID string) (map[string]interface{}, error) {
	claim, ok := s.claims[claimID]
	if !ok {
		return nil, fmt.Errorf("理赔不存在")
	}

	if claim.Status != "voting" {
		return nil, fmt.Errorf("理赔未完成投票")
	}

	policy := s.policies[claim.PolicyID]
	pool := s.pools[policy.Protocol]

	approved := claim.VotesFor > claim.VotesAgainst

	if approved {
		claim.Status = "approved"
		policy.ClaimApproved = true
		pool.ClaimsPaid.Add(pool.ClaimsPaid, claim.Amount)
		pool.TotalStaked.Sub(pool.TotalStaked, claim.Amount)
	} else {
		claim.Status = "rejected"
	}

	claim.ResolvedAt = time.Now()

	result := map[string]interface{}{
		"claim_id":      claimID,
		"status":        claim.Status,
		"votes_for":     claim.VotesFor,
		"votes_against": claim.VotesAgainst,
		"payout": func() string {
			if approved {
				return claim.Amount.String()
			}
			return "0"
		}(),
	}

	s.EmitEvent("claim_resolved", "", "", result)

	s.updateState()
	return result, nil
}

// GetPoolInfo 获取保险池信息
func (s *InsuranceSimulator) GetPoolInfo(protocol string) map[string]interface{} {
	pool, ok := s.pools[protocol]
	if !ok {
		return nil
	}

	capacity := new(big.Int).Mul(pool.TotalStaked, big.NewInt(int64(s.capacityFactor)))
	availableCapacity := new(big.Int).Sub(capacity, pool.ActiveCover)

	return map[string]interface{}{
		"protocol":           pool.Protocol,
		"total_staked":       pool.TotalStaked.String(),
		"active_cover":       pool.ActiveCover.String(),
		"capacity":           capacity.String(),
		"available_capacity": availableCapacity.String(),
		"capacity_used":      fmt.Sprintf("%.1f%%", pool.CapacityUsed),
		"premium_earned":     pool.PremiumEarned.String(),
		"claims_paid":        pool.ClaimsPaid.String(),
		"staker_count":       pool.StakerCount,
		"risk_score":         pool.RiskScore,
		"min_premium_rate":   fmt.Sprintf("%.2f%%", pool.MinPremiumRate*100),
	}
}

// GetRealWorldCases 获取真实案例
func (s *InsuranceSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"protocol": "Nexus Mutual",
			"event":    "bZx闪电贷攻击",
			"date":     "2020年2月",
			"payout":   "约$36,000",
			"outcome":  "理赔批准",
		},
		{
			"protocol": "Cover Protocol",
			"event":    "Pickle Finance攻击",
			"date":     "2020年11月",
			"payout":   "$19.7M (部分)",
			"outcome":  "理赔批准",
		},
		{
			"protocol": "Nexus Mutual",
			"event":    "Yearn v1 DAI金库攻击",
			"date":     "2021年2月",
			"payout":   "$2.8M",
			"outcome":  "理赔批准",
		},
	}
}

// updateState 更新状态
func (s *InsuranceSimulator) updateState() {
	totalStaked := big.NewInt(0)
	totalActiveCover := big.NewInt(0)

	for _, pool := range s.pools {
		totalStaked.Add(totalStaked, pool.TotalStaked)
		totalActiveCover.Add(totalActiveCover, pool.ActiveCover)
	}

	s.SetGlobalData("pool_count", len(s.pools))
	s.SetGlobalData("policy_count", len(s.policies))
	s.SetGlobalData("claim_count", len(s.claims))
	s.SetGlobalData("total_staked", totalStaked.String())
	s.SetGlobalData("total_active_cover", totalActiveCover.String())

	summary := fmt.Sprintf("当前共有 %d 个保险池、%d 份保单和 %d 条理赔记录。", len(s.pools), len(s.policies), len(s.claims))
	nextHint := "可以继续购买保单、提交理赔或投票处理理赔，观察池子偿付能力如何变化。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"insurance_claims",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"pool_count": len(s.pools), "policy_count": len(s.policies), "claim_count": len(s.claims)},
	)
}

func (s *InsuranceSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "stake_pool":
		protocol := "aave"
		staker := "lp-1"
		amount := big.NewInt(1000)
		if raw, ok := params["protocol"].(string); ok && raw != "" {
			protocol = raw
		}
		if raw, ok := params["staker"].(string); ok && raw != "" {
			staker = raw
		}
		if raw, ok := params["amount"].(float64); ok && raw > 0 {
			amount = big.NewInt(int64(raw))
		}
		if err := s.StakeToPool(staker, protocol, amount); err != nil {
			return nil, err
		}
		return defiActionResult("已向保险池质押资金。", map[string]interface{}{"protocol": protocol, "staker": staker, "amount": amount.String()}, &types.ActionFeedback{
			Summary:     "新的质押已增加保险池偿付能力。",
			NextHint:    "继续购买保单或提交理赔，观察保险池总保障额度如何变化。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"protocol": protocol, "amount": amount.String()},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported insurance action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// InsuranceFactory 保险工厂
type InsuranceFactory struct{}

// Create 创建演示器
func (f *InsuranceFactory) Create() engine.Simulator {
	return NewInsuranceSimulator()
}

// GetDescription 获取描述
func (f *InsuranceFactory) GetDescription() types.Description {
	return NewInsuranceSimulator().GetDescription()
}

// NewInsuranceFactory 创建工厂
func NewInsuranceFactory() *InsuranceFactory {
	return &InsuranceFactory{}
}

var _ engine.SimulatorFactory = (*InsuranceFactory)(nil)
