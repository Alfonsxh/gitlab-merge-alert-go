package web

import (
	"embed"
	"io/fs"
	"net/http"

	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// 嵌入前端构建后的静态文件
// 注意：embed 指令的路径是相对于当前文件的位置
//
//go:embed frontend_dist
var frontendFS embed.FS

// GetFrontendFS 返回前端文件系统，去掉路径前缀
func GetFrontendFS() (http.FileSystem, error) {
	subFS, err := fs.Sub(frontendFS, "frontend_dist")
	if err != nil {
		return nil, err
	}
	return http.FS(subFS), nil
}

// SetupStaticFiles 配置静态文件服务
func SetupStaticFiles(router *gin.Engine) error {
	logger.GetLogger().Info("Serving static files from embedded frontend bundle")

	assetSub, err := fs.Sub(frontendFS, "frontend_dist/assets")
	if err != nil {
		return err
	}
	router.StaticFS("/assets", http.FS(assetSub))

	router.GET("/vite.svg", func(c *gin.Context) {
		data, err := frontendFS.ReadFile("frontend_dist/vite.svg")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	return nil
}

// ServeIndexHTML 提供 index.html
func ServeIndexHTML(c *gin.Context) {
	data, err := frontendFS.ReadFile("frontend_dist/index.html")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}
