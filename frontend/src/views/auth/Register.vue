<template>
  <div class="register-container">
    <div class="register-box">
      <div class="register-header">
        <h1 class="register-title">GitLab Merge Alert</h1>
        <p class="register-subtitle">注册普通用户账户</p>
      </div>

      <form @submit.prevent="handleRegister" class="register-form">
        <div class="form-group">
          <label for="username" class="form-label">用户名</label>
          <input
            id="username"
            v-model="registerForm.username"
            type="text"
            class="form-control"
            placeholder="请输入用户名"
            required
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="email" class="form-label">邮箱</label>
          <input
            id="email"
            v-model="registerForm.email"
            type="email"
            class="form-control"
            placeholder="请输入邮箱"
            required
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="password" class="form-label">密码</label>
          <input
            id="password"
            v-model="registerForm.password"
            type="password"
            class="form-control"
            placeholder="请输入密码"
            required
            minlength="6"
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="confirmPassword" class="form-label">确认密码</label>
          <input
            id="confirmPassword"
            v-model="registerForm.confirmPassword"
            type="password"
            class="form-control"
            placeholder="请再次输入密码"
            required
            minlength="6"
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label for="gitlabToken" class="form-label">GitLab Token</label>
          <input
            id="gitlabToken"
            v-model="registerForm.gitlabToken"
            type="password"
            class="form-control"
            placeholder="请输入 GitLab Personal Access Token"
            required
            :disabled="loading"
          />
          <p class="token-hint">
            <span class="token-hint__text">需具备 <code>api</code> / <code>read_api</code> / <code>read_user</code> 权限。</span>
            <a class="token-hint__link" :href="gitlabPatLink" target="_blank" rel="noopener">前往生成 Personal Access Token</a>
          </p>
        </div>

        <button
          type="submit"
          class="btn btn-primary btn-block"
          :disabled="loading"
        >
          <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
          {{ loading ? '注册中...' : '注册' }}
        </button>
      </form>

      <p v-if="error" class="error-banner">
        {{ error }}
      </p>

      <p class="register-note">
        系统仅允许默认 admin 账号拥有管理员权限，注册后将创建普通用户。
      </p>

      <div class="register-footer">
        已有账户？
        <RouterLink to="/login">立即登录</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { gitlabApi } from '@/api/gitlab'

const router = useRouter()
const authStore = useAuthStore()

const registerForm = ref({
	username: '',
	email: '',
	password: '',
	confirmPassword: '',
	gitlabToken: ''
})

const loading = ref(false)
const error = ref('')
const gitlabUrl = ref('')

const gitlabPatLink = computed(() => {
	const base = gitlabUrl.value?.replace(/\/$/, '')
	return base ? `${base}/-/profile/personal_access_tokens` : 'https://gitlab.com/-/profile/personal_access_tokens'
})

	onMounted(async () => {
		try {
			const res = await gitlabApi.getPublicConfig()
			gitlabUrl.value = res.data.gitlab_url
		} catch (err) {
			console.error('Failed to load GitLab config:', err)
		}
	})

const handleRegister = async () => {
	if (registerForm.value.password !== registerForm.value.confirmPassword) {
		error.value = '两次输入的密码不一致'
		return
	}

	if (!registerForm.value.gitlabToken.trim()) {
		error.value = 'GitLab Token 不能为空'
		return
	}

	loading.value = true
	error.value = ''

	try {
		await authStore.register({
			username: registerForm.value.username.trim(),
			email: registerForm.value.email.trim(),
			password: registerForm.value.password,
			gitlab_personal_access_token: registerForm.value.gitlabToken.trim()
		})

    const redirect = (router.currentRoute.value.query.redirect as string) || '/'
    router.push(redirect)
  } catch (err: any) {
    error.value = err.response?.data?.error || '注册失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
}

.register-box {
  width: 100%;
  max-width: 420px;
  padding: 2rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.register-header {
  text-align: center;
  margin-bottom: 2rem;
}

.register-title {
  font-size: 1.75rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 0.5rem;
}

.register-subtitle {
  color: #666;
  font-size: 0.95rem;
}

.register-form {
  margin-top: 1.5rem;
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
  line-height: 1.5;
  color: #495057;
  background-color: #fff;
  background-clip: padding-box;
  border: 1px solid #ced4da;
  border-radius: 0.25rem;
  transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
}

.form-control:focus {
  color: #495057;
  background-color: #fff;
  border-color: #80bdff;
  outline: 0;
  box-shadow: 0 0 0 0.2rem rgba(0, 123, 255, 0.25);
}

.form-control:disabled {
  background-color: #e9ecef;
  opacity: 1;
}

.token-hint {
  margin-top: 0.5rem;
  font-size: 0.85rem;
  color: #666;
  line-height: 1.5;
}

.token-hint__text {
  display: block;
  margin-bottom: 0.25rem;
}

.token-hint code {
  background-color: #f5f5f5;
  padding: 0 4px;
  border-radius: 4px;
  font-size: 0.8rem;
}

.token-hint__link {
  display: inline-block;
  color: #409eff;
  text-decoration: none;
}

.token-hint__link:hover {
  text-decoration: underline;
}

.error-banner {
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  background: #fdecea;
  color: #d93025;
  border: 1px solid rgba(217, 48, 37, 0.35);
  border-radius: 6px;
  font-size: 0.95rem;
  font-weight: 600;
}

.btn {
  display: inline-block;
  font-weight: 400;
  text-align: center;
  white-space: nowrap;
  vertical-align: middle;
  user-select: none;
  padding: 0.5rem 1rem;
  font-size: 1rem;
  line-height: 1.5;
  border-radius: 0.25rem;
  width: 100%;
}

.btn-block {
  width: 100%;
}

.register-footer {
  margin-top: 1.5rem;
  text-align: center;
  color: #666;
}

.register-note {
  margin-top: 1.5rem;
  font-size: 0.9rem;
  color: #888;
  text-align: center;
  line-height: 1.4;
}

.register-footer a {
  color: #0d6efd;
  text-decoration: none;
  margin-left: 0.25rem;
}

.register-footer a:hover {
  text-decoration: underline;
}
</style>
