package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"

	"gitlab-merge-alert-go/internal/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateProfile(c *gin.Context) {
	accountID := c.GetUint("account_id")
	
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Email != "" {
		var existingAccount models.Account
		if err := h.db.Where("email = ? AND id != ?", req.Email, accountID).First(&existingAccount).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已被使用"})
			return
		}
		updates["email"] = req.Email
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有需要更新的内容"})
		return
	}

	if err := h.db.Model(&models.Account{}).Where("id = ?", accountID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	var account models.Account
	if err := h.db.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取更新后的账户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"account": account.ToResponse(),
	})
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	accountID := c.GetUint("account_id")
	
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法获取上传文件"})
		return
	}
	defer file.Close()

	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过5MB"})
		return
	}

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能上传图片文件"})
		return
	}

	buffer := make([]byte, header.Size)
	if _, err := file.Read(buffer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	dataURI := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(buffer)

	if err := h.db.Model(&models.Account{}).Where("id = ?", accountID).Update("avatar", dataURI).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存头像失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "头像上传成功",
		"avatar": dataURI,
	})
}

