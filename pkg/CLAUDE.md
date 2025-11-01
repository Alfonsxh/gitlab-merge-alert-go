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
  - 内部函数：camelCase（如 `validateClaims`）
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
- 第三方：golang-jwt/jwt/v5, golang.org/x/crypto

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
- 支持自定义过期时间（默认24小时）
- 包含用户ID、用户名、角色信息
- 错误类型区分（过期/无效）

**核心结构**：
```go
type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

type JWTManager struct {
    secretKey []byte
    duration  time.Duration
}
```

**核心 API**：
```go
// 创建管理器
func NewJWTManager(secretKey string, duration time.Duration) *JWTManager

// 生成 token（返回 token 字符串和过期时间戳）
func (m *JWTManager) Generate(userID uint, username, role string) (string, int64, error)

// 验证并解析 token
func (m *JWTManager) Verify(tokenString string) (*Claims, error)
```

**使用示例**：
```go
// 创建 JWT 管理器
manager := auth.NewJWTManager("your-secret-key", 24*time.Hour)

// 生成 token
token, expiresAt, err := manager.Generate(1, "admin", "admin")
if err != nil {
    return err
}

// 验证 token
claims, err := manager.Verify(token)
if err == auth.ErrExpiredToken {
    // Token 已过期，需要刷新
} else if err == auth.ErrInvalidToken {
    // Token 无效
} else {
    // 获取用户信息
    userID := claims.UserID
    role := claims.Role
}
```

### 密码处理
位置：`auth/password.go`

**功能特点**：
- 使用 bcrypt 算法
- 自动加盐保护
- 默认成本因子 14（平衡安全性和性能）
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
// 用户注册时加密密码
hashedPassword, err := auth.HashPassword("user-password-123")
if err != nil {
    return fmt.Errorf("password hashing failed: %w", err)
}
// 存储 hashedPassword 到数据库

// 用户登录时验证密码
valid := auth.CheckPassword("user-input-password", hashedPassword)
if !valid {
    return errors.New("incorrect password")
}
```

### 日志管理器
位置：`logger/logger.go`

**功能特点**：
- 全局单例模式
- 支持日志级别控制（debug/info/warn/error/fatal）
- 同时输出到文件和控制台
- JSON 格式化（生产环境）或文本格式（开发环境）
- 自动创建日志目录
- 包含时间戳、调用位置、字段信息

**初始化和使用**：
```go
// 初始化日志（在 main 函数中调用一次）
logger.Init("info")  // 设置日志级别

// 获取日志实例并使用
log := logger.GetLogger()

// 基本日志
log.Info("Server started successfully")
log.Warn("Configuration file not found, using defaults")
log.Error("Failed to connect to database")

// 带字段的结构化日志
log.WithFields(logrus.Fields{
    "user_id":  123,
    "username": "john",
    "action":   "login",
    "ip":       "192.168.1.1",
}).Info("User login successful")

// 错误日志
log.WithError(err).Error("Operation failed")

// 致命错误（会导致程序退出）
log.Fatal("Critical error, shutting down")
```

**日志输出格式**：
```json
{
  "time": "2024-08-11T10:30:45+08:00",
  "level": "info",
  "msg": "User login successful",
  "user_id": 123,
  "username": "john",
  "action": "login",
  "ip": "192.168.1.1"
}
```

## 配置映射关系

| 配置项 | 环境变量 | 默认值 | 影响功能 |
|---------|---------|--------|---------|
| log_level | LOG_LEVEL | info | 日志输出级别 |
| log_file | LOG_FILE | logs/app.log | 日志文件路径 |
| jwt_secret | JWT_SECRET | - | Token 签名密钥 |
| jwt_duration | JWT_DURATION | 24h | Token 有效期 |
| bcrypt_cost | BCRYPT_COST | 14 | 密码加密强度 |

## 开发指南

### 添加新的工具包

#### 步骤说明
1. 在 `pkg/` 创建新目录
2. 实现功能并导出公共 API
3. 编写单元测试
4. 更新文档

#### 代码示例
```go
// pkg/crypto/aes.go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

// AES 加密
func Encrypt(key []byte, plaintext string) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }
    
    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
    
    return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// AES 解密
