package services

import (
	"encoding/json"
	"fmt"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

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
	if err := s.db.Where(&models.Project{GitLabProjectID: webhookData.Project.ID}).First(&project).Error; err != nil {
		return fmt.Errorf("project not found: %v", err)
	}

	// 获取项目关联的webhooks
	if err := s.db.Preload("Webhooks").First(&project, project.ID).Error; err != nil {
		return fmt.Errorf("failed to load project webhooks: %v", err)
	}

	// 获取指派人信息（邮箱和用户名）
	assigneeInfo := make([]models.AssigneeInfo, len(webhookData.Assignees))
	assigneeEmails := make([]string, len(webhookData.Assignees))
	for i, assignee := range webhookData.Assignees {
		assigneeInfo[i] = models.AssigneeInfo{
			Email:    assignee.Email,
			Username: assignee.Username,
		}
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

	// 将邮箱数组转换为JSON字符串（保持向后兼容）
	if len(assigneeEmails) > 0 {
		emailsJSON, _ := json.Marshal(assigneeEmails)
		notification.AssigneeEmails = string(emailsJSON)
	}

	// 发送通知
	if err := s.sendNotifications(&project, webhookData, assigneeInfo); err != nil {
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

func (s *NotificationService) sendNotifications(project *models.Project, webhookData *models.GitLabWebhookData, assigneeInfo []models.AssigneeInfo) error {
	logger.GetLogger().Infof("开始处理通知发送 - 项目: %s", project.Name)

	// 记录从 webhook 获取的指派人信息
	if len(assigneeInfo) > 0 {
		logger.GetLogger().Infof("从 GitLab webhook 获取到 %d 个指派人:", len(assigneeInfo))
		for i, info := range assigneeInfo {
			logger.GetLogger().Infof("  指派人 %d: 邮箱=%s, 用户名=%s", i+1, info.Email, info.Username)
		}
	} else {
		logger.GetLogger().Warnf("没有找到指派人信息")
	}

	// 获取指派人的手机号
	var mentionedMobiles []string
	var assigneeEmails []string // 用于消息格式化，保持向后兼容

	if len(assigneeInfo) > 0 {
		// 准备查询条件：优先使用邮箱，如果邮箱被脱敏则使用GitLab用户名
		var emailList []string
		var usernameList []string

		for _, info := range assigneeInfo {
			assigneeEmails = append(assigneeEmails, info.Email) // 保持向后兼容
			if info.Email != "" && info.Email != "[REDACTED]" {
				emailList = append(emailList, info.Email)
			} else if info.Username != "" {
				usernameList = append(usernameList, info.Username)
			}
		}

		var users []models.User
		var err error

		// 先通过邮箱查询用户
		if len(emailList) > 0 {
			if err = s.db.Where("email IN ?", emailList).Find(&users).Error; err != nil {
				logger.GetLogger().Errorf("通过邮箱查询用户数据库失败: %v", err)
			}
		}

		// 再通过GitLab用户名查询用户
		if len(usernameList) > 0 {
			var usernameUsers []models.User
			if err = s.db.Where("gitlab_username IN ?", usernameList).Find(&usernameUsers).Error; err != nil {
				logger.GetLogger().Errorf("通过GitLab用户名查询用户数据库失败: %v", err)
			} else {
				users = append(users, usernameUsers...)
			}
		}

		if len(users) > 0 {
			logger.GetLogger().Infof("从数据库中找到 %d 个匹配的用户记录:", len(users))
			phoneMap := make(map[string]bool) // 去重
			for i, user := range users {
				logger.GetLogger().Infof("  用户 %d: 邮箱=%s, GitLab用户名=%s, 手机号=%s, 姓名=%s",
					i+1, user.Email, user.GitLabUsername, user.Phone, user.Name)
				if user.Phone != "" && !phoneMap[user.Phone] {
					mentionedMobiles = append(mentionedMobiles, user.Phone)
					phoneMap[user.Phone] = true
				} else if user.Phone == "" {
					logger.GetLogger().Warnf("用户 %s (GitLab: %s) 的手机号为空，无法进行@", user.Email, user.GitLabUsername)
				}
			}

			// 检查是否有指派人没有找到对应用户
			for _, info := range assigneeInfo {
				found := false
				for _, user := range users {
					if (info.Email != "" && info.Email != "[REDACTED]" && user.Email == info.Email) ||
						(info.Username != "" && user.GitLabUsername == info.Username) {
						found = true
						break
					}
				}
				if !found {
					if info.Email != "" && info.Email != "[REDACTED]" {
						logger.GetLogger().Warnf("邮箱 %s 在用户数据库中没有找到对应记录", info.Email)
					} else if info.Username != "" {
						logger.GetLogger().Warnf("GitLab用户名 %s 在用户数据库中没有找到对应记录", info.Username)
					}
				}
			}
		}
	}

	logger.GetLogger().Infof("最终获得 %d 个手机号用于@功能:", len(mentionedMobiles))
	for i, mobile := range mentionedMobiles {
		logger.GetLogger().Infof("  手机号 %d: %s", i+1, mobile)
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

	// 发送到所有关联的webhook，使用去重逻辑防止重复发送
	sentWebhooks := make(map[uint]bool) // 记录已发送的webhook ID
	for _, webhook := range project.Webhooks {
		if !webhook.IsActive {
			continue
		}

		// 检查是否已经发送过这个webhook
		if sentWebhooks[webhook.ID] {
			continue
		}

		if err := s.wechatService.SendMessage(webhook.URL, content, mentionedMobiles); err != nil {
			return fmt.Errorf("failed to send to webhook %s: %v", webhook.Name, err)
		}

		// 标记这个webhook已发送
		sentWebhooks[webhook.ID] = true
	}

	return nil
}
