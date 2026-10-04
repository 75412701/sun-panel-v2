<template>
	<div
		class="flex flex-col h-screen bg-white dark:bg-gray-800"
		@contextmenu.prevent
	>
		<!-- 顶部标题栏 -->
		<div class="px-4 py-2.5 border-b flex items-center justify-between bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 relative">
			<!-- 移动端左栏展开按钮 -->
			<div
				v-if="isMobile"
				@click="togglePanel"
				class="flex items-center justify-center w-10 h-10 rounded-full bg-transparent text-gray-700 dark:text-white cursor-pointer transition-all hover:bg-gray-100 dark:hover:bg-gray-700"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</div>

			<div
				@click="goBackToHome"
				class="flex items-center justify-center w-10 h-10 rounded-full bg-transparent text-gray-700 dark:text-white cursor-pointer transition-all hover:bg-gray-100 dark:hover:bg-gray-700"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</div>

			<h1 class="text-xl font-bold text-gray-800 dark:text-white flex-1 text-center">{{ t('bookmarkManager.management') }}</h1>

			<!-- 自定义下拉菜单 -->
			<div class="relative custom-dropdown">
				<div
					class="flex items-center justify-center w-10 h-10 rounded-full bg-transparent text-gray-700 dark:text-white cursor-pointer transition-all hover:bg-gray-100 dark:hover:bg-gray-700 mr-2"
					@click="isDropdownOpen = !isDropdownOpen"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
					</svg>
				</div>
				<!-- 下拉菜单内容 -->
				<div
					v-if="isDropdownOpen"
					class="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 text-gray-700 dark:text-white rounded-md shadow-lg py-1 z-[100000]"
				>
					<button
					class="block px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 w-full text-left"
					@click.stop="bookmarkType = 'bookmark'; createNewBookmark(); isDropdownOpen = false"
				>
					{{ t('bookmarkManager.addBookmark') }}
				</button>
				<button
					class="block px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 w-full text-left"
					@click.stop="bookmarkType = 'folder'; createNewBookmark(); isDropdownOpen = false"
				>
					{{ t('bookmarkManager.addFolder') }}
				</button>
				<div class="border-t border-gray-200 dark:border-gray-700 my-1"></div>
					<button
						class="block px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 w-full text-left"
						@click.stop="triggerImportBookmarks(); isDropdownOpen = false"
					>
						{{ t('bookmarkManager.importBookmarks') }}
					</button>
					<button
						class="block px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 w-full text-left"
						@click.stop="exportBookmarks(); isDropdownOpen = false"
					>
						{{ t('bookmarkManager.exportBookmarks') }}
					</button>
				</div>
			</div>

			<!-- 隐藏的文件输入框 -->
			<input
				ref="fileInput"
				type="file"
				accept=".html"
				style="display: none;"
				@change="handleFileChange"
			/>
		</div>

		<!-- 主内容区域 -->
		<div class="flex flex-1 overflow-hidden">
			<!-- 遮罩层：移动端左栏打开时 -->
			<div
				v-if="isMobile && showLeftPanel"
				class="fixed inset-0 bg-black bg-opacity-30 z-40"
				@click="collapsePanel"
			></div>

			<!-- 左侧书签树 -->
			<div
				v-show="showLeftPanel"
				:class="[
    isMobile ? 'fixed top-0 left-0 h-full bg-white dark:bg-gray-800 z-50 rounded-r-lg shadow-lg overflow-auto transition-all duration-300 ease-in-out' : 'h-full bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 overflow-auto',
    isMobile && isPanelExpanded ? 'w-3/4' : isMobile ? 'w-12' : ''
  ]"
				:style="{ width: !isMobile ? leftPanelWidth + 'px' : '' }"
			>
				<n-tree
					:data="bookmarkTree"
					:default-expanded-keys="defaultExpandedKeys"
					:selected-keys="selectedKeysRef"
					block-line
					@update:selected-keys="handleSelect"
					@expand="handleNodeExpand"
					:render-label="renderTreeLabel"
					:render-expand-icon="renderExpandIcon"
					ref="treeRef"
				/>
			</div>

			<!-- 可拖动分割线（桌面端） -->
			<div
				v-if="!isMobile"
				class="w-1 bg-gray-200 cursor-col-resize hover:bg-blue-300 flex items-center justify-center"
				@mousedown="startResize"
				:style="{ height: '100%', userSelect: 'none' }"
			>
				<div class="w-px h-12 bg-gray-400"></div>
			</div>

			<!-- 右侧书签列表 -->
		<div class="flex-1 flex flex-col overflow-hidden">
			<div class="sticky top-0 z-10 p-2 border-b flex flex-col bg-white dark:bg-gray-800">
					<!-- 面包屑导航 -->
				<div class="flex items-center text-sm mb-2 text-gray-600 dark:text-gray-400">
					<span v-for="(crumb, index) in currentPath" :key="index"
					      class="cursor-pointer hover:text-blue-600"
					      @click="handleBreadcrumbClick(crumb.id)">
						{{ crumb.name }}
						<span v-if="index < currentPath.length - 1" class="mx-1">/</span>
					</span>
				</div>
				<div class="flex-1 rounded-full overflow-hidden border-2 border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800">
					<n-input
						v-model:value="searchQuery"
						:placeholder="t('bookmarkManager.searchPlaceholder')"
						clearable
						@input="handleSearch"
						class="w-full bg-transparent text-gray-800 dark:text-white bookmark-search-input"
					/>
				</div>
			</div>

				<!-- 书签列表 - 简洁列表样式 -->
				<div class="flex-1 relative overflow-auto bg-white dark:bg-gray-800" @dragover.prevent="handleContainerDragOver">
					<div v-if="filteredBookmarks.length === 0" class="text-center py-8 text-gray-400 dark:text-gray-500">
						{{ t('bookmarkManager.noData') }}
					</div>

					<div v-else class="py-2" @dragover.prevent="handleContainerDragOver">
						<!-- 全局拖拽指示线 -->
						<div
							v-if="dragIndicatorTop !== null"
							class="absolute left-4 right-4 z-20 flex items-center pointer-events-none transition-all duration-75"
							:style="{ top: dragIndicatorTop + 'px', transform: 'translateY(-50%)' }"
						>
							<div class="w-full h-[2px] bg-blue-500"></div>
						</div>
						<template v-for="item in filteredBookmarks" :key="item.id">
							<!-- 相对定位容器,用于包含绝对定位的插入横线 -->
							<div
								class="relative py-[2px]"
								:draggable="true"
								@dragstart="handleDragStart($event, item)"
								@dragend="handleDragEnd($event); dragOverTarget = null; dragInsertPosition = null; dragIndicatorTop = null"
								@dragover="handleDragOver($event, item)"
								@dragleave="handleDragLeave($event)"
								@drop="handleDrop($event, item); dragOverTarget = null; dragInsertPosition = null; dragIndicatorTop = null"
							>

								<div
									:class="[
										'flex items-center px-4 py-2 cursor-pointer transition-colors group',
										dragOverTarget === item.id && item.isFolder && dragInsertPosition === null
											? 'bg-blue-50 dark:bg-blue-900'
											: selectedBookmarkId === String(item.id) || (item.isFolder && selectedFolder === String(item.id))
											? 'bg-gray-100 dark:bg-gray-700'
											: 'hover:bg-gray-50 dark:hover:bg-gray-700'
									]"
									@contextmenu.prevent="!isMobile ? openContextMenu($event, item) : null"
									@click="focusedItemId = String(item.id)"
									@dblclick="item.isFolder ? openFolder(item.id) : openBookmark(item)"
								>
									<!-- 图标 -->
									<div class="flex-shrink-0 w-4 h-4 flex items-center justify-center mr-3">
	<span v-if="item.isFolder" class="text-gray-400 dark:text-gray-500 group-hover:text-[#4285F4] dark:group-hover:text-[#4285F4] transition-colors">
		<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
			<path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z" />
		</svg>
	</span>
										<img v-else-if="item.iconJson"
											:src="item.iconJson.startsWith('data:') ? item.iconJson : 'data:image/png;base64,' + item.iconJson"
											alt=""
											class="w-4 h-4 rounded"
											onerror="this.style.display='none'; this.nextElementSibling.style.display='block';"
										>
										<span v-else class="w-4 h-4 bg-gray-200 dark:bg-gray-600 rounded" style="display: none;"></span>
									</div>

									<!-- 文本 - 聚焦时在标题后显示URL -->
									<div class="flex-1 min-w-0">
										<div class="text-sm text-gray-900 dark:text-white truncate">
											<span>{{ item.title }}</span>
											<span v-if="!item.isFolder && item.url && focusedItemId === String(item.id)" class="text-gray-500 dark:text-gray-400 transition-opacity ml-2">
												{{ item.url }}
											</span>
											<span v-else-if="item.isFolder && focusedItemId === String(item.id)" class="text-gray-400 dark:text-gray-500 italic transition-opacity ml-2">
												{{ t('bookmarkManager.folder') || '文件夹' }}
											</span>
										</div>
									</div>

									<!-- 三个点菜单 - 一直显示，颜色与右上角一致 -->
									<div
										class="flex-shrink-0 w-6 h-6 flex items-center justify-center ml-2 cursor-pointer text-gray-700 dark:text-white"
										@click.stop="openContextMenu($event, item)"
									>
										<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
										</svg>
									</div>
								</div>


							</div>
						</template>
					</div>
				</div>
			</div>
		</div>

		<!-- 编辑书签对话框 -->
		<div v-if="isEditDialogOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
			<div class="bg-white dark:bg-gray-800 p-6 rounded-lg w-96">
				<h3 class="text-xl font-bold text-gray-800 dark:text-white mb-4">{{ isCreateMode ? (bookmarkType === 'bookmark' ? t('bookmarkManager.createBookmark') : t('bookmarkManager.createFolder')) : t('bookmarkManager.editBookmark') }}</h3>
				<div class="mb-4">
					<label class="block mb-2 text-gray-800 dark:text-white">{{ t('bookmarkManager.title') }}</label>
					<input
				v-model="currentEditBookmark.title"
				class="w-full px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 text-gray-800 dark:text-white"
				:placeholder="t('bookmarkManager.title')"
			/>
				</div>
				<div v-if="(!isCreateMode && !currentBookmark?.isFolder) || (isCreateMode && bookmarkType === 'bookmark')" class="mb-4">
					<label class="block mb-2 text-gray-800 dark:text-white">{{ t('bookmarkManager.url') }}</label>
					<input
				v-model="currentEditBookmark.url"
				class="w-full px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 text-gray-800 dark:text-white"
				:placeholder="t('bookmarkManager.enterUrl')"
			/>
				</div>
				<div class="mb-4">
					<label class="block mb-2 text-gray-800 dark:text-white">{{ t('bookmarkManager.folder') }}</label>
					<n-tree-select
						v-model:value="currentEditBookmark.folderId"
						:options="folderTreeOptions"
						key-field="key"
						label-field="label"
						:placeholder="t('bookmarkManager.folder')"
						default-expand-all
						class="w-full"
					/>
				</div>
				<div class="flex justify-end gap-2">
					<button @click="closeEditDialog" class="px-4 py-2 border border-gray-300 rounded-md text-gray-800 dark:text-white">{{ t('bookmarkManager.cancel') }}</button>
					<button @click="saveBookmarkChanges" class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors shadow-md">{{ t('bookmarkManager.confirm') }}</button>
				</div>
			</div>
		</div>

		<!-- 右键菜单 -->
		<div
			v-if="isContextMenuOpen"
			:style="contextMenuStyle"
			class="fixed bg-white dark:bg-gray-800 text-gray-700 dark:text-white shadow-lg py-1 z-50 w-40 context-menu border border-gray-200 dark:border-gray-700"
			@contextmenu.prevent.stop
		>
			<div v-if="!currentBookmark?.isFolder" @click="handleEditBookmark" class="px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer">{{ t('bookmarkManager.edit') }}</div>
			<div @click="handleDeleteBookmark" class="px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer">{{ t('bookmarkManager.delete') }}</div>
		</div>
	</div>
