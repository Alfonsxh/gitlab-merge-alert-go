<template>
  <div class="page-container">
    <h1 class="page-title">仪表板</h1>
    
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-body">
            <div class="stat-icon" style="background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);">
              <el-icon :size="24"><User /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_users }}</div>
              <div class="stat-label">用户总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-body">
            <div class="stat-icon" style="background: linear-gradient(135deg, #67c23a 0%, #409eff 100%);">
              <el-icon :size="24"><FolderOpened /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_projects }}</div>
              <div class="stat-label">项目总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-body">
            <div class="stat-icon" style="background: linear-gradient(135deg, #e6a23c 0%, #f56c6c 100%);">
              <el-icon :size="24"><Link /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_webhooks }}</div>
              <div class="stat-label">Webhook总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-body">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f56c6c 0%, #e6a23c 100%);">
              <el-icon :size="24"><Bell /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_notifications }}</div>
              <div class="stat-label">通知总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 整合的内容卡片 -->
    <el-card class="main-content-card">
      <div class="scrollable-content">
        <!-- 统计图表 -->
        <div class="chart-section">
          <h2 class="section-title">数据统计</h2>
          <el-row :gutter="20">
            <el-col :xs="24" :md="12">
              <div class="chart-wrapper">
                <div class="chart-header">
                  <span class="chart-title">项目每日 Merge 数趋势</span>
                  <el-select v-model="projectChartDays" size="small" @change="loadProjectStats">
                    <el-option :value="7" label="最近 7 天" />
                    <el-option :value="14" label="最近 14 天" />
                    <el-option :value="30" label="最近 30 天" />
                  </el-select>
                </div>
                <el-skeleton :loading="projectStatsLoading" animated :rows="10">
                  <LineChart
                    :data="projectChartData"
                    :height="'350px'"
                    :show-data-zoom="false"
                  />
                </el-skeleton>
              </div>
            </el-col>
            
            <el-col :xs="24" :md="12">
              <div class="chart-wrapper">
                <div class="chart-header">
                  <span class="chart-title">Webhook 每日 Merge 数趋势</span>
                  <el-select v-model="webhookChartDays" size="small" @change="loadWebhookStats">
                    <el-option :value="7" label="最近 7 天" />
                    <el-option :value="14" label="最近 14 天" />
                    <el-option :value="30" label="最近 30 天" />
                  </el-select>
                </div>
                <el-skeleton :loading="webhookStatsLoading" animated :rows="10">
                  <LineChart
                    :data="webhookChartData"
                    :height="'350px'"
                    :show-data-zoom="false"
                  />
                </el-skeleton>
              </div>
            </el-col>
          </el-row>
        </div>

        <!-- 最近通知 -->
        <div class="notifications-section">
          <div class="section-header">
            <h2 class="section-title">最近通知记录</h2>
            <el-tag type="info" size="small">
              <el-icon><Clock /></el-icon>
              最近 24 小时
            </el-tag>
          </div>
          
          <el-empty v-if="notifications.length === 0" description="暂无通知记录">
            <template #image>
              <el-icon :size="64"><Bell /></el-icon>
            </template>
          </el-empty>
          
          <div v-else class="notifications-grouped">
        <el-collapse accordion>
          <el-collapse-item
            v-for="(projectNotifications, projectName) in groupedNotifications"
            :key="projectName"
          >
            <template #title>
              <div class="project-header">
                <div class="project-info">
                  <el-icon class="project-icon"><Folder /></el-icon>
                  <span class="project-name">{{ projectName }}</span>
                  <el-badge :value="projectNotifications.length" type="primary" />
                </div>
              </div>
            </template>
            
            <el-table
              :data="projectNotifications"
              size="small"
              stripe
            >
              <el-table-column label="标题" prop="title" width="250">
                <template #default="{ row }">
                  <el-link
                    v-if="getMergeRequestUrl(row)"
                    :href="getMergeRequestUrl(row)"
                    target="_blank"
                    type="primary"
                  >
                    {{ row.title }}
                  </el-link>
                  <span v-else>{{ row.title }}</span>
                </template>
              </el-table-column>
              
              <el-table-column label="分支" width="200">
                <template #default="{ row }">
                  <div class="branch-flow">
                    <el-tag size="small" type="danger">
                      <el-icon><Share /></el-icon>
                      {{ row.source_branch }}
                    </el-tag>
                    <el-icon class="branch-arrow"><Right /></el-icon>
                    <el-tag size="small" type="primary">
                      <el-icon><Share /></el-icon>
                      {{ row.target_branch }}
                    </el-tag>
                  </div>
                </template>
              </el-table-column>
              
              <el-table-column label="提交者" prop="author_email">
                <template #default="{ row }">
                  {{ extractNameFromEmail(row.author_email) }}
                </template>
              </el-table-column>
              
              <el-table-column label="Merge者">
                <template #default="{ row }">
                  <el-space v-if="row.assignee_emails?.length" wrap>
                    <el-tag
                      v-for="email in row.assignee_emails"
                      :key="email"
                      size="small"
                    >
                      {{ extractNameFromEmail(email) }}
                    </el-tag>
                  </el-space>
                  <span v-else class="text-muted">-</span>
                </template>
              </el-table-column>
              
              <el-table-column label="状态" width="120">
                <template #default="{ row }">
                  <div class="status-wrapper">
                    <el-tag 
                      :type="row.notification_sent ? 'success' : 'danger'"
                      size="small"
                    >
                      <el-icon>
                        <CircleCheck v-if="row.notification_sent" />
                        <CircleClose v-else />
                      </el-icon>
                      {{ row.notification_sent ? '已发送' : '发送失败' }}
                    </el-tag>
                    <el-tooltip
                      v-if="!row.notification_sent && row.error_message"
                      :content="row.error_message"
                      placement="top"
                    >
                      <el-icon class="error-info-icon"><InfoFilled /></el-icon>
                    </el-tooltip>
                  </div>
                </template>
              </el-table-column>
              
              <el-table-column label="时间" prop="created_at" width="160">
                <template #default="{ row }">
                  {{ formatDate(row.created_at) }}
                </template>
              </el-table-column>
            </el-table>
          </el-collapse-item>
            </el-collapse>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  User,
  FolderOpened,
  Link,
  Bell,
  Clock,
  Folder,
  Share,
  Right,
  CircleCheck,
  CircleClose,
  InfoFilled
} from '@element-plus/icons-vue'
import { statsApi, notificationsApi, projectsApi } from '@/api'
import type { Stats, Notification, Project, ProjectDailyStats, WebhookDailyStats } from '@/api'
import { formatDate, extractNameFromEmail } from '@/utils/format'
import LineChart from '@/components/charts/LineChart.vue'

