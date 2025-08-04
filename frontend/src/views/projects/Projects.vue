<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">项目管理</h2>
      <el-button type="primary" @click="showAddModal">
        <el-icon><Plus /></el-icon>
        添加项目
      </el-button>
    </div>
    
    <el-card>
      <el-empty v-if="projects.length === 0" description="还没有添加任何项目" class="empty-container">
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
              style="width: 100%;"
            >
              <el-table-column label="GitLab项目ID" prop="gitlab_project_id" width="150" />
              
              <el-table-column label="项目名称" prop="name" min-width="300">
                <template #default="{ row }">
                  <el-link :href="row.url" target="_blank" type="primary">
                    {{ row.name }}
                    <el-icon><Link /></el-icon>
                  </el-link>
                </template>
              </el-table-column>
              
              <el-table-column label="描述" prop="description" min-width="250" show-overflow-tooltip />
              
              <el-table-column label="GitLab Webhook" width="180">
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
              
              <el-table-column label="关联Webhook" min-width="200">
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
              
              <el-table-column label="创建时间" prop="created_at" width="180">
                <template #default="{ row }">
                  {{ formatDate(row.created_at) }}
                </template>
              </el-table-column>
              
              <el-table-column label="操作" width="320" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="manageWebhooks(row)">
                    <el-icon><Connection /></el-icon>
                    管理Webhook
                  </el-button>
                  <el-button link type="primary" @click="editProject(row)">
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
                      <el-button link type="danger">
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
      width="650px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="projectFormRef"
        :model="currentProject"
        :rules="projectRules"
        label-width="140px"
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
          <div style="display: flex; align-items: center; width: 100%;">
            <el-switch v-model="currentProject.auto_manage_webhook" style="margin-right: 12px;" />
            <span style="font-size: 13px; color: #909399; line-height: 1.2;">
              开启后，系统会自动在GitLab上创建和管理Webhook
            </span>
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
      width="700px"
      :close-on-click-modal="false"
    >
      <div v-if="managingProject" class="webhook-manage-dialog">
        <div class="project-info-section">
          <div class="project-label">当前项目</div>
          <div class="project-details">
            <div class="project-name">
              <el-icon><FolderOpened /></el-icon>
              {{ managingProject.name }}
            </div>
            <div class="project-url">{{ managingProject.url }}</div>
          </div>
        </div>
        
        <div class="webhook-selection">
          <div class="section-title">选择要关联的企业微信机器人</div>
          <el-empty v-if="availableWebhooks.length === 0" description="暂无可用的Webhook">
            <el-button type="primary" @click="goToWebhooks">前往创建</el-button>
          </el-empty>
          
          <div v-else class="webhook-list">
            <div
              v-for="webhook in availableWebhooks"
              :key="webhook.id"
              class="webhook-item"
              :class="{ selected: selectedWebhookIds.includes(webhook.id) }"
              @click="toggleWebhook(webhook.id)"
            >
              <el-checkbox
                :model-value="selectedWebhookIds.includes(webhook.id)"
                @change="toggleWebhook(webhook.id)"
                class="webhook-checkbox"
              />
              <div class="webhook-content">
                <div class="webhook-main">
                  <el-icon class="webhook-icon"><Connection /></el-icon>
                  <span class="webhook-name">{{ webhook.name }}</span>
                </div>
                <div class="webhook-url">{{ formatWebhookUrl(webhook.url) }}</div>
              </div>
              <el-tag
                v-if="selectedWebhookIds.includes(webhook.id)"
                type="success"
                size="small"
                class="selected-tag"
              >
                已选择
              </el-tag>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <el-button @click="webhookModalVisible = false">取消</el-button>
        <el-button type="primary" @click="saveProjectWebhooks" :loading="submitting">
          保存（{{ selectedWebhookIds.length }}）
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { useRouter } from 'vue-router'
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

const router = useRouter()
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
    // 从项目URL中提取完整的组路径
    // 例如: https://gitlab.woqutech.com/QDataPro/qpatchs/Q-20220721-001
    // 提取: QDataPro/qpatchs
    const urlParts = project.url.split('/')
    const hostIndex = urlParts.findIndex(part => part.includes('.'))
    if (hostIndex >= 0 && urlParts.length > hostIndex + 2) {
      // 获取除了最后一个部分（项目名）之外的所有路径部分
      const pathParts = urlParts.slice(hostIndex + 1, -1)
      const groupName = pathParts.join('/')
      
      if (!groups[groupName]) {
        groups[groupName] = []
      }
      groups[groupName].push(project)
    } else {
      // 无法解析的URL，放入"其他"组
      if (!groups['其他']) {
        groups['其他'] = []
      }
      groups['其他'].push(project)
    }
  })
  
  // 按组名排序，组内按 GitLab项目ID 增序排列
  const sortedGroups: Record<string, Project[]> = {}
  Object.keys(groups).sort().forEach(groupName => {
    sortedGroups[groupName] = groups[groupName].sort((a, b) => a.gitlab_project_id - b.gitlab_project_id)
  })
  
  return sortedGroups
})

