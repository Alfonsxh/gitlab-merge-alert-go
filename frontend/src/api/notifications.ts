import apiClient from './client'

export interface Notification {
  id: number
  project_id: number
  project_name: string
  merge_request_id: number
  title: string
  source_branch: string
  target_branch: string
  author_email: string
  assignee_emails?: string[]
  notification_sent: boolean
  created_at: string
}

export const notificationsApi = {
  getNotifications(params?: { page_size?: number }) {
    return apiClient.get<any, { data: Notification[] }>('/notifications', { params })
  }
}