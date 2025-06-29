package handlers

import (
	"net/http"
	"strconv"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// 转换为响应格式
	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Phone:     user.Phone,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
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
		Email: req.Email,
		Phone: req.Phone,
		Name:  req.Name,
	}

	if err := h.db.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	response := models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
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
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	response := models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

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