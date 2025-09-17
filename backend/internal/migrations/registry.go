package migrations

import (
	"gorm.io/gorm"
)

// GetAllMigrations 获取所有迁移
func GetAllMigrations() []Migration {
	return []Migration{
		&Migration001CreateInitialTables{},
		&Migration002FixNotNullConstraints{},
		&Migration003FixGitLabWebhookIDColumnName{},
		&Migration004AddGitLabUsernameToUsers{},
		&Migration005CreateAccountsTable{},
		&Migration006AddCreatedByFields{},
		&Migration007{},
		&Migration008AddGitLabAccessTokenToAccounts{},
	}
}

// SetupMigrator 设置迁移器
func SetupMigrator(db *gorm.DB) *Migrator {
	migrator := NewMigrator(db)

	// 注册所有迁移
	for _, migration := range GetAllMigrations() {
		migrator.AddMigration(migration)
	}

	return migrator
}
