package services

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"gorm.io/gorm"
)

type notificationService struct {
	db            *gorm.DB
	senderFactory SenderFactory
}

func NewNotificationService(db *gorm.DB, factory SenderFactory) NotificationService {
	return &notificationService{
		db:            db,
		senderFactory: factory,
	}
}

func (s *notificationService) ProcessMergeRequest(webhookData *models.GitLabWebhookData) error {
	if webhookData.ObjectAttributes.State != "opened" {
		return nil
	}

	var project models.Project
	if err := s.db.Where(&models.Project{GitLabProjectID: webhookData.Project.ID}).First(&project).Error; err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	if err := s.db.Preload("Webhooks").Preload("Webhooks.Settings").First(&project, project.ID).Error; err != nil {
		return fmt.Errorf("failed to load project webhooks: %w", err)
	}

	assigneeInfo, assigneeEmails := buildAssigneeInfo(webhookData)

	mentionedMobiles, err := s.lookupMentionedMobiles(assigneeInfo)
	if err != nil {
		logger.GetLogger().Warnf("查询指派人手机号失败: %v", err)
	}

	authorEmail := webhookData.User.Email
	if authorEmail == "[REDACTED]" {
		authorEmail = webhookData.User.Name
	}

	payload := &MergeRequestPayload{
		ProjectName:       project.Name,
		SourceBranch:      webhookData.ObjectAttributes.SourceBranch,
		TargetBranch:      webhookData.ObjectAttributes.TargetBranch,
		AuthorName:        webhookData.User.Name,
		Title:             webhookData.ObjectAttributes.Title,
		URL:               webhookData.ObjectAttributes.URL,
		MentionedMobiles:  mentionedMobiles,
		MentionedAccounts: assigneeEmails,
		Assignees:         assigneeInfo,
	}

	notification := &models.Notification{
		ProjectID:      project.ID,
		MergeRequestID: webhookData.ObjectAttributes.IID,
		Title:          webhookData.ObjectAttributes.Title,
		SourceBranch:   webhookData.ObjectAttributes.SourceBranch,
		TargetBranch:   webhookData.ObjectAttributes.TargetBranch,
		AuthorEmail:    authorEmail,
		Status:         webhookData.ObjectAttributes.State,
	}

	if len(assigneeEmails) > 0 {
		if emailsJSON, err := json.Marshal(assigneeEmails); err == nil {
			notification.AssigneeEmails = string(emailsJSON)
		}
	}

	if err := s.sendNotifications(context.Background(), &project, payload); err != nil {
		notification.ErrorMessage = err.Error()
		notification.NotificationSent = false
	} else {
		notification.NotificationSent = true
	}

	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	return nil
}

func (s *notificationService) sendNotifications(ctx context.Context, project *models.Project, payload *MergeRequestPayload) error {
	logger.GetLogger().Infof("开始处理通知发送 - 项目: %s", project.Name)

	if len(payload.Assignees) > 0 {
		logger.GetLogger().Infof("从 GitLab webhook 获取到 %d 个指派人", len(payload.Assignees))
		for i, info := range payload.Assignees {
			logger.GetLogger().Infof("  指派人 %d: 邮箱=%s, 用户名=%s", i+1, info.Email, info.Username)
		}
	} else {
		logger.GetLogger().Warnf("没有找到指派人信息")
	}

	logger.GetLogger().Infof("最终获得 %d 个手机号用于@功能", len(payload.MentionedMobiles))
	for i, mobile := range payload.MentionedMobiles {
		logger.GetLogger().Infof("  手机号 %d: %s", i+1, mobile)
	}

	sentWebhooks := make(map[uint]bool)
	for _, webhook := range project.Webhooks {
		if !webhook.IsActive {
			continue
		}
		if sentWebhooks[webhook.ID] {
			continue
		}

		webhook.ApplyDefaults()
		sender, err := s.senderFactory.SenderFor(&webhook)
		if err != nil {
			return fmt.Errorf("failed to find sender for webhook %d: %w", webhook.ID, err)
		}

		if err := sender.Send(ctx, &webhook, payload); err != nil {
			return fmt.Errorf("failed to send via webhook %s (%d): %w", webhook.Name, webhook.ID, err)
		}

		sentWebhooks[webhook.ID] = true
	}

	return nil
}

