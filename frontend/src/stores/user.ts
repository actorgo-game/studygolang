import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Me } from '@/types'
import { getCurrentUser, login as apiLogin, logout as apiLogout } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const me = ref<Me | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!me.value && me.value.uid > 0)
  const isAdmin = computed(() => !!me.value && me.value.is_root)

  async function fetchCurrentUser() {
    loading.value = true
    try {
      me.value = await getCurrentUser()
    } catch {
      me.value = null
    } finally {
      loading.value = false
    }
  }

  async function login(username: string, passwd: string, rememberMe = false) {
    await apiLogin({ username, passwd, remember_me: rememberMe ? '1' : '0' })
    await fetchCurrentUser()
  }

  async function logout() {
    await apiLogout()
    me.value = null
  }

  function updateMsgNum(num: number) {
    if (me.value) {
      me.value.msgnum = num
    }
  }

  return { me, loading, isLoggedIn, isAdmin, fetchCurrentUser, login, logout, updateMsgNum }
})
