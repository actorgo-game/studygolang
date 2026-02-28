<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NList, NListItem, NAvatar, NSpace, NText, NSpin, NEmpty } from 'naive-ui'
import { getDauRank, getRichRank } from '@/api/misc'

const props = defineProps<{ type: string }>()
const list = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try { list.value = (props.type === 'rich' ? await getRichRank() : await getDauRank()) || [] } catch {}
  loading.value = false
})
</script>

<template>
  <NCard :title="type === 'rich' ? '财富排行榜' : '活跃排行榜'">
    <NSpin :show="loading">
      <NEmpty v-if="!loading && !list.length" description="暂无数据" />
      <NList v-else :show-divider="true">
        <NListItem v-for="(item, i) in list" :key="i">
          <NSpace align="center">
            <NText strong style="width: 30px; text-align: center">{{ i + 1 }}</NText>
            <NAvatar :src="item.avatar" :size="32" round />
            <router-link :to="`/user/${item.username}`" style="text-decoration: none; color: inherit">
              <NText>{{ item.username }}</NText>
            </router-link>
            <NText type="warning">{{ item.weight || item.balance || item.num }}</NText>
          </NSpace>
        </NListItem>
      </NList>
    </NSpin>
  </NCard>
</template>
