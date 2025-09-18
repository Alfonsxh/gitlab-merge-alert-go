package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type gitLabService struct {
	baseURL     string
	accessToken string
	client      *http.Client
}

func NewGitLabService(baseURL, accessToken string) GitLabService {
	return &gitLabService{
		baseURL:     baseURL,
		accessToken: accessToken,
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

type GitLabProjectInfo struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	Description       string `json:"description"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
}

type ParsedGitLabURL struct {
	BaseURL     string
	ProjectPath string
	IsGroup     bool // 是否为组URL
	IsValid     bool
	Error       string
}

// GitLabGroupInfo GitLab组信息
type GitLabGroupInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
	WebURL   string `json:"web_url"`
}

// GitLabWebhook GitLab项目Webhook结构
type GitLabWebhook struct {
	ID                       int    `json:"id"`
	URL                      string `json:"url"`
	MergeRequestsEvents      bool   `json:"merge_requests_events"`
	PushEvents               bool   `json:"push_events"`
	IssuesEvents             bool   `json:"issues_events"`
	ConfidentialIssuesEvents bool   `json:"confidential_issues_events"`
	TagPushEvents            bool   `json:"tag_push_events"`
	NoteEvents               bool   `json:"note_events"`
	PipelineEvents           bool   `json:"pipeline_events"`
	WikiPageEvents           bool   `json:"wiki_page_events"`
	DeploymentEvents         bool   `json:"deployment_events"`
	JobEvents                bool   `json:"job_events"`
	ReleasesEvents           bool   `json:"releases_events"`
	SubgroupEvents           bool   `json:"subgroup_events"`
	EnableSSLVerification    bool   `json:"enable_ssl_verification"`
	Token                    string `json:"token,omitempty"`
	CreatedAt                string `json:"created_at"`
	UpdatedAt                string `json:"updated_at"`
}

// CreateWebhookRequest 创建Webhook请求结构
type CreateWebhookRequest struct {
	URL                      string `json:"url"`
	MergeRequestsEvents      bool   `json:"merge_requests_events"`
	PushEvents               bool   `json:"push_events"`
	IssuesEvents             bool   `json:"issues_events"`
	ConfidentialIssuesEvents bool   `json:"confidential_issues_events"`
	TagPushEvents            bool   `json:"tag_push_events"`
	NoteEvents               bool   `json:"note_events"`
	PipelineEvents           bool   `json:"pipeline_events"`
	WikiPageEvents           bool   `json:"wiki_page_events"`
	DeploymentEvents         bool   `json:"deployment_events"`
	JobEvents                bool   `json:"job_events"`
	ReleasesEvents           bool   `json:"releases_events"`
	SubgroupEvents           bool   `json:"subgroup_events"`
	EnableSSLVerification    bool   `json:"enable_ssl_verification"`
	Token                    string `json:"token,omitempty"`
}

// ParseGitLabURL 解析GitLab项目URL，提取基础URL和项目路径
func (s *gitLabService) ParseGitLabURL(projectURL string) *ParsedGitLabURL {
	result := &ParsedGitLabURL{}

	// 清理URL，移除末尾的斜杠和可能的片段
	projectURL = strings.TrimSpace(projectURL)
	if projectURL == "" {
		result.Error = "URL不能为空"
		return result
	}

	// 解析URL
	parsedURL, err := url.Parse(projectURL)
	if err != nil {
		result.Error = "URL格式无效"
		return result
	}

	if parsedURL.Scheme == "" {
		result.Error = "URL必须包含协议(http或https)"
		return result
	}

	if parsedURL.Host == "" {
		result.Error = "URL必须包含主机名"
		return result
	}

	// 构建基础URL
	result.BaseURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	if parsedURL.Port() != "" && parsedURL.Port() != "80" && parsedURL.Port() != "443" {
		result.BaseURL = fmt.Sprintf("%s:%s", result.BaseURL, parsedURL.Port())
	}

	// 提取项目路径
	path := strings.TrimPrefix(parsedURL.Path, "/")
	path = strings.TrimSuffix(path, "/")

	// 移除GitLab特有的路径后缀
	gitlabSuffixes := []string{
		"/-/tree/",
		"/-/blob/",
		"/-/commits/",
		"/-/merge_requests",
		"/-/issues",
		"/-/wiki",
		"/-/settings",
	}

	for _, suffix := range gitlabSuffixes {
		if idx := strings.Index(path, suffix); idx != -1 {
			path = path[:idx]
			break
		}
	}

	// 验证项目路径格式
	if path == "" {
		result.Error = "无法从URL中提取项目路径"
		return result
	}

	// 判断是组还是项目
	// 组路径: group 或 group/subgroup
	// 项目路径: group/project 或 group/subgroup/project
	pathParts := strings.Split(path, "/")

	if len(pathParts) == 1 {
		// 单层路径，可能是组
		result.IsGroup = true
	} else if len(pathParts) >= 2 {
		// 多层路径，需要通过API判断是组还是项目
		// 先假设是项目，如果API调用失败再尝试作为组
		result.IsGroup = false
	}

	// GitLab路径格式验证（组和项目都适用）
	pathRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)*$`)
	if !pathRegex.MatchString(path) {
		result.Error = "路径格式无效，应为 group 或 group/project 格式"
		return result
	}

	result.ProjectPath = path
	result.IsValid = true
	return result
}

