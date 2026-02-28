<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NLayout, NLayoutSider, NLayoutContent, NMenu, NIcon, NText, NSpace } from 'naive-ui'
import { PeopleOutline, ChatbubblesOutline, LibraryOutline, GitNetworkOutline, ReaderOutline, SettingsOutline, HomeOutline, NewspaperOutline } from '@vicons/ionicons5'
import { h } from 'vue'

const route = useRoute()
const router = useRouter()

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions = [
  { label: '仪表盘', key: '/admin', icon: renderIcon(HomeOutline) },
  { label: '用户管理', key: '/admin/user/user/list', icon: renderIcon(PeopleOutline) },
  { label: '主题管理', key: '/admin/community/topic/list', icon: renderIcon(ChatbubblesOutline) },
  { label: '文章管理', key: '/admin/crawl/article/list', icon: renderIcon(NewspaperOutline) },
  { label: '节点管理', key: '/admin/community/node/list', icon: renderIcon(GitNetworkOutline) },
  { label: '爬取规则', key: '/admin/crawl/rule/list', icon: renderIcon(LibraryOutline) },
  { label: '晨读管理', key: '/admin/reading/list', icon: renderIcon(ReaderOutline) },
  { label: '系统设置', key: '/admin/setting', icon: renderIcon(SettingsOutline) },
]

const activeKey = computed(() => route.path)
const collapsed = ref(false)

function onMenuSelect(key: string) {
  router.push(key)
}
</script>

<template>
  <NLayout has-sider style="min-height: 100vh">
    <NLayoutSider
      bordered
      :collapsed="collapsed"
      collapse-mode="width"
      :collapsed-width="64"
      :width="220"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
      :native-scrollbar="false"
      style="background: #fff"
    >
      <div style="padding: 16px; text-align: center">
        <NText strong style="font-size: 16px">{{ collapsed ? 'SG' : '管理后台' }}</NText>
      </div>
      <NMenu
        :options="menuOptions"
        :value="activeKey"
        @update:value="onMenuSelect"
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
      />
    </NLayoutSider>
    <NLayout>
      <NLayoutContent style="padding: 24px; background: #f5f5f5">
        <router-view />
      </NLayoutContent>
    </NLayout>
  </NLayout>
</template>
