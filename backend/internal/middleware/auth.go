package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/internal/services"
	"gitlab-merge-alert-go/pkg/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	ContextKeyAccount   = "account"
	ContextKeyAccountID = "account_id"
	ContextKeyRole      = "role"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
	db         *gorm.DB
}

func NewAuthMiddleware(db *gorm.DB, jwtSecret string) *AuthMiddleware {
	jwtManager := auth.NewJWTManager(jwtSecret, 0)
	return &AuthMiddleware{
		jwtManager: jwtManager,
		db:         db,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			return
		}

		claims, err := m.jwtManager.Verify(token)
		if err != nil {
			m.handleTokenError(c, err)
			return
		}

		account, err := m.loadAccount(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if !account.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Account is not active"})
			return
		}

		if account.ForcePasswordReset {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Password reset required"})
			return
		}

		m.setContext(c, account)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.jwtManager.Verify(token)
		if err != nil {
			c.Next()
			return
		}

		account, err := m.loadAccount(claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		if !account.IsActive || account.ForcePasswordReset {
			c.Next()
			return
		}

		m.setContext(c, account)
		c.Next()
	}
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return ""
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(authHeader, BearerPrefix)
}

func (m *AuthMiddleware) handleTokenError(c *gin.Context, err error) {
	if err == auth.ErrExpiredToken {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
		return
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
}

func (m *AuthMiddleware) loadAccount(id uint) (*models.Account, error) {
	var account models.Account
	if err := m.db.First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &account, nil
}

func (m *AuthMiddleware) setContext(c *gin.Context, account *models.Account) {
	c.Set(ContextKeyAccount, account)
	c.Set(ContextKeyAccountID, account.ID)
	c.Set(ContextKeyRole, account.Role)
}

func GetAccountID(c *gin.Context) (uint, bool) {
	accountID, exists := c.Get(ContextKeyAccountID)
	if !exists {
		return 0, false
	}

	id, ok := accountID.(uint)
	return id, ok
}

func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	return roleStr, ok
}

func IsAdmin(c *gin.Context) bool {
	role, exists := GetRole(c)
	return exists && role == models.RoleAdmin
}

func RequireResourcePermission(db *gorm.DB, resourceType models.ResourceType) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, exists := GetAccountID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		role, _ := GetRole(c)

		if role == models.RoleAdmin {
			c.Next()
			return
		}

		var resourceID uint
		idParam := c.Param("id")
		if idParam != "" {
			var id int
			if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
				c.Abort()
				return
			}
			resourceID = uint(id)
		}

		if c.Request.Method == "GET" && resourceID == 0 {
			c.Set("filter_by_permission", true)
			c.Next()
			return
		}

		if resourceID > 0 {
			rmService := services.NewResourceManagerService(db)
			if !rmService.HasPermission(accountID, role, resourceID, resourceType) {
				c.JSON(http.StatusForbidden, gin.H{"error": "没有权限访问此资源"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func RequireSelfOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, exists := GetAccountID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		role, _ := GetRole(c)

		if role == models.RoleAdmin {
			c.Next()
			return
		}

		targetIDStr := c.Param("id")
		if targetIDStr != "" {
			var targetID uint
			if _, err := fmt.Sscanf(targetIDStr, "%d", &targetID); err == nil {
				if targetID != accountID {
					c.JSON(http.StatusForbidden, gin.H{"error": "只能访问自己的资源"})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
