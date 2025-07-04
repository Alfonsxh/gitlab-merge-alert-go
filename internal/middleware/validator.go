package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ValidateID 验证路径中的ID参数
func ValidateID(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param(paramName)
		if idStr == "" {
			response := models.ValidationErrorResponse(fmt.Sprintf("%s参数不能为空", paramName))
			c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
			c.Abort()
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response := models.ValidationErrorResponse(fmt.Sprintf("无效的%s格式", paramName))
			c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
			c.Abort()
			return
		}

		// 将解析后的ID存储到context中，避免重复解析
		c.Set(paramName+"_parsed", uint(id))
		c.Next()
	}
}

// ValidateJSON 验证JSON请求体
func ValidateJSON(target interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(target); err != nil {
			logger.GetLogger().Warnf("JSON validation failed: %v", err)
			response := models.ValidationErrorResponse(err.Error())
			c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetParsedID 从context中获取已解析的ID
func GetParsedID(c *gin.Context, paramName string) uint {
	if value, exists := c.Get(paramName + "_parsed"); exists {
		if id, ok := value.(uint); ok {
			return id
		}
	}
	return 0
}

// CORS 跨域处理中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token,X-Token,X-User-Id")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
