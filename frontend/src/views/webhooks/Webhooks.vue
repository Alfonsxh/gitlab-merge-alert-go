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
      <div class="table-scroll-wrapper" ref="tableWrapperRef">
        <el-table
          :data="webhooks"
          v-loading="loading"
          stripe
          style="min-width: 1100px; width: 100%;"
        >
          <el-table-column prop="id" label="ID" width="80" />

          <el-table-column prop="name" label="名称" min-width="200">
            <template #default="{ row }">
              <div class="webhook-name">
                <el-icon><Connection /></el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ renderWebhookType(row.type, row.url) }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column prop="url" label="URL" min-width="360">
            <template #default="{ row }">
              <div class="webhook-url">
                <el-text class="url-text" truncated>{{ row.url }}</el-text>
                <el-button link type="primary" size="small" @click="copyUrl(row.url)">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="description" label="描述" min-width="220" show-overflow-tooltip />

          <el-table-column prop="is_active" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
                {{ row.is_active ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="关联项目" min-width="260">
            <template #default="{ row }">
              <div v-if="row.projects?.length" class="project-tags">
                <el-popover
                  v-if="row.projects.length > 3"
                  placement="top"
                  trigger="hover"
                  width="400"
                >
                  <template #reference>
                    <div class="project-tags-compact">
                      <el-tag
                        v-for="project in row.projects.slice(0, 3)"
                        :key="project.id"
                        size="small"
                        type="info"
                        class="project-tag"
                        @click="goToProject(project)"
                      >
                        {{ project.name }}
                      </el-tag>
                      <el-tag size="small" type="warning">
                        +{{ row.projects.length - 3 }}
                      </el-tag>
                    </div>
                  </template>
                  <div class="project-list-popover">
                    <div class="popover-title">所有关联项目 ({{ row.projects.length }})</div>
                    <div class="project-list">
                      <el-tag
                        v-for="project in row.projects"
                        :key="project.id"
                        size="small"
                        type="info"
                        class="project-tag"
                        @click="goToProject(project)"
                      >
                        {{ project.name }}
                      </el-tag>
                    </div>
                  </div>
                </el-popover>
                <div v-else class="project-tags-compact">
                  <el-tag
                    v-for="project in row.projects"
                    :key="project.id"
                    size="small"
                    type="info"
                    class="project-tag"
                    @click="goToProject(project)"
                  >
                    {{ project.name }}
                  </el-tag>
                </div>
              </div>
              <span v-else class="text-muted">无</span>
            </template>
          </el-table-column>

          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </el-table-column>

          <el-table-column label="操作" width="320" fixed="right">
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
      </div>
    </el-card>

    <el-dialog
      v-model="modalVisible"
      :title="isEditing ? '编辑Webhook' : '添加Webhook'"
      width="680px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="currentWebhook" :rules="rules" label-width="110px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="currentWebhook.name" placeholder="为这个Webhook起一个易识别的名称">
            <template #prefix>
              <el-icon><Connection /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item prop="url">
          <template #label>
            <span class="label-inline">Webhook URL</span>
          </template>
          <el-input
            v-model="currentWebhook.url"
            placeholder="请粘贴完整的 Webhook 地址"
            class="full-width-input"
          >
            <template #prefix>
              <el-icon><Link /></el-icon>
            </template>
          </el-input>
          <div class="form-item-help">填写有效地址后会自动识别渠道</div>
        </el-form-item>

        <el-form-item label="Webhook 类型" prop="type">
          <el-select v-model="currentWebhook.type" placeholder="选择通知渠道">
            <el-option v-for="option in webhookTypeOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <div class="form-item-help" v-if="currentWebhook.url">
            当前识别结果：<strong>{{ renderWebhookType(effectiveType) }}</strong>
            <span class="type-hint">（{{ selectedType === 'auto' ? '自动识别' : '已手动指定' }}）</span>
          </div>
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input v-model="currentWebhook.description" type="textarea" :rows="3" placeholder="描述这个Webhook的用途" />
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="currentWebhook.is_active" active-text="启用" inactive-text="禁用" />
        </el-form-item>

        <el-divider v-if="effectiveType === 'dingtalk'">钉钉配置</el-divider>

        <template v-if="effectiveType === 'dingtalk'">
          <el-alert
            title="钉钉自定义机器人需开启安全策略（关键词、加签或IP白名单）"
            type="warning"
            :closable="false"
            show-icon
            class="form-alert"
          />

          <el-form-item label="加签 Secret" prop="secret">
            <el-input v-model="currentWebhook.secret" placeholder="请输入钉钉机器人加签Secret" />
            <div class="form-item-help">加签用于生成签名，确保请求来源可靠。</div>
          </el-form-item>

          <el-form-item label="签名算法">
            <el-input v-model="currentWebhook.signature_method" disabled />
          </el-form-item>

          <el-form-item label="安全关键词">
            <el-select
              v-model="currentWebhook.security_keywords"
              multiple
              filterable
              allow-create
              default-first-option
              placeholder="输入关键词后回车"
            >
              <el-option
                v-for="keyword in currentWebhook.security_keywords || []"
                :key="keyword"
                :label="keyword"
                :value="keyword"
              />
            </el-select>
            <div class="form-item-help">如开启关键词策略，请保证消息体包含任意一个关键词。</div>
          </el-form-item>
        </template>

        <el-divider v-if="effectiveType === 'custom'">自定义Webhook提示</el-divider>

        <template v-if="effectiveType === 'custom' && currentWebhook.url">
          <el-alert
            title="自定义Webhook仅用于记录，GitLab Merge Alert 不会转发消息。请在 GitLab 中直接配置该地址。"
            type="info"
            :closable="false"
            show-icon
            class="form-alert"
            style="margin: 0 0 12px 110px; width: calc(100% - 110px);"
          />

          <el-form-item label="自定义 Header">
            <div class="custom-headers">
              <div class="header-row" v-for="(item, index) in customHeaders" :key="index">
                <el-input v-model="item.key" placeholder="Header 名称" class="header-input" />
                <el-input v-model="item.value" placeholder="Header 值" class="header-input" />
                <el-button
                  v-if="customHeaders.length > 1"
                  link
                  type="danger"
                  class="header-remove"
                  @click="removeCustomHeader(index)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button type="primary" size="small" class="header-add" @click="addCustomHeader">
                +
              </el-button>
            </div>
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="modalVisible = false">取消</el-button>
        <el-button type="primary" @click="saveWebhook" :loading="submitting">
          {{ isEditing ? '更新' : '添加' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="testModalVisible"
      title="测试Webhook"
      width="520px"
      class="webhook-test-dialog"
    >
      <div v-if="testingWebhook" class="webhook-test-content">
        <div class="webhook-info-card">
          <div class="info-row">
            <span class="info-label">名称</span>
            <span class="info-value">{{ testingWebhook.name }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">类型</span>
            <span class="info-value">{{ renderWebhookType(testingWebhook.type, testingWebhook.url) }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">URL</span>
            <div class="info-value url">
              <el-text class="url-text" truncated>{{ testingWebhook.url }}</el-text>
              <el-button
                v-if="testingWebhook?.url"
                link
                size="small"
                type="primary"
                @click="copyTestingWebhookUrl"
              >
                <el-icon><CopyDocument /></el-icon>
                复制
              </el-button>
            </div>
          </div>
        </div>

        <div class="test-actions">
          <el-button type="primary" :loading="testingSending" @click="sendTestMessage">
            <el-icon><Promotion /></el-icon>
            <span>{{ testingSending ? '发送中...' : '发送测试消息' }}</span>
          </el-button>

          <el-alert
            v-if="testResult"
            :type="testResult.success ? 'success' : 'error'"
            :title="testResult.message"
            :closable="false"
            class="test-result"
          />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useRouter } from 'vue-router'
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
import type { Webhook, WebhookType, UpsertWebhookPayload } from '@/api'
import { formatDate } from '@/utils/format'

const router = useRouter()
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
const tableWrapperRef = ref<HTMLDivElement | null>(null)

const webhookTypeLabels: Record<string, string> = {
  auto: '自动识别',
  wechat: '企业微信',
  dingtalk: '钉钉',
  custom: '自定义'
}

const webhookTypeOptions = [
  { label: '自动识别', value: 'auto' },
  { label: '企业微信', value: 'wechat' },
  { label: '钉钉', value: 'dingtalk' },
  { label: '自定义', value: 'custom' }
]

const detectWebhookType = (url?: string): WebhookType => {
  if (!url) return 'custom'
  try {
    const parsed = new URL(url)
    const host = parsed.host.toLowerCase()
    if (host.includes('dingtalk.com') || host.includes('dingtalk')) {
      return 'dingtalk'
    }
    if (host.includes('qyapi.weixin.qq.com') || host.includes('work.weixin.qq.com')) {
      return 'wechat'
    }
    return 'custom'
  } catch (error) {
    return 'custom'
  }
}

const currentWebhook = reactive<UpsertWebhookPayload & { id?: number }>(
  {
    id: undefined,
    name: '',
    url: '',
    description: '',
    type: 'auto',
    signature_method: 'hmac_sha256',
    secret: '',
    security_keywords: [],
    custom_headers: {},
    is_active: true
  }
)

const customHeaders = ref<Array<{ key: string; value: string }>>([{ key: '', value: '' }])

const selectedType = computed<WebhookType>(() => (currentWebhook.type as WebhookType) || 'auto')
const effectiveType = computed<WebhookType>(() =>
  selectedType.value === 'auto' ? detectWebhookType(currentWebhook.url) : selectedType.value
)

const rules = reactive<FormRules<UpsertWebhookPayload & { id?: number }>>({
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [
    { required: true, message: '请输入Webhook URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: ['blur', 'change'] }
  ],
  secret: [
    {
      validator: (_rule, value, callback) => {
        if (effectiveType.value === 'dingtalk' && !value) {
          callback(new Error('钉钉安全加签需要填写Secret'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
})

watch(effectiveType, newType => {
  if (newType === 'dingtalk' && !currentWebhook.signature_method) {
    currentWebhook.signature_method = 'hmac_sha256'
  }
  if (newType !== 'dingtalk') {
    currentWebhook.security_keywords = currentWebhook.security_keywords?.length ? currentWebhook.security_keywords : []
  }
})

const renderWebhookType = (type?: string, url?: string) => {
  const resolved = type && type !== 'auto' ? type : detectWebhookType(url)
  return webhookTypeLabels[resolved] || webhookTypeLabels.custom
}

const resetCustomHeaders = (headers?: Record<string, string>) => {
  const entries = headers ? Object.entries(headers) : []
  customHeaders.value = entries.length
    ? entries.map(([key, value]) => ({ key, value }))
    : [{ key: '', value: '' }]
}

const buildCustomHeadersPayload = (): Record<string, string> => {
  const payload: Record<string, string> = {}
  customHeaders.value.forEach(({ key, value }) => {
    const trimmedKey = key.trim()
    if (trimmedKey) {
      payload[trimmedKey] = value
    }
  })
  return payload
}

const loadWebhooks = async () => {
  loading.value = true
  try {
    const res = await webhooksApi.getWebhooks()
    webhooks.value = (res.data || []).map(item => ({
      ...item,
      type: (item.type || detectWebhookType(item.url)) as WebhookType
    }))
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
    type: 'auto' as WebhookType,
    signature_method: 'hmac_sha256',
    secret: '',
    security_keywords: [],
    custom_headers: {},
    is_active: true
  })
  resetCustomHeaders()
  isEditing.value = false
  modalVisible.value = true
}

const editWebhook = (webhook: Webhook) => {
  Object.assign(currentWebhook, {
    id: webhook.id,
    name: webhook.name,
    url: webhook.url,
    description: webhook.description,
    type: (webhook.type as WebhookType) || 'wechat',
    signature_method: webhook.signature_method || 'hmac_sha256',
    secret: webhook.secret || '',
    security_keywords: webhook.security_keywords ? [...webhook.security_keywords] : [],
    custom_headers: webhook.custom_headers ? { ...webhook.custom_headers } : {},
    is_active: webhook.is_active
  })
  resetCustomHeaders(webhook.custom_headers || {})
  isEditing.value = true
  modalVisible.value = true
}

const saveWebhook = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload: UpsertWebhookPayload = {
      name: currentWebhook.name,
      url: currentWebhook.url,
      description: currentWebhook.description,
      type: selectedType.value,
      signature_method: currentWebhook.signature_method,
      is_active: currentWebhook.is_active,
      secret: effectiveType.value === 'dingtalk' ? (currentWebhook.secret || '') : '',
      security_keywords: (currentWebhook.security_keywords || []).map(keyword => keyword.trim()).filter(Boolean),
      custom_headers: effectiveType.value === 'custom' ? buildCustomHeadersPayload() : {}
    }

    if (isEditing.value && currentWebhook.id) {
      await webhooksApi.updateWebhook(currentWebhook.id, payload)
    } else {
      await webhooksApi.createWebhook(payload)
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
  navigator.clipboard
    .writeText(url)
    .then(() => {
      ElMessage.success('已复制到剪贴板')
    })
    .catch(() => {
      ElMessage.error('复制失败')
    })
}

const copyTestingWebhookUrl = () => {
  if (!testingWebhook.value?.url) return
  copyUrl(testingWebhook.value.url)
}

const addCustomHeader = () => {
  customHeaders.value.push({ key: '', value: '' })
}

const removeCustomHeader = (index: number) => {
  customHeaders.value.splice(index, 1)
  if (customHeaders.value.length === 0) {
    customHeaders.value.push({ key: '', value: '' })
  }
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
    const response = await webhooksApi.sendTestMessage(testingWebhook.value.id)

    if (response.channel && testingWebhook.value) {
      testingWebhook.value = { ...testingWebhook.value, type: response.channel as WebhookType }
    }

    testResult.value = {
      success: true,
      message: response.message || '测试消息发送成功！'
    }

    ElMessage.success(`测试消息已发送到 ${response.webhook_name}`)
  } catch (error: any) {
    testResult.value = {
      success: false,
      message: error.response?.data?.details || error.response?.data?.error || '测试消息发送失败，请检查Webhook配置'
    }

    ElMessage.error(testResult.value.message)
  } finally {
    testingSending.value = false
  }
}

const goToProject = (_project: any) => {
  router.push('/projects')
}

onMounted(() => {
  loadWebhooks()

  const wrapper = tableWrapperRef.value
  if (wrapper) {
    wrapper.addEventListener('wheel', handleHorizontalScroll, { passive: false })
  }
})

onBeforeUnmount(() => {
  const wrapper = tableWrapperRef.value
  if (wrapper) {
    wrapper.removeEventListener('wheel', handleHorizontalScroll)
  }
})

const handleHorizontalScroll = (event: WheelEvent) => {
  const wrapper = tableWrapperRef.value
  if (!wrapper) return

  if (Math.abs(event.deltaY) > Math.abs(event.deltaX)) {
    event.preventDefault()
    wrapper.scrollLeft += event.deltaY
  }
}
</script>

<style scoped lang="less">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

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

  .project-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .project-tags-compact {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .project-tag {
    cursor: pointer;
  }
}

.table-scroll-wrapper {
  overflow-x: auto;
}

.text-muted {
  color: #909399;
}

.form-item-help {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.full-width-input {
  width: 100%;
}

.label-inline {
  white-space: nowrap;
}

.type-hint {
  margin-left: 4px;
  color: #606266;
}

.form-alert {
  margin-bottom: 12px;
}

.custom-headers {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;

  .header-row {
    display: flex;
    gap: 12px;
    align-items: center;
    width: 100%;
  }

  .header-input {
    flex: 1;
  }

  .header-remove {
    flex-shrink: 0;
  }

  .header-add {
    align-self: flex-start;
    margin-top: 8px;
    min-width: 32px;
    font-size: 18px;
    font-weight: 500;
  }
}

.webhook-test-dialog {
  .webhook-info-card {
    background: #f5f7fa;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 16px;

    .info-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 10px;

      &:last-child {
        margin-bottom: 0;
      }

      .info-label {
        font-weight: 600;
        color: #606266;
      }

      .info-value {
        color: #303133;
        max-width: 320px;
        text-align: right;

        &.url {
          display: flex;
          align-items: center;
          gap: 8px;

          .url-text {
            flex: 1;
            text-align: right;
          }
        }
      }
    }
  }

  .test-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;

    .test-result {
      margin-top: 8px;
    }
  }
}
</style>
