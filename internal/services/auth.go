package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Alfonsxh/gitlab-merge-alert-go/internal/models"
	"github.com/Alfonsxh/gitlab-merge-alert-go/pkg/auth"
	"github.com/Alfonsxh/gitlab-merge-alert-go/pkg/logger"
	"github.com/Alfonsxh/gitlab-merge-alert-go/pkg/security"

	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials      = errors.New("invalid username or password")
	ErrAccountNotActive        = errors.New("account is not active")
	ErrAccountNotFound         = errors.New("account not found")
	ErrUsernameExists          = errors.New("username already exists")
	ErrEmailExists             = errors.New("email already exists")
	ErrAdminLocked             = errors.New("admin account is reserved")
	ErrPasswordResetRequired   = errors.New("password reset required")
	ErrAdminAlreadyInitialized = errors.New("admin account already initialized")
	ErrInvalidSetupToken       = errors.New("invalid setup token")
	ErrWeakPassword            = errors.New("password does not meet complexity requirements")
)

type AuthService interface {
	Login(username, password string) (*models.LoginResponse, error)
	RefreshToken(oldToken string) (*models.LoginResponse, error)
	GetAccountByID(id uint) (*models.Account, error)
	ChangePassword(accountID uint, oldPassword, newPassword string) error
	RegisterUser(username, email, password, gitlabToken string) (*models.LoginResponse, error)
	InitializeAdminAccount() error
	IsAdminSetupRequired() (bool, error)
	CompleteAdminSetup(token, email, password string) error
}

type authService struct {
	db              *gorm.DB
	jwtManager      *auth.JWTManager
	passwordManager *auth.PasswordManager
	encryptionKey   string
}

func NewAuthService(db *gorm.DB, jwtSecret string, jwtDuration time.Duration, encryptionKey string) AuthService {
	return &authService{
		db:              db,
		jwtManager:      auth.NewJWTManager(jwtSecret, jwtDuration),
		passwordManager: auth.NewPasswordManager(),
		encryptionKey:   encryptionKey,
	}
}

func (s *authService) Login(username, password string) (*models.LoginResponse, error) {
	var account models.Account

	if err := s.db.Where("username = ?", username).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !account.IsActive {
		return nil, ErrAccountNotActive
	}

	if account.ForcePasswordReset {
		return nil, ErrPasswordResetRequired
	}

	if err := s.passwordManager.VerifyPassword(account.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := s.jwtManager.Generate(account.ID, account.Username, account.Role)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	account.LastLoginAt = &now
	if err := s.db.Save(&account).Error; err != nil {
		logger.GetLogger().Warnf("Failed to update last login time: %v", err)
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      account.ToResponse(),
	}, nil
}

func (s *authService) RefreshToken(oldToken string) (*models.LoginResponse, error) {
	claims, err := s.jwtManager.Verify(oldToken)
	if err != nil && err != auth.ErrExpiredToken {
		return nil, err
	}

	account, err := s.GetAccountByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	if !account.IsActive {
		return nil, ErrAccountNotActive
	}

	if account.ForcePasswordReset {
		return nil, ErrPasswordResetRequired
	}

	token, expiresAt, err := s.jwtManager.Refresh(oldToken)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      account.ToResponse(),
	}, nil
}

