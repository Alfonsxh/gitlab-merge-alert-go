package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

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
		// 解析JSON字符串为数组
		if notification.AssigneeEmails != "" {
			if err := json.Unmarshal([]byte(notification.AssigneeEmails), &assigneeEmails); err != nil {
				// 解析失败时记录日志但继续处理
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

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// DailyStats represents daily statistics
type DailyStats struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ProjectDailyStats represents daily stats for a project
type ProjectDailyStats struct {
	ProjectID   uint         `json:"project_id"`
	ProjectName string       `json:"project_name"`
	Data        []DailyStats `json:"data"`
}

// WebhookDailyStats represents daily stats for a webhook
type WebhookDailyStats struct {
	WebhookID   uint         `json:"webhook_id"`
	WebhookName string       `json:"webhook_name"`
	Data        []DailyStats `json:"data"`
}

// GetProjectDailyStats returns daily merge request statistics per project
func (h *Handler) GetProjectDailyStats(c *gin.Context) {
	// Get days parameter (default 7 days)
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := time.ParseDuration(d + "d"); err == nil {
			days = int(parsed.Hours() / 24)
		}
	}

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Get all projects
	var projects []models.Project
	projectQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.Project{}), "projects")
	if err := projectQuery.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	// Prepare result
	result := make([]ProjectDailyStats, 0)

	// For each project, get daily stats
	for _, project := range projects {
		var dailyStats []DailyStats

		// Query notifications grouped by date
		query := h.db.Model(&models.Notification{}).
			Select("DATE(created_at) as date, COUNT(*) as count").
			Where("project_id = ? AND created_at >= ? AND created_at <= ?", project.ID, startDate, endDate).
			Group("DATE(created_at)").
			Order("date ASC")

		// Apply ownership filter
		query = middleware.ApplyOwnershipFilter(c, query, "notifications")

		// Execute query
		rows, err := query.Rows()
		if err != nil {
			continue
		}
		defer rows.Close()

		// Parse results
		for rows.Next() {
			var stat DailyStats
			if err := rows.Scan(&stat.Date, &stat.Count); err == nil {
				dailyStats = append(dailyStats, stat)
			}
		}

		// Fill missing dates with zero
		dailyStats = fillMissingDates(dailyStats, startDate, endDate)

		result = append(result, ProjectDailyStats{
			ProjectID:   project.ID,
			ProjectName: project.Name,
			Data:        dailyStats,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetWebhookDailyStats returns daily merge request statistics per webhook
func (h *Handler) GetWebhookDailyStats(c *gin.Context) {
	// Get days parameter (default 7 days)
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := time.ParseDuration(d + "d"); err == nil {
			days = int(parsed.Hours() / 24)
		}
	}

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Get all webhooks
	var webhooks []models.Webhook
	webhookQuery := middleware.ApplyOwnershipFilter(c, h.db.Model(&models.Webhook{}), "webhooks")
	if err := webhookQuery.Find(&webhooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}

	// Prepare result
	result := make([]WebhookDailyStats, 0)

	// For each webhook, get daily stats through project associations
	for _, webhook := range webhooks {
		var dailyStats []DailyStats

		// Get projects associated with this webhook
		var projectIDs []uint
		h.db.Table("project_webhooks").
			Where("webhook_id = ?", webhook.ID).
			Pluck("project_id", &projectIDs)

		if len(projectIDs) > 0 {
			// Query notifications grouped by date for projects associated with this webhook
			query := h.db.Model(&models.Notification{}).
				Select("DATE(created_at) as date, COUNT(*) as count").
				Where("project_id IN ? AND created_at >= ? AND created_at <= ?", projectIDs, startDate, endDate).
				Group("DATE(created_at)").
				Order("date ASC")

			// Apply ownership filter
			query = middleware.ApplyOwnershipFilter(c, query, "notifications")

			// Execute query
			rows, err := query.Rows()
			if err != nil {
				continue
			}
			defer rows.Close()

			// Parse results
			for rows.Next() {
				var stat DailyStats
				if err := rows.Scan(&stat.Date, &stat.Count); err == nil {
					dailyStats = append(dailyStats, stat)
				}
			}
		}

		// Fill missing dates with zero
		dailyStats = fillMissingDates(dailyStats, startDate, endDate)

		result = append(result, WebhookDailyStats{
			WebhookID:   webhook.ID,
			WebhookName: webhook.Name,
			Data:        dailyStats,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// fillMissingDates fills missing dates in the stats with zero values
func fillMissingDates(stats []DailyStats, startDate, endDate time.Time) []DailyStats {
	// Create a map for quick lookup
	statsMap := make(map[string]int64)
	for _, stat := range stats {
		statsMap[stat.Date] = stat.Count
	}

	// Create complete date range
	var result []DailyStats
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		count := int64(0)
		if c, exists := statsMap[dateStr]; exists {
			count = c
		}
		result = append(result, DailyStats{
			Date:  dateStr,
			Count: count,
		})
	}

	return result
}
