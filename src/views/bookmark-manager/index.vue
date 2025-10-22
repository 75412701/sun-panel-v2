<template>
	<div
		class="flex flex-col h-screen bg-white"
		@contextmenu.prevent
	>
		<!-- 顶部标题栏 -->
		<div class="p-4 border-b flex items-center justify-between bg-gray-50 relative">
			<!-- 移动端左栏展开按钮 -->
			<NButton
				v-if="isMobile"
				@click="togglePanel"
				type="primary"
				ghost
				size="small"
				class="flex items-center gap-1 text-blue-600 hover:bg-blue-100 transition-colors border border-blue-200"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</NButton>

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
				@click="createNewBookmark"
				type="primary"
				ghost
				size="small"
				class="flex items-center gap-1 text-purple-600 hover:bg-purple-100 transition-colors border border-purple-200"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				新建
			</NButton>

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
    isMobile ? 'fixed top-0 left-0 h-full bg-white z-50 rounded-r-lg shadow-lg overflow-auto transition-all duration-300 ease-in-out' : 'h-full bg-white border-r border-gray-200 overflow-auto',
    isMobile && isPanelExpanded ? 'w-3/4' : isMobile ? 'w-12' : ''
  ]"
				:style="{ width: !isMobile ? leftPanelWidth + 'px' : '' }"
			>
				<n-tree
					:data="bookmarkTree"
					:default-expand-all="true"
					block-line
					@update:selected-keys="handleSelect"
					:render-label="renderTreeLabel"
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
						<div v-if="filteredBookmarks.length === 0" class="text-center py-8 text-gray-400">
							暂无书签数据
						</div>

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

		<!-- 编辑书签对话框 -->
		<div v-if="isEditDialogOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
			<div class="bg-white p-6 rounded-lg w-96">
				<h3 class="text-xl font-bold mb-4">{{ isCreateMode ? '新建' : '修改' }}书签</h3>
				<div v-if="isCreateMode" class="mb-4">
					<label class="block mb-2">类型</label>
					<select
						v-model="bookmarkType"
						class="w-full px-3 py-2 border border-gray-300 rounded-md"
					>
						<option value="bookmark">书签</option>
						<option value="folder">文件夹</option>
					</select>
				</div>
				<div class="mb-4">
					<label class="block mb-2">标题</label>
					<input
						v-model="currentEditBookmark.title"
						class="w-full px-3 py-2 border border-gray-300 rounded-md"
						placeholder="请输入{{ bookmarkType === 'folder' ? '文件夹' : '书签' }}标题"
					/>
				</div>
				<div v-if="!isCreateMode || bookmarkType === 'bookmark'" class="mb-4">
					<label class="block mb-2">URL</label>
					<input
						v-model="currentEditBookmark.url"
						class="w-full px-3 py-2 border border-gray-300 rounded-md"
						placeholder="请输入书签URL"
					/>
				</div>
				<div class="mb-4">
					<label class="block mb-2">上级文件夹</label>
					<select v-model="currentEditBookmark.folderId" class="w-full px-3 py-2 border border-gray-300 rounded-md">
						<option v-for="folder in allFolders" :key="folder.value" :value="folder.value">{{ folder.label }}</option>
					</select>
				</div>
				<div class="flex justify-end gap-2">
					<button @click="closeEditDialog" class="px-4 py-2 border border-gray-300 rounded-md">取消</button>
					<button @click="saveBookmarkChanges" class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors shadow-md">确定</button>
				</div>
			</div>
		</div>

		<!-- 右键菜单 -->
		<div
			v-if="isContextMenuOpen"
			:style="contextMenuStyle"
			class="fixed bg-white text-gray-700 shadow-lg rounded-md py-1 z-50 w-40 context-menu border border-gray-200"
			@contextmenu.prevent.stop
		>
			<div @click="handleEditBookmark" class="px-4 py-2 hover:bg-gray-100 cursor-pointer">编辑</div>
			<div @click="handleDeleteBookmark" class="px-4 py-2 hover:bg-gray-100 cursor-pointer">删除</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted,h} from 'vue'
import { NTree, NInput, NButton, useMessage } from 'naive-ui'
import { useRouter } from 'vue-router'
import { addMultiple as addMultipleBookmarks, add as addBookmark, getList as getBookmarksList, deletes, update as updateBookmark } from '@/api/panel/bookmark'
import { t } from '@/locales'
import { dialog } from '@/utils/request/apiMessage'
import { ss } from '@/utils/storage/local'

// 首页书签缓存键名
const BOOKMARKS_CACHE_KEY = 'bookmarksTreeCache'

const router = useRouter()
const ms = useMessage()
const isMobile = ref(false)      // 是否移动端
const showLeftPanel = ref(true)  // 左栏是否显示
// 左侧面板宽度
const leftPanelWidth = ref(256) // 默认256px
const isResizing = ref(false)

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


