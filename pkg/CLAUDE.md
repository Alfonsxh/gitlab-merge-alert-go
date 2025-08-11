# pkg - 可重用包

## 概述

`pkg` 目录包含可被外部项目引用的公共包，提供通用的功能实现。这些包设计为独立、可重用的组件，不依赖于项目特定的业务逻辑。目前包含认证（JWT、密码）和日志两个核心功能包。

## 目录结构

```
pkg/
├── auth/                # 认证相关功能
│   ├── jwt.go          # JWT 令牌生成和验证
│   └── password.go     # 密码哈希和验证（bcrypt）
└── logger/             # 日志工具
    └── logger.go       # Logrus 日志封装和配置
```

## 技术栈

- **核心技术**：Go 1.23 标准库
- **依赖库**：
  - golang-jwt/jwt/v5 v5.3.0 - JWT 处理
  - golang.org/x/crypto v0.39.0 - bcrypt 密码加密
  - Logrus v1.9.3 - 结构化日志
- **工具**：gofmt, golangci-lint

## 编码规范

- **命名规则**：
  - 包名：小写单词（如 `auth`, `logger`）
  - 导出类型：PascalCase（如 `JWTManager`）
  - 导出函数：PascalCase（如 `HashPassword`）
  - 内部函数：camelCase（如 `generateSalt`）
- **代码风格**：
  - 每个包专注单一职责
  - 提供清晰的公共 API
  - 充分的错误处理
- **文件组织**：
  - 相关功能放在同一包内
  - 避免循环依赖

## 依赖关系

**auth 包依赖于**：
- 标准库：time, errors
- 第三方：jwt-go, bcrypt

**logger 包依赖于**：
- 标准库：os, path
- 第三方：logrus

**被依赖方**：
- internal/middleware：使用 auth 包进行认证
- internal/services：使用 auth 包生成 token
- 所有模块：使用 logger 包记录日志

## 关键功能说明

### JWT 管理器
位置：`auth/jwt.go`

**功能特点**：
- 生成带过期时间的 JWT token
- 验证 token 有效性
- 解析 token 获取用户信息
- 支持自定义过期时间
- 错误类型区分（过期/无效）

**核心 API**：
```go
type JWTManager struct {
    secret        string
    tokenDuration time.Duration
}

// 创建管理器
func NewJWTManager(secret string, duration time.Duration) *JWTManager

// 生成 token
func (m *JWTManager) Generate(userID uint, username string, role string) (string, error)

// 验证 token
func (m *JWTManager) Verify(tokenString string) (*UserClaims, error)
```

**使用示例**：
```go
manager := auth.NewJWTManager("secret-key", 24*time.Hour)

// 生成 token
token, err := manager.Generate(1, "admin", "admin")

// 验证 token
claims, err := manager.Verify(token)
if err == auth.ErrExpiredToken {
    // token 已过期
} else if err != nil {
    // token 无效
}
```

### 密码处理
位置：`auth/password.go`

**功能特点**：
- 使用 bcrypt 算法
- 自动加盐
- 可配置成本因子（默认 10）
- 恒定时间比较防止时序攻击

**核心 API**：
```go
// 加密密码
func HashPassword(password string) (string, error)

// 验证密码
func CheckPassword(password, hash string) bool
```

**使用示例**：
```go
// 加密密码
hash, err := auth.HashPassword("user-password")
if err != nil {
    return err
}

// 验证密码
valid := auth.CheckPassword("user-password", hash)
if !valid {
    return errors.New("密码错误")
}
```

### 日志管理器
位置：`logger/logger.go`

**功能特点**：
- 全局单例模式
- 支持日志级别控制
- 输出到文件和控制台
- JSON 格式化输出
- 自动日志轮转（需配置）
- 包含调用位置信息

**核心 API**：
```go
// 初始化日志
func InitLogger()

// 获取日志实例
func GetLogger() *logrus.Logger

// 使用示例
logger.GetLogger().Info("操作成功")
logger.GetLogger().WithFields(logrus.Fields{
    "user_id": 123,
    "action":  "login",
}).Info("用户登录")
logger.GetLogger().Error("操作失败:", err)
```

