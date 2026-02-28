<script setup lang="ts">
import { NSpace, NText } from 'naive-ui'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
</script>

<template>
  <footer class="app-footer">
    <div class="footer-inner">
      <NSpace vertical align="center" :size="8">
        <NText depth="3">
          当前在线：{{ appStore.onlineUsers }} 人 | 历史最高：{{ appStore.maxOnlineUsers }} 人
        </NText>
        <NSpace :size="16">
          <router-link to="/links"><NText depth="3">友情链接</NText></router-link>
          <router-link to="/nodes"><NText depth="3">节点列表</NText></router-link>
        </NSpace>
        <NText depth="3" style="font-size: 12px">
          &copy; Go语言中文网 | Powered by Vue 3 + Naive UI
        </NText>
      </NSpace>
    </div>
  </footer>
</template>

<style scoped>
.app-footer {
  border-top: 1px solid #e8e8e8;
  padding: 24px 16px;
  margin-top: 48px;
  background: #fafafa;
}
.footer-inner {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
}
.app-footer a {
  text-decoration: none;
}
</style>
