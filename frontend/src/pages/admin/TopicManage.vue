<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NCard, NDataTable, NSpin, NButton, NIcon, NPopconfirm, NSpace, NPagination, useMessage } from 'naive-ui'
import { TrashOutline, OpenOutline } from '@vicons/ionicons5'
import type { Topic } from '@/types'
import { getTopics, deleteTopic } from '@/api/topic'

const message = useMessage()
const topics = ref<Topic[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const perPage = 20

const columns = [
  { title: 'TID', key: 'tid', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '作者', key: 'user.username', width: 100, render: (row: Topic) => row.user?.username || '' },
  { title: '阅读', key: 'viewnum', width: 70 },
  { title: '评论', key: 'cmtnum', width: 70 },
  { title: '创建时间', key: 'ctime', width: 170 },
  {
    title: '操作', key: 'actions', width: 140,
    render(row: Topic) {
      return h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, onClick: () => window.open(`/topics/${row.tid}`, '_blank') },
          { icon: () => h(NIcon, { component: OpenOutline, size: 14 }) }),
        h(NPopconfirm, { onPositiveClick: () => handleDelete(row.tid) },
          {
            trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' },
              { icon: () => h(NIcon, { component: TrashOutline, size: 14 }) }),
            default: () => '确定删除该主题？'
          }),
      ])
    }
  },
]

async function load() {
  loading.value = true
  try { const data = await getTopics({ p: page.value }); topics.value = data?.list || []; total.value = data?.total || 0 } catch {}
  loading.value = false
}

async function handleDelete(tid: number) {
  try { await deleteTopic(tid); message.success('删除成功'); load() } catch (e: any) { message.error(e.message) }
}

onMounted(load)
</script>

<template>
  <NCard title="主题管理">
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="topics" :bordered="false" :pagination="false" />
    </NSpin>
    <NPagination v-if="total > perPage" v-model:page="page" :page-count="Math.ceil(total / perPage)" style="margin-top: 16px; justify-content: flex-end" @update:page="load" />
  </NCard>
</template>
