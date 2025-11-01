package models

import (
	"time"
)

type ResourceType string

const (
	ResourceTypeProject ResourceType = "project"
	ResourceTypeWebhook ResourceType = "webhook"
	ResourceTypeUser    ResourceType = "user"
)

type ResourceManager struct {
	ID           uint         `json:"id" gorm:"column:id;primarykey"`
	ResourceID   uint         `json:"resource_id" gorm:"column:resource_id;not null;index:idx_resource"`
	ResourceType ResourceType `json:"resource_type" gorm:"column:resource_type;type:varchar(20);not null;index:idx_resource"`
	ManagerID    uint         `json:"manager_id" gorm:"column:manager_id;not null;index"`
	CreatedBy    uint         `json:"created_by" gorm:"column:created_by;not null"`
	CreatedAt    time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"column:updated_at"`

	Manager Account `json:"manager,omitempty" gorm:"foreignKey:ManagerID"`
	Creator Account `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

type AssignManagerRequest struct {
	ResourceID   uint         `json:"resource_id" binding:"required"`
	ResourceType ResourceType `json:"resource_type" binding:"required,oneof=project webhook user"`
	ManagerID    uint         `json:"manager_id" binding:"required"`
}

type RemoveManagerRequest struct {
	ResourceID   uint         `json:"resource_id" binding:"required"`
	ResourceType ResourceType `json:"resource_type" binding:"required,oneof=project webhook user"`
	ManagerID    uint         `json:"manager_id" binding:"required"`
}

type ResourceManagerResponse struct {
	ID           uint             `json:"id"`
	ResourceID   uint             `json:"resource_id"`
	ResourceType ResourceType     `json:"resource_type"`
	ManagerID    uint             `json:"manager_id"`
	Manager      *AccountResponse `json:"manager,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}
