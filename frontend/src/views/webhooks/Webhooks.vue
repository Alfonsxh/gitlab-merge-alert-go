<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Webhook管理</h1>
      <el-button type="primary" @click="showAddModal" size="large">
        <el-icon><Plus /></el-icon>
        添加Webhook
      </el-button>
    </div>
    
    <el-card>
      <el-table
        :data="webhooks"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="60" />
        
        <el-table-column prop="name" label="名称" min-width="150">
          <template #default="{ row }">
            <div class="webhook-name">
              <el-icon><Connection /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="url" label="URL" min-width="300">
          <template #default="{ row }">
            <div class="webhook-url">
              <el-text class="url-text" truncated>{{ row.url }}</el-text>
              <el-button
                link
                type="primary"
                size="small"
                @click="copyUrl(row.url)"
              >
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        
        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.is_active ? 'success' : 'info'"
              size="small"
            >
              {{ row.is_active ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="关联项目" width="200">
          <template #default="{ row }">
            <el-space v-if="row.projects?.length" wrap>
              <el-tag
                v-for="project in row.projects"
                :key="project.id"
                size="small"
                type="info"
              >
                {{ project.name }}
              </el-tag>
            </el-space>
            <span v-else class="text-muted">无</span>
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="testWebhook(row)">
              <el-icon><VideoPlay /></el-icon>
              测试
            </el-button>
            <el-button link type="primary" size="small" @click="editWebhook(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-popconfirm
              title="确定要删除这个Webhook吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="deleteWebhook(row.id)"
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
    </el-card>
    
    <!-- 添加/编辑Webhook对话框 -->
    <el-dialog
      v-model="modalVisible"
      :title="isEditing ? '编辑Webhook' : '添加Webhook'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="currentWebhook"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="名称" prop="name">
          <el-input
            v-model="currentWebhook.name"
            placeholder="为这个Webhook起一个易识别的名称"
          >
            <template #prefix>
              <el-icon><Connection /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="Webhook URL" prop="url">
          <el-input
            v-model="currentWebhook.url"
            placeholder="企业微信机器人的Webhook URL"
          >
            <template #prefix>
              <el-icon><Link /></el-icon>
            </template>
          </el-input>
          <div class="form-item-help">
            请输入企业微信群机器人的完整Webhook地址
          </div>
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="currentWebhook.description"
            type="textarea"
            :rows="3"
            placeholder="描述这个Webhook的用途"
          />
        </el-form-item>
        
        <el-form-item label="状态">
          <el-switch
            v-model="currentWebhook.is_active"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="modalVisible = false">取消</el-button>
        <el-button type="primary" @click="saveWebhook" :loading="submitting">
          {{ isEditing ? '更新' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 测试Webhook对话框 -->
    <el-dialog
      v-model="testModalVisible"
      title="测试Webhook"
      width="500px"
    >
      <div v-if="testingWebhook" class="test-content">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="名称">
            {{ testingWebhook.name }}
          </el-descriptions-item>
          <el-descriptions-item label="URL">
            <el-text class="url-text" truncated>{{ testingWebhook.url }}</el-text>
          </el-descriptions-item>
        </el-descriptions>
        
        <el-divider />
        
        <el-button
          type="primary"
          :loading="testingSending"
          @click="sendTestMessage"
          style="width: 100%"
        >
          <el-icon><Promotion /></el-icon>
          发送测试消息
        </el-button>
        
        <el-alert
          v-if="testResult"
          :type="testResult.success ? 'success' : 'error'"
          :title="testResult.message"
          :closable="false"
          class="test-result"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  VideoPlay,
  Connection,
  Link,
  CopyDocument,
  Promotion
} from '@element-plus/icons-vue'
import { webhooksApi } from '@/api'
import type { Webhook } from '@/api'
import { formatDate } from '@/utils/format'

const webhooks = ref<Webhook[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const testModalVisible = ref(false)
const submitting = ref(false)
const testingSending = ref(false)
const isEditing = ref(false)
const formRef = ref<FormInstance>()
const testingWebhook = ref<Webhook | null>(null)
const testResult = ref<{ success: boolean; message: string } | null>(null)

const currentWebhook = reactive<Partial<Webhook>>({
  name: '',
  url: '',
  description: '',
  is_active: true
})

const rules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '请输入Webhook URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: ['blur', 'change'] }
  ]
}

const loadWebhooks = async () => {
  loading.value = true
  try {
    const res = await webhooksApi.getWebhooks()
    webhooks.value = res.data || []
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    loading.value = false
  }
}

const showAddModal = () => {
  Object.assign(currentWebhook, {
    id: undefined,
    name: '',
    url: '',
    description: '',
    is_active: true
  })
  isEditing.value = false
  modalVisible.value = true
}

const editWebhook = (webhook: Webhook) => {
  Object.assign(currentWebhook, webhook)
  isEditing.value = true
  modalVisible.value = true
}

const saveWebhook = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  
  submitting.value = true
  try {
    if (isEditing.value && currentWebhook.id) {
      await webhooksApi.updateWebhook(currentWebhook.id, currentWebhook)
    } else {
      await webhooksApi.createWebhook(currentWebhook)
    }
    
    ElMessage.success(isEditing.value ? '更新成功' : '添加成功')
    modalVisible.value = false
    await loadWebhooks()
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    submitting.value = false
  }
}

const deleteWebhook = async (webhookId: number) => {
  try {
    await webhooksApi.deleteWebhook(webhookId)
    ElMessage.success('删除成功')
    await loadWebhooks()
  } catch (error) {
    // 错误已在 API 客户端处理
  }
}

const copyUrl = (url: string) => {
  navigator.clipboard.writeText(url).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

const testWebhook = (webhook: Webhook) => {
  testingWebhook.value = webhook
  testResult.value = null
  testModalVisible.value = true
}

const sendTestMessage = async () => {
  if (!testingWebhook.value) return
  
  testingSending.value = true
  testResult.value = null
  
  try {
    // TODO: 调用后端API发送测试消息
    // 模拟发送
    await new Promise(resolve => setTimeout(resolve, 1500))
    
    testResult.value = {
      success: true,
      message: '测试消息发送成功！'
    }
  } catch (error) {
    testResult.value = {
      success: false,
      message: '测试消息发送失败，请检查Webhook配置'
    }
  } finally {
    testingSending.value = false
  }
}

onMounted(() => {
  loadWebhooks()
})
</script>

<style scoped lang="less">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  
  :deep(.el-button) {
    height: 40px;
    font-size: 15px;
  }
}

.el-card {
  :deep(.el-card__body) {
    padding: 0;
  }
}

:deep(.el-table) {
  .webhook-name {
    display: flex;
    align-items: center;
    gap: 8px;
    
    .el-icon {
      color: #409eff;
    }
  }
  
  .webhook-url {
    display: flex;
    align-items: center;
    gap: 8px;
    
    .url-text {
      flex: 1;
      font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
      font-size: 13px;
      background: #f5f7fa;
      padding: 2px 8px;
      border-radius: 4px;
    }
  }
}

.form-item-help {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

.test-content {
  .el-descriptions {
    margin-bottom: 20px;
    
    .url-text {
      font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
      font-size: 12px;
    }
  }
  
  .test-result {
    margin-top: 20px;
  }
}

.text-muted {
  color: #909399;
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
}

// 响应式设计
@media screen and (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    
    .page-title {
      margin-bottom: 0;
    }
  }
  
  :deep(.el-table) {
    font-size: 12px;
  }
  
  :deep(.el-dialog) {
    width: 90% !important;
  }
}
</style>