const bookmarkTree = ref<any[]>([])

// 当前选中的文件夹
const selectedFolder = ref<string>('')



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

// 当前选中的节点键引用
const selectedKeysRef = ref<(string | number)[]>([]);

// 点击树节点
function handleSelect(keys: (string | number)[]) {
	// 更新选中的节点键引用
	selectedKeysRef.value = keys;

	// 确保类型安全的赋值方式
	if (keys && Array.isArray(keys) && keys.length > 0) {
		const key = keys[0];
		selectedFolder.value = String(key);
	} else {
		selectedFolder.value = '';
	}
}

// 搜索时清空选中
function handleSearch() {
	if (searchQuery.value) {
		selectedFolder.value = ''
	}
}



// 右键菜单相关状态
const isContextMenuOpen = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const currentBookmark = ref<Bookmark | null>(null);

// 树组件引用
const treeRef = ref<InstanceType<typeof NTree> | null>(null);

// 计算属性 - 所有文件夹（用于下拉选择）
const allFolders = computed(() => {
    const folders: { label: string; value: string }[] = [{ label: '根目录', value: '0' }];
    const collectFolders = (nodes: any[]) => {
        for (const node of nodes) {
            if (!node.isLeaf) {
                folders.push({ label: node.label, value: String(node.key) });
                if (node.children) {
                    collectFolders(node.children);
                }
            }
        }
    };
    collectFolders(bookmarkTree.value);
    return folders;
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
} as { id: number; title: string; url: string; folderId?: string | number | undefined; });


// 右键菜单样式
const contextMenuStyle = computed(() => ({
	top: `${contextMenuY.value}px`,
	left: `${contextMenuX.value}px`,
}));

// 打开右侧列表右键菜单
function openContextMenu(event: MouseEvent, bookmark: Bookmark) {
	event.preventDefault()
	event.stopPropagation()
	isContextMenuOpen.value = true
	contextMenuX.value = event.clientX
	contextMenuY.value = event.clientY
	currentBookmark.value = bookmark
}

// ✅ 左侧树节点右键菜单（与右侧一样）
function handleTreeContextMenu({ node, event }: { node: any; event: MouseEvent }) {
	event.preventDefault()
	event.stopPropagation()

	// 打开菜单
	isContextMenuOpen.value = true
	contextMenuX.value = event.clientX
	contextMenuY.value = event.clientY

	// 当前节点信息
	currentBookmark.value = {
		id: node.key,
		title: node.label,
		url: node.bookmark?.url || '',
		folderId: node.bookmark?.folderId || '0',
		isFolder: !node.isLeaf,
	}
}

// 渲染每个节点时，给 label 添加右键事件
function renderTreeLabel({ option }: { option: any }) {
	return h(
		'span',
		{
			class: 'block px-1 py-0.5 rounded hover:bg-gray-100 cursor-default',
			onContextmenu: (e: MouseEvent) => {
				handleTreeContextMenu({ node: option, event: e })
			},
		},
		option.label
	)
}

function handleGlobalClick(event: MouseEvent) {
	const path = event.composedPath() as HTMLElement[]
	const clickedInsideMenu = path.some(
		(el) => el.classList && el.classList.contains('context-menu')
	)
	if (!clickedInsideMenu) {
		closeContextMenu()
	}
}


