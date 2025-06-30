package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"column:id;primarykey"`
	Email     string    `json:"email" gorm:"column:email;uniqueIndex;not null;default:''"`
	Phone     string    `json:"phone" gorm:"column:phone;not null;default:''"`
	Name      string    `json:"name" gorm:"column:name"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
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
