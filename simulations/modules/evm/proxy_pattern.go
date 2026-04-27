package evm

import (
	"encoding/hex"
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"golang.org/x/crypto/sha3"
)

// =============================================================================
// 代理模式演示器
// =============================================================================

// ProxyType 代理类型
type ProxyType string

const (
	ProxyTypeTransparent ProxyType = "Transparent"
	ProxyTypeUUPS        ProxyType = "UUPS"
	ProxyTypeBeacon      ProxyType = "Beacon"
	ProxyTypeDiamond     ProxyType = "Diamond"
	ProxyTypeMinimal     ProxyType = "Minimal"
)

// ProxyInfo 代理信息
type ProxyInfo struct {
	Type               ProxyType `json:"type"`
	ProxyAddress       string    `json:"proxy_address"`
	ImplementationSlot string    `json:"implementation_slot"`
	AdminSlot          string    `json:"admin_slot,omitempty"`
	BeaconSlot         string    `json:"beacon_slot,omitempty"`
}

// ProxyPatternSimulator 代理模式演示器
// 演示各种代理合约模式:
// - Transparent Proxy
// - UUPS (Universal Upgradeable Proxy Standard)
// - Beacon Proxy
// - Diamond (EIP-2535)
// - Minimal Proxy (EIP-1167)
type ProxyPatternSimulator struct {
	*base.BaseSimulator
	proxies []*ProxyInfo
}

// NewProxyPatternSimulator 创建代理模式演示器
func NewProxyPatternSimulator() *ProxyPatternSimulator {
	sim := &ProxyPatternSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"proxy_pattern",
			"代理模式演示器",
			"演示Transparent、UUPS、Beacon、Diamond等代理模式",
			"evm",
			types.ComponentDemo,
		),
		proxies: make([]*ProxyInfo, 0),
	}

	return sim
}

// Init 初始化
func (s *ProxyPatternSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// CalculateEIP1967Slots 计算EIP-1967标准存储槽
func (s *ProxyPatternSimulator) CalculateEIP1967Slots() map[string]string {
	// EIP-1967定义的标准存储槽
	// 使用keccak256(label) - 1来避免与正常存储冲突

	slots := make(map[string]string)

	// Implementation slot: keccak256("eip1967.proxy.implementation") - 1
	implLabel := "eip1967.proxy.implementation"
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(implLabel))
	implHash := h.Sum(nil)
	implSlot := subtractOne(implHash)
	slots["implementation"] = "0x" + hex.EncodeToString(implSlot)

	// Admin slot: keccak256("eip1967.proxy.admin") - 1
	adminLabel := "eip1967.proxy.admin"
	h = sha3.NewLegacyKeccak256()
	h.Write([]byte(adminLabel))
	adminHash := h.Sum(nil)
	adminSlot := subtractOne(adminHash)
	slots["admin"] = "0x" + hex.EncodeToString(adminSlot)

	// Beacon slot: keccak256("eip1967.proxy.beacon") - 1
	beaconLabel := "eip1967.proxy.beacon"
	h = sha3.NewLegacyKeccak256()
	h.Write([]byte(beaconLabel))
	beaconHash := h.Sum(nil)
	beaconSlot := subtractOne(beaconHash)
	slots["beacon"] = "0x" + hex.EncodeToString(beaconSlot)

	s.EmitEvent("eip1967_slots_calculated", "", "", map[string]interface{}{
		"implementation": slots["implementation"],
		"admin":          slots["admin"],
		"beacon":         slots["beacon"],
	})

	return slots
}

// subtractOne 从32字节数组中减1
func subtractOne(data []byte) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	borrow := byte(1)
	for i := len(result) - 1; i >= 0 && borrow > 0; i-- {
		if result[i] >= borrow {
			result[i] -= borrow
			borrow = 0
		} else {
			result[i] = 255
		}
	}
	return result
}

// ExplainTransparentProxy 解释透明代理
func (s *ProxyPatternSimulator) ExplainTransparentProxy() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Transparent Proxy Pattern",
		"eip":         "EIP-1967",
		"description": "管理员和用户使用不同的调用路径",
		"components": map[string]string{
			"Proxy":          "存储状态，转发调用",
			"ProxyAdmin":     "管理升级的独立合约",
			"Implementation": "业务逻辑",
		},
		"mechanism": []string{
			"1. 如果调用者是admin，执行代理管理函数",
			"2. 如果调用者不是admin，delegatecall到implementation",
			"3. admin不能调用业务函数，避免选择器冲突",
		},
		"storage_slots": s.CalculateEIP1967Slots(),
		"pros": []string{
			"避免函数选择器冲突",
			"升级逻辑清晰",
			"广泛使用和审计",
		},
		"cons": []string{
			"每次调用都要检查caller",
			"需要额外的ProxyAdmin合约",
			"gas成本略高",
		},
		"code_example": `// TransparentUpgradeableProxy
fallback() external payable {
    if (msg.sender == admin) {
        // 管理员调用代理函数
        _dispatchAdmin();
    } else {
        // 普通用户delegatecall到implementation
        _delegate(implementation);
    }
}`,
	}
}

