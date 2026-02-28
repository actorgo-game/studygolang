<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NGrid, NGi, NAvatar, NSpace, NText, NSpin, NEmpty } from 'naive-ui'
import type { User } from '@/types'
import { getUsers } from '@/api/user'

const users = ref<User[]>([])
const loading = ref(true)

onMounted(async () => {
  try { const data = await getUsers({ p: 1 }); users.value = (data as any)?.list || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="社区成员">
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !users.length" description="暂无用户" />
      <NGrid v-else :cols="4" :x-gap="16" :y-gap="16">
        <NGi v-for="u in users" :key="u.uid">
          <router-link :to="`/user/${u.username}`" style="text-decoration: none; color: inherit; text-align: center; display: block">
            <NAvatar :src="u.avatar" :size="48" round style="margin: 0 auto" />
            <NText style="margin-top: 4px; display: block; font-size: 13px">{{ u.username }}</NText>
          </router-link>
        </NGi>
      </NGrid>
    </NSpin>
  </NCard>
</template>
