package handlers

import (
	"encoding/json"
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

	// 记录完整的 webhook 数据到日志
	if webhookJSON, err := json.MarshalIndent(webhookData, "", "  "); err == nil {
		logger.GetLogger().Infof("收到 GitLab Webhook 请求:\n%s", string(webhookJSON))
	}

	// 特别记录 assignees 信息
	logger.GetLogger().Infof("Webhook 详情 - 项目: %s (ID: %d), 合并请求: %s, 状态: %s",
		webhookData.Project.Name,
		webhookData.Project.ID,
		webhookData.ObjectAttributes.Title,
		webhookData.ObjectAttributes.State)

	if len(webhookData.Assignees) > 0 {
		logger.GetLogger().Infof("发现 %d 个指派人:", len(webhookData.Assignees))
		for i, assignee := range webhookData.Assignees {
			logger.GetLogger().Infof("  指派人 %d: %s (%s) - 邮箱: %s", i+1, assignee.Name, assignee.Username, assignee.Email)
		}
	} else {
		logger.GetLogger().Warnf("此合并请求没有指派人")
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