func Decrypt(key []byte, ciphertext string) (string, error) {
    data, err := base64.URLEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    if len(data) < aes.BlockSize {
        return "", errors.New("ciphertext too short")
    }
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    iv := data[:aes.BlockSize]
    data = data[aes.BlockSize:]
    
    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(data, data)
    
    return string(data), nil
}
```

### 扩展现有功能

#### 添加 Token 刷新功能
```go
// auth/jwt.go
func (m *JWTManager) Refresh(tokenString string) (string, int64, error) {
    // 验证旧 token（允许过期）
    claims, err := m.parseToken(tokenString)
    if err != nil && !errors.Is(err, ErrExpiredToken) {
        return "", 0, err
    }
    
    // 检查刷新时间窗口（过期后7天内可刷新）
    if time.Since(claims.ExpiresAt.Time) > 7*24*time.Hour {
        return "", 0, errors.New("refresh window expired")
    }
    
    // 生成新 token
    return m.Generate(claims.UserID, claims.Username, claims.Role)
}
```

#### 添加密码强度验证
```go
// auth/password_strength.go
type PasswordStrength int

const (
    Weak PasswordStrength = iota
    Medium
    Strong
    VeryStrong
)

func CheckPasswordStrength(password string) PasswordStrength {
    score := 0
    
    // 长度检查
    if len(password) >= 8 {
        score++
    }
    if len(password) >= 12 {
        score++
    }
    
    // 字符类型检查
    if regexp.MustCompile(`[a-z]`).MatchString(password) {
        score++
    }
    if regexp.MustCompile(`[A-Z]`).MatchString(password) {
        score++
    }
    if regexp.MustCompile(`[0-9]`).MatchString(password) {
        score++
    }
    if regexp.MustCompile(`[!@#$%^&*]`).MatchString(password) {
        score++
    }
    
    switch {
    case score < 3:
        return Weak
    case score < 5:
        return Medium
    case score < 6:
        return Strong
    default:
        return VeryStrong
    }
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
        instance.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: "2006-01-02 15:04:05",
        })
        instance.SetLevel(logrus.InfoLevel)
        
        // 设置输出
        file, _ := os.OpenFile("logs/app.log", 
            os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        instance.SetOutput(io.MultiWriter(file, os.Stdout))
    })
    return instance
}
```

### 选项模式（配置）
```go
// auth/options.go
type JWTOption func(*JWTManager)

func WithDuration(d time.Duration) JWTOption {
    return func(m *JWTManager) {
        m.duration = d
    }
}

func WithSigningMethod(method jwt.SigningMethod) JWTOption {
    return func(m *JWTManager) {
        m.signingMethod = method
    }
}

func NewJWTManagerWithOptions(secret string, opts ...JWTOption) *JWTManager {
    m := &JWTManager{
        secretKey: []byte(secret),
        duration:  24 * time.Hour,  // 默认值
    }
    
    for _, opt := range opts {
        opt(m)
    }
    
    return m
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
go tool cover -html=coverage.out -o coverage.html

# 基准测试
go test -bench=. -benchmem ./pkg/auth

# 运行特定测试
go test -run TestJWTManager ./pkg/auth

# 代码检查
golangci-lint run ./pkg/...

# 格式化代码
gofmt -w pkg/
```

## 注意事项

- **重要**：包必须保持独立，不依赖 internal 目录的任何代码
- **性能**：
  - bcrypt 成本因子影响性能和安全性，生产环境建议 12-14
  - JWT 验证是 CPU 密集型操作，考虑缓存验证结果
- **安全**：
  - JWT 密钥必须足够复杂（至少 32 字符）且定期轮换
  - 密码必须使用 bcrypt，永远不要存储明文或简单哈希
  - 日志绝不记录敏感信息（密码、完整 token、密钥等）
  - Token 应该通过 HTTPS 传输
- **兼容性**：
  - 保持向后兼容，避免破坏性变更
  - 使用语义化版本管理
- **测试**：所有公共 API 必须有单元测试，覆盖率目标 80%+
- **文档**：所有导出的函数、类型、常量必须有 godoc 注释