// GetProjectByURL 通过URL和token获取项目信息
func (s *gitLabService) GetProjectByURL(projectURL, accessToken string) (*GitLabProjectInfo, error) {
	// 解析URL
	parsed := s.ParseGitLabURL(projectURL)
	if !parsed.IsValid {
		return nil, fmt.Errorf("URL解析失败: %s", parsed.Error)
	}

	// 首先尝试作为项目获取
	projectInfo, err := s.GetProjectByPath(parsed.BaseURL, parsed.ProjectPath, accessToken)
	if err == nil {
		return projectInfo, nil
	}

	// 如果失败了，检查是否因为这是一个组URL
	if strings.Contains(err.Error(), "项目不存在") || strings.Contains(err.Error(), "404") {
		// 返回特殊错误标识这是组URL
		return nil, fmt.Errorf("GROUP_URL")
	}

	// 其他错误直接返回
	return nil, err
}

// GetProjectByPath 通过路径获取项目信息
func (s *gitLabService) GetProjectByPath(baseURL, projectPath, accessToken string) (*GitLabProjectInfo, error) {
	// URL编码项目路径
	encodedPath := url.QueryEscape(projectPath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s", baseURL, encodedPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置认证头
	if accessToken != "" {
		// GitLab支持多种token格式
		if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		} else {
			req.Header.Set("PRIVATE-TOKEN", accessToken)
		}
	}

	req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 处理不同的HTTP状态码
	switch resp.StatusCode {
	case http.StatusOK:
		// 成功，继续处理
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("访问令牌无效或已过期")
	case http.StatusForbidden:
		return nil, fmt.Errorf("没有权限访问此项目")
	case http.StatusNotFound:
		return nil, fmt.Errorf("项目不存在或无权限访问")
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("API请求频率过高，请稍后重试")
	default:
		return nil, fmt.Errorf("GitLab API返回错误状态: %d", resp.StatusCode)
	}

	var project GitLabProjectInfo
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &project, nil
}

// GetProject 通过项目ID获取项目信息（保持原有兼容性）
// 接受一个可选的accessToken参数，如果提供，则优先使用该token进行认证
func (s *gitLabService) GetProject(projectID int, accessToken ...string) (*GitLabProjectInfo, error) {
	token := s.accessToken
	if len(accessToken) > 0 && accessToken[0] != "" {
		token = accessToken[0]
	}
	return s.GetProjectByPath(s.baseURL, fmt.Sprintf("%d", projectID), token)
}

