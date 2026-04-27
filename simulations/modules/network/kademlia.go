package network

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// =============================================================================
// Kademlia DHT演示器
// =============================================================================

// KademliaNode Kademlia节点
type KademliaNode struct {
	ID       [20]byte          `json:"id"`      // 160-bit节点ID
	IDHex    string            `json:"id_hex"`  // 十六进制表示
	Address  string            `json:"address"` // 网络地址
	KBuckets [160]*KBucket     `json:"-"`       // K桶数组
	Data     map[string][]byte `json:"-"`       // 存储的数据
	IsOnline bool              `json:"is_online"`
}

// KBucket K桶
type KBucket struct {
	Nodes    []*KademliaNode `json:"nodes"`
	LastSeen time.Time       `json:"last_seen"`
}

// KademliaLookup 查找记录
type KademliaLookup struct {
	TargetID   string        `json:"target_id"`
	Initiator  string        `json:"initiator"`
	Steps      []*LookupStep `json:"steps"`
	FoundNode  *KademliaNode `json:"found_node,omitempty"`
	FoundValue string        `json:"found_value,omitempty"`
	TotalHops  int           `json:"total_hops"`
	Duration   time.Duration `json:"duration"`
}

// LookupStep 查找步骤
type LookupStep struct {
	Step         int      `json:"step"`
	CurrentNode  string   `json:"current_node"`
	QueriedNodes []string `json:"queried_nodes"`
	CloserNodes  []string `json:"closer_nodes"`
	XORDistance  string   `json:"xor_distance"`
}

// KademliaSimulator Kademlia DHT演示器
// 演示分布式哈希表的核心概念:
//
// 1. XOR距离度量 - 节点间距离 = ID XOR
// 2. K桶 - 按距离分层存储节点
// 3. 节点查找 - 迭代查找最近节点
// 4. 值存储/检索 - 分布式键值存储
//
// 参数:
// - K = 20 (K桶大小)
// - α = 3 (并行查询数)
// - ID长度 = 160 bits (SHA-1)
type KademliaSimulator struct {
	*base.BaseSimulator
	nodes   map[string]*KademliaNode
	k       int // K桶大小
	alpha   int // 并行度
	lookups []*KademliaLookup
}

// NewKademliaSimulator 创建Kademlia演示器
func NewKademliaSimulator() *KademliaSimulator {
	sim := &KademliaSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"kademlia",
			"Kademlia DHT演示器",
			"演示分布式哈希表的XOR距离、K桶和节点查找",
			"network",
			types.ComponentProcess,
		),
		nodes:   make(map[string]*KademliaNode),
		k:       20,
		alpha:   3,
		lookups: make([]*KademliaLookup, 0),
	}

	sim.AddParam(types.Param{
		Key:         "node_count",
		Name:        "节点数量",
		Description: "DHT网络中的节点数量",
		Type:        types.ParamTypeInt,
		Default:     50,
		Min:         10,
		Max:         1000,
	})

	sim.AddParam(types.Param{
		Key:         "k_bucket_size",
		Name:        "K桶大小",
		Description: "每个K桶的最大节点数",
		Type:        types.ParamTypeInt,
		Default:     20,
		Min:         1,
		Max:         100,
	})

	return sim
}

// Init 初始化
func (s *KademliaSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	nodeCount := 50
	if v, ok := config.Params["node_count"]; ok {
		if n, ok := v.(float64); ok {
			nodeCount = int(n)
		}
	}

	if v, ok := config.Params["k_bucket_size"]; ok {
		if n, ok := v.(float64); ok {
			s.k = int(n)
		}
	}

	s.initializeNetwork(nodeCount)
	s.updateState()
	return nil
}

// initializeNetwork 初始化网络
func (s *KademliaSimulator) initializeNetwork(nodeCount int) {
	s.nodes = make(map[string]*KademliaNode)

	for i := 0; i < nodeCount; i++ {
		node := s.createNode(fmt.Sprintf("node-%d", i))
		s.nodes[node.IDHex] = node
	}

	// 建立路由表
	for _, node := range s.nodes {
		s.buildRoutingTable(node)
	}
}

