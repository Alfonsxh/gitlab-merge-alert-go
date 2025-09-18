package models

import (
	"strings"
	"time"
)

type Account struct {
	ID                uint       `json:"id" gorm:"column:id;primarykey"`
	Username          string     `json:"username" gorm:"column:username;uniqueIndex;not null"`
	PasswordHash      string     `json:"-" gorm:"column:password_hash;not null"`
	Email             string     `json:"email" gorm:"column:email;uniqueIndex;not null"`
	Role              string     `json:"role" gorm:"column:role;default:'user'"`
	Avatar            string     `json:"avatar" gorm:"column:avatar;type:text"`
	GitLabAccessToken string     `json:"-" gorm:"column:gitlab_access_token"`
	IsActive          bool       `json:"is_active" gorm:"column:is_active;default:true"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty" gorm:"column:last_login_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string           `json:"token"`
	ExpiresAt int64            `json:"expires_at"`
	User      *AccountResponse `json:"user"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

type CreateAccountRequest struct {
	Username                  string `json:"username" binding:"required,min=3,max=50"`
	Password                  string `json:"password" binding:"required,min=6"`
	Email                     string `json:"email" binding:"required,email"`
	Role                      string `json:"role" binding:"omitempty,oneof=admin user"`
	GitLabPersonalAccessToken string `json:"gitlab_personal_access_token"`
}

type UpdateAccountRequest struct {
	Email                     string  `json:"email" binding:"omitempty,email"`
	Role                      string  `json:"role" binding:"omitempty,oneof=admin user"`
	IsActive                  *bool   `json:"is_active"`
	GitLabPersonalAccessToken *string `json:"gitlab_personal_access_token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type AccountResponse struct {
	ID                           uint       `json:"id"`
	Username                     string     `json:"username"`
	Email                        string     `json:"email"`
	Role                         string     `json:"role"`
	Avatar                       string     `json:"avatar"`
	IsActive                     bool       `json:"is_active"`
	LastLoginAt                  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt                    time.Time  `json:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at"`
	HasGitLabPersonalAccessToken bool       `json:"has_gitlab_personal_access_token"`
}

type UpdateProfileRequest struct {
	Email                     string  `json:"email" binding:"omitempty,email"`
	Avatar                    string  `json:"avatar"`
	GitLabPersonalAccessToken *string `json:"gitlab_personal_access_token"`
}

func (a *Account) IsAdmin() bool {
	return a.Role == RoleAdmin
}

func (a *Account) ToResponse() *AccountResponse {
	return &AccountResponse{
		ID:                           a.ID,
		Username:                     a.Username,
		Email:                        a.Email,
		Role:                         a.Role,
		Avatar:                       a.Avatar,
		IsActive:                     a.IsActive,
		LastLoginAt:                  a.LastLoginAt,
		CreatedAt:                    a.CreatedAt,
		UpdatedAt:                    a.UpdatedAt,
		HasGitLabPersonalAccessToken: strings.TrimSpace(a.GitLabAccessToken) != "",
	}
}