// TestConnection 测试GitLab连接和token有效性
func (s *gitLabService) TestConnection(baseURL, accessToken string) error {
	apiURL := fmt.Sprintf("%s/api/v4/user", baseURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	if accessToken != "" {
		if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		} else {
			req.Header.Set("PRIVATE-TOKEN", accessToken)
		}
	}

	req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("访问令牌无效")
	case http.StatusForbidden:
		return fmt.Errorf("访问令牌权限不足")
	default:
		return fmt.Errorf("连接测试失败，状态码: %d", resp.StatusCode)
	}
}

// GetGroupProjects 获取组下所有项目（包括子组项目）
func (s *gitLabService) GetGroupProjects(baseURL, groupPath, accessToken string) ([]*GitLabProjectInfo, error) {
	var allProjects []*GitLabProjectInfo

	// 获取组直接下的项目
	projects, err := s.getGroupDirectProjects(baseURL, groupPath, accessToken)
	if err != nil {
		return nil, err
	}
	allProjects = append(allProjects, projects...)

	// 获取子组
	subgroups, err := s.getSubgroups(baseURL, groupPath, accessToken)
	if err != nil {
		return nil, err
	}

	// 递归获取子组的项目
	for _, subgroup := range subgroups {
		subgroupProjects, err := s.GetGroupProjects(baseURL, subgroup.FullPath, accessToken)
		if err != nil {
			// 记录错误但继续处理其他子组
			continue
		}
		allProjects = append(allProjects, subgroupProjects...)
	}

	return allProjects, nil
}

// getGroupDirectProjects 获取组直接下的项目
func (s *gitLabService) getGroupDirectProjects(baseURL, groupPath, accessToken string) ([]*GitLabProjectInfo, error) {
	encodedPath := url.QueryEscape(groupPath)
	apiURL := fmt.Sprintf("%s/api/v4/groups/%s/projects", baseURL, encodedPath)

	var allProjects []*GitLabProjectInfo
	page := 1
	perPage := 100

	for {
		// 添加分页参数
		paginatedURL := fmt.Sprintf("%s?page=%d&per_page=%d&simple=false", apiURL, page, perPage)

		req, err := http.NewRequest("GET", paginatedURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}

		// 设置认证头
		if accessToken != "" {
			if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			} else {
				req.Header.Set("PRIVATE-TOKEN", accessToken)
			}
		}

		req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		// 处理HTTP状态码
		switch resp.StatusCode {
		case http.StatusOK:
			// 成功，继续处理
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("访问令牌无效或已过期")
		case http.StatusForbidden:
			return nil, fmt.Errorf("没有权限访问此组")
		case http.StatusNotFound:
			return nil, fmt.Errorf("组不存在或无权限访问")
		default:
			return nil, fmt.Errorf("GitLab API返回错误状态: %d", resp.StatusCode)
		}

		var projects []GitLabProjectInfo
		if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		// 转换为指针数组
		for i := range projects {
			allProjects = append(allProjects, &projects[i])
		}

		// 检查是否还有更多页面
		if len(projects) < perPage {
			break
		}
		page++
	}

	return allProjects, nil
}

// getSubgroups 获取子组
func (s *gitLabService) getSubgroups(baseURL, groupPath, accessToken string) ([]*GitLabGroupInfo, error) {
	encodedPath := url.QueryEscape(groupPath)
	apiURL := fmt.Sprintf("%s/api/v4/groups/%s/subgroups", baseURL, encodedPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置认证头
	if accessToken != "" {
		if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		} else {
			req.Header.Set("PRIVATE-TOKEN", accessToken)
		}
	}

	req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []*GitLabGroupInfo{}, nil // 没有子组或无权限访问，返回空数组
	}

	var groups []GitLabGroupInfo
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 转换为指针数组
	var result []*GitLabGroupInfo
	for i := range groups {
		result = append(result, &groups[i])
	}

	return result, nil
}

