<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NSpace, NAvatar, NTooltip } from 'naive-ui'
import type { User } from '@/types'
import { getActiveUsers } from '@/api/sidebar'

const users = ref<User[]>([])

onMounted(async () => {
  try { users.value = await getActiveUsers() } catch {}
})
</script>

<template>
  <NCard size="small" title="活跃会员">
    <NSpace>
      <router-link v-for="u in users" :key="u.uid" :to="`/user/${u.username}`">
        <NTooltip>
          <template #trigger>
            <NAvatar :src="u.avatar" :size="32" round :fallback-src="`https://www.gravatar.com/avatar/?d=identicon`" />
          </template>
          {{ u.username }}
        </NTooltip>
      </router-link>
    </NSpace>
  </NCard>
</template>
