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
			logger.GetLogger().Errorf("Template rendering panic for %s: %v", templateName, r)
			// 尝试渲染错误页面
			h.renderErrorPage(c, "模板渲染失败")
		}
	}()

	logger.GetLogger().Debugf("Rendering template: %s", templateName)
	c.HTML(http.StatusOK, templateName, data)
	return nil
}

func (h *Handler) renderErrorPage(c *gin.Context, errorMsg string) {
	defer func() {
		if r := recover(); r != nil {
			logger.GetLogger().Errorf("Error page rendering panic: %v", r)
			// 最后的备用方案：返回简单的JSON错误
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "页面渲染失败",
				"details": errorMsg,
			})
		}
	}()

	c.HTML(http.StatusInternalServerError, "error.html", gin.H{
		"error": errorMsg,
	})
}