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
	// 添加 created_by 字段到 projects 表
	if err := db.Exec(`
		ALTER TABLE projects ADD COLUMN created_by INTEGER;
		CREATE INDEX idx_projects_created_by ON projects(created_by);
	`).Error; err != nil {
		return err
	}

	// 添加 created_by 字段到 webhooks 表
	if err := db.Exec(`
		ALTER TABLE webhooks ADD COLUMN created_by INTEGER;
		CREATE INDEX idx_webhooks_created_by ON webhooks(created_by);
	`).Error; err != nil {
		return err
	}

	// 添加 created_by 字段到 users 表
	if err := db.Exec(`
		ALTER TABLE users ADD COLUMN created_by INTEGER;
		CREATE INDEX idx_users_created_by ON users(created_by);
	`).Error; err != nil {
		return err
	}

	// 添加 owner_id 字段到 notifications 表
	if err := db.Exec(`
		ALTER TABLE notifications ADD COLUMN owner_id INTEGER;
		CREATE INDEX idx_notifications_owner_id ON notifications(owner_id);
	`).Error; err != nil {
		return err
	}

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