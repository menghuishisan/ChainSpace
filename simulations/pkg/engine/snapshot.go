package engine

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

// SnapshotStore 快照存储
type SnapshotStore struct {
	mu        sync.RWMutex
	snapshots map[string]*types.Snapshot
	maxCount  int
}

// NewSnapshotStore 创建快照存储
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		snapshots: make(map[string]*types.Snapshot),
		maxCount:  100, // 最多保存100个快照
	}
}

// Save 保存快照
func (ss *SnapshotStore) Save(name string, state *types.State) (*types.SnapshotInfo, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// 序列化状态
	stateData, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	// 创建快照
	info := types.SnapshotInfo{
		ID:        uuid.New().String(),
		Name:      name,
		Tick:      state.Tick,
		CreatedAt: time.Now(),
		Size:      int64(len(stateData)),
	}

	snapshot := &types.Snapshot{
		Info:  info,
		State: stateData,
	}

	// 检查是否超过最大数量
	if len(ss.snapshots) >= ss.maxCount {
		// 删除最旧的快照
		var oldest *types.Snapshot
		for _, s := range ss.snapshots {
			if oldest == nil || s.Info.CreatedAt.Before(oldest.Info.CreatedAt) {
				oldest = s
			}
		}
		if oldest != nil {
			delete(ss.snapshots, oldest.Info.ID)
		}
	}

	ss.snapshots[info.ID] = snapshot
	return &info, nil
}

// Load 加载快照
func (ss *SnapshotStore) Load(id string) (*types.State, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	snapshot, ok := ss.snapshots[id]
	if !ok {
		return nil, fmt.Errorf("snapshot not found: %s", id)
	}

	var state types.State
	if err := json.Unmarshal(snapshot.State, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// LoadByName 按名称加载快照
func (ss *SnapshotStore) LoadByName(name string) (*types.State, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	for _, snapshot := range ss.snapshots {
		if snapshot.Info.Name == name {
			var state types.State
			if err := json.Unmarshal(snapshot.State, &state); err != nil {
				return nil, fmt.Errorf("failed to unmarshal state: %w", err)
			}
			return &state, nil
		}
	}
	return nil, fmt.Errorf("snapshot not found: %s", name)
}

// Delete 删除快照
func (ss *SnapshotStore) Delete(id string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if _, ok := ss.snapshots[id]; !ok {
		return fmt.Errorf("snapshot not found: %s", id)
	}
	delete(ss.snapshots, id)
	return nil
}

// List 列出所有快照
func (ss *SnapshotStore) List() []types.SnapshotInfo {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	var infos []types.SnapshotInfo
	for _, snapshot := range ss.snapshots {
		infos = append(infos, snapshot.Info)
	}
	return infos
}

// Get 获取快照信息
func (ss *SnapshotStore) Get(id string) (*types.SnapshotInfo, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	snapshot, ok := ss.snapshots[id]
	if !ok {
		return nil, fmt.Errorf("snapshot not found: %s", id)
	}
	return &snapshot.Info, nil
}

// Clear 清空所有快照
func (ss *SnapshotStore) Clear() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.snapshots = make(map[string]*types.Snapshot)
}

// Count 快照数量
func (ss *SnapshotStore) Count() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return len(ss.snapshots)
}

// Export 导出所有快照
func (ss *SnapshotStore) Export() ([]byte, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return json.Marshal(ss.snapshots)
}

// Import 导入快照
func (ss *SnapshotStore) Import(data []byte) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	var snapshots map[string]*types.Snapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return fmt.Errorf("failed to unmarshal snapshots: %w", err)
	}
	ss.snapshots = snapshots
	return nil
}

// GetByTick 按tick获取最近的快照
func (ss *SnapshotStore) GetByTick(tick uint64) *types.SnapshotInfo {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	var closest *types.SnapshotInfo
	var minDiff uint64 = ^uint64(0)

	for _, snapshot := range ss.snapshots {
		var diff uint64
		if snapshot.Info.Tick > tick {
			diff = snapshot.Info.Tick - tick
		} else {
			diff = tick - snapshot.Info.Tick
		}
		if diff < minDiff {
			minDiff = diff
			info := snapshot.Info
			closest = &info
		}
	}
	return closest
}

// AutoSnapshot 自动快照管理
type AutoSnapshot struct {
	store      *SnapshotStore
	stateStore *StateStore
	interval   uint64 // 每隔多少tick自动保存
	enabled    bool
	mu         sync.RWMutex
}

// NewAutoSnapshot 创建自动快照管理器
func NewAutoSnapshot(store *SnapshotStore, stateStore *StateStore, interval uint64) *AutoSnapshot {
	return &AutoSnapshot{
		store:      store,
		stateStore: stateStore,
		interval:   interval,
		enabled:    false,
	}
}

// Enable 启用自动快照
func (as *AutoSnapshot) Enable() {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.enabled = true
}

// Disable 禁用自动快照
func (as *AutoSnapshot) Disable() {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.enabled = false
}

// Check 检查是否需要保存快照
func (as *AutoSnapshot) Check(tick uint64) {
	as.mu.RLock()
	enabled := as.enabled
	interval := as.interval
	as.mu.RUnlock()

	if !enabled || interval == 0 {
		return
	}

	if tick%interval == 0 {
		state := as.stateStore.GetState()
		if state != nil {
			name := fmt.Sprintf("auto_%d", tick)
			as.store.Save(name, state)
		}
	}
}

// SetInterval 设置自动快照间隔
func (as *AutoSnapshot) SetInterval(interval uint64) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.interval = interval
}
