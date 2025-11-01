//go:build !embed

package web

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Alfonsxh/gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

const frontendDir = "internal/web/frontend_dist"

func GetFrontendFS() (http.FileSystem, error) {
	if _, err := os.Stat(frontendDir); err != nil {
		return nil, err
	}
	return http.Dir(frontendDir), nil
}

func SetupStaticFiles(router *gin.Engine) error {
	assetPath := filepath.Join(frontendDir, "assets")
	if info, err := os.Stat(assetPath); err == nil {
		if info.IsDir() {
			router.StaticFS("/assets", http.Dir(assetPath))
		} else {
			logger.GetLogger().Warn("frontend assets path exists but is not a directory; skipping /assets mount")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		logger.GetLogger().Warn("Local frontend assets directory missing; skipping /assets mount")
	} else {
		return err
	}

	router.GET("/vite.svg", func(c *gin.Context) {
		data, err := os.ReadFile(filepath.Join(frontendDir, "vite.svg"))
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	return nil
}

func ServeIndexHTML(c *gin.Context) {
	data, err := os.ReadFile(filepath.Join(frontendDir, "index.html"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}
