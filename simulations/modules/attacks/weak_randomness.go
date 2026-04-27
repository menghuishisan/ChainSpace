package attacks

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// RandomnessSource 描述一种随机源的安全性特征。
type RandomnessSource struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Predictable bool   `json:"predictable"`
	Manipulable bool   `json:"manipulable"`
}

// WeakRandomnessSimulator 演示链上弱随机数带来的攻击面。
type WeakRandomnessSimulator struct {
	*base.BaseSimulator
	records []map[string]interface{}
}

// NewWeakRandomnessSimulator 创建弱随机数模拟器。
func NewWeakRandomnessSimulator() *WeakRandomnessSimulator {
	return &WeakRandomnessSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"weak_randomness",
			"弱随机数攻击演示器",
			"演示 block.timestamp、blockhash、私有种子等弱随机源如何被预测或操纵，并展示更安全的随机数方案。",
			"attacks",
			types.ComponentAttack,
		),
		records: make([]map[string]interface{}, 0),
	}
}

// Init 初始化基础状态。
func (s *WeakRandomnessSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}
	s.records = make([]map[string]interface{}, 0)
	s.updateState()
	return nil
}

// ShowUnsafeSources 返回常见不安全随机源。
func (s *WeakRandomnessSimulator) ShowUnsafeSources() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"source":   "block.timestamp",
			"code":     `uint random = uint(keccak256(abi.encode(block.timestamp))) % 100;`,
			"attack":   "矿工可以在允许范围内轻微调整时间戳，从而影响开奖结果。",
			"severity": "High",
		},
		{
			"source":     "blockhash",
			"code":       `uint random = uint(blockhash(block.number - 1)) % 100;`,
			"attack":     "区块生产者可以在收益足够高时选择是否发布候选区块。",
			"severity":   "High",
			"limitation": "只能读取最近 256 个区块的哈希，且仍可能被区块生产者影响。",
		},
		{
			"source":   "链上私有种子",
			"code":     `uint private seed; // 链上可读`,
			"attack":   "攻击者可以通过状态读取接口直接获取种子值，然后预测未来结果。",
			"severity": "Critical",
		},
		{
			"source":   "msg.sender + block.number",
			"code":     `uint random = uint(keccak256(abi.encode(msg.sender, block.number)));`,
			"attack":   "攻击者可以控制发送地址和发起时机，使随机值落入期望范围。",
			"severity": "High",
		},
	}
}

// SimulatePrediction 演示弱随机数被攻击者预测并利用。
func (s *WeakRandomnessSimulator) SimulatePrediction() map[string]interface{} {
	data := map[string]interface{}{
		"scenario": "弱随机数预测",
		"vulnerable_lottery": `function play() external payable {
    require(msg.value == 1 ether);

    uint random = uint(keccak256(abi.encode(
        block.timestamp,
        block.difficulty,
        msg.sender
    ))) % 10;

    if (random == 7) {
        payable(msg.sender).transfer(address(this).balance);
    }
}`,
		"attack_contract": `contract Attacker {
    Lottery lottery;

    function attack() external payable {
        uint prediction = uint(keccak256(abi.encode(
            block.timestamp,
            block.difficulty,
            address(this)
        ))) % 10;

        if (prediction == 7) {
            lottery.play{value: 1 ether}();
        }
    }
}`,
		"summary": "攻击者只在预测中奖时才参与，极大提高了中奖概率。",
		"flow": []string{
			"攻击者观察合约使用的伪随机输入来源。",
			"在本地复现同样的哈希计算，提前得到候选结果。",
			"只有当结果满足中奖条件时，才发送真实交易参与游戏。",
		},
	}

	s.records = append(s.records, data)
	s.EmitEvent("weak_randomness_prediction", "", "", data)
	s.updateState()
	return data
}

