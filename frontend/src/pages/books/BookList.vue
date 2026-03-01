<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NSpace, NText, NButton, NIcon, NPagination, NEmpty, NSpin, NTag, NDivider } from 'naive-ui'
import { HeartOutline, ChatbubbleOutline, EyeOutline } from '@vicons/ionicons5'
import type { Book } from '@/types'
import { getBooks } from '@/api/book'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const books = ref<Book[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20
const loading = ref(true)

async function load() {
  loading.value = true
  try { const data = await getBooks({ p: page.value }); books.value = data?.list || []; total.value = data?.total || 0 } catch {}
  loading.value = false
}
watch(() => route.query, () => { page.value = Number(route.query.p) || 1; load() })
onMounted(load)
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <h2 style="margin: 0">图书推荐</h2>
      <NButton type="primary" @click="userStore.isLoggedIn ? router.push('/book/new') : appStore.openLoginModal()">推荐图书</NButton>
    </NSpace>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !books.length" description="暂无图书" />
      <div v-else class="book-list">
        <NCard v-for="b in books" :key="b.id" size="small" hoverable class="book-item">
          <div class="book-row">
            <img v-if="b.cover" :src="b.cover" class="book-cover" />
            <div class="book-info">
              <router-link :to="`/book/${b.id}`" class="book-title">{{ b.name }}</router-link>
              <div class="book-meta">
                <NText depth="3" style="font-size: 13px">
                  <span v-if="b.author">[作] {{ b.author }}</span>
                  <span v-if="b.translator"> &nbsp; [译] {{ b.translator }}</span>
                </NText>
                <NText v-if="b.pub_date" depth="3" style="font-size: 12px; margin-left: 12px">{{ b.pub_date }}</NText>
              </div>
              <p class="book-desc" v-if="b.desc">{{ b.desc.substring(0, 150) }}<router-link v-if="b.desc.length > 150" :to="`/book/${b.id}`" class="more-link">[...]</router-link></p>
              <NSpace :size="16" class="book-stats">
                <NSpace :size="4" align="center">
                  <NIcon :component="HeartOutline" size="14" /><NText depth="3" style="font-size: 12px">{{ b.likenum || 0 }}推荐</NText>
                </NSpace>
                <NSpace :size="4" align="center">
                  <NIcon :component="ChatbubbleOutline" size="14" /><NText depth="3" style="font-size: 12px">{{ b.cmtnum || 0 }}评论</NText>
                </NSpace>
                <NSpace :size="4" align="center">
                  <NIcon :component="EyeOutline" size="14" /><NText depth="3" style="font-size: 12px">{{ b.viewnum || 0 }} 阅读</NText>
                </NSpace>
              </NSpace>
            </div>
          </div>
        </NCard>
      </div>
    </NSpin>
    <NPagination v-if="total > perPage" v-model:page="page" :page-count="Math.ceil(total / perPage)" style="margin-top: 16px; justify-content: center" @update:page="(p: number) => router.push({ query: { ...route.query, p: String(p) } })" />
  </div>
</template>

<style scoped>
.book-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.book-row {
  display: flex;
  gap: 16px;
}
.book-cover {
  width: 90px;
  height: 120px;
  object-fit: cover;
  border-radius: 4px;
  flex-shrink: 0;
}
.book-info {
  flex: 1;
  min-width: 0;
}
.book-title {
  font-size: 16px;
  font-weight: 600;
  text-decoration: none;
  color: inherit;
  display: block;
  margin-bottom: 4px;
}
.book-title:hover {
  color: #18a058;
}
.book-meta {
  margin-bottom: 6px;
}
.book-desc {
  color: #666;
  font-size: 13px;
  line-height: 1.6;
  margin: 4px 0 8px;
}
.more-link {
  color: #18a058;
  text-decoration: none;
  font-size: 13px;
}
.book-stats {
  margin-top: 4px;
}
.book-item a {
  text-decoration: none;
  color: inherit;
}
</style>
