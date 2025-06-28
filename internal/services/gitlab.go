package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GitLabService struct {
	baseURL     string
	accessToken string
	client      *http.Client
}

func NewGitLabService(baseURL, accessToken string) *GitLabService {
	return &GitLabService{
		baseURL:     baseURL,
		accessToken: accessToken,
		client:      &http.Client{},
	}
}

type GitLabProjectInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL      string `json:"web_url"`
	Description string `json:"description"`
}

func (s *GitLabService) GetProject(projectID int) (*GitLabProjectInfo, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d", s.baseURL, projectID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if s.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.accessToken)
	}
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
	}
	
	var project GitLabProjectInfo
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}
	
	return &project, nil
}

func (s *GitLabService) ValidateProjectURL(projectURL string) (int, error) {
	// 简单的URL解析来提取项目ID
	// 实际实现中可能需要更复杂的逻辑
	return 0, fmt.Errorf("not implemented")
}