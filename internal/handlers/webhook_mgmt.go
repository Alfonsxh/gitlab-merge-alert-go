package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) GetWebhooks(c *gin.Context) {
	var webhooks []models.Webhook
	if err := h.db.Preload("Projects").Find(&webhooks).Error; err != nil {
		logger.GetLogger().Errorf("Failed to fetch webhooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}

	logger.GetLogger().Debugf("Successfully fetched %d webhooks", len(webhooks))

	// 转换为响应格式
	var responses []models.WebhookResponse
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

	webhook := &models.Webhook{
		Name:        req.Name,
		URL:         req.URL,
		Description: req.Description,
		IsActive:    true,
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

	// 建立关联
	if err := h.db.Model(&project).Association("Webhooks").Append(&webhook); err != nil {
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
