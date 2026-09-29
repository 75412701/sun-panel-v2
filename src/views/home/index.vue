<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, NButtonGroup, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage } from 'naive-ui'
import { nextTick, onMounted, onActivated, ref, h } from 'vue'
import { AppIcon, AppStarter, EditItem, NotePad } from './components'
import { Clock, SystemMonitor } from '@/components/deskModule'
import SearchBoxWithSuggestions from '@/components/deskModule/SearchBoxWithSuggestions/index.vue'
import { SvgIcon } from '@/components/common'
import { deletes, getListByGroupId, saveSort } from '@/api/panel/itemIcon'

import { setTitle, updateLocalUserInfo, openUrlWithoutReferer } from '@/utils/cmn'
import { useAuthStore, usePanelState } from '@/store'
import { PanelPanelConfigStyleEnum, PanelStateNetworkModeEnum } from '@/enums'
import { VisitMode } from '@/enums/auth'
import { router } from '@/router'
import { onBeforeRouteUpdate } from 'vue-router'
import { t } from '@/locales'
import {  computed } from "vue"
import { useWindowSize } from "@vueuse/core"
import { NDrawer, NDrawerContent, NTree,  } from "naive-ui"
interface ItemGroup extends Panel.ItemIconGroup {
  sortStatus?: boolean
  hoverStatus: boolean
  items?: Panel.ItemInfo[]
}

const ms = useMessage()
const dialog = useDialog()
const panelState = usePanelState()
const authStore = useAuthStore()


const scrollContainerRef = ref<HTMLElement | undefined>(undefined)

const editItemInfoShow = ref<boolean>(false)
const editItemInfoData = ref<Panel.ItemInfo | null>(null)
const windowShow = ref<boolean>(false)
const windowSrc = ref<string>('')
const windowTitle = ref<string>('')

const windowIframeRef = ref(null)
const windowIframeIsLoad = ref<boolean>(false)

const dropdownMenuX = ref(0)
const dropdownMenuY = ref(0)
const dropdownShow = ref(false)
const currentRightSelectItem = ref<Panel.ItemInfo | null>(null)
const currentAddItenIconGroupId = ref<number | undefined>()
// 刷新函数 - 删除缓存并重新加载数据
async function handleRefreshData() {
  try {
    // 删除除用户登录信息外的所有缓存
    ss.remove(BOOKMARKS_CACHE_KEY)
    ss.remove(GROUP_LIST_CACHE_KEY)
    ss.remove('searchEngineListCache')

    // 直接清除所有localStorage中的图标列表缓存
    // 由于ss没有getAllKeys方法，我们直接使用原生localStorage API
    Object.keys(localStorage).forEach(key => {
      if (key.startsWith(ITEM_ICON_LIST_CACHE_KEY_PREFIX)) {
        ss.remove(key)
      }
    })

    // 重新加载数据
    getList()
    // 调用loadBookmarkTree并传入true参数以强制刷新
    await loadBookmarkTree(true)

    // 刷新便签数据（仅登录状态下）
    if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN && notepadInstance.value) {
        // @ts-ignore
        notepadInstance.value.refreshData?.()
    }

    // 检查是否需要重新获取网络壁纸
    if (panelState.panelConfig.autoNetworkWallpaper) {
      try {
        // 添加时间戳参数，确保每次请求的URL不同，避免浏览器缓存
        const timestamp = new Date().getTime()
        const baseUrl = panelState.panelConfig.autoNetworkWallpaperApi || 'https://img.xjh.me/random_img.php?return=302&type=bg&ctype=nature'
        const apiUrl = baseUrl.includes('?') ? `${baseUrl}&t=${timestamp}` : `${baseUrl}?t=${timestamp}`
        panelState.panelConfig.backgroundImageSrc = apiUrl
        panelState.recordState()
      } catch (error) {
        console.error('重新获取网络壁纸失败', error)
      }
    }

    ms.success(t('common.refreshSuccess'))
  } catch (error) {
    console.error('刷新数据失败:', error)
    ms.error(t('common.refreshFailed'))
  }
}
// 1. 定义树形节点的接口（包含 children）
interface TreeItem {
	key: string | number;
	label: string;
	isLeaf: boolean;
	bookmark?: {
		id: number;
		title: string;
		url: string;
		folderId: string | null;
	};
	children?: TreeItem[]; // 补充 children 属性（可选，叶子节点没有）
}

const settingModalShow = ref(false)

const items = ref<ItemGroup[]>([])
const filterItems = ref<ItemGroup[]>([])


const drawerVisible = ref(false)
const notepadVisible = ref(false)
const notepadInstance = ref(null) // 便签实例
const { width } = useWindowSize()

// 移动端判断
const isMobile = computed(() => width.value < 768)

// 从API导入获取书签列表的函数
import { getList as getBookmarksList } from '@/api/panel/bookmark'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'
import { ss } from '@/utils/storage/local'


// 书签数据树
const treeData = ref<any[]>([])
// 缓存键名
const BOOKMARKS_CACHE_KEY = 'bookmarksTreeCache'
const GROUP_LIST_CACHE_KEY = 'groupListCache'
// 图标列表缓存键前缀
const ITEM_ICON_LIST_CACHE_KEY_PREFIX = 'itemIconList_'

// 智能内外网检测缓存 (有效期 10 秒)
let intranetCheckCache: { isLan: boolean; timestamp: number } | null = null

// 检测内网连接
async function checkIntranetConnection(): Promise<boolean> {
  const now = Date.now()
  if (intranetCheckCache && now - intranetCheckCache.timestamp < 10000) {
    return intranetCheckCache.isLan
  }

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 1000)

  try {
    const response = await fetch('/ping', {
      method: 'GET',
      signal: controller.signal,
    })
    clearTimeout(timeoutId)
    if (response.ok) {
      const res = await response.json()
      const isLan = !!(res?.data?.isLan ?? false)
      intranetCheckCache = { isLan, timestamp: Date.now() }
      return isLan
    }
  } catch (e) {
    // 忽略异常，降级为非内网
  }
  return false
}


