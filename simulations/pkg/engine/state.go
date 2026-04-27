package engine

import (
	"encoding/json"
	"reflect"
	"sync"
	"time"

	"github.com/chainspace/simulations/pkg/types"
)

// StateStore 状态存储
type StateStore struct {
	mu      sync.RWMutex
	current *types.State
	history []*types.State
	maxHist int
}

// NewStateStore 创建状态存储
func NewStateStore() *StateStore {
	return &StateStore{
		current: &types.State{
			Tick:       0,
			Status:     types.StatusIdle,
			Nodes:      make(map[types.NodeID]*types.NodeState),
			GlobalData: make(map[string]interface{}),
			UpdatedAt:  time.Now(),
		},
		history: make([]*types.State, 0),
		maxHist: 1000, // 保留最近1000个状态
	}
}

// GetState 获取当前状态
func (ss *StateStore) GetState() *types.State {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.current
}

// SetState 设置当前状态
func (ss *StateStore) SetState(state *types.State) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// 保存历史
	if ss.current != nil {
		ss.history = append(ss.history, ss.current)
		if len(ss.history) > ss.maxHist {
			ss.history = ss.history[1:]
		}
	}

	state.UpdatedAt = time.Now()
	ss.current = state
}

// UpdateState 更新状态
func (ss *StateStore) UpdateState(updater func(*types.State)) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// 深拷贝当前状态
	stateCopy := ss.copyState(ss.current)

	// 保存历史
	ss.history = append(ss.history, ss.current)
	if len(ss.history) > ss.maxHist {
		ss.history = ss.history[1:]
	}

	// 应用更新
	updater(stateCopy)
	stateCopy.UpdatedAt = time.Now()
	ss.current = stateCopy
}

// copyState 深拷贝状态
func (ss *StateStore) copyState(state *types.State) *types.State {
	if state == nil {
		return nil
	}

	data, _ := json.Marshal(state)
	var copied types.State
	json.Unmarshal(data, &copied)
	return &copied
}

// GetHistory 获取历史状态
func (ss *StateStore) GetHistory(limit int) []*types.State {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if limit > len(ss.history) {
		limit = len(ss.history)
	}
	result := make([]*types.State, limit)
	copy(result, ss.history[len(ss.history)-limit:])
	return result
}

// GetStateAtTick 获取指定tick的状态
func (ss *StateStore) GetStateAtTick(tick uint64) *types.State {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if ss.current != nil && ss.current.Tick == tick {
		return ss.current
	}

	for i := len(ss.history) - 1; i >= 0; i-- {
		if ss.history[i].Tick == tick {
			return ss.history[i]
		}
	}
	return nil
}

// Reset 重置状态
func (ss *StateStore) Reset() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.current = &types.State{
		Tick:       0,
		Status:     types.StatusIdle,
		Nodes:      make(map[types.NodeID]*types.NodeState),
		GlobalData: make(map[string]interface{}),
		UpdatedAt:  time.Now(),
	}
	ss.history = make([]*types.State, 0)
}

// Export 导出状态
func (ss *StateStore) Export() (json.RawMessage, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return json.Marshal(ss.current)
}

// Import 导入状态
func (ss *StateStore) Import(data json.RawMessage) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	var state types.State
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	ss.current = &state
	return nil
}

// GetNodeState 获取节点状态
func (ss *StateStore) GetNodeState(nodeID types.NodeID) *types.NodeState {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if ss.current == nil {
		return nil
	}
	return ss.current.Nodes[nodeID]
}

// SetNodeState 设置节点状态
func (ss *StateStore) SetNodeState(nodeID types.NodeID, nodeState *types.NodeState) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.current == nil {
		ss.current = &types.State{
			Nodes: make(map[types.NodeID]*types.NodeState),
		}
	}
	ss.current.Nodes[nodeID] = nodeState
	ss.current.UpdatedAt = time.Now()
}

// GetGlobalData 获取全局数据
func (ss *StateStore) GetGlobalData(key string) interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if ss.current == nil {
		return nil
	}
	return ss.current.GlobalData[key]
}

// SetGlobalData 设置全局数据
func (ss *StateStore) SetGlobalData(key string, value interface{}) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.current == nil {
		ss.current = &types.State{
			GlobalData: make(map[string]interface{}),
		}
	}
	ss.current.GlobalData[key] = value
	ss.current.UpdatedAt = time.Now()
}

// IncrementTick 增加tick
func (ss *StateStore) IncrementTick() uint64 {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.current == nil {
		ss.current = &types.State{}
	}
	ss.current.Tick++
	ss.current.UpdatedAt = time.Now()
	return ss.current.Tick
}

// SetTick 设置tick
func (ss *StateStore) SetTick(tick uint64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.current == nil {
		ss.current = &types.State{}
	}
	ss.current.Tick = tick
	ss.current.UpdatedAt = time.Now()
}

// SetStatus 设置状态
func (ss *StateStore) SetStatus(status types.SimulatorStatus) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.current == nil {
		ss.current = &types.State{}
	}
	ss.current.Status = status
	ss.current.UpdatedAt = time.Now()
}

// StateDiff 状态差异
type StateDiff struct {
	Tick       *FieldDiff            `json:"tick,omitempty"`
	Status     *FieldDiff            `json:"status,omitempty"`
	Nodes      map[string]*NodeDiff  `json:"nodes,omitempty"`
	GlobalData map[string]*FieldDiff `json:"global_data,omitempty"`
	HasChanges bool                  `json:"has_changes"`
}

// FieldDiff 字段差异
type FieldDiff struct {
	Action string      `json:"action"` // added, removed, changed
	Old    interface{} `json:"old,omitempty"`
	New    interface{} `json:"new,omitempty"`
}

