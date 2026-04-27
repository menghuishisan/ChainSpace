package blockchain

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
)

// GasOperation Gas操作
type GasOperation struct {
	Name     string `json:"name"`
	GasCost  uint64 `json:"gas_cost"`
	Category string `json:"category"`
}

// GasSimulator Gas机制演示器
type GasSimulator struct {
	*base.BaseSimulator
	operations  map[string]*GasOperation
	gasPrice    uint64
	gasLimit    uint64
	gasUsed     uint64
	execHistory []map[string]interface{}
}

// NewGasSimulator 创建Gas演示器
func NewGasSimulator() *GasSimulator {
	sim := &GasSimulator{
		BaseSimulator: base.NewBaseSimulator(
			"gas",
			"Gas机制演示器",
			"展示以太坊Gas计费机制和EIP-1559",
			"blockchain",
			types.ComponentTool,
		),
		operations:  make(map[string]*GasOperation),
		gasPrice:    20,
		gasLimit:    100000,
		execHistory: make([]map[string]interface{}, 0),
	}

	sim.AddParam(types.Param{
		Key: "gas_price", Name: "Gas价格(Gwei)", Type: types.ParamTypeInt,
		Default: 20, Min: 1, Max: 500,
	})
	sim.AddParam(types.Param{
		Key: "gas_limit", Name: "Gas上限", Type: types.ParamTypeInt,
		Default: 100000, Min: 21000, Max: 30000000,
	})
	return sim
}

// Init 初始化
func (s *GasSimulator) Init(config types.Config) error {
	if err := s.BaseSimulator.Init(config); err != nil {
		return err
	}

	if v, ok := config.Params["gas_price"]; ok {
		if n, ok := v.(float64); ok {
			s.gasPrice = uint64(n)
		}
	}
	if v, ok := config.Params["gas_limit"]; ok {
		if n, ok := v.(float64); ok {
			s.gasLimit = uint64(n)
		}
	}

	s.operations = map[string]*GasOperation{
		"ADD":      {Name: "ADD", GasCost: 3, Category: "arithmetic"},
		"MUL":      {Name: "MUL", GasCost: 5, Category: "arithmetic"},
		"SUB":      {Name: "SUB", GasCost: 3, Category: "arithmetic"},
		"DIV":      {Name: "DIV", GasCost: 5, Category: "arithmetic"},
		"SLOAD":    {Name: "SLOAD", GasCost: 800, Category: "storage"},
		"SSTORE":   {Name: "SSTORE", GasCost: 20000, Category: "storage"},
		"CALL":     {Name: "CALL", GasCost: 700, Category: "call"},
		"CREATE":   {Name: "CREATE", GasCost: 32000, Category: "contract"},
		"TRANSFER": {Name: "TRANSFER", GasCost: 21000, Category: "transfer"},
		"LOG":      {Name: "LOG", GasCost: 375, Category: "log"},
	}

	s.updateState()
	return nil
}

// ExecuteOperation 执行操作
func (s *GasSimulator) ExecuteOperation(opName string) error {
	op := s.operations[opName]
	if op == nil {
		return fmt.Errorf("unknown operation: %s", opName)
	}

	if s.gasUsed+op.GasCost > s.gasLimit {
		s.EmitEvent("out_of_gas", "", "", map[string]interface{}{
			"operation": opName, "required": op.GasCost, "remaining": s.gasLimit - s.gasUsed,
		})
		return fmt.Errorf("out of gas")
	}

	s.gasUsed += op.GasCost
	cost := op.GasCost * s.gasPrice

	s.execHistory = append(s.execHistory, map[string]interface{}{
		"operation": opName, "gas_cost": op.GasCost, "eth_cost": cost,
	})

	s.EmitEvent("operation_executed", "", "", map[string]interface{}{
		"operation": opName, "gas_cost": op.GasCost, "gas_used": s.gasUsed,
	})
	s.updateState()
	return nil
}

// EstimateGas 估算Gas
func (s *GasSimulator) EstimateGas(operations []string) uint64 {
	var total uint64
	for _, opName := range operations {
		if op := s.operations[opName]; op != nil {
			total += op.GasCost
		}
	}
	return total
}

// Reset 重置
func (s *GasSimulator) Reset() error {
	s.gasUsed = 0
	s.execHistory = make([]map[string]interface{}, 0)
	s.updateState()
	return s.BaseSimulator.Reset()
}

func (s *GasSimulator) updateState() {
	s.SetGlobalData("gas_price", s.gasPrice)
	s.SetGlobalData("gas_limit", s.gasLimit)
	s.SetGlobalData("gas_used", s.gasUsed)
	s.SetGlobalData("gas_remaining", s.gasLimit-s.gasUsed)
	s.SetGlobalData("operations", s.operations)
	s.SetGlobalData("exec_history", s.execHistory)

	summary := fmt.Sprintf("当前 Gas 价格为 %d Gwei，已消耗 %d / %d Gas。", s.gasPrice, s.gasUsed, s.gasLimit)
	nextHint := "可以继续执行操作，观察不同指令的 Gas 成本和剩余额度如何变化。"
	setBlockchainTeachingState(
		s.BaseSimulator,
		"blockchain",
		"准备 Gas 计费",
		summary,
		nextHint,
		0.45,
		map[string]interface{}{"gas_used": s.gasUsed, "gas_limit": s.gasLimit, "gas_price": s.gasPrice},
	)
}

func (s *GasSimulator) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	switch action {
	case "execute_operation":
		opName := "SSTORE"
		if raw, ok := params["operation"].(string); ok && raw != "" {
			opName = raw
		}
		if err := s.ExecuteOperation(opName); err != nil {
			return nil, err
		}
		return blockchainActionResult("已执行一次 Gas 计费操作。", map[string]interface{}{"operation": opName, "gas_used": s.gasUsed}, &types.ActionFeedback{
			Summary:     "当前操作的 Gas 已计入执行历史。",
			NextHint:    "继续比较不同操作的 Gas 消耗，并观察何时触发 out of gas。",
			EffectScope: "blockchain",
			ResultState: map[string]interface{}{"operation": opName, "gas_used": s.gasUsed},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported gas action: %s", action)
	}
}

type GasFactory struct{}

func (f *GasFactory) Create() engine.Simulator          { return NewGasSimulator() }
func (f *GasFactory) GetDescription() types.Description { return NewGasSimulator().GetDescription() }
func NewGasFactory() *GasFactory                        { return &GasFactory{} }

var _ engine.SimulatorFactory = (*GasFactory)(nil)
