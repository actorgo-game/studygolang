<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NDataTable, NSpin, NInput, NSpace, NButton, useMessage } from 'naive-ui'
import type { User } from '@/types'
import { getUsers } from '@/api/user'

const message = useMessage()
const users = ref<User[]>([])
const loading = ref(true)
const searchQuery = ref('')

const columns = [
  { title: 'UID', key: 'uid', width: 80 },
  { title: '用户名', key: 'username', width: 120 },
  { title: '邮箱', key: 'email', width: 200 },
  { title: '角色', key: 'role_name', width: 100 },
  { title: '状态', key: 'status', width: 80 },
  { title: '注册时间', key: 'ctime', width: 180 },
]

onMounted(async () => {
  try { const data = await getUsers({ p: 1 }); users.value = (data as any)?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="用户管理">
    <NSpace style="margin-bottom: 16px">
      <NInput v-model:value="searchQuery" placeholder="搜索用户..." style="width: 300px" />
    </NSpace>
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="users" :pagination="{ pageSize: 20 }" :bordered="false" />
    </NSpin>
  </NCard>
</template>
