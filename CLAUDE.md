# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

GitLab Merge Request 企业微信通知服务 - 一个基于 Go 开发的 Webhook 服务，用于将 GitLab MR 事件通知发送到企业微信群。

**技术栈**:
- 后端: Go 1.23 + Gin + GORM + SQLite
- 前端: Vue 3 + TypeScript + Element Plus + Vite
- 部署: Docker + Docker Compose

## 开发命令

### 后端开发
```bash
# 在项目根目录运行
go mod download                   # 安装依赖
go run ./cmd/main.go              # 运行服务（默认端口 1688）
go test ./...                     # 运行测试
go fmt ./...                      # 格式化代码
```

### 前端开发
```bash
# 在 frontend 目录下运行
npm install                       # 安装依赖
npm run dev                       # 开发模式（端口 5173）
npm run build                     # 构建生产版本
```

### 构建和部署
```bash
# 在项目根目录运行
make init                         # 初始化数据目录
make deps                         # 安装前后端依赖
make dev                          # 同时启动前后端开发服务器
make build                        # 构建带嵌入前端的二进制文件
make docker-build                 # 构建 Docker 镜像
make docker-run                   # 运行 Docker 容器
```

### 数据库迁移
```bash
# 在项目根目录运行
make migrate                      # 运行数据库迁移
make migrate-status               # 查看迁移状态
make migrate-rollback             # 回滚最后一个迁移
```

## 核心架构

### 分层架构
```
cmd/main.go                       # 应用入口
    ↓
internal/handlers/                # HTTP 处理层
    ↓
internal/services/                # 业务逻辑层
    ↓
internal/models/                  # 数据模型层
    ↓
internal/database/                # 数据持久化层
```

### 关键模块

**认证系统** (`internal/handlers/auth.go:41`, `internal/services/auth.go:23`)
- JWT 认证机制，Token 有效期 24 小时
- 支持管理员首次初始化流程（Setup Token）
- 基于 owner_id 的多租户资源隔离

**Webhook 处理** (`internal/handlers/webhook.go:13`, `internal/services/notification.go:25`)
- 接收 GitLab MR 事件，仅处理 `opened` 状态
- 自动映射 GitLab 用户邮箱到企业微信手机号
- 异步发送通知，避免阻塞响应

**GitLab 集成** (`internal/services/gitlab.go:20`)
- 自动管理 GitLab Project Webhooks
- 支持 GitLab Personal Access Token 认证
- 批量导入 GitLab 项目和用户

**资源管理** (`internal/middleware/owner.go:15`, `internal/services/resource_manager.go:19`)
- 所有资源基于 owner_id 隔离
- 支持资源统计和配额管理
- 级联删除相关资源

**数据库迁移** (`internal/migrations/registry.go:7`)
- 版本化迁移管理系统
- 自动注册和执行迁移脚本
- 支持迁移回滚

### 前端架构

**技术栈**: Vue 3 + TypeScript + Element Plus + Pinia

**核心模块**:
- `src/stores/`: Pinia 状态管理
- `src/api/`: Axios API 客户端
- `src/views/`: 页面组件
- `src/router/`: Vue Router 路由配置

## 配置管理

配置优先级: 环境变量 > config.local.yaml > config.yaml

**关键配置**:
```yaml
host: 0.0.0.0                     # 服务监听地址
port: 1688                        # 服务端口
database_path: ./data/gitlab-merge-alert.db      # SQLite 数据库路径
gitlab_url: https://gitlab.com    # GitLab 服务器地址
jwt_secret: your-secret-key       # JWT 签名密钥
encryption_key: your-encrypt-key  # 敏感数据加密密钥
```

环境变量前缀: `GMA_` (例如 `GMA_PORT=1688`)

## API 接口

### 认证接口
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/register` - 注册
- `POST /api/v1/auth/logout` - 登出
- `POST /api/v1/auth/setup-admin` - 初始化管理员

### 核心接口
- `POST /api/v1/webhook/gitlab` - 接收 GitLab Webhook
- `GET/POST/PUT/DELETE /api/v1/users` - 用户管理
- `GET/POST/PUT/DELETE /api/v1/projects` - 项目管理
- `GET/POST/PUT/DELETE /api/v1/webhooks` - Webhook 管理
- `GET/POST/PUT/DELETE /api/v1/accounts` - 账号管理

### GitLab 集成
- `POST /api/v1/projects/:id/sync-gitlab-webhook` - 同步 GitLab Webhook
- `GET /api/v1/projects/:id/gitlab-webhook-status` - 获取 Webhook 状态
- `POST /api/v1/gitlab/test-connection` - 测试 GitLab 连接

## 数据模型

**核心表结构**:
- `accounts` - 用户认证账号（登录用）
- `users` - GitLab 用户映射（邮箱→手机号）
- `projects` - GitLab 项目配置
- `webhooks` - 企业微信机器人配置
- `project_webhooks` - 项目与机器人关联
- `notifications` - 通知历史记录

**关系说明**:
- 一个 Account 可管理多个 Project/User/Webhook（owner_id）
- 一个 Project 可关联多个 Webhook（多对多）
- 通知基于 User 邮箱映射发送

## 安全措施

- 密码使用 bcrypt 加密存储
- GitLab Token 使用 AES 加密存储
- JWT 认证，支持 Token 过期
- 基于 owner_id 的资源隔离
- 敏感信息不记录到日志

## 故障排查

**常见问题**:
1. 数据库连接失败 - 检查 `database_path` 和文件权限
2. GitLab Webhook 失败 - 验证网络连接和 Token 权限
3. 企业微信通知失败 - 检查机器人 URL 和消息格式

**日志位置**:
- 本地开发: `./logs/app.log`
- Docker: `docker logs gitlab-merge-alert`

**调试技巧**:
- 所有 Webhook 数据记录在日志中
- 使用 `GMA_LOG_LEVEL=debug` 启用详细日志
- 检查 `/api/v1/stats` 查看系统状态
