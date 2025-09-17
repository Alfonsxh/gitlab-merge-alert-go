# GitLab Merge Alert Go - 企业级合并请求通知服务

## 快速导航
- 🚀 [快速开始](#快速开始) - 5分钟上手
- 🏗️ [项目结构](#项目结构) - 理解项目架构  
- 📝 [开发命令](#开发命令) - 开发必备命令
- 🔄 [核心业务流程](#核心业务流程) - 通知处理机制
- 🏛️ [架构决策](#架构决策记录adr) - 重要设计决策
- 🌐 [API 文档](#api-文档) - 接口详细说明
- ❓ [常见问题](#常见问题解决方案) - 故障排查指南
- ⚠️ [注意事项](#注意事项) - 必须遵守的规则
- 🔧 [核心目录说明](#核心目录说明) - 目录功能详解

## 项目概述

这是一个 GitLab Merge Request 通知服务的 Go 语言重构版本，采用 B/S 架构，用于将 GitLab 的合并请求通知发送到企业微信群机器人。支持多项目、多 Webhook 管理，具备完整的用户认证和权限管理系统。

### 核心功能
- **GitLab 集成**：自动接收并处理 Merge Request 事件
- **企业微信通知**：支持多机器人、多群组配置
- **项目管理**：批量导入、URL 解析、组扫描
- **用户系统**：JWT 认证、角色权限、资源隔离
- **Web 管理界面**：Vue.js 单页应用，响应式设计

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
| JWT 有效期 | 24小时 |
| 默认管理员 | admin / admin123 (首次启动自动创建) |

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

### 前置要求
- Go 1.23+
- Node.js 18+
- Make 工具
- Git

### 1. 克隆项目
```bash
git clone https://github.com/your-org/gitlab-merge-alert-go.git
cd gitlab-merge-alert-go
```

### 2. 安装依赖
```bash
# 安装所有依赖（后端 + 前端）
make deps && cd frontend && npm install && cd ..
```

### 3. 配置环境
```bash
# 复制配置文件模板
cp config.example.yaml config.local.yaml

# 编辑 config.local.yaml，填入真实配置
vim config.local.yaml
```

**必须配置的项目**：
```yaml
# GitLab 配置
gitlab_url: "https://gitlab.example.com"

# JWT 配置（生产环境必须更改）
jwt_secret: "your-super-secret-key-at-least-32-chars"

# 数据加密密钥（用于加密 GitLab Token 等敏感信息）
encryption_key: "a-32-characters-long-secret"

# 公开 Webhook URL（GitLab 回调地址）
public_webhook_url: "https://your-domain.com"
```

> 注意：GitLab Personal Access Token 需要在应用启动后，通过「账户管理」或「个人中心」页面配置，而非直接写入配置文件。

### 4. 初始化并启动
```bash
# 初始化数据目录和数据库
make init

# 启动开发服务器（前后端同时启动）
make dev

# 或分别启动
make run                    # 后端 (localhost:1688)
cd frontend && npm run dev  # 前端 (localhost:5173)
```

### 5. 访问系统
- 打开浏览器访问：http://localhost:1688
- 默认管理员账号：admin / admin123
- 首次登录后请立即修改密码

### 6. 配置 GitLab Webhook
1. 登录系统后，进入"项目管理"
2. 添加 GitLab 项目
3. 点击"同步 GitLab Webhook"
4. 或手动在 GitLab 项目设置中添加：
   - URL: `http://your-server:1688/api/v1/webhook/gitlab`
   - Secret Token: 留空
   - Trigger: Merge request events

### 构建生产版本
```bash
# 完整构建（前端 + 后端）
make build

# Docker 构建
make docker-build
make docker-run
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
```mermaid
GitLab 推送 Webhook
       ↓
HandleGitLabWebhook (handlers/webhook.go:13)
    ├─ 解析 JSON 数据
    ├─ 记录完整日志
    └─ 验证事件类型
       ↓
[只处理 merge_request + opened 状态]
       ↓
ProcessMergeRequest (services/notification.go:25)
    ├─ 查找项目配置（GitLab Project ID）
    ├─ 加载关联的 Webhooks
    └─ 提取指派人信息
       ↓
格式化企业微信消息
    ├─ Markdown 格式
    ├─ @指派人手机号
    └─ 包含 MR 链接
       ↓
批量发送通知
    ├─ 遍历所有关联的机器人
    ├─ 异步发送避免阻塞
    └─ 记录发送结果
       ↓
保存通知历史
    └─ 包含成功/失败状态
```

### 用户认证流程
```mermaid
用户登录请求
    ├─ 用户名
    └─ 密码
       ↓
验证账号 (services/auth.go)
    ├─ 查询账号表
    ├─ bcrypt 验证密码
    └─ 检查账号状态
       ↓
生成 JWT Token
    ├─ 包含: user_id, username, role
    ├─ 有效期: 24小时
    └─ 签名: HMAC-SHA256
       ↓
返回响应
    ├─ token 字符串
    ├─ expires_at 时间戳
    └─ 用户信息
       ↓
前端处理
    ├─ localStorage 存储 token
    ├─ Axios 拦截器自动附加
    └─ 401 响应自动跳转登录
```

### 项目-Webhook 关联流程
```mermaid
创建项目
    └─ 输入 GitLab URL 或 Project ID
       ↓
解析项目信息
    ├─ 调用 GitLab API
    └─ 获取项目名称、路径
       ↓
创建 Webhook
    └─ 输入企业微信机器人 URL
       ↓
建立关联 (多对多)
    ├─ 一个项目 → 多个机器人
    └─ 一个机器人 → 多个项目
       ↓
同步 GitLab (可选)
    ├─ 自动在 GitLab 创建 Webhook
    └─ 配置 Merge Request 事件
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
**症状**：GitLab 显示 webhook 发送失败，系统未收到通知

**检查步骤**：
```bash
# 1. 检查服务是否运行
curl http://localhost:1688/api/v1/health

# 2. 查看最近的日志
tail -f logs/app.log | grep webhook

# 3. 测试 webhook 端点
curl -X POST http://localhost:1688/api/v1/webhook/gitlab \
  -H "Content-Type: application/json" \
  -d '{"object_kind":"merge_request"}'
```

**解决方案**：
- 确保 `config.yaml` 中 `public_webhook_url` 配置为 GitLab 可访问的地址
- 检查防火墙是否开放 1688 端口：`sudo ufw allow 1688`
- 如果使用内网穿透，确保隧道正常运行
- GitLab 项目设置中检查 Webhook URL 是否正确

### Q: 企业微信通知发送失败
**症状**：日志显示通知发送失败，群里收不到消息

**检查步骤**：
```bash
# 1. 查看通知错误日志
grep "notification failed" logs/app.log

# 2. 测试机器人 URL
curl -X POST https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY \
  -H "Content-Type: application/json" \
  -d '{"msgtype":"text","text":{"content":"测试消息"}}'

# 3. 检查数据库中的通知记录
sqlite3 data/gitlab-merge-alert.db "SELECT * FROM notifications ORDER BY created_at DESC LIMIT 5;"
```

**解决方案**：
- 在 Webhook 管理页面更新正确的机器人 URL
- 确保机器人 URL 中的 key 参数正确
- 在用户管理页面设置 GitLab 邮箱到企业微信手机号的映射
- 检查企业微信机器人是否被管理员禁用或删除
- 确认消息格式符合企业微信 Markdown 要求

### Q: 数据库迁移失败
**症状**：启动时报数据库错误，表结构不匹配

**检查步骤**：
```bash
# 1. 查看迁移状态
make migrate-status

# 2. 检查数据库文件权限
ls -la data/gitlab-merge-alert.db

# 3. 查看迁移历史
sqlite3 data/gitlab-merge-alert.db "SELECT * FROM schema_migrations;"
```

**解决方案**：
```bash
# 方案1：回滚最后的迁移
make migrate-rollback

# 方案2：重置数据库（会丢失数据）
mv data/gitlab-merge-alert.db data/gitlab-merge-alert.db.bak
make init
make migrate

# 方案3：手动修复权限
chmod 666 data/gitlab-merge-alert.db
chown $(whoami) data/gitlab-merge-alert.db
```

### Q: 前端页面无法访问
**症状**：后端正常但前端页面 404 或空白

**检查步骤**：
```bash
# 1. 检查前端是否构建
ls -la frontend/dist/

# 2. 检查静态文件服务
curl http://localhost:1688/assets/index.js

# 3. 查看浏览器控制台错误
# F12 打开开发者工具查看 Console 和 Network
```

**解决方案**：
```bash
# 重新构建前端
cd frontend
npm install
npm run build
cd ..

# 重启服务
make build
make run

# 开发模式（前后端分离）
make run                    # 终端1：后端
cd frontend && npm run dev  # 终端2：前端
```

### Q: JWT Token 认证失败
**症状**：登录后仍然提示未授权，频繁跳转登录页

**检查步骤**：
```bash
# 1. 检查 JWT 配置
grep jwt_secret config.local.yaml

# 2. 查看认证错误日志
grep "JWT" logs/app.log

# 3. 检查浏览器 localStorage
# 浏览器控制台执行：localStorage.getItem('token')
```

**解决方案**：
- 确保 `jwt_secret` 在所有环境中保持一致
- 清除浏览器缓存和 localStorage：`localStorage.clear()`
- 检查系统时间是否正确（Token 有效期验证）
- 重新登录获取新 Token

### Q: 项目无法批量导入
**症状**：扫描 GitLab 组时无项目返回或报错

**检查步骤**：
```bash
# 1. 测试 GitLab 连接
curl -H "PRIVATE-TOKEN: YOUR_TOKEN" \
  https://gitlab.example.com/api/v4/projects

# 2. 检查 GitLab Token 权限
# Token 需要 api 或 read_api 权限

# 3. 查看扫描错误日志
grep "scan group" logs/app.log
```

**解决方案**：
- 重新生成 GitLab Personal Access Token，确保具备 `api` 或 `read_api` 权限
- 在系统「账户管理」或「个人中心」页面更新 GitLab Token
- 确认 GitLab URL 格式正确（不要带尾部斜杠）
- 检查用户是否有访问目标组的权限

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

## 性能优化建议

### 数据库优化
- **连接池配置**：调整 GORM 连接池大小
- **索引优化**：为常用查询字段添加索引
- **查询优化**：使用 Preload 避免 N+1 查询

### 前端优化
- **路由懒加载**：按需加载页面组件
- **图片优化**：使用 WebP 格式，启用懒加载
- **缓存策略**：合理设置静态资源缓存

### 后端优化
- **并发处理**：使用 goroutine 池处理通知发送
- **缓存机制**：Redis 缓存热点数据（可选）
- **限流保护**：添加 API 访问频率限制

## 监控和日志

### 日志管理
```bash
# 查看实时日志
tail -f logs/app.log

# 按级别过滤
grep "ERROR" logs/app.log

# 按时间查询
grep "2024-08-11" logs/app.log

# 日志轮转（使用 logrotate）
sudo nano /etc/logrotate.d/gitlab-merge-alert
```

### 健康检查
```bash
# API 健康检查
curl http://localhost:1688/api/v1/health

# 数据库连接检查
sqlite3 data/gitlab-merge-alert.db "SELECT datetime('now');"

# 进程监控
ps aux | grep gitlab-merge-alert
```

### 性能监控
- 使用 Prometheus + Grafana 监控系统指标
- 集成 pprof 进行性能分析
- 添加自定义业务指标

## 安全最佳实践

### 认证安全
- 使用强 JWT 密钥（至少 32 字符）
- 定期轮换 Token
- 实现 Token 刷新机制
- 添加登录失败限制

### 数据安全
- 敏感配置使用环境变量
- 数据库定期备份
- 日志脱敏处理
- HTTPS 传输加密

### 代码安全
- 定期更新依赖
- 使用安全扫描工具
- 代码审查流程
- 最小权限原则

## 部署建议

### Docker 部署
```bash
# 构建镜像
docker build -t gitlab-merge-alert:latest .

# 运行容器
docker run -d \
  --name gitlab-merge-alert \
  -p 1688:1688 \
  -v $(pwd)/data:/data \
  -v $(pwd)/logs:/logs \
  -v $(pwd)/config.yaml:/config.yaml \
  gitlab-merge-alert:latest
```

### Systemd 服务
```ini
[Unit]
Description=GitLab Merge Alert Service
After=network.target

[Service]
Type=simple
User=gitlab-alert
WorkingDirectory=/opt/gitlab-merge-alert
ExecStart=/opt/gitlab-merge-alert/bin/gitlab-merge-alert-go
Restart=always

[Install]
WantedBy=multi-user.target
```

### Nginx 反向代理
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://127.0.0.1:1688;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 相关文档
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM 文档](https://gorm.io/zh_CN/)
- [Vue.js 3 文档](https://cn.vuejs.org/)
- [Element Plus 文档](https://element-plus.org/zh-CN/)
- [企业微信机器人文档](https://developer.work.weixin.qq.com/document/path/91770)
- [GitLab API 文档](https://docs.gitlab.com/ee/api/)