**日志级别**：
- Debug：调试信息
- Info：一般信息
- Warn：警告信息
- Error：错误信息
- Fatal：致命错误（会退出程序）

## 配置映射关系

| 环境变量 | 配置项 | 影响功能 |
|---------|--------|---------|
| LOG_LEVEL | 日志级别 | 控制日志输出详细程度 |
| LOG_FILE | 日志文件路径 | 日志输出位置 |
| JWT_SECRET | JWT 密钥 | Token 签名密钥 |
| BCRYPT_COST | bcrypt 成本 | 密码加密强度 |

## 开发指南

### 添加新的工具包

#### 步骤说明
1. 在 `pkg/` 创建新目录
2. 实现功能并导出公共 API
3. 编写单元测试
4. 更新文档

#### 代码示例
```go
// pkg/validator/validator.go
package validator

import (
    "regexp"
    "errors"
)

var (
    ErrInvalidEmail = errors.New("invalid email format")
    ErrInvalidPhone = errors.New("invalid phone format")
)

// 验证邮箱格式
func ValidateEmail(email string) error {
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    match, _ := regexp.MatchString(pattern, email)
    if !match {
        return ErrInvalidEmail
    }
    return nil
}

// 验证手机号格式
func ValidatePhone(phone string) error {
    pattern := `^1[3-9]\d{9}$`
    match, _ := regexp.MatchString(pattern, phone)
    if !match {
        return ErrInvalidPhone
    }
    return nil
}
```

### 扩展 JWT 功能

#### 添加刷新 token
```go
// auth/jwt.go
func (m *JWTManager) Refresh(tokenString string) (string, error) {
    claims, err := m.Verify(tokenString)
    if err != nil && err != ErrExpiredToken {
        return "", err
    }
    
    // 生成新 token
    return m.Generate(claims.UserID, claims.Username, claims.Role)
}
```

#### 添加 token 黑名单
```go
// auth/blacklist.go
type TokenBlacklist interface {
    Add(token string, expiry time.Time) error
    Contains(token string) bool
    Clean() // 清理过期项
}

type MemoryBlacklist struct {
    tokens map[string]time.Time
    mu     sync.RWMutex
}

func (b *MemoryBlacklist) Add(token string, expiry time.Time) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.tokens[token] = expiry
    return nil
}
```

## 常用设计模式

### 单例模式（日志）
```go
// logger/logger.go
var (
    instance *logrus.Logger
    once     sync.Once
)

func GetLogger() *logrus.Logger {
    once.Do(func() {
        instance = logrus.New()
        // 配置日志
    })
    return instance
}
```

### 工厂模式（认证）
```go
// auth/factory.go
type Authenticator interface {
    Authenticate(credentials interface{}) (User, error)
}

func NewAuthenticator(authType string) Authenticator {
    switch authType {
    case "jwt":
        return &JWTAuthenticator{}
    case "oauth":
        return &OAuthAuthenticator{}
    default:
        return &BasicAuthenticator{}
    }
}
```

### 策略模式（密码强度）
```go
// auth/password_policy.go
type PasswordPolicy interface {
    Validate(password string) error
}

type StrongPasswordPolicy struct{}

func (p *StrongPasswordPolicy) Validate(password string) error {
    if len(password) < 12 {
        return errors.New("密码长度不足12位")
    }
    // 其他强度检查
    return nil
}
```

## 相关命令

```bash
# 运行测试
go test ./pkg/...

# 测试覆盖率
go test -cover ./pkg/...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out

# 基准测试
go test -bench=. ./pkg/auth

# 代码检查
golangci-lint run ./pkg/...

# 格式化代码
gofmt -w pkg/
```

## 注意事项

- **重要**：包必须保持独立，不依赖 internal 目录
- **性能**：bcrypt 成本因子影响性能，生产环境建议 10-12
- **安全**：
  - JWT 密钥必须足够复杂且定期轮换
  - 密码必须使用 bcrypt 而非明文或简单哈希
  - 日志不要记录敏感信息（密码、token 等）
- **兼容性**：保持向后兼容，避免破坏性变更
- **测试**：所有公共 API 必须有单元测试
- **文档**：导出的函数和类型必须有注释