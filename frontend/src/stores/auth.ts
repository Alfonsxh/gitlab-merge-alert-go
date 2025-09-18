import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/api/auth'
import type { AccountResponse, LoginResponse } from '@/api/types/auth'
import { setFaviconFromAvatar } from '@/utils/favicon'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const refreshToken = ref<string | null>(localStorage.getItem('refreshToken'))
  const user = ref<AccountResponse | null>(null)
  const lastActivityTime = ref<number>(Date.now())

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const username = computed(() => user.value?.username || '')
  const email = computed(() => user.value?.email || '')
  const hasGitLabToken = computed(() => !!user.value?.has_gitlab_personal_access_token)

  const login = async (username: string, password: string) => {
    try {
      const response = await authAPI.login({ username, password })
      setAuthData(response)
      await fetchProfile()
      // 设置用户头像为 favicon
      if (user.value?.avatar) {
        setFaviconFromAvatar(user.value.avatar)
      }
    } catch (error) {
      clearAuthData()
      throw error
    }
  }

  const register = async (payload: { username: string; email: string; password: string }) => {
    try {
      const response = await authAPI.register(payload)
      setAuthData(response)
      await fetchProfile()
      if (user.value?.avatar) {
        setFaviconFromAvatar(user.value.avatar)
      }
    } catch (error) {
      clearAuthData()
      throw error
    }
  }

  const logout = async () => {
    try {
      await authAPI.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      clearAuthData()
      // 重置为默认 favicon
      setFaviconFromAvatar(null)
    }
  }

  const fetchProfile = async () => {
    if (!token.value) {
      throw new Error('No token available')
    }
    
    try {
      const profile = await authAPI.getProfile()
      user.value = profile
      // 更新 favicon
      if (profile.avatar) {
        setFaviconFromAvatar(profile.avatar)
      }
    } catch (error) {
      if ((error as any).response?.status === 401) {
        clearAuthData()
      }
      throw error
    }
  }

  const refreshAccessToken = async () => {
    if (!refreshToken.value) {
      throw new Error('No refresh token available')
    }

    try {
      const response = await authAPI.refreshToken(refreshToken.value)
      setAuthData(response)
      return response.token
    } catch (error) {
      clearAuthData()
      throw error
    }
  }

  const updateLastActivity = () => {
    lastActivityTime.value = Date.now()
  }

  const checkTokenExpiry = () => {
    // 检查是否超过30分钟无活动
    const thirtyMinutes = 30 * 60 * 1000
    if (Date.now() - lastActivityTime.value > thirtyMinutes) {
      clearAuthData()
      return false
    }
    return true
  }

  const setAuthData = (data: LoginResponse) => {
    token.value = data.token
    refreshToken.value = null // 后端暂不支持 refresh token
    user.value = data.user

    localStorage.setItem('token', data.token)

    updateLastActivity()
  }

  const clearAuthData = () => {
    token.value = null
    refreshToken.value = null
    user.value = null
    
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  const hasPermission = (resource: string, action: string): boolean => {
    if (!user.value) return false
    
    // 管理员拥有所有权限
    if (user.value.role === 'admin') return true
    
    // 普通用户权限检查
    const userPermissions: Record<string, string[]> = {
      'projects': ['view', 'create', 'update', 'delete'],
      'webhooks': ['view', 'create', 'update', 'delete'],
      'users': ['view', 'create', 'update', 'delete'],
      'notifications': ['view'],
      'profile': ['view', 'update'],
      'accounts': [] // 普通用户没有账户管理权限
    }
    
    return userPermissions[resource]?.includes(action) || false
  }

  const canAccessResource = (resource: string): boolean => {
    return hasPermission(resource, 'view')
  }

  return {
    // State
    token,
    refreshToken,
    user,
    lastActivityTime,
    
    // Getters
    isAuthenticated,
    isAdmin,
    username,
    email,
    hasGitLabToken,
    
    // Actions
    login,
    register,
    logout,
    fetchProfile,
    refreshAccessToken,
    updateLastActivity,
    checkTokenExpiry,
    clearAuthData,
    hasPermission,
    canAccessResource
  }
})
