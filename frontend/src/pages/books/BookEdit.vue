<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NInputNumber, NButton, NSpace, NSwitch, NSpin, useMessage } from 'naive-ui'
import { createBook, modifyBook, getBookDetail } from '@/api/book'

const router = useRouter()
const route = useRoute()
const message = useMessage()

const isEdit = computed(() => !!route.params.id)
const loading = ref(false)
const pageLoading = ref(false)
const form = ref({
  id: 0,
  name: '',
  ename: '',
  cover: '',
  author: '',
  translator: '',
  pub_date: '',
  desc: '',
  tags: '',
  catalogue: '',
  is_free: false,
  online_url: '',
  download_url: '',
  buy_url: '',
  price: 0,
})

onMounted(async () => {
  if (route.params.id) {
    pageLoading.value = true
    try {
      const data = await getBookDetail(Number(route.params.id))
      const book = data?.book
      if (book) {
        form.value = {
          id: book.id,
          name: book.name || '',
          ename: book.ename || '',
          cover: book.cover || '',
          author: book.author || '',
          translator: book.translator || '',
          pub_date: book.pub_date || '',
          desc: book.desc || '',
          tags: book.tags || '',
          catalogue: book.catalogue || '',
          is_free: book.is_free || false,
          online_url: book.online_url || '',
          download_url: book.download_url || '',
          buy_url: book.buy_url || '',
          price: book.price || 0,
        }
      }
    } catch {}
    pageLoading.value = false
  }
})

async function handleSubmit() {
  if (!form.value.name.trim()) { message.warning('请输入书名'); return }
  loading.value = true
  try {
    const payload: Record<string, any> = { ...form.value, is_free: form.value.is_free ? '1' : '0' }
    if (isEdit.value) {
      payload.id = String(form.value.id)
      await modifyBook(payload)
      message.success('修改成功')
    } else {
      delete payload.id
      await createBook(payload)
      message.success('创建成功')
    }
    router.push('/books')
  } catch (e: any) { message.error(e.message || '操作失败') }
  loading.value = false
}
</script>

<template>
  <NSpin :show="pageLoading">
    <NCard :title="isEdit ? '编辑图书' : '推荐图书'" style="max-width: 800px; margin: 0 auto">
      <NForm label-placement="top">
        <NFormItem label="书名" required>
          <NInput v-model:value="form.name" placeholder="图书名称" />
        </NFormItem>
        <NFormItem label="英文名">
          <NInput v-model:value="form.ename" placeholder="英文书名（可选）" />
        </NFormItem>
        <NFormItem label="封面图片">
          <NInput v-model:value="form.cover" placeholder="封面图片 URL" />
        </NFormItem>
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 0 16px">
          <NFormItem label="作者">
            <NInput v-model:value="form.author" placeholder="作者" />
          </NFormItem>
          <NFormItem label="译者">
            <NInput v-model:value="form.translator" placeholder="译者（可选）" />
          </NFormItem>
          <NFormItem label="出版日期">
            <NInput v-model:value="form.pub_date" placeholder="如：2024-01" />
          </NFormItem>
          <NFormItem label="价格">
            <NInputNumber v-model:value="form.price" :min="0" :precision="2" placeholder="0 表示免费" style="width: 100%" />
          </NFormItem>
        </div>
        <NFormItem label="简介">
          <NInput v-model:value="form.desc" type="textarea" placeholder="图书简介" :rows="6" />
        </NFormItem>
        <NFormItem label="目录">
          <NInput v-model:value="form.catalogue" type="textarea" placeholder="图书目录（可选）" :rows="4" />
        </NFormItem>
        <NFormItem label="标签">
          <NInput v-model:value="form.tags" placeholder="多个标签用逗号分隔" />
        </NFormItem>
        <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0 16px">
          <NFormItem label="在线阅读 URL">
            <NInput v-model:value="form.online_url" placeholder="在线阅读地址" />
          </NFormItem>
          <NFormItem label="下载 URL">
            <NInput v-model:value="form.download_url" placeholder="下载地址" />
          </NFormItem>
          <NFormItem label="购买 URL">
            <NInput v-model:value="form.buy_url" placeholder="购买地址" />
          </NFormItem>
        </div>
        <NFormItem label="免费">
          <NSwitch v-model:value="form.is_free" />
        </NFormItem>
        <NSpace justify="end">
          <NButton @click="router.back()">取消</NButton>
          <NButton type="primary" :loading="loading" @click="handleSubmit">{{ isEdit ? '保存' : '提交' }}</NButton>
        </NSpace>
      </NForm>
    </NCard>
  </NSpin>
</template>
