# GitLab Merge Alert (Golang 版本)

![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?logo=go) ![License](https://img.shields.io/badge/license-MIT-green) ![Status](https://img.shields.io/badge/status-active-success)

> 一款专为 GitLab Merge Request 打造的自动化通知服务，使用 Go 语言重构，开箱即用地推送到企业微信机器人。

## 目录
- [项目简介](#项目简介)
- [核心特性](#核心特性)
  - [功能特性](#功能特性)
  - [管理能力](#管理能力)
  - [技术特性](#技术特性)
- [系统架构概览](#系统架构概览)
- [快速开始](#快速开始)
  - [环境要求](#环境要求)
  - [本地开发指南](#本地开发指南)
  - [管理员首次初始化](#管理员首次初始化)
  - [Docker 部署](#docker-部署)
- [配置指南](#配置指南)
  - [配置文件说明](#配置文件说明)
  - [环境变量对照表](#环境变量对照表)
  - [GitLab Webhook 设置](#gitlab-webhook-设置)
  - [企业微信机器人接入](#企业微信机器人接入)
- [开发与测试](#开发与测试)
  - [常用 Make 命令](#常用-make-命令)
  - [数据库迁移](#数据库迁移)
- [从 Python 版本迁移](#从-python-版本迁移)
- [故障排除](#故障排除)
  - [常见问题](#常见问题)
  - [日志查看](#日志查看)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

## 项目简介
GitLab Merge Alert 是基于 Go 语言的 GitLab Merge Request 通知平台，用于实时捕获 GitLab Webhook 事件并将通知推送到企业微信群机器人。项目包含现代化 Web 管理界面、灵活的配置能力以及完善的数据统计功能，适用于企业内的代码协作与质量把控场景。

## 核心特性

### 功能特性
- ✅ 自动接收并解析 GitLab Merge Request Webhook 事件
- ✅ 将格式化消息推送至企业微信群机器人
- ✅ 多项目、多机器人关联管理
- ✅ GitLab 邮箱与企业微信手机号智能映射
- ✅ 通知历史记录与统计分析

### 管理能力
- ✅ 基于 Bootstrap 5 + Vue.js 3 的管理后台
- ✅ 用户映射管理与批量导入
- ✅ GitLab 项目 URL 解析与组扫描
- ✅ 企业微信机器人配置与状态监控
- ✅ GitLab Webhook 自动同步与管理

### 技术特性
- ✅ Go 1.23 + Gin + GORM + SQLite
- ✅ 支持配置文件与环境变量双通道配置
- ✅ 内置数据库迁移与版本管理
- ✅ 结构化日志与错误追踪
- ✅ Docker 容器化部署与一键脚本支持

## 系统架构概览
```
GitLab Webhook --> 后端服务 (Gin) --> 业务处理层 --> 通知通道 (企业微信)
                            |--> 数据库 (SQLite)
                            |--> 前端管理界面 (Vue.js + Bootstrap)
```

## 快速开始

### 环境要求
- Go 1.23+
- Make
- SQLite3（无需单独安装，默认随应用内置）
- 企业微信机器人（用于接收通知）

### 本地开发指南
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

   ⚠️ **安全提示：请勿提交敏感配置到版本库。**

   **方式一：本地配置文件（推荐）**
   ```bash
   cp config.example.yaml config.local.yaml
   # 编辑 config.local.yaml 并填入真实配置
   ```

   **方式二：环境变量**
   ```bash
   export GMA_GITLAB_URL="https://your-gitlab-server.com"
   export GMA_DEFAULT_WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR-REAL-KEY"
   ```

   配置文件约定：
   - `config.local.yaml`：仅用于本地敏感信息（已加入 `.gitignore`）
   - `config.yaml`：安全示例配置，可提交
   - `config.example.yaml`：模板文件，方便复制

5. **启动服务**
   ```bash
   make run
   ```
   打开浏览器访问 `http://localhost:1688` 查看管理界面。

### 管理员首次初始化
首次启动时系统会在日志中输出一次性 "Admin setup token"。请按以下流程完成初始化：
1. 在服务日志或容器日志中找到最新的 Setup Token。
2. 访问前端页面 `/setup-admin`（或根据登录页提示跳转）。
3. 输入 token、管理员邮箱和新密码完成初始化。
4. 完成后 token 即失效，若需重新生成，请重启服务。

### Docker 部署

#### 方式一：使用 Makefile（推荐）
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

#### 方式二：手动 Docker 操作
1. **构建镜像**
   ```bash
   docker build -t gitlab-merge-alert:latest .
   ```
2. **准备数据目录**
   ```bash
   mkdir -p ./docker-data
   chmod 755 ./docker-data
   ```
3. **运行容器**
   ```bash
   docker run -d \
     --name gitlab-merge-alert \
     -p 1688:1688 \
     -v $PWD/docker-data:/app/data \
     -v $PWD/config.local.yaml:/app/config.local.yaml:ro \
     gitlab-merge-alert:latest
   ```
4. **查看日志**
   ```bash
   docker logs -f gitlab-merge-alert
   ```

## 配置指南

### 配置文件说明
项目默认会加载 `config.yaml`，并自动合并 `config.local.yaml`（如果存在）。配置优先级如下：

```
环境变量 > config.local.yaml > config.yaml > 内置默认值
```

### 环境变量对照表
| 名称 | 说明 | 默认值 |
| ---- | ---- | ------ |
| `GMA_HOST` | 服务监听地址 | `0.0.0.0` |
| `GMA_PORT` | 服务监听端口 | `1688` |
| `GMA_DATABASE_PATH` | SQLite 数据库文件路径 | `./data/app.db` |
| `GMA_LOG_LEVEL` | 日志级别（debug/info/warn/error） | `info` |
| `GMA_ENVIRONMENT` | 运行环境（development/production） | `development` |
| `GMA_GITLAB_URL` | GitLab 服务器地址 | 必填 |
| `GMA_DEFAULT_WEBHOOK_URL` | 默认企业微信机器人地址 | 可选 |
| `GMA_ENCRYPTION_KEY` | 敏感信息加密密钥（为空时回退到 `JWT_SECRET`） | 可选 |
| `GMA_REDIRECT_SERVER_URL` | 重定向服务器地址 | 可选 |
| `GMA_PUBLIC_WEBHOOK_URL` | 公共 Webhook 地址 | 可选 |

> GitLab Personal Access Token 请在系统管理界面中单独配置，不建议写入环境变量或配置文件。

### GitLab Webhook 设置
1. 打开 GitLab 项目 `Settings → Webhooks`
2. 将 URL 设置为 `http://your-server:1688/api/v1/webhook/gitlab`
3. 选择 `Merge request events`
4. 点击 **Add webhook** 完成配置

### 企业微信机器人接入
1. 在目标企业微信群中添加机器人，获取 Webhook URL
2. 在系统后台中创建企业微信机器人配置，并粘贴 Webhook URL
3. 将机器人与对应 GitLab 项目进行关联
4. 提交一次测试 Merge Request 验证通知流程

## 开发与测试

### 常用 Make 命令
```bash
make deps    # 安装依赖
make init    # 初始化数据与日志目录
make run     # 启动开发环境
make build   # 构建二进制
make fmt     # GoFmt 全量格式化
make test    # 运行测试用例
make lint    # 静态检查
```

### 数据库迁移
```bash
make migrate           # 执行所有待运行的迁移
make migrate-status    # 查看迁移状态
make migrate-rollback  # 回滚最近一次迁移
```

数据表说明：
- `users`：GitLab 邮箱与企业微信手机号映射
- `projects`：GitLab 项目信息
- `webhooks`：企业微信机器人配置
- `project_webhooks`：项目与 Webhook 的关联关系
- `notifications`：通知历史记录

## 从 Python 版本迁移
1. 导出现有配置和映射数据
2. 部署并配置 Golang 版本
3. 将导出的数据导入新系统
4. 更新 GitLab Webhook 指向新服务
5. 验证企业微信通知是否正常

## 故障排除

### 常见问题
1. **数据库连接失败**
   - 确认数据库路径与文件权限
   - 确保数据目录已创建
2. **GitLab Webhook 请求未达**
   - 检查网络连通性与防火墙配置
   - 核对 Webhook URL 是否正确
3. **企业微信通知发送失败**
   - 核对机器人 Webhook URL
   - 检查机器人权限或所在群组设置

### 日志查看
```bash
make docker-logs   # Docker 环境查看日志
mkdir -p logs && tail -f logs/app.log  # 本地环境实时查看日志
```

## 贡献指南
欢迎通过 Issue 或 Pull Request 贡献功能、修复缺陷或完善文档。在提交前请确保通过 `make fmt` 与 `make test`。

## 许可证
本项目基于 [MIT License](./LICENSE) 开源。
