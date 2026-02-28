<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NList, NListItem, NSpace, NText, NAvatar, NTag, NIcon, NPagination, NButton, NEmpty, NSpin } from 'naive-ui'
import { ChatbubbleOutline, EyeOutline, HeartOutline } from '@vicons/ionicons5'
import type { Topic } from '@/types'
import { getTopics, getNoReplyTopics, getLastTopics, getNodeTopics } from '@/api/topic'
import { timeAgo } from '@/utils/time'
import { useUserStore } from '@/stores/user'

const props = defineProps<{ tab?: string }>()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const topics = ref<Topic[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20
const loading = ref(true)
const currentTab = ref(props.tab || 'all')

async function loadTopics() {
  loading.value = true
  try {
    const nid = route.params.nid ? Number(route.params.nid) : 0
    let data: any
    if (nid) {
      data = await getNodeTopics(nid, { p: page.value })
    } else if (currentTab.value === 'no_reply') {
      data = await getNoReplyTopics({ p: page.value })
    } else if (currentTab.value === 'last') {
      data = await getLastTopics({ p: page.value })
    } else {
      data = await getTopics({ p: page.value, tab: currentTab.value })
    }
    topics.value = data?.list || []
    total.value = data?.total || 0
  } catch {}
  loading.value = false
}

watch(() => [route.params, route.query], () => {
  page.value = Number(route.query.p) || 1
  loadTopics()
})

onMounted(loadTopics)
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <h2 style="margin: 0">主题列表</h2>
      <NButton v-if="userStore.isLoggedIn" type="primary" @click="router.push('/topics/new')">
        发布主题
      </NButton>
    </NSpace>

    <NSpin :show="loading">
      <NEmpty v-if="!loading && !topics.length" description="暂无主题" />
      <NList v-else :show-divider="true" hoverable>
        <NListItem v-for="t in topics" :key="t.tid">
          <div class="topic-row">
            <router-link v-if="t.user" :to="`/user/${t.user.username}`">
              <NAvatar :src="t.user?.avatar" :size="40" round />
            </router-link>
            <div class="topic-info">
              <router-link :to="`/topics/${t.tid}`" class="topic-title">
                <NTag v-if="t.top > 0" type="error" size="tiny">置顶</NTag>
                {{ t.title }}
              </router-link>
              <NSpace :size="12" class="topic-meta">
                <NText depth="3">{{ t.user?.username }}</NText>
                <NText depth="3">{{ timeAgo(t.ctime) }}</NText>
                <NSpace :size="4" align="center">
                  <NIcon :component="EyeOutline" size="14" /><NText depth="3">{{ t.viewnum }}</NText>
                </NSpace>
                <NSpace :size="4" align="center">
                  <NIcon :component="ChatbubbleOutline" size="14" /><NText depth="3">{{ t.cmtnum }}</NText>
                </NSpace>
                <NSpace :size="4" align="center">
                  <NIcon :component="HeartOutline" size="14" /><NText depth="3">{{ t.likenum }}</NText>
                </NSpace>
                <NTag v-if="t.node?.name" size="tiny" round>{{ t.node?.name }}</NTag>
              </NSpace>
            </div>
          </div>
        </NListItem>
      </NList>
    </NSpin>

    <NPagination
      v-if="total > perPage"
      v-model:page="page"
      :page-count="Math.ceil(total / perPage)"
      style="margin-top: 16px; justify-content: center"
      @update:page="(p: number) => router.push({ query: { ...route.query, p: String(p) } })"
    />
  </div>
</template>

<style scoped>
.topic-row { display: flex; gap: 12px; align-items: flex-start; }
.topic-info { flex: 1; min-width: 0; }
.topic-title { font-size: 15px; font-weight: 500; text-decoration: none; color: inherit; display: flex; align-items: center; gap: 6px; }
.topic-title:hover { color: #18a058; }
.topic-meta { font-size: 12px; margin-top: 4px; }
</style>
