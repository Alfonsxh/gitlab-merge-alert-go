<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">账户管理</h1>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        创建账户
      </el-button>
    </div>

    <el-card>
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-form :inline="true">
          <el-form-item label="搜索">
            <el-input
              v-model="searchForm.search"
              placeholder="用户名或邮箱"
              clearable
              @clear="handleSearch"
              @keyup.enter="handleSearch"
            />
          </el-form-item>
          <el-form-item label="角色">
            <el-select v-model="searchForm.role" clearable @change="handleSearch" style="width: 150px" placeholder="全部">
              <el-option label="全部" value="" />
              <el-option label="管理员" value="admin" />
              <el-option label="普通用户" value="user" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">搜索</el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 账户列表 -->
      <el-table
        v-loading="loading"
        :data="accountList"
        stripe
        style="width: 100%"
      >
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="role" label="角色" width="100">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'primary'">
            {{ row.role === 'admin' ? '管理员' : '普通用户' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'danger'">
            {{ row.is_active ? '正常' : '已禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_login_at" label="最后登录" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.last_login_at) }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button
            type="primary"
            size="small"
            link
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button
            type="warning"
            size="small"
            link
            @click="handleResetPassword(row)"
          >
            重置密码
          </el-button>
          <el-button
            type="danger"
            size="small"
            link
            @click="handleDelete(row)"
            :disabled="row.id === authStore.user?.id"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchAccounts"
          @current-change="fetchAccounts"
        />
      </div>
    </el-card>

    <!-- 创建账户对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建账户"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="createForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="createForm.password"
            type="password"
            placeholder="请输入密码（至少6位）"
            show-password
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="createForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="createLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑账户对话框 -->
    <el-dialog
      v-model="showEditDialog"
      title="编辑账户"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="80px"
      >
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="editForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="is_active">
          <el-switch v-model="editForm.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdate" :loading="editLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="showResetPasswordDialog"
      title="重置密码"
      width="400px"
      destroy-on-close
    >
      <el-form
        ref="resetPasswordFormRef"
        :model="resetPasswordForm"
        :rules="resetPasswordRules"
        label-width="80px"
      >
        <el-form-item label="用户名">
          <el-input v-model="resetPasswordForm.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="resetPasswordForm.newPassword"
            type="password"
            placeholder="请输入新密码（至少6位）"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showResetPasswordDialog = false">取消</el-button>
        <el-button type="primary" @click="handleResetPasswordConfirm" :loading="resetPasswordLoading">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { accountAPI } from '@/api/auth'
import type { AccountResponse } from '@/api/types/auth'

const authStore = useAuthStore()

// 数据
const loading = ref(false)
const accountList = ref<AccountResponse[]>([])
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 搜索表单
const searchForm = reactive({
  search: '',
  role: ''
})

// 创建账户
const showCreateDialog = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({
  username: '',
  password: '',
  email: '',
  role: 'user'
})
const createRules = reactive<FormRules>({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 30, message: '用户名长度在 3 到 30 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
})

// 编辑账户
const showEditDialog = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive({
  id: 0,
  username: '',
  email: '',
  role: '',
  is_active: true
})
const editRules = reactive<FormRules>({
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
})

// 重置密码
const showResetPasswordDialog = ref(false)
const resetPasswordLoading = ref(false)
const resetPasswordFormRef = ref<FormInstance>()
const resetPasswordForm = reactive({
  id: 0,
  username: '',
  newPassword: ''
})
const resetPasswordRules = reactive<FormRules>({
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
})

// 格式化时间
const formatDateTime = (dateTime?: string) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

// 获取账户列表
const fetchAccounts = async () => {
  loading.value = true
  try {
    const response = await accountAPI.getAccounts({
      page: pagination.page,
      page_size: pagination.pageSize,
      search: searchForm.search,
      role: searchForm.role
    })
    accountList.value = response.data
    pagination.total = response.total
  } catch (error) {
    console.error('Failed to fetch accounts:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchAccounts()
}

// 重置搜索
const resetSearch = () => {
  searchForm.search = ''
  searchForm.role = ''
  handleSearch()
}

// 创建账户
const handleCreate = async () => {
  if (!createFormRef.value) return
  
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    createLoading.value = true
    try {
      await accountAPI.createAccount(createForm)
      ElMessage.success('账户创建成功')
      showCreateDialog.value = false
      fetchAccounts()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '创建失败')
    } finally {
      createLoading.value = false
    }
  })
}

// 编辑账户
const handleEdit = (row: AccountResponse) => {
  editForm.id = row.id
  editForm.username = row.username
  editForm.email = row.email
  editForm.role = row.role
  editForm.is_active = row.is_active
  showEditDialog.value = true
}

// 更新账户
const handleUpdate = async () => {
  if (!editFormRef.value) return
  
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    editLoading.value = true
    try {
      await accountAPI.updateAccount(editForm.id, {
        email: editForm.email,
        role: editForm.role,
        is_active: editForm.is_active
      })
      ElMessage.success('账户更新成功')
      showEditDialog.value = false
      fetchAccounts()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '更新失败')
    } finally {
      editLoading.value = false
    }
  })
}

// 重置密码
const handleResetPassword = (row: AccountResponse) => {
  resetPasswordForm.id = row.id
  resetPasswordForm.username = row.username
  resetPasswordForm.newPassword = ''
  showResetPasswordDialog.value = true
}

// 确认重置密码
const handleResetPasswordConfirm = async () => {
  if (!resetPasswordFormRef.value) return
  
  await resetPasswordFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    resetPasswordLoading.value = true
    try {
      await accountAPI.resetPassword(resetPasswordForm.id, resetPasswordForm.newPassword)
      ElMessage.success('密码重置成功')
      showResetPasswordDialog.value = false
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '重置失败')
    } finally {
      resetPasswordLoading.value = false
    }
  })
}

// 删除账户
const handleDelete = async (row: AccountResponse) => {
  await ElMessageBox.confirm(
    `确定要删除账户 "${row.username}" 吗？删除后无法恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
  
  try {
    await accountAPI.deleteAccount(row.id)
    ElMessage.success('账户删除成功')
    fetchAccounts()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '删除失败')
  }
}

onMounted(() => {
  fetchAccounts()
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

.el-card {
  :deep(.el-card__body) {
    padding: 0;
  }
}

.search-bar {
  padding: 16px 20px;
  border-bottom: 1px solid #e6e8eb;
  flex-shrink: 0;
  
  :deep(.el-card__body) {
    padding-bottom: 0;
  }
}

// 表格样式已默认设置

.pagination-wrapper {
  padding: 16px 20px;
  border-top: 1px solid #e6e8eb;
  flex-shrink: 0;
  background: #fff;
  display: flex;
  justify-content: center;
}
</style>