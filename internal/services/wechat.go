package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	message := WeChatMessage{
		MsgType: "text",
	}
	message.Text.Content = content
	message.Text.MentionedMobileList = mentionedMobiles

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := s.client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WeChat API returned status %d", resp.StatusCode)
	}

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