</template>

<script setup lang="ts">

import { ref, computed, onMounted, onUnmounted, h, watch, nextTick } from 'vue'
// 不再直接导入SVG文件，使用内联方式
import { NTree, NInput, useMessage, NTreeSelect } from 'naive-ui'
import { useRouter } from 'vue-router'
import { getList as getBookmarksList, add as addBookmark, update, deletes, saveSort as saveBookmarkSort, addMultiple as addMultipleBookmarks } from '@/api/panel/bookmark'
import { t } from '@/locales'
import { dialog } from '@/utils/request/apiMessage'
import { ss } from '@/utils/storage/local'
import { openUrlWithoutReferer } from '@/utils/cmn'


// 组件顶部（import 之后）添加核心接口
interface TreeOption {
	key: string; // 节点唯一标识
	label: string; // 显示文本
	isLeaf: boolean; // 是否为叶子节点（书签）
	isFolder: boolean; // 是否为文件夹（核心标识）
	bookmark?: Omit<Bookmark, 'folderId'>; // 书签信息（非文件夹节点）
	children: TreeOption[]; // 子节点（必须初始化为数组）
	rawNode: any; // 后端原始数据
	disabledExpand: boolean; // 关键：是否禁用折叠（控制图标显示）
	sort?: number; // 排序字段
	ParentId?: string; // 父节点ID（缓存时使用）
}

const  BOOKMARKS_CACHE_KEY= 'bookmarksTreeCache'// 完整书签数据缓存 (与首页缓存键一致)

const router = useRouter()
const ms = useMessage()
const isMobile = ref(false)      // 是否移动端
const showLeftPanel = ref(true)  // 左栏是否显示
// 左侧面板宽度
const leftPanelWidth = ref(256) // 默认256px
const isResizing = ref(false)
// 只展开第一级节点的键
const defaultExpandedKeys = ref<string[]>([])

// 导入相关
const fileInput = ref<HTMLInputElement>()
const uploadLoading = ref(false)
const jsonData = ref<string | null>(null)
const importWarning = ref<string[]>([])
const isPanelExpanded = ref(false)
// 返回首页
function goBackToHome() {
  router.push('/')
}


function togglePanel() {
	showLeftPanel.value = true
	isPanelExpanded.value = true
}

function collapsePanel() {
	isPanelExpanded.value = false
	setTimeout(() => {
		showLeftPanel.value = false
	}, 300) // 等待动画完成再隐藏
}

// 组件挂载时加载书签数据
onMounted(() => {
	// 延迟执行，优先保证页面切换流畅
	setTimeout(() => {
		refreshBookmarks();
	}, 20);

	// 添加全局事件监听器
	document.addEventListener('mousemove', handleMouseMove);
	document.addEventListener('mouseup', stopResize);
	document.addEventListener('click', handleGlobalClick);
	// 全局处理dragover，防止出现禁用光标 (使用捕获阶段，确保优先执行)
	document.addEventListener('dragover', handleGlobalDragOver, true);
	document.addEventListener('dragenter', handleGlobalDragOver, true);

	handleResize();
	window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
	document.removeEventListener('mouseup', stopResize);
	document.removeEventListener('click', handleGlobalClick);
	document.removeEventListener('dragover', handleGlobalDragOver, true);
	document.removeEventListener('dragenter', handleGlobalDragOver, true);
	window.removeEventListener('resize', handleResize);
});

// 全局dragover/dragenter处理
function handleGlobalDragOver(e: DragEvent) {
	e.preventDefault();
	if (e.dataTransfer) {
		e.dataTransfer.dropEffect = 'move';
	}
}

// 开始调整大小
function startResize(e: MouseEvent) {
  isResizing.value = true
  // 防止文本选择
  e.preventDefault()
}

// 处理鼠标移动
function handleMouseMove(e: MouseEvent) {
  if (!isResizing.value) return

  // 获取主内容区域的位置
  const container = document.querySelector('.flex-1.overflow-hidden') as HTMLElement
  if (!container) return

  const containerRect = container.getBoundingClientRect()
  const newWidth = e.clientX - containerRect.left

  // 设置最小宽度和最大宽度限制
  if (newWidth > 150 && newWidth < containerRect.width - 200) {
    leftPanelWidth.value = newWidth
  }
}

// 结束调整大小
function stopResize() {
  isResizing.value = false
}


interface Bookmark {
	id: number
	title: string
	url: string
	folderId?: string | number
	isFolder?: boolean
	parentId?: number
	createTime?: string
	updateTime?: string
	iconJson?: string
	sort?: number

	icon?: any | null
	lanUrl?: string
	openMethod?: number
}


const bookmarkTree = ref<any[]>([])

// 当前选中的文件夹，初始化为根目录'0'
const selectedFolder = ref<string>('0')

// 新增：响应式的完整数据变量
const fullData = ref<any[]>([])

// 新增：响应式的原始数据缓存变量
const rawDataCache = ref<any>(null)

// 计算属性 - 所有书签和文件夹
const allItems = computed<(Bookmark & { isFolder?: boolean })[]>(() => {
	const items: (Bookmark & { isFolder?: boolean })[] = []

	// 获取完整数据，优先使用响应式的fullData变量
	const data = fullData.value.length > 0 ? fullData.value : bookmarkTree.value

	// 遍历所有文件夹和书签
	function traverseItems(nodes: any[], folderId: string) {
		for (const node of nodes) {
			// 如果是文件夹，添加到列表并继续遍历
			if (node.isFolder || (node.children && node.children.length > 0)) {
				items.push({
					id: Number(node.key),
					title: node.label,
					url: '',
					folderId: String(folderId),
					isFolder: true,
					sort: node.sort || node.rawNode?.sort || 0,  // 添加sort值
					iconJson: node.rawNode?.iconJson || '',
					lanUrl: node.rawNode?.lanUrl || '',
					openMethod: node.rawNode?.openMethod || 0,
					icon: node.rawNode?.icon || null
				})
			}
			// 如果是书签节点，添加到列表
			if ((node.isLeaf && node.bookmark) || (!node.isFolder && node.id !== undefined) || (!node.isLeaf && node.url)) {
				const bookmarkData = node.bookmark || node;
				items.push({
					...bookmarkData,
					folderId: String(folderId),
					isFolder: false
				})
			}
			// 继续遍历子节点
			if (node.children && node.children.length > 0) {
				traverseItems(node.children, String(node.key))
			}
		}
	}

	// 开始遍历完整数据
	traverseItems(data, '0')
	return items
})

// 搜索
const searchQuery = ref('')
const filteredBookmarks = computed(() => {
	// 如果选中了具体书签，直接返回该书签
	if (selectedBookmarkId.value) {
		const bookmark = allItems.value.find(b => String(b.id) === selectedBookmarkId.value);
		return bookmark ? [bookmark] : [];
	}

	// 先获取所有项目（书签和文件夹）
	let items = allItems.value

	// 当没有搜索查询时，按当前选中的文件夹过滤；有搜索时显示所有文件夹的内容
	if (!searchQuery.value.trim()) {
		const targetFolderId = selectedFolder.value || '0'
		items = items.filter(item => {
			const itemFolderId = String(item.folderId || '0') // 确保folderId存在
			return itemFolderId === targetFolderId
		})
	}


	// 应用搜索过滤
	if (searchQuery.value.trim()) {
		const query = searchQuery.value.toLowerCase()
		items = items.filter(item =>
			item.title.toLowerCase().includes(query) ||
			(item.url && item.url.toLowerCase().includes(query))
		)
	}

	return items
})

// 当前选中的节点键引用
const selectedKeysRef = ref<(string | number)[]>([]);
// 左侧树选中的键
// 移除未使用的treeSelectedKeys

// 当前选中的书签ID（用于显示单个书签）
const selectedBookmarkId = ref<string>('')

// 当前聚焦的项目ID（用于单击显示URL）
const focusedItemId = ref<string>('')

// 拖拽相关
const draggedItem = ref<any>(null)

// 查找节点并构建路径
function findNodePath(nodes: any[], targetId: string, currentPath: {id: string, name: string}[] = []): {id: string, name: string}[] | null {
	for (const node of nodes) {
		// 先将当前节点添加到路径中
		const newPath = [...currentPath, {id: String(node.key), name: node.label}];

		// 如果找到目标节点，返回完整路径
		if (String(node.key) === targetId) {
			return newPath;
		}

		// 如果有子节点，递归查找
		if (node.children && node.children.length > 0) {
			const foundPath = findNodePath(node.children, targetId, newPath);
			if (foundPath) {
				return foundPath;
			}
		}
	}
	return null;
}

// 当前路径
const currentPath = computed(() => {
	// 根路径始终是路径的起点
	const rootPath = [{id: '0', name: t('bookmarkManager.rootDirectory') }];

	// 如果没有选中文件夹或选中的是根文件夹，只返回根路径
	if (!selectedFolder.value || selectedFolder.value === '0') {
		return rootPath;
	}

	// 查找选中文件夹的完整路径
	const fullPath = findNodePath(bookmarkTree.value, selectedFolder.value, []);

	// 如果找到完整路径，将根路径与子路径合并
	if (fullPath) {
		return rootPath.concat(fullPath);
	}

	// 如果找不到，返回根路径 + 当前选中文件夹
	return rootPath.concat([{id: selectedFolder.value, name: '...'}]);
});

// 处理面包屑点击
function handleBreadcrumbClick(id: string) {
	selectedFolder.value = id;
	selectedBookmarkId.value = '';
}

