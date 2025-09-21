package handlers

import (
	"net/http"

	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/internal/services"

	"github.com/gin-gonic/gin"
)

// Register 注册新的普通用户账户
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	resp, err := h.authService.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		switch err {
		case services.ErrUsernameExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		case services.ErrEmailExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		case services.ErrAdminLocked:
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin account is reserved"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register account"})
			return
		}
	}

	h.response.Success(c, resp)
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 调用认证服务
	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		if err == services.ErrAccountNotActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is not active"})
			return
		}
		if err == services.ErrPasswordResetRequired {
			c.JSON(http.StatusForbidden, gin.H{"error": "Password reset required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	h.response.Success(c, resp)
}

// Logout 用户登出
func (h *Handler) Logout(c *gin.Context) {
	// JWT 是无状态的，客户端清除 token 即可
	h.response.Success(c, gin.H{"message": "Logout successful"})
}

// RefreshToken 刷新 Token
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	resp, err := h.authService.RefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token"})
		return
	}

	h.response.Success(c, resp)
}

// GetProfile 获取当前用户信息
func (h *Handler) GetProfile(c *gin.Context) {
	accountID, exists := middleware.GetAccountID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	account, err := h.authService.GetAccountByID(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	h.response.Success(c, account.ToResponse())
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	accountID, exists := middleware.GetAccountID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	err := h.authService.ChangePassword(accountID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	h.response.Success(c, gin.H{"message": "Password changed successfully"})
}
