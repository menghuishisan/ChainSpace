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
// veToken演示器
// =============================================================================

// VeLock 锁定记录
type VeLock struct {
	ID         string    `json:"id"`
	Owner      string    `json:"owner"`
	Amount     *big.Int  `json:"amount"`      // 锁定的代币数量
	UnlockTime time.Time `json:"unlock_time"` // 解锁时间
	VePower    *big.Int  `json:"ve_power"`    // 投票权重
	CreatedAt  time.Time `json:"created_at"`
}

// Gauge Gauge池
type Gauge struct {
	ID          string   `json:"id"`
	Pool        string   `json:"pool"`
	Weight      *big.Int `json:"weight"` // 投票权重
	Votes       *big.Int `json:"votes"`  // 收到的投票
	RewardRate  float64  `json:"reward_rate"`
	TotalStaked *big.Int `json:"total_staked"`
}

// GaugeVote Gauge投票
type GaugeVote struct {
	Voter   string   `json:"voter"`
	GaugeID string   `json:"gauge_id"`
	Weight  uint64   `json:"weight"` // 0-10000 (万分比)
	VePower *big.Int `json:"ve_power"`
}

// VeTokenSimulator veToken演示器
// 演示投票托管代币机制:
//
// 1. 锁定机制
//   - 锁定代币获得veToken
//   - 锁定时间越长，veToken越多
//   - veToken随时间线性衰减
//
// 2. 投票权重
//   - veToken = amount × (lock_time / max_lock_time)
//   - 最长锁定4年
//
// 3. Gauge投票
//   - veToken持有者投票决定奖励分配
//   - 贿赂(Bribes)激励投票
//
// 参考: Curve veCRV, Balancer veBAL
type VeTokenSimulator struct {
	*base.BaseSimulator
	tokenSymbol  string
	veSymbol     string
	maxLockTime  time.Duration // 最长锁定时间
	locks        map[string]*VeLock
	gauges       map[string]*Gauge
	gaugeVotes   map[string][]*GaugeVote // voter -> votes
	totalVePower *big.Int
	totalLocked  *big.Int
	currentEpoch uint64
}

// NewVeTokenSimulator 创建veToken演示器
func NewVeTokenSimulator() *VeTokenSimulator {
	sim := &VeTokenSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"ve_token",
			"veToken演示器",
			"演示投票托管代币的锁定、投票权重、Gauge投票等机制",
			"defi",
			types.ComponentDeFi,
		),
		locks:      make(map[string]*VeLock),
		gauges:     make(map[string]*Gauge),
		gaugeVotes: make(map[string][]*GaugeVote),
	}

	sim.AddParam(types.Param{
		Key:         "max_lock_years",
		Name:        "最长锁定年数",
		Description: "代币最长锁定时间(年)",
		Type:        types.ParamTypeInt,
		Default:     4,
		Min:         1,
		Max:         10,
	})

	return sim
}

// Init 初始化
func (s *VeTokenSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	maxLockYears := 4
	if v, ok := config.Params["max_lock_years"]; ok {
		if n, ok := v.(float64); ok {
			maxLockYears = int(n)
		}
	}

	s.tokenSymbol = "CRV"
	s.veSymbol = "veCRV"
	s.maxLockTime = time.Duration(maxLockYears) * 365 * 24 * time.Hour
	s.locks = make(map[string]*VeLock)
	s.gauges = make(map[string]*Gauge)
	s.gaugeVotes = make(map[string][]*GaugeVote)
	s.totalVePower = big.NewInt(0)
	s.totalLocked = big.NewInt(0)
	s.currentEpoch = 0

	// 初始化示例Gauge
	s.gauges = map[string]*Gauge{
		"gauge-3pool":     {ID: "gauge-3pool", Pool: "3pool (DAI/USDC/USDT)", Weight: big.NewInt(0), Votes: big.NewInt(0), TotalStaked: big.NewInt(0)},
		"gauge-steth":     {ID: "gauge-steth", Pool: "stETH/ETH", Weight: big.NewInt(0), Votes: big.NewInt(0), TotalStaked: big.NewInt(0)},
		"gauge-fraxusdc":  {ID: "gauge-fraxusdc", Pool: "FRAX/USDC", Weight: big.NewInt(0), Votes: big.NewInt(0), TotalStaked: big.NewInt(0)},
		"gauge-tricrypto": {ID: "gauge-tricrypto", Pool: "USDT/WBTC/ETH", Weight: big.NewInt(0), Votes: big.NewInt(0), TotalStaked: big.NewInt(0)},
	}

	s.updateState()
	return nil
}

// =============================================================================
// veToken机制解释
// =============================================================================

