<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NSelect, NButton, NSpace, useMessage } from 'naive-ui'
import type { TopicNode } from '@/types'
import { getNodes, createTopic, modifyTopic } from '@/api/topic'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const isEdit = ref(false)
const loading = ref(false)
const nodes = ref<TopicNode[]>([])
const form = ref({ tid: 0, title: '', content: '', nid: 0, tags: '' })

const nodeOptions = ref<{ label: string; value: number }[]>([])

onMounted(async () => {
  try {
    nodes.value = await getNodes()
    nodeOptions.value = nodes.value.map(n => ({ label: n.name, value: n.nid }))
  } catch {}
  if (route.query.tid) {
    isEdit.value = true
    form.value.tid = Number(route.query.tid)
  }
})

async function handleSubmit() {
  if (!form.value.title.trim()) { message.warning('请输入标题'); return }
  if (!form.value.nid) { message.warning('请选择节点'); return }
  if (!form.value.content.trim()) { message.warning('请输入内容'); return }

  loading.value = true
  try {
    if (isEdit.value) {
      await modifyTopic(form.value)
      message.success('修改成功')
      router.push(`/topics/${form.value.tid}`)
    } else {
      const data = await createTopic(form.value) as any
      message.success('发布成功')
      router.push(`/topics/${data?.tid || ''}`)
    }
  } catch (e: any) {
    message.error(e.message || '操作失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <NCard :title="isEdit ? '编辑主题' : '发布新主题'">
    <NForm label-placement="top">
      <NFormItem label="标题" required>
        <NInput v-model:value="form.title" placeholder="请输入标题" />
      </NFormItem>
      <NFormItem label="节点" required>
        <NSelect v-model:value="form.nid" :options="nodeOptions" placeholder="请选择节点" />
      </NFormItem>
      <NFormItem label="标签">
        <NInput v-model:value="form.tags" placeholder="多个标签用逗号分隔" />
      </NFormItem>
      <NFormItem label="内容" required>
        <MarkdownEditor v-model="form.content" />
      </NFormItem>
      <NSpace justify="end">
        <NButton @click="router.back()">取消</NButton>
        <NButton type="primary" :loading="loading" @click="handleSubmit">
          {{ isEdit ? '保存修改' : '发布' }}
        </NButton>
      </NSpace>
    </NForm>
  </NCard>
</template>
