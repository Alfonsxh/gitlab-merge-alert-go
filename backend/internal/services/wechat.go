package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab-merge-alert-go/pkg/logger"
	"net/http"
)

type weChatService struct {
	client *http.Client
}

func NewWeChatService() WeChatService {
	return &weChatService{
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

func (s *weChatService) SendMessage(webhookURL, content string, mentionedMobiles []string) error {
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

func (s *weChatService) FormatMergeRequestMessage(projectName, sourceBranch, targetBranch, mergeFrom, mergeTitle, clickURL string, mergeToList []string, mentionedMobiles []string) string {

	payload := &MergeRequestPayload{
		ProjectName:       projectName,
		SourceBranch:      sourceBranch,
		TargetBranch:      targetBranch,
		AuthorName:        mergeFrom,
		Title:             mergeTitle,
		URL:               clickURL,
		MentionedAccounts: mergeToList,
	}

	return FormatMergeRequestPayloadTextWithPhones(payload, mentionedMobiles)
}
