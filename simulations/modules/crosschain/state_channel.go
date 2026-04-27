package crosschain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 状态通道演示器
// 演示Layer2状态通道的核心机制
//
// 核心概念:
// 1. 状态通道: 链下交易，链上结算
// 2. 双向通道: 双方可以互相转账
// 3. 争议解决: 提交最新状态，挑战期
// 4. 通道网络: 多跳支付路由
//
// 优势:
// - 即时确认
// - 极低手续费
// - 高吞吐量
//
// 参考: Bitcoin Lightning Network, Ethereum Raiden Network
// =============================================================================

// ChannelState 通道状态
type ChannelState string

const (
	ChannelStateOpen     ChannelState = "open"
	ChannelStateActive   ChannelState = "active"
	ChannelStateDisputed ChannelState = "disputed"
	ChannelStateClosing  ChannelState = "closing"
	ChannelStateClosed   ChannelState = "closed"
)

// StateChannel 状态通道
type StateChannel struct {
	ChannelID       string       `json:"channel_id"`
	Participant1    string       `json:"participant1"`
	Participant2    string       `json:"participant2"`
	Balance1        *big.Int     `json:"balance1"`
	Balance2        *big.Int     `json:"balance2"`
	TotalDeposit    *big.Int     `json:"total_deposit"`
	Nonce           uint64       `json:"nonce"`
	State           ChannelState `json:"state"`
	LastUpdate      time.Time    `json:"last_update"`
	DisputeDeadline time.Time    `json:"dispute_deadline"`
	OpenTxHash      string       `json:"open_tx_hash"`
	CloseTxHash     string       `json:"close_tx_hash"`
	CreatedAt       time.Time    `json:"created_at"`
}

// ChannelUpdate 通道更新
type ChannelUpdate struct {
	UpdateID   string    `json:"update_id"`
	ChannelID  string    `json:"channel_id"`
	Nonce      uint64    `json:"nonce"`
	Balance1   *big.Int  `json:"balance1"`
	Balance2   *big.Int  `json:"balance2"`
	Signature1 string    `json:"signature1"`
	Signature2 string    `json:"signature2"`
	Timestamp  time.Time `json:"timestamp"`
}