const loadProjects = async () => {
  loading.value = true
  try {
    const res = await projectsApi.getProjects()
    projects.value = res.data || []
    
    // 默认展开第一个组
    const firstGroup = Object.keys(groupedProjects.value)[0]
    if (firstGroup && activeGroups.value.length === 0) {
      activeGroups.value = [firstGroup]
    }
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

const toggleWebhook = (webhookId: number) => {
  const index = selectedWebhookIds.value.indexOf(webhookId)
  if (index > -1) {
    selectedWebhookIds.value.splice(index, 1)
  } else {
    selectedWebhookIds.value.push(webhookId)
  }
}

const formatWebhookUrl = (url: string) => {
  // 保留URL的主要部分，隐藏敏感的key参数
  const keyIndex = url.indexOf('key=')
  if (keyIndex > -1) {
    // 保留key=之前的部分，加上省略号
    return url.substring(0, keyIndex + 4) + '...'
  }
  // 如果没有key参数，返回原URL（但限制最大长度）
  if (url.length > 60) {
    return url.substring(0, 60) + '...'
  }
  return url
}

const goToWebhooks = () => {
  webhookModalVisible.value = false
  router.push('/webhooks')
}

onMounted(() => {
  loadProjects()
})
</script>

<style scoped lang="less">
.page-container {
  // 样式已在主布局中定义
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
  
  .page-title {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: #303133;
    display: flex;
    align-items: center;
    
    &::before {
      content: '';
      width: 4px;
      height: 20px;
      background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
      border-radius: 2px;
      margin-right: 12px;
    }
  }
}

.mb-4 {
  margin-bottom: 16px;
}

.el-card {
  max-height: calc(100vh - 200px);
  display: flex;
  flex-direction: column;
  
  :deep(.el-card__body) {
    padding: 20px;
    overflow: auto;
    
    &::-webkit-scrollbar {
      width: 8px;
    }
    
    &::-webkit-scrollbar-track {
      background: #f5f7fa;
    }
    
    &::-webkit-scrollbar-thumb {
      background: #dcdfe6;
      border-radius: 4px;
      
      &:hover {
        background: #c0c4cc;
      }
    }
  }
}

.empty-container {
  padding: 60px 20px;
}

.projects-content {
  // 内容会随高度自适应
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
    
    &::before {
      content: '';
      width: 4px;
      height: 16px;
      background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
      border-radius: 2px;
      margin-right: 8px;
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
  
  // 增加操作按钮的大小
  .el-button {
    font-size: 14px !important;
    
    .el-icon {
      font-size: 16px !important;
      margin-right: 4px;
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

// Webhook管理对话框样式
.webhook-manage-dialog {
  .project-info-section {
    margin-bottom: 24px;
    padding: 20px;
    background: linear-gradient(135deg, #f5f8ff 0%, #f0f4ff 100%);
    border-radius: 10px;
    border: 1px solid #e6e8eb;
    
    .project-label {
      font-size: 13px;
      color: #909399;
      font-weight: 500;
      letter-spacing: 0.5px;
      text-transform: uppercase;
      margin-bottom: 8px;
    }
    
    .project-details {
      .project-name {
        display: flex;
        align-items: center;
        gap: 10px;
        font-size: 20px;
        font-weight: 600;
        color: #303133;
        margin-bottom: 6px;
        
        .el-icon {
          font-size: 24px;
          color: #409eff;
        }
      }
      
      .project-url {
        font-size: 13px;
        color: #606266;
        font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
        padding-left: 34px;
      }
    }
  }
  
  .webhook-selection {
    .section-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      margin-bottom: 16px;
      padding-left: 4px;
      border-left: 3px solid #409eff;
      padding-left: 12px;
    }
    
    .webhook-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
      
      .webhook-item {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        padding: 16px;
        border: 1px solid #e6e8eb;
        border-radius: 8px;
        cursor: pointer;
        transition: all 0.3s;
        position: relative;
        
        &:hover {
          border-color: #409eff;
          background: #f8f9fb;
          box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
        }
        
        &.selected {
          border-color: #409eff;
          background: #ecf5ff;
          box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
        }
        
        .webhook-checkbox {
          margin-top: 2px;
        }
        
        .webhook-content {
          flex: 1;
          
          .webhook-main {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 6px;
            
            .webhook-icon {
              font-size: 18px;
              color: #409eff;
            }
            
            .webhook-name {
              font-size: 15px;
              font-weight: 500;
              color: #303133;
            }
          }
          
          .webhook-url {
            font-size: 12px;
            color: #909399;
            font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
            word-break: break-all;
            line-height: 1.4;
            padding-left: 26px;
          }
        }
        
        .selected-tag {
          position: absolute;
          top: 16px;
          right: 16px;
        }
      }
    }
  }
}

// 表格列宽度优化
:deep(.el-table) {
  // 确保表格能够自适应容器宽度
  table-layout: fixed;
  width: 100%;
  
  // 优化长文本显示
  .cell {
    word-break: break-word;
    word-wrap: break-word;
  }
  
  // 优化固定列的显示
  .el-table__fixed-right {
    box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1);
  }
}

// 响应式设计
@media screen and (max-width: 1200px) {
  :deep(.el-table) {
    // 在中等屏幕上调整项目名称列宽度
    .el-table__body .el-table__row td:nth-child(2) {
      width: 28% !important;
    }
    // 调整描述列宽度
    .el-table__body .el-table__row td:nth-child(3) {
      width: 22% !important;
    }
    // 调整关联Webhook列宽度
    .el-table__body .el-table__row td:nth-child(5) {
      width: 18% !important;
    }
  }
}

@media screen and (max-width: 768px) {
  :deep(.el-table) {
    font-size: 12px;
    
    // 在小屏幕上隐藏描述列和GitLab Webhook列
    .el-table__header th:nth-child(3),
    .el-table__body td:nth-child(3),
    .el-table__header th:nth-child(4),
    .el-table__body td:nth-child(4) {
      display: none;
    }
    
    // 调整项目名称列在小屏幕上的显示
    .el-table__body .el-table__row td:nth-child(2) {
      width: 40% !important;
    }
    
    // 调整关联Webhook列在小屏幕上的显示
    .el-table__body .el-table__row td:nth-child(5) {
      width: 35% !important;
    }
  }
  
  :deep(.el-dialog) {
    width: 90% !important;
  }
}
</style>