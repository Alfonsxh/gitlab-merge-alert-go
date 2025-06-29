package handlers

import (
	"net/http"
	"strconv"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

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

// ParseProjectURL 解析GitLab项目URL并返回项目信息
func (h *Handler) ParseProjectURL(c *gin.Context) {
	var req models.ParseProjectURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 使用GitLab服务解析URL并获取项目信息
	projectInfo, err := h.gitlabService.GetProjectByURL(req.URL, req.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	response := models.ParseProjectURLResponse{
		GitLabProjectID:   projectInfo.ID,
		Name:              projectInfo.Name,
		Description:       projectInfo.Description,
		WebURL:            projectInfo.WebURL,
		PathWithNamespace: projectInfo.PathWithNamespace,
		DefaultBranch:     projectInfo.DefaultBranch,
		Visibility:        projectInfo.Visibility,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// TestGitLabConnection 测试GitLab连接
func (h *Handler) TestGitLabConnection(c *gin.Context) {
	var req models.GitLabConnectionTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 首先解析URL以提取基础URL
	parsed := h.gitlabService.ParseGitLabURL(req.URL)
	if !parsed.IsValid {
		response := models.GitLabConnectionTestResponse{
			Success: false,
			Message: "URL格式无效: " + parsed.Error,
		}
		c.JSON(http.StatusOK, gin.H{"data": response})
		return
	}

	// 测试连接
	err := h.gitlabService.TestConnection(parsed.BaseURL, req.AccessToken)

	response := models.GitLabConnectionTestResponse{}
	if err != nil {
		response.Success = false
		response.Message = err.Error()
	} else {
		response.Success = true
		response.Message = "连接成功"
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// ScanGroupProjects 扫描组项目
func (h *Handler) ScanGroupProjects(c *gin.Context) {
	var req models.ScanGroupProjectsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 解析URL
	parsed := h.gitlabService.ParseGitLabURL(req.URL)
	if !parsed.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL解析失败: " + parsed.Error})
		return
	}

	// 首先尝试作为组解析
	groupInfo, err := h.gitlabService.GetGroupByPath(parsed.BaseURL, parsed.ProjectPath, req.AccessToken)
	if err != nil {
		// 如果不是组，尝试作为项目解析
		projectInfo, projectErr := h.gitlabService.GetProjectByPath(parsed.BaseURL, parsed.ProjectPath, req.AccessToken)
		if projectErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法识别为组或项目: " + err.Error()})
			return
		}

		// 是单个项目，返回单个项目信息
		response := models.ScanGroupProjectsResponse{
			GroupInfo: nil,
			Projects: []*models.GitLabProjectInfo{
				{
					ID:                projectInfo.ID,
					Name:              projectInfo.Name,
					PathWithNamespace: projectInfo.PathWithNamespace,
					WebURL:            projectInfo.WebURL,
					Description:       projectInfo.Description,
					DefaultBranch:     projectInfo.DefaultBranch,
					Visibility:        projectInfo.Visibility,
					Selected:          true, // 单个项目默认选中
				},
			},
		}
		c.JSON(http.StatusOK, gin.H{"data": response})
		return
	}

	// 是组，获取组下所有项目
	projects, err := h.gitlabService.GetGroupProjects(parsed.BaseURL, parsed.ProjectPath, req.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取组项目失败: " + err.Error()})
		return
	}

	// 转换为响应格式
	var projectInfos []*models.GitLabProjectInfo
	for _, project := range projects {
		projectInfos = append(projectInfos, &models.GitLabProjectInfo{
			ID:                project.ID,
			Name:              project.Name,
			PathWithNamespace: project.PathWithNamespace,
			WebURL:            project.WebURL,
			Description:       project.Description,
			DefaultBranch:     project.DefaultBranch,
			Visibility:        project.Visibility,
			Selected:          false, // 批量项目默认不选中，让用户选择
		})
	}

	response := models.ScanGroupProjectsResponse{
		GroupInfo: &models.GitLabGroupInfo{
			ID:       groupInfo.ID,
			Name:     groupInfo.Name,
			Path:     groupInfo.Path,
			FullPath: groupInfo.FullPath,
			WebURL:   groupInfo.WebURL,
		},
		Projects: projectInfos,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// BatchCreateProjects 批量创建项目
func (h *Handler) BatchCreateProjects(c *gin.Context) {
	var req models.BatchCreateProjectsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	var results []models.BatchProjectResult
	successCount := 0
	failureCount := 0

	// 开始数据库事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 处理webhook配置
	var webhookID uint
	if req.WebhookConfig.UseUnified {
		if req.WebhookConfig.NewWebhook != nil {
			// 创建新的统一webhook
			webhook := &models.Webhook{
				Name:        req.WebhookConfig.NewWebhook.Name,
				URL:         req.WebhookConfig.NewWebhook.URL,
				Description: req.WebhookConfig.NewWebhook.Description,
				IsActive:    true,
			}
			if req.WebhookConfig.NewWebhook.IsActive != nil {
				webhook.IsActive = *req.WebhookConfig.NewWebhook.IsActive
			}

			if err := tx.Create(webhook).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建统一webhook失败: " + err.Error()})
				return
			}
			webhookID = webhook.ID
		} else if req.WebhookConfig.UnifiedWebhookID != nil {
			// 使用现有的统一webhook
			webhookID = *req.WebhookConfig.UnifiedWebhookID
		}
	}

	// 批量创建项目
	for _, projectInfo := range req.Projects {
		result := models.BatchProjectResult{
			GitLabProjectID: projectInfo.GitLabProjectID,
			Name:            projectInfo.Name,
		}

		// 检查项目是否已存在
		var existingProject models.Project
		if err := tx.Where("gitlab_project_id = ?", projectInfo.GitLabProjectID).First(&existingProject).Error; err == nil {
			result.Success = false
			result.Error = "项目已存在"
			results = append(results, result)
			failureCount++
			continue
		}

		// 创建项目
		project := &models.Project{
			GitLabProjectID: projectInfo.GitLabProjectID,
			Name:            projectInfo.Name,
			URL:             projectInfo.URL,
			Description:     projectInfo.Description,
			AccessToken:     req.AccessToken,
		}

		if err := tx.Create(project).Error; err != nil {
			result.Success = false
			result.Error = "创建项目失败: " + err.Error()
			results = append(results, result)
			failureCount++
			continue
		}

		result.Success = true
		result.ProjectID = project.ID
		results = append(results, result)
		successCount++

		// 关联webhook
		if req.WebhookConfig.UseUnified && webhookID > 0 {
			// 使用统一webhook
			projectWebhook := &models.ProjectWebhook{
				ProjectID: project.ID,
				WebhookID: webhookID,
			}
			if err := tx.Create(projectWebhook).Error; err != nil {
				// webhook关联失败不影响项目创建，记录错误即可
				result.Error = "项目创建成功，但webhook关联失败: " + err.Error()
			}
		} else if !req.WebhookConfig.UseUnified {
			// 使用单独配置的webhook
			for _, mapping := range req.WebhookConfig.ProjectWebhooks {
				if mapping.GitLabProjectID == projectInfo.GitLabProjectID {
					projectWebhook := &models.ProjectWebhook{
						ProjectID: project.ID,
						WebhookID: mapping.WebhookID,
					}
					if err := tx.Create(projectWebhook).Error; err != nil {
						result.Error = "项目创建成功，但webhook关联失败: " + err.Error()
					}
					break
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败: " + err.Error()})
		return
	}

	response := models.BatchCreateProjectsResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

func (h *Handler) ProjectsPage(c *gin.Context) {
	data := gin.H{
		"title":       "项目管理",
		"currentPage": "projects",
	}

	if err := h.renderTemplate(c, "projects.html", data); err != nil {
		logger.GetLogger().Errorf("Failed to render projects template: %v", err)
		h.renderErrorPage(c, "项目管理页面加载失败")
		return
	}
}
