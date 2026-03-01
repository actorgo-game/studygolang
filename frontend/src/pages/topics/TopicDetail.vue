<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NSpace, NText, NAvatar, NTag, NButton, NIcon, NSpin, NPopconfirm, useMessage } from 'naive-ui'
import { HeartOutline, Heart, BookmarkOutline, Bookmark, TrashOutline } from '@vicons/ionicons5'
import type { Topic } from '@/types'
import { getTopicDetail, deleteTopic } from '@/api/topic'
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
const isOwner = computed(() => userStore.me && topic.value && (topic.value.uid === userStore.me.uid || userStore.isAdmin))

const topic = ref<Topic | null>(null)
const loading = ref(true)
const liked = ref(false)
const favorited = ref(false)

async function load() {
  loading.value = true
  try {
    const tid = Number(route.params.tid)
    const data = await getTopicDetail(tid)
    topic.value = data?.topic || null
    if (userStore.isLoggedIn && topic.value) {
      hadLike(topic.value.tid, 0).then(res => { liked.value = res?.liked ?? false }).catch(() => {})
      hadFavorite(topic.value.tid, 0).then(res => { favorited.value = res?.favorited ?? false }).catch(() => {})
    }
  } catch {}
  loading.value = false
}

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!topic.value) return
  try {
    const res = await toggleLike(topic.value.tid, { objtype: 0 })
    const nowLiked = res?.liked ?? false
    topic.value.likenum += nowLiked ? 1 : -1
    liked.value = nowLiked
    message.success(nowLiked ? '点赞成功' : '已取消点赞')
  } catch (e: any) { message.error(e.message) }
}

async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!topic.value) return
  try {
    const res = await toggleFavorite(topic.value.tid, { objtype: 0 })
    favorited.value = res?.favorited ?? false
    message.success(favorited.value ? '收藏成功' : '已取消收藏')
  } catch (e: any) { message.error(e.message) }
}

async function handleDelete() {
  if (!topic.value) return
  try {
    await deleteTopic(topic.value.tid)
    message.success('删除成功')
    router.push('/topics')
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
        <NButton @click="handleLike" quaternary :type="liked ? 'error' : 'default'">
          <template #icon><NIcon :component="liked ? Heart : HeartOutline" /></template>
          {{ topic.likenum }} 赞
        </NButton>
        <NButton @click="handleFavorite" quaternary :type="favorited ? 'warning' : 'default'">
          <template #icon><NIcon :component="favorited ? Bookmark : BookmarkOutline" /></template>
          {{ favorited ? '已收藏' : '收藏' }}
        </NButton>
        <NPopconfirm v-if="isOwner" @positive-click="handleDelete">
          <template #trigger>
            <NButton quaternary type="error">
              <template #icon><NIcon :component="TrashOutline" /></template>
              删除
            </NButton>
          </template>
          确定要删除这个主题吗？
        </NPopconfirm>
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
