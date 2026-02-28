<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NList, NListItem, NEllipsis } from 'naive-ui'
import type { Topic } from '@/types'
import { getRecentTopics } from '@/api/sidebar'

const topics = ref<Topic[]>([])

onMounted(async () => {
  try { topics.value = await getRecentTopics() } catch {}
})
</script>

<template>
  <NCard size="small" title="最新主题">
    <NList :show-divider="false" size="small">
      <NListItem v-for="t in topics" :key="t.tid">
        <router-link :to="`/topics/${t.tid}`" class="topic-link">
          <NEllipsis>{{ t.title }}</NEllipsis>
        </router-link>
      </NListItem>
    </NList>
  </NCard>
</template>

<style scoped>
.topic-link {
  text-decoration: none;
  color: inherit;
  font-size: 13px;
}
.topic-link:hover {
  color: #18a058;
}
</style>