// 点击树节点
function handleSelect(keys: (string | number)[]) {
	// 更新选中的节点键引用
	selectedKeysRef.value = keys;

	// 重置选中的书签ID
	selectedBookmarkId.value = '';

	// 确保类型安全的赋值方式
	if (keys && Array.isArray(keys) && keys.length > 0) {
		const key = String(keys[0]);

		// 查找当前点击的节点
		function findNodeById(nodes: any[], id: string): any | null {
			for (const node of nodes) {
				if (String(node.key) === id) {
					return node;
				}
				if (node.children && node.children.length > 0) {
					const found = findNodeById(node.children, id);
					if (found) return found;
				}
			}
			return null;
		}

		const selectedNode = findNodeById(bookmarkTree.value, key);

		// 确保node存在
		if (selectedNode) {
			// 检查是否为具体书签节点
			if (selectedNode.isLeaf && selectedNode.bookmark) {
				// 如果是具体书签，设置selectedBookmarkId
				selectedBookmarkId.value = key;
				selectedFolder.value = ''; // 清空选中的文件夹
			}
			// 如果是文件夹，设置selectedFolder
			else if (selectedNode.isFolder || !selectedNode.isLeaf || !selectedNode.bookmark?.url) {
				selectedFolder.value = key;
				// 强制保持选中状态
					selectedKeysRef.value = [key];
			}
		}
	}
	// 不在这里清空selectedFolder，这样当取消选择时仍能保持文件夹筛选状态
	// 只有当选择了其他类型节点或明确取消选择时才清空
}

// 文件夹展开时自动选中并显示内部书签
function handleNodeExpand(node: any) {
	if (node && node.key) {
		const key = String(node.key);
		// 设置选中的文件夹
		selectedFolder.value = key;
		// 确保选中状态正确更新
		selectedKeysRef.value = [key];

	}
}

// 搜索时清空选中
function handleSearch() {
	if (searchQuery.value) {
		selectedFolder.value = ''
		selectedBookmarkId.value = '' // 同时清空选中的书签
	}
}

// 打开文件夹（清空搜索）
function openFolder(folderId: string | number) {
	selectedFolder.value = String(folderId)
	searchQuery.value = '' // 清空搜索框
	selectedKeysRef.value = [String(folderId)] // 同步左侧树选中状态
}



// 右键菜单相关状态
const isContextMenuOpen = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const currentBookmark = ref<(Bookmark & { isFolder?: boolean }) | null>(null);

// 右上角菜单相关状态
const isDropdownOpen = ref(false);

// 树组件引用
const treeRef = ref<InstanceType<typeof NTree> | null>(null);



// 用于TreeSelect的文件夹树选项
const folderTreeOptions = computed(() => {
  const processNodes = (nodes: any[]): any[] => {
    // 过滤并转换节点
    return nodes
      .map(node => {
        // 判断是否为文件夹
        const isFolder = node.isFolder || (!node.isLeaf && !node.bookmark?.url);

        // 如果不是文件夹，跳过
        if (!isFolder) return null;

        const option: any = {
          key: String(node.key),
          label: node.label || node.title || 'Unknown',
        };

        if (node.children && node.children.length > 0) {
          const children = processNodes(node.children);
          if (children.length > 0) {
            option.children = children;
          }
        }
        return option;
      })
      .filter(node => node !== null);
  };

  // 基础根目录选项
  const rootOption = {
    key: '0',
    label: t('bookmarkManager.rootDirectory'),
  };

  const treeData = fullData.value.length > 0 ? fullData.value : bookmarkTree.value;
  return [rootOption, ...processNodes(treeData)];
});



// 编辑书签相关
const isEditDialogOpen = ref(false);
const isCreateMode = ref(false); // 是否为创建模式
const bookmarkType = ref<'bookmark' | 'folder'>('bookmark'); // 书签类型
const currentEditBookmark = ref({
	id: 0,
	title: '',
	url: '',
	folderId: '0' as string | number | undefined,
	iconJson: '',
} as { id: number; title: string; url: string; folderId?: string | number | undefined; iconJson?: string; });


// 右键菜单样式
const contextMenuStyle = computed(() => {
  // 菜单默认宽度和高度估算
  const menuWidth = 160; // 对应w-40
  const menuHeight = 80; // 估算两个菜单项的高度

  // 获取屏幕尺寸
  const screenWidth = window.innerWidth;
  const screenHeight = window.innerHeight;

  // 计算菜单位置，确保不超出屏幕
  let left = contextMenuX.value;
  let top = contextMenuY.value;

  // 如果菜单右侧会超出屏幕，调整左侧位置
  if (left + menuWidth > screenWidth) {
    left = screenWidth - menuWidth;
  }

  // 如果菜单底部会超出屏幕，调整顶部位置
  if (top + menuHeight > screenHeight) {
    top = screenHeight - menuHeight;
  }

  // 确保菜单位置不会小于0
  left = Math.max(0, left);
  top = Math.max(0, top);

  return {
    top: `${top}px`,
    left: `${left}px`,
  };
});

// 打开右侧列表右键菜单
function openContextMenu(event: MouseEvent, bookmark: Bookmark & { isFolder?: boolean }) {
	event.preventDefault()
	event.stopPropagation()
	isContextMenuOpen.value = true
	contextMenuX.value = event.clientX
	contextMenuY.value = event.clientY
	currentBookmark.value = bookmark
}

// 双击进入文件夹
// 移除未使用的enterFolder函数

// 左侧树节点右键菜单
function handleTreeContextMenu({ node, event }: { node: any; event: MouseEvent }) {
	event.preventDefault()
	event.stopPropagation()

	// 打开菜单
	isContextMenuOpen.value = true
	contextMenuX.value = event.clientX
	contextMenuY.value = event.clientY

	// 查找当前节点的父文件夹ID
	let parentFolderId = '0'; // 默认根目录

	// 如果是从bookmark对象中可以直接获取folderId
	if (node.bookmark?.folderId) {
		parentFolderId = node.bookmark.folderId;
	} else {
		// 否则尝试通过遍历树结构找到父节点
		function findParentId(treeNodes: any[], targetId: string | number): string {
			for (const item of treeNodes) {
				if (item.children) {
					// 检查当前节点的子节点中是否有目标节点
					const childFound = item.children.some((child: any) => String(child.key) === String(targetId));
					if (childFound) {
						return String(item.key); // 返回找到的父节点ID，确保是字符串类型
					}
					// 递归搜索子节点
					const parentId = findParentId(item.children, targetId);
					if (parentId !== '0') {
						return parentId; // parentId已经是字符串类型
					}
				}
			}
			return '0'; // 确保返回字符串类型
		}
		parentFolderId = findParentId(bookmarkTree.value, String(node.key));
	}

	// 当前节点信息
	currentBookmark.value = {
		id: node.key,
		title: node.label,
		url: node.bookmark?.url || '',
		folderId: parentFolderId, // 使用找到的父文件夹ID
		isFolder: !node.isLeaf,
	}
}

// 渲染标签函数，使用SVG图标
// 自定义折叠图标渲染函数
const renderExpandIcon = ({ option }: { option: TreeOption }) => {
	// 现在我们已经通过设置isLeaf属性来控制折叠图标显示
	// 这里保持简单，让Tree组件根据isLeaf属性自动决定
	return undefined;
};

const renderTreeLabel = ({ option }: { option: any }) => {
  // 检查是否是文件夹节点
  const isFolder = option.isFolder || (!option.isLeaf && !option.bookmark?.url);

  // 尝试从不同的可能属性获取标题，增加健壮性
  const nodeTitle = option.title || option.label || option.name || '未命名';

  // 创建内容数组
  const content = [];

  // 如果是文件夹节点，添加SVG图标
  if (isFolder) {
    // 检查当前节点是否被选中
    const isSelected = selectedKeysRef.value.includes(option.key);
    const iconColor = isSelected ? '#4285F4' : '#9CA3AF'; // 选中时蓝色，未选中时灰色

    content.push(
      h('svg', {
        xmlns: 'http://www.w3.org/2000/svg',
        class: 'w-5 h-5 inline-block mr-1',
        width: '24',
        height: '24',
        fill: iconColor,
        viewBox: '0 0 24 24'
      }, [
        h('path', {
          d: 'M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z'
        })
      ])
    );

    // 显示文件夹标题
    content.push(nodeTitle);
  } else {
    // 非文件夹节点显示图标和标题
    if (option.bookmark?.iconJson) {
      // 处理base64图标，确保有数据URL前缀
      let iconSrc = option.bookmark?.iconJson;
      if (!iconSrc.startsWith('data:')) {
        // 默认添加png格式的base64前缀
        iconSrc = 'data:image/png;base64,' + iconSrc;
      }
      content.push(
        h('img', {
          src: iconSrc,
          class: 'w-4 h-4 inline-block mr-1 rounded-full',
          alt: 'bookmark icon'
        })
      );
    } else {
      // 默认图标
      content.push(
        h('img', {
          src: 'https://www.google.com/s2/favicons?domain=' + option.bookmark?.url,
          class: 'w-5 h-5 inline-block mr-1 rounded-full',
          alt: 'default icon'
        })
      );
    }
    content.push(nodeTitle);
  }

  return h(
    'div',
    {
      class: 'px-1 py-0.5 rounded hover:bg-gray-100 cursor-default flex items-center text-gray-700 dark:text-white',
      draggable: true,
      onDragstart: (e: DragEvent) => {
        // 将树节点转换为可拖动的格式
        const treeItem = {
          id: option.key,
          title: option.label,
          url: option.bookmark?.url || '',
          isFolder: option.isFolder || false,
          folderId: option.rawNode?.parentId || '0',
          iconJson: option.bookmark?.iconJson || '',
          sort: option.sort || 0
        };
        handleDragStart(e, treeItem);
      },
      onDragend: (e: DragEvent) => {
        handleDragEnd(e);
      },
      onDragover: (e: DragEvent) => {
        e.preventDefault();
        if (e.dataTransfer) {
          e.dataTransfer.dropEffect = 'move';
        }
        // 添加悬停效果
        if (e.currentTarget instanceof HTMLElement) {
          e.currentTarget.classList.add('bg-blue-100', 'dark:bg-blue-900');
        }
      },
      onDragleave: (e: DragEvent) => {
        // 移除悬停效果
        if (e.currentTarget instanceof HTMLElement) {
          e.currentTarget.classList.remove('bg-blue-100', 'dark:bg-blue-900');
        }
      },
      onDrop: (e: DragEvent) => {
        e.preventDefault();
        // 移除悬停效果
        if (e.currentTarget instanceof HTMLElement) {
          e.currentTarget.classList.remove('bg-blue-100', 'dark:bg-blue-900');
        }
        // 将树节点转换为目标项格式
        const treeItem = {
          id: option.key,
          title: option.label,
          url: option.bookmark?.url || '',
          isFolder: option.isFolder || false,
          folderId: option.rawNode?.parentId || '0',
          label: option.label,
          iconJson: option.bookmark?.iconJson || '',
          sort: option.sort || 0
        };
        handleDrop(e, treeItem);
      },
      onContextmenu: (e: MouseEvent) => {
        e.preventDefault()
        e.stopPropagation()
        if (!isMobile.value) {
          handleTreeContextMenu({ node: option, event: e })
        }
      },
    },
    content
  );
}

function handleGlobalClick(event: MouseEvent) {
	const path = event.composedPath() as HTMLElement[]
	const clickedInsideMenu = path.some(
		(el) => el.classList && (el.classList.contains('context-menu') || el.closest('.custom-dropdown'))
	)
	if (!clickedInsideMenu) {
		closeContextMenu()
		isDropdownOpen.value = false
	}
}


// 关闭右键菜单
function closeContextMenu() {
	isContextMenuOpen.value = false;
}


