package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		logger.GetLogger().Errorf("Failed to fetch users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	logger.GetLogger().Debugf("Successfully fetched %d users", len(users))

	// 转换为响应格式
	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, models.UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			Phone:          user.Phone,
			Name:           user.Name,
			GitLabUsername: user.GitLabUsername,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		Email:          req.Email,
		Phone:          req.Phone,
		Name:           req.Name,
		GitLabUsername: req.GitLabUsername,
	}

	if err := h.db.Create(user).Error; err != nil {
		logger.GetLogger().Errorf("Failed to create user [Email: %s, Phone: %s]: %v", req.Email, req.Phone, err)

		// 检查是否是唯一约束冲突
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱地址已存在，请使用其他邮箱"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		}
		return
	}

	logger.GetLogger().Infof("Successfully created user [ID: %d, Email: %s]", user.ID, user.Email)

	response := models.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		Phone:          user.Phone,
		Name:           user.Name,
		GitLabUsername: user.GitLabUsername,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Warnf("User not found [ID: %d]", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			logger.GetLogger().Errorf("Failed to fetch user [ID: %d]: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 更新字段
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	// 允许清空 GitLab 用户名
	user.GitLabUsername = req.GitLabUsername

	if err := h.db.Save(&user).Error; err != nil {
		logger.GetLogger().Errorf("Failed to update user [ID: %d]: %v", id, err)

		// 检查是否是唯一约束冲突
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱地址已存在，请使用其他邮箱"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		}
		return
	}

	logger.GetLogger().Infof("Successfully updated user [ID: %d, Email: %s]", user.ID, user.Email)

	response := models.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		Phone:          user.Phone,
		Name:           user.Name,
		GitLabUsername: user.GitLabUsername,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.db.Delete(&models.User{}, id).Error; err != nil {
		logger.GetLogger().Errorf("Failed to delete user [ID: %d]: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	logger.GetLogger().Infof("Successfully deleted user [ID: %d]", id)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *Handler) UsersPage(c *gin.Context) {
	data := gin.H{
		"title":       "用户管理",
		"currentPage": "users",
	}

	if err := h.renderTemplate(c, "users.html", data); err != nil {
		logger.GetLogger().Errorf("Failed to render users template: %v", err)
		h.renderErrorPage(c, "用户管理页面加载失败")
		return
	}
}
