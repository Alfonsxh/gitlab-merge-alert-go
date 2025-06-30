package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gitlab-merge-alert-go/pkg/logger"
)

type WeChatService struct {
	client *http.Client
}

func NewWeChatService() *WeChatService {
	return &WeChatService{
		client: &http.Client{},
	}
}

type WeChatMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content             string   `json:"content"`
		MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
	} `json:"text"`
}

func (s *WeChatService) SendMessage(webhookURL, content string, mentionedMobiles []string) error {
	logger.GetLogger().Infof("准备发送企业微信消息到: %s", webhookURL)
	logger.GetLogger().Infof("消息内容: %s", content)
	logger.GetLogger().Infof("需要@的手机号列表: %v", mentionedMobiles)

	message := WeChatMessage{
		MsgType: "text",
	}
	message.Text.Content = content
	message.Text.MentionedMobileList = mentionedMobiles

	// 记录完整的发送数据
	if messageJSON, err := json.MarshalIndent(message, "", "  "); err == nil {
		logger.GetLogger().Infof("发送给企业微信的完整消息结构:\n%s", string(messageJSON))
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.GetLogger().Errorf("序列化消息失败: %v", err)
		return err
	}

	resp, err := s.client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.GetLogger().Errorf("发送企业微信消息失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GetLogger().Errorf("企业微信 API 返回错误状态码: %d", resp.StatusCode)
		return fmt.Errorf("WeChat API returned status %d", resp.StatusCode)
	}

	logger.GetLogger().Infof("企业微信消息发送成功")
	return nil
}

func (s *WeChatService) FormatMergeRequestMessage(projectName, sourceBranch, targetBranch, mergeFrom, mergeTitle, clickURL string, mergeToList []string) string {
	content := fmt.Sprintf("%s\nProject: %s\nFrom: %s(%s)\nTo: %s\nMerge Info -> %s\nClick -> %s",
		strings.Repeat("=", 32)+" Merge Request "+strings.Repeat("=", 32),
		projectName,
		sourceBranch,
		mergeFrom,
		targetBranch,
		mergeTitle,
		clickURL,
	)

	return content
}
