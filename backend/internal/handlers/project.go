package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab-merge-alert-go/internal/middleware"
	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/internal/services"
	"gitlab-merge-alert-go/pkg/logger"
	"gitlab-merge-alert-go/pkg/security"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	errGitLabTokenMissing = errors.New("gitlab personal access token not configured")
	errUnauthorized       = errors.New("unauthorized")
)

func (h *Handler) resolveGitLabToken(c *gin.Context, provided string) (string, error) {
	if token := strings.TrimSpace(provided); token != "" {
		return token, nil
	}

	accountID, exists := middleware.GetAccountID(c)
	if !exists {
		return "", errUnauthorized
	}

	var account models.Account
	if err := h.db.Select("gitlab_access_token").First(&account, accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errUnauthorized
		}
		return "", err
	}

	if account.GitLabAccessToken == "" {
		return "", errGitLabTokenMissing
	}

	decrypted, err := security.Decrypt(h.config.EncryptionKey, account.GitLabAccessToken)
	var token string
	if err != nil {
		logger.GetLogger().Warnf("Failed to decrypt GitLab token for account %d, fallback to legacy plaintext: %v", accountID, err)
		token = strings.TrimSpace(account.GitLabAccessToken)
	} else {
		token = strings.TrimSpace(decrypted)
	}
	if token == "" {
		return "", errGitLabTokenMissing
	}

	return token, nil
}

