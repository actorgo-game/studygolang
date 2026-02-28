<script setup lang="ts">
import { NResult, NButton } from 'naive-ui'
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div style="padding: 80px 0; text-align: center">
    <NResult status="404" title="404" description="页面未找到">
      <template #footer>
        <NButton @click="router.push('/')">返回首页</NButton>
      </template>
    </NResult>
  </div>
</template>
