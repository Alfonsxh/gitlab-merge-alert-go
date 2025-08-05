import apiClient from './client'

export interface Stats {
  total_users: number
  total_projects: number
  total_webhooks: number
  total_notifications: number
}

export interface DailyStats {
  date: string
  count: number
}

export interface ProjectDailyStats {
  project_id: number
  project_name: string
  data: DailyStats[]
}

export interface WebhookDailyStats {
  webhook_id: number
  webhook_name: string
  data: DailyStats[]
}

export const statsApi = {
  getStats() {
    return apiClient.get<any, { data: Stats }>('/stats')
  },
  
  getProjectDailyStats(days?: number) {
    const params = days ? { days } : {}
    return apiClient.get<any, { data: ProjectDailyStats[] }>('/stats/projects/daily', { params })
  },
  
  getWebhookDailyStats(days?: number) {
    const params = days ? { days } : {}
    return apiClient.get<any, { data: WebhookDailyStats[] }>('/stats/webhooks/daily', { params })
  }
}