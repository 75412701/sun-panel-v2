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
      <!-- 固定底部的一言区域 -->
      <div class="foot" style="position: fixed; bottom: 0; left: 0; right: 0;  color: white; text-align: center; padding: 10px; z-index: 9999;">
        <p id="hitokoto">:D 获取中...</p>
      </div>
    </NaiveProvider>
  </NConfigProvider>
</template>
