package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// 女巫攻击演示器
// =============================================================================

// SybilIdentity 女巫身份
type SybilIdentity struct {
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	NodeID     string    `json:"node_id"`    // 伪造的节点ID
	Controller string    `json:"controller"` // 控制者
	CreatedAt  time.Time `json:"created_at"`
	IsActive   bool      `json:"is_active"`
}

// SybilAttackState 女巫攻击状态
type SybilAttackState struct {
	AttackerID         string           `json:"attacker_id"`
	SybilIdentities    []*SybilIdentity `json:"sybil_identities"`
	TargetRegion       string           `json:"target_region"`       // DHT中的目标区域
	NetworkPenetration float64          `json:"network_penetration"` // 网络渗透率
	Phase              string           `json:"phase"`
}

// SybilAttackSimulator 女巫攻击演示器
// 演示女巫攻击的原理和影响:
//
// 攻击原理:
// 攻击者创建大量虚假身份(节点)，
// 以低成本获得对网络的不成比例影响力
//
// 在区块链中的应用:
// 1. DHT污染 - 在Kademlia中占据特定ID空间
// 2. 投票操纵 - 在PoS中伪造大量验证者
// 3. 声誉系统攻击 - 伪造评价
// 4. 日蚀攻击辅助 - 提供大量恶意节点
//
// 防御机制:
// 1. 资源证明 - PoW/PoS
// 2. 身份验证 - KYC/信任网络
// 3. 成本机制 - 质押要求
type SybilAttackSimulator struct {
	*base.BaseSimulator
	honestNodes map[string]bool
	sybilNodes  map[string]*SybilIdentity
	attackState *SybilAttackState
}

// NewSybilAttackSimulator 创建女巫攻击演示器
func NewSybilAttackSimulator() *SybilAttackSimulator {
	sim := &SybilAttackSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"sybil_attack",
			"女巫攻击演示器",
			"演示女巫攻击如何通过创建虚假身份影响网络",
			"network",
			types.ComponentAttack,
		),
		honestNodes: make(map[string]bool),
		sybilNodes:  make(map[string]*SybilIdentity),
	}

	sim.AddParam(types.Param{
		Key:         "honest_nodes",
		Name:        "诚实节点数",
		Description: "网络中的诚实节点数量",
		Type:        types.ParamTypeInt,
		Default:     100,
		Min:         10,
		Max:         1000,
	})

	sim.AddParam(types.Param{
		Key:         "sybil_ratio",
		Name:        "女巫比例",
		Description: "女巫节点相对于诚实节点的比例",
		Type:        types.ParamTypeFloat,
		Default:     2.0,
		Min:         0.1,
		Max:         10.0,
	})

	return sim
}

// Init 初始化
func (s *SybilAttackSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	honestNodes := 100
	sybilRatio := 2.0

	if v, ok := config.Params["honest_nodes"]; ok {
		if n, ok := v.(float64); ok {
			honestNodes = int(n)
		}
	}
	if v, ok := config.Params["sybil_ratio"]; ok {
		if n, ok := v.(float64); ok {
			sybilRatio = n
		}
	}

	s.initializeNetwork(honestNodes, int(float64(honestNodes)*sybilRatio))
	s.updateState()
	return nil
}

// initializeNetwork 初始化网络
func (s *SybilAttackSimulator) initializeNetwork(honestCount, sybilCount int) {
	s.honestNodes = make(map[string]bool)
	s.sybilNodes = make(map[string]*SybilIdentity)

	// 创建诚实节点
	for i := 0; i < honestCount; i++ {
		nodeID := fmt.Sprintf("honest-%d", i)
		s.honestNodes[nodeID] = true
	}

	// 创建女巫节点
	for i := 0; i < sybilCount; i++ {
		sybilID := fmt.Sprintf("sybil-%d", i)
		s.sybilNodes[sybilID] = &SybilIdentity{
			ID:         sybilID,
			Address:    fmt.Sprintf("10.%d.%d.%d:30303", rand.Intn(256), rand.Intn(256), rand.Intn(256)),
			NodeID:     fmt.Sprintf("0x%x", rand.Int63()),
			Controller: "attacker-0",
			CreatedAt:  time.Now(),
			IsActive:   true,
		}
	}
}

// ExplainAttack 解释女巫攻击
func (s *SybilAttackSimulator) ExplainAttack() map[string]interface{} {
	return map[string]interface{}{
		"name":        "女巫攻击 (Sybil Attack)",
		"origin":      "来自小说《Sybil》中多重人格的主角",
		"definition":  "攻击者创建大量虚假身份以获得对系统的不成比例影响",
		"key_insight": "在无身份验证的系统中，创建新身份几乎没有成本",
		"attack_scenarios": []map[string]string{
			{"scenario": "P2P网络", "impact": "占据路由表，控制信息流"},
			{"scenario": "投票系统", "impact": "用虚假身份操纵投票结果"},
			{"scenario": "声誉系统", "impact": "伪造评价，提升/降低声誉"},
			{"scenario": "DHT", "impact": "占据特定ID空间，控制数据存储"},
			{"scenario": "PoS共识", "impact": "如果质押要求低，可伪造大量验证者"},
		},
		"cost_analysis": map[string]interface{}{
			"without_defense": "创建身份成本接近0",
			"with_pow":        "每个身份需要算力成本",
			"with_pos":        "每个身份需要质押成本",
			"with_kyc":        "每个身份需要真实身份",
		},
	}
}