// 获取书签数据并转换为前端需要的格式
async function loadBookmarkTree(forceRefresh = false) {
  try {
    // 如果不是强制刷新且缓存存在，则使用缓存
    if (!forceRefresh) {
      const cachedData = ss.get(BOOKMARKS_CACHE_KEY)
      if (cachedData) {
        // 处理缓存的原始fullData格式数据
        let treeDataResult = [];

        // 检查是否已经是树形结构（直接包含children字段）
        if (Array.isArray(cachedData) && cachedData.length > 0 && 'children' in cachedData[0]) {
          treeDataResult = convertServerTreeToFrontendTree(cachedData)
        } else if (cachedData.list && Array.isArray(cachedData.list)) {
          // 后端返回的是带list字段的结构
          const serverBookmarks = cachedData.list
          if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
            treeDataResult = convertServerTreeToFrontendTree(serverBookmarks)
          } else {
            treeDataResult = buildBookmarkTree(serverBookmarks)
          }
        } else {
          treeDataResult = buildBookmarkTree(Array.isArray(cachedData) ? cachedData : [])
        }

        treeData.value = treeDataResult
        return
      }
    } else {
      // 强制刷新时清除缓存
      ss.remove(BOOKMARKS_CACHE_KEY)
    }
    const response = await getBookmarksList()
    if (response.code === 0) {
      // 检查数据结构
      const data: any = response.data || []
      let treeDataResult = []

      // 检查是否已经是树形结构（直接包含children字段）
      if (Array.isArray(data) && data.length > 0 && 'children' in data[0]) {
        treeDataResult = convertServerTreeToFrontendTree(data)
      } else if (data.list && Array.isArray(data.list)) {
        // 后端返回的是带list字段的结构
        const serverBookmarks = data.list
        if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
          treeDataResult = convertServerTreeToFrontendTree(serverBookmarks)
        } else {
          treeDataResult = buildBookmarkTree(serverBookmarks)
        }
      } else {
        treeDataResult = buildBookmarkTree(Array.isArray(data) ? data : [])
      }

      // 更新treeData
      treeData.value = treeDataResult
      ss.set(BOOKMARKS_CACHE_KEY, data)
    }
  } catch (error) {
    console.error('获取书签数据失败:', error)
    // 出错时尝试使用缓存
    const cachedData = ss.get(BOOKMARKS_CACHE_KEY)
    if (cachedData) {
      // 处理缓存的原始fullData格式数据
      let treeDataResult = [];

      // 检查是否已经是树形结构（直接包含children字段）
      if (Array.isArray(cachedData) && cachedData.length > 0 && 'children' in cachedData[0]) {
        treeDataResult = convertServerTreeToFrontendTree(cachedData)
      } else if (cachedData.list && Array.isArray(cachedData.list)) {
        // 后端返回的是带list字段的结构
        const serverBookmarks = cachedData.list
        if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
          treeDataResult = convertServerTreeToFrontendTree(serverBookmarks)
        } else {
          treeDataResult = buildBookmarkTree(serverBookmarks)
        }
      } else {
        treeDataResult = buildBookmarkTree(Array.isArray(cachedData) ? cachedData : [])
      }
      treeData.value = treeDataResult
    }
  }
}

// 将服务器返回的树形结构转换为前端组件需要的格式
function convertServerTreeToFrontendTree(serverTree: any[]): any[] {
  // 先对顶层节点按sort字段排序
  const sortedServerTree = [...serverTree].sort((a, b) => (a.sort || 0) - (b.sort || 0));
  const result = sortedServerTree.map(node => {
    // 处理两种可能的节点结构：
    // 1. 服务器原始数据格式 (id, title, isFolder, url, iconJson)
    // 2. 前端节点格式 (key, label, isFolder, bookmark)
    const isFrontendFormat = node.hasOwnProperty('key') && node.hasOwnProperty('label');

    // 提取基本属性
    const nodeId = isFrontendFormat ? node.key : node.id;
    const title = isFrontendFormat ? node.label : node.title;
    const isFolder = isFrontendFormat ? (node.isFolder ? 1 : 0) : node.isFolder;
    const url = isFrontendFormat ? (node.bookmark?.url || '') : node.url;
    const iconJson = isFrontendFormat ? (node.bookmark?.iconJson || '') : node.iconJson;
    const parentId = isFrontendFormat ? (node.rawNode?.parentId || node.ParentId || '0') : (node.parentId || node.ParentId || '0');

    // 提取排序字段
    const sortOrder = node.sort || 0;

    // 处理bookmark对象
    let bookmarkObj = undefined;
    if (isFolder !== 1 && url) {
      // 确保folderId是字符串类型
      const folderId = parentId !== undefined ? String(parentId) : null;
      bookmarkObj = {
        id: nodeId,
        title: title,
        url: url,
        folderId: folderId,
        iconJson: iconJson, // 保存base64图标数据
        sort: sortOrder // 保存排序字段到书签对象
      };
    }

    const frontendNode = {
        key: nodeId,
        label: title || '未命名',
        isLeaf: isFolder !== 1,
        isFolder: isFolder === 1, // 添加isFolder属性
        sort: sortOrder, // 保存排序字段到前端节点
        bookmark: bookmarkObj
    };

    // 递归处理子节点
    if (node.children && node.children.length > 0) {
      // 对子节点先按sort字段排序再递归转换
      const sortedChildren = [...node.children].sort((a, b) => (a.sort || 0) - (b.sort || 0));
      (frontendNode as TreeItem).children = convertServerTreeToFrontendTree(sortedChildren);
    }

    return frontendNode;
  });

  return result;
}

