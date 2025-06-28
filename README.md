# GitLab Merge Alert (Golang版本)

GitLab Merge Request 通知服务的 Golang 重构版本，用于将 GitLab 的合并请求通知发送到企业微信群机器人。

## 功能特性

- ✅ 接收 GitLab Merge Request Webhook 事件
- ✅ 发送格式化通知到企业微信群
- ✅ 用户管理：GitLab 邮箱与企业微信手机号映射
- ✅ 项目管理：GitLab 项目与企业微信机器人关联
- ✅ Webhook 管理：企业微信机器人配置
- ✅ 通知历史记录和统计
- ✅ 现代化 Web 管理界面
- ✅ SQLite 数据库存储
- ✅ Docker 容器化部署

## 技术栈

- **后端**: Go 1.21 + Gin + GORM + SQLite
- **前端**: HTML5 + Bootstrap 5 + Vue.js 3
- **数据库**: SQLite
- **部署**: Docker

## 快速开始

### 本地开发

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd gitlab-merge-alert-go
   ```

2. **安装依赖**
   ```bash
   make deps
   ```

3. **初始化数据目录**
   ```bash
   make init
   ```

4. **配置应用**
   
   编辑 `config.yaml` 文件：
   ```yaml
   host: 0.0.0.0
   port: 1688
   environment: development
   log_level: info
   database_path: ./data/gitlab-merge-alert.db
   gitlab_url_prefix: https://your-gitlab.com
   default_webhook_url: https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your-key
   ```

5. **运行应用**
   ```bash
   make run
   ```

   访问 http://localhost:1688 查看管理界面

### Docker 部署

1. **构建镜像**
   ```bash
   make docker-build
   ```

2. **运行容器**
   ```bash
   make docker-run
   ```

3. **查看日志**
   ```bash
   make docker-logs
   ```

## 项目结构

```
gitlab-merge-alert-go/
├── cmd/server/           # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── models/          # 数据模型
│   ├── handlers/        # HTTP 处理器
│   ├── services/        # 业务逻辑
│   └── middleware/      # 中间件
├── web/
│   ├── static/         # 静态资源
│   └── templates/      # HTML 模板
├── pkg/                # 公共包
├── data/              # 数据文件目录
├── config.yaml        # 配置文件
├── Dockerfile         # Docker 构建文件
├── Makefile          # 构建脚本
└── README.md         # 说明文档
```

## API 接口

### GitLab Webhook
- `POST /api/v1/webhook/gitlab` - 接收 GitLab webhook

### 用户管理
- `GET /api/v1/users` - 获取用户列表
- `POST /api/v1/users` - 创建用户
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

### 项目管理
- `GET /api/v1/projects` - 获取项目列表
- `POST /api/v1/projects` - 创建项目
- `PUT /api/v1/projects/:id` - 更新项目
- `DELETE /api/v1/projects/:id` - 删除项目

### Webhook 管理
- `GET /api/v1/webhooks` - 获取 webhook 列表
- `POST /api/v1/webhooks` - 创建 webhook
- `PUT /api/v1/webhooks/:id` - 更新 webhook
- `DELETE /api/v1/webhooks/:id` - 删除 webhook

### 项目-Webhook 关联
- `POST /api/v1/projects/:project_id/webhooks/:webhook_id` - 关联项目和 webhook
- `DELETE /api/v1/projects/:project_id/webhooks/:webhook_id` - 取消关联

### 统计和通知
- `GET /api/v1/stats` - 获取统计信息
- `GET /api/v1/notifications` - 获取通知历史

## Web 界面

- `/` - 仪表板：统计信息和最近通知
- `/users` - 用户管理：GitLab 用户与企业微信映射
- `/projects` - 项目管理：GitLab 项目配置
- `/webhooks` - Webhook 管理：企业微信机器人配置

## 配置说明

### 环境变量

可以通过环境变量覆盖配置文件中的设置，环境变量前缀为 `GMA_`：

- `GMA_HOST` - 服务监听地址
- `GMA_PORT` - 服务端口
- `GMA_DATABASE_PATH` - SQLite 数据库文件路径
- `GMA_GITLAB_URL_PREFIX` - GitLab 服务器地址
- `GMA_LOG_LEVEL` - 日志级别

### GitLab Webhook 配置

在 GitLab 项目设置中添加 Webhook：

1. 进入项目 Settings → Webhooks
2. URL: `http://your-server:1688/api/v1/webhook/gitlab`
3. Trigger: `Merge request events`
4. 点击 "Add webhook"

### 企业微信机器人

1. 在企业微信群中添加机器人
2. 获取 Webhook URL
3. 在系统中添加 Webhook 配置
4. 将 Webhook 关联到相应项目

## 开发

### 添加新功能
```bash
# 格式化代码
make fmt

# 运行测试
make test

# 代码检查
make lint
```

### 数据库结构

- `users` - 用户表（邮箱与手机号映射）
- `projects` - 项目表（GitLab 项目信息）
- `webhooks` - Webhook 表（企业微信机器人配置）
- `project_webhooks` - 项目-Webhook 关联表
- `notifications` - 通知记录表

## 从 Python 版本迁移

1. 导出现有配置数据
2. 安装并配置新版本
3. 导入配置数据
4. 更新 GitLab Webhook URL
5. 测试通知功能

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库文件路径和权限
   - 确保数据目录存在

2. **GitLab Webhook 调用失败**
   - 检查网络连接和防火墙设置
   - 验证 Webhook URL 配置

3. **企业微信通知发送失败**
   - 检查 Webhook URL 是否正确
   - 验证机器人权限设置

### 日志查看

```bash
# Docker 环境
make docker-logs

# 本地环境
tail -f logs/app.log
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License