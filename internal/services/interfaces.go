package services

import (
	"github.com/Alfonsxh/gitlab-merge-alert-go/internal/models"
)

// UserService 用户服务接口
type UserService interface {
	CreateUser(req *models.CreateUserRequest) (*models.UserResponse, error)
	GetUserByID(id uint) (*models.UserResponse, error)
	GetAllUsers() ([]models.UserResponse, error)
	UpdateUser(id uint, req *models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(id uint) error
	FindUsersByEmailsOrUsernames(emails, usernames []string) ([]models.User, error)
}

// ProjectService 项目服务接口
type ProjectService interface {
	CreateProject(req *models.CreateProjectRequest) (*models.ProjectResponse, error)
	GetProjectByID(id uint) (*models.ProjectResponse, error)
	GetAllProjects() ([]models.ProjectResponse, error)
	UpdateProject(id uint, req *models.UpdateProjectRequest) (*models.ProjectResponse, error)
	DeleteProject(id uint) error
	ParseProjectURL(req *models.ParseProjectURLRequest) (*models.ParseProjectURLResponse, error)
	TestGitLabConnection(req *models.GitLabConnectionTestRequest) (*models.GitLabConnectionTestResponse, error)
	ScanGroupProjects(req *models.ScanGroupProjectsRequest) (*models.ScanGroupProjectsResponse, error)
	BatchCreateProjects(req *models.BatchCreateProjectsRequest) (*models.BatchCreateProjectsResponse, error)
	SyncGitLabWebhook(projectID uint) (*models.SyncGitLabWebhookResponse, error)
	DeleteGitLabWebhook(projectID uint) error
	GetGitLabWebhookStatus(projectID uint) (*models.GitLabWebhookStatusResponse, error)
}

// WebhookService Webhook服务接口
type WebhookService interface {
	CreateWebhook(req *models.CreateWebhookRequest) (*models.WebhookResponse, error)
	GetWebhookByID(id uint) (*models.WebhookResponse, error)
	GetAllWebhooks() ([]models.WebhookResponse, error)
	UpdateWebhook(id uint, req *models.UpdateWebhookRequest) (*models.WebhookResponse, error)
	DeleteWebhook(id uint) error
	LinkProjectWebhook(projectID, webhookID uint) error
	UnlinkProjectWebhook(projectID, webhookID uint) error
}

// NotificationService 通知服务接口
type NotificationService interface {
	ProcessMergeRequest(webhookData *models.GitLabWebhookData) error
	GetAllNotifications() ([]models.NotificationResponse, error)
	GetNotificationsByProjectID(projectID uint) ([]models.NotificationResponse, error)
	GetRecentNotifications(limit int) ([]models.NotificationResponse, error)
	GetNotificationStats() (map[string]interface{}, error)
}

// GitLabService GitLab服务接口
type GitLabService interface {
	ParseGitLabURL(projectURL string) *ParsedGitLabURL
	GetProjectByURL(projectURL, accessToken string) (*GitLabProjectInfo, error)
	GetProjectByPath(baseURL, projectPath, accessToken string) (*GitLabProjectInfo, error)
	GetProject(projectID int, accessToken ...string) (*GitLabProjectInfo, error)
	TestConnection(baseURL, accessToken string) error
	GetGroupProjects(baseURL, groupPath, accessToken string) ([]*GitLabProjectInfo, error)
	GetGroupByPath(baseURL, groupPath, accessToken string) (*GitLabGroupInfo, error)
	ValidateProjectURL(projectURL string) (int, error)
	CreateProjectWebhook(baseURL string, projectID int, webhookURL, accessToken string) (*GitLabWebhook, error)
	ListProjectWebhooks(baseURL string, projectID int, accessToken string) ([]*GitLabWebhook, error)
	DeleteProjectWebhook(baseURL string, projectID, webhookID int, accessToken string) error
	FindWebhookByURL(baseURL string, projectID int, webhookURL, accessToken string) (*GitLabWebhook, error)
	FindAllWebhooksByURL(baseURL string, projectID int, webhookURL, accessToken string) ([]*GitLabWebhook, error)
	DeleteAllWebhooksByURL(baseURL string, projectID int, webhookURL, accessToken string) (int, error)
	BuildWebhookURL(publicBaseURL string) string
}

// WeChatService 微信服务接口
type WeChatService interface {
	SendMessage(webhookURL, content string, mentionedMobiles []string) error
	FormatMergeRequestMessage(projectName, sourceBranch, targetBranch, mergeFrom, mergeTitle, clickURL string, mergeToList []string, mentionedMobiles []string) string
}