// GetGroupByPath 获取组信息
func (s *gitLabService) GetGroupByPath(baseURL, groupPath, accessToken string) (*GitLabGroupInfo, error) {
	encodedPath := url.QueryEscape(groupPath)
	apiURL := fmt.Sprintf("%s/api/v4/groups/%s", baseURL, encodedPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置认证头
	if accessToken != "" {
		if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		} else {
			req.Header.Set("PRIVATE-TOKEN", accessToken)
		}
	}

	req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 处理不同的HTTP状态码
	switch resp.StatusCode {
	case http.StatusOK:
		// 成功，继续处理
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("访问令牌无效或已过期")
	case http.StatusForbidden:
		return nil, fmt.Errorf("没有权限访问此组")
	case http.StatusNotFound:
		return nil, fmt.Errorf("组不存在或无权限访问")
	default:
		return nil, fmt.Errorf("GitLab API返回错误状态: %d", resp.StatusCode)
	}

	var group GitLabGroupInfo
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &group, nil
}

// ValidateProjectURL 验证项目URL并返回项目ID（保持向后兼容性）
func (s *gitLabService) ValidateProjectURL(projectURL string) (int, error) {
	parsed := s.ParseGitLabURL(projectURL)
	if !parsed.IsValid {
		return 0, fmt.Errorf(parsed.Error)
	}

	project, err := s.GetProjectByPath(parsed.BaseURL, parsed.ProjectPath, s.accessToken)
	if err != nil {
		return 0, err
	}

	return project.ID, nil
}

// CreateProjectWebhook 在GitLab项目中创建webhook
func (s *gitLabService) CreateProjectWebhook(baseURL string, projectID int, webhookURL, accessToken string) (*GitLabWebhook, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/hooks", baseURL, projectID)

	webhookRequest := CreateWebhookRequest{
		URL:                   webhookURL,
		MergeRequestsEvents:   true, // 只关注合并请求事件
		PushEvents:            false,
		IssuesEvents:          false,
		EnableSSLVerification: false, // 对于测试环境可以关闭SSL验证
	}

	requestBody, err := json.Marshal(webhookRequest)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		} else {
			req.Header.Set("PRIVATE-TOKEN", accessToken)
		}
	}
	req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 处理不同的HTTP状态码
	switch resp.StatusCode {
	case http.StatusCreated:
		// 成功创建，继续处理
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("访问令牌无效或已过期")
	case http.StatusForbidden:
		return nil, fmt.Errorf("没有权限在此项目中创建webhook")
	case http.StatusNotFound:
		return nil, fmt.Errorf("项目不存在或无权限访问")
	case http.StatusUnprocessableEntity:
		return nil, fmt.Errorf("Webhook URL已存在或格式无效")
	default:
		return nil, fmt.Errorf("GitLab API返回错误状态: %d", resp.StatusCode)
	}

	var webhook GitLabWebhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &webhook, nil
}

// ListProjectWebhooks 获取项目的所有webhooks
func (s *gitLabService) ListProjectWebhooks(baseURL string, projectID int, accessToken string) ([]*GitLabWebhook, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/hooks", baseURL, projectID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置认证头
	if accessToken != "" {
		if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		} else {
			req.Header.Set("PRIVATE-TOKEN", accessToken)
		}
	}
	req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 处理不同的HTTP状态码
	switch resp.StatusCode {
	case http.StatusOK:
		// 成功，继续处理
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("访问令牌无效或已过期")
	case http.StatusForbidden:
		return nil, fmt.Errorf("没有权限访问此项目的webhooks")
	case http.StatusNotFound:
		return nil, fmt.Errorf("项目不存在或无权限访问")
	default:
		return nil, fmt.Errorf("GitLab API返回错误状态: %d", resp.StatusCode)
	}

	var webhooks []GitLabWebhook
	if err := json.NewDecoder(resp.Body).Decode(&webhooks); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 转换为指针数组
	var result []*GitLabWebhook
	for i := range webhooks {
		result = append(result, &webhooks[i])
	}

	return result, nil
}

