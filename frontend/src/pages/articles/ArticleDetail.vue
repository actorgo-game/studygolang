<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NText, NAvatar, NTag, NButton, NIcon, NSpin, useMessage } from 'naive-ui'
import { HeartOutline, BookmarkOutline } from '@vicons/ionicons5'
import type { Article } from '@/types'
import { getArticleDetail } from '@/api/article'
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

const article = ref<Article | null>(null)
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const data = await getArticleDetail(id)
    article.value = data?.article || null
  } catch {}
  loading.value = false
}

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!article.value) return
  try {
    await toggleLike(article.value.id, { objtype: 1 })
    article.value.likenum++
    message.success('点赞成功')
  } catch (e: any) { message.error(e.message) }
}

async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!article.value) return
  try {
    await toggleFavorite(article.value.id, { objtype: 1 })
    message.success('收藏成功')
  } catch (e: any) { message.error(e.message) }
}

onMounted(load)
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="article">
      <template #header>
        <h1 style="font-size: 22px; margin: 0">{{ article.title }}</h1>
      </template>
      <NSpace align="center" :size="12" style="margin-bottom: 16px">
        <NText>{{ article.author_txt || article.author }}</NText>
        <NText depth="3">{{ timeAgo(article.pub_date || article.ctime) }}</NText>
        <NText depth="3">{{ article.viewnum }} 阅读</NText>
      </NSpace>

      <div class="markdown-body" v-html="renderMarkdown(article.content)" />

      <NSpace style="margin-top: 24px" :size="12">
        <NButton @click="handleLike" quaternary>
          <template #icon><NIcon :component="HeartOutline" /></template>
          {{ article.likenum }} 赞
        </NButton>
        <NButton @click="handleFavorite" quaternary>
          <template #icon><NIcon :component="BookmarkOutline" /></template>
          收藏
        </NButton>
      </NSpace>

      <CommentList :objid="article.id" :objtype="1" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>
.markdown-body { line-height: 1.8; }
.markdown-body :deep(pre) { background: #2d2d2d; padding: 16px; border-radius: 4px; overflow-x: auto; }
.markdown-body :deep(img) { max-width: 100%; }
a { text-decoration: none; color: inherit; }
</style>
