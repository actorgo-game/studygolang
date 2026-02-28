<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import SiteStat from '@/components/sidebar/SiteStat.vue'
import HotNodes from '@/components/sidebar/HotNodes.vue'
import ViewRank from '@/components/sidebar/ViewRank.vue'
import RecentTopics from '@/components/sidebar/RecentTopics.vue'
import RecentArticles from '@/components/sidebar/RecentArticles.vue'
import ActiveUsers from '@/components/sidebar/ActiveUsers.vue'
import FriendLinks from '@/components/sidebar/FriendLinks.vue'

const route = useRoute()

const showNodes = computed(() => route.path.startsWith('/topics') || route.path === '/')
const showArticles = computed(() => route.path.startsWith('/articles') || route.path.startsWith('/projects') || route.path.startsWith('/wiki'))
</script>

<template>
  <aside class="app-sidebar">
    <SiteStat />
    <HotNodes v-if="showNodes" />
    <ViewRank />
    <RecentTopics />
    <RecentArticles v-if="showArticles" />
    <ActiveUsers />
    <FriendLinks />
  </aside>
</template>

<style scoped>
.app-sidebar {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
