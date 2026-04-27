package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"golang.org/x/crypto/sha3"
)

// =============================================================================
// 合约创建演示器
// =============================================================================

// DeploymentInfo 部署信息
type DeploymentInfo struct {
	Method        string `json:"method"`
	DeployerAddr  string `json:"deployer_address"`
	Nonce         uint64 `json:"nonce,omitempty"`
	Salt          string `json:"salt,omitempty"`
	InitCodeHash  string `json:"init_code_hash"`
	ContractAddr  string `json:"contract_address"`
	Deterministic bool   `json:"deterministic"`
}

// CreateCreate2Simulator 合约创建演示器
// 演示CREATE和CREATE2的区别:
//
// CREATE:
//
//	address = keccak256(rlp([sender, nonce]))[12:]
//	- 地址依赖部署者地址和nonce
//	- 同一合约在不同链上地址不同
//
// CREATE2 (EIP-1014):
//
//	address = keccak256(0xff ++ sender ++ salt ++ keccak256(init_code))[12:]
//	- 地址可预先计算
//	- 相同参数在任何链上得到相同地址
type CreateCreate2Simulator struct {
	*base.BaseSimulator
	deployments []*DeploymentInfo
}

// NewCreateCreate2Simulator 创建合约创建演示器
func NewCreateCreate2Simulator() *CreateCreate2Simulator {
	sim := &CreateCreate2Simulator{
		BaseSimulator: base.NewBaseSimulator(
			"create_create2",
			"合约创建演示器",
			"演示CREATE和CREATE2的地址计算方法",
			"evm",
			types.ComponentDemo,
		),
		deployments: make([]*DeploymentInfo, 0),
	}

	return sim
}

// Init 初始化
func (s *CreateCreate2Simulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.updateState()
	return nil
}

// CalculateCreateAddress 计算CREATE地址
// address = keccak256(rlp([sender, nonce]))[12:]
func (s *CreateCreate2Simulator) CalculateCreateAddress(deployerHex string, nonce uint64) *DeploymentInfo {
	// 解析部署者地址
	deployer := parseAddress(deployerHex)

	// RLP编码: [sender, nonce]
	// 简化的RLP编码
	var rlpEncoded []byte

	// 地址长度前缀 (0x80 + 20 = 0x94)
	rlpEncoded = append(rlpEncoded, 0x94)
	rlpEncoded = append(rlpEncoded, deployer[:]...)

	// Nonce编码
	if nonce == 0 {
		rlpEncoded = append(rlpEncoded, 0x80)
	} else if nonce < 128 {
		rlpEncoded = append(rlpEncoded, byte(nonce))
	} else {
		nonceBytes := big.NewInt(int64(nonce)).Bytes()
		rlpEncoded = append(rlpEncoded, 0x80+byte(len(nonceBytes)))
		rlpEncoded = append(rlpEncoded, nonceBytes...)
	}

	// 计算总长度前缀
	totalLen := len(rlpEncoded)
	var finalRlp []byte
	if totalLen < 56 {
		finalRlp = append([]byte{0xc0 + byte(totalLen)}, rlpEncoded...)
	} else {
		lenBytes := big.NewInt(int64(totalLen)).Bytes()
		finalRlp = append([]byte{0xf7 + byte(len(lenBytes))}, lenBytes...)
		finalRlp = append(finalRlp, rlpEncoded...)
	}

	// 计算keccak256
	h := sha3.NewLegacyKeccak256()
	h.Write(finalRlp)
	hash := h.Sum(nil)

	// 取后20字节作为地址
	contractAddr := "0x" + hex.EncodeToString(hash[12:])

	info := &DeploymentInfo{
		Method:        "CREATE",
		DeployerAddr:  deployerHex,
		Nonce:         nonce,
		ContractAddr:  contractAddr,
		Deterministic: false,
	}

	s.deployments = append(s.deployments, info)

	s.EmitEvent("create_address_calculated", "", "", map[string]interface{}{
		"deployer": deployerHex,
		"nonce":    nonce,
		"address":  contractAddr,
	})

	s.updateState()
	return info
}

