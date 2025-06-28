package handlers

import (
	"net/http"
	"strconv"

	"gitlab-merge-alert-go/internal/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetProjects(c *gin.Context) {
	var projects []models.Project
	if err := h.db.Preload("Webhooks").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	// 转换为响应格式
	var responses []models.ProjectResponse
	for _, project := range projects {
		response := models.ProjectResponse{
			ID:              project.ID,
			GitLabProjectID: project.GitLabProjectID,
			Name:            project.Name,
			URL:             project.URL,
			Description:     project.Description,
			CreatedAt:       project.CreatedAt,
			UpdatedAt:       project.UpdatedAt,
		}

		// 转换关联的webhooks
		for _, webhook := range project.Webhooks {
			response.Webhooks = append(response.Webhooks, models.WebhookResponse{
				ID:          webhook.ID,
				Name:        webhook.Name,
				URL:         webhook.URL,
				Description: webhook.Description,
				IsActive:    webhook.IsActive,
				CreatedAt:   webhook.CreatedAt,
				UpdatedAt:   webhook.UpdatedAt,
			})
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证GitLab项目是否存在
	if h.gitlabService != nil {
		_, err := h.gitlabService.GetProject(req.GitLabProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "GitLab project not found or access denied"})
			return
		}
	}

	project := &models.Project{
		GitLabProjectID: req.GitLabProjectID,
		Name:            req.Name,
		URL:             req.URL,
		Description:     req.Description,
		AccessToken:     req.AccessToken,
	}

	if err := h.db.Create(project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	response := models.ProjectResponse{
		ID:              project.ID,
		GitLabProjectID: project.GitLabProjectID,
		Name:            project.Name,
		URL:             project.URL,
		Description:     project.Description,
		CreatedAt:       project.CreatedAt,
		UpdatedAt:       project.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 更新字段
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.URL != "" {
		project.URL = req.URL
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.AccessToken != "" {
		project.AccessToken = req.AccessToken
	}

	if err := h.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	response := models.ProjectResponse{
		ID:              project.ID,
		GitLabProjectID: project.GitLabProjectID,
		Name:            project.Name,
		URL:             project.URL,
		Description:     project.Description,
		CreatedAt:       project.CreatedAt,
		UpdatedAt:       project.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.db.Delete(&models.Project{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func (h *Handler) ProjectsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "projects.html", gin.H{
		"title": "项目管理",
	})
}