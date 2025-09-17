package migrations

import (
	"gorm.io/gorm"
)

type Migration005CreateAccountsTable struct{}

func (m *Migration005CreateAccountsTable) ID() string {
	return "005_create_accounts_table"
}

func (m *Migration005CreateAccountsTable) Description() string {
	return "Create accounts table for user authentication"
}

func (m *Migration005CreateAccountsTable) Up(db *gorm.DB) error {
	// 创建 accounts 表
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			role VARCHAR(20) DEFAULT 'user',
			is_active BOOLEAN DEFAULT TRUE,
			last_login_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);
		CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);
		CREATE INDEX IF NOT EXISTS idx_accounts_role ON accounts(role);
	`).Error
}

func (m *Migration005CreateAccountsTable) Down(db *gorm.DB) error {
	// 删除索引
	if err := db.Exec("DROP INDEX IF EXISTS idx_accounts_role").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP INDEX IF EXISTS idx_accounts_email").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP INDEX IF EXISTS idx_accounts_username").Error; err != nil {
		return err
	}
	
	// 删除表
	return db.Exec("DROP TABLE IF EXISTS accounts").Error
}