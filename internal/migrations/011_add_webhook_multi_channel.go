package migrations

import (
	"fmt"
	"time"

	"github.com/Alfonsxh/gitlab-merge-alert-go/internal/models"
	"gorm.io/gorm"
)

type Migration011AddWebhookMultiChannel struct{}

func (m Migration011AddWebhookMultiChannel) ID() string {
	return "011_add_webhook_multi_channel"
}

func (m Migration011AddWebhookMultiChannel) Description() string {
	return "Introduce webhook type metadata, settings table, and delivery stats"
}

func (m Migration011AddWebhookMultiChannel) Up(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&models.Webhook{}, &models.WebhookSetting{}, &models.WebhookDeliveryStat{}); err != nil {
			return fmt.Errorf("auto migrate tables failed: %w", err)
		}

		var webhooks []models.Webhook
		if err := tx.Find(&webhooks).Error; err != nil {
			return fmt.Errorf("load existing webhooks failed: %w", err)
		}

		periodStart := time.Now().UTC().Truncate(24 * time.Hour)
		periodStart = time.Date(periodStart.Year(), periodStart.Month(), 1, 0, 0, 0, 0, time.UTC)

		for _, webhook := range webhooks {
			detectedType := models.DetectWebhookType(webhook.URL)
			if err := tx.Model(&models.Webhook{}).Where("id = ?", webhook.ID).Update("type", detectedType).Error; err != nil {
				return fmt.Errorf("update webhook %d type failed: %w", webhook.ID, err)
			}

			setting := models.WebhookSetting{
				WebhookID:       webhook.ID,
				SignatureMethod: models.SignatureMethodHMACSHA256,
			}
			if err := tx.Where("webhook_id = ?", webhook.ID).FirstOrCreate(&setting).Error; err != nil {
				return fmt.Errorf("init webhook setting %d failed: %w", webhook.ID, err)
			}

			stat := models.WebhookDeliveryStat{
				WebhookID:   webhook.ID,
				Channel:     detectedType,
				PeriodStart: periodStart,
			}
			if err := tx.Where("webhook_id = ? AND period_start = ?", webhook.ID, stat.PeriodStart).FirstOrCreate(&stat).Error; err != nil {
				return fmt.Errorf("initialize webhook delivery stat failed: %w", err)
			}
		}

		return nil
	})
}

func (m Migration011AddWebhookMultiChannel) Down(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable(&models.WebhookDeliveryStat{}) {
			if err := tx.Migrator().DropTable(&models.WebhookDeliveryStat{}); err != nil {
				return err
			}
		}

		if tx.Migrator().HasTable(&models.WebhookSetting{}) {
			if err := tx.Migrator().DropTable(&models.WebhookSetting{}); err != nil {
				return err
			}
		}

		if err := tx.Exec(`
            CREATE TABLE webhooks_backup (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL DEFAULT '',
                url TEXT NOT NULL DEFAULT '',
                description TEXT,
                type TEXT NOT NULL DEFAULT 'wechat',
                is_active BOOLEAN DEFAULT 1,
                created_by INTEGER,
                created_at DATETIME,
                updated_at DATETIME
            )
        `).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
            INSERT INTO webhooks_backup (
                id, name, url, description, type, is_active, created_by, created_at, updated_at
            )
            SELECT
                id, name, url, description, type, is_active, created_by, created_at, updated_at
            FROM webhooks
        `).Error; err != nil {
			return err
		}

		if err := tx.Exec(`DROP TABLE webhooks`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`ALTER TABLE webhooks_backup RENAME TO webhooks`).Error; err != nil {
			return err
		}

		return nil
	})
}