// createNode 创建节点
func (s *KademliaSimulator) createNode(seed string) *KademliaNode {
	hash := sha256.Sum256([]byte(seed))
	var id [20]byte
	copy(id[:], hash[:20])

	node := &KademliaNode{
		ID:       id,
		IDHex:    hex.EncodeToString(id[:]),
		Address:  fmt.Sprintf("192.168.1.%d:4001", len(s.nodes)+1),
		Data:     make(map[string][]byte),
		IsOnline: true,
	}

	// 初始化K桶
	for i := 0; i < 160; i++ {
		node.KBuckets[i] = &KBucket{
			Nodes:    make([]*KademliaNode, 0, s.k),
			LastSeen: time.Now(),
		}
	}

	return node
}

// buildRoutingTable 构建路由表
func (s *KademliaSimulator) buildRoutingTable(node *KademliaNode) {
	for _, other := range s.nodes {
		if other.IDHex == node.IDHex {
			continue
		}

		bucketIndex := s.getBucketIndex(node.ID, other.ID)
		bucket := node.KBuckets[bucketIndex]

		if len(bucket.Nodes) < s.k {
			bucket.Nodes = append(bucket.Nodes, other)
		}
	}
}

// ExplainXORDistance 解释XOR距离
func (s *KademliaSimulator) ExplainXORDistance() map[string]interface{} {
	return map[string]interface{}{
		"concept": "XOR距离度量",
		"formula": "distance(A, B) = A XOR B",
		"properties": []map[string]string{
			{"property": "同一性", "description": "d(A, A) = 0"},
			{"property": "对称性", "description": "d(A, B) = d(B, A)"},
			{"property": "三角不等式", "description": "d(A, C) ≤ d(A, B) + d(B, C)"},
		},
		"advantages": []string{
			"简单高效的位运算",
			"单向性: 对于任何距离只有一个节点",
			"自然形成树状结构",
		},
		"example": s.demonstrateXOR(),
	}
}

// demonstrateXOR 演示XOR运算
func (s *KademliaSimulator) demonstrateXOR() map[string]interface{} {
	a := []byte{0b10110100}
	b := []byte{0b10100010}
	xor := a[0] ^ b[0]

	return map[string]interface{}{
		"node_a":      fmt.Sprintf("%08b", a[0]),
		"node_b":      fmt.Sprintf("%08b", b[0]),
		"xor":         fmt.Sprintf("%08b", xor),
		"distance":    xor,
		"explanation": "XOR结果的最高有效位决定距离的量级",
	}
}

// ExplainKBuckets 解释K桶
func (s *KademliaSimulator) ExplainKBuckets() map[string]interface{} {
	return map[string]interface{}{
		"concept":     "K桶路由表",
		"description": "每个节点维护160个K桶，第i个桶存储距离在[2^i, 2^(i+1))范围内的节点",
		"structure": []map[string]interface{}{
			{"bucket": 0, "distance_range": "[1, 2)", "contains": "最近的节点(1 bit差异)"},
			{"bucket": 1, "distance_range": "[2, 4)", "contains": "2 bits差异的节点"},
			{"bucket": 2, "distance_range": "[4, 8)", "contains": "3 bits差异的节点"},
			{"bucket": "...", "distance_range": "...", "contains": "..."},
			{"bucket": 159, "distance_range": "[2^159, 2^160)", "contains": "最远的节点"},
		},
		"properties": []string{
			"每个桶最多存储K个节点 (默认K=20)",
			"越近的节点，桶越容易被填满",
			"越远的节点，单个桶覆盖的ID空间越大",
		},
		"lru_policy": "最近最少使用的节点可能被替换",
	}
}