// 构建书签树
function buildBookmarkTree(bookmarks: any[]): any[] {
  // 首先分离文件夹和书签
  const folders = bookmarks.filter(b => {
    return (b.isFolder === 1 || (b.isFolder && typeof b.isFolder === 'boolean'));
  });
  const items = bookmarks.filter(b => {
    return (b.isFolder === 0 || (!b.isFolder && typeof b.isFolder === 'boolean'));
  });

  // 构建文件夹树
  const rootFolders: any[] = []
  const folderMap = new Map<string, any>() // 使用字符串键

  // 先创建所有文件夹节点
  folders.forEach(folder => {
    // 处理两种可能的文件夹结构
    const isFrontendFormat = folder.hasOwnProperty('key') && folder.hasOwnProperty('label');
    const folderId = isFrontendFormat ? folder.key : folder.id;
    const folderTitle = isFrontendFormat ? folder.label : folder.title;
    const folderSort = folder.sort || 0;
    const folderNode = {
      key: folderId,
      label: folderTitle,
      children: [],
      isFolder: true,
      sort: folderSort // 保存排序字段
    };
    // 使用id作为map的键
    folderMap.set(folderId.toString(), folderNode);
    // 同时也将文件夹名称作为键，以便处理嵌套关系
    folderMap.set(folderTitle, folderNode);
  });

  // 将文件夹添加到其父文件夹中
  folders.forEach(folder => {
    const folderNode = folderMap.get(folder.id.toString())
    // 检查是否有ParentUrl并且不是根节点(0)
    if (folder.ParentUrl && folder.ParentUrl !== '0' && folder.ParentUrl !== 0) {
      // 尝试用不同的方式查找父文件夹
      let parentFolder = folderMap.get(folder.ParentUrl.toString())

      if (!parentFolder) {
        // 如果找不到，尝试用文件夹标题匹配
        parentFolder = folderMap.get(folder.ParentUrl)
      }

      if (parentFolder) {
        parentFolder.children.push(folderNode)
        return
      }
    }
    // 如果没有父文件夹或父文件夹不存在，则作为根文件夹
    rootFolders.push(folderNode)
  })

  // 将书签项添加到对应的文件夹中
  items.forEach(item => {
    // 处理两种可能的书签结构
    const isFrontendFormat = item.hasOwnProperty('key') && item.hasOwnProperty('label');
    // 提取书签基本信息
    const bookmarkId = isFrontendFormat ? item.key : item.id;
    const bookmarkTitle = isFrontendFormat ? item.label : (item.title || '未命名');
    const bookmarkUrl = isFrontendFormat ? (item.bookmark?.url || '') : (item.url || '');
    const bookmarkIconJson = isFrontendFormat ? (item.bookmark?.iconJson || '') : (item.iconJson || '');
    // 确保folderId是字符串类型
    const folderId = isFrontendFormat ? (item.rawNode?.parentId || item.ParentId || '0') : (item.parentId || item.ParentId || '0');
    const stringFolderId = String(folderId);
    // 获取排序字段
    const sortOrder = isFrontendFormat ? (item.rawNode?.sort || 0) : (item.sort || 0);

    let targetFolder;

    if (stringFolderId === '0' || stringFolderId === 'null' || stringFolderId === 'undefined') {
      // 根目录的书签，创建一个"未分类"文件夹
      targetFolder = folderMap.get('未分类');
      if (!targetFolder) {
        targetFolder = {
          key: '未分类',
          label: '未分类',
          children: [],
          isFolder: true,
          sort: 0 // 设置默认排序
        };
        folderMap.set('未分类', targetFolder);
        rootFolders.push(targetFolder);
      }
    } else {
      // 查找对应的文件夹
      targetFolder = folderMap.get(stringFolderId);
    }

    if (targetFolder) {
      // 创建书签节点
      const bookmarkNode = {
        key: bookmarkId,
        label: bookmarkTitle,
        isLeaf: true,
        sort: sortOrder, // 保存排序字段
        bookmark: {
          id: bookmarkId,
          title: bookmarkTitle,
          url: bookmarkUrl,
          folderId: stringFolderId,
          iconJson: bookmarkIconJson
        }
      };
      targetFolder.children.push(bookmarkNode);
    }
  })

  // 递归排序所有节点的子节点
  function sortTreeNodes(nodes: any[]) {
    nodes.sort((a, b) => (a.sort || 0) - (b.sort || 0));
    nodes.forEach(node => {
      if (node.isFolder && node.children) {
        sortTreeNodes(node.children);
      }
    });
  }

  sortTreeNodes(rootFolders);

  return rootFolders
}



function openPage(openMethod: number, url: string, title?: string) {
  switch (openMethod) {
    case 1:
      window.location.replace(url)
      break
    case 2:
      openUrlWithoutReferer(url, '_blank')
      break
    case 3:
      windowShow.value = true
      windowSrc.value = url
      windowTitle.value = title || url
      windowIframeIsLoad.value = true
      break

    default:
      break
  }
}

async function handleItemClick(itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus) {
    handleEditItem(item)
    return
  }

  // 辅助函数：标准化URL（自动添加http://）
  const normalizeUrl = (url: string | undefined | null) => {
    if (!url) return ''
    let trimmed = url.trim()
    if (!trimmed) return ''

    // 如果是 javascript: 等特殊协议或已经是 http/https 开头，或者是相对路径，则不处理
    if (/^[a-z]+:/i.test(trimmed) || trimmed.startsWith('/') || trimmed.startsWith('./') || trimmed.startsWith('../')) {
      return trimmed
    }

    // 默认为 http
    return 'http://' + trimmed
  }

  // 辅助函数：检查URL是否有效
  const isValidUrl = (url: string | undefined | null) => {
    if (!url) return false
    const trimmed = url.trim()
    return trimmed !== '' && trimmed !== 'null' && trimmed !== 'undefined'
  }

  // 预处理 URL
  const publicUrl = normalizeUrl(item.url)
  const lanUrl = normalizeUrl(item.lanUrl)

  // 默认使用公网地址
  let jumpUrl = publicUrl

  // 检查是否需要进行内网探测
  // 条件：有内网地址 AND 内网地址有效
  const shouldCheckIntranet = isValidUrl(item.lanUrl)

  if (shouldCheckIntranet) {
    // 情况1：新窗口打开 (openMethod === 2)
    // 需要先打开空白窗口避开拦截，然后异步探测
    if (item.openMethod === 2) {
      const newWindow = window.open('about:blank', '_blank')
      if (newWindow) {
        // 探测
        const isIntranet = await checkIntranetConnection()

        // 确定最终URL
        let finalUrl = publicUrl
        if (isIntranet && isValidUrl(item.lanUrl)) {
          finalUrl = lanUrl
        }

        newWindow.location.href = finalUrl
        return // 结束，不执行后面的 openPage
      }
    }

    // 情况2：当前窗口或弹窗 (openMethod === 1, 3等)
    const isIntranet = await checkIntranetConnection()
    if (isIntranet && isValidUrl(item.lanUrl)) {
      jumpUrl = lanUrl
    }
  }

  // 执行打开页面 (如果是新窗口且上面处理过了，这里就不会执行)
  openPage(item.openMethod, jumpUrl, item.title)
}

