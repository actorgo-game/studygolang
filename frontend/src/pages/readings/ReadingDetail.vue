<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NText, NTag, NSpin } from 'naive-ui'
import type { Reading } from '@/types'
import { getReadingDetail } from '@/api/reading'
import { timeAgo } from '@/utils/time'
import { renderMarkdown } from '@/utils/markdown'
import CommentList from '@/components/comment/CommentList.vue'

const route = useRoute()
const reading = ref<Reading | null>(null)
const loading = ref(true)

onMounted(async () => {
  try { const data = await getReadingDetail(Number(route.params.id)); reading.value = data?.reading || null } catch {}
  loading.value = false
})
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="reading">
      <template #header><h1 style="font-size: 22px; margin: 0">每日晨读 #{{ reading.id }}</h1></template>
      <NSpace :size="12" style="margin-bottom: 16px">
        <NText>{{ reading.username }}</NText>
        <NText depth="3">{{ timeAgo(reading.ctime) }}</NText>
        <NTag size="small">{{ reading.rtype === 1 ? 'Go晨读' : '综合晨读' }}</NTag>
      </NSpace>
      <div class="markdown-body" v-html="renderMarkdown(reading.content)" />
      <a v-if="reading.url" :href="reading.url" target="_blank" rel="noopener" style="display: inline-block; margin-top: 12px; color: #18a058">查看原文 →</a>
      <CommentList :objid="reading.id" :objtype="6" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>.markdown-body { line-height: 1.8; }</style>
