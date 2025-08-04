package handlers

import (
	"net/http"
	
	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db            *gorm.DB
	config        *config.Config
	gitlabService services.GitLabService
	wechatService services.WeChatService
	notifyService services.NotificationService
	response      *middleware.ResponseHelper
}

func New(db *gorm.DB, cfg *config.Config) *Handler {
	gitlabService := services.NewGitLabService(cfg.GitLabURL, cfg.GitLabPersonalAccessToken)
	wechatService := services.NewWeChatService()
	notifyService := services.NewNotificationService(db, wechatService)

	return &Handler{
		db:            db,
		config:        cfg,
		gitlabService: gitlabService,
		wechatService: wechatService,
		notifyService: notifyService,
		response:      middleware.NewResponseHelper(),
	}
}

// GetGitLabConfig 获取GitLab配置信息
func (h *Handler) GetGitLabConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"gitlab_url": h.config.GitLabURL,
		},
	})
}

