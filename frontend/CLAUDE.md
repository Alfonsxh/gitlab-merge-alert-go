# frontend - Vue.js SPA 前端应用

## 概述

`frontend` 目录包含基于 Vue.js 3 + TypeScript 的单页应用（SPA），提供了完整的 Web 管理界面。采用 Element Plus 作为 UI 框架，通过 RESTful API 与后端通信，实现了用户管理、项目配置、Webhook 管理等功能。

## 目录结构

```
frontend/
├── src/
│   ├── api/                 # API 客户端封装
│   │   ├── client.ts       # Axios 实例和拦截器配置
│   │   ├── auth.ts         # 认证相关 API
│   │   ├── users.ts        # 用户管理 API
│   │   ├── projects.ts     # 项目管理 API
│   │   ├── webhooks.ts     # Webhook 管理 API
│   │   ├── gitlab.ts       # GitLab 集成 API
│   │   ├── notifications.ts # 通知历史 API
│   │   ├── stats.ts        # 统计数据 API
│   │   ├── resource-manager.ts # 资源管理 API
│   │   ├── types/          # TypeScript 类型定义
│   │   │   └── auth.ts     # 认证相关类型
│   │   └── index.ts        # API 统一导出
│   ├── views/               # 页面组件
│   │   ├── auth/           # 认证页面
│   │   │   └── Login.vue   # 登录页
│   │   ├── dashboard/      # 仪表板
│   │   │   └── Dashboard.vue # 统计概览页
│   │   ├── users/          # 用户管理
│   │   │   └── Users.vue   # 用户列表页
│   │   ├── projects/       # 项目管理
│   │   │   └── Projects.vue # 项目配置页
│   │   ├── webhooks/       # Webhook 管理
│   │   │   └── Webhooks.vue # Webhook 配置页
│   │   ├── accounts/       # 账号管理
│   │   │   └── Accounts.vue # 账号列表页
│   │   └── profile/        # 个人资料
│   │       └── Profile.vue # 个人信息页
│   ├── components/          # 通用组件
│   │   └── charts/         # 图表组件
│   │       └── LineChart.vue # 折线图组件
│   ├── layouts/             # 布局组件
│   │   └── MainLayout.vue  # 主布局（侧边栏+顶栏）
│   ├── stores/              # Pinia 状态管理
│   │   ├── auth.ts         # 认证状态
│   │   └── index.ts        # Store 配置
│   ├── router/              # 路由配置
│   │   └── index.ts        # Vue Router 配置
│   ├── utils/               # 工具函数
│   │   ├── format.ts       # 格式化工具
│   │   └── favicon.ts      # 动态 Favicon
│   ├── assets/              # 静态资源
│   │   └── styles/         # 样式文件
│   │       └── index.css   # 全局样式
│   ├── App.vue             # 根组件
│   ├── main.ts             # 应用入口
│   └── vite-env.d.ts       # 类型声明
├── public/                  # 公共静态资源
│   └── vite.svg           # 默认图标
├── dist/                    # 构建输出目录
├── package.json            # 项目配置和依赖
├── tsconfig.json           # TypeScript 配置
├── vite.config.ts          # Vite 构建配置
└── README.md               # 前端说明文档
```

## 技术栈

- **核心技术**：
  - Vue.js 3.5.17 - 渐进式 JavaScript 框架
  - TypeScript 5.8.3 - 类型安全的 JavaScript
  - Vite 7.0.4 - 快速构建工具
- **依赖库**：
  - Element Plus 2.9.2 - UI 组件库
  - Vue Router 4.5.1 - 路由管理
  - Pinia 3.0.3 - 状态管理
  - Axios 1.11.0 - HTTP 客户端
  - ECharts 5.6.0 - 数据可视化
  - dayjs 1.11.10 - 日期处理
- **开发工具**：
  - vue-tsc 2.2.12 - TypeScript 编译器
  - Less 4.4.0 - CSS 预处理器

## 编码规范

- **命名规则**：
  - 组件名：PascalCase（如 `Dashboard.vue`）
  - 函数名：camelCase（如 `fetchUsers`）
  - 文件名：kebab-case（如 `resource-manager.ts`）
  - CSS 类名：kebab-case（如 `.main-container`）
- **代码风格**：
  - 使用 Composition API
  - 使用 `<script setup>` 语法
  - TypeScript 严格模式
- **文件组织**：
  - 单文件组件（SFC）
  - 按功能模块组织视图

## 依赖关系

**views 依赖于**：
- api：后端接口调用
- stores：全局状态管理
- components：可复用组件
- utils：工具函数

**api 依赖于**：
- client：Axios 实例
- types：TypeScript 类型定义

**stores 依赖于**：
- api：数据获取
- router：路由跳转

## 关键业务逻辑说明