// SimulateLookup 模拟节点查找
func (s *KademliaSimulator) SimulateLookup(startNodeID, targetID string) *KademliaLookup {
	startNode := s.nodes[startNodeID]
	if startNode == nil {
		// 使用第一个节点
		for _, node := range s.nodes {
			startNode = node
			break
		}
	}

	targetBytes, _ := hex.DecodeString(targetID)
	if len(targetBytes) < 20 {
		// 生成随机目标
		hash := sha256.Sum256([]byte(targetID))
		targetBytes = hash[:20]
		targetID = hex.EncodeToString(targetBytes)
	}

	var target [20]byte
	copy(target[:], targetBytes[:20])

	lookup := &KademliaLookup{
		TargetID:  targetID,
		Initiator: startNode.IDHex,
		Steps:     make([]*LookupStep, 0),
	}

	startTime := time.Now()

	// 迭代查找
	currentNode := startNode
	visited := make(map[string]bool)
	visited[currentNode.IDHex] = true

	for step := 0; step < 20; step++ { // 最多20步
		closestNodes := s.findClosestNodes(currentNode, target, s.alpha)

		stepRecord := &LookupStep{
			Step:         step + 1,
			CurrentNode:  currentNode.IDHex[:16] + "...",
			QueriedNodes: make([]string, 0),
			CloserNodes:  make([]string, 0),
			XORDistance:  s.calculateXORDistance(currentNode.ID, target),
		}

		var nextNode *KademliaNode
		currentDistance := s.xorDistanceBigInt(currentNode.ID, target)

		for _, node := range closestNodes {
			stepRecord.QueriedNodes = append(stepRecord.QueriedNodes, node.IDHex[:16]+"...")

			if visited[node.IDHex] {
				continue
			}
			visited[node.IDHex] = true

			nodeDistance := s.xorDistanceBigInt(node.ID, target)
			if nodeDistance.Cmp(currentDistance) < 0 {
				stepRecord.CloserNodes = append(stepRecord.CloserNodes, node.IDHex[:16]+"...")
				if nextNode == nil || nodeDistance.Cmp(s.xorDistanceBigInt(nextNode.ID, target)) < 0 {
					nextNode = node
				}
			}

			// 检查是否找到目标
			if node.IDHex == targetID {
				lookup.FoundNode = node
				break
			}
		}

		lookup.Steps = append(lookup.Steps, stepRecord)

		if lookup.FoundNode != nil {
			break
		}

		if nextNode == nil {
			break // 没有更近的节点了
		}

		currentNode = nextNode
	}

	lookup.TotalHops = len(lookup.Steps)
	lookup.Duration = time.Since(startTime)

	s.lookups = append(s.lookups, lookup)

	s.EmitEvent("lookup_completed", "", "", map[string]interface{}{
		"target":    targetID[:16] + "...",
		"hops":      lookup.TotalHops,
		"found":     lookup.FoundNode != nil,
		"initiator": lookup.Initiator[:16] + "...",
	})

	s.updateState()
	return lookup
}

// findClosestNodes 查找最近的节点
func (s *KademliaSimulator) findClosestNodes(node *KademliaNode, target [20]byte, count int) []*KademliaNode {
	type nodeDistance struct {
		node     *KademliaNode
		distance *big.Int
	}

	var candidates []nodeDistance

	// 收集所有已知节点
	for _, bucket := range node.KBuckets {
		if bucket == nil {
			continue
		}
		for _, n := range bucket.Nodes {
			if n != nil && n.IsOnline {
				candidates = append(candidates, nodeDistance{
					node:     n,
					distance: s.xorDistanceBigInt(n.ID, target),
				})
			}
		}
	}

	// 按距离排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance.Cmp(candidates[j].distance) < 0
	})

	// 返回最近的count个
	result := make([]*KademliaNode, 0, count)
	for i := 0; i < len(candidates) && i < count; i++ {
		result = append(result, candidates[i].node)
	}

	return result
}

// getBucketIndex 获取K桶索引
func (s *KademliaSimulator) getBucketIndex(a, b [20]byte) int {
	for i := 0; i < 20; i++ {
		xor := a[i] ^ b[i]
		if xor != 0 {
			// 找到第一个不同的字节，计算最高位
			for bit := 7; bit >= 0; bit-- {
				if (xor>>bit)&1 == 1 {
					return (i * 8) + (7 - bit)
				}
			}
		}
	}
	return 159
}

// xorDistanceBigInt 计算XOR距离(大整数)
func (s *KademliaSimulator) xorDistanceBigInt(a, b [20]byte) *big.Int {
	var xor [20]byte
	for i := 0; i < 20; i++ {
		xor[i] = a[i] ^ b[i]
	}
	return new(big.Int).SetBytes(xor[:])
}

// calculateXORDistance 计算XOR距离(字符串表示)
func (s *KademliaSimulator) calculateXORDistance(a, b [20]byte) string {
	distance := s.xorDistanceBigInt(a, b)
	return distance.Text(16)[:8] + "..." // 截断显示
}

