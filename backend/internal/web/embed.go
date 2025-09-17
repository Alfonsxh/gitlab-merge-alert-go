package web

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// 嵌入前端构建后的静态文件
// 注意：embed 指令的路径是相对于当前文件的位置
//
//go:embed frontend_dist/*
var frontendFS embed.FS

// GetFrontendFS 返回前端文件系统，去掉路径前缀
func GetFrontendFS() (http.FileSystem, error) {
	// 去掉 frontend_dist 前缀，直接访问文件
	subFS, err := fs.Sub(frontendFS, "frontend_dist")
	if err != nil {
		return nil, err
	}
	return http.FS(subFS), nil
}

// SetupStaticFiles 配置静态文件服务
func SetupStaticFiles(router *gin.Engine) error {
	// 尝试多个可能的前端文件路径（开发模式）
	devPaths := []string{
		"frontend/dist",           // 从项目根目录运行
		"../frontend/dist",        // 从 backend 目录运行
		"../../frontend/dist",     // 从 backend/cmd/server 运行
	}

	for _, path := range devPaths {
		if _, err := os.Stat(path); err == nil {
			logger.GetLogger().Infof("Using filesystem for static files (development mode): %s", path)
			router.Static("/assets", path+"/assets")
			router.StaticFile("/vite.svg", path+"/vite.svg")
			return nil
		}
	}

	// 生产模式：使用嵌入的文件
	logger.GetLogger().Info("Using embedded static files (production mode)")
	fs, err := GetFrontendFS()
	if err != nil {
		return err
	}

	// 提供 /assets 目录下的所有文件
	router.StaticFS("/assets", fs)

	// 单独处理 vite.svg
	router.GET("/vite.svg", func(c *gin.Context) {
		c.FileFromFS("vite.svg", fs)
	})

	return nil
}

// ServeIndexHTML 提供 index.html
func ServeIndexHTML(c *gin.Context) {
	// 尝试多个可能的前端文件路径（开发模式）
	devPaths := []string{
		"frontend/dist/index.html",           // 从项目根目录运行
		"../frontend/dist/index.html",        // 从 backend 目录运行
		"../../frontend/dist/index.html",     // 从 backend/cmd/server 运行
	}

	for _, path := range devPaths {
		if _, err := os.Stat(path); err == nil {
			c.File(path)
			return
		}
	}

	// 生产模式：使用嵌入的文件
	fs, err := GetFrontendFS()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
		return
	}

	c.FileFromFS("index.html", fs)
}