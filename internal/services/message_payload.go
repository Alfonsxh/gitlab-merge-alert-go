package services

import (
	"context"

	"github.com/Alfonsxh/gitlab-merge-alert-go/internal/models"
)

type MergeRequestPayload struct {
	ProjectName       string
	SourceBranch      string
	TargetBranch      string
	AuthorName        string
	Title             string
	URL               string
	MentionedMobiles  []string
	MentionedAccounts []string
	Assignees         []models.AssigneeInfo
}

type MessageSender interface {
	Send(ctx context.Context, webhook *models.Webhook, payload *MergeRequestPayload) error
}

type SenderFactory interface {
	SenderFor(webhook *models.Webhook) (MessageSender, error)
}
