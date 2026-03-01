<script setup lang="ts">
import AppHeader from './AppHeader.vue'
import AppFooter from './AppFooter.vue'
import AppSidebar from './AppSidebar.vue'
import { useHeartbeat } from '@/composables/useHeartbeat'

useHeartbeat()
</script>

<template>
  <div class="app-container">
    <AppHeader />
    <div class="main-wrapper">
      <main class="main-content">
        <router-view />
      </main>
      <AppSidebar />
    </div>
    <AppFooter />
  </div>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  background: #f5f5f5;
}
.main-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px 16px;
  display: flex;
  gap: 20px;
}
.main-content {
  flex: 1;
  min-width: 0;
}
</style>
