<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NCard, NDataTable, NSpin, NButton, NIcon, NPopconfirm, NSpace, useMessage } from 'naive-ui'
import { TrashOutline, OpenOutline, CreateOutline } from '@vicons/ionicons5'
import type { Wiki } from '@/types'
import { getWikiList, deleteWiki } from '@/api/wiki'

const message = useMessage()
const wikis = ref<Wiki[]>([])
const loading = ref(true)

const columns = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: 'URI', key: 'uri', width: 120 },
  { title: '阅读', key: 'viewnum', width: 70 },
  { title: '创建时间', key: 'ctime', width: 170 },
  {
    title: '操作', key: 'actions', width: 180,
    render(row: Wiki) {
      return h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, onClick: () => window.open(`/wiki/${row.uri || row.id}`, '_blank') },
          { icon: () => h(NIcon, { component: OpenOutline, size: 14 }) }),
        h(NButton, { size: 'small', quaternary: true, onClick: () => window.open(`/wiki/edit/${row.id}`, '_blank') },
          { icon: () => h(NIcon, { component: CreateOutline, size: 14 }) }),
        h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) },
          {
            trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' },
              { icon: () => h(NIcon, { component: TrashOutline, size: 14 }) }),
            default: () => '确定删除该Wiki？'
          }),
      ])
    }
  },
]

async function load() {
  loading.value = true
  try { const data = await getWikiList({ p: 1 }); wikis.value = data?.list || [] } catch {}
  loading.value = false
}

async function handleDelete(id: number) {
  try { await deleteWiki(id); message.success('删除成功'); load() } catch (e: any) { message.error(e.message) }
}

onMounted(load)
</script>

<template>
  <NCard title="Wiki管理">
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="wikis" :pagination="{ pageSize: 20 }" :bordered="false" />
    </NSpin>
  </NCard>
</template>
