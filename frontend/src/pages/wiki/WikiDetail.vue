<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpace, NText, NSpin } from 'naive-ui'
import type { Wiki } from '@/types'
import { getWikiDetail } from '@/api/wiki'
import { timeAgo } from '@/utils/time'
import { renderMarkdown } from '@/utils/markdown'
import CommentList from '@/components/comment/CommentList.vue'

const route = useRoute()
const wiki = ref<Wiki | null>(null)
const loading = ref(true)

onMounted(async () => {
  try { const data = await getWikiDetail(String(route.params.uri)); wiki.value = data?.wiki || null } catch {}
  loading.value = false
})
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="wiki">
      <template #header><h1 style="font-size: 22px; margin: 0">{{ wiki.title }}</h1></template>
      <NSpace :size="12" style="margin-bottom: 16px">
        <NText>{{ wiki.user?.username }}</NText>
        <NText depth="3">{{ timeAgo(wiki.ctime) }}</NText>
        <NText depth="3">{{ wiki.viewnum }} 阅读</NText>
      </NSpace>
      <div class="markdown-body" v-html="renderMarkdown(wiki.content)" />
      <CommentList :objid="wiki.id" :objtype="3" style="margin-top: 24px" />
    </NCard>
  </NSpin>
</template>

<style scoped>.markdown-body { line-height: 1.8; } .markdown-body :deep(pre) { background: #2d2d2d; padding: 16px; border-radius: 4px; overflow-x: auto; } .markdown-body :deep(img) { max-width: 100%; }</style>
