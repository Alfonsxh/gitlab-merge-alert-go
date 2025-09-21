<template>
  <div class="setup-container">
    <div class="setup-box">
      <div class="setup-header">
        <h1 class="setup-title">管理员账户初始化</h1>
        <p class="setup-subtitle">使用启动日志中提供的 Setup Token 完成首次密码设置</p>
      </div>

      <div class="info-card">
        <p>请在服务器启动日志中查找 <code>Admin setup token</code>，并在下方输入该 token 以及新的管理员邮箱和密码。</p>
        <p class="tip">出于安全考虑，token 仅用于首次初始化，一旦设置成功将立即失效。</p>
      </div>

      <form class="setup-form" @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="token" class="form-label">Setup Token</label>
          <input
            id="token"
            v-model.trim="form.token"
            type="text"
            class="form-control"
            placeholder="请输入 Setup Token"
            required
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="email" class="form-label">管理员邮箱</label>
          <input
            id="email"
            v-model.trim="form.email"
            type="email"
            class="form-control"
            placeholder="请输入管理员邮箱"
            required
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="password" class="form-label">新密码</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            class="form-control"
            placeholder="请输入至少 6 位的新密码"
            minlength="6"
            required
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="confirm" class="form-label">确认密码</label>
          <input
            id="confirm"
            v-model="form.confirm"
            type="password"
            class="form-control"
            placeholder="请再次输入新密码"
            minlength="6"
            required
            :disabled="loading"
          />
        </div>

        <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
          <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
          {{ loading ? '提交中...' : '完成初始化' }}
        </button>
      </form>

      <div v-if="error" class="alert alert-danger mt-3">
        {{ error }}
      </div>

      <div v-if="success" class="alert alert-success mt-3">
        {{ success }}
        <div class="mt-2">
          <button class="btn btn-link p-0" type="button" @click="goToLogin">返回登录</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useSystemStore } from '@/stores/system'

const router = useRouter()
const route = useRoute()
const systemStore = useSystemStore()

const form = reactive({
  token: '',
  email: '',
  password: '',
  confirm: ''
})

const loading = ref(false)
const error = ref('')
const success = ref('')

const goToLogin = () => {
  const redirect = (route.query.redirect as string) || '/login'
  router.push(redirect)
}

const handleSubmit = async () => {
  error.value = ''
  success.value = ''

  if (form.password !== form.confirm) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  try {
    await systemStore.setupAdmin({
      token: form.token,
      email: form.email,
      password: form.password
    })

    success.value = '管理员账户已成功初始化，请使用新密码登录。'
  } catch (err: any) {
    error.value = err.response?.data?.error || '初始化失败，请检查 token 是否正确'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.setup-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
  padding: 1.5rem;
}

.setup-box {
  width: 100%;
  max-width: 520px;
  padding: 2rem;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.setup-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.setup-title {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.setup-subtitle {
  color: #666;
  font-size: 0.95rem;
}

.info-card {
  background-color: #f0f7ff;
  border: 1px solid #c8e1ff;
  border-radius: 6px;
  padding: 1rem;
  margin-bottom: 1.5rem;
  color: #1f4b8c;
}

.info-card code {
  background: rgba(0, 0, 0, 0.04);
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
}

.tip {
  margin-top: 0.5rem;
  font-size: 0.9rem;
}

.setup-form {
  margin-top: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #333;
}

.form-control {
  width: 100%;
  padding: 0.5rem 0.75rem;
  font-size: 1rem;
  border: 1px solid #ced4da;
  border-radius: 4px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.form-control:focus {
  outline: none;
  border-color: #409eff;
  box-shadow: 0 0 0 0.2rem rgba(64, 158, 255, 0.2);
}

.form-control:disabled {
  background-color: #f0f0f0;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.btn-block {
  width: 100%;
}

.alert {
  border-radius: 4px;
  padding: 0.75rem 1rem;
}

.alert-danger {
  background-color: #fdeaea;
  border: 1px solid #f5c2c7;
  color: #842029;
}

.alert-success {
  background-color: #ecfdf3;
  border: 1px solid #a3e2b6;
  color: #0f5132;
}

.btn-link {
  color: #409eff;
  text-decoration: none;
}

.btn-link:hover {
  text-decoration: underline;
}
</style>