// PaymentRoute 支付路由
type PaymentRoute struct {
	RouteID  string   `json:"route_id"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Amount   *big.Int `json:"amount"`
	Hops     []string `json:"hops"`
	TotalFee *big.Int `json:"total_fee"`
	Success  bool     `json:"success"`
}

// DisputeRecord 争议记录
type DisputeRecord struct {
	DisputeID       string        `json:"dispute_id"`
	ChannelID       string        `json:"channel_id"`
	Challenger      string        `json:"challenger"`
	SubmittedNonce  uint64        `json:"submitted_nonce"`
	ChallengePeriod time.Duration `json:"challenge_period"`
	Deadline        time.Time     `json:"deadline"`
	Resolved        bool          `json:"resolved"`
	Winner          string        `json:"winner"`
}

// StateChannelSimulator 状态通道演示器
type StateChannelSimulator struct {
	*base.BaseSimulator
	channels        map[string]*StateChannel
	updates         map[string][]*ChannelUpdate
	disputes        map[string]*DisputeRecord
	routes          map[string]*PaymentRoute
	challengePeriod time.Duration
	totalVolume     *big.Int
	totalFees       *big.Int
}

// NewStateChannelSimulator 创建状态通道演示器
func NewStateChannelSimulator() *StateChannelSimulator {
	sim := &StateChannelSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"state_channel",
			"状态通道演示器",
			"演示Layer2状态通道的开启、链下交易、争议解决和关闭流程",
			"crosschain",
			types.ComponentProcess,
		),
		channels:    make(map[string]*StateChannel),
		updates:     make(map[string][]*ChannelUpdate),
		disputes:    make(map[string]*DisputeRecord),
		routes:      make(map[string]*PaymentRoute),
		totalVolume: big.NewInt(0),
		totalFees:   big.NewInt(0),
	}

	sim.AddParam(types.Param{
		Key:         "challenge_period_hours",
		Name:        "挑战期(小时)",
		Description: "争议解决的挑战期时长",
		Type:        types.ParamTypeInt,
		Default:     24,
		Min:         1,
		Max:         168,
	})

	return sim
}

// Init 初始化
func (s *StateChannelSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.challengePeriod = 24 * time.Hour

	if v, ok := config.Params["challenge_period_hours"]; ok {
		if n, ok := v.(float64); ok {
			s.challengePeriod = time.Duration(n) * time.Hour
		}
	}

	s.channels = make(map[string]*StateChannel)
	s.updates = make(map[string][]*ChannelUpdate)
	s.disputes = make(map[string]*DisputeRecord)
	s.routes = make(map[string]*PaymentRoute)
	s.totalVolume = big.NewInt(0)
	s.totalFees = big.NewInt(0)

	s.updateState()
	return nil
}

// =============================================================================
// 状态通道机制解释
// =============================================================================

// ExplainStateChannel 解释状态通道
func (s *StateChannelSimulator) ExplainStateChannel() map[string]interface{} {
	return map[string]interface{}{
		"overview": "状态通道是Layer2扩容方案，允许双方在链下进行多次交易，只在开启和关闭时上链",
		"lifecycle": []map[string]string{
			{"phase": "1. 开启", "action": "双方在链上锁定资金，创建通道"},
			{"phase": "2. 交易", "action": "链下签名更新余额状态"},
			{"phase": "3. 关闭", "action": "提交最终状态到链上"},
			{"phase": "4. 结算", "action": "等待挑战期后分配资金"},
		},
		"key_concepts": []map[string]interface{}{
			{
				"name":        "Nonce递增",
				"description": "每次更新nonce+1，只有最大nonce的状态有效",
				"purpose":     "防止提交旧状态",
			},
			{
				"name":        "双签名",
				"description": "每个状态更新都需要双方签名",
				"purpose":     "确保双方同意",
			},
			{
				"name":        "挑战期",
				"description": fmt.Sprintf("关闭后等待%s可被挑战", s.challengePeriod),
				"purpose":     "给对方提交更新状态的机会",
			},
		},
		"advantages": []string{
			"即时确认: 无需等待区块确认",
			"低成本: 只有开关通道上链",
			"高吞吐: 理论上无限TPS",
			"隐私性: 链下交易不公开",
		},
		"limitations": []string{
			"需要双方在线",
			"资金锁定在通道中",
			"需要监控对手行为",
			"通道容量有限",
		},
		"implementations": []map[string]string{
			{"name": "Lightning Network", "chain": "Bitcoin", "feature": "支付通道网络"},
			{"name": "Raiden Network", "chain": "Ethereum", "feature": "ERC20支持"},
			{"name": "Celer Network", "chain": "Multi-chain", "feature": "状态通道框架"},
		},
	}
}

// =============================================================================
// 通道操作
// =============================================================================

// OpenChannel 开启通道
func (s *StateChannelSimulator) OpenChannel(participant1, participant2 string, deposit1, deposit2 *big.Int) (*StateChannel, error) {
	channelData := fmt.Sprintf("%s-%s-%d", participant1, participant2, time.Now().UnixNano())
	channelHash := sha256.Sum256([]byte(channelData))
	channelID := fmt.Sprintf("channel-%s", hex.EncodeToString(channelHash[:8]))

	totalDeposit := new(big.Int).Add(deposit1, deposit2)

	channel := &StateChannel{
		ChannelID:    channelID,
		Participant1: participant1,
		Participant2: participant2,
		Balance1:     new(big.Int).Set(deposit1),
		Balance2:     new(big.Int).Set(deposit2),
		TotalDeposit: totalDeposit,
		Nonce:        0,
		State:        ChannelStateOpen,
		LastUpdate:   time.Now(),
		OpenTxHash:   fmt.Sprintf("0x%s", hex.EncodeToString(channelHash[:])),
		CreatedAt:    time.Now(),
	}

	s.channels[channelID] = channel
	s.updates[channelID] = make([]*ChannelUpdate, 0)

	s.EmitEvent("channel_opened", "", "", map[string]interface{}{
		"channel_id":   channelID,
		"participant1": participant1,
		"participant2": participant2,
		"deposit1":     deposit1.String(),
		"deposit2":     deposit2.String(),
		"total":        totalDeposit.String(),
	})

	s.updateState()
	return channel, nil
}

// UpdateChannel 更新通道状态(链下)
func (s *StateChannelSimulator) UpdateChannel(channelID string, newBalance1, newBalance2 *big.Int) (*ChannelUpdate, error) {
	channel, ok := s.channels[channelID]
	if !ok {
		return nil, fmt.Errorf("通道不存在: %s", channelID)
	}

	if channel.State != ChannelStateOpen && channel.State != ChannelStateActive {
		return nil, fmt.Errorf("通道状态不允许更新: %s", channel.State)
	}

	totalNew := new(big.Int).Add(newBalance1, newBalance2)
	if totalNew.Cmp(channel.TotalDeposit) != 0 {
		return nil, fmt.Errorf("余额总和不等于存款: %s != %s", totalNew.String(), channel.TotalDeposit.String())
	}

	channel.Nonce++
	channel.Balance1 = new(big.Int).Set(newBalance1)
	channel.Balance2 = new(big.Int).Set(newBalance2)
	channel.State = ChannelStateActive
	channel.LastUpdate = time.Now()

	updateData := fmt.Sprintf("%s-%d-%d", channelID, channel.Nonce, time.Now().UnixNano())
	updateHash := sha256.Sum256([]byte(updateData))

	update := &ChannelUpdate{
		UpdateID:   fmt.Sprintf("update-%s", hex.EncodeToString(updateHash[:8])),
		ChannelID:  channelID,
		Nonce:      channel.Nonce,
		Balance1:   new(big.Int).Set(newBalance1),
		Balance2:   new(big.Int).Set(newBalance2),
		Signature1: hex.EncodeToString(updateHash[:16]),
		Signature2: hex.EncodeToString(updateHash[16:]),
		Timestamp:  time.Now(),
	}

	s.updates[channelID] = append(s.updates[channelID], update)

	transferAmount := new(big.Int).Abs(new(big.Int).Sub(newBalance1, channel.Balance1))
	s.totalVolume.Add(s.totalVolume, transferAmount)

	s.EmitEvent("channel_updated", "", "", map[string]interface{}{
		"channel_id": channelID,
		"nonce":      channel.Nonce,
		"balance1":   newBalance1.String(),
		"balance2":   newBalance2.String(),
		"offchain":   true,
	})

	s.updateState()
	return update, nil
}

// InitiateClose 发起关闭通道
func (s *StateChannelSimulator) InitiateClose(channelID, initiator string) error {
	channel, ok := s.channels[channelID]
	if !ok {
		return fmt.Errorf("通道不存在: %s", channelID)
	}

	if initiator != channel.Participant1 && initiator != channel.Participant2 {
		return fmt.Errorf("非通道参与者: %s", initiator)
	}

	channel.State = ChannelStateClosing
	channel.DisputeDeadline = time.Now().Add(s.challengePeriod)

	s.EmitEvent("channel_closing", "", "", map[string]interface{}{
		"channel_id":       channelID,
		"initiator":        initiator,
		"submitted_nonce":  channel.Nonce,
		"balance1":         channel.Balance1.String(),
		"balance2":         channel.Balance2.String(),
		"dispute_deadline": channel.DisputeDeadline,
	})

	s.updateState()
	return nil
}

// ChallengeClose 挑战关闭(提交更新的状态)
func (s *StateChannelSimulator) ChallengeClose(channelID, challenger string, update *ChannelUpdate) error {
	channel, ok := s.channels[channelID]
	if !ok {
		return fmt.Errorf("通道不存在: %s", channelID)
	}

	if channel.State != ChannelStateClosing {
		return fmt.Errorf("通道未在关闭中: %s", channel.State)
	}

	if time.Now().After(channel.DisputeDeadline) {
		return fmt.Errorf("挑战期已过")
	}

	if update.Nonce <= channel.Nonce {
		return fmt.Errorf("提交的nonce不是更新的: %d <= %d", update.Nonce, channel.Nonce)
	}

	channel.Nonce = update.Nonce
	channel.Balance1 = new(big.Int).Set(update.Balance1)
	channel.Balance2 = new(big.Int).Set(update.Balance2)
	channel.State = ChannelStateDisputed
	channel.DisputeDeadline = time.Now().Add(s.challengePeriod)

	dispute := &DisputeRecord{
		DisputeID:       fmt.Sprintf("dispute-%s-%d", channelID, time.Now().UnixNano()),
		ChannelID:       channelID,
		Challenger:      challenger,
		SubmittedNonce:  update.Nonce,
		ChallengePeriod: s.challengePeriod,
		Deadline:        channel.DisputeDeadline,
		Resolved:        false,
	}
	s.disputes[channelID] = dispute

	s.EmitEvent("channel_challenged", "", "", map[string]interface{}{
		"channel_id":   channelID,
		"challenger":   challenger,
		"new_nonce":    update.Nonce,
		"new_balance1": update.Balance1.String(),
		"new_balance2": update.Balance2.String(),
		"new_deadline": channel.DisputeDeadline,
	})

	s.updateState()
	return nil
}

// FinalizeClose 完成关闭
func (s *StateChannelSimulator) FinalizeClose(channelID string) error {
	channel, ok := s.channels[channelID]
	if !ok {
		return fmt.Errorf("通道不存在: %s", channelID)
	}

	if channel.State != ChannelStateClosing && channel.State != ChannelStateDisputed {
		return fmt.Errorf("通道未在关闭中: %s", channel.State)
	}

	channel.State = ChannelStateClosed

	closeData := fmt.Sprintf("%s-close-%d", channelID, time.Now().UnixNano())
	closeHash := sha256.Sum256([]byte(closeData))
	channel.CloseTxHash = fmt.Sprintf("0x%s", hex.EncodeToString(closeHash[:]))

	if dispute, ok := s.disputes[channelID]; ok {
		dispute.Resolved = true
		dispute.Winner = "honest_party"
	}

	s.EmitEvent("channel_closed", "", "", map[string]interface{}{
		"channel_id":     channelID,
		"final_nonce":    channel.Nonce,
		"final_balance1": channel.Balance1.String(),
		"final_balance2": channel.Balance2.String(),
		"close_tx_hash":  channel.CloseTxHash,
	})

	s.updateState()
	return nil
}

// SimulateChannelLifecycle 模拟完整生命周期
func (s *StateChannelSimulator) SimulateChannelLifecycle() map[string]interface{} {
	steps := make([]map[string]interface{}, 0)

	deposit1 := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	deposit2 := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	channel, _ := s.OpenChannel("Alice", "Bob", deposit1, deposit2)
	steps = append(steps, map[string]interface{}{
		"step": 1, "action": "链上开启通道",
		"channel_id": channel.ChannelID,
		"deposits":   "Alice: 10 ETH, Bob: 10 ETH",
		"onchain":    true,
	})

	for i := 1; i <= 5; i++ {
		newBal1 := new(big.Int).Sub(deposit1, big.NewInt(int64(i)*1e18))
		newBal2 := new(big.Int).Add(deposit2, big.NewInt(int64(i)*1e18))
		s.UpdateChannel(channel.ChannelID, newBal1, newBal2)
	}
	steps = append(steps, map[string]interface{}{
		"step": 2, "action": "链下交易5次",
		"nonce":    channel.Nonce,
		"balance":  fmt.Sprintf("Alice: %s, Bob: %s", channel.Balance1.String(), channel.Balance2.String()),
		"offchain": true,
		"instant":  true,
	})

	s.InitiateClose(channel.ChannelID, "Alice")
	steps = append(steps, map[string]interface{}{
		"step": 3, "action": "Alice发起关闭通道",
		"submitted_state":  fmt.Sprintf("nonce=%d", channel.Nonce),
		"challenge_period": s.challengePeriod.String(),
		"onchain":          true,
	})

	steps = append(steps, map[string]interface{}{
		"step": 4, "action": "等待挑战期(无挑战)",
		"deadline": channel.DisputeDeadline,
	})

	s.FinalizeClose(channel.ChannelID)
	steps = append(steps, map[string]interface{}{
		"step": 5, "action": "完成关闭，资金分配",
		"final_balance": fmt.Sprintf("Alice: %s, Bob: %s", channel.Balance1.String(), channel.Balance2.String()),
		"onchain":       true,
	})

	return map[string]interface{}{
		"channel_id":         channel.ChannelID,
		"total_transactions": channel.Nonce,
		"onchain_txs":        2,
		"steps":              steps,
		"summary": map[string]interface{}{
			"offchain_txs": channel.Nonce,
			"cost_saved":   fmt.Sprintf("%d个链上交易费", channel.Nonce-2),
		},
	}
}

// GetChannelInfo 获取通道信息
func (s *StateChannelSimulator) GetChannelInfo(channelID string) map[string]interface{} {
	channel, ok := s.channels[channelID]
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"channel_id":    channel.ChannelID,
		"participant1":  channel.Participant1,
		"participant2":  channel.Participant2,
		"balance1":      channel.Balance1.String(),
		"balance2":      channel.Balance2.String(),
		"total_deposit": channel.TotalDeposit.String(),
		"nonce":         channel.Nonce,
		"state":         string(channel.State),
		"update_count":  len(s.updates[channelID]),
		"created_at":    channel.CreatedAt,
		"last_update":   channel.LastUpdate,
	}
}

// GetStatistics 获取统计
func (s *StateChannelSimulator) GetStatistics() map[string]interface{} {
	openChannels := 0
	closedChannels := 0
	totalUpdates := 0

	for _, channel := range s.channels {
		if channel.State == ChannelStateClosed {
			closedChannels++
		} else {
			openChannels++
		}
	}

	for _, updates := range s.updates {
		totalUpdates += len(updates)
	}

	return map[string]interface{}{
		"total_channels":   len(s.channels),
		"open_channels":    openChannels,
		"closed_channels":  closedChannels,
		"total_updates":    totalUpdates,
		"total_disputes":   len(s.disputes),
		"total_volume":     s.totalVolume.String(),
		"challenge_period": s.challengePeriod.String(),
	}
}

// updateState 更新状态
func (s *StateChannelSimulator) updateState() {
	s.SetGlobalData("channel_count", len(s.channels))
	s.SetGlobalData("total_volume", s.totalVolume.String())

	setCrosschainTeachingState(
		s.BaseSimulator,
		"crosschain",
		"state_channel",
		"当前可以开启状态通道，再观察离链更新和关闭结算流程。",
		"先开启一条通道，随后执行一次链下更新或关闭流程，比较不同阶段的成本与速度。",
		0,
		map[string]interface{}{
			"channel_count": len(s.channels),
			"total_volume":  s.totalVolume.String(),
		},
	)
}

func (s *StateChannelSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "open_channel":
		channel, err := s.OpenChannel("alice", "bob", big.NewInt(10), big.NewInt(10))
		if err != nil {
			return nil, err
		}
		return crosschainActionResult(
			"已开启状态通道",
			map[string]interface{}{"channel_id": channel.ChannelID},
			&types.ActionFeedback{
				Summary:     "状态通道已经建立，可继续进行离链更新或发起关闭。",
				NextHint:    "执行一次通道更新，观察状态如何在链下推进。",
				EffectScope: "crosschain",
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported state channel action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// StateChannelFactory 状态通道工厂
type StateChannelFactory struct{}

func (f *StateChannelFactory) Create() engine.Simulator { return NewStateChannelSimulator() }
func (f *StateChannelFactory) GetDescription() types.Description {
	return NewStateChannelSimulator().GetDescription()
}
func NewStateChannelFactory() *StateChannelFactory { return &StateChannelFactory{} }

var _ engine.SimulatorFactory = (*StateChannelFactory)(nil)