// CalculateCreate2Address 计算CREATE2地址
// address = keccak256(0xff ++ sender ++ salt ++ keccak256(init_code))[12:]
func (s *CreateCreate2Simulator) CalculateCreate2Address(deployerHex, saltHex, initCodeHex string) *DeploymentInfo {
	// 解析部署者地址
	deployer := parseAddress(deployerHex)

	// 解析salt (32字节)
	salt := parseSalt(saltHex)

	// 解析initCode并计算哈希
	if len(initCodeHex) > 2 && initCodeHex[:2] == "0x" {
		initCodeHex = initCodeHex[2:]
	}
	initCode, _ := hex.DecodeString(initCodeHex)

	h := sha3.NewLegacyKeccak256()
	h.Write(initCode)
	initCodeHash := h.Sum(nil)

	// 构造输入: 0xff ++ sender ++ salt ++ keccak256(init_code)
	var input []byte
	input = append(input, 0xff)
	input = append(input, deployer[:]...)
	input = append(input, salt[:]...)
	input = append(input, initCodeHash...)

	// 计算keccak256
	h = sha3.NewLegacyKeccak256()
	h.Write(input)
	hash := h.Sum(nil)

	// 取后20字节作为地址
	contractAddr := "0x" + hex.EncodeToString(hash[12:])

	info := &DeploymentInfo{
		Method:        "CREATE2",
		DeployerAddr:  deployerHex,
		Salt:          saltHex,
		InitCodeHash:  "0x" + hex.EncodeToString(initCodeHash),
		ContractAddr:  contractAddr,
		Deterministic: true,
	}

	s.deployments = append(s.deployments, info)

	s.EmitEvent("create2_address_calculated", "", "", map[string]interface{}{
		"deployer":       deployerHex,
		"salt":           saltHex,
		"init_code_hash": info.InitCodeHash,
		"address":        contractAddr,
	})

	s.updateState()
	return info
}

// CompareCreateMethods 对比CREATE和CREATE2
func (s *CreateCreate2Simulator) CompareCreateMethods() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"method":  "CREATE",
			"formula": "keccak256(rlp([sender, nonce]))[12:]",
			"inputs":  []string{"部署者地址", "部署者nonce"},
			"features": []string{
				"地址依赖nonce，每次部署不同",
				"无法预先计算地址",
				"同一合约在不同链上地址不同",
			},
			"use_cases": []string{
				"普通合约部署",
				"不需要预知地址的场景",
			},
		},
		{
			"method":  "CREATE2",
			"formula": "keccak256(0xff ++ sender ++ salt ++ keccak256(init_code))[12:]",
			"inputs":  []string{"0xff固定前缀", "部署者地址", "salt(32字节)", "initCode哈希"},
			"features": []string{
				"地址可预先计算",
				"相同参数在任何链上得到相同地址",
				"salt可用于生成不同地址",
			},
			"use_cases": []string{
				"Counterfactual部署 (先使用后部署)",
				"跨链确定性地址",
				"工厂合约模式",
				"状态通道",
				"CREATE2 + selfdestruct 可重新部署到同一地址",
			},
		},
	}
}

// SimulateFactoryPattern 模拟工厂模式
func (s *CreateCreate2Simulator) SimulateFactoryPattern(factoryAddr string, userAddr string) map[string]interface{} {
	// 模拟工厂为用户创建合约

	// 使用用户地址作为salt的一部分
	h := sha3.NewLegacyKeccak256()
	addr := parseAddress(userAddr)
	h.Write(addr[:])
	salt := "0x" + hex.EncodeToString(h.Sum(nil))

	// 示例initCode
	initCode := "0x608060405234801561001057600080fd5b50610150806100206000396000f3fe"

	result := s.CalculateCreate2Address(factoryAddr, salt, initCode)

	return map[string]interface{}{
		"factory":           factoryAddr,
		"user":              userAddr,
		"salt":              salt,
		"predicted_address": result.ContractAddr,
		"explanation":       "工厂使用用户地址派生salt，确保每个用户获得唯一但可预测的合约地址",
	}
}

