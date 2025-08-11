package middleware

import (
	"net/http"
	"strconv"

	"gitlab-merge-alert-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OwnershipChecker struct {
	db *gorm.DB
}

func NewOwnershipChecker(db *gorm.DB) *OwnershipChecker {
	return &OwnershipChecker{db: db}
}

// CheckProjectOwnership 检查项目所有权
func (o *OwnershipChecker) CheckProjectOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是管理员，直接放行
		if IsAdmin(c) {
			c.Next()
			return
		}

		// 获取当前用户 ID
		accountID, exists := GetAccountID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// 获取项目 ID
		projectIDStr := c.Param("id")
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			c.Abort()
			return
		}

		// 检查项目是否属于当前用户
		var count int64
		if err := o.db.Model(&models.Project{}).
			Where("id = ? AND created_by = ?", projectID, accountID).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckWebhookOwnership 检查 Webhook 所有权
func (o *OwnershipChecker) CheckWebhookOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是管理员，直接放行
		if IsAdmin(c) {
			c.Next()
			return
		}

		// 获取当前用户 ID
		accountID, exists := GetAccountID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// 获取 Webhook ID
		webhookIDStr := c.Param("id")
		webhookID, err := strconv.ParseUint(webhookIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
			c.Abort()
			return
		}

		// 检查 Webhook 是否属于当前用户
		var count int64
		if err := o.db.Model(&models.Webhook{}).
			Where("id = ? AND created_by = ?", webhookID, accountID).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckUserOwnership 检查用户（GitLab 用户映射）所有权
func (o *OwnershipChecker) CheckUserOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是管理员，直接放行
		if IsAdmin(c) {
			c.Next()
			return
		}

		// 获取当前用户 ID
		accountID, exists := GetAccountID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// 获取用户 ID
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// 检查用户是否属于当前账户
		var count int64
		if err := o.db.Model(&models.User{}).
			Where("id = ? AND created_by = ?", userID, accountID).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ApplyOwnershipFilter 应用所有权过滤器到查询
func ApplyOwnershipFilter(c *gin.Context, query *gorm.DB, tableName string) *gorm.DB {
	// 如果是管理员且请求查看所有数据
	if IsAdmin(c) {
		showAll := c.Query("all") == "true"
		if showAll {
			return query
		}
	}

	// 获取当前用户 ID
	accountID, exists := GetAccountID(c)
	if !exists {
		// 未登录用户看不到任何数据
		return query.Where("1 = 0")
	}

	// 根据不同的表使用不同的字段
	switch tableName {
	case "notifications":
		// notifications 表：通过项目权限控制，查看用户有权限访问的项目的通知
		return query.Where(tableName+".project_id IN (SELECT id FROM projects WHERE created_by = ? OR id IN (SELECT resource_id FROM resource_managers WHERE manager_id = ? AND resource_type = 'project'))", accountID, accountID)
	case "projects":
		// projects 表：查询用户创建的或被分配管理的项目
		return query.Where(tableName+".created_by = ? OR "+tableName+".id IN (SELECT resource_id FROM resource_managers WHERE manager_id = ? AND resource_type = 'project')", accountID, accountID)
	case "webhooks":
		// webhooks 表：查询用户创建的或被分配管理的 webhook
		return query.Where(tableName+".created_by = ? OR "+tableName+".id IN (SELECT resource_id FROM resource_managers WHERE manager_id = ? AND resource_type = 'webhook')", accountID, accountID)
	case "users":
		// users 表：查询用户创建的或被分配管理的用户
		return query.Where(tableName+".created_by = ? OR "+tableName+".id IN (SELECT resource_id FROM resource_managers WHERE manager_id = ? AND resource_type = 'user')", accountID, accountID)
	default:
		// 其他表使用 created_by
		return query.Where(tableName+".created_by = ?", accountID)
	}
}