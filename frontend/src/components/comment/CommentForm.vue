<script setup lang="ts">
import { ref } from 'vue'
import { NInput, NButton, NSpace, useMessage } from 'naive-ui'
import { createComment } from '@/api/comment'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  objid: number
  objtype: number
}>()

const emit = defineEmits<{
  created: []
}>()

const userStore = useUserStore()
const appStore = useAppStore()
const message = useMessage()

const content = ref('')
const loading = ref(false)

async function handleSubmit() {
  if (!userStore.isLoggedIn) {
    appStore.openLoginModal()
    return
  }
  if (!content.value.trim()) {
    message.warning('请输入评论内容')
    return
  }
  loading.value = true
  try {
    await createComment(props.objid, { objtype: props.objtype, content: content.value })
    message.success('评论成功')
    content.value = ''
    emit('created')
  } catch (e: any) {
    message.error(e.message || '评论失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="comment-form">
    <NInput
      v-model:value="content"
      type="textarea"
      placeholder="写下你的评论...（支持 Markdown）"
      :rows="4"
    />
    <NSpace justify="end" style="margin-top: 8px">
      <NButton type="primary" :loading="loading" @click="handleSubmit">发表评论</NButton>
    </NSpace>
  </div>
</template>
