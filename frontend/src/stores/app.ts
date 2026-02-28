import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const showLoginModal = ref(false)
  const onlineUsers = ref(0)
  const maxOnlineUsers = ref(0)
  const sidebarCollapsed = ref(false)

  function openLoginModal() {
    showLoginModal.value = true
  }

  function closeLoginModal() {
    showLoginModal.value = false
  }

  function setOnlineUsers(online: number, maxOnline: number) {
    onlineUsers.value = online
    maxOnlineUsers.value = maxOnline
  }

  return {
    showLoginModal,
    onlineUsers,
    maxOnlineUsers,
    sidebarCollapsed,
    openLoginModal,
    closeLoginModal,
    setOnlineUsers,
  }
})
