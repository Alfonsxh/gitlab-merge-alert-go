import { apiClient } from './client'

export interface AssignManagerRequest {
  resource_id: number
  resource_type: 'project' | 'webhook' | 'user'
  manager_id: number
}

export interface ResourceManagerResponse {
  id: number
  resource_id: number
  resource_type: string
  manager_id: number
  manager?: any
  created_at: string
}

export const resourceManagerAPI = {
  // 分配管理员
  assignManager: async (data: AssignManagerRequest): Promise<void> => {
    await apiClient.post('/resource-managers/assign', data)
  },

  // 移除管理员
  removeManager: async (data: {
    resource_id: number
    resource_type: string
    manager_id: number
  }): Promise<void> => {
    await apiClient.post('/resource-managers/remove', data)
  },

  // 获取资源的管理员列表
  getResourceManagers: async (resourceId: number, resourceType: string): Promise<{
    managers: ResourceManagerResponse[]
    total: number
  }> => {
    return await apiClient.get('/resource-managers', {
      params: {
        resource_id: resourceId,
        resource_type: resourceType
      }
    })
  },

  // 获取账户管理的资源列表
  getManagedResources: async (accountId: number, resourceType: string): Promise<{
    resource_ids: number[]
    total: number
  }> => {
    return await apiClient.get(`/resource-managers/managed/${accountId}`, {
      params: {
        resource_type: resourceType
      }
    })
  },

  // 批量分配资源
  batchAssign: async (accountId: number, assignments: AssignManagerRequest[]): Promise<void> => {
    await apiClient.post(`/resource-managers/batch-assign/${accountId}`, {
      assignments
    })
  }
}