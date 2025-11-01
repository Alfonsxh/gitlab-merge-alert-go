package models

import "time"

// WebhookDeliveryStat 记录渠道发送配额和次数
type WebhookDeliveryStat struct {
	ID          uint      `json:"id" gorm:"column:id;primarykey"`
	WebhookID   uint      `json:"webhook_id" gorm:"column:webhook_id;index;not null"`
	Channel     string    `json:"channel" gorm:"column:channel;not null;default:'';index"`
	PeriodStart time.Time `json:"period_start" gorm:"column:period_start;not null;index"`
	Count       uint      `json:"count" gorm:"column:count;not null;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (WebhookDeliveryStat) TableName() string {
	return "webhook_delivery_stats"
}
