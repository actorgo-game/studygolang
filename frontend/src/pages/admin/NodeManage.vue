<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NCard, NDataTable, NSpin, NButton, NIcon, NPopconfirm, NSpace, NModal, NForm, NFormItem, NInput, NInputNumber, NSwitch, useMessage } from 'naive-ui'
import { CreateOutline, TrashOutline, AddOutline } from '@vicons/ionicons5'
import type { TopicNode } from '@/types'
import { getNodes, modifyNode, deleteNode } from '@/api/topic'

const message = useMessage()
const nodes = ref<TopicNode[]>([])
const loading = ref(true)
const showModal = ref(false)
const saving = ref(false)
const editForm = ref({ nid: 0, name: '', ename: '', intro: '', seq: 0, parent: 0, logo: '', show_index: false })

const columns = [
  { title: 'NID', key: 'nid', width: 80 },
  { title: '名称', key: 'name', width: 160 },
  { title: '英文名', key: 'ename', width: 120 },
  { title: '简介', key: 'intro', ellipsis: { tooltip: true } },
  { title: '排序', key: 'seq', width: 80 },
  { title: '首页显示', key: 'show_index', width: 90, render(row: TopicNode) { return row.show_index ? '是' : '否' } },
  { title: '创建时间', key: 'ctime', width: 170 },
  {
    title: '操作', key: 'actions', width: 140,
    render(row: TopicNode) {
      return h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, onClick: () => openEdit(row) },
          { icon: () => h(NIcon, { component: CreateOutline, size: 14 }) }),
        h(NPopconfirm, { onPositiveClick: () => handleDelete(row.nid) },
          {
            trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' },
              { icon: () => h(NIcon, { component: TrashOutline, size: 14 }) }),
            default: () => '确定删除该节点？'
          }),
      ])
    }
  },
]

function openEdit(node?: TopicNode) {
  if (node) {
    editForm.value = { nid: node.nid, name: node.name, ename: node.ename || '', intro: node.intro || '', seq: node.seq || 0, parent: (node as any).parent || 0, logo: (node as any).logo || '', show_index: !!node.show_index }
  } else {
    editForm.value = { nid: 0, name: '', ename: '', intro: '', seq: 0, parent: 0, logo: '', show_index: false }
  }
  showModal.value = true
}

async function handleSave() {
  if (!editForm.value.name.trim()) { message.warning('请输入节点名称'); return }
  saving.value = true
  try {
    await modifyNode({
      nid: editForm.value.nid,
      name: editForm.value.name,
      ename: editForm.value.ename,
      intro: editForm.value.intro,
      seq: editForm.value.seq,
      parent: editForm.value.parent,
      logo: editForm.value.logo,
      show_index: editForm.value.show_index ? '1' : '0',
    })
    message.success(editForm.value.nid ? '修改成功' : '创建成功')
    showModal.value = false
    load()
  } catch (e: any) { message.error(e.message || '操作失败') }
  saving.value = false
}

async function handleDelete(nid: number) {
  try { await deleteNode(nid); message.success('删除成功'); load() } catch (e: any) { message.error(e.message) }
}

async function load() {
  loading.value = true
  try { nodes.value = await getNodes() || [] } catch {}
  loading.value = false
}

onMounted(load)
</script>

<template>
  <NCard title="节点管理">
    <template #header-extra>
      <NButton type="primary" size="small" @click="openEdit()">
        <template #icon><NIcon :component="AddOutline" /></template>
        新建节点
      </NButton>
    </template>
    <NSpin :show="loading">
      <NDataTable :columns="columns" :data="nodes" :pagination="{ pageSize: 20 }" :bordered="false" />
    </NSpin>
  </NCard>

  <NModal v-model:show="showModal" preset="dialog" :title="editForm.nid ? '编辑节点' : '新建节点'" style="width: 520px" :mask-closable="false">
    <NForm label-placement="left" label-width="80" style="margin-top: 16px">
      <NFormItem label="名称">
        <NInput v-model:value="editForm.name" placeholder="节点名称" />
      </NFormItem>
      <NFormItem label="英文名">
        <NInput v-model:value="editForm.ename" placeholder="英文标识" />
      </NFormItem>
      <NFormItem label="简介">
        <NInput v-model:value="editForm.intro" type="textarea" placeholder="节点简介" :rows="3" />
      </NFormItem>
      <NFormItem label="Logo">
        <NInput v-model:value="editForm.logo" placeholder="Logo URL" />
      </NFormItem>
      <NFormItem label="父节点ID">
        <NInputNumber v-model:value="editForm.parent" :min="0" placeholder="0 表示顶级节点" style="width: 100%" />
      </NFormItem>
      <NFormItem label="排序">
        <NInputNumber v-model:value="editForm.seq" :min="0" style="width: 100%" />
      </NFormItem>
      <NFormItem label="首页显示">
        <NSwitch v-model:value="editForm.show_index" />
      </NFormItem>
      <NSpace justify="end">
        <NButton @click="showModal = false">取消</NButton>
        <NButton type="primary" :loading="saving" @click="handleSave">保存</NButton>
      </NSpace>
    </NForm>
  </NModal>
</template>
