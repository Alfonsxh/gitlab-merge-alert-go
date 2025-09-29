# 多渠道 Webhook 设计与实施指引

> 更新时间：2025-09-28

## 1. 背景概述

GitLab Merge Alert 最初仅支持企业微信机器人。为了适配更多协作场景，本次迭代扩展为多渠道架构，新增钉钉机器人与自定义 HTTP Webhook 支持，同时保留企业微信的最佳体验。目标如下：

- 为主流国产 IM/协作工具提供开箱即用的通知能力；
- 保留单入口配置体验，自动识别 URL 所属渠道，管理员可手动覆盖；
- 内建钉钉安全策略与限流约束，减少触发平台风控的风险；
- 自定义 Webhook 用于记录和授权管理，消息仍由 GitLab 原生触发。

## 2. 渠道能力矩阵

| 渠道 | 对接方式 | 安全策略 / 限流 | 特殊说明 |
| --- | --- | --- | --- |
| 企业微信 (wechat) | 继续使用原有机器人 URL | 平台内部限制；支持手机号 @ | 表单无新增字段 |
| 钉钉 (dingtalk) | URL 自动识别 + 手动切换 | 20 次/分钟令牌桶；月度配额 5000 次（可配置） | 需填写加签 Secret、可选关键词；支持自定义 Header 预留 |
| 自定义 (custom) | 记录 HTTP 目标地址 | 平台不转发消息 | 提示在 GitLab 直接配置；支持记录额外 Header |

## 3. 数据结构与迁移

- `models.Webhook` 新增字段：`Type`、`SignatureMethod`、`Secret`、`SecurityKeywords`、`CustomHeaders`。
- 自定义 JSON 字段通过自定义类型 `StringList`、`StringMap` 序列化，兼容 SQLite。
- 迁移 `011_add_webhook_multi_channel.go`：
  1. `AutoMigrate` 扩展 `webhooks` 表；
  2. 根据 URL 调用 `DetectWebhookType` 回填历史数据；
  3. 初始化 `webhook_delivery_stats` 用于月度配额统计。
- Webhook 响应结构与创建/更新请求同步扩展，前端 API 与表单字段随之调整。

## 4. 发送链路重构

```
GitLab Webhook -> NotificationService -> MessageSenderFactory
                                   ├─ WeComSender (企业微信)
                                   ├─ DingTalkSender (钉钉)
                                   └─ CustomSender (记录提示)
```

- `MessageSender` 接口统一收敛三种渠道的发送逻辑。
- `DingTalkSender`
  - 使用 `pkg/ratelimit.TokenBucket` 实现 20 rpm 令牌桶；
  - 生成时间戳 + HMAC-SHA256 签名，按钉钉要求附加 `timestamp` 与 `sign`；
  - `webhook_delivery_stats` 记录月度调用次数，超额时抛出 `ErrDingTalkQuotaExceeded`；
  - 日志中打印请求体与限流命中，便于排障。
- `CustomSender` 仅记录提示，提醒管理员在 GitLab 直接维护该 Webhook。

## 5. 配置与运维

在 `config.example.yaml` 和 `config.Config` 中新增：

```yaml
notification:
  dingtalk:
    rate_limit_per_minute: 20
    monthly_quota: 5000
    request_timeout: 5s
    retry_attempts: 3
```

- 可通过环境变量 `GMA_NOTIFICATION__DINGTALK__*` 或 `notification.dingtalk.*` 覆盖。
- 当钉钉接口返回错误或限流命中，会在通知记录中写入 `error_message` 并保持历史。

## 6. 前端交互

- Webhook 管理页：
  - 新增「Webhook 类型」选择器与自动识别提示；
  - 钉钉类型显示 Secret 输入框与关键词选择（支持 `allow-create`）；
  - 自定义类型展示 Header 动态表单及 GitLab 手动配置提醒；
  - 列表中使用类型标签区分各渠道。
- 项目与账户授权页同步显示渠道标签，帮助管理员快速识别。

## 7. 测试与验证建议

- **数据库迁移**：在 SQLite 测试库上执行迁移，确认字段与统计表创建成功。
- **钉钉签名**：准备多组 Secret/时间戳用例，验证签名 URL 是否符合官方文档。
- **限流场景**：模拟 20rpm 以上流量，确认触发 `ErrDingTalkRateLimited` 且记录日志。
- **月度配额**：手工写入 `webhook_delivery_stats`，验证配额超限时提示可读。
- **前端**：分别创建三种类型 Webhook，检查表单字段显隐与测试按钮行为。自定义类型触发测试应返回“请在 GitLab 配置”提示。

## 8. 后续迭代方向

- 将钉钉配额统计对接 Redis 等外部存储，支持多实例部署。
- 抽象消息模板，实现 Markdown / 卡片消息针对不同渠道的最佳呈现。
- 扩展更多渠道（如 Slack、飞书），复用 Sender 接口实现。
