<script setup lang="ts">
import { computed } from 'vue'
import { NSpace, NTag } from 'naive-ui'

const props = defineProps<{
  tags: string
}>()

const tagList = computed(() => {
  if (!props.tags) return []
  return props.tags.split(',').map((t) => t.trim()).filter(Boolean)
})
</script>

<template>
  <NSpace v-if="tagList.length" :size="4">
    <router-link v-for="tag in tagList" :key="tag" :to="`/search?q=${tag}`">
      <NTag size="small" round>{{ tag }}</NTag>
    </router-link>
  </NSpace>
</template>