function handWindowIframeIdLoad(payload: Event) {
  windowIframeIsLoad.value = false
}

function navigateToBookmarkManager() {
  // 跳转到书签管理页面
  router.push('/bookmark-manager')
}

// 处理树节点选择事件
function handleTreeSelect(keys: (string | number)[]) {
  if (keys && keys.length > 0) {
    // 查找选中的节点
    const selectedKey = keys[0]
    const selectedNode = findNodeByKey(treeData.value, selectedKey)

    // 如果是书签项（非文件夹），则打开链接
    if (selectedNode && selectedNode.isLeaf && selectedNode.bookmark && selectedNode.bookmark.url) {
      openUrlWithoutReferer(selectedNode.bookmark.url, '_blank')
    }
  }
}

// 根据key查找节点
function findNodeByKey(nodes: any[], key: string | number): any {
  for (const node of nodes) {
    if (node.key === key) {
      return node
    }
    if (node.children && node.children.length > 0) {
      const found = findNodeByKey(node.children, key)
      if (found) {
        return found
      }
    }
  }
  return null
}

// 根据网络模式过滤项目
function filterItemsByNetworkMode() {
  // 只有在公网模式下才需要过滤
  if (panelState.networkMode === PanelStateNetworkModeEnum.wan) {
    const filteredGroups = items.value.map(group => {
      if (group.items) {
        // 过滤掉lanOnly为1的项目
        const filteredItems = group.items.filter(item => item.lanOnly !== 1)
        return { ...group, items: filteredItems }
      }
      return group
    })
    // 过滤掉没有项目的组
    filterItems.value = filteredGroups.filter(group => !group.items || group.items.length > 0)
  } else {
    // 私密模式下显示所有项目
    filterItems.value = items.value
  }
}

async function getList() {
  try {
    // 1. 首先尝试从缓存读取数据
    const cachedData = ss.get(GROUP_LIST_CACHE_KEY)
    if (cachedData) {
      items.value = cachedData
      // 为每个分组加载图标数据
      for (let i = 0; i < cachedData.length; i++) {
        const element = cachedData[i]
        if (element.id)
          updateItemIconGroupByNet(i, element.id)
      }
      // 应用网络模式过滤
      filterItemsByNetworkMode()
      return
    }

    // 2. 缓存中没有数据，请求接口获取数据
    const response = await getGroupList<Common.ListResponse<ItemGroup[]>>()
    if (response.code === 0) {
      items.value = response.data.list
      // 3. 将数据永久保存到缓存中
      ss.set(GROUP_LIST_CACHE_KEY, response.data.list)

      // 为每个分组加载图标数据
      for (let i = 0; i < response.data.list.length; i++) {
        const element = response.data.list[i]
        if (element.id)
          updateItemIconGroupByNet(i, element.id)
      }
      // 应用网络模式过滤
      filterItemsByNetworkMode()
    }
  } catch (error) {
    // 出错时尝试从缓存获取
    const cachedData = ss.get(GROUP_LIST_CACHE_KEY)
    if (cachedData) {
      items.value = cachedData
      // 为每个分组加载图标数据
      for (let i = 0; i < cachedData.length; i++) {
        const element = cachedData[i]
        if (element.id)
          updateItemIconGroupByNet(i, element.id)
      }
      // 应用网络模式过滤
      filterItemsByNetworkMode()
    }
  }
}

// 从后端获取组下面的图标
async function updateItemIconGroupByNet(itemIconGroupIndex: number, itemIconGroupId: number) {
  try {
    // 1. 定义缓存键
    const cacheKey = `${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemIconGroupId}`

    // 2. 首先尝试从缓存读取数据
    const cachedData = ss.get(cacheKey)
    if (cachedData) {
      items.value[itemIconGroupIndex].items = cachedData

      // 当所有组的数据都加载完成后，应用网络模式过滤
      const allGroupsLoaded = items.value.every(group => group.items !== undefined)
      if (allGroupsLoaded) {
        filterItemsByNetworkMode()
      }
      return
    }
    const res = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(itemIconGroupId)

    if (res.code === 0) {
      items.value[itemIconGroupIndex].items = res.data.list
      // 4. 将数据永久保存到缓存中
      ss.set(cacheKey, res.data.list)

      // 当所有组的数据都加载完成后，应用网络模式过滤
      const allGroupsLoaded = items.value.every(group => group.items !== undefined)
      if (allGroupsLoaded) {
        filterItemsByNetworkMode()
      }
    }
  } catch (error) {
    // 出错时尝试从缓存获取
    const cacheKey = `${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemIconGroupId}`
    const cachedData = ss.get(cacheKey)
    if (cachedData) {
      items.value[itemIconGroupIndex].items = cachedData

      // 当所有组的数据都加载完成后，应用网络模式过滤
      const allGroupsLoaded = items.value.every(group => group.items !== undefined)
      if (allGroupsLoaded) {
        filterItemsByNetworkMode()
      }
    }
  }
}

// 组件激活时刷新书签数据确保显示最新顺序
onActivated(() => {
  // 延迟执行，优先保证页面切换流畅
  setTimeout(() => {
    loadBookmarkTree(false);
  }, 20);
});

