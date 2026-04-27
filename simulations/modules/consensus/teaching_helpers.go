package consensus

import (
	"fmt"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/types"
)

func consensusActionResult(message string, data map[string]interface{}, feedback *types.ActionFeedback) *types.ActionResult {
	return &types.ActionResult{
		Success:  true,
		Message:  message,
		Data:     data,
		Feedback: feedback,
	}
}

func buildConsensusLinkedEffects(sim *base.BaseSimulator) []types.LinkedEffect {
	effects := make([]types.LinkedEffect, 0)
	for _, fault := range sim.GetActiveFaults() {
		effects = append(effects, types.LinkedEffect{
			ID:       fault.ID,
			Scope:    "consensus",
			Target:   string(fault.Target),
			Summary:  fmt.Sprintf("故障 %s 正在影响节点可用性、投票传播或提交推进。", fault.Type),
			Severity: "high",
			Blocking: false,
			Metrics:  map[string]interface{}{"kind": "fault", "type": string(fault.Type)},
		})
	}
	for _, attack := range sim.GetActiveAttacks() {
		effects = append(effects, types.LinkedEffect{
			ID:       attack.ID,
			Scope:    "consensus",
			Target:   attack.Target,
			Summary:  fmt.Sprintf("攻击 %s 正在干扰共识投票、链分叉或链头选择。", attack.Type),
			Severity: "high",
			Blocking: false,
			Metrics:  map[string]interface{}{"kind": "attack", "type": string(attack.Type)},
		})
	}
	return effects
}

func setConsensusTeachingState(
	sim *base.BaseSimulator,
	stage string,
	summary string,
	nextHint string,
	progress float64,
	result map[string]interface{},
) {
	sim.SetLinkedEffects(buildConsensusLinkedEffects(sim))
	sim.SetProcessFeedback(&types.ProcessFeedback{
		Stage:    stage,
		Summary:  summary,
		NextHint: nextHint,
		Progress: progress,
		Result:   result,
	})
}
