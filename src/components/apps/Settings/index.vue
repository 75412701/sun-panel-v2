<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NButton, NCard, NTag } from 'naive-ui'
import { t } from '@/locales'

// 自动判断内外网测试
const testStatus = ref<'idle' | 'testing' | 'success_lan' | 'success_wan' | 'failed'>('idle')
const isTesting = ref(false)
const clientIp = ref('')
const serverPublicIp = ref('')

// 测试连接与内外网识别
async function testConnection() {
  isTesting.value = true
  testStatus.value = 'testing'

  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 3000)

    const response = await fetch('/ping', {
      method: 'GET',
      signal: controller.signal,
    })

    clearTimeout(timeoutId)

    if (response.ok) {
      const res = await response.json()
      if (res.code === 0 && res.data) {
        clientIp.value = res.data.clientIp || ''
        serverPublicIp.value = res.data.serverPublicIp || ''
        if (res.data.isLan) {
          testStatus.value = 'success_lan'
        } else {
          testStatus.value = 'success_wan'
        }
        return
      }
    }
    testStatus.value = 'failed'
  } catch (error: any) {
    testStatus.value = 'failed'
  } finally {
    isTesting.value = false
  }
}

onMounted(() => {
  testConnection()
})
</script>

<template>
  <div class="bg-slate-200 dark:bg-zinc-900 rounded-[10px] p-[8px] overflow-auto">
    <!-- 智能内外网检测设置 -->
    <NCard style="border-radius:10px" size="small">
      <div class="text-slate-500 mb-[5px] font-bold text-base">
        {{ t('apps.settings.networkDetection') }}
      </div>

      <div class="mt-[10px] text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
        {{ t('apps.settings.networkDetectionDesc') }}
      </div>

      <!-- 测试连接按钮及结果 -->
      <div class="flex flex-wrap items-center gap-3 mt-[15px]">
        <NButton
          size="small"
          type="primary"
          :loading="isTesting"
          @click="testConnection"
        >
          {{ t('apps.settings.testConnection') }}
        </NButton>

        <!-- 状态显示 -->
        <NTag
          v-if="testStatus === 'success_lan'"
          type="success"
          size="small"
          round
        >
          ✓ {{ t('apps.settings.connectionSuccess') }}
        </NTag>
        <NTag
          v-else-if="testStatus === 'success_wan'"
          type="info"
          size="small"
          round
        >
          🌐 {{ t('apps.settings.connectionWan') }}
        </NTag>
        <NTag
          v-else-if="testStatus === 'failed'"
          type="error"
          size="small"
          round
        >
          ✗ {{ t('apps.settings.connectionFailed') }}
        </NTag>
        <span
          v-else-if="testStatus === 'testing'"
          class="text-sm text-gray-600 dark:text-gray-400"
        >
          {{ t('apps.settings.testing') }}
        </span>
      </div>

      <!-- 诊断详情 -->
      <div v-if="clientIp || serverPublicIp" class="mt-[12px] p-[10px] bg-gray-100 dark:bg-zinc-800 rounded-lg text-xs space-y-1 text-gray-700 dark:text-gray-300">
        <div v-if="clientIp">
          <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('apps.settings.clientIp') }}:</span>
          <span class="ml-2 font-mono">{{ clientIp }}</span>
        </div>
        <div v-if="serverPublicIp">
          <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('apps.settings.serverPublicIp') }}:</span>
          <span class="ml-2 font-mono">{{ serverPublicIp }}</span>
        </div>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.text-shadow {
  text-shadow: 0px 0px 5px gray;
}
</style>
