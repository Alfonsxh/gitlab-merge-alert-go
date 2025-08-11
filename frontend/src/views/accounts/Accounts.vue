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
      <el-table-column label="操作" width="280" fixed="right">
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
            type="success"
            size="small"
            link
            @click="handleAssignResources(row)"
          >
            分配资源
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

    <!-- 资源分配对话框 -->
    <el-dialog
      v-model="showAssignDialog"
      title="资源分配"
      width="800px"
      destroy-on-close
    >
      <div class="assign-header">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="账户">
            {{ assignForm.username }}
          </el-descriptions-item>
          <el-descriptions-item label="角色">
            <el-tag :type="assignForm.role === 'admin' ? 'danger' : 'primary'">
              {{ assignForm.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <el-tabs v-model="activeResourceTab" class="resource-tabs">
        <el-tab-pane label="项目管理权限" name="projects">
          <div class="resource-section">
            <div class="section-header">
              <span>选择该账户可以管理的项目</span>
              <el-button size="small" @click="selectAllProjects">全选</el-button>
              <el-button size="small" @click="clearAllProjects">清空</el-button>
            </div>
            <el-transfer
              v-model="assignedProjects"
              :data="projectList"
              :titles="['可选项目', '已分配项目']"
              :props="{
                key: 'id',
                label: 'name',
                disabled: 'disabled'
              }"
              filterable
              filter-placeholder="搜索项目"
            >
              <template #default="{ option }">
                <span>{{ option.name }}</span>
                <el-tag v-if="option.assigned_to && option.assigned_to !== assignForm.id" 
                  type="info" size="small" style="margin-left: 8px">
                  已分配给: {{ option.assigned_to_name }}
                </el-tag>
              </template>
            </el-transfer>
          </div>
        </el-tab-pane>

        <el-tab-pane label="Webhook管理权限" name="webhooks">
          <div class="resource-section">
            <div class="section-header">
              <span>选择该账户可以管理的Webhook</span>
              <el-button size="small" @click="selectAllWebhooks">全选</el-button>
              <el-button size="small" @click="clearAllWebhooks">清空</el-button>
            </div>
            <el-transfer
              v-model="assignedWebhooks"
              :data="webhookList"
              :titles="['可选Webhook', '已分配Webhook']"
              :props="{
                key: 'id',
                label: 'name',
                disabled: 'disabled'
              }"
              filterable
              filter-placeholder="搜索Webhook"
            >
              <template #default="{ option }">
                <span>{{ option.name }}</span>
                <el-tag v-if="option.assigned_to && option.assigned_to !== assignForm.id" 
                  type="info" size="small" style="margin-left: 8px">
                  已分配给: {{ option.assigned_to_name }}
                </el-tag>
              </template>
            </el-transfer>
          </div>
        </el-tab-pane>

        <el-tab-pane label="用户管理权限" name="users">
          <div class="resource-section">
            <div class="section-header">
              <span>选择该账户可以管理的用户</span>
              <el-button size="small" @click="selectAllUsers">全选</el-button>
              <el-button size="small" @click="clearAllUsers">清空</el-button>
            </div>
            <el-transfer
              v-model="assignedUsers"
              :data="userList"
              :titles="['可选用户', '已分配用户']"
              :props="{
                key: 'id',
                label: 'name',
                disabled: 'disabled'
              }"
              filterable
              filter-placeholder="搜索用户"
            >
              <template #default="{ option }">
                <span>{{ option.name }}</span>
                <el-tag v-if="option.assigned_to && option.assigned_to !== assignForm.id" 
                  type="info" size="small" style="margin-left: 8px">
                  已分配给: {{ option.assigned_to_name }}
                </el-tag>
              </template>
            </el-transfer>
          </div>
        </el-tab-pane>
      </el-tabs>

      <template #footer>
        <el-button @click="showAssignDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveAssignment" :loading="assignLoading">
          保存分配
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
import { projectsApi } from '@/api/projects'
import { webhooksApi } from '@/api/webhooks'
import { usersApi } from '@/api/users'
import { resourceManagerAPI } from '@/api/resource-manager'
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

// 资源分配
const showAssignDialog = ref(false)
const assignLoading = ref(false)
const activeResourceTab = ref('projects')
const assignForm = reactive({
  id: 0,
  username: '',
  role: ''
})
const projectList = ref<any[]>([])
const webhookList = ref<any[]>([])
const userList = ref<any[]>([])
const assignedProjects = ref<number[]>([])
const assignedWebhooks = ref<number[]>([])
const assignedUsers = ref<number[]>([])

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

// 资源分配
const handleAssignResources = async (row: AccountResponse) => {
  assignForm.id = row.id
  assignForm.username = row.username
  assignForm.role = row.role
  
  // 加载资源列表
  await loadResourceLists()
  
  // 加载所有资源的分配情况
  await loadAllResourceAssignments()
  
  // 加载已分配的资源
  await loadAssignedResources(row.id)
  
  showAssignDialog.value = true
}