// ShowSecureSolutions 返回安全随机数方案。
func (s *WeakRandomnessSimulator) ShowSecureSolutions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "Chainlink VRF",
			"description": "通过链下证明和链上验证提供可校验随机数。",
			"code": `import "@chainlink/contracts/src/v0.8/VRFConsumerBaseV2.sol";

function requestRandom() external returns (uint256 requestId) {
    requestId = COORDINATOR.requestRandomWords(
        keyHash,
        subscriptionId,
        requestConfirmations,
        callbackGasLimit,
        numWords
    );
}

function fulfillRandomWords(uint256 requestId, uint256[] memory randomWords) internal override {
    // 使用随机结果
}`,
			"pros": []string{"可验证", "难以操纵", "适合高价值场景"},
			"cons": []string{"依赖预言机网络", "成本高于本地伪随机"},
		},
		{
			"name":        "Commit-Reveal",
			"description": "参与者先提交承诺，之后再揭示原始值，降低单方操纵概率。",
			"code": `// Phase 1: Commit
function commit(bytes32 hash) external {
    commits[msg.sender] = Commit(hash, block.number);
}

// Phase 2: Reveal (after N blocks)
function reveal(uint256 number, bytes32 salt) external {
    require(keccak256(abi.encode(number, salt)) == commits[msg.sender].hash);
    require(block.number > commits[msg.sender].block + 10);
    // use number
}`,
			"pros": []string{"去中心化程度较高", "适合多人参与场景"},
			"cons": []string{"需要两阶段交互", "存在未揭示时的额外处理成本"},
		},
		{
			"name":        "RANDAO / prevrandao",
			"description": "PoS 网络中可利用协议层提供的随机信号作为更安全的基础随机源。",
			"code":        `uint256 random = block.prevrandao;`,
			"pros":        []string{"无需额外预言机", "与共识过程结合紧密"},
			"cons":        []string{"仍需结合具体协议分析，不能视为完全不可操纵"},
		},
	}
}

// GetRealWorldCases 返回真实案例。
func (s *WeakRandomnessSimulator) GetRealWorldCases() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":  "SmartBillions",
			"date":  "2017",
			"loss":  "约 400 ETH",
			"issue": "攻击者利用 blockhash 设计缺陷反复尝试，最终绕过随机机制。",
		},
		{
			"name":  "Fomo3D",
			"date":  "2018",
			"issue": "链上弱随机和矿工行为被广泛讨论，说明高价值游戏尤其不能依赖伪随机。",
		},
	}
}

// updateState 更新状态。
func (s *WeakRandomnessSimulator) updateState() {
	s.SetGlobalData("attack_count", len(s.records))

	if len(s.records) == 0 {
		s.SetGlobalData("latest_attack", nil)
		s.SetGlobalData("attack_summary", "当前尚未执行弱随机数攻击，请观察攻击者如何预测随机结果并只在有利时机出手。")
		s.SetGlobalData("steps", []interface{}{})
		s.SetGlobalData("max_depth", 0)
		setAttackTeachingState(
			s.BaseSimulator,
			"execution",
			"idle",
			"等待弱随机数攻击场景。",
			"可以先触发一次预测攻击，观察伪随机输入如何被本地复算并提前利用。",
			0,
			map[string]interface{}{
				"attack_count": len(s.records),
			},
		)
		return
	}

	latest := s.records[len(s.records)-1]
	flow, _ := latest["flow"].([]string)
	steps := make([]map[string]interface{}, 0, len(flow))
	for index, item := range flow {
		steps = append(steps, map[string]interface{}{
			"index":       index + 1,
			"title":       fmt.Sprintf("步骤 %d", index+1),
			"description": item,
		})
	}

	s.SetGlobalData("latest_attack", latest)
	s.SetGlobalData("attack_summary", latest["summary"])
	s.SetGlobalData("steps", steps)
	s.SetGlobalData("max_depth", len(steps))

	setAttackTeachingState(
		s.BaseSimulator,
		"execution",
		"predicted",
		fmt.Sprintf("%v", latest["summary"]),
		"重点观察伪随机输入能否被攻击者在链下提前推演出来。",
		1.0,
		map[string]interface{}{
			"scenario":   latest["scenario"],
			"step_count": len(steps),
		},
	)
}

// ExecuteAction 执行弱随机数攻击动作。
func (s *WeakRandomnessSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "simulate_prediction":
		result := s.SimulatePrediction()
		return actionResultWithFeedback(
			"已执行弱随机数预测攻击演示。",
			map[string]interface{}{
				"result": result,
			},
			&types.ActionFeedback{
				Summary:     "已进入通过预测伪随机输入来提前筛选有利交易的攻击流程。",
				NextHint:    "重点观察攻击者为何可以在发送真实交易前就知道中奖条件是否满足。",
				EffectScope: "execution",
				ResultState: result,
			},
		), nil
	case "reset_attack":
		return resetAttackScene(s)
	default:
		return nil, fmt.Errorf("unsupported weak randomness action: %s", action)
	}
}

// WeakRandomnessFactory 创建弱随机数模拟器。
type WeakRandomnessFactory struct{}

func (f *WeakRandomnessFactory) Create() engine.Simulator { return NewWeakRandomnessSimulator() }
func (f *WeakRandomnessFactory) GetDescription() types.Description {
	return NewWeakRandomnessSimulator().GetDescription()
}
func NewWeakRandomnessFactory() *WeakRandomnessFactory { return &WeakRandomnessFactory{} }

var _ engine.SimulatorFactory = (*WeakRandomnessFactory)(nil)
