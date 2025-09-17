# internal - 核心业务逻辑层

## 概述

`internal` 目录包含应用的所有内部业务逻辑实现，遵循 Go 语言约定，该目录下的代码不能被外部项目引用。这是整个应用的核心层，实现了从 HTTP 处理、业务逻辑到数据持久化的完整功能。

## 目录结构

```
internal/
├── config/              # 配置管理层
│   └── config.go       # Viper 配置加载和管理
├── database/            # 数据库连接层  
│   └── db.go          # GORM 数据库初始化和连接管理
├── handlers/            # HTTP 处理器层（控制器）
│   ├── handler.go      # Handler 基础结构和依赖注入
│   ├── webhook.go      # GitLab webhook 接收处理
│   ├── auth.go         # 认证相关接口
│   ├── user.go         # 用户管理接口
│   ├── project.go      # 项目管理接口
│   ├── webhook_mgmt.go # Webhook 管理接口
│   ├── dashboard.go    # 仪表板统计接口
│   ├── account.go      # 账号管理接口
│   ├── profile.go      # 个人资料管理
│   └── resource_manager.go # 资源管理器接口
├── middleware/          # HTTP 中间件
│   ├── auth.go         # JWT 认证中间件
│   ├── owner.go        # 资源所有权验证
│   ├── error_handler.go # 全局错误处理
│   └── validator.go    # 请求参数验证
├── models/              # 数据模型定义
│   ├── account.go      # 账号模型（用户认证）
│   ├── user.go         # 用户模型（GitLab 映射）
│   ├── project.go      # 项目模型
│   ├── webhook.go      # Webhook 模型和 GitLab 数据结构
│   ├── notification.go # 通知记录模型
│   ├── resource_manager.go # 资源管理模型
│   └── response.go     # 统一响应结构
├── services/            # 业务逻辑服务层
│   ├── interfaces.go   # 服务接口定义
│   ├── notification.go # 通知处理核心逻辑
│   ├── wechat.go      # 企业微信 API 集成
│   ├── gitlab.go      # GitLab API 集成
│   ├── auth.go        # 认证服务
│   └── resource_manager.go # 资源管理服务
├── migrations/          # 数据库迁移脚本
│   ├── migration.go    # 迁移基础结构
│   ├── registry.go     # 迁移注册中心
│   └── 00X_*.go      # 具体迁移脚本
└── utils/              # 工具函数
    └── function_helpers.go # 辅助函数
```

## 技术栈

- **核心技术**：Go 1.23, Gin Web Framework
- **依赖库**：
  - GORM v1.30.0 - ORM 框架
  - Viper v1.20.1 - 配置管理
  - JWT v5.3.0 - 认证令牌
  - Logrus v1.9.3 - 结构化日志
- **工具**：gofmt, golangci-lint

## 编码规范

- **命名规则**：
  - 类型名：PascalCase（如 `NotificationService`）
  - 函数名：PascalCase（导出）/ camelCase（内部）
  - 文件名：snake_case（如 `webhook_mgmt.go`）
- **代码风格**：遵循 Go 官方风格指南
- **文件组织**：按功能模块组织，每个文件专注单一职责

## 依赖关系

### 分层架构
```
handlers (HTTP 层)
    ↓ 依赖注入
services (业务逻辑层)
    ↓ 数据操作
models (数据模型层)
    ↓ ORM 映射
database (持久化层)
```

### 模块依赖关系

**handlers 依赖于**：
- services：业务逻辑处理
- models：请求/响应数据结构
- middleware：请求拦截和验证

**services 依赖于**：
- models：数据模型操作
- database：数据库连接
- pkg/logger：日志记录

**middleware 依赖于**：
- pkg/auth：JWT 验证
- models：用户权限模型

## 关键业务逻辑说明

### GitLab Webhook 处理
位置：`handlers/webhook.go:13` → `services/notification.go:25`

**实现特点**：
- 只处理 `merge_request` 类型且状态为 `opened` 的事件
- 自动记录所有 webhook 数据到日志便于调试
- 支持批量指派人通知（多个 assignee）
- 异步发送通知，避免阻塞响应

**处理流程**：
1. 接收 GitLab POST 请求
2. 解析 JSON 数据到 `GitLabWebhookData` 结构
3. 验证事件类型和状态
4. 查找对应项目配置
5. 映射用户邮箱到企业微信手机号
6. 格式化消息并发送
7. 记录通知历史

