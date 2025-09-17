package migrations

import "gorm.io/gorm"

type Migration008AddGitLabAccessTokenToAccounts struct{}

func (m Migration008AddGitLabAccessTokenToAccounts) ID() string {
	return "008_add_gitlab_access_token_to_accounts"
}

func (m Migration008AddGitLabAccessTokenToAccounts) Description() string {
	return "Add gitlab_access_token column to accounts table"
}

func (m Migration008AddGitLabAccessTokenToAccounts) Up(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE accounts 
		ADD COLUMN gitlab_access_token TEXT
	`).Error
}

func (m Migration008AddGitLabAccessTokenToAccounts) Down(db *gorm.DB) error {
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
				created_at,
				updated_at,
				avatar
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

		return nil
	})
}
