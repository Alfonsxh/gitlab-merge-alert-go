package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) GetWebhooks(c *gin.Context) {
	var webhooks []models.Webhook

	// 应用所有权过滤
	query := h.db.Preload("Projects")
	query = middleware.ApplyOwnershipFilter(c, query, "webhooks")

	if err := query.Find(&webhooks).Error; err != nil {
		logger.GetLogger().Errorf("Failed to fetch webhooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}

	logger.GetLogger().Debugf("Successfully fetched %d webhooks", len(webhooks))

	// 转换为响应格式
	responses := make([]models.WebhookResponse, 0, len(webhooks))
	for _, webhook := range webhooks {
		response := models.WebhookResponse{
			ID:          webhook.ID,
			Name:        webhook.Name,
			URL:         webhook.URL,
			Description: webhook.Description,
			IsActive:    webhook.IsActive,
			CreatedAt:   webhook.CreatedAt,
			UpdatedAt:   webhook.UpdatedAt,
		}

		// 转换关联的项目
		for _, project := range webhook.Projects {
			response.Projects = append(response.Projects, models.ProjectResponse{
				ID:              project.ID,
				GitLabProjectID: project.GitLabProjectID,
				Name:            project.Name,
				URL:             project.URL,
				Description:     project.Description,
				CreatedAt:       project.CreatedAt,
				UpdatedAt:       project.UpdatedAt,
			})
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *Handler) CreateWebhook(c *gin.Context) {
	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	accountID, _ := middleware.GetAccountID(c)

	webhook := &models.Webhook{
		Name:        req.Name,
		URL:         req.URL,
		Description: req.Description,
		IsActive:    true,
		CreatedBy:   &accountID,
	}

	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := h.db.Create(webhook).Error; err != nil {
		logger.GetLogger().Errorf("Failed to create webhook [Name: %s, URL: %s]: %v", req.Name, req.URL, err)

		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "Webhook名称或URL已存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建Webhook失败"})
		}
		return
	}

	logger.GetLogger().Infof("Successfully created webhook [ID: %d, Name: %s]", webhook.ID, webhook.Name)

	response := models.WebhookResponse{
		ID:          webhook.ID,
		Name:        webhook.Name,
		URL:         webhook.URL,
		Description: webhook.Description,
		IsActive:    webhook.IsActive,
		CreatedAt:   webhook.CreatedAt,
		UpdatedAt:   webhook.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

func (h *Handler) UpdateWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var req models.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var webhook models.Webhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Warnf("Webhook not found [ID: %d]", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			logger.GetLogger().Errorf("Failed to fetch webhook [ID: %d]: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 更新字段
	if req.Name != "" {
		webhook.Name = req.Name
	}
	if req.URL != "" {
		webhook.URL = req.URL
	}
	if req.Description != "" {
		webhook.Description = req.Description
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := h.db.Save(&webhook).Error; err != nil {
		logger.GetLogger().Errorf("Failed to update webhook [ID: %d]: %v", id, err)

		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "Webhook名称或URL已存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新Webhook失败"})
		}
		return
	}

	logger.GetLogger().Infof("Successfully updated webhook [ID: %d, Name: %s]", webhook.ID, webhook.Name)

	response := models.WebhookResponse{
		ID:          webhook.ID,
		Name:        webhook.Name,
		URL:         webhook.URL,
		Description: webhook.Description,
		IsActive:    webhook.IsActive,
		CreatedAt:   webhook.CreatedAt,
		UpdatedAt:   webhook.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *Handler) DeleteWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	if err := h.db.Delete(&models.Webhook{}, id).Error; err != nil {
		logger.GetLogger().Errorf("Failed to delete webhook [ID: %d]: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除Webhook失败"})
		return
	}

	logger.GetLogger().Infof("Successfully deleted webhook [ID: %d]", id)

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
}

func (h *Handler) LinkProjectWebhook(c *gin.Context) {
	var req struct {
		ProjectID uint `json:"project_id" binding:"required"`
		WebhookID uint `json:"webhook_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectID := uint64(req.ProjectID)
	webhookID := uint64(req.WebhookID)

	// 检查项目和webhook是否存在
	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var webhook models.Webhook
	if err := h.db.First(&webhook, webhookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// 检查是否已经关联，避免重复记录
	var existing []models.ProjectWebhook
	if err := h.db.Where("project_id = ? AND webhook_id = ?", project.ID, webhook.ID).Find(&existing).Error; err != nil {
		logger.GetLogger().Errorf("Failed to check existing project-webhook link [project_id=%d, webhook_id=%d]: %v", project.ID, webhook.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link project and webhook"})
		return
	}

	if len(existing) > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Project and webhook already linked"})
		return
	}

	association := &models.ProjectWebhook{
		ProjectID: project.ID,
		WebhookID: webhook.ID,
	}

	if err := h.db.Create(association).Error; err != nil {
		logger.GetLogger().Errorf("Failed to create project-webhook link [project_id=%d, webhook_id=%d]: %v", project.ID, webhook.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link project and webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project and webhook linked successfully"})
}

func (h *Handler) UnlinkProjectWebhook(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("project_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	webhookID, err := strconv.ParseUint(c.Param("webhook_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	// 检查项目是否存在
	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var webhook models.Webhook
	if err := h.db.First(&webhook, webhookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// 删除关联
	if err := h.db.Model(&project).Association("Webhooks").Delete(&webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink project and webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project and webhook unlinked successfully"})
}

func (h *Handler) SendTestMessage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	// 查找webhook
	var webhook models.Webhook
	query := h.db.Model(&models.Webhook{})
	query = middleware.ApplyOwnershipFilter(c, query, "webhooks")

	if err := query.First(&webhook, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Warnf("Webhook not found [ID: %d]", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			logger.GetLogger().Errorf("Failed to fetch webhook [ID: %d]: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 检查webhook是否启用
	if !webhook.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook is not active"})
		return
	}

	// 构建测试消息内容
	testMessage := fmt.Sprintf(
		"🔔 GitLab Merge Alert 测试消息\n\n"+
			"✅ Webhook连接测试成功！\n"+
			"📌 Webhook名称：%s\n"+
			"🕐 测试时间：%s\n\n"+
			"如果您看到这条消息，说明企业微信机器人配置正确。",
		webhook.Name,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// 发送测试消息
	if err := h.wechatService.SendMessage(webhook.URL, testMessage, nil); err != nil {
		logger.GetLogger().Errorf("Failed to send test message to webhook [ID: %d, Name: %s]: %v",
			webhook.ID, webhook.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "发送测试消息失败",
			"details": err.Error(),
		})
		return
	}

	// 记录通知历史
	accountID, _ := middleware.GetAccountID(c)
	notification := &models.Notification{
		Title:            "Webhook测试",
		Status:           "success",
		NotificationSent: true,
		OwnerID:          &accountID,
	}

	if err := h.db.Create(notification).Error; err != nil {
		logger.GetLogger().Warnf("Failed to save test notification history: %v", err)
		// 不影响主要功能，只记录警告
	}

	logger.GetLogger().Infof("Successfully sent test message to webhook [ID: %d, Name: %s]",
		webhook.ID, webhook.Name)

	c.JSON(http.StatusOK, gin.H{
		"message":      "测试消息发送成功",
		"webhook_name": webhook.Name,
		"sent_at":      time.Now().Format("2006-01-02 15:04:05"),
	})
}
