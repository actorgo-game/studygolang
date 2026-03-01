<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NCard, NDataTable, NSpin, NButton, NIcon, NSpace, NPagination, useMessage } from 'naive-ui'
import { OpenOutline } from '@vicons/ionicons5'
import type { User } from '@/types'
import { getUsers } from '@/api/user'

const message = useMessage()
const users = ref<User[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const perPage = 20

const columns = [
  { title: 'UID', key: 'uid', width: 70 },
  { title: '用户名', key: 'username', width: 120 },
  { title: '邮箱', key: 'email', width: 200 },
  { title: '角色', key: 'role_name', width: 100 },
  { title: '状态', key: 'status', width: 70 },
  { title: '注册时间', key: 'ctime', width: 170 },
  {
    title: '操作', key: 'actions', width: 80,
    render(row: User) {
      return h(NButton, { size: 'small', quaternary: true, onClick: () => window.open(`/user/${row.username}`, '_blank') },
        { icon: () => h(NIcon, { component: OpenOutline, size: 14 }) })
    }
  },
]

async function load() {
  loading.value = true
  try { const data = await getUsers({ p: page.value }); users.value = (data as any)?.list || []; total.value = (data as any)?.total || 0 } catch {}
  loading.value = false
}

onMounted(load)
</script>

<template>
  <NCard title="用户管理">
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="users" :bordered="false" :pagination="false" />
    </NSpin>
    <NPagination v-if="total > perPage" v-model:page="page" :page-count="Math.ceil(total / perPage)" style="margin-top: 16px; justify-content: flex-end" @update:page="load" />
  </NCard>
</template>
