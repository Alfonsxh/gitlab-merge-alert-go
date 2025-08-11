package handlers

import (
	"fmt"
	"net/http"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/internal/services"
	"github.com/gin-gonic/gin"
)

func (h *Handler) AssignManager(c *gin.Context) {
	accountID := c.GetUint("account_id")
	role := c.GetString("role")
	
	if role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以分配管理员"})
		return
	}

	var req models.AssignManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rmService := services.NewResourceManagerService(h.db)
	if err := rmService.AssignManager(accountID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "管理员分配成功"})
}

func (h *Handler) RemoveManager(c *gin.Context) {
	accountID := c.GetUint("account_id")
	role := c.GetString("role")
	
	if role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以移除管理员"})
		return
	}

	var req models.RemoveManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rmService := services.NewResourceManagerService(h.db)
	if err := rmService.RemoveManager(accountID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "管理员移除成功"})
}

func (h *Handler) GetResourceManagers(c *gin.Context) {
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	
	if resourceType == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源类型和ID必须提供"})
		return
	}

	var rid uint
	if _, err := fmt.Sscanf(resourceID, "%d", &rid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	rmService := services.NewResourceManagerService(h.db)
	managers, err := rmService.GetResourceManagers(rid, models.ResourceType(resourceType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取管理员列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"managers": managers,
		"total": len(managers),
	})
}

func (h *Handler) GetManagedResources(c *gin.Context) {
	idStr := c.Param("id")
	var targetAccountID uint
	if _, err := fmt.Sscanf(idStr, "%d", &targetAccountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}
	
	resourceType := c.Query("resource_type")
	if resourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源类型必须提供"})
		return
	}

	rmService := services.NewResourceManagerService(h.db)
	resourceIDs, err := rmService.GetManagedResources(targetAccountID, models.ResourceType(resourceType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取管理的资源失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resource_ids": resourceIDs,
		"total": len(resourceIDs),
	})
}

func (h *Handler) BatchAssignResources(c *gin.Context) {
	accountID := c.GetUint("account_id")
	
	idStr := c.Param("id")
	var targetAccountID uint
	if _, err := fmt.Sscanf(idStr, "%d", &targetAccountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	var req struct {
		Assignments []models.AssignManagerRequest `json:"assignments"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rmService := services.NewResourceManagerService(h.db)
	
	// 构建新分配的映射
	newAssignments := make(map[string]map[uint]bool)
	for _, resourceType := range []models.ResourceType{
		models.ResourceTypeProject,
		models.ResourceTypeWebhook,
		models.ResourceTypeUser,
	} {
		newAssignments[string(resourceType)] = make(map[uint]bool)
	}
	
	for _, assignment := range req.Assignments {
		newAssignments[string(assignment.ResourceType)][assignment.ResourceID] = true
	}
	
	// 获取现有分配并处理需要删除的
	for _, resourceType := range []models.ResourceType{
		models.ResourceTypeProject,
		models.ResourceTypeWebhook,
		models.ResourceTypeUser,
	} {
		existingResources, _ := rmService.GetManagedResources(targetAccountID, resourceType)
		
		// 删除不在新分配列表中的资源
		for _, resourceID := range existingResources {
			if !newAssignments[string(resourceType)][resourceID] {
				rmService.RemoveManager(accountID, &models.RemoveManagerRequest{
					ResourceID:   resourceID,
					ResourceType: resourceType,
					ManagerID:    targetAccountID,
				})
			}
		}
	}
	
	// 添加新的分配（会自动跳过已存在的）
	for _, assignment := range req.Assignments {
		assignment.ManagerID = targetAccountID
		if err := rmService.AssignManager(accountID, &assignment); err != nil {
			// 如果是"已经被分配"的错误，忽略它
			if err.Error() != "该管理员已经被分配到此资源" {
				c.Error(err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "资源批量分配成功"})
}