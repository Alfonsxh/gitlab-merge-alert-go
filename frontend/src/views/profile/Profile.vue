<template>
  <div class="page-container">
    <div class="page-header">
      <h2>个人中心</h2>
    </div>
    
    <el-tabs v-model="activeTab" class="profile-tabs">
      <el-tab-pane label="基本信息" name="info">
        <el-card>
          <div class="profile-header">
            <div class="avatar-section">
              <el-upload
                class="avatar-uploader"
                :show-file-list="false"
                :before-upload="beforeAvatarUpload"
                :on-success="handleAvatarSuccess"
                action="#"
                :http-request="uploadAvatar"
              >
                <el-avatar 
                  :size="120" 
                  :src="avatarUrl"
                  class="avatar-display"
                >
                  <el-icon :size="50"><User /></el-icon>
                </el-avatar>
                <div class="avatar-overlay">
                  <el-icon :size="20"><Camera /></el-icon>
                  <span>更换头像</span>
                </div>
              </el-upload>
            </div>
            <div class="info-section">
              <h3>{{ authStore.user?.username || '加载中...' }}</h3>
              <el-tag v-if="authStore.user" :type="authStore.user?.role === 'admin' ? 'danger' : 'primary'">
                {{ authStore.user?.role === 'admin' ? '管理员' : '普通用户' }}
              </el-tag>
            </div>
          </div>
          
          <el-divider />
          
          <el-descriptions :column="1" border>
            <el-descriptions-item label="用户名">
              {{ authStore.user?.username }}
            </el-descriptions-item>
            <el-descriptions-item label="邮箱">
              <div v-if="!editingEmail">
                {{ authStore.user?.email }}
                <el-button link type="primary" @click="startEditEmail">
                  <el-icon><Edit /></el-icon>
                </el-button>
              </div>
              <div v-else style="display: flex; gap: 10px;">
                <el-input v-model="profileData.email" size="small" />
                <el-button type="primary" size="small" @click="saveEmail">保存</el-button>
                <el-button size="small" @click="cancelEditEmail">取消</el-button>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="角色">
              <el-tag :type="authStore.user?.role === 'admin' ? 'danger' : 'primary'">
                {{ authStore.user?.role === 'admin' ? '管理员' : '普通用户' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="账户状态">
              <el-tag :type="authStore.user?.is_active ? 'success' : 'danger'">
                {{ authStore.user?.is_active ? '正常' : '已禁用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="最后登录时间">
              {{ formatDateTime(authStore.user?.last_login_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="注册时间">
              {{ formatDateTime(authStore.user?.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="GitLab Token">
              <div class="token-status">
                <el-tag :type="authStore.user?.has_gitlab_personal_access_token ? 'success' : 'info'">
                  {{ authStore.user?.has_gitlab_personal_access_token ? '已配置' : '未配置' }}
                </el-tag>
                <el-button type="primary" link @click="openTokenDialog">
                  {{ authStore.user?.has_gitlab_personal_access_token ? '更新' : '配置' }}
                </el-button>
              </div>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-tab-pane>
      
      <el-tab-pane label="修改密码" name="password">
        <el-card>
          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-width="120px"
            style="max-width: 500px"
          >
            <el-form-item label="原密码" prop="oldPassword">
              <el-input
                v-model="passwordForm.oldPassword"
                type="password"
                placeholder="请输入原密码"
                show-password
              />
            </el-form-item>
            
            <el-form-item label="新密码" prop="newPassword">
              <el-input
                v-model="passwordForm.newPassword"
                type="password"
                placeholder="请输入新密码（至少6位）"
                show-password
              />
            </el-form-item>
            
            <el-form-item label="确认新密码" prop="confirmPassword">
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                placeholder="请再次输入新密码"
                show-password
              />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="handleChangePassword" :loading="loading">
                确认修改
              </el-button>
              <el-button @click="resetPasswordForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog
      v-model="showTokenDialog"
      title="配置 GitLab Personal Access Token"
      width="600px"
      :close-on-click-modal="false"
      class="token-config-dialog"
    >
      <div class="dialog-content">
        <el-alert type="info" :closable="false" style="margin-bottom: 20px;">
          <template #default>
            <span>GitLab Personal Access Token 用于授权系统访问 GitLab 项目和管理 Webhook</span>
          </template>
        </el-alert>

        <el-form :model="tokenForm" label-width="110px">
          <el-form-item label="Access Token">
            <div style="display: flex; gap: 10px; width: 100%;">
              <el-input
                v-model="tokenForm.gitlab_personal_access_token"
                type="password"
                placeholder="请输入 GitLab Personal Access Token（留空表示清除当前 Token）"
                show-password
                clearable
                style="flex: 1;"
              />
              <el-button
                :icon="Connection"
                @click="testProfileToken"
                :loading="tokenTesting"
              >
                测试
              </el-button>
            </div>
          </el-form-item>
        </el-form>

        <div class="token-help-info">
          <div class="permission-title">
            <el-icon><InfoFilled /></el-icon>
            <span>令牌需具备权限：</span>
          </div>
          <div class="permission-items">
            <div class="permission-item">
              <el-tag size="small">read_api</el-tag>
              <span>- 读取项目信息</span>
            </div>
            <div class="permission-item">
              <el-tag size="small">api</el-tag>
              <span>- 管理 Webhook（可选）</span>
            </div>
          </div>
          <div class="help-links">
            <el-link
              v-if="gitlabUrl"
              :href="`${gitlabUrl}/-/user_settings/personal_access_tokens`"
              target="_blank"
              type="primary"
              :underline="false"
            >
              <el-icon><Link /></el-icon>
              前往 GitLab 创建 Token
            </el-link>
            <el-link
              :href="gitlabPatDocUrl"
              target="_blank"
              type="info"
              :underline="false"
              style="margin-left: 12px;"
            >
              <el-icon><Document /></el-icon>
              查看生成指南
            </el-link>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showTokenDialog = false">取消</el-button>
          <el-button type="primary" :loading="tokenLoading" @click="submitToken">
            <el-icon v-if="!tokenLoading"><Check /></el-icon>
            保存配置
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules, type UploadProps } from 'element-plus'
import { User, Camera, Edit, Connection, InfoFilled, Document, Check, Link } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { authAPI } from '@/api/auth'
import { gitlabApi } from '@/api'

const route = useRoute()
const authStore = useAuthStore()

const activeTab = ref('info')
const loading = ref(false)
const passwordFormRef = ref<FormInstance>()
const editingEmail = ref(false)
const gitlabUrl = ref('')
const tokenTesting = ref(false)
const gitlabPatDocUrl = 'https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html'

// 使用 computed 来动态获取头像，避免初始化时的错误
const avatarUrl = computed(() => {
  if (!authStore.user) return ''
  return (authStore.user as any).avatar || ''
})

const profileData = reactive({
  email: authStore.user?.email || '',
  avatar: ''
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入新密码'))
  } else if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = reactive<FormRules>({
  oldPassword: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
})

const loadGitLabConfig = async () => {
  try {
    const res = await gitlabApi.getConfig()
    gitlabUrl.value = res.data.gitlab_url
  } catch (error) {
    console.error('Failed to fetch GitLab config:', error)
  }
}

const testProfileToken = async () => {
  const token = tokenForm.gitlab_personal_access_token.trim()
  if (!token && !authStore.user?.has_gitlab_personal_access_token) {
    ElMessage.warning('请先输入 GitLab Token')
    return
  }

  tokenTesting.value = true
  try {
    const payload: Record<string, string> = {}
    if (token) {
      payload.access_token = token
    }
    if (gitlabUrl.value) {
      payload.gitlab_url = gitlabUrl.value
    }

    const res: any = await gitlabApi.testToken(payload)
    const result = res?.data ?? res
    if (result?.success) {
      ElMessage.success(result.message || '连接成功')
    } else {
      ElMessage.error(result?.message || '连接失败，请检查 Token 或 GitLab 配置')
    }
  } finally {
    tokenTesting.value = false
  }
}

const showTokenDialog = ref(false)
const tokenLoading = ref(false)
const tokenForm = reactive({
  gitlab_personal_access_token: ''
})

const formatDateTime = (dateTime?: string) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

const handleChangePassword = async () => {
  if (!passwordFormRef.value) return
  
  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    loading.value = true
    try {
      await authAPI.changePassword({
        old_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword
      })
      
      ElMessage.success('密码修改成功，请重新登录')
      
      // 清除认证信息并跳转到登录页
      setTimeout(() => {
        authStore.clearAuthData()
        window.location.href = '/login'
      }, 1500)
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '密码修改失败')
    } finally {
      loading.value = false
    }
  })
}

const resetPasswordForm = () => {
  passwordFormRef.value?.resetFields()
}

const startEditEmail = () => {
  profileData.email = authStore.user?.email || ''
  editingEmail.value = true
}

const cancelEditEmail = () => {
  editingEmail.value = false
  profileData.email = authStore.user?.email || ''
}

const saveEmail = async () => {
  if (!profileData.email) {
    ElMessage.error('邮箱不能为空')
    return
  }
  
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(profileData.email)) {
    ElMessage.error('请输入有效的邮箱地址')
    return
  }
  
  loading.value = true
  try {
    await authAPI.updateProfile({ email: profileData.email })
    
    if (authStore.user) {
      authStore.user.email = profileData.email
    }
    
    ElMessage.success('邮箱更新成功')
    editingEmail.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '更新失败')
  } finally {
    loading.value = false
  }
}

const openTokenDialog = () => {
  tokenForm.gitlab_personal_access_token = ''
  showTokenDialog.value = true
}

const submitToken = async () => {
  tokenLoading.value = true
  try {
    await authAPI.updateProfile({
      gitlab_personal_access_token: tokenForm.gitlab_personal_access_token
    })

    await authStore.fetchProfile()
    ElMessage.success(tokenForm.gitlab_personal_access_token ? 'Token 更新成功' : 'Token 已清除')
    showTokenDialog.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || 'Token 更新失败')
  } finally {
    tokenLoading.value = false
  }
}

const beforeAvatarUpload: UploadProps['beforeUpload'] = (rawFile) => {
  if (!rawFile.type.startsWith('image/')) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (rawFile.size / 1024 / 1024 > 5) {
    ElMessage.error('图片大小不能超过 5MB')
    return false
  }
  return true
}

const uploadAvatar = async ({ file }: any) => {
  const formData = new FormData()
  formData.append('avatar', file)
  
  try {
    const response = await authAPI.uploadAvatar(formData)
    
    // 重新获取用户信息以更新 avatar
    try {
      await authStore.fetchProfile()
    } catch (err) {
      console.error('Failed to refresh profile:', err)
    }
    
    ElMessage.success('头像上传成功')
    return response
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '头像上传失败')
    throw error
  }
}

const handleAvatarSuccess = () => {
  // 不需要手动更新，computed 会自动响应
}

onMounted(async () => {
  // 如果用户信息不存在，尝试获取
  if (!authStore.user) {
    try {
      await authStore.fetchProfile()
    } catch (error) {
      console.error('Failed to fetch profile:', error)
    }
  }
  
  // 初始化数据
  if (authStore.user) {
    profileData.email = authStore.user.email || ''
  }
  
  // 检查是否需要切换到密码标签页
  if (route.query.tab === 'password') {
    activeTab.value = 'password'
  }

  loadGitLabConfig()
})
</script>

<style scoped lang="less">
.page-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
  
  h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: #303133;
  }
}

.profile-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 20px;
  }
}

