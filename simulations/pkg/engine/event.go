package engine

import (
	"sync"
	"time"

	"github.com/chainspace/simulations/pkg/types"
	"github.com/google/uuid"
)

// EventBus 事件总线
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan types.Event
	events      []types.Event
	maxEvents   int
}

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan types.Event),
		events:      make([]types.Event, 0),
		maxEvents:   10000, // 最多保留10000条事件
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType string) <-chan types.Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan types.Event, 100)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

// SubscribeAll 订阅所有事件
func (eb *EventBus) SubscribeAll() <-chan types.Event {
	return eb.Subscribe("*")
}

// Unsubscribe 取消订阅
func (eb *EventBus) Unsubscribe(eventType string, ch <-chan types.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs := eb.subscribers[eventType]
	for i, sub := range subs {
		if sub == ch {
			eb.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			close(sub)
			return
		}
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(event types.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// 设置事件ID和时间戳
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 保存事件
	eb.events = append(eb.events, event)
	if len(eb.events) > eb.maxEvents {
		eb.events = eb.events[len(eb.events)-eb.maxEvents:]
	}

	// 通知特定类型的订阅者
	for _, ch := range eb.subscribers[event.Type] {
		select {
		case ch <- event:
		default:
			// 通道满了，跳过
		}
	}

	// 通知所有事件订阅者
	for _, ch := range eb.subscribers["*"] {
		select {
		case ch <- event:
		default:
		}
	}
}

// PublishAsync 异步发布事件
func (eb *EventBus) PublishAsync(event types.Event) {
	go eb.Publish(event)
}

// GetEvents 获取事件
func (eb *EventBus) GetEvents(since uint64) []types.Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var result []types.Event
	for _, event := range eb.events {
		if event.Tick >= since {
			result = append(result, event)
		}
	}
	return result
}

// GetEventsByType 按类型获取事件
func (eb *EventBus) GetEventsByType(eventType string, limit int) []types.Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var result []types.Event
	for i := len(eb.events) - 1; i >= 0 && len(result) < limit; i-- {
		if eb.events[i].Type == eventType {
			result = append(result, eb.events[i])
		}
	}
	return result
}

// GetLatestEvents 获取最新事件
func (eb *EventBus) GetLatestEvents(limit int) []types.Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if limit > len(eb.events) {
		limit = len(eb.events)
	}
	return eb.events[len(eb.events)-limit:]
}

// Clear 清空事件
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.events = make([]types.Event, 0)
}

// Count 事件数量
func (eb *EventBus) Count() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.events)
}

// EventLogger 事件日志器
type EventLogger struct {
	eventBus *EventBus
	events   []types.Event
	mu       sync.RWMutex
	maxSize  int
}

// NewEventLogger 创建事件日志器
func NewEventLogger(eventBus *EventBus, maxSize int) *EventLogger {
	logger := &EventLogger{
		eventBus: eventBus,
		events:   make([]types.Event, 0),
		maxSize:  maxSize,
	}

	// 订阅所有事件
	go func() {
		ch := eventBus.SubscribeAll()
		for event := range ch {
			logger.log(event)
		}
	}()

	return logger
}

// log 记录事件
func (el *EventLogger) log(event types.Event) {
	el.mu.Lock()
	defer el.mu.Unlock()

	el.events = append(el.events, event)
	if len(el.events) > el.maxSize {
		el.events = el.events[1:]
	}
}

// GetLogs 获取日志
func (el *EventLogger) GetLogs(limit int) []types.Event {
	el.mu.RLock()
	defer el.mu.RUnlock()

	if limit > len(el.events) {
		limit = len(el.events)
	}
	result := make([]types.Event, limit)
	copy(result, el.events[len(el.events)-limit:])
	return result
}

// Filter 过滤日志
func (el *EventLogger) Filter(filterFn func(types.Event) bool) []types.Event {
	el.mu.RLock()
	defer el.mu.RUnlock()

	var result []types.Event
	for _, event := range el.events {
		if filterFn(event) {
			result = append(result, event)
		}
	}
	return result
}
