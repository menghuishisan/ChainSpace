package network

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// NAT穿透演示器
// =============================================================================

// NATType NAT类型
type NATType string

const (
	NATFull           NATType = "full_cone"       // 全锥型
	NATRestricted     NATType = "restricted_cone" // 受限锥型
	NATPortRestricted NATType = "port_restricted" // 端口受限锥型
	NATSymmetric      NATType = "symmetric"       // 对称型
)

// NATNode NAT节点
type NATNode struct {
	ID         string  `json:"id"`
	PrivateIP  string  `json:"private_ip"`
	PublicIP   string  `json:"public_ip"`
	NATType    NATType `json:"nat_type"`
	PortMapped int     `json:"port_mapped"`
	CanReceive bool    `json:"can_receive_inbound"`
}

// TraversalMethod 穿透方法
type TraversalMethod struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Applicable  []NATType `json:"applicable_nat_types"`
	SuccessRate float64   `json:"success_rate"`
}

// TraversalAttempt 穿透尝试
type TraversalAttempt struct {
	FromNode string   `json:"from_node"`
	ToNode   string   `json:"to_node"`
	Method   string   `json:"method"`
	Success  bool     `json:"success"`
	Steps    []string `json:"steps"`
	Latency  int      `json:"latency_ms"`
}

// NATTraversalSimulator NAT穿透演示器
// 演示P2P网络中的NAT穿透技术:
//
// NAT类型:
// 1. 全锥型 - 最容易穿透
// 2. 受限锥型 - 需要先发出请求
// 3. 端口受限锥型 - 需要精确端口匹配
// 4. 对称型 - 最难穿透
//
// 穿透技术:
// 1. STUN - 发现公网地址
// 2. TURN - 中继服务器
// 3. 打洞 - 利用NAT映射
// 4. UPnP/NAT-PMP - 主动端口映射
type NATTraversalSimulator struct {
	*base.BaseSimulator
	nodes    map[string]*NATNode
	attempts []*TraversalAttempt
}

// NewNATTraversalSimulator 创建NAT穿透演示器
func NewNATTraversalSimulator() *NATTraversalSimulator {
	sim := &NATTraversalSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"nat_traversal",
			"NAT穿透演示器",
			"演示P2P网络中的NAT类型识别和穿透技术",
			"network",
			types.ComponentDemo,
		),
		nodes:    make(map[string]*NATNode),
		attempts: make([]*TraversalAttempt, 0),
	}

	return sim
}

// Init 初始化
func (s *NATTraversalSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	s.initializeNodes()
	s.updateState()
	return nil
}

// initializeNodes 初始化节点
func (s *NATTraversalSimulator) initializeNodes() {
	s.nodes = map[string]*NATNode{
		"node-A": {
			ID:         "node-A",
			PrivateIP:  "192.168.1.100",
			PublicIP:   "203.0.113.50",
			NATType:    NATFull,
			PortMapped: 30303,
			CanReceive: true,
		},
		"node-B": {
			ID:         "node-B",
			PrivateIP:  "10.0.0.50",
			PublicIP:   "198.51.100.25",
			NATType:    NATRestricted,
			PortMapped: 30303,
			CanReceive: false,
		},
		"node-C": {
			ID:         "node-C",
			PrivateIP:  "172.16.0.100",
			PublicIP:   "192.0.2.100",
			NATType:    NATSymmetric,
			PortMapped: 0, // 每次不同
			CanReceive: false,
		},
	}
}

// ExplainNATTypes 解释NAT类型
func (s *NATTraversalSimulator) ExplainNATTypes() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":        "full_cone",
			"name":        "全锥型NAT",
			"description": "一旦内部地址(iAddr:iPort)映射到外部地址(eAddr:ePort)，任何外部主机都可以通过(eAddr:ePort)到达内部主机",
			"traversal":   "最容易穿透",
			"diagram": `
内部: 192.168.1.100:5000 ─┐
                         ├──► NAT ──► 203.0.113.50:30000
任意外部主机 ◄────────────┘
            `,
		},
		{
			"type":        "restricted_cone",
			"name":        "受限锥型NAT",
			"description": "只有内部主机曾经发送过数据包的外部IP才能发送数据包回来",
			"traversal":   "需要先发出请求",
			"diagram": `
内部: 192.168.1.100:5000 ──► NAT ──► 203.0.113.50:30000
                                         │
只有通信过的IP ◄────────────────────────┘
            `,
		},
		{
			"type":        "port_restricted",
			"name":        "端口受限锥型NAT",
			"description": "只有内部主机曾经发送过数据包的外部IP:Port才能发送数据包回来",
			"traversal":   "需要精确端口匹配",
			"diagram": `
内部: 192.168.1.100:5000 ──► NAT ──► 203.0.113.50:30000
                                         │
只有通信过的IP:Port ◄──────────────────┘
            `,
		},
		{
			"type":        "symmetric",
			"name":        "对称型NAT",
			"description": "对于每个不同的外部目标，NAT使用不同的端口映射",
			"traversal":   "最难穿透，通常需要TURN中继",
			"diagram": `
内部 ──► 目标A ──► NAT使用端口30001
     └─► 目标B ──► NAT使用端口30002
     └─► 目标C ──► NAT使用端口30003
            `,
		},
	}
}

