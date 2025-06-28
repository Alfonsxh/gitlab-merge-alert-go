package handlers

import (
	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/services"

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
	gitlabService := services.NewGitLabService(cfg.GitLabURLPrefix, cfg.GitLabPersonalAccessToken)
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