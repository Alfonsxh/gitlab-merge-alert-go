import { apiClient } from './client'
import type {
  LoginRequest,
  LoginResponse,
  ChangePasswordRequest,
  AccountResponse,
  RegisterRequest
} from './types/auth'

export const authAPI = {
  // 用户注册
  register: async (data: RegisterRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/auth/register', data)
    return response.data
  },

  // 登录
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<any>('/auth/login', data)
    // 后端返回的是包装后的响应，需要提取 data 字段
    return response.data
  },

  // 登出
  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout')
  },

  // 刷新Token
  refreshToken: async (token: string): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/auth/refresh', { token })
    return response.data
  },

  // 获取当前用户信息
  getProfile: async (): Promise<AccountResponse> => {
    const response = await apiClient.get<AccountResponse>('/auth/profile')
    return response.data
  },

  // 修改密码
  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    await apiClient.post('/auth/change-password', data)
  },

  // 更新个人资料
  updateProfile: async (data: { email?: string; avatar?: string; gitlab_personal_access_token?: string }): Promise<AccountResponse> => {
    const response = await apiClient.put<AccountResponse>('/auth/profile', data)
    return response.data
  },

  // 上传头像
  uploadAvatar: async (formData: FormData): Promise<{ avatar: string }> => {
    const response = await apiClient.post<{ avatar: string }>('/auth/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    return response.data
  }
}

// 账户管理API（仅管理员）
export const accountAPI = {
  // 获取账户列表
  getAccounts: async (params?: {
    page?: number
    page_size?: number
    search?: string
    role?: string
  }): Promise<{
    total: number
    data: AccountResponse[]
    page: number
    page_size: number
  }> => {
    const response = await apiClient.get('/accounts', { params })
    return response.data
  },

  // 创建账户
  createAccount: async (data: {
    username: string
    password: string
    email: string
    role?: string
    gitlab_personal_access_token?: string
  }): Promise<AccountResponse> => {
    const response = await apiClient.post<AccountResponse>('/accounts', data)
    return response.data
  },

  // 更新账户
  updateAccount: async (id: number, data: {
    email?: string
    role?: string
    is_active?: boolean
    gitlab_personal_access_token?: string
  }): Promise<AccountResponse> => {
    const response = await apiClient.put<AccountResponse>(`/accounts/${id}`, data)
    return response.data
  },

  // 删除账户
  deleteAccount: async (id: number): Promise<void> => {
    await apiClient.delete(`/accounts/${id}`)
  },

  // 重置密码（管理员）
  resetPassword: async (id: number, newPassword: string): Promise<void> => {
    await apiClient.post(`/accounts/${id}/reset-password`, { new_password: newPassword })
  }
}
