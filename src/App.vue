<script setup lang="ts">
import { NConfigProvider } from 'naive-ui'
import { NaiveProvider } from '@/components/common'
import { useTheme } from '@/hooks/useTheme'
import { useLanguage } from '@/hooks/useLanguage'
import { onMounted } from 'vue'

const { theme, themeOverrides } = useTheme()
const { language } = useLanguage()

// 在组件挂载时动态加载一言脚本
onMounted(() => {
  const script = document.createElement('script')
  script.src = 'https://v1.hitokoto.cn/?encode=js&select=%23hitokoto'
  script.defer = true
  document.body.appendChild(script)
})
</script>

<template>
  <NConfigProvider
    class="h-full"
    :theme="theme"
    :theme-overrides="themeOverrides"
    :locale="language"
  >
    <NaiveProvider>
      <RouterView />
    </NaiveProvider>
  </NConfigProvider>
</template>