### JWT 认证机制
位置：`middleware/auth.go`

**实现特点**：
- Token 有效期 24 小时
- 支持角色权限（admin/user）
- 自动从 Authorization Header 提取 token
- 过期 token 返回特定错误码

**安全措施**：
- 使用 bcrypt 加密密码
- JWT 密钥从配置文件读取
- 支持 token 黑名单（待实现）

### 资源管理器
位置：`services/resource_manager.go`

**实现特点**：
- 多租户资源隔离
- 基于 owner_id 的权限控制
- 支持资源统计和配额管理
- 级联删除相关资源

### 数据库迁移系统
位置：`migrations/`

**实现特点**：
- 版本化迁移管理
- 支持迁移和回滚
- 自动记录迁移历史
- 迁移脚本自注册机制

**迁移命名规范**：
```
00X_description.go
其中 X 是递增的版本号
```

## 配置映射关系

| 配置文件 | 配置项 | 影响功能 |
|---------|--------|---------|
| config.yaml | `host`, `port` | 服务监听地址 |
| config.yaml | `database_path` | SQLite 数据库位置 |
| config.yaml | `gitlab_url` | GitLab API 地址 |
| config.yaml | `gitlab_personal_access_token` | GitLab API 认证 |
| config.yaml | `jwt_secret` | JWT 签名密钥 |
| config.yaml | `public_webhook_url` | 对外 webhook 地址 |

## 开发指南

### 添加新的 API 接口

#### 步骤说明
1. 在 `handlers/` 创建处理器文件
2. 在 `services/` 实现业务逻辑
3. 在 `models/` 定义数据结构
4. 在 `cmd/server/main.go` 注册路由

#### 代码示例
```go
// handlers/example.go
func (h *Handler) CreateExample(c *gin.Context) {
    var req models.CreateExampleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    result, err := h.exampleService.Create(&req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, result)
}
```

### 添加新的数据库表

#### 步骤说明
1. 在 `models/` 定义模型结构
2. 创建迁移脚本 `migrations/00X_create_table.go`
3. 在 `registry.go` 注册迁移
4. 运行 `make migrate`

#### 代码示例
```go
// models/example.go
type Example struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Name      string    `gorm:"not null" json:"name"`
    OwnerID   uint      `json:"owner_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// migrations/008_create_example_table.go
func init() {
    RegisterMigration(&Migration{
        Version: "008",
        Name:    "create_example_table",
        Up: func(db *gorm.DB) error {
            return db.AutoMigrate(&models.Example{})
        },
        Down: func(db *gorm.DB) error {
            return db.Migrator().DropTable(&models.Example{})
        },
    })
}
```

### 添加中间件

#### 代码示例
```go
// middleware/example.go
func ExampleMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        logger.GetLogger().Info("Request received")
        
        // 继续处理
        c.Next()
        
        // 后置处理
        logger.GetLogger().Info("Request completed")
    }
}
```

## 常用设计模式

### 依赖注入模式
```go
// handlers/handler.go
type Handler struct {
    db               *gorm.DB
    notifyService    services.NotificationService
    wechatService    services.WeChatService
    gitlabService    services.GitLabService
}

func NewHandler(db *gorm.DB, ...) *Handler {
    return &Handler{
        db:            db,
        notifyService: services.NewNotificationService(db, ...),
        // ...
    }
}
```

### 服务接口模式
```go
// services/interfaces.go
type NotificationService interface {
    ProcessMergeRequest(*models.GitLabWebhookData) error
    GetNotifications(ownerID uint, limit int) ([]models.Notification, error)
}

// 实现分离，便于测试和替换
type notificationService struct {
    db *gorm.DB
}
```

## 相关命令

```bash
# 运行服务
go run cmd/server/main.go

# 运行迁移
go run cmd/migrate/main.go up

# 格式化代码
gofmt -w internal/

# 运行测试
go test ./internal/...

# 检查代码质量
golangci-lint run ./internal/...
```

## 注意事项

- **重要**：所有数据库操作必须检查 `owner_id` 进行权限隔离
- **性能**：避免 N+1 查询，使用 `Preload` 预加载关联数据
- **安全**：敏感信息（如 token）不要记录到日志
- **错误处理**：始终返回有意义的错误信息，便于问题定位
- **事务处理**：批量操作使用数据库事务确保数据一致性