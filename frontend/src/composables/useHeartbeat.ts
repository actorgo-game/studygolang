import { ref, onMounted, onUnmounted } from 'vue'
import { get } from '@/api/request'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

export function useHeartbeat(interval = 30000) {
  const active = ref(false)
  let timer: ReturnType<typeof setInterval> | null = null

  async function poll() {
    const userStore = useUserStore()
    const appStore = useAppStore()

    try {
      const data = await get<{ online: number; maxonline: number; msgnum: number }>('/user/heartbeat')
      appStore.setOnlineUsers(data.online ?? 0, data.maxonline ?? 0)
      if (userStore.me) {
        userStore.updateMsgNum(data.msgnum ?? 0)
      }
    } catch {
      // silently ignore heartbeat failures
    }
  }

  function start() {
    if (active.value) return
    active.value = true
    poll()
    timer = setInterval(poll, interval)
  }

  function stop() {
    active.value = false
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  onMounted(start)
  onUnmounted(stop)

  return { active, stop }
}
