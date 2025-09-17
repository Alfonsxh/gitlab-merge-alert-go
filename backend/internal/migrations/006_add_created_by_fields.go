package migrations

import (
	"gorm.io/gorm"
)

type Migration006AddCreatedByFields struct{}

func (m *Migration006AddCreatedByFields) ID() string {
	return "006_add_created_by_fields"
}

func (m *Migration006AddCreatedByFields) Description() string {
	return "Add created_by field to projects, webhooks, users tables for multi-tenancy"
}

func (m *Migration006AddCreatedByFields) Up(db *gorm.DB) error {
	// 检查并添加 created_by 字段到 projects 表
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('projects') WHERE name = 'created_by'").Scan(&count)
	if count == 0 {
		if err := db.Exec(`ALTER TABLE projects ADD COLUMN created_by INTEGER`).Error; err != nil {
			return err
		}
	}
	// 创建索引（如果不存在）
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by)`)

	// 检查并添加 created_by 字段到 webhooks 表
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('webhooks') WHERE name = 'created_by'").Scan(&count)
	if count == 0 {
		if err := db.Exec(`ALTER TABLE webhooks ADD COLUMN created_by INTEGER`).Error; err != nil {
			return err
		}
	}
	// 创建索引（如果不存在）
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_webhooks_created_by ON webhooks(created_by)`)

	// 检查并添加 created_by 字段到 users 表
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = 'created_by'").Scan(&count)
	if count == 0 {
		if err := db.Exec(`ALTER TABLE users ADD COLUMN created_by INTEGER`).Error; err != nil {
			return err
		}
	}
	// 创建索引（如果不存在）
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_created_by ON users(created_by)`)

	// 检查并添加 owner_id 字段到 notifications 表
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('notifications') WHERE name = 'owner_id'").Scan(&count)
	if count == 0 {
		if err := db.Exec(`ALTER TABLE notifications ADD COLUMN owner_id INTEGER`).Error; err != nil {
			return err
		}
	}
	// 创建索引（如果不存在）
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_owner_id ON notifications(owner_id)`)

	return nil
}

func (m *Migration006AddCreatedByFields) Down(db *gorm.DB) error {
	// 删除索引
	indexes := []string{
		"idx_projects_created_by",
		"idx_webhooks_created_by",
		"idx_users_created_by",
		"idx_notifications_owner_id",
	}
	
	for _, idx := range indexes {
		if err := db.Exec("DROP INDEX IF EXISTS " + idx).Error; err != nil {
			return err
		}
	}

	// SQLite 不支持直接删除列，需要重建表
	// 这里为了简化，只删除索引
	return nil
}