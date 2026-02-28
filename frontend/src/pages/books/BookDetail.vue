<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NText, NButton, NImage, NIcon, NSpin, useMessage } from 'naive-ui'
import { HeartOutline, BookmarkOutline, CartOutline } from '@vicons/ionicons5'
import type { Book } from '@/types'
import { getBookDetail } from '@/api/book'
import { toggleLike, toggleFavorite } from '@/api/interact'
import { renderMarkdown } from '@/utils/markdown'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import CommentList from '@/components/comment/CommentList.vue'

const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()
const message = useMessage()
const book = ref<Book | null>(null)
const loading = ref(true)

onMounted(async () => {
  try { const data = await getBookDetail(Number(route.params.id)); book.value = data?.book || null } catch {}
  loading.value = false
})

async function handleLike() {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  if (!book.value) return
  try { await toggleLike(book.value.id, { objtype: 5 }); book.value.likenum++; message.success('推荐成功') } catch (e: any) { message.error(e.message) }
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
            </NSpace>
          </NSpace>
        </div>
      </div>
      <div class="markdown-body" v-html="renderMarkdown(book.desc)" />
      <NSpace style="margin-top: 24px" :size="12">
        <NButton @click="handleLike" quaternary><template #icon><NIcon :component="HeartOutline" /></template>{{ book.likenum }} 推荐</NButton>
      </NSpace>
      <CommentList :objid="book.id" :objtype="5" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>
.markdown-body { line-height: 1.8; }
a { text-decoration: none; color: inherit; }
</style>
