<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">仪表板</h1>
    </div>
    
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-body">
            <div class="stat-icon" style="background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);">
              <el-icon :size="28"><User /></el-icon>
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
              <el-icon :size="28"><FolderOpened /></el-icon>
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
              <el-icon :size="28"><Link /></el-icon>
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
              <el-icon :size="28"><Bell /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_notifications }}</div>
              <div class="stat-label">通知总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近通知 -->
    <el-card class="notifications-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">最近通知记录</span>
          <el-tag type="info" size="small">
            <el-icon><Clock /></el-icon>
            最近 24 小时
          </el-tag>
        </div>
      </template>
      
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
import type { Stats, Notification, Project } from '@/api'
import { formatDate, extractNameFromEmail } from '@/utils/format'

const stats = ref<Stats>({
  total_users: 0,
  total_projects: 0,
  total_webhooks: 0,
  total_notifications: 0
})

const notifications = ref<Notification[]>([])
const projects = ref<Project[]>([])
const projectsMap = ref<Record<number, Project>>({})

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
    stats.value = res.data
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  }
}

const loadProjects = async () => {
  try {
    const res = await projectsApi.getProjects()
    projects.value = res.data || []
    
    // 建立项目ID到项目信息的映射
    projectsMap.value = {}
    projects.value.forEach(project => {
      projectsMap.value[project.id] = project
    })
  } catch (error) {
    ElMessage.error('加载项目数据失败')
  }
}

const loadNotifications = async () => {
  try {
    const res = await notificationsApi.getNotifications({ page_size: 10 })
    notifications.value = res.data || []
  } catch (error) {
    ElMessage.error('加载通知数据失败')
  }
}

onMounted(() => {
  loadStats()
  loadProjects()
  loadNotifications()
})
</script>

<style scoped lang="less">
.stat-cards {
  margin-bottom: 24px;
  
  .stat-card {
    border: none;
    margin-bottom: 20px;
    
    :deep(.el-card__body) {
      padding: 20px;
    }
    
    .stat-card-body {
      display: flex;
      align-items: center;
      gap: 20px;
      
      .stat-icon {
        width: 64px;
        height: 64px;
        border-radius: 16px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;
        transition: transform 0.3s ease;
      }
      
      .stat-content {
        flex: 1;
        
        .stat-value {
          font-size: 32px;
          font-weight: 700;
          line-height: 1.2;
          color: #303133;
          margin-bottom: 4px;
        }
        
        .stat-label {
          font-size: 14px;
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

.notifications-card {
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    
    .card-title {
      font-size: 18px;
      font-weight: 600;
    }
  }
  
  :deep(.el-card__body) {
    padding: 0;
  }
  
  .notifications-grouped {
    padding: 20px;
    
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
@media screen and (max-width: 768px) {
  .stat-cards {
    .el-col {
      margin-bottom: 12px;
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