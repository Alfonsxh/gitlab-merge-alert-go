package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID               uint           `json:"id" gorm:"primarykey"`
	ProjectID        uint           `json:"project_id" gorm:"not null"`
	MergeRequestID   int            `json:"merge_request_id" gorm:"not null"`
	Title            string         `json:"title"`
	SourceBranch     string         `json:"source_branch"`
	TargetBranch     string         `json:"target_branch"`
	AuthorEmail      string         `json:"author_email"`
	AssigneeEmails   string         `json:"assignee_emails"` // JSON array as string
	Status           string         `json:"status"`
	NotificationSent bool           `json:"notification_sent" gorm:"default:false"`
	ErrorMessage     string         `json:"error_message"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Project Project `json:"project" gorm:"foreignKey:ProjectID"`
}

type GitLabWebhookData struct {
	ObjectKind       string             `json:"object_kind"`
	User             GitLabUser         `json:"user"`
	Project          GitLabProject      `json:"project"`
	Repository       GitLabRepository   `json:"repository"`
	ObjectAttributes GitLabMergeRequest `json:"object_attributes"`
	Assignees        []GitLabUser       `json:"assignees"`
}

type GitLabUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type GitLabProject struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	WebURL    string `json:"web_url"`
	Namespace struct {
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"namespace"`
}

type GitLabRepository struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
}

type GitLabMergeRequest struct {
	ID           int    `json:"id"`
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	State        string `json:"state"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	URL          string `json:"url"`
}

type NotificationResponse struct {
	ID               uint      `json:"id"`
	ProjectID        uint      `json:"project_id"`
	ProjectName      string    `json:"project_name"`
	MergeRequestID   int       `json:"merge_request_id"`
	Title            string    `json:"title"`
	SourceBranch     string    `json:"source_branch"`
	TargetBranch     string    `json:"target_branch"`
	AuthorEmail      string    `json:"author_email"`
	AssigneeEmails   []string  `json:"assignee_emails"`
	Status           string    `json:"status"`
	NotificationSent bool      `json:"notification_sent"`
	ErrorMessage     string    `json:"error_message"`
	CreatedAt        time.Time `json:"created_at"`
}
