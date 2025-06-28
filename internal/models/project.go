package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID               uint           `json:"id" gorm:"primarykey"`
	GitLabProjectID  int            `json:"gitlab_project_id" gorm:"uniqueIndex;not null"`
	Name             string         `json:"name" gorm:"not null"`
	URL              string         `json:"url" gorm:"not null"`
	Description      string         `json:"description"`
	AccessToken      string         `json:"-"` // 不在JSON中显示敏感信息
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联关系
	Webhooks         []Webhook      `json:"webhooks,omitempty" gorm:"many2many:project_webhooks;"`
}

type CreateProjectRequest struct {
	GitLabProjectID int    `json:"gitlab_project_id" binding:"required"`
	Name            string `json:"name" binding:"required"`
	URL             string `json:"url" binding:"required,url"`
	Description     string `json:"description"`
	AccessToken     string `json:"access_token"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	URL         string `json:"url" binding:"omitempty,url"`
	Description string `json:"description"`
	AccessToken string `json:"access_token"`
}

type ProjectResponse struct {
	ID              uint      `json:"id"`
	GitLabProjectID int       `json:"gitlab_project_id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Webhooks        []WebhookResponse `json:"webhooks,omitempty"`
}