package migrations

import (
	"gitlab-merge-alert-go/internal/models"
	"gorm.io/gorm"
)

type Migration001CreateInitialTables struct{}

func (m *Migration001CreateInitialTables) ID() string {
	return "001_create_initial_tables"
}

func (m *Migration001CreateInitialTables) Description() string {
	return "Create initial database tables"
}

func (m *Migration001CreateInitialTables) Up(db *gorm.DB) error {
	// 检查表是否已存在，如果已存在则跳过创建
	if db.Migrator().HasTable(&models.User{}) {
		// 表已存在，只更新结构
		return db.AutoMigrate(
			&models.User{},
			&models.Project{},
			&models.Webhook{},
			&models.ProjectWebhook{},
			&models.Notification{},
		)
	}

	// 创建所有表
	return db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Webhook{},
		&models.ProjectWebhook{},
		&models.Notification{},
	)
}

func (m *Migration001CreateInitialTables) Down(db *gorm.DB) error {
	// 删除所有表（谨慎使用）
	return db.Migrator().DropTable(
		&models.Notification{},
		&models.ProjectWebhook{},
		&models.Webhook{},
		&models.Project{},
		&models.User{},
	)
}
