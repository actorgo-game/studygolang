import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import { useRouter } from 'vue-router'

export function useAuth() {
  const userStore = useUserStore()
  const appStore = useAppStore()
  const router = useRouter()

  function requireLogin(callback?: () => void) {
    if (!userStore.isLoggedIn) {
      appStore.openLoginModal()
      return false
    }
    callback?.()
    return true
  }

  async function handleLogout() {
    await userStore.logout()
    router.push('/')
  }

  return { requireLogin, handleLogout }
}
