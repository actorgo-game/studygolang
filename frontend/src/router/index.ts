import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    children: [
      { path: '', name: 'Home', component: () => import('@/pages/home/IndexPage.vue') },
      // Topics
      { path: 'topics', name: 'Topics', component: () => import('@/pages/topics/TopicList.vue') },
      { path: 'topics/no_reply', name: 'TopicsNoReply', component: () => import('@/pages/topics/TopicList.vue'), props: { tab: 'no_reply' } },
      { path: 'topics/last', name: 'TopicsLast', component: () => import('@/pages/topics/TopicList.vue'), props: { tab: 'last' } },
      { path: 'topics/node/:nid', name: 'TopicsByNode', component: () => import('@/pages/topics/TopicList.vue') },
      { path: 'topics/:tid', name: 'TopicDetail', component: () => import('@/pages/topics/TopicDetail.vue') },
      { path: 'topics/new', name: 'TopicNew', component: () => import('@/pages/topics/TopicEdit.vue'), meta: { requireAuth: true } },
      { path: 'topics/modify', name: 'TopicModify', component: () => import('@/pages/topics/TopicEdit.vue'), meta: { requireAuth: true } },
      // Articles
      { path: 'articles', name: 'Articles', component: () => import('@/pages/articles/ArticleList.vue') },
      { path: 'articles/:id', name: 'ArticleDetail', component: () => import('@/pages/articles/ArticleDetail.vue') },
      { path: 'articles/new', name: 'ArticleNew', component: () => import('@/pages/articles/ArticleEdit.vue'), meta: { requireAuth: true } },
      { path: 'articles/modify', name: 'ArticleModify', component: () => import('@/pages/articles/ArticleEdit.vue'), meta: { requireAuth: true } },
      // Resources
      { path: 'resources', name: 'Resources', component: () => import('@/pages/resources/ResourceList.vue') },
      { path: 'resources/:id', name: 'ResourceDetail', component: () => import('@/pages/resources/ResourceDetail.vue') },
      { path: 'resources/new', name: 'ResourceNew', component: () => import('@/pages/resources/ResourceEdit.vue'), meta: { requireAuth: true } },
      // Projects
      { path: 'projects', name: 'Projects', component: () => import('@/pages/projects/ProjectList.vue') },
      { path: 'p/:uri', name: 'ProjectDetail', component: () => import('@/pages/projects/ProjectDetail.vue') },
      { path: 'project/new', name: 'ProjectNew', component: () => import('@/pages/projects/ProjectEdit.vue'), meta: { requireAuth: true } },
      // Books
      { path: 'books', name: 'Books', component: () => import('@/pages/books/BookList.vue') },
      { path: 'book/:id', name: 'BookDetail', component: () => import('@/pages/books/BookDetail.vue') },
      // Wiki
      { path: 'wiki', name: 'Wiki', component: () => import('@/pages/wiki/WikiList.vue') },
      { path: 'wiki/new', name: 'WikiNew', component: () => import('@/pages/wiki/WikiEdit.vue'), meta: { requireAuth: true } },
      { path: 'wiki/edit/:id', name: 'WikiEdit', component: () => import('@/pages/wiki/WikiEdit.vue'), meta: { requireAuth: true } },
      { path: 'wiki/:uri', name: 'WikiDetail', component: () => import('@/pages/wiki/WikiDetail.vue') },
      // Readings
      { path: 'readings', name: 'Readings', component: () => import('@/pages/readings/ReadingList.vue') },
      { path: 'readings/:id', name: 'ReadingDetail', component: () => import('@/pages/readings/ReadingDetail.vue') },
      // User
      { path: 'user/:username', name: 'UserProfile', component: () => import('@/pages/user/UserProfile.vue') },
      { path: 'users', name: 'Users', component: () => import('@/pages/user/UserList.vue') },
      // Account
      { path: 'account/login', name: 'Login', component: () => import('@/pages/account/LoginPage.vue') },
      { path: 'account/register', name: 'Register', component: () => import('@/pages/account/RegisterPage.vue') },
      { path: 'account/edit', name: 'AccountEdit', component: () => import('@/pages/user/UserSettings.vue'), meta: { requireAuth: true } },
      // Messages
      { path: 'message/:msgtype', name: 'Messages', component: () => import('@/pages/messages/MessageList.vue'), meta: { requireAuth: true } },
      { path: 'message/send', name: 'MessageSend', component: () => import('@/pages/messages/MessageSend.vue'), meta: { requireAuth: true } },
      // Search
      { path: 'search', name: 'Search', component: () => import('@/pages/search/SearchPage.vue') },
      // Misc
      { path: 'favorites/:username', name: 'Favorites', component: () => import('@/pages/misc/FavoritesPage.vue') },
      { path: 'mission/daily', name: 'Mission', component: () => import('@/pages/misc/MissionPage.vue'), meta: { requireAuth: true } },
      { path: 'balance', name: 'Balance', component: () => import('@/pages/misc/BalancePage.vue'), meta: { requireAuth: true } },
      { path: 'gift', name: 'Gift', component: () => import('@/pages/misc/GiftPage.vue') },
      { path: 'top/dau', name: 'DauRank', component: () => import('@/pages/misc/RankPage.vue'), props: { type: 'dau' } },
      { path: 'top/rich', name: 'RichRank', component: () => import('@/pages/misc/RankPage.vue'), props: { type: 'rich' } },
      { path: 'nodes', name: 'Nodes', component: () => import('@/pages/topics/NodeList.vue') },
      { path: 'links', name: 'Links', component: () => import('@/pages/misc/LinksPage.vue') },
    ],
  },
  // Admin
  {
    path: '/admin',
    component: () => import('@/pages/admin/AdminLayout.vue'),
    meta: { requireAuth: true, requireAdmin: true },
    children: [
      { path: '', name: 'AdminDashboard', component: () => import('@/pages/admin/Dashboard.vue') },
      { path: 'user/user/list', name: 'AdminUsers', component: () => import('@/pages/admin/UserManage.vue') },
      { path: 'community/topic/list', name: 'AdminTopics', component: () => import('@/pages/admin/TopicManage.vue') },
      { path: 'crawl/article/list', name: 'AdminArticles', component: () => import('@/pages/admin/ArticleManage.vue') },
      { path: 'resource/list', name: 'AdminResources', component: () => import('@/pages/admin/ResourceManage.vue') },
      { path: 'book/list', name: 'AdminBooks', component: () => import('@/pages/admin/BookManage.vue') },
      { path: 'wiki/list', name: 'AdminWikis', component: () => import('@/pages/admin/WikiManage.vue') },
      { path: 'community/node/list', name: 'AdminNodes', component: () => import('@/pages/admin/NodeManage.vue') },
      { path: 'reading/list', name: 'AdminReadings', component: () => import('@/pages/admin/ReadingManage.vue') },
      { path: 'setting', name: 'AdminSetting', component: () => import('@/pages/admin/SettingManage.vue') },
    ],
  },
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/pages/misc/NotFound.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to, _from, next) => {
  if (to.meta.requireAuth || to.meta.requireAdmin) {
    const { useUserStore } = await import('@/stores/user')
    const userStore = useUserStore()

    if (!userStore.me) {
      await userStore.fetchCurrentUser()
    }

    if (!userStore.isLoggedIn) {
      const { useAppStore } = await import('@/stores/app')
      const appStore = useAppStore()
      appStore.openLoginModal()
      return next('/')
    }

    if (to.meta.requireAdmin && !userStore.isAdmin) {
      return next('/')
    }
  }
  next()
})

export default router
