<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NGrid, NGi, NStatistic, NSpin } from 'naive-ui'
import { getSiteStat } from '@/api/sidebar'
import type { SiteStat } from '@/types'

const stat = ref<SiteStat | null>(null)
const loading = ref(true)

onMounted(async () => {
  try { stat.value = await getSiteStat() } catch {}
  loading.value = false
})
</script>

<template>
  <div>
    <h2 style="margin: 0 0 24px">管理仪表盘</h2>
    <NSpin :show="loading">
      <NGrid :cols="4" :x-gap="16" :y-gap="16" v-if="stat">
        <NGi><NCard><NStatistic label="注册用户" :value="stat.user" /></NCard></NGi>
        <NGi><NCard><NStatistic label="主题数量" :value="stat.topic" /></NCard></NGi>
        <NGi><NCard><NStatistic label="文章数量" :value="stat.article" /></NCard></NGi>
        <NGi><NCard><NStatistic label="评论数量" :value="stat.comment" /></NCard></NGi>
        <NGi><NCard><NStatistic label="资源数量" :value="stat.resource" /></NCard></NGi>
        <NGi><NCard><NStatistic label="项目数量" :value="stat.project" /></NCard></NGi>
        <NGi><NCard><NStatistic label="图书数量" :value="stat.book" /></NCard></NGi>
      </NGrid>
    </NSpin>
  </div>
</template>
