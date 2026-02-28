<script setup lang="ts">
import { NCard, NSpace, NText, NTag, NAvatar, NIcon } from 'naive-ui'
import { EyeOutline, ChatbubbleOutline, HeartOutline } from '@vicons/ionicons5'
import { timeAgo } from '@/utils/time'

defineProps<{
  title: string
  url: string
  author?: string
  authorUrl?: string
  avatar?: string
  time?: string
  tags?: string
  viewnum?: number
  cmtnum?: number
  likenum?: number
  summary?: string
  cover?: string
}>()
</script>

<template>
  <NCard size="small" hoverable class="content-card">
    <div class="card-layout">
      <img v-if="cover" :src="cover" class="card-cover" />
      <div class="card-body">
        <router-link :to="url" class="card-title">{{ title }}</router-link>
        <p v-if="summary" class="card-summary">{{ summary }}</p>
        <NSpace align="center" :size="12" class="card-meta">
          <NSpace v-if="author" align="center" :size="4">
            <router-link v-if="authorUrl" :to="authorUrl">
              <NAvatar :src="avatar" :size="20" round />
            </router-link>
            <router-link v-if="authorUrl" :to="authorUrl">
              <NText depth="3" style="font-size: 13px">{{ author }}</NText>
            </router-link>
            <NText v-else depth="3" style="font-size: 13px">{{ author }}</NText>
          </NSpace>
          <NText v-if="time" depth="3" style="font-size: 12px">{{ timeAgo(time) }}</NText>
          <NSpace v-if="viewnum !== undefined" :size="4" align="center">
            <NIcon :component="EyeOutline" size="14" /><NText depth="3" style="font-size: 12px">{{ viewnum }}</NText>
          </NSpace>
          <NSpace v-if="cmtnum !== undefined" :size="4" align="center">
            <NIcon :component="ChatbubbleOutline" size="14" /><NText depth="3" style="font-size: 12px">{{ cmtnum }}</NText>
          </NSpace>
          <NSpace v-if="likenum !== undefined" :size="4" align="center">
            <NIcon :component="HeartOutline" size="14" /><NText depth="3" style="font-size: 12px">{{ likenum }}</NText>
          </NSpace>
        </NSpace>
        <NSpace v-if="tags" :size="4" style="margin-top: 4px">
          <NTag v-for="tag in tags.split(',')" :key="tag" size="tiny" round>{{ tag.trim() }}</NTag>
        </NSpace>
      </div>
    </div>
  </NCard>
</template>

<style scoped>
.content-card a {
  text-decoration: none;
  color: inherit;
}
.card-layout {
  display: flex;
  gap: 12px;
}
.card-cover {
  width: 120px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
  flex-shrink: 0;
}
.card-body {
  flex: 1;
  min-width: 0;
}
.card-title {
  font-size: 15px;
  font-weight: 500;
  line-height: 1.4;
  display: block;
}
.card-title:hover {
  color: #18a058;
}
.card-summary {
  color: #666;
  font-size: 13px;
  margin: 4px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-meta {
  margin-top: 4px;
}
</style>
