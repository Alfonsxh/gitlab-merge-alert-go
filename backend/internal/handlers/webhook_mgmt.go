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
	"gitlab-merge-alert-go/internal/services"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) GetWebhooks(c *gin.Context) {
	var webhooks []models.Webhook

	// 应用所有权过滤
	query := h.db.Preload("Projects").Preload("Settings")
	query = middleware.ApplyOwnershipFilter(c, query, "webhooks")

	if err := query.Find(&webhooks).Error; err != nil {
		logger.GetLogger().Errorf("Failed to fetch webhooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}

	logger.GetLogger().Debugf("Successfully fetched %d webhooks", len(webhooks))

	// 转换为响应格式
	responses := make([]models.WebhookResponse, 0, len(webhooks))
	for idx := range webhooks {
		responses = append(responses, buildWebhookResponse(&webhooks[idx]))
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

	targetURL := webhook.URL
	channel := strings.ToLower(strings.TrimSpace(req.Type))
	if channel == "" || channel == models.WebhookTypeAuto {
		channel = models.DetectWebhookType(targetURL)
	}

	signatureMethod := req.SignatureMethod
	if signatureMethod == "" {
		signatureMethod = models.SignatureMethodHMACSHA256
	}

	webhook.Type = channel
	webhook.ApplyDefaults()

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

	secret := req.Secret
	keywords := req.SecurityKeywords
	headers := req.CustomHeaders
	if err := h.upsertWebhookSettings(webhook.ID, &signatureMethod, &secret, &keywords, &headers); err != nil {
		logger.GetLogger().Errorf("Failed to persist webhook settings [WebhookID: %d]: %v", webhook.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建Webhook失败"})
		return
	}

	if err := h.db.Preload("Settings").Preload("Projects").First(webhook, webhook.ID).Error; err != nil {
		logger.GetLogger().Errorf("Failed to reload webhook [ID: %d]: %v", webhook.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建Webhook失败"})
		return
	}

	response := buildWebhookResponse(webhook)

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
	if err := h.db.Preload("Settings").First(&webhook, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Warnf("Webhook not found [ID: %d]", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		} else {
			logger.GetLogger().Errorf("Failed to fetch webhook [ID: %d]: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	targetURL := webhook.URL

	if req.Name != "" {
		webhook.Name = req.Name
	}
	if req.URL != "" {
		webhook.URL = req.URL
		targetURL = req.URL
	}
	if req.Description != "" {
		webhook.Description = req.Description
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if req.Type != "" {
		channel := strings.ToLower(strings.TrimSpace(req.Type))
		if channel == models.WebhookTypeAuto {
			channel = models.DetectWebhookType(targetURL)
		}
		webhook.Type = channel
	} else if req.URL != "" {
		webhook.Type = models.DetectWebhookType(targetURL)
	}

	var signaturePtr *string
	if req.SignatureMethod != "" {
		sig := req.SignatureMethod
		signaturePtr = &sig
	}

	var secretPtr *string
	if req.Secret != nil {
		secretPtr = req.Secret
	}

	var keywordsPtr *[]string
	if req.SecurityKeywords != nil {
		keywords := req.SecurityKeywords
		keywordsPtr = &keywords
	}

	var headersPtr *map[string]string
	if req.CustomHeaders != nil {
		headers := req.CustomHeaders
		headersPtr = &headers
	}

	webhook.ApplyDefaults()

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

	if err := h.upsertWebhookSettings(webhook.ID, signaturePtr, secretPtr, keywordsPtr, headersPtr); err != nil {
		logger.GetLogger().Errorf("Failed to update webhook settings [ID: %d]: %v", webhook.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新Webhook失败"})
		return
	}

	if err := h.db.Preload("Settings").Preload("Projects").First(&webhook, id).Error; err != nil {
		logger.GetLogger().Errorf("Failed to reload webhook [ID: %d]: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新Webhook失败"})
		return
	}

	response := buildWebhookResponse(&webhook)

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
	query := h.db.Model(&models.Webhook{}).Preload("Settings")
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

	channel := strings.ToLower(strings.TrimSpace(webhook.Type))
	if channel == "" || channel == models.WebhookTypeAuto {
		channel = models.DetectWebhookType(webhook.URL)
	}

	if channel == models.WebhookTypeCustom {
		c.JSON(http.StatusOK, gin.H{
			"message":      "自定义 Webhook 不支持在平台内测试，请在 GitLab 中直接验证",
			"webhook_name": webhook.Name,
			"webhook_url":  webhook.URL,
			"channel":      channel,
		})
		return
	}

	webhook.ApplyDefaults()
	sender, err := h.senderFactory.SenderFor(&webhook)
	if err != nil {
		logger.GetLogger().Errorf("Failed to resolve sender for webhook [ID: %d]: %v", webhook.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "解析Webhook类型失败",
			"details": err.Error(),
		})
		return
	}

	payload := &services.MergeRequestPayload{
		ProjectName:  fmt.Sprintf("Webhook测试 - %s", webhook.Name),
		SourceBranch: "test/source",
		TargetBranch: "test/target",
		AuthorName:   "GitLab Merge Alert",
		Title:        fmt.Sprintf("Webhook [%s] 连接测试", webhook.Name),
		URL:          h.config.PublicWebhookURL,
	}

	if err := sender.Send(c.Request.Context(), &webhook, payload); err != nil {
		logger.GetLogger().Errorf("Failed to send test message to webhook [ID: %d, Name: %s]: %v", webhook.ID, webhook.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "发送测试消息失败",
			"details": err.Error(),
		})
		return
	}

	accountID, _ := middleware.GetAccountID(c)
	notification := &models.Notification{
		Title:            "Webhook测试",
		Status:           "success",
		NotificationSent: true,
		OwnerID:          &accountID,
	}

	if err := h.db.Create(notification).Error; err != nil {
		logger.GetLogger().Warnf("Failed to save test notification history: %v", err)
	}

	logger.GetLogger().Infof("Successfully sent test message to webhook [ID: %d, Name: %s]", webhook.ID, webhook.Name)

	c.JSON(http.StatusOK, gin.H{
		"message":      "测试消息发送成功",
		"webhook_name": webhook.Name,
		"sent_at":      time.Now().Format("2006-01-02 15:04:05"),
		"channel":      channel,
	})
}

func buildWebhookResponse(webhook *models.Webhook) models.WebhookResponse {
	if webhook == nil {
		return models.WebhookResponse{}
	}

	webhook.ApplyDefaults()

	signatureMethod := models.SignatureMethodHMACSHA256
	secret := ""
	if webhook.Settings != nil {
		if webhook.Settings.SignatureMethod != "" {
			signatureMethod = webhook.Settings.SignatureMethod
		}
		secret = webhook.Settings.Secret
	}

	response := models.WebhookResponse{
		ID:               webhook.ID,
		Name:             webhook.Name,
		URL:              webhook.URL,
		Description:      webhook.Description,
		Type:             webhook.Type,
		SignatureMethod:  signatureMethod,
		Secret:           secret,
		SecurityKeywords: webhook.SecurityKeywordsAsSlice(),
		CustomHeaders:    webhook.CustomHeadersAsMap(),
		IsActive:         webhook.IsActive,
		CreatedAt:        webhook.CreatedAt,
		UpdatedAt:        webhook.UpdatedAt,
	}

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

	return response
}

func (h *Handler) upsertWebhookSettings(webhookID uint, signature *string, secret *string, keywords *[]string, headers *map[string]string) error {
	var setting models.WebhookSetting
	err := h.db.Where("webhook_id = ?", webhookID).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting = models.WebhookSetting{WebhookID: webhookID}
	} else if err != nil {
		return err
	}

	if signature != nil && *signature != "" {
		setting.SignatureMethod = *signature
	}
	if secret != nil {
		setting.Secret = *secret
	}
	if keywords != nil {
		setting.SecurityKeywords = models.ToStringList(*keywords)
	}
	if headers != nil {
		setting.CustomHeaders = models.ToStringMap(*headers)
	}

	setting.ApplyDefaults()

	if setting.ID == 0 {
		return h.db.Create(&setting).Error
	}

	return h.db.Save(&setting).Error
}