// ExplainVeToken 解释veToken
func (s *VeTokenSimulator) ExplainVeToken() map[string]interface{} {
	return map[string]interface{}{
		"name":         "投票托管代币 (Vote-Escrowed Token)",
		"core_concept": "用户锁定代币获得veToken，veToken代表治理权和收益权",
		"formula":      "veToken = lockedAmount × (lockTime / maxLockTime)",
		"properties": []map[string]string{
			{"property": "不可转让", "description": "veToken绑定地址，不能交易"},
			{"property": "时间衰减", "description": "veToken随时间线性减少"},
			{"property": "可续期", "description": "可以延长锁定时间增加veToken"},
		},
		"example": map[string]interface{}{
			"locked_amount": "1000 CRV",
			"lock_time":     "4年",
			"ve_power":      "1000 veCRV (最大)",
		},
		"decay": []map[string]interface{}{
			{"time_remaining": "4年", "ve_power": "1000 veCRV"},
			{"time_remaining": "3年", "ve_power": "750 veCRV"},
			{"time_remaining": "2年", "ve_power": "500 veCRV"},
			{"time_remaining": "1年", "ve_power": "250 veCRV"},
			{"time_remaining": "0", "ve_power": "0 veCRV (可解锁)"},
		},
		"benefits": []string{
			"协议费用分成",
			"治理投票权",
			"Gauge投票决定奖励分配",
			"Boost LP挖矿收益",
		},
	}
}

// =============================================================================
// 锁定操作
// =============================================================================

// CreateLock 创建锁定
func (s *VeTokenSimulator) CreateLock(owner string, amount *big.Int, lockDuration time.Duration) (*VeLock, error) {
	if lockDuration > s.maxLockTime {
		lockDuration = s.maxLockTime
	}
	if lockDuration < 7*24*time.Hour {
		return nil, fmt.Errorf("最短锁定时间为1周")
	}

	// 计算veToken权重
	vePower := s.calculateVePower(amount, lockDuration)

	lockID := fmt.Sprintf("lock-%s-%d", owner, time.Now().UnixNano())
	lock := &VeLock{
		ID:         lockID,
		Owner:      owner,
		Amount:     amount,
		UnlockTime: time.Now().Add(lockDuration),
		VePower:    vePower,
		CreatedAt:  time.Now(),
	}

	s.locks[lockID] = lock
	s.totalLocked.Add(s.totalLocked, amount)
	s.totalVePower.Add(s.totalVePower, vePower)

	s.EmitEvent("lock_created", "", "", map[string]interface{}{
		"lock_id":       lockID,
		"owner":         owner,
		"amount":        amount.String(),
		"lock_duration": lockDuration.String(),
		"ve_power":      vePower.String(),
		"unlock_time":   lock.UnlockTime,
	})

	s.updateState()
	return lock, nil
}

// IncreaseLockAmount 增加锁定数量
func (s *VeTokenSimulator) IncreaseLockAmount(lockID string, additionalAmount *big.Int) error {
	lock, ok := s.locks[lockID]
	if !ok {
		return fmt.Errorf("锁定不存在")
	}

	if time.Now().After(lock.UnlockTime) {
		return fmt.Errorf("锁定已过期")
	}

	// 重新计算veToken
	remainingDuration := time.Until(lock.UnlockTime)
	additionalVePower := s.calculateVePower(additionalAmount, remainingDuration)

	lock.Amount.Add(lock.Amount, additionalAmount)
	lock.VePower.Add(lock.VePower, additionalVePower)
	s.totalLocked.Add(s.totalLocked, additionalAmount)
	s.totalVePower.Add(s.totalVePower, additionalVePower)

	s.EmitEvent("lock_amount_increased", "", "", map[string]interface{}{
		"lock_id":           lockID,
		"additional_amount": additionalAmount.String(),
		"new_ve_power":      lock.VePower.String(),
	})

	s.updateState()
	return nil
}

// ExtendLockTime 延长锁定时间
func (s *VeTokenSimulator) ExtendLockTime(lockID string, newUnlockTime time.Time) error {
	lock, ok := s.locks[lockID]
	if !ok {
		return fmt.Errorf("锁定不存在")
	}

	if newUnlockTime.Before(lock.UnlockTime) {
		return fmt.Errorf("新解锁时间必须晚于当前解锁时间")
	}

	maxUnlock := time.Now().Add(s.maxLockTime)
	if newUnlockTime.After(maxUnlock) {
		newUnlockTime = maxUnlock
	}

	// 重新计算veToken
	newDuration := time.Until(newUnlockTime)
	newVePower := s.calculateVePower(lock.Amount, newDuration)
	vePowerIncrease := new(big.Int).Sub(newVePower, lock.VePower)

	lock.UnlockTime = newUnlockTime
	lock.VePower = newVePower
	s.totalVePower.Add(s.totalVePower, vePowerIncrease)

	s.EmitEvent("lock_time_extended", "", "", map[string]interface{}{
		"lock_id":         lockID,
		"new_unlock_time": newUnlockTime,
		"new_ve_power":    newVePower.String(),
	})

	s.updateState()
	return nil
}

