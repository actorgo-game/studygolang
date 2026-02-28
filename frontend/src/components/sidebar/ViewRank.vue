<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NTabs, NTabPane, NList, NListItem, NEllipsis } from 'naive-ui'
import { getViewRank } from '@/api/sidebar'

const todayRank = ref<any[]>([])
const weekRank = ref<any[]>([])

onMounted(async () => {
  try {
    todayRank.value = await getViewRank({ objtype: 0, rank_type: 'today', limit: 10 })
  } catch {}
  try {
    weekRank.value = await getViewRank({ objtype: 0, rank_type: 'week', limit: 10 })
  } catch {}
})
</script>

<template>
  <NCard size="small" title="阅读排行">
    <NTabs type="line" size="small">
      <NTabPane name="today" tab="今日">
        <NList :show-divider="false" size="small">
          <NListItem v-for="(item, i) in todayRank" :key="i">
            <router-link :to="item.url || '#'" class="rank-link">
              <NEllipsis>{{ i + 1 }}. {{ item.title }}</NEllipsis>
            </router-link>
          </NListItem>
        </NList>
      </NTabPane>
      <NTabPane name="week" tab="本周">
        <NList :show-divider="false" size="small">
          <NListItem v-for="(item, i) in weekRank" :key="i">
            <router-link :to="item.url || '#'" class="rank-link">
              <NEllipsis>{{ i + 1 }}. {{ item.title }}</NEllipsis>
            </router-link>
          </NListItem>
        </NList>
      </NTabPane>
    </NTabs>
  </NCard>
</template>

<style scoped>
.rank-link {
  text-decoration: none;
  color: inherit;
  font-size: 13px;
}
.rank-link:hover {
  color: #18a058;
}
</style>
