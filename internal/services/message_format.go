package services

import (
	"fmt"
	"strings"
)

func FormatMergeRequestPayloadText(payload *MergeRequestPayload) string {
	if payload == nil {
		return ""
	}

	divider := strings.Repeat("=", 32) + " Merge Request " + strings.Repeat("=", 32)
	content := fmt.Sprintf(`%s
Project: %s
   From: %s -> %s (%s)
MR Info: %s
Click -> %s`,
		divider,
		payload.ProjectName,
		payload.SourceBranch,
		payload.TargetBranch,
		payload.AuthorName,
		payload.Title,
		payload.URL,
	)

	if len(payload.MentionedAccounts) > 0 {
		mentions := ""
		for _, account := range payload.MentionedAccounts {
			mentions += fmt.Sprintf(" @%s", account)
		}
		if mentions != "" {
			content += "\n" + strings.TrimSpace(mentions)
		}
	}

	return content
}

func FormatMergeRequestPayloadTextWithPhones(payload *MergeRequestPayload, mentionedMobiles []string) string {
	if payload == nil {
		return ""
	}

	divider := strings.Repeat("=", 32) + " Merge Request " + strings.Repeat("=", 32)
	content := fmt.Sprintf(`%s
Project: %s
   From: %s -> %s (%s)
MR Info: %s
Click -> %s`,
		divider,
		payload.ProjectName,
		payload.SourceBranch,
		payload.TargetBranch,
		payload.AuthorName,
		payload.Title,
		payload.URL,
	)

	return content
}
