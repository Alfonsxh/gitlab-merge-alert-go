import apiClient from './client'

export interface Stats {
  total_users: number
  total_projects: number
  total_webhooks: number
  total_notifications: number
}

export const statsApi = {
  getStats() {
    return apiClient.get<any, { data: Stats }>('/stats')
  }
}