<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">项目管理</h2>
      <el-space>
        <el-button
          v-if="projects.length > 0"
          @click="batchCheckWebhookStatus"
          :loading="batchChecking"
        >
          <el-icon><Refresh /></el-icon>
          批量刷新Webhook状态
        </el-button>
        <el-button type="primary" @click="showAddModal">
          <el-icon><Plus /></el-icon>
          添加项目
        </el-button>
      </el-space>
    </div>

    <el-alert
      v-if="!hasGitLabToken"
      class="mb-4"
      type="warning"
      show-icon
      title="未配置 GitLab Personal Access Token"
    >
      <template #description>
        <span>项目操作需要 GitLab 访问令牌，请先前往个人中心或账户管理配置 Token。</span>
        <el-button type="primary" link @click="goToProfile" style="margin-left: 8px;">立即配置</el-button>
      </template>
    </el-alert>

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
    
    <!-- 添加项目对话框（支持单项目和批量导入） -->
    <el-dialog
      v-model="projectModalVisible"
      :title="isGroupMode ? `批量导入 - ${groupInfo?.name || ''}` : '添加项目'"
      :width="isGroupMode ? '900px' : '750px'"
      :close-on-click-modal="true"
      :close-on-press-escape="true"
      :append-to-body="true"
      class="add-project-dialog"
      @closed="handleDialogClose"
    >
      <el-form
        ref="projectFormRef"
        :model="currentProject"
        label-width="80px"
        :hide-required-asterisk="true"
      >
        <!-- URL输入区域 -->
        <el-form-item label="项目URL" v-if="!isEditing">
          <div class="url-input-container">
            <el-input
              v-model="currentProject.url"
              placeholder="输入项目或组的GitLab URL"
              :disabled="parsingUrl || urlParsed"
              class="url-input"
            >
              <template #prefix>
                <el-icon><Link /></el-icon>
              </template>
            </el-input>
            <el-button
              type="primary"
              @click="parseProjectUrl"
              :loading="parsingUrl"
              :disabled="!currentProject.url || urlParsed"
              class="parse-button"
            >
              {{ urlParsed ? '重新解析' : '解析' }}
            </el-button>
          </div>
          <div class="form-item-help">
            <span v-if="!parsingUrl && !urlParsed">支持项目URL或组URL</span>
            <span v-else-if="parsingUrl" style="color: #409eff">
              <el-icon class="is-loading" style="margin-right: 4px; vertical-align: middle;"><Loading /></el-icon>
              正在检测URL类型...
            </span>
            <span v-else-if="urlParsed && isGroupMode" style="color: #67c23a">
              检测到GitLab组，发现 {{ groupProjects.length }} 个项目
            </span>
            <span v-else-if="urlParsed && !isGroupMode" style="color: #67c23a">
              项目信息已解析
            </span>
          </div>
        </el-form-item>

        <!-- 解析中状态 -->
        <div v-if="parsingUrl && !isEditing" class="parsing-container">
          <el-icon class="is-loading" :size="48" color="#409eff"><Loading /></el-icon>
          <p>正在解析URL，请稍候...</p>
        </div>

        <!-- 组模式：显示项目列表 -->
        <template v-if="isGroupMode && !parsingUrl">
          <el-alert type="info" :closable="false" class="mb-4">
            发现 {{ groupProjects.length }} 个新项目，已选择 {{ selectedGroupProjects.length }} 个
          </el-alert>

          <!-- 全选/取消全选 -->
          <div class="mb-3">
            <el-checkbox
              :model-value="selectedGroupProjects.length === groupProjects.length"
              :indeterminate="selectedGroupProjects.length > 0 && selectedGroupProjects.length < groupProjects.length"
              @change="toggleAllGroupProjects"
            >
              全选/取消全选
            </el-checkbox>
          </div>

          <!-- 项目列表 -->
          <div class="project-list">
            <div
              v-for="project in groupProjects"
              :key="project.id"
              class="project-item"
            >
              <el-checkbox
                :model-value="selectedGroupProjects.includes(project.id)"
                @change="toggleGroupProject(project.id)"
              >
                <div class="project-info">
                  <div class="project-name">{{ project.name }}</div>
                  <div class="project-path">{{ project.path_with_namespace }}</div>
                  <div v-if="project.description" class="project-desc">{{ project.description }}</div>
                </div>
              </el-checkbox>
            </div>
          </div>

          <!-- Webhook配置 -->
          <el-divider />
          <div class="webhook-config">
            <h4>Webhook 配置</h4>
            <el-radio-group v-model="webhookConfig.useUnified">
              <el-radio :label="true">为所有项目使用统一的 Webhook</el-radio>
              <el-radio :label="false">为每个项目单独配置（稍后配置）</el-radio>
            </el-radio-group>

            <el-select
              v-if="webhookConfig.useUnified"
              v-model="webhookConfig.webhookId"
              placeholder="请选择 Webhook"
              class="mt-3"
              style="width: 100%"
            >
              <el-option
                v-for="webhook in availableWebhooks"
                :key="webhook.id"
                :label="webhook.name"
                :value="webhook.id"
              >
                <span>{{ webhook.name }}</span>
                <span style="color: #999; margin-left: 10px">{{ formatWebhookUrl(webhook.url) }}</span>
              </el-option>
            </el-select>
          </div>
        </template>

        <!-- 单项目模式：显示项目信息和Webhook选择 -->
        <template v-else-if="!isGroupMode && !parsingUrl && urlParsed">
          <div class="project-info-box">
            <div class="project-info-item">
              <span class="info-label">项目名称:</span>
              <span class="info-value">{{ currentProject.name }}</span>
            </div>
            <div class="project-info-item">
              <span class="info-label">GitLab项目ID:</span>
              <span class="info-value">{{ currentProject.gitlab_project_id }}</span>
            </div>
            <div class="project-info-item" v-if="currentProject.description">
              <span class="info-label">描述:</span>
              <span class="info-value">{{ currentProject.description }}</span>
            </div>
          </div>

          <!-- Webhook配置 -->
          <el-divider />
          <div class="webhook-config">
            <h4>Webhook 配置</h4>
            <el-select
              v-model="singleProjectWebhookId"
              placeholder="请选择要关联的企业微信机器人"
              style="width: 100%"
            >
              <el-option
                v-for="webhook in availableWebhooks"
                :key="webhook.id"
                :label="webhook.name"
                :value="webhook.id"
              >
                <span>{{ webhook.name }}</span>
                <span style="color: #999; margin-left: 10px">{{ formatWebhookUrl(webhook.url) }}</span>
              </el-option>
            </el-select>
          </div>
        </template>
      </el-form>

      <template #footer v-if="urlParsed">
        <div style="text-align: center;">
          <el-button
            type="primary"
            @click="isGroupMode ? saveBatchProjects() : saveProject()"
            :loading="submitting"
            :disabled="parsingUrl"
            size="large"
          >
            {{ isGroupMode ? `批量导入 (${selectedGroupProjects.length})` : '添加项目' }}
          </el-button>
        </div>
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

    <!-- 批量导入对话框（已整合到主对话框中） -->
    <!-- <BatchImport
      v-model="batchModalVisible"
      :projects="batchProjects"
      @success="handleBatchImportSuccess"
    /> -->
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
  FolderOpened,
  Loading
} from '@element-plus/icons-vue'
import { projectsApi, webhooksApi } from '@/api'
import type { Project, Webhook } from '@/api'
import { formatDate } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'
// import BatchImport from './BatchImport.vue' // 不再需要，功能已整合到主对话框

