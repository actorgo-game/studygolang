<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NSpace, useMessage } from 'naive-ui'
import { createArticle } from '@/api/article'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const form = ref({ title: '', content: '', url: '', tags: '', author_txt: '' })

async function handleSubmit() {
  if (!form.value.title.trim()) { message.warning('请输入标题'); return }
  loading.value = true
  try {
    await createArticle(form.value)
    message.success('发布成功')
    router.push('/articles')
  } catch (e: any) {
    message.error(e.message || '发布失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <NCard title="发布文章">
    <NForm label-placement="top">
      <NFormItem label="标题" required>
        <NInput v-model:value="form.title" placeholder="请输入文章标题" />
      </NFormItem>
      <NFormItem label="原文链接">
        <NInput v-model:value="form.url" placeholder="文章原文URL（可选）" />
      </NFormItem>
      <NFormItem label="标签">
        <NInput v-model:value="form.tags" placeholder="多个标签用逗号分隔" />
      </NFormItem>
      <NFormItem label="内容" required>
        <MarkdownEditor v-model="form.content" />
      </NFormItem>
      <NSpace justify="end">
        <NButton @click="router.back()">取消</NButton>
        <NButton type="primary" :loading="loading" @click="handleSubmit">发布</NButton>
      </NSpace>
    </NForm>
  </NCard>
</template>
