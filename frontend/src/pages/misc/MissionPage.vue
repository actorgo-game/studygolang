<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NList, NListItem, NButton, NText, NSpace, NTag, NSpin, useMessage } from 'naive-ui'
import type { Mission } from '@/types'
import { getDailyMission, redeemDailyMission } from '@/api/misc'

const message = useMessage()
const missions = ref<Mission[]>([])
const loading = ref(true)

onMounted(async () => {
  try { const data = await getDailyMission(); missions.value = data?.missions || [] } catch {}
  loading.value = false
})

async function handleRedeem() {
  try { await redeemDailyMission(); message.success('领取成功！') } catch (e: any) { message.error(e.message || '领取失败') }
}
</script>

<template>
  <NCard title="每日任务">
    <NButton type="primary" @click="handleRedeem" style="margin-bottom: 16px">领取每日登录奖励</NButton>
    <NSpin :show="loading">
      <NList :show-divider="true">
        <NListItem v-for="m in missions" :key="m.id">
          <NSpace justify="space-between" align="center" style="width: 100%">
            <div><NText strong>{{ m.name }}</NText><NText depth="3" style="margin-left: 8px">+{{ m.award }} 铜币</NText></div>
            <NTag :type="m.state === 1 ? 'success' : 'default'" size="small">{{ m.state === 1 ? '已完成' : '未完成' }}</NTag>
          </NSpace>
        </NListItem>
      </NList>
    </NSpin>
  </NCard>
</template>
