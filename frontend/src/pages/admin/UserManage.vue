<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NCard, NDataTable, NSpin, NButton, NIcon, NSpace, NPagination, NPopconfirm, NTag, NModal, NForm, NFormItem, NSelect, useMessage } from 'naive-ui'
import { OpenOutline, CreateOutline, TrashOutline } from '@vicons/ionicons5'
import type { User } from '@/types'
import { getUsers, adminChangeUserStatus, adminDeleteUser } from '@/api/user'

const message = useMessage()
const users = ref<User[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const perPage = 20

const showEditModal = ref(false)
const editUid = ref(0)
const editUsername = ref('')
const editStatus = ref(0)
const saving = ref(false)

const statusOptions = [
  { label: '正常', value: 0 },
  { label: '冻结', value: 1 },
  { label: '停号', value: 2 },
]

function statusTag(status: number) {
  if (status === 0) return h(NTag, { size: 'small', type: 'success' }, () => '正常')
  if (status === 1) return h(NTag, { size: 'small', type: 'warning' }, () => '冻结')
  if (status === 2) return h(NTag, { size: 'small', type: 'error' }, () => '停号')
  return h(NTag, { size: 'small' }, () => String(status))
}

const columns = [
  { title: 'UID', key: 'uid', width: 70 },
  { title: '用户名', key: 'username', width: 120 },
  { title: '邮箱', key: 'email', width: 200 },
  { title: '角色', key: 'role_name', width: 100 },
  { title: '状态', key: 'status', width: 80, render(row: User) { return statusTag((row as any).status ?? 0) } },
  { title: '注册时间', key: 'ctime', width: 170 },
  {
    title: '操作', key: 'actions', width: 160,
    render(row: User) {
      return h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, onClick: () => window.open(`/user/${row.username}`, '_blank') },
          { icon: () => h(NIcon, { component: OpenOutline, size: 14 }) }),
        h(NButton, { size: 'small', quaternary: true, onClick: () => openEdit(row) },
          { icon: () => h(NIcon, { component: CreateOutline, size: 14 }) }),
        h(NPopconfirm, { onPositiveClick: () => handleDelete(row.uid) },
          {
            trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' },
              { icon: () => h(NIcon, { component: TrashOutline, size: 14 }) }),
            default: () => `确定删除用户 ${row.username} 的所有内容？此操作不可恢复！`
          }),
      ])
    }
  },
]

function openEdit(user: User) {
  editUid.value = user.uid
  editUsername.value = user.username
  editStatus.value = (user as any).status ?? 0
  showEditModal.value = true
}

async function handleSaveStatus() {
  saving.value = true
  try {
    await adminChangeUserStatus(editUid.value, editStatus.value)
    message.success('状态修改成功')
    showEditModal.value = false
    load()
  } catch (e: any) { message.error(e.message || '操作失败') }
  saving.value = false
}

async function handleDelete(uid: number) {
  try {
    await adminDeleteUser(uid)
    message.success('已删除该用户的所有内容')
    load()
  } catch (e: any) { message.error(e.message) }
}

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

  <NModal v-model:show="showEditModal" preset="dialog" title="编辑用户状态" style="width: 420px" :mask-closable="false">
    <NForm label-placement="left" label-width="80" style="margin-top: 16px">
      <NFormItem label="用户">
        <span>{{ editUsername }} (UID: {{ editUid }})</span>
      </NFormItem>
      <NFormItem label="状态">
        <NSelect v-model:value="editStatus" :options="statusOptions" />
      </NFormItem>
      <NSpace justify="end">
        <NButton @click="showEditModal = false">取消</NButton>
        <NButton type="primary" :loading="saving" @click="handleSaveStatus">保存</NButton>
      </NSpace>
    </NForm>
  </NModal>
</template>
