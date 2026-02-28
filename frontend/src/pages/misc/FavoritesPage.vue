<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NList, NListItem, NText, NEmpty, NSpin } from 'naive-ui'
import { getFavorites } from '@/api/interact'
import { timeAgo } from '@/utils/time'

const route = useRoute()
const favorites = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try { const data = await getFavorites(String(route.params.username), { p: 1 }); favorites.value = data?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="我的收藏">
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !favorites.length" description="暂无收藏" />
      <NList v-else :show-divider="true">
        <NListItem v-for="(f, i) in favorites" :key="i">
          <router-link :to="f.url || '#'" style="text-decoration: none; color: inherit">
            <div style="font-weight: 500">{{ f.title }}</div>
            <NText depth="3" style="font-size: 12px">{{ timeAgo(f.ctime) }}</NText>
          </router-link>
        </NListItem>
      </NList>
    </NSpin>
  </NCard>
</template>
