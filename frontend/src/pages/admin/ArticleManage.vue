<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NDataTable, NSpin } from 'naive-ui'
import type { Article } from '@/types'
import { getArticles } from '@/api/article'

const articles = ref<Article[]>([])
const loading = ref(true)

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '作者', key: 'author_txt', width: 120 },
  { title: '阅读', key: 'viewnum', width: 80 },
  { title: '评论', key: 'cmtnum', width: 80 },
  { title: '状态', key: 'status', width: 80 },
  { title: '创建时间', key: 'ctime', width: 180 },
]

onMounted(async () => {
  try { const data = await getArticles({ p: 1 }); articles.value = data?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="文章管理">
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="articles" :pagination="{ pageSize: 20 }" :bordered="false" />
    </NSpin>
  </NCard>
</template>
