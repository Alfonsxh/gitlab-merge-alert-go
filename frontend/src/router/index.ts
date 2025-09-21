import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import { setFaviconFromAvatar } from '@/utils/favicon'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { title: '登录', requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/Register.vue'),
    meta: { title: '注册', requiresAuth: false }
  },
  {
    path: '/setup-admin',
    name: 'SetupAdmin',
    component: () => import('@/views/system/SetupAdmin.vue'),
    meta: { title: '管理员初始化', requiresAuth: false }
  },
  {
    path: '/',
    component: MainLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/Dashboard.vue'),
        meta: { title: '仪表板', requiresAuth: true }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/Users.vue'),
        meta: { title: '用户管理', requiresAuth: true }
      },
      {
        path: 'projects',
        name: 'Projects',
        component: () => import('@/views/projects/Projects.vue'),
        meta: { title: '项目管理', requiresAuth: true }
      },
      {
        path: 'webhooks',
        name: 'Webhooks',
        component: () => import('@/views/webhooks/Webhooks.vue'),
        meta: { title: 'Webhook管理', requiresAuth: true }
      },
      {
        path: 'accounts',
        name: 'Accounts',
        component: () => import('@/views/accounts/Accounts.vue'),
        meta: { title: '账户管理', requiresAuth: true, requiresAdmin: true }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/Profile.vue'),
        meta: { title: '个人中心', requiresAuth: true }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 全局路由守卫
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  const systemStore = useSystemStore()

  const setupRequired = await systemStore.checkAdminSetup()

  if (setupRequired && to.path !== '/setup-admin') {
    const redirect = to.fullPath !== '/setup-admin' ? to.fullPath : undefined
    return next({ path: '/setup-admin', query: redirect ? { redirect } : undefined })
  }

  if (!setupRequired && to.path === '/setup-admin') {
    const redirect = (to.query.redirect as string) || '/login'
    return next(redirect)
  }

  // 设置页面标题
  document.title = `${to.meta.title} - GitLab Merge Alert` || 'GitLab Merge Alert'
  
  // 检查路由是否需要认证
  const requiresAuth = to.meta.requiresAuth !== false
  const requiresAdmin = to.meta.requiresAdmin === true
  
  if (requiresAuth) {
    // 检查是否已认证
    if (!authStore.isAuthenticated) {
      // 未认证，跳转到登录页
      setFaviconFromAvatar(null) // 重置为默认 favicon
      return next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
    }
    
    // 检查Token是否过期
    if (!authStore.checkTokenExpiry()) {
      // Token过期，跳转到登录页
      setFaviconFromAvatar(null) // 重置为默认 favicon
      return next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
    }
    
    // 如果没有用户信息，尝试获取
    if (!authStore.user) {
      try {
        await authStore.fetchProfile()
      } catch (error) {
        // 获取用户信息失败，跳转到登录页
        setFaviconFromAvatar(null) // 重置为默认 favicon
        return next({
          path: '/login',
          query: { redirect: to.fullPath }
        })
      }
    }
    
    // 设置用户头像为 favicon
    if (authStore.user?.avatar) {
      setFaviconFromAvatar(authStore.user.avatar)
    }
    
    // 检查管理员权限
    if (requiresAdmin && !authStore.isAdmin) {
      // 没有管理员权限，跳转到首页
      return next('/')
    }
  } else {
    // 登录页面，重置为默认 favicon
    setFaviconFromAvatar(null)
  }
  
  // 已登录用户访问登录页，重定向到首页
  if (['/login', '/register'].includes(to.path) && authStore.isAuthenticated) {
    return next('/')
  }
  
  next()
})

// 监听用户活动事件
window.addEventListener('userActivity', () => {
  const authStore = useAuthStore()
  authStore.updateLastActivity()
})

export default router
