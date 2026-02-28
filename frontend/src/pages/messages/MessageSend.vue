<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NInputNumber, NButton, NSpace, useMessage } from 'naive-ui'
import { sendMessage } from '@/api/message'

const router = useRouter()
const message = useMessage()
const form = ref({ to: 0, content: '' })
const loading = ref(false)

async function handleSend() {
  if (!form.value.to || !form.value.content.trim()) { message.warning('请填写完整信息'); return }
  loading.value = true
  try { await sendMessage(form.value); message.success('发送成功'); router.push('/message/outbox') } catch (e: any) { message.error(e.message || '发送失败') } finally { loading.value = false }
}
</script>

<template>
  <NCard title="发送消息">
    <NForm label-placement="top">
      <NFormItem label="收件人UID"><NInputNumber v-model:value="form.to" placeholder="收件人用户ID" style="width: 100%" /></NFormItem>
      <NFormItem label="内容"><NInput v-model:value="form.content" type="textarea" :rows="4" placeholder="消息内容" /></NFormItem>
      <NSpace justify="end"><NButton @click="router.back()">取消</NButton><NButton type="primary" :loading="loading" @click="handleSend">发送</NButton></NSpace>
    </NForm>
  </NCard>
</template>