// Withdraw 提取(锁定到期后)
func (s *VeTokenSimulator) Withdraw(lockID string) (*big.Int, error) {
	lock, ok := s.locks[lockID]
	if !ok {
		return nil, fmt.Errorf("锁定不存在")
	}

	if time.Now().Before(lock.UnlockTime) {
		return nil, fmt.Errorf("锁定尚未到期，剩余 %v", time.Until(lock.UnlockTime))
	}

	amount := new(big.Int).Set(lock.Amount)

	s.totalLocked.Sub(s.totalLocked, lock.Amount)
	s.totalVePower.Sub(s.totalVePower, lock.VePower)
	delete(s.locks, lockID)

	// 清除投票
	delete(s.gaugeVotes, lock.Owner)

	s.EmitEvent("withdrawn", "", "", map[string]interface{}{
		"lock_id": lockID,
		"amount":  amount.String(),
	})

	s.updateState()
	return amount, nil
}

// calculateVePower 计算veToken权重
func (s *VeTokenSimulator) calculateVePower(amount *big.Int, duration time.Duration) *big.Int {
	// vePower = amount × (duration / maxLockTime)
	ratio := float64(duration) / float64(s.maxLockTime)
	if ratio > 1 {
		ratio = 1
	}

	power := new(big.Float).SetInt(amount)
	power.Mul(power, big.NewFloat(ratio))

	result := new(big.Int)
	power.Int(result)
	return result
}

// =============================================================================
// Gauge投票
// =============================================================================

// VoteForGauge 为Gauge投票
func (s *VeTokenSimulator) VoteForGauge(voter string, votes map[string]uint64) error {
	// 验证投票权
	vePower := s.getVePower(voter)
	if vePower.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("没有投票权")
	}

	// 验证权重总和
	var totalWeight uint64
	for _, weight := range votes {
		totalWeight += weight
	}
	if totalWeight > 10000 {
		return fmt.Errorf("权重总和超过100%%")
	}

	// 清除旧投票
	oldVotes := s.gaugeVotes[voter]
	for _, vote := range oldVotes {
		if gauge, ok := s.gauges[vote.GaugeID]; ok {
			gauge.Votes.Sub(gauge.Votes, vote.VePower)
		}
	}

	// 记录新投票
	newVotes := make([]*GaugeVote, 0)
	for gaugeID, weight := range votes {
		if weight == 0 {
			continue
		}

		gauge, ok := s.gauges[gaugeID]
		if !ok {
			continue
		}

		votePower := new(big.Int).Mul(vePower, big.NewInt(int64(weight)))
		votePower.Div(votePower, big.NewInt(10000))

		vote := &GaugeVote{
			Voter:   voter,
			GaugeID: gaugeID,
			Weight:  weight,
			VePower: votePower,
		}
		newVotes = append(newVotes, vote)

		gauge.Votes.Add(gauge.Votes, votePower)
	}

	s.gaugeVotes[voter] = newVotes

	s.EmitEvent("gauge_voted", "", "", map[string]interface{}{
		"voter":    voter,
		"ve_power": vePower.String(),
		"votes":    votes,
	})

	s.updateGaugeWeights()
	s.updateState()
	return nil
}

// updateGaugeWeights 更新Gauge权重
func (s *VeTokenSimulator) updateGaugeWeights() {
	totalVotes := big.NewInt(0)
	for _, gauge := range s.gauges {
		totalVotes.Add(totalVotes, gauge.Votes)
	}

	if totalVotes.Cmp(big.NewInt(0)) == 0 {
		return
	}

	for _, gauge := range s.gauges {
		// 权重 = Gauge投票 / 总投票
		weight := new(big.Int).Mul(gauge.Votes, big.NewInt(10000))
		weight.Div(weight, totalVotes)
		gauge.Weight = weight
	}
}

// getVePower 获取用户veToken权重
func (s *VeTokenSimulator) getVePower(owner string) *big.Int {
	totalPower := big.NewInt(0)
	for _, lock := range s.locks {
		if lock.Owner == owner && time.Now().Before(lock.UnlockTime) {
			totalPower.Add(totalPower, lock.VePower)
		}
	}
	return totalPower
}

