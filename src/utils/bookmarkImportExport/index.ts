import CryptoJS from 'crypto-js'
import moment from 'moment'

const VERSION = 1 // 当前书签文件版本
const ALLOW_LOW_VERSION = 1 // 最小支持的书签文件版本号
const APPNAME = 'Sun-Panel-Bookmark'

// 书签接口定义
export interface BookmarkItem {
  id: number
  title: string
  url: string
  folderId?: number
  isFolder: boolean
  icon?: string
}

export interface BookmarkFolder {
  id: number
  title: string
  children: (BookmarkItem | BookmarkFolder)[]
  isFolder: boolean
}

export interface BookmarkJsonStructure {
  version: number
  appName: 'Sun-Panel-Bookmark'
  exportTime: string
  appVersion: string
  bookmarks: BookmarkFolder[]
  md5: string
}

export class FormatError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'FormatError'
  }
}

export class ConfigVersionLowError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'ConfigVersionLowError'
  }
}

interface ExportBookmarkResult {
  addBookmarksData(datas: BookmarkFolder[]): ExportBookmarkResult
  exportFile(): void
  string(): string
}

// 导出书签数据
export function exportBookmarkJson(appVersion?: string): ExportBookmarkResult {
  const jsonData: BookmarkJsonStructure = {
    version: VERSION,
    appName: APPNAME,
    exportTime: moment().format('YYYY-MM-DD HH:mm:ss'),
    appVersion: appVersion || '',
    bookmarks: [],
    md5: '',
  }

  // MD5 生成函数
  function generateMD5AndUpdate() {
    jsonData.md5 = generateMD5(JSON.stringify(jsonData))
  }

  return {
    // 添加书签信息
    addBookmarksData(datas: BookmarkFolder[]) {
      jsonData.bookmarks = datas
      return this
    },

    // 导出json文件
    exportFile() {
      generateMD5AndUpdate()
      const jsonString = JSON.stringify(jsonData)
      if (jsonString) {
        const blob = new Blob([jsonString], { type: 'application/json' })
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = `SunPanel-Bookmark${moment().format('YYYYMMDDHHmm')}.sun-panel.json`
        link.click()
      }
    },

    // 返回字符串
    string() {
      generateMD5AndUpdate()
      return JSON.stringify(jsonData)
    },
  }
}

export interface ImportBookmarkResult {
  isPassCheckMd5: () => boolean
  isPassCheckConfigVersionOld: () => boolean
  isPassCheckConfigVersionNew: () => boolean
  isPassCheckConfigVersionBest: () => boolean
  jsonStruct: BookmarkJsonStructure
  hasProperty: (key: string) => boolean
  getBookmarks: () => BookmarkFolder[]
}

// 导入书签json数据
export function importBookmarkJsonString(jsonString: string): ImportBookmarkResult | null {
  let data: any
  try {
    data = JSON.parse(jsonString)
  } catch (error) {
    throw new FormatError('file format error')
  }

  const jsonStruct = transformJson(data)
  const md5 = generateMD5(jsonString)

  if (!jsonStruct) {
    throw new FormatError('file format error')
  }

  if (data.version < ALLOW_LOW_VERSION) {
    throw new ConfigVersionLowError('')
  }

  return {
    isPassCheckMd5: () => md5 === jsonStruct.md5,
    isPassCheckConfigVersionOld: () => !(jsonStruct.version < VERSION),
    isPassCheckConfigVersionNew: () => !(jsonStruct.version > VERSION),
    isPassCheckConfigVersionBest: () => jsonStruct.version === VERSION,
    jsonStruct,
    hasProperty: (key: string): boolean => {
      return key in jsonStruct
    },
    getBookmarks: (): BookmarkFolder[] => {
      return jsonStruct.bookmarks || []
    },
  }
}

function transformJson(jsonData: any): BookmarkJsonStructure | null {
  // 检查必须存在的键
  const requiredKeys: Array<keyof BookmarkJsonStructure> = ['version', 'appName', 'exportTime', 'appVersion', 'md5']
  for (const key of requiredKeys) {
    if (!(key in jsonData)) {
      return null
    }
  }

  // 使用类型断言将 JSON 数据转换为指定类型
  const transformedData: BookmarkJsonStructure = jsonData as BookmarkJsonStructure

  // 返回转换后的数据
  return transformedData
}

