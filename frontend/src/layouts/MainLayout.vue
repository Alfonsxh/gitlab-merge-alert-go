<template>
  <el-container class="layout-container">
    <el-header class="layout-header">
      <div class="header-content">
        <div class="logo">
          <el-icon :size="24"><Monitor /></el-icon>
          <span>GitLab Merge Alert</span>
        </div>
        <div class="header-actions">
          <el-button :icon="Refresh" circle @click="handleRefresh" />
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" :icon="UserFilled" />
              <span class="username">{{ authStore.username }}</span>
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item :icon="UserFilled" command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item :icon="Lock" command="changePassword">修改密码</el-dropdown-item>
                <el-dropdown-item v-if="authStore.isAdmin" :icon="Setting" command="settings" divided>系统设置</el-dropdown-item>
                <el-dropdown-item :icon="SwitchButton" command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>
    
    <el-container>
      <el-aside :width="isCollapse ? '64px' : '220px'" class="layout-aside">
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          class="layout-menu"
        >
          <el-menu-item index="/">
            <el-icon><DataAnalysis /></el-icon>
            <span>仪表板</span>
          </el-menu-item>
          
          <el-menu-item index="/users">
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          
          <el-menu-item index="/projects">
            <el-icon><FolderOpened /></el-icon>
            <span>项目管理</span>
          </el-menu-item>
          
          <el-menu-item index="/webhooks">
            <el-icon><Link /></el-icon>
            <span>Webhook管理</span>
          </el-menu-item>
          
          <el-menu-item v-if="authStore.isAdmin" index="/accounts">
            <el-icon><UserFilled /></el-icon>
            <span>账户管理</span>
          </el-menu-item>
        </el-menu>
        
        <div class="collapse-btn" @click="isCollapse = !isCollapse">
          <el-icon v-if="isCollapse"><Expand /></el-icon>
          <el-icon v-else><Fold /></el-icon>
        </div>
      </el-aside>
      
      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import {
  Monitor,
  DataAnalysis,
  User,
  FolderOpened,
  Link,
  Expand,
  Fold,
  Refresh,
  Setting,
  UserFilled,
  Lock,
  SwitchButton,
  ArrowDown
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const isCollapse = ref(false)

const activeMenu = computed(() => route.path)

const handleRefresh = () => {
  location.reload()
}

const handleCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'changePassword':
      router.push('/profile?tab=password')
      break
    case 'settings':
      ElMessage.info('系统设置功能开发中...')
      break
    case 'logout':
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await authStore.logout()
      router.push('/login')
      ElMessage.success('已退出登录')
      break
  }
}
</script>

<style scoped lang="less">
.layout-container {
  height: 100vh;
  background-color: #f5f7fa;
}

.layout-header {
  background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
  color: #fff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
  height: 60px;
  padding: 0;
  
  .header-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 100%;
    padding: 0 20px;
    
    .logo {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 20px;
      font-weight: 600;
      letter-spacing: 0.5px;
      text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }
    
    .header-actions {
      display: flex;
      gap: 15px;
      align-items: center;
      
      :deep(.el-button) {
        background: rgba(255, 255, 255, 0.2);
        border: none;
        color: #fff;
        
        &:hover {
          background: rgba(255, 255, 255, 0.3);
        }
      }
      
      .user-info {
        display: flex;
        align-items: center;
        gap: 8px;
        cursor: pointer;
        padding: 6px 12px;
        border-radius: 20px;
        background: rgba(255, 255, 255, 0.2);
        transition: all 0.3s;
        
        &:hover {
          background: rgba(255, 255, 255, 0.3);
        }
        
        .username {
          color: #fff;
          font-size: 14px;
          max-width: 150px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        
        .el-icon--right {
          color: #fff;
          margin-left: 4px;
        }
      }
    }
  }
}

.layout-aside {
  background: #fff;
  box-shadow: 2px 0 12px rgba(0, 0, 0, 0.06);
  position: relative;
  transition: width 0.3s;
  
  .layout-menu {
    height: calc(100% - 50px);
    border-right: none;
    padding: 8px;
    
    :deep(.el-menu-item) {
      height: 48px;
      margin-bottom: 4px;
      border-radius: 8px;
      transition: all 0.3s;
      
      &:hover {
        background-color: #f5f7fa;
      }
      
      &.is-active {
        background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
        color: #fff;
        
        .el-icon {
          color: #fff;
        }
      }
    }
    
    :deep(.el-icon) {
      font-size: 20px;
    }
  }
  
  .collapse-btn {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    border-top: 1px solid #e6e8eb;
    transition: all 0.3s;
    
    &:hover {
      background-color: #f5f7fa;
      color: #409eff;
    }
    
    .el-icon {
      font-size: 16px;
    }
  }
}

.layout-main {
  padding: 0;
  overflow: hidden;
  position: relative;
  
  :deep(.page-container) {
    height: 100%;
    overflow-y: auto;
    
    &::-webkit-scrollbar {
      width: 8px;
    }
    
    &::-webkit-scrollbar-track {
      background: transparent;
    }
    
    &::-webkit-scrollbar-thumb {
      background: #ddd;
      border-radius: 4px;
      
      &:hover {
        background: #ccc;
      }
    }
  }
}

// 页面切换动画
.fade-transform-leave-active,
.fade-transform-enter-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(30px);
}

// 响应式设计
@media screen and (max-width: 768px) {
  .layout-aside {
    position: fixed;
    left: 0;
    top: 60px;
    bottom: 0;
    z-index: 999;
    transform: translateX(-100%);
    transition: transform 0.3s;
    
    &.is-open {
      transform: translateX(0);
    }
  }
  
  .layout-main {
    margin-left: 0;
  }
}
</style>