// ExplainTraversalMethods 解释穿透方法
func (s *NATTraversalSimulator) ExplainTraversalMethods() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"method":      "STUN",
			"name":        "Session Traversal Utilities for NAT",
			"description": "通过公共STUN服务器发现自己的公网IP和端口",
			"works_for":   []string{"full_cone", "restricted_cone", "port_restricted"},
			"fails_for":   []string{"symmetric (部分情况)"},
			"process": []string{
				"1. 客户端向STUN服务器发送请求",
				"2. STUN服务器返回观察到的IP:Port",
				"3. 客户端用这个地址告诉其他节点",
			},
		},
		{
			"method":       "TURN",
			"name":         "Traversal Using Relays around NAT",
			"description":  "当直接连接失败时，通过中继服务器转发数据",
			"works_for":    []string{"所有NAT类型"},
			"disadvantage": "增加延迟，中继服务器需要带宽",
			"process": []string{
				"1. 客户端连接到TURN服务器",
				"2. TURN服务器分配中继地址",
				"3. 数据通过TURN服务器中转",
			},
		},
		{
			"method":      "hole_punching",
			"name":        "UDP打洞",
			"description": "两个NAT后的节点同时向对方发送数据包，创建NAT映射",
			"works_for":   []string{"full_cone", "restricted_cone", "port_restricted"},
			"fails_for":   []string{"symmetric + symmetric"},
			"process": []string{
				"1. 两节点通过信令服务器交换公网地址",
				"2. 同时向对方发送UDP包",
				"3. NAT创建映射，后续包可以通过",
			},
		},
		{
			"method":      "UPnP/NAT-PMP",
			"name":        "主动端口映射",
			"description": "请求路由器主动创建端口转发规则",
			"works_for":   []string{"支持UPnP的路由器"},
			"security":    "可能被禁用因为安全考虑",
			"process": []string{
				"1. 发现网关设备",
				"2. 请求添加端口映射",
				"3. 路由器创建转发规则",
			},
		},
	}
}

// SimulateTraversal 模拟NAT穿透
func (s *NATTraversalSimulator) SimulateTraversal(fromNode, toNode string) *TraversalAttempt {
	from := s.nodes[fromNode]
	to := s.nodes[toNode]

	if from == nil || to == nil {
		return nil
	}

	attempt := &TraversalAttempt{
		FromNode: fromNode,
		ToNode:   toNode,
		Steps:    make([]string, 0),
	}

	// 确定穿透方法
	switch {
	case to.NATType == NATFull:
		attempt.Method = "direct"
		attempt.Steps = append(attempt.Steps, "目标节点是全锥型NAT，可以直接连接")
		attempt.Steps = append(attempt.Steps, fmt.Sprintf("连接到 %s:%d", to.PublicIP, to.PortMapped))
		attempt.Success = true
		attempt.Latency = 50

	case from.NATType != NATSymmetric && to.NATType != NATSymmetric:
		attempt.Method = "hole_punching"
		attempt.Steps = append(attempt.Steps, "双方NAT类型支持打洞")
		attempt.Steps = append(attempt.Steps, "通过信令服务器交换公网地址")
		attempt.Steps = append(attempt.Steps, fmt.Sprintf("%s 向 %s:%d 发送UDP包", from.ID, to.PublicIP, to.PortMapped))
		attempt.Steps = append(attempt.Steps, fmt.Sprintf("%s 向 %s:%d 发送UDP包", to.ID, from.PublicIP, from.PortMapped))
		attempt.Steps = append(attempt.Steps, "NAT映射建立，打洞成功")
		attempt.Success = true
		attempt.Latency = 100

	case from.NATType == NATSymmetric && to.NATType == NATSymmetric:
		attempt.Method = "TURN_relay"
		attempt.Steps = append(attempt.Steps, "双方都是对称型NAT，打洞失败")
		attempt.Steps = append(attempt.Steps, "回退到TURN中继")
		attempt.Steps = append(attempt.Steps, "连接到TURN服务器: turn.example.com")
		attempt.Steps = append(attempt.Steps, "分配中继地址")
		attempt.Steps = append(attempt.Steps, "通过中继建立连接")
		attempt.Success = true
		attempt.Latency = 200

	default:
		attempt.Method = "TURN_relay"
		attempt.Steps = append(attempt.Steps, "NAT类型组合不支持直接打洞")
		attempt.Steps = append(attempt.Steps, "使用TURN中继")
		attempt.Success = true
		attempt.Latency = 150
	}

	s.attempts = append(s.attempts, attempt)

	s.EmitEvent("traversal_attempt", "", "", map[string]interface{}{
		"from":    fromNode,
		"to":      toNode,
		"method":  attempt.Method,
		"success": attempt.Success,
		"latency": attempt.Latency,
	})

	s.updateState()
	return attempt
}

