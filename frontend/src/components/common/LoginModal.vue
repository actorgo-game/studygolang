<script setup lang="ts">
import { ref } from 'vue'
import { NModal, NCard, NForm, NFormItem, NInput, NButton, NCheckbox, NSpace, useMessage } from 'naive-ui'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'

const appStore = useAppStore()
const userStore = useUserStore()
const message = useMessage()

const form = ref({ username: '', passwd: '', remember_me: false })
const loading = ref(false)

async function handleLogin() {
  if (!form.value.username || !form.value.passwd) {
    message.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await userStore.login(form.value.username, form.value.passwd, form.value.remember_me)
    message.success('登录成功')
    appStore.closeLoginModal()
    form.value = { username: '', passwd: '', remember_me: false }
  } catch (e: any) {
    message.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <NModal v-model:show="appStore.showLoginModal" :mask-closable="true">
    <NCard title="登录" style="width: 400px" :bordered="false" closable @close="appStore.closeLoginModal()">
      <NForm @submit.prevent="handleLogin">
        <NFormItem label="用户名/邮箱">
          <NInput v-model:value="form.username" placeholder="请输入用户名或邮箱" />
        </NFormItem>
        <NFormItem label="密码">
          <NInput v-model:value="form.passwd" type="password" placeholder="请输入密码" show-password-on="click" />
        </NFormItem>
        <NSpace justify="space-between" align="center">
          <NCheckbox v-model:checked="form.remember_me">记住登录</NCheckbox>
          <router-link to="/account/forgetpwd" @click="appStore.closeLoginModal()">忘记密码？</router-link>
        </NSpace>
        <NButton type="primary" block :loading="loading" attr-type="submit" style="margin-top: 16px">
          登录
        </NButton>
      </NForm>
      <div style="text-align: center; margin-top: 16px">
        还没有账号？
        <router-link to="/account/register" @click="appStore.closeLoginModal()">立即注册</router-link>
      </div>
    </NCard>
  </NModal>
</template>
