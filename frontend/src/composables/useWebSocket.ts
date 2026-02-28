import { ref, onMounted, onUnmounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

export function useWebSocket() {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null

  function connect() {
    const userStore = useUserStore()
    const appStore = useAppStore()

    if (!userStore.me?.uid) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws?uid=${userStore.me.uid}`

    ws.value = new WebSocket(url)

    ws.value.onopen = () => {
      connected.value = true
      heartbeatTimer = setInterval(() => {
        ws.value?.send('ping')
      }, 15000)
    }

    ws.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 0) {
          userStore.updateMsgNum(data.body?.msgnum ?? 0)
        } else if (data.type === 1) {
          appStore.setOnlineUsers(data.body?.online ?? 0, data.body?.maxonline ?? 0)
        }
      } catch {
        // ignore non-JSON messages
      }
    }

    ws.value.onclose = () => {
      connected.value = false
      if (heartbeatTimer) clearInterval(heartbeatTimer)
      setTimeout(connect, 5000)
    }

    ws.value.onerror = () => {
      ws.value?.close()
    }
  }

  function disconnect() {
    if (heartbeatTimer) clearInterval(heartbeatTimer)
    ws.value?.close()
    ws.value = null
  }

  onMounted(connect)
  onUnmounted(disconnect)

  return { connected, disconnect }
}
