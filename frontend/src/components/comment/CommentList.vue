<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { NList, NListItem, NAvatar, NSpace, NText, NButton, NPagination, NEmpty } from 'naive-ui'
import type { Comment } from '@/types'
import { getComments } from '@/api/comment'
import { timeAgo } from '@/utils/time'
import { renderMarkdown } from '@/utils/markdown'
import CommentForm from './CommentForm.vue'

const props = defineProps<{
  objid: number
  objtype: number
}>()

const comments = ref<Comment[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20
const loading = ref(false)

async function loadComments() {
  loading.value = true
  try {
    const data = await getComments({ objid: props.objid, objtype: props.objtype, p: page.value })
    comments.value = data.list || []
    total.value = data.total || 0
  } catch {}
  loading.value = false
}

function onCommentCreated() {
  page.value = 1
  loadComments()
}

watch(() => props.objid, loadComments)
onMounted(loadComments)
</script>

<template>
  <div class="comment-section">
    <h3>{{ total }} 条评论</h3>
    <CommentForm :objid="objid" :objtype="objtype" @created="onCommentCreated" />

    <NEmpty v-if="!comments.length && !loading" description="暂无评论" style="margin: 24px 0" />

    <NList v-else :show-divider="true" style="margin-top: 16px">
      <NListItem v-for="c in comments" :key="c.cid">
        <div class="comment-item">
          <router-link :to="`/user/${c.user?.username}`">
            <NAvatar :src="c.user?.avatar" :size="36" round />
          </router-link>
          <div class="comment-body">
            <NSpace align="center" :size="8">
              <router-link :to="`/user/${c.user?.username}`">
                <NText strong>{{ c.user?.username }}</NText>
              </router-link>
              <NText depth="3" style="font-size: 12px">#{{ c.floor }}</NText>
              <NText depth="3" style="font-size: 12px">{{ timeAgo(c.ctime) }}</NText>
            </NSpace>
            <div class="comment-content" v-html="renderMarkdown(c.content)" />
          </div>
        </div>
      </NListItem>
    </NList>

    <NPagination
      v-if="total > perPage"
      v-model:page="page"
      :page-count="Math.ceil(total / perPage)"
      style="margin-top: 16px; justify-content: center"
      @update:page="loadComments"
    />
  </div>
</template>

<style scoped>
.comment-item {
  display: flex;
  gap: 12px;
}
.comment-body {
  flex: 1;
  min-width: 0;
}
.comment-content {
  margin-top: 4px;
  line-height: 1.6;
}
.comment-content :deep(p) {
  margin: 4px 0;
}
.comment-section a {
  text-decoration: none;
  color: inherit;
}
</style>
