import { defineStore } from 'pinia'
import type { ModuleConfigState } from './helper'
import { getLocalState, setLocalState } from './helper'
import { getValueByName, save } from '@/api/system/moduleConfig'
import { ss } from '@/utils/storage'

export const useModuleConfig = defineStore('module-config-store', {
  state: (): ModuleConfigState => getLocalState(),
  actions: {

    // 获取值（带缓存）
    async getValueByNameFromCloud<T>(name: string) {
      const moduleName = `module-${name}`
      const cacheKey = `moduleConfig_${moduleName}`

      try {
        // 1. 首先尝试从缓存读取数据
        const cachedData = ss.get(cacheKey)
        if (cachedData) {
          return cachedData
        }

        // 2. 缓存中没有数据，请求接口获取数据
        const response = await getValueByName<T>(moduleName)

        // 3. 将数据永久保存到缓存中
        ss.set(cacheKey, response)

        return response
      } catch (error) {
        console.error(`获取模块配置失败: ${name}`, error)
        // 如果出错，尝试从缓存获取
        const cachedData = ss.get(cacheKey)
        if (cachedData) {
          return cachedData
        }
        // 如果缓存也没有，抛出错误
        throw error
      }
    },

    // 保存到网络
    async saveToCloud(name: string, value: any) {
      const moduleName = `module-${name}`
      const cacheKey = `moduleConfig_${moduleName}`

      // 保存至网络
      const response = await save(moduleName, value)

      // 如果保存成功，清除对应缓存，下次获取时重新缓存
      if (response.code === 0) {
        ss.remove(cacheKey)
      }

      return response
    },

    // 从网络同步
    // syncFromCloud(moduleName: string) {
    //   getValueByName<any>(moduleName).then(({ code, data, msg }) => {
    //     if (code === 0)
    //       this.$state[moduleName] = data
    //   })
    // },

    recordState() {
      setLocalState(this.$state)
    },
  },
})
