package auth

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordManager struct {
	cost int
}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		cost: bcrypt.DefaultCost,
	}
}

func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (pm *PasswordManager) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (pm *PasswordManager) IsValidPassword(password string) bool {
	// 密码至少6个字符
	return len(password) >= 6
}
