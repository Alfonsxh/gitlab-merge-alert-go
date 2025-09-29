package models

import (
	"database/sql/driver"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const (
	WebhookTypeWeCom    = "wechat"
	WebhookTypeDingTalk = "dingtalk"
	WebhookTypeCustom   = "custom"
	WebhookTypeAuto     = "auto"

	SignatureMethodHMACSHA256 = "hmac_sha256"
)

type StringList []string

type StringMap map[string]string

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*s = nil
			return nil
		}
		return json.Unmarshal(v, s)
	case string:
		if v == "" {
			*s = nil
			return nil
		}
		return json.Unmarshal([]byte(v), s)
	default:
		return json.Unmarshal([]byte("[]"), s)
	}
}

func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(s))
}

func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*m = nil
			return nil
		}
		return json.Unmarshal(v, m)
	case string:
		if v == "" {
			*m = nil
			return nil
		}
		return json.Unmarshal([]byte(v), m)
	default:
		return json.Unmarshal([]byte("{}"), m)
	}
}

func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]string(m))
}

type Webhook struct {
	ID          uint      `json:"id" gorm:"column:id;primarykey"`
	Name        string    `json:"name" gorm:"column:name;not null;default:''"`
	URL         string    `json:"url" gorm:"column:url;not null;default:''"`
	Description string    `json:"description" gorm:"column:description"`
	Type        string    `json:"type" gorm:"column:type;not null;default:'wechat'"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedBy   *uint     `json:"created_by,omitempty" gorm:"column:created_by;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`

	Settings *WebhookSetting `json:"settings,omitempty" gorm:"constraint:OnDelete:CASCADE;"`

	Projects []Project `json:"projects,omitempty" gorm:"many2many:project_webhooks;"`
}

type WebhookSetting struct {
	ID               uint       `json:"-" gorm:"column:id;primarykey"`
	WebhookID        uint       `json:"-" gorm:"column:webhook_id;uniqueIndex"`
	SignatureMethod  string     `json:"signature_method" gorm:"column:signature_method;not null;default:'hmac_sha256'"`
	Secret           string     `json:"secret" gorm:"column:secret"`
	SecurityKeywords StringList `json:"security_keywords" gorm:"column:security_keywords;type:json"`
	CustomHeaders    StringMap  `json:"custom_headers" gorm:"column:custom_headers;type:json"`
	CreatedAt        time.Time  `json:"-" gorm:"column:created_at"`
	UpdatedAt        time.Time  `json:"-" gorm:"column:updated_at"`
}

type ProjectWebhook struct {
	ID        uint      `json:"id" gorm:"column:id;primarykey"`
	ProjectID uint      `json:"project_id" gorm:"column:project_id;not null;default:0"`
	WebhookID uint      `json:"webhook_id" gorm:"column:webhook_id;not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`

	Project Project `json:"project" gorm:"foreignKey:ProjectID"`
	Webhook Webhook `json:"webhook" gorm:"foreignKey:WebhookID"`
}

type CreateWebhookRequest struct {
	Name             string            `json:"name" binding:"required"`
	URL              string            `json:"url" binding:"required,url"`
	Description      string            `json:"description"`
	Type             string            `json:"type" binding:"omitempty,oneof=wechat dingtalk custom auto"`
	SignatureMethod  string            `json:"signature_method" binding:"omitempty,oneof=hmac_sha256"`
	Secret           string            `json:"secret"`
	SecurityKeywords []string          `json:"security_keywords"`
	CustomHeaders    map[string]string `json:"custom_headers"`
	IsActive         *bool             `json:"is_active"`
}

type UpdateWebhookRequest struct {
	Name             string            `json:"name"`
	URL              string            `json:"url" binding:"omitempty,url"`
	Description      string            `json:"description"`
	Type             string            `json:"type" binding:"omitempty,oneof=wechat dingtalk custom auto"`
	SignatureMethod  string            `json:"signature_method" binding:"omitempty,oneof=hmac_sha256"`
	Secret           *string           `json:"secret"`
	SecurityKeywords []string          `json:"security_keywords"`
	CustomHeaders    map[string]string `json:"custom_headers"`
	IsActive         *bool             `json:"is_active"`
}

type WebhookResponse struct {
	ID               uint              `json:"id"`
	Name             string            `json:"name"`
	URL              string            `json:"url"`
	Description      string            `json:"description"`
	Type             string            `json:"type"`
	SignatureMethod  string            `json:"signature_method"`
	Secret           string            `json:"secret,omitempty"`
	SecurityKeywords []string          `json:"security_keywords,omitempty"`
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`
	IsActive         bool              `json:"is_active"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	Projects         []ProjectResponse `json:"projects,omitempty"`
}

type LinkProjectWebhookRequest struct {
	ProjectID uint `json:"project_id" binding:"required"`
	WebhookID uint `json:"webhook_id" binding:"required"`
}

func (w *Webhook) ApplyDefaults() {
	if w.Type == "" {
		w.Type = WebhookTypeWeCom
	}
	if w.Settings != nil {
		w.Settings.ApplyDefaults()
	}
}

func (s *WebhookSetting) ApplyDefaults() {
	if s == nil {
		return
	}
	if s.SignatureMethod == "" {
		s.SignatureMethod = SignatureMethodHMACSHA256
	}
}

func (w *Webhook) EnsureSettings() *WebhookSetting {
	if w.Settings == nil {
		w.Settings = &WebhookSetting{}
	}
	return w.Settings
}

func DetectWebhookType(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return WebhookTypeCustom
	}

	host := strings.ToLower(parsed.Host)
	switch {
	case strings.Contains(host, "dingtalk.com") || strings.Contains(host, "ding"):
		return WebhookTypeDingTalk
	case strings.Contains(host, "qyapi.weixin.qq.com") || strings.Contains(host, "work.weixin.qq.com"):
		return WebhookTypeWeCom
	default:
		return WebhookTypeCustom
	}
}

func ToStringList(values []string) StringList {
	if len(values) == 0 {
		return nil
	}
	return StringList(values)
}

func ToStringMap(values map[string]string) StringMap {
	if len(values) == 0 {
		return nil
	}
	return StringMap(values)
}

func (w *Webhook) SecurityKeywordsAsSlice() []string {
	if w.Settings == nil || len(w.Settings.SecurityKeywords) == 0 {
		return nil
	}
	return append([]string(nil), w.Settings.SecurityKeywords...)
}

func (w *Webhook) CustomHeadersAsMap() map[string]string {
	if w.Settings == nil || len(w.Settings.CustomHeaders) == 0 {
		return nil
	}
	out := make(map[string]string, len(w.Settings.CustomHeaders))
	for k, v := range w.Settings.CustomHeaders {
		out[k] = v
	}
	return out
}

func (s StringList) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(s))
}

func (m StringMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]string(m))
}
