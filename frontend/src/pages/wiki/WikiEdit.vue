<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NSpace, useMessage } from 'naive-ui'
import { createWiki } from '@/api/wiki'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const form = ref({ title: '', uri: '', content: '', tags: '' })

async function handleSubmit() {
  if (!form.value.title.trim()) { message.warning('请输入标题'); return }
  loading.value = true
  try { await createWiki(form.value); message.success('创建成功'); router.push('/wiki') } catch (e: any) { message.error(e.message || '创建失败') } finally { loading.value = false }
}
</script>

<template>
  <NCard title="新建Wiki">
    <NForm label-placement="top">
      <NFormItem label="标题" required><NInput v-model:value="form.title" placeholder="Wiki标题" /></NFormItem>
      <NFormItem label="URI"><NInput v-model:value="form.uri" placeholder="URL标识（如 my-wiki）" /></NFormItem>
      <NFormItem label="标签"><NInput v-model:value="form.tags" placeholder="多个标签用逗号分隔" /></NFormItem>
      <NFormItem label="内容" required><MarkdownEditor v-model="form.content" /></NFormItem>
      <NSpace justify="end"><NButton @click="router.back()">取消</NButton><NButton type="primary" :loading="loading" @click="handleSubmit">创建</NButton></NSpace>
    </NForm>
  </NCard>
</template>