// SimulateDHTAttack 模拟DHT女巫攻击
func (s *SybilAttackSimulator) SimulateDHTAttack(targetKeyHash string) map[string]interface{} {
	// 模拟攻击者创建大量接近目标Key的节点ID
	s.attackState = &SybilAttackState{
		AttackerID:      "attacker-0",
		SybilIdentities: make([]*SybilIdentity, 0),
		TargetRegion:    targetKeyHash,
		Phase:           "preparation",
	}

	// 生成接近目标的女巫节点
	sybilCount := len(s.sybilNodes) / 10 // 使用10%的女巫节点
	for i := 0; i < sybilCount; i++ {
		// 模拟生成接近目标hash的节点ID
		sybil := &SybilIdentity{
			ID:         fmt.Sprintf("sybil-targeted-%d", i),
			NodeID:     fmt.Sprintf("%s%04x", targetKeyHash[:len(targetKeyHash)-4], i),
			Controller: "attacker-0",
			CreatedAt:  time.Now(),
			IsActive:   true,
		}
		s.attackState.SybilIdentities = append(s.attackState.SybilIdentities, sybil)
	}

	s.attackState.Phase = "execution"
	s.attackState.NetworkPenetration = float64(len(s.sybilNodes)) / float64(len(s.honestNodes)+len(s.sybilNodes)) * 100

	s.EmitEvent("sybil_dht_attack", "", "", map[string]interface{}{
		"target_key":  targetKeyHash[:16] + "...",
		"sybil_nodes": len(s.attackState.SybilIdentities),
		"penetration": fmt.Sprintf("%.1f%%", s.attackState.NetworkPenetration),
	})

	return map[string]interface{}{
		"attack":      "DHT女巫攻击",
		"target":      targetKeyHash,
		"sybil_nodes": len(s.attackState.SybilIdentities),
		"attack_flow": []string{
			"1. 分析目标Key的哈希值",
			"2. 生成大量接近目标哈希的节点ID",
			"3. 将女巫节点加入DHT网络",
			"4. 女巫节点成为目标Key的最近节点",
			"5. 控制对该Key的查询和存储",
		},
		"impact": []string{
			"阻止对目标数据的访问",
			"返回虚假数据",
			"审查特定内容",
		},
	}
}

// SimulateVotingAttack 模拟投票女巫攻击
func (s *SybilAttackSimulator) SimulateVotingAttack(proposalID string) map[string]interface{} {
	totalVotes := len(s.honestNodes) + len(s.sybilNodes)
	sybilVotes := len(s.sybilNodes)
	honestVotes := len(s.honestNodes)

	// 假设诚实节点平均分布投票
	honestYes := honestVotes / 2
	honestNo := honestVotes - honestYes

	// 女巫节点全部投yes
	sybilVote := "yes"

	result := map[string]interface{}{
		"proposal":    proposalID,
		"total_votes": totalVotes,
		"honest_votes": map[string]int{
			"yes": honestYes,
			"no":  honestNo,
		},
		"sybil_votes": map[string]int{
			sybilVote: sybilVotes,
		},
		"final_result": map[string]int{
			"yes": honestYes + sybilVotes,
			"no":  honestNo,
		},
		"legitimate_result":  "50/50 (如果只计算诚实节点)",
		"manipulated_result": fmt.Sprintf("%.1f%% yes", float64(honestYes+sybilVotes)/float64(totalVotes)*100),
		"attack_successful":  true,
	}

	s.EmitEvent("sybil_voting_attack", "", "", result)

	return result
}

// ShowDefenses 显示防御方法
func (s *SybilAttackSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"defense":       "工作量证明 (PoW)",
			"description":   "创建每个身份需要消耗计算资源",
			"example":       "比特币挖矿",
			"effectiveness": "高",
			"cost":          "能源消耗",
		},
		{
			"defense":       "权益证明 (PoS)",
			"description":   "创建验证者需要质押代币",
			"example":       "以太坊PoS需要32 ETH质押",
			"effectiveness": "高",
			"cost":          "资本锁定",
		},
		{
			"defense":       "身份验证 (KYC)",
			"description":   "要求真实身份验证",
			"example":       "交易所KYC",
			"effectiveness": "高",
			"cost":          "隐私损失",
		},
		{
			"defense":       "信任网络 (Web of Trust)",
			"description":   "新身份需要现有成员担保",
			"example":       "PGP信任网络",
			"effectiveness": "中等",
			"cost":          "入门门槛",
		},
		{
			"defense":       "验证码 (CAPTCHA)",
			"description":   "区分人类和机器",
			"example":       "网站注册",
			"effectiveness": "低-中",
			"cost":          "用户体验",
		},
		{
			"defense":       "社交图分析",
			"description":   "检测异常的身份创建模式",
			"example":       "社交网络反机器人",
			"effectiveness": "中等",
			"cost":          "计算复杂",
		},
	}
}

