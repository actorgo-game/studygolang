<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NGrid, NGi, NSpace, NSpin, NEmpty } from 'naive-ui'
import type { FriendLink } from '@/types'
import { getFriendLinks } from '@/api/sidebar'

const links = ref<FriendLink[]>([])
const loading = ref(true)

onMounted(async () => {
  try { links.value = await getFriendLinks() || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="友情链接">
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !links.length" description="暂无链接" />
      <NGrid v-else :cols="3" :x-gap="16" :y-gap="16">
        <NGi v-for="link in links" :key="link.id">
          <a :href="link.url" target="_blank" rel="noopener" style="text-decoration: none; color: inherit; display: block; padding: 12px; border: 1px solid #e8e8e8; border-radius: 8px; text-align: center">
            <img v-if="link.logo" :src="link.logo" style="height: 24px; margin-bottom: 4px" />
            <div>{{ link.name }}</div>
          </a>
        </NGi>
      </NGrid>
    </NSpin>
  </NCard>
</template>
