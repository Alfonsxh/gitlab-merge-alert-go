package migrations

import (
	"gorm.io/gorm"
)

type Migration002FixNotNullConstraints struct{}

func (m *Migration002FixNotNullConstraints) ID() string {
	return "002_fix_not_null_constraints"
}

func (m *Migration002FixNotNullConstraints) Description() string {
	return "Fix NOT NULL constraints by adding default values and cleaning up data"
}

func (m *Migration002FixNotNullConstraints) Up(db *gorm.DB) error {
	// 清理用户表中的空值数据
	if err := m.cleanupUserData(db); err != nil {
		return err
	}

	// 清理项目表中的空值数据
	if err := m.cleanupProjectData(db); err != nil {
		return err
	}

	// 清理Webhook表中的空值数据
	if err := m.cleanupWebhookData(db); err != nil {
		return err
	}

	// 清理通知表中的空值数据
	if err := m.cleanupNotificationData(db); err != nil {
		return err
	}

	return nil
}

func (m *Migration002FixNotNullConstraints) Down(db *gorm.DB) error {
	// 回滚操作通常不需要处理，因为我们只是清理了数据
	// 如果需要，可以在这里添加相反的操作
	return nil
}

func (m *Migration002FixNotNullConstraints) cleanupUserData(db *gorm.DB) error {
	// 删除email或phone为空的用户记录
	if err := db.Exec("DELETE FROM users WHERE email = '' OR email IS NULL OR phone = '' OR phone IS NULL").Error; err != nil {
		return err
	}

	// 更新剩余记录确保非空
	if err := db.Exec("UPDATE users SET email = 'unknown@example.com' WHERE email = '' OR email IS NULL").Error; err != nil {
		return err
	}

	if err := db.Exec("UPDATE users SET phone = '00000000000' WHERE phone = '' OR phone IS NULL").Error; err != nil {
		return err
	}

	return nil
}

func (m *Migration002FixNotNullConstraints) cleanupProjectData(db *gorm.DB) error {
	// 删除必要字段为空的项目记录
	if err := db.Exec("DELETE FROM projects WHERE gitlab_project_id = 0 OR name = '' OR name IS NULL OR url = '' OR url IS NULL").Error; err != nil {
		return err
	}

	// 更新剩余记录确保非空
	if err := db.Exec("UPDATE projects SET name = 'Unknown Project' WHERE name = '' OR name IS NULL").Error; err != nil {
		return err
	}

	if err := db.Exec("UPDATE projects SET url = 'https://example.com' WHERE url = '' OR url IS NULL").Error; err != nil {
		return err
	}

	return nil
}

func (m *Migration002FixNotNullConstraints) cleanupWebhookData(db *gorm.DB) error {
	// 删除name或url为空的webhook记录
	if err := db.Exec("DELETE FROM webhooks WHERE name = '' OR name IS NULL OR url = '' OR url IS NULL").Error; err != nil {
		return err
	}

	// 更新剩余记录确保非空
	if err := db.Exec("UPDATE webhooks SET name = 'Unknown Webhook' WHERE name = '' OR name IS NULL").Error; err != nil {
		return err
	}

	if err := db.Exec("UPDATE webhooks SET url = 'https://example.com/webhook' WHERE url = '' OR url IS NULL").Error; err != nil {
		return err
	}

	return nil
}

func (m *Migration002FixNotNullConstraints) cleanupNotificationData(db *gorm.DB) error {
	// 删除project_id或merge_request_id为0的通知记录
	if err := db.Exec("DELETE FROM notifications WHERE project_id = 0 OR merge_request_id = 0").Error; err != nil {
		return err
	}

	// 清理project_webhooks表中的无效记录
	if err := db.Exec("DELETE FROM project_webhooks WHERE project_id = 0 OR webhook_id = 0").Error; err != nil {
		return err
	}

	return nil
}
