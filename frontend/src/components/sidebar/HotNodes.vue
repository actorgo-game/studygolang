<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NSpace, NTag } from 'naive-ui'
import type { TopicNode } from '@/types'
import { getHotNodes } from '@/api/sidebar'

const nodes = ref<TopicNode[]>([])

onMounted(async () => {
  try { nodes.value = await getHotNodes() } catch {}
})
</script>

<template>
  <NCard size="small" title="热门节点">
    <NSpace>
      <router-link v-for="node in nodes" :key="node.nid" :to="`/topics/node/${node.nid}`">
        <NTag size="small" round>{{ node.name }}</NTag>
      </router-link>
    </NSpace>
  </NCard>
</template>