// ExplainUUPSProxy 解释UUPS代理
func (s *ProxyPatternSimulator) ExplainUUPSProxy() map[string]interface{} {
	return map[string]interface{}{
		"name":        "UUPS (Universal Upgradeable Proxy Standard)",
		"eip":         "EIP-1822",
		"description": "升级逻辑在Implementation中而非Proxy",
		"components": map[string]string{
			"Proxy":          "最小代理，只做delegatecall",
			"Implementation": "包含业务逻辑和升级函数",
		},
		"mechanism": []string{
			"1. Proxy总是delegatecall到implementation",
			"2. 升级函数(upgradeTo)在implementation中",
			"3. 由于是delegatecall，升级函数修改proxy的存储",
		},
		"pros": []string{
			"Proxy合约更小，部署成本低",
			"没有选择器冲突问题",
			"gas效率更高",
		},
		"cons": []string{
			"如果implementation没有升级函数，将永远无法升级",
			"升级函数需要访问控制",
			"需要确保每个implementation都包含升级逻辑",
		},
		"code_example": `// UUPSUpgradeable Implementation
function upgradeTo(address newImplementation) external onlyOwner {
    _authorizeUpgrade(newImplementation);
    // 这里修改的是proxy的存储
    StorageSlot.getAddressSlot(_IMPLEMENTATION_SLOT).value = newImplementation;
}`,
	}
}

// ExplainBeaconProxy 解释Beacon代理
func (s *ProxyPatternSimulator) ExplainBeaconProxy() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Beacon Proxy Pattern",
		"eip":         "EIP-1967",
		"description": "多个代理共享同一个Beacon，实现批量升级",
		"components": map[string]string{
			"Beacon":         "存储implementation地址",
			"BeaconProxy":    "指向Beacon",
			"Implementation": "业务逻辑",
		},
		"mechanism": []string{
			"1. 多个BeaconProxy指向同一个Beacon",
			"2. BeaconProxy每次调用时从Beacon读取implementation",
			"3. 升级Beacon即可升级所有BeaconProxy",
		},
		"pros": []string{
			"一次升级影响所有代理",
			"适合大量相同合约的场景",
			"如工厂创建的用户合约",
		},
		"cons": []string{
			"每次调用需要额外读取Beacon",
			"无法单独升级某个代理",
			"Beacon是中心化点",
		},
		"use_cases": []string{
			"NFT集合中每个NFT的独立合约",
			"每用户一个金库合约",
			"批量部署的策略合约",
		},
	}
}

// ExplainDiamondProxy 解释Diamond代理
func (s *ProxyPatternSimulator) ExplainDiamondProxy() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Diamond Pattern (Multi-Facet Proxy)",
		"eip":         "EIP-2535",
		"description": "一个代理可以有多个implementation(Facet)",
		"components": map[string]string{
			"Diamond":    "代理合约，路由到不同Facet",
			"Facets":     "多个逻辑合约",
			"DiamondCut": "管理Facet的添加/替换/删除",
			"Loupe":      "查询Diamond结构",
		},
		"mechanism": []string{
			"1. Diamond维护selector->facet的映射",
			"2. 根据calldata的函数选择器路由到对应Facet",
			"3. 可以添加、替换、删除任意函数",
		},
		"pros": []string{
			"突破24KB合约大小限制",
			"模块化升级，可以只升级部分功能",
			"一个地址提供多个合约的功能",
		},
		"cons": []string{
			"复杂度高",
			"存储管理需要特别注意",
			"审计难度大",
		},
		"storage_pattern": "Diamond Storage - 每个Facet使用唯一的存储结构",
	}
}

// ExplainMinimalProxy 解释最小代理
func (s *ProxyPatternSimulator) ExplainMinimalProxy() map[string]interface{} {
	// EIP-1167 Minimal Proxy字节码
	minimalBytecode := "363d3d373d3d3d363d73" + "bebebebebebebebebebebebebebebebebebebebe" + "5af43d82803e903d91602b57fd5bf3"

	return map[string]interface{}{
		"name":        "Minimal Proxy (Clone)",
		"eip":         "EIP-1167",
		"description": "最小化的代理合约，用于低成本克隆",
		"bytecode":    "0x" + minimalBytecode,
		"size":        "45 bytes",
		"deploy_cost": "约10000 gas (vs 32000+ for CREATE)",
		"mechanism": []string{
			"1. 固定的字节码，只需替换implementation地址",
			"2. 无法升级implementation",
			"3. 纯粹的delegatecall转发",
		},
		"pros": []string{
			"部署成本极低",
			"字节码小且固定",
			"适合大量部署相同合约",
		},
		"cons": []string{
			"不可升级",
			"每次调用有额外gas开销",
		},
		"use_cases": []string{
			"Gnosis Safe多签钱包",
			"Uniswap V2 Pair合约",
			"NFT工厂",
		},
	}
}