// GetCreate2UseCases 获取CREATE2使用场景
func (s *CreateCreate2Simulator) GetCreate2UseCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "Counterfactual Instantiation",
			"description": "在合约部署前就使用其地址",
			"example":     "用户可以先向计算出的地址转账，之后再部署合约",
		},
		{
			"name":        "Single-use Addresses",
			"description": "创建一次性使用的地址",
			"example":     "交易所为每个用户生成唯一充值地址",
		},
		{
			"name":        "State Channels",
			"description": "状态通道中的争议解决合约",
			"example":     "只在需要时才部署争议解决合约",
		},
		{
			"name":        "Upgradeable Contracts",
			"description": "结合selfdestruct实现可升级",
			"example":     "销毁后可重新部署到同一地址(注意: EIP-6780后受限)",
		},
		{
			"name":        "Cross-chain Deployment",
			"description": "在多条链上部署到相同地址",
			"example":     "DeFi协议在所有链上使用相同地址",
		},
	}
}

// parseAddress 解析地址
func parseAddress(addrHex string) [20]byte {
	var addr [20]byte
	if len(addrHex) > 2 && addrHex[:2] == "0x" {
		addrHex = addrHex[2:]
	}
	decoded, _ := hex.DecodeString(addrHex)
	if len(decoded) > 20 {
		decoded = decoded[len(decoded)-20:]
	}
	copy(addr[20-len(decoded):], decoded)
	return addr
}

// parseSalt 解析salt
func parseSalt(saltHex string) [32]byte {
	var salt [32]byte
	if len(saltHex) > 2 && saltHex[:2] == "0x" {
		saltHex = saltHex[2:]
	}
	decoded, _ := hex.DecodeString(saltHex)
	if len(decoded) > 32 {
		decoded = decoded[len(decoded)-32:]
	}
	copy(salt[32-len(decoded):], decoded)
	return salt
}

// updateState 更新状态
func (s *CreateCreate2Simulator) updateState() {
	s.SetGlobalData("deployment_count", len(s.deployments))

	if len(s.deployments) > 0 {
		recent := s.deployments
		if len(recent) > 5 {
			recent = recent[len(recent)-5:]
		}
		s.SetGlobalData("recent_deployments", recent)
	}

	setEVMTeachingState(
		s.BaseSimulator,
		"evm",
		func() string {
			if len(s.deployments) > 0 {
				return "deployment_completed"
			}
			return "deployment_ready"
		}(),
		func() string {
			if len(s.deployments) > 0 {
				return fmt.Sprintf("当前已经生成 %d 条部署地址推导记录。", len(s.deployments))
			}
			return "当前还没有执行地址推导，可以先比较 CREATE 和 CREATE2 的差异。"
		}(),
		"重点观察 nonce、salt 和 init code 如何共同决定最终合约地址。",
		func() float64 {
			if len(s.deployments) > 0 {
				return 0.85
			}
			return 0.2
		}(),
		map[string]interface{}{"deployment_count": len(s.deployments)},
	)
}

// ExecuteAction 为 CREATE / CREATE2 实验提供交互动作。
func (s *CreateCreate2Simulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "calculate_create":
		result := s.CalculateCreateAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1)
		return evmActionResult("已计算 CREATE 地址。", map[string]interface{}{"deployment": result}, &types.ActionFeedback{
			Summary:     "CREATE 地址已经根据部署者地址和 nonce 推导完成。",
			NextHint:    "继续对比 CREATE2，观察 salt 和 init code 如何让地址变得可预测。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"address": result.ContractAddr},
		}), nil
	case "calculate_create2":
		result := s.CalculateCreate2Address("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0x01", "0x6000")
		return evmActionResult("已计算 CREATE2 地址。", map[string]interface{}{"deployment": result}, &types.ActionFeedback{
			Summary:     "CREATE2 地址已经根据 deployer、salt 和 init code 哈希推导完成。",
			NextHint:    "继续观察工厂模式如何利用 CREATE2 预先规划地址。",
			EffectScope: "evm",
			ResultState: map[string]interface{}{"address": result.ContractAddr},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported create/create2 action: %s", action)
	}
}

// =============================================================================
// 工厂
// =============================================================================

// CreateCreate2Factory 合约创建工厂
type CreateCreate2Factory struct{}

func (f *CreateCreate2Factory) Create() engine.Simulator {
	return NewCreateCreate2Simulator()
}

func (f *CreateCreate2Factory) GetDescription() types.Description {
	return NewCreateCreate2Simulator().GetDescription()
}

func NewCreateCreate2Factory() *CreateCreate2Factory {
	return &CreateCreate2Factory{}
}

var _ engine.SimulatorFactory = (*CreateCreate2Factory)(nil)
