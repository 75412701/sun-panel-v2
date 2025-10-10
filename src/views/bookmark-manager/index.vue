<template>
	<div class="flex flex-col h-screen bg-white">
		<!-- 顶部标题栏（左右两侧共同的顶部） -->
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
				class="flex items-center gap-1 text-green-600 hover:bg-green-100 transition-colors border border-green-200"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
				</svg>
				导入书签
			</NButton>
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
				<!-- 书签表格 -->
				<div class="flex-1 p-4 relative">
					<div class="grid grid-cols-1 gap-2">
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
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { NTree, NDataTable, NInput, NDropdown, NButton, NMenu } from 'naive-ui'
import { useRouter } from 'vue-router'

const router = useRouter()

// 左侧面板宽度
const leftPanelWidth = ref(256) // 默认256px
const isResizing = ref(false)

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
	folderId?: number
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
const selectedFolder = ref<number | null>(null)

// 所有书签扁平化
const allBookmarks = computed<Bookmark[]>(() => {
	const result: Bookmark[] = []
	function traverse(nodes: any[], parentId?: number) {
		for (const node of nodes) {
			if (node.isLeaf && node.bookmark) {
				result.push(node.bookmark)
			} else if (node.children) {
				traverse(node.children, node.key)
			}
		}
	}
	traverse(bookmarkTree.value)
	return result
})

// 搜索
const searchQuery = ref('')
const filteredBookmarks = computed(() => {
	if (searchQuery.value) {
		return allBookmarks.value.filter(b =>
			b.title.includes(searchQuery.value) || b.url.includes(searchQuery.value)
		)
	}
	return allBookmarks.value.filter(b => b.folderId === selectedFolder.value)
})

const columns = [
	{ title: '标题', key: 'title' },
	{ title: '网址', key: 'url' }
]

// 点击树节点
function handleSelect(keys: (string | number)[]) {
	selectedFolder.value = typeof keys[0] === 'number' ? keys[0] : null
}

// 搜索时清空选中
function handleSearch() {
	if (searchQuery.value) {
		selectedFolder.value = null
	}
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


</script>
