<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NSpace, useMessage } from 'naive-ui'
import { createProject } from '@/api/project'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const form = ref({ name: '', category: '', uri: '', home: '', src: '', doc: '', desc: '', logo: '', tags: '', author: '' })

async function handleSubmit() {
  if (!form.value.name.trim()) { message.warning('请输入项目名称'); return }
  loading.value = true
  try { await createProject(form.value); message.success('发布成功'); router.push('/projects') } catch (e: any) { message.error(e.message || '发布失败') } finally { loading.value = false }
}
</script>

<template>
  <NCard title="发布项目">
    <NForm label-placement="top">
      <NFormItem label="项目名称" required><NInput v-model:value="form.name" placeholder="项目名称" /></NFormItem>
      <NFormItem label="分类"><NInput v-model:value="form.category" placeholder="分类" /></NFormItem>
      <NFormItem label="URI"><NInput v-model:value="form.uri" placeholder="项目URI（用于URL）" /></NFormItem>
      <NFormItem label="主页"><NInput v-model:value="form.home" placeholder="项目主页URL" /></NFormItem>
      <NFormItem label="源码"><NInput v-model:value="form.src" placeholder="源码仓库URL" /></NFormItem>
      <NFormItem label="标签"><NInput v-model:value="form.tags" placeholder="多个标签用逗号分隔" /></NFormItem>
      <NFormItem label="Logo URL"><NInput v-model:value="form.logo" placeholder="Logo图片URL" /></NFormItem>
      <NFormItem label="项目描述"><MarkdownEditor v-model="form.desc" /></NFormItem>
      <NSpace justify="end"><NButton @click="router.back()">取消</NButton><NButton type="primary" :loading="loading" @click="handleSubmit">发布</NButton></NSpace>
    </NForm>
  </NCard>
</template>
