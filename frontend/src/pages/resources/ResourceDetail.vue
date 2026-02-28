<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NText, NButton, NIcon, NSpin, useMessage } from 'naive-ui'
import { HeartOutline, BookmarkOutline, LinkOutline } from '@vicons/ionicons5'
import type { Resource } from '@/types'
import { getResourceDetail } from '@/api/resource'
import { toggleLike, toggleFavorite } from '@/api/interact'
import { timeAgo } from '@/utils/time'
import { renderMarkdown } from '@/utils/markdown'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import CommentList from '@/components/comment/CommentList.vue'

const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()
const message = useMessage()
const resource = ref<Resource | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const data = await getResourceDetail(Number(route.params.id))
    resource.value = data?.resource || null
  } catch {}
  loading.value = false
})

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!resource.value) return
  try { await toggleLike(resource.value.id, { objtype: 2 }); resource.value.likenum++; message.success('点赞成功') } catch (e: any) { message.error(e.message) }
}
async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!resource.value) return
  try { await toggleFavorite(resource.value.id, { objtype: 2 }); message.success('收藏成功') } catch (e: any) { message.error(e.message) }
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
        <NButton @click="handleLike" quaternary><template #icon><NIcon :component="HeartOutline" /></template>{{ resource.likenum }} 赞</NButton>
        <NButton @click="handleFavorite" quaternary><template #icon><NIcon :component="BookmarkOutline" /></template>收藏</NButton>
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
