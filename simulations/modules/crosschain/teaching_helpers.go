package crosschain

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/types"
)

func crosschainActionResult(message string, data map[string]interface{}, feedback *types.ActionFeedback) *types.ActionResult {
	return &types.ActionResult{
		Success:  true,
		Message:  message,
		Data:     data,
		Feedback: feedback,
	}
}

func buildCrosschainLinkedEffects(sim *base.BaseSimulator, scope string) []types.LinkedEffect {
	effects := make([]types.LinkedEffect, 0)
	for _, fault := range sim.GetActiveFaults() {
		effects = append(effects, types.LinkedEffect{
			ID:       fault.ID,
			Scope:    scope,
			Target:   string(fault.Target),
			Summary:  fmt.Sprintf("故障 %s 正在影响跨链消息确认、证明生成或目标链执行。", fault.Type),
			Severity: "high",
			Blocking: true,
			Metrics:  map[string]interface{}{"kind": "fault", "type": string(fault.Type)},
		})
	}
	for _, attack := range sim.GetActiveAttacks() {
		effects = append(effects, types.LinkedEffect{
			ID:       attack.ID,
			Scope:    scope,
			Target:   attack.Target,
			Summary:  fmt.Sprintf("攻击 %s 正在尝试改写跨链验证、签名聚合或目标链执行结果。", attack.Type),
			Severity: "high",
			Blocking: true,
			Metrics:  map[string]interface{}{"kind": "attack", "type": string(attack.Type)},
		})
	}
	return effects
}

func setCrosschainTeachingState(
	sim *base.BaseSimulator,
	scope string,
	stage string,
	summary string,
	nextHint string,
	progress float64,
	result map[string]interface{},
) {
	sim.SetLinkedEffects(buildCrosschainLinkedEffects(sim, scope))
	sim.SetProcessFeedback(&types.ProcessFeedback{
		Stage:    stage,
		Summary:  summary,
		NextHint: nextHint,
		Progress: progress,
		Result:   result,
	})
}
