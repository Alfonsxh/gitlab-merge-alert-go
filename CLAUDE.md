# GitLab Merge Alert Go - 企业级合并请求通知服务

## 快速导航
- 🚀 [快速开始](#快速开始) - 5分钟上手
- 🏗️ [项目结构](#项目结构) - 理解项目架构  
- 📝 [开发命令](#开发命令) - 开发必备命令
- ⚠️ [注意事项](#注意事项) - 必须遵守的规则
- 🔧 [核心目录说明](#核心目录说明) - 目录功能详解
- 🌐 [API 文档](#api-文档) - 接口详细说明
- 🔄 [核心业务流程](#核心业务流程) - 通知处理机制

## 项目概述

这是一个 GitLab Merge Request 通知服务的 Go 语言重构版本，采用 B/S 架构，用于将 GitLab 的合并请求通知发送到企业微信群机器人。支持多项目、多 Webhook 管理，具备完整的用户认证和权限管理系统。

## 关键信息

| 项目 | 说明 |
|------|------|
| 主要语言 | Go 1.23 |
| 前端框架 | Vue.js 3.5 + TypeScript 5.8 |
| 默认端口 | 1688 |
| 主配置文件 | `config.local.yaml` (本地) / `config.yaml` (默认) |
| 日志位置 | `logs/app.log` |
| 数据库 | SQLite (`data/gitlab-merge-alert.db`) |
| API 前缀 | `/api/v1` |
| GitLab Webhook | `POST /api/v1/webhook/gitlab` |

## 技术栈

### 后端技术栈
- **编程语言**：Go 1.23.0
- **Web 框架**：Gin v1.10.0
- **ORM**：GORM v1.30.0
- **数据库**：SQLite (glebarez/sqlite v1.11.0)
- **认证**：JWT (golang-jwt/jwt/v5 v5.3.0)
- **配置管理**：Viper v1.20.1
- **日志**：Logrus v1.9.3
- **密码加密**：bcrypt (golang.org/x/crypto v0.39.0)

### 前端技术栈
- **框架**：Vue.js 3.5.17 + TypeScript 5.8.3
- **UI 框架**：Element Plus 2.9.2
- **路由**：Vue Router 4.5.1
- **状态管理**：Pinia 3.0.3
- **HTTP 客户端**：Axios 1.11.0
- **图表库**：ECharts 5.6.0 + vue-echarts 7.0.3
- **构建工具**：Vite 7.0.4
- **日期处理**：dayjs 1.11.10

## 项目结构

```
gitlab-merge-alert-go/
├── cmd/                      # 应用入口点
│   ├── server/              # Web 服务器入口 (main.go)
│   └── migrate/             # 数据库迁移工具
├── internal/                 # 内部包（业务逻辑）
│   ├── config/              # Viper 配置管理
│   ├── database/            # GORM 数据库连接
│   ├── models/              # 数据模型和请求/响应结构
│   ├── handlers/            # HTTP 处理器 (Gin 路由)
│   ├── services/            # 业务逻辑服务层
│   ├── middleware/          # HTTP 中间件（认证、权限、错误处理）
│   ├── migrations/          # 数据库迁移脚本
│   └── utils/               # 工具函数
├── pkg/                      # 可重用包
│   ├── auth/                # JWT 和密码处理
│   └── logger/              # 日志工具包
├── frontend/                 # Vue.js SPA 前端应用
│   ├── src/
│   │   ├── api/            # API 客户端封装
│   │   ├── views/          # 页面组件
│   │   ├── components/     # 通用组件
│   │   ├── stores/         # Pinia 状态管理
│   │   ├── router/         # 路由配置
│   │   └── utils/          # 工具函数
│   └── dist/               # 构建输出
├── data/                     # 数据目录 (SQLite 数据库)
├── logs/                     # 日志目录
├── config.example.yaml       # 配置文件示例
├── Makefile                  # 构建和开发命令
└── Dockerfile               # Docker 镜像定义
```

## 快速开始

### 安装依赖
```bash
# 后端依赖
make deps          # 安装 Go 依赖并整理模块

# 前端依赖
cd frontend && npm install
```

### 配置环境
```bash
# 复制配置文件模板
cp config.example.yaml config.local.yaml

# 编辑 config.local.yaml，填入真实配置
# 重点配置项：
# - gitlab_url: GitLab 服务器地址
# - gitlab_personal_access_token: GitLab 访问令牌
# - jwt_secret: JWT 密钥（用于用户认证）
```

### 开发环境
```bash
# 初始化数据和日志目录
make init

# 启动后端开发服务器 (localhost:1688)
make run

# 启动前端开发服务器 (另开终端)
cd frontend && npm run dev
```

### 构建项目
```bash
# 构建后端二进制文件
make build         # 输出到 bin/gitlab-merge-alert-go

# 构建前端
cd frontend && npm run build
```

### 运行测试
```bash
# 运行后端测试
make test

# 运行代码检查
make lint
```

## 开发命令

### 后端开发
```bash
make deps          # 安装依赖并整理模块
make init          # 初始化数据和日志目录  
make run           # 运行开发服务器 (localhost:1688)
make build         # 构建二进制文件到 bin/
make fmt           # 格式化代码
make lint          # 运行 golangci-lint 检查
make test          # 运行所有测试
make clean         # 删除构建文件和数据库
```

### 数据库管理
```bash
make migrate           # 运行数据库迁移
make migrate-status    # 查看迁移状态  
make migrate-rollback  # 回滚最后一个迁移
```

### Docker 部署
```bash
make docker-build    # 构建 Docker 镜像
make docker-run      # 运行容器
make docker-logs     # 查看容器日志
make docker-stop     # 停止并删除容器
make docker-restart  # 重启容器
```

### 前端开发
```bash
cd frontend
npm install        # 安装依赖
npm run dev        # 开发服务器 (支持热更新)
npm run build      # 生产构建
npm run preview    # 预览生产构建
```

## 架构决策记录（ADR）

### ADR-001: 采用 SQLite 作为数据库
- **决策**：使用 SQLite 替代 MySQL/PostgreSQL
- **原因**：项目规模适中，SQLite 足够满足需求，部署简单，无需额外数据库服务
- **影响**：简化部署流程，但不适合高并发场景

### ADR-002: 前后端分离架构
- **决策**：前端采用独立的 Vue.js SPA，通过 API 与后端通信
- **原因**：提高开发效率，前后端可独立部署和扩展
- **影响**：需要处理跨域问题，增加了认证复杂度

### ADR-003: JWT 认证机制
- **决策**：使用 JWT 进行用户认证，而非 Session
- **原因**：无状态认证，适合前后端分离架构，支持横向扩展
- **影响**：需要妥善处理 token 刷新和过期问题

### ADR-004: 多对多项目-Webhook 关联
- **决策**：项目与企业微信机器人采用多对多关系
- **原因**：一个项目可能需要通知多个群，一个群可能接收多个项目的通知
- **影响**：增加了配置灵活性，但管理复杂度略有提升

### ADR-005: 资源管理器模式
- **决策**：实现资源管理器（ResourceManager）进行细粒度权限控制
- **原因**：支持多租户场景，不同用户管理各自的资源
- **影响**：提高了系统的安全性和隔离性

## 核心业务流程

### GitLab Webhook 处理流程
```
GitLab 推送 Webhook
       ↓
HandleGitLabWebhook (handlers/webhook.go:13)
       ↓
验证事件类型（只处理 merge_request）
       ↓
ProcessMergeRequest (services/notification.go:25)
       ↓
查找项目配置（根据 GitLab 项目 ID）
       ↓
获取指派人信息（邮箱映射到企业微信手机号）
       ↓
格式化消息（Markdown 格式）
       ↓
发送到关联的企业微信机器人
       ↓
记录通知历史
```

### 用户认证流程
```
用户登录（用户名/密码）
       ↓
验证账号信息 (services/auth.go)
       ↓
生成 JWT Token（24小时有效期）
       ↓
前端存储 Token
       ↓
请求携带 Token（Authorization: Bearer xxx）
       ↓
中间件验证 Token (middleware/auth.go)
       ↓
提取用户信息和角色
       ↓
权限检查（管理员/普通用户）
```

## API 文档

### GitLab Webhook
- `POST /api/v1/webhook/gitlab` - 接收 GitLab 合并请求事件

### 认证相关
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出
- `GET /api/v1/auth/profile` - 获取当前用户信息
- `PUT /api/v1/auth/profile` - 更新用户资料
- `PUT /api/v1/auth/password` - 修改密码

### 用户管理（需要管理员权限）
- `GET /api/v1/users` - 获取用户列表
- `POST /api/v1/users` - 创建用户
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

### 项目管理
- `GET /api/v1/projects` - 获取项目列表
- `POST /api/v1/projects` - 创建项目
- `PUT /api/v1/projects/:id` - 更新项目
- `DELETE /api/v1/projects/:id` - 删除项目
- `POST /api/v1/projects/parse-url` - 解析项目 URL
- `POST /api/v1/projects/scan-group` - 扫描 GitLab 组项目
- `POST /api/v1/projects/batch-create` - 批量创建项目
- `POST /api/v1/projects/:id/sync-gitlab-webhook` - 同步 GitLab webhook
- `DELETE /api/v1/projects/:id/sync-gitlab-webhook` - 删除 GitLab webhook
- `GET /api/v1/projects/:id/gitlab-webhook-status` - 获取 GitLab webhook 状态

### Webhook 管理
- `GET /api/v1/webhooks` - 获取 webhook 列表
- `POST /api/v1/webhooks` - 创建 webhook
- `PUT /api/v1/webhooks/:id` - 更新 webhook
- `DELETE /api/v1/webhooks/:id` - 删除 webhook

### 项目-Webhook 关联
- `POST /api/v1/project-webhooks` - 关联项目和 webhook
- `DELETE /api/v1/project-webhooks/:project_id/:webhook_id` - 取消关联

### GitLab 集成
- `POST /api/v1/gitlab/test-connection` - 测试 GitLab 连接
- `GET /api/v1/gitlab/config` - 获取 GitLab 配置

### 统计和通知
- `GET /api/v1/stats` - 获取统计信息
- `GET /api/v1/notifications` - 获取通知历史

### 资源管理（多租户）
- `GET /api/v1/resource-manager/stats` - 获取资源统计
- `GET /api/v1/resource-manager/resources` - 获取用户资源列表

## 常见问题解决方案

### Q: GitLab Webhook 无法送达
**检查步骤**：
1. 确认 GitLab 能访问到服务地址
2. 检查 GitLab webhook 配置是否正确
3. 查看日志文件 `logs/app.log`

**解决方案**：
- 确保 `public_webhook_url` 配置正确
- 检查防火墙设置
- 使用 `curl` 测试 webhook 端点可达性

### Q: 企业微信通知发送失败
**检查步骤**：
1. 检查企业微信机器人 webhook URL 是否正确
2. 查看通知历史中的错误信息
3. 验证用户邮箱是否已映射手机号

**解决方案**：
- 在 Webhook 管理页面更新正确的机器人地址
- 在用户管理页面设置邮箱-手机号映射
- 检查企业微信机器人是否被禁用

### Q: 数据库迁移失败
**检查步骤**：
1. `make migrate-status` 查看当前迁移状态
2. 检查数据库文件权限

**解决方案**：
- `make migrate-rollback` 回滚失败的迁移
- 确保 `data/` 目录有写权限
- 必要时删除数据库重新初始化：`rm data/*.db && make init`

### Q: 前端构建失败
**检查步骤**：
1. 检查 Node.js 版本（需要 18+）
2. 清理 node_modules 重新安装

**解决方案**：
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

## 开发规范

- **代码风格**：使用 `gofmt` 和 `golangci-lint` 进行格式化和检查
- **提交规范**：使用语义化提交消息（feat/fix/docs/refactor/test/chore）
- **分支策略**：
  - `main` - 主分支，保持稳定
  - `feature/*` - 功能开发
  - `fix/*` - 缺陷修复
- **代码审查**：所有代码需要通过 lint 检查

## 注意事项

### 重要警告
- **绝不**将包含真实 token 的配置文件提交到版本控制
- **绝不**在生产环境使用默认的 JWT 密钥
- **绝不**直接修改数据库文件，使用迁移脚本

### 安全要求
- **必须**定期轮换 GitLab Personal Access Token
- **必须**使用 HTTPS 部署生产环境
- **必须**设置强密码策略

### 最佳实践
- **始终**使用 `config.local.yaml` 存储敏感配置
- **始终**在修改数据模型后创建迁移脚本
- **始终**运行 `make lint` 检查代码质量
- **始终**记录关键操作的日志

## 核心目录说明

### internal/ - 核心业务逻辑
包含所有内部业务逻辑，不对外暴露。详见 [internal/CLAUDE.md](internal/CLAUDE.md)

### frontend/ - Vue.js 前端应用  
独立的 SPA 应用，提供 Web 管理界面。详见 [frontend/CLAUDE.md](frontend/CLAUDE.md)

### cmd/ - 应用入口
包含 main 函数的可执行程序入口。详见 [cmd/CLAUDE.md](cmd/CLAUDE.md)

### pkg/ - 可重用包
可被其他项目引用的通用包。详见 [pkg/CLAUDE.md](pkg/CLAUDE.md)

## 相关文档
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM 文档](https://gorm.io/zh_CN/)
- [Vue.js 3 文档](https://cn.vuejs.org/)
- [Element Plus 文档](https://element-plus.org/zh-CN/)
- [企业微信机器人文档](https://developer.work.weixin.qq.com/document/path/91770)