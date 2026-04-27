package base

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

// BaseSimulator 基础模拟器实现
type BaseSimulator struct {
	mu          sync.RWMutex
	id          string
	name        string
	description string
	category    string
	compType    types.ComponentType
	caps        []types.Capability
	params      map[string]types.Param
	state       *types.State
	events      []types.Event
	snapshots   map[string]*types.Snapshot
	faults      map[string]*types.Fault
	attacks     map[string]*types.Attack
	running     bool
	paused      bool
	speed       float64
	ctx         context.Context
	cancel      context.CancelFunc
	tickChan    chan uint64
	onTick      func(tick uint64) error
}

// NewBaseSimulator 创建基础模拟器
func NewBaseSimulator(id, name, description, category string, compType types.ComponentType) *BaseSimulator {
	return &BaseSimulator{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		compType:    compType,
		caps:        defaultCapabilities(compType, false),
		params:      make(map[string]types.Param),
		state: &types.State{
			Tick:       0,
			Status:     types.StatusIdle,
			Nodes:      make(map[types.NodeID]*types.NodeState),
			GlobalData: make(map[string]interface{}),
			UpdatedAt:  time.Now(),
		},
		events:    make([]types.Event, 0),
		snapshots: make(map[string]*types.Snapshot),
		faults:    make(map[string]*types.Fault),
		attacks:   make(map[string]*types.Attack),
		// 默认速度保持在 1x，由基础间隔统一控制教学节奏。
		speed:     1.0,
		tickChan:  make(chan uint64, 100),
	}
}

// defaultCapabilities 根据组件类型返回默认能力
func defaultCapabilities(compType types.ComponentType, hasTick bool) []types.Capability {
	switch compType {
	case types.ComponentTool:
		return []types.Capability{types.CapabilityParamPanel}
	case types.ComponentDemo:
		return []types.Capability{types.CapabilityParamPanel, types.CapabilityStateMonitor}
	case types.ComponentProcess:
		caps := []types.Capability{
			types.CapabilityParamPanel,
			types.CapabilityStateMonitor,
			types.CapabilityEventLog,
			types.CapabilitySnapshot,
		}
		if hasTick {
			caps = append(caps, types.CapabilityTimeControl)
		}
		return caps
	case types.ComponentAttack:
		caps := []types.Capability{
			types.CapabilityParamPanel,
			types.CapabilityStateMonitor,
			types.CapabilityEventLog,
		}
		if hasTick {
			caps = append(caps, types.CapabilityTimeControl)
		}
		return caps
	case types.ComponentDeFi:
		return []types.Capability{
			types.CapabilityParamPanel,
			types.CapabilityStateMonitor,
			types.CapabilityEventLog,
			types.CapabilitySnapshot,
		}
	default:
		return []types.Capability{types.CapabilityParamPanel}
	}
}

// SetOnTick 设置tick回调
func (s *BaseSimulator) SetOnTick(fn func(tick uint64) error) {
	s.onTick = fn
	s.mu.Lock()
	defer s.mu.Unlock()
	s.caps = defaultCapabilities(s.compType, fn != nil)
}

// Init 初始化
func (s *BaseSimulator) Init(config types.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range config.Params {
		if param, ok := s.params[key]; ok {
			param.Value = value
			s.params[key] = param
		}
	}
	s.syncDisturbanceStateLocked()
	s.clearTeachingStateLocked()
	return nil
}

// Start 启动
func (s *BaseSimulator) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// 模拟器运行循环不能绑定到单次 HTTP 请求上下文。
	// 否则请求一结束，ctx 会立即取消，导致“启动成功但不会自动推进”。
	// 这里统一改为独立的后台上下文，由 Stop 显式控制生命周期。
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true
	s.paused = false
	s.state.Status = types.StatusRunning

	go s.runLoop()
	return nil
}

