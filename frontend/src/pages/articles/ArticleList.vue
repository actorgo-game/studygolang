<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NSpace, NButton, NPagination, NEmpty, NSpin } from 'naive-ui'
import type { Article } from '@/types'
import { getArticles } from '@/api/article'
import { useUserStore } from '@/stores/user'
import ContentCard from '@/components/common/ContentCard.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const articles = ref<Article[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20
const loading = ref(true)

async function loadArticles() {
  loading.value = true
  try {
    const data = await getArticles({ p: page.value })
    articles.value = data?.list || []
    total.value = data?.total || 0
  } catch {}
  loading.value = false
}

watch(() => route.query, () => {
  page.value = Number(route.query.p) || 1
  loadArticles()
})

onMounted(loadArticles)
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <h2 style="margin: 0">文章列表</h2>
      <NButton v-if="userStore.isLoggedIn" type="primary" @click="router.push('/articles/new')">
        发布文章
      </NButton>
    </NSpace>

    <NSpin :show="loading">
      <NEmpty v-if="!loading && !articles.length" description="暂无文章" />
      <div v-else class="article-list">
        <ContentCard
          v-for="a in articles"
          :key="a.id"
          :title="a.title"
          :url="`/articles/${a.id}`"
          :author="a.author_txt || a.author"
          :author-url="a.user ? `/user/${a.user.username}` : undefined"
          :avatar="a.user?.avatar"
          :time="a.ctime"
          :tags="a.tags"
          :viewnum="a.viewnum"
          :cmtnum="a.cmtnum"
          :likenum="a.likenum"
          :summary="a.txt?.substring(0, 150)"
          :cover="a.cover"
        />
      </div>
    </NSpin>

    <NPagination
      v-if="total > perPage"
      v-model:page="page"
      :page-count="Math.ceil(total / perPage)"
      style="margin-top: 16px; justify-content: center"
      @update:page="(p: number) => router.push({ query: { ...route.query, p: String(p) } })"
    />
  </div>
</template>

<style scoped>
.article-list { display: flex; flex-direction: column; gap: 12px; }
</style>
