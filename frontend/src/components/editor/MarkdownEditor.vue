<script setup lang="ts">
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <MdEditor
    :model-value="modelValue"
    @update:model-value="(v: string) => emit('update:modelValue', v)"
    language="zh-CN"
    :placeholder="placeholder || '请输入内容（支持 Markdown）'"
    :preview="true"
    style="height: 400px"
  />
</template>