// 打开书签URL
function openBookmark(bookmark: Bookmark) {
	if (bookmark.url && !bookmark.isFolder) {
		// 使用openUrlWithoutReferer在新标签页中打开书签URL，防止referer导致的403错误
		openUrlWithoutReferer(bookmark.url, '_blank');
	}
}

// 检查是否是子节点（防止循环移动）
function checkIsDescendant(parentId: string | number, childId: string | number): boolean {
	const findNode = (nodes: any[], targetId: string | number): any => {
		for (const node of nodes) {
			if (String(node.id || node.key) === String(targetId)) {
				return node;
			}
			if (node.children && node.children.length > 0) {
				const found = findNode(node.children, targetId);
				if (found) return found;
			}
		}
		return null;
	};

	const parentNode = findNode(fullData.value, parentId);
	if (!parentNode) return false;

	// 递归检查子节点中是否包含childId
	const checkChildren = (node: any): boolean => {
		if (!node.children || node.children.length === 0) return false;
		for (const child of node.children) {
			if (String(child.id || child.key) === String(childId)) {
				return true;
			}
			if (checkChildren(child)) return true;
		}
		return false;
	};

	return checkChildren(parentNode);
}

// 拖拽开始
function handleDragStart(event: DragEvent, item: any) {
	draggedItem.value = item;
	if (event.dataTransfer) {
		event.dataTransfer.effectAllowed = 'all'; // 允许所有效果，防止初始闪烁
		event.dataTransfer.dropEffect = 'move';
		event.dataTransfer.setData('text/plain', item.id.toString());
	}
	// 延迟添加视觉效果，防止影响拖拽初始化导致的闪烁
	if (event.currentTarget instanceof HTMLElement) {
		const element = event.currentTarget;
		setTimeout(() => {
			element.classList.add('opacity-50');
		}, 0);
	}
}

// 拖拽结束
function handleDragEnd(event: DragEvent) {
	// 移除拖拽时的视觉效果
	if (event.currentTarget instanceof HTMLElement) {
		event.currentTarget.classList.remove('opacity-50');
	}
	draggedItem.value = null;
}



// 拖拽悬停
const dragOverTarget = ref<string | number | null>(null);
// 拖拽插入位置
const dragInsertPosition = ref<'before' | 'after' | null>(null);
// 全局拖拽指示线位置 (top值)
const dragIndicatorTop = ref<number | null>(null);

function handleDragOver(event: DragEvent, item?: any) {
	// 阻止默认行为并设置允许放置
	event.preventDefault();
	// event.stopPropagation(); // 允许冒泡，以便容器也能处理dragover事件，防止禁用光标闪烁

	// 始终设置为move效果,避免显示禁用光标
	if (event.dataTransfer) {
		event.dataTransfer.dropEffect = 'move';
	}

	if (!item || !event.currentTarget) {
		dragOverTarget.value = null;
		dragInsertPosition.value = null;
		return;
	}

	const wrapper = event.currentTarget as HTMLElement;
	const rect = wrapper.getBoundingClientRect();
	const mouseY = event.clientY;
	const itemHeight = rect.height;
	const mousePosition = mouseY - rect.top;

	if (item && (item.isFolder || item.isFolder === true)) {
		// 文件夹逻辑：上1/3插入前，中1/3移动到文件夹，下1/3插入后
		if (mousePosition < itemHeight / 3) {
			// 上1/3 - 插入到文件夹前
			dragOverTarget.value = item.id;
			dragInsertPosition.value = 'before';
			dragIndicatorTop.value = wrapper.offsetTop;
		} else if (mousePosition > (2 * itemHeight) / 3) {
			// 下1/3 - 插入到文件夹后
			dragOverTarget.value = item.id;
			dragInsertPosition.value = 'after';
			dragIndicatorTop.value = wrapper.offsetTop + wrapper.offsetHeight;
		} else {
			// 中间 - 移动到文件夹内
			dragOverTarget.value = item.id;
			dragInsertPosition.value = null;
			dragIndicatorTop.value = null;
		}
	} else {
		// 普通项逻辑：上半部分插入前，下半部分插入后
		dragOverTarget.value = item.id;
		if (mousePosition < itemHeight / 2) {
			dragInsertPosition.value = 'before';
			dragIndicatorTop.value = wrapper.offsetTop;
		} else {
			dragInsertPosition.value = 'after';
			dragIndicatorTop.value = wrapper.offsetTop + wrapper.offsetHeight;
		}
	}
}

// 容器级别的dragover处理,防止在间隙处显示禁用光标
function handleContainerDragOver(event: DragEvent) {
	event.preventDefault();
	if (event.dataTransfer) {
		event.dataTransfer.dropEffect = 'move';
	}
}


function handleDragLeave(event: DragEvent) {
	// 阻止默认行为
	event.preventDefault();
	// 不在dragleave时清除状态,避免在元素间移动时出现禁用光标
	// 状态会在dragend时清除
	// if (event.currentTarget === event.target || !(event.currentTarget as HTMLElement)?.contains(event.relatedTarget as Node)) {
	// 	dragOverTarget.value = null;
	// 	dragInsertPosition.value = null;
	// }
}

// 定义一个响应式数组用于前端临时排序展示
const sortedItems = ref<any[]>([]);

// 监听filteredBookmarks变化，自动更新sortedItems
watch(() => filteredBookmarks.value, (newValue) => {
  // 直接使用过滤后的数据，不做额外排序
  sortedItems.value = [...newValue];
}, { immediate: true, deep: true });

// 在组件挂载时，确保立即执行一次过滤和排序
watch(fullData, () => {
  // 强制使用根目录过滤后的数据
  setTimeout(() => {
    sortedItems.value = [...filteredBookmarks.value];
  }, 0);
}, { immediate: true, deep: true });

// 监听selectedFolder变化,同步左侧树的展开状态和选中状态
watch(selectedFolder, (newFolderId) => {
  if (newFolderId && newFolderId !== '0') {
    // 构建从根到当前文件夹的路径
    const buildPath = (nodes: any[], targetId: string, path: string[] = []): string[] | null => {
      for (const node of nodes) {
        const nodeId = String(node.key);
        const currentPath = [...path, nodeId];

        if (nodeId === targetId) {
          return currentPath;
        }

        if (node.children && node.children.length > 0) {
          const found = buildPath(node.children, targetId, currentPath);
          if (found) return found;
        }
      }
      return null;
    };

    const path = buildPath(bookmarkTree.value, newFolderId);
    if (path) {
      // 展开路径上的所有父文件夹(不包括当前文件夹本身)
      const parentPath = path.slice(0, -1);
      defaultExpandedKeys.value = [...new Set([...defaultExpandedKeys.value, ...parentPath])];

      // 高亮当前文件夹
      selectedKeysRef.value = [newFolderId];
    }
  } else if (newFolderId === '0') {
    // 如果是根目录,清除选中状态
    selectedKeysRef.value = [];
  }
}, { immediate: false });


// 处理拖拽放置 - 支持移动到文件夹或排序
async function handleDrop(event: DragEvent, targetItem: any) {

	event.preventDefault();

	// 移除拖拽时的视觉效果
	if (event.currentTarget instanceof HTMLElement) {
		event.currentTarget.classList.remove('opacity-50');
	}

	// 确保拖拽项存在且不是同一个项目
	if (!draggedItem.value || draggedItem.value.id === targetItem.id) {

		draggedItem.value = null;
		return;
	}

	const draggedItemData = draggedItem.value;


	// 检查目标项是否为文件夹
	const isTargetFolder = targetItem.isFolder || targetItem.isFolder === true;

	// 如果目标是文件夹且没有明确的插入位置(before/after),则将拖动的项移动到该文件夹内
	// 如果有明确的插入位置,则表示要在文件夹前后进行排序,而不是移入文件夹
	if (isTargetFolder && !dragInsertPosition.value) {


		// 检查是否试图将文件夹移动到自己的子文件夹中（防止循环）
		if (draggedItemData.isFolder) {
			const isDescendant = checkIsDescendant(draggedItemData.id, targetItem.id);
			if (isDescendant) {
				ms.warning('不能将文件夹移动到自己的子文件夹中');
				draggedItem.value = null;
				return;
			}
		}

		// 获取目标文件夹的ID
		const targetFolderId = String(targetItem.id);

		// 检查是否试图移动到自己的位置
		const draggedParentId = String(draggedItemData.folderId || '0');
		if (draggedParentId === targetFolderId) {

			draggedItem.value = null;
			return;
		}

		// 获取目标文件夹下的最大sort值
		const targetFolderItems = allItems.value.filter(item =>
			String(item.folderId || '0') === targetFolderId
		);
		const maxSort = targetFolderItems.length > 0
			? Math.max(...targetFolderItems.map(item => item.sort || 0))
			: 0;

		// 准备更新数据
			const updateData = {
				id: Number(draggedItemData.id),
				title: draggedItemData.title,
				url: draggedItemData.isFolder ? draggedItemData.title : (draggedItemData.url || ''),
				parentId: Number(targetFolderId), // 使用targetFolderId作为parentId
				sort: maxSort + 1,
				lanUrl: draggedItemData.lanUrl || '',
				openMethod: draggedItemData.openMethod || 0,
				icon: draggedItemData.icon || null,
				iconJson: draggedItemData.iconJson || ''
			};

		try {
			// 调用更新接口
			const response = await update(updateData);
			if (response && response.code === 0) {
				// 更新本地缓存
				updateCacheAfterUpdate(response.data);
			}
		} catch (error) {
			console.error('移动书签失败:', error);
			ms.error(`${t('bookmarkManager.moveFailed') || '移动失败'}: ${(error as Error).message || ''}`);
		}

		draggedItem.value = null;
		return;
	}

	// 其他情况:执行排序逻辑(包括文件夹与文件夹之间的排序、书签与书签之间的排序)
	// 注意:如果是拖到文件夹上且dragInsertPosition为null,说明是要移动到文件夹内,已经在上面处理了
	// 这里只处理排序的情况(dragInsertPosition为'before'或'after')
	const draggedFolderId = String(draggedItemData.folderId || '0');
	const targetFolderId = String(targetItem.folderId || '0');

	// 特殊情况:如果目标是文件夹,且拖拽项不在同一父文件夹,则视为移动到该文件夹
	// (即使dragInsertPosition不为null,也应该移动到文件夹内,而不是排序)
	if (isTargetFolder && draggedFolderId !== targetFolderId) {
		// 获取目标文件夹的ID
		const targetFolderIdValue = String(targetItem.id);

		// 检查是否试图将文件夹移动到自己的子文件夹中（防止循环）
		if (draggedItemData.isFolder) {
			const isDescendant = checkIsDescendant(draggedItemData.id, targetItem.id);
			if (isDescendant) {
				ms.warning('不能将文件夹移动到自己的子文件夹中');
				draggedItem.value = null;
				return;
			}
		}

		// 获取目标文件夹下的最大sort值
		const targetFolderItems = allItems.value.filter(item =>
			String(item.folderId || '0') === targetFolderIdValue
		);
		const maxSort = targetFolderItems.length > 0
			? Math.max(...targetFolderItems.map(item => item.sort || 0))
			: 0;

		// 准备更新数据
			const updateData = {
				id: Number(draggedItemData.id),
				title: draggedItemData.title,
				url: draggedItemData.isFolder ? draggedItemData.title : (draggedItemData.url || ''),
				parentId: Number(targetFolderIdValue), // 使用targetFolderId作为parentId
				sort: maxSort + 1,
				lanUrl: draggedItemData.lanUrl || '',
				openMethod: draggedItemData.openMethod || 0,
				icon: draggedItemData.icon || null,
				iconJson: draggedItemData.iconJson || ''
			};

		try {
			// 调用更新接口
			const response = await update(updateData);
			if (response && response.code === 0) {
				// 更新本地缓存
				updateCacheAfterUpdate(response.data);
			}
		} catch (error) {
			console.error('移动书签失败:', error);
			ms.error(`${t('bookmarkManager.moveFailed') || '移动失败'}: ${(error as Error).message || ''}`);
		}

		draggedItem.value = null;
		return;
	}

	// 确保它们在同一个文件夹中才能排序
	if (draggedFolderId !== targetFolderId) {
		ms.warning('只能在同一文件夹内拖拽排序');
		draggedItem.value = null;
		return;
	}

	const currentFolderItems = allItems.value.filter(item =>
		String(item.folderId || '0') === draggedFolderId
	);

	// 找到拖拽项和目标项在当前文件夹中的索引
	const draggedIndex = currentFolderItems.findIndex(item => String(item.id) === String(draggedItemData.id));
	const targetIndex = currentFolderItems.findIndex(item => String(item.id) === String(targetItem.id));



	if (draggedIndex !== -1 && targetIndex !== -1) {
		// 实现插入排序的逻辑
		// 步骤1：获取当前文件夹的所有项目并按sort排序
		const sortedFolderItems = [...currentFolderItems].sort((a, b) => (a.sort || 0) - (b.sort || 0));

		// 步骤2：找到拖拽项和目标项在排序后的索引
		const sortedDraggedIndex = sortedFolderItems.findIndex(item => String(item.id) === String(draggedItemData.id));
		const sortedTargetIndex = sortedFolderItems.findIndex(item => String(item.id) === String(targetItem.id));

		// 步骤3：根据插入位置确定新的索引
		let newIndex: number;
		if (dragInsertPosition.value === 'before') {
			newIndex = sortedTargetIndex;
		} else if (dragInsertPosition.value === 'after') {
			newIndex = sortedTargetIndex + 1;
		} else {
			// 默认行为 - 如果没有插入位置，则根据原始索引决定
			newIndex = draggedIndex < targetIndex ? targetIndex : targetIndex + 1;
		}

		// 步骤4：从排序数组中移除拖拽项并插入到新位置
		const updatedItems = [...sortedFolderItems];
		updatedItems.splice(sortedDraggedIndex, 1);
		// 修正插入索引：如果新索引在拖拽项之后，需要减1（因为移除操作使后续项前移了）
		const adjustedNewIndex = newIndex > sortedDraggedIndex ? newIndex - 1 : newIndex;
		updatedItems.splice(adjustedNewIndex, 0, draggedItemData);

		// 步骤5：为当前文件夹所有项分配连续且唯一的 sort 值 (1, 2, 3...)
		const sortItems = updatedItems.map((item, index) => ({
			id: Number(item.id),
			sort: index + 1,
			parentId: Number(draggedFolderId),
		}));

		// 先执行乐观更新，平滑刷新本地树形与缓存
		updateCacheAfterSort(sortItems);

		// 异步同步到服务端保存
		try {
			const res = await saveBookmarkSort(sortItems);
			if (res && res.code !== 0) {
				console.error('服务器排序保存失败:', res);
				ms.error('排序保存失败,已恢复');
				refreshBookmarks(true);
			}
		} catch (error) {
			console.error('服务器排序保存失败,回滚本地数据:', error);
			ms.error('排序保存失败,已恢复');
			refreshBookmarks(true);
		}
	} else {
		ms.warning('排序更新失败：找不到相关项目');
	}

	// 重置拖拽状态
	draggedItem.value = null;

}

