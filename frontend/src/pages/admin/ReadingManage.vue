<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NDataTable, NSpin } from 'naive-ui'
import type { Reading } from '@/types'
import { getReadings } from '@/api/reading'

const readings = ref<Reading[]>([])
const loading = ref(true)

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '内容', key: 'content', ellipsis: { tooltip: true } },
  { title: '类型', key: 'rtype', width: 80 },
  { title: '提交者', key: 'username', width: 120 },
  { title: '创建时间', key: 'ctime', width: 180 },
]

onMounted(async () => {
  try { const data = await getReadings({ p: 1 }); readings.value = data?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="晨读管理">
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="readings" :pagination="{ pageSize: 20 }" :bordered="false" />
    </NSpin>
  </NCard>
</template>
