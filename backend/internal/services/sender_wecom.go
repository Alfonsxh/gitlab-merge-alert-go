package services

import (
	"context"

	"gitlab-merge-alert-go/internal/models"
)

type WeComSender struct {
	service WeChatService
}

func NewWeComSender(service WeChatService) *WeComSender {
	return &WeComSender{service: service}
}

func (s *WeComSender) Send(ctx context.Context, webhook *models.Webhook, payload *MergeRequestPayload) error {
	if payload == nil {
		return nil
	}
	content := s.service.FormatMergeRequestMessage(
		payload.ProjectName,
		payload.SourceBranch,
		payload.TargetBranch,
		payload.AuthorName,
		payload.Title,
		payload.URL,
		payload.MentionedAccounts,
	)

	return s.service.SendMessage(webhook.URL, content, payload.MentionedMobiles)
}
