package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/chainspace/simulations/pkg/types"
)

// Simulator 模拟器接口
type Simulator interface {
	// 生命周期
	Init(config types.Config) error
	Start(ctx context.Context) error
	Stop() error
	Reset() error

	// 时间控制
	Pause() error
	Resume() error
	Step() (*types.State, error)
	SetSpeed(multiplier float64) error
	Seek(tick uint64) error

	// 状态管理
	GetState() *types.State
	GetEvents(since uint64) []types.Event
	ExportState() (json.RawMessage, error)
	ImportState(data json.RawMessage) error

	// 快照管理
	SaveSnapshot(name string) error
	LoadSnapshot(name string) error
	ListSnapshots() []types.SnapshotInfo

	// 参数控制
	GetParams() map[string]types.Param
	SetParam(key string, value interface{}) error

	// 故障/攻击注入
	InjectFault(fault *types.Fault) error
	RemoveFault(faultID string) error
	InjectAttack(attack *types.Attack) error
	RemoveAttack(attackID string) error
	ClearFaults() error
	ClearAttacks() error

	// 元信息
	GetType() types.ComponentType
	GetCapabilities() []types.Capability
	GetDescription() types.Description
}

// SimulatorFactory 模拟器工厂
type SimulatorFactory interface {
	Create() Simulator
	GetDescription() types.Description
}

// Engine 引擎管理器
type Engine struct {
	mu         sync.RWMutex
	active     Simulator
	reg        *Registry
	eventBus   *EventBus
	timeCtrl   *TimeController
	stateStore *StateStore
	snapStore  *SnapshotStore
	config     types.Config
	status     types.SimulatorStatus
}

// NewEngine 创建引擎
func NewEngine() *Engine {
	return &Engine{
		reg:        NewRegistry(),
		eventBus:   NewEventBus(),
		timeCtrl:   NewTimeController(),
		stateStore: NewStateStore(),
		snapStore:  NewSnapshotStore(),
		status:     types.StatusIdle,
	}
}

// Registry 获取模块注册表
func (e *Engine) Registry() *Registry {
	return e.reg
}

// Register 注册模拟器工厂
func (e *Engine) Register(name string, factory SimulatorFactory) {
	e.reg.Register(name, factory)
}

// ListSimulators 列出可用模拟器
func (e *Engine) ListSimulators() []types.Description {
	return e.reg.List()
}

// Init 初始化模拟器
func (e *Engine) Init(config types.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	factory, ok := e.reg.Get(config.Module)
	if !ok {
		return fmt.Errorf("simulator not found: %s", config.Module)
	}

	e.active = factory.Create()
	e.config = config

	if err := e.active.Init(config); err != nil {
		return fmt.Errorf("failed to init simulator: %w", err)
	}

	e.status = types.StatusIdle
	e.eventBus.Publish(types.Event{
		Type: "engine.initialized",
		Data: map[string]interface{}{
			"module": config.Module,
		},
	})

	return nil
}

// Start 启动模拟
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}

	if err := e.active.Start(ctx); err != nil {
		return err
	}

	e.status = types.StatusRunning
	e.timeCtrl.Start()
	e.eventBus.Publish(types.Event{
		Type: "engine.started",
	})

	return nil
}

// Stop 停止模拟
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return nil
	}

	if err := e.active.Stop(); err != nil {
		return err
	}

	e.status = types.StatusStopped
	e.timeCtrl.Stop()
	e.eventBus.Publish(types.Event{
		Type: "engine.stopped",
	})

	return nil
}

// Pause 暂停模拟
func (e *Engine) Pause() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}

	if err := e.active.Pause(); err != nil {
		return err
	}

	e.status = types.StatusPaused
	e.timeCtrl.Pause()
	e.eventBus.Publish(types.Event{
		Type: "engine.paused",
	})

	return nil
}

// Resume 恢复模拟
func (e *Engine) Resume() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}

	if err := e.active.Resume(); err != nil {
		return err
	}

	e.status = types.StatusRunning
	e.timeCtrl.Resume()
	e.eventBus.Publish(types.Event{
		Type: "engine.resumed",
	})

	return nil
}

// Step 单步执行
func (e *Engine) Step() (*types.State, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return nil, fmt.Errorf("no active simulator")
	}

	state, err := e.active.Step()
	if err != nil {
		return nil, err
	}

	e.timeCtrl.Step()
	e.eventBus.Publish(types.Event{
		Type: "engine.step",
		Tick: state.Tick,
	})

	return state, nil
}

// Reset 重置模拟
func (e *Engine) Reset() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}

	if err := e.active.Reset(); err != nil {
		return err
	}

	e.status = types.StatusIdle
	e.timeCtrl.Reset()
	e.eventBus.Publish(types.Event{
		Type: "engine.reset",
	})

	return nil
}

