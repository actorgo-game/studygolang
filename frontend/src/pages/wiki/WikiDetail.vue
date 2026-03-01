<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NSpace, NText, NButton, NIcon, NSpin, NPopconfirm, useMessage } from 'naive-ui'
import { HeartOutline, Heart, BookmarkOutline, Bookmark, TrashOutline, CreateOutline } from '@vicons/ionicons5'
import type { Wiki } from '@/types'
import { getWikiDetail, deleteWiki } from '@/api/wiki'
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
const wiki = ref<Wiki | null>(null)
const wikiUser = ref<any>(null)
const loading = ref(true)
const liked = ref(false)
const favorited = ref(false)

const isOwner = computed(() => userStore.me && wiki.value && (wiki.value.uid === userStore.me.uid || userStore.isAdmin))

onMounted(async () => {
  try {
    const data = await getWikiDetail(String(route.params.uri))
    wiki.value = data?.wiki || null
    wikiUser.value = (data as any)?.wiki_user || null
    if (userStore.isLoggedIn && wiki.value) {
      hadLike(wiki.value.id, 3).then(res => { liked.value = res?.liked ?? false }).catch(() => {})
      hadFavorite(wiki.value.id, 3).then(res => { favorited.value = res?.favorited ?? false }).catch(() => {})
    }
  } catch {}
  loading.value = false
})

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!wiki.value) return
  try {
    const res = await toggleLike(wiki.value.id, { objtype: 3 })
    liked.value = res?.liked ?? false
    message.success(liked.value ? '点赞成功' : '已取消点赞')
  } catch (e: any) { message.error(e.message) }
}

async function handleFavorite() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!wiki.value) return
  try {
    const res = await toggleFavorite(wiki.value.id, { objtype: 3 })
    favorited.value = res?.favorited ?? false
    message.success(favorited.value ? '收藏成功' : '已取消收藏')
  } catch (e: any) { message.error(e.message) }
}

async function handleDelete() {
  if (!wiki.value) return
  try {
    await deleteWiki(wiki.value.id)
    message.success('删除成功')
    router.push('/wiki')
  } catch (e: any) { message.error(e.message) }
}
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="wiki">
      <template #header><h1 style="font-size: 22px; margin: 0">{{ wiki.title }}</h1></template>
      <NSpace :size="12" style="margin-bottom: 16px">
        <router-link v-if="wikiUser" :to="`/user/${wikiUser.username}`" style="text-decoration: none; color: inherit">
          <NText>{{ wikiUser.username }}</NText>
        </router-link>
        <NText depth="3">{{ timeAgo(wiki.ctime) }}</NText>
        <NText depth="3">{{ wiki.viewnum }} 阅读</NText>
      </NSpace>
      <div class="markdown-body" v-html="renderMarkdown(wiki.content)" />
      <NSpace style="margin-top: 24px" :size="12">
        <NButton @click="handleLike" quaternary :type="liked ? 'error' : 'default'">
          <template #icon><NIcon :component="liked ? Heart : HeartOutline" /></template>
          赞
        </NButton>
        <NButton @click="handleFavorite" quaternary :type="favorited ? 'warning' : 'default'">
          <template #icon><NIcon :component="favorited ? Bookmark : BookmarkOutline" /></template>
          {{ favorited ? '已收藏' : '收藏' }}
        </NButton>
        <NButton v-if="isOwner" quaternary @click="router.push(`/wiki/edit/${wiki.id}`)">
          <template #icon><NIcon :component="CreateOutline" /></template>
          编辑
        </NButton>
        <NPopconfirm v-if="isOwner" @positive-click="handleDelete">
          <template #trigger>
            <NButton quaternary type="error">
              <template #icon><NIcon :component="TrashOutline" /></template>
              删除
            </NButton>
          </template>
          确定要删除这个Wiki吗？
        </NPopconfirm>
      </NSpace>
      <CommentList :objid="wiki.id" :objtype="3" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>.markdown-body { line-height: 1.8; } .markdown-body :deep(pre) { background: #2d2d2d; padding: 16px; border-radius: 4px; overflow-x: auto; } .markdown-body :deep(img) { max-width: 100%; }</style>
