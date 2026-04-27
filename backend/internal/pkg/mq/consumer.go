package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// Consumer 消息消费者
type Consumer struct {
	client   *Client
	handlers map[string]MessageHandler
}

// MessageHandler 消息处理函数
type MessageHandler func(ctx context.Context, msg *Message) error

// NewConsumer 创建消费者
func NewConsumer(client *Client) *Consumer {
	return &Consumer{
		client:   client,
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler 注册消息处理器
func (c *Consumer) RegisterHandler(queue string, handler MessageHandler) {
	c.handlers[queue] = handler
}

// Start 启动消费者
func (c *Consumer) Start(ctx context.Context) error {
	for queue, handler := range c.handlers {
		if err := c.startQueueConsumer(ctx, queue, handler); err != nil {
			return fmt.Errorf("start consumer for %s: %w", queue, err)
		}
		logger.Info("Started consumer", zap.String("queue", queue))
	}
	return nil
}

// startQueueConsumer 启动单个队列消费者
func (c *Consumer) startQueueConsumer(ctx context.Context, queue string, handler MessageHandler) error {
	c.client.mu.RLock()
	channel := c.client.channel
	c.client.mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("channel not available")
	}

	msgs, err := channel.Consume(
		queue,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					logger.Warn("Consumer channel closed", zap.String("queue", queue))
					// 尝试重新连接
					time.Sleep(5 * time.Second)
					c.startQueueConsumer(ctx, queue, handler)
					return
				}

				var msg Message
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					logger.Error("Unmarshal message failed",
						zap.String("queue", queue),
						zap.Error(err),
					)
					d.Nack(false, false) // 发送到死信队列
					continue
				}

				// 执行处理器
				if err := handler(ctx, &msg); err != nil {
					logger.Error("Handle message failed",
						zap.String("queue", queue),
						zap.String("msg_id", msg.ID),
						zap.Int("retry_count", msg.RetryCount),
						zap.Error(err),
					)

					// 重试逻辑
					msg.RetryCount++
					if msg.RetryCount < 3 {
						// 重新发布到队列进行重试
						c.client.Publish(ctx, queue, &msg)
						d.Ack(false)
					} else {
						// 超过重试次数，发送到死信队列
						d.Nack(false, false)
					}
					continue
				}

				d.Ack(false)
				logger.Debug("Message processed",
					zap.String("queue", queue),
					zap.String("msg_id", msg.ID),
				)
			}
		}
	}()

	return nil
}

// ConsumeWithRetry 带重试的消费（使用延迟队列）
func (c *Consumer) ConsumeWithRetry(ctx context.Context, queue string, handler MessageHandler, maxRetries int) error {
	return c.startQueueConsumer(ctx, queue, func(ctx context.Context, msg *Message) error {
		err := handler(ctx, msg)
		if err != nil && msg.RetryCount < maxRetries {
			// 计算延迟时间（指数退避）
			delay := time.Duration(1<<uint(msg.RetryCount)) * time.Second
			if delay > 5*time.Minute {
				delay = 5 * time.Minute
			}

			logger.Info("Scheduling retry",
				zap.String("msg_id", msg.ID),
				zap.Int("retry_count", msg.RetryCount+1),
				zap.Duration("delay", delay),
			)

			// 延迟后重新发布
			time.AfterFunc(delay, func() {
				msg.RetryCount++
				c.client.Publish(context.Background(), queue, msg)
			})
			return nil // 返回nil避免立即重试
		}
		return err
	})
}