// Switch 切换模拟器
func (e *Engine) Switch(moduleName string, preserveState bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 保存当前状态
	var savedState json.RawMessage
	if preserveState && e.active != nil {
		var err error
		savedState, err = e.active.ExportState()
		if err != nil {
			return fmt.Errorf("failed to export state: %w", err)
		}
	}

	// 停止当前模拟器
	if e.active != nil {
		e.active.Stop()
	}

	// 创建新模拟器
	factory, ok := e.reg.Get(moduleName)
	if !ok {
		return fmt.Errorf("simulator not found: %s", moduleName)
	}

	e.active = factory.Create()
	e.config.Module = moduleName

	if err := e.active.Init(e.config); err != nil {
		return fmt.Errorf("failed to init simulator: %w", err)
	}

	// 恢复状态
	if preserveState && savedState != nil {
		if err := e.active.ImportState(savedState); err != nil {
			return fmt.Errorf("failed to import state: %w", err)
		}
	}

	e.status = types.StatusIdle
	e.eventBus.Publish(types.Event{
		Type: "engine.switched",
		Data: map[string]interface{}{
			"module": moduleName,
		},
	})

	return nil
}

// GetState 获取状态
func (e *Engine) GetState() *types.State {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.active == nil {
		return nil
	}
	return e.active.GetState()
}

// GetEvents 获取事件
func (e *Engine) GetEvents(since uint64) []types.Event {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.active == nil {
		return nil
	}
	return e.active.GetEvents(since)
}

// GetParams 获取参数
func (e *Engine) GetParams() map[string]types.Param {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.active == nil {
		return nil
	}
	return e.active.GetParams()
}

// SetParam 设置参数
func (e *Engine) SetParam(key string, value interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.SetParam(key, value)
}

// SetSpeed 设置速度
func (e *Engine) SetSpeed(multiplier float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	e.timeCtrl.SetSpeed(multiplier)
	return e.active.SetSpeed(multiplier)
}

// InjectFault 注入故障
func (e *Engine) InjectFault(fault *types.Fault) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.InjectFault(fault)
}

// RemoveFault 移除故障
func (e *Engine) RemoveFault(faultID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.RemoveFault(faultID)
}

// InjectAttack 注入攻击
func (e *Engine) InjectAttack(attack *types.Attack) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.InjectAttack(attack)
}

// ExecuteAction 执行当前模块暴露的教学动作。
func (e *Engine) ExecuteAction(action string, params map[string]interface{}) (*types.ActionResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return nil, fmt.Errorf("no active simulator")
	}

	handler, ok := e.active.(types.ActionHandler)
	if !ok {
		return nil, fmt.Errorf("simulator does not support custom actions")
	}

	return handler.ExecuteAction(action, params)
}

// RemoveAttack 移除攻击
func (e *Engine) RemoveAttack(attackID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.RemoveAttack(attackID)
}

// SaveSnapshot 保存快照
func (e *Engine) SaveSnapshot(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.SaveSnapshot(name)
}

// LoadSnapshot 加载快照
func (e *Engine) LoadSnapshot(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.LoadSnapshot(name)
}

// ListSnapshots 列出快照
func (e *Engine) ListSnapshots() []types.SnapshotInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.active == nil {
		return nil
	}
	return e.active.ListSnapshots()
}

// DeleteSnapshot 删除快照
func (e *Engine) DeleteSnapshot(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.snapStore.Delete(id)
}

// ExportState 导出状态
func (e *Engine) ExportState() (json.RawMessage, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.active == nil {
		return nil, fmt.Errorf("no active simulator")
	}
	return e.active.ExportState()
}

// ImportState 导入状态
func (e *Engine) ImportState(data interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return e.active.ImportState(jsonData)
}

// ClearFaults 清除所有故障
func (e *Engine) ClearFaults() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.ClearFaults()
}

// ClearAttacks 清除所有攻击
func (e *Engine) ClearAttacks() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active == nil {
		return fmt.Errorf("no active simulator")
	}
	return e.active.ClearAttacks()
}

// GetDescription 获取当前模拟器描述
func (e *Engine) GetDescription() *types.Description {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.active == nil {
		return nil
	}
	desc := e.active.GetDescription()
	return &desc
}

// GetStatus 获取引擎状态
func (e *Engine) GetStatus() types.SimulatorStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// GetEventBus 获取事件总线
func (e *Engine) GetEventBus() *EventBus {
	return e.eventBus
}

// GetTimeController 获取时间控制器
func (e *Engine) GetTimeController() *TimeController {
	return e.timeCtrl
}