// 加载所有资源的分配情况
const loadAllResourceAssignments = async () => {
  try {
    // 获取所有账户的资源分配情况
    const accounts = await accountAPI.getAccounts({ page_size: 1000 })
    
    // 为每个资源标记是否已分配及分配给谁
    const projectAssignments = new Map<number, { id: number, name: string }>()
    const webhookAssignments = new Map<number, { id: number, name: string }>()
    const userAssignments = new Map<number, { id: number, name: string }>()
    
    // 并行获取所有账户的资源分配
    await Promise.all(accounts.data.map(async (account: any) => {
      if (account.id === assignForm.id) return // 跳过当前账户
      
      const [projects, webhooks, users] = await Promise.all([
        resourceManagerAPI.getManagedResources(account.id, 'project').catch(() => ({ resource_ids: [] })),
        resourceManagerAPI.getManagedResources(account.id, 'webhook').catch(() => ({ resource_ids: [] })),
        resourceManagerAPI.getManagedResources(account.id, 'user').catch(() => ({ resource_ids: [] }))
      ])
      
      projects.resource_ids?.forEach((id: number) => {
        projectAssignments.set(id, { id: account.id, name: account.username })
      })
      webhooks.resource_ids?.forEach((id: number) => {
        webhookAssignments.set(id, { id: account.id, name: account.username })
      })
      users.resource_ids?.forEach((id: number) => {
        userAssignments.set(id, { id: account.id, name: account.username })
      })
    }))
    
    // 更新资源列表的分配状态
    projectList.value = projectList.value.map(p => ({
      ...p,
      disabled: projectAssignments.has(p.id),
      assigned_to: projectAssignments.get(p.id)?.id,
      assigned_to_name: projectAssignments.get(p.id)?.name
    }))
    
    webhookList.value = webhookList.value.map(w => ({
      ...w,
      disabled: webhookAssignments.has(w.id),
      assigned_to: webhookAssignments.get(w.id)?.id,
      assigned_to_name: webhookAssignments.get(w.id)?.name
    }))
    
    userList.value = userList.value.map(u => ({
      ...u,
      disabled: userAssignments.has(u.id),
      assigned_to: userAssignments.get(u.id)?.id,
      assigned_to_name: userAssignments.get(u.id)?.name
    }))
  } catch (error) {
    console.error('Failed to load resource assignments:', error)
  }
}

const loadResourceLists = async () => {
  try {
    // 并行加载所有资源列表
    const [projects, webhooks, users] = await Promise.all([
      projectsApi.getProjects({ page_size: 1000 }), // 获取所有资源
      webhooksApi.getWebhooks({ page_size: 1000 }),
      usersApi.getUsers({ page_size: 1000 })
    ])
    
    projectList.value = projects.data.map((p: any) => ({
      id: p.id,
      name: p.name || p.path,
      disabled: false // 标记资源是否已分配给其他账户
    }))
    
    webhookList.value = webhooks.data.map((w: any) => ({
      id: w.id,
      name: w.name,
      disabled: false
    }))
    
    userList.value = users.data.map((u: any) => ({
      id: u.id,
      name: `${u.name || u.email} (${u.email})`,
      disabled: false
    }))
  } catch (error) {
    console.error('Failed to load resource lists:', error)
    ElMessage.error('加载资源列表失败')
  }
}

const loadAssignedResources = async (accountId: number) => {
  try {
    const [projects, webhooks, users] = await Promise.all([
      resourceManagerAPI.getManagedResources(accountId, 'project'),
      resourceManagerAPI.getManagedResources(accountId, 'webhook'),
      resourceManagerAPI.getManagedResources(accountId, 'user')
    ])
    
    console.log('Loaded assigned resources for account', accountId, {
      projects,
      webhooks,
      users
    })
    
    // 后端返回格式: { resource_ids: number[], total: number }
    // apiClient 已经解包了 response.data，所以直接访问即可
    assignedProjects.value = projects?.resource_ids || []
    assignedWebhooks.value = webhooks?.resource_ids || []
    assignedUsers.value = users?.resource_ids || []
    
    console.log('Set assigned values:', {
      assignedProjects: assignedProjects.value,
      assignedWebhooks: assignedWebhooks.value,
      assignedUsers: assignedUsers.value
    })
  } catch (error) {
    console.error('Failed to load assigned resources:', error)
    // 初始化为空数组
    assignedProjects.value = []
    assignedWebhooks.value = []
    assignedUsers.value = []
  }
}

const selectAllProjects = () => {
  assignedProjects.value = projectList.value.map(p => p.id)
}

const clearAllProjects = () => {
  assignedProjects.value = []
}

const selectAllWebhooks = () => {
  assignedWebhooks.value = webhookList.value.map(w => w.id)
}

const clearAllWebhooks = () => {
  assignedWebhooks.value = []
}

const selectAllUsers = () => {
  assignedUsers.value = userList.value.map(u => u.id)
}

const clearAllUsers = () => {
  assignedUsers.value = []
}

const handleSaveAssignment = async () => {
  assignLoading.value = true
  
  try {
    // 保存各类资源的分配
    const assignments = [
      ...assignedProjects.value.map(id => ({ 
        resource_id: id, 
        resource_type: 'project' as const,
        manager_id: assignForm.id
      })),
      ...assignedWebhooks.value.map(id => ({ 
        resource_id: id, 
        resource_type: 'webhook' as const,
        manager_id: assignForm.id
      })),
      ...assignedUsers.value.map(id => ({ 
        resource_id: id, 
        resource_type: 'user' as const,
        manager_id: assignForm.id
      }))
    ]
    
    // 批量更新资源分配
    await resourceManagerAPI.batchAssign(assignForm.id, assignments)
    
    ElMessage.success('资源分配成功')
    showAssignDialog.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '资源分配失败')
  } finally {
    assignLoading.value = false
  }
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

.assign-header {
  margin-bottom: 20px;
}

.resource-tabs {
  margin-top: 20px;
  
  .resource-section {
    .section-header {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 15px;
      
      span {
        flex: 1;
        font-weight: 500;
        color: #606266;
      }
    }
    
    :deep(.el-transfer) {
      .el-transfer-panel {
        width: 300px;
      }
    }
  }
}
</style>