<template>
  <div class="page-container">
    <div class="page-header">
      <h2>个人中心</h2>
    </div>
    
    <el-tabs v-model="activeTab" class="profile-tabs">
      <el-tab-pane label="基本信息" name="info">
        <el-card>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="用户名">
              {{ authStore.user?.username }}
            </el-descriptions-item>
            <el-descriptions-item label="邮箱">
              {{ authStore.user?.email }}
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { authAPI } from '@/api/auth'

const route = useRoute()
const authStore = useAuthStore()

const activeTab = ref('info')
const loading = ref(false)
const passwordFormRef = ref<FormInstance>()

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

onMounted(() => {
  // 检查是否需要切换到密码标签页
  if (route.query.tab === 'password') {
    activeTab.value = 'password'
  }
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
</style>