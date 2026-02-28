<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NDataTable, NSpin } from 'naive-ui'
import type { Topic } from '@/types'
import { getTopics } from '@/api/topic'

const topics = ref<Topic[]>([])
const loading = ref(true)

const columns = [
  { title: 'TID', key: 'tid', width: 80 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '作者', key: 'user.username', width: 120, render: (row: Topic) => row.user?.username || '' },
  { title: '阅读', key: 'viewnum', width: 80 },
  { title: '评论', key: 'cmtnum', width: 80 },
  { title: '创建时间', key: 'ctime', width: 180 },
]

onMounted(async () => {
  try { const data = await getTopics({ p: 1 }); topics.value = data?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="主题管理">
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="topics" :pagination="{ pageSize: 20 }" :bordered="false" />
    </NSpin>
  </NCard>
</template>
