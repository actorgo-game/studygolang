<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NSpace, NButton, NEmpty, NSpin } from 'naive-ui'
import type { Wiki } from '@/types'
import { getWikiList } from '@/api/wiki'
import { useUserStore } from '@/stores/user'
import ContentCard from '@/components/common/ContentCard.vue'

const router = useRouter()
const userStore = useUserStore()
const wikis = ref<Wiki[]>([])
const loading = ref(true)

onMounted(async () => {
  try { const data = await getWikiList({ p: 1 }); wikis.value = data?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <h2 style="margin: 0">Wiki</h2>
      <NButton v-if="userStore.isLoggedIn" type="primary" @click="router.push('/wiki/new')">新建Wiki</NButton>
    </NSpace>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !wikis.length" description="暂无Wiki" />
      <div v-else class="list"><ContentCard v-for="w in wikis" :key="w.id" :title="w.title" :url="`/wiki/${w.uri}`" :author="w.user?.username" :author-url="w.user ? `/user/${w.user.username}` : undefined" :time="w.ctime" :viewnum="w.viewnum" :cmtnum="w.cmtnum" /></div>
    </NSpin>
  </div>
</template>

<style scoped>.list { display: flex; flex-direction: column; gap: 12px; }</style>
