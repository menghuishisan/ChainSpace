package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Client RabbitMQ客户端
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
	mu      sync.RWMutex
	closed  bool

	// 重连配置
	reconnectDelay time.Duration
	maxRetries     int
}

// Config RabbitMQ配置
type Config struct {
	URL            string
	ReconnectDelay time.Duration
	MaxRetries     int
}

// Message 消息结构
type Message struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	CreatedAt  time.Time              `json:"created_at"`
	RetryCount int                    `json:"retry_count"`
}

// Queue 队列名称常量
const (
	QueueEnvCreate    = "env.create"
	QueueEnvDestroy   = "env.destroy"
	QueueNotification = "notification"
	QueueVulnSync     = "vuln.sync"
	QueueVulnConvert  = "vuln.convert"
	QueueEmailSend    = "email.send"
)

// Exchange 交换机名称
const (
	ExchangeDirect = "chainspace.direct"
	ExchangeTopic  = "chainspace.topic"
	ExchangeDLX    = "chainspace.dlx" // 死信交换机
)

// NewClient 创建RabbitMQ客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg.ReconnectDelay == 0 {
		cfg.ReconnectDelay = 5 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 10
	}

	client := &Client{
		url:            cfg.URL,
		reconnectDelay: cfg.ReconnectDelay,
		maxRetries:     cfg.MaxRetries,
	}

	if err := client.connect(); err != nil {
		return nil, err
	}

	// 启动连接监控
	go client.watchConnection()

	return client, nil
}

// connect 建立连接
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	c.conn, err = amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 设置QoS
	if err := c.channel.Qos(10, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// 声明交换机
	if err := c.declareExchanges(); err != nil {
		return err
	}

	// 声明队列
	if err := c.declareQueues(); err != nil {
		return err
	}

	logger.Info("RabbitMQ connected successfully")
	return nil
}

// declareExchanges 声明交换机
func (c *Client) declareExchanges() error {
	exchanges := []struct {
		name string
		kind string
	}{
		{ExchangeDirect, "direct"},
		{ExchangeTopic, "topic"},
		{ExchangeDLX, "direct"},
	}

	for _, ex := range exchanges {
		if err := c.channel.ExchangeDeclare(
			ex.name,
			ex.kind,
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,
		); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.name, err)
		}
	}
	return nil
}

// declareQueues 声明队列
func (c *Client) declareQueues() error {
	queues := []string{
		QueueEnvCreate,
		QueueEnvDestroy,
		QueueNotification,
		QueueVulnSync,
		QueueVulnConvert,
		QueueEmailSend,
	}

	for _, queueName := range queues {
		// 死信队列
		dlqName := queueName + ".dlq"
		if _, err := c.channel.QueueDeclare(
			dlqName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,
		); err != nil {
			return fmt.Errorf("failed to declare DLQ %s: %w", dlqName, err)
		}

		// 绑定死信队列到DLX
		if err := c.channel.QueueBind(dlqName, queueName, ExchangeDLX, false, nil); err != nil {
			return fmt.Errorf("failed to bind DLQ %s: %w", dlqName, err)
		}

		// 主队列（带死信配置）
		args := amqp.Table{
			"x-dead-letter-exchange":    ExchangeDLX,
			"x-dead-letter-routing-key": queueName,
		}
		if _, err := c.channel.QueueDeclare(
			queueName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			args,
		); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}

		// 绑定到交换机
		if err := c.channel.QueueBind(queueName, queueName, ExchangeDirect, false, nil); err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}
	return nil
}

// watchConnection 监控连接状态并自动重连
func (c *Client) watchConnection() {
	for {
		c.mu.RLock()
		if c.closed {
			c.mu.RUnlock()
			return
		}
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			time.Sleep(c.reconnectDelay)
			continue
		}

		// 等待连接关闭通知
		notifyClose := conn.NotifyClose(make(chan *amqp.Error, 1))
		err := <-notifyClose

		c.mu.RLock()
		if c.closed {
			c.mu.RUnlock()
			return
		}
		c.mu.RUnlock()

		if err != nil {
			logger.Warn("RabbitMQ connection closed", zap.Error(err))
		}

		// 尝试重连
		for i := 0; i < c.maxRetries; i++ {
			logger.Info("Attempting to reconnect to RabbitMQ", zap.Int("attempt", i+1))
			time.Sleep(c.reconnectDelay)

			if err := c.connect(); err != nil {
				logger.Error("Failed to reconnect", zap.Error(err))
				continue
			}
			break
		}
	}
}

// Publish 发布消息
func (c *Client) Publish(ctx context.Context, queue string, msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return fmt.Errorf("channel is not available")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.channel.PublishWithContext(
		ctx,
		ExchangeDirect,
		queue,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			MessageId:    msg.ID,
			Timestamp:    msg.CreatedAt,
		},
	)
}

// Consume 消费消息
func (c *Client) Consume(queue string, handler func(*Message) error) error {
	c.mu.RLock()
	channel := c.channel
	c.mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("channel is not available")
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
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			var msg Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				logger.Error("Failed to unmarshal message", zap.Error(err))
				d.Nack(false, false) // 发送到死信队列
				continue
			}

			if err := handler(&msg); err != nil {
				logger.Error("Failed to handle message",
					zap.String("queue", queue),
					zap.String("msg_id", msg.ID),
					zap.Error(err),
				)
				d.Nack(false, false) // 发送到死信队列
				continue
			}

			d.Ack(false)
		}
	}()

	return nil
}

// Close 关闭连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && !c.conn.IsClosed()
}
