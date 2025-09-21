import { apiClient } from './client'

export interface BootstrapStatusResponse {
  admin_setup_required: boolean
}

export interface SetupAdminPayload {
  token: string
  email: string
  password: string
}

export const systemAPI = {
  async getBootstrapStatus(): Promise<BootstrapStatusResponse> {
    const response = await apiClient.get<BootstrapStatusResponse>('/system/bootstrap')
    return response as unknown as BootstrapStatusResponse
  },

  async setupAdmin(payload: SetupAdminPayload): Promise<void> {
    await apiClient.post('/system/setup-admin', payload)
  }
}
