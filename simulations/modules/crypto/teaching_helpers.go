package crypto

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/types"
)

func cryptoActionResult(message string, data map[string]interface{}, feedback *types.ActionFeedback) *types.ActionResult {
	return &types.ActionResult{
		Success:  true,
		Message:  message,
		Data:     data,
		Feedback: feedback,
	}
}

func buildCryptoLinkedEffects(sim *base.BaseSimulator, scope string) []types.LinkedEffect {
	effects := make([]types.LinkedEffect, 0)
	for _, fault := range sim.GetActiveFaults() {
		effects = append(effects, types.LinkedEffect{
			ID:       fault.ID,
			Scope:    scope,
			Target:   string(fault.Target),
			Summary:  fmt.Sprintf("故障 %s 正在影响 %s 的密码学过程。", fault.Type, fault.Target),
			Severity: "medium",
			Blocking: false,
			Metrics: map[string]interface{}{
				"kind": "fault",
				"type": string(fault.Type),
			},
		})
	}
	for _, attack := range sim.GetActiveAttacks() {
		effects = append(effects, types.LinkedEffect{
			ID:       attack.ID,
			Scope:    scope,
			Target:   attack.Target,
			Summary:  fmt.Sprintf("攻击 %s 正在改写 %s 的密码学结果或验证路径。", attack.Type, attack.Target),
			Severity: "high",
			Blocking: false,
			Metrics: map[string]interface{}{
				"kind": "attack",
				"type": string(attack.Type),
			},
		})
	}
	return effects
}

func setCryptoTeachingState(
	sim *base.BaseSimulator,
	scope string,
	stage string,
	summary string,
	nextHint string,
	progress float64,
	result map[string]interface{},
) {
	sim.SetLinkedEffects(buildCryptoLinkedEffects(sim, scope))
	sim.SetProcessFeedback(&types.ProcessFeedback{
		Stage:    stage,
		Summary:  summary,
		NextHint: nextHint,
		Progress: progress,
		Result:   result,
	})
}
