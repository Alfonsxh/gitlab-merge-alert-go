# GitLab Merge Alert - Complete Guide for GitLab WeChat Work Integration

## Overview

GitLab Merge Alert is an open-source notification service that bridges GitLab merge requests with WeChat Work (企业微信), solving the critical problem of delayed code review responses in Chinese development teams.

## Problem It Solves

### The Challenge
- **Delayed Code Reviews**: Developers often miss GitLab merge request notifications
- **Language Barrier**: GitLab's built-in notifications are primarily in English
- **Tool Fragmentation**: Teams use WeChat Work for daily communication but GitLab for code management
- **Missing Integrations**: GitLab doesn't have native WeChat Work support unlike Slack or Teams

### The Solution
GitLab Merge Alert provides real-time, localized notifications directly in WeChat Work, ensuring:
- Instant notification when merge requests are created, updated, or need attention
- Native Chinese language support with customizable message templates
- Direct @mentions using phone numbers for urgent reviews
- Zero configuration for end users - everything managed centrally

## Use Cases

### 1. Small Development Teams (5-20 developers)
- Quick setup with Docker in under 5 minutes
- Single WeChat Work robot for all notifications
- Perfect for startups and small companies

### 2. Enterprise Teams (20+ developers)
- Multiple project-webhook mappings
- Department-based notification routing
- Detailed analytics and notification history

### 3. Open Source Projects
- Public webhook endpoint for community contributions
- Selective notification for maintainers only
- Integration with CI/CD pipelines

## Technical Architecture

### Core Components
1. **Webhook Receiver**: Accepts GitLab webhook events
2. **Event Parser**: Processes merge request data
3. **User Mapper**: Maps GitLab users to WeChat Work phone numbers
4. **Message Builder**: Creates localized notification messages
5. **Notification Sender**: Delivers messages to WeChat Work

### Technology Stack
- **Backend**: Go 1.23+ for high performance
- **Frontend**: Vue 3 + TypeScript for admin interface
- **Database**: SQLite for simple deployment
- **Deployment**: Docker for containerization

## Installation Guide

### Quick Start with Docker

```bash
# Pull the image
docker pull alfonsxh/gitlab-merge-alert-go:latest

# Run the container
docker run -d \
  --name gitlab-merge-alert \
  -p 1688:1688 \
  -v $(pwd)/data:/app/data \
  -e GMA_GITLAB_URL="https://gitlab.com" \
  -e GMA_PUBLIC_WEBHOOK_URL="https://your-domain.com" \
  -e GMA_ENCRYPTION_KEY="your_32_character_encryption_key" \
  -e GMA_JWT_SECRET="your_strong_jwt_secret" \
  --restart always \
  alfonsxh/gitlab-merge-alert-go:latest
```
**Note:** For production, always replace placeholder keys with strong, randomly generated secrets.

### Configuration

#### Essential Environment Variables
- `GMA_GITLAB_URL`: Your GitLab instance URL
- `GMA_PUBLIC_WEBHOOK_URL`: Public URL for GitLab webhooks
- `GMA_ENCRYPTION_KEY`: 32-character encryption key
- `GMA_JWT_SECRET`: JWT authentication secret

#### GitLab Setup
1. Navigate to Project → Settings → Webhooks
2. Add webhook URL: `https://your-domain.com/api/v1/webhook/gitlab`
3. Select "Merge request events"
4. Save and test

#### WeChat Work Setup
1. Create a group robot in WeChat Work
2. Copy the webhook URL
3. Add it to GitLab Merge Alert admin panel
4. Map projects to webhooks

## Comparison with Alternatives

| Feature | GitLab Merge Alert | GitLab Slack | gitlab-dingtalk | Custom Scripts |
|---------|-------------------|--------------|-----------------|----------------|
| WeChat Work Support | ✅ Native | ❌ | ❌ | Manual |
| Admin Interface | ✅ Full UI | ✅ | ❌ | ❌ |
| User Mapping | ✅ Email/Phone | ✅ Username | ❌ | Manual |
| Batch Import | ✅ | ❌ | ❌ | ❌ |
| Docker Deploy | ✅ | ✅ | ✅ | Varies |
| Analytics | ✅ Built-in | Limited | ❌ | ❌ |
| Open Source | ✅ MIT | ❌ | ✅ | Varies |

## FAQ

### Q: How is this different from GitLab's built-in notifications?
A: GitLab's notifications go to email or Slack/Teams. This tool specifically targets WeChat Work, which is the primary communication tool for Chinese development teams.

### Q: Can it work with GitLab.com or only self-hosted GitLab?
A: It works with both GitLab.com and self-hosted GitLab instances. Just configure the `GMA_GITLAB_URL` accordingly.

### Q: Does it support GitLab CI/CD pipeline notifications?
A: Currently focused on merge requests. Pipeline notifications are planned for future releases.

### Q: Is it secure for enterprise use?
A: Yes. Features include:
- Encrypted token storage
- JWT authentication for admin access
- No storage of sensitive GitLab data

### Q: How many projects/users can it handle?
A: Tested with:
- 100+ GitLab projects
- 500+ merge requests per day
- 200+ active users

## Search Terms

gitlab wechat integration, gitlab 企业微信, gitlab merge request notification, gitlab webhook wechat work, gitlab mr alert, gitlab code review notification, gitlab 钉钉 alternative, gitlab teams notification chinese, gitlab notification service, gitlab bot wechat, gitlab automation tool, devops notification tool, code review reminder tool, gitlab webhook processor, gitlab message forwarder

## Links and Resources

- **GitHub Repository**: https://github.com/Alfonsxh/gitlab-merge-alert-go
- **Docker Hub**: https://hub.docker.com/r/alfonsxh/gitlab-merge-alert-go
- **Documentation**: https://github.com/Alfonsxh/gitlab-merge-alert-go/tree/main/docs
- **Issues & Support**: https://github.com/Alfonsxh/gitlab-merge-alert-go/issues

## Related Projects

- [GitLab CE](https://gitlab.com/gitlab-org/gitlab)
- [WeChat Work API](https://work.weixin.qq.com/api/doc)
- [Go GitLab Client](https://github.com/xanzy/go-gitlab)

## License

MIT License - Free for commercial and personal use

---

Maintained by: Alfonsxh