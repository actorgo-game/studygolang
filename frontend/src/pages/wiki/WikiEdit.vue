<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NSpace, NSpin, useMessage } from 'naive-ui'
import { createWiki, modifyWiki, getWikiDetail } from '@/api/wiki'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const loading = ref(false)
const pageLoading = ref(false)
const isEdit = ref(false)
const form = ref({ id: 0, title: '', uri: '', content: '', tags: '' })

onMounted(async () => {
  const editId = route.params.id
  if (editId) {
    isEdit.value = true
    pageLoading.value = true
    try {
      const data = await getWikiDetail(String(editId))
      const wiki = data?.wiki
      if (wiki) {
        form.value = { id: wiki.id, title: wiki.title, uri: wiki.uri, content: wiki.content, tags: wiki.tags || '' }
      }
    } catch {}
    pageLoading.value = false
  }
})

async function handleSubmit() {
  if (!form.value.title.trim()) { message.warning('请输入标题'); return }
  loading.value = true
  try {
    if (isEdit.value) {
      await modifyWiki(form.value)
      message.success('修改成功')
    } else {
      await createWiki(form.value)
      message.success('创建成功')
    }
    router.push('/wiki')
  } catch (e: any) { message.error(e.message || '操作失败') } finally { loading.value = false }
}
</script>

<template>
  <NSpin :show="pageLoading">
    <NCard :title="isEdit ? '编辑Wiki' : '新建Wiki'">
      <NForm label-placement="top">
        <NFormItem label="标题" required><NInput v-model:value="form.title" placeholder="Wiki标题" /></NFormItem>
        <NFormItem label="URI"><NInput v-model:value="form.uri" placeholder="URL标识（如 my-wiki）" /></NFormItem>
        <NFormItem label="标签"><NInput v-model:value="form.tags" placeholder="多个标签用逗号分隔" /></NFormItem>
        <NFormItem label="内容" required><MarkdownEditor v-model="form.content" /></NFormItem>
        <NSpace justify="end"><NButton @click="router.back()">取消</NButton><NButton type="primary" :loading="loading" @click="handleSubmit">{{ isEdit ? '保存' : '创建' }}</NButton></NSpace>
      </NForm>
    </NCard>
  </NSpin>
</template>
