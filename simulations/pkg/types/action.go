package types

// ActionFeedback 描述教学动作执行后的直接反馈。
// 这些字段主要服务前端主舞台，让学生在执行动作后马上看到：
// 1. 当前动作做了什么
// 2. 下一步应该观察什么
// 3. 这次动作主要影响了哪一部分实验过程
type ActionFeedback struct {
	Summary     string                 `json:"summary,omitempty"`
	NextHint    string                 `json:"next_hint,omitempty"`
	EffectScope string                 `json:"effect_scope,omitempty"`
	ResultState map[string]interface{} `json:"result_state,omitempty"`
}

// ActionResult 描述模块动作执行后的结果。
// Data 保留给各模块返回结构化状态快照；
// Feedback 用于承载统一的教学反馈信息。
type ActionResult struct {
	Success  bool                   `json:"success"`
	Message  string                 `json:"message,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Feedback *ActionFeedback        `json:"feedback,omitempty"`
}

// ActionHandler 允许模拟器暴露模块级教学动作。
// 这类动作用于触发“发起攻击、模拟交易、切换轮次、注入步骤”等课堂交互。
type ActionHandler interface {
	ExecuteAction(action string, params map[string]interface{}) (*ActionResult, error)
}
