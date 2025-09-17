package migrations

import (
	"gorm.io/gorm"
)

type Migration003FixGitLabWebhookIDColumnName struct{}

func (m *Migration003FixGitLabWebhookIDColumnName) ID() string {
	return "003_fix_gitlab_webhook_id_column_name"
}

func (m *Migration003FixGitLabWebhookIDColumnName) Description() string {
	return "Fix GitLab webhook ID column name from git_lab_webhook_id to gitlab_webhook_id"
}

func (m *Migration003FixGitLabWebhookIDColumnName) Up(db *gorm.DB) error {
	// 检查是否存在旧的字段名
	var oldColumnExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('projects') WHERE name = 'git_lab_webhook_id'").Scan(&oldColumnExists).Error; err != nil {
		return err
	}

	// 检查是否存在新的字段名
	var newColumnExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('projects') WHERE name = 'gitlab_webhook_id'").Scan(&newColumnExists).Error; err != nil {
		return err
	}

	// 如果旧字段存在但新字段不存在，则进行重命名
	if oldColumnExists && !newColumnExists {
		// 1. 添加新字段
		if err := db.Exec("ALTER TABLE projects ADD COLUMN gitlab_webhook_id INTEGER").Error; err != nil {
			return err
		}

		// 2. 复制数据从旧字段到新字段
		if err := db.Exec("UPDATE projects SET gitlab_webhook_id = git_lab_webhook_id").Error; err != nil {
			return err
		}

		// 3. 删除旧索引（如果存在）
		db.Exec("DROP INDEX IF EXISTS idx_projects_git_lab_webhook_id")

		// 4. 创建新索引
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_gitlab_webhook_id ON projects(gitlab_webhook_id)").Error; err != nil {
			return err
		}

		// 5. 在SQLite中我们不能直接删除列，但可以通过创建临时表的方式重建表结构
		// 为了安全起见，我们保留旧字段，只是不再使用它
		// 实际生产环境中可以考虑完整的表重建操作
	}

	return nil
}

func (m *Migration003FixGitLabWebhookIDColumnName) Down(db *gorm.DB) error {
	// 检查是否存在新字段
	var newColumnExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('projects') WHERE name = 'gitlab_webhook_id'").Scan(&newColumnExists).Error; err != nil {
		return err
	}

	// 检查是否存在旧字段
	var oldColumnExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('projects') WHERE name = 'git_lab_webhook_id'").Scan(&oldColumnExists).Error; err != nil {
		return err
	}

	if newColumnExists && oldColumnExists {
		// 回滚：将数据从新字段复制回旧字段
		if err := db.Exec("UPDATE projects SET git_lab_webhook_id = gitlab_webhook_id").Error; err != nil {
			return err
		}

		// 删除新索引
		db.Exec("DROP INDEX IF EXISTS idx_projects_gitlab_webhook_id")

		// 重建旧索引
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_git_lab_webhook_id ON projects(git_lab_webhook_id)").Error; err != nil {
			return err
		}

		// 注意：我们不删除新字段，以免数据丢失
		// 在SQLite中删除列需要重建整个表
	}

	return nil
}