// runLoop 运行循环
func (s *BaseSimulator) runLoop() {
	// 过程类实验默认以 2.5s 为基础间隔，再乘以速度倍率控制，
	// 让默认自动模式落在 2~3 秒一步，方便课堂观察。
	interval := time.Duration(float64(2500*time.Millisecond) / s.speed)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			paused := s.paused
			s.mu.RUnlock()

			if !paused {
				s.mu.Lock()
				s.state.Tick++
				tick := s.state.Tick
				s.state.UpdatedAt = time.Now()
				s.mu.Unlock()

				if s.onTick != nil {
					s.onTick(tick)
				}
			}

			// 更新ticker间隔
			s.mu.RLock()
			newInterval := time.Duration(float64(2500*time.Millisecond) / s.speed)
			s.mu.RUnlock()
			if newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
			}
		}
	}
}

// Stop 停止
func (s *BaseSimulator) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}
	s.running = false
	s.state.Status = types.StatusStopped
	return nil
}

// Reset 重置
func (s *BaseSimulator) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = &types.State{
		Tick:       0,
		Status:     types.StatusIdle,
		Nodes:      make(map[types.NodeID]*types.NodeState),
		GlobalData: make(map[string]interface{}),
		UpdatedAt:  time.Now(),
	}
	s.events = make([]types.Event, 0)
	s.faults = make(map[string]*types.Fault)
	s.attacks = make(map[string]*types.Attack)
	s.syncDisturbanceStateLocked()
	s.clearTeachingStateLocked()
	return nil
}

// Pause 暂停
func (s *BaseSimulator) Pause() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.paused = true
	s.state.Status = types.StatusPaused
	return nil
}

// Resume 恢复
func (s *BaseSimulator) Resume() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.paused = false
	s.state.Status = types.StatusRunning
	return nil
}

// Step 单步执行
func (s *BaseSimulator) Step() (*types.State, error) {
	s.mu.Lock()
	s.state.Tick++
	tick := s.state.Tick
	s.state.UpdatedAt = time.Now()
	s.mu.Unlock()

	if s.onTick != nil {
		s.onTick(tick)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.copyState(), nil
}

// SetSpeed 设置速度
func (s *BaseSimulator) SetSpeed(multiplier float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if multiplier < 0.1 {
		multiplier = 0.1
	}
	if multiplier > 10.0 {
		multiplier = 10.0
	}
	s.speed = multiplier
	return nil
}

// Seek 跳转到指定tick
func (s *BaseSimulator) Seek(tick uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state.Tick = tick
	s.state.UpdatedAt = time.Now()
	return nil
}

// GetState 获取状态
func (s *BaseSimulator) GetState() *types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.copyState()
}

func (s *BaseSimulator) copyState() *types.State {
	data, _ := json.Marshal(s.state)
	var copied types.State
	json.Unmarshal(data, &copied)
	return &copied
}

// GetEvents 获取事件
func (s *BaseSimulator) GetEvents(since uint64) []types.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []types.Event
	for _, e := range s.events {
		if e.Tick >= since {
			result = append(result, e)
		}
	}
	return result
}

// EmitEvent 发出事件
func (s *BaseSimulator) EmitEvent(eventType string, source, target types.NodeID, data map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := types.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Tick:      s.state.Tick,
		Timestamp: time.Now(),
		Source:    source,
		Target:    target,
		Data:      data,
	}
	s.events = append(s.events, event)

	// 限制事件数量
	if len(s.events) > 10000 {
		s.events = s.events[len(s.events)-10000:]
	}
}

// ExportState 导出状态
func (s *BaseSimulator) ExportState() (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.state)
}

// ImportState 导入状态
func (s *BaseSimulator) ImportState(data json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var state types.State
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	s.state = &state
	return nil
}

// SaveSnapshot 保存快照
func (s *BaseSimulator) SaveSnapshot(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stateData, err := json.Marshal(s.state)
	if err != nil {
		return err
	}

	info := types.SnapshotInfo{
		ID:        uuid.New().String(),
		Name:      name,
		Tick:      s.state.Tick,
		CreatedAt: time.Now(),
		Size:      int64(len(stateData)),
	}

	s.snapshots[info.ID] = &types.Snapshot{
		Info:  info,
		State: stateData,
	}
	return nil
}