func (s *notificationService) lookupMentionedMobiles(assignees []models.AssigneeInfo) ([]string, error) {
	if len(assignees) == 0 {
		return nil, nil
	}

	var emailList []string
	var usernameList []string
	for _, info := range assignees {
		if info.Email != "" && info.Email != "[REDACTED]" {
			emailList = append(emailList, info.Email)
		} else if info.Username != "" {
			usernameList = append(usernameList, info.Username)
		}
	}

	phoneMap := make(map[string]bool)
	var mentionedMobiles []string

	if len(emailList) > 0 {
		var users []models.User
		if err := s.db.Where("email IN ?", emailList).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			if user.Phone != "" && !phoneMap[user.Phone] {
				mentionedMobiles = append(mentionedMobiles, user.Phone)
				phoneMap[user.Phone] = true
			}
		}
	}

	if len(usernameList) > 0 {
		var users []models.User
		if err := s.db.Where("gitlab_username IN ?", usernameList).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			if user.Phone != "" && !phoneMap[user.Phone] {
				mentionedMobiles = append(mentionedMobiles, user.Phone)
				phoneMap[user.Phone] = true
			}
		}
	}

	return mentionedMobiles, nil
}

func buildAssigneeInfo(webhookData *models.GitLabWebhookData) ([]models.AssigneeInfo, []string) {
	info := make([]models.AssigneeInfo, len(webhookData.Assignees))
	emails := make([]string, len(webhookData.Assignees))
	for i, assignee := range webhookData.Assignees {
		email := assignee.Email
		if email == "[REDACTED]" {
			email = assignee.Name
		}
		info[i] = models.AssigneeInfo{
			Email:    email,
			Username: assignee.Username,
		}
		emails[i] = email
	}
	return info, emails
}

func (s *notificationService) GetAllNotifications() ([]models.NotificationResponse, error) {
	var notifications []models.Notification
	if err := s.db.Preload("Project").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return buildNotificationResponses(notifications)
}

func (s *notificationService) GetNotificationsByProjectID(projectID uint) ([]models.NotificationResponse, error) {
	var notifications []models.Notification
	if err := s.db.Where("project_id = ?", projectID).Preload("Project").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return buildNotificationResponses(notifications)
}

func (s *notificationService) GetRecentNotifications(limit int) ([]models.NotificationResponse, error) {
	var notifications []models.Notification
	if err := s.db.Preload("Project").Order("created_at desc").Limit(limit).Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent notifications: %w", err)
	}

	return buildNotificationResponses(notifications)
}

func (s *notificationService) GetNotificationStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCount int64
	if err := s.db.Model(&models.Notification{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_notifications"] = totalCount

	var successCount int64
	if err := s.db.Model(&models.Notification{}).Where("notification_sent = ?", true).Count(&successCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get success count: %w", err)
	}
	stats["success_notifications"] = successCount

	var failureCount int64
	if err := s.db.Model(&models.Notification{}).Where("notification_sent = ?", false).Count(&failureCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get failure count: %w", err)
	}
	stats["failure_notifications"] = failureCount

	var todayCount int64
	if err := s.db.Model(&models.Notification{}).Where("DATE(created_at) = CURRENT_DATE").Count(&todayCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get today count: %w", err)
	}
	stats["today_notifications"] = todayCount

	return stats, nil
}

func buildNotificationResponses(notifications []models.Notification) ([]models.NotificationResponse, error) {
	responses := make([]models.NotificationResponse, 0, len(notifications))

	for _, notification := range notifications {
		var assigneeEmails []string
		if notification.AssigneeEmails != "" {
			if err := json.Unmarshal([]byte(notification.AssigneeEmails), &assigneeEmails); err != nil {
				logger.GetLogger().Warnf("Failed to unmarshal assignee emails: %v, raw: %s", err, notification.AssigneeEmails)
			}
		}

		responses = append(responses, models.NotificationResponse{
			ID:               notification.ID,
			ProjectID:        notification.ProjectID,
			ProjectName:      notification.Project.Name,
			MergeRequestID:   notification.MergeRequestID,
			Title:            notification.Title,
			SourceBranch:     notification.SourceBranch,
			TargetBranch:     notification.TargetBranch,
			AuthorEmail:      notification.AuthorEmail,
			AssigneeEmails:   assigneeEmails,
			Status:           notification.Status,
			NotificationSent: notification.NotificationSent,
			ErrorMessage:     notification.ErrorMessage,
			CreatedAt:        notification.CreatedAt,
		})
	}

	return responses, nil
}