// CompareProxyPatterns 对比代理模式
func (s *ProxyPatternSimulator) CompareProxyPatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"pattern":       "Transparent",
			"upgradeable":   true,
			"complexity":    "中等",
			"gas_overhead":  "中等",
			"upgrade_scope": "单个代理",
		},
		{
			"pattern":       "UUPS",
			"upgradeable":   true,
			"complexity":    "低",
			"gas_overhead":  "低",
			"upgrade_scope": "单个代理",
		},
		{
			"pattern":       "Beacon",
			"upgradeable":   true,
			"complexity":    "中等",
			"gas_overhead":  "高",
			"upgrade_scope": "批量代理",
		},
		{
			"pattern":       "Diamond",
			"upgradeable":   true,
			"complexity":    "高",
			"gas_overhead":  "中等",
			"upgrade_scope": "模块化",
		},
		{
			"pattern":       "Minimal",
			"upgradeable":   false,
			"complexity":    "低",
			"gas_overhead":  "低",
			"upgrade_scope": "N/A",
		},
	}
}

// updateState 更新状态
func (s *ProxyPatternSimulator) updateState() {
	s.SetGlobalData("proxy_count", len(s.proxies))
	s.SetGlobalData("patterns", []string{"Transparent", "UUPS", "Beacon", "Diamond", "Minimal"})

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		"pattern_ready",
		"代理模式实验已就绪，可以查看不同代理模式的结构、升级范围和风险点。",
		"建议先比较 Transparent 与 UUPS，再理解 Beacon、Diamond 和 Minimal Proxy 的适用场景。",
		0.3,
		map[string]interface{}{"patterns": []string{"Transparent", "UUPS", "Beacon", "Diamond", "Minimal"}},
	)
}

// ExecuteAction 为代理模式实验提供交互动作。
func (s *ProxyPatternSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "explain_transparent":
		result := s.ExplainTransparentProxy()
		return evmActionResult("已加载 Transparent Proxy 说明。", result, &types.ActionFeedback{
			Summary:     "Transparent Proxy 的调用分流与管理逻辑已经展开。",
			NextHint:    "重点观察 admin 和普通用户为何会走不同路径，以及它如何避免选择器冲突。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	case "explain_uups":
		result := s.ExplainUUPSProxy()
		return evmActionResult("已加载 UUPS 代理说明。", result, &types.ActionFeedback{
			Summary:     "UUPS 代理的升级逻辑已经展开。",
			NextHint:    "重点观察为什么升级逻辑被放进实现合约，以及这种方式的风险边界。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	case "explain_beacon":
		result := s.ExplainBeaconProxy()
		return evmActionResult("已加载 Beacon 代理说明。", result, &types.ActionFeedback{
			Summary:     "Beacon 代理的批量升级路径已经展开。",
			NextHint:    "重点观察 Beacon 作为中间层如何同时影响多个代理实例。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	case "explain_diamond":
		result := s.ExplainDiamondProxy()
		return evmActionResult("已加载 Diamond 代理说明。", result, &types.ActionFeedback{
			Summary:     "Diamond 多 Facet 路由结构已经展开。",
			NextHint:    "重点观察 selector 到 facet 的映射关系，以及它如何突破单合约体积限制。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	case "explain_minimal":
		result := s.ExplainMinimalProxy()
		return evmActionResult("已加载 Minimal Proxy 说明。", result, &types.ActionFeedback{
			Summary:     "Minimal Proxy 的克隆机制已经展开。",
			NextHint:    "重点观察它为什么部署成本低，以及为什么不能原生升级。",
			EffectScope: "evm",
			ResultState: result,
		}), nil
	case "compare_patterns":
		result := s.CompareProxyPatterns()
		return evmActionResult("已生成代理模式对比。", map[string]interface{}{"patterns": result}, &types.ActionFeedback{
			Summary:     "代理模式的升级范围、复杂度和 gas 开销差异已经整理完成。",
			NextHint:    "继续结合具体模式说明，判断不同业务场景该选择哪一类代理。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"count": len(result)},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported proxy pattern action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// ProxyPatternFactory 代理模式工厂
type ProxyPatternFactory struct{}

func (f *ProxyPatternFactory) Create() engine.Simulator {
	return NewProxyPatternSimulator()
}

func (f *ProxyPatternFactory) GetDescription() types.Description {
	return NewProxyPatternSimulator().GetDescription()
}

func NewProxyPatternFactory() *ProxyPatternFactory {
	return &ProxyPatternFactory{}
}

var _ engine.SimulatorFactory = (*ProxyPatternFactory)(nil)
