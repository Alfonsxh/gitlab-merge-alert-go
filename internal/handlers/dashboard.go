package handlers

import (
	"net/http"
	"time"

	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/models"

	"github.com/gin-gonic/gin"
)

type Stats struct {
	TotalUsers              int64 `json:"total_users"`
	TotalProjects           int64 `json:"total_projects"`
	TotalWebhooks           int64 `json:"total_webhooks"`
	TotalNotifications      int64 `json:"total_notifications"`
	RecentNotifications     int64 `json:"recent_notifications"`
	SuccessfulNotifications int64 `json:"successful_notifications"`
}

func (h *Handler) GetStats(c *gin.Context) {
	var stats Stats

	// 应用所有权过滤
	userQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.User{}), "users")
	projectQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.Project{}), "projects")
	webhookQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.Webhook{}), "webhooks")
	notificationQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.Notification{}), "notifications")

	// 统计用户数量
	userQuery.Count(&stats.TotalUsers)

	// 统计项目数量
	projectQuery.Count(&stats.TotalProjects)

	// 统计Webhook数量
	webhookQuery.Count(&stats.TotalWebhooks)

	// 统计通知总数
	notificationQuery.Count(&stats.TotalNotifications)

	// 统计最近24小时的通知数量
	yesterday := time.Now().Add(-24 * time.Hour)
	recentQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.Notification{}), "notifications")
	recentQuery.Where("created_at > ?", yesterday).Count(&stats.RecentNotifications)

	// 统计成功发送的通知数量
	successQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.Notification{}), "notifications")
	successQuery.Where("notification_sent = ?", true).Count(&stats.SuccessfulNotifications)

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *Handler) GetNotifications(c *gin.Context) {
	var notifications []models.Notification
	
	// 应用所有权过滤
	query := h.db.Preload("Project").Order("created_at DESC")
	query = middleware.ApplyOwnershipFilter(c, query, "notifications")

	// 简单分页实现
	query = query.Limit(20)

	if err := query.Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	// 转换为响应格式
	responses := make([]models.NotificationResponse, 0) // 确保是空数组而不是nil
	for _, notification := range notifications {
		var assigneeEmails []string
		// 简单解析JSON字符串，实际应该使用json.Unmarshal
		if notification.AssigneeEmails != "" {
			assigneeEmails = []string{notification.AssigneeEmails}
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

	c.JSON(http.StatusOK, gin.H{"data": responses})
}
