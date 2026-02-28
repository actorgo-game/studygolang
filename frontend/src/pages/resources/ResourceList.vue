<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NSpace, NButton, NPagination, NEmpty, NSpin } from 'naive-ui'
import type { Resource } from '@/types'
import { getResources } from '@/api/resource'
import { useUserStore } from '@/stores/user'
import ContentCard from '@/components/common/ContentCard.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const resources = ref<Resource[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const catid = route.params.catid ? Number(route.params.catid) : undefined
    const data = await getResources({ p: page.value, catid })
    resources.value = data?.list || []
    total.value = data?.total || 0
  } catch {}
  loading.value = false
}

watch(() => route.query, () => { page.value = Number(route.query.p) || 1; load() })
onMounted(load)
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <h2 style="margin: 0">资源列表</h2>
      <NButton v-if="userStore.isLoggedIn" type="primary" @click="router.push('/resources/new')">分享资源</NButton>
    </NSpace>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !resources.length" description="暂无资源" />
      <div v-else class="list"><ContentCard v-for="r in resources" :key="r.id" :title="r.title" :url="`/resources/${r.id}`" :author="r.user?.username" :author-url="r.user ? `/user/${r.user.username}` : undefined" :avatar="r.user?.avatar" :time="r.ctime" :viewnum="r.viewnum" :cmtnum="r.cmtnum" :likenum="r.likenum" /></div>
    </NSpin>
    <NPagination v-if="total > perPage" v-model:page="page" :page-count="Math.ceil(total / perPage)" style="margin-top: 16px; justify-content: center" @update:page="(p: number) => router.push({ query: { ...route.query, p: String(p) } })" />
  </div>
</template>

<style scoped>.list { display: flex; flex-direction: column; gap: 12px; }</style>
