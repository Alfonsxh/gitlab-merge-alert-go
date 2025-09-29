import apiClient from './client'

export type WebhookType = 'wechat' | 'dingtalk' | 'custom' | 'auto'

export interface Webhook {
  id: number
  name: string
  url: string
  description?: string
  type: WebhookType
  signature_method?: string
  secret?: string
  security_keywords?: string[]
  custom_headers?: Record<string, string>
  is_active: boolean
  created_at: string
  updated_at: string
  projects?: any[]
}

export type UpsertWebhookPayload = Partial<Omit<Webhook, 'id' | 'created_at' | 'updated_at'>>

export const webhooksApi = {
  getWebhooks(params?: { page?: number; page_size?: number }) {
    return apiClient.get<any, { data: Webhook[]; total: number }>('/webhooks', { params })
  },

  createWebhook(webhook: UpsertWebhookPayload) {
    return apiClient.post<any, { data: Webhook }>('/webhooks', webhook)
  },

  updateWebhook(id: number, webhook: UpsertWebhookPayload) {
    return apiClient.put<any, { data: Webhook }>(`/webhooks/${id}`, webhook)
  },

  deleteWebhook(id: number) {
    return apiClient.delete(`/webhooks/${id}`)
  },

  sendTestMessage(id: number) {
    return apiClient.post<any, { message: string; webhook_name: string; sent_at: string; channel?: string }>(`/webhooks/${id}/test`)
  }
}
