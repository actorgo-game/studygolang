<script setup lang="ts">
import { NAvatar } from 'naive-ui'

defineProps<{
  src?: string
  username?: string
  size?: number
}>()
</script>

<template>
  <router-link v-if="username" :to="`/user/${username}`" class="user-avatar-link">
    <NAvatar
      :src="src"
      :size="size || 32"
      round
      :fallback-src="`https://www.gravatar.com/avatar/?d=identicon`"
    />
  </router-link>
  <NAvatar
    v-else
    :src="src"
    :size="size || 32"
    round
    :fallback-src="`https://www.gravatar.com/avatar/?d=identicon`"
  />
</template>

<style scoped>
.user-avatar-link {
  display: inline-flex;
}
</style>
