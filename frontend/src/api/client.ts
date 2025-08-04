import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { AxiosError, InternalAxiosRequestConfig } from 'axios'
import router from '@/router'

const apiClient = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    // 更新最后活动时间（通过自定义事件）
    window.dispatchEvent(new CustomEvent('userActivity'))
    
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  response => {
    return response.data
  },
  async (error: AxiosError) => {
    const { response, config } = error
    
    if (response?.status === 401) {
      // Token 失效，尝试刷新
      const refreshToken = localStorage.getItem('refreshToken')
      
      if (refreshToken && config && !(config as any)._retry) {
        (config as any)._retry = true
        
        try {
          // 动态导入以避免循环依赖
          const { useAuthStore } = await import('@/stores/auth')
          const authStore = useAuthStore()
          const newToken = await authStore.refreshAccessToken()
          
          // 更新原请求的 Authorization
          config.headers.Authorization = `Bearer ${newToken}`
          
          // 重试原请求
          return apiClient(config)
        } catch (refreshError) {
          // 刷新失败，清除认证信息并跳转到登录页
          const { useAuthStore } = await import('@/stores/auth')
          const authStore = useAuthStore()
          authStore.clearAuthData()
          
          router.push({
            path: '/login',
            query: { redirect: router.currentRoute.value.fullPath }
          })
          
          return Promise.reject(refreshError)
        }
      } else {
        // 没有 refresh token 或已经重试过，跳转到登录页
        const { useAuthStore } = await import('@/stores/auth')
        const authStore = useAuthStore()
        authStore.clearAuthData()
        
        router.push({
          path: '/login',
          query: { redirect: router.currentRoute.value.fullPath }
        })
      }
    } else if (response?.status === 403) {
      ElMessage.error('您没有权限执行此操作')
    } else {
      const message = (response?.data as any)?.error || error.message || '请求失败'
      ElMessage.error(message)
    }
    
    return Promise.reject(error)
  }
)

export default apiClient
export { apiClient }