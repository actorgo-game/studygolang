<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NText, NButton, NIcon, NTag, NSpin, useMessage } from 'naive-ui'
import { HeartOutline, BookmarkOutline, LogoGithub, LinkOutline } from '@vicons/ionicons5'
import type { Project } from '@/types'
import { getProjectDetail } from '@/api/project'
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
const project = ref<Project | null>(null)
const loading = ref(true)

onMounted(async () => {
  try { const data = await getProjectDetail(String(route.params.uri)); project.value = data?.project || null } catch {}
  loading.value = false
})

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!project.value) return
  try { await toggleLike(project.value.id, { objtype: 4 }); project.value.likenum++; message.success('点赞成功') } catch (e: any) { message.error(e.message) }
}
async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!project.value) return
  try { await toggleFavorite(project.value.id, { objtype: 4 }); message.success('收藏成功') } catch (e: any) { message.error(e.message) }
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
        <NButton @click="handleLike" quaternary><template #icon><NIcon :component="HeartOutline" /></template>{{ project.likenum }} 赞</NButton>
        <NButton @click="handleFavorite" quaternary><template #icon><NIcon :component="BookmarkOutline" /></template>收藏</NButton>
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
