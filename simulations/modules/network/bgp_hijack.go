package network

import (
	"fmt"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// BGP劫持攻击演示器
// =============================================================================

// ASInfo 自治系统信息
type ASInfo struct {
	ASN         uint32   `json:"asn"`
	Name        string   `json:"name"`
	Prefixes    []string `json:"prefixes"`
	Neighbors   []uint32 `json:"neighbors"`
	IsMalicious bool     `json:"is_malicious"`
}

// BGPAnnouncement BGP公告
type BGPAnnouncement struct {
	Prefix    string    `json:"prefix"`
	Origin    uint32    `json:"origin_asn"`
	ASPath    []uint32  `json:"as_path"`
	NextHop   string    `json:"next_hop"`
	Timestamp time.Time `json:"timestamp"`
	IsLegit   bool      `json:"is_legitimate"`
}

// BGPHijackState BGP劫持状态
type BGPHijackState struct {
	AttackerAS      uint32             `json:"attacker_as"`
	VictimPrefix    string             `json:"victim_prefix"`
	HijackType      string             `json:"hijack_type"`
	AffectedASes    []uint32           `json:"affected_ases"`
	TrafficDiverted float64            `json:"traffic_diverted_percent"`
	StartTime       time.Time          `json:"start_time"`
	Announcements   []*BGPAnnouncement `json:"announcements"`
}

// BGPHijackSimulator BGP劫持攻击演示器
// 演示BGP劫持对区块链网络的影响:
//
// BGP劫持类型:
// 1. 前缀劫持 - 宣告他人的IP前缀
// 2. 子前缀劫持 - 宣告更具体的子网
// 3. AS路径篡改 - 缩短AS路径吸引流量
//
// 对区块链的影响:
// 1. 延迟区块传播
// 2. 分区攻击
// 3. 双花攻击
// 4. 挖矿算力窃取
type BGPHijackSimulator struct {
	*base.BaseSimulator
	asMap       map[uint32]*ASInfo
	hijackState *BGPHijackState
}

// NewBGPHijackSimulator 创建BGP劫持演示器
func NewBGPHijackSimulator() *BGPHijackSimulator {
	sim := &BGPHijackSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"bgp_hijack",
			"BGP劫持攻击演示器",
			"演示BGP劫持如何影响区块链网络的连接和安全",
			"network",
			types.ComponentAttack,
		),
		asMap: make(map[uint32]*ASInfo),
	}

	return sim
}

// Init 初始化
func (s *BGPHijackSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.initializeASTopology()
	s.updateState()
	return nil
}

// initializeASTopology 初始化AS拓扑
func (s *BGPHijackSimulator) initializeASTopology() {
	s.asMap = map[uint32]*ASInfo{
		64500: {
			ASN:         64500,
			Name:        "Victim Mining Pool",
			Prefixes:    []string{"203.0.113.0/24"},
			Neighbors:   []uint32{64501, 64502},
			IsMalicious: false,
		},
		64501: {
			ASN:         64501,
			Name:        "Transit Provider A",
			Prefixes:    []string{"198.51.100.0/24"},
			Neighbors:   []uint32{64500, 64502, 64503},
			IsMalicious: false,
		},
		64502: {
			ASN:         64502,
			Name:        "Transit Provider B",
			Prefixes:    []string{"192.0.2.0/24"},
			Neighbors:   []uint32{64500, 64501, 64503, 64504},
			IsMalicious: false,
		},
		64503: {
			ASN:         64503,
			Name:        "Attacker AS",
			Prefixes:    []string{"10.0.0.0/8"},
			Neighbors:   []uint32{64501, 64502},
			IsMalicious: true,
		},
		64504: {
			ASN:         64504,
			Name:        "Bitcoin Nodes Network",
			Prefixes:    []string{"172.16.0.0/16"},
			Neighbors:   []uint32{64502},
			IsMalicious: false,
		},
	}
}

// ExplainBGP 解释BGP
func (s *BGPHijackSimulator) ExplainBGP() map[string]interface{} {
	return map[string]interface{}{
		"what_is_bgp": "边界网关协议(Border Gateway Protocol)是互联网的核心路由协议",
		"how_it_works": []string{
			"互联网由多个自治系统(AS)组成",
			"每个AS通过BGP宣告自己的IP前缀",
			"AS之间交换路由信息，决定流量转发路径",
			"BGP基于信任模型，没有内置验证",
		},
		"vulnerability": "BGP设计时假设所有参与者都是诚实的，缺乏认证机制",
		"hijack_types": []map[string]string{
			{"type": "prefix_hijack", "description": "宣告他人的IP前缀"},
			{"type": "subprefix_hijack", "description": "宣告更具体的子网(优先级更高)"},
			{"type": "path_manipulation", "description": "伪造更短的AS路径"},
		},
	}
}