const stats = ref<Stats>({
  total_users: 0,
  total_projects: 0,
  total_webhooks: 0,
  total_notifications: 0
})

const notifications = ref<Notification[]>([])
const projects = ref<Project[]>([])
const projectsMap = ref<Record<number, Project>>({})

// 图表相关数据
const projectChartDays = ref(7)
const webhookChartDays = ref(7)
const projectStatsLoading = ref(false)
const webhookStatsLoading = ref(false)
const projectDailyStats = ref<ProjectDailyStats[]>([])
const webhookDailyStats = ref<WebhookDailyStats[]>([])

// 处理图表数据
const projectChartData = computed(() => {
  return projectDailyStats.value
    .map(item => ({
      name: item.project_name,
      data: item.data,
      total: item.data.reduce((sum, d) => sum + d.count, 0)
    }))
    .sort((a, b) => b.total - a.total) // 按总merge数降序排序
    .slice(0, 10) // 只取前10个
    .map(item => ({
      name: item.name,
      data: item.data
    }))
})

const webhookChartData = computed(() => {
  return webhookDailyStats.value
    .map(item => ({
      name: item.webhook_name,
      data: item.data,
      total: item.data.reduce((sum, d) => sum + d.count, 0)
    }))
    .sort((a, b) => b.total - a.total) // 按总merge数降序排序
    .slice(0, 10) // 只取前10个
    .map(item => ({
      name: item.name,
      data: item.data
    }))
})