// DeleteProjectWebhook 删除项目webhook
func (s *gitLabService) DeleteProjectWebhook(baseURL string, projectID, webhookID int, accessToken string) error {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/hooks/%d", baseURL, projectID, webhookID)

	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置认证头
	if accessToken != "" {
		if strings.HasPrefix(accessToken, "glpat-") || strings.HasPrefix(accessToken, "glcbt-") {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		} else {
			req.Header.Set("PRIVATE-TOKEN", accessToken)
		}
	}
	req.Header.Set("User-Agent", "GitLab-Merge-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 处理不同的HTTP状态码
	switch resp.StatusCode {
	case http.StatusNoContent:
		// 成功删除
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("访问令牌无效或已过期")
	case http.StatusForbidden:
		return fmt.Errorf("没有权限删除此项目的webhook")
	case http.StatusNotFound:
		return fmt.Errorf("项目或webhook不存在")
	default:
		return fmt.Errorf("GitLab API返回错误状态: %d", resp.StatusCode)
	}
}

// BuildWebhookURL 构建本服务的webhook接收URL
func (s *gitLabService) BuildWebhookURL(publicBaseURL string) string {
	return fmt.Sprintf("%s/api/v1/webhook/gitlab", strings.TrimSuffix(publicBaseURL, "/"))
}

// FindWebhookByURL 根据URL查找项目中的webhook
func (s *gitLabService) FindWebhookByURL(baseURL string, projectID int, webhookURL, accessToken string) (*GitLabWebhook, error) {
	webhooks, err := s.ListProjectWebhooks(baseURL, projectID, accessToken)
	if err != nil {
		return nil, err
	}

	for _, webhook := range webhooks {
		if webhook.URL == webhookURL {
			return webhook, nil
		}
	}

	return nil, nil // 未找到
}

// FindAllWebhooksByURL 根据URL查找项目中所有匹配的webhook
func (s *gitLabService) FindAllWebhooksByURL(baseURL string, projectID int, webhookURL, accessToken string) ([]*GitLabWebhook, error) {
	webhooks, err := s.ListProjectWebhooks(baseURL, projectID, accessToken)
	if err != nil {
		return nil, err
	}

	var matchingWebhooks []*GitLabWebhook
	for _, webhook := range webhooks {
		if webhook.URL == webhookURL {
			matchingWebhooks = append(matchingWebhooks, webhook)
		}
	}

	return matchingWebhooks, nil
}

// DeleteAllWebhooksByURL 删除项目中所有匹配URL的webhook
func (s *gitLabService) DeleteAllWebhooksByURL(baseURL string, projectID int, webhookURL, accessToken string) (int, error) {
	// 首先找到所有匹配的webhook
	matchingWebhooks, err := s.FindAllWebhooksByURL(baseURL, projectID, webhookURL, accessToken)
	if err != nil {
		return 0, fmt.Errorf("查找匹配的webhook失败: %v", err)
	}

	if len(matchingWebhooks) == 0 {
		return 0, nil // 没有找到匹配的webhook
	}

	var deletedCount int
	var errors []string

	// 删除每个匹配的webhook
	for _, webhook := range matchingWebhooks {
		err := s.DeleteProjectWebhook(baseURL, projectID, webhook.ID, accessToken)
		if err != nil {
			errors = append(errors, fmt.Sprintf("删除webhook ID %d 失败: %v", webhook.ID, err))
		} else {
			deletedCount++
		}
	}

	// 如果有错误，返回部分成功的结果
	if len(errors) > 0 {
		return deletedCount, fmt.Errorf("部分删除失败 (%d/%d 成功): %s", deletedCount, len(matchingWebhooks), strings.Join(errors, "; "))
	}

	return deletedCount, nil
}
