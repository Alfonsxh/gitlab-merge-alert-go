# GitLab Merge Alert 前端

基于 Vue 3 + TypeScript + Arco Design Vue 构建的现代化前端界面。

## 技术栈

- **框架**: Vue 3.4 + TypeScript
- **构建工具**: Vite 5
- **UI 组件库**: @arco-design/web-vue
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **HTTP 客户端**: Axios

## 开发

### 前置要求

- Node.js 18+
- npm 或 yarn

### 安装依赖

```bash
cd frontend
npm install
```

### 启动开发服务器

```bash
npm run dev
# 或使用项目根目录的 Makefile
make frontend-dev
```

开发服务器将在 http://localhost:3000 启动，API 请求会自动代理到后端服务器 http://localhost:1688。

### 同时启动前后端

```bash
make dev
```

## 构建

### 构建前端

```bash
npm run build
# 或使用项目根目录的 Makefile
make frontend-build
```

构建产物将输出到 `web/dist` 目录。

### 完整构建（前端 + 后端）

```bash
make build
```

## 项目结构

```
frontend/
├── src/
│   ├── api/          # API 接口封装
│   ├── assets/       # 静态资源
│   ├── components/   # 公共组件
│   ├── layouts/      # 布局组件
│   ├── router/       # 路由配置
│   ├── stores/       # Pinia 状态管理
│   ├── utils/        # 工具函数
│   └── views/        # 页面组件
│       ├── dashboard/    # 仪表板
│       ├── users/        # 用户管理
│       ├── projects/     # 项目管理
│       └── webhooks/     # Webhook管理
```

## 功能特性

- 📊 **仪表板**: 系统统计数据和最近通知记录
- 👥 **用户管理**: GitLab 邮箱与企业微信手机号映射
- 📁 **项目管理**: GitLab 项目配置和 Webhook 同步
- 🔗 **Webhook管理**: 企业微信机器人配置

## 部署模式

系统支持两种部署模式：

1. **SPA 模式**（推荐）：当存在前端构建文件时，自动使用 SPA 模式
2. **SSR 模式**：当没有前端构建文件时，回退到服务端渲染模式

## 开发指南

### API 封装

所有 API 请求都封装在 `src/api` 目录下，按模块划分：

- `stats.ts` - 统计相关 API
- `users.ts` - 用户管理 API
- `projects.ts` - 项目管理 API
- `webhooks.ts` - Webhook 管理 API
- `notifications.ts` - 通知相关 API
- `gitlab.ts` - GitLab 集成 API

### 添加新页面

1. 在 `src/views` 下创建页面组件
2. 在 `src/router/index.ts` 中添加路由
3. 在 `src/layouts/MainLayout.vue` 中添加导航菜单

### 主题定制

Arco Design Vue 支持主题定制，可以通过配置修改主题色、字体等样式。详见 [Arco Design 主题配置](https://arco.design/vue/docs/theme)。