<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { register } from '@/api/user'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const message = useMessage()

const form = ref({ username: '', email: '', passwd: '', passwd2: '' })
const loading = ref(false)

async function handleRegister() {
  if (!form.value.username || !form.value.email || !form.value.passwd) {
    message.warning('请填写完整信息')
    return
  }
  if (form.value.passwd !== form.value.passwd2) {
    message.warning('两次密码不一致')
    return
  }
  loading.value = true
  try {
    await register({ username: form.value.username, email: form.value.email, passwd: form.value.passwd })
    message.success('注册成功')
    await userStore.fetchCurrentUser()
    router.push('/')
  } catch (e: any) {
    message.error(e.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div style="max-width: 480px; margin: 40px auto">
    <NCard title="注册">
      <NForm @submit.prevent="handleRegister">
        <NFormItem label="用户名" required>
          <NInput v-model:value="form.username" placeholder="请输入用户名" />
        </NFormItem>
        <NFormItem label="邮箱" required>
          <NInput v-model:value="form.email" placeholder="请输入邮箱" />
        </NFormItem>
        <NFormItem label="密码" required>
          <NInput v-model:value="form.passwd" type="password" placeholder="请输入密码" show-password-on="click" />
        </NFormItem>
        <NFormItem label="确认密码" required>
          <NInput v-model:value="form.passwd2" type="password" placeholder="请再次输入密码" show-password-on="click" />
        </NFormItem>
        <NButton type="primary" block :loading="loading" attr-type="submit" style="margin-top: 8px">注册</NButton>
      </NForm>
      <div style="text-align: center; margin-top: 16px">
        已有账号？<router-link to="/account/login">去登录</router-link>
      </div>
    </NCard>
  </div>
</template>