// 处理编辑书签
function handleEditBookmark() {
  if (currentBookmark.value) {
    currentEditBookmark.value = {
      ...currentBookmark.value,
    };
    // 根据是否为文件夹设置书签类型
    bookmarkType.value = currentBookmark.value.isFolder ? 'folder' : 'bookmark';
    // 设置为修改模式
    isCreateMode.value = false;
    isEditDialogOpen.value = true;
  }
  isContextMenuOpen.value = false;
}

// 处理删除书签
function handleDeleteBookmark() {
	if (currentBookmark.value) {
		deleteBookmark(currentBookmark.value);
	}
	isContextMenuOpen.value = false;
}

// 创建新书签
function createNewBookmark() {
	// 重置表单
	// 设置默认父文件夹为当前选中的文件夹
	const defaultFolderId = (selectedFolder.value && selectedFolder.value !== '0') ? selectedFolder.value : '0';

	currentEditBookmark.value = {
		id: 0,
		title: '',
		url: '',
		folderId: defaultFolderId,
	};
	// 设置为创建模式
	isCreateMode.value = true;
	// 打开编辑对话框
	isEditDialogOpen.value = true;
}

// 关闭编辑对话框
function closeEditDialog() {
	isEditDialogOpen.value = false;
}

// 保存书签修改
async function saveBookmarkChanges() {
	try {
		// 数据验证
		if (!currentEditBookmark.value.title.trim()) {
			ms.error(t('bookmarkManager.titleRequired'));
			return;
		}

		// URL验证和处理 - 只有在编辑书签或创建书签时才需要URL
		// 修改模式
		// 根据是否为文件夹构建不同的更新数据
		const isFolderItem = bookmarkType.value === 'folder';

		// 如果不是文件夹且需要URL但未提供，显示错误
		if (!isFolderItem && !currentEditBookmark.value.url.trim()) {
			ms.error(t('bookmarkManager.urlRequired'));
			return;
		}

		// 只有书签才需要添加URL协议检查和补充
		if (!isFolderItem && currentEditBookmark.value.url.trim()) {
			let url = currentEditBookmark.value.url.trim();
			if (!url.startsWith('http://') && !url.startsWith('https://')) {
				url = 'https://' + url;
				currentEditBookmark.value.url = url;
			}
		}

		// 移除favicon自动获取逻辑，避免CORS问题
		// 前端将使用项目内置的方式展示favicon

		// 根据模式决定调用哪个接口
		if (isCreateMode.value) {
			// 创建新模式
			// 根据当前选择的folderId获取parentId
				const parentId = currentEditBookmark.value.folderId ? Number(currentEditBookmark.value.folderId) : 0;

			const createData = {
				title: currentEditBookmark.value.title,
				url: '',
				// 使用文件夹ID作为parentId
				parentId: parentId,
				sort: 9999,
				lanUrl: '',
				icon: null,
				openMethod: 0,
				iconJson: currentEditBookmark.value.iconJson || ''
			};

			// 根据类型添加相应的字段
			if (bookmarkType.value === 'folder') {
				// 文件夹：设置isFolder为1，URL设为标题
				Object.assign(createData, {
					isFolder: 1,
					url: currentEditBookmark.value.title
				});
			} else {
				// 书签：添加URL
				Object.assign(createData, {
					isFolder: 0,
					url: currentEditBookmark.value.url
				});
			}

			// 调用添加接口
			const createResponse = await addBookmark(createData);

			// 检查响应状态
				if (createResponse && createResponse.code === 0) {
					// 更新本地缓存数据，而不是强制刷新
					updateCacheAfterAdd(createResponse.data);
					isEditDialogOpen.value = false;
					isCreateMode.value = false;
				} else {
					ms.error(`${t('bookmarkManager.createFailed')} ${createResponse?.msg || t('bookmarkManager.unknownError')}`);
				}
		} else {
			// 修改模式
			// 根据是否为文件夹构建不同的更新数据
			const isFolderItem = currentBookmark.value?.isFolder;
			// 直接使用folderId作为parentId
			let parentId = 0;
			const selectedFolderId = currentEditBookmark.value.folderId ? currentEditBookmark.value.folderId.toString() : '0';
			if (selectedFolderId !== '0') {
				parentId = Number(selectedFolderId);
			}
			const updateData = {
				id: Number(currentEditBookmark.value.id),
				title: currentEditBookmark.value.title,
				parentId: parentId,
				sort: 9999,
				lanUrl: '',
				icon: null,
				openMethod: 0,
				url: '', // 添加url属性以满足bookmarkInfo类型要求
				iconJson: currentEditBookmark.value.iconJson || ''
			};

			// 文件夹特殊处理：将URL设置为标题或空
			if (isFolderItem) {
				updateData.url = currentEditBookmark.value.title;
			} else {
				// 书签正常处理URL
				updateData.url = currentEditBookmark.value.url;
			}

			const updateResponse = await update(updateData);

				// 检查响应状态
				if (updateResponse && updateResponse.code === 0) {
					// 更新本地缓存数据，而不是强制刷新
					updateCacheAfterUpdate(updateResponse.data);
					isEditDialogOpen.value = false;
				} else {
					ms.error(`${t('bookmarkManager.updateFailed')} ${updateResponse?.msg || t('bookmarkManager.unknownError')}`);
				}
			}
	} catch (error) {
		console.error('保存书签失败', error);
		ms.error(`${t('bookmarkManager.updateFailed')} ${(error as Error).message || t('common.networkError')}`);
	}
}

// 删除书签或文件夹
async function deleteBookmark(bookmark: Bookmark) {
	// 根据是否为文件夹显示不同的确认消息
	const confirmMessage = bookmark.isFolder
		? t('bookmarkManager.deleteFolderConfirm').replace('name', "【"+bookmark.title+"】")
		 : t('bookmarkManager.deleteBookmarkConfirm').replace('name', "【"+bookmark.title+"】");

	// 直接使用从apiMessage导入的dialog对象
	dialog.warning({
		title: t('bookmarkManager.confirmDelete'),
		content: confirmMessage,
		positiveText: t('bookmarkManager.confirm'),
		negativeText: t('bookmarkManager.cancel'),
		onPositiveClick: async () => {
			 try {
				 const response = await deletes([Number(bookmark.id)]);
				 if (response.code === 0) {
					 // 重置选中状态
					 if (selectedBookmarkId.value === bookmark.id.toString()) {
						 selectedBookmarkId.value = '';
					 }
					 // 如果删除的是当前选中的文件夹，清空选中状态
					 if (bookmark.isFolder && selectedFolder.value === bookmark.id.toString()) {
						 selectedFolder.value = '0';
					 }

					 // 更新本地缓存数据，而不是强制刷新
					 updateCacheAfterDelete(Number(bookmark.id), bookmark.isFolder || false);

				 } else {
					 ms.error(`${t('common.failed')}: ${response.msg}`);
				 }
			 } catch (error) {
				 ms.error(`${t('common.failed')} ${(error as Error).message || t('bookmarkManager.unknownError')}`);
			 }
		 }
	});
}

// 触发导入书签
function triggerImportBookmarks() {
	fileInput.value?.click();
}