// SimulatePrefixHijack 模拟前缀劫持
func (s *BGPHijackSimulator) SimulatePrefixHijack(victimPrefix string) *BGPHijackState {
	attackerAS := s.asMap[64503]

	s.hijackState = &BGPHijackState{
		AttackerAS:    attackerAS.ASN,
		VictimPrefix:  victimPrefix,
		HijackType:    "prefix_hijack",
		AffectedASes:  make([]uint32, 0),
		StartTime:     time.Now(),
		Announcements: make([]*BGPAnnouncement, 0),
	}

	// 合法公告
	legitAnnouncement := &BGPAnnouncement{
		Prefix:    victimPrefix,
		Origin:    64500,
		ASPath:    []uint32{64500},
		NextHop:   "203.0.113.1",
		Timestamp: time.Now().Add(-time.Hour),
		IsLegit:   true,
	}
	s.hijackState.Announcements = append(s.hijackState.Announcements, legitAnnouncement)

	// 恶意公告
	maliciousAnnouncement := &BGPAnnouncement{
		Prefix:    victimPrefix,
		Origin:    64503,
		ASPath:    []uint32{64503}, // 假装是origin
		NextHop:   "10.0.0.1",
		Timestamp: time.Now(),
		IsLegit:   false,
	}
	s.hijackState.Announcements = append(s.hijackState.Announcements, maliciousAnnouncement)

	// 计算受影响的AS
	for asn := range s.asMap {
		if asn != 64500 && asn != 64503 {
			s.hijackState.AffectedASes = append(s.hijackState.AffectedASes, asn)
		}
	}
	s.hijackState.TrafficDiverted = 60.0 // 假设60%流量被劫持

	s.EmitEvent("bgp_hijack", "", "", map[string]interface{}{
		"type":             "prefix_hijack",
		"attacker":         attackerAS.Name,
		"victim_prefix":    victimPrefix,
		"affected_ases":    len(s.hijackState.AffectedASes),
		"traffic_diverted": fmt.Sprintf("%.1f%%", s.hijackState.TrafficDiverted),
	})

	s.updateState()
	return s.hijackState
}

// SimulateSubprefixHijack 模拟子前缀劫持
func (s *BGPHijackSimulator) SimulateSubprefixHijack(victimPrefix, subprefix string) *BGPHijackState {
	attackerAS := s.asMap[64503]

	s.hijackState = &BGPHijackState{
		AttackerAS:    attackerAS.ASN,
		VictimPrefix:  victimPrefix,
		HijackType:    "subprefix_hijack",
		AffectedASes:  make([]uint32, 0),
		StartTime:     time.Now(),
		Announcements: make([]*BGPAnnouncement, 0),
	}

	// 合法公告
	legitAnnouncement := &BGPAnnouncement{
		Prefix:    victimPrefix,
		Origin:    64500,
		ASPath:    []uint32{64500},
		NextHop:   "203.0.113.1",
		Timestamp: time.Now().Add(-time.Hour),
		IsLegit:   true,
	}
	s.hijackState.Announcements = append(s.hijackState.Announcements, legitAnnouncement)

	// 恶意子前缀公告 (更具体的前缀优先级更高)
	maliciousAnnouncement := &BGPAnnouncement{
		Prefix:    subprefix, // 例如 203.0.113.0/25 比 /24 更具体
		Origin:    64503,
		ASPath:    []uint32{64503},
		NextHop:   "10.0.0.1",
		Timestamp: time.Now(),
		IsLegit:   false,
	}
	s.hijackState.Announcements = append(s.hijackState.Announcements, maliciousAnnouncement)

	// 子前缀劫持影响更广
	for asn := range s.asMap {
		if asn != 64503 {
			s.hijackState.AffectedASes = append(s.hijackState.AffectedASes, asn)
		}
	}
	s.hijackState.TrafficDiverted = 95.0 // 子前缀劫持几乎100%有效

	s.EmitEvent("bgp_hijack", "", "", map[string]interface{}{
		"type":             "subprefix_hijack",
		"attacker":         attackerAS.Name,
		"victim_prefix":    victimPrefix,
		"malicious_prefix": subprefix,
		"traffic_diverted": fmt.Sprintf("%.1f%%", s.hijackState.TrafficDiverted),
	})

	s.updateState()
	return s.hijackState
}

