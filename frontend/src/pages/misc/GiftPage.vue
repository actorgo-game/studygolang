<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NGrid, NGi, NButton, NText, NImage, NSpin, NEmpty, useMessage } from 'naive-ui'
import type { Gift } from '@/types'
import { getGifts, exchangeGift } from '@/api/misc'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

const userStore = useUserStore()
const appStore = useAppStore()
const message = useMessage()
const gifts = ref<Gift[]>([])
const loading = ref(true)

onMounted(async () => {
  try { gifts.value = await getGifts() || [] } catch {}
  loading.value = false
})

async function handleExchange(id: number) {
  if (!userStore.isLoggedIn) { appStore.openLoginModal(); return }
  try { await exchangeGift({ gift_id: id }); message.success('兑换成功！') } catch (e: any) { message.error(e.message || '兑换失败') }
}
</script>

<template>
  <NCard title="礼品商城">
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !gifts.length" description="暂无礼品" />
      <NGrid v-else :cols="3" :x-gap="16" :y-gap="16">
        <NGi v-for="g in gifts" :key="g.id">
          <NCard size="small" hoverable>
            <NImage v-if="g.image" :src="g.image" height="120" style="width: 100%; object-fit: cover; border-radius: 4px" />
            <h4 style="margin: 8px 0 4px">{{ g.name }}</h4>
            <NText depth="3" style="font-size: 13px">{{ g.desc }}</NText>
            <div style="margin-top: 8px"><NText type="warning">{{ g.price }} 铜币</NText><NText depth="3" style="margin-left: 8px">剩余 {{ g.remain_num }}</NText></div>
            <NButton type="primary" size="small" style="margin-top: 8px" block @click="handleExchange(g.id)" :disabled="g.remain_num <= 0">兑换</NButton>
          </NCard>
        </NGi>
      </NGrid>
    </NSpin>
  </NCard>
</template>
