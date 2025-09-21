import apiClient from './client'

export interface Project {
  id: number
  name: string
  url: string
  gitlab_project_id: number
  description?: string
  webhook_synced: boolean
  created_at: string
  updated_at: string
  webhooks?: any[]
}

export interface ProjectUpdatePayload extends Partial<Project> {
  webhook_ids?: number[]
}

export const projectsApi = {
  getProjects(params?: { page?: number; page_size?: number }) {
    return apiClient.get<any, { data: Project[]; total: number }>('/projects', { params })
  },

  createProject(project: Partial<Project>) {
    return apiClient.post<any, { data: Project }>('/projects', project)
  },

  updateProject(id: number, project: ProjectUpdatePayload) {
    return apiClient.put<any, { data: Project }>(`/projects/${id}`, project)
  },

  deleteProject(id: number) {
    return apiClient.delete(`/projects/${id}`)
  },

  parseProjectUrl(url: string) {
    return apiClient.post('/projects/parse-url', { url })
  },

  scanGroupProjects(url: string, access_token?: string) {
    const payload: Record<string, any> = { url }
    if (access_token) {
      payload.access_token = access_token
    }
    return apiClient.post('/projects/scan-group', payload)
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

  batchCheckWebhookStatus() {
    return apiClient.post('/projects/batch-check-webhook-status')
  },

  createProjectWebhook(data: { project_id: number; webhook_id: number }) {
    return apiClient.post('/project-webhooks', data)
  },
  
  deleteProjectWebhook(projectId: number, webhookId: number) {
    return apiClient.delete(`/project-webhooks/${projectId}/${webhookId}`)
  }
}
