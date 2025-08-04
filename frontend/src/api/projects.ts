import apiClient from './client'

export interface Project {
  id: number
  name: string
  url: string
  gitlab_project_id: number
  description?: string
  webhook_synced: boolean
  auto_manage_webhook: boolean
  created_at: string
  updated_at: string
  webhooks?: any[]
}

export const projectsApi = {
  getProjects() {
    return apiClient.get<any, { data: Project[] }>('/projects')
  },

  createProject(project: Partial<Project>) {
    return apiClient.post<any, { data: Project }>('/projects', project)
  },

  updateProject(id: number, project: Partial<Project>) {
    return apiClient.put<any, { data: Project }>(`/projects/${id}`, project)
  },

  deleteProject(id: number) {
    return apiClient.delete(`/projects/${id}`)
  },

  parseProjectUrl(url: string) {
    return apiClient.post('/projects/parse-url', { url })
  },

  scanGroupProjects(url: string, access_token: string) {
    return apiClient.post('/projects/scan-group', { url, access_token })
  },

  batchCreateProjects(data: any) {
    return apiClient.post('/projects/batch-create', data)
  },

  syncGitLabWebhook(id: number) {
    return apiClient.post(`/projects/${id}/sync-gitlab-webhook`)
  },

  deleteGitLabWebhook(id: number) {
    return apiClient.delete(`/projects/${id}/sync-gitlab-webhook`)
  },

  getGitLabWebhookStatus(id: number) {
    return apiClient.get(`/projects/${id}/gitlab-webhook-status`)
  },
  
  createProjectWebhook(data: { project_id: number; webhook_id: number }) {
    return apiClient.post('/project-webhooks', data)
  },
  
  deleteProjectWebhook(projectId: number, webhookId: number) {
    return apiClient.delete(`/project-webhooks/${projectId}/${webhookId}`)
  }
}