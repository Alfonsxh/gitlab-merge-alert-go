<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">项目管理</h1>
      <el-button type="primary" @click="showAddModal" size="large">
        <el-icon><Plus /></el-icon>
        添加项目
      </el-button>
    </div>
    
    <el-card>
      <el-empty v-if="projects.length === 0" description="还没有添加任何项目">
        <el-button type="primary" @click="showAddModal">开始添加项目</el-button>
      </el-empty>
      
      <div v-else>
        <el-alert 
          title=""
          type="info" 
          :closable="false"
          show-icon
          class="mb-4"
        >
          找到 {{ projects.length }} 个项目，分组数量：{{ Object.keys(groupedProjects).length }}
        </el-alert>
        
        <!-- 按组分组显示项目 -->
        <el-collapse v-model="activeGroups">
          <el-collapse-item
            v-for="(groupProjects, groupName) in groupedProjects"
            :key="groupName"
            :name="groupName"
          >
            <template #title>
              <div class="group-header">
                <span class="group-name">{{ groupName }}</span>
                <el-badge :value="groupProjects.length" type="primary" />
              </div>
            </template>
            
            <el-table
              :data="groupProjects"
              size="small"
              stripe
            >
              <el-table-column label="ID" prop="id" width="60" />
              
              <el-table-column label="项目名称" prop="name" min-width="200">
                <template #default="{ row }">
                  <el-link :href="row.url" target="_blank" type="primary">
                    {{ row.name }}
                    <el-icon><Link /></el-icon>
                  </el-link>
                </template>
              </el-table-column>
              
              <el-table-column label="GitLab项目ID" prop="gitlab_project_id" width="120" />
              
              <el-table-column label="描述" prop="description" show-overflow-tooltip />
              
              <el-table-column label="GitLab Webhook" width="150">
                <template #default="{ row }">
                  <el-space>
                    <el-tag 
                      :type="row.webhook_synced ? 'success' : 'warning'"
                      size="small"
                    >
                      {{ row.webhook_synced ? '已同步' : '未同步' }}
                    </el-tag>
                    <el-button
                      v-if="row.auto_manage_webhook"
                      size="small"
                      link
                      type="primary"
                      @click="syncGitLabWebhook(row)"
                      :loading="syncingWebhook === row.id"
                    >
                      <el-icon><Refresh /></el-icon>
                    </el-button>
                  </el-space>
                </template>
              </el-table-column>
              
              <el-table-column label="关联Webhook" width="200">
                <template #default="{ row }">
                  <el-space v-if="row.webhooks?.length" wrap>
                    <el-tag 
                      v-for="webhook in row.webhooks" 
                      :key="webhook.id" 
                      size="small"
                      type="info"
                    >
                      {{ webhook.name }}
                    </el-tag>
                  </el-space>
                  <span v-else class="text-muted">无</span>
                </template>
              </el-table-column>
              
              <el-table-column label="创建时间" prop="created_at" width="160">
                <template #default="{ row }">
                  {{ formatDate(row.created_at) }}
                </template>
              </el-table-column>
              
              <el-table-column label="操作" width="240" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" size="small" @click="manageWebhooks(row)">
                    <el-icon><Connection /></el-icon>
                    管理Webhook
                  </el-button>
                  <el-button link type="primary" size="small" @click="editProject(row)">
                    <el-icon><Edit /></el-icon>
                    编辑
                  </el-button>
                  <el-popconfirm
                    title="确定要删除这个项目吗？"
                    confirm-button-text="确定"
                    cancel-button-text="取消"
                    @confirm="deleteProject(row.id)"
                  >
                    <template #reference>
                      <el-button link type="danger" size="small">
                        <el-icon><Delete /></el-icon>
                        删除
                      </el-button>
                    </template>
                  </el-popconfirm>
                </template>
              </el-table-column>
            </el-table>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-card>
    
    <!-- 添加/编辑项目对话框 -->
    <el-dialog
      v-model="projectModalVisible"
      :title="isEditing ? '编辑项目' : '添加项目'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="projectFormRef"
        :model="currentProject"
        :rules="projectRules"
        label-width="120px"
      >
        <el-form-item label="项目URL" prop="url">
          <el-input
            v-model="currentProject.url"
            placeholder="例如: https://gitlab.com/group/project"
            @blur="parseProjectUrl"
          >
            <template #prefix>
              <el-icon><Link /></el-icon>
            </template>
          </el-input>
          <div class="form-item-help">
            输入项目URL后，系统会自动解析项目信息
          </div>
        </el-form-item>
        
        <el-form-item label="项目名称" prop="name">
          <el-input
            v-model="currentProject.name"
            placeholder="项目名称"
            :disabled="autoFilled"
          >
            <template #prefix>
              <el-icon><FolderOpened /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="GitLab项目ID" prop="gitlab_project_id">
          <el-input-number
            v-model="currentProject.gitlab_project_id"
            :min="1"
            placeholder="GitLab项目ID"
            :disabled="autoFilled"
            style="width: 100%"
          />
        </el-form-item>
        
        <el-form-item label="项目描述" prop="description">
          <el-input
            v-model="currentProject.description"
            type="textarea"
            :rows="3"
            placeholder="项目描述（可选）"
          />
        </el-form-item>
        
        <el-form-item label="自动管理Webhook">
          <el-switch v-model="currentProject.auto_manage_webhook" />
          <div class="form-item-help">
            开启后，系统会自动在GitLab上创建和管理Webhook
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="projectModalVisible = false">取消</el-button>
        <el-button type="primary" @click="saveProject" :loading="submitting">
          {{ isEditing ? '更新' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
    
    <!-- Webhook管理对话框 -->
    <el-dialog
      v-model="webhookModalVisible"
      title="管理项目Webhook"
      width="600px"
    >
      <el-form v-if="managingProject">
        <el-form-item label="项目名称">
          <el-tag type="info">{{ managingProject.name }}</el-tag>
        </el-form-item>
        
        <el-form-item label="选择Webhook">
          <el-checkbox-group v-model="selectedWebhookIds">
            <el-checkbox
              v-for="webhook in availableWebhooks"
              :key="webhook.id"
              :label="webhook.id"
              :value="webhook.id"
            >
              {{ webhook.name }}
              <el-tag size="small" type="info" style="margin-left: 8px">
                {{ webhook.url }}
              </el-tag>
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="webhookModalVisible = false">取消</el-button>
        <el-button type="primary" @click="saveProjectWebhooks" :loading="submitting">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  Link,
  Connection,
  Refresh,
  FolderOpened
} from '@element-plus/icons-vue'
import { projectsApi, webhooksApi } from '@/api'
import type { Project, Webhook } from '@/api'
import { formatDate } from '@/utils/format'

const projects = ref<Project[]>([])
const availableWebhooks = ref<Webhook[]>([])
const loading = ref(false)
const syncingWebhook = ref<number | null>(null)
const projectModalVisible = ref(false)
const webhookModalVisible = ref(false)
const submitting = ref(false)
const isEditing = ref(false)
const autoFilled = ref(false)
const managingProject = ref<Project | null>(null)
const selectedWebhookIds = ref<number[]>([])
const activeGroups = ref<string[]>([])

const projectFormRef = ref<FormInstance>()

const currentProject = reactive<Partial<Project>>({
  name: '',
  url: '',
  gitlab_project_id: undefined,
  description: '',
  auto_manage_webhook: true
})

const projectRules = {
  url: [
    { required: true, message: '请输入项目URL', trigger: 'blur' },
    { pattern: /^https?:\/\/.+/, message: '请输入有效的URL', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入项目名称', trigger: 'blur' }
  ],
  gitlab_project_id: [
    { required: true, message: '请输入GitLab项目ID', trigger: 'blur' }
  ]
}

const groupedProjects = computed(() => {
  const groups: Record<string, Project[]> = {}
  
  projects.value.forEach(project => {
    // 从项目URL中提取组名
    const match = project.url.match(/\/\/[^\/]+\/([^\/]+)/)
    const groupName = match ? match[1] : '其他'
    
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push(project)
  })
  
  // 按组名排序
  const sortedGroups: Record<string, Project[]> = {}
  Object.keys(groups).sort().forEach(groupName => {
    sortedGroups[groupName] = groups[groupName].sort((a, b) => a.name.localeCompare(b.name))
  })
  
  return sortedGroups
})

const loadProjects = async () => {
  loading.value = true
  try {
    const res = await projectsApi.getProjects()
    projects.value = res.data || []
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    loading.value = false
  }
}

const loadWebhooks = async () => {
  try {
    const res = await webhooksApi.getWebhooks()
    availableWebhooks.value = res.data || []
  } catch (error) {
    // 错误已在 API 客户端处理
  }
}

const showAddModal = () => {
  Object.assign(currentProject, {
    id: undefined,
    name: '',
    url: '',
    gitlab_project_id: undefined,
    description: '',
    auto_manage_webhook: true
  })
  isEditing.value = false
  autoFilled.value = false
  projectModalVisible.value = true
}

const editProject = (project: Project) => {
  Object.assign(currentProject, project)
  isEditing.value = true
  autoFilled.value = false
  projectModalVisible.value = true
}

const parseProjectUrl = async () => {
  if (!currentProject.url || isEditing.value) return
  
  try {
    const res = await projectsApi.parseProjectUrl(currentProject.url)
    if (res.data) {
      currentProject.name = res.data.name
      currentProject.gitlab_project_id = res.data.gitlab_project_id
      autoFilled.value = true
      ElMessage.success('项目信息解析成功')
    }
  } catch (error) {
    autoFilled.value = false
  }
}

const saveProject = async () => {
  const valid = await projectFormRef.value?.validate().catch(() => false)
  if (!valid) return
  
  submitting.value = true
  try {
    if (isEditing.value && currentProject.id) {
      await projectsApi.updateProject(currentProject.id, currentProject)
    } else {
      await projectsApi.createProject(currentProject)
    }
    
    ElMessage.success(isEditing.value ? '更新成功' : '添加成功')
    projectModalVisible.value = false
    await loadProjects()
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    submitting.value = false
  }
}

const deleteProject = async (projectId: number) => {
  try {
    await projectsApi.deleteProject(projectId)
    ElMessage.success('删除成功')
    await loadProjects()
  } catch (error) {
    // 错误已在 API 客户端处理
  }
}

const manageWebhooks = async (project: Project) => {
  managingProject.value = project
  selectedWebhookIds.value = project.webhooks?.map(w => w.id) || []
  await loadWebhooks()
  webhookModalVisible.value = true
}

const saveProjectWebhooks = async () => {
  if (!managingProject.value) return
  
  submitting.value = true
  try {
    // 先清除所有关联
    const currentWebhookIds = managingProject.value.webhooks?.map(w => w.id) || []
    for (const webhookId of currentWebhookIds) {
      await projectsApi.deleteProjectWebhook(managingProject.value.id, webhookId)
    }
    
    // 添加新的关联
    for (const webhookId of selectedWebhookIds.value) {
      await projectsApi.createProjectWebhook({
        project_id: managingProject.value.id,
        webhook_id: webhookId
      })
    }
    
    ElMessage.success('Webhook关联更新成功')
    webhookModalVisible.value = false
    await loadProjects()
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    submitting.value = false
  }
}

const syncGitLabWebhook = async (project: Project) => {
  syncingWebhook.value = project.id
  try {
    await projectsApi.syncGitLabWebhook(project.id)
    ElMessage.success('同步成功')
    await loadProjects()
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    syncingWebhook.value = null
  }
}

onMounted(() => {
  loadProjects()
})
</script>

<style scoped lang="less">
.mb-4 {
  margin-bottom: 16px;
}

.el-card {
  :deep(.el-card__body) {
    padding: 20px;
  }
}

:deep(.el-alert) {
  .el-alert__content {
    font-weight: 500;
  }
}

:deep(.el-collapse) {
  border: none;
  
  .el-collapse-item {
    margin-bottom: 12px;
    border: 1px solid #e6e8eb;
    border-radius: 8px;
    overflow: hidden;
    
    &:last-child {
      margin-bottom: 0;
    }
    
    .el-collapse-item__header {
      background: #f5f7fa;
      padding: 16px 20px;
      height: auto;
      line-height: normal;
      
      &:hover {
        background: #eff2f7;
      }
    }
    
    .el-collapse-item__content {
      padding: 0;
    }
  }
}

.group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  
  .group-name {
    font-weight: 600;
    font-size: 16px;
    color: #303133;
    display: flex;
    align-items: center;
    gap: 8px;
    
    &::before {
      content: '';
      width: 4px;
      height: 16px;
      background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
      border-radius: 2px;
    }
  }
}

:deep(.el-table) {
  .el-link {
    font-weight: 500;
    
    .el-icon {
      margin-left: 4px;
    }
  }
}

.text-muted {
  color: #909399;
}

.form-item-help {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

:deep(.el-dialog) {
  .el-dialog__header {
    border-bottom: 1px solid #e6e8eb;
    padding: 20px;
  }
  
  .el-dialog__body {
    padding: 30px 20px;
  }
  
  .el-dialog__footer {
    border-top: 1px solid #e6e8eb;
    padding: 20px;
  }
  
  .el-checkbox-group {
    display: flex;
    flex-direction: column;
    gap: 12px;
    
    .el-checkbox {
      display: flex;
      align-items: center;
      width: 100%;
      
      .el-checkbox__label {
        display: flex;
        align-items: center;
        flex: 1;
      }
    }
  }
}

// 响应式设计
@media screen and (max-width: 768px) {
  :deep(.el-table) {
    font-size: 12px;
  }
  
  :deep(.el-dialog) {
    width: 90% !important;
  }
}
</style>