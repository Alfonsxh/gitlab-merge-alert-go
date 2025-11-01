package migrations

import (
	"gorm.io/gorm"
)

type Migration007 struct{}

func (m Migration007) ID() string {
	return "007_add_user_auth_and_resource_management"
}

func (m Migration007) Description() string {
	return "Add avatar field to accounts table and create resource_managers table"
}

func (m Migration007) Up(db *gorm.DB) error {
	if err := db.Exec(`
		ALTER TABLE accounts 
		ADD COLUMN avatar TEXT
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS resource_managers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			resource_id INTEGER NOT NULL,
			resource_type VARCHAR(20) NOT NULL,
			manager_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (manager_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (created_by) REFERENCES accounts(id) ON DELETE CASCADE,
			UNIQUE(resource_id, resource_type, manager_id)
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX idx_resource ON resource_managers (resource_id, resource_type)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX idx_manager ON resource_managers (manager_id)
	`).Error; err != nil {
		return err
	}

	return nil
}

func (m Migration007) Down(db *gorm.DB) error {
	if err := db.Exec(`DROP TABLE IF EXISTS resource_managers`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE accounts_backup AS SELECT 
			id, username, password_hash, email, role, is_active, last_login_at, created_at, updated_at
		FROM accounts
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`DROP TABLE accounts`).Error; err != nil {
		return err
	}

	if err := db.Exec(`ALTER TABLE accounts_backup RENAME TO accounts`).Error; err != nil {
		return err
	}

	return nil
}
