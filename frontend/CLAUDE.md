# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 常用命令

### 开发命令
```bash
# 安装依赖
npm install

# 启动开发服务器（默认端口 3000，自动代理 API 到 localhost:1688）
npm run dev

# TypeScript 类型检查和构建
npm run build

# 预览生产构建
npm run preview
```

### 运行测试
项目目前未配置测试框架。如需添加测试，建议使用 Vitest。

## 架构概览

### 技术栈
- **框架**: Vue 3.5 + TypeScript 5.8（Composition API + `<script setup>`）
- **构建工具**: Vite 7
- **UI 库**: Element Plus 2.9
- **路由**: Vue Router 4（基于文件的路由）
- **状态管理**: Pinia 3（组合式 API 风格）
- **HTTP 客户端**: Axios 1.11
- **图表**: ECharts 5.6 + vue-echarts

### 核心架构模式

#### API 层架构
- **client.ts**: Axios 实例配置，包含请求/响应拦截器、自动 token 注入、401 自动刷新 token
- **API 模块**: 按功能域划分（auth、users、projects、webhooks、gitlab、notifications、stats）
- **类型定义**: `api/types/` 目录集中管理 TypeScript 接口定义

#### 认证流程
- JWT Token 存储在 localStorage
- 401 响应触发 token 刷新机制（虽然后端暂未实现 refresh token）
- 路由守卫实现权限控制（requiresAuth、requiresAdmin）
- 30分钟无活动自动登出
- 用户头像自动设为网站 favicon

#### 状态管理架构
- **auth store**: 管理认证状态、用户信息、权限判断
- **system store**: 管理系统配置、管理员初始化状态
- 使用 Pinia 组合式 API 风格（setup store）

#### 路由架构
- 公开路由：/login、/register、/setup-admin
- 受保护路由：需要认证才能访问
- 管理员路由：需要 admin 角色（如 /accounts）
- 全局前置守卫处理认证、权限、管理员初始化检查

### 目录结构说明
```
src/
├── api/          # API 接口封装
│   ├── client.ts # Axios 配置和拦截器
│   └── *.ts      # 各模块 API
├── views/        # 页面组件（路由页面）
├── components/   # 可复用组件
├── layouts/      # 布局组件（MainLayout）
├── stores/       # Pinia 状态管理
├── router/       # 路由配置
└── utils/        # 工具函数
```

## 关键业务逻辑

### API 请求拦截器链
1. 请求拦截：注入 Bearer token、触发用户活动事件
2. 响应拦截：401 尝试刷新 token、403 显示权限错误、其他错误统一处理
3. Token 刷新失败自动跳转登录页

### 管理员初始化流程
- 系统启动时检查是否需要初始化管理员（/setup-admin）
- 未初始化时所有路由重定向到初始化页面
- 初始化完成后自动跳转原目标页面

### 权限控制机制
- 基于角色的权限（admin/user）
- Store 中的 hasPermission 方法判断资源操作权限
- 路由 meta 配置 requiresAdmin 控制页面访问

## 开发注意事项

### API 调用模式
所有 API 调用已封装，直接使用：
```typescript
import { projectApi } from '@/api/projects'
const { data } = await projectApi.list()
```

### 组件开发模式
使用 Vue 3 Composition API + TypeScript：
```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
// 逻辑代码
</script>
```

### 路由配置模式
在 `router/index.ts` 中添加路由，配置 meta 属性控制权限：
```typescript
{
  path: '/example',
  meta: {
    title: '页面标题',
    requiresAuth: true,    // 需要登录
    requiresAdmin: false   // 需要管理员
  }
}
```

### Vite 配置要点
- 开发服务器端口：3000
- API 代理：`/api` -> `http://localhost:1688`
- 路径别名：`@` -> `./src`
- 构建输出：`dist` 目录