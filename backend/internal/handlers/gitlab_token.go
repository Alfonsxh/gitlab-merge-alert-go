package handlers

import (
	"errors"
	"net/http"
	"strings"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

type testGitLabTokenRequest struct {
	AccessToken string `json:"access_token"`
	GitLabURL   string `json:"gitlab_url"`
}

func (h *Handler) TestGitLabToken(c *gin.Context) {
	var req testGitLabTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	baseURL := strings.TrimSpace(req.GitLabURL)
	if baseURL == "" {
		baseURL = h.config.GitLabURL
	}

	if baseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未配置 GitLab 服务器地址"})
		return
	}

	token, err := h.resolveGitLabToken(c, req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, errUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		case errors.Is(err, errGitLabTokenMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前账户未配置 GitLab Personal Access Token"})
		default:
			logger.GetLogger().Errorf("Failed to resolve GitLab token for testing: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析GitLab Token失败"})
		}
		return
	}

	if err := h.gitlabService.TestConnection(baseURL, token); err != nil {
		c.JSON(http.StatusOK, gin.H{"data": models.GitLabConnectionTestResponse{Success: false, Message: err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.GitLabConnectionTestResponse{Success: true, Message: "连接成功"}})
}
