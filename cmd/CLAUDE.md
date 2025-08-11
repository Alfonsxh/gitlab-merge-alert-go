# cmd - 应用程序入口

## 概述

`cmd` 目录包含应用程序的可执行入口点，遵循 Go 项目的标准布局。每个子目录代表一个独立的可执行程序，包含 `main` 函数。目前包含两个程序：Web 服务器和数据库迁移工具。

## 目录结构

```
cmd/
├── server/              # Web 服务器应用
│   └── main.go         # HTTP 服务器入口，路由注册，依赖注入
└── migrate/            # 数据库迁移工具
    └── main.go         # 迁移命令行工具，支持 up/down/status
```

## 技术栈

- **核心技术**：Go 1.23 标准库
- **依赖库**：
  - Gin v1.10.0 - Web 框架（server）
  - GORM v1.30.0 - ORM（migrate）
  - Viper v1.20.1 - 配置管理
  - Logrus v1.9.3 - 日志系统

## 编码规范

- **命名规则**：
  - 目录名：小写，描述功能（如 `server`, `migrate`）
  - 文件名：始终为 `main.go`
  - 函数名：`main()` 作为入口点
- **代码风格**：保持 main 函数简洁，复杂逻辑抽取到 internal
- **文件组织**：每个命令一个目录，避免代码混杂

## 依赖关系

**server 依赖于**：
- internal/config：配置加载
- internal/database：数据库连接
- internal/handlers：HTTP 处理器
- internal/services：业务服务
- internal/middleware：中间件
- pkg/logger：日志工具

**migrate 依赖于**：
- internal/config：配置加载
- internal/database：数据库连接
- internal/migrations：迁移脚本
- pkg/logger：日志工具

## 关键业务逻辑说明

### Web 服务器启动流程
位置：`server/main.go`

**启动顺序**：
1. 初始化日志系统
2. 加载配置文件（config.yaml/config.local.yaml）
3. 连接数据库（SQLite）
4. 自动运行数据库迁移
5. 初始化服务层（依赖注入）
6. 注册 HTTP 路由
7. 配置中间件（CORS、日志、错误处理）
8. 启动 HTTP 服务器（默认端口 1688）

**路由注册结构**：
```
/                          # 静态文件服务（前端）
/api/v1
  ├── /webhook/gitlab      # GitLab webhook 接收（无需认证）
  ├── /auth/*             # 认证相关接口
  ├── /gitlab/*           # GitLab 集成接口
  └── [需认证]
      ├── /users/*        # 用户管理
      ├── /projects/*     # 项目管理
      ├── /webhooks/*     # Webhook 管理
      ├── /stats          # 统计信息
      └── /notifications  # 通知历史
```

### 数据库迁移工具
位置：`migrate/main.go`

**支持的命令**：
- `up` - 执行所有未运行的迁移
- `down` - 回滚最后一个迁移
- `status` - 查看迁移状态
- `version` - 查看当前版本

**执行流程**：
1. 解析命令行参数
2. 加载配置文件
3. 连接数据库
4. 执行相应的迁移操作
5. 记录迁移历史

## 配置映射关系

| 配置文件 | 配置项 | 影响功能 |
|---------|--------|---------|
| config.yaml | `host` | 服务监听地址 |
| config.yaml | `port` | 服务监听端口 |
| config.yaml | `database_path` | 数据库文件路径 |
| config.yaml | `log_level` | 日志级别 |
| config.yaml | `jwt_secret` | JWT 签名密钥 |

## 开发指南

### 添加新的命令行工具

#### 步骤说明
1. 在 `cmd/` 创建新目录
2. 创建 `main.go` 文件
3. 实现命令逻辑
4. 更新 Makefile 添加构建目标

