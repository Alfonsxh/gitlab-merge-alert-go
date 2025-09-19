<template>
  <el-dialog
    v-model="visible"
    class="batch-import-dialog"
    title="批量导入项目"
    :width="dialogWidth"
    :close-on-click-modal="false"
    :body-style="dialogBodyStyle"
    destroy-on-close
    @closed="handleClose"
  >
    <div v-if="projects.length > 0" class="dialog-content">
      <div class="project-section">
        <el-alert type="info" :closable="false" class="mb-4">
          发现 {{ projects.length }} 个项目，已选择 {{ selectedProjects.length }} 个
        </el-alert>

        <!-- 全选/取消全选 -->
        <div class="mb-3 selection-bar">
          <el-checkbox
            :model-value="selectedProjects.length === projects.length"
            :indeterminate="selectedProjects.length > 0 && selectedProjects.length < projects.length"
            @change="toggleAllProjects"
          >
            全选/取消全选
          </el-checkbox>
        </div>

        <!-- 项目列表 -->
        <div class="project-list custom-scroll">
          <div
            v-for="project in projects"
            :key="project.id"
            class="project-item"
          >
            <el-checkbox
              :model-value="selectedProjects.includes(project.id)"
              @change="toggleProject(project.id)"
            >
              <div class="project-info">
                <div class="project-name">{{ project.name }}</div>
                <div class="project-path">{{ project.path_with_namespace }}</div>
                <div v-if="project.description" class="project-desc">{{ project.description }}</div>
              </div>
            </el-checkbox>
          </div>
        </div>
      </div>

      <!-- Webhook 配置 -->
      <div class="webhook-section">
        <el-divider class="section-divider" />
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
            class="mt-3 webhook-select"
          >
            <el-option
              v-for="webhook in availableWebhooks"
              :key="webhook.id"
              :label="webhook.name"
              :value="webhook.id"
            >
              <span>{{ webhook.name }}</span>
              <span style="color: #999; margin-left: 10px">{{ webhook.url.substring(0, 50) }}...</span>
            </el-option>
          </el-select>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="selectedProjects.length === 0"
        @click="handleSubmit"
      >
        批量导入 ({{ selectedProjects.length }})
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { projectsApi, webhooksApi } from '@/api'

interface Props {
  modelValue: boolean
  projects: any[]
}

const props = defineProps<Props>()
const emit = defineEmits(['update:modelValue', 'success'])

const visible = ref(false)
const selectedProjects = ref<number[]>([])
const availableWebhooks = ref<any[]>([])
const submitting = ref(false)
const dialogBodyStyle = {
  padding: '0 24px 24px',
  overflow: 'hidden'
}

const dialogWidth = 'clamp(340px, 80vw, 900px)'

const webhookConfig = ref({
  useUnified: true,
  webhookId: null as number | null
})

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    // 初始化选中的项目
    selectedProjects.value = props.projects
      .filter(p => p.selected !== false)
      .map(p => p.id)
    // 加载可用的 webhooks
    loadWebhooks()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const loadWebhooks = async () => {
  try {
    const res = await webhooksApi.getWebhooks()
    availableWebhooks.value = res.data || []
  } catch (error) {
    // 错误已在 API 客户端处理
  }
}

const toggleProject = (projectId: number) => {
  const index = selectedProjects.value.indexOf(projectId)
  if (index > -1) {
    selectedProjects.value.splice(index, 1)
  } else {
    selectedProjects.value.push(projectId)
  }
}

const toggleAllProjects = () => {
  if (selectedProjects.value.length === props.projects.length) {
    selectedProjects.value = []
  } else {
    selectedProjects.value = props.projects.map(p => p.id)
  }
}

const handleSubmit = async () => {
  if (selectedProjects.value.length === 0) {
    ElMessage.warning('请至少选择一个项目')
    return
  }

  if (webhookConfig.value.useUnified && !webhookConfig.value.webhookId) {
    ElMessage.warning('请选择要关联的 Webhook')
    return
  }

  submitting.value = true
  try {
    // 准备批量创建的数据
    const selectedProjectsData = props.projects
      .filter(p => selectedProjects.value.includes(p.id))
      .map(p => ({
        gitlab_project_id: p.id,
        name: p.name,
        url: p.web_url,
        description: p.description || ''
      }))

    const batchData = {
      projects: selectedProjectsData,
      webhook_config: {
        use_unified: webhookConfig.value.useUnified,
        unified_webhook_id: webhookConfig.value.webhookId
      }
    }

    const res = await projectsApi.batchCreateProjects(batchData)

    if (res.data) {
      const { success_count, failure_count } = res.data
      ElMessage.success(`批量导入完成：成功 ${success_count} 个，失败 ${failure_count} 个`)
      visible.value = false
      emit('success')
    }
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    submitting.value = false
  }
}

const handleClose = () => {
  selectedProjects.value = []
  webhookConfig.value = {
    useUnified: true,
    webhookId: null
  }
}
</script>

<style scoped lang="less">
.batch-import-dialog {
  :deep(.el-dialog) {
    width: 100%;
    max-width: 900px;
    max-height: calc(100vh - 48px);
    display: flex;
    flex-direction: column;
    margin: 24px auto !important;
  }

  :deep(.el-dialog__header) {
    padding: 20px 24px 12px;
  }

  :deep(.el-dialog__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  :deep(.el-dialog__footer) {
    padding: 16px 24px 24px;
  }
}

.dialog-content {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.project-section {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.webhook-section {
  flex-shrink: 0;
  padding-bottom: 8px;
}

.selection-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.project-list {
.project-list {
  flex: 1 1 auto;
  max-height: calc(70vh - 230px);
  overflow-y: auto;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background-color: #fafafa;
  margin-bottom: 24px;

  .project-item {
    padding: 10px 12px;
    border-bottom: 1px solid #f0f0f0;

    &:last-child {
      border-bottom: none;
    }

    .project-info {
      margin-left: 8px;
      display: flex;
      flex-direction: column;
      gap: 4px;

      .project-name {
        font-weight: 600;
        color: #303133;
      }

      .project-path {
        font-size: 13px;
        color: #909399;
        word-break: break-all;
      }

      .project-desc {
        font-size: 13px;
        color: #606266;
        line-height: 1.4;
        word-break: break-word;
      }
    }
  }
}

.custom-scroll {
  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    border-radius: 8px;
    background-color: rgba(0, 0, 0, 0.2);
  }

  scrollbar-width: thin;
  scrollbar-color: rgba(0, 0, 0, 0.2) transparent;
}

.webhook-config {
  margin-top: 8px;

  h4 {
    margin-bottom: 16px;
    color: #303133;
  }
}

.webhook-select {
  width: 100%;
}

.section-divider {
  margin: 16px 0 12px;
}

.mb-3 {
  margin-bottom: 12px;
}

.mb-4 {
  margin-bottom: 16px;
}

.mt-3 {
  margin-top: 12px;
}
</style>
