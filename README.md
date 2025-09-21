# GitLab Merge Alert (Golang版本)

GitLab Merge Request 通知服务的 Golang 重构版本，用于将 GitLab 的合并请求通知发送到企业微信群机器人。

## 功能特性

### 核心功能
- ✅ 接收 GitLab Merge Request Webhook 事件
- ✅ 发送格式化通知到企业微信群机器人
- ✅ 智能用户映射：GitLab 邮箱与企业微信手机号关联
- ✅ 多项目支持：一个项目可关联多个企业微信机器人
- ✅ 通知历史记录和详细统计分析

### 管理功能
- ✅ 现代化 Web 管理界面 (Bootstrap 5 + Vue.js 3)
- ✅ 用户管理：批量导入和管理用户映射关系
- ✅ 项目管理：支持 GitLab 项目 URL 解析和组扫描
- ✅ Webhook 管理：企业微信机器人配置和状态监控
- ✅ GitLab 集成：自动同步和管理 GitLab Webhook

### 技术特性
- ✅ SQLite 轻量级数据库存储
- ✅ 数据库版本化迁移管理
- ✅ Docker 容器化部署
- ✅ 环境变量和配置文件双重配置支持
- ✅ 结构化日志记录和错误跟踪

## 技术栈

- **后端**: Go 1.23 + Gin + GORM + SQLite
- **前端**: HTML5 + Bootstrap 5 + Vue.js 3
- **数据库**: SQLite
- **配置管理**: Viper
- **日志**: Logrus
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
   
   **⚠️ 安全警告**: 请勿将敏感信息提交到版本控制系统！
   
   **方法1: 本地配置文件 (推荐)**
   ```bash
   # 复制配置模板
   cp config.example.yaml config.local.yaml
   
   # 编辑 config.local.yaml，填入真实的敏感信息
   vim config.local.yaml
   ```
   
   **方法2: 环境变量**
   ```bash
   export GMA_GITLAB_URL="https://your-gitlab-server.com"
   export GMA_DEFAULT_WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR-REAL-KEY"
   ```
   
   **配置说明**:
   - `config.local.yaml`: 包含敏感信息，仅本地使用 (已在 .gitignore 中)
   - `config.yaml`: 示例配置，安全可提交
   - `config.example.yaml`: 配置模板，参考使用

5. **运行应用**
   ```bash
   make run
   ```

   访问 http://localhost:1688 查看管理界面

### 管理员首次初始化

首次启动时后台会自动创建占位的管理员账号，并在日志中输出一次性 "Admin setup token"。使用流程如下：

1. 在服务启动日志或容器日志中复制最新的 Setup Token。
2. 打开前端页面的 `/setup-admin`（或在登录页收到提示后跳转），输入 token、管理员邮箱以及新密码完成初始化。
3. 初始化成功后，使用新密码登录，原 token 会立即失效。若需要重新生成，可重启服务以获取新的 token。

### Docker 部署

#### 方法1: 使用 Makefile (推荐)

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

#### 方法2: 手动 Docker 部署

1. **构建 Docker 镜像**
   ```bash
   # 构建镜像，标记为 gitlab-merge-alert:latest
   docker build -t gitlab-merge-alert:latest .
   ```

2. **准备数据目录**
   ```bash
   # 创建持久化数据目录
   mkdir -p ./docker-data
   chmod 755 ./docker-data
   ```

3. **准备配置文件**
   
   **方法A: 使用配置文件**
   ```bash
   # 复制配置模板到数据目录
   cp config.example.yaml ./docker-data/config.yaml
   
   # 编辑配置文件，填入真实的敏感信息
   vim ./docker-data/config.yaml
   ```
   
   **方法B: 使用环境变量文件**
   ```bash
   # 创建环境变量文件
   cat > ./docker-data/.env << EOF
   GMA_GITLAB_URL=https://your-gitlab-server.com
   GMA_DEFAULT_WEBHOOK_URL=https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR-REAL-KEY
   GMA_DATABASE_PATH=/data/app.db
   GMA_HOST=0.0.0.0
   GMA_PORT=1688
   EOF
   ```

4. **运行容器**
   
   **方法A: 使用配置文件**
   ```bash
   docker run -d \
     --name gitlab-merge-alert \
     --restart unless-stopped \
     -p 1688:1688 \
     -v $(pwd)/docker-data:/data \
     -e GMA_DATABASE_PATH=/data/app.db \
     gitlab-merge-alert:latest
   ```
   
   **方法B: 使用环境变量**
   ```bash
   docker run -d \
     --name gitlab-merge-alert \
     --restart unless-stopped \
     -p 1688:1688 \
     -v $(pwd)/docker-data:/data \
     --env-file ./docker-data/.env \
     gitlab-merge-alert:latest
   ```

5. **验证部署**
   ```bash
   # 检查容器状态
   docker ps | grep gitlab-merge-alert
   
   # 查看应用日志
   docker logs gitlab-merge-alert
   
   # 测试应用是否正常启动
   curl http://localhost:1688/
   ```

6. **管理容器**
   ```bash
   # 停止容器
   docker stop gitlab-merge-alert
   
   # 启动容器
   docker start gitlab-merge-alert
   
   # 重启容器
   docker restart gitlab-merge-alert
   
   # 删除容器（注意：数据会保留在 docker-data 目录中）
   docker rm gitlab-merge-alert
   
   # 进入容器调试
   docker exec -it gitlab-merge-alert sh
   
   # 实时查看日志
   docker logs -f gitlab-merge-alert
   ```