// NodeDiff 节点差异
type NodeDiff struct {
	Action     string                `json:"action"` // added, removed, changed
	OldState   *types.NodeState      `json:"old_state,omitempty"`
	NewState   *types.NodeState      `json:"new_state,omitempty"`
	FieldDiffs map[string]*FieldDiff `json:"field_diffs,omitempty"`
}

// Diff 计算状态差异
func (ss *StateStore) Diff(oldState, newState *types.State) *StateDiff {
	diff := &StateDiff{
		Nodes:      make(map[string]*NodeDiff),
		GlobalData: make(map[string]*FieldDiff),
		HasChanges: false,
	}

	if oldState == nil && newState == nil {
		return diff
	}

	if oldState == nil {
		diff.HasChanges = true
		diff.Tick = &FieldDiff{Action: "added", New: newState.Tick}
		diff.Status = &FieldDiff{Action: "added", New: newState.Status}
		for nodeID, node := range newState.Nodes {
			diff.Nodes[string(nodeID)] = &NodeDiff{Action: "added", NewState: node}
		}
		return diff
	}

	if newState == nil {
		diff.HasChanges = true
		diff.Tick = &FieldDiff{Action: "removed", Old: oldState.Tick}
		diff.Status = &FieldDiff{Action: "removed", Old: oldState.Status}
		for nodeID, node := range oldState.Nodes {
			diff.Nodes[string(nodeID)] = &NodeDiff{Action: "removed", OldState: node}
		}
		return diff
	}

	// 比较Tick
	if oldState.Tick != newState.Tick {
		diff.Tick = &FieldDiff{Action: "changed", Old: oldState.Tick, New: newState.Tick}
		diff.HasChanges = true
	}

	// 比较Status
	if oldState.Status != newState.Status {
		diff.Status = &FieldDiff{Action: "changed", Old: oldState.Status, New: newState.Status}
		diff.HasChanges = true
	}

	// 比较节点状态
	for nodeID, newNode := range newState.Nodes {
		oldNode, exists := oldState.Nodes[nodeID]
		if !exists {
			diff.Nodes[string(nodeID)] = &NodeDiff{Action: "added", NewState: newNode}
			diff.HasChanges = true
		} else if !ss.compareNodeState(oldNode, newNode) {
			nodeDiff := &NodeDiff{
				Action:     "changed",
				OldState:   oldNode,
				NewState:   newNode,
				FieldDiffs: ss.diffNodeFields(oldNode, newNode),
			}
			diff.Nodes[string(nodeID)] = nodeDiff
			diff.HasChanges = true
		}
	}

	// 检查被删除的节点
	for nodeID, oldNode := range oldState.Nodes {
		if _, exists := newState.Nodes[nodeID]; !exists {
			diff.Nodes[string(nodeID)] = &NodeDiff{Action: "removed", OldState: oldNode}
			diff.HasChanges = true
		}
	}

	// 比较全局数据
	for key, newVal := range newState.GlobalData {
		oldVal, exists := oldState.GlobalData[key]
		if !exists {
			diff.GlobalData[key] = &FieldDiff{Action: "added", New: newVal}
			diff.HasChanges = true
		} else if !reflect.DeepEqual(oldVal, newVal) {
			diff.GlobalData[key] = &FieldDiff{Action: "changed", Old: oldVal, New: newVal}
			diff.HasChanges = true
		}
	}

	for key, oldVal := range oldState.GlobalData {
		if _, exists := newState.GlobalData[key]; !exists {
			diff.GlobalData[key] = &FieldDiff{Action: "removed", Old: oldVal}
			diff.HasChanges = true
		}
	}

	// 清理空map
	if len(diff.Nodes) == 0 {
		diff.Nodes = nil
	}
	if len(diff.GlobalData) == 0 {
		diff.GlobalData = nil
	}

	return diff
}

// compareNodeState 比较两个节点状态是否相等
func (ss *StateStore) compareNodeState(a, b *types.NodeState) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.ID != b.ID || a.Status != b.Status || a.IsByzantine != b.IsByzantine {
		return false
	}
	return reflect.DeepEqual(a.Data, b.Data)
}

// diffNodeFields 比较节点字段差异
func (ss *StateStore) diffNodeFields(oldNode, newNode *types.NodeState) map[string]*FieldDiff {
	diffs := make(map[string]*FieldDiff)

	if oldNode.Status != newNode.Status {
		diffs["status"] = &FieldDiff{Action: "changed", Old: oldNode.Status, New: newNode.Status}
	}
	if oldNode.IsByzantine != newNode.IsByzantine {
		diffs["is_byzantine"] = &FieldDiff{Action: "changed", Old: oldNode.IsByzantine, New: newNode.IsByzantine}
	}

	// 比较Data字段
	for key, newVal := range newNode.Data {
		oldVal, exists := oldNode.Data[key]
		if !exists {
			diffs["data."+key] = &FieldDiff{Action: "added", New: newVal}
		} else if !reflect.DeepEqual(oldVal, newVal) {
			diffs["data."+key] = &FieldDiff{Action: "changed", Old: oldVal, New: newVal}
		}
	}

	for key, oldVal := range oldNode.Data {
		if _, exists := newNode.Data[key]; !exists {
			diffs["data."+key] = &FieldDiff{Action: "removed", Old: oldVal}
		}
	}

	if len(diffs) == 0 {
		return nil
	}
	return diffs
}

// Clone 克隆状态
func (ss *StateStore) Clone() *types.State {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.copyState(ss.current)
}
