<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NGrid, NGi, NTag, NSpin } from 'naive-ui'
import type { TopicNode } from '@/types'
import { getNodes } from '@/api/topic'

const nodes = ref<TopicNode[]>([])
const loading = ref(true)

onMounted(async () => {
  try { nodes.value = await getNodes() } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="所有节点">
    <NSpin :show="loading">
      <NGrid :cols="4" :x-gap="12" :y-gap="12">
        <NGi v-for="node in nodes" :key="node.nid">
          <router-link :to="`/topics/node/${node.nid}`">
            <NTag size="medium" round style="width: 100%; justify-content: center; cursor: pointer">
              {{ node.name }}
            </NTag>
          </router-link>
        </NGi>
      </NGrid>
    </NSpin>
  </NCard>
</template>
