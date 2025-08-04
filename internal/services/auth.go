package services

import (
	"errors"
	"time"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/auth"
	"gitlab-merge-alert-go/pkg/logger"

	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountNotActive   = errors.New("account is not active")
	ErrAccountNotFound    = errors.New("account not found")
)

type AuthService interface {
	Login(username, password string) (*models.LoginResponse, error)
	RefreshToken(oldToken string) (*models.LoginResponse, error)
	GetAccountByID(id uint) (*models.Account, error)
	ChangePassword(accountID uint, oldPassword, newPassword string) error
	InitializeAdminAccount() error
}

type authService struct {
	db              *gorm.DB
	jwtManager      *auth.JWTManager
	passwordManager *auth.PasswordManager
}

func NewAuthService(db *gorm.DB, jwtSecret string, jwtDuration time.Duration) AuthService {
	return &authService{
		db:              db,
		jwtManager:      auth.NewJWTManager(jwtSecret, jwtDuration),
		passwordManager: auth.NewPasswordManager(),
	}
}

func (s *authService) Login(username, password string) (*models.LoginResponse, error) {
	var account models.Account
	
	// 查找账户
	if err := s.db.Where("username = ?", username).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 检查账户是否激活
	if !account.IsActive {
		return nil, ErrAccountNotActive
	}

	// 验证密码
	if err := s.passwordManager.VerifyPassword(account.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 生成 JWT token
	token, expiresAt, err := s.jwtManager.Generate(account.ID, account.Username, account.Role)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	account.LastLoginAt = &now
	if err := s.db.Save(&account).Error; err != nil {
		logger.GetLogger().Warnf("Failed to update last login time: %v", err)
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      &account,
	}, nil
}

func (s *authService) RefreshToken(oldToken string) (*models.LoginResponse, error) {
	// 验证旧 token 并获取信息
	claims, err := s.jwtManager.Verify(oldToken)
	if err != nil && err != auth.ErrExpiredToken {
		return nil, err
	}

	// 获取账户信息
	account, err := s.GetAccountByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// 检查账户是否激活
	if !account.IsActive {
		return nil, ErrAccountNotActive
	}

	// 刷新 token
	token, expiresAt, err := s.jwtManager.Refresh(oldToken)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      account,
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
	// 获取账户
	account, err := s.GetAccountByID(accountID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := s.passwordManager.VerifyPassword(account.PasswordHash, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	// 生成新密码哈希
	newHash, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	account.PasswordHash = newHash
	return s.db.Save(account).Error
}

func (s *authService) InitializeAdminAccount() error {
	// 检查是否已存在管理员账户
	var count int64
	if err := s.db.Model(&models.Account{}).Where("role = ?", models.RoleAdmin).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.GetLogger().Info("Admin account already exists")
		return nil
	}

	// 创建默认管理员账户
	passwordHash, err := s.passwordManager.HashPassword("admin123456")
	if err != nil {
		return err
	}

	admin := &models.Account{
		Username:     "admin",
		PasswordHash: passwordHash,
		Email:        "admin@example.com",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}

	if err := s.db.Create(admin).Error; err != nil {
		return err
	}

	logger.GetLogger().Info("Default admin account created (username: admin, password: admin123456)")
	logger.GetLogger().Warn("Please change the default admin password immediately!")
	
	return nil
}