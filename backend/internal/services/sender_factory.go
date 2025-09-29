package services

import (
	"fmt"
	"strings"

	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/models"
	"gorm.io/gorm"
)

type messageSenderFactory struct {
	wecom    MessageSender
	dingtalk MessageSender
	custom   MessageSender
}

func NewMessageSenderFactory(db *gorm.DB, cfg *config.Config, wechatService WeChatService) SenderFactory {
	dingTalkSender := NewDingTalkSender(db, cfg.Notification.DingTalk)
	return &messageSenderFactory{
		wecom:    NewWeComSender(wechatService),
		dingtalk: dingTalkSender,
		custom:   NewCustomSender(),
	}
}

func (f *messageSenderFactory) SenderFor(webhook *models.Webhook) (MessageSender, error) {
	if webhook == nil {
		return nil, fmt.Errorf("nil webhook")
	}

	channel := strings.ToLower(strings.TrimSpace(webhook.Type))
	if channel == "" || channel == models.WebhookTypeAuto {
		channel = models.DetectWebhookType(webhook.URL)
	}

	switch channel {
	case models.WebhookTypeDingTalk:
		return f.dingtalk, nil
	case models.WebhookTypeCustom:
		return f.custom, nil
	case models.WebhookTypeWeCom:
		fallthrough
	default:
		return f.wecom, nil
	}
}