// 处理文件选择变化
function handleFileChange(event: Event) {
	const target = event.target as HTMLInputElement;
	const file = target.files?.[0];

	if (file) {
		uploadLoading.value = true;
		const reader = new FileReader();
		reader.onload = (e) => {
			if (e.target?.result) {
				jsonData.value = e.target.result as string;
				importCheck(file.name);
			} else {
				ms.error(`${t('common.failed')}: ${t('common.repeatLater')}`);
				uploadLoading.value = false;
			}
		};
		reader.onerror = () => {
			ms.error(`${t('common.failed')}: ${t('common.fileReadError')}`);
			uploadLoading.value = false;
		};
		reader.readAsText(file);
	} else {
		uploadLoading.value = false;
	}

	// 清空input的值，以便可以重复选择同一个文件
	target.value = '';
}

// 验证导入文件
function importCheck(fileName: string) {
	importWarning.value = [];

	try {
		if (fileName.endsWith('.html')) {
			// 直接将HTML内容传递给后端解析
			importBookmarksToServerWithHTML(jsonData.value!);
		} else {
			ms.error(t('bookmarkManager.onlySupportHtml'));
		}
	} catch (error) {
		ms.error(`${t('common.failed')}: ${(error as Error).message || t('common.unknownError')}`);
	} finally {
		uploadLoading.value = false;
	}
}

// 导入HTML书签到服务器
async function importBookmarksToServerWithHTML(htmlContent: string) {
	uploadLoading.value = true;
	try {
		// 将HTML内容传递给后端
		// 注意：后端期望的字段名是HtmlContent（首字母大写）
		const response = await addMultipleBookmarks({
			HtmlContent: htmlContent
		} as any);

		if (response.code === 0) {
			ss.remove(BOOKMARKS_CACHE_KEY)
			// 刷新书签列表
			await refreshBookmarks();
		} else {
			ms.error(`${t('common.failed')}: ${response.msg}`);
		}
	} catch (error) {
		ms.error(`${t('common.failed')}: ${(error as Error).message || t('common.unknownError')}`);
	} finally {
		uploadLoading.value = false;
	}
}

// 导出书签为HTML格式
function exportBookmarks() {
	// 准备HTML文件头
	let html = `<!DOCTYPE NETSCAPE-Bookmark-file-1>
`;
	html += `<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
`;
	html += `<TITLE>Bookmarks</TITLE>
`;
	html += `<H1>Bookmarks</H1>
`;
	html += `<DL><p>
`;

	// 转换书签树为HTML
	function convertTreeToHtml(nodes: TreeOption[], level: number = 0): string {
		let result = ``;

		for (const node of nodes) {
			if (node.isFolder) {
				// 创建文件夹
				result += `${'  '.repeat(level)}<DT><H3>${node.label}</H3>
`;
				result += `${'  '.repeat(level)}<DL><p>
`;
				// 递归处理子节点
				result += convertTreeToHtml(node.children, level + 1);
				result += `${'  '.repeat(level)}</DL><p>
`;
			} else if (node.bookmark) {
				// 创建书签
			const title = node.bookmark.title || '';
			const url = node.bookmark.url || '';
			const icon = node.bookmark.iconJson || node.rawNode?.iconJson || node.rawNode?.icon || '';
			result += `${'  '.repeat(level)}<DT><A HREF="${url}" ${icon ? `ICON="${icon}"` : ''}>${title}</A>
`;
			}
		}

		return result;
	}

	// 转换所有节点
	html += convertTreeToHtml(fullData.value.length > 0 ? fullData.value : bookmarkTree.value);

	// 关闭根DL标签
	html += `</DL><p>
`;

	// 下载文件
	const blob = new Blob([html], { type: 'text/html;charset=utf-8' });
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = 'bookmarks.html';
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}

// 不再需要此函数，后端已直接处理排序值


// 过滤函数：从完整树结构中只保留文件夹
function filterFoldersOnly(nodes: TreeOption[]): TreeOption[] {

	if (!Array.isArray(nodes)) {
		console.error('filterFoldersOnly: 输入数据不是数组:', nodes);
		return [];
	}

	const result: TreeOption[] = [];
	for (const node of nodes) {
		if (node && node.isFolder) {
			const folderNode: TreeOption = {
				...node,
				bookmark: undefined,
				children: []
			};

			if (Array.isArray(node.children) && node.children.length > 0) {
				const folderChildren = node.children.filter(child => child && child.isFolder);
				if (folderChildren.length > 0) {
					folderNode.children = filterFoldersOnly(folderChildren);
					// 子节点不为空 → 启用折叠，设置isLeaf为false
					folderNode.isLeaf = false;
					folderNode.disabledExpand = false;
				} else {
					folderNode.children = [];
					// 子节点为空 → 禁用折叠，设置isLeaf为true
					folderNode.isLeaf = true;
					folderNode.disabledExpand = true;
				}
			} else {
				// 没有子节点 → 设置isLeaf为true
				folderNode.isLeaf = true;
				folderNode.disabledExpand = true;
			}

			result.push(folderNode);
		}
	}

	return result;
}

// 刷新书签列表
	async function refreshBookmarks(forceRefresh = false) {
	try {
		// 首先尝试从BOOKMARKS_CACHE_KEY获取完整数据（与首页缓存一致）
	let folderTreeData: TreeOption[] = [];

		// 强制刷新时清除缓存（与首页保持一致）
		if (forceRefresh) {
			ss.remove(BOOKMARKS_CACHE_KEY);
		}

		// 检查缓存是否存在
		if (!forceRefresh) {
			const cachedData = ss.get(BOOKMARKS_CACHE_KEY);
			if (cachedData) {
				// 与首页保持一致的处理逻辑
				let treeDataResult: TreeOption[] = [];
				const data = cachedData;

				if (Array.isArray(data) && data.length > 0 && 'children' in data[0]) {
					// 已经是树形结构，转换为前端需要的格式
					treeDataResult = convertServerTreeToFrontendTree(data);
				} else if ((data as { list?: any[] }).list && Array.isArray((data as { list?: any[] }).list)) {
					// 后端返回的是带list字段的结构
					const serverBookmarks = (data as { list: any[] }).list;
					if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
						// list字段中已经是树形结构
						treeDataResult = convertServerTreeToFrontendTree(serverBookmarks);
					} else {
						// 构建树形结构
						treeDataResult = buildBookmarkTree(serverBookmarks);
					}
				} else {
					// 作为列表数据构建树形结构
					treeDataResult = buildBookmarkTree(Array.isArray(data) ? data : []);
				}

				// 先更新完整数据（树结构）
			fullData.value = treeDataResult;
			// 保存原始数据缓存
			rawDataCache.value = cachedData;

			// 过滤文件夹
				folderTreeData = filterFoldersOnly(treeDataResult);

				// 严格检查过滤结果
if (Array.isArray(folderTreeData) && folderTreeData.length > 0) {
				// 深拷贝避免引用问题
				bookmarkTree.value = JSON.parse(JSON.stringify(folderTreeData));

				// 默认展开第一栏文件夹
				if (folderTreeData.length > 0 && folderTreeData[0].key) {
					defaultExpandedKeys.value = [folderTreeData[0].key];
				} else {
					defaultExpandedKeys.value = [];
				}

				// 保留全局变量更新
				if (globalThis) {
					Object.defineProperty(globalThis, '__bookmarksFullData', { value: treeDataResult, configurable: true });
				}

				return;
			} else {
				// 设置空数组确保组件能正确渲染
				bookmarkTree.value = [];
				defaultExpandedKeys.value = [];
				// 如果缓存数据有内容，返回以避免请求服务器
				if (Array.isArray(treeDataResult) && treeDataResult.length > 0) {
					return;
				}
			}
			}

		}

		// 缓存不存在或强制刷新时从服务器获取数据

const response = await getBookmarksList();
if (response.code === 0) {
	// 后端返回格式为 { data: { list: [], count: number }, code: 0, msg: "" }
	// 获取实际的书签数据列表
	let serverBookmarks = [];

	// 检查响应结构
	if (response.data) {
		// 如果data本身是数组，直接使用
		if (Array.isArray(response.data)) {
			serverBookmarks = response.data;
		}
		// 如果data是包含list字段的对象，使用list字段
				else if (Array.isArray((response.data as any).list)) {
					serverBookmarks = (response.data as any).list;
				}
	}

	let treeDataResult = [];

	// 检查是否已经是树形结构（直接包含children字段）
	if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
		// 已经是树形结构，转换为前端需要的格式

		treeDataResult = convertServerTreeToFrontendTree(serverBookmarks);
	} else {
		// 构建树形结构

		treeDataResult = buildBookmarkTree(serverBookmarks);
	}

	// 存储到缓存（与首页保持一致，缓存原始服务器数据）
	// 缓存response.data，确保包含完整的服务器响应数据
	ss.set(BOOKMARKS_CACHE_KEY, response.data);

	// 更新完整数据（树结构）
	fullData.value = treeDataResult;

	// 过滤文件夹
	const folderTreeData = filterFoldersOnly(treeDataResult);

			// 检查过滤结果
			if (Array.isArray(folderTreeData) && folderTreeData.length > 0) {
				// 深拷贝避免引用问题
				bookmarkTree.value = JSON.parse(JSON.stringify(folderTreeData));

				// 默认展开第一栏文件夹
				if (folderTreeData.length > 0 && folderTreeData[0].key) {
					defaultExpandedKeys.value = [folderTreeData[0].key];
				} else {
					defaultExpandedKeys.value = [];
				}

				// 保留全局变量更新
				if (globalThis) {
					Object.defineProperty(globalThis, '__bookmarksFullData', { value: treeDataResult, configurable: true });
				}
			} else {
				// 设置空数组确保组件能正确渲染
				bookmarkTree.value = [];
				sortedItems.value = [];
				defaultExpandedKeys.value = [];
			}
		} else {
			// 服务器返回错误
			console.error('获取书签数据失败:', response);
			// 设置空数据确保组件能正确渲染
			fullData.value = [];
			bookmarkTree.value = [];
			sortedItems.value = [];
			defaultExpandedKeys.value = [];

			// 全局变量更新为空
			if (globalThis) {
				Object.defineProperty(globalThis, '__bookmarksFullData', { value: [], configurable: true });
			}
		}
	} catch (error) {
		console.error('刷新书签列表发生异常:', error);
		// 清空所有相关数据
		bookmarkTree.value = [];
		fullData.value = [];
		sortedItems.value = [];
		defaultExpandedKeys.value = [];
	}
 }

