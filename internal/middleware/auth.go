package middleware

import (
	"net/http"
	"strings"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/auth"

	"github.com/gin-gonic/gin"
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
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	// 这里使用较长的过期时间，实际的过期时间由 JWT token 自身控制
	jwtManager := auth.NewJWTManager(jwtSecret, 0)
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth 验证用户是否已登录
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization token",
			})
			c.Abort()
			return
		}

		// 验证 token
		claims, err := m.jwtManager.Verify(token)
		if err != nil {
			if err == auth.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token has expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
				})
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(ContextKeyAccountID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireAdmin 验证用户是否为管理员
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		if role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选的认证，如果有 token 则验证，没有也放行
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		// 验证 token
		claims, err := m.jwtManager.Verify(token)
		if err == nil {
			// token 有效，存入上下文
			c.Set(ContextKeyAccountID, claims.UserID)
			c.Set(ContextKeyRole, claims.Role)
		}
		// 无论 token 是否有效都继续
		c.Next()
	}
}

// extractToken 从请求头中提取 token
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

// GetAccountID 从上下文中获取当前用户 ID
func GetAccountID(c *gin.Context) (uint, bool) {
	accountID, exists := c.Get(ContextKeyAccountID)
	if !exists {
		return 0, false
	}
	
	id, ok := accountID.(uint)
	return id, ok
}

// GetRole 从上下文中获取当前用户角色
func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	
	roleStr, ok := role.(string)
	return roleStr, ok
}

// IsAdmin 检查当前用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	role, exists := GetRole(c)
	return exists && role == models.RoleAdmin
}