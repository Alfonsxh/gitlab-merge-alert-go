# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个 GitLab Merge Request 通知服务的 Go 语言重构版本，用于将 GitLab 的合并请求通知发送到企业微信群机器人。

**技术栈**:
- **后端**: Go 1.21 + Gin + GORM + SQLite  
- **前端**: HTML5 + Bootstrap 5 + Vue.js 3
- **数据库**: SQLite
- **部署**: Docker

## 开发命令

### 本地开发
```bash
make deps          # 安装依赖并整理模块
make init          # 初始化数据和日志目录
make run           # 运行开发服务器 (localhost:1688)
make build         # 构建二进制文件到 bin/
```

### 代码质量
```bash
make fmt           # 格式化代码
make lint          # 运行 golangci-lint 检查
make test          # 运行所有测试
```

### Docker 部署
```bash
make docker-build    # 构建 Docker 镜像
make docker-run      # 运行容器
make docker-logs     # 查看容器日志
make docker-stop     # 停止并删除容器
make docker-restart  # 重启容器
```

### 清理
```bash
make clean         # 删除构建文件和数据库
```

## 代码架构

### 项目结构
```
cmd/server/           # 应用入口点
internal/
├── config/          # Viper 配置管理
├── database/        # GORM 数据库连接和迁移
├── models/          # 数据库模型和请求/响应结构
├── handlers/        # Gin HTTP 处理器 (路由层)
├── services/        # 业务逻辑服务层
└── middleware/      # HTTP 中间件
pkg/logger/          # 日志工具包
web/                 # 静态资源和 HTML 模板
```

### 核心组件

**Handler 层** (`internal/handlers/`):
- 负责 HTTP 请求处理和路由
- 主要文件: `handler.go` (基础结构), `webhook.go` (GitLab webhook), `user.go`, `project.go`, `webhook_mgmt.go`, `dashboard.go`

**Service 层** (`internal/services/`):
- `notification.go`: 处理合并请求通知的核心业务逻辑
- `wechat.go`: 企业微信 API 集成和消息格式化
- `gitlab.go`: GitLab API 集成

**数据模型** (`internal/models/`):
- `user.go`: 用户模型 (GitLab 邮箱 ↔ 企业微信手机号映射)
- `project.go`: 项目模型 (GitLab 项目配置)
- `webhook.go`: Webhook 模型 (企业微信机器人配置)
- `notification.go`: 通知记录模型

### 配置管理

**⚠️ 安全注意事项**：
- **绝不** 将包含敏感信息的配置文件提交到版本控制系统
- 生产环境必须使用环境变量或本地配置文件
- 定期轮换API密钥和Webhook密钥

**配置文件优先级**:
1. `config.local.yaml` (本地敏感配置，已在 .gitignore 中)
2. `config.yaml` (示例配置，安全可提交)  
3. 环境变量 (前缀: `GMA_`)

**配置方法**:

**方法1: 本地配置文件**
```bash
# 复制模板并填入真实值
cp config.example.yaml config.local.yaml
# 编辑 config.local.yaml，填入真实的敏感信息
```

**方法2: 环境变量**
```bash
export GMA_GITLAB_URL="https://your-gitlab-server.com"
export GMA_DEFAULT_WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR-KEY"
export GMA_GITLAB_PERSONAL_ACCESS_TOKEN="your-gitlab-token"
```

**主要配置项**:
- `host/port`: 服务监听地址
- `database_path`: SQLite 数据库路径
- `gitlab_url`: GitLab 服务器地址 (**敏感**)
- `default_webhook_url`: 企业微信机器人地址 (**敏感**)
- `gitlab_personal_access_token`: GitLab 访问令牌 (**敏感**)

### 数据库

使用 SQLite 作为数据库，通过 GORM 进行操作：
- **自动迁移**: 启动时运行 `database.Migrate()`
- **核心表**: users, projects, webhooks, project_webhooks (多对多关联), notifications

### API 端点

**GitLab Webhook**: `POST /api/v1/webhook/gitlab`
**管理 API**:
- Users: `/api/v1/users` (CRUD)
- Projects: `/api/v1/projects` (CRUD) 
- Webhooks: `/api/v1/webhooks` (CRUD)
- Project-Webhook 关联: `/api/v1/project-webhooks`
- 统计数据: `/api/v1/stats`, `/api/v1/notifications`

**Web 界面**:
- `/`: 仪表板
- `/users`, `/projects`, `/webhooks`: 管理页面

### 业务流程

1. **GitLab Webhook** → `handlers.HandleGitLabWebhook`
2. **通知处理** → `services.NotificationService.ProcessMergeRequest`
3. **用户映射** → 查找 assignee 邮箱对应的手机号
4. **消息发送** → `services.WeChatService.SendMessage`
5. **记录保存** → 保存通知历史到数据库

### 开发注意事项

- 只处理 `merge_request` 类型且状态为 `opened` 的 webhook 事件
- 项目与企业微信机器人为多对多关系，一个项目可以关联多个机器人
- 用户的 GitLab 邮箱与企业微信手机号需要预先在系统中建立映射
- 所有错误都会记录到通知历史中，便于故障排除
- 使用 Gin 的日志中间件和自定义 logger 进行日志记录

### Docker 部署

- 多阶段构建，最终镜像基于 Alpine Linux
- 数据目录 `/data` 需要挂载持久化存储
- 默认暴露端口 1688
- 环境变量 `GMA_DATABASE_PATH` 指向容器内数据库路径