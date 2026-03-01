<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NSpace, NText, NButton, NIcon, NSpin, NPopconfirm, useMessage } from 'naive-ui'
import { HeartOutline, Heart, BookmarkOutline, Bookmark, LinkOutline, TrashOutline } from '@vicons/ionicons5'
import type { Resource } from '@/types'
import { getResourceDetail, deleteResource } from '@/api/resource'
import { toggleLike, hadLike, toggleFavorite, hadFavorite } from '@/api/interact'
import { timeAgo } from '@/utils/time'
import { renderMarkdown } from '@/utils/markdown'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import CommentList from '@/components/comment/CommentList.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const message = useMessage()
const isOwner = computed(() => userStore.me && resource.value && (resource.value.uid === userStore.me.uid || userStore.isAdmin))
const resource = ref<Resource | null>(null)
const loading = ref(true)
const liked = ref(false)
const favorited = ref(false)

onMounted(async () => {
  try {
    const data = await getResourceDetail(Number(route.params.id))
    resource.value = data?.resource || null
    if (userStore.isLoggedIn && resource.value) {
      hadLike(resource.value.id, 2).then(res => { liked.value = res?.liked ?? false }).catch(() => {})
      hadFavorite(resource.value.id, 2).then(res => { favorited.value = res?.favorited ?? false }).catch(() => {})
    }
  } catch {}
  loading.value = false
})

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!resource.value) return
  try {
    const res = await toggleLike(resource.value.id, { objtype: 2 })
    const nowLiked = res?.liked ?? false
    resource.value.likenum += nowLiked ? 1 : -1
    liked.value = nowLiked
    message.success(nowLiked ? '点赞成功' : '已取消点赞')
  } catch (e: any) { message.error(e.message) }
}
async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!resource.value) return
  try {
    const res = await toggleFavorite(resource.value.id, { objtype: 2 })
    favorited.value = res?.favorited ?? false
    message.success(favorited.value ? '收藏成功' : '已取消收藏')
  } catch (e: any) { message.error(e.message) }
}
async function handleDelete() {
  if (!resource.value) return
  try {
    await deleteResource(resource.value.id)
    message.success('删除成功')
    router.push('/resources')
  } catch (e: any) { message.error(e.message) }
}
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="resource">
      <template #header><h1 style="font-size: 22px; margin: 0">{{ resource.title }}</h1></template>
      <NSpace align="center" :size="12" style="margin-bottom: 16px">
        <NText>{{ resource.user?.username }}</NText>
        <NText depth="3">{{ timeAgo(resource.ctime) }}</NText>
        <NText depth="3">{{ resource.viewnum }} 阅读</NText>
        <a v-if="resource.url" :href="resource.url" target="_blank" rel="noopener">
          <NButton size="small" quaternary><template #icon><NIcon :component="LinkOutline" /></template>访问链接</NButton>
        </a>
      </NSpace>
      <div class="markdown-body" v-html="renderMarkdown(resource.content)" />
      <NSpace style="margin-top: 24px" :size="12">
        <NButton @click="handleLike" quaternary :type="liked ? 'error' : 'default'"><template #icon><NIcon :component="liked ? Heart : HeartOutline" /></template>{{ resource.likenum }} 赞</NButton>
        <NButton @click="handleFavorite" quaternary :type="favorited ? 'warning' : 'default'"><template #icon><NIcon :component="favorited ? Bookmark : BookmarkOutline" /></template>{{ favorited ? '已收藏' : '收藏' }}</NButton>
        <NPopconfirm v-if="isOwner" @positive-click="handleDelete"><template #trigger><NButton quaternary type="error"><template #icon><NIcon :component="TrashOutline" /></template>删除</NButton></template>确定要删除这个资源吗？</NPopconfirm>
      </NSpace>
      <CommentList :objid="resource.id" :objtype="2" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>
.markdown-body { line-height: 1.8; }
.markdown-body :deep(pre) { background: #2d2d2d; padding: 16px; border-radius: 4px; overflow-x: auto; }
.markdown-body :deep(img) { max-width: 100%; }
a { text-decoration: none; color: inherit; }
</style>
