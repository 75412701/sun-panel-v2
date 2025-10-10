<template>
	<div class="flex flex-col h-screen bg-white">
		<!-- 顶部标题栏 -->
		<div class="p-4 border-b flex items-center justify-between bg-gray-50">
			<NButton
				@click="goBackToHome"
				type="primary"
				ghost
				size="small"
				class="flex items-center gap-1 text-blue-600 hover:bg-blue-100 transition-colors border border-blue-200"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
				返回
			</NButton>
			<h1 class="text-xl font-bold text-gray-800 flex-1 text-center">书签管理</h1>
			<NButton
				@click="triggerImportBookmarks"
				type="success"
				ghost
				size="small"
				:loading="uploadLoading"
				class="flex items-center gap-1 text-green-600 hover:bg-green-100 transition-colors border border-green-200"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
				</svg>
				导入书签
			</NButton>
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
			<!-- 左侧书签树 -->
			<div :style="{ width: leftPanelWidth + 'px' }" class="border-r p-2 overflow-auto">
				<n-tree
					:data="bookmarkTree"
					:default-expand-all="true"
					block-line
					@update:selected-keys="handleSelect"
				/>
			</div>

			<!-- 可拖动分割线 -->
			<div
				class="w-1 bg-gray-200 cursor-col-resize hover:bg-blue-300 flex items-center justify-center"
				@mousedown="startResize"
				:style="{ height: '100%', userSelect: 'none' }"
			>
				<div class="w-px h-12 bg-gray-400"></div>
			</div>

			<!-- 右侧书签列表 -->
			<div class="flex-1 flex flex-col overflow-hidden">
				<div class="p-2 border-b flex items-center">
					<n-input
						v-model:value="searchQuery"
						placeholder="搜索书签..."
						clearable
						@input="handleSearch"
					/>
				</div>

				<!-- 书签列表 -->
				<div class="flex-1 p-4 relative overflow-auto">
					<div class="grid grid-cols-1 gap-2">
						<!-- 如果没有数据，显示空状态提示 -->
						<div v-if="filteredBookmarks.length === 0" class="text-center py-8 text-gray-400">
							暂无书签数据
						</div>

						<!-- 渲染书签列表 -->
						<div
							v-for="bookmark in filteredBookmarks"
							:key="bookmark.id"
							class="flex items-center justify-between border p-2 rounded hover:bg-gray-50 cursor-pointer"
							@contextmenu.prevent="openContextMenu($event, bookmark)"
						>
							<div class="flex items-center space-x-2">
								<span class="font-medium text-slate-700">{{ bookmark.title }}</span>
								<span class="text-slate-400 text-sm">{{ bookmark.url }}</span>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- 右键菜单 -->
		<NDropdown
			trigger="manual"
			placement="bottom-start"
			:x="contextMenu.x"
			:y="contextMenu.y"
			:options="contextOptions"
			:show="contextMenu.show"
			@select="handleContextSelect"
			@clickoutside="contextMenu.show = false"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { NTree, NDataTable, NInput, NDropdown, NButton, NMenu, NAlert, useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import type { ImportBookmarkResult } from '@/utils/bookmarkImportExport'
import { flattenBookmarkTree } from '@/utils/bookmarkImportExport'
import { addMultiple as addMultipleBookmarks, getList as getBookmarksList } from '@/api/panel/bookmark'
import { t } from '@/locales'

const router = useRouter()
const ms = useMessage()

// 左侧面板宽度
const leftPanelWidth = ref(256) // 默认256px
const isResizing = ref(false)

// 导入相关
const fileInput = ref<HTMLInputElement>()
const uploadLoading = ref(false)
const jsonData = ref<string | null>(null)
const importWarning = ref<string[]>([])
const importObj = ref<ImportBookmarkResult | null>(null)

// 返回首页
function goBackToHome() {
  router.push('/')
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
}


const bookmarkTree = ref([
	{
		key: 1,
		label: '工作',
		children: [
			{ key: 11, label: 'Vue官网', isLeaf: true, bookmark: { id: 11, title: 'Vue官网', url: 'https://vuejs.org', folderId: 1 } },
			{ key: 12, label: 'Naive UI', isLeaf: true, bookmark: { id: 12, title: 'Naive UI', url: 'https://www.naiveui.com', folderId: 1 } }
		]
	},
	{
		key: 2,
		label: '娱乐',
		children: [
			{ key: 21, label: 'YouTube', isLeaf: true, bookmark: { id: 21, title: 'YouTube', url: 'https://youtube.com', folderId: 2 } }
		]
	}
])

// 当前选中的文件夹
const selectedFolder = ref<string | null>(null)



// 计算属性 - 所有书签
const allBookmarks = computed<Bookmark[]>(() => {
	const bookmarks: Bookmark[] = []

	// 遍历所有文件夹和书签
	function traverseBookmarks(nodes: any[], folderId: string) {
		for (const node of nodes) {
			// 如果是文件夹，继续遍历
			if (node.children && node.children.length > 0) {
				traverseBookmarks(node.children, String(node.key))
			}
			// 如果是书签，添加到结果集
			else if (node.isLeaf && node.bookmark) {
				// 确保每个书签都有folderId属性，并转换为字符串格式便于比较
				bookmarks.push({
					...node.bookmark,
					folderId: String(folderId)
				})
			}
		}
	}

	// 开始遍历根节点
	traverseBookmarks(bookmarkTree.value, '0')
	return bookmarks
})

// 搜索
const searchQuery = ref('')
const filteredBookmarks = computed(() => {
	// 先获取所有书签
	let bookmarks = allBookmarks.value

	// 1. 应用文件夹过滤
	if (selectedFolder.value) {
		const targetFolderId = String(selectedFolder.value)
		bookmarks = bookmarks.filter(bookmark => {
			// 直接比较字符串形式的folderId
			return String(bookmark.folderId) === targetFolderId
		})
	}

	// 2. 应用搜索过滤
	if (searchQuery.value.trim()) {
		const query = searchQuery.value.toLowerCase()
		bookmarks = bookmarks.filter(bookmark =>
			bookmark.title.toLowerCase().includes(query) ||
			bookmark.url.toLowerCase().includes(query)
		)
	}

	return bookmarks
})

// 点击树节点
function handleSelect(keys: (string | number)[]) {
	if (keys && keys.length > 0) {
		const key = keys[0];
		// 无论key是数字还是字符串，都转换为字符串类型
		selectedFolder.value = String(key);
	} else {
		selectedFolder.value = null;
	}
}

// 搜索时清空选中
function handleSearch() {
	if (searchQuery.value) {
		selectedFolder.value = null
	}
}

// 根据文件夹ID获取文件夹名称
function getFolderName(folderId: string): string {
	// 在bookmarkTree中查找对应的文件夹
	function findFolder(nodes: any[]): string | null {
		for (const node of nodes) {
			if (String(node.key) === folderId) {
				return node.label
			}
			if (node.children && node.children.length > 0) {
				const found = findFolder(node.children)
				if (found) {
					return found
				}
			}
		}
		return null
	}

	const folderName = findFolder(bookmarkTree.value)
	return folderName || '未知文件夹'
}

// 测试文件夹匹配的状态
const testFolderId = ref('')
const testResults = ref<any[]>([])

// 测试文件夹ID匹配
function testFolderMatch() {
	if (!testFolderId.value) {
		testResults.value = []
		return
	}


	testResults.value = allBookmarks.value.map(bookmark => {
		// 尝试多种匹配方式
		const folderIdStr = String(bookmark.folderId || '')
		const folderIdNum = Number(bookmark.folderId || NaN)
		const testIdStr = String(testFolderId.value)
		const testIdNum = Number(testFolderId.value)

		const isStrMatch = folderIdStr === testIdStr
		const isNumMatch = !isNaN(folderIdNum) && !isNaN(testIdNum) && folderIdNum === testIdNum
		const isMatch = isStrMatch || isNumMatch


		return {
			...bookmark,
			isMatch,
			matchType: isStrMatch ? '字符串匹配' : (isNumMatch ? '数字匹配' : '不匹配')
		}
	})

	// 统计匹配数量
	const matchCount = testResults.value.filter(r => r.isMatch).length
}

// 右键菜单
const menuX = ref(0)
const menuY = ref(0)
const showMenu = ref(false)
const selectedRow = ref<Bookmark | null>(null)


// 右键菜单状态
const contextMenu = ref({
	show: false,
	x: 0,
	y: 0,
	bookmark: null as Bookmark | null,
});

const contextOptions = [
	{ label: "新窗口打开", key: "open" },
	{ label: "修改", key: "edit" },
	{ label: "删除", key: "delete" },
];
const openContextMenu = (e: MouseEvent, bookmark: Bookmark) => {
	contextMenu.value = {
		show: true,
		x: e.clientX,
		y: e.clientY,
		bookmark,
	};
};

const handleContextSelect = (key: string) => {
	const bm = contextMenu.value.bookmark;
	if (!bm) return;

	switch (key) {
		case "open":
			window.open(bm.url, "_blank");
			break;
		case "edit":
			alert(`修改书签：${bm.title}`);
			break;
		case "delete":
			break;
	}
	contextMenu.value.show = false;
};

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
			ms.error('只支持HTML格式的书签文件导入');
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
		// 直接将HTML内容传递给后端
		const response = await addMultipleBookmarks({ htmlContent });
		if (response.code === 0) {
			ms.success(`${t('common.success')}，成功导入 ${response.data.count} 个书签`);
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

// 旧的导入方法（保留用于向后兼容）
async function importBookmarksToServer(bookmarks: any[]) {
	uploadLoading.value = true;
	try {
		// 扁平化书签树
		const flatBookmarks = flattenBookmarkTree(bookmarks);

		// 转换为服务器需要的格式
		const bookmarksToImport = flatBookmarks.map(bookmark => ({
			title: bookmark.title,
			url: bookmark.url,
			parentId: bookmark.folderId || 0,
			isFolder: bookmark.isFolder ? 1 : 0,
			sort: 9999, // 默认排序
			lanUrl: '', // 局域网地址留空
			IconJson: '' // 图标留空
		}));

		// 批量导入到服务器
		const response = await addMultipleBookmarks(bookmarksToImport);
		if (response.code === 0) {
			ms.success(`${t('common.success')}，成功导入 ${bookmarksToImport.length} 个书签`);
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

// 刷新书签列表
async function refreshBookmarks() {
	try {
		const response = await getBookmarksList();
		if (response.code === 0) {
			// 检查数据结构，如果已经是树形结构则直接使用
			const data = response.data || [];

			// 检查是否已经是树形结构（直接包含children字段）
			let treeData = [];
			if (Array.isArray(data) && data.length > 0 && 'children' in data[0]) {
				// 已经是树形结构，转换为前端需要的格式
				treeData = convertServerTreeToFrontendTree(data);
			} else if (data.list && Array.isArray(data.list)) {
				// 后端返回的是带list字段的结构
				const serverBookmarks = data.list;
				if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
					// list字段中已经是树形结构
					treeData = convertServerTreeToFrontendTree(serverBookmarks);
				} else {
					// 构建树形结构
					treeData = buildBookmarkTree(serverBookmarks);
				}
			} else {
				// 作为列表数据构建树形结构
				treeData = buildBookmarkTree(Array.isArray(data) ? data : []);
			}

			bookmarkTree.value = treeData;
		}
	} catch (error) {
		console.error('刷新书签列表失败:', error);
	}
}

// 将服务器返回的树形结构转换为前端组件需要的格式
function convertServerTreeToFrontendTree(serverTree: any[]): any[] {
	const result = serverTree.map(node => {
		// 处理bookmark对象
		let bookmarkObj = undefined;
		if (node.isFolder !== 1 && node.url) {
			// 确保folderId是字符串类型，以便与selectedFolder.value进行比较
			const folderId = node.ParentUrl !== undefined ? String(node.ParentUrl) : null;
			bookmarkObj = {
				id: node.id,
				title: node.title,
				url: node.url,
				folderId: folderId
			};
		}

		const frontendNode = {
			key: node.id,
			label: node.title,
			isLeaf: node.isFolder !== 1,
			bookmark: bookmarkObj,
			// 添加原始节点信息用于调试
			rawNode: node
		};

		// 递归处理子节点
		if (node.children && node.children.length > 0) {
			frontendNode.children = convertServerTreeToFrontendTree(node.children);
		}

		return frontendNode;
	});
	return result;
}

// 构建书签树
function buildBookmarkTree(bookmarks: any[]): any[] {
	// 首先分离文件夹和书签
	const folders = bookmarks.filter(b => b.isFolder === 1);
	const items = bookmarks.filter(b => b.isFolder === 0);

	// 构建文件夹树
	const rootFolders: any[] = [];
	const folderMap = new Map<string, any>(); // 使用字符串键，因为ParentUrl可能是字符串



	// 先创建所有文件夹节点
	folders.forEach(folder => {
		const folderNode = {
			key: folder.id,
			label: folder.title,
			children: [],
			isFolder: true
		};
		// 使用id作为map的键，因为ParentUrl可能是文件夹的名称或其他标识
		folderMap.set(folder.id.toString(), folderNode);
		// 同时也将文件夹名称作为键，以便处理嵌套关系
		folderMap.set(folder.title, folderNode);
	});

	// 将文件夹添加到其父文件夹中
	folders.forEach(folder => {
		const folderNode = folderMap.get(folder.id.toString());
		// 检查是否有ParentUrl并且不是根节点(0)
		if (folder.ParentUrl && folder.ParentUrl !== '0' && folder.ParentUrl !== 0) {
			// 尝试用不同的方式查找父文件夹
			let parentFolder = folderMap.get(folder.ParentUrl.toString());

			if (!parentFolder) {
				// 如果找不到，尝试用文件夹标题匹配
				parentFolder = folderMap.get(folder.ParentUrl);
			}

			if (parentFolder) {
				parentFolder.children.push(folderNode);
				return;
			}
		}
		// 如果没有父文件夹或者父文件夹不存在，则添加到根节点
		rootFolders.push(folderNode);
	});

	// 将书签添加到对应的文件夹中
	items.forEach(item => {
		const bookmarkItem = {
			key: item.id,
			label: item.title,
			isLeaf: true,
			bookmark: {
				id: item.id,
				title: item.title,
				url: item.url,
				// 确保folderId始终是字符串类型
				folderId: item.ParentUrl !== undefined ? String(item.ParentUrl) : null
			}
		};

		// 检查是否有ParentUrl并且不是根节点(0)
		if (item.ParentUrl && item.ParentUrl !== '0' && item.ParentUrl !== 0) {
			// 尝试用不同的方式查找父文件夹
			let parentFolder = folderMap.get(item.ParentUrl.toString());

			if (!parentFolder) {
				// 如果找不到，尝试用文件夹标题匹配
				parentFolder = folderMap.get(item.ParentUrl);
			}

			if (parentFolder) {
				parentFolder.children.push(bookmarkItem);
				return;
			}
		}
		// 如果没有指定文件夹或文件夹不存在，则添加到根节点
		rootFolders.push(bookmarkItem);
	});
	// 如果没有书签数据，使用默认数据
	if (rootFolders.length === 0) {
		return [
			{
				key: 1,
				label: '工作',
				children: [
					{ key: 11, label: 'Vue官网', isLeaf: true, bookmark: { id: 11, title: 'Vue官网', url: 'https://vuejs.org', folderId: 1 } },
					{ key: 12, label: 'Naive UI', isLeaf: true, bookmark: { id: 12, title: 'Naive UI', url: 'https://www.naiveui.com', folderId: 1 } }
				]
			},
			{
				key: 2,
				label: '娱乐',
				children: [
					{ key: 21, label: 'YouTube', isLeaf: true, bookmark: { id: 21, title: 'YouTube', url: 'https://youtube.com', folderId: 2 } }
				]
			}
		];
	}

	return rootFolders;
}

// 组件挂载时加载书签
onMounted(() => {
	refreshBookmarks();

	// 添加全局事件监听器
	document.addEventListener('mousemove', handleMouseMove);
	document.addEventListener('mouseup', stopResize);
});

// 组件卸载时移除事件监听器
onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
	document.removeEventListener('mouseup', stopResize);
});

</script>