// 关闭右键菜单
function closeContextMenu() {
	isContextMenuOpen.value = false;
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
	currentEditBookmark.value = {
		id: 0,
		title: '',
		url: '',
		folderId: '0',
	};
	// 设置为创建模式
	isCreateMode.value = true;
	// 默认选择书签类型
	bookmarkType.value = 'bookmark';
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
			ms.error('标题不能为空');
			return;
		}

		// 如果是创建书签类型，验证URL
		if ((!isCreateMode.value || bookmarkType.value === 'bookmark') && !currentEditBookmark.value.url.trim()) {
			ms.error('书签URL不能为空');
			return;
		}

		// 添加URL协议检查和补充
		if ((!isCreateMode.value || bookmarkType.value === 'bookmark') && currentEditBookmark.value.url.trim()) {
			let url = currentEditBookmark.value.url.trim();
			if (!url.startsWith('http://') && !url.startsWith('https://')) {
				url = 'https://' + url;
				currentEditBookmark.value.url = url;
			}
		}

		// 清除首页的书签缓存，确保下次访问首页时重新获取最新数据
		ss.remove(BOOKMARKS_CACHE_KEY)

		// 根据模式决定调用哪个接口
		if (isCreateMode.value) {
			// 创建新模式
			const createData = {
				title: currentEditBookmark.value.title,
				url: '',
				// 注意后端JSON标签是小写的parentUrl
				parentUrl: currentEditBookmark.value.folderId ? currentEditBookmark.value.folderId.toString() : '0',
				sort: 9999,
				lanUrl: '',
				icon: null,
				openMethod: 0,
				// IconJson在后端被标记为json:"-",不参与JSON序列化
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
				await refreshBookmarks();
				ms.success('创建成功');
				isEditDialogOpen.value = false;
				isCreateMode.value = false;
			} else {
				ms.error(`创建失败: ${createResponse?.msg || '未知错误'}`);
			}
		} else {
			// 修改模式
			// 使用update接口更新书签
			const updateResponse = await updateBookmark({
				id: Number(currentEditBookmark.value.id),
				title: currentEditBookmark.value.title,
				url: currentEditBookmark.value.url,
				parentUrl: currentEditBookmark.value.folderId ? currentEditBookmark.value.folderId.toString() : '0',
				sort: 9999,
				lanUrl: '',
				icon: null,
				openMethod: 0,
				// IconJson在后端被标记为json:"-",不参与JSON序列化
			});

				// 检查响应状态
				if (updateResponse && updateResponse.code === 0) {
					await refreshBookmarks();
					ms.success('书签修改成功');
					isEditDialogOpen.value = false;
				} else {
					ms.error(`书签修改失败: ${updateResponse?.msg || '未知错误'}`);
				}
			}
	} catch (error) {
		console.error('保存书签失败', error);
		ms.error(`书签修改失败: ${(error as Error).message || '网络错误'}`);
	}
}

// 删除书签或文件夹
async function deleteBookmark(bookmark: Bookmark) {
	// 根据是否为文件夹显示不同的确认消息
	const confirmMessage = bookmark.isFolder
		? `确定要删除文件夹 "${bookmark.title}" 吗？删除后，该文件夹下的所有内容也将被删除。`
		: `确定要删除书签 "${bookmark.title}" 吗？`;

	// 直接使用从apiMessage导入的dialog对象
	dialog.warning({
		title: '确认删除',
		content: confirmMessage,
		positiveText: '确定',
		negativeText: '取消',
		onPositiveClick: async () => {
			 try {
				 const response = await deletes([Number(bookmark.id)]);
				 if (response.code === 0) {
					 // 清除书签缓存
					 ss.remove(BOOKMARKS_CACHE_KEY)
					 ms.success('删除成功');
					 // 刷新书签列表
					 await refreshBookmarks();
				 } else {
					 ms.error(`删除失败: ${response.msg}`);
				 }
			 } catch (error) {
				 ms.error(`删除失败: ${(error as Error).message || '未知错误'}`);
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
		const response = await addMultipleBookmarks({ htmlContent } as any);
		if (response.code === 0) {
			ms.success(`${t('common.success')}，成功导入 ${(response.data as any).count} 个书签`);
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
		// 1. 首先尝试从缓存读取数据
		const cachedData = ss.get(BOOKMARKS_CACHE_KEY)
		if (cachedData) {
			bookmarkTree.value = cachedData
			return
		}

		// 2. 缓存中没有数据，请求接口获取数据
		const response = await getBookmarksList();
		if (response.code === 0) {
			// 检查数据结构，如果已经是树形结构则直接使用
			const data: any = response.data || [];

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

			// 3. 将数据永久保存到缓存中
			ss.set(BOOKMARKS_CACHE_KEY, treeData)
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
		children: [] as any[], // 明确指定类型为 any[]
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


	return rootFolders;
}
const handleResize = () => {
	isMobile.value = window.innerWidth < 768
	// 移动端默认隐藏左栏
	showLeftPanel.value = isMobile.value ? false : true
}
// 组件挂载时加载书签
onMounted(async () => {
	await refreshBookmarks();
	// 添加全局事件监听器
	document.addEventListener('mousemove', handleMouseMove);
	document.addEventListener('mouseup', stopResize);
	document.addEventListener('click', handleGlobalClick);

	handleResize()
	window.addEventListener('resize', handleResize)
});

// 组件卸载时移除事件监听器
onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
	document.removeEventListener('mouseup', stopResize);
	document.removeEventListener('click', handleGlobalClick);

	window.removeEventListener('resize', handleResize)
});



</script>

<style>
.context-menu {
position: fixed !important;
z-index: 99999 !important;
background-color: white;
}
/* 移动端左栏滑动过渡 */
.left-panel-transition {
	transition: transform 0.3s ease-in-out;
}

/* 阴影和圆角让左栏浮动更美观 */
.fixed.bg-white {
	border-right: 1px solid #e5e7eb; /* 柔和边框 */
	border-radius: 0 0.75rem 0.75rem 0; /* 左边圆角 */
	box-shadow: 0 4px 12px rgba(0,0,0,0.1);
	transition: width 0.3s ease-in-out, transform 0.3s ease-in-out;
}


</style>
