package main

import (
	"fmt"
	"log"
	"strings"

	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/database"
	"gitlab-merge-alert-go/internal/handlers"
	"gitlab-merge-alert-go/internal/web"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.LogLevel)

	// 初始化数据库
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 运行数据库迁移
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 为 Vue SPA 服务静态文件
	if err := web.SetupStaticFiles(router); err != nil {
		log.Fatalf("Failed to setup static files: %v", err)
	}

	// 初始化处理器
	h := handlers.New(db, cfg)

	// 初始化默认管理员账户
	if err := h.InitializeAdminAccount(); err != nil {
		log.Fatalf("Failed to initialize admin account: %v", err)
	}

	// 注册路由
	setupRoutes(router, h)

	// 添加根路径处理，返回 index.html
	router.GET("/", func(c *gin.Context) {
		web.ServeIndexHTML(c)
	})

	// 启动服务器
	logger.GetLogger().Infof("Starting server on %s:%d", cfg.Host, cfg.Port)
	if err := router.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(router *gin.Engine, h *handlers.Handler) {
	// API路由
	api := router.Group("/api/v1")
	{
		system := api.Group("/system")
		{
			system.GET("/bootstrap", h.GetBootstrapStatus)
			system.POST("/setup-admin", h.SetupAdmin)
		}

		// 公开路由（无需认证）
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
			auth.POST("/logout", h.Logout)
			auth.POST("/refresh", h.RefreshToken)
		}

		// GitLab Webhook接收（无需认证，使用 webhook 自身的验证）
		api.POST("/webhook/gitlab", h.HandleGitLabWebhook)

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(h.GetAuthMiddleware().RequireAuth())
		{
			// 认证相关
			authProtected := protected.Group("/auth")
			{
				authProtected.GET("/profile", h.GetProfile)
				authProtected.PUT("/profile", h.UpdateProfile)
				authProtected.POST("/avatar", h.UploadAvatar)
				authProtected.POST("/change-password", h.ChangePassword)
			}

			// 账户管理（仅管理员）
			admin := protected.Group("")
			admin.Use(h.GetAuthMiddleware().RequireAdmin())
			{
				accounts := admin.Group("/accounts")
				{
					accounts.GET("", h.GetAccounts)
					accounts.POST("", h.CreateAccount)
					accounts.PUT("/:id", h.UpdateAccount)
					accounts.DELETE("/:id", h.DeleteAccount)
					accounts.PUT("/:id/password", h.ResetPassword)
				}
			}

			// 用户管理API（GitLab 用户映射）
			users := protected.Group("/users")
			{
				users.GET("", h.GetUsers)
				users.POST("", h.CreateUser)
				users.PUT("/:id", h.UpdateUser).Use(h.GetOwnershipChecker().CheckUserOwnership())
				users.DELETE("/:id", h.DeleteUser).Use(h.GetOwnershipChecker().CheckUserOwnership())
			}

			// 项目管理API
			projects := protected.Group("/projects")
			{
				projects.GET("", h.GetProjects)
				projects.POST("", h.CreateProject)
				projects.PUT("/:id", h.UpdateProject).Use(h.GetOwnershipChecker().CheckProjectOwnership())
				projects.DELETE("/:id", h.DeleteProject).Use(h.GetOwnershipChecker().CheckProjectOwnership())
				projects.POST("/parse-url", h.ParseProjectURL)
				projects.POST("/scan-group", h.ScanGroupProjects)
				projects.POST("/batch-create", h.BatchCreateProjects)

				// GitLab Webhook管理API
				projects.POST("/:id/sync-gitlab-webhook", h.SyncGitLabWebhook).Use(h.GetOwnershipChecker().CheckProjectOwnership())
				projects.DELETE("/:id/sync-gitlab-webhook", h.DeleteGitLabWebhook).Use(h.GetOwnershipChecker().CheckProjectOwnership())
				projects.GET("/:id/gitlab-webhook-status", h.GetGitLabWebhookStatus).Use(h.GetOwnershipChecker().CheckProjectOwnership())
				projects.POST("/batch-check-webhook-status", h.BatchCheckWebhookStatus)
			}

			// GitLab相关API
			gitlab := protected.Group("/gitlab")
			{
				gitlab.POST("/test-connection", h.TestGitLabConnection)
				gitlab.GET("/config", h.GetGitLabConfig)
				gitlab.POST("/test-token", h.TestGitLabToken)
			}

			// Webhook管理API
			webhooks := protected.Group("/webhooks")
			{
				webhooks.GET("", h.GetWebhooks)
				webhooks.POST("", h.CreateWebhook)
				webhooks.PUT("/:id", h.UpdateWebhook).Use(h.GetOwnershipChecker().CheckWebhookOwnership())
				webhooks.DELETE("/:id", h.DeleteWebhook).Use(h.GetOwnershipChecker().CheckWebhookOwnership())
				webhooks.POST("/:id/test", h.SendTestMessage).Use(h.GetOwnershipChecker().CheckWebhookOwnership())
			}

			// 项目-Webhook关联API
			protected.POST("/project-webhooks", h.LinkProjectWebhook)
			protected.DELETE("/project-webhooks/:project_id/:webhook_id", h.UnlinkProjectWebhook)

			// 资源管理API（仅管理员）
			resourceManager := protected.Group("/resource-managers")
			resourceManager.Use(h.GetAuthMiddleware().RequireAdmin())
			{
				resourceManager.POST("/assign", h.AssignManager)
				resourceManager.POST("/remove", h.RemoveManager)
				resourceManager.GET("", h.GetResourceManagers)
				resourceManager.GET("/managed/:id", h.GetManagedResources)
				resourceManager.POST("/batch-assign/:id", h.BatchAssignResources)
			}

			// 统计API
			protected.GET("/stats", h.GetStats)
			protected.GET("/notifications", h.GetNotifications)
			protected.GET("/stats/projects/daily", h.GetProjectDailyStats)
			protected.GET("/stats/webhooks/daily", h.GetWebhookDailyStats)
		}
	}

	// 配置 SPA 路由 - 必须在 API 路由之后定义
	router.NoRoute(func(c *gin.Context) {
		// API 路由不存在时返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		// 其他路由返回 index.html
		web.ServeIndexHTML(c)
	})
}