func (h *Handler) GetProjects(c *gin.Context) {
	var projects []models.Project

	// 应用所有权过滤
	query := h.db.Preload("Webhooks")
	query = middleware.ApplyOwnershipFilter(c, query, "projects")

	if err := query.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	// 转换为响应格式
	responses := make([]models.ProjectResponse, 0) // 确保是空数组而不是nil
	for _, project := range projects {
		response := models.ProjectResponse{
			ID:                project.ID,
			GitLabProjectID:   project.GitLabProjectID,
			Name:              project.Name,
			URL:               project.URL,
			Description:       project.Description,
			GitLabWebhookID:   project.GitLabWebhookID,
			WebhookSynced:     project.WebhookSynced,
			AutoManageWebhook: project.AutoManageWebhook,
			LastSyncAt:        project.LastSyncAt,
			CreatedAt:         project.CreatedAt,
			UpdatedAt:         project.UpdatedAt,
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

	h.response.Success(c, responses)
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	token, err := h.resolveGitLabToken(c, req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, errUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		case errors.Is(err, errGitLabTokenMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前账户未配置 GitLab Personal Access Token，请先在账户管理中设置"})
		default:
			logger.GetLogger().Errorf("Failed to resolve GitLab token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析凭证失败"})
		}
		return
	}

	// 验证GitLab项目是否存在
	if h.gitlabService != nil {
		// 使用解析后的token进行验证
		if _, err := h.gitlabService.GetProject(req.GitLabProjectID, token); err != nil {
			logger.GetLogger().Errorf("Failed to fetch GitLab project [ID: %d]: %v", req.GitLabProjectID, err)
			h.response.ErrorWithMessage(c, "保存项目失败: GitLab项目不存在或访问被拒绝")
			return
		}
	}

	// 设置默认值
	autoManageWebhook := true
	if req.AutoManageWebhook != nil {
		autoManageWebhook = *req.AutoManageWebhook
	}

	// 检查项目是否已存在
	var existingProject models.Project
	err = h.db.Where(&models.Project{GitLabProjectID: req.GitLabProjectID}).First(&existingProject).Error
	if err == nil {
		// 项目已存在
		logger.GetLogger().Warnf("Attempt to create existing project [GitLab ID: %d, Name: %s] from IP: %s", req.GitLabProjectID, req.Name, c.ClientIP())
		h.response.Conflict(c, "GitLab项目ID已存在，如需重新配置请先删除现有项目")
		return
	} else if err != gorm.ErrRecordNotFound {
		// 其他数据库错误
		logger.GetLogger().Errorf("Database error while checking existing project [GitLab ID: %d]: %v", req.GitLabProjectID, err)
		h.response.InternalError(c, "数据库查询失败")
		return
	}

	// 获取当前用户ID
	accountID, _ := middleware.GetAccountID(c)

	// 创建新项目
	project := &models.Project{
		GitLabProjectID:   req.GitLabProjectID,
		Name:              req.Name,
		URL:               req.URL,
		Description:       req.Description,
		AutoManageWebhook: autoManageWebhook,
		WebhookSynced:     false,
		CreatedBy:         &accountID,
	}

	if err := h.db.Create(project).Error; err != nil {
		logger.GetLogger().Errorf("Failed to create project [GitLab ID: %d]: %v", req.GitLabProjectID, err)

		// 处理UNIQUE约束冲突
		if strings.Contains(err.Error(), "UNIQUE constraint failed: projects.gitlab_project_id") {
			h.response.Conflict(c, "GitLab项目ID已存在，如需重新配置请先删除现有项目")
		} else {
			h.response.InternalError(c, "创建项目���败")
		}
		return
	}

	// 关联 Webhook
	if req.WebhookID != nil {
		projectWebhook := &models.ProjectWebhook{
			ProjectID: project.ID,
			WebhookID: *req.WebhookID,
		}
		if err := h.db.Create(projectWebhook).Error; err != nil {
			// 即使关联失败，项目也已创建成功，只记录日志
			logger.GetLogger().Errorf("Failed to associate webhook [ID: %d] with project [ID: %d]: %v",
				*req.WebhookID, project.ID, err)
		}
	}

	// 如果启用了自动管理webhook，尝试创建GitLab webhook
	if project.AutoManageWebhook {
		h.autoCreateGitLabWebhook(project, token)
	}

	logger.GetLogger().Infof("Successfully created project [ID: %d, GitLab ID: %d, Name: %s] from IP: %s",
		project.ID, project.GitLabProjectID, project.Name, c.ClientIP())

	response := models.ProjectResponse{
		ID:                project.ID,
		GitLabProjectID:   project.GitLabProjectID,
		Name:              project.Name,
		URL:               project.URL,
		Description:       project.Description,
		GitLabWebhookID:   project.GitLabWebhookID,
		WebhookSynced:     project.WebhookSynced,
		AutoManageWebhook: project.AutoManageWebhook,
		LastSyncAt:        project.LastSyncAt,
		CreatedAt:         project.CreatedAt,
		UpdatedAt:         project.UpdatedAt,
	}

	h.response.Created(c, response)
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
	providedToken := strings.TrimSpace(req.AccessToken)
	if req.AutoManageWebhook != nil {
		project.AutoManageWebhook = *req.AutoManageWebhook
	}

	// 更新 Webhook 关联
	if req.WebhookIDs != nil {
		// 使用事务确保原子性
		err := h.db.Transaction(func(tx *gorm.DB) error {
			// 1. 删除现有所有关联
			if err := tx.Where("project_id = ?", project.ID).Delete(&models.ProjectWebhook{}).Error; err != nil {
				return err
			}

			// 2. 创建新的关联
			if len(req.WebhookIDs) > 0 {
				var newAssociations []models.ProjectWebhook
				for _, webhookID := range req.WebhookIDs {
					newAssociations = append(newAssociations, models.ProjectWebhook{
						ProjectID: project.ID,
						WebhookID: webhookID,
					})
				}
				if err := tx.Create(&newAssociations).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			logger.GetLogger().Errorf("Failed to update project webhooks [ID: %d]: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目关联的webhook失败"})
			return
		}
	}

	if project.AutoManageWebhook {
		token, err := h.resolveGitLabToken(c, providedToken)
		if err != nil {
			if errors.Is(err, errGitLabTokenMissing) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "当前账户未配置 GitLab Personal Access Token，无法启用自动Webhook管理"})
				return
			}
			if errors.Is(err, errUnauthorized) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
			logger.GetLogger().Errorf("Failed to resolve GitLab token for project update [ID: %d]: %v", project.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目失败"})
			return
		}
		h.autoCreateGitLabWebhook(&project, token)
	}

	if err := h.db.Save(&project).Error; err != nil {
		logger.GetLogger().Errorf("Failed to update project [ID: %d]: %v", id, err)

		if strings.Contains(err.Error(), "UNIQUE constraint failed: projects.gitlab_project_id") {
			c.JSON(http.StatusConflict, gin.H{"error": "GitLab项目ID已存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目失败"})
		}
		return
	}

	logger.GetLogger().Infof("Successfully updated project [ID: %d, Name: %s]", project.ID, project.Name)

	// 重新加载项目以包含更新后的 Webhooks
	if err := h.db.Preload("Webhooks").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found after update"})
		return
	}

	response := models.ProjectResponse{
		ID:                project.ID,
		GitLabProjectID:   project.GitLabProjectID,
		Name:              project.Name,
		URL:               project.URL,
		Description:       project.Description,
		GitLabWebhookID:   project.GitLabWebhookID,
		WebhookSynced:     project.WebhookSynced,
		AutoManageWebhook: project.AutoManageWebhook,
		LastSyncAt:        project.LastSyncAt,
		CreatedAt:         project.CreatedAt,
		UpdatedAt:         project.UpdatedAt,
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

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// 获取项目信息用于清理webhook
	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 如果有GitLab webhook，尝试删除
	if project.GitLabWebhookID != nil {
		token, tokenErr := h.resolveGitLabToken(c, "")
		if tokenErr != nil {
			if errors.Is(tokenErr, errUnauthorized) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
			if errors.Is(tokenErr, errGitLabTokenMissing) {
				logger.GetLogger().Warnf("Skip GitLab webhook cleanup for project %d: token not configured", project.ID)
			} else {
				logger.GetLogger().Warnf("Skip GitLab webhook cleanup for project %d due to token error: %v", project.ID, tokenErr)
			}
		} else if token != "" {
			h.autoDeleteGitLabWebhook(&project, token)
		}
	}

	if err := h.db.Delete(&models.Project{}, id).Error; err != nil {
		logger.GetLogger().Errorf("Failed to delete project [ID: %d]: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除项目失败"})
		return
	}

	logger.GetLogger().Infof("Successfully deleted project [ID: %d]", id)

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// ParseProjectURL 解析GitLab项目URL并返回项目信息
func (h *Handler) ParseProjectURL(c *gin.Context) {
	var req models.ParseProjectURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	token, err := h.resolveGitLabToken(c, req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, errUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		case errors.Is(err, errGitLabTokenMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先在账户管理中配置 GitLab Personal Access Token"})
		default:
			logger.GetLogger().Errorf("Failed to resolve GitLab token for parse URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析凭证失败"})
		}
		return
	}

	// 使用GitLab服务解析URL并获取项目信息
	projectInfo, err := h.gitlabService.GetProjectByURL(req.URL, token)
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

	token, err := h.resolveGitLabToken(c, req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, errUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		case errors.Is(err, errGitLabTokenMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先在账户管理中配置 GitLab Personal Access Token"})
		default:
			logger.GetLogger().Errorf("Failed to resolve GitLab token for connection test: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析凭证失败"})
		}
		return
	}

	// 测试连接
	err = h.gitlabService.TestConnection(parsed.BaseURL, token)

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

	token, err := h.resolveGitLabToken(c, req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, errUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		case errors.Is(err, errGitLabTokenMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先在账户管理中配置 GitLab Personal Access Token"})
		default:
			logger.GetLogger().Errorf("Failed to resolve GitLab token for scanning group: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析凭证失败"})
		}
		return
	}

	// 首先尝试作为组解析
	groupInfo, err := h.gitlabService.GetGroupByPath(parsed.BaseURL, parsed.ProjectPath, token)
	if err != nil {
		// 如果不是组，尝试作为项目解析
		projectInfo, projectErr := h.gitlabService.GetProjectByPath(parsed.BaseURL, parsed.ProjectPath, token)
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
	projects, err := h.gitlabService.GetGroupProjects(parsed.BaseURL, parsed.ProjectPath, token)
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

	// 获取当前用户ID
	accountID, _ := middleware.GetAccountID(c)

	token, err := h.resolveGitLabToken(c, req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, errUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		case errors.Is(err, errGitLabTokenMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先在账户管理中配置 GitLab Personal Access Token"})
		default:
			logger.GetLogger().Errorf("Failed to resolve GitLab token for batch create: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析凭证失败"})
		}
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
				CreatedBy:   &accountID,
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

		var projectToAssociate *models.Project

		// 检查项目是否已存在
		var existingProject models.Project
		err := tx.Where(&models.Project{GitLabProjectID: projectInfo.GitLabProjectID}).First(&existingProject).Error
		if err == nil {
			// 项目已存在
			result.Success = false
			result.Error = "项目已存在"
			results = append(results, result)
			failureCount++
			continue
		} else if err != gorm.ErrRecordNotFound {
			// 数据库查询错误
			result.Success = false
			result.Error = "数据库查询失败: " + err.Error()
			results = append(results, result)
			failureCount++
			continue
		}

		// 创建新项目
		project := &models.Project{
			GitLabProjectID:   projectInfo.GitLabProjectID,
			Name:              projectInfo.Name,
			URL:               projectInfo.URL,
			Description:       projectInfo.Description,
			AutoManageWebhook: true, // 批量创建时默认启用自动管理
			WebhookSynced:     false,
			CreatedBy:         &accountID,
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
		projectToAssociate = project

		// 在事务提交后异步创建GitLab webhook
		defer func(p *models.Project) {
			go h.autoCreateGitLabWebhook(p, token)
		}(project)

		// 关联webhook
		if projectToAssociate != nil {
			if req.WebhookConfig.UseUnified && webhookID > 0 {
				// 使用统一webhook - 检查是否已存在关联
				var existingAssociation models.ProjectWebhook
				err := tx.Where("project_id = ? AND webhook_id = ?", projectToAssociate.ID, webhookID).First(&existingAssociation).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 不存在关联，创建新的
					projectWebhook := &models.ProjectWebhook{
						ProjectID: projectToAssociate.ID,
						WebhookID: webhookID,
					}
					if err := tx.Create(projectWebhook).Error; err != nil {
						// webhook关联失败不影响项目创建，记录错误即可
						result.Error = "项目创建成功，但webhook关联失败: " + err.Error()
					}
				} else if err != nil {
					// 数据库查询错误
					result.Error = "项目创建成功，但检查webhook关联失败: " + err.Error()
				}
				// 如果关联已存在，则不需要任何操作
			} else if !req.WebhookConfig.UseUnified {
				// 使用单独配置的webhook
				for _, mapping := range req.WebhookConfig.ProjectWebhooks {
					if mapping.GitLabProjectID == projectInfo.GitLabProjectID {
						// 检查是否已存在关联
						var existingAssociation models.ProjectWebhook
						err := tx.Where("project_id = ? AND webhook_id = ?", projectToAssociate.ID, mapping.WebhookID).First(&existingAssociation).Error
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// 不存在关联，创建新的
							projectWebhook := &models.ProjectWebhook{
								ProjectID: projectToAssociate.ID,
								WebhookID: mapping.WebhookID,
							}
							if err := tx.Create(projectWebhook).Error; err != nil {
								result.Error = "项目创建成功，但webhook关联失败: " + err.Error()
							}
						} else if err != nil {
							// 数据库查询错误
							result.Error = "项目创建成功，但检查webhook关联失败: " + err.Error()
						}
						// 如果关联已存在，则不需要任何操作
						break
					}
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败: " + err.Error()})
		return
	}

	// 转换为新的批量响应格式
	var batchResults []models.BatchResultItem
	for _, result := range results {
		batchResults = append(batchResults, models.BatchResultItem{
			ID:      result.GitLabProjectID,
			Name:    result.Name,
			Success: result.Success,
			Error:   result.Error,
			Data:    map[string]interface{}{"project_id": result.ProjectID},
		})
	}

	// 使用新的统一响应格式
	h.response.BatchOperation(c, successCount, failureCount, batchResults)
}

// SyncGitLabWebhook 同步GitLab Webhook
func (h *Handler) SyncGitLabWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 检查是否启用自动管理webhook
	if !project.AutoManageWebhook {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目未启用自动webhook管理"})
		return
	}

	// 解析GitLab URL以获取基础URL
	parsed := h.gitlabService.ParseGitLabURL(project.URL)
	if !parsed.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目URL格式无效: " + parsed.Error})
		return
	}

	// 构建webhook URL
	webhookURL := h.gitlabService.BuildWebhookURL(h.config.PublicWebhookURL)

	token, err := h.resolveGitLabToken(c, "")
	if err != nil {
		if errors.Is(err, errUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		} else if errors.Is(err, errGitLabTokenMissing) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前账户未配置 GitLab Personal Access Token"})
		} else {
			logger.GetLogger().Errorf("Failed to resolve GitLab token for sync [project ID: %d]: %v", project.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "同步失败: 无法解析凭证"})
		}
		return
	}

	// 检查是否已存在相同的webhook
	existingWebhook, err := h.gitlabService.FindWebhookByURL(parsed.BaseURL, project.GitLabProjectID, webhookURL, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查现有webhook失败: " + err.Error()})
		return
	}

	var response models.SyncGitLabWebhookResponse
	now := time.Now()

	if existingWebhook != nil {
		// webhook已存在，更新项目状态
		project.GitLabWebhookID = &existingWebhook.ID
		project.WebhookSynced = true
		project.LastSyncAt = &now

		response = models.SyncGitLabWebhookResponse{
			Success:         true,
			Message:         "Webhook已存在，状态已更新",
			GitLabWebhookID: &existingWebhook.ID,
			WebhookURL:      webhookURL,
		}
	} else {
		// 创建新的webhook
		gitlabService := services.NewGitLabService(parsed.BaseURL, token)
		webhook, err := gitlabService.CreateProjectWebhook(parsed.BaseURL, project.GitLabProjectID, webhookURL, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建GitLab webhook失败: " + err.Error()})
			return
		}

		// 更新项目状态
		project.GitLabWebhookID = &webhook.ID
		project.WebhookSynced = true
		project.LastSyncAt = &now

		response = models.SyncGitLabWebhookResponse{
			Success:         true,
			Message:         "GitLab webhook创建成功",
			GitLabWebhookID: &webhook.ID,
			WebhookURL:      webhookURL,
		}
	}

	// 保存项目状态
	if err := h.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存项目状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// DeleteGitLabWebhook 删除GitLab Webhook
