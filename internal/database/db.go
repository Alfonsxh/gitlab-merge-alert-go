package database

import (
	"os"
	"path/filepath"

	"gitlab-merge-alert-go/internal/migrations"

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
	// 使用迁移系统而不是简单的AutoMigrate
	return RunMigrations(db)
}

// RunMigrations 运行数据库迁移
func RunMigrations(db *gorm.DB) error {
	migrator := migrations.SetupMigrator(db)
	return migrator.Up()
}

// MigrationsStatus 查看迁移状态
func MigrationsStatus(db *gorm.DB) error {
	migrator := migrations.SetupMigrator(db)
	return migrator.Status()
}

// RollbackMigration 回滚迁移
func RollbackMigration(db *gorm.DB) error {
	migrator := migrations.SetupMigrator(db)
	return migrator.Rollback()
}
