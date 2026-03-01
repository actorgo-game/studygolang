<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NSpace, NText, NButton, NIcon, NTag, NSpin, NPopconfirm, useMessage } from 'naive-ui'
import { HeartOutline, Heart, BookmarkOutline, Bookmark, LogoGithub, LinkOutline, TrashOutline } from '@vicons/ionicons5'
import type { Project } from '@/types'
import { getProjectDetail, deleteProject } from '@/api/project'
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
const isOwner = computed(() => userStore.me && project.value && (project.value.username === userStore.me.username || userStore.isAdmin))
const project = ref<Project | null>(null)
const loading = ref(true)
const liked = ref(false)
const favorited = ref(false)

onMounted(async () => {
  try {
    const data = await getProjectDetail(String(route.params.uri))
    project.value = data?.project || null
    if (userStore.isLoggedIn && project.value) {
      hadLike(project.value.id, 4).then(res => { liked.value = res?.liked ?? false }).catch(() => {})
      hadFavorite(project.value.id, 4).then(res => { favorited.value = res?.favorited ?? false }).catch(() => {})
    }
  } catch {}
  loading.value = false
})

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!project.value) return
  try {
    const res = await toggleLike(project.value.id, { objtype: 4 })
    const nowLiked = res?.liked ?? false
    project.value.likenum += nowLiked ? 1 : -1
    liked.value = nowLiked
    message.success(nowLiked ? '点赞成功' : '已取消点赞')
  } catch (e: any) { message.error(e.message) }
}
async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!project.value) return
  try {
    const res = await toggleFavorite(project.value.id, { objtype: 4 })
    favorited.value = res?.favorited ?? false
    message.success(favorited.value ? '收藏成功' : '已取消收藏')
  } catch (e: any) { message.error(e.message) }
}
async function handleDelete() {
  if (!project.value) return
  try {
    await deleteProject(project.value.id)
    message.success('删除成功')
    router.push('/projects')
  } catch (e: any) { message.error(e.message) }
}
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="project">
      <template #header>
        <NSpace align="center" :size="12">
          <img v-if="project.logo" :src="project.logo" style="width: 48px; height: 48px; border-radius: 8px" />
          <div><h1 style="font-size: 22px; margin: 0">{{ project.name }}</h1><NText depth="3">{{ project.category }}</NText></div>
        </NSpace>
      </template>
      <NSpace :size="12" style="margin-bottom: 16px">
        <NText depth="3">{{ project.author || project.username }}</NText>
        <NText depth="3">{{ timeAgo(project.ctime) }}</NText>
        <NText depth="3">{{ project.viewnum }} 阅读</NText>
        <a v-if="project.home" :href="project.home" target="_blank" rel="noopener"><NButton size="small" quaternary><template #icon><NIcon :component="LinkOutline" /></template>主页</NButton></a>
        <a v-if="project.src" :href="project.src" target="_blank" rel="noopener"><NButton size="small" quaternary><template #icon><NIcon :component="LogoGithub" /></template>源码</NButton></a>
      </NSpace>
      <NSpace v-if="project.tags" :size="4" style="margin-bottom: 12px"><NTag v-for="tag in project.tags.split(',')" :key="tag" size="small" round>{{ tag.trim() }}</NTag></NSpace>
      <div class="markdown-body" v-html="renderMarkdown(project.desc)" />
      <NSpace style="margin-top: 24px" :size="12">
        <NButton @click="handleLike" quaternary :type="liked ? 'error' : 'default'"><template #icon><NIcon :component="liked ? Heart : HeartOutline" /></template>{{ project.likenum }} 赞</NButton>
        <NButton @click="handleFavorite" quaternary :type="favorited ? 'warning' : 'default'"><template #icon><NIcon :component="favorited ? Bookmark : BookmarkOutline" /></template>{{ favorited ? '已收藏' : '收藏' }}</NButton>
        <NPopconfirm v-if="isOwner" @positive-click="handleDelete"><template #trigger><NButton quaternary type="error"><template #icon><NIcon :component="TrashOutline" /></template>删除</NButton></template>确定要删除这个项目吗？</NPopconfirm>
      </NSpace>
      <CommentList :objid="project.id" :objtype="4" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>
.markdown-body { line-height: 1.8; }
.markdown-body :deep(pre) { background: #2d2d2d; padding: 16px; border-radius: 4px; overflow-x: auto; }
.markdown-body :deep(img) { max-width: 100%; }
a { text-decoration: none; color: inherit; }
</style>