function handleRightMenuSelect(key: string | number) {
  dropdownShow.value = false
  let jumpUrl = panelState.networkMode === PanelStateNetworkModeEnum.lan ? currentRightSelectItem.value?.lanUrl : currentRightSelectItem.value?.url
  if (currentRightSelectItem.value?.lanUrl === '')
    jumpUrl = currentRightSelectItem.value.url
  switch (key) {
    case 'newWindows':
      if (jumpUrl) {
        openUrlWithoutReferer(jumpUrl, '_blank')
      }
      break
    case 'openWanUrl':
      if (currentRightSelectItem.value)
        openPage(currentRightSelectItem.value?.openMethod, currentRightSelectItem.value?.url, currentRightSelectItem.value?.title)
      break
    case 'openLanUrl':
      if (currentRightSelectItem.value && currentRightSelectItem.value.lanUrl)
        openPage(currentRightSelectItem.value?.openMethod, currentRightSelectItem.value.lanUrl, currentRightSelectItem.value?.title)
      break
    case 'edit':
      // 这里有个奇怪的问题，如果不使用{...}的方式 父组件的值会同步修改 标记一下
      handleEditItem({ ...currentRightSelectItem.value } as Panel.ItemInfo)
      break
    case 'delete':
      dialog.warning({
        title: t('common.warning'),
        content: t('common.deleteConfirmByName', { name: currentRightSelectItem.value?.title }),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: () => {
          if (currentRightSelectItem.value) {
            const itemIconGroupId = currentRightSelectItem.value.itemIconGroupId
            deletes([currentRightSelectItem.value.id as number]).then(({ code, msg }) => {
              if (code === 0) {
                ms.success(t('common.deleteSuccess'))
                // 清除该分组的图标缓存
                ss.remove(`${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemIconGroupId}`)
                getList()
              }
              else {
                ms.error(`${t('common.deleteFail')}:${msg}`)
              }
            })
          }
        },
      })

      break
    default:
      break
  }
}

function handleContextMenu(e: MouseEvent, itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus)
    return

  e.preventDefault()
  currentRightSelectItem.value = item
  dropdownShow.value = false
  nextTick().then(() => {
    dropdownShow.value = true
    dropdownMenuX.value = e.clientX
    dropdownMenuY.value = e.clientY
  })
}

function onClickoutside() {
  // message.info('clickoutside')
  dropdownShow.value = false
}

function handleEditSuccess(item: Panel.ItemInfo) {
  // 查找编辑的图标所属的分组
  for (let i = 0; i < items.value.length; i++) {
    const group = items.value[i]
    if (group.id === item.itemIconGroupId) {
      // 清除该分组的图标缓存
      ss.remove(`${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${item.itemIconGroupId}`)
      break
    }
  }
  getList()
}

// function handleChangeNetwork(mode: PanelStateNetworkModeEnum) {
//   panelState.setNetworkMode(mode)
//   if (mode === PanelStateNetworkModeEnum.lan)
//     ms.success(t('panelHome.changeToLanModelSuccess'))
//
//   else
//     ms.success(t('panelHome.changeToWanModelSuccess'))
//
//   // 切换网络模式后，重新应用过滤
//   filterItemsByNetworkMode()
// }


function handleSaveSort(itemGroup: ItemGroup) {
  const saveItems: Common.SortItemRequest[] = []
  if (itemGroup.items) {
    for (let i = 0; i < itemGroup.items.length; i++) {
      const element = itemGroup.items[i]
      saveItems.push({
        id: element.id as number,
        sort: i + 1,
      })
    }

    saveSort({ itemIconGroupId: itemGroup.id as number, sortItems: saveItems }).then(({ code, msg }) => {
      if (code === 0) {
        // 清除该分组的图标缓存
        ss.remove(`${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemGroup.id}`)
        // itemGroup.sortStatus = false // 不要自动关闭排序状态，允许用户继续操作
      }
      else {
        console.error(`${t('common.saveFail')}:${msg}`)
      }
    })
  }
}

function getDropdownMenuOptions() {
  const dropdownMenuOptions = [
    {
      label: t('iconItem.newWindowOpen'),
      key: 'newWindows',
    },

  ]

  // 当图标有公网地址时，显示打开公网地址选项
  if (currentRightSelectItem.value?.url) {
    dropdownMenuOptions.push({
      label: t('panelHome.openWanUrl'),
      key: 'openWanUrl',
    })
  }

  // 当图标有内网地址时，显示打开内网地址选项
  if (currentRightSelectItem.value?.lanUrl) {
    dropdownMenuOptions.push({
      label: t('panelHome.openLanUrl'),
      key: 'openLanUrl',
    })
  }

  if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN) {
    dropdownMenuOptions.push({
      label: t('common.edit'),
      key: 'edit',
    }, {
      label: t('common.delete'),
      key: 'delete',
    })
  }

  return dropdownMenuOptions
}

onMounted(async () => {
  // 更新用户信息
  updateLocalUserInfo()
  getList()

  // 预热内外网环境状态
  checkIntranetConnection()

  // 更新同步云端配置
  panelState.updatePanelConfigByCloud()

  // 检查是否需要自动获取网络壁纸
  if (panelState.panelConfig.autoNetworkWallpaper) {
    try {
      const apiUrl = panelState.panelConfig.autoNetworkWallpaperApi || 'https://img.xjh.me/random_img.php?return=302&type=bg&ctype=nature'
      panelState.panelConfig.backgroundImageSrc = apiUrl
      panelState.recordState()
    } catch (error) {
      console.error('自动获取网络壁纸失败', error)
    }
  }

  // 设置标题
  if (panelState.panelConfig.logoText)
    setTitle(panelState.panelConfig.logoText)

  // 确保公开模式下始终使用公网模式
  if (authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC) {
    panelState.setNetworkMode(PanelStateNetworkModeEnum.wan)
  }

  // 加载书签数据，使用forceRefresh=true确保获取最新排序
  await loadBookmarkTree(false)
})

onActivated(() => {
  // Reload bookmark tree when returning from manager to reflect cache changes
  loadBookmarkTree(true)
})
onBeforeRouteUpdate(() => {
  // Reload bookmark tree when route is updated
  loadBookmarkTree(true)
})

// 递归搜索书签树函数已被移除，搜索功能已迁移到SearchBoxWithSuggestions组件

// 前端搜索过滤
function itemFrontEndSearch(keyword?: string) {
  const trimmedKeyword = keyword?.trim() || ''
  if (trimmedKeyword !== '' && panelState.panelConfig.searchBoxSearchIcon) {
    const filteredData = ref<ItemGroup[]>([])
    const lowerCaseKeyword = trimmedKeyword.toLowerCase()

    // 只搜索原有图标（首页书签），不再搜索左侧书签
    for (let i = 0; i < items.value.length; i++) {
      const element = items.value[i].items?.filter((item: Panel.ItemInfo) => {
        // 首先应用网络模式过滤
        const networkModeMatch = panelState.networkMode !== PanelStateNetworkModeEnum.wan || item.lanOnly !== 1
        if (!networkModeMatch) return false

        // 然后应用搜索关键词过滤
        return (
          item.title.toLowerCase().includes(lowerCaseKeyword)
          || item.url.toLowerCase().includes(lowerCaseKeyword)
          || item.description?.toLowerCase().includes(lowerCaseKeyword)
        )
      })
      if (element && element.length > 0) {
        filteredData.value.push({ items: element, hoverStatus: false })
      }
    }

    filterItems.value = filteredData.value
  }
  else {
    // 没有搜索关键词时，应用网络模式过滤
    filterItemsByNetworkMode()
  }
}

