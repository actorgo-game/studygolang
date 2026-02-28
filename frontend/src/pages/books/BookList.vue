<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NGrid, NGi, NSpace, NText, NTag, NButton, NPagination, NEmpty, NSpin, NImage } from 'naive-ui'
import type { Book } from '@/types'
import { getBooks } from '@/api/book'
import { timeAgo } from '@/utils/time'

const route = useRoute()
const router = useRouter()
const books = ref<Book[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20
const loading = ref(true)

async function load() {
  loading.value = true
  try { const data = await getBooks({ p: page.value }); books.value = data?.list || []; total.value = data?.total || 0 } catch {}
  loading.value = false
}
watch(() => route.query, () => { page.value = Number(route.query.p) || 1; load() })
onMounted(load)
</script>

<template>
  <div>
    <h2 style="margin-bottom: 16px">图书推荐</h2>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !books.length" description="暂无图书" />
      <NGrid v-else :cols="2" :x-gap="16" :y-gap="16">
        <NGi v-for="b in books" :key="b.id">
          <NCard hoverable size="small">
            <div style="display: flex; gap: 12px">
              <NImage v-if="b.cover" :src="b.cover" width="80" style="border-radius: 4px; flex-shrink: 0" />
              <div style="flex: 1; min-width: 0">
                <router-link :to="`/book/${b.id}`" style="font-weight: 500; font-size: 15px; text-decoration: none; color: inherit">{{ b.name }}</router-link>
                <div style="margin-top: 4px"><NText depth="3" style="font-size: 13px">{{ b.author }}{{ b.translator ? ` / ${b.translator} 译` : '' }}</NText></div>
                <NText depth="3" style="font-size: 12px">{{ b.intro?.substring(0, 80) }}</NText>
              </div>
            </div>
          </NCard>
        </NGi>
      </NGrid>
    </NSpin>
    <NPagination v-if="total > perPage" v-model:page="page" :page-count="Math.ceil(total / perPage)" style="margin-top: 16px; justify-content: center" @update:page="(p: number) => router.push({ query: { ...route.query, p: String(p) } })" />
  </div>
</template>