#### Docker Compose 部署（可选）

创建 `docker-compose.yml` 文件：
```yaml
version: '3.8'

services:
  gitlab-merge-alert:
    build: .
    container_name: gitlab-merge-alert
    restart: unless-stopped
    ports:
      - "1688:1688"
    volumes:
      - ./docker-data:/data
    environment:
      - GMA_DATABASE_PATH=/data/app.db
      - GMA_HOST=0.0.0.0
      - GMA_PORT=1688
      # 添加其他环境变量
      - GMA_GITLAB_URL=${GITLAB_URL}
      - GMA_DEFAULT_WEBHOOK_URL=${WEBHOOK_URL}
```

使用 Docker Compose：
```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### 生产环境部署建议

1. **安全配置**
   ```bash
   # 设置适当的文件权限
   chmod 600 ./docker-data/config.yaml
   chmod 600 ./docker-data/.env
   
   # 使用非 root 用户运行容器
   docker run -d \
     --name gitlab-merge-alert \
     --restart unless-stopped \
     -p 1688:1688 \
     -v $(pwd)/docker-data:/data \
     --user $(id -u):$(id -g) \
     --env-file ./docker-data/.env \
     gitlab-merge-alert:latest
   ```

2. **反向代理配置（Nginx 示例）**
   ```nginx
   server {
       listen 80;
       server_name your-domain.com;
       
       location / {
           proxy_pass http://localhost:1688;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   ```

3. **监控和日志**
   ```bash
   # 设置日志轮转
   docker run -d \
     --name gitlab-merge-alert \
     --restart unless-stopped \
     -p 1688:1688 \
     -v $(pwd)/docker-data:/data \
     --log-driver json-file \
     --log-opt max-size=10m \
     --log-opt max-file=3 \
     --env-file ./docker-data/.env \
     gitlab-merge-alert:latest
   ```

## 架构设计

本项目采用清晰的分层架构：

- **Handler 层**: 处理 HTTP 请求和路由，负责请求验证和响应格式化
- **Service 层**: 实现核心业务逻辑，包括通知处理、GitLab 集成、企业微信消息发送
- **Model 层**: 定义数据结构和数据库模型，使用 GORM 进行 ORM 映射
- **配置层**: 使用 Viper 实现灵活的配置管理，支持文件和环境变量
- **迁移层**: 版本化数据库迁移管理，确保数据库结构一致性

## 项目结构

```
gitlab-merge-alert-go/
├── cmd/
│   ├── server/          # 应用入口点 (main.go)
│   └── migrate/         # 数据库迁移工具
├── internal/
│   ├── config/          # Viper 配置管理
│   ├── database/        # GORM 数据库连接和迁移
│   ├── models/          # 数据库模型和请求/响应结构
│   ├── handlers/        # Gin HTTP 处理器 (路由层)
│   ├── services/        # 业务逻辑服务层
│   ├── migrations/      # 数据库迁移脚本
│   └── middleware/      # HTTP 中间件 (暂未使用)
├── web/
│   ├── static/         # CSS, JS, 字体文件
│   └── templates/      # HTML 模板文件
├── pkg/logger/         # 日志工具包
├── data/              # 数据文件目录
├── config.example.yaml # 配置模板
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
- `POST /api/v1/projects/parse-url` - 解析 GitLab 项目 URL
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

### GitLab 集成
- `POST /api/v1/gitlab/test-connection` - 测试 GitLab 连接
- `GET /api/v1/gitlab/config` - 获取 GitLab 配置

### 项目-Webhook 关联
- `POST /api/v1/project-webhooks` - 关联项目和 webhook
- `DELETE /api/v1/project-webhooks/:project_id/:webhook_id` - 取消关联

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

**基础配置**:
- `GMA_HOST` - 服务监听地址 (默认: 0.0.0.0)
- `GMA_PORT` - 服务端口 (默认: 1688)
- `GMA_DATABASE_PATH` - SQLite 数据库文件路径
- `GMA_LOG_LEVEL` - 日志级别 (debug/info/warn/error)
- `GMA_ENVIRONMENT` - 运行环境 (development/production)

**GitLab 集成** (敏感信息):
- `GMA_GITLAB_URL` - GitLab 服务器地址
- `GMA_DEFAULT_WEBHOOK_URL` - 默认企业微信机器人地址
- `GMA_ENCRYPTION_KEY` - 敏感数据加密密钥（若未设置则回退到 `JWT_SECRET`）

> 注意：GitLab Personal Access Token 需要在应用的账户管理或个人中心页面中配置，不再通过环境变量或配置文件提供。

**服务配置**:
- `GMA_REDIRECT_SERVER_URL` - 重定向服务器地址
- `GMA_PUBLIC_WEBHOOK_URL` - 公共 Webhook 地址

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

### 开发命令
```bash
# 安装依赖
make deps

# 初始化数据和日志目录
make init

# 运行开发服务器
make run

# 构建二进制文件
make build

# 格式化代码
make fmt

# 运行测试
make test

# 代码检查
make lint
```

### 数据库迁移
```bash
# 运行数据库迁移
make migrate

# 查看迁移状态
make migrate-status

# 回滚最后一个迁移
make migrate-rollback
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