function handleSetHoverStatus(groupIndex: number, hoverStatus: boolean) {
  if (items.value[groupIndex])
    items.value[groupIndex].hoverStatus = hoverStatus
}

function handleSetSortStatus(itemGroup: ItemGroup, sortStatus: boolean) {
  itemGroup.sortStatus = sortStatus

  // 并未保存排序重新更新数据
  if (!sortStatus) {
    if (itemGroup.id) {
       // Find the index in the original items array to ensure data consistency
       const idx = items.value.findIndex(x => x.id === itemGroup.id)
       if (idx !== -1) {
           updateItemIconGroupByNet(idx, itemGroup.id as number)
       }
    }
  }
}

function handleEditItem(item: Panel.ItemInfo) {
  editItemInfoData.value = item
  editItemInfoShow.value = true
  currentAddItenIconGroupId.value = undefined
}

function handleAddItem(itemIconGroupId?: number) {
  editItemInfoData.value = null
  editItemInfoShow.value = true
  if (itemIconGroupId)
    currentAddItenIconGroupId.value = itemIconGroupId
}

// 修改renderTreeLabel函数以正确使用SVG组件
const renderTreeLabel = ({ option }: { option: any }) => {
// 安全获取节点数据
const nodeData = option || {};

// 判断是否为文件夹节点
const isFolder = nodeData.isFolder === 1 || !nodeData.isLeaf;
const displayText = nodeData.label || nodeData.title || '未命名';

try {
return h('div', { class: 'flex items-center' }, [
isFolder ? h('svg', {
        xmlns: 'http://www.w3.org/2000/svg',
        class: 'w-4 h-4 mr-2',
        width: '24',
        height: '24',
        fill: '#4285F4',
        viewBox: '0 0 24 24'
      }, [
        h('path', {
          d: 'M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z'
        })
      ]) : (() => {
        let iconSrc = option.bookmark?.iconJson || '';
        if (iconSrc && !iconSrc.startsWith('data:')) {
          iconSrc = 'data:image/png;base64,' + iconSrc;
        }
        return h('img', {
          src: iconSrc, // 只使用缓存的base64数据，不请求外部图标
          class: 'w-4 h-4 mr-2 rounded-full',
          alt: 'bookmark icon'
        });
      })(), // 非文件夹节点显示图标
// 节点文本
h('span', {}, displayText)
]);
} catch (error) {
console.error('渲染树节点标签时出错:', error);
// 错误处理
return h('span', {}, displayText);
}
};

// 网络模式切换处理
function handleChangeNetwork(targetMode: PanelStateNetworkModeEnum) {
  // 从公网模式切换到私密模式需要验证密码
  if (panelState.networkMode === PanelStateNetworkModeEnum.wan && targetMode === PanelStateNetworkModeEnum.lan) {
    // 显示密码输入对话框
    const passwordInput = ref('')

    dialog.create({
      title: t('panelHome.verifyPassword'),
      content: () => h('div', { class: 'mt-4' }, [
        h('div', { class: 'mb-2 text-sm text-gray-600 dark:text-gray-400' }, t('panelHome.enterPasswordToSwitchLan')),
        h('input', {
          type: 'password',
          class: 'w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white',
          placeholder: t('common.password'),
          value: passwordInput.value,
          onInput: (e: Event) => {
            passwordInput.value = (e.target as HTMLInputElement).value
          },
          onKeydown: (e: KeyboardEvent) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              // 触发确定按钮
              const positiveButton = document.querySelector('.n-dialog__action button:last-child') as HTMLButtonElement
              if (positiveButton) positiveButton.click()
            }
          }
        })
      ]),
      positiveText: t('common.confirm'),
      negativeText: t('common.cancel'),
      onPositiveClick: async () => {
        if (!passwordInput.value) {
          ms.warning(t('panelHome.passwordRequired'))
          return false // 阻止对话框关闭
        }

        try {
          // 验证密码 - 调用登录接口验证
          const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
              username: authStore.userInfo?.username,
              password: passwordInput.value,
            }),
          })

          const result = await response.json()

          if (result.code === 0) {
            // 密码正确,切换模式
            // 切换前自动保存排序状态
            items.value.forEach(group => {
              if (group.sortStatus) {
                handleSaveSort(group)
                group.sortStatus = false // 立即关闭排序状态
              }
            })

            panelState.setNetworkMode(targetMode)
            ms.success(t('panelHome.changeToLanModelSuccess'))
            filterItemsByNetworkMode() // 确保视图更新
            return true
          } else {
            // 密码错误
            ms.error(t('panelHome.passwordIncorrect'))
            return false // 阻止对话框关闭
          }
        } catch (error) {
          console.error('验证密码失败:', error)
          ms.error(t('common.networkError'))
          return false
        }
      },
    })
  } else {
    // 切换前自动保存排序状态
    items.value.forEach(group => {
      if (group.sortStatus) {
        handleSaveSort(group)
        group.sortStatus = false // 立即关闭排序状态
      }
    })

    // 从内网切换到公网,或其他情况,直接切换
    panelState.setNetworkMode(targetMode)
    filterItemsByNetworkMode() // 确保视图更新
    if (targetMode === PanelStateNetworkModeEnum.wan) {
      ms.success(t('panelHome.changeToWanModelSuccess'))
    }
  }
}
</script>

