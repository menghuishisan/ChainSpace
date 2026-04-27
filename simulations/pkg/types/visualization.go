package types

// DisturbanceSnapshot 描述当前正在生效的联动覆盖层。
// 前端通过该结构直接展示“故障/攻击作用在谁身上、当前影响是什么”。
type DisturbanceSnapshot struct {
	ID       string                 `json:"id"`
	Kind     string                 `json:"kind"`
	Type     string                 `json:"type"`
	Target   string                 `json:"target"`
	Label    string                 `json:"label"`
	Summary  string                 `json:"summary"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Duration uint64                 `json:"duration,omitempty"`
}

// LinkedEffect 描述联动覆盖层对主实验过程造成的具体影响。
// 例如“投票不足”“价格偏移”“桥验证绕过”“传播路径断裂”等。
type LinkedEffect struct {
	ID        string                 `json:"id,omitempty"`
	Scope     string                 `json:"scope"`
	Target    string                 `json:"target,omitempty"`
	Summary   string                 `json:"summary"`
	Severity  string                 `json:"severity,omitempty"`
	Blocking  bool                   `json:"blocking,omitempty"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

// ProcessFeedback 描述当前主过程的教学状态。
// 它不是原始模拟状态，而是为了前端主舞台提供统一的过程解释。
type ProcessFeedback struct {
	Stage    string                 `json:"stage,omitempty"`
	Summary  string                 `json:"summary,omitempty"`
	NextHint string                 `json:"next_hint,omitempty"`
	Progress float64                `json:"progress,omitempty"`
	Result   map[string]interface{} `json:"result,omitempty"`
}