// LoadSnapshot 加载快照
func (s *BaseSimulator) LoadSnapshot(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, snap := range s.snapshots {
		if snap.Info.Name == name {
			var state types.State
			if err := json.Unmarshal(snap.State, &state); err != nil {
				return err
			}
			s.state = &state
			return nil
		}
	}
	return fmt.Errorf("snapshot not found: %s", name)
}

// ListSnapshots 列出快照
func (s *BaseSimulator) ListSnapshots() []types.SnapshotInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []types.SnapshotInfo
	for _, snap := range s.snapshots {
		infos = append(infos, snap.Info)
	}
	return infos
}

// GetParams 获取参数
func (s *BaseSimulator) GetParams() map[string]types.Param {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]types.Param)
	for k, v := range s.params {
		result[k] = v
	}
	return result
}

// SetParam 设置参数
func (s *BaseSimulator) SetParam(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	param, ok := s.params[key]
	if !ok {
		return fmt.Errorf("unknown parameter: %s", key)
	}
	param.Value = value
	s.params[key] = param
	return nil
}

// AddParam 添加参数定义
func (s *BaseSimulator) AddParam(param types.Param) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.params[param.Key] = param
}

// InjectFault 注入故障
func (s *BaseSimulator) InjectFault(fault *types.Fault) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fault.ID == "" {
		fault.ID = uuid.New().String()
	}
	fault.StartTick = s.state.Tick
	fault.Active = true
	s.faults[fault.ID] = fault
	s.syncDisturbanceStateLocked()
	return nil
}

// RemoveFault 移除故障
func (s *BaseSimulator) RemoveFault(faultID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.faults[faultID]; !ok {
		return fmt.Errorf("fault not found: %s", faultID)
	}
	delete(s.faults, faultID)
	s.syncDisturbanceStateLocked()
	return nil
}

// ClearFaults 清除所有故障
func (s *BaseSimulator) ClearFaults() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faults = make(map[string]*types.Fault)
	s.syncDisturbanceStateLocked()
	return nil
}

// GetActiveFaults 获取活跃故障
func (s *BaseSimulator) GetActiveFaults() []*types.Fault {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var faults []*types.Fault
	for _, f := range s.faults {
		if f.Active {
			faults = append(faults, f)
		}
	}
	return faults
}

// InjectAttack 注入攻击
func (s *BaseSimulator) InjectAttack(attack *types.Attack) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if attack.ID == "" {
		attack.ID = uuid.New().String()
	}
	attack.StartTick = s.state.Tick
	attack.Active = true
	s.attacks[attack.ID] = attack
	s.syncDisturbanceStateLocked()
	return nil
}

// RemoveAttack 移除攻击
func (s *BaseSimulator) RemoveAttack(attackID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.attacks[attackID]; !ok {
		return fmt.Errorf("attack not found: %s", attackID)
	}
	delete(s.attacks, attackID)
	s.syncDisturbanceStateLocked()
	return nil
}

// ClearAttacks 清除所有攻击
func (s *BaseSimulator) ClearAttacks() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attacks = make(map[string]*types.Attack)
	s.syncDisturbanceStateLocked()
	return nil
}

// GetActiveAttacks 获取活跃攻击
func (s *BaseSimulator) GetActiveAttacks() []*types.Attack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var attacks []*types.Attack
	for _, a := range s.attacks {
		if a.Active {
			attacks = append(attacks, a)
		}
	}
	return attacks
}

// GetType 获取组件类型
func (s *BaseSimulator) GetType() types.ComponentType {
	return s.compType
}

// GetCapabilities 获取能力列表
func (s *BaseSimulator) GetCapabilities() []types.Capability {
	return s.caps
}

