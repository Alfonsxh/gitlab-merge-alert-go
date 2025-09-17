import apiClient from './client'

export interface GitLabConfig {
  gitlab_url: string
}

export const gitlabApi = {
  testConnection(data: { url: string; access_token?: string }) {
    return apiClient.post('/gitlab/test-connection', data)
  },

  getConfig() {
    return apiClient.get<any, { data: GitLabConfig }>('/gitlab/config')
  },

  testToken(data: { access_token?: string; gitlab_url?: string }) {
    return apiClient.post('/gitlab/test-token', data)
  }
}