### 认证流程
位置：`stores/auth.ts`, `api/client.ts`

**实现特点**：
- JWT Token 存储在 localStorage
- Axios 拦截器自动添加 Authorization header
- 401 响应自动跳转登录页
- Token 过期自动清理

**认证状态管理**：
```typescript
// stores/auth.ts
const authStore = useAuthStore()
authStore.login(username, password) // 登录
authStore.logout() // 登出
authStore.isAuthenticated // 认证状态
```

### API 请求封装
位置：`api/client.ts`

**实现特点**：
- 统一的错误处理
- 请求/响应拦截器
- 自动添加认证 header
- 统一的 API 前缀配置

**请求示例**：
```typescript
// api/projects.ts
export const projectApi = {
  list: () => client.get('/projects'),
  create: (data) => client.post('/projects', data),
  update: (id, data) => client.put(`/projects/${id}`, data),
  delete: (id) => client.delete(`/projects/${id}`)
}
```

### 路由守卫
位置：`router/index.ts`

**实现特点**：
- 基于认证状态的路由保护
- 角色权限检查（admin/user）
- 自动重定向到登录页
- 登录后返回原页面

### 数据可视化
位置：`components/charts/`, `views/dashboard/Dashboard.vue`

**实现特点**：
- 使用 ECharts 渲染图表
- 响应式图表尺寸
- 实时数据更新
- 主题自适应

## 配置映射关系

| 配置文件 | 配置项 | 影响功能 |
|---------|--------|---------|
| vite.config.ts | `server.proxy` | API 代理配置 |
| vite.config.ts | `base` | 部署路径 |
| .env | `VITE_API_URL` | API 服务器地址 |
| tsconfig.json | `compilerOptions` | TypeScript 编译选项 |

## 开发指南

### 添加新页面

#### 步骤说明
1. 在 `views/` 创建页面组件
2. 在 `router/index.ts` 添加路由
3. 在 `api/` 创建对应 API 模块
4. 更新导航菜单

#### 代码示例
```vue
<!-- views/example/Example.vue -->
<template>
  <div class="example-container">
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="name" label="名称" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { exampleApi } from '@/api/example'

const list = ref([])
const loading = ref(false)

const fetchData = async () => {
  loading.value = true
  try {
    const { data } = await exampleApi.list()
    list.value = data
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

### 添加 API 模块

#### 代码示例
```typescript
// api/example.ts
import { client } from './client'

export interface Example {
  id: number
  name: string
  createdAt: string
}

export const exampleApi = {
  list: () => client.get<Example[]>('/examples'),
  create: (data: Partial<Example>) => 
    client.post<Example>('/examples', data),
  update: (id: number, data: Partial<Example>) => 
    client.put<Example>(`/examples/${id}`, data),
  delete: (id: number) => 
    client.delete(`/examples/${id}`)
}
```

### 添加状态管理

#### 代码示例
```typescript
// stores/example.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { exampleApi } from '@/api/example'

export const useExampleStore = defineStore('example', () => {
  const list = ref([])
  const loading = ref(false)
  
  const fetchList = async () => {
    loading.value = true
    try {
      const { data } = await exampleApi.list()
      list.value = data
    } finally {
      loading.value = false
    }
  }
  
  return {
    list,
    loading,
    fetchList
  }
})
```

## 常用设计模式

### 组合式 API 模式
```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

// 响应式状态
const count = ref(0)

// 计算属性
const doubled = computed(() => count.value * 2)

// 生命周期
onMounted(() => {
  console.log('Component mounted')
})
</script>
```

### 自定义 Hook 模式
```typescript
// composables/useTable.ts
export function useTable<T>() {
  const data = ref<T[]>([])
  const loading = ref(false)
  const currentPage = ref(1)
  const pageSize = ref(10)
  
  const fetchData = async (apiFn: Function) => {
    loading.value = true
    try {
      const result = await apiFn({
        page: currentPage.value,
        size: pageSize.value
      })
      data.value = result.data
    } finally {
      loading.value = false
    }
  }
  
  return {
    data,
    loading,
    currentPage,
    pageSize,
    fetchData
  }
}
```

## 相关命令

```bash
# 安装依赖
npm install

# 开发服务器（热更新）
npm run dev

# TypeScript 类型检查
npm run type-check

# 生产构建
npm run build

# 预览生产构建
npm run preview

# 代码格式化
npm run format

# 代码检查
npm run lint
```

## 注意事项

- **重要**：生产环境必须配置正确的 API 地址
- **性能**：大数据列表使用虚拟滚动或分页
- **安全**：敏感信息不要存储在前端代码中
- **兼容性**：支持现代浏览器（Chrome, Firefox, Safari, Edge）
- **响应式**：确保界面在不同屏幕尺寸下正常显示
- **错误处理**：所有 API 调用都要有错误处理和用户提示