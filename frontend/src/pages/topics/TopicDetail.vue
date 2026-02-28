<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NText, NAvatar, NTag, NButton, NIcon, NSpin, useMessage } from 'naive-ui'
import { HeartOutline, BookmarkOutline, ShareSocialOutline } from '@vicons/ionicons5'
import type { Topic } from '@/types'
import { getTopicDetail } from '@/api/topic'
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

const topic = ref<Topic | null>(null)
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const tid = Number(route.params.tid)
    const data = await getTopicDetail(tid)
    topic.value = data?.topic || null
  } catch {}
  loading.value = false
}

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!topic.value) return
  try {
    await toggleLike(topic.value.tid, { objtype: 0 })
    topic.value.likenum++
    message.success('点赞成功')
  } catch (e: any) { message.error(e.message) }
}

async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!topic.value) return
  try {
    await toggleFavorite(topic.value.tid, { objtype: 0 })
    message.success('收藏成功')
  } catch (e: any) { message.error(e.message) }
}

onMounted(load)
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="topic">
      <template #header>
        <h1 style="font-size: 22px; margin: 0">{{ topic.title }}</h1>
      </template>
      <NSpace align="center" :size="12" style="margin-bottom: 16px">
        <router-link v-if="topic.user" :to="`/user/${topic.user.username}`">
          <NSpace align="center" :size="8">
            <NAvatar :src="topic.user.avatar" :size="28" round />
            <NText>{{ topic.user.username }}</NText>
          </NSpace>
        </router-link>
        <NText depth="3">{{ timeAgo(topic.ctime) }}</NText>
        <NText depth="3">{{ topic.viewnum }} 阅读</NText>
        <NTag v-if="topic.node?.name" size="small" round>{{ topic.node?.name }}</NTag>
      </NSpace>

      <div class="markdown-body" v-html="renderMarkdown(topic.content)" />

      <NSpace style="margin-top: 24px" :size="12">
        <NButton @click="handleLike" quaternary>
          <template #icon><NIcon :component="HeartOutline" /></template>
          {{ topic.likenum }} 赞
        </NButton>
        <NButton @click="handleFavorite" quaternary>
          <template #icon><NIcon :component="BookmarkOutline" /></template>
          收藏
        </NButton>
      </NSpace>

      <CommentList :objid="topic.tid" :objtype="0" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>
.markdown-body { line-height: 1.8; }
.markdown-body :deep(pre) { background: #2d2d2d; padding: 16px; border-radius: 4px; overflow-x: auto; }
.markdown-body :deep(code) { font-family: 'Fira Code', monospace; }
.markdown-body :deep(img) { max-width: 100%; }
.markdown-body :deep(blockquote) { border-left: 4px solid #18a058; padding-left: 16px; color: #666; }
a { text-decoration: none; color: inherit; }
</style>
