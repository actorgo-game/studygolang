<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NSpace, NButton, NPagination, NEmpty, NSpin } from 'naive-ui'
import type { Project } from '@/types'
import { getProjects } from '@/api/project'
import { useUserStore } from '@/stores/user'
import ContentCard from '@/components/common/ContentCard.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const projects = ref<Project[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20
const loading = ref(true)

async function load() {
  loading.value = true
  try { const data = await getProjects({ p: page.value }); projects.value = data?.list || []; total.value = data?.total || 0 } catch {}
  loading.value = false
}
watch(() => route.query, () => { page.value = Number(route.query.p) || 1; load() })
onMounted(load)
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <h2 style="margin: 0">开源项目</h2>
      <NButton v-if="userStore.isLoggedIn" type="primary" @click="router.push('/project/new')">发布项目</NButton>
    </NSpace>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !projects.length" description="暂无项目" />
      <div v-else class="list"><ContentCard v-for="p in projects" :key="p.id" :title="p.name" :url="`/p/${p.uri}`" :author="p.username || p.author" :time="p.ctime" :viewnum="p.viewnum" :cmtnum="p.cmtnum" :likenum="p.likenum" :summary="p.desc?.substring(0, 150)" :cover="p.logo" :tags="p.tags" /></div>
    </NSpin>
    <NPagination v-if="total > perPage" v-model:page="page" :page-count="Math.ceil(total / perPage)" style="margin-top: 16px; justify-content: center" @update:page="(p: number) => router.push({ query: { ...route.query, p: String(p) } })" />
  </div>
</template>

<style scoped>.list { display: flex; flex-direction: column; gap: 12px; }</style>