.el-card {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.el-descriptions {
  :deep(.el-descriptions__label) {
    font-weight: 600;
    color: #606266;
  }
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 30px;
  margin-bottom: 20px;
  
  .avatar-section {
    position: relative;
    
    .avatar-uploader {
      position: relative;
      cursor: pointer;
      
      &:hover .avatar-overlay {
        opacity: 1;
      }
    }
    
    .avatar-display {
      border: 2px solid #e4e7ed;
      background-color: #f5f7fa;
    }
    
    .avatar-overlay {
      position: absolute;
      top: 0;
      left: 0;
      width: 120px;
      height: 120px;
      border-radius: 50%;
      background-color: rgba(0, 0, 0, 0.5);
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      color: white;
      opacity: 0;
      transition: opacity 0.3s;
      
      span {
        margin-top: 5px;
        font-size: 12px;
      }
    }
  }
  
  .info-section {
    flex: 1;
    
    h3 {
      margin: 0 0 10px 0;
      font-size: 24px;
      font-weight: 600;
      color: #303133;
    }
  }
}

.token-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-item-help {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

.token-input-group {
  display: flex;
  align-items: center;
  gap: 8px;

  :deep(.el-input) {
    flex: 1;
  }
}

// Token 配置弹框样式
.token-config-dialog {
  :deep(.el-dialog__body) {
    padding: 20px;
  }

  .dialog-content {
    .token-help-info {
      background-color: #f5f7fa;
      padding: 12px 16px;
      border-radius: 4px;
      margin-top: 16px;

      .permission-title {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 13px;
        color: #606266;
        margin-bottom: 8px;

        .el-icon {
          color: #909399;
        }
      }

      .permission-items {
        padding-left: 22px;
        margin-bottom: 12px;

        .permission-item {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 13px;
          color: #606266;
          line-height: 24px;

          .el-tag {
            flex-shrink: 0;
          }
        }
      }

      .help-links {
        padding-left: 22px;

        .el-link {
          font-size: 13px;
          display: inline-flex;
          align-items: center;
          gap: 4px;

          .el-icon {
            font-size: 14px;
          }
        }
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;

  .el-button {
    min-width: 100px;

    .el-icon {
      margin-right: 4px;
    }
  }
}
</style>
