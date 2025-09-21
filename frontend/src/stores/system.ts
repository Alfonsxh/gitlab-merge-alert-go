import { defineStore } from 'pinia'
import { ref } from 'vue'
import { systemAPI } from '@/api/system'
import type { SetupAdminPayload } from '@/api/system'

export const useSystemStore = defineStore('system', () => {
  const adminSetupRequired = ref<boolean | null>(null)
  const loading = ref(false)

  const checkAdminSetup = async (force = false): Promise<boolean> => {
    if (!force && adminSetupRequired.value !== null) {
      return adminSetupRequired.value
    }

    loading.value = true
    try {
      const { admin_setup_required } = await systemAPI.getBootstrapStatus()
      adminSetupRequired.value = admin_setup_required
      return admin_setup_required
    } catch (error) {
      console.error('Failed to fetch bootstrap status', error)
      adminSetupRequired.value = false
      return false
    } finally {
      loading.value = false
    }
  }

  const setupAdmin = async (payload: SetupAdminPayload): Promise<void> => {
    await systemAPI.setupAdmin(payload)
    adminSetupRequired.value = false
  }

  const markAdminSetupRequired = () => {
    adminSetupRequired.value = true
  }

  return {
    adminSetupRequired,
    loading,
    checkAdminSetup,
    setupAdmin,
    markAdminSetupRequired
  }
})