function generateMD5(jsonString: string): string {
  try {
    // 解析 JSON 字符串
    const data: any = JSON.parse(jsonString)
    // 移除 md5 字段及其对应的值
    removeMD5Field(data)
    // 将修改后的 JSON 对象转换回字符串
    const modifiedJsonString = JSON.stringify(data)
    // 使用 crypto-js 计算 MD5 值
    const md5 = CryptoJS.MD5(modifiedJsonString).toString()
    return md5
  } catch (error) {
    return ''
  }
}

function removeMD5Field(obj: any): void {
  for (const key in obj) {
    if (key === 'md5') {
      // 移除 md5 字段
      delete obj[key]
      return
    }
  }
}

// 解析浏览器导出的HTML书签文件
export function parseBrowserBookmarkHTML(htmlContent: string): BookmarkFolder[] {
  const parser = new DOMParser()
  const doc = parser.parseFromString(htmlContent, 'text/html')
  const rootFolder: BookmarkFolder = {
    id: 0,
    title: '书签',
    children: [],
    isFolder: true
  }

  // 查找所有DL标签（书签文件夹）
  const dlElements = doc.querySelectorAll('dl')
  if (dlElements.length > 0) {
    // 处理第一个DL标签作为根文件夹
    const firstDl = dlElements[0]
    parseDLElement(firstDl, rootFolder)
  }

  return [rootFolder]
}

function parseDLElement(dlElement: Element, parentFolder: BookmarkFolder): void {
  let currentFolder = parentFolder
  let currentTitle = ''

  // 遍历DL标签内的所有子节点
  for (let i = 0; i < dlElement.childNodes.length; i++) {
    const node = dlElement.childNodes[i]
    
    if (node.nodeType === Node.ELEMENT_NODE) {
      const element = node as Element
      
      // 处理DT标签
      if (element.tagName.toLowerCase() === 'dt') {
        // 检查是否包含H3标签（文件夹标题）
        const h3Element = element.querySelector('h3')
        if (h3Element) {
          currentTitle = h3Element.textContent?.trim() || '未命名文件夹'
          
          // 创建新文件夹
          const newFolder: BookmarkFolder = {
            id: Date.now() + Math.floor(Math.random() * 1000), // 生成临时ID
            title: currentTitle,
            children: [],
            isFolder: true
          }
          
          currentFolder.children.push(newFolder)
          currentFolder = newFolder
        }
        // 检查是否包含A标签（书签项）
        else {
          const aElement = element.querySelector('a')
          if (aElement) {
            const bookmarkTitle = aElement.textContent?.trim() || '未命名书签'
            const bookmarkUrl = aElement.getAttribute('href') || ''
            
            // 创建书签项
            const bookmarkItem: BookmarkItem = {
              id: Date.now() + Math.floor(Math.random() * 1000), // 生成临时ID
              title: bookmarkTitle,
              url: bookmarkUrl,
              folderId: parentFolder.id,
              isFolder: false
            }
            
            parentFolder.children.push(bookmarkItem)
          }
        }
      }
      // 处理嵌套的DL标签
      else if (element.tagName.toLowerCase() === 'dl') {
        parseDLElement(element, currentFolder)
      }
    }
  }
}

// 生成唯一ID的工具函数
export function generateUniqueId(): number {
  return Date.now() + Math.floor(Math.random() * 1000)
}

// 扁平化书签树
export function flattenBookmarkTree(bookmarks: BookmarkFolder[]): (BookmarkItem | BookmarkFolder)[] {
  const result: (BookmarkItem | BookmarkFolder)[] = []
  
  function traverse(nodes: (BookmarkItem | BookmarkFolder)[], parentId?: number) {
    for (const node of nodes) {
      if (node.isFolder) {
        const folder = node as BookmarkFolder
        // 包含文件夹本身
        result.push({
          ...folder,
          folderId: parentId
        })
        // 继续遍历子节点
        if (folder.children) {
          traverse(folder.children, folder.id)
        }
      } else {
        const bookmark = node as BookmarkItem
        result.push({
          ...bookmark,
          folderId: parentId
        })
      }
    }
  }
  
  traverse(bookmarks)
  return result
}