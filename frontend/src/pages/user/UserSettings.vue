<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NForm, NFormItem, NInput, NButton, NAvatar, NSpace, useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { modifyUser, changePassword } from '@/api/user'

const userStore = useUserStore()
const message = useMessage()

const form = ref({ name: '', city: '', company: '', github: '', website: '', introduce: '' })
const pwdForm = ref({ cur_passwd: '', passwd: '', passwd2: '' })
const loading = ref(false)
const pwdLoading = ref(false)

onMounted(() => {
  if (userStore.me) {
    form.value = { name: (userStore.me as any).name || '', city: '', company: '', github: '', website: '', introduce: '' }
  }
})

async function handleSave() {
  loading.value = true
  try { await modifyUser(form.value); message.success('保存成功'); userStore.fetchCurrentUser() } catch (e: any) { message.error(e.message || '保存失败') } finally { loading.value = false }
}

async function handleChangePwd() {
  if (pwdForm.value.passwd !== pwdForm.value.passwd2) { message.warning('两次密码不一致'); return }
  pwdLoading.value = true
  try { await changePassword({ cur_passwd: pwdForm.value.cur_passwd, passwd: pwdForm.value.passwd }); message.success('密码修改成功'); pwdForm.value = { cur_passwd: '', passwd: '', passwd2: '' } } catch (e: any) { message.error(e.message || '修改失败') } finally { pwdLoading.value = false }
}
</script>

<template>
  <div>
    <NCard title="个人资料">
      <NSpace align="center" style="margin-bottom: 24px">
        <NAvatar :src="userStore.me?.avatar" :size="64" round />
        <NButton size="small">更换头像</NButton>
      </NSpace>
      <NForm label-placement="left" label-width="80">
        <NFormItem label="昵称"><NInput v-model:value="form.name" /></NFormItem>
        <NFormItem label="城市"><NInput v-model:value="form.city" /></NFormItem>
        <NFormItem label="公司"><NInput v-model:value="form.company" /></NFormItem>
        <NFormItem label="GitHub"><NInput v-model:value="form.github" /></NFormItem>
        <NFormItem label="个人网站"><NInput v-model:value="form.website" /></NFormItem>
        <NFormItem label="简介"><NInput v-model:value="form.introduce" type="textarea" :rows="3" /></NFormItem>
        <NFormItem><NButton type="primary" :loading="loading" @click="handleSave">保存</NButton></NFormItem>
      </NForm>
    </NCard>

    <NCard title="修改密码" style="margin-top: 16px">
      <NForm label-placement="left" label-width="80">
        <NFormItem label="当前密码"><NInput v-model:value="pwdForm.cur_passwd" type="password" show-password-on="click" /></NFormItem>
        <NFormItem label="新密码"><NInput v-model:value="pwdForm.passwd" type="password" show-password-on="click" /></NFormItem>
        <NFormItem label="确认密码"><NInput v-model:value="pwdForm.passwd2" type="password" show-password-on="click" /></NFormItem>
        <NFormItem><NButton type="primary" :loading="pwdLoading" @click="handleChangePwd">修改密码</NButton></NFormItem>
      </NForm>
    </NCard>
  </div>
</template>
