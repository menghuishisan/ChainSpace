package safego

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// Go 安全地启动goroutine，带panic恢复
func Go(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Goroutine panic recovered",
					zap.String("name", name),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()
		fn()
	}()
}

// GoWithContext 安全地启动带context的goroutine
func GoWithContext(ctx context.Context, name string, fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Goroutine panic recovered",
					zap.String("name", name),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()
		fn(ctx)
	}()
}

// GoWithTimeout 安全地启动带超时的goroutine
func GoWithTimeout(name string, timeout time.Duration, fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Goroutine panic recovered",
					zap.String("name", name),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			fn(ctx)
		}()

		select {
		case <-done:
			// 正常完成
		case <-ctx.Done():
			logger.Warn("Goroutine timeout",
				zap.String("name", name),
				zap.Duration("timeout", timeout),
			)
		}
	}()
}

// RunWithRecovery 同步执行函数，带panic恢复，返回error
func RunWithRecovery(name string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Function panic recovered",
				zap.String("name", name),
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())),
			)
			err = &PanicError{Name: name, Value: r}
		}
	}()
	return fn()
}

// PanicError panic错误包装
type PanicError struct {
	Name  string
	Value interface{}
}

func (e *PanicError) Error() string {
	return "panic in " + e.Name
}
