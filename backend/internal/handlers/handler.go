package handlers

import (
	"net/http"
	"time"

	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db               *gorm.DB
	config           *config.Config
	gitlabService    services.GitLabService
	wechatService    services.WeChatService
	notifyService    services.NotificationService
	authService      services.AuthService
	authMiddleware   *middleware.AuthMiddleware
	ownershipChecker *middleware.OwnershipChecker
	response         *middleware.ResponseHelper
}

func New(db *gorm.DB, cfg *config.Config) *Handler {
	gitlabService := services.NewGitLabService(cfg.GitLabURL, "")
	wechatService := services.NewWeChatService()
	notifyService := services.NewNotificationService(db, wechatService)

	// 使用配置中的 JWT 设置，如果没有则使用默认值
	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "gitlab-merge-alert-secret-key" // 默认密钥，生产环境应使用配置
	}
	jwtDuration := cfg.JWTDuration
	if jwtDuration == 0 {
		jwtDuration = 24 * time.Hour // 默认 24 小时
	}

	authService := services.NewAuthService(db, jwtSecret, jwtDuration)
	authMiddleware := middleware.NewAuthMiddleware(db, jwtSecret)
	ownershipChecker := middleware.NewOwnershipChecker(db)

	return &Handler{
		db:               db,
		config:           cfg,
		gitlabService:    gitlabService,
		wechatService:    wechatService,
		notifyService:    notifyService,
		authService:      authService,
		authMiddleware:   authMiddleware,
		ownershipChecker: ownershipChecker,
		response:         middleware.NewResponseHelper(),
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

// InitializeAdminAccount 初始化管理员账户
func (h *Handler) InitializeAdminAccount() error {
	return h.authService.InitializeAdminAccount()
}

// GetAuthMiddleware 获取认证中间件
func (h *Handler) GetAuthMiddleware() *middleware.AuthMiddleware {
	return h.authMiddleware
}

// GetOwnershipChecker 获取所有权检查器
func (h *Handler) GetOwnershipChecker() *middleware.OwnershipChecker {
	return h.ownershipChecker
}
