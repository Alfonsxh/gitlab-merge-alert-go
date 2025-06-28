package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Phone     string         `json:"phone" gorm:"not null"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required"`
	Name  string `json:"name"`
}

type UpdateUserRequest struct {
	Email string `json:"email" binding:"email"`
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}