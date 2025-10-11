import { post } from '@/utils/request'

/**
 * 批量添加书签
 */
export function addMultiple<T>(req: Panel.Info[]) {
  return post<T>({
    url: '/panel/bookmark/addMultiple',
    data: req,
  })
}

/**
 * 添加单个书签
 */
export function add<T>(req: Panel.bookmarkInfo) {
  return post<T>({
    url: '/panel/bookmark/add',
    data: req,
  })
}

/**
 * 获取书签列表
 */
export function getList<T>() {
  return post<T>({
    url: '/panel/bookmark/getList',
  })
}

/**
 * 更新书签
 */
export function update<T>(req: Panel.bookmarkInfo) {
  return post<T>({
    url: '/panel/bookmark/update',
    data: req,
  })
}

/**
 * 删除书签
 */
export function deletes<T>(ids: number[]) {
  return post<T>({
    url: '/panel/bookmark/deletes',
    data: { ids },
  })
}