// ExplainGaugeVoting 解释Gauge投票
func (s *VeTokenSimulator) ExplainGaugeVoting() map[string]interface{} {
	return map[string]interface{}{
		"purpose": "决定协议奖励如何分配给不同的流动性池",
		"mechanism": []string{
			"veToken持有者每周可以投票",
			"将投票权分配给不同的Gauge",
			"Gauge获得的投票越多，奖励份额越大",
		},
		"bribes": map[string]interface{}{
			"concept":     "项目方贿赂veToken持有者投票给自己的池",
			"platforms":   []string{"Votium", "Convex", "Hidden Hand"},
			"typical_apr": "10-50% APR 贿赂收益",
		},
		"vote_locking": "投票后锁定10天，防止投票后立即卖出",
	}
}

// GetGaugeInfo 获取Gauge信息
func (s *VeTokenSimulator) GetGaugeInfo() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, gauge := range s.gauges {
		result = append(result, map[string]interface{}{
			"id":           gauge.ID,
			"pool":         gauge.Pool,
			"votes":        gauge.Votes.String(),
			"weight":       fmt.Sprintf("%.2f%%", float64(gauge.Weight.Int64())/100),
			"total_staked": gauge.TotalStaked.String(),
		})
	}
	return result
}

// SimulateDecay 模拟veToken衰减
func (s *VeTokenSimulator) SimulateDecay(lockID string, months int) []map[string]interface{} {
	lock, ok := s.locks[lockID]
	if !ok {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	now := time.Now()

	for i := 0; i <= months; i++ {
		futureTime := now.Add(time.Duration(i) * 30 * 24 * time.Hour)
		if futureTime.After(lock.UnlockTime) {
			result = append(result, map[string]interface{}{
				"month":    i,
				"ve_power": "0 (已解锁)",
				"decay":    "100%",
			})
			break
		}

		remaining := lock.UnlockTime.Sub(futureTime)
		power := s.calculateVePower(lock.Amount, remaining)
		decay := 1 - float64(power.Int64())/float64(lock.VePower.Int64())

		result = append(result, map[string]interface{}{
			"month":    i,
			"ve_power": power.String(),
			"decay":    fmt.Sprintf("%.1f%%", decay*100),
		})
	}

	return result
}

// updateState 更新状态
func (s *VeTokenSimulator) updateState() {
	s.SetGlobalData("total_locked", s.totalLocked.String())
	s.SetGlobalData("total_ve_power", s.totalVePower.String())
	s.SetGlobalData("lock_count", len(s.locks))
	s.SetGlobalData("gauge_count", len(s.gauges))

	summary := fmt.Sprintf("当前共有 %d 个锁仓和 %d 个 Gauge。", len(s.locks), len(s.gauges))
	nextHint := "可以继续锁仓或对 Gauge 投票，观察 ve 权重如何影响激励分配。"
	setDeFiTeachingState(
		s.BaseSimulator,
		"defi",
		"ve_governance",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"lock_count": len(s.locks), "gauge_count": len(s.gauges)},
	)
}

func (s *VeTokenSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "lock_tokens":
		owner := "alice"
		amount := big.NewInt(100)
		weeks := 52
		if raw, ok := params["owner"].(string); ok && raw != "" {
			owner = raw
		}
		if raw, ok := params["amount"].(float64); ok && raw > 0 {
			amount = big.NewInt(int64(raw))
		}
		if raw, ok := params["weeks"].(float64); ok && raw > 0 {
			weeks = int(raw)
		}
		lock, err := s.CreateLock(owner, amount, time.Duration(weeks)*7*24*time.Hour)
		if err != nil {
			return nil, err
		}
		return defiActionResult("已创建一个 veToken 锁仓。", map[string]interface{}{"lock_id": lock.ID}, &types.ActionFeedback{
			Summary:     "新的锁仓已建立，对应 ve 权重已经生成。",
			NextHint:    "继续进行 Gauge 投票，观察 ve 权重如何决定奖励流向。",
			EffectScope: "defi",
			ResultState: map[string]interface{}{"lock_id": lock.ID, "lock_count": len(s.locks)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported ve token action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// VeTokenFactory veToken工厂
type VeTokenFactory struct{}

// Create 创建演示器
func (f *VeTokenFactory) Create() engine.Simulator {
	return NewVeTokenSimulator()
}

// GetDescription 获取描述
func (f *VeTokenFactory) GetDescription() types.Description {
	return NewVeTokenSimulator().GetDescription()
}

// NewVeTokenFactory 创建工厂
func NewVeTokenFactory() *VeTokenFactory {
	return &VeTokenFactory{}
}

var _ engine.SimulatorFactory = (*VeTokenFactory)(nil)
