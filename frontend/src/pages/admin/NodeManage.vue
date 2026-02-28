<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NDataTable, NSpin } from 'naive-ui'
import type { TopicNode } from '@/types'
import { getNodes } from '@/api/topic'

const nodes = ref<TopicNode[]>([])
const loading = ref(true)

const columns = [
  { title: 'NID', key: 'nid', width: 80 },
  { title: '名称', key: 'name', width: 200 },
  { title: '简介', key: 'intro', ellipsis: { tooltip: true } },
  { title: '排序', key: 'seq', width: 80 },
  { title: '创建时间', key: 'ctime', width: 180 },
]

onMounted(async () => {
  try { nodes.value = await getNodes() || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="节点管理">
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="nodes" :pagination="{ pageSize: 20 }" :bordered="false" />
    </NSpin>
  </NCard>
</template>
