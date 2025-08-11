import apiClient from './client'

export interface Webhook {
  id: number
  name: string
  url: string
  description?: string
  is_active: boolean
  created_at: string
  updated_at: string
  projects?: any[]
}

export const webhooksApi = {
  getWebhooks(params?: { page?: number; page_size?: number }) {
    return apiClient.get<any, { data: Webhook[]; total: number }>('/webhooks', { params })
  },

  createWebhook(webhook: Partial<Webhook>) {
    return apiClient.post<any, { data: Webhook }>('/webhooks', webhook)
  },

  updateWebhook(id: number, webhook: Partial<Webhook>) {
    return apiClient.put<any, { data: Webhook }>(`/webhooks/${id}`, webhook)
  },

  deleteWebhook(id: number) {
    return apiClient.delete(`/webhooks/${id}`)
  }
}