// GetDescription 获取描述
func (s *BaseSimulator) GetDescription() types.Description {
	s.mu.RLock()
	defer s.mu.RUnlock()

	params := make([]types.Param, 0, len(s.params))
	for _, p := range s.params {
		params = append(params, p)
	}

	return types.Description{
		ID:           s.id,
		Name:         s.name,
		Description:  s.description,
		Category:     s.category,
		Type:         s.compType,
		Capabilities: s.caps,
		Params:       params,
		Version:      "1.0.0",
	}
}

// SetNodeState 设置节点状态
func (s *BaseSimulator) SetNodeState(nodeID types.NodeID, state *types.NodeState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Nodes[nodeID] = state
}

// ClearNodeStates 在模块重建场景前清空旧节点状态。
// 这样前端重新渲染时不会混入上一次场景残留的节点。
func (s *BaseSimulator) ClearNodeStates() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Nodes = make(map[types.NodeID]*types.NodeState)
}

// GetNodeState 获取节点状态
func (s *BaseSimulator) GetNodeState(nodeID types.NodeID) *types.NodeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Nodes[nodeID]
}

// SetGlobalData 设置全局数据
func (s *BaseSimulator) SetGlobalData(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.GlobalData[key] = value
}

// syncDisturbanceStateLocked 把当前激活的故障和攻击覆盖层统一导出到全局状态。
// 这样前端主舞台可以直接读取“当前有哪些联动影响正在生效”，而不必从零散事件里反推。
func (s *BaseSimulator) syncDisturbanceStateLocked() {
	activeFaults := make([]types.DisturbanceSnapshot, 0)
	for _, fault := range s.faults {
		if !fault.Active {
			continue
		}
		activeFaults = append(activeFaults, types.DisturbanceSnapshot{
			ID:       fault.ID,
			Kind:     "fault",
			Type:     string(fault.Type),
			Target:   string(fault.Target),
			Label:    "故障联动",
			Summary:  fmt.Sprintf("故障类型 %s 当前作用于 %s。", fault.Type, fault.Target),
			Params:   fault.Params,
			Duration: fault.Duration,
		})
	}

	activeAttacks := make([]types.DisturbanceSnapshot, 0)
	for _, attack := range s.attacks {
		if !attack.Active {
			continue
		}
		activeAttacks = append(activeAttacks, types.DisturbanceSnapshot{
			ID:       attack.ID,
			Kind:     "attack",
			Type:     string(attack.Type),
			Target:   attack.Target,
			Label:    "攻击联动",
			Summary:  fmt.Sprintf("攻击类型 %s 当前作用于 %s。", attack.Type, attack.Target),
			Params:   attack.Params,
			Duration: attack.Duration,
		})
	}

	s.state.GlobalData["active_faults"] = activeFaults
	s.state.GlobalData["active_attacks"] = activeAttacks
	s.state.GlobalData["fault_count"] = len(activeFaults)
	s.state.GlobalData["attack_overlay_count"] = len(activeAttacks)
}

// GetGlobalData 获取全局数据
func (s *BaseSimulator) GetGlobalData(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.GlobalData[key]
}

func (s *BaseSimulator) SetLinkedEffects(effects []types.LinkedEffect) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.GlobalData["linked_effects"] = effects
	s.state.GlobalData["linked_effect_count"] = len(effects)
}

func (s *BaseSimulator) SetProcessFeedback(feedback *types.ProcessFeedback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.GlobalData["process_feedback"] = feedback
}

func (s *BaseSimulator) ClearTeachingState() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearTeachingStateLocked()
}

func (s *BaseSimulator) clearTeachingStateLocked() {
	s.state.GlobalData["linked_effects"] = []types.LinkedEffect{}
	s.state.GlobalData["linked_effect_count"] = 0
	s.state.GlobalData["process_feedback"] = nil
}

// Ensure BaseSimulator implements engine.Simulator
var _ engine.Simulator = (*BaseSimulator)(nil)
