# 多 Webhook 渠道支持设计

## 背景
- 现有系统仅支持企业微信渠道，所有 webhook 记录缺乏类型标识，服务层以企业微信发送逻辑为核心。
- 新需求要求服务自动识别并同时支持企业微信、钉钉以及用户自定义的 HTTP webhook，以覆盖更多通知场景。
- 钉钉自定义机器人在 2024-2025 年间强化了安全策略、限流与商业化约束，需要在设计中专项应对。

## 目标与范围
- 创建统一的 webhook 模型与配置体验，能够保存渠道类型、认证信息、动态安全配置。
- 后端服务在处理 GitLab Merge Request 事件时，能够按渠道路由不同 Sender 并支持钉钉加签、限流、重试。
- 前端配置界面提供类型选择、字段动态展示以及关键安全提示，确保管理员理解钉钉与自定义渠道的差异。
- 文档更新，输出一份集中说明多渠道支持、配置步骤与实施方案的设计稿。

## 关键约束
- 钉钉机器人调用限流：单机器人 20 次/分钟；免费额度 5000 次/自然月，需要可配置的阈值与提醒。
- 安全策略要求至少启用关键词、加签或 IP 白名单之一；若选择加签，需在服务端按 HMAC-SHA256 规则生成签名。
- 自定义 webhook 不经过中转时需提示直接在 GitLab 配置，避免与内置渠道冲突。
- 老数据全部为企业微信，需要平滑迁移并保障通知链路不中断。

## 需求细化
- **渠道识别**：系统在新增/编辑 webhook 时应自动解析 URL，判定是 `wechat`、`dingtalk` 还是 `custom`，管理员可手动覆盖。
- **密钥管理**：钉钉 webhook 需存储 `secret`、签名算法、关键词列表等信息；企业微信沿用现有字段；自定义 webhook 可额外保存自定义 Header。
- **通知链路**：MR 事件经过通知服务后，根据 webhook 类型调用对应 Sender，发送成功与否需写入通知表以便审计。
- **限流控制**：钉钉 Sender 在发送前检查 Token Bucket，记录限流命中情况并返回友好提示。
- **配置入口**：前端在 Webhook 表单中分别展示钉钉安全策略说明、自定义 webhook 的 GitLab 配置提醒。

## 设计方案
### 数据模型
- `backend/internal/models/webhook.go`
  - 主表保留通用字段：`Type`、`Name`、`URL`、`Description`、`IsActive` 等；新增一对一 `WebhookSetting` 记录渠道配置（签名算法、Secret、关键词、自定义 Header）。
  - 提供 `DetectWebhookType(url string)` 辅助函数自动判定渠道类型。
  - 迁移脚本 `backend/internal/migrations/011_add_webhook_multi_channel.go`：创建 `webhook_settings` 与 `webhook_delivery_stats`，同时回填历史数据并推断渠道类型。

### 业务流程
- `backend/internal/services/notification.go`
  - 抽象 `MessageSender` 接口及 `MessageSenderFactory`，根据 webhook 类型返回 `WeComSender`、`DingTalkSender` 或 `GenericHTTPSender`。
  - 原先内嵌的企业微信发送逻辑迁移至 `WeComSender` 实现。
  - `DingTalkSender` 负责拼装加签 URL、注入限流器并处理钉钉特有错误码。
  - 自定义 webhook 支持额外 Header，并在系统返回说明「需在 GitLab 保持同样的 webhook 配置」。

### 配置与限流
- 新增 `pkg/ratelimit`，实现可注入的 Token Bucket，默认值通过环境变量 `GMA_DINGTALK_RPM`、`GMA_DINGTALK_MONTHLY_QUOTA` 配置。
- `config/config.go` 增加钉钉配置段（限流、重试次数、超时、超额提示内容）。
- 通知记录新增配额超限与安全策略缺失的错误信息，方便管理员排查。

### 文档交付
- `README_zh.md` 将加入「多渠道通知」章节，说明三类 webhook 的配置路径与差异。
- 新建 `docs/webhook-multi-channel.md`，沉淀架构背景、字段解释、运维建议及钉钉安全策略操作步骤。

## 前端方案
- `frontend/src/views/webhooks/Webhooks.vue`
  - 表单新增「Webhook 类型」字段，默认显示自动识别结果并允许手动修改。
  - 类型为钉钉时显示 `secret` 输入框、限流提示与安全策略说明；类型为自定义时提供自定义 Header 配置及 GitLab 手动配置指引。
  - 列表页展示渠道标签，支持依据类型筛选。
- `frontend/src/views/projects/Projects.vue`、`frontend/src/views/accounts/Accounts.vue`
  - 在列表、分配组件中同步展示 webhook 类型，确保项目与账户绑定时信息明确。

## 任务规划
| 模块 | 子任务 | 关联文件/位置 | 备注 |
| --- | --- | --- | --- |
| 数据模型 | 拆分 Webhook 主表与 `WebhookSetting` 子表，保留类型字段并实现自动识别 | `backend/internal/models/webhook.go` | 保证向后兼容 |
| 数据迁移 | 创建 `webhook_settings` 与统计表，迁移/回填历史数据 | `backend/internal/migrations/011_add_webhook_multi_channel.go` | 按「建表→回填→初始化统计」顺序 |
| 服务层 | 重构通知服务，引入 Sender 工厂与钉钉实现 | `backend/internal/services/notification.go`、`backend/internal/services/*.go` | 拆分 Sender，便于扩展 |
| 限流配置 | 搭建限流器与钉钉配置项 | `pkg/ratelimit/`、`config/config.go`、`.env.example` | RPM、月度配额均可覆盖 |
| 前端 | 表单与列表适配多类型字段 | `frontend/src/views/webhooks/Webhooks.vue`、`frontend/src/views/projects/Projects.vue`、`frontend/src/views/accounts/Accounts.vue` | 包含自动识别交互 |
| 文档 | 更新 README 并新增设计稿 | `README_zh.md`、`docs/webhook-multi-channel.md` | 输出渠道差异与配置指南 |
