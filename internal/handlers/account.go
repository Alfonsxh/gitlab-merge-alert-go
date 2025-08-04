package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAccounts 获取账户列表（仅管理员）
func (h *Handler) GetAccounts(c *gin.Context) {
	var accounts []models.Account
	
	// 构建查询
	query := h.db.Model(&models.Account{})
	
	// 分页参数
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}
	
	// 搜索参数
	if search := c.Query("search"); search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// 角色过滤
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	
	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count accounts"})
		return
	}
	
	// 获取数据
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get accounts"})
		return
	}
	
	// 转换为响应格式
	var responses []models.AccountResponse
	for _, account := range accounts {
		responses = append(responses, *account.ToResponse())
	}
	
	h.response.Success(c, gin.H{
		"total": total,
		"data":  responses,
		"page":  page,
		"page_size": pageSize,
	})
}

// CreateAccount 创建账户（仅管理员）
func (h *Handler) CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	
	// 如果没有指定角色，默认为普通用户
	if req.Role == "" {
		req.Role = models.RoleUser
	}
	
	// 检查用户名是否已存在
	var count int64
	if err := h.db.Model(&models.Account{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}
	
	// 检查邮箱是否已存在
	if err := h.db.Model(&models.Account{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}
	
	// 加密密码
	passwordManager := auth.NewPasswordManager()
	passwordHash, err := passwordManager.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// 创建账户
	account := &models.Account{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
		Role:         req.Role,
		IsActive:     true,
	}
	
	if err := h.db.Create(account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}
	
	h.response.Success(c, account.ToResponse())
}

// UpdateAccount 更新账户（仅管理员）
func (h *Handler) UpdateAccount(c *gin.Context) {
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}
	
	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	
	// 获取账户
	var account models.Account
	if err := h.db.First(&account, accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	// 更新字段
	updates := make(map[string]interface{})
	if req.Email != "" && req.Email != account.Email {
		// 检查邮箱是否已被其他账户使用
		var count int64
		if err := h.db.Model(&models.Account{}).Where("email = ? AND id != ?", req.Email, accountID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		updates["email"] = req.Email
	}
	
	if req.Role != "" && req.Role != account.Role {
		// 不能移除最后一个管理员
		if account.Role == models.RoleAdmin && req.Role != models.RoleAdmin {
			var adminCount int64
			if err := h.db.Model(&models.Account{}).Where("role = ? AND id != ?", models.RoleAdmin, accountID).Count(&adminCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
			if adminCount == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the last admin"})
				return
			}
		}
		updates["role"] = req.Role
	}
	
	if req.IsActive != nil {
		// 不能禁用最后一个管理员
		if account.Role == models.RoleAdmin && !*req.IsActive {
			var activeAdminCount int64
			if err := h.db.Model(&models.Account{}).Where("role = ? AND is_active = ? AND id != ?", models.RoleAdmin, true, accountID).Count(&activeAdminCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
			if activeAdminCount == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot disable the last active admin"})
				return
			}
		}
		updates["is_active"] = *req.IsActive
	}
	
	// 执行更新
	if len(updates) > 0 {
		if err := h.db.Model(&account).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
			return
		}
	}
	
	// 重新加载账户数据
	if err := h.db.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload account"})
		return
	}
	
	h.response.Success(c, account.ToResponse())
}

// DeleteAccount 删除账户（仅管理员）
func (h *Handler) DeleteAccount(c *gin.Context) {
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}
	
	// 获取当前登录用户ID
	currentAccountID, _ := middleware.GetAccountID(c)
	
	// 不能删除自己的账户
	if uint(accountID) == currentAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}
	
	// 获取要删除的账户
	var account models.Account
	if err := h.db.First(&account, accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	// 不能删除最后一个管理员
	if account.Role == models.RoleAdmin {
		var adminCount int64
		if err := h.db.Model(&models.Account{}).Where("role = ? AND id != ?", models.RoleAdmin, accountID).Count(&adminCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if adminCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the last admin"})
			return
		}
	}
	
	// 删除账户
	if err := h.db.Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}
	
	h.response.Success(c, gin.H{"message": "Account deleted successfully"})
}

// ResetPassword 重置密码（仅管理员）
func (h *Handler) ResetPassword(c *gin.Context) {
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}
	
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	
	// 获取账户
	var account models.Account
	if err := h.db.First(&account, accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	// 加密新密码
	passwordManager := auth.NewPasswordManager()
	passwordHash, err := passwordManager.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// 更新密码
	if err := h.db.Model(&account).Update("password_hash", passwordHash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}
	
	h.response.Success(c, gin.H{"message": "Password reset successfully"})
}