<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NList, NListItem, NEllipsis } from 'naive-ui'
import type { Article } from '@/types'
import { getRecentArticles } from '@/api/sidebar'

const articles = ref<Article[]>([])

onMounted(async () => {
  try { articles.value = await getRecentArticles() } catch {}
})
</script>

<template>
  <NCard size="small" title="最新文章">
    <NList :show-divider="false" size="small">
      <NListItem v-for="a in articles" :key="a.id">
        <router-link :to="`/articles/${a.id}`" class="article-link">
          <NEllipsis>{{ a.title }}</NEllipsis>
        </router-link>
      </NListItem>
    </NList>
  </NCard>
</template>

<style scoped>
.article-link {
  text-decoration: none;
  color: inherit;
  font-size: 13px;
}
.article-link:hover {
  color: #18a058;
}
</style>