const router = useRouter()
const authStore = useAuthStore()
const projects = ref<Project[]>([])
const availableWebhooks = ref<Webhook[]>([])
const loading = ref(false)
const syncingWebhook = ref<number | null>(null)
const batchChecking = ref(false)
const projectModalVisible = ref(false)
const webhookModalVisible = ref(false)
const submitting = ref(false)
const isEditing = ref(false)
const autoFilled = ref(false)
const managingProject = ref<Project | null>(null)
const selectedWebhookIds = ref<number[]>([])
const activeGroups = ref<string[]>([])
// 以下变量已不再需要，功能已整合到主对话框
// const batchModalVisible = ref(false)
// const batchProjects = ref<any[]>([])

// 新增状态：URL解析和组模式
const parsingUrl = ref(false)  // URL解析中
const isGroupMode = ref(false)  // 是否为组模式
const groupInfo = ref<any>(null)  // 组信息
const urlParsed = ref(false)  // URL已解析标志
const groupProjects = ref<any[]>([])  // 组内项目列表
const selectedGroupProjects = ref<number[]>([])  // 选中的组项目
const webhookConfig = reactive({
  useUnified: true,
  webhookId: null as number | null
})

const hasGitLabToken = computed(() => authStore.hasGitLabToken)

const ensureGitLabToken = () => {
  if (!hasGitLabToken.value) {
    ElMessage.warning('请先在个人中心配置 GitLab Personal Access Token')
    return false
  }
  return true
}

const goToProfile = () => {
  router.push({ path: '/profile' })
}

const projectFormRef = ref<FormInstance>()
const singleProjectWebhookId = ref<number | null>(null)  // 单项目模式的Webhook选择