// 将服务器返回的树形结构或首页缓存的TreeOption数据转换为前端组件需要的格式
function convertServerTreeToFrontendTree(serverTree: any[]): TreeOption[] {
	const result: TreeOption[] = [];

	function processNode(node: any): TreeOption {
		// 检查是否已经是TreeOption对象（首页缓存）
		const isTreeOption = node.hasOwnProperty('key') && node.hasOwnProperty('label');

		let nodeKey: string;
		let isFolder: boolean;
		let title: string;
		let url: string;
		let iconJson: string;
		let children: any[];
		let rawNode: any;

		if (isTreeOption) {
			// 处理首页缓存的TreeOption数据
			nodeKey = String(node.key);
			isFolder = node.isFolder;
			title = node.label;
			url = node.bookmark?.url || '';
			iconJson = node.bookmark?.iconJson || '';
			children = node.children || [];
			rawNode = {
				...node.rawNode,
				parentId: node.rawNode?.parentId || '0'
			};
		} else {
			// 处理服务器原始数据
			nodeKey = String(node.id);
			isFolder = node.isFolder === 1;
			title = node.title;
			url = node.url || '';
			iconJson = node.iconJson || '';
			children = node.children || [];
			rawNode = {
				...node,
				parentId: node.parentId || '0'
			};

			// 优化：如果是书签，从rawNode中移除iconJson以节省缓存空间
			if (!isFolder && rawNode.iconJson) {
				delete rawNode.iconJson;
			}
		}

		const hasChildren = Array.isArray(children) && children.length > 0;

		// 构建基本节点结构
		const frontendNode: TreeOption = {
    key: nodeKey,
    label: title || '未命名',
    isLeaf: !isFolder || !hasChildren,
    isFolder: isFolder,
    bookmark: isFolder ? undefined : { id: node.id || nodeKey, title: title, url: url || '', iconJson: iconJson || '', sort: node.sort || 0 },
    children: [],
    rawNode: rawNode,
    disabledExpand: !hasChildren,
    sort: node.sort || rawNode.sort || 0
};

		// 递归处理子节点
		if (hasChildren) {
			frontendNode.children = children.map((child: any) => processNode(child));
			// 按sort值对children排序
			frontendNode.children.sort((a, b) => {
				const sortA = a.sort ?? a.bookmark?.sort ?? a.rawNode?.sort ?? 0;
				const sortB = b.sort ?? b.bookmark?.sort ?? b.rawNode?.sort ?? 0;
				return sortA - sortB;
			});
		}

		return frontendNode;
	}

	// 处理根节点
	for (const node of serverTree) {
		result.push(processNode(node));
	}

	// 按sort值对根节点排序
	result.sort((a, b) => {
		const sortA = a.sort ?? a.bookmark?.sort ?? a.rawNode?.sort ?? 0;
		const sortB = b.sort ?? b.bookmark?.sort ?? b.rawNode?.sort ?? 0;
		return sortA - sortB;
	});

	return result;
}

function buildBookmarkTree(bookmarks: any[]): TreeOption[] {
	// 创建节点映射，用于快速查找
	const nodeMap = new Map<string, TreeOption>();
	const rootNodes: TreeOption[] = [];

	// 第一步：创建所有节点并添加到映射
	for (const bookmark of bookmarks) {
		// 处理两种可能的节点结构：
		// 1. 服务器原始数据格式 (id, title, isFolder, url, iconJson)
		// 2. 前端节点格式 (key, label, isFolder, bookmark)
		const isFrontendFormat = bookmark.hasOwnProperty('key') && bookmark.hasOwnProperty('label');

		// 提取基本属性
		const nodeId = isFrontendFormat ? bookmark.key : bookmark.id || bookmark.Key || '0';
		const nodeKey = String(nodeId);
		const title = isFrontendFormat ? bookmark.label : bookmark.title;
		const isFolder = isFrontendFormat ? bookmark.isFolder : (bookmark.isFolder === 1);
		const url = isFrontendFormat ? (bookmark.bookmark?.url || '') : bookmark.url || '';
		const iconJson = isFrontendFormat ? (bookmark.bookmark?.iconJson || '') : bookmark.iconJson || '';
		// 使用parentId关联父子集
		const parentId = isFrontendFormat ? (bookmark.rawNode?.parentId || bookmark.ParentId || '0') : (bookmark.parentId || bookmark.ParentId || '0');
		const sort = isFrontendFormat ? (bookmark.bookmark?.sort || 0) : (bookmark.sort || 0);


		// 优化：构建rawNode时，如果是书签，移除iconJson以节省缓存空间
		const rawNodeData = {
			...bookmark,
			parentId: parentId
		};
		if (!isFolder && rawNodeData.iconJson) {
			delete rawNodeData.iconJson;
		}

		const node: TreeOption = {
			key: nodeKey,
			label: title || '未命名',
			isLeaf: !isFolder,
			isFolder: isFolder,
			bookmark: isFolder ? undefined : {
				id: nodeId,
				title: title || '未命名',
				url: url,
				iconJson: iconJson,
				sort: sort
			},
			rawNode: rawNodeData,
			children: [],
			disabledExpand: true
		};
		nodeMap.set(nodeKey, node);
	}

	// 第二步：构建树形结构
	for (const bookmark of bookmarks) {
		const isFrontendFormat = bookmark.hasOwnProperty('key') && bookmark.hasOwnProperty('label');
		const nodeKey = String(bookmark.id || bookmark.Key || '0');
		const node = nodeMap.get(nodeKey);
		if (!node) continue;

		// 获取父节点ID，只使用parentId
		const parentId = isFrontendFormat ? (bookmark.rawNode?.parentId || bookmark.ParentId || '0') : (bookmark.parentId || bookmark.ParentId || '0');
		const parentKey = String(parentId);

		if (parentId === '0' || parentId === '' || parentId === 'null' || parentId === null || parentId === undefined) {
			// 根节点
			rootNodes.push(node);
		} else {
			// 查找父节点并建立关系
			const parentNode = nodeMap.get(parentKey);
			if (parentNode && parentNode.key !== node.key) {
				parentNode.children.push(node);
				// 按sort值排序子节点
				parentNode.children.sort((a, b) => {
					const sortA = a.sort ?? a.bookmark?.sort ?? a.rawNode?.sort ?? 0;
					const sortB = b.sort ?? b.bookmark?.sort ?? b.rawNode?.sort ?? 0;
					return sortA - sortB;
				});
				parentNode.disabledExpand = false;
				parentNode.isLeaf = false;
			} else {
				// 如果父节点不存在，作为根节点
				rootNodes.push(node);
			}
		}
	}

	// 按sort值排序根节点
	rootNodes.sort((a, b) => {
		const sortA = a.sort ?? a.bookmark?.sort ?? a.rawNode?.sort ?? 0;
		const sortB = b.sort ?? b.bookmark?.sort ?? b.rawNode?.sort ?? 0;
		return sortA - sortB;
	});

	// 第三步：确保文件夹正确设置isLeaf属性
	for (const node of nodeMap.values()) {
		if (node.isFolder) {
			node.isLeaf = node.children.length === 0;
		}
	}

	return rootNodes;
}

// 更新本地缓存数据的辅助函数
// 将书签数据转换为缓存格式
function convertBookmarkToCacheNode(bookmark: any): any {
	const isFolder = bookmark.isFolder === 1 || bookmark.isFolder === true;
	return {
		id: bookmark.id,
		title: bookmark.title,
		url: bookmark.url || '',
		iconJson: bookmark.iconJson || '',
		isFolder: isFolder ? 1 : 0,
		parentId: bookmark.parentId || '0',
		sort: bookmark.sort || 0,
		children: []
	};
}

// 在树中查找节点（支持通过id或title查找）
function findNodeInTree(nodes: any[], id: string | number): any {
	for (const node of nodes) {
		// 通过id或key匹配
		if (String(node.id || node.key) === String(id)) {
			return node;
		}
		if (node.children && node.children.length > 0) {
			const found = findNodeInTree(node.children, id);
			if (found) return found;
		}
	}
	return null;
}

// 从树中删除节点
function removeNodeFromTree(nodes: any[], id: string | number): boolean {
	for (let i = 0; i < nodes.length; i++) {
		if (String(nodes[i].id || nodes[i].key) === String(id)) {
			nodes.splice(i, 1);
			return true;
		}
		if (nodes[i].children && nodes[i].children.length > 0) {
			if (removeNodeFromTree(nodes[i].children, id)) {
				return true;
			}
		}
	}
	return false;
}

// 更新缓存：添加书签
function updateCacheAfterAdd(bookmark: any) {
	try {
		const cachedData = ss.get(BOOKMARKS_CACHE_KEY);
		if (!cachedData) {
			// 如果没有缓存，刷新数据
			refreshBookmarks(false);
			return;
		}

		// 将新书签转换为缓存格式
		const newBookmark = convertBookmarkToCacheNode(bookmark);

		// 处理缓存数据
		let cacheList: any[] = [];
		if (Array.isArray(cachedData)) {
			cacheList = cachedData;
		} else if (cachedData.list && Array.isArray(cachedData.list)) {
			cacheList = cachedData.list;
		} else {
			// 如果格式不对，刷新数据
			refreshBookmarks(false);
			return;
		}

		// 如果是根节点（parentId为'0'），直接添加到列表
		if (newBookmark.parentId === '0') {
			cacheList.push(newBookmark);
			// 按sort排序根列表
			cacheList.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
		} else {
			// 查找父节点并添加到其children中
			const parentNode = findNodeInTree(cacheList, newBookmark.parentId);
			if (parentNode) {
				if (!parentNode.children) {
					parentNode.children = [];
				}
				parentNode.children.push(newBookmark);
				// 按sort排序
				parentNode.children.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
			} else {
				// 如果找不到父节点，添加到根列表
				cacheList.push(newBookmark);
				// 按sort排序根列表
				cacheList.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
			}
		}

		// 更新缓存
		ss.set(BOOKMARKS_CACHE_KEY, cacheList);

		// 重新构建树并更新UI
		const treeDataResult = convertServerTreeToFrontendTree(cacheList);
		fullData.value = treeDataResult;
		const folderTreeData = filterFoldersOnly(treeDataResult);
		if (Array.isArray(folderTreeData) && folderTreeData.length > 0) {
			bookmarkTree.value = JSON.parse(JSON.stringify(folderTreeData));
		}
	} catch (error) {
		console.error('更新缓存失败，刷新数据:', error);
		refreshBookmarks(false);
	}
}

