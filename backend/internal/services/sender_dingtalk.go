package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"
	"gitlab-merge-alert-go/pkg/ratelimit"

	"gorm.io/gorm"
)

var (
	ErrDingTalkRateLimited   = errors.New("dingtalk rate limit reached")
	ErrDingTalkQuotaExceeded = errors.New("dingtalk monthly quota exceeded")
)

type DingTalkSender struct {
	client       *http.Client
	limiter      *ratelimit.TokenBucket
	monthlyQuota int
	db           *gorm.DB
}

type dingTalkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type dingTalkMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		Mobiles []string `json:"atMobiles,omitempty"`
		IsAtAll bool     `json:"isAtAll"`
	} `json:"at"`
}

func NewDingTalkSender(db *gorm.DB, cfg config.DingTalkConfig) *DingTalkSender {
	timeout := cfg.RequestTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	limiter := ratelimit.NewTokenBucket(cfg.RateLimitPerMinute)

	return &DingTalkSender{
		client:       &http.Client{Timeout: timeout},
		limiter:      limiter,
		monthlyQuota: cfg.MonthlyQuota,
		db:           db,
	}
}

func (s *DingTalkSender) Send(ctx context.Context, webhook *models.Webhook, payload *MergeRequestPayload) error {
	if payload == nil {
		return errors.New("nil payload")
	}

	if s.monthlyQuota > 0 {
		exceeded, current, err := s.isQuotaExceeded(webhook.ID)
		if err != nil {
			return err
		}
		if exceeded {
			logger.GetLogger().Warnf("钉钉 webhook %d 已达到月度配额: %d", webhook.ID, current)
			return ErrDingTalkQuotaExceeded
		}
	}

	if !s.limiter.Allow() {
		logger.GetLogger().Warnf("钉钉 webhook %d 触发速率限制", webhook.ID)
		return ErrDingTalkRateLimited
	}

	webhook.ApplyDefaults()

	secret := ""
	if webhook.Settings != nil {
		secret = webhook.Settings.Secret
	}

	content := FormatMergeRequestPayloadText(payload)
	signedURL, timestamp := buildSignedDingTalkURL(webhook.URL, secret)

	message := dingTalkMessage{MsgType: "text"}
	message.Text.Content = content
	message.At.Mobiles = payload.MentionedMobiles
	message.At.IsAtAll = false

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal dingtalk message failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, signedURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create dingtalk request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	logger.GetLogger().Infof("发送钉钉通知 webhook=%s timestamp=%d body=%s", webhook.URL, timestamp, string(body))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send dingtalk message failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dingtalk http status %d", resp.StatusCode)
	}

	var response dingTalkResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decode dingtalk response failed: %w", err)
	}

	if response.ErrCode != 0 {
		return fmt.Errorf("dingtalk error %d: %s", response.ErrCode, response.ErrMsg)
	}

	if err := s.incrementQuota(webhook.ID); err != nil {
		return err
	}

	return nil
}

func (s *DingTalkSender) isQuotaExceeded(webhookID uint) (bool, uint, error) {
	if s.monthlyQuota <= 0 {
		return false, 0, nil
	}

	stat, err := s.ensureStatRecord(webhookID)
	if err != nil {
		return false, 0, err
	}

	return stat.Count >= uint(s.monthlyQuota), stat.Count, nil
}

func (s *DingTalkSender) incrementQuota(webhookID uint) error {
	if s.monthlyQuota <= 0 {
		return nil
	}

	stat, err := s.ensureStatRecord(webhookID)
	if err != nil {
		return err
	}

	return s.db.Model(&models.WebhookDeliveryStat{}).
		Where("id = ?", stat.ID).
		UpdateColumn("count", gorm.Expr("count + ?", 1)).Error
}

func (s *DingTalkSender) ensureStatRecord(webhookID uint) (*models.WebhookDeliveryStat, error) {
	period := startOfMonth(time.Now().UTC())
	stat := models.WebhookDeliveryStat{}
	err := s.db.Where("webhook_id = ? AND period_start = ?", webhookID, period).
		First(&stat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		stat = models.WebhookDeliveryStat{
			WebhookID:   webhookID,
			Channel:     models.WebhookTypeDingTalk,
			PeriodStart: period,
			Count:       0,
		}
		if err := s.db.Create(&stat).Error; err != nil {
			return nil, err
		}
		return &stat, nil
	}
	if err != nil {
		return nil, err
	}
	return &stat, nil
}

func buildSignedDingTalkURL(rawURL, secret string) (string, int64) {
	if secret == "" {
		return rawURL, 0
	}

	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	sign = url.QueryEscape(sign)

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, timestamp
	}

	query := parsed.Query()
	query.Set("timestamp", fmt.Sprintf("%d", timestamp))
	query.Set("sign", sign)
	parsed.RawQuery = query.Encode()
	return parsed.String(), timestamp
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
