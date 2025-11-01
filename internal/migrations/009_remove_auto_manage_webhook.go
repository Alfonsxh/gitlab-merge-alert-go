package migrations

import "gorm.io/gorm"

type Migration009RemoveAutoManageWebhook struct{}

func (m Migration009RemoveAutoManageWebhook) ID() string {
	return "009_remove_auto_manage_webhook"
}

func (m Migration009RemoveAutoManageWebhook) Description() string {
	return "Remove auto_manage_webhook column from projects table"
}

func (m Migration009RemoveAutoManageWebhook) Up(db *gorm.DB) error {
	// SQLite 不支持直接删除列，需要重建表
	return db.Transaction(func(tx *gorm.DB) error {
		// 创建新表（不包含 auto_manage_webhook 字段）
		if err := tx.Exec(`
			CREATE TABLE projects_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				gitlab_project_id INTEGER NOT NULL DEFAULT 0,
				name TEXT NOT NULL DEFAULT '',
				url TEXT NOT NULL DEFAULT '',
				description TEXT,
				access_token TEXT,
				gitlab_webhook_id INTEGER,
				webhook_synced BOOLEAN DEFAULT 0,
				last_sync_at DATETIME,
				created_by INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`).Error; err != nil {
			return err
		}

		// 复制数据（不包含 auto_manage_webhook）
		if err := tx.Exec(`
			INSERT INTO projects_new (
				id, gitlab_project_id, name, url, description, access_token,
				gitlab_webhook_id, webhook_synced, last_sync_at,
				created_by, created_at, updated_at
			)
			SELECT
				id, gitlab_project_id, name, url, description, access_token,
				gitlab_webhook_id, webhook_synced, last_sync_at,
				created_by, created_at, updated_at
			FROM projects
		`).Error; err != nil {
			return err
		}

		// 删除旧表
		if err := tx.Exec(`DROP TABLE projects`).Error; err != nil {
			return err
		}

		// 重命名新表
		if err := tx.Exec(`ALTER TABLE projects_new RENAME TO projects`).Error; err != nil {
			return err
		}

		// 重建索引
		if err := tx.Exec(`CREATE UNIQUE INDEX idx_projects_gitlab_project_id ON projects (gitlab_project_id)`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX idx_projects_gitlab_webhook_id ON projects (gitlab_webhook_id)`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX idx_projects_created_by ON projects (created_by)`).Error; err != nil {
			return err
		}

		return nil
	})
}

func (m Migration009RemoveAutoManageWebhook) Down(db *gorm.DB) error {
	// 恢复 auto_manage_webhook 字段
	return db.Exec(`
		ALTER TABLE projects
		ADD COLUMN auto_manage_webhook BOOLEAN DEFAULT 1
	`).Error
}