// SimulateBlockchainImpact 模拟对区块链的影响
func (s *BGPHijackSimulator) SimulateBlockchainImpact() map[string]interface{} {
	if s.hijackState == nil {
		return map[string]interface{}{"error": "需要先模拟BGP劫持"}
	}

	return map[string]interface{}{
		"hijack_state": s.hijackState.HijackType,
		"impacts": []map[string]interface{}{
			{
				"impact":      "区块传播延迟",
				"description": "劫持者可以延迟转发区块",
				"consequence": "增加孤块率，降低网络效率",
				"delay_added": "10-60秒",
			},
			{
				"impact":         "网络分区",
				"description":    "将部分节点与网络隔离",
				"consequence":    "可能导致链分叉",
				"affected_nodes": fmt.Sprintf("约%.0f%%", s.hijackState.TrafficDiverted),
			},
			{
				"impact":      "双花攻击",
				"description": "配合网络分区进行双花",
				"consequence": "交易可能被回滚",
			},
			{
				"impact":      "挖矿算力窃取",
				"description": "劫持矿池连接，替换区块模板",
				"consequence": "矿工为攻击者挖矿",
			},
			{
				"impact":      "中间人攻击",
				"description": "监听和修改节点通信",
				"consequence": "窃取私钥或篡改交易",
			},
		},
		"real_world_research": map[string]string{
			"paper":   "Hijacking Bitcoin: Routing Attacks on Cryptocurrencies",
			"authors": "Apostolaki et al.",
			"year":    "2017",
			"finding": "AS级攻击者可以有效分区比特币网络",
		},
	}
}

// ShowDefenses 显示防御方法
func (s *BGPHijackSimulator) ShowDefenses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"defense":       "RPKI (Resource Public Key Infrastructure)",
			"description":   "为IP前缀提供密码学验证",
			"status":        "逐步部署中",
			"effectiveness": "高",
		},
		{
			"defense":       "BGPsec",
			"description":   "为AS路径提供签名验证",
			"status":        "标准化完成，部署缓慢",
			"effectiveness": "高",
		},
		{
			"defense":       "路由监控",
			"description":   "监控BGP公告异常",
			"tools":         []string{"BGPStream", "RIPE RIS", "RouteViews"},
			"effectiveness": "检测用，不能防止",
		},
		{
			"defense":       "多路径连接",
			"description":   "通过多个ISP连接，减少单点依赖",
			"effectiveness": "中等",
		},
		{
			"defense":       "加密通信",
			"description":   "使用TLS/加密协议，即使被劫持也无法篡改",
			"effectiveness": "防止窃听和篡改",
		},
		{
			"defense":       "Tor/VPN",
			"description":   "隐藏真实IP，增加劫持难度",
			"effectiveness": "中等",
		},
	}
}

// GetRealWorldCases 获取真实案例
func (s *BGPHijackSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"event":  "MyEtherWallet DNS劫持",
			"date":   "2018-04-24",
			"method": "BGP劫持Amazon Route 53 DNS服务器",
			"impact": "用户被引导至钓鱼网站",
			"loss":   "~$150,000 ETH被盗",
		},
		{
			"event":    "比特币矿池BGP劫持",
			"date":     "2014",
			"attacker": "加拿大ISP",
			"impact":   "矿池流量被劫持约19个月",
			"loss":     "约$83,000 BTC被窃取",
		},
		{
			"event":    "YouTube全球中断",
			"date":     "2008-02-24",
			"attacker": "巴基斯坦电信",
			"method":   "意外宣告YouTube的子前缀",
			"impact":   "全球约2小时无法访问YouTube",
		},
	}
}

