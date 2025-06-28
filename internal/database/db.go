package database

import (
	"os"
	"path/filepath"

	"gitlab-merge-alert-go/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(databasePath string) (*gorm.DB, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(databasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 连接SQLite数据库
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	// 自动迁移数据库表结构
	return db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Webhook{},
		&models.ProjectWebhook{},
		&models.Notification{},
	)
}