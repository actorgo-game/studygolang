<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NGrid, NGi, NStatistic, NSpin } from 'naive-ui'
import { getSiteStat } from '@/api/sidebar'
import type { SiteStat } from '@/types'

const router = useRouter()
const stat = ref<SiteStat | null>(null)
const loading = ref(true)

onMounted(async () => {
  try { stat.value = await getSiteStat() } catch {}
  loading.value = false
})

const cards = [
  { label: '注册用户', key: 'user', path: '/admin/user/user/list' },
  { label: '主题数量', key: 'topic', path: '/admin/community/topic/list' },
  { label: '文章数量', key: 'article', path: '/admin/crawl/article/list' },
  { label: '评论数量', key: 'comment', path: '' },
  { label: '资源数量', key: 'resource', path: '/admin/resource/list' },
  { label: '项目数量', key: 'project', path: '' },
  { label: '图书数量', key: 'book', path: '/admin/book/list' },
]
</script>

<template>
  <div>
    <h2 style="margin: 0 0 24px">管理仪表盘</h2>
    <NSpin :show="loading">
      <NGrid :cols="4" :x-gap="16" :y-gap="16" v-if="stat">
        <NGi v-for="c in cards" :key="c.key">
          <NCard hoverable style="cursor: pointer" @click="c.path && router.push(c.path)">
            <NStatistic :label="c.label" :value="(stat as any)[c.key] ?? 0" />
          </NCard>
        </NGi>
      </NGrid>
    </NSpin>
  </div>
</template>
