<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NSpace } from 'naive-ui'
import type { FriendLink } from '@/types'
import { getFriendLinks } from '@/api/sidebar'

const links = ref<FriendLink[]>([])

onMounted(async () => {
  try { links.value = await getFriendLinks() } catch {}
})
</script>

<template>
  <NCard size="small" title="友情链接">
    <NSpace>
      <a v-for="link in links" :key="link.id" :href="link.url" target="_blank" rel="noopener" class="friend-link">
        {{ link.name }}
      </a>
    </NSpace>
  </NCard>
</template>

<style scoped>
.friend-link {
  text-decoration: none;
  color: #666;
  font-size: 13px;
}
.friend-link:hover {
  color: #18a058;
}
</style>
