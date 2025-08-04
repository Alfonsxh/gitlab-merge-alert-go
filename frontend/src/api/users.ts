import apiClient from './client'

export interface User {
  id: number
  email: string
  phone: string
  gitlab_username?: string
  created_at: string
  updated_at: string
}

export const usersApi = {
  getUsers() {
    return apiClient.get<any, { data: User[] }>('/users')
  },

  createUser(user: Partial<User>) {
    return apiClient.post<any, { data: User }>('/users', user)
  },

  updateUser(id: number, user: Partial<User>) {
    return apiClient.put<any, { data: User }>(`/users/${id}`, user)
  },

  deleteUser(id: number) {
    return apiClient.delete(`/users/${id}`)
  }
}