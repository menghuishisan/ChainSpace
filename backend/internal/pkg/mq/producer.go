package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Producer 消息生产者
type Producer struct {
	client *Client
}

// NewProducer 创建生产者
func NewProducer(client *Client) *Producer {
	return &Producer{client: client}
}

// PublishEnvCreate 发布环境创建任务
func (p *Producer) PublishEnvCreate(ctx context.Context, experimentID, userID, schoolID uint, snapshotURL string) error {
	msg := &Message{
		ID:        uuid.New().String(),
		Type:      "env.create",
		CreatedAt: time.Now(),
		Payload: map[string]interface{}{
			"experiment_id": experimentID,
			"user_id":       userID,
			"school_id":     schoolID,
			"snapshot_url":  snapshotURL,
		},
	}
	return p.client.Publish(ctx, QueueEnvCreate, msg)
}

// PublishEnvDestroy 发布环境销毁任务
func (p *Producer) PublishEnvDestroy(ctx context.Context, envID string) error {
	msg := &Message{
		ID:        uuid.New().String(),
		Type:      "env.destroy",
		CreatedAt: time.Now(),
		Payload: map[string]interface{}{
			"env_id": envID,
		},
	}
	return p.client.Publish(ctx, QueueEnvDestroy, msg)
}

// PublishNotification 发布通知任务
func (p *Producer) PublishNotification(ctx context.Context, userID uint, notifyType, title, content string, extra map[string]interface{}) error {
	msg := &Message{
		ID:        uuid.New().String(),
		Type:      "notification",
		CreatedAt: time.Now(),
		Payload: map[string]interface{}{
			"user_id": userID,
			"type":    notifyType,
			"title":   title,
			"content": content,
			"extra":   extra,
		},
	}
	return p.client.Publish(ctx, QueueNotification, msg)
}

// PublishVulnSync 发布漏洞同步任务
func (p *Producer) PublishVulnSync(ctx context.Context, sourceID uint) error {
	msg := &Message{
		ID:        uuid.New().String(),
		Type:      "vuln.sync",
		CreatedAt: time.Now(),
		Payload: map[string]interface{}{
			"source_id": sourceID,
		},
	}
	return p.client.Publish(ctx, QueueVulnSync, msg)
}

// PublishVulnConvert 发布漏洞转化任务
func (p *Producer) PublishVulnConvert(ctx context.Context, vulnID uint) error {
	msg := &Message{
		ID:        uuid.New().String(),
		Type:      "vuln.convert",
		CreatedAt: time.Now(),
		Payload: map[string]interface{}{
			"vuln_id": vulnID,
		},
	}
	return p.client.Publish(ctx, QueueVulnConvert, msg)
}

// PublishEmail 发布邮件发送任务
func (p *Producer) PublishEmail(ctx context.Context, to, subject, body string) error {
	msg := &Message{
		ID:        uuid.New().String(),
		Type:      "email.send",
		CreatedAt: time.Now(),
		Payload: map[string]interface{}{
			"to":      to,
			"subject": subject,
			"body":    body,
		},
	}
	return p.client.Publish(ctx, QueueEmailSend, msg)
}