// GetNATCompatibilityMatrix 获取NAT兼容性矩阵
func (s *NATTraversalSimulator) GetNATCompatibilityMatrix() map[string]interface{} {
	return map[string]interface{}{
		"description": "不同NAT类型组合的穿透可能性",
		"matrix": []map[string]interface{}{
			{"from": "Full Cone", "to": "Full Cone", "method": "直接连接", "success": "100%"},
			{"from": "Full Cone", "to": "Restricted", "method": "直接连接", "success": "100%"},
			{"from": "Full Cone", "to": "Port Restricted", "method": "直接连接", "success": "100%"},
			{"from": "Full Cone", "to": "Symmetric", "method": "直接连接", "success": "100%"},
			{"from": "Restricted", "to": "Restricted", "method": "UDP打洞", "success": "~95%"},
			{"from": "Restricted", "to": "Port Restricted", "method": "UDP打洞", "success": "~90%"},
			{"from": "Restricted", "to": "Symmetric", "method": "UDP打洞", "success": "~50%"},
			{"from": "Port Restricted", "to": "Port Restricted", "method": "UDP打洞", "success": "~85%"},
			{"from": "Port Restricted", "to": "Symmetric", "method": "端口预测", "success": "~30%"},
			{"from": "Symmetric", "to": "Symmetric", "method": "TURN中继", "success": "需要中继"},
		},
	}
}

// updateState 更新状态
func (s *NATTraversalSimulator) updateState() {
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("attempt_count", len(s.attempts))

	stage := "network_ready"
	summary := "当前 NAT 穿透实验已就绪，可以发起一轮连接建立尝试。"
	nextHint := "选择两种不同 NAT 类型的节点，观察是直接连接、打洞还是回退到中继。"
	progress := 0.2
	result := map[string]interface{}{
		"node_count":     len(s.nodes),
		"attempt_count":  len(s.attempts),
	}

	if len(s.attempts) > 0 {
		latest := s.attempts[len(s.attempts)-1]
		stage = "traversal_completed"
		summary = fmt.Sprintf("最近一次连接尝试采用 %s 方法，结果为%s。", latest.Method, map[bool]string{true: "成功", false: "失败"}[latest.Success])
		nextHint = "继续对比不同 NAT 组合下的成功率和时延差异，理解何时必须回退到 TURN。"
		progress = 0.8
		result["latest_method"] = latest.Method
		result["latest_success"] = latest.Success
		result["latest_latency"] = latest.Latency
	}

	setNetworkTeachingState(s.BaseSimulator, "network", stage, summary, nextHint, progress, result)
}

// ExecuteAction 为 NAT 穿透实验提供交互动作。
func (s *NATTraversalSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_traversal":
		fromNode, _ := params["from_node"].(string)
		toNode, _ := params["to_node"].(string)
		result := s.SimulateTraversal(fromNode, toNode)
		return networkActionResult(
			"已完成一轮 NAT 穿透连接尝试。",
			map[string]interface{}{
				"from":    result.FromNode,
				"to":      result.ToNode,
				"method":  result.Method,
				"success": result.Success,
				"latency": result.Latency,
			},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("连接尝试采用 %s 方法，结果为%s。", result.Method, map[bool]string{true: "成功", false: "失败"}[result.Success]),
				NextHint:    "重点观察 NAT 类型组合如何决定是直接连接、UDP 打洞还是 TURN 中继。",
				EffectScope: "network",
				ResultState: map[string]interface{}{
					"method":  result.Method,
					"success": result.Success,
				},
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported nat traversal action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// NATTraversalFactory NAT穿透工厂
type NATTraversalFactory struct{}

// Create 创建演示器
func (f *NATTraversalFactory) Create() engine.Simulator {
	return NewNATTraversalSimulator()
}

// GetDescription 获取描述
func (f *NATTraversalFactory) GetDescription() types.Description {
	return NewNATTraversalSimulator().GetDescription()
}

// NewNATTraversalFactory 创建工厂
func NewNATTraversalFactory() *NATTraversalFactory {
	return &NATTraversalFactory{}
}

var _ engine.SimulatorFactory = (*NATTraversalFactory)(nil)
