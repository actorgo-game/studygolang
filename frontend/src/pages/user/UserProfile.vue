<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NAvatar, NSpace, NText, NTabs, NTabPane, NList, NListItem, NTag, NEmpty, NSpin } from 'naive-ui'
import type { User, Topic, Article, Comment } from '@/types'
import { getUserProfile, getUserTopics, getUserArticles, getUserComments } from '@/api/user'
import { timeAgo } from '@/utils/time'

const route = useRoute()
const user = ref<User | null>(null)
const topics = ref<Topic[]>([])
const articles = ref<Article[]>([])
const comments = ref<Comment[]>([])
const loading = ref(true)

async function load() {
  loading.value = true
  const username = String(route.params.username)
  try {
    const data = await getUserProfile(username)
    user.value = data?.user || null
    const [t, a, c] = await Promise.all([
      getUserTopics(username, { p: 1 }).catch(() => ({ list: [] })),
      getUserArticles(username, { p: 1 }).catch(() => ({ list: [] })),
      getUserComments(username, { p: 1 }).catch(() => ({ list: [] })),
    ])
    topics.value = (t as any)?.list || []
    articles.value = (a as any)?.list || []
    comments.value = (c as any)?.list || []
  } catch {}
  loading.value = false
}

watch(() => route.params.username, load)
onMounted(load)
</script>

<template>
  <NSpin :show="loading">
    <NCard v-if="user">
      <div class="profile-header">
        <NAvatar :src="user.avatar" :size="80" round />
        <div>
          <h2 style="margin: 0">{{ user.name || user.username }}</h2>
          <NText depth="3">@{{ user.username }}</NText>
          <NSpace :size="12" style="margin-top: 8px">
            <NText v-if="user.city" depth="3">{{ user.city }}</NText>
            <NText v-if="user.company" depth="3">{{ user.company }}</NText>
            <NTag v-if="user.role_name" size="small" type="info">{{ user.role_name }}</NTag>
          </NSpace>
          <p v-if="user.introduce" style="margin: 8px 0 0; color: #666">{{ user.introduce }}</p>
          <NSpace :size="12" style="margin-top: 8px">
            <a v-if="user.github" :href="`https://github.com/${user.github}`" target="_blank" rel="noopener">GitHub</a>
            <a v-if="user.website" :href="user.website" target="_blank" rel="noopener">个人网站</a>
          </NSpace>
        </div>
      </div>
    </NCard>

    <NTabs type="line" style="margin-top: 16px" v-if="user">
      <NTabPane name="topics" :tab="`主题 (${topics.length})`">
        <NEmpty v-if="!topics.length" description="暂无主题" />
        <NList v-else :show-divider="true">
          <NListItem v-for="t in topics" :key="t.tid">
            <router-link :to="`/topics/${t.tid}`" style="text-decoration: none; color: inherit">
              <div style="font-weight: 500">{{ t.title }}</div>
              <NText depth="3" style="font-size: 12px">{{ timeAgo(t.ctime) }} · {{ t.viewnum }} 阅读 · {{ t.cmtnum }} 评论</NText>
            </router-link>
          </NListItem>
        </NList>
      </NTabPane>
      <NTabPane name="articles" :tab="`文章 (${articles.length})`">
        <NEmpty v-if="!articles.length" description="暂无文章" />
        <NList v-else :show-divider="true">
          <NListItem v-for="a in articles" :key="a.id">
            <router-link :to="`/articles/${a.id}`" style="text-decoration: none; color: inherit">
              <div style="font-weight: 500">{{ a.title }}</div>
              <NText depth="3" style="font-size: 12px">{{ timeAgo(a.ctime) }} · {{ a.viewnum }} 阅读</NText>
            </router-link>
          </NListItem>
        </NList>
      </NTabPane>
      <NTabPane name="comments" :tab="`评论 (${comments.length})`">
        <NEmpty v-if="!comments.length" description="暂无评论" />
        <NList v-else :show-divider="true">
          <NListItem v-for="c in comments" :key="c.cid">
            <div style="font-size: 13px; color: #666">{{ c.content?.substring(0, 200) }}</div>
            <NText depth="3" style="font-size: 12px">{{ timeAgo(c.ctime) }}</NText>
          </NListItem>
        </NList>
      </NTabPane>
    </NTabs>
  </NSpin>
</template>

<style scoped>
.profile-header { display: flex; gap: 20px; align-items: flex-start; }
a { color: #18a058; text-decoration: none; }
</style>
