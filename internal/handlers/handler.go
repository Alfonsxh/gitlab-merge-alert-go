package handlers

import (
	"net/http"

	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/services"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db             *gorm.DB
	config         *config.Config
	gitlabService  *services.GitLabService
	wechatService  *services.WeChatService
	notifyService  *services.NotificationService
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

func (h *Handler) renderTemplate(c *gin.Context, templateName string, data gin.H) error {
	defer func() {
		if r := recover(); r != nil {
			logger.GetLogger().Errorf("Template rendering panic: %v", r)
		}
	}()

	c.HTML(http.StatusOK, templateName, data)
	return nil
}