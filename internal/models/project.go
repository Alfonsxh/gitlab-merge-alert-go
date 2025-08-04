package models

import (
	"time"
)

type Project struct {
	ID              uint   `json:"id" gorm:"column:id;primarykey"`
	GitLabProjectID int    `json:"gitlab_project_id" gorm:"column:gitlab_project_id;uniqueIndex;not null;default:0"`
	Name            string `json:"name" gorm:"column:name;not null;default:''"`
	URL             string `json:"url" gorm:"column:url;not null;default:''"`
	Description     string `json:"description" gorm:"column:description"`
	AccessToken     string `json:"-" gorm:"column:access_token"` // 不在JSON中显示敏感信息

	// GitLab Webhook相关字段
	GitLabWebhookID   *int       `json:"gitlab_webhook_id,omitempty" gorm:"column:gitlab_webhook_id;index"`  // GitLab中webhook的ID
	WebhookSynced     bool       `json:"webhook_synced" gorm:"column:webhook_synced;default:false"`          // webhook同步状态
	AutoManageWebhook bool       `json:"auto_manage_webhook" gorm:"column:auto_manage_webhook;default:true"` // 是否自动管理webhook
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty" gorm:"column:last_sync_at"`                  // 最后同步时间

	CreatedBy *uint     `json:"created_by,omitempty" gorm:"column:created_by;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 关联关系
	Webhooks []Webhook `json:"webhooks,omitempty" gorm:"many2many:project_webhooks;"`
}

type CreateProjectRequest struct {
	GitLabProjectID   int    `json:"gitlab_project_id" binding:"required"`
	Name              string `json:"name" binding:"required"`
	URL               string `json:"url" binding:"required,url"`
	Description       string `json:"description"`
	AccessToken       string `json:"access_token"`
	AutoManageWebhook *bool  `json:"auto_manage_webhook,omitempty"` // 可选，默认为true
	WebhookID         *uint  `json:"webhook_id,omitempty"`          // 可选，用于关联的webhook
}

type UpdateProjectRequest struct {
	Name              string `json:"name"`
	URL               string `json:"url" binding:"omitempty,url"`
	Description       string `json:"description"`
	AccessToken       string `json:"access_token"`
	AutoManageWebhook *bool  `json:"auto_manage_webhook,omitempty"`
	WebhookIDs        []uint `json:"webhook_ids"`
}

type ProjectResponse struct {
	ID                uint              `json:"id"`
	GitLabProjectID   int               `json:"gitlab_project_id"`
	Name              string            `json:"name"`
	URL               string            `json:"url"`
	Description       string            `json:"description"`
	GitLabWebhookID   *int              `json:"gitlab_webhook_id,omitempty"`
	WebhookSynced     bool              `json:"webhook_synced"`
	AutoManageWebhook bool              `json:"auto_manage_webhook"`
	LastSyncAt        *time.Time        `json:"last_sync_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Webhooks          []WebhookResponse `json:"webhooks,omitempty"`
}

// ParseProjectURLRequest 解析GitLab项目URL的请求结构
type ParseProjectURLRequest struct {
	URL         string `json:"url" binding:"required,url"`
	AccessToken string `json:"access_token" binding:"required"`
}

// ParseProjectURLResponse 解析GitLab项目URL的响应结构
type ParseProjectURLResponse struct {
	GitLabProjectID   int    `json:"gitlab_project_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url"`
	PathWithNamespace string `json:"path_with_namespace"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
}

// GitLabConnectionTestRequest 测试GitLab连接的请求结构
type GitLabConnectionTestRequest struct {
	URL         string `json:"url" binding:"required,url"`
	AccessToken string `json:"access_token" binding:"required"`
}

// GitLabConnectionTestResponse 测试GitLab连接的响应结构
type GitLabConnectionTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ScanGroupProjectsRequest 扫描组项目的请求结构
type ScanGroupProjectsRequest struct {
	URL         string `json:"url" binding:"required,url"`
	AccessToken string `json:"access_token" binding:"required"`
}

// ScanGroupProjectsResponse 扫描组项目的响应结构
type ScanGroupProjectsResponse struct {
	GroupInfo *GitLabGroupInfo     `json:"group_info"`
	Projects  []*GitLabProjectInfo `json:"projects"`
}

// GitLabGroupInfo GitLab组信息
type GitLabGroupInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
	WebURL   string `json:"web_url"`
}

// GitLabProjectInfo GitLab项目信息
type GitLabProjectInfo struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	Description       string `json:"description"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
	Selected          bool   `json:"selected"` // 前端选择状态
}

// BatchCreateProjectsRequest 批量创建项目的请求结构
type BatchCreateProjectsRequest struct {
	Projects      []BatchProjectInfo `json:"projects" binding:"required"`
	AccessToken   string             `json:"access_token"`
	WebhookConfig BatchWebhookConfig `json:"webhook_config"`
}

// BatchProjectInfo 批量创建项目的项目信息
type BatchProjectInfo struct {
	GitLabProjectID int    `json:"gitlab_project_id" binding:"required"`
	Name            string `json:"name" binding:"required"`
	URL             string `json:"url" binding:"required"`
	Description     string `json:"description"`
}

// BatchWebhookConfig 批量创建项目的webhook配置
type BatchWebhookConfig struct {
	UseUnified       bool                    `json:"use_unified"`                  // 是否使用统一webhook
	UnifiedWebhookID *uint                   `json:"unified_webhook_id,omitempty"` // 统一webhook ID
	NewWebhook       *CreateWebhookRequest   `json:"new_webhook,omitempty"`        // 新建webhook信息
	ProjectWebhooks  []ProjectWebhookMapping `json:"project_webhooks,omitempty"`   // 项目-webhook映射
}

// ProjectWebhookMapping 项目-webhook映射
type ProjectWebhookMapping struct {
	GitLabProjectID int  `json:"gitlab_project_id"`
	WebhookID       uint `json:"webhook_id"`
}

// BatchCreateProjectsResponse 批量创建项目的响应结构
type BatchCreateProjectsResponse struct {
	SuccessCount int                  `json:"success_count"`
	FailureCount int                  `json:"failure_count"`
	Results      []BatchProjectResult `json:"results"`
}

// BatchProjectResult 批量创建项目的单个结果
type BatchProjectResult struct {
	GitLabProjectID int    `json:"gitlab_project_id"`
	Name            string `json:"name"`
	Success         bool   `json:"success"`
	Error           string `json:"error,omitempty"`
	ProjectID       uint   `json:"project_id,omitempty"`
}

// SyncGitLabWebhookRequest 同步GitLab Webhook请求
type SyncGitLabWebhookRequest struct {
	ProjectID uint `json:"project_id" binding:"required"`
}

// SyncGitLabWebhookResponse 同步GitLab Webhook响应
type SyncGitLabWebhookResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	GitLabWebhookID *int   `json:"gitlab_webhook_id,omitempty"`
	WebhookURL      string `json:"webhook_url,omitempty"`
}

// GitLabWebhookStatusResponse GitLab Webhook状态响应
type GitLabWebhookStatusResponse struct {
	ProjectID       uint       `json:"project_id"`
	WebhookSynced   bool       `json:"webhook_synced"`
	GitLabWebhookID *int       `json:"gitlab_webhook_id,omitempty"`
	WebhookURL      string     `json:"webhook_url,omitempty"`
	LastSyncAt      *time.Time `json:"last_sync_at,omitempty"`
	CanManage       bool       `json:"can_manage"` // 是否有权限管理webhook
}
