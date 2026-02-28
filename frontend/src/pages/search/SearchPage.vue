<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NInput, NIcon, NList, NListItem, NText, NEmpty, NSpin, NSpace } from 'naive-ui'
import { SearchOutline } from '@vicons/ionicons5'
import { search } from '@/api/interact'

const route = useRoute()
const query = ref(String(route.query.q || ''))
const results = ref<any[]>([])
const loading = ref(false)

async function doSearch() {
  if (!query.value.trim()) return
  loading.value = true
  try { const data = await search({ q: query.value }); results.value = data?.list || data || [] } catch {}
  loading.value = false
}

function handleSearch() {
  window.history.replaceState(null, '', `/search?q=${encodeURIComponent(query.value)}`)
  doSearch()
}

watch(() => route.query.q, (v) => { if (v) { query.value = String(v); doSearch() } })
onMounted(() => { if (query.value) doSearch() })
</script>

<template>
  <NCard title="搜索">
    <NInput v-model:value="query" placeholder="搜索..." round clearable @keyup.enter="handleSearch" size="large" style="margin-bottom: 16px">
      <template #prefix><NIcon :component="SearchOutline" /></template>
    </NInput>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !results.length && query" description="未找到结果" />
      <NList v-else-if="results.length" :show-divider="true">
        <NListItem v-for="(item, i) in results" :key="i">
          <router-link :to="item.url || '#'" style="text-decoration: none; color: inherit">
            <div style="font-weight: 500">{{ item.title || item.name }}</div>
            <NText depth="3" style="font-size: 13px">{{ item.desc?.substring(0, 200) || item.content?.substring(0, 200) }}</NText>
          </router-link>
        </NListItem>
      </NList>
    </NSpin>
  </NCard>
</template>