// CalculateSybilResistance 计算女巫抵抗力
func (s *SybilAttackSimulator) CalculateSybilResistance() map[string]interface{} {
	totalNodes := len(s.honestNodes) + len(s.sybilNodes)
	sybilRatio := float64(len(s.sybilNodes)) / float64(len(s.honestNodes))

	return map[string]interface{}{
		"network_size":     totalNodes,
		"honest_nodes":     len(s.honestNodes),
		"sybil_nodes":      len(s.sybilNodes),
		"sybil_ratio":      fmt.Sprintf("%.2f:1", sybilRatio),
		"sybil_percentage": fmt.Sprintf("%.1f%%", float64(len(s.sybilNodes))/float64(totalNodes)*100),
		"vulnerability_assessment": map[string]string{
			"routing": s.assessVulnerability(sybilRatio, 1.0),
			"voting":  s.assessVulnerability(sybilRatio, 0.5),
			"storage": s.assessVulnerability(sybilRatio, 0.3),
		},
	}
}

// assessVulnerability 评估脆弱性
func (s *SybilAttackSimulator) assessVulnerability(ratio, threshold float64) string {
	if ratio >= threshold*3 {
		return "极高风险"
	} else if ratio >= threshold*2 {
		return "高风险"
	} else if ratio >= threshold {
		return "中等风险"
	}
	return "低风险"
}

// updateState 更新状态
func (s *SybilAttackSimulator) updateState() {
	s.SetGlobalData("honest_nodes", len(s.honestNodes))
	s.SetGlobalData("sybil_nodes", len(s.sybilNodes))
	s.SetGlobalData("sybil_ratio", float64(len(s.sybilNodes))/float64(len(s.honestNodes)))

	ratio := 0.0
	if len(s.honestNodes) > 0 {
		ratio = float64(len(s.sybilNodes)) / float64(len(s.honestNodes))
	}

	setNetworkTeachingState(
		s.BaseSimulator,
		"network",
		"sybil_pressure",
		fmt.Sprintf("当前网络中有 %d 个诚实节点、%d 个 Sybil 节点，伪造身份比例约为 %.2f。", len(s.honestNodes), len(s.sybilNodes), ratio),
		"可以继续模拟投票攻击或 DHT 路由污染，观察虚假身份如何放大网络偏差。",
		minFloat(0.95, ratio/2+0.25),
		map[string]interface{}{
			"honest_nodes": len(s.honestNodes),
			"sybil_nodes":  len(s.sybilNodes),
			"sybil_ratio":  ratio,
		},
	)
}

// ExecuteAction 为 Sybil 攻击实验提供交互动作。
func (s *SybilAttackSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_dht_attack":
		targetHash, _ := params["target_key_hash"].(string)
		if targetHash == "" {
			targetHash = "deadbeefcafebabe"
		}
		result := s.SimulateDHTAttack(targetHash)
		return networkActionResult(
			"已模拟一轮 DHT 路由污染攻击。",
			result,
			&types.ActionFeedback{
				Summary:     "攻击者正在通过伪造身份占据更多路由位置，影响目标键的查找结果。",
				NextHint:    "继续观察目标键附近的路由节点是否被 Sybil 身份占据，以及查找路径是否被扭曲。",
				EffectScope: "network",
				ResultState: result,
			},
		), nil
	case "simulate_voting_attack":
		proposalID, _ := params["proposal_id"].(string)
		if proposalID == "" {
			proposalID = "proposal-demo"
		}
		result := s.SimulateVotingAttack(proposalID)
		return networkActionResult(
			"已模拟一轮投票型 Sybil 攻击。",
			result,
			&types.ActionFeedback{
				Summary:     "伪造身份已经开始放大投票权重，真实多数意见可能被覆盖。",
				NextHint:    "重点观察诚实票数与最终结果之间的偏差，以及伪造身份比例对操纵成功率的影响。",
				EffectScope: "network",
				ResultState: result,
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported sybil attack action: %s", action)
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// 工厂
// =============================================================================

// SybilAttackFactory 女巫攻击工厂
type SybilAttackFactory struct{}

// Create 创建演示器
func (f *SybilAttackFactory) Create() engine.Simulator {
	return NewSybilAttackSimulator()
}

// GetDescription 获取描述
func (f *SybilAttackFactory) GetDescription() types.Description {
	return NewSybilAttackSimulator().GetDescription()
}

// NewSybilAttackFactory 创建工厂
func NewSybilAttackFactory() *SybilAttackFactory {
	return &SybilAttackFactory{}
}

var _ engine.SimulatorFactory = (*SybilAttackFactory)(nil)