// updateState 更新状态
func (s *BGPHijackSimulator) updateState() {
	s.SetGlobalData("as_count", len(s.asMap))
	if s.hijackState != nil {
		s.SetGlobalData("hijack_active", true)
		s.SetGlobalData("hijack_type", s.hijackState.HijackType)
		s.SetGlobalData("traffic_diverted", s.hijackState.TrafficDiverted)
	} else {
		s.SetGlobalData("hijack_active", false)
	}

	stage := "network_ready"
	summary := "当前没有进行 BGP 劫持，跨自治系统的路由仍然保持正常。"
	nextHint := "可以先模拟前缀劫持，再观察流量被重定向后对区块链传播与连通性的影响。"
	progress := 0.2
	result := map[string]interface{}{
		"as_count":       len(s.asMap),
		"hijack_active":  s.hijackState != nil,
	}

	if s.hijackState != nil {
		stage = "hijack_active"
		summary = fmt.Sprintf("当前正在进行 %s，预计可重定向 %.1f%% 的目标流量。", s.hijackState.HijackType, s.hijackState.TrafficDiverted)
		nextHint = "继续观察受影响的自治系统数量、受害前缀和区块链传播延迟如何变化。"
		progress = 0.85
		result["hijack_type"] = s.hijackState.HijackType
		result["traffic_diverted"] = s.hijackState.TrafficDiverted
		result["affected_ases"] = len(s.hijackState.AffectedASes)
	}

	setNetworkTeachingState(s.BaseSimulator, "network", stage, summary, nextHint, progress, result)
}

// ExecuteAction 为 BGP 劫持实验提供交互动作。
func (s *BGPHijackSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_prefix_hijack":
		prefix, _ := params["victim_prefix"].(string)
		if prefix == "" {
			prefix = "203.0.113.0/24"
		}
		state := s.SimulatePrefixHijack(prefix)
		return networkActionResult(
			"已模拟一轮前缀劫持。",
			map[string]interface{}{
				"hijack_type":      state.HijackType,
				"traffic_diverted": state.TrafficDiverted,
				"affected_ases":    len(state.AffectedASes),
			},
			&types.ActionFeedback{
				Summary:     "恶意自治系统已经宣布受害前缀，网络路由开始偏向攻击者。",
				NextHint:    "继续观察传播延迟、链分区和区块链节点可达性如何被改变。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"hijack_type": state.HijackType},
			},
		), nil
	case "simulate_subprefix_hijack":
		prefix, _ := params["victim_prefix"].(string)
		subprefix, _ := params["subprefix"].(string)
		if prefix == "" {
			prefix = "203.0.113.0/24"
		}
		if subprefix == "" {
			subprefix = "203.0.113.0/25"
		}
		state := s.SimulateSubprefixHijack(prefix, subprefix)
		return networkActionResult(
			"已模拟一轮子前缀劫持。",
			map[string]interface{}{
				"hijack_type":      state.HijackType,
				"traffic_diverted": state.TrafficDiverted,
				"affected_ases":    len(state.AffectedASes),
			},
			&types.ActionFeedback{
				Summary:     "更具体的子前缀声明已经发布，攻击者对流量的控制能力进一步增强。",
				NextHint:    "继续观察被重定向的流量比例和区块链传播异常是否进一步扩大。",
				EffectScope: "network",
				ResultState: map[string]interface{}{"hijack_type": state.HijackType},
			},
		), nil
	case "simulate_blockchain_impact":
		result := s.SimulateBlockchainImpact()
		if _, ok := result["error"]; ok {
			return &types.ActionResult{Success: false, Message: "需要先模拟一轮 BGP 劫持。"}, nil
		}
		return networkActionResult(
			"已分析当前劫持对区块链传播的影响。",
			result,
			&types.ActionFeedback{
				Summary:     "当前劫持已经开始影响区块传播、链分区风险和中间人攻击面。",
				NextHint:    "重点观察传播延迟、分区概率和节点视图是否出现明显偏差。",
				EffectScope: "network",
				ResultState: result,
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported bgp hijack action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// BGPHijackFactory BGP劫持工厂
type BGPHijackFactory struct{}

// Create 创建演示器
func (f *BGPHijackFactory) Create() engine.Simulator {
	return NewBGPHijackSimulator()
}

// GetDescription 获取描述
func (f *BGPHijackFactory) GetDescription() types.Description {
	return NewBGPHijackSimulator().GetDescription()
}

// NewBGPHijackFactory 创建工厂
func NewBGPHijackFactory() *BGPHijackFactory {
	return &BGPHijackFactory{}
}

var _ engine.SimulatorFactory = (*BGPHijackFactory)(nil)
