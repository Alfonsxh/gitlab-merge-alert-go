package main

import (
	"fmt"
	"log"
	"strings"

	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/database"
	"gitlab-merge-alert-go/internal/handlers"
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
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/vite.svg", "./frontend/dist/vite.svg")

	// 初始化处理器
	h := handlers.New(db, cfg)

	// 注册路由
	setupRoutes(router, h)

	// 启动服务器
	logger.GetLogger().Infof("Starting server on %s:%d", cfg.Host, cfg.Port)
	if err := router.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(router *gin.Engine, h *handlers.Handler) {
	// 配置 SPA 路由
	router.NoRoute(func(c *gin.Context) {
		// API 路由不存在时返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		// 其他路由返回 index.html
		c.File("./frontend/dist/index.html")
	})

	// API路由
	api := router.Group("/api/v1")
	{
		// GitLab Webhook接收
		api.POST("/webhook/gitlab", h.HandleGitLabWebhook)

		// 用户管理API
		users := api.Group("/users")
		{
			users.GET("", h.GetUsers)
			users.POST("", h.CreateUser)
			users.PUT("/:id", h.UpdateUser)
			users.DELETE("/:id", h.DeleteUser)
		}

		// 项目管理API
		projects := api.Group("/projects")
		{
			projects.GET("", h.GetProjects)
			projects.POST("", h.CreateProject)
			projects.PUT("/:id", h.UpdateProject)
			projects.DELETE("/:id", h.DeleteProject)
			projects.POST("/parse-url", h.ParseProjectURL)
			projects.POST("/scan-group", h.ScanGroupProjects)
			projects.POST("/batch-create", h.BatchCreateProjects)

			// GitLab Webhook管理API
			projects.POST("/:id/sync-gitlab-webhook", h.SyncGitLabWebhook)
			projects.DELETE("/:id/sync-gitlab-webhook", h.DeleteGitLabWebhook)
			projects.GET("/:id/gitlab-webhook-status", h.GetGitLabWebhookStatus)
		}

		// GitLab相关API
		gitlab := api.Group("/gitlab")
		{
			gitlab.POST("/test-connection", h.TestGitLabConnection)
			gitlab.GET("/config", h.GetGitLabConfig)
		}

		// Webhook管理API
		webhooks := api.Group("/webhooks")
		{
			webhooks.GET("", h.GetWebhooks)
			webhooks.POST("", h.CreateWebhook)
			webhooks.PUT("/:id", h.UpdateWebhook)
			webhooks.DELETE("/:id", h.DeleteWebhook)
		}

		// 项目-Webhook关联API
		api.POST("/project-webhooks", h.LinkProjectWebhook)
		api.DELETE("/project-webhooks/:project_id/:webhook_id", h.UnlinkProjectWebhook)

		// 统计API
		api.GET("/stats", h.GetStats)
		api.GET("/notifications", h.GetNotifications)
	}
}
