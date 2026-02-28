<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NList, NListItem, NAvatar, NSpace, NText, NEmpty, NSpin, NTabs, NTabPane } from 'naive-ui'
import type { Message } from '@/types'
import { getMessages } from '@/api/message'
import { timeAgo } from '@/utils/time'

const route = useRoute()
const messages = ref<Message[]>([])
const loading = ref(true)

async function load() {
  loading.value = true
  const msgtype = String(route.params.msgtype || 'system')
  try { const data = await getMessages(msgtype, { p: 1 }); messages.value = (data as any)?.list || [] } catch {}
  loading.value = false
}

watch(() => route.params.msgtype, load)
onMounted(load)
</script>

<template>
  <NCard title="消息中心">
    <NTabs type="line" @update:value="(v: string) => $router.push(`/message/${v}`)">
      <NTabPane name="system" tab="系统消息" />
      <NTabPane name="inbox" tab="收件箱" />
      <NTabPane name="outbox" tab="发件箱" />
    </NTabs>
    <NSpin :show="loading" style="margin-top: 16px">
      <NEmpty v-if="!loading && !messages.length" description="暂无消息" />
      <NList v-else :show-divider="true">
        <NListItem v-for="m in messages" :key="m.id">
          <NSpace>
            <NAvatar v-if="m.from_user" :src="m.from_user.avatar" :size="32" round />
            <div>
              <div>{{ m.content }}</div>
              <NText depth="3" style="font-size: 12px">{{ timeAgo(m.ctime) }}</NText>
            </div>
          </NSpace>
        </NListItem>
      </NList>
    </NSpin>
  </NCard>
</template>