// StoreValue 存储值
func (s *KademliaSimulator) StoreValue(key string, value []byte) map[string]interface{} {
	keyHash := sha256.Sum256([]byte(key))
	var keyID [20]byte
	copy(keyID[:], keyHash[:20])
	keyHex := hex.EncodeToString(keyID[:])

	// 找到最近的K个节点存储
	storedNodes := make([]string, 0)
	for _, node := range s.nodes {
		if len(storedNodes) >= s.k {
			break
		}
		node.Data[keyHex] = value
		storedNodes = append(storedNodes, node.IDHex[:16]+"...")
	}

	s.EmitEvent("value_stored", "", "", map[string]interface{}{
		"key":          key,
		"key_hash":     keyHex[:16] + "...",
		"stored_nodes": len(storedNodes),
	})

	return map[string]interface{}{
		"key":          key,
		"key_hash":     keyHex,
		"stored_nodes": storedNodes,
		"replication":  len(storedNodes),
	}
}

// updateState 更新状态
func (s *KademliaSimulator) updateState() {
	s.SetGlobalData("node_count", len(s.nodes))
	s.SetGlobalData("k_bucket_size", s.k)
	s.SetGlobalData("alpha", s.alpha)
	s.SetGlobalData("lookup_count", len(s.lookups))

	stage := "network_ready"
	summary := "当前 DHT 网络已经就绪，可以执行一次节点查找或键值存储。"
	nextHint := "先观察 XOR 距离和 K 桶如何帮助逐步缩小查找范围。"
	progress := 0.2
	result := map[string]interface{}{
		"node_count":    len(s.nodes),
		"k_bucket_size": s.k,
		"alpha":         s.alpha,
		"lookup_count":  len(s.lookups),
	}

	if len(s.lookups) > 0 {
		latest := s.lookups[len(s.lookups)-1]
		stage = "lookup_completed"
		summary = fmt.Sprintf("最近一次查找共经历 %d 跳，目标节点%s。", latest.TotalHops, map[bool]string{true: "已找到", false: "未找到"}[latest.FoundNode != nil])
		nextHint = "继续对比不同目标距离下的跳数变化，以及 α 并行度对查找速度的影响。"
		progress = 0.85
		result["latest_hops"] = latest.TotalHops
		result["target_found"] = latest.FoundNode != nil
	}

	setNetworkTeachingState(s.BaseSimulator, "network", stage, summary, nextHint, progress, result)
}

// ExecuteAction 为 Kademlia 实验提供交互动作。
func (s *KademliaSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "lookup_node":
		startNode, _ := params["start_node"].(string)
		targetID, _ := params["target_id"].(string)
		lookup := s.SimulateLookup(startNode, targetID)
		return networkActionResult(
			"已完成一轮 Kademlia 节点查找。",
			map[string]interface{}{
				"initiator": lookup.Initiator,
				"target":    lookup.TargetID,
				"total_hops": lookup.TotalHops,
				"found":     lookup.FoundNode != nil,
			},
			&types.ActionFeedback{
				Summary:     fmt.Sprintf("查找从 %s 发起，共经历 %d 跳。", lookup.Initiator[:16]+"...", lookup.TotalHops),
				NextHint:    "继续观察每一步返回的候选节点是否在不断逼近目标 ID。",
				EffectScope: "network",
				ResultState: map[string]interface{}{
					"total_hops": lookup.TotalHops,
					"found":      lookup.FoundNode != nil,
				},
			},
		), nil
	case "store_value":
		key, _ := params["key"].(string)
		valueText, _ := params["value"].(string)
		if key == "" {
			key = "demo-key"
		}
		if valueText == "" {
			valueText = "demo-value"
		}
		result := s.StoreValue(key, []byte(valueText))
		return networkActionResult(
			"已将键值写入最近的 K 个节点。",
			result,
			&types.ActionFeedback{
				Summary:     "数据已经根据键的哈希分布到距离最近的节点集合中。",
				NextHint:    "继续执行节点查找，观察目标键如何映射到对应的 K 桶和最近邻。",
				EffectScope: "network",
				ResultState: result,
			},
		), nil
	default:
		return nil, fmt.Errorf("unsupported kademlia action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// KademliaFactory Kademlia工厂
type KademliaFactory struct{}

// Create 创建演示器
func (f *KademliaFactory) Create() engine.Simulator {
	return NewKademliaSimulator()
}

// GetDescription 获取描述
func (f *KademliaFactory) GetDescription() types.Description {
	return NewKademliaSimulator().GetDescription()
}

// NewKademliaFactory 创建工厂
func NewKademliaFactory() *KademliaFactory {
	return &KademliaFactory{}
}

var _ engine.SimulatorFactory = (*KademliaFactory)(nil)
