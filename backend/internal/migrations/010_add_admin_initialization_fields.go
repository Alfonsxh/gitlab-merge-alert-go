package migrations

import "gorm.io/gorm"

type Migration010AddAdminInitializationFields struct{}

func (m Migration010AddAdminInitializationFields) ID() string {
	return "010_add_admin_initialization_fields"
}

func (m Migration010AddAdminInitializationFields) Description() string {
	return "Add admin initialization columns to accounts table"
}

func (m Migration010AddAdminInitializationFields) Up(db *gorm.DB) error {
	statements := []string{
		"ALTER TABLE accounts ADD COLUMN force_password_reset BOOLEAN NOT NULL DEFAULT 0",
		"ALTER TABLE accounts ADD COLUMN password_initialized_at DATETIME",
		"ALTER TABLE accounts ADD COLUMN admin_setup_token_hash TEXT",
		"ALTER TABLE accounts ADD COLUMN admin_setup_token_generated_at DATETIME",
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range statements {
			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (m Migration010AddAdminInitializationFields) Down(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
            CREATE TABLE accounts_backup AS SELECT 
                id,
                username,
                password_hash,
                email,
                role,
                is_active,
                last_login_at,
                avatar,
                gitlab_access_token,
                created_at,
                updated_at
            FROM accounts
        `).Error; err != nil {
			return err
		}

		if err := tx.Exec(`DROP TABLE accounts`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`ALTER TABLE accounts_backup RENAME TO accounts`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username)`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email)`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_role ON accounts(role)`).Error; err != nil {
			return err
		}

		return nil
	})
}