#### 代码示例
```go
// cmd/example/main.go
package main

import (
    "flag"
    "fmt"
    "log"
    
    "gitlab-merge-alert-go/internal/config"
    "gitlab-merge-alert-go/internal/database"
    "gitlab-merge-alert-go/pkg/logger"
)

func main() {
    // 解析命令行参数
    var configPath string
    flag.StringVar(&configPath, "config", "config.yaml", "配置文件路径")
    flag.Parse()
    
    // 初始化日志
    logger.InitLogger()
    
    // 加载配置
    cfg, err := config.Load(configPath)
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    // 连接数据库
    db, err := database.Connect(cfg.DatabasePath)
    if err != nil {
        log.Fatal("连接数据库失败:", err)
    }
    
    // 执行业务逻辑
    fmt.Println("Example command executed")
}
```

### 修改服务器路由

#### 代码位置
`server/main.go` - setupRoutes 函数

#### 添加新路由示例
```go
func setupRoutes(router *gin.Engine, handler *handlers.Handler) {
    api := router.Group("/api/v1")
    
    // 添加新的公开路由
    api.POST("/public/example", handler.PublicExample)
    
    // 添加需要认证的路由
    protected := api.Group("")
    protected.Use(authMiddleware.RequireAuth())
    {
        // 新的受保护路由
        protected.GET("/example", handler.GetExample)
        protected.POST("/example", handler.CreateExample)
        
        // 需要管理员权限的路由
        admin := protected.Group("")
        admin.Use(authMiddleware.RequireAdmin())
        {
            admin.DELETE("/example/:id", handler.DeleteExample)
        }
    }
}
```

### 优化服务启动

#### 添加优雅关闭
```go
// server/main.go
func main() {
    // ... 初始化代码
    
    srv := &http.Server{
        Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Handler: router,
    }
    
    // 启动服务器
    go func() {
        if err := srv.ListenAndServe(); err != nil {
            log.Printf("服务器错误: %v", err)
        }
    }()
    
    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("服务器关闭失败:", err)
    }
    
    log.Println("服务器已优雅关闭")
}
```

## 常用设计模式

### 依赖注入模式
```go
// server/main.go
func initializeServices(db *gorm.DB, cfg *config.Config) *handlers.Handler {
    // 创建服务实例
    wechatService := services.NewWeChatService()
    gitlabService := services.NewGitLabService(cfg.GitLabURL, cfg.GitLabToken)
    notifyService := services.NewNotificationService(db, wechatService)
    authService := services.NewAuthService(db, cfg.JWTSecret)
    
    // 注入到处理器
    return handlers.NewHandler(
        db,
        notifyService,
        wechatService,
        gitlabService,
        authService,
    )
}
```

### 命令模式（迁移工具）
```go
// migrate/main.go
type Command interface {
    Execute() error
}

type UpCommand struct {
    db *gorm.DB
}

func (c *UpCommand) Execute() error {
    return migrations.RunUp(c.db)
}

// 使用
commands := map[string]Command{
    "up":     &UpCommand{db: db},
    "down":   &DownCommand{db: db},
    "status": &StatusCommand{db: db},
}

if cmd, ok := commands[os.Args[1]]; ok {
    if err := cmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

## 相关命令

```bash
# 运行服务器
go run cmd/server/main.go

# 指定配置文件运行
go run cmd/server/main.go -config=config.local.yaml

# 构建服务器二进制
go build -o bin/server cmd/server/main.go

# 运行数据库迁移
go run cmd/migrate/main.go up

# 查看迁移状态
go run cmd/migrate/main.go status

# 回滚迁移
go run cmd/migrate/main.go down

# 使用 Makefile
make run        # 运行服务器
make migrate    # 运行迁移
```

## 注意事项

- **重要**：main 函数应保持简洁，复杂逻辑放到 internal
- **性能**：使用连接池管理数据库连接
- **安全**：生产环境必须使用环境变量或安全的配置管理
- **日志**：所有关键操作都要记录日志
- **错误处理**：启动失败应该快速失败（fail fast）
- **信号处理**：实现优雅关闭，避免数据丢失