<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NList, NListItem, NSpace, NText, NTag, NEmpty, NSpin } from 'naive-ui'
import type { Reading } from '@/types'
import { getReadings } from '@/api/reading'
import { timeAgo } from '@/utils/time'

const readings = ref<Reading[]>([])
const loading = ref(true)

onMounted(async () => {
  try { const data = await getReadings({ p: 1 }); readings.value = data?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <div>
    <h2 style="margin-bottom: 16px">每日晨读</h2>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !readings.length" description="暂无晨读" />
      <NList v-else :show-divider="true" hoverable>
        <NListItem v-for="r in readings" :key="r.id">
          <router-link :to="`/readings/${r.id}`" style="text-decoration: none; color: inherit; display: block">
            <div style="font-size: 15px; font-weight: 500; margin-bottom: 4px">{{ r.content?.substring(0, 100) }}</div>
            <NSpace :size="12"><NText depth="3">{{ r.username }}</NText><NText depth="3">{{ timeAgo(r.ctime) }}</NText><NTag size="tiny">{{ r.rtype === 1 ? 'Go晨读' : '综合晨读' }}</NTag></NSpace>
          </router-link>
        </NListItem>
      </NList>
    </NSpin>
  </div>
</template>
