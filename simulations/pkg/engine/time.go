package engine

import (
	"sync"
	"time"
)

// TimeController 时间控制器
type TimeController struct {
	mu       sync.RWMutex
	tick     uint64
	speed    float64 // 速度倍率 0.1x - 10x
	paused   bool
	running  bool
	tickChan chan uint64
	stopChan chan struct{}
	interval time.Duration // 基础间隔
}

// NewTimeController 创建时间控制器
func NewTimeController() *TimeController {
	return &TimeController{
		tick:     0,
		speed:    1.0,
		paused:   false,
		running:  false,
		tickChan: make(chan uint64, 100),
		stopChan: make(chan struct{}),
		// 默认基础间隔放慢到 2.5s，配合 1x 默认速度时约为 2.5 秒一步。
		interval: 2500 * time.Millisecond,
	}
}

// Start 启动时间控制器
func (tc *TimeController) Start() {
	tc.mu.Lock()
	if tc.running {
		tc.mu.Unlock()
		return
	}
	tc.running = true
	tc.paused = false
	tc.mu.Unlock()

	go tc.run()
}

// run 运行循环
func (tc *TimeController) run() {
	ticker := time.NewTicker(tc.getActualInterval())
	defer ticker.Stop()

	for {
		select {
		case <-tc.stopChan:
			return
		case <-ticker.C:
			tc.mu.RLock()
			paused := tc.paused
			tc.mu.RUnlock()

			if !paused {
				tc.mu.Lock()
				tc.tick++
				currentTick := tc.tick
				tc.mu.Unlock()

				// 非阻塞发送
				select {
				case tc.tickChan <- currentTick:
				default:
				}
			}

			// 更新ticker间隔
			ticker.Reset(tc.getActualInterval())
		}
	}
}

// getActualInterval 获取实际间隔
func (tc *TimeController) getActualInterval() time.Duration {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return time.Duration(float64(tc.interval) / tc.speed)
}

// Stop 停止时间控制器
func (tc *TimeController) Stop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if !tc.running {
		return
	}
	tc.running = false
	close(tc.stopChan)
	tc.stopChan = make(chan struct{})
}

// Pause 暂停
func (tc *TimeController) Pause() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.paused = true
}

// Resume 恢复
func (tc *TimeController) Resume() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.paused = false
}

// Step 单步
func (tc *TimeController) Step() {
	tc.mu.Lock()
	tc.tick++
	currentTick := tc.tick
	tc.mu.Unlock()

	select {
	case tc.tickChan <- currentTick:
	default:
	}
}

// Reset 重置
func (tc *TimeController) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tick = 0
	tc.paused = false
}

// GetTick 获取当前tick
func (tc *TimeController) GetTick() uint64 {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.tick
}

// SetTick 设置tick（用于跳转）
func (tc *TimeController) SetTick(tick uint64) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tick = tick
}

// SetSpeed 设置速度
func (tc *TimeController) SetSpeed(speed float64) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if speed < 0.1 {
		speed = 0.1
	}
	if speed > 10.0 {
		speed = 10.0
	}
	tc.speed = speed
}

// GetSpeed 获取速度
func (tc *TimeController) GetSpeed() float64 {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.speed
}

// IsPaused 是否暂停
func (tc *TimeController) IsPaused() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.paused
}

// IsRunning 是否运行中
func (tc *TimeController) IsRunning() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.running
}

// SetInterval 设置基础间隔
func (tc *TimeController) SetInterval(interval time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.interval = interval
}

// TickChan 获取tick通道
func (tc *TimeController) TickChan() <-chan uint64 {
	return tc.tickChan
}

// WaitTick 等待指定tick
func (tc *TimeController) WaitTick(targetTick uint64) {
	for {
		tc.mu.RLock()
		currentTick := tc.tick
		tc.mu.RUnlock()

		if currentTick >= targetTick {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Elapsed 获取经过的模拟时间
func (tc *TimeController) Elapsed() time.Duration {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return time.Duration(tc.tick) * tc.interval
}
