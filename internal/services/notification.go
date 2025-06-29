package services

import (
	"encoding/json"
	"fmt"

	"gitlab-merge-alert-go/internal/models"

	"gorm.io/gorm"
)

type NotificationService struct {
	db            *gorm.DB
	wechatService *WeChatService
}

func NewNotificationService(db *gorm.DB, wechatService *WeChatService) *NotificationService {
	return &NotificationService{
		db:            db,
		wechatService: wechatService,
	}
}

func (s *NotificationService) ProcessMergeRequest(webhookData *models.GitLabWebhookData) error {
	// 只处理打开状态的合并请求
	if webhookData.ObjectAttributes.State != "opened" {
		return nil
	}

	// 查找项目
	var project models.Project
	if err := s.db.Where("gitlab_project_id = ?", webhookData.Project.ID).First(&project).Error; err != nil {
		return fmt.Errorf("project not found: %v", err)
	}

	// 获取项目关联的webhooks
	if err := s.db.Preload("Webhooks").First(&project, project.ID).Error; err != nil {
		return fmt.Errorf("failed to load project webhooks: %v", err)
	}

	// 获取指派人的邮箱列表
	assigneeEmails := make([]string, len(webhookData.Assignees))
	for i, assignee := range webhookData.Assignees {
		assigneeEmails[i] = assignee.Email
	}

	// 记录通知
	notification := &models.Notification{
		ProjectID:      project.ID,
		MergeRequestID: webhookData.ObjectAttributes.ID,
		Title:          webhookData.ObjectAttributes.Title,
		SourceBranch:   webhookData.ObjectAttributes.SourceBranch,
		TargetBranch:   webhookData.ObjectAttributes.TargetBranch,
		AuthorEmail:    webhookData.User.Email,
		Status:         webhookData.ObjectAttributes.State,
	}

	// 将邮箱数组转换为JSON字符串
	if len(assigneeEmails) > 0 {
		emailsJSON, _ := json.Marshal(assigneeEmails)
		notification.AssigneeEmails = string(emailsJSON)
	}

	// 发送通知
	if err := s.sendNotifications(&project, webhookData, assigneeEmails); err != nil {
		notification.ErrorMessage = err.Error()
		notification.NotificationSent = false
	} else {
		notification.NotificationSent = true
	}

	// 保存通知记录
	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	return nil
}

func (s *NotificationService) sendNotifications(project *models.Project, webhookData *models.GitLabWebhookData, assigneeEmails []string) error {
	// 获取指派人的手机号
	var mentionedMobiles []string
	if len(assigneeEmails) > 0 {
		var users []models.User
		if err := s.db.Where("email IN ?", assigneeEmails).Find(&users).Error; err == nil {
			for _, user := range users {
				mentionedMobiles = append(mentionedMobiles, user.Phone)
			}
		}
	}

	// 格式化消息内容
	content := s.wechatService.FormatMergeRequestMessage(
		project.Name,
		webhookData.ObjectAttributes.SourceBranch,
		webhookData.ObjectAttributes.TargetBranch,
		webhookData.User.Name,
		webhookData.ObjectAttributes.Title,
		webhookData.ObjectAttributes.URL,
		assigneeEmails,
	)

	// 发送到所有关联的webhook
	for _, webhook := range project.Webhooks {
		if !webhook.IsActive {
			continue
		}

		if err := s.wechatService.SendMessage(webhook.URL, content, mentionedMobiles); err != nil {
			return fmt.Errorf("failed to send to webhook %s: %v", webhook.Name, err)
		}
	}

	return nil
}
