package engine

import (
	"fmt"
	"sync"

	"github.com/chainspace/simulations/pkg/types"
)

// Registry 模块注册表
type Registry struct {
	mu        sync.RWMutex
	factories map[string]SimulatorFactory
	modules   map[string]types.Description
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]SimulatorFactory),
		modules:   make(map[string]types.Description),
	}
}

// globalRegistry 全局注册表
var globalRegistry = NewRegistry()

// GetGlobalRegistry 获取全局注册表
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// Register 注册模拟器工厂
func (r *Registry) Register(name string, factory SimulatorFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("simulator already registered: %s", name)
	}

	r.factories[name] = factory
	r.modules[name] = factory.GetDescription()
	return nil
}

// Unregister 注销模拟器
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.factories, name)
	delete(r.modules, name)
}

// Get 获取模拟器工厂
func (r *Registry) Get(name string) (SimulatorFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[name]
	return factory, ok
}

// Create 创建模拟器实例
func (r *Registry) Create(name string) (Simulator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("simulator not found: %s", name)
	}
	return factory.Create(), nil
}

// List 列出所有已注册的模块
func (r *Registry) List() []types.Description {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var descriptions []types.Description
	for _, desc := range r.modules {
		descriptions = append(descriptions, desc)
	}
	return descriptions
}

// ListByCategory 按类别列出模块
func (r *Registry) ListByCategory(category string) []types.Description {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var descriptions []types.Description
	for _, desc := range r.modules {
		if desc.Category == category {
			descriptions = append(descriptions, desc)
		}
	}
	return descriptions
}

// ListByType 按类型列出模块
func (r *Registry) ListByType(componentType types.ComponentType) []types.Description {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var descriptions []types.Description
	for _, desc := range r.modules {
		if desc.Type == componentType {
			descriptions = append(descriptions, desc)
		}
	}
	return descriptions
}

// GetDescription 获取模块描述
func (r *Registry) GetDescription(name string) (*types.Description, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	desc, ok := r.modules[name]
	if !ok {
		return nil, false
	}
	return &desc, true
}

// Count 已注册模块数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.factories)
}

// Categories 获取所有类别
func (r *Registry) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categorySet := make(map[string]struct{})
	for _, desc := range r.modules {
		categorySet[desc.Category] = struct{}{}
	}

	var categories []string
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	return categories
}

// Register 全局注册函数
func Register(name string, factory SimulatorFactory) error {
	return globalRegistry.Register(name, factory)
}

// MustRegister 必须成功注册，失败则panic
func MustRegister(name string, factory SimulatorFactory) {
	if err := globalRegistry.Register(name, factory); err != nil {
		panic(err)
	}
}

// Get 全局获取工厂
func Get(name string) (SimulatorFactory, bool) {
	return globalRegistry.Get(name)
}

// Create 全局创建模拟器
func Create(name string) (Simulator, error) {
	return globalRegistry.Create(name)
}

// ListModules 全局列出模块
func ListModules() []types.Description {
	return globalRegistry.List()
}