const groupedNotifications = computed(() => {
  const groups: Record<string, Notification[]> = {}
  
  notifications.value.forEach(notification => {
    const projectName = notification.project_name || '未知项目'
    if (!groups[projectName]) {
      groups[projectName] = []
    }
    groups[projectName].push(notification)
  })
  
  // 按项目名称排序
  const sortedGroups: Record<string, Notification[]> = {}
  Object.keys(groups).sort().forEach(projectName => {
    sortedGroups[projectName] = groups[projectName].sort((a, b) => 
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  })
  
  return sortedGroups
})

const getMergeRequestUrl = (notification: Notification) => {
  const project = projectsMap.value[notification.project_id]
  if (!project || !project.url || !notification.merge_request_id) {
    return undefined
  }
  return `${project.url}/-/merge_requests/${notification.merge_request_id}`
}

const loadStats = async () => {
  try {
    const res = await statsApi.getStats()
    stats.value = res.data || res // 兼容处理
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  }
}

const loadProjects = async () => {
  try {
    const res = await projectsApi.getProjects()
    // 确保 projects.value 是数组
    const projectData = res.data || res || []
    projects.value = Array.isArray(projectData) ? projectData : []
    
    // 建立项目ID到项目信息的映射
    projectsMap.value = {}
    projects.value.forEach(project => {
      projectsMap.value[project.id] = project
    })
  } catch (error) {
    console.error('加载项目数据失败:', error)
    projects.value = []
    ElMessage.error('加载项目数据失败')
  }
}

const loadNotifications = async () => {
  try {
    const res = await notificationsApi.getNotifications({ page_size: 10 })
    // 确保 notifications.value 是数组
    const notificationData = res.data || res || []
    notifications.value = Array.isArray(notificationData) ? notificationData : []
  } catch (error) {
    console.error('加载通知数据失败:', error)
    notifications.value = []
    ElMessage.error('加载通知数据失败')
  }
}

const loadProjectStats = async () => {
  projectStatsLoading.value = true
  try {
    const res = await statsApi.getProjectDailyStats(projectChartDays.value)
    projectDailyStats.value = res.data || []
  } catch (error) {
    console.error('加载项目统计数据失败:', error)
    ElMessage.error('加载项目统计数据失败')
  } finally {
    projectStatsLoading.value = false
  }
}

const loadWebhookStats = async () => {
  webhookStatsLoading.value = true
  try {
    const res = await statsApi.getWebhookDailyStats(webhookChartDays.value)
    webhookDailyStats.value = res.data || []
  } catch (error) {
    console.error('加载Webhook统计数据失败:', error)
    ElMessage.error('加载Webhook统计数据失败')
  } finally {
    webhookStatsLoading.value = false
  }
}

onMounted(() => {
  loadStats()
  loadProjects()
  loadNotifications()
  loadProjectStats()
  loadWebhookStats()
})
</script>

<style scoped lang="less">
.page-title {
  margin: 0 0 20px 0;
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

.stat-cards {
  margin-bottom: 0;
  
  .stat-card {
    border: none;
    margin-bottom: 10px;
    height: 80px;
    
    :deep(.el-card__body) {
      padding: 0 16px;
      height: 100%;
      display: flex;
      align-items: center;
    }
    
    .stat-card-body {
      display: flex;
      align-items: center;
      gap: 16px;
      width: 100%;
      
      .stat-icon {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;
        transition: transform 0.3s ease;
      }
      
      .stat-content {
        flex: 1;
        
        .stat-value {
          font-size: 26px;
          font-weight: 700;
          line-height: 1.2;
          color: #303133;
          margin-bottom: 2px;
        }
        
        .stat-label {
          font-size: 13px;
          color: #909399;
        }
      }
    }
    
    &:hover {
      .stat-icon {
        transform: scale(1.1);
      }
    }
  }
}

// 主内容卡片
.main-content-card {
  min-height: calc(100vh - 164px);
  height: auto;
  display: flex;
  flex-direction: column;
  margin-top: 10px;
  margin-bottom: 20px;
  position: relative;
  
  :deep(.el-card__body) {
    padding: 0;
    height: auto;
    overflow: visible;
  }
  
  .scrollable-content {
    height: auto;
    overflow: visible;
    padding: 20px;
    
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

// 章节标题
.section-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 20px 0;
  padding-left: 12px;
  position: relative;
  
  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 4px;
    height: 20px;
    background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
    border-radius: 2px;
  }
}

// 图表部分
.chart-section {
  margin-bottom: 40px;
  
  .chart-wrapper {
    background: #f5f7fa;
    border-radius: 8px;
    padding: 20px;
    height: 100%;
    
    .chart-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 20px;
      
      .chart-title {
        font-size: 16px;
        font-weight: 500;
        color: #303133;
      }
      
      .el-select {
        width: 120px;
      }
    }
  }
  
  .el-col {
    margin-bottom: 20px;
    
    @media screen and (min-width: 992px) {
      margin-bottom: 0;
    }
  }
}

// 通知部分
.notifications-section {
  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
  }
  
  .notifications-grouped {
    
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
    
    .project-header {
      width: 100%;
      
      .project-info {
        display: flex;
        align-items: center;
        gap: 12px;
        
        .project-icon {
          font-size: 20px;
          color: #409eff;
        }
        
        .project-name {
          font-weight: 600;
          font-size: 16px;
          color: #303133;
          margin-right: 12px;
        }
      }
    }
  }
}

.branch-flow {
  display: flex;
  align-items: center;
  gap: 8px;
  
  .branch-arrow {
    color: #909399;
    font-size: 16px;
  }
  
  :deep(.el-tag) {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-weight: 500;
  }
}

.status-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  
  .error-info-icon {
    color: #f56c6c;
    cursor: help;
  }
}

.text-muted {
  color: #909399;
}

// 响应式设计
@media screen and (max-width: 1200px) {
  .main-content-card {
    min-height: calc(100vh - 224px);
    height: auto;
    margin-bottom: 20px;
  }
}

@media screen and (max-width: 768px) {
  .stat-cards {
    .el-col {
      margin-bottom: 10px;
    }
    
    .stat-card {
      margin-bottom: 10px;
      
      :deep(.el-card__body) {
        padding: 16px;
      }
      
      .stat-card-body {
        gap: 16px;
        
        .stat-icon {
          width: 44px;
          height: 44px;
        }
        
        .stat-value {
          font-size: 24px;
        }
      }
    }
  }
  
  .main-content-card {
    min-height: calc(100vh - 264px);
    height: auto;
    margin-top: 10px;
    margin-bottom: 20px;
  }
  
  .scrollable-content {
    padding: 16px;
  }
  
  .section-title {
    font-size: 18px;
    margin-bottom: 16px;
  }
  
  .chart-section {
    margin-bottom: 30px;
    
    .chart-wrapper {
      padding: 16px;
    }
  }
  
  .notifications-grouped {
    padding: 12px;
    
    :deep(.el-table) {
      font-size: 12px;
    }
  }
}
</style>