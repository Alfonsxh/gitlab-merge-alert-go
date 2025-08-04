<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">用户管理</h1>
      <el-button type="primary" @click="showAddModal" size="large">
        <el-icon><Plus /></el-icon>
        添加用户
      </el-button>
    </div>
    
    <el-card>
      <el-table
        :data="users"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="60" />
        
        <el-table-column prop="email" label="邮箱" width="280" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="email-cell">
              <el-icon><Message /></el-icon>
              <span>{{ row.email }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="phone" label="手机号" width="140">
          <template #default="{ row }">
            <el-tag type="info" class="phone-tag">
              {{ formatPhone(row.phone) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="gitlab_username" label="GitLab用户名" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="gitlab-username">
              <el-icon><UserFilled /></el-icon>
              {{ row.gitlab_username || '-' }}
            </span>
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editUser(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-popconfirm
              title="确定要删除这个用户吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="deleteUser(row.id)"
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
    
    <!-- 添加/编辑用户对话框 -->
    <el-dialog
      v-model="modalVisible"
      :title="isEditing ? '编辑用户' : '添加用户'"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="currentUser"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="currentUser.email"
            placeholder="请输入GitLab上的用户邮箱"
            @input="onEmailInput"
          >
            <template #prefix>
              <el-icon><Message /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="currentUser.phone"
            placeholder="请输入企业微信上注册的手机号"
          >
            <template #prefix>
              <el-icon><Phone /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="GitLab用户名" prop="gitlab_username">
          <el-input
            v-model="currentUser.gitlab_username"
            placeholder="请输入GitLab中的用户名"
            @input="onGitLabUsernameInput"
          >
            <template #prefix>
              <el-icon><UserFilled /></el-icon>
            </template>
          </el-input>
          <div class="form-item-help">
            系统会自动从邮箱提取用户名，您也可以手动修改
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="modalVisible = false">取消</el-button>
        <el-button type="primary" @click="saveUser" :loading="submitting">
          {{ isEditing ? '更新' : '添加' }}
        </el-button>
      </template>
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
  Message,
  Phone,
  UserFilled
} from '@element-plus/icons-vue'
import { usersApi } from '@/api'
import type { User } from '@/api'
import { formatDate, formatPhone } from '@/utils/format'

const users = ref<User[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const submitting = ref(false)
const isEditing = ref(false)
const userModifiedGitLabUsername = ref(false)
const formRef = ref<FormInstance>()

const currentUser = reactive<Partial<User>>({
  email: '',
  phone: '',
  gitlab_username: ''
})

const rules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: ['blur', 'change'] }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1\d{10}$/, message: '请输入有效的手机号', trigger: ['blur', 'change'] }
  ]
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await usersApi.getUsers()
    users.value = res.data || []
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    loading.value = false
  }
}

const showAddModal = () => {
  Object.assign(currentUser, {
    id: undefined,
    email: '',
    phone: '',
    gitlab_username: ''
  })
  isEditing.value = false
  userModifiedGitLabUsername.value = false
  modalVisible.value = true
}

const editUser = (user: User) => {
  Object.assign(currentUser, user)
  isEditing.value = true
  userModifiedGitLabUsername.value = true
  modalVisible.value = true
}

const saveUser = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  
  submitting.value = true
  try {
    if (isEditing.value && currentUser.id) {
      await usersApi.updateUser(currentUser.id, currentUser)
    } else {
      await usersApi.createUser(currentUser)
    }
    
    ElMessage.success(isEditing.value ? '更新成功' : '添加成功')
    modalVisible.value = false
    await loadUsers()
  } catch (error) {
    // 错误已在 API 客户端处理
  } finally {
    submitting.value = false
  }
}

const deleteUser = async (userId: number) => {
  try {
    await usersApi.deleteUser(userId)
    ElMessage.success('删除成功')
    await loadUsers()
  } catch (error) {
    // 错误已在 API 客户端处理
  }
}

const onEmailInput = () => {
  // 只在新增模式且用户没有手动修改过GitLab用户名时自动填充
  if (!isEditing.value && !userModifiedGitLabUsername.value) {
    const email = currentUser.email
    if (email && email.includes('@')) {
      const username = email.split('@')[0]
      currentUser.gitlab_username = username
    }
  }
}

const onGitLabUsernameInput = () => {
  userModifiedGitLabUsername.value = true
}

onMounted(() => {
  loadUsers()
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
  .email-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    
    .el-icon {
      color: #409eff;
    }
  }
  
  .phone-tag {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
    font-size: 13px;
  }
  
  .gitlab-username {
    display: flex;
    align-items: center;
    gap: 6px;
    
    .el-icon {
      color: #67c23a;
    }
  }
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