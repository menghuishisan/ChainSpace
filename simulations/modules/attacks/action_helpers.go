package attacks

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/chainspace/simulations/modules/base"
	"github.com/chainspace/simulations/pkg/types"
)

type resettableAttackSimulator interface {
	Reset() error
	Init(config types.Config) error
	GetParams() map[string]types.Param
}

// actionInt 从动作参数中读取整数，缺失时返回默认值。
func actionInt(params map[string]interface{}, key string, fallback int) int {
	if params == nil {
		return fallback
	}

	value, ok := params[key]
	if !ok || value == nil {
		return fallback
	}

	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(typed)
		if err == nil {
			return parsed
		}
	}

	return fallback
}

// actionInt64 从动作参数中读取 int64，缺失时返回默认值。
func actionInt64(params map[string]interface{}, key string, fallback int64) int64 {
	if params == nil {
		return fallback
	}

	value, ok := params[key]
	if !ok || value == nil {
		return fallback
	}

	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float32:
		return int64(typed)
	case float64:
		return int64(typed)
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err == nil {
			return parsed
		}
	}

	return fallback
}

// actionFloat64 从动作参数中读取浮点数，缺失时返回默认值。
func actionFloat64(params map[string]interface{}, key string, fallback float64) float64 {
	if params == nil {
		return fallback
	}

	value, ok := params[key]
	if !ok || value == nil {
		return fallback
	}

	switch typed := value.(type) {
	case int:
		return float64(typed)
	case int8:
		return float64(typed)
	case int16:
		return float64(typed)
	case int32:
		return float64(typed)
	case int64:
		return float64(typed)
	case float32:
		return float64(typed)
	case float64:
		return typed
	case string:
		parsed, err := strconv.ParseFloat(typed, 64)
		if err == nil {
			return parsed
		}
	}

	return fallback
}

// actionString 从动作参数中读取字符串，缺失时返回默认值。
func actionString(params map[string]interface{}, key string, fallback string) string {
	if params == nil {
		return fallback
	}

	value, ok := params[key]
	if !ok || value == nil {
		return fallback
	}

	if typed, ok := value.(string); ok && typed != "" {
		return typed
	}

	return fallback
}

// actionBool 从动作参数中读取布尔值，缺失时返回默认值。
func actionBool(params map[string]interface{}, key string, fallback bool) bool {
	if params == nil {
		return fallback
	}

	value, ok := params[key]
	if !ok || value == nil {
		return fallback
	}

	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		if typed == "true" {
			return true
		}
		if typed == "false" {
			return false
		}
	}

	return fallback
}

// actionResult 构造统一的动作执行结果。
func actionResult(message string, data map[string]interface{}) *types.ActionResult {
	return &types.ActionResult{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// actionResultWithFeedback 构造包含教学反馈的动作结果。
func actionResultWithFeedback(message string, data map[string]interface{}, feedback *types.ActionFeedback) *types.ActionResult {
	return &types.ActionResult{
		Success:  true,
		Message:  message,
		Data:     data,
		Feedback: feedback,
	}
}

// resetAttackScene 使用当前参数重新初始化攻击场景。
func resetAttackScene(sim resettableAttackSimulator) (*types.ActionResult, error) {
	params := sim.GetParams()
	config := types.Config{
		Params: make(map[string]interface{}, len(params)),
	}

	for key, param := range params {
		if param.Value != nil {
			config.Params[key] = param.Value
			continue
		}
		config.Params[key] = param.Default
	}

	if err := sim.Reset(); err != nil {
		return nil, err
	}

	if err := sim.Init(config); err != nil {
		return nil, err
	}

	return actionResultWithFeedback(
		"已重置当前攻击场景。",
		nil,
		&types.ActionFeedback{
			Summary:     "攻击场景已恢复到初始状态。",
			NextHint:    "可以重新选择攻击动作，观察过程和结果如何变化。",
			EffectScope: "attack",
		},
	), nil
}

// buildLinkedEffects 把当前激活的攻击和故障转换为前端可消费的联动影响。
func buildLinkedEffects(sim *base.BaseSimulator, scope string) []types.LinkedEffect {
	effects := make([]types.LinkedEffect, 0)

	for _, fault := range sim.GetActiveFaults() {
		effects = append(effects, types.LinkedEffect{
			ID:       fault.ID,
			Scope:    scope,
			Target:   string(fault.Target),
			Summary:  fmt.Sprintf("故障 %s 正在影响 %s。", fault.Type, fault.Target),
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
			Summary:  fmt.Sprintf("攻击 %s 正在影响 %s。", attack.Type, attack.Target),
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

// setAttackTeachingState 为攻击类实验统一导出过程反馈和联动影响。
func setAttackTeachingState(
	sim *base.BaseSimulator,
	scope string,
	stage string,
	summary string,
	nextHint string,
	progress float64,
	result map[string]interface{},
) {
	sim.SetLinkedEffects(buildLinkedEffects(sim, scope))
	sim.SetProcessFeedback(&types.ProcessFeedback{
		Stage:    stage,
		Summary:  summary,
		NextHint: nextHint,
		Progress: progress,
		Result:   result,
	})
}

// formatEther 将 wei 近似格式化为 ETH 文本。
func formatEther(value *big.Int) string {
	if value == nil {
		return "0.0000"
	}

	ether := new(big.Float).Quo(
		new(big.Float).SetInt(value),
		new(big.Float).SetInt64(1e18),
	)
	return fmt.Sprintf("%.4f", ether)
}

// formatBigInt 将大整数格式化为字符串。
func formatBigInt(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}

// formatTokenAmount 将代币数量格式化为便于展示的文本。
func formatTokenAmount(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}

var leadingNumberPattern = regexp.MustCompile(`[-+]?\d+(\.\d+)?`)

// parseAmountText 从描述性金额文本中提取首个数字，便于前端做教学可视化。
func parseAmountText(value string) float64 {
	match := leadingNumberPattern.FindString(value)
	if match == "" {
		return 0
	}

	parsed, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0
	}
	return parsed
}