// 更新缓存：批量保存书签排序
function updateCacheAfterSort(sortItems: { id: number; sort: number }[]) {
	try {
		const sortMap = new Map(sortItems.map(item => [Number(item.id), item.sort]));

		// 1. 同步更新 fullData 和 bookmarkTree
		const updateNodeSort = (nodes: any[]) => {
			if (!Array.isArray(nodes)) return;
			for (const node of nodes) {
				const nodeId = Number(node.id || node.key || node.bookmark?.id);
				if (sortMap.has(nodeId)) {
					const newSort = sortMap.get(nodeId)!;
					if (node.bookmark) {
						node.bookmark.sort = newSort;
					}
					node.sort = newSort;
					if (node.rawNode) {
						node.rawNode.sort = newSort;
					}
				}
				if (node.children && Array.isArray(node.children)) {
					updateNodeSort(node.children);
					node.children.sort((a: any, b: any) => {
						const aSort = a.sort ?? a.bookmark?.sort ?? a.rawNode?.sort ?? 0;
						const bSort = b.sort ?? b.bookmark?.sort ?? b.rawNode?.sort ?? 0;
						return aSort - bSort;
					});
				}
			}
		};

		if (Array.isArray(fullData.value)) {
			updateNodeSort(fullData.value);
			fullData.value.sort((a, b) => {
				const aSort = a.sort ?? a.bookmark?.sort ?? a.rawNode?.sort ?? 0;
				const bSort = b.sort ?? b.bookmark?.sort ?? b.rawNode?.sort ?? 0;
				return aSort - bSort;
			});
			fullData.value = [...fullData.value];
		}

		if (Array.isArray(bookmarkTree.value)) {
			updateNodeSort(bookmarkTree.value);
			bookmarkTree.value.sort((a, b) => {
				const aSort = a.sort ?? a.bookmark?.sort ?? a.rawNode?.sort ?? 0;
				const bSort = b.sort ?? b.bookmark?.sort ?? b.rawNode?.sort ?? 0;
				return aSort - bSort;
			});
			bookmarkTree.value = [...bookmarkTree.value];
		}

		// 2. 更新持久化本地缓存
		const cachedData = ss.get(BOOKMARKS_CACHE_KEY);
		if (cachedData) {
			let cacheList: any[] = [];
			if (Array.isArray(cachedData)) {
				cacheList = cachedData;
			} else if (cachedData.list && Array.isArray(cachedData.list)) {
				cacheList = cachedData.list;
			}

			if (cacheList.length > 0) {
				const updateCacheListSort = (nodes: any[]) => {
					for (const node of nodes) {
						const nodeId = Number(node.id || node.key);
						if (sortMap.has(nodeId)) {
							node.sort = sortMap.get(nodeId);
						}
						if (node.children && Array.isArray(node.children)) {
							updateCacheListSort(node.children);
							node.children.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
						}
					}
				};
				updateCacheListSort(cacheList);
				cacheList.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
				ss.set(BOOKMARKS_CACHE_KEY, cachedData);
			}
		}
	} catch (error) {
		console.error('更新本地排序状态失败:', error);
	}
}

// 更新缓存：修改书签
function updateCacheAfterUpdate(bookmark: any) {
    try {
        const cachedData = ss.get(BOOKMARKS_CACHE_KEY);
        if (!cachedData) {
            refreshBookmarks(false);
            return;
        }

        let cacheList: any[] = [];
        if (Array.isArray(cachedData)) {
            cacheList = cachedData;
        } else if (cachedData.list && Array.isArray(cachedData.list)) {
            cacheList = cachedData.list;
        } else {
            refreshBookmarks(false);
            return;
        }

        // 查找并更新节点
        const node = findNodeInTree(cacheList, bookmark.id);
        if (node) {
            // 保存原来的parentId用于比较
            const oldParentId = node.parentId || '0';
            const newParentId = bookmark.parentId || bookmark.ParentId || '0';

            // 更新节点属性
            node.title = bookmark.title;
            node.url = bookmark.url || '';
            node.iconJson = bookmark.iconJson || '';
            node.sort = bookmark.sort || 0;
            node.parentId = newParentId;

            // 如果parentId改变了，需要移动节点
            if (String(oldParentId) !== String(newParentId)) {
                // 从原位置删除
                removeNodeFromTree(cacheList, bookmark.id);
                // 添加到新位置
                if (newParentId === '0' || newParentId === '' || newParentId === 'null' || newParentId === null || newParentId === undefined) {
                    cacheList.push(node);
                    // 按sort排序
                    cacheList.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
                } else {
                    const parentNode = findNodeInTree(cacheList, newParentId);
                    if (parentNode) {
                        if (!parentNode.children) {
                            parentNode.children = [];
                        }
                        parentNode.children.push(node);
                        parentNode.children.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
                    } else {
                        // 如果找不到父节点，添加到根列表
                        cacheList.push(node);
                        cacheList.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0));
                    }
                }
            }
        } else {
            // 如果找不到节点，刷新数据
            refreshBookmarks(false);
            return;
        }

        // 更新缓存
        ss.set(BOOKMARKS_CACHE_KEY, cacheList);

        // 重新构建树并更新UI
        const treeDataResult = convertServerTreeToFrontendTree(cacheList);
        fullData.value = treeDataResult;
        const folderTreeData = filterFoldersOnly(treeDataResult);
        if (Array.isArray(folderTreeData) && folderTreeData.length > 0) {
            bookmarkTree.value = JSON.parse(JSON.stringify(folderTreeData));
        }
    } catch (error) {
        console.error('更新缓存失败，刷新数据:', error);
        refreshBookmarks(false);
    }
}

// 更新缓存：删除书签（包括递归删除子项）
function updateCacheAfterDelete(bookmarkId: number, isFolder: boolean) {
	try {
		const cachedData = ss.get(BOOKMARKS_CACHE_KEY);
		if (!cachedData) {
			refreshBookmarks(false);
			return;
		}

		// 处理缓存数据格式（可能是树形结构或列表）
		let cacheList: any[] = [];
		if (Array.isArray(cachedData)) {
			// 检查是否是树形结构（有children字段）
			if (cachedData.length > 0 && cachedData[0].children !== undefined) {
				cacheList = cachedData; // 树形结构
			} else {
				cacheList = cachedData; // 扁平列表
			}
		} else if (cachedData.list && Array.isArray(cachedData.list)) {
			cacheList = cachedData.list;
		} else {
			refreshBookmarks(false);
			return;
		}

		// 收集所有需要删除的ID（包括文件夹下的所有子项）
		const collectAllIdsToDelete = (nodes: any[], targetId: number): number[] => {
			const idsToDelete: number[] = [];

			// 查找目标节点（支持id和key匹配）
			const findNode = (nodeList: any[]): any => {
				for (const node of nodeList) {
					const nodeId = Number(node.id || node.key || 0);
					if (nodeId === targetId) {
						return node;
					}
					if (node.children && node.children.length > 0) {
						const found = findNode(node.children);
						if (found) return found;
					}
				}
				return null;
			};

			const targetNode = findNode(nodes);
			if (!targetNode) {

				return idsToDelete;
			}

			// 添加目标节点ID
			const targetNodeId = Number(targetNode.id || targetNode.key || bookmarkId);
			idsToDelete.push(targetNodeId);

			// 如果是文件夹，递归收集所有子项ID
			if ((targetNode.isFolder === 1 || targetNode.isFolder === true) && targetNode.children && targetNode.children.length > 0) {
				const collectChildrenIds = (children: any[]) => {
					for (const child of children) {
						const childId = Number(child.id || child.key || 0);
						if (childId > 0) {
							idsToDelete.push(childId);
						}
						// 如果子项也是文件夹，递归收集其子项
						if ((child.isFolder === 1 || child.isFolder === true) && child.children && child.children.length > 0) {
							collectChildrenIds(child.children);
						}
					}
				};
				collectChildrenIds(targetNode.children);
			}

			return idsToDelete;
		};

		// 收集所有需要删除的ID
		const idsToDelete = collectAllIdsToDelete(cacheList, bookmarkId);


		if (idsToDelete.length === 0) {

			refreshBookmarks(false);
			return;
		}

		// 递归删除函数（删除指定ID的节点，支持id和key匹配）
		const deleteNodeById = (nodes: any[], id: number): boolean => {
			for (let i = 0; i < nodes.length; i++) {
				const nodeId = Number(nodes[i].id || nodes[i].key || 0);
				if (nodeId === id) {
					nodes.splice(i, 1);
					return true;
				}
				if (nodes[i].children && nodes[i].children.length > 0) {
					if (deleteNodeById(nodes[i].children, id)) {
						return true;
					}
				}
			}
			return false;
		};

		// 删除所有收集到的节点（按ID从大到小排序，先删除子项再删除父项）
		const sortedIds = [...idsToDelete].sort((a, b) => b - a);
		for (const id of sortedIds) {
			deleteNodeById(cacheList, id);
		}



		// 更新缓存
		ss.set(BOOKMARKS_CACHE_KEY, cacheList);


		// 重新构建树并更新UI
		const treeDataResult = convertServerTreeToFrontendTree(cacheList);
		fullData.value = treeDataResult;
		const folderTreeData = filterFoldersOnly(treeDataResult);
		if (Array.isArray(folderTreeData) && folderTreeData.length > 0) {
			bookmarkTree.value = JSON.parse(JSON.stringify(folderTreeData));
		} else {
			bookmarkTree.value = [];
		}
	} catch (error) {
		console.error('更新缓存失败，刷新数据:', error);
		refreshBookmarks(false);
	}
}

const handleResize = () => {
	isMobile.value = window.innerWidth < 768
	// 移动端默认隐藏左栏
	showLeftPanel.value = isMobile.value ? false : true
}
// 组件挂载时加载书签
onMounted(async () => {
	// 首先设置selectedFolder为'0'
	selectedFolder.value = '0'

	// 加载数据
	await refreshBookmarks();

	// 完全绕过计算属性和watch，直接操作sortedItems
	// 使用较长的延迟确保数据完全处理
	setTimeout(() => {


		// 移除未使用的调试代码

		// 严格过滤根目录项目
		const rootItems = allItems.value.filter(item => {
			const folderId = String(item.folderId || '0');
			return folderId === '0';
		});


		// 清空并重新设置sortedItems
		sortedItems.value = [];
		nextTick(() => {
			sortedItems.value = [...rootItems];
		});
	}, 500);

	// 添加全局事件监听器
	document.addEventListener('mousemove', handleMouseMove);
	document.addEventListener('mouseup', stopResize);
	document.addEventListener('click', handleGlobalClick);

	handleResize()
	window.addEventListener('resize', handleResize)
});

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
	document.removeEventListener('mouseup', stopResize);
	document.removeEventListener('click', handleGlobalClick);

	window.removeEventListener('resize', handleResize)
	// 清理全局存储的完整数据，避免内存泄漏
	if ((globalThis as any).__bookmarksFullData) {
		delete (globalThis as any).__bookmarksFullData;
	}
});



</script>

<style scoped>
.context-menu {
    position: fixed !important;
    z-index: 99999 !important;
    border-radius: 0.375rem !important;
    border-right: 1px solid #e5e7eb !important;
}

.dark .context-menu {
  border-right: 1px solid #4a5568 !important;
}


/* 阴影和圆角让左栏浮动更美观 */
.fixed.bg-white {
	border-right: 1px solid #e5e7eb; /* 柔和边框 */
	border-radius: 0 0.75rem 0.75rem 0; /* 左边圆角 */
	box-shadow: 0 4px 12px rgba(0,0,0,0.1);
	transition: width 0.3s ease-in-out, transform 0.3s ease-in-out;
}

    /* 移除输入组件根元素的边框 */
    .bookmark-search-input.n-input {
      border: none !important;
      box-shadow: none !important;
      outline: none !important;
    }
    /* 移除内部输入组件的所有边框、阴影和圆角 */
    .bookmark-search-input.n-input :deep(*) {
      border: none !important;
      outline: none !important;
      box-shadow: none !important;
      border-radius: 0 !important;
    }
    /* 确保正确的布局 */
    .bookmark-search-input.n-input :deep(.n-input__inner),
    .bookmark-search-input.n-input :deep(.n-input__input-wrap) {
      overflow: hidden !important;
      width: 100% !important;
      height: 100% !important;
    }

</style>