<template>
	<div>
		<!-- 左上角抽屉按钮 - 大众常用样式 -->
		<div class="fixed top-4 left-4 z-50">
			<NButton
				circle
				color="#2a2a2a6b"
				class="w-10 h-10 !p-0 shadow-[0_0_10px_2px_rgba(0,0,0,0.2)] no-focus-outline no-tap-highlight"
				tabindex="-1"
				@click="drawerVisible = !drawerVisible"
			>
				<svg viewBox="0 0 24 24" class="w-6 h-6 text-white" v-if="drawerVisible">
					<path
						d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12
         5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
						fill="currentColor"
					/>
				</svg>
				<svg viewBox="0 0 24 24" class="w-6 h-6 text-white" v-else>
					<path
						d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"
						fill="currentColor"
					/>
				</svg>
			</NButton>
		</div>

		<!-- 右上角便签按钮 -->
		<div v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="fixed top-4 right-4 z-50 cursor-pointer transition-opacity hover:opacity-80 no-tap-highlight" @click="notepadVisible = !notepadVisible">
			<SvgIcon icon="note" class="text-white text-[32px] drop-shadow-md" />
		</div>

		<!-- 左侧抽屉 -->
		<NDrawer
			v-model:show="drawerVisible"
			placement="left"
			:width="isMobile ? '80%' : 300"
			style="overflow-y: auto;"
		>
			<NDrawerContent style=" min-height: 100vh;">
				<template #header>
					<div class="flex items-center justify-between w-full">
						<span class="text-lg font-medium">{{ t('bookmarkManager.bookmarkList') }}</span>
						<NButton type="info" size="small" round @click="navigateToBookmarkManager">
							{{ t('bookmarkManager.management') }}
						</NButton>
					</div>
				</template>
				<NTree
					:data="treeData"
					block-line
					expand-on-click
					:default-expanded-keys="treeData.length > 0 ? [treeData[0].key] : []"
					@update:selected-keys="handleTreeSelect"
					:render-label="renderTreeLabel"
				/>
			</NDrawerContent>
		</NDrawer>
	</div>
  <div class="w-full h-full sun-main">
    <div
      class="cover wallpaper" :style="{
        filter: `blur(${panelState.panelConfig.backgroundBlur}px)`,
        background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <div class="mask" :style="{ backgroundColor: `rgba(0,0,0,${panelState.panelConfig.backgroundMaskNumber})` }" />
    <div ref="scrollContainerRef" class="absolute w-full h-full overflow-auto">
      <div
        class="p-2.5 mx-auto"
        :style="{
          marginTop: `${panelState.panelConfig.marginTop}%`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <!-- 头 -->
        <div class="mx-[auto] w-[80%]">
          <div class="flex mx-[auto] items-center justify-center text-white">
            <div class="logo">
              <span class="text-2xl md:text-6xl font-bold text-shadow">
                {{ panelState.panelConfig.logoText }}
              </span>
            </div>
            <div class="divider text-base lg:text-2xl mx-[10px]">
              |
            </div>
            <div class="text-shadow">
              <Clock :hide-second="!panelState.panelConfig.clockShowSecond" />
            </div>
          </div>
          <div v-if="panelState.panelConfig.searchBoxShow" class="flex mt-[20px] mx-auto sm:w-full lg:w-[80%]">
            <SearchBoxWithSuggestions @itemSearch="itemFrontEndSearch" />
          </div>
        </div>

        <!-- 应用盒子 -->
        <div :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <!-- 系统监控状态 -->
          <div
            v-if="panelState.panelConfig.systemMonitorShow
              && ((panelState.panelConfig.systemMonitorPublicVisitModeShow && authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
                || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN)"
            class="flex mx-auto"
          >
            <SystemMonitor
              :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
              :show-title="panelState.panelConfig.systemMonitorShowTitle"
            />
          </div>

          <!-- 组纵向排列 -->
          <div
            v-for="(itemGroup, itemGroupIndex) in filterItems" :key="itemGroupIndex"
            class="item-list mt-[50px]"
            :class="itemGroup.sortStatus ? 'shadow-2xl border shadow-[0_0_30px_10px_rgba(0,0,0,0.3)]  p-[10px] rounded-2xl' : ''"
            @mouseenter="handleSetHoverStatus(itemGroupIndex, true)"
            @mouseleave="handleSetHoverStatus(itemGroupIndex, false)"
          >
            <!-- 分组标题 -->
            <div class="text-white text-xl font-extrabold mb-[20px] ml-[10px] flex items-center">
              <span class="group-title text-shadow">
                {{ itemGroup.title }}
              </span>
              <div
                v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN && panelState.networkMode === PanelStateNetworkModeEnum.lan"
                class="group-buttons ml-2 delay-100 transition-opacity flex"
              >
                <span class="mr-2 cursor-pointer" :title="t('common.add')" @click="handleAddItem(itemGroup.id)">
                  <SvgIcon class="text-white font-xl" icon="typcn:plus" />
                </span>
                <span class="mr-2 cursor-pointer " :title="t('common.sort')" @click="handleSetSortStatus(itemGroup, !itemGroup.sortStatus)">
                  <SvgIcon class="text-white font-xl" icon="ri:drag-drop-line" />
                </span>
              </div>
            </div>

            <!-- 详情图标 -->
            <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.info">
              <div v-if="itemGroup.items">
                <VueDraggable
                  v-model="itemGroup.items" item-key="sort" :animation="300"
                  class="icon-info-box"
                  filter=".not-drag"
                  :disabled="!itemGroup.sortStatus"
                  @end="handleSaveSort(itemGroup)"
                >
                  <div v-for="item, index in itemGroup.items" :key="index" :title="item.description" @contextmenu="(e) => handleContextMenu(e, itemGroupIndex, item)">
                    <AppIcon
                      :class="itemGroup.sortStatus ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="item"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="panelState.panelConfig.iconTextInfoHideDescription || false"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="0"
                      @click="handleItemClick(itemGroupIndex, item)"
                    />
                  </div>

                  <div v-if="itemGroup.items.length === 0" class="not-drag">
                    <AppIcon
                      :class="itemGroup.sortStatus ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="{ icon: { itemType: 3, text: 'subway:add' }, title: t('common.add'), url: '', openMethod: 0 }"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="panelState.panelConfig.iconTextInfoHideDescription || false"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="0"
                      @click="handleAddItem(itemGroup.id)"
                    />
                  </div>
                </VueDraggable>
              </div>
            </div>

            <!-- APP图标宫型盒子 -->
            <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.icon">
              <div v-if="itemGroup.items">
                <VueDraggable
                  v-model="itemGroup.items" item-key="sort" :animation="300"
                  class="icon-small-box"

                  filter=".not-drag"
                  :disabled="!itemGroup.sortStatus"
                  @end="handleSaveSort(itemGroup)"
                >
                  <div v-for="item, index in itemGroup.items" :key="index" :title="item.description" @contextmenu="(e) => handleContextMenu(e, itemGroupIndex, item)">
                    <AppIcon
                      :class="itemGroup.sortStatus ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="item"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="!panelState.panelConfig.iconTextInfoHideDescription"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="1"
                      @click="handleItemClick(itemGroupIndex, item)"
                    />
                  </div>

                  <div v-if="itemGroup.items.length === 0" class="not-drag">
                    <AppIcon
                      class="cursor-pointer"
                      :item-info="{ icon: { itemType: 3, text: 'subway:add' }, title: $t('common.add'), url: '', openMethod: 0 }"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="!panelState.panelConfig.iconTextInfoHideDescription"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="1"
                      @click="handleAddItem(itemGroup.id)"
                    />
                  </div>
                </vuedraggable>
              </div>
            </div>

            <!-- 编辑栏 -->

          </div>
        </div>
        <div class="mt-5 footer" v-html="panelState.panelConfig.footerHtml" />
      </div>
    </div>

    <!-- 右键菜单 -->
    <NDropdown
      placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY"
      :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect"
    />

    <!-- 悬浮按钮 -->
    <div class="fixed-element shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
      <NButtonGroup vertical>
        <!-- 网络模式切换按钮组 -->
        <NButton
          v-if="panelState.networkMode === PanelStateNetworkModeEnum.lan && panelState.panelConfig.netModeChangeButtonShow && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" color="#2a2a2a6b"
          :title="t('panelHome.changeToWanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.wan)"
        >
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="material-symbols:lan-outline-rounded" />
          </template>
        </NButton>

        <NButton
          v-if="panelState.networkMode === PanelStateNetworkModeEnum.wan && panelState.panelConfig.netModeChangeButtonShow && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" color="#2a2a2a6b"
          :title="t('panelHome.changeToLanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.lan)"
        >
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="mdi:wan" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" color="#2a2a2a6b" @click="settingModalShow = !settingModalShow">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="majesticons-applications" />
          </template>
        </NButton>

        <NButton color="#2a2a2a6b" :title="t('panelHome.refreshData')" @click="handleRefreshData">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="shuaxin" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC" color="#2a2a2a6b" :title="$t('panelHome.goToLogin')" @click="router.push('/login')">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="material-symbols:account-circle" />
          </template>
        </NButton>
      </NButtonGroup>

      <AppStarter v-model:visible="settingModalShow" />
      <NotePad ref="notepadInstance" v-model:visible="notepadVisible" />
      <!-- <Setting v-model:visible="settingModalShow" /> -->
    </div>

    <NBackTop
      :listen-to="() => scrollContainerRef"
      :right="10"
      :bottom="10"
      style="background-color:transparent;border: none;box-shadow: none;"
    >
      <div class="shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
        <NButton color="#2a2a2a6b">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="icon-park-outline:to-top" />
          </template>
        </NButton>
      </div>
    </NBackTop>

    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />

    <!-- 弹窗 -->
    <NModal
      v-model:show="windowShow" :mask-closable="false" preset="card"
      style="max-width: 1000px;height: 600px;border-radius: 1rem;" :bordered="true" size="small" role="dialog"
      aria-modal="true"
    >
      <template #header>
        <div class="flex items-center">
          <span class="mr-[20px]">
            {{ windowTitle }}
          </span>

          <NSpin v-if="windowIframeIsLoad" size="small" />
        </div>
      </template>
      <div class="w-full h-full rounded-2xl overflow-hidden border dark:border-zinc-700">
        <div v-if="windowIframeIsLoad" class="flex flex-col p-5">
          <NSkeleton height="50px" width="100%" class="rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
        </div>
        <iframe
          v-show="!windowIframeIsLoad" id="windowIframeId" ref="windowIframeRef" :src="windowSrc"
          class="w-full h-full" frameborder="0" @load="handWindowIframeIdLoad"
        />
      </div>
    </NModal>
  </div>
</template>

<style>
body,
html {
  overflow: hidden;
  background-color: rgb(54, 54, 54);
}
</style>

<style scoped>
.mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.sun-main {
  user-select: none;
}

.cover {
  position: absolute;
  width: 100%;
  height: 100%;
  overflow: hidden;
  /* background: url(@/assets/start_sky.jpg) no-repeat; */

  transform: scale(1.05);
}

.text-shadow {
  text-shadow: 2px 2px 50px rgb(0, 0, 0);
}

.app-icon-text-shadow {
  text-shadow: 2px 2px 5px rgb(0, 0, 0);
}

.fixed-element {
  position: fixed;
  /* 将元素固定在屏幕上 */
  right: 10px;
  /* 距离屏幕顶部的距离 */
  bottom: 50px;
  /* 距离屏幕左侧的距离 */
}

.icon-info-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(200px, 100%), 1fr));
  gap: 18px;

}

.icon-small-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
  gap: 18px;

}

/* 响应式图标块布局 */
@media (max-width: 1024px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(160px, 100%), 1fr));
    gap: 14px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(85px, 100%), 1fr));
    gap: 14px;
  }
}

/* 响应式图标块布局 - 继续使用grid布局，但减小最小宽度 */
@media (max-width: 768px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
    gap: 12px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(60px, 100%), 1fr));
    gap: 12px;
  }
}

@media (max-width: 480px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
    gap: 10px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(60px, 100%), 1fr));
    gap: 10px;
  }
}

@media (max-width: 360px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
    gap: 8px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(60px, 100%), 1fr));
    gap: 8px;
  }
}



/* 优化条状按钮阴影 */
/* 优化条状按钮阴影 - 已移除，避免污染全局 .fixed 类 */


:deep(.no-focus-outline:focus) {
  box-shadow: none !important;
}

.no-tap-highlight {
  -webkit-tap-highlight-color: transparent !important;
  outline: none !important;
}
</style>