func (h *Handler) DeleteGitLabWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 解析GitLab URL以获取基础URL
	parsed := h.gitlabService.ParseGitLabURL(project.URL)
	if !parsed.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目URL格式无效: " + parsed.Error})
		return
	}

	// 构建webhook URL
	webhookURL := h.gitlabService.BuildWebhookURL(h.config.PublicWebhookURL)

	token, err := h.resolveGitLabToken(c, "")
	if err != nil {
		if errors.Is(err, errUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		} else if errors.Is(err, errGitLabTokenMissing) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前账户未配置 GitLab Personal Access Token"})
		} else {
			logger.GetLogger().Errorf("Failed to resolve GitLab token for delete webhook [project ID: %d]: %v", project.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除GitLab webhook失败: 无法解析凭证"})
		}
		return
	}

	// 删除GitLab中所有匹配的webhook
	gitlabService := services.NewGitLabService(parsed.BaseURL, token)
	deletedCount, err := gitlabService.DeleteAllWebhooksByURL(parsed.BaseURL, project.GitLabProjectID, webhookURL, token)

	var responseMessage string
	if err != nil {
		logger.GetLogger().Warnf("删除GitLab webhook失败: %v (已删除 %d 个)", err, deletedCount)
		if deletedCount > 0 {
			responseMessage = fmt.Sprintf("部分删除成功 (已删除 %d 个webhook)", deletedCount)
		} else {
			responseMessage = "删除失败，webhook可能已被手动删除"
		}
	} else if deletedCount > 0 {
		if deletedCount == 1 {
			responseMessage = "GitLab webhook已删除"
		} else {
			responseMessage = fmt.Sprintf("GitLab webhook已删除 (共删除 %d 个重复webhook)", deletedCount)
		}
	} else {
		responseMessage = "未找到匹配的webhook，可能已被删除"
	}

	// 更新项目状态
	project.GitLabWebhookID = nil
	project.WebhookSynced = false
	now := time.Now()
	project.LastSyncAt = &now

	if err := h.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存项目状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      responseMessage,
		"deletedCount": deletedCount,
	})
}

