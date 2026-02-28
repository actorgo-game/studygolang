<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NTabs, NTabPane, NList, NListItem, NSpace, NText, NAvatar, NTag, NButton, NIcon } from 'naive-ui'
import { ChatbubbleOutline, EyeOutline, HeartOutline } from '@vicons/ionicons5'
import type { Topic, Article } from '@/types'
import { getTopics } from '@/api/topic'
import { getArticles } from '@/api/article'
import { timeAgo } from '@/utils/time'

const topics = ref<Topic[]>([])
const articles = ref<Article[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [topicData, articleData] = await Promise.all([
      getTopics({ p: 1 }),
      getArticles({ p: 1 }),
    ])
    topics.value = topicData?.list || []
    articles.value = articleData?.list || []
  } catch {}
  loading.value = false
})
</script>

<template>
  <div>
    <NTabs type="line" animated>
      <NTabPane name="topics" tab="最新主题">
        <NList :show-divider="true" hoverable clickable>
          <NListItem v-for="t in topics" :key="t.tid">
            <router-link :to="`/topics/${t.tid}`" class="topic-item">
              <div class="topic-row">
                <router-link v-if="t.user" :to="`/user/${t.user.username}`">
                  <NAvatar :src="t.user.avatar" :size="40" round />
                </router-link>
                <div class="topic-info">
                  <div class="topic-title">
                    <NTag v-if="t.top > 0" type="error" size="tiny">置顶</NTag>
                    {{ t.title }}
                  </div>
                  <NSpace :size="12" class="topic-meta">
                    <NText depth="3">{{ t.user?.username }}</NText>
                    <NText depth="3">{{ timeAgo(t.ctime) }}</NText>
                    <NSpace :size="4" align="center">
                      <NIcon :component="EyeOutline" size="14" />
                      <NText depth="3">{{ t.viewnum }}</NText>
                    </NSpace>
                    <NSpace :size="4" align="center">
                      <NIcon :component="ChatbubbleOutline" size="14" />
                      <NText depth="3">{{ t.cmtnum }}</NText>
                    </NSpace>
                    <NSpace :size="4" align="center">
                      <NIcon :component="HeartOutline" size="14" />
                      <NText depth="3">{{ t.likenum }}</NText>
                    </NSpace>
                    <NTag v-if="t.node?.name" size="tiny" round>{{ t.node?.name }}</NTag>
                  </NSpace>
                </div>
              </div>
            </router-link>
          </NListItem>
        </NList>
        <div style="text-align: center; margin-top: 16px">
          <router-link to="/topics"><NButton>查看更多主题</NButton></router-link>
        </div>
      </NTabPane>

      <NTabPane name="articles" tab="最新文章">
        <NList :show-divider="true" hoverable clickable>
          <NListItem v-for="a in articles" :key="a.id">
            <router-link :to="`/articles/${a.id}`" class="article-item">
              <div class="article-row">
                <img v-if="a.cover" :src="a.cover" class="article-cover" />
                <div class="article-info">
                  <div class="article-title">{{ a.title }}</div>
                  <p class="article-summary">{{ a.txt?.substring(0, 120) }}</p>
                  <NSpace :size="12" class="topic-meta">
                    <NText depth="3">{{ a.author_txt || a.author }}</NText>
                    <NText depth="3">{{ timeAgo(a.ctime) }}</NText>
                    <NSpace :size="4" align="center">
                      <NIcon :component="EyeOutline" size="14" />
                      <NText depth="3">{{ a.viewnum }}</NText>
                    </NSpace>
                    <NSpace :size="4" align="center">
                      <NIcon :component="ChatbubbleOutline" size="14" />
                      <NText depth="3">{{ a.cmtnum }}</NText>
                    </NSpace>
                  </NSpace>
                </div>
              </div>
            </router-link>
          </NListItem>
        </NList>
        <div style="text-align: center; margin-top: 16px">
          <router-link to="/articles"><NButton>查看更多文章</NButton></router-link>
        </div>
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
.topic-item, .article-item {
  text-decoration: none;
  color: inherit;
  display: block;
}
.topic-row, .article-row {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}
.topic-info, .article-info {
  flex: 1;
  min-width: 0;
}
.topic-title, .article-title {
  font-size: 15px;
  font-weight: 500;
  line-height: 1.4;
  display: flex;
  align-items: center;
  gap: 6px;
}
.article-summary {
  color: #666;
  font-size: 13px;
  margin: 4px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.topic-meta {
  font-size: 12px;
  margin-top: 4px;
}
.article-cover {
  width: 120px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
  flex-shrink: 0;
}
</style>