const currentProject = reactive<Partial<Project>>({
  name: '',
  url: '',
  gitlab_project_id: undefined,
  description: ''
})

// 不再需要表单验证规则，通过解析按钮来验证URL

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

const showAddModal = async () => {
  if (!ensureGitLabToken()) {
    return
  }
  Object.assign(currentProject, {
    id: undefined,
    name: '',
    url: '',
    gitlab_project_id: undefined,
    description: ''
  })
  isEditing.value = false
  autoFilled.value = false
  isGroupMode.value = false
  parsingUrl.value = false
  // 加载可用的webhooks，为批量导入做准备
  await loadWebhooks()
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
  if (!ensureGitLabToken()) return

  // 如果已经解析过，先重置
  if (urlParsed.value) {
    urlParsed.value = false
    isGroupMode.value = false
    currentProject.name = ''
    currentProject.gitlab_project_id = undefined
    currentProject.description = ''
    autoFilled.value = false
    groupProjects.value = []
    selectedGroupProjects.value = []
  }

  parsingUrl.value = true
  try {
    const res: any = await projectsApi.parseProjectUrl(currentProject.url)

    // 检查响应中的 is_group 标志
    if (res.is_group) {
      // 切换到组模式
      isGroupMode.value = true
      groupInfo.value = res.data.group_info
      groupProjects.value = res.data.projects
      selectedGroupProjects.value = res.data.projects.map((p: any) => p.id) // 默认全选

      // 清空单项目表单
      currentProject.name = ''
      currentProject.gitlab_project_id = undefined
      currentProject.description = ''
      autoFilled.value = false

      ElMessage.success(`检测到 GitLab 组：${groupInfo.value.name}，发现 ${groupProjects.value.length} 个新项目`)
    } else if (res.data) {
      // 单项目模式
      isGroupMode.value = false
      currentProject.name = res.data.name
      currentProject.gitlab_project_id = res.data.gitlab_project_id
      currentProject.description = res.data.description || ''
      autoFilled.value = true
      ElMessage.success('项目信息解析成功')
    }
    urlParsed.value = true
  } catch (error) {
    autoFilled.value = false
    isGroupMode.value = false
    urlParsed.value = false
    ElMessage.warning('URL解析失败，请检查URL格式或手动填写项目信息')
  } finally {
    parsingUrl.value = false
  }
}

