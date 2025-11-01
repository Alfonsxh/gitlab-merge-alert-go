package migrations

import (
	"gorm.io/gorm"
)

type Migration004AddGitLabUsernameToUsers struct{}

func (m *Migration004AddGitLabUsernameToUsers) ID() string {
	return "004_add_gitlab_username_to_users"
}

func (m *Migration004AddGitLabUsernameToUsers) Description() string {
	return "Add gitlab_username column to users table"
}

func (m *Migration004AddGitLabUsernameToUsers) Up(db *gorm.DB) error {
	// 检查是否已经存在 gitlab_username 字段
	var columnExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('users') WHERE name = 'gitlab_username'").Scan(&columnExists).Error; err != nil {
		return err
	}

	// 如果字段不存在，则添加
	if !columnExists {
		// 1. 添加 gitlab_username 字段
		if err := db.Exec("ALTER TABLE users ADD COLUMN gitlab_username TEXT DEFAULT ''").Error; err != nil {
			return err
		}

		// 2. 创建唯一索引（允许空值）
		if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_gitlab_username ON users(gitlab_username) WHERE gitlab_username != ''").Error; err != nil {
			return err
		}
	}

	return nil
}

func (m *Migration004AddGitLabUsernameToUsers) Down(db *gorm.DB) error {
	// 检查字段是否存在
	var columnExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('users') WHERE name = 'gitlab_username'").Scan(&columnExists).Error; err != nil {
		return err
	}

	if columnExists {
		// 删除索引
		db.Exec("DROP INDEX IF EXISTS idx_users_gitlab_username")

		// 注意：SQLite 不支持直接删除列，需要重建表
		// 在生产环境中，可以考虑通过重建表的方式来删除列
		// 这里为了安全起见，我们保留字段但清空数据
		if err := db.Exec("UPDATE users SET gitlab_username = ''").Error; err != nil {
			return err
		}
	}

	return nil
}
