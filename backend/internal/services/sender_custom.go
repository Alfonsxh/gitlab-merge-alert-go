package services

import (
	"context"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"
)

type CustomSender struct{}

func NewCustomSender() *CustomSender {
	return &CustomSender{}
}

func (s *CustomSender) Send(ctx context.Context, webhook *models.Webhook, payload *MergeRequestPayload) error {
	logger.GetLogger().Infof("自定义 webhook (%d) 使用 GitLab 原生通知，请确保已在 GitLab 配置该地址: %s", webhook.ID, webhook.URL)
	return nil
}