const saveProject = async () => {
  if (!ensureGitLabToken()) return

  submitting.value = true
  try {
    if (isEditing.value && currentProject.id) {
      await projectsApi.updateProject(currentProject.id, currentProject)
    } else {
      // 创建项目，包含webhook_id（如果选择了）
      const projectData = {
        ...currentProject,
        webhook_id: singleProjectWebhookId.value || undefined
      }
      await projectsApi.createProject(projectData)
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
  if (!ensureGitLabToken()) return
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

const batchCheckWebhookStatus = async () => {
  if (!ensureGitLabToken()) return

  batchChecking.value = true
  try {
    const { data } = await projectsApi.batchCheckWebhookStatus()

    // 更新本地项目列表的状态
    if (data.data && Array.isArray(data.data)) {
      data.data.forEach((result: any) => {
        const project = projects.value.find(p => p.id === result.project_id)
        if (project) {
          project.webhook_synced = result.webhook_synced
        }
      })
    }

    // 显示汇总信息
    const summary = data.summary
    if (summary) {
      ElMessage.success(
        `检查完成: ${summary.total} 个项目, ${summary.success} 个成功, ${summary.errors} 个失败, ${summary.status_changed} 个状态已更新`
      )
    } else {
      ElMessage.success('批量状态检查完成')
    }
  } catch (error: any) {
    console.error('批量检查失败:', error)
    ElMessage.error(error.response?.data?.error || '批量检查失败')
  } finally {
    batchChecking.value = false
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

// 不再需要这个函数，功能已整合到saveBatchProjects中
// const handleBatchImportSuccess = () => {
//   loadProjects()
// }

// 批量保存项目
const saveBatchProjects = async () => {
  if (selectedGroupProjects.value.length === 0) {
    ElMessage.warning('请至少选择一个项目')
    return
  }

  if (webhookConfig.useUnified && !webhookConfig.webhookId) {
    ElMessage.warning('请选择要关联的 Webhook')
    return
  }

  submitting.value = true
  try {
    const selectedProjectsData = groupProjects.value
      .filter((p: any) => selectedGroupProjects.value.includes(p.id))
      .map((p: any) => ({
        gitlab_project_id: p.id,
        name: p.name,
        url: p.web_url,
        description: p.description || ''
      }))

    const batchData = {
      projects: selectedProjectsData,
      webhook_config: {
        use_unified: webhookConfig.useUnified,
        unified_webhook_id: webhookConfig.webhookId
      }
    }

    const res = await projectsApi.batchCreateProjects(batchData)

    if (res.data) {
      const { success_count, failure_count } = res.data
      ElMessage.success(`批量导入完成：成功 ${success_count} 个，失败 ${failure_count} 个`)
      projectModalVisible.value = false
      await loadProjects()
    }
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    submitting.value = false
  }
}

// 切换组项目选择
const toggleGroupProject = (projectId: number) => {
  const index = selectedGroupProjects.value.indexOf(projectId)
  if (index > -1) {
    selectedGroupProjects.value.splice(index, 1)
  } else {
    selectedGroupProjects.value.push(projectId)
  }
}

// 全选/取消全选组项目
const toggleAllGroupProjects = () => {
  if (selectedGroupProjects.value.length === groupProjects.value.length) {
    selectedGroupProjects.value = []
  } else {
    selectedGroupProjects.value = groupProjects.value.map((p: any) => p.id)
  }
}

// 处理对话框关闭
const handleDialogClose = () => {
  // 重置所有状态
  isGroupMode.value = false
  groupInfo.value = null
  groupProjects.value = []
  selectedGroupProjects.value = []
  parsingUrl.value = false
  autoFilled.value = false
  urlParsed.value = false
  currentProject.url = ''
  currentProject.name = ''
  currentProject.gitlab_project_id = undefined
  currentProject.description = ''
  webhookConfig.useUnified = true
  webhookConfig.webhookId = null
  singleProjectWebhookId.value = null
  isEditing.value = false
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

.mb-3 {
  margin-bottom: 12px;
}

.mt-3 {
  margin-top: 12px;
}

.parsing-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: #909399;

  p {
    margin-top: 16px;
    font-size: 14px;
  }
}

.project-list {
  max-height: 300px;
  min-height: 150px;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 10px;

  .project-item {
    padding: 10px;
    border-bottom: 1px solid #f0f0f0;

    &:last-child {
      border-bottom: none;
    }

    .project-info {
      margin-left: 8px;

      .project-name {
        font-weight: 500;
        color: #303133;
      }

      .project-path {
        font-size: 12px;
        color: #909399;
        margin-top: 2px;
      }

      .project-desc {
        font-size: 12px;
        color: #909399;
        margin-top: 4px;
      }
    }
  }
}

.webhook-config {
  margin-top: 20px;

  h4 {
    margin-bottom: 15px;
    color: #303133;
  }
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

// 添加项目对话框样式
.add-project-dialog {
  // 覆盖对话框 wrapper 样式，确保相对视口定位
  :deep(.el-overlay) {
    position: fixed !important;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    overflow: auto;
  }

  :deep(.el-overlay-dialog) {
    position: fixed !important;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 8vh;
    padding-bottom: 8vh;
    overflow: auto;
  }

  :deep(.el-dialog) {
    margin: 0 auto !important;
    max-height: 84vh;
    display: flex;
    flex-direction: column;
    position: relative !important;
  }

  :deep(.el-dialog__body) {
    flex: 1 1 auto;
    overflow-y: auto;
    overflow-x: hidden;
    max-height: calc(84vh - 160px); // 减去header和footer的高度
    padding: 20px 24px;

    &::-webkit-scrollbar {
      width: 8px;
    }

    &::-webkit-scrollbar-track {
      background: #f5f7fa;
      border-radius: 4px;
    }

    &::-webkit-scrollbar-thumb {
      background: #dcdfe6;
      border-radius: 4px;

      &:hover {
        background: #c0c4cc;
      }
    }
  }

  :deep(.el-dialog__header) {
    flex-shrink: 0;
    padding: 20px 24px;
    border-bottom: 1px solid #ebeef5;
  }

  :deep(.el-dialog__footer) {
    flex-shrink: 0;
    padding: 16px 24px;
    border-top: 1px solid #ebeef5;
  }
}

// URL输入容器样式
.url-input-container {
  display: flex;
  gap: 10px;
  width: 100%;

  .url-input {
    flex: 1;
  }

  .parse-button {
    flex-shrink: 0;
  }
}

// 项目信息展示框样式
.project-info-box {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  margin: 20px 0;

  .project-info-item {
    display: flex;
    margin-bottom: 12px;

    &:last-child {
      margin-bottom: 0;
    }

    .info-label {
      font-weight: 500;
      color: #606266;
      width: 120px;
      flex-shrink: 0;
    }

    .info-value {
      color: #303133;
      flex: 1;
      word-break: break-word;
    }
  }
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
