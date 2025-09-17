package models

import (
	"time"
)

type Webhook struct {
	ID          uint      `json:"id" gorm:"column:id;primarykey"`
	Name        string    `json:"name" gorm:"column:name;not null;default:''"`
	URL         string    `json:"url" gorm:"column:url;not null;default:''"`
	Description string    `json:"description" gorm:"column:description"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedBy   *uint     `json:"created_by,omitempty" gorm:"column:created_by;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 关联关系
	Projects []Project `json:"projects,omitempty" gorm:"many2many:project_webhooks;"`
}

type ProjectWebhook struct {
	ID        uint      `json:"id" gorm:"column:id;primarykey"`
	ProjectID uint      `json:"project_id" gorm:"column:project_id;not null;default:0"`
	WebhookID uint      `json:"webhook_id" gorm:"column:webhook_id;not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`

	// 关联关系
	Project Project `json:"project" gorm:"foreignKey:ProjectID"`
	Webhook Webhook `json:"webhook" gorm:"foreignKey:WebhookID"`
}

type CreateWebhookRequest struct {
	Name        string `json:"name" binding:"required"`
	URL         string `json:"url" binding:"required,url"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateWebhookRequest struct {
	Name        string `json:"name"`
	URL         string `json:"url" binding:"omitempty,url"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type WebhookResponse struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Description string            `json:"description"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Projects    []ProjectResponse `json:"projects,omitempty"`
}

type LinkProjectWebhookRequest struct {
	ProjectID uint `json:"project_id" binding:"required"`
	WebhookID uint `json:"webhook_id" binding:"required"`
}
