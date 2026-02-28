<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NStatistic, NList, NListItem, NText, NSpin } from 'naive-ui'
import { getBalance } from '@/api/misc'

const balance = ref(0)
const records = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try { const data = await getBalance(); balance.value = data?.balance || 0; records.value = data?.records || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard title="我的余额">
    <NStatistic label="当前铜币" :value="balance" style="margin-bottom: 24px" />
    <NSpin :show="loading">
      <h3>交易记录</h3>
      <NList :show-divider="true">
        <NListItem v-for="(r, i) in records" :key="i">
          <div>{{ r.desc || r.type }}</div>
          <NText :type="r.amount > 0 ? 'success' : 'error'">{{ r.amount > 0 ? '+' : '' }}{{ r.amount }}</NText>
        </NListItem>
      </NList>
    </NSpin>
  </NCard>
</template>