func (s *authService) GetAccountByID(id uint) (*models.Account, error) {
	var account models.Account
	if err := s.db.First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (s *authService) ChangePassword(accountID uint, oldPassword, newPassword string) error {
	account, err := s.GetAccountByID(accountID)
	if err != nil {
		return err
	}

	if err := s.passwordManager.VerifyPassword(account.PasswordHash, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	newHash, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return err
	}

	account.PasswordHash = newHash
	account.PasswordInitializedAt = ptrTime(time.Now())

	return s.db.Save(account).Error
}

func (s *authService) RegisterUser(username, email, password, gitlabToken string) (*models.LoginResponse, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	gitlabToken = strings.TrimSpace(gitlabToken)
	if strings.EqualFold(username, "admin") {
		return nil, ErrAdminLocked
	}

	if gitlabToken == "" {
		return nil, fmt.Errorf("gitlab personal access token is required")
	}

	var count int64
	if err := s.db.Model(&models.Account{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUsernameExists
	}

	if err := s.db.Model(&models.Account{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrEmailExists
	}

	passwordHash, err := s.passwordManager.HashPassword(password)
	if err != nil {
		return nil, err
	}

	encryptedToken, err := security.Encrypt(s.encryptionKey, gitlabToken)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	account := &models.Account{
		Username:              username,
		PasswordHash:          passwordHash,
		Email:                 email,
		Role:                  models.RoleUser,
		IsActive:              true,
		ForcePasswordReset:    false,
		PasswordInitializedAt: &now,
		GitLabAccessToken:     encryptedToken,
	}

	if err := s.db.Create(account).Error; err != nil {
		return nil, err
	}

	token, expiresAt, err := s.jwtManager.Generate(account.ID, account.Username, account.Role)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      account.ToResponse(),
	}, nil
}

func (s *authService) InitializeAdminAccount() error {
	var admin models.Account
	err := s.db.Where("role = ?", models.RoleAdmin).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.createBootstrapAdmin()
	}
	if err != nil {
		return err
	}

	if admin.ForcePasswordReset {
		token, rotateErr := s.rotateAdminSetupToken(&admin)
		if rotateErr != nil {
			return rotateErr
		}
		logger.GetLogger().Warnf("Admin account requires initial setup. Use the setup token to finish initialization.")
		logger.GetLogger().Warnf("Admin setup token: %s", token)
		return nil
	}

	logger.GetLogger().Info("Admin account already initialized")
	return nil
}

func (s *authService) IsAdminSetupRequired() (bool, error) {
	var count int64
	if err := s.db.Model(&models.Account{}).
		Where("role = ? AND force_password_reset = ?", models.RoleAdmin, true).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *authService) CompleteAdminSetup(token, email, password string) error {
	token = strings.TrimSpace(token)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if token == "" {
		return ErrInvalidSetupToken
	}

	if !s.passwordManager.IsValidPassword(password) {
		return ErrWeakPassword
	}

	var admin models.Account
	if err := s.db.Where("role = ?", models.RoleAdmin).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAccountNotFound
		}
		return err
	}

	if !admin.ForcePasswordReset {
		return ErrAdminAlreadyInitialized
	}

	if admin.AdminSetupTokenHash == "" {
		return ErrInvalidSetupToken
	}

	if err := s.passwordManager.VerifyPassword(admin.AdminSetupTokenHash, token); err != nil {
		return ErrInvalidSetupToken
	}

	newHash, err := s.passwordManager.HashPassword(password)
	if err != nil {
		return err
	}

	now := time.Now()
	admin.PasswordHash = newHash
	admin.Email = email
	admin.ForcePasswordReset = false
	admin.AdminSetupTokenHash = ""
	admin.AdminSetupTokenGeneratedAt = nil
	admin.PasswordInitializedAt = &now
	admin.IsActive = true

	if err := s.db.Save(&admin).Error; err != nil {
		return err
	}

	logger.GetLogger().Info("Admin account successfully initialized via setup flow")

	return nil
}

func (s *authService) createBootstrapAdmin() error {
	placeholder, err := generateRandomSecret(32)
	if err != nil {
		return err
	}

	passwordHash, err := s.passwordManager.HashPassword(placeholder)
	if err != nil {
		return err
	}

	token, tokenHash, err := s.generateSetupToken()
	if err != nil {
		return err
	}

	now := time.Now()
	admin := &models.Account{
		Username:                   "admin",
		PasswordHash:               passwordHash,
		Email:                      "admin@example.com",
		Role:                       models.RoleAdmin,
		IsActive:                   true,
		ForcePasswordReset:         true,
		AdminSetupTokenHash:        tokenHash,
		AdminSetupTokenGeneratedAt: &now,
	}

	if err := s.db.Create(admin).Error; err != nil {
		return err
	}

	logger.GetLogger().Warnf("Bootstrap admin account created. Complete setup using the token shown below.")
	logger.GetLogger().Warnf("Admin setup token: %s", token)

	return nil
}

func (s *authService) rotateAdminSetupToken(admin *models.Account) (string, error) {
	token, tokenHash, err := s.generateSetupToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	update := map[string]interface{}{
		"admin_setup_token_hash":         tokenHash,
		"admin_setup_token_generated_at": now,
		"force_password_reset":           true,
	}

	if err := s.db.Model(admin).Updates(update).Error; err != nil {
		return "", err
	}

	admin.AdminSetupTokenHash = tokenHash
	admin.AdminSetupTokenGeneratedAt = &now
	admin.ForcePasswordReset = true

	return token, nil
}

func (s *authService) generateSetupToken() (string, string, error) {
	token, err := generateRandomSecret(32)
	if err != nil {
		return "", "", err
	}

	hash, err := s.passwordManager.HashPassword(token)
	if err != nil {
		return "", "", err
	}

	return token, hash, nil
}

func generateRandomSecret(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
