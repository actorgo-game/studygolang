<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NSpace, NText, NButton, NImage, NIcon, NSpin, NPopconfirm, NTag, useMessage } from 'naive-ui'
import { HeartOutline, Heart, CartOutline, CreateOutline, TrashOutline } from '@vicons/ionicons5'
import type { Book } from '@/types'
import { getBookDetail, deleteBook } from '@/api/book'
import { toggleLike, hadLike } from '@/api/interact'
import { renderMarkdown } from '@/utils/markdown'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import CommentList from '@/components/comment/CommentList.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const message = useMessage()
const book = ref<Book | null>(null)
const loading = ref(true)
const liked = ref(false)
const isOwner = computed(() => userStore.me && book.value && (book.value.uid === userStore.me.uid || userStore.isAdmin))

onMounted(async () => {
  try {
    const data = await getBookDetail(Number(route.params.id))
    book.value = data?.book || null
    if (userStore.isLoggedIn && book.value) {
      hadLike(book.value.id, 5).then(res => { liked.value = res?.liked ?? false }).catch(() => {})
    }
  } catch {}
  loading.value = false
})

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!book.value) return
  try {
    const res = await toggleLike(book.value.id, { objtype: 5 })
    const nowLiked = res?.liked ?? false
    book.value.likenum += nowLiked ? 1 : -1
    liked.value = nowLiked
    message.success(nowLiked ? '推荐成功' : '已取消推荐')
  } catch (e: any) { message.error(e.message) }
}

async function handleDelete() {
  if (!book.value) return
  try {
    await deleteBook(book.value.id)
    message.success('删除成功')
    router.push('/books')
  } catch (e: any) { message.error(e.message) }
}
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="book">
      <div style="display: flex; gap: 24px; margin-bottom: 24px">
        <NImage v-if="book.cover" :src="book.cover" width="150" style="border-radius: 8px; flex-shrink: 0" />
        <div>
          <h1 style="font-size: 22px; margin: 0 0 8px 0">{{ book.name }}</h1>
          <NSpace vertical :size="4">
            <NText>作者：{{ book.author }}</NText>
            <NText v-if="book.translator">译者：{{ book.translator }}</NText>
            <NText v-if="book.pub_date">出版日期：{{ book.pub_date }}</NText>
            <NText v-if="book.price">价格：¥{{ book.price }}</NText>
            <NSpace :size="8" style="margin-top: 8px">
              <a v-if="book.buy_url" :href="book.buy_url" target="_blank" rel="noopener"><NButton type="primary" size="small"><template #icon><NIcon :component="CartOutline" /></template>购买</NButton></a>
              <a v-if="book.online_url" :href="book.online_url" target="_blank" rel="noopener"><NButton size="small">在线阅读</NButton></a>
              <a v-if="book.download_url" :href="book.download_url" target="_blank" rel="noopener"><NButton size="small">下载</NButton></a>
              <NTag v-if="book.is_free" type="success" size="small">免费</NTag>
            </NSpace>
          </NSpace>
        </div>
      </div>
      <div class="markdown-body" v-html="renderMarkdown(book.desc)" />
      <template v-if="book.catalogue">
        <h3 style="margin-top: 24px">目录</h3>
        <div class="markdown-body" v-html="renderMarkdown(book.catalogue)" />
      </template>
      <NSpace style="margin-top: 24px" :size="12">
        <NButton @click="handleLike" quaternary :type="liked ? 'error' : 'default'"><template #icon><NIcon :component="liked ? Heart : HeartOutline" /></template>{{ book.likenum }} 推荐</NButton>
        <NButton v-if="isOwner" quaternary @click="router.push(`/book/edit/${book.id}`)"><template #icon><NIcon :component="CreateOutline" /></template>编辑</NButton>
        <NPopconfirm v-if="isOwner" @positive-click="handleDelete">
          <template #trigger><NButton quaternary type="error"><template #icon><NIcon :component="TrashOutline" /></template>删除</NButton></template>
          确定要删除这本图书吗？
        </NPopconfirm>
      </NSpace>
      <CommentList :objid="book.id" :objtype="5" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>
.markdown-body { line-height: 1.8; }
a { text-decoration: none; color: inherit; }
</style>