// GetGitLabWebhookStatus 获取GitLab Webhook状态
func (h *Handler) GetGitLabWebhookStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	webhookURL := h.gitlabService.BuildWebhookURL(h.config.PublicWebhookURL)

	// 检查是否有权限管理webhook（通过测试连接来判断）
	canManage := false
	token, tokenErr := h.resolveGitLabToken(c, "")
	if tokenErr != nil {
		if errors.Is(tokenErr, errUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		if tokenErr != errGitLabTokenMissing {
			logger.GetLogger().Warnf("Failed to resolve GitLab token when checking webhook status for project %d: %v", project.ID, tokenErr)
		}
	} else if token != "" {
		parsed := h.gitlabService.ParseGitLabURL(project.URL)
		if parsed.IsValid {
			if err := h.gitlabService.TestConnection(parsed.BaseURL, token); err == nil {
				canManage = true
			}
		}
	}

	response := models.GitLabWebhookStatusResponse{
		ProjectID:       project.ID,
		WebhookSynced:   project.WebhookSynced,
		GitLabWebhookID: project.GitLabWebhookID,
		WebhookURL:      webhookURL,
		LastSyncAt:      project.LastSyncAt,
		CanManage:       canManage,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// autoCreateGitLabWebhook 自动创建GitLab webhook
func (h *Handler) autoCreateGitLabWebhook(project *models.Project, token string) {
	// 解析GitLab URL
	parsed := h.gitlabService.ParseGitLabURL(project.URL)
	if !parsed.IsValid {
		logger.GetLogger().Warnf("项目 %d URL格式无效，跳过webhook创建: %s", project.ID, parsed.Error)
		return
	}

	// 构建webhook URL
	webhookURL := h.gitlabService.BuildWebhookURL(h.config.PublicWebhookURL)

	// 检查是否已存在相同的webhook
	existingWebhook, err := h.gitlabService.FindWebhookByURL(parsed.BaseURL, project.GitLabProjectID, webhookURL, token)
	if err != nil {
		logger.GetLogger().Warnf("检查项目 %d 现有webhook失败: %v", project.ID, err)
		return
	}

	now := time.Now()

	if existingWebhook != nil {
		// webhook已存在，更新项目状态
		project.GitLabWebhookID = &existingWebhook.ID
		project.WebhookSynced = true
		project.LastSyncAt = &now
		logger.GetLogger().Infof("项目 %d 的GitLab webhook已存在，状态已更新", project.ID)
	} else {
		// 创建新的webhook
		gitlabService := services.NewGitLabService(parsed.BaseURL, token)
		webhook, err := gitlabService.CreateProjectWebhook(parsed.BaseURL, project.GitLabProjectID, webhookURL, token)
		if err != nil {
			logger.GetLogger().Warnf("为项目 %d 创建GitLab webhook失败: %v", project.ID, err)
			return
		}

		// 更新项目状态
		project.GitLabWebhookID = &webhook.ID
		project.WebhookSynced = true
		project.LastSyncAt = &now
		logger.GetLogger().Infof("项目 %d 的GitLab webhook创建成功，ID: %d", project.ID, webhook.ID)
	}

	// 保存项目状态（忽略错误，避免影响主流程）
	if err := h.db.Save(project).Error; err != nil {
		logger.GetLogger().Warnf("保存项目 %d webhook状态失败: %v", project.ID, err)
	}
}

// autoDeleteGitLabWebhook 自动删除GitLab webhook（支持删除多个重复的webhook）
func (h *Handler) autoDeleteGitLabWebhook(project *models.Project, token string) {
	// 解析GitLab URL
	parsed := h.gitlabService.ParseGitLabURL(project.URL)
	if !parsed.IsValid {
		logger.GetLogger().Warnf("项目 %d URL格式无效，跳过webhook删除: %s", project.ID, parsed.Error)
		return
	}

	// 构建webhook URL
	webhookURL := h.gitlabService.BuildWebhookURL(h.config.PublicWebhookURL)

	// 删除GitLab中所有匹配的webhook
	gitlabService := services.NewGitLabService(parsed.BaseURL, token)
	deletedCount, err := gitlabService.DeleteAllWebhooksByURL(parsed.BaseURL, project.GitLabProjectID, webhookURL, token)
	if err != nil {
		logger.GetLogger().Warnf("删除项目 %d 的GitLab webhook失败: %v (已删除 %d 个)", project.ID, err, deletedCount)
		// 即使删除失败也继续，可能是webhook已经被手动删除
	} else if deletedCount > 0 {
		if deletedCount == 1 {
			logger.GetLogger().Infof("项目 %d 的GitLab webhook已删除", project.ID)
		} else {
			logger.GetLogger().Infof("项目 %d 的GitLab webhook已删除 (共删除 %d 个重复webhook)", project.ID, deletedCount)
		}
	} else {
		logger.GetLogger().Infof("项目 %d 未找到匹配的GitLab webhook，可能已被手动删除", project.ID)
	}
}
