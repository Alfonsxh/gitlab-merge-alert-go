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
  getWebhooks() {
    return apiClient.get<any, { data: Webhook[] }>('/webhooks')
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