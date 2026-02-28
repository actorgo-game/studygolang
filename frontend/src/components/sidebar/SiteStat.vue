<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NStatistic, NGrid, NGi } from 'naive-ui'
import type { SiteStat } from '@/types'
import { getSiteStat } from '@/api/sidebar'

const stat = ref<SiteStat | null>(null)

onMounted(async () => {
  try { stat.value = await getSiteStat() } catch {}
})
</script>

<template>
  <NCard size="small" title="社区统计">
    <NGrid :cols="2" :x-gap="8" :y-gap="8" v-if="stat">
      <NGi><NStatistic label="会员" :value="stat.user" /></NGi>
      <NGi><NStatistic label="主题" :value="stat.topic" /></NGi>
      <NGi><NStatistic label="文章" :value="stat.article" /></NGi>
      <NGi><NStatistic label="回复" :value="stat.comment" /></NGi>
    </NGrid>
  </NCard>
</template>
