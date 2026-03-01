<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { NMenu, NButton, NAvatar, NDropdown, NBadge, NInput, NSpace, NIcon } from 'naive-ui'
import { SearchOutline, NotificationsOutline } from '@vicons/ionicons5'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const { handleLogout } = useAuth()

const searchQuery = ref('')

const menuOptions = [
  { label: '主题', key: '/topics' },
  { label: '文章', key: '/articles' },
  { label: '项目', key: '/projects' },
  { label: '资源', key: '/resources' },
  { label: '图书', key: '/books' },
  { label: '下载', key: 'https://go.dev/dl/' },
]

const userMenuOptions = computed(() => {
  const items: any[] = [
    { label: '我的主页', key: 'profile' },
    { label: '个人设置', key: 'settings' },
    { label: '我的收藏', key: 'favorites' },
    { label: '我的余额', key: 'balance' },
    { label: '消息中心', key: 'messages' },
  ]
  if (userStore.isAdmin) {
    items.push({ type: 'divider', key: 'd0' })
    items.push({ label: '管理后台', key: 'admin' })
  }
  items.push({ type: 'divider', key: 'd1' })
  items.push({ label: '退出登录', key: 'logout' })
  return items
})

function onMenuSelect(key: string) {
  if (key.startsWith('http')) {
    window.open(key, '_blank')
  } else {
    router.push(key)
  }
}

function onSearch() {
  if (searchQuery.value.trim()) {
    router.push({ path: '/search', query: { q: searchQuery.value } })
  }
}

async function onUserMenuSelect(key: string) {
  const username = userStore.me?.username
  switch (key) {
    case 'profile':
      router.push(`/user/${username}`)
      break
    case 'settings':
      router.push('/account/edit')
      break
    case 'favorites':
      router.push(`/favorites/${username}`)
      break
    case 'balance':
      router.push('/balance')
      break
    case 'messages':
      router.push('/message/system')
      break
    case 'admin':
      router.push('/admin')
      break
    case 'logout':
      await handleLogout()
      break
  }
}
</script>

<template>
  <header class="app-header">
    <div class="header-inner">
      <router-link to="/" class="logo">
        <strong>Go语言中文网</strong>
      </router-link>

      <NMenu
        mode="horizontal"
        :options="menuOptions"
        :on-update:value="onMenuSelect"
        class="nav-menu"
      />

      <div class="header-right">
        <NInput
          v-model:value="searchQuery"
          placeholder="搜索..."
          size="small"
          round
          clearable
          @keyup.enter="onSearch"
          style="width: 200px"
        >
          <template #prefix>
            <NIcon :component="SearchOutline" />
          </template>
        </NInput>

        <template v-if="userStore.isLoggedIn">
          <router-link to="/message/system">
            <NBadge :value="userStore.me?.msgnum || 0" :max="99">
              <NIcon :component="NotificationsOutline" size="22" style="cursor: pointer" />
            </NBadge>
          </router-link>

          <NDropdown :options="userMenuOptions" @select="onUserMenuSelect" trigger="click">
            <NSpace align="center" style="cursor: pointer" :size="8">
              <NAvatar
                :src="userStore.me?.avatar"
                :size="32"
                round
                :fallback-src="`https://www.gravatar.com/avatar/?d=identicon`"
              />
              <span>{{ userStore.me?.username }}</span>
            </NSpace>
          </NDropdown>
        </template>

        <template v-else>
          <NSpace>
            <NButton size="small" @click="appStore.openLoginModal()">登录</NButton>
            <NButton size="small" type="primary" @click="router.push('/account/register')">注册</NButton>
          </NSpace>
        </template>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  position: sticky;
  top: 0;
  z-index: 100;
}
.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  align-items: center;
  height: 56px;
  gap: 16px;
}
.logo {
  text-decoration: none;
  color: #18a058;
  font-size: 18px;
  white-space: nowrap;
}
.nav-menu {
  flex: 1;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
  white-space: nowrap;
}
</style>
