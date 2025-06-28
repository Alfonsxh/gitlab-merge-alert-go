package handlers

import (
	"net/http"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

func (h *Handler) HandleGitLabWebhook(c *gin.Context) {
	var webhookData models.GitLabWebhookData
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		logger.GetLogger().Errorf("Failed to parse webhook data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
		return
	}

	// 只处理合并请求事件
	if webhookData.ObjectKind != "merge_request" {
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	if err := h.notifyService.ProcessMergeRequest(&webhookData); err != nil {
		logger.GetLogger().Errorf("Failed to process merge request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}