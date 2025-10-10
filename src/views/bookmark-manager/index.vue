<template>
	<div class="flex flex-col h-screen bg-white">
		<!-- 顶部标题栏（左右两侧共同的顶部） -->
		<div class="p-4 border-b flex items-center justify-between bg-gray-50">
			<NButton
				@click="goBackToHome"
				ghost
				circle
				size="small"
				icon="lucide:arrow-left"
				class="hover:bg-gray-200 transition-colors"
			/>
			<h1 class="text-xl font-bold text-gray-800 flex-1 text-center">书签管理</h1>

		</div>

		<!-- 主内容区域 -->
		<div class="flex flex-1 overflow-hidden">
			<!-- 左侧书签树 -->
			<div class="w-64 border-r p-2 overflow-auto">
				<n-tree
					:data="bookmarkTree"
					:default-expand-all="true"
					block-line
					@update:selected-keys="handleSelect"
				/>
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
				<div class="flex-1 overflow-auto p-2">
					<n-data-table
						:columns="columns"
						:data="filteredBookmarks"
						:bordered="false"
						size="small"
						@row-contextmenu="handleRowContextMenu"
					/>
				</div>
			</div>
		</div>

		<!-- 右键菜单 -->
		<n-dropdown
			trigger="manual"
			placement="bottom-start"
			:x="menuX"
			:y="menuY"
			:options="menuOptions"
			:show="showMenu"
			@clickoutside="showMenu = false"
			@select="handleMenuSelect"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { NTree, NDataTable, NInput, NDropdown, NButton } from 'naive-ui'
import { useRouter } from 'vue-router'

const router = useRouter()

// 返回首页
function goBackToHome() {
  router.push('/')
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

const menuOptions = [
	{ label: '新窗口打开', key: 'open' },
	{ label: '修改', key: 'edit' },
	{ label: '删除', key: 'delete' }
]

function handleRowContextMenu(row: Bookmark, e: MouseEvent) {
	e.preventDefault()
	selectedRow.value = row
	menuX.value = e.clientX
	menuY.value = e.clientY
	showMenu.value = true
}

function handleMenuSelect(key: string) {
	showMenu.value = false
	if (!selectedRow.value) return
	switch (key) {
		case 'open':
			window.open(selectedRow.value.url, '_blank')
			break
		case 'edit':
			alert('编辑功能待实现: ' + selectedRow.value.title)
			break
		case 'delete':
			alert('删除功能待实现: ' + selectedRow.value.title)
			break
	}
}
